package router

import (
	"net"
	"sync"
)

type Context struct {
	Conn     net.Conn
	MsgID    uint32
	Data     []byte
	handlers []HandlerFunc
	index    int
	Keys     map[string]interface{}
	mu       sync.RWMutex
}

func NewContext(conn net.Conn, msgID uint32, data []byte) *Context {
	return &Context{
		Conn:  conn,
		MsgID: msgID,
		Data:  data,
		Keys:  make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Keys[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.Keys[key]
	return value, ok
}

func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}
