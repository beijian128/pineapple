package router

import (
	"bytes"
	"net"
	"sync"

	pineapplenet "github.com/beijian128/pineapple/internal/net"
	"github.com/beijian128/pineapple/internal/utils"
	"go.uber.org/zap"
)

type Router struct {
	handlers    map[uint32]Handler
	middlewares []HandlerFunc
	mu          sync.RWMutex
	codec       *pineapplenet.Codec
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint32]Handler),
		codec:    pineapplenet.NewCodec(),
	}
}

func (r *Router) Use(middleware ...HandlerFunc) {
	r.middlewares = append(r.middlewares, middleware...)
}

func (r *Router) Register(msgID uint32, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgID] = handler
	utils.Logger.Debug("handler registered", zap.Uint32("msg_id", msgID))
}

func (r *Router) RegisterFunc(msgID uint32, fn HandlerFunc) {
	r.Register(msgID, fn.Wrapper())
}

func (r *Router) Handle(conn net.Conn, data []byte) {
	buf := bytes.NewBuffer(data)
	for buf.Len() > 0 {
		msg, err := r.codec.Decode(buf)
		if err != nil {
			utils.Logger.Error("decode message error", zap.Error(err))
			break
		}
		r.dispatch(conn, msg)
	}
}

func (r *Router) dispatch(conn net.Conn, msg *pineapplenet.Message) {
	r.mu.RLock()
	handler, exists := r.handlers[msg.MsgID]
	r.mu.RUnlock()

	ctx := NewContext(conn, msg.MsgID, msg.Data)
	ctx.Set("router", r)

	if exists {
		ctx.handlers = make([]HandlerFunc, 0, len(r.middlewares)+1)
		ctx.handlers = append(ctx.handlers, r.middlewares...)
		ctx.handlers = append(ctx.handlers, func(c *Context) {
			handler.Handle(c)
		})
	} else {
		utils.Logger.Warn("no handler found for message", zap.Uint32("msg_id", msg.MsgID))
		ctx.handlers = r.middlewares
	}

	ctx.Next()
}

func (r *Router) Send(conn net.Conn, msg *pineapplenet.Message) error {
	data, err := r.codec.Encode(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}
