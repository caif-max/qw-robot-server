package main

import (
	"fmt"
	"qw-robot-server/common/config"
	"qw-robot-server/common/log"
	"qw-robot-server/common/websocket"
)

func main() {
	fmt.Println("Hello World")
	config.Init()
	log.GetLogger().Info("Hello World")

	websocket.Init(8080)
}
