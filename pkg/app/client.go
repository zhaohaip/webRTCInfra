package app

import (
	"encoding/json"
	"fmt"
	"log"
	"signalingServer/pkg/common"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserName string
	UserID   string
	SendMsg  chan []byte
}

func NewClient(userName string, userID string, conn *websocket.Conn) *Client {
	return &Client{
		Conn:     conn,
		UserName: userName,
		UserID:   userID,
		SendMsg:  make(chan []byte, 32),
	}
}

func (c *Client) generateSignalingResponse(req common.SignalingRequest, err error) *common.SignalingResponse {
	return &common.SignalingResponse{
		Type:     req.Type,
		ErrorMsg: err.Error(),
		SDPMessage: common.SDPMessage{
			From: req.From,
			SDP:  req.SDP,
			To:   req.To,
		},
	}
}

// 校验信令
func (c *Client) validateSignalingRequest(req *common.SignalingRequest) error {
	switch req.Type {
	case common.SignallingTypeOffer, common.SignallingTypeAnswer:
		if len(req.SDP) == 0 || len(req.From) == 0 || len(req.To) == 0 {
			return fmt.Errorf("sdp、from、to are cannot be empty")
		}
	}
	return nil
}

// 解析信令
func (c *Client) analyzeSignalingRequest(data []byte) (*common.SignalingRequest, error) {
	var req common.SignalingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("[WsReadLoop] unmarshal signaling request failed: %v", err)
	}

	if err := c.validateSignalingRequest(&req); err != nil {
		return nil, err
	}

	return &req, nil
}
func (c *Client) WsReadLoop(cm *ClientManager) {
	defer func() {
		cm.RemoveClient(c.UserID)
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("[WsReadLoop] client %s disconnected: %v", c.UserID, err)
			break
		}

		// 解析请求数据
		req, err := c.analyzeSignalingRequest(data)
		if err != nil {
			res, _ := common.NewWebsocketServiceResponse(c.UserID, common.SignallingTypeError, err)
			c.SendMsg <- res
			continue
		}

		switch req.Type {
		case common.SignallingTypeOffer, common.SignallingTypeAnswer:
			if client, ok := cm.GetClient(req.To); ok {
				select {
				case client.SendMsg <- data:
				default:
					// 缓存满了，删除该用户，并断开连接
					close(client.SendMsg)
					cm.RemoveClient(req.To)
					log.Printf("[WsReadLoop] client %s send message to %s failed: channel is full", c.UserID, req.To)
				}
			} else {
				// 所发送的客户端不存在
				res, _ := common.NewWebsocketServiceResponse(c.UserID, common.SignallingTypeError, fmt.Errorf("[WsReadLoop] send to client %s not found", req.To))
				c.SendMsg <- res
				continue
			}
		case common.SignallingTypeClose: // 客户端主动关闭请求
			break
		default:
			res, _ := common.NewWebsocketServiceResponse(c.UserID, common.SignallingTypeError, fmt.Errorf("[WsReadLoop] unsupported signaling type: %s", req.Type))
			c.SendMsg <- res
			continue
		}
	}

}

func (c *Client) WsWriteLoop() {
	for {
		for msg := range c.SendMsg {
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}
