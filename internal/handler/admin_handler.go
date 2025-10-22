package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService  *service.AdminService
	userService   *service.UserService
	quotaService  *service.QuotaService
	donateService *service.DonateService
	logger        *logrus.Logger
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(
	adminService *service.AdminService,
	userService *service.UserService,
	quotaService *service.QuotaService,
	donateService *service.DonateService,
	logger *logrus.Logger,
) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		userService:   userService,
		quotaService:  quotaService,
		donateService: donateService,
		logger:        logger,
	}
}

// GetConfig 获取管理员配置
// @Summary 获取配置
// @Description 获取系统配置信息
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/config [get]
// @Security BearerAuth
func (h *AdminHandler) GetConfig(c *gin.Context) {
	config, err := h.adminService.GetConfig(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get admin config, returning default config")
		// 返回默认配置，避免 500 错误阻塞前端
		defaultConfig := &model.AdminConfigResponse{
			ClaimQuota:                  500000,
			SessionConfigured:           false,
			KeysAPIURL:                  "",
			KeysAuthorizationConfigured: false,
			GroupID:                     1,
			UpdatedAt:                   0,
		}
		c.JSON(http.StatusOK, model.NewResponse(defaultConfig, "Config retrieved (default)"))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(config, "Config retrieved"))
}

// UpdateConfig 更新管理员配置
// @Summary 更新配置
// @Description 更新系统配置
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body model.UpdateConfigRequest true "Update config request"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/config [put]
// @Security BearerAuth
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req model.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update config request")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("invalid request body", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"claim_quota":        req.ClaimQuota,
		"session_provided":   req.Session != nil && *req.Session != "",
		"new_api_user":       req.NewAPIUser,
		"keys_api_url":       req.KeysAPIURL,
		"keys_auth_provided": req.KeysAuthorization != nil && *req.KeysAuthorization != "",
		"group_id":           req.GroupID,
	}).Info("Received config update request")

	if err := h.adminService.UpdateConfig(c.Request.Context(), &req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"error_type": fmt.Sprintf("%T", err),
			"error_msg":  err.Error(),
		}).Error("Failed to update config")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to update config", err))
		return
	}

	h.logger.Info("Admin config updated successfully")
	c.JSON(http.StatusOK, model.NewResponse(nil, "Config updated successfully"))
}

// GetSystemStats 获取系统统计
// @Summary 获取系统统计
// @Description 获取系统的统计信息
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/stats [get]
// @Security BearerAuth
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.adminService.GetSystemStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system stats")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get stats", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(stats, "Stats retrieved"))
}

// GetDashboard 获取仪表板数据
// @Summary 获取仪表板
// @Description 获取管理员仪表板的详细数据
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/dashboard [get]
// @Security BearerAuth
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard stats")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get dashboard", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(stats, "Dashboard data retrieved"))
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取所有用户的分页列表
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} model.PaginationResult
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/users [get]
// @Security BearerAuth
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.adminService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to list users", err))
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAllStatistics 获取所有用户统计
// @Summary 获取所有用户统计
// @Description 获取所有用户的统计信息
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/statistics [get]
// @Security BearerAuth
func (h *AdminHandler) GetAllStatistics(c *gin.Context) {
	stats, err := h.adminService.ListAllStatistics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all statistics")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get statistics", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(stats, "Statistics retrieved"))
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户及其所有数据
// @Tags Admin
// @Accept json
// @Produce json
// @Param linux_do_id path string true "Linux Do ID"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/users/{linux_do_id} [delete]
// @Security BearerAuth
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	linuxDoID := c.Param("linux_do_id")
	if linuxDoID == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("linux_do_id is required", nil))
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), linuxDoID); err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to delete user", err))
		return
	}

	h.logger.WithField("linux_do_id", linuxDoID).Info("User deleted by admin")
	c.JSON(http.StatusOK, model.NewResponse(nil, "User deleted successfully"))
}

// ListAllClaims 获取所有领取记录
// @Summary 获取所有领取记录
// @Description 获取所有用户的领取记录
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} model.PaginationResult
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/claims [get]
// @Security BearerAuth
func (h *AdminHandler) ListAllClaims(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.quotaService.ListAllClaims(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list all claims")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to list claims", err))
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	result := &model.PaginationResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
		Data:       records,
	}

	c.JSON(http.StatusOK, result)
}

// ListAllDonates 获取所有投喂记录
// @Summary 获取所有投喂记录
// @Description 获取所有用户的投喂记录
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} model.PaginationResult
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/donates [get]
// @Security BearerAuth
func (h *AdminHandler) ListAllDonates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.donateService.ListAllDonates(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list all donates")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to list donates", err))
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	result := &model.PaginationResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
		Data:       records,
	}

	c.JSON(http.StatusOK, result)
}

