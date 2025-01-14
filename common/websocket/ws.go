package websocket

import (
	"context"
	"errors"
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

type Client struct {
	conn      *websocket.Conn
	messageCh chan []byte
	done      chan struct{}
	sync.Mutex
}

type clients struct {
	conn map[string]*Client
	sync.Mutex
}

var accounts = make(map[string]*clients)
var accountsMu sync.Mutex

var (
	connStatusMu sync.Mutex
	connStatus   = make(map[*websocket.Conn]time.Time)
)

type message struct {
	Code    string `json:"code"`
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

type license struct {
	Id             string `json:"id" bson:"_id"`
	Account        string `json:"account"`                              //公司id
	Product        string `json:"product"`                              //产品id
	Price          int64  `json:"price"`                                //价格int
	Num            int64  `json:"num"`                                  //数量
	Type           string `json:"type"`                                 //try试用、use正式
	Begin          string `json:"begin"`                                //开始时间
	End            string `json:"end"`                                  //到期时间
	LastUpdateTime string `json:"lastUpdateTime" bson:"lastUpdateTime"` //最近更新时间
	CreateUser     string `json:"createUser" bson:"createUser"`         //创建人
	CreateTime     string `json:"createTime" bson:"createTime"`         //创建时间
}

func Init(port string) {
	// 添加定期检查连接状态
	go func() {
		ticker := time.NewTicker(time.Minute) // 每分钟检查一次
		defer ticker.Stop()

		for range ticker.C {
			checkConnStatus()
		}
	}()

	// 设置 websocket 路由
	http.HandleFunc("/ws", handleConnections)

	// 启动 HTTP 服务器
	log.GetLogger().Infof("WebSocket server starting on port: " + port)

	addr := ":" + port
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.GetLogger().Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	log.GetLogger().Infof("handleConnections: %v", r)
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

	client := &Client{
		conn:      ws,
		messageCh: make(chan []byte, 100),
		done:      make(chan struct{}),
	}

	go handleMessages(client)

	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accounts[account] = &clients{
			conn: make(map[string]*Client),
		}
	}
	accounts[account].Lock()
	if oldClient, exists := accounts[account].conn[user_id]; exists {
		// 创建一个带超时的 context
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// 使用 goroutine 发送消息
		go func() {
			done := make(chan struct{})
			go func() {
				oldClient.conn.WriteJSON(message{
					Code:    "relogin",
					UserId:  user_id,
					Message: "当前账户已经在别的地方登录了",
				})
				close(done)
			}()

			select {
			case <-ctx.Done():
				log.GetLogger().Warn("Send relogin message timeout")
			case <-done:
				log.GetLogger().Info("Relogin message sent successfully")
			}
		}()

		close(oldClient.done)
		oldClient.conn.Close()
	}
	accounts[account].conn[user_id] = client
	accounts[account].Unlock()
	accountsMu.Unlock()
	connStatusMu.Lock()
	connStatus[ws] = time.Now()
	connStatusMu.Unlock()

	ws.SetPingHandler(func(appData string) error {
		log.GetLogger().Infof("SetPingHandler: %v", appData)
		connStatusMu.Lock()
		defer connStatusMu.Unlock()

		if err := ws.WriteControl(websocket.PongMessage, []byte("Pong"),
			time.Now().Add(time.Second)); err != nil {
			log.GetLogger().Errorf("Failed to send pong: %v", err)
		}

		connStatus[ws] = time.Now()
		return nil
	})

	ws.SetCloseHandler(func(code int, text string) error {
		log.GetLogger().Infof("SetCloseHandler: %v", code)
		removeConnection(ws, account, user_id)
		return nil
	})

	ws.WriteJSON(message{
		Code:    "login",
		UserId:  user_id,
		Message: "登录成功",
	})

	defer ws.Close()
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.GetLogger().Infof("ReadMessage error: %v", err)
			removeConnection(ws, account, user_id)
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
	// url := config.GetConf("bill.api") + "account/getLicense?product=qw-robot"
	// data := map[string]string{
	// 	"account": account,
	// 	"product": "qw-robot",
	// }
	// jsonData, err := json.Marshal(data)
	// response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	log.GetLogger().Errorf("Failed to get license: %v", err)
	// 	return false
	// }
	// defer response.Body.Close()
	// license := license{}
	// json.NewDecoder(response.Body).Decode(&license)

	// // 检查license是否过期
	// if license.End < time.Now().Format("2006-01-02") {
	// 	log.GetLogger().Errorf("License expired: %v", license)
	// 	return false
	// }

	// // 检查当前账户下的连接数是否超过license数量
	// accountsMu.Lock()
	// count := len(accounts[account].conn)
	// accountsMu.Unlock()
	// if count >= int(license.Num) {
	// 	log.GetLogger().Errorf("License count exceeded: %v", license)
	// 	return false
	// }

	// Example authentication logic - replace with your actual authentication system
	return true // temporary return for demonstration
}

func SendMessageToUser(account, user_id, message string) error {
	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accountsMu.Unlock()
		return errors.New("account not found")
	}
	accounts[account].Lock()
	client, exists := accounts[account].conn[user_id]
	accounts[account].Unlock()
	accountsMu.Unlock()

	if !exists {
		return errors.New("user not found")
	}

	select {
	case client.messageCh <- []byte(message):
		log.GetLogger().Info("SendMessageToUser: " + message)
		return nil
	default:
		return errors.New("message channel full")
	}
}

