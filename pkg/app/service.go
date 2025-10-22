package app

import (
	"log"
	"net/http"
	"signalingServer/pkg/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Service struct {
	upGrader websocket.Upgrader // 定义 WebSocket Upgrader，用于把普通 HTTP 请求升级为 WebSocket 连接
	cm       *ClientManager
}

func NewService() *Service {
	cm := NewClientManager()
	return &Service{
		cm: cm,
		upGrader: websocket.Upgrader{
			ReadBufferSize:  1024, // 读取缓冲区大小
			WriteBufferSize: 1024, // 写入缓冲区大小
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Service) WebsocketSignalHandler(c *gin.Context) {
	userName := c.Query("name")
	if len(userName) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing username"})
		return
	}

	conn, err := s.upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade failed, error: ", err)
		return
	}

	// 创建一个新客户端，并将客户端加入
	clientId := uuid.NewString()
	client := NewClient(userName, clientId, conn)
	s.cm.AddClient(client)

	go client.WsReadLoop(s.cm)
	go client.WsWriteLoop()

	// 将clientID返回给客户端
	res, _ := common.NewWebsocketServiceResponse(clientId, common.SignallingTypeRegister, nil)
	if err = conn.WriteMessage(websocket.TextMessage, res); err != nil {
		log.Println("[WebsocketSignalHandler] send clientID failed")
	}
}

func (s *Service) ListSignalClients(c *gin.Context) {
	clients := s.cm.ListClients()

	clientList := make([]common.ClientMessage, 0)
	for _, client := range clients {
		clientList = append(clientList, common.ClientMessage{Name: client.UserName, ID: client.UserID})
	}

	c.JSON(http.StatusOK, common.ListClientsResponse{Clients: clientList})
}
