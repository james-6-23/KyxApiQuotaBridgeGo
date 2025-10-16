package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kyx-api-quota-bridge/internal/models"
	"golang.org/x/oauth2"
)

// handleIndex 首页
func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// handleLogin 登录
func (s *Server) handleLogin(c *gin.Context) {
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.LinuxDoClientID,
		ClientSecret: s.config.LinuxDoClientSecret,
		RedirectURL:  s.config.LinuxDoRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.LinuxDoAuthURL,
			TokenURL: s.config.LinuxDoTokenURL,
		},
		Scopes: []string{"read"},
	}

	state := uuid.New().String()
	url := oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// handleCallback OAuth 回调
func (s *Server) handleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		sendError(c, http.StatusBadRequest, "缺少授权码")
		return
	}

	oauth2Config := &oauth2.Config{
		ClientID:     s.config.LinuxDoClientID,
		ClientSecret: s.config.LinuxDoClientSecret,
		RedirectURL:  s.config.LinuxDoRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.config.LinuxDoAuthURL,
			TokenURL: s.config.LinuxDoTokenURL,
		},
	}

	token, err := oauth2Config.Exchange(c.Request.Context(), code)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取令牌失败")
		return
	}

	// 获取用户信息
	client := oauth2Config.Client(c.Request.Context(), token)
	resp, err := client.Get(s.config.LinuxDoUserInfoURL)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Avatar   string `json:"avatar_url"`
	}

	if err := parseJSON(resp.Body, &userInfo); err != nil {
		sendError(c, http.StatusInternalServerError, "解析用户信息失败")
		return
	}

	// 创建会话
	sessionID := uuid.New().String()
	session := &models.Session{
		SessionID: sessionID,
		LinuxDoID: fmt.Sprintf("%d", userInfo.ID),
		Username:  userInfo.Username,
		Name:      userInfo.Name,
		AvatarURL: userInfo.Avatar,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	if err := s.db.SaveSession(session); err != nil {
		sendError(c, http.StatusInternalServerError, "保存会话失败")
		return
	}

	// 设置 Cookie
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// handleBind 绑定公益站账号
func (s *Server) handleBind(c *gin.Context) {
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

	var req struct {
		Username string `json:"username"`
	}
	if err := c.BindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Username == "" {
		sendError(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	// 获取管理员配置
	config, err := s.db.GetAdminConfig()
	if err != nil || config.Session == "" {
		sendError(c, http.StatusInternalServerError, "系统配置错误，请联系管理员")
		return
	}

	// 搜索公益站用户
	kyxResult, err := searchKyxUser(req.Username, config.Session, config.NewAPIUser, s.config.KyxAPIBase)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	if !kyxResult.Success {
		sendError(c, http.StatusBadRequest, "查询失败: "+kyxResult.Message)
		return
	}

	// 精确匹配用户名
	kyxUser := findExactUser(kyxResult, req.Username)
	if kyxUser == nil {
		sendError(c, http.StatusNotFound, "未找到该用户，请确认用户名输入正确")
		return
	}

	// 验证 Linux Do ID 是否匹配
	if kyxUser.LinuxDoID != session.LinuxDoID {
		msg := fmt.Sprintf("Linux Do ID 不匹配！\n您当前登录的 Linux Do ID: %s\n用户 %s 的 Linux Do ID: %s\n请使用正确的 Linux Do 账号登录后再绑定此用户名。",
			session.LinuxDoID, req.Username, kyxUser.LinuxDoID)
		sendError(c, http.StatusBadRequest, msg)
		return
	}

	// 保存用户绑定
	user := &models.User{
		LinuxDoID: session.LinuxDoID,
		Username:  kyxUser.Username,
		KyxUserID: kyxUser.ID,
		CreatedAt: time.Now().Unix(),
	}

	if err := s.db.SetUser(user); err != nil {
		sendError(c, http.StatusInternalServerError, "绑定失败")
		return
	}

	sendSuccessWithMessage(c, "绑定成功", user)
}

// handleLogout 登出
func (s *Server) handleLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("session_id")
	if sessionID != "" {
		s.db.DeleteSession(sessionID)
	}
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	sendSuccess(c, nil)
}

// handleGetUserQuota 获取用户额度
func (s *Server) handleGetUserQuota(c *gin.Context) {
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

	// 获取管理员配置
	config, err := s.db.GetAdminConfig()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	// 查询公益站用户信息
	kyxResult, err := searchKyxUser(user.Username, config.Session, config.NewAPIUser, s.config.KyxAPIBase)
	if err != nil || !kyxResult.Success {
		sendError(c, http.StatusInternalServerError, "查询额度失败")
		return
	}

	kyxUser := findExactUser(kyxResult, user.Username)
	if kyxUser == nil {
		sendError(c, http.StatusNotFound, "未找到用户信息")
		return
	}

	// 检查今天是否已领取
	claimedToday, _ := s.db.GetClaimToday(user.LinuxDoID)

	data := map[string]interface{}{
		"username":      kyxUser.Username,
		"display_name":  kyxUser.DisplayName,
		"linux_do_id":   user.LinuxDoID,
		"avatar_url":    session.AvatarURL,
		"name":          session.Name,
		"quota":         kyxUser.Quota,
		"used_quota":    kyxUser.UsedQuota,
		"total":         kyxUser.Quota + kyxUser.UsedQuota,
		"can_claim":     kyxUser.Quota < s.config.MinQuotaThreshold && !claimedToday,
		"claimed_today": claimedToday,
	}

	sendSuccess(c, data)
}

// handleGetUserClaimRecords 获取用户领取记录
func (s *Server) handleGetUserClaimRecords(c *gin.Context) {
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

	records, err := s.db.GetUserClaimRecords(session.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	sendSuccess(c, records)
}

// handleGetUserDonateRecords 获取用户投喂记录
func (s *Server) handleGetUserDonateRecords(c *gin.Context) {
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

	records, err := s.db.GetUserDonateRecords(session.LinuxDoID)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	sendSuccess(c, records)
}
