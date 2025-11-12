package common

import (
	"encoding/json"
)

func NewWebsocketServiceResponse(clientID, msgType string, err error) ([]byte, error) {
	res := &SignalingResponse{
		Type: msgType,
	}

	if err != nil {
		res.ErrorMsg = err.Error()
	} else {
		res.SDPMessage.From = "service"
		res.SDPMessage.To = clientID
	}

	return json.Marshal(res)
}
