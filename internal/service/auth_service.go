package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
)

// AuthService 认证服务
type AuthService struct {
	linuxDoClient  *LinuxDoClient
	sessionRepo    *repository.SessionRepository
	userRepo       *repository.UserRepository
	cacheService   *CacheService
	jwtSecret      string
	adminPassword  string
	sessionTimeout time.Duration
	logger         *logrus.Logger
}

// AuthServiceConfig 认证服务配置
type AuthServiceConfig struct {
	JWTSecret      string
	AdminPassword  string
	SessionTimeout time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(
	linuxDoClient *LinuxDoClient,
	sessionRepo *repository.SessionRepository,
	userRepo *repository.UserRepository,
	cacheService *CacheService,
	config AuthServiceConfig,
	logger *logrus.Logger,
) *AuthService {
	if config.SessionTimeout == 0 {
		config.SessionTimeout = 24 * time.Hour
	}

	return &AuthService{
		linuxDoClient:  linuxDoClient,
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		cacheService:   cacheService,
		jwtSecret:      config.JWTSecret,
		adminPassword:  config.AdminPassword,
		sessionTimeout: config.SessionTimeout,
		logger:         logger,
	}
}

// GenerateState 生成OAuth状态参数
func (s *AuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		s.logger.WithError(err).Error("Failed to generate state")
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}

// GetAuthorizationURL 获取OAuth授权URL
func (s *AuthService) GetAuthorizationURL(ctx context.Context) (string, string, error) {
	// 生成state
	state, err := s.GenerateState()
	if err != nil {
		return "", "", err
	}

	// 将state存储到缓存（15分钟有效期）
	stateKey := "oauth:state:" + state
	if err := s.cacheService.Set(ctx, stateKey, "1", 15*time.Minute); err != nil {
		s.logger.WithError(err).Error("Failed to cache OAuth state")
		return "", "", fmt.Errorf("failed to cache state: %w", err)
	}

	// 获取授权URL
	authURL := s.linuxDoClient.GetAuthorizationURL(state)

	s.logger.WithField("state", state).Debug("Generated OAuth authorization URL")
	return authURL, state, nil
}

// ValidateState 验证OAuth状态参数
func (s *AuthService) ValidateState(ctx context.Context, state string) error {
	if state == "" {
		return fmt.Errorf("state is required")
	}

	stateKey := "oauth:state:" + state
	exists, err := s.cacheService.Exists(ctx, stateKey)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check state existence")
		return fmt.Errorf("failed to validate state: %w", err)
	}

	if !exists {
		s.logger.WithField("state", state).Warn("Invalid or expired OAuth state")
		return fmt.Errorf("invalid or expired state")
	}

	// 删除已使用的state（防止重放攻击）
	_ = s.cacheService.Del(ctx, stateKey)

	return nil
}

// HandleCallback 处理OAuth回调
func (s *AuthService) HandleCallback(ctx context.Context, code string, state string) (*model.User, string, error) {
	// 验证state
	if err := s.ValidateState(ctx, state); err != nil {
		return nil, "", err
	}

	// 交换授权码获取访问令牌
	tokenResp, err := s.linuxDoClient.ExchangeCode(ctx, code)
	if err != nil {
		s.logger.WithError(err).Error("Failed to exchange authorization code")
		return nil, "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// 获取用户信息
	userInfo, err := s.linuxDoClient.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user info")
		return nil, "", fmt.Errorf("failed to get user info: %w", err)
	}

	// 将 Linux.do 用户ID转换为字符串
	linuxDoID := strconv.Itoa(userInfo.ID)

	// 查找或创建用户
	user, err := s.userRepo.GetByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user from database")
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// 用户不存在，创建新用户
		user = &model.User{
			LinuxDoID: linuxDoID,
			Username:  userInfo.Username,
			KyxUserID: 0, // 未绑定
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			s.logger.WithError(err).Error("Failed to create user")
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}
		s.logger.WithFields(logrus.Fields{
			"user_id":     user.ID,
			"linux_do_id": user.LinuxDoID,
			"username":    user.Username,
		}).Info("New user created")
	} else {
		// 更新用户名（可能已改名）
		if user.Username != userInfo.Username {
			user.Username = userInfo.Username
			if err := s.userRepo.Update(ctx, user); err != nil {
				s.logger.WithError(err).Warn("Failed to update username")
			}
		}
	}

	// 创建会话
	sessionID, err := s.CreateSession(ctx, user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create session")
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"linux_do_id": user.LinuxDoID,
		"username":    user.Username,
		"session_id":  sessionID,
	}).Info("User logged in successfully")

	return user, sessionID, nil
}

