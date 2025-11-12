package stun

import (
	"encoding/binary"
	"net"
)

// 消息类型
const (
	MessageTypeBindingRequest  uint16 = 0x0001
	MessageTypeBindingResponse uint16 = 0x0101
)

// 属性类型
const (
	AttributeTypeXORMappedAddress uint16 = 0x0020
	AttributeTypeErrorCode               = 0x0009
)

const (
	IPV4 = 0x01
	IPV6 = 0x02
)

type Message struct {
	Type          uint16
	TransactionID [12]byte
	Attributes    map[uint16][]byte
}

func NewMessage(type_ uint16, transactionID [12]byte) *Message {
	return &Message{
		Type:          type_,
		TransactionID: transactionID,
		Attributes:    make(map[uint16][]byte),
	}
}

func (m *Message) SetXORMappedAddress(ip net.IP, port int) {
	var family byte
	var ipBytes []byte
	if ip.To4() != nil {
		family = IPV4
		ipBytes = ip.To4()
	} else {
		family = IPV6
		ipBytes = ip.To16()
	}

	// 编码端口
	portBytes := make([]byte, 2)
	magic := binary.BigEndian.Uint16(magicCookie[:2])
	binary.BigEndian.PutUint16(portBytes, uint16(port^int(magic)))

	xorIP := make([]byte, len(ipBytes))
	for i := 0; i < len(ipBytes); i++ {
		xorIP[i] = ipBytes[i] ^ magicCookie[i%4]
	}

	value := []byte{0x00, family}
	value = append(value, portBytes...)
	value = append(value, xorIP...)
	m.Attributes[AttributeTypeXORMappedAddress] = value
}
