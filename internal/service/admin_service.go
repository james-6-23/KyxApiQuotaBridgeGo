package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
)

// AdminService 管理员服务
type AdminService struct {
	adminConfigRepo *repository.AdminConfigRepository
	userRepo        *repository.UserRepository
	claimRepo       *repository.ClaimRepository
	donateRepo      *repository.DonateRepository
	keyRepo         *repository.KeyRepository
	sessionRepo     *repository.SessionRepository
	kyxClient       *KyxClient
	cacheService    *CacheService
	logger          *logrus.Logger
}

// NewAdminService 创建管理员服务
func NewAdminService(
	adminConfigRepo *repository.AdminConfigRepository,
	userRepo *repository.UserRepository,
	claimRepo *repository.ClaimRepository,
	donateRepo *repository.DonateRepository,
	keyRepo *repository.KeyRepository,
	sessionRepo *repository.SessionRepository,
	kyxClient *KyxClient,
	cacheService *CacheService,
	logger *logrus.Logger,
) *AdminService {
	return &AdminService{
		adminConfigRepo: adminConfigRepo,
		userRepo:        userRepo,
		claimRepo:       claimRepo,
		donateRepo:      donateRepo,
		keyRepo:         keyRepo,
		sessionRepo:     sessionRepo,
		kyxClient:       kyxClient,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// GetConfig 获取管理员配置
func (s *AdminService) GetConfig(ctx context.Context) (*model.AdminConfigResponse, error) {
	config, err := s.adminConfigRepo.Get(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get admin config")
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if config == nil {
		// 返回默认配置
		return &model.AdminConfigResponse{
			ClaimQuota:                  500000,
			SessionConfigured:           false,
			KeysAPIURL:                  "",
			KeysAuthorizationConfigured: false,
			GroupID:                     1,
			UpdatedAt:                   0,
		}, nil
	}

	response := &model.AdminConfigResponse{
		ClaimQuota:                  config.ClaimQuota,
		SessionConfigured:           config.Session.Valid && config.Session.String != "",
		KeysAPIURL:                  config.KeysAPIURL.String,
		KeysAuthorizationConfigured: config.KeysAuthorization.Valid && config.KeysAuthorization.String != "",
		GroupID:                     config.GroupID,
		UpdatedAt:                   config.UpdatedAt.Unix(),
	}

	return response, nil
}

// UpdateConfig 更新管理员配置
func (s *AdminService) UpdateConfig(ctx context.Context, req *model.UpdateConfigRequest) error {
	// 获取当前配置
	currentConfig, err := s.adminConfigRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// 构建更新map
	updates := make(map[string]interface{})

	if req.ClaimQuota != nil {
		if *req.ClaimQuota < 0 {
			return fmt.Errorf("claim quota cannot be negative")
		}
		updates["claim_quota"] = *req.ClaimQuota
		s.logger.WithField("claim_quota", *req.ClaimQuota).Info("Updating claim quota")
	}

	if req.Session != nil {
		updates["session"] = *req.Session
		// 更新 KyxClient 的 session
		s.kyxClient.UpdateSession(*req.Session)
		s.logger.Info("Updating Kyx session")
	}

	if req.NewAPIUser != nil {
		updates["new_api_user"] = *req.NewAPIUser
	}

	if req.KeysAPIURL != nil {
		updates["keys_api_url"] = *req.KeysAPIURL
		s.logger.WithField("keys_api_url", *req.KeysAPIURL).Info("Updating Keys API URL")
	}

	if req.KeysAuthorization != nil {
		updates["keys_authorization"] = *req.KeysAuthorization
		s.logger.Info("Updating Keys API authorization")
	}

	if req.GroupID != nil {
		if *req.GroupID < 0 {
			return fmt.Errorf("group ID cannot be negative")
		}
		updates["group_id"] = *req.GroupID
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	// 如果配置不存在，创建新配置
	if currentConfig == nil {
		newConfig := &model.AdminConfig{
			ClaimQuota: 500000,
			GroupID:    1,
		}

		if val, ok := updates["claim_quota"].(int64); ok {
			newConfig.ClaimQuota = val
			s.logger.WithField("claim_quota", val).Debug("Setting claim_quota from int64")
		} else {
			s.logger.WithFields(logrus.Fields{
				"value": updates["claim_quota"],
				"type":  fmt.Sprintf("%T", updates["claim_quota"]),
			}).Debug("claim_quota type mismatch or not provided")
		}
		
		if val, ok := updates["session"].(string); ok {
			newConfig.Session = sql.NullString{String: val, Valid: val != ""}
		}
		if val, ok := updates["new_api_user"].(string); ok {
			newConfig.NewAPIUser = sql.NullString{String: val, Valid: val != ""}
		}
		if val, ok := updates["keys_api_url"].(string); ok {
			newConfig.KeysAPIURL = sql.NullString{String: val, Valid: val != ""}
		}
		if val, ok := updates["keys_authorization"].(string); ok {
			newConfig.KeysAuthorization = sql.NullString{String: val, Valid: val != ""}
		}
		if val, ok := updates["group_id"].(int); ok {
			newConfig.GroupID = val
			s.logger.WithField("group_id", val).Debug("Setting group_id from int")
		} else {
			s.logger.WithFields(logrus.Fields{
				"value": updates["group_id"],
				"type":  fmt.Sprintf("%T", updates["group_id"]),
			}).Debug("group_id type mismatch or not provided")
		}

		s.logger.WithFields(logrus.Fields{
			"new_config": newConfig,
			"updates":    updates,
		}).Info("Creating new admin config")

		if err := s.adminConfigRepo.Create(ctx, newConfig); err != nil {
			s.logger.WithError(err).Error("Failed to create admin config")
			return fmt.Errorf("failed to create config: %w", err)
		}

		s.logger.Info("Admin config created successfully")
		return nil
	}

	// 更新现有配置
	if err := s.adminConfigRepo.UpdatePartial(ctx, updates); err != nil {
		s.logger.WithError(err).Error("Failed to update admin config")
		return fmt.Errorf("failed to update config: %w", err)
	}

	s.logger.WithField("updates", updates).Info("Admin config updated successfully")
	return nil
}

// GetSystemStats 获取系统统计信息
func (s *AdminService) GetSystemStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 用户统计
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total users count")
	} else {
		stats["total_users"] = totalUsers
	}

	// 领取统计
	totalClaims, err := s.claimRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total claims count")
	} else {
		stats["total_claims"] = totalClaims
	}

	todayClaimCount, todayClaimQuota, err := s.claimRepo.GetTodayStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get today's claim stats")
	} else {
		stats["today_claims"] = todayClaimCount
		stats["today_claim_quota"] = todayClaimQuota
		stats["today_claim_quota_usd"] = model.QuotaToDollar(todayClaimQuota)
	}

	// 投喂统计
	totalDonates, err := s.donateRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total donates count")
	} else {
		stats["total_donates"] = totalDonates
	}

	todayDonateCount, todayDonateKeys, todayDonateQuota, err := s.donateRepo.GetTodayStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get today's donate stats")
	} else {
		stats["today_donates"] = todayDonateCount
		stats["today_donate_keys"] = todayDonateKeys
		stats["today_donate_quota"] = todayDonateQuota
		stats["today_donate_quota_usd"] = model.QuotaToDollar(todayDonateQuota)
	}

	// Key统计
	totalKeys, err := s.keyRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total keys count")
	} else {
		stats["total_keys"] = totalKeys
	}

	todayKeys, err := s.keyRepo.GetTodayCount(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get today's keys count")
	} else {
		stats["today_keys"] = todayKeys
	}

	// 唯一用户数
	uniqueUsers, err := s.keyRepo.GetUniqueUsers(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get unique users count")
	} else {
		stats["unique_donate_users"] = uniqueUsers
	}

	// 缓存统计
	cacheStats, err := s.cacheService.Stats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get cache stats")
	} else {
		stats["cache"] = cacheStats
	}

	// 系统时间
	stats["server_time"] = time.Now().Unix()

	return stats, nil
}

