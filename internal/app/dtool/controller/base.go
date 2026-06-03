package controller

import (
	"context"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// getSafeTokenManager 创建 Safe Token 管理器（从配置读取）
func getSafeTokenManager() *common.SafeTokenManager {
	password := component.ConfigViper.GetString("safe.password")
	appName := component.ConfigViper.GetString("app.name")
	return common.NewSafeTokenManager(password, appName)
}

func splitConfiguredPorts(raw string) []string {
	portList := strings.Split(raw, `,`)
	ret := make([]string, 0, len(portList))
	seen := make(map[string]bool)
	for _, port := range portList {
		port = strings.TrimSpace(port)
		if port == `` || seen[port] {
			continue
		}
		seen[port] = true
		ret = append(ret, port)
	}
	return ret
}

func getConfiguredAPIPorts(cfg *viper.Viper) []string {
	apiPorts := splitConfiguredPorts(cfg.GetString(`run.api_port`))
	if len(apiPorts) > 0 {
		return apiPorts
	}
	return splitConfiguredPorts(cfg.GetString(`run.ports`))
}

func getConfiguredSsePorts(cfg *viper.Viper) []string {
	ssePorts := splitConfiguredPorts(cfg.GetString(`run.sse_port`))
	if len(ssePorts) > 0 {
		return ssePorts
	}
	apiPorts := getConfiguredAPIPorts(cfg)
	if len(apiPorts) == 0 {
		return nil
	}
	return []string{apiPorts[0]}
}

func buildLoginStatusData(cfg *viper.Viper, enabled bool, loggedIn bool, expireAt int64) map[string]any {
	return map[string]any{
		`enabled`:   enabled,
		`logged_in`: loggedIn,
		`expire_at`: expireAt,
		`api_port`:  getConfiguredAPIPorts(cfg),
		`sse_ports`: getConfiguredSsePorts(cfg),
	}
}

func buildLoginData(cfg *viper.Viper, enabled bool, token string, expireAt int64) map[string]any {
	return map[string]any{
		`enabled`:   enabled,
		`token`:     token,
		`expire_at`: expireAt,
		`api_port`:  getConfiguredAPIPorts(cfg),
		`sse_ports`: getConfiguredSsePorts(cfg),
		`local_ip`:  GetLANIP(),
	}
}

// BaseLogin Safe 登录接口
// 使用配置文件中的 safe.password 进行验证
func BaseLogin(c *gin.Context) {
	// 获取请求参数（兼容旧字段和新字段）
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseError(c, `请求参数错误`, map[string]any{
			`token`: ``,
		})
		return
	}

	// 获取输入密码（兼容 password 和 Password 字段）
	inputPassword := ``
	if p, ok := reqMap[`password`]; ok {
		inputPassword = cast.ToString(p)
	} else if p, ok := reqMap[`Password`]; ok {
		inputPassword = cast.ToString(p)
	}

	// 创建 Token 管理器
	tokenManager := getSafeTokenManager()

	// 检查是否启用了密码保护
	if !tokenManager.IsEnabled() {
		gsgin.GinResponseSuccess(c, `未启用密码保护，无需登录`, buildLoginData(component.ConfigViper, false, ``, 0))
		return
	}

	// 验证密码
	if !tokenManager.VerifyPassword(inputPassword) {
		gsgin.GinResponseError(c, `密码错误`, map[string]any{
			`token`: ``,
		})
		return
	}

	// 生成 Token
	token, expireAt, tokenErr := tokenManager.GenerateToken()
	if tokenErr != nil {
		gsgin.GinResponseError(c, `登录失败（`+tokenErr.Error()+`）`, map[string]any{
			`token`: ``,
		})
		return
	}

	gsgin.GinResponseSuccess(c, `登录成功`, buildLoginData(component.ConfigViper, true, token, expireAt))
}

// BaseLoginStatus 检查登录状态接口
// 前端启动时调用，判断是否需要弹出登录框
func BaseLoginStatus(c *gin.Context) {
	// 从请求头获取 Token
	token := c.GetHeader("Token")

	// 创建 Token 管理器
	tokenManager := getSafeTokenManager()

	// 检查是否启用了密码保护
	if !tokenManager.IsEnabled() {
		gsgin.GinResponseSuccess(c, `获取成功`, buildLoginStatusData(component.ConfigViper, false, true, 0))
		return
	}

	// 解析并验证 Token
	claims, errCode, _ := tokenManager.ParseToken(token)

	isLoggedIn := errCode == 0
	expireAt := int64(0)
	if claims != nil {
		expireAt = claims.ExpireAt
	}

	gsgin.GinResponseSuccess(c, `获取成功`, buildLoginStatusData(component.ConfigViper, true, isLoggedIn, expireAt))
}

