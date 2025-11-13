package stun

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// 构造辅助函数
func buildValidStunPacket(magicCookie []byte, transactionID []byte, attrType uint16, attrValue []byte) []byte {
	var buf bytes.Buffer

	// 写入 STUN Header
	binary.Write(&buf, binary.BigEndian, uint16(0x0001))           // Type: Binding Request
	binary.Write(&buf, binary.BigEndian, uint16(4+len(attrValue))) // Length = 4 bytes header + value length
	buf.Write(magicCookie)
	buf.Write(transactionID)

	// 写入属性
	binary.Write(&buf, binary.BigEndian, attrType)
	binary.Write(&buf, binary.BigEndian, uint16(len(attrValue)))
	buf.Write(attrValue)

	return buf.Bytes()
}

func TestDecode(t *testing.T) {
	transactionID := []byte{
		0x63, 0x61, 0x66, 0x65, 0x62, 0x61,
		0x62, 0x65, 0x66, 0x61, 0x63, 0x65,
	}

	type args struct {
		date []byte
	}
	tests := []struct {
		name    string
		args    []byte
		want    *Message
		wantErr bool
		check   func(t *testing.T, msg *Message)
	}{
		// TODO: Add test cases.
		{
			name: "valid STUN message",
			args: buildValidStunPacket(magicCookie, transactionID, 0x0001, []byte{0xDE, 0xAD, 0xBE, 0xEF}),
			check: func(t *testing.T, msg *Message) {
				if msg.Type != 0x0001 {
					t.Errorf("expected type 0x0001, got %x", msg.Type)
				}
				attr, ok := msg.Attributes[0x0001]
				if !ok {
					t.Fatalf("expected attribute 0x0001 not found")
				}
				if !bytes.Equal(attr, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
					t.Errorf("expected attr value DEADBEEF, got %X", attr)
				}
			},
		},
		{
			name:    "packet too short",
			args:    make([]byte, 10),
			wantErr: true,
		},
		{
			name: "invalid magic cookie",
			args: func() []byte {
				b := buildValidStunPacket([]byte{0x00, 0x00, 0x00, 0x00}, transactionID, 0x0001, []byte{0x01, 0x02})
				return b
			}(),
			wantErr: true,
		},
		{
			name: "attribute length mismatch",
			args: func() []byte {
				var buf bytes.Buffer
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(8))      // Wrong length
				buf.Write(magicCookie)
				buf.Write(transactionID)
				binary.Write(&buf, binary.BigEndian, uint16(0x0001))
				binary.Write(&buf, binary.BigEndian, uint16(10)) // Invalid
				buf.Write([]byte{0x01, 0x02, 0x03})
				return buf.Bytes()
			}(),
			wantErr: true,
		},
		{
			name: "message length mismatch",
			args: func() []byte {
				var buf bytes.Buffer
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(100))    // Length too large
				buf.Write(magicCookie)
				buf.Write(transactionID)
				// No actual attribute data to match the claimed length
				return buf.Bytes()
			}(),
			wantErr: true,
		}, {
			name:    "empty message",
			args:    []byte{},
			wantErr: true,
		},
		{
			name: "message with no attributes",
			args: func() []byte {
				var buf bytes.Buffer
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(0))      // No attributes
				buf.Write(magicCookie)
				buf.Write(transactionID)
				return buf.Bytes()
			}(),
			check: func(t *testing.T, msg *Message) {
				if msg.Type != 0x0001 {
					t.Errorf("expected type 0x0001, got %x", msg.Type)
				}
				if len(msg.Attributes) != 0 {
					t.Errorf("expected no attributes, got %d", len(msg.Attributes))
				}
			},
		},
		{
			name: "message with multiple attributes",
			args: func() []byte {
				var buf bytes.Buffer
				// Header
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(16))     // Total attribute length

				buf.Write(magicCookie)
				buf.Write(transactionID)

				// First attribute
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // MAPPED-ADDRESS
				binary.Write(&buf, binary.BigEndian, uint16(4))      // Length
				buf.Write([]byte{0x01, 0x02, 0x03, 0x04})

				// Second attribute
				binary.Write(&buf, binary.BigEndian, uint16(0x0002)) // USERNAME
				binary.Write(&buf, binary.BigEndian, uint16(4))      // Length
				buf.Write([]byte{0x05, 0x06, 0x07, 0x08})

				return buf.Bytes()
			}(),
			check: func(t *testing.T, msg *Message) {
				if msg.Type != 0x0001 {
					t.Errorf("expected type 0x0001, got %x", msg.Type)
				}

				// Check first attribute
				attr1, ok := msg.Attributes[0x0001]
				if !ok {
					t.Errorf("expected attribute 0x0001 not found")
				}
				if !bytes.Equal(attr1, []byte{0x01, 0x02, 0x03, 0x04}) {
					t.Errorf("expected attr1 value 01020304, got %X", attr1)
				}

				// Check second attribute
				attr2, ok := msg.Attributes[0x0002]
				if !ok {
					t.Errorf("expected attribute 0x0002 not found")
				}
				if !bytes.Equal(attr2, []byte{0x05, 0x06, 0x07, 0x08}) {
					t.Errorf("expected attr2 value 05060708, got %X", attr2)
				}
			},
		},
		{
			name: "attribute with padding",
			args: func() []byte {
				var buf bytes.Buffer
				// Header
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(8))      // Attribute length

				buf.Write(magicCookie)
				buf.Write(transactionID)

				// Attribute with length not multiple of 4 (3 bytes)
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Attribute type
				binary.Write(&buf, binary.BigEndian, uint16(3))      // Length = 3 bytes
				buf.Write([]byte{0x01, 0x02, 0x03})                  // Actual data
				// Padding should be added during encoding, but decoder should handle it correctly

				return buf.Bytes()
			}(),
			check: func(t *testing.T, msg *Message) {
				attr, ok := msg.Attributes[0x0001]
				if !ok {
					t.Errorf("expected attribute 0x0001 not found")
				}
				if !bytes.Equal(attr, []byte{0x01, 0x02, 0x03}) {
					t.Errorf("expected attr value 010203, got %X", attr)
				}
			},
		},
		{
			name: "attribute too short",
			args: func() []byte {
				var buf bytes.Buffer
				binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // Binding Request
				binary.Write(&buf, binary.BigEndian, uint16(2))      // Only 2 bytes, not enough for attribute header
				buf.Write(magicCookie)
				buf.Write(transactionID)
				buf.Write([]byte{0x01}) // Incomplete attribute data
				return buf.Bytes()
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}

		})
	}
}

