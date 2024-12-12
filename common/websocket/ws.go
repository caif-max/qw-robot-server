package websocket

import (
	"errors"
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

type clients struct {
	conn       map[string]*websocket.Conn
	sync.Mutex // Add mutex for thread safety
}

var accounts = make(map[string]*clients)
var accountsMu sync.Mutex

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
	// Add authentication check
	account := r.URL.Query().Get("account")
	user_id := r.URL.Query().Get("user_id")

	if !authenticateUser(account, user_id) {
		log.GetLogger().Error("Authentication failed")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.GetLogger().Error(err)
		return
	}

	// 检查是否有这个account的连接池
	// 如果没有，则创建一个空链接池
	// 检查这个account的连接池中是否有这个user_id的连接
	// 如果有，则关闭这个连接，然后创建一个新的连接
	// 如果没有，则创建一个新的连接
	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accounts[account] = &clients{
			conn: make(map[string]*websocket.Conn),
		}
	}
	accounts[account].Lock()
	if _, exists := accounts[account].conn[user_id]; !exists {
		accounts[account].conn[user_id] = ws
	} else {
		// todo 给旧连接发一个消息，告诉客户端当前账户已经做别的地方登录了
		log.GetLogger().Infof("User %s is already logged in on another device", user_id)
		accounts[account].conn[user_id].Close()
		accounts[account].conn[user_id] = ws
	}
	accounts[account].Unlock()
	accountsMu.Unlock()
	connStatusMu.Lock()
	connStatus[ws] = time.Now()
	connStatusMu.Unlock()

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
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.GetLogger().Error(err)
			return
		}
		// 处理消息...
		log.GetLogger().Infof("Received message: %d, %s", mt, string(message))

	}
}

// Add new authentication function
func authenticateUser(account, user_id string) bool {
	// TODO: Implement your actual authentication logic here
	if account == "" || user_id == "" {
		return false
	}

	// Example authentication logic - replace with your actual authentication system
	return true // temporary return for demonstration
}

func SendMessageToUser(account, user_id, message string) error {
	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accountsMu.Unlock()
		log.GetLogger().Errorf("Account %s not found", account)
		return errors.New("account not found")
	}
	accounts[account].Lock()
	if _, exists := accounts[account].conn[user_id]; !exists {
		accounts[account].Unlock()
		accountsMu.Unlock()
		log.GetLogger().Errorf("User %s not found", user_id)
		return errors.New("user not found")
	}
	accounts[account].conn[user_id].WriteMessage(websocket.TextMessage, []byte(message))
	accounts[account].Unlock()
	accountsMu.Unlock()
	return nil
}

func SendMessageToAll(account, message string) error {
	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accountsMu.Unlock()
		log.GetLogger().Errorf("Account %s not found", account)
		return errors.New("account not found")
	}
	accounts[account].Lock()
	for _, conn := range accounts[account].conn {
		conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
	accounts[account].Unlock()
	accountsMu.Unlock()
	return nil
}
