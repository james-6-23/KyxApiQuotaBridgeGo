package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kyx-api-quota-bridge/internal/config"
	"github.com/kyx-api-quota-bridge/internal/store"
)

// Server API 服务器
type Server struct {
	config *config.Config
	db     *store.DB
	router *gin.Engine
}

// NewServer 创建新的 API 服务器
func NewServer(cfg *config.Config, db *store.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	s := &Server{
		config: cfg,
		db:     db,
		router: router,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 页面路由
	s.router.GET("/", s.handleUserPage)
	s.router.GET("/admin", s.handleAdminPage)

	// OAuth 回调
	s.router.GET("/oauth/callback", s.handleCallback)

	// 用户 API
	userAPI := s.router.Group("/api")
	{
		// 认证相关
		userAPI.POST("/auth/bind", s.handleBind)
		userAPI.POST("/auth/logout", s.handleLogout)

		// 用户信息
		userAPI.GET("/user/quota", s.handleGetUserQuota)
		userAPI.GET("/user/records/claim", s.handleGetUserClaimRecords)
		userAPI.GET("/user/records/donate", s.handleGetUserDonateRecords)

		// 领取和投喂
		userAPI.POST("/claim/daily", s.handleDailyClaim)
		userAPI.POST("/donate/validate", s.handleDonateValidate)

		// Key 测试
		userAPI.POST("/test/key", s.handleTestKey)
	}

	// 管理员 API
	adminAPI := s.router.Group("/api/admin")
	{
		// 登录
		adminAPI.POST("/login", s.handleAdminLogin)

		// 配置管理
		adminAPI.GET("/config", s.handleGetAdminConfig)
		adminAPI.PUT("/config/quota", s.handleUpdateQuota)
		adminAPI.PUT("/config/session", s.handleUpdateSession)
		adminAPI.PUT("/config/new-api-user", s.handleUpdateNewAPIUser)
		adminAPI.PUT("/config/keys-api-url", s.handleUpdateKeysAPIURL)
		adminAPI.PUT("/config/keys-authorization", s.handleUpdateKeysAuthorization)
		adminAPI.PUT("/config/group-id", s.handleUpdateGroupID)

		// Keys 管理
		adminAPI.GET("/keys/export", s.handleExportKeys)
		adminAPI.POST("/keys/test", s.handleBatchTestKeys)
		adminAPI.POST("/keys/delete", s.handleDeleteKeys)

		// 记录查询
		adminAPI.GET("/records/claim", s.handleGetAllClaimRecords)
		adminAPI.GET("/records/donate", s.handleGetAllDonateRecords)

		// 用户管理
		adminAPI.GET("/users", s.handleGetAllUsers)
		adminAPI.POST("/rebind-user", s.handleRebindUser)

		// 重新推送 Keys
		adminAPI.POST("/retry-push", s.handleRetryPush)
	}
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// sendJSON 发送 JSON 响应
func sendJSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

// sendError 发送错误响应
func sendError(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Success: false,
		Message: message,
	})
}

// sendSuccess 发送成功响应
func sendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// sendSuccessWithMessage 发送带消息的成功响应
func sendSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}
