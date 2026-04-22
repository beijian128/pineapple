package main

import (
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/beijian128/pineapple/internal/dao"
	"github.com/beijian128/pineapple/internal/discovery"
	"github.com/beijian128/pineapple/internal/model"
	pineapplenet "github.com/beijian128/pineapple/internal/net"
	"github.com/beijian128/pineapple/internal/router"
	"github.com/beijian128/pineapple/internal/session"
	"github.com/beijian128/pineapple/internal/storage"
	"github.com/beijian128/pineapple/internal/utils"
	"go.uber.org/zap"
)

const (
	MsgIDRegisterRequest  = 1003
	MsgIDRegisterResponse = 1004
	MsgIDLogoutRequest    = 1005
	MsgIDLogoutResponse   = 1006
)

type LoginHandler struct {
	router  *router.Router
	userDAO *dao.UserDAO
}

func NewLoginHandler(r *router.Router, dao *dao.UserDAO) *LoginHandler {
	return &LoginHandler{
		router:  r,
		userDAO: dao,
	}
}

func (h *LoginHandler) OnConnect(conn net.Conn) {
	utils.Logger.Info("client connected", zap.String("remote", conn.RemoteAddr().String()))
}

func (h *LoginHandler) OnMessage(conn net.Conn, data []byte) {
	h.router.Handle(conn, data)
}

func (h *LoginHandler) OnDisconnect(conn net.Conn) {
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
		Name:    "login",
		Addr:    "127.0.0.1",
		Port:    9000,
		Version: utils.AppConfig.Server.Version,
	}
	if discovery.GlobalDiscovery != nil {
		if err := discovery.GlobalDiscovery.Register(serviceInfo); err != nil {
			utils.Logger.Error("service register failed", zap.Error(err))
		}
		defer func() {
			_ = discovery.GlobalDiscovery.Unregister(serviceInfo)
		}()
	}

	userDAO := dao.NewUserDAO()
	r := router.NewRouter()

	r.Use(router.Recovery())
	r.Use(router.Logger())

	handler := NewLoginHandler(r, userDAO)

	r.RegisterFunc(router.MsgIDLoginRequest, handler.LoginHandler)
	r.RegisterFunc(MsgIDRegisterRequest, handler.RegisterHandler)
	r.RegisterFunc(router.MsgIDHeartbeat, HeartbeatHandler)

	var servers []*pineapplenet.TCPServer

	tcpServer := pineapplenet.NewTCPServer(9000, handler)
	if err := tcpServer.Start(); err != nil {
		utils.Logger.Fatal("tcp server start failed", zap.Error(err))
	}
	servers = append(servers, tcpServer)

	utils.Logger.Info("login server started",
		zap.String("name", utils.AppConfig.Server.Name),
		zap.String("version", utils.AppConfig.Server.Version))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Logger.Info("shutting down...")
	for _, s := range servers {
		s.Stop()
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

func (h *LoginHandler) LoginHandler(c *router.Context) {
	var req LoginRequest
	if err := json.Unmarshal(c.Data, &req); err != nil {
		h.sendResponse(c, router.MsgIDLoginResponse, 400, "invalid request", nil)
		return
	}

	utils.Logger.Info("login request", zap.String("username", req.Username))

	user, err := h.userDAO.FindByUsername(req.Username)
	if err != nil {
		utils.Logger.Warn("user not found", zap.String("username", req.Username), zap.Error(err))
		h.sendResponse(c, router.MsgIDLoginResponse, 401, "invalid username or password", nil)
		return
	}

	if user.Status != model.UserStatusNormal {
		utils.Logger.Warn("user banned", zap.String("username", req.Username))
		h.sendResponse(c, router.MsgIDLoginResponse, 403, "account disabled", nil)
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		utils.Logger.Warn("invalid password", zap.String("username", req.Username))
		h.sendResponse(c, router.MsgIDLoginResponse, 401, "invalid username or password", nil)
		return
	}

	if err := h.userDAO.UpdateLastLogin(user.ID); err != nil {
		utils.Logger.Warn("update last login failed", zap.Error(err))
	}

	token, err := session.CreateSession(user)
	if err != nil {
		utils.Logger.Error("create session failed", zap.Error(err))
		h.sendResponse(c, router.MsgIDLoginResponse, 500, "internal error", nil)
		return
	}

	utils.Logger.Info("login success", zap.String("username", req.Username))

	respData := LoginResponse{
		Code:    0,
		Message: "success",
		Token:   token,
		UserID:  user.ID.Hex(),
	}
	data, _ := json.Marshal(respData)
	h.sendResponse(c, router.MsgIDLoginResponse, 0, "success", data)
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname,omitempty"`
}

type RegisterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func (h *LoginHandler) RegisterHandler(c *router.Context) {
	var req RegisterRequest
	if err := json.Unmarshal(c.Data, &req); err != nil {
		h.sendResponse(c, MsgIDRegisterResponse, 400, "invalid request", nil)
		return
	}

	utils.Logger.Info("register request", zap.String("username", req.Username))

	if req.Username == "" || req.Password == "" {
		h.sendResponse(c, MsgIDRegisterResponse, 400, "username or password required", nil)
		return
	}

	exists, err := h.userDAO.ExistsUsername(req.Username)
	if err != nil {
		utils.Logger.Error("check username failed", zap.Error(err))
		h.sendResponse(c, MsgIDRegisterResponse, 500, "internal error", nil)
		return
	}
	if exists {
		h.sendResponse(c, MsgIDRegisterResponse, 409, "username already exists", nil)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Logger.Error("hash password failed", zap.Error(err))
		h.sendResponse(c, MsgIDRegisterResponse, 500, "internal error", nil)
		return
	}

	user := model.NewUser(req.Username, hashedPassword)
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if err := h.userDAO.Create(user); err != nil {
		utils.Logger.Error("create user failed", zap.Error(err))
		h.sendResponse(c, MsgIDRegisterResponse, 500, "internal error", nil)
		return
	}

	utils.Logger.Info("register success", zap.String("username", req.Username))

	respData := RegisterResponse{
		Code:    0,
		Message: "success",
		UserID:  user.ID.Hex(),
	}
	data, _ := json.Marshal(respData)
	h.sendResponse(c, MsgIDRegisterResponse, 0, "success", data)
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

func (h *LoginHandler) sendResponse(c *router.Context, msgID uint32, code int, message string, data []byte) {
	resp := LoginResponse{
		Code:    code,
		Message: message,
	}
	if data == nil {
		data, _ = json.Marshal(resp)
	}

	msg := &pineapplenet.Message{
		MsgID: msgID,
		Data:  data,
	}
	if r, ok := c.Get("router"); ok {
		if router, ok := r.(*router.Router); ok {
			_ = router.Send(c.Conn, msg)
		}
	}
}
