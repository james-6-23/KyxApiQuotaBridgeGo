package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kyx-api-quota-bridge/internal/models"
)

// handleAdminLogin 管理员登录
func (s *Server) handleAdminLogin(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Password != s.config.AdminPassword {
		sendError(c, http.StatusUnauthorized, "密码错误")
		return
	}

	// 创建管理员会话
	sessionID := uuid.New().String()
	session := &models.Session{
		SessionID: sessionID,
		Admin:     true,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	if err := s.db.SaveSession(session); err != nil {
		sendError(c, http.StatusInternalServerError, "保存会话失败")
		return
	}

	c.SetCookie("admin_session", sessionID, 86400, "/", "", false, true)
	sendSuccessWithMessage(c, "登录成功", nil)
}

// handleGetAdminConfig 获取管理员配置
func (s *Server) handleGetAdminConfig(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	config, err := s.db.GetAdminConfig()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	data := map[string]interface{}{
		"claim_quota":                   config.ClaimQuota,
		"session_configured":            config.Session != "",
		"keys_api_url":                  config.KeysAPIURL,
		"keys_authorization_configured": config.KeysAuthorization != "",
		"group_id":                      config.GroupID,
		"updated_at":                    config.UpdatedAt,
	}

	sendSuccess(c, data)
}

// handleUpdateQuota 更新领取额度
func (s *Server) handleUpdateQuota(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		ClaimQuota int64 `json:"claim_quota"`
	}

	if err := c.BindJSON(&req); err != nil || req.ClaimQuota <= 0 {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := s.db.UpdateAdminConfigField("claim_quota", req.ClaimQuota); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "领取额度已更新", nil)
}

// handleUpdateSession 更新 Session
func (s *Server) handleUpdateSession(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		Session string `json:"session"`
	}

	if err := c.BindJSON(&req); err != nil || req.Session == "" {
		sendError(c, http.StatusBadRequest, "Session 不能为空")
		return
	}

	if err := s.db.UpdateAdminConfigField("session", req.Session); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "Session 已更新", nil)
}

// handleUpdateNewAPIUser 更新 new-api-user
func (s *Server) handleUpdateNewAPIUser(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		NewAPIUser string `json:"new_api_user"`
	}

	if err := c.BindJSON(&req); err != nil || req.NewAPIUser == "" {
		sendError(c, http.StatusBadRequest, "new-api-user 不能为空")
		return
	}

	if err := s.db.UpdateAdminConfigField("new_api_user", req.NewAPIUser); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "new-api-user 已更新", nil)
}

// handleUpdateKeysAPIURL 更新 Keys API URL
func (s *Server) handleUpdateKeysAPIURL(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		KeysAPIURL string `json:"keys_api_url"`
	}

	if err := c.BindJSON(&req); err != nil || req.KeysAPIURL == "" {
		sendError(c, http.StatusBadRequest, "Keys API URL 不能为空")
		return
	}

	if err := s.db.UpdateAdminConfigField("keys_api_url", req.KeysAPIURL); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "Keys API URL 已更新", nil)
}

// handleUpdateKeysAuthorization 更新 Keys Authorization
func (s *Server) handleUpdateKeysAuthorization(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		KeysAuthorization string `json:"keys_authorization"`
	}

	if err := c.BindJSON(&req); err != nil || req.KeysAuthorization == "" {
		sendError(c, http.StatusBadRequest, "Keys Authorization 不能为空")
		return
	}

	if err := s.db.UpdateAdminConfigField("keys_authorization", req.KeysAuthorization); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "Keys Authorization 已更新", nil)
}

// handleUpdateGroupID 更新 Group ID
func (s *Server) handleUpdateGroupID(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		GroupID int `json:"group_id"`
	}

	if err := c.BindJSON(&req); err != nil || req.GroupID <= 0 {
		sendError(c, http.StatusBadRequest, "Group ID 不能为空")
		return
	}

	if err := s.db.UpdateAdminConfigField("group_id", req.GroupID); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	sendSuccessWithMessage(c, "Group ID 已更新", nil)
}

// handleExportKeys 导出所有投喂的 Keys
func (s *Server) handleExportKeys(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	keys, err := s.db.GetAllDonatedKeys()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	sendSuccess(c, keys)
}

// handleBatchTestKeys 批量测试 Keys
func (s *Server) handleBatchTestKeys(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		Keys []string `json:"keys"`
	}

	if err := c.BindJSON(&req); err != nil || len(req.Keys) == 0 {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	results := make([]map[string]interface{}, len(req.Keys))
	for i, key := range req.Keys {
		isValid := validateModelScopeKey(key, s.config.ModelScopeAPIBase)
		results[i] = map[string]interface{}{
			"key":   key,
			"valid": isValid,
		}
	}

	sendSuccess(c, results)
}

// handleDeleteKeys 删除指定的 Keys
func (s *Server) handleDeleteKeys(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		Keys []string `json:"keys"`
	}

	if err := c.BindJSON(&req); err != nil || len(req.Keys) == 0 {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := s.db.DeleteKeys(req.Keys); err != nil {
		sendError(c, http.StatusInternalServerError, "删除失败")
		return
	}

	sendSuccessWithMessage(c, fmt.Sprintf("已删除 %d 个 Keys", len(req.Keys)), nil)
}

// handleGetAllClaimRecords 获取所有领取记录
func (s *Server) handleGetAllClaimRecords(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	records, err := s.db.GetAllClaimRecords()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	sendSuccess(c, records)
}

// handleGetAllDonateRecords 获取所有投喂记录
func (s *Server) handleGetAllDonateRecords(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	records, err := s.db.GetAllDonateRecords()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	sendSuccess(c, records)
}