func TestEncode(t *testing.T) {
	transactionID := [12]byte{
		0x63, 0x61, 0x66, 0x65, 0x62, 0x61,
		0x62, 0x65, 0x66, 0x61, 0x63, 0x65,
	}

	tests := []struct {
		name  string
		msg   Message
		check func(t *testing.T, encoded []byte)
	}{
		{
			name: "message with no attributes",
			msg: Message{
				Type:          0x0001,
				TransactionID: transactionID,
				Attributes:    make(map[uint16][]byte),
			},
			check: func(t *testing.T, encoded []byte) {
				if len(encoded) < 20 {
					t.Fatalf("encoded message too short: %d", len(encoded))
				}

				// 检查类型
				msgType := binary.BigEndian.Uint16(encoded[0:2])
				if msgType != 0x0001 {
					t.Errorf("expected type 0x0001, got %x", msgType)
				}

				// 检查长度
				msgLen := binary.BigEndian.Uint16(encoded[2:4])
				if msgLen != 0 {
					t.Errorf("expected length 0, got %d", msgLen)
				}

				// 检查magic cookie
				cookie := encoded[4:8]
				if !bytes.Equal(cookie, magicCookie) {
					t.Errorf("magic cookie mismatch")
				}

				// 检查事务ID
				transID := encoded[8:20]
				if !bytes.Equal(transID, transactionID[:]) {
					t.Errorf("transaction ID mismatch")
				}

				// 总长度应为20字节(header)
				if len(encoded) != 20 {
					t.Errorf("expected total length 20, got %d", len(encoded))
				}
			},
		},
		{
			name: "message with one attribute",
			msg: Message{
				Type:          0x0001,
				TransactionID: transactionID,
				Attributes: map[uint16][]byte{
					0x0001: {0xDE, 0xAD, 0xBE, 0xEF},
				},
			},
			check: func(t *testing.T, encoded []byte) {
				// 检查基本header信息
				msgType := binary.BigEndian.Uint16(encoded[0:2])
				if msgType != 0x0001 {
					t.Errorf("expected type 0x0001, got %x", msgType)
				}

				// 检查magic cookie
				cookie := encoded[4:8]
				if !bytes.Equal(cookie, magicCookie) {
					t.Errorf("magic cookie mismatch")
				}

				// 检查事务ID
				transID := encoded[8:20]
				if !bytes.Equal(transID, transactionID[:]) {
					t.Errorf("transaction ID mismatch")
				}

				// 检查属性长度(4字节属性数据 + 4字节属性头 = 8字节)
				msgLen := binary.BigEndian.Uint16(encoded[2:4])
				if msgLen != 8 {
					t.Errorf("expected attribute length 8, got %d", msgLen)
				}

				// 总长度应为28字节(20字节header + 8字节属性)
				if len(encoded) != 28 {
					t.Errorf("expected total length 28, got %d", len(encoded))
				}

				// 检查属性
				if len(encoded) >= 24 {
					attrType := binary.BigEndian.Uint16(encoded[20:22])
					if attrType != 0x0001 {
						t.Errorf("expected attribute type 0x0001, got %x", attrType)
					}

					attrLen := binary.BigEndian.Uint16(encoded[22:24])
					if attrLen != 4 {
						t.Errorf("expected attribute length 4, got %d", attrLen)
					}

					attrValue := encoded[24:28]
					if !bytes.Equal(attrValue, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
						t.Errorf("expected attribute value DEADBEEF, got %X", attrValue)
					}
				}
			},
		},
		{
			name: "message with attribute requiring padding",
			msg: Message{
				Type:          0x0001,
				TransactionID: transactionID,
				Attributes: map[uint16][]byte{
					0x0001: {0x01, 0x02, 0x03}, // 3字节，需要填充到4字节边界
				},
			},
			check: func(t *testing.T, encoded []byte) {
				// 检查属性长度(3字节属性数据 + 4字节属性头 + 1字节填充 = 8字节)
				msgLen := binary.BigEndian.Uint16(encoded[2:4])
				if msgLen != 8 {
					t.Errorf("expected attribute length 8, got %d", msgLen)
				}

				// 总长度应为28字节(20字节header + 8字节属性)
				if len(encoded) != 28 {
					t.Errorf("expected total length 28, got %d", len(encoded))
				}

				// 检查属性值
				if len(encoded) >= 27 {
					attrValue := encoded[24:27]
					if !bytes.Equal(attrValue, []byte{0x01, 0x02, 0x03}) {
						t.Errorf("expected attribute value 010203, got %X", attrValue)
					}
				}
			},
		},
		{
			name: "message with multiple attributes",
			msg: Message{
				Type:          0x0001,
				TransactionID: transactionID,
				Attributes: map[uint16][]byte{
					0x0001: {0x01, 0x02, 0x03, 0x04},
					0x0002: {0x05, 0x06},
				},
			},
			check: func(t *testing.T, encoded []byte) {
				// 检查总属性长度
				msgLen := binary.BigEndian.Uint16(encoded[2:4])
				// 第一个属性: 4字节数据 + 4字节头 = 8字节
				// 第二个属性: 2字节数据 + 4字节头 + 2字节填充 = 8字节
				// 总计: 16字节
				if msgLen != 16 {
					t.Errorf("expected attribute length 16, got %d", msgLen)
				}

				// 总长度应为36字节(20字节header + 16字节属性)
				if len(encoded) != 36 {
					t.Errorf("expected total length 36, got %d", len(encoded))
				}
			},
		},
		{
			name: "roundtrip encode/decode",
			msg: Message{
				Type:          0x0101,
				TransactionID: transactionID,
				Attributes: map[uint16][]byte{
					0x0001: {0xAA, 0xBB, 0xCC, 0xDD},
					0x0002: {0xEE, 0xFF},
				},
			},
			check: func(t *testing.T, encoded []byte) {
				// 使用Decode函数解码编码后的数据，验证一致性
				decoded, err := Decode(encoded)
				if err != nil {
					t.Fatalf("Decode failed: %v", err)
				}

				if decoded.Type != 0x0101 {
					t.Errorf("expected type 0x0101, got %x", decoded.Type)
				}

				if !bytes.Equal(decoded.TransactionID[:], transactionID[:]) {
					t.Errorf("transaction ID mismatch")
				}

				attr1, ok := decoded.Attributes[0x0001]
				if !ok {
					t.Error("missing attribute 0x0001")
				} else if !bytes.Equal(attr1, []byte{0xAA, 0xBB, 0xCC, 0xDD}) {
					t.Errorf("attribute 0x0001 value mismatch: %X", attr1)
				}

				attr2, ok := decoded.Attributes[0x0002]
				if !ok {
					t.Error("missing attribute 0x0002")
				} else if !bytes.Equal(attr2, []byte{0xEE, 0xFF}) {
					t.Errorf("attribute 0x0002 value mismatch: %X", attr2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(&tt.msg)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
