package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
)

// QuotaService 额度服务
type QuotaService struct {
	claimRepo       *repository.ClaimRepository
	userRepo        *repository.UserRepository
	adminConfigRepo *repository.AdminConfigRepository
	kyxClient       *KyxClient
	cacheService    *CacheService
	logger          *logrus.Logger
}

// NewQuotaService 创建额度服务
func NewQuotaService(
	claimRepo *repository.ClaimRepository,
	userRepo *repository.UserRepository,
	adminConfigRepo *repository.AdminConfigRepository,
	kyxClient *KyxClient,
	cacheService *CacheService,
	logger *logrus.Logger,
) *QuotaService {
	return &QuotaService{
		claimRepo:       claimRepo,
		userRepo:        userRepo,
		adminConfigRepo: adminConfigRepo,
		kyxClient:       kyxClient,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// ClaimQuota 领取每日额度
func (s *QuotaService) ClaimQuota(ctx context.Context, linuxDoID string) (*model.ClaimRecord, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("User not found for claim")
		return nil, fmt.Errorf("user not found")
	}

	// 检查是否已绑定账号
	if user.KyxUserID == 0 {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("Attempt to claim without bound account")
		return nil, fmt.Errorf("account not bound, please bind first")
	}

	// 检查今天是否已领取
	claimed, err := s.CanClaim(ctx, linuxDoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check claim status: %w", err)
	}

	if !claimed {
		s.logger.WithField("linux_do_id", linuxDoID).Warn("Already claimed today")
		return nil, fmt.Errorf("already claimed today, please try again tomorrow")
	}

	// 获取领取额度配置
	claimQuota, err := s.adminConfigRepo.GetClaimQuota(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get claim quota config")
		return nil, fmt.Errorf("failed to get claim quota: %w", err)
	}

	if claimQuota <= 0 {
		s.logger.Warn("Claim quota not configured")
		return nil, fmt.Errorf("claim quota not configured")
	}

	// 调用公益站API添加额度
	err = s.kyxClient.AddQuota(ctx, user.KyxUserID, claimQuota)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"kyx_user_id": user.KyxUserID,
			"quota":       claimQuota,
		}).Error("Failed to add quota via Kyx API")
		return nil, fmt.Errorf("failed to add quota: %w", err)
	}

	// 创建领取记录
	record := &model.ClaimRecord{
		LinuxDoID:  linuxDoID,
		Username:   user.Username,
		QuotaAdded: claimQuota,
	}

	if err := s.claimRepo.Create(ctx, record); err != nil {
		s.logger.WithError(err).Error("Failed to create claim record")
		// 即使记录创建失败，额度已经添加，不应该返回错误
		// 只记录警告
		s.logger.Warn("Quota added but failed to save record")
	}

	// 标记今天已领取（缓存）
	if err := s.cacheService.MarkClaimedToday(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to mark claimed in cache")
	}

	// 清除用户额度缓存
	_ = s.cacheService.ClearUserQuota(ctx, linuxDoID)

	s.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"username":    user.Username,
		"kyx_user_id": user.KyxUserID,
		"quota":       claimQuota,
		"record_id":   record.ID,
	}).Info("Quota claimed successfully")

	return record, nil
}

// CanClaim 检查用户是否可以领取
func (s *QuotaService) CanClaim(ctx context.Context, linuxDoID string) (bool, error) {
	// 先检查缓存
	claimed, err := s.cacheService.HasClaimedToday(ctx, linuxDoID)
	if err == nil {
		if claimed {
			return false, nil
		}
	}

	// 检查数据库
	claimed, err = s.claimRepo.HasClaimedToday(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to check claim status")
		return false, fmt.Errorf("failed to check claim status: %w", err)
	}

	return !claimed, nil
}

// GetClaimHistory 获取用户的领取历史
func (s *QuotaService) GetClaimHistory(ctx context.Context, linuxDoID string, page, pageSize int) ([]*model.ClaimRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.claimRepo.GetByLinuxDoID(ctx, linuxDoID, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get claim history")
		return nil, 0, fmt.Errorf("failed to get claim history: %w", err)
	}

	total, err := s.claimRepo.CountByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count claim records")
		return nil, 0, fmt.Errorf("failed to count claim records: %w", err)
	}

	return records, total, nil
}

// GetUserClaimStats 获取用户的领取统计
func (s *QuotaService) GetUserClaimStats(ctx context.Context, linuxDoID string) (totalClaims int64, totalQuota int64, err error) {
	totalClaims, err = s.claimRepo.CountByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count user claims")
		return 0, 0, fmt.Errorf("failed to count claims: %w", err)
	}

	totalQuota, err = s.claimRepo.GetTotalQuota(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total claimed quota")
		return 0, 0, fmt.Errorf("failed to get total quota: %w", err)
	}

	return totalClaims, totalQuota, nil
}