// handleGetAllUsers 获取所有用户
func (s *Server) handleGetAllUsers(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	users, err := s.db.GetAllUsers()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	claimRecords, _ := s.db.GetAllClaimRecords()
	donateRecords, _ := s.db.GetAllDonateRecords()

	// 统计每个用户的数据
	type UserStats struct {
		Username    string `json:"username"`
		LinuxDoID   string `json:"linux_do_id"`
		CreatedAt   int64  `json:"created_at"`
		DonateCount int    `json:"donate_count"`
		DonateQuota int64  `json:"donate_quota"`
		ClaimCount  int    `json:"claim_count"`
		ClaimQuota  int64  `json:"claim_quota"`
		TotalQuota  int64  `json:"total_quota"`
	}

	stats := make([]UserStats, len(users))
	for i, user := range users {
		userStat := UserStats{
			Username:  user.Username,
			LinuxDoID: user.LinuxDoID,
			CreatedAt: user.CreatedAt,
		}

		// 统计领取记录
		for _, record := range claimRecords {
			if record.LinuxDoID == user.LinuxDoID {
				userStat.ClaimCount++
				userStat.ClaimQuota += record.QuotaAdded
			}
		}

		// 统计投喂记录
		for _, record := range donateRecords {
			if record.LinuxDoID == user.LinuxDoID {
				userStat.DonateCount += record.KeysCount
				userStat.DonateQuota += record.TotalQuotaAdded
			}
		}

		userStat.TotalQuota = userStat.ClaimQuota + userStat.DonateQuota
		stats[i] = userStat
	}

	sendSuccess(c, stats)
}

// handleRebindUser 管理员重新绑定用户账号
func (s *Server) handleRebindUser(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		LinuxDoID   string `json:"linux_do_id"`
		NewUsername string `json:"new_username"`
	}

	if err := c.BindJSON(&req); err != nil || req.LinuxDoID == "" || req.NewUsername == "" {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 获取当前用户
	currentUser, err := s.db.GetUser(req.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	if currentUser == nil {
		sendError(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 获取管理员配置
	config, err := s.db.GetAdminConfig()
	if err != nil || config.Session == "" {
		sendError(c, http.StatusInternalServerError, "系统配置错误，请联系管理员")
		return
	}

	// 搜索新用户名
	kyxResult, err := searchKyxUser(req.NewUsername, config.Session, config.NewAPIUser, s.config.KyxAPIBase)
	if err != nil || !kyxResult.Success {
		sendError(c, http.StatusBadRequest, "查询失败")
		return
	}

	// 精确匹配用户名
	kyxUser := findExactUser(kyxResult, req.NewUsername)
	if kyxUser == nil {
		sendError(c, http.StatusNotFound, "未找到该用户")
		return
	}

	// 验证 Linux Do ID 是否匹配
	if kyxUser.LinuxDoID != req.LinuxDoID {
		msg := fmt.Sprintf("Linux Do ID 不匹配，当前用户的 Linux Do ID 是 %s，但搜索到的用户 %s 的 Linux Do ID 是 %s",
			req.LinuxDoID, req.NewUsername, kyxUser.LinuxDoID)
		sendError(c, http.StatusBadRequest, msg)
		return
	}

	// 更新用户绑定
	currentUser.Username = kyxUser.Username
	currentUser.KyxUserID = kyxUser.ID

	if err := s.db.SetUser(currentUser); err != nil {
		sendError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	message := fmt.Sprintf("用户重新绑定成功，从 %s 更新为 %s", currentUser.Username, kyxUser.Username)
	sendSuccessWithMessage(c, message, nil)
}

// handleRetryPush 重新推送失败的 Keys
func (s *Server) handleRetryPush(c *gin.Context) {
	if !s.checkAdminAuth(c) {
		return
	}

	var req struct {
		LinuxDoID string `json:"linux_do_id"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := c.BindJSON(&req); err != nil || req.LinuxDoID == "" || req.Timestamp == 0 {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 获取投喂记录
	record, err := s.db.GetDonateRecord(req.LinuxDoID, req.Timestamp)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询记录失败")
		return
	}

	if record == nil {
		sendError(c, http.StatusNotFound, "未找到投喂记录")
		return
	}

	if len(record.FailedKeys) == 0 {
		sendError(c, http.StatusBadRequest, "没有失败的 Keys")
		return
	}

	// 获取管理员配置
	config, err := s.db.GetAdminConfig()
	if err != nil || config.KeysAuthorization == "" {
		sendError(c, http.StatusInternalServerError, "未配置推送授权")
		return
	}

	// 重新推送
	success, message, failedKeys := pushKeysToGroup(record.FailedKeys, config.KeysAPIURL, config.KeysAuthorization, config.GroupID)

	// 更新记录
	if success {
		record.PushStatus = "success"
		record.PushMessage = "推送成功"
		record.FailedKeys = nil
	} else {
		record.PushStatus = "failed"
		record.PushMessage = message
		record.FailedKeys = failedKeys
	}

	if err := s.db.UpdateDonateRecord(req.LinuxDoID, req.Timestamp, record); err != nil {
		sendError(c, http.StatusInternalServerError, "更新记录失败")
		return
	}

	if success {
		sendSuccessWithMessage(c, "重新推送成功", nil)
	} else {
		sendError(c, http.StatusInternalServerError, "重新推送失败: "+message)
	}
}

// checkAdminAuth 检查管理员权限
func (s *Server) checkAdminAuth(c *gin.Context) bool {
	sessionID, err := c.Cookie("admin_session")
	if err != nil {
		sendError(c, http.StatusUnauthorized, "未授权")
		return false
	}

	session, err := s.db.GetSession(sessionID)
	if err != nil || session == nil || !session.Admin {
		sendError(c, http.StatusUnauthorized, "未授权")
		return false
	}

	return true
}
