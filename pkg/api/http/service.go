package http

import (
	"log"
	"net/http"
	"signalingServer/pkg/common"
	"signalingServer/pkg/service/sdp"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	upGrader   websocket.Upgrader // 定义 WebSocket Upgrader，用于把普通 HTTP 请求升级为 WebSocket 连接
	sdpService *sdp.Service
}

func NewHandler(sdpSvc *sdp.Service) *Handler {
	return &Handler{
		sdpService: sdpSvc,
		upGrader: websocket.Upgrader{
			ReadBufferSize:  1024, // 读取缓冲区大小
			WriteBufferSize: 1024, // 写入缓冲区大小
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Handler) WebsocketSignalHandler(c *gin.Context) {
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
	clientId := s.sdpService.RegistryClient(userName, conn)

	// 将clientID返回给客户端
	res, _ := common.NewWebsocketServiceResponse(clientId, common.SignallingTypeRegister, nil)
	if err = conn.WriteMessage(websocket.TextMessage, res); err != nil {
		log.Printf("send register response failed: %v", err)
		conn.Close()
	}
}

func (s *Handler) ListSignalClients(c *gin.Context) {
	clients := s.sdpService.ListClients()
	c.JSON(http.StatusOK, common.ListClientsResponse{Clients: clients})
}
