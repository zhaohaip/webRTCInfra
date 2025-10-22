package common

import (
	"encoding/json"
)

func NewWebsocketServiceResponse(clientID, msgType string, err error) ([]byte, error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	res := &SignalingResponse{
		Type:     msgType,
		ErrorMsg: errMsg,
		SDPMessage: SDPMessage{
			From: "service",
			To:   clientID,
		},
	}

	resByte, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return resByte, nil
}
