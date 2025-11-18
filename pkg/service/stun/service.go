package stun

import (
	"log"
	"webRTCInfra/pkg/network/udp"
	"webRTCInfra/pkg/protocol/stun"
)

type Service struct {
	udpSvc *udp.Server
}

func NewService(udpSvc *udp.Server) *Service {
	service := &Service{
		udpSvc: udpSvc,
	}
	udpSvc.SetOnPacket(service.handlePacket)
	return service
}

func (s *Service) Start() error {
	return s.udpSvc.Start()
}

func (s *Service) Close() {
	s.udpSvc.Close()
}

func (s *Service) handlePacket(conn *udp.Connection, data []byte) {
	msg, err := stun.Decode(data)
	if err != nil {
		log.Printf("failed to decode STUN message: %v", err)
		return
	}

	switch msg.Type {
	case stun.MessageTypeBindingRequest:
		s.handleBindingRequest(conn, msg)
	default:
		log.Printf("unknown STUN message type: 0x%x", msg.Type)
		return
	}
}

func (s *Service) handleBindingRequest(conn *udp.Connection, msg *stun.Message) {
	clientAddr := conn.GetRemoteAddr()
	clientIP := clientAddr.IP
	clientPort := clientAddr.Port

	// 创建响应消息
	resp := stun.NewMessage(stun.MessageTypeBindingResponse, msg.TransactionID)

	// 设置XOR-MAPPED-ADDRESS
	resp.SetXORMappedAddress(clientIP, clientPort)

	// 编码响应消息
	data := stun.Encode(resp)

	// 发送响应
	if err := conn.Write(data); err != nil {
		log.Printf("failed to send STUN response: %v", err)
		return
	} else {
		log.Printf("send STUN response to %s:%d", clientIP, clientPort)
	}
}
