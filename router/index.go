package router

import (
	"bytes"
	"io"
	"net/http"
	"qw-robot-server/common/log"
	"qw-robot-server/handler"

	"github.com/gin-gonic/gin"
)

// 添加中间件函数来记录请求信息
func logRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		// 获取请求URL
		requestURL := c.Request.URL.String()

		// 读取原始请求数据
		rawData, _ := c.GetRawData()
		// 由于读取了raw data，需要重新设置body供后续中间件使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))

		// 获取所有请求参数
		params := c.Request.URL.Query()

		// 记录请求信息
		log.GetLogger().Infof("Client IP: %s, URL: %s, Params: %v, Raw Data: %s",
			clientIP, requestURL, params, string(rawData))

		c.Next()
	}
}

func Init(engine *gin.Engine) {
	route := engine.Group("/api")
	// 添加日志中间件
	route.Use(logRequestMiddleware())

	route.Static("/apidoc", "apidoc")

	route.GET("/healthcheck", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "consul health check ok .")
	})
	// route.Use(middleware.VisitLog())

	//内部服务接口，暂时不验证
	// serviceRoutes := route.Group("/service")
	// service.InitServiceRouter(serviceRoutes)

	route.POST("/sendMessageToUser", handler.SendMessageToUser)
	route.POST("/sendMessageToAll", handler.SendMessageToAll)
}
