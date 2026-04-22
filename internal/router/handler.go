package router

type HandlerFunc func(*Context)

type Handler interface {
	Handle(*Context)
}

type HandlerFuncWrapper struct {
	fn HandlerFunc
}

func (h *HandlerFuncWrapper) Handle(c *Context) {
	h.fn(c)
}

func (fn HandlerFunc) Wrapper() Handler {
	return &HandlerFuncWrapper{fn: fn}
}
