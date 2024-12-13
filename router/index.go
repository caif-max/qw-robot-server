package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine) {
	route := engine.Group("/api")
	route.Static("/apidoc", "apidoc")

	route.GET("/healthcheck", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "consul health check ok .")
	})
	// route.Use(middleware.VisitLog())

	//内部服务接口，暂时不验证
	// serviceRoutes := route.Group("/service")
	// service.InitServiceRouter(serviceRoutes)
}
