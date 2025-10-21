package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authService *service.AuthService
	logger      *logrus.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authService *service.AuthService, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth 要求用户认证
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Cookie获取session_id
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			m.logger.WithField("path", c.Request.URL.Path).Debug("No session cookie found")
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("unauthorized", nil))
			c.Abort()
			return
		}

		// 验证会话
		user, err := m.authService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			m.logger.WithError(err).WithField("session_id", sessionID).Warn("Invalid session")
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("session expired or invalid", err))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Set("session_id", sessionID)
		c.Set("linux_do_id", user.LinuxDoID)
		c.Set("username", user.Username)

		m.logger.WithFields(logrus.Fields{
			"linux_do_id": user.LinuxDoID,
			"username":    user.Username,
			"path":        c.Request.URL.Path,
		}).Debug("User authenticated")

		c.Next()
	}
}

// RequireAdmin 要求管理员认证
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.WithField("path", c.Request.URL.Path).Debug("No authorization header")
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("unauthorized", nil))
			c.Abort()
			return
		}

		// 提取Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.WithField("auth_header", authHeader).Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("invalid authorization header", nil))
			c.Abort()
			return
		}

		token := parts[1]

		// 验证admin token
		err := m.authService.ValidateAdminToken(token)
		if err != nil {
			m.logger.WithError(err).Warn("Invalid admin token")
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("invalid or expired token", err))
			c.Abort()
			return
		}

		// 设置管理员标识
		c.Set("is_admin", true)

		m.logger.WithField("path", c.Request.URL.Path).Debug("Admin authenticated")

		c.Next()
	}
}

// OptionalAuth 可选认证（不强制要求）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从Cookie获取session_id
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			// 没有session，继续执行
			c.Next()
			return
		}

		// 尝试验证会话
		user, err := m.authService.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			// 会话无效，但不阻止请求
			m.logger.WithError(err).Debug("Optional auth: invalid session")
			c.Next()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Set("session_id", sessionID)
		c.Set("linux_do_id", user.LinuxDoID)
		c.Set("username", user.Username)

		c.Next()
	}
}

// GetUser 从上下文获取用户信息
func GetUser(c *gin.Context) (*model.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*model.User); ok {
			return u, true
		}
	}
	return nil, false
}

// GetLinuxDoID 从上下文获取LinuxDoID
func GetLinuxDoID(c *gin.Context) (string, bool) {
	if linuxDoID, exists := c.Get("linux_do_id"); exists {
		if id, ok := linuxDoID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name, true
		}
	}
	return "", false
}

// GetSessionID 从上下文获取SessionID
func GetSessionID(c *gin.Context) (string, bool) {
	if sessionID, exists := c.Get("session_id"); exists {
		if id, ok := sessionID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// IsAdmin 检查是否为管理员
func IsAdmin(c *gin.Context) bool {
	if isAdmin, exists := c.Get("is_admin"); exists {
		if admin, ok := isAdmin.(bool); ok {
			return admin
		}
	}
	return false
}
