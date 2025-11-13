package udp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock UDP连接用于测试
type mockUDPConn struct {
	*net.UDPConn
}

func (m *mockUDPConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return len(b), nil
}

func TestServer_handlePackets(t *testing.T) {
	// 测试数据包处理
	t.Run("处理客户端数据包", func(t *testing.T) {
		server := NewService(":0", func(conn *Connection, data []byte) {
			// 验证回调被正确调用
			assert.Equal(t, []byte("test data"), data)
		})

		// 创建模拟UDP连接
		mockConn := &mockUDPConn{}
		clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
		conn := NewUDPConnection(mockConn.UDPConn, clientAddr)

		// 启动处理协程
		done := make(chan bool)
		go func() {
			server.handlePackets(conn)
			done <- true
		}()

		// 发送测试数据包
		testData := []byte("test data")
		conn.SavePacket(testData)

		// 验证数据包被处理
		time.Sleep(100 * time.Millisecond)

		// 清理
		server.Close()
	})

	// 测试资源清理
	t.Run("客户端断开后资源清理", func(t *testing.T) {
		server := NewService(":0", func(conn *Connection, data []byte) {})

		mockConn := &mockUDPConn{}
		clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12347")
		conn := NewUDPConnection(mockConn.UDPConn, clientAddr)

		// 添加客户端到服务器
		clientKey := clientAddr.String()
		server.mu.Lock()
		server.clients[clientKey] = conn
		server.mu.Unlock()

		// 启动处理协程
		go func() {
			server.handlePackets(conn)
		}()

		// 等待处理完成
		time.Sleep(UDPTimeOut + 5*time.Second)

		// 验证客户端已被移除
		server.mu.RLock()
		_, exists := server.clients[clientKey]
		server.mu.RUnlock()
		assert.False(t, exists)
	})
}
