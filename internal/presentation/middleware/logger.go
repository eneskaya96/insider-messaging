package middleware

import (
	"time"

	"github.com/eneskaya/insider-messaging/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Get().Error("request error", append(fields, zap.Error(e.Err))...)
			}
		} else {
			if statusCode >= 500 {
				logger.Get().Error("server error", fields...)
			} else if statusCode >= 400 {
				logger.Get().Warn("client error", fields...)
			} else {
				logger.Get().Info("request completed", fields...)
			}
		}
	}
}
