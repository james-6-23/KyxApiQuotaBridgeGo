package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/middleware"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
	logger      *logrus.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// GetAuthURL 获取OAuth授权URL
// @Summary 获取OAuth授权URL
// @Description 生成Linux.do OAuth授权链接
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 500 {object} model.ErrorResponse
// @Router /api/auth/url [get]
func (h *AuthHandler) GetAuthURL(c *gin.Context) {
	authURL, state, err := h.authService.GetAuthorizationURL(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate authorization URL")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"failed to generate authorization URL",
			err,
		))
		return
	}

	h.logger.WithField("state", state).Info("Authorization URL generated")

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"auth_url": authURL,
			"state":    state,
		},
		"Authorization URL generated successfully",
	))
}

// HandleCallback 处理OAuth回调
// @Summary OAuth回调处理
// @Description 处理Linux.do OAuth授权回调，创建用户会话
// @Tags Auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/auth/callback [get]
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.logger.Warn("OAuth callback missing code")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			"missing authorization code",
			nil,
		))
		return
	}

	if state == "" {
		h.logger.Warn("OAuth callback missing state")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			"missing state parameter",
			nil,
		))
		return
	}

	// 处理回调
	user, sessionID, err := h.authService.HandleCallback(c.Request.Context(), code, state)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"code":  code,
			"state": state,
		}).Error("Failed to handle OAuth callback")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"failed to complete authentication",
			err,
		))
		return
	}

	// 设置Cookie (使用 http.Cookie 以支持 SameSite 属性)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Domain:   "",
		MaxAge:   86400 * 7, // 7天
		Secure:   true,      // HTTPS 环境必须为 true
		HttpOnly: true,      // 防止 XSS 攻击
		SameSite: http.SameSiteLaxMode, // 防止 CSRF 攻击
	})

	h.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"linux_do_id": user.LinuxDoID,
		"username":    user.Username,
	}).Info("User authenticated via OAuth")

	// 重定向到前端回调页面,由前端处理后续跳转
	// 前端会检查 cookie 中的 session_id 来确认登录状态
	c.Redirect(http.StatusFound, "/user/dashboard")
}

// Logout 用户登出
// @Summary 用户登出
// @Description 删除用户会话并清除Cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/auth/logout [post]
// @Security SessionAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"not authenticated",
			nil,
		))
		return
	}

	// 删除会话
	if err := h.authService.DeleteSession(c.Request.Context(), sessionID); err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete session")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"failed to logout",
			err,
		))
		return
	}

	// 清除Cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.logger.WithField("session_id", sessionID).Info("User logged out")

	c.JSON(http.StatusOK, model.NewResponse(
		nil,
		"Logout successful",
	))
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户
// @Description 获取当前已登录用户的信息
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Router /api/auth/me [get]
// @Security SessionAuth
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := middleware.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"not authenticated",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(
		user,
		"User information retrieved",
	))
}

// AdminLogin 管理员登录
// @Summary 管理员登录
// @Description 使用密码进行管理员登录，返回JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.AdminLoginRequest true "Login credentials"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/auth/admin/login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req model.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid admin login request")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			"invalid request body",
			err,
		))
		return
	}

	// 验证密码并生成token
	token, err := h.authService.AdminLogin(c.Request.Context(), req.Password)
	if err != nil {
		h.logger.WithError(err).Warn("Admin login failed")
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"invalid password",
			err,
		))
		return
	}

	h.logger.Info("Admin logged in successfully")

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"token": token,
		},
		"Admin login successful",
	))
}

// RefreshSession 刷新会话
// @Summary 刷新会话
// @Description 延长当前会话的有效期
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/auth/refresh [post]
// @Security SessionAuth
func (h *AuthHandler) RefreshSession(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"not authenticated",
			nil,
		))
		return
	}

	// 刷新会话
	if err := h.authService.RefreshSession(c.Request.Context(), sessionID); err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to refresh session")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"failed to refresh session",
			err,
		))
		return
	}

	h.logger.WithField("session_id", sessionID).Debug("Session refreshed")

	c.JSON(http.StatusOK, model.NewResponse(
		nil,
		"Session refreshed successfully",
	))
}

// CheckAuth 检查认证状态
// @Summary 检查认证状态
// @Description 检查用户是否已登录
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Router /api/auth/check [get]
func (h *AuthHandler) CheckAuth(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusOK, model.NewResponse(
			gin.H{
				"authenticated": false,
			},
			"Not authenticated",
		))
		return
	}

	// 验证会话
	user, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusOK, model.NewResponse(
			gin.H{
				"authenticated": false,
			},
			"Session expired or invalid",
		))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"authenticated": true,
			"user":          user,
		},
		"Authenticated",
	))
}
