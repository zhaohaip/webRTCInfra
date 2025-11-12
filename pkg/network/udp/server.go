package udp

import (
	"log"
	"net"
	"sync"
)

// Server UDP服务器，负责监听端口并分发数据包
type Server struct {
	addr     string
	conn     *net.UDPConn
	clients  map[string]*Connection // 客户端链接映射
	mu       sync.RWMutex
	onPacket func(*Connection, []byte)
	close    bool
}

func NewService(addr string, onPacket func(*Connection, []byte)) *Server {
	return &Server{
		addr:     addr,
		clients:  make(map[string]*Connection),
		onPacket: onPacket,
	}
}

func (s *Server) SetOnPacket(fn func(*Connection, []byte)) {
	s.onPacket = fn
}

func (s *Server) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.conn = conn
	s.close = false

	go s.ListenLoop()

	return nil
}

func (s *Server) ListenLoop() {
	buf := make([]byte, 1024)
	for !s.close {
		n, clientAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if !s.close {
				log.Printf("read from udp error: %v", err)
			}
			return
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])

		conn := s.getOrCreateClient(clientAddr)

		// 分发数据到对应客户端通道
		go conn.SavePacket(packet)

	}
}

func (s *Server) getOrCreateClient(clientAddr *net.UDPAddr) *Connection {
	clientKey := clientAddr.String()

	s.mu.RLock()
	conn, ok := s.clients[clientKey]
	s.mu.RUnlock()
	if !ok {
		conn = NewUDPConnection(s.conn, clientAddr)
		s.mu.Lock()
		s.clients[clientKey] = conn
		s.mu.Unlock()

		go s.handleClient(conn)
	}
	log.Printf("client %s connected", clientAddr.String())

	return conn
}

func (s *Server) handleClient(conn *Connection) {
	clientAddr := conn.GetRemoteAddr().String()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientAddr)
		s.mu.Unlock()
		conn.Close()
		log.Printf("client %s disconnected", clientAddr)
	}()

	// 从通道中读取分发的数据包，调用业务层回调处理
	for pocket := range conn.Receive() {
		s.onPacket(conn, pocket)
	}
}

func (s *Server) Close() {
	s.close = true
	if s.conn != nil {
		s.conn.Close()
	}
	log.Println("udp service closed")
}