// CreateSession 创建会话
func (s *AuthService) CreateSession(ctx context.Context, user *model.User) (string, error) {
	// 生成会话ID
	sessionID, err := s.GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// 创建会话数据
	sessionData := model.JSONMap{
		"user_id":     user.ID,
		"linux_do_id": user.LinuxDoID,
		"username":    user.Username,
		"created_at":  time.Now().Unix(),
	}

	// 保存会话
	session := &model.Session{
		SessionID: sessionID,
		Data:      sessionData,
		ExpiresAt: time.Now().Add(s.sessionTimeout),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session, s.sessionTimeout); err != nil {
		s.logger.WithError(err).Error("Failed to save session")
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return sessionID, nil
}

// GetSession 获取会话
func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get session")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// ValidateSession 验证会话
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*model.User, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 从会话数据中提取用户信息
	linuxDoID, ok := session.Data["linux_do_id"].(string)
	if !ok {
		s.logger.WithField("session_id", sessionID).Error("Invalid session data: missing linux_do_id")
		return nil, fmt.Errorf("invalid session data")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("User not found for valid session")
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// RefreshSession 刷新会话
func (s *AuthService) RefreshSession(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// 更新过期时间
	newExpiry := time.Now().Add(s.sessionTimeout)
	session.ExpiresAt = newExpiry

	if err := s.sessionRepo.Update(ctx, sessionID, session.Data, s.sessionTimeout); err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to refresh session")
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	s.logger.WithField("session_id", sessionID).Debug("Session refreshed")
	return nil
}

// DeleteSession 删除会话（登出）
func (s *AuthService) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		s.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.WithField("session_id", sessionID).Info("Session deleted")
	return nil
}

// AdminLogin 管理员登录
func (s *AuthService) AdminLogin(ctx context.Context, password string) (string, error) {
	if s.adminPassword == "" {
		s.logger.Error("Admin password not configured")
		return "", fmt.Errorf("admin authentication not configured")
	}

	if password != s.adminPassword {
		s.logger.Warn("Invalid admin password attempt")
		return "", fmt.Errorf("invalid password")
	}

	// 生成JWT token
	token, err := s.GenerateAdminToken()
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate admin token")
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("Admin logged in successfully")
	return token, nil
}

// GenerateAdminToken 生成管理员JWT token
func (s *AuthService) GenerateAdminToken() (string, error) {
	if s.jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	// 创建token
	claims := jwt.MapClaims{
		"role": "admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateAdminToken 验证管理员JWT token
func (s *AuthService) ValidateAdminToken(tokenString string) error {
	if s.jwtSecret == "" {
		return fmt.Errorf("JWT secret not configured")
	}

	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		s.logger.WithError(err).Warn("Invalid admin token")
		return fmt.Errorf("invalid token: %w", err)
	}

	// 验证claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		s.logger.Warn("Invalid token claims")
		return fmt.Errorf("invalid token claims")
	}

	// 验证角色
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		s.logger.Warn("Invalid admin role in token")
		return fmt.Errorf("invalid role")
	}

	// 验证过期时间
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid expiration time")
	}

	if time.Now().Unix() > int64(exp) {
		return fmt.Errorf("token expired")
	}

	return nil
}

// IsAdmin 检查是否为管理员
func (s *AuthService) IsAdmin(tokenString string) bool {
	return s.ValidateAdminToken(tokenString) == nil
}

// CleanExpiredSessions 清理过期会话
func (s *AuthService) CleanExpiredSessions(ctx context.Context) (int64, error) {
	count, err := s.sessionRepo.CleanExpired(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to clean expired sessions")
		return 0, err
	}

	if count > 0 {
		s.logger.WithField("count", count).Info("Expired sessions cleaned")
	}

	return count, nil
}

// GetUserFromSession 从会话ID获取用户信息
func (s *AuthService) GetUserFromSession(ctx context.Context, sessionID string) (*model.User, error) {
	return s.ValidateSession(ctx, sessionID)
}

// UpdateSessionData 更新会话数据
func (s *AuthService) UpdateSessionData(ctx context.Context, sessionID string, key string, value interface{}) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Data[key] = value
	return s.sessionRepo.Update(ctx, sessionID, session.Data, s.sessionTimeout)
}

// GetSessionData 获取会话数据
func (s *AuthService) GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error) {
	return s.sessionRepo.GetSessionData(ctx, sessionID, key)
}
