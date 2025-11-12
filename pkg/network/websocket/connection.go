package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

// 连接回调函数定义

type OnMessageFunc func(userID string, data []byte) // 收到消息时回调
type OnCloseFunc func(userID string)                // 连接关闭时回调

// Connection 封装单个WebSocket连接，仅处理网络读写
type Connection struct {
	Conn      *websocket.Conn
	UserName  string
	UserID    string
	SendMsg   chan []byte
	OnMessage OnMessageFunc
	OnClose   OnCloseFunc
}

func NewConnection(userName string, userID string, conn *websocket.Conn, onClose OnCloseFunc) *Connection {
	return &Connection{
		Conn:     conn,
		UserName: userName,
		UserID:   userID,
		SendMsg:  make(chan []byte, 32),
		OnClose:  onClose,
	}
}

func (c *Connection) SetOnMessage(fn OnMessageFunc) {
	c.OnMessage = fn
}

// ReadLoop 读取WebSocket消息（仅负责读，不解析业务）
func (c *Connection) ReadLoop() {
	defer func() {
		c.OnClose(c.UserID) // 通知管理器移除该客户端
		c.Conn.Close()
		close(c.SendMsg)
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("user [%s/%s] read error: %v", c.UserName, c.UserID, err)
			break
		}

		if c.OnMessage != nil {
			c.OnMessage(c.UserID, data)
		}
	}

}

// WriteLoop 发送消息（仅负责写，不处理业务）
func (c *Connection) WriteLoop() {
	for {
		for msg := range c.SendMsg {
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("user [%s/%s] write error: %v", c.UserName, c.UserID, err)
				return
			}
		}
	}
}

func (c *Connection) Send(data []byte) bool {
	select {
	case c.SendMsg <- data:
		return true
	default:
		log.Printf("user [%s/%s] send queue full", c.UserName, c.UserID)
		return false
	}
}