func SendMessageToAll(account string, userIds []string, message string) error {
	accountsMu.Lock()
	if _, exists := accounts[account]; !exists {
		accountsMu.Unlock()
		return errors.New("account not found")
	}
	accounts[account].Lock()
	clients := make([]*Client, 0, len(accounts[account].conn))
	for _, client := range accounts[account].conn {
		clients = append(clients, client)
	}
	accounts[account].Unlock()
	accountsMu.Unlock()

	msg := []byte(message)
	for _, client := range clients {
		select {
		case client.messageCh <- msg:
		default:
			log.GetLogger().Errorf("Failed to queue message: channel full")
		}
	}

	return nil
}

func checkConnStatus() {
	timeout := time.Minute // 1分钟超时

	connStatusMu.Lock()
	now := time.Now()
	var expiredConns []*websocket.Conn

	// 找出所有超时的连接
	for conn, lastActive := range connStatus {
		if now.Sub(lastActive) > timeout {
			expiredConns = append(expiredConns, conn)
		}
	}
	connStatusMu.Unlock()

	// 处理超时的连接
	for _, conn := range expiredConns {
		// 关闭连接
		conn.Close()

		// 从 connStatus 中移除
		connStatusMu.Lock()
		delete(connStatus, conn)
		connStatusMu.Unlock()

		// 从 accounts 中移除
		accountsMu.Lock()
		for _, clients := range accounts {
			clients.Lock()
			for userID, c := range clients.conn {
				if c.conn == conn {
					delete(clients.conn, userID)
					log.GetLogger().Infof("Removed expired connection for user: %s", userID)
				}
			}
			clients.Unlock()
		}
		accountsMu.Unlock()
	}

	if len(expiredConns) > 0 {
		log.GetLogger().Infof("Cleaned up %d expired connections", len(expiredConns))
	}
}

// removeConnection 从所有相关的数据结构中移除连接
func removeConnection(conn *websocket.Conn, account, userID string) {
	// 从 connStatus 移除
	connStatusMu.Lock()
	delete(connStatus, conn)
	connStatusMu.Unlock()

	// 从 accounts 中移除
	accountsMu.Lock()
	if clients, exists := accounts[account]; exists {
		clients.Lock()
		if c, ok := clients.conn[userID]; ok && c.conn == conn {
			delete(clients.conn, userID)
			log.GetLogger().Infof("Remove connection for user: %s in account: %s", userID, account)
		}
		clients.Unlock()
	}
	accountsMu.Unlock()
}

func handleMessages(client *Client) {
	for {
		select {
		case msg := <-client.messageCh:
			client.Lock()
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			client.Unlock()
			if err != nil {
				log.GetLogger().Errorf("Failed to send message: %v", err)
				return
			}
		case <-client.done:
			return
		}
	}
}