// GetDashboardStats 获取仪表板统计（更详细）
func (s *AdminService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 基础统计
	systemStats, err := s.GetSystemStats(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range systemStats {
		stats[k] = v
	}

	// 本周统计
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	weekClaimCount, weekClaimQuota, err := s.claimRepo.GetDateRangeStats(
		ctx,
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"),
	)
	if err == nil {
		stats["week_claims"] = weekClaimCount
		stats["week_claim_quota"] = weekClaimQuota
		stats["week_claim_quota_usd"] = model.QuotaToDollar(weekClaimQuota)
	}

	weekDonateCount, weekDonateKeys, weekDonateQuota, err := s.donateRepo.GetDateRangeStats(
		ctx,
		weekStart,
		weekEnd,
	)
	if err == nil {
		stats["week_donates"] = weekDonateCount
		stats["week_donate_keys"] = weekDonateKeys
		stats["week_donate_quota"] = weekDonateQuota
		stats["week_donate_quota_usd"] = model.QuotaToDollar(weekDonateQuota)
	}

	// 本月统计
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	monthClaimCount, monthClaimQuota, err := s.claimRepo.GetDateRangeStats(
		ctx,
		monthStart.Format("2006-01-02"),
		monthEnd.Format("2006-01-02"),
	)
	if err == nil {
		stats["month_claims"] = monthClaimCount
		stats["month_claim_quota"] = monthClaimQuota
		stats["month_claim_quota_usd"] = model.QuotaToDollar(monthClaimQuota)
	}

	monthDonateCount, monthDonateKeys, monthDonateQuota, err := s.donateRepo.GetDateRangeStats(
		ctx,
		monthStart,
		monthEnd,
	)
	if err == nil {
		stats["month_donates"] = monthDonateCount
		stats["month_donate_keys"] = monthDonateKeys
		stats["month_donate_quota"] = monthDonateQuota
		stats["month_donate_quota_usd"] = model.QuotaToDollar(monthDonateQuota)
	}

	return stats, nil
}

// ListUsers 获取用户列表
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) (*model.PaginationResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, err := s.userRepo.List(ctx, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count users")
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &model.PaginationResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
		Data:       users,
	}, nil
}

