package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct {
	logger *logrus.Logger
}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware(logger *logrus.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// Handler 返回日志处理函数
func (m *LoggerMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// 构建日志字段
		fields := logrus.Fields{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency.Milliseconds(),
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"body_size":  bodySize,
		}

		// 添加查询参数（如果有）
		if len(c.Request.URL.RawQuery) > 0 {
			fields["query"] = c.Request.URL.RawQuery
		}

		// 添加用户代理（如果有）
		if userAgent != "" {
			fields["user_agent"] = userAgent
		}

		// 添加用户信息（如果已认证）
		if linuxDoID, exists := c.Get("linux_do_id"); exists {
			fields["linux_do_id"] = linuxDoID
		}
		if username, exists := c.Get("username"); exists {
			fields["username"] = username
		}

		// 添加错误信息（如果有）
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 根据状态码选择日志级别
		entry := m.logger.WithFields(fields)

		switch {
		case statusCode >= 500:
			entry.Error("Server error")
		case statusCode >= 400:
			entry.Warn("Client error")
		case statusCode >= 300:
			entry.Info("Redirection")
		default:
			entry.Info("Request completed")
		}
	}
}

// SkipPaths 返回跳过日志记录的路径
func (m *LoggerMiddleware) SkipPaths(skipPaths []string) gin.HandlerFunc {
	skipPathsMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipPathsMap[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果路径在跳过列表中，直接处理请求
		if skipPathsMap[path] {
			c.Next()
			return
		}

		// 否则使用正常的日志处理
		m.Handler()(c)
	}
}

// DetailedHandler 返回详细的日志处理函数（包含请求体等）
func (m *LoggerMiddleware) DetailedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		proto := c.Request.Proto

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		fields := logrus.Fields{
			"method":        method,
			"path":          path,
			"status":        statusCode,
			"latency":       latency.Milliseconds(),
			"latency_human": latency.String(),
			"client_ip":     clientIP,
			"body_size":     bodySize,
			"proto":         proto,
		}

		if len(c.Request.URL.RawQuery) > 0 {
			fields["query"] = c.Request.URL.RawQuery
		}

		if userAgent != "" {
			fields["user_agent"] = userAgent
		}

		if referer := c.Request.Referer(); referer != "" {
			fields["referer"] = referer
		}

		if linuxDoID, exists := c.Get("linux_do_id"); exists {
			fields["linux_do_id"] = linuxDoID
		}
		if username, exists := c.Get("username"); exists {
			fields["username"] = username
		}

		// 添加请求头（可选）
		if m.logger.Level >= logrus.DebugLevel {
			headers := make(map[string]string)
			for key, values := range c.Request.Header {
				if len(values) > 0 {
					// 隐藏敏感信息
					if key == "Authorization" || key == "Cookie" {
						headers[key] = "[REDACTED]"
					} else {
						headers[key] = values[0]
					}
				}
			}
			fields["headers"] = headers
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		entry := m.logger.WithFields(fields)

		switch {
		case statusCode >= 500:
			entry.Error("Server error")
		case statusCode >= 400:
			entry.Warn("Client error")
		case statusCode >= 300:
			entry.Info("Redirection")
		default:
			entry.Debug("Request completed")
		}
	}
}

// SimpleHandler 返回简单的日志处理函数（只记录基本信息）
func (m *LoggerMiddleware) SimpleHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		latency := time.Since(startTime)

		m.logger.WithFields(logrus.Fields{
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"status":  c.Writer.Status(),
			"latency": latency.Milliseconds(),
		}).Info("Request")
	}
}
