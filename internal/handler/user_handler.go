package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/middleware"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService   *service.UserService
	quotaService  *service.QuotaService
	donateService *service.DonateService
	logger        *logrus.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(
	userService *service.UserService,
	quotaService *service.QuotaService,
	donateService *service.DonateService,
	logger *logrus.Logger,
) *UserHandler {
	return &UserHandler{
		userService:   userService,
		quotaService:  quotaService,
		donateService: donateService,
		logger:        logger,
	}
}

// BindAccount 绑定公益站账号
// @Summary 绑定账号
// @Description 绑定Linux.do用户与公益站账号
// @Tags User
// @Accept json
// @Produce json
// @Param request body model.BindAccountRequest true "Bind request"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/bind [post]
// @Security SessionAuth
func (h *UserHandler) BindAccount(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	var req model.BindAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid bind account request")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("invalid request body", err))
		return
	}

	// 绑定账号
	response, err := h.userService.BindAccount(c.Request.Context(), linuxDoID, req.Username)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"username":    req.Username,
		}).Error("Failed to bind account")
		
		// 判断是否是 Session 未配置错误
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "session not configured") {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse("请联系管理员配置 Session 密钥后再进行绑定", err))
			return
		}
		
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to bind account", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"linux_do_id":   linuxDoID,
		"username":      req.Username,
		"is_first_bind": response.IsFirstBind,
	}).Info("Account bound successfully")

	c.JSON(http.StatusOK, model.NewResponse(response, "Account bound successfully"))
}

// GetQuota 获取用户额度信息
// @Summary 获取额度信息
// @Description 获取用户在公益站的额度信息
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/quota [get]
// @Security SessionAuth
func (h *UserHandler) GetQuota(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	// 获取额度信息
	quotaInfo, err := h.userService.GetQuotaInfo(c.Request.Context(), linuxDoID)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get quota info")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get quota info", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(quotaInfo, "Quota info retrieved"))
}

// ClaimQuota 领取每日额度
// @Summary 领取每日额度
// @Description 领取每日免费额度
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/claim [post]
// @Security SessionAuth
func (h *UserHandler) ClaimQuota(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	// 领取额度
	record, err := h.quotaService.ClaimQuota(c.Request.Context(), linuxDoID)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to claim quota")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("failed to claim quota", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"quota_added": record.QuotaAdded,
	}).Info("Quota claimed successfully")

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"record":      record,
			"quota_added": record.QuotaAdded,
			"quota_usd":   model.QuotaToDollar(record.QuotaAdded),
		},
		"Quota claimed successfully",
	))
}

// GetClaimHistory 获取领取历史
// @Summary 获取领取历史
// @Description 获取用户的领取历史记录
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} model.PaginationResult
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/claims [get]
// @Security SessionAuth
func (h *UserHandler) GetClaimHistory(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.quotaService.GetClaimHistory(c.Request.Context(), linuxDoID, page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get claim history")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get claim history", err))
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

// DonateKeys 投喂Keys
// @Summary 投喂Keys
// @Description 投喂OpenAI Keys并获得额度奖励
// @Tags User
// @Accept json
// @Produce json
// @Param request body model.DonateRequest true "Donate request"
// @Success 200 {object} model.Response
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/donate [post]
// @Security SessionAuth
func (h *UserHandler) DonateKeys(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	var req model.DonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid donate request")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("invalid request body", err))
		return
	}

	if len(req.Keys) == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse("no keys provided", nil))
		return
	}

	// 投喂Keys
	response, err := h.donateService.DonateKeys(c.Request.Context(), linuxDoID, req.Keys)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"keys_count":  len(req.Keys),
		}).Error("Failed to donate keys")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to donate keys", err))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"valid_keys":  response.ValidKeys,
		"quota_added": response.QuotaAdded,
	}).Info("Keys donated successfully")

	c.JSON(http.StatusOK, model.NewResponse(response, "Keys donated successfully"))
}

// GetDonateHistory 获取投喂历史
// @Summary 获取投喂历史
// @Description 获取用户的投喂历史记录
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} model.PaginationResult
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/donates [get]
// @Security SessionAuth
func (h *UserHandler) GetDonateHistory(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.donateService.GetDonateHistory(c.Request.Context(), linuxDoID, page, pageSize)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get donate history")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get donate history", err))
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

// GetStatistics 获取用户统计信息
// @Summary 获取统计信息
// @Description 获取用户的完整统计信息
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/statistics [get]
// @Security SessionAuth
func (h *UserHandler) GetStatistics(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	// 获取用户统计
	stats, err := h.userService.GetStatistics(c.Request.Context(), linuxDoID)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user statistics")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get statistics", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(stats, "Statistics retrieved"))
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取用户的详细资料信息
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/user/profile [get]
// @Security SessionAuth
func (h *UserHandler) GetProfile(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUser(c.Request.Context(), linuxDoID)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user profile")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to get profile", err))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse("user not found", nil))
		return
	}

	// 获取统计信息
	stats, _ := h.userService.GetStatistics(c.Request.Context(), linuxDoID)

	// 获取额度信息（如果已绑定）
	var quotaInfo *model.QuotaInfo
	if user.KyxUserID > 0 {
		quotaInfo, _ = h.userService.GetQuotaInfo(c.Request.Context(), linuxDoID)
	}

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"user":       user,
			"statistics": stats,
			"quota":      quotaInfo,
		},
		"Profile retrieved",
	))
}

// CheckBindStatus 检查绑定状态
// @Summary 检查绑定状态
// @Description 检查用户是否已绑定公益站账号
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Failure 401 {object} model.ErrorResponse
// @Router /api/user/bind/status [get]
// @Security SessionAuth
func (h *UserHandler) CheckBindStatus(c *gin.Context) {
	linuxDoID, exists := middleware.GetLinuxDoID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse("not authenticated", nil))
		return
	}

	isBound, err := h.userService.IsAccountBound(c.Request.Context(), linuxDoID)
	if err != nil {
		h.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to check bind status")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse("failed to check bind status", err))
		return
	}

	c.JSON(http.StatusOK, model.NewResponse(
		gin.H{
			"is_bound": isBound,
		},
		"Bind status retrieved",
	))
}
