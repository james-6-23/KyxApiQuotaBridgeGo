package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
)

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	cacheService *service.CacheService
	logger       *logrus.Logger
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Limit  int64         // 限制次数
	Window time.Duration // 时间窗口
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(cacheService *service.CacheService, logger *logrus.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cacheService: cacheService,
		logger:       logger,
	}
}

// RateLimitByIP 基于IP的限流
func (m *RateLimitMiddleware) RateLimitByIP(limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := model.RateLimitAPI + "ip:" + clientIP

		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, limit, window)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check rate limit")
			// 如果检查失败，允许请求通过（优雅降级）
			c.Next()
			return
		}

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":    clientIP,
				"path":  c.Request.URL.Path,
				"limit": limit,
			}).Warn("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"rate limit exceeded, please try again later",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser 基于用户的限流
func (m *RateLimitMiddleware) RateLimitByUser(limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户标识
		linuxDoID, exists := c.Get("linux_do_id")
		if !exists {
			// 如果未认证，使用IP限流
			m.RateLimitByIP(limit, window)(c)
			return
		}

		identifier := linuxDoID.(string)
		key := model.RateLimitAPI + "user:" + identifier

		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, limit, window)
		if err != nil {
			m.logger.WithError(err).WithField("user", identifier).Error("Failed to check rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"user":  identifier,
				"path":  c.Request.URL.Path,
				"limit": limit,
			}).Warn("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"rate limit exceeded, please try again later",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit 登录限流（更严格）
func (m *RateLimitMiddleware) LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := model.RateLimitLogin + clientIP

		// 登录限流：每小时最多10次
		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, 10, time.Hour)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Failed to check login rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":   clientIP,
				"path": c.Request.URL.Path,
			}).Warn("Login rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"too many login attempts, please try again later",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// DonateRateLimit 投喂限流
func (m *RateLimitMiddleware) DonateRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		linuxDoID, exists := c.Get("linux_do_id")
		if !exists {
			// 未认证用户不应该访问投喂接口
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse("unauthorized", nil))
			c.Abort()
			return
		}

		identifier := linuxDoID.(string)
		key := model.RateLimitDonate + identifier

		// 投喂限流：每天最多10次
		now := time.Now()
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		window := tomorrow.Sub(now)

		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, 10, window)
		if err != nil {
			m.logger.WithError(err).WithField("user", identifier).Error("Failed to check donate rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", tomorrow.Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"user": identifier,
				"path": c.Request.URL.Path,
			}).Warn("Donate rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"daily donate limit exceeded (max 10 times per day)",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRateLimit 自定义限流
func (m *RateLimitMiddleware) CustomRateLimit(keyFunc func(*gin.Context) string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			// 如果无法获取key，跳过限流
			c.Next()
			return
		}

		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, limit, window)
		if err != nil {
			m.logger.WithError(err).WithField("key", key).Error("Failed to check custom rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"key":   key,
				"path":  c.Request.URL.Path,
				"limit": limit,
			}).Warn("Custom rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"rate limit exceeded, please try again later",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查限流状态
func (m *RateLimitMiddleware) checkRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (allowed bool, remaining int64, err error) {
	allowed, err = m.cacheService.CheckRateLimit(ctx, key, limit, window)
	if err != nil {
		return false, 0, err
	}

	if !allowed {
		return false, 0, nil
	}

	remaining, err = m.cacheService.GetRateLimitRemaining(ctx, key, limit)
	if err != nil {
		return true, 0, err
	}

	return true, remaining, nil
}

// GlobalRateLimit 全局限流（所有请求）
func (m *RateLimitMiddleware) GlobalRateLimit(limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用用户标识，其次使用IP
		var identifier string
		if linuxDoID, exists := c.Get("linux_do_id"); exists {
			identifier = "user:" + linuxDoID.(string)
		} else {
			identifier = "ip:" + c.ClientIP()
		}

		key := model.RateLimitAPI + identifier

		allowed, remaining, err := m.checkRateLimit(c.Request.Context(), key, limit, window)
		if err != nil {
			m.logger.WithError(err).WithField("identifier", identifier).Error("Failed to check global rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		if !allowed {
			m.logger.WithFields(logrus.Fields{
				"identifier": identifier,
				"path":       c.Request.URL.Path,
				"limit":      limit,
			}).Warn("Global rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				"rate limit exceeded, please try again later",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// BypassRateLimitForAdmin 管理员绕过限流
func (m *RateLimitMiddleware) BypassRateLimitForAdmin(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为管理员
		if isAdmin, exists := c.Get("is_admin"); exists && isAdmin.(bool) {
			// 管理员直接跳过限流
			next(c)
			return
		}

		// 非管理员执行限流检查
		next(c)
	}
}