func GetLANIP() string {
	// 获取主机的所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		gstool.FmtPrintlnLogTime(`%s`, `获取主机的所有网络接口失败:`+err.Error())
		return ""
	}

	for _, iface := range interfaces {
		// 添加一些过滤逻辑，例如跳过 down 掉的网卡
		if iface.Flags&net.FlagUp == 0 {
			continue // 接口未启用
		}
		// 跳过回环接口
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取接口下的所有地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 检查 IP 是否为 nil，且只处理 IPv4 (如果你想兼容 IPv6，可去掉 !ip.To4() 判断)
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// 过滤 IPv6，只保留 IPv4
			if ip = ip.To4(); ip == nil {
				continue
			}

			return ip.String()
		}
	}
	return ""
}

// BaseCheckUnikeyExist 检查是否需要登录
func BaseCheckUnikeyExist(c *gin.Context) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		gsgin.GinResponseSuccess(c, err.Error(), nil)
		return
	}
	reqConsMap := gstool.ConsNewMap(reqMap)
	unikey := reqConsMap[`Unikey`]
	if unikey.IsEmpty() {
		gsgin.GinResponseSuccess(c, `Unikey不能为空`, nil)
		return
	}

	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`NeedLogin`: `0`,
	})
}

// GetLocalIP 获取局域网IP
func GetLocalIP(c *gin.Context) {
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]any{
		`local_ip`: GetLANIP(),
	})
}

// BaseRegisterService 注册各类服务
func BaseRegisterService(c *gin.Context) {
	gsgin.GinResponseSuccess(c, `ok`, nil)
}

// GetGlobalReqParamsM 拿到全局参数 返回map
func GetGlobalReqParamsM(c *gin.Context) (map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, err
	}
	return reqMap, nil
}

func BaseRedisCheckKeyExist(redisCli *redis.Client, key string) error {
	//判断是否存在
	if existInt := redisCli.Exists(context.Background(), key).Val(); existInt <= 0 {
		return errors.New(fmt.Sprintf(`%s 不存在`, key))
	}
	return nil
}

func BaseResponseByError(c *gin.Context, err error) {
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), ``)
	} else {
		gsgin.GinResponseSuccess(c, ``, ``)
	}
}

func BaseSshList(c *gin.Context) {
	sshList, _ := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`ssh_list`: sshList,
	})
}

// Ip 外网IP
func Ip(c *gin.Context) {
	ip, _ := p_common.TBaseClient.GetPublicIPWithSTUN()
	gsgin.GinResponseSuccess(c, `获取成功`, map[string]string{
		`ip`: ip,
	})
}

func Upload(c *gin.Context) {
	file, err := c.FormFile(`file`)
	if err != nil {
		gsgin.GinResponseError(c, `上传失败:`+err.Error(), ``)
		return
	}
	uploadDir := filepath.Join(component.EnvClient.RootPath, `upload`)
	_ = gstool.DirCreatePath(uploadDir)
	// 生成新名字：时间戳+扩展名
	//ext := filepath.Ext(file.Filename)
	//newName := fmt.Sprintf("%d%s", time.Now().UnixMicro(), ext)
	dst := filepath.Join(uploadDir, file.Filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		gsgin.GinResponseError(c, `上传存储文件失败:`+err.Error(), ``)
		return
	}

	gsgin.GinResponseSuccess(c, `上传成功`, map[string]string{
		`url`: dst,
	})
}

// DownloadWebFile 下载 web/download 目录下的单个文件。
func DownloadWebFile(c *gin.Context) {
	fileName := strings.TrimSpace(c.Param(`name`))
	if !isValidWebDownloadFileName(fileName) {
		gsgin.GinResponseError(c, `文件名不合法`, ``)
		return
	}

	targetPath := buildWebDownloadFilePath(fileName)
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			gsgin.GinResponseError(c, `文件不存在`, ``)
			return
		}
		gsgin.GinResponseError(c, `读取文件失败:`+err.Error(), ``)
		return
	}
	if fileInfo.IsDir() {
		gsgin.GinResponseError(c, `文件不存在`, ``)
		return
	}

	c.FileAttachment(targetPath, fileName)
	c.Status(http.StatusOK)
}

// isValidWebDownloadFileName 校验下载文件名只能是 web/download 下的顶层文件。
func isValidWebDownloadFileName(fileName string) bool {
	if fileName == `` || fileName == `.` || fileName == `..` {
		return false
	}
	if strings.Contains(fileName, `/`) || strings.Contains(fileName, `\`) {
		return false
	}
	return filepath.Base(fileName) == fileName
}

// buildWebDownloadFilePath 拼接 web/download 目录中的目标文件路径。
func buildWebDownloadFilePath(fileName string) string {
	rootPath := ``
	if component.EnvClient != nil {
		rootPath = component.EnvClient.RootPath
	}
	return filepath.Join(rootPath, `web`, `download`, fileName)
}
