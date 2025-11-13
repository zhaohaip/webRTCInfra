package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

// Connection 封装单个WebSocket连接，仅处理网络读写
type Connection struct {
	Conn      *websocket.Conn
	UserName  string
	UserID    string
	SendMsg   chan []byte
	OnMessage func(userID string, data []byte)
	OnClose   func(userID string)
}

func NewConnection(userName string, userID string, conn *websocket.Conn, onClose func(userID string)) *Connection {
	return &Connection{
		Conn:     conn,
		UserName: userName,
		UserID:   userID,
		SendMsg:  make(chan []byte, 32),
		OnClose:  onClose,
	}
}

func (c *Connection) SetOnMessage(fn func(userID string, data []byte)) {
	c.OnMessage = fn
}

// ReadLoop 读取WebSocket消息（仅负责读，不解析业务）
func (c *Connection) ReadLoop() {
	defer func() {
		c.Conn.Close()
		close(c.SendMsg)
		if c.OnClose != nil {
			c.OnClose(c.UserID)
		}
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			// 区分正常关闭和异常情况
			if _, ok := err.(*websocket.CloseError); ok {
				log.Printf("user [%s/%s] connection closed normally: %v", c.UserName, c.UserID, err)
			} else {
				log.Printf("user [%s/%s] read error: %v", c.UserName, c.UserID, err)
			}
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
