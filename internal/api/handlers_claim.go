package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kyx-api-quota-bridge/internal/models"
)

// handleDailyClaim 每日领取额度
func (s *Server) handleDailyClaim(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		sendError(c, http.StatusUnauthorized, "未登录")
		return
	}

	session, err := s.db.GetSession(sessionID)
	if err != nil || session == nil {
		sendError(c, http.StatusUnauthorized, "会话无效")
		return
	}

	user, err := s.db.GetUser(session.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取用户失败")
		return
	}

	if user == nil {
		sendError(c, http.StatusBadRequest, "未绑定账号")
		return
	}

	// 检查今天是否已领取
	claimedToday, err := s.db.GetClaimToday(user.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询领取记录失败")
		return
	}

	if claimedToday {
		sendError(c, http.StatusBadRequest, "今天已经领取过了")
		return
	}

	// 获取管理员配置
	config, err := s.db.GetAdminConfig()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	// 查询公益站用户信息
	kyxResult, err := searchKyxUser(user.Username, config.Session, config.NewAPIUser, s.config.KyxAPIBase)
	if err != nil || !kyxResult.Success {
		sendError(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	kyxUser := findExactUser(kyxResult, user.Username)
	if kyxUser == nil {
		sendError(c, http.StatusNotFound, "未找到用户信息")
		return
	}

	// 检查额度是否充足
	if kyxUser.Quota >= s.config.MinQuotaThreshold {
		sendError(c, http.StatusBadRequest, "额度充足，未达到领取要求")
		return
	}

	// 更新额度
	newQuota := kyxUser.Quota + config.ClaimQuota
	err = updateKyxUserQuota(user.KyxUserID, newQuota, config.Session, config.NewAPIUser,
		kyxUser.Username, kyxUser.Group, s.config.KyxAPIBase)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "额度添加失败: "+err.Error())
		return
	}

	// 记录领取
	record := &models.ClaimRecord{
		LinuxDoID:  user.LinuxDoID,
		Username:   user.Username,
		QuotaAdded: config.ClaimQuota,
		Timestamp:  time.Now().Unix(),
		Date:       time.Now().Format("2006-01-02"),
	}

	if err := s.db.AddClaimRecord(record); err != nil {
		// 记录失败不影响领取成功
		fmt.Printf("Failed to save claim record: %v\n", err)
	}

	quotaCNY := float64(config.ClaimQuota) * 0.02
	message := fmt.Sprintf("领取成功！已为您增加 ¥%.2f 额度", quotaCNY)
	sendSuccessWithMessage(c, message, map[string]interface{}{
		"quota_added": config.ClaimQuota,
		"new_quota":   newQuota,
	})
}

// handleTestKey 测试单个 Key
func (s *Server) handleTestKey(c *gin.Context) {
	var req struct {
		Key string `json:"key"`
	}

	if err := c.BindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Key == "" {
		sendError(c, http.StatusBadRequest, "Key 不能为空")
		return
	}

	isValid := validateModelScopeKey(req.Key, s.config.ModelScopeAPIBase)

	sendSuccess(c, map[string]interface{}{
		"key":   req.Key[:10] + "...",
		"valid": isValid,
	})
}
