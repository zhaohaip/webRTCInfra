package sdp

import (
	"webRTCInfra/pkg/common"
	netwebsocket "webRTCInfra/pkg/network/websocket"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Service SDP业务服务
type Service struct {
	connMgr  *netwebsocket.Manager
	signaler *Signaler // 信令处理器
}

func NewService(mgr *netwebsocket.Manager) *Service {
	return &Service{
		connMgr:  mgr,
		signaler: NewSignaler(mgr),
	}
}

func (s *Service) RegistryClient(userName string, conn *websocket.Conn) string {
	// 生成唯一客户端ID
	clientId := uuid.NewString()
	// 创建网络层连接
	wsConn := netwebsocket.NewConnection(userName, clientId, conn, s.connMgr.RemoveClient)
	// 设置消息回调，交给信令处理器进行处理
	wsConn.SetOnMessage(s.signaler.HandleMessage)
	// 将客户端添加到管理器
	s.connMgr.AddClient(wsConn)

	go wsConn.ReadLoop()
	go wsConn.WriteLoop()

	return clientId
}

func (s *Service) ListClients() []common.ClientMessage {
	clients := s.connMgr.ListClients()
	clientList := make([]common.ClientMessage, 0)
	for _, client := range clients {
		clientList = append(clientList, common.ClientMessage{
			Name: client.UserName,
			ID:   client.UserID,
		})
	}
	return clientList
}
