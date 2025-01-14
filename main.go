package main

import (
	"fmt"
	"io"
	"os"
	"qw-robot-server/common/config"
	"qw-robot-server/common/log"
	"qw-robot-server/common/websocket"
	"qw-robot-server/router"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hello World")
	config.Init()
	log.GetLogger().Info("Hello World")

	go websocket.Init(config.GetConf("websocket.port"))

	fmt.Println("启动http服务")
	httpLog := config.GetConf("log.dir") + "/middleware.log"
	f, _ := os.Create(httpLog)
	gin.DefaultWriter = io.MultiWriter(f)
	e := gin.Default()

	// 初始化路由
	fmt.Println("初始化路由")
	router.Init(e)
	err := e.Run(":" + config.GetConf("http.port"))
	if err != nil {
		panic(err)
	}
}
