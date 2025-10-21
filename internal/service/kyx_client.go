package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
)

// KyxClient 公益站API客户端
type KyxClient struct {
	baseURL    string
	httpClient *http.Client
	session    string
	logger     *logrus.Logger
}

// KyxClientConfig 公益站客户端配置
type KyxClientConfig struct {
	BaseURL string
	Session string
	Timeout time.Duration
}

// NewKyxClient 创建公益站客户端
func NewKyxClient(config KyxClientConfig, logger *logrus.Logger) *KyxClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &KyxClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		session: config.Session,
		logger:  logger,
	}
}

// UpdateSession 更新Session
func (c *KyxClient) UpdateSession(session string) {
	c.session = session
	c.logger.Info("Kyx client session updated")
}

// SearchUser 搜索用户
func (c *KyxClient) SearchUser(ctx context.Context, linuxDoID string) (*model.KyxUser, error) {
	if c.session == "" {
		return nil, fmt.Errorf("session not configured")
	}

	// 构建搜索URL
	searchURL := fmt.Sprintf("%s/api/user?keyword=%s", c.baseURL, url.QueryEscape(linuxDoID))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create search request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.session))
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to search user")
		return nil, fmt.Errorf("failed to search user: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Search user request failed")
		return nil, fmt.Errorf("search user failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var searchResp model.KyxSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		c.logger.WithError(err).WithField("response", string(body)).Error("Failed to parse search response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !searchResp.Success {
		c.logger.WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"message":     searchResp.Message,
		}).Warn("Search user returned unsuccessful")
		return nil, fmt.Errorf("search failed: %s", searchResp.Message)
	}

	// 查找匹配的用户
	for _, user := range searchResp.Data {
		if user.LinuxDoID == linuxDoID {
			c.logger.WithFields(logrus.Fields{
				"linux_do_id": linuxDoID,
				"username":    user.Username,
				"kyx_user_id": user.ID,
			}).Info("User found in Kyx API")
			return &user, nil
		}
	}

	c.logger.WithField("linux_do_id", linuxDoID).Warn("User not found in search results")
	return nil, fmt.Errorf("user not found")
}

// GetUserByID 根据ID获取用户信息
func (c *KyxClient) GetUserByID(ctx context.Context, kyxUserID int) (*model.KyxUser, error) {
	if c.session == "" {
		return nil, fmt.Errorf("session not configured")
	}

	// 构建URL
	userURL := fmt.Sprintf("%s/api/user/%d", c.baseURL, kyxUserID)

	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create get user request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.session))
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).WithField("kyx_user_id", kyxUserID).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Get user request failed")
		return nil, fmt.Errorf("get user failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var user model.KyxUser
	if err := json.Unmarshal(body, &user); err != nil {
		c.logger.WithError(err).WithField("response", string(body)).Error("Failed to parse user response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"kyx_user_id": kyxUserID,
		"username":    user.Username,
	}).Debug("User info retrieved")

	return &user, nil
}

// AddQuota 为用户增加额度
func (c *KyxClient) AddQuota(ctx context.Context, kyxUserID int, quota int64) error {
	if c.session == "" {
		return fmt.Errorf("session not configured")
	}

	// 构建URL
	addQuotaURL := fmt.Sprintf("%s/api/user/%d/quota", c.baseURL, kyxUserID)

	// 构建请求体
	requestBody := map[string]interface{}{
		"quota": quota,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal request body")
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", addQuotaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create add quota request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.session))
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"kyx_user_id": kyxUserID,
			"quota":       quota,
		}).Error("Failed to add quota")
		return fmt.Errorf("failed to add quota: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"kyx_user_id": kyxUserID,
			"quota":       quota,
		}).Error("Add quota request failed")
		return fmt.Errorf("add quota failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.WithFields(logrus.Fields{
		"kyx_user_id": kyxUserID,
		"quota":       quota,
	}).Info("Quota added successfully")

	return nil
}

// GetQuota 获取用户额度信息
func (c *KyxClient) GetQuota(ctx context.Context, kyxUserID int) (quota int64, usedQuota int64, err error) {
	user, err := c.GetUserByID(ctx, kyxUserID)
	if err != nil {
		return 0, 0, err
	}

	return user.Quota, user.UsedQuota, nil
}

// ValidateSession 验证Session是否有效
func (c *KyxClient) ValidateSession(ctx context.Context) error {
	if c.session == "" {
		return fmt.Errorf("session not configured")
	}

	// 尝试获取用户列表来验证Session
	testURL := fmt.Sprintf("%s/api/user?page=1&page_size=1", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.session))
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.logger.Warn("Session validation failed - unauthorized")
		return fmt.Errorf("session is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithField("status_code", resp.StatusCode).Warn("Session validation returned non-OK status")
		return fmt.Errorf("session validation failed with status %d", resp.StatusCode)
	}

	c.logger.Info("Session validated successfully")
	return nil
}

// UpdateGroup 更新用户组
func (c *KyxClient) UpdateGroup(ctx context.Context, kyxUserID int, groupID int) error {
	if c.session == "" {
		return fmt.Errorf("session not configured")
	}

	// 构建URL
	updateURL := fmt.Sprintf("%s/api/user/%d/group", c.baseURL, kyxUserID)

	// 构建请求体
	requestBody := map[string]interface{}{
		"group_id": groupID,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal request body")
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create update group request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", fmt.Sprintf("session=%s", c.session))
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"kyx_user_id": kyxUserID,
			"group_id":    groupID,
		}).Error("Failed to update group")
		return fmt.Errorf("failed to update group: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Update group request failed")
		return fmt.Errorf("update group failed with status %d", resp.StatusCode)
	}

	c.logger.WithFields(logrus.Fields{
		"kyx_user_id": kyxUserID,
		"group_id":    groupID,
	}).Info("User group updated successfully")

	return nil
}

// Ping 测试连接
func (c *KyxClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: status %d", resp.StatusCode)
	}

	return nil
}
