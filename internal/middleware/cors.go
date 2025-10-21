package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
	logger           *logrus.Logger
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware(config CORSConfig, logger *logrus.Logger) *CORSMiddleware {
	// 设置默认值
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Requested-With",
		}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 3600 // 1小时
	}

	return &CORSMiddleware{
		allowedOrigins:   config.AllowedOrigins,
		allowedMethods:   config.AllowedMethods,
		allowedHeaders:   config.AllowedHeaders,
		exposedHeaders:   config.ExposedHeaders,
		allowCredentials: config.AllowCredentials,
		maxAge:           config.MaxAge,
		logger:           logger,
	}
}

// Handler 返回CORS处理函数
func (m *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该源
		allowedOrigin := m.getAllowedOrigin(origin)

		if allowedOrigin != "" {
			// 设置CORS响应头
			c.Header("Access-Control-Allow-Origin", allowedOrigin)

			if m.allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if len(m.exposedHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(m.exposedHeaders, ", "))
			}
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			m.handlePreflightRequest(c, allowedOrigin)
			return
		}

		c.Next()
	}
}

// handlePreflightRequest 处理预检请求
func (m *CORSMiddleware) handlePreflightRequest(c *gin.Context, allowedOrigin string) {
	if allowedOrigin == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// 设置允许的方法
	c.Header("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))

	// 设置允许的请求头
	requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
	if requestHeaders != "" {
		// 如果客户端请求了特定的头，检查是否允许
		if m.areHeadersAllowed(requestHeaders) {
			c.Header("Access-Control-Allow-Headers", requestHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
		}
	} else {
		c.Header("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
	}

	// 设置预检请求缓存时间
	c.Header("Access-Control-Max-Age", string(rune(m.maxAge)))

	m.logger.WithFields(logrus.Fields{
		"origin": allowedOrigin,
		"method": "OPTIONS",
	}).Debug("Preflight request handled")

	c.AbortWithStatus(http.StatusNoContent)
}

// getAllowedOrigin 获取允许的源
func (m *CORSMiddleware) getAllowedOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	// 如果允许所有源
	if len(m.allowedOrigins) == 1 && m.allowedOrigins[0] == "*" {
		return "*"
	}

	// 检查是否在允许列表中
	for _, allowed := range m.allowedOrigins {
		if allowed == "*" {
			return origin
		}
		if allowed == origin {
			return origin
		}
		// 支持通配符匹配
		if m.matchOrigin(allowed, origin) {
			return origin
		}
	}

	return ""
}

// matchOrigin 匹配源（支持简单的通配符）
func (m *CORSMiddleware) matchOrigin(pattern, origin string) bool {
	// 简单的通配符匹配：*.example.com
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[2:]
		return strings.HasSuffix(origin, domain) || strings.HasSuffix(origin, "."+domain)
	}
	return false
}

// areHeadersAllowed 检查请求头是否允许
func (m *CORSMiddleware) areHeadersAllowed(requestHeaders string) bool {
	headers := strings.Split(requestHeaders, ",")
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if !m.isHeaderAllowed(header) {
			return false
		}
	}
	return true
}

// isHeaderAllowed 检查单个请求头是否允许
func (m *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(strings.TrimSpace(header))
	for _, allowed := range m.allowedHeaders {
		if strings.ToLower(allowed) == header {
			return true
		}
	}
	return false
}

// DefaultCORS 返回默认的CORS中间件
func DefaultCORS(logger *logrus.Logger) *CORSMiddleware {
	return NewCORSMiddleware(CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}, logger)
}
