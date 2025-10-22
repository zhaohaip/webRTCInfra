package common

const (
	SignallingTypeRegister = "register"
	SignallingTypeOffer    = "offer"
	SignallingTypeAnswer   = "answer"
	SignallingTypeClose    = "close"
	SignallingTypeError    = "error"
)

type SDPMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	SDP  string `json:"sdp"`
}

type SignalingRequest struct {
	Type string `json:"type" validate:"required"`
	SDPMessage
}

type SignalingResponse struct {
	Type     string `json:"type"`
	ErrorMsg string `json:"errorMsg"`
	SDPMessage
}

type ClientMessage struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ListClientsResponse struct {
	Clients []ClientMessage `json:"clients"`
}