// ListAllStatistics 获取所有用户统计
func (s *AdminService) ListAllStatistics(ctx context.Context) ([]*model.UserStatistics, error) {
	return s.userRepo.GetAllStatistics(ctx)
}

// DeleteUser 删除用户及其所有数据
func (s *AdminService) DeleteUser(ctx context.Context, linuxDoID string) error {
	// 删除领取记录
	if err := s.claimRepo.DeleteByLinuxDoID(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete claim records")
	}

	// 删除投喂记录
	if err := s.donateRepo.DeleteByLinuxDoID(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete donate records")
	}

	// 删除已使用的Key记录
	if err := s.keyRepo.DeleteByLinuxDoID(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete key records")
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 清除缓存
	_ = s.cacheService.ClearUserCache(ctx, linuxDoID)

	s.logger.WithField("linux_do_id", linuxDoID).Info("User and all related data deleted")
	return nil
}

// CleanExpiredSessions 清理过期会话
func (s *AdminService) CleanExpiredSessions(ctx context.Context) (int64, error) {
	count, err := s.sessionRepo.CleanExpired(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to clean expired sessions")
		return 0, fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	if count > 0 {
		s.logger.WithField("count", count).Info("Expired sessions cleaned")
	}

	return count, nil
}

// CleanOldKeys 清理旧的Key记录
func (s *AdminService) CleanOldKeys(ctx context.Context, daysOld int) (int64, error) {
	if daysOld < 1 {
		daysOld = 90 // 默认90天
	}

	olderThan := time.Now().AddDate(0, 0, -daysOld)
	count, err := s.keyRepo.DeleteOlderThan(ctx, olderThan)
	if err != nil {
		s.logger.WithError(err).Error("Failed to clean old keys")
		return 0, fmt.Errorf("failed to clean old keys: %w", err)
	}

	if count > 0 {
		s.logger.WithFields(logrus.Fields{
			"count":    count,
			"days_old": daysOld,
		}).Info("Old keys cleaned")
	}

	return count, nil
}

// ClearCache 清除指定类型的缓存
func (s *AdminService) ClearCache(ctx context.Context, cacheType string) error {
	switch cacheType {
	case "all":
		if err := s.cacheService.ClearAllUserCaches(ctx); err != nil {
			return fmt.Errorf("failed to clear all caches: %w", err)
		}
		_ = s.adminConfigRepo.ClearCache(ctx)
		s.logger.Info("All caches cleared")

	case "user":
		if err := s.cacheService.ClearAllUserCaches(ctx); err != nil {
			return fmt.Errorf("failed to clear user caches: %w", err)
		}
		s.logger.Info("User caches cleared")

	case "config":
		if err := s.adminConfigRepo.ClearCache(ctx); err != nil {
			return fmt.Errorf("failed to clear config cache: %w", err)
		}
		s.logger.Info("Config cache cleared")

	default:
		return fmt.Errorf("invalid cache type: %s", cacheType)
	}

	return nil
}

// ValidateKyxSession 验证公益站Session是否有效
func (s *AdminService) ValidateKyxSession(ctx context.Context) error {
	return s.kyxClient.ValidateSession(ctx)
}

// TestKyxConnection 测试公益站连接
func (s *AdminService) TestKyxConnection(ctx context.Context) error {
	return s.kyxClient.Ping(ctx)
}

// GetRecentActivity 获取最近的活动
func (s *AdminService) GetRecentActivity(ctx context.Context, limit int) (map[string]interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	activity := make(map[string]interface{})

	// 最近的领取记录
	recentClaims, err := s.claimRepo.List(ctx, limit, 0)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent claims")
	} else {
		activity["recent_claims"] = recentClaims
	}

	// 最近的投喂记录
	recentDonates, err := s.donateRepo.List(ctx, limit, 0)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent donates")
	} else {
		activity["recent_donates"] = recentDonates
	}

	// 最近注册的用户
	recentUsers, err := s.userRepo.List(ctx, limit, 0)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get recent users")
	} else {
		activity["recent_users"] = recentUsers
	}

	return activity, nil
}

