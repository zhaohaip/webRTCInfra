package stun

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var magicCookie = []byte{0x21, 0x12, 0xa4, 0x42}

func Encode(msg *Message) []byte {
	// 计算属性部分的长度
	var attrBytes []byte
	for attrType, value := range msg.Attributes {
		// 属性长度
		attrLen := len(value)
		padLen := attrLen
		if padLen%4 != 0 {
			padLen += 4 - padLen%4
		}

		attrHeader := make([]byte, 4)
		binary.BigEndian.PutUint16(attrHeader[:2], attrType)
		binary.BigEndian.PutUint16(attrHeader[2:], uint16(attrLen))
		attrBytes = append(attrBytes, attrHeader...)
		attrBytes = append(attrBytes, value...)

		// 四字节对齐
		if padLen > attrLen {
			attrBytes = append(attrBytes, bytes.Repeat([]byte{0}, padLen-attrLen)...)
		}
	}

	// 消息头部
	header := make([]byte, 20)
	binary.BigEndian.PutUint16(header[:2], msg.Type)
	binary.BigEndian.PutUint16(header[2:4], uint16(len(attrBytes)))
	copy(header[4:8], magicCookie)
	copy(header[8:20], msg.TransactionID[:])

	return append(header, attrBytes...)
}

func Decode(date []byte) (*Message, error) {
	if len(date) < 20 {
		return nil, fmt.Errorf("stun: packet too short")
	}

	msgType := binary.BigEndian.Uint16(date[0:2])
	msgLen := binary.BigEndian.Uint16(date[2:4])
	cookie := date[4:8]

	// 获取事务ID，通过拷贝的方式，避免修改原始数据
	var transactionID [12]byte
	copy(transactionID[:], date[8:20])

	// 校验magicCookie
	if !bytes.Equal(cookie, magicCookie) {
		return nil, fmt.Errorf("stun: magic cookie mismatch")
	}

	// 校验消息长度
	if int(msgLen)+20 != len(date) {
		return nil, fmt.Errorf("stun: message length mismatch")
	}

	msg := &Message{
		Type:          msgType,
		TransactionID: transactionID,
		Attributes:    make(map[uint16][]byte),
	}

	var attributeDate = make([]byte, msgLen)
	copy(attributeDate, date[20:])

	offset := 0
	// 属性值可能存在一个或多个
	for offset < len(attributeDate) {
		if offset+4 > len(attributeDate) {
			return nil, fmt.Errorf("stun: attribute too short")
		}

		attrType := binary.BigEndian.Uint16(attributeDate[offset : offset+2])
		attrLen := binary.BigEndian.Uint16(attributeDate[offset+2 : offset+4])
		offset += 4

		if offset+int(attrLen) > len(attributeDate) {
			return nil, fmt.Errorf("stun: attribute length mismatch")
		}

		var value = make([]byte, attrLen)
		copy(value, attributeDate[offset:offset+int(attrLen)])
		msg.Attributes[attrType] = value

		// 属性长度按4个字节对其
		offset += int(attrLen)
		if pad := offset % 4; pad != 0 {
			offset += 4 - pad
		}
	}
	return msg, nil
}
