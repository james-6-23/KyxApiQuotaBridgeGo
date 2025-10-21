package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
)

// LinuxDoClient Linux Do OAuth客户端
type LinuxDoClient struct {
	clientID     string
	clientSecret string
	redirectURI  string
	baseURL      string
	httpClient   *http.Client
	logger       *logrus.Logger
}

// LinuxDoClientConfig Linux Do客户端配置
type LinuxDoClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	BaseURL      string
	Timeout      time.Duration
}

// NewLinuxDoClient 创建Linux Do客户端
func NewLinuxDoClient(config LinuxDoClientConfig, logger *logrus.Logger) *LinuxDoClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://connect.linux.do"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &LinuxDoClient{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		redirectURI:  config.RedirectURI,
		baseURL:      config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// GetAuthorizationURL 获取OAuth授权URL
func (c *LinuxDoClient) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", c.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "read")
	params.Set("state", state)

	authURL := fmt.Sprintf("%s/oauth2/authorize?%s", c.baseURL, params.Encode())

	c.logger.WithFields(logrus.Fields{
		"state":        state,
		"redirect_uri": c.redirectURI,
	}).Debug("Generated authorization URL")

	return authURL
}

// ExchangeCode 交换授权码获取访问令牌
func (c *LinuxDoClient) ExchangeCode(ctx context.Context, code string) (*model.LinuxDoTokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/oauth2/token", c.baseURL)

	// 构建请求参数
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create token exchange request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to exchange authorization code")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read token response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Token exchange failed")
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var tokenResp model.LinuxDoTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		c.logger.WithError(err).WithField("response", string(body)).Error("Failed to parse token response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"token_type": tokenResp.TokenType,
		"expires_in": tokenResp.ExpiresIn,
	}).Info("Successfully exchanged authorization code for token")

	return &tokenResp, nil
}

// RefreshToken 刷新访问令牌
func (c *LinuxDoClient) RefreshToken(ctx context.Context, refreshToken string) (*model.LinuxDoTokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/oauth2/token", c.baseURL)

	// 构建请求参数
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create token refresh request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to refresh token")
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read refresh token response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Token refresh failed")
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var tokenResp model.LinuxDoTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		c.logger.WithError(err).WithField("response", string(body)).Error("Failed to parse refresh token response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("Successfully refreshed access token")

	return &tokenResp, nil
}

// GetUserInfo 获取用户信息
func (c *LinuxDoClient) GetUserInfo(ctx context.Context, accessToken string) (*model.LinuxDoUserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/api/user", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create user info request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read user info response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Warn("Access token is invalid or expired")
		return nil, fmt.Errorf("access token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Get user info failed")
		return nil, fmt.Errorf("get user info failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var userInfo model.LinuxDoUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		c.logger.WithError(err).WithField("response", string(body)).Error("Failed to parse user info response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"user_id":  userInfo.ID,
		"username": userInfo.Username,
	}).Info("Successfully retrieved user info")

	return &userInfo, nil
}

// ValidateToken 验证访问令牌是否有效
func (c *LinuxDoClient) ValidateToken(ctx context.Context, accessToken string) error {
	_, err := c.GetUserInfo(ctx, accessToken)
	return err
}

// RevokeToken 撤销访问令牌
func (c *LinuxDoClient) RevokeToken(ctx context.Context, token string) error {
	revokeURL := fmt.Sprintf("%s/oauth2/revoke", c.baseURL)

	// 构建请求参数
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create token revoke request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to revoke token")
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read revoke response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Warn("Token revoke request returned non-OK status")
		// 撤销失败不应该阻止流程，只记录警告
		return nil
	}

	c.logger.Info("Successfully revoked token")
	return nil
}

// Ping 测试连接
func (c *LinuxDoClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "KyxQuotaBridge/1.0")

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

// GetClientID 获取客户端ID
func (c *LinuxDoClient) GetClientID() string {
	return c.clientID
}

// UpdateConfig 更新客户端配置
func (c *LinuxDoClient) UpdateConfig(clientID, clientSecret, redirectURI string) {
	if clientID != "" {
		c.clientID = clientID
	}
	if clientSecret != "" {
		c.clientSecret = clientSecret
	}
	if redirectURI != "" {
		c.redirectURI = redirectURI
	}
	c.logger.Info("LinuxDo client configuration updated")
}
