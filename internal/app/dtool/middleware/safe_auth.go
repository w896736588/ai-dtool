package middleware

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// SafeAuthWhiteList 不需要鉴权的接口白名单
var SafeAuthWhiteList = map[string]bool{
	"/api/BaseLogin":            true,
	"/api/BaseLoginStatus":      true,
	"/api/Ip":                   true,
	"/api/BaseRegisterService":  true,
	"/api/BaseCheckUnikeyExist": true,
	"/api/Upload":               true, // 上传接口暂时放行，避免阻塞
}

// SafeAuthPathPrefixSkips 按前缀放行的接口（任务 10：录制 iframe 链路自洽）。
// 仅限录制专用接口，其他未授权接口严禁加入此白名单。
var SafeAuthPathPrefixSkips = []string{
	"/api/e2e/recorder/",       // iframe proxy.html / 后续静态资源
	"/api/e2e/record/by_token/", // ws_token 鉴权由 RecorderTokenAuthMiddleware 处理
}

// shouldSkipSafeAuth 判断请求是否在白名单或前缀放行列表中。
func shouldSkipSafeAuth(path string) bool {
	if SafeAuthWhiteList[path] {
		return true
	}
	for _, prefix := range SafeAuthPathPrefixSkips {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// getSafeTokenManager 创建 Safe Token 管理器（从配置读取）
func getSafeTokenManager() *common.SafeTokenManager {
	password := component.ConfigViper.GetString("safe.password")
	appName := component.ConfigViper.GetString("app.name")
	return common.NewSafeTokenManager(password, appName)
}

// SafeAuthContextKey 存储在 gin.Context 中的认证信息 key
const SafeAuthContextKey = "safe_auth_claims"

// SafeAuthMiddleware Safe 鉴权中间件
func SafeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 检查是否是白名单接口
		path := c.Request.URL.Path
		isSsePath := strings.HasPrefix(path, "/sse")
		if shouldSkipSafeAuth(path) {
			if isSsePath {
				gstool.FmtPrintlnLogTime(`[SSE-Auth] 白名单放行 path=%s`, path)
			}
			c.Next()
			return
		}

		// 2. 读取当前配置
		tokenManager := getSafeTokenManager()

		// 3. 如果未启用密码保护，直接放行
		if !tokenManager.IsEnabled() {
			if isSsePath {
				gstool.FmtPrintlnLogTime(`[SSE-Auth] 未启用密码保护，直接放行 path=%s`, path)
			}
			c.Next()
			return
		}

		// 4. 从请求头获取 Token，如果没有则从 URL 查询参数获取（兼容 SSE）
		token := c.GetHeader("Token")
		tokenSource := "header"
		if token == "" {
			// 尝试从 Cookie 获取（兼容某些场景）
			token, _ = c.Cookie("safe_token")
			if token != "" {
				tokenSource = "cookie"
			}
		}
		if token == "" {
			// 尝试从 URL 查询参数获取（SSE 场景）
			token = c.Query("token")
			if token != "" {
				tokenSource = "query"
			}
		}

		if isSsePath {
			clientId := c.Query("client_id")
			gstool.FmtPrintlnLogTime(`[SSE-Auth] path=%s client_id=%s token_source=%s token_len=%d password_enabled=%v`, path, clientId, tokenSource, len(token), true)
		}

		// 5. 解析并验证 Token
		claims, errCode, err := tokenManager.ParseToken(token)
		if err != nil {
			if isSsePath {
				gstool.FmtPrintlnLogTime(`[SSE-Auth] Token解析失败 path=%s errCode=%d err=%s`, path, errCode, err.Error())
			}
			respondAuthError(c, errCode, err.Error())
			c.Abort()
			return
		}

		// 6. 将认证信息存入 Context，便于后续使用
		c.Set(SafeAuthContextKey, claims)

		if isSsePath {
			gstool.FmtPrintlnLogTime(`[SSE-Auth] Token验证通过 path=%s`, path)
		}

		c.Next()
	}
}

// respondAuthError 返回鉴权错误响应
func respondAuthError(c *gin.Context, errCode int, errMsg string) {
	// 对于密码版本不匹配(40103)或登录过期(40102)的错误，推送 SSE 事件通知前端弹窗
	if errCode == 40103 || errCode == 40102 {
		go sendSafeAuthRequiredSSE(c, errCode)
	}

	// 统一错误响应格式
	c.JSON(http.StatusOK, map[string]interface{}{
		"ErrCode": errCode,
		"ErrMsg":  getAuthErrorMessage(errCode, errMsg),
		"Data":    nil,
	})
}

// sendSafeAuthRequiredSSE 发送安全认证失效 SSE 事件到对应客户端
func sendSafeAuthRequiredSSE(c *gin.Context, errCode int) {
	defer func() {
		// 忽略 panic，避免影响主流程
		recover()
	}()

	// 获取 SSE client_id
	clientId := c.GetHeader("SseClientId")
	if clientId == "" {
		return
	}

	// 获取 SSE 连接
	sse := gsgin.SseGetByClientId(clientId)
	if sse == nil {
		return
	}

	// 构造消息
	message := "登录态已过期，请重新登录"
	if errCode == 40103 {
		message = "登录态与当前密码不匹配，请重新登录"
	}

	data := map[string]any{
		"message": message,
		"time":    gstool.TimeNowUnixToString(`Y-m-d H:i:s`),
	}

	sseMsg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseSafeAuthRequired,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	})

	// 发送消息（非阻塞）
	_ = sse.SendToChan(sseMsg)
}

// getAuthErrorMessage 获取错误码对应的标准错误消息
func getAuthErrorMessage(errCode int, defaultMsg string) string {
	switch errCode {
	case 40101:
		return "未登录或 token 缺失"
	case 40102:
		return "登录态已过期，请重新登录"
	case 40103:
		return "登录态与当前密码不匹配，请重新登录"
	case 40104:
		return "登录态非法或解密失败，请重新登录"
	default:
		return defaultMsg
	}
}

// GetSafeAuthClaims 从 gin.Context 获取认证信息
func GetSafeAuthClaims(c *gin.Context) *common.SafeTokenClaims {
	claims, exists := c.Get(SafeAuthContextKey)
	if !exists {
		return nil
	}
	if c, ok := claims.(*common.SafeTokenClaims); ok {
		return c
	}
	return nil
}

// IsSafeAuthWhiteList 检查路径是否在白名单中
func IsSafeAuthWhiteList(path string) bool {
	// 去除可能的查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	return SafeAuthWhiteList[path]
}

// AddSafeAuthWhiteList 添加白名单路径（用于动态扩展）
func AddSafeAuthWhiteList(path string) {
	SafeAuthWhiteList[path] = true
}