// ExportData 导出数据（用于备份或分析）
func (s *AdminService) ExportData(ctx context.Context, dataType string) (interface{}, error) {
	switch dataType {
	case "users":
		return s.userRepo.List(ctx, 10000, 0)
	case "claims":
		return s.claimRepo.List(ctx, 10000, 0)
	case "donates":
		return s.donateRepo.List(ctx, 10000, 0)
	case "statistics":
		return s.userRepo.GetAllStatistics(ctx)
	default:
		return nil, fmt.Errorf("invalid data type: %s", dataType)
	}
}

// GetHealthStatus 获取系统健康状态
func (s *AdminService) GetHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	health := make(map[string]interface{})
	health["status"] = "healthy"
	health["timestamp"] = time.Now().Unix()

	// 检查数据库
	if _, err := s.userRepo.Count(ctx); err != nil {
		health["database"] = "unhealthy"
		health["database_error"] = err.Error()
		health["status"] = "unhealthy"
	} else {
		health["database"] = "healthy"
	}

	// 检查Redis
	if err := s.cacheService.Ping(ctx); err != nil {
		health["redis"] = "unhealthy"
		health["redis_error"] = err.Error()
		health["status"] = "unhealthy"
	} else {
		health["redis"] = "healthy"
	}

	// 检查公益站连接
	if err := s.kyxClient.Ping(ctx); err != nil {
		health["kyx_api"] = "unhealthy"
		health["kyx_api_error"] = err.Error()
		// 不影响整体状态
	} else {
		health["kyx_api"] = "healthy"
	}

	return health, nil
}

// InitializeDefaultConfig 初始化默认配置
func (s *AdminService) InitializeDefaultConfig(ctx context.Context) error {
	return s.adminConfigRepo.InitializeDefault(ctx)
}