// GetRecentActivity 获取最近活动
// @Summary 获取最近活动
// @Description 获取系统最近的活动记录
// @Tags Admin
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/activity [get]
// @Security BearerAuth
func (h *AdminHandler) GetRecentActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	activity, err := h.adminService.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get recent activity")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get activity", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(activity, "Activity retrieved"))
}

// CleanExpiredSessions 清理过期会话
// @Summary 清理过期会话
// @Description 清理所有过期的用户会话
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/maintenance/sessions [post]
// @Security BearerAuth
func (h *AdminHandler) CleanExpiredSessions(c *gin.Context) {
	count, err := h.adminService.CleanExpiredSessions(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to clean expired sessions")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to clean sessions", err))
		return
	}

	h.logger.WithField("count", count).Info("Expired sessions cleaned by admin")
	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{"cleaned_count": count},
		"Expired sessions cleaned",
	))
}

// CleanOldKeys 清理旧Key记录
// @Summary 清理旧Key记录
// @Description 清理指定天数之前的Key记录
// @Tags Admin
// @Accept json
// @Produce json
// @Param days query int false "Days old" default(90)
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/maintenance/keys [post]
// @Security BearerAuth
func (h *AdminHandler) CleanOldKeys(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))

	count, err := h.adminService.CleanOldKeys(c.Request.Context(), days)
	if err != nil {
		h.logger.WithError(err).Error("Failed to clean old keys")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to clean keys", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"count": count,
		"days":  days,
	}).Info("Old keys cleaned by admin")

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"cleaned_count": count,
			"days":          days,
		},
		"Old keys cleaned",
	))
}

// ClearCache 清除缓存
// @Summary 清除缓存
// @Description 清除指定类型的缓存
// @Tags Admin
// @Accept json
// @Produce json
// @Param type query string true "Cache type" Enums(all, user, config)
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/cache/clear [post]
// @Security BearerAuth
func (h *AdminHandler) ClearCache(c *gin.Context) {
	cacheType := c.Query("type")
	if cacheType == "" {
		cacheType = "all"
	}

	if err := h.adminService.ClearCache(c.Request.Context(), cacheType); err != nil {
		h.logger.WithError(err).WithField("type", cacheType).Error("Failed to clear cache")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("failed to clear cache", err))
		return
	}

	h.logger.WithField("type", cacheType).Info("Cache cleared by admin")
	c.JSON(http.StatusOK, model.NewResponse(nil, "Cache cleared successfully"))
}

// TestKyxConnection 测试公益站连接
// @Summary 测试公益站连接
// @Description 测试与公益站的连接是否正常
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/test/kyx [get]
// @Security BearerAuth
func (h *AdminHandler) TestKyxConnection(c *gin.Context) {
	if err := h.adminService.TestKyxConnection(c.Request.Context()); err != nil {
		h.logger.WithError(err).Warn("Kyx connection test failed")
		c.JSON(http.StatusOK, model.NewResponse(
			gin.H{
				"status":  "failed",
				"message": err.Error(),
			},
			"Connection test failed",
		))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{"status": "ok"},
		"Connection test successful",
	))
}

// ValidateKyxSession 验证公益站Session
// @Summary 验证公益站Session
// @Description 验证当前配置的公益站Session是否有效
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/test/session [get]
// @Security BearerAuth
func (h *AdminHandler) ValidateKyxSession(c *gin.Context) {
	if err := h.adminService.ValidateKyxSession(c.Request.Context()); err != nil {
		h.logger.WithError(err).Warn("Kyx session validation failed")
		c.JSON(http.StatusOK, model.NewResponse(
			gin.H{
				"valid":   false,
				"message": err.Error(),
			},
			"Session validation failed",
		))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{"valid": true},
		"Session is valid",
	))
}

// GetHealthStatus 获取系统健康状态
// @Summary 获取健康状态
// @Description 获取系统各组件的健康状态
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/health [get]
// @Security BearerAuth
func (h *AdminHandler) GetHealthStatus(c *gin.Context) {
	health, err := h.adminService.GetHealthStatus(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get health status")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get health status", err))
		return
	}

	// 根据健康状态决定HTTP状态码
	statusCode := http.StatusOK
	if status, ok := health["status"].(string); ok && status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, model.NewResponse(health, "Health status retrieved"))
}

// ExportData 导出数据
// @Summary 导出数据
// @Description 导出指定类型的数据用于备份或分析
// @Tags Admin
// @Accept json
// @Produce json
// @Param type query string true "Data type" Enums(users, claims, donates, statistics)
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/admin/export [get]
// @Security BearerAuth
func (h *AdminHandler) ExportData(c *gin.Context) {
	dataType := c.Query("type")
	if dataType == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("data type is required", nil))
		return
	}

	data, err := h.adminService.ExportData(c.Request.Context(), dataType)
	if err != nil {
		h.logger.WithError(err).WithField("type", dataType).Error("Failed to export data")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to export data", err))
		return
	}

	h.logger.WithField("type", dataType).Info("Data exported by admin")
	c.JSON(http.StatusOK, model.NewResponse(data, "Data exported successfully"))
}