// GetTodayStats 获取今日领取统计
func (s *QuotaService) GetTodayStats(ctx context.Context) (count int64, totalQuota int64, err error) {
	return s.claimRepo.GetTodayStats(ctx)
}

// GetDateRangeStats 获取日期范围内的领取统计
func (s *QuotaService) GetDateRangeStats(ctx context.Context, startDate, endDate string) (count int64, totalQuota int64, err error) {
	return s.claimRepo.GetDateRangeStats(ctx, startDate, endDate)
}

// ListAllClaims 获取所有领取记录（管理员）
func (s *QuotaService) ListAllClaims(ctx context.Context, page, pageSize int) ([]*model.ClaimRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.claimRepo.List(ctx, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list all claims")
		return nil, 0, fmt.Errorf("failed to list claims: %w", err)
	}

	total, err := s.claimRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count all claims")
		return nil, 0, fmt.Errorf("failed to count claims: %w", err)
	}

	return records, total, nil
}

// GetClaimsByDate 获取指定日期的领取记录
func (s *QuotaService) GetClaimsByDate(ctx context.Context, date string, page, pageSize int) ([]*model.ClaimRecord, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	records, err := s.claimRepo.GetByDate(ctx, date, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).WithField("date", date).Error("Failed to get claims by date")
		return nil, fmt.Errorf("failed to get claims by date: %w", err)
	}

	return records, nil
}

// DeleteClaimRecord 删除领取记录（管理员）
func (s *QuotaService) DeleteClaimRecord(ctx context.Context, recordID int) error {
	if err := s.claimRepo.Delete(ctx, recordID); err != nil {
		s.logger.WithError(err).WithField("record_id", recordID).Error("Failed to delete claim record")
		return fmt.Errorf("failed to delete claim record: %w", err)
	}

	s.logger.WithField("record_id", recordID).Info("Claim record deleted")
	return nil
}

// GetClaimQuotaConfig 获取领取额度配置
func (s *QuotaService) GetClaimQuotaConfig(ctx context.Context) (int64, error) {
	return s.adminConfigRepo.GetClaimQuota(ctx)
}

// UpdateClaimQuotaConfig 更新领取额度配置（管理员）
func (s *QuotaService) UpdateClaimQuotaConfig(ctx context.Context, quota int64) error {
	if quota < 0 {
		return fmt.Errorf("quota cannot be negative")
	}

	if err := s.adminConfigRepo.UpdateClaimQuota(ctx, quota); err != nil {
		s.logger.WithError(err).WithField("quota", quota).Error("Failed to update claim quota config")
		return fmt.Errorf("failed to update claim quota: %w", err)
	}

	s.logger.WithField("quota", quota).Info("Claim quota config updated")
	return nil
}

// GetAllClaimStats 获取所有领取统计
func (s *QuotaService) GetAllClaimStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 今日统计
	todayCount, todayQuota, err := s.GetTodayStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get today stats")
	} else {
		stats["today_count"] = todayCount
		stats["today_quota"] = todayQuota
		stats["today_quota_usd"] = model.QuotaToDollar(todayQuota)
	}

	// 总统计
	totalCount, err := s.claimRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total count")
	} else {
		stats["total_count"] = totalCount
	}

	// 本周统计
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)
	weekCount, weekQuota, err := s.GetDateRangeStats(
		ctx,
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"),
	)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get week stats")
	} else {
		stats["week_count"] = weekCount
		stats["week_quota"] = weekQuota
		stats["week_quota_usd"] = model.QuotaToDollar(weekQuota)
	}

	// 本月统计
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthCount, monthQuota, err := s.GetDateRangeStats(
		ctx,
		monthStart.Format("2006-01-02"),
		monthEnd.Format("2006-01-02"),
	)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get month stats")
	} else {
		stats["month_count"] = monthCount
		stats["month_quota"] = monthQuota
		stats["month_quota_usd"] = model.QuotaToDollar(monthQuota)
	}

	return stats, nil
}

// ResetDailyClaim 重置每日领取状态（用于测试或手动重置）
func (s *QuotaService) ResetDailyClaim(ctx context.Context, linuxDoID string) error {
	// 清除缓存中的今日领取标记
	key := s.cacheService.ClaimTodayKey(linuxDoID)
	if err := s.cacheService.Del(ctx, key); err != nil {
		s.logger.WithError(err).Warn("Failed to clear claim cache")
	}

	s.logger.WithField("linux_do_id", linuxDoID).Info("Daily claim status reset")
	return nil
}

// GetRecentClaims 获取最近的领取记录
func (s *QuotaService) GetRecentClaims(ctx context.Context, limit int) ([]*model.ClaimRecord, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	records, err := s.claimRepo.List(ctx, limit, 0)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get recent claims")
		return nil, fmt.Errorf("failed to get recent claims: %w", err)
	}

	return records, nil
}
