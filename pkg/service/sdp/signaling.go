package sdp

import (
	"encoding/json"
	"fmt"
	"webRTCInfra/pkg/common"
	"webRTCInfra/pkg/network/websocket"
)

// Signaler 信令处理器
type Signaler struct {
	connMgr *websocket.Manager
}

func NewSignaler(mgr *websocket.Manager) *Signaler {
	return &Signaler{
		connMgr: mgr,
	}
}

// HandleMessage 处理收到的信令消息
func (s *Signaler) HandleMessage(userID string, data []byte) {
	// 解析信令请求
	req, err := s.parseRequest(data)
	if err != nil {
		s.sendError(userID, err.Error())
		return
	}

	// 校验信令
	if err = s.validateRequest(req); err != nil {
		s.sendError(userID, err.Error())
		return
	}

	// 处理不同类型的信令
	switch req.Type {
	case common.SignallingTypeOffer, common.SignallingTypeAnswer:
		s.forwardSignaling(userID, req.To, data) // 转发信令给目标客户端
	case common.SignallingTypeClose:
		s.handleClose(userID) // 处理关闭请求
	default:
		s.sendError(userID, fmt.Sprintf("unsupported signaling type: %s", req.Type))
	}
}

// 解析信令请求
func (s *Signaler) parseRequest(data []byte) (*common.SignalingRequest, error) {
	var req common.SignalingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("invalid signaling format: %v", err)
	}
	return &req, nil
}

// 校验信令合法性
func (s *Signaler) validateRequest(req *common.SignalingRequest) error {
	if req.From == "" {
		return fmt.Errorf("missing 'from' field")
	}
	if req.To == "" && req.Type != common.SignallingTypeClose {
		return fmt.Errorf("missing 'to' field")
	}
	if (req.Type == common.SignallingTypeOffer || req.Type == common.SignallingTypeAnswer) && req.SDP == "" {
		return fmt.Errorf("missing 'sdp' field")
	}
	return nil
}

func (s *Signaler) forwardSignaling(sourceUserID, targetUserID string, data []byte) {
	targetConn, ok := s.connMgr.GetClient(targetUserID)
	if !ok {
		s.sendError(sourceUserID, fmt.Sprintf("target user %s not found", targetUserID))
		return
	}

	if !targetConn.Send(data) {
		// 客户端消息处理不过来，关闭客户端
		targetConn.Conn.Close()
		s.handleClose(targetUserID)
	}
}

func (s *Signaler) handleClose(userID string) {
	s.connMgr.RemoveClient(userID)
}

// 发送错误响应
func (s *Signaler) sendError(userID, msg string) {
	errMsg, _ := common.NewWebsocketServiceResponse("", common.SignallingTypeError, fmt.Errorf(msg))
	if conn, ok := s.connMgr.GetClient(userID); ok {
		if !conn.Send(errMsg) {
			conn.Conn.Close()
			s.handleClose(userID)
		}
	}
}
