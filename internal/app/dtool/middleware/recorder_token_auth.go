package middleware

import (
	"dev_tool/internal/app/dtool/component/e2e/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// RecorderTokenAuthMiddleware 校验 ws_token 查询参数（一次性 token），用于 /api/e2e/record/by_token/* 系列接口。
// 通过后把 token 与 row 注入 gin.Context，供下游 handler 直接用；会话已关闭（committed / discarded）则拒绝。
func RecorderTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("ws_token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ErrCode": http.StatusUnauthorized, "ErrMsg": "缺少 ws_token"})
			return
		}
		rs := store.NewRecordSessionStore()
		row, err := rs.FindByToken(token)
		if err != nil || row == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ErrCode": http.StatusUnauthorized, "ErrMsg": "ws_token 无效"})
			return
		}
		status := cast.ToString(row["status"])
		if status == "committed" || status == "discarded" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ErrCode": http.StatusUnauthorized, "ErrMsg": "会话已关闭"})
			return
		}
		c.Set("ws_token", token)
		c.Set("recorder_session_row", row)
		c.Next()
	}
}