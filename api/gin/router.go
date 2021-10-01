package gin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRouter() (router *gin.Engine) {
	router = gin.Default()
	router.Use(Cors())
	//展示所有的
	router.GET(`/api/redis/list`, RedisList)
	//查询某个key
	router.POST(`/api/search`, Search)
	//模糊搜索key
	router.POST(`/api/keys`, Keys)
	//获取key类型
	router.POST(`/api/key/type`, GetKeyType)
	//序列化和反序列化
	router.POST(`/api/serialize`, PhpSerialize)
	router.POST(`/api/unserialize`, PhpUnSerialize)
	//保存string
	router.POST(`/api/save/string`, SaveString)
	//删除key
	router.POST(`/api/del/key`, DelKey)
	//更改ttl
	router.POST(`/api/edit/ttl`, EditTtl)
	return router
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
