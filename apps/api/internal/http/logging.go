package http

import (
	"time"

	"github.com/bytedance/calandar/apps/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header(requestIDHeader, requestID)

		c.Next()

		status := c.Writer.Status()
		message := "api_request request_id=%s method=%s path=%s status=%d latency_ms=%d client_ip=%s"
		args := []any{
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			status,
			time.Since(start).Milliseconds(),
			c.ClientIP(),
		}

		switch {
		case status >= 500:
			logger.Errorf(message, args...)
		case status >= 400:
			logger.Warningf(message, args...)
		case c.Request.URL.Path == "/api/health":
			logger.Debugf(message, args...)
		default:
			logger.Infof(message, args...)
		}
	}
}
