package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
)

// RecoveryMiddleware 恢复中间件
type RecoveryMiddleware struct {
	logger            *logrus.Logger
	stackAll          bool // 是否打印所有goroutine的堆栈
	stackSize         int  // 堆栈大小限制
	disablePrintStack bool // 是否禁用打印堆栈
}

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
	StackAll          bool
	StackSize         int
	DisablePrintStack bool
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(logger *logrus.Logger, config RecoveryConfig) *RecoveryMiddleware {
	if config.StackSize == 0 {
		config.StackSize = 4 << 10 // 4KB
	}

	return &RecoveryMiddleware{
		logger:            logger,
		stackAll:          config.StackAll,
		stackSize:         config.StackSize,
		disablePrintStack: config.DisablePrintStack,
	}
}

// Handler 返回恢复处理函数
func (m *RecoveryMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				m.handlePanic(c, err)
			}
		}()

		c.Next()
	}
}

// handlePanic 处理panic
func (m *RecoveryMiddleware) handlePanic(c *gin.Context, err interface{}) {
	// 获取堆栈信息
	stack := m.getStack()

	// 构建错误信息
	errMsg := fmt.Sprintf("%v", err)

	// 构建日志字段
	fields := logrus.Fields{
		"error":      errMsg,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}

	// 添加用户信息（如果已认证）
	if linuxDoID, exists := c.Get("linux_do_id"); exists {
		fields["linux_do_id"] = linuxDoID
	}
	if username, exists := c.Get("username"); exists {
		fields["username"] = username
	}

	// 添加查询参数
	if len(c.Request.URL.RawQuery) > 0 {
		fields["query"] = c.Request.URL.RawQuery
	}

	// 添加堆栈信息
	if !m.disablePrintStack {
		fields["stack"] = stack
	}

	// 记录错误日志
	m.logger.WithFields(fields).Error("Panic recovered")

	// 检查连接是否已断开
	if m.isBrokenPipe(err) {
		m.logger.WithFields(fields).Warn("Connection broken, client disconnected")
		c.Abort()
		return
	}

	// 返回错误响应
	c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
		"internal server error",
		fmt.Errorf("%v", err),
	))

	c.Abort()
}

// getStack 获取堆栈信息
func (m *RecoveryMiddleware) getStack() string {
	if m.stackAll {
		return string(debug.Stack())
	}

	// 获取指定大小的堆栈
	buf := make([]byte, m.stackSize)
	n := debug.Stack()
	if len(n) < m.stackSize {
		buf = n
	} else {
		buf = n[:m.stackSize]
	}

	return string(buf)
}

// isBrokenPipe 检查是否为连接中断错误
func (m *RecoveryMiddleware) isBrokenPipe(err interface{}) bool {
	if err == nil {
		return false
	}

	errStr := fmt.Sprintf("%v", err)
	brokenPipeErrors := []string{
		"broken pipe",
		"connection reset by peer",
		"write: broken pipe",
		"write: connection reset by peer",
		"read: connection reset by peer",
	}

	for _, msg := range brokenPipeErrors {
		if strings.Contains(strings.ToLower(errStr), msg) {
			return true
		}
	}

	return false
}

// CustomRecovery 自定义恢复处理
func (m *RecoveryMiddleware) CustomRecovery(handler func(*gin.Context, interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录日志
				stack := m.getStack()
				fields := logrus.Fields{
					"error":      fmt.Sprintf("%v", err),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}

				if !m.disablePrintStack {
					fields["stack"] = stack
				}

				m.logger.WithFields(fields).Error("Panic recovered with custom handler")

				// 调用自定义处理函数
				handler(c, err)
			}
		}()

		c.Next()
	}
}

// DefaultRecovery 返回默认的恢复中间件
func DefaultRecovery(logger *logrus.Logger) *RecoveryMiddleware {
	return NewRecoveryMiddleware(logger, RecoveryConfig{
		StackAll:          false,
		StackSize:         4 << 10, // 4KB
		DisablePrintStack: false,
	})
}

// RecoveryWithWriter 带自定义writer的恢复中间件
func (m *RecoveryMiddleware) RecoveryWithWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否为连接中断
				if m.isBrokenPipe(err) {
					// 如果是连接中断，只记录警告
					m.logger.WithFields(logrus.Fields{
						"error":     fmt.Sprintf("%v", err),
						"method":    c.Request.Method,
						"path":      c.Request.URL.Path,
						"client_ip": c.ClientIP(),
					}).Warn("Connection broken during request")
					c.Abort()
					return
				}

				// 其他panic
				m.handlePanic(c, err)
			}
		}()

		c.Next()
	}
}

// SafeHandler 安全的处理函数包装器（用于异步操作）
func (m *RecoveryMiddleware) SafeHandler(handler func()) {
	defer func() {
		if err := recover(); err != nil {
			stack := m.getStack()
			m.logger.WithFields(logrus.Fields{
				"error": fmt.Sprintf("%v", err),
				"stack": stack,
			}).Error("Panic recovered in safe handler")
		}
	}()

	handler()
}

// SafeGo 安全的goroutine启动
func (m *RecoveryMiddleware) SafeGo(handler func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := m.getStack()
				m.logger.WithFields(logrus.Fields{
					"error": fmt.Sprintf("%v", err),
					"stack": stack,
				}).Error("Panic recovered in goroutine")
			}
		}()

		handler()
	}()
}

// PanicLogger panic日志记录器
type PanicLogger struct {
	logger *logrus.Logger
}

// NewPanicLogger 创建panic日志记录器
func NewPanicLogger(logger *logrus.Logger) *PanicLogger {
	return &PanicLogger{logger: logger}
}

// Log 记录panic日志
func (p *PanicLogger) Log(err interface{}, stack string) {
	p.logger.WithFields(logrus.Fields{
		"error": fmt.Sprintf("%v", err),
		"stack": stack,
	}).Error("Panic occurred")
}

// RecoveryWithCallback 带回调的恢复中间件
func (m *RecoveryMiddleware) RecoveryWithCallback(callback func(*gin.Context, interface{}, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := m.getStack()

				// 记录日志
				m.logger.WithFields(logrus.Fields{
					"error":     fmt.Sprintf("%v", err),
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"client_ip": c.ClientIP(),
					"stack":     stack,
				}).Error("Panic recovered with callback")

				// 调用回调
				if callback != nil {
					callback(c, err, stack)
				}

				// 返回错误响应
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
						"internal server error",
						fmt.Errorf("%v", err),
					))
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}
