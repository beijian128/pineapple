
package net

import (
	"fmt"
	"net"
	"sync"

	"github.com/beijian128/pineapple/internal/utils"
	"go.uber.org/zap"
)

type ConnHandler interface {
	OnConnect(conn net.Conn)
	OnMessage(conn net.Conn, data []byte)
	OnDisconnect(conn net.Conn)
}

type TCPServer struct {
	addr        string
	port        int
	handler     ConnHandler
	listener    net.Listener
	connections map[net.Conn]struct{}
	mu          sync.RWMutex
	running     bool
}

func NewTCPServer(port int, handler ConnHandler) *TCPServer {
	return &TCPServer{
		port:        port,
		addr:        fmt.Sprintf(":%d", port),
		handler:     handler,
		connections: make(map[net.Conn]struct{}),
	}
}

func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %w", err)
	}

	s.running = true
	utils.Logger.Info("tcp server started", zap.Int("port", s.port))

	go s.acceptLoop()
	return nil
}

func (s *TCPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	if s.listener != nil {
		_ = s.listener.Close()
	}

	for conn := range s.connections {
		_ = conn.Close()
		delete(s.connections, conn)
	}

	utils.Logger.Info("tcp server stopped")
}

func (s *TCPServer) acceptLoop() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				utils.Logger.Error("accept connection error", zap.Error(err))
			}
			continue
		}

		s.mu.Lock()
		s.connections[conn] = struct{}{}
		s.mu.Unlock()

		if s.handler != nil {
			s.handler.OnConnect(conn)
		}

		go s.readLoop(conn)
	}
}

func (s *TCPServer) readLoop(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()

		if s.handler != nil {
			s.handler.OnDisconnect(conn)
		}
		_ = conn.Close()
	}()

	buf := make([]byte, 4096)
	for s.running {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		if s.handler != nil {
			s.handler.OnMessage(conn, buf[:n])
		}
	}
}
