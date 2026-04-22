
package router

import (
	"time"

	"github.com/beijian128/pineapple/internal/utils"
	"go.uber.org/zap"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		utils.Logger.Debug("request handled",
			zap.Uint32("msg_id", c.MsgID),
			zap.Duration("duration", duration),
			zap.String("remote", c.Conn.RemoteAddr().String()))
	}
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				utils.Logger.Error("handler panic",
					zap.Any("error", err),
					zap.Uint32("msg_id", c.MsgID))
			}
		}()
		c.Next()
	}
}
