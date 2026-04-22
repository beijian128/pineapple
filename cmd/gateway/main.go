package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/beijian128/pineapple/internal/discovery"
	pineapplenet "github.com/beijian128/pineapple/internal/net"
	"github.com/beijian128/pineapple/internal/router"
	"github.com/beijian128/pineapple/internal/storage"
	"github.com/beijian128/pineapple/internal/utils"
	"go.uber.org/zap"
)

type GatewayHandler struct {
	router *router.Router
}

func NewGatewayHandler(r *router.Router) *GatewayHandler {
	return &GatewayHandler{
		router: r,
	}
}

func (h *GatewayHandler) OnConnect(conn net.Conn) {
	utils.Logger.Info("client connected", zap.String("remote", conn.RemoteAddr().String()))
}

func (h *GatewayHandler) OnMessage(conn net.Conn, data []byte) {
	h.router.Handle(conn, data)
}

func (h *GatewayHandler) OnDisconnect(conn net.Conn) {
	utils.Logger.Info("client disconnected", zap.String("remote", conn.RemoteAddr().String()))
}

func main() {
	configPath := "./config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := utils.LoadConfig(configPath); err != nil {
		panic(err)
	}

	if err := utils.InitLogger(&utils.AppConfig.Log); err != nil {
		panic(err)
	}
	defer utils.SyncLogger()

	if err := discovery.InitDiscovery(&utils.AppConfig.Etcd); err != nil {
		utils.Logger.Warn("etcd init failed", zap.Error(err))
	}
	defer discovery.CloseDiscovery()

	if err := storage.InitMongoDB(&utils.AppConfig.MongoDB); err != nil {
		utils.Logger.Warn("mongodb init failed", zap.Error(err))
	}
	defer storage.CloseMongoDB()

	if err := storage.InitRedis(&utils.AppConfig.Redis); err != nil {
		utils.Logger.Warn("redis init failed", zap.Error(err))
	}
	defer storage.CloseRedis()

	serviceInfo := &discovery.ServiceInfo{
		Name:    "gateway",
		Addr:    "127.0.0.1",
		Port:    8888,
		Version: "1.0.0",
	}
	if discovery.GlobalDiscovery != nil {
		if err := discovery.GlobalDiscovery.Register(serviceInfo); err != nil {
			utils.Logger.Error("service register failed", zap.Error(err))
		}
		defer func() {
			_ = discovery.GlobalDiscovery.Unregister(serviceInfo)
		}()
	}

	r := router.NewRouter()

	r.Use(router.Recovery())
	r.Use(router.Logger())

	r.RegisterFunc(router.MsgIDHeartbeat, HeartbeatHandler)
	r.RegisterFunc(router.MsgIDLoginRequest, LoginHandler)

	handler := NewGatewayHandler(r)
	var servers []*pineapplenet.TCPServer

	tcpServer := pineapplenet.NewTCPServer(8888, handler)
	if err := tcpServer.Start(); err != nil {
		utils.Logger.Fatal("tcp server start failed", zap.Error(err))
	}
	servers = append(servers, tcpServer)

	utils.Logger.Info("gateway server started",
		zap.String("name", "gateway"),
		zap.String("version", "1.0.0"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Logger.Info("shutting down...")
	for _, s := range servers {
		s.Stop()
	}
}

func HeartbeatHandler(c *router.Context) {
	utils.Logger.Debug("heartbeat received", zap.String("remote", c.Conn.RemoteAddr().String()))

	resp := &pineapplenet.Message{
		MsgID: router.MsgIDHeartbeat,
		Data:  nil,
	}

	if r, ok := c.Get("router"); ok {
		if router, ok := r.(*router.Router); ok {
			_ = router.Send(c.Conn, resp)
		}
	}
}

func LoginHandler(c *router.Context) {
	utils.Logger.Info("login request received",
		zap.String("remote", c.Conn.RemoteAddr().String()),
		zap.Int("data_size", len(c.Data)))

	resp := &pineapplenet.Message{
		MsgID: router.MsgIDLoginResponse,
		Data:  []byte(`{"code":0,"message":"ok"}`),
	}

	if r, ok := c.Get("router"); ok {
		if router, ok := r.(*router.Router); ok {
			_ = router.Send(c.Conn, resp)
		}
	}
}
