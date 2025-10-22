package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"signalingServer/pkg/common"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var serverURL = "127.0.0.1:8080"

func listClientsHttpRequest(t *testing.T) *common.ListClientsResponse {
	resp, err := http.Get("http://" + serverURL + "/clients")
	assert.NoError(t, err)
	defer resp.Body.Close()

	var res common.ListClientsResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.NoError(t, err)

	return &res
}

func createConnectWS(userName string) (*websocket.Conn, error) {
	u := url.URL{
		Scheme:   "ws",
		Host:     serverURL,
		Path:     "ws/signaling",
		RawQuery: "name=" + userName,
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return conn, err
}

// 建立 WebSocket 连接的帮助函数
func connectWS(t *testing.T, userName string) (*websocket.Conn, string) {

	conn, err := createConnectWS(userName)
	assert.NoError(t, err, "WebSocket 连接失败")

	// wait for initial message from server (with timeout)
	deadline := time.After(6 * time.Second)
	msgCh := make(chan common.SignalingResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}

		res := common.SignalingResponse{}
		err = json.Unmarshal(msg, &res)
		if err != nil {
			errCh <- err
			return
		}

		if res.Type != common.SignallingTypeRegister {
			errCh <- fmt.Errorf("无效的初始消息类型")
			return
		}

		msgCh <- res
	}()

	select {
	case <-deadline:
		t.Fatal("超时等待初始消息")
	case msg := <-msgCh:
		return conn, msg.To
	case err := <-errCh:
		t.Fatalf("读取初始消息失败: %v", err)
	}
	return conn, ""
}

type ClientMsg struct {
	Conn *websocket.Conn
	Id   string
}

func ClientSendMessage(t *testing.T, ty string, fromClient *ClientMsg, toClient *ClientMsg) {
	msg := fmt.Sprintf(`{"type":"%s","from":"%s","to":"%s","sdp":"v=0..."}`, ty, fromClient.Id, toClient.Id)
	err := fromClient.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	assert.NoError(t, err)

	// 检查接收端是否已收到消息
	done := make(chan struct{})
	var received string
	go func() {
		_, msg, _ := toClient.Conn.ReadMessage()
		received = string(msg)
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, msg, received)
	case <-time.After(5 * time.Second):
		t.Fatal("超时等待消息")
	}
}

// --- helper 函数：读取JSON消息 ---
func readJSON(t *testing.T, conn *websocket.Conn) map[string]any {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err, "read message failed")

	var m map[string]any
	err = json.Unmarshal(msg, &m)
	assert.NoError(t, err)
	return m
}

// 测试发送异常消息
func ClientSendAbnormalMessage(t *testing.T, msg string, fromClient *ClientMsg, containsErr string) {
	err := fromClient.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	assert.NoError(t, err)

	// 检查是否接收到报错消息
	resp := readJSON(t, fromClient.Conn)
	assert.Equal(t, "error", resp["type"])
	assert.Contains(t, resp["errorMsg"], containsErr)
}

func TestSignalingServerE2E(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("E2E测试-客户端正常连接和注册", func(t *testing.T) {
		clientA, idA := connectWS(t, "Alice")
		defer clientA.Close()

		assert.NotEmpty(t, idA)
	})
	t.Run("E2E测试-客户端正常发送offer信令", func(t *testing.T) {
		clientA, idA := connectWS(t, "Alice")
		defer clientA.Close()

		clientB, idB := connectWS(t, "Bob")
		defer clientB.Close()

		// 模拟：Alice 向 Bob 发送 SDP Offer
		ClientSendMessage(t, common.SignallingTypeOffer, &ClientMsg{clientA, idA}, &ClientMsg{clientB, idB})
	})
	t.Run("异常：无用户名连接", func(t *testing.T) {
		_, err := createConnectWS("")
		assert.Error(t, err, "should fail without name param")
	})

	t.Run("异常：发送非法信令", func(t *testing.T) {
		clientA, idA := connectWS(t, "Alice")
		defer clientA.Close()
		clientB, idB := connectWS(t, "Bob")
		defer clientB.Close()

		assert.NotEmpty(t, idA)
		assert.NotEmpty(t, idB)

		// 向服务器发送非法json数据
		ClientSendAbnormalMessage(t, "this is not json", &ClientMsg{clientA, idA}, "invalid character")

		// 向服务器发送不支持的信令类型
		ClientSendAbnormalMessage(t, `{"type":"invalid"}`, &ClientMsg{clientA, idA}, "unsupported signaling type")

		// 向服务器发送空的信令内容
		ClientSendAbnormalMessage(t, `{"type":"offer"}`, &ClientMsg{clientA, idA}, "cannot be empty")

		// 向服务器发送信令给不存在的客户端
		msg := fmt.Sprintf(`{"type":"offer","from":"%s","to":"none","sdp":"v=0..."}`, idA)
		ClientSendAbnormalMessage(t, msg, &ClientMsg{clientA, idA}, "client none not found")

		// 由于服务端解析失败不会中断连接，应仍能继续收发
		time.Sleep(200 * time.Millisecond)
		ClientSendMessage(t, common.SignallingTypeOffer, &ClientMsg{clientA, idA}, &ClientMsg{clientB, idB})

	})

	t.Run("异常：客户端异常断开", func(t *testing.T) {
		clientA, idA := connectWS(t, "Alice")
		assert.NotEmpty(t, idA)

		clientA.Close()
		time.Sleep(500 * time.Millisecond)

		clients := listClientsHttpRequest(t)
		assert.Len(t, clients.Clients, 0)
	})
}
