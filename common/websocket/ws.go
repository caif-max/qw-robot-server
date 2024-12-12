package websocket

import (
	"fmt"
	"net/http"
	"qw-robot-server/common/log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type client struct {
	conn map[string]*websocket.Conn
}

var clients = make(map[string]client)

var (
	connStatusMu sync.Mutex
	connStatus   = make(map[*websocket.Conn]time.Time)
)

func Init(port int) {
	// 设置 websocket 路由
	http.HandleFunc("/ws", handleConnections)

	// 启动 HTTP 服务器
	address := fmt.Sprintf(":%d", port)
	log.GetLogger().Infof("WebSocket server starting on port %d", port)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.GetLogger().Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.GetLogger().Error(err)
		return
	}

	ws.SetPingHandler(func(appData string) error {
		connStatusMu.Lock()
		defer connStatusMu.Unlock()

		if err := ws.WriteControl(websocket.PongMessage, []byte{},
			time.Now().Add(time.Second)); err != nil {
			log.GetLogger().Errorf("Failed to send pong: %v", err)
		}

		connStatus[ws] = time.Now()
		return nil
	})

	defer ws.Close()
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.GetLogger().Error(err)
			return
		}
		// 处理消息...
	}
}
