package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo        *repository.UserRepository
	claimRepo       *repository.ClaimRepository
	donateRepo      *repository.DonateRepository
	adminConfigRepo *repository.AdminConfigRepository
	kyxClient       *KyxClient
	linuxDoClient   *LinuxDoClient
	cacheService    *CacheService
	logger          *logrus.Logger
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo *repository.UserRepository,
	claimRepo *repository.ClaimRepository,
	donateRepo *repository.DonateRepository,
	adminConfigRepo *repository.AdminConfigRepository,
	kyxClient *KyxClient,
	linuxDoClient *LinuxDoClient,
	cacheService *CacheService,
	logger *logrus.Logger,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		claimRepo:       claimRepo,
		donateRepo:      donateRepo,
		adminConfigRepo: adminConfigRepo,
		kyxClient:       kyxClient,
		linuxDoClient:   linuxDoClient,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, linuxDoID string) (*model.User, error) {
	// 先从缓存获取
	cacheKey := s.cacheService.UserKey(linuxDoID)
	var user model.User
	err := s.cacheService.GetJSON(ctx, cacheKey, &user)
	if err == nil && user.ID > 0 {
		s.logger.WithField("linux_do_id", linuxDoID).Debug("User retrieved from cache")
		return &user, nil
	}

	// 从数据库获取
	dbUser, err := s.userRepo.GetByLinuxDoID(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if dbUser == nil {
		return nil, nil
	}

	// 缓存用户信息（1小时）
	_ = s.cacheService.SetJSON(ctx, cacheKey, dbUser, time.Hour)

	return dbUser, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByUsername 根据用户名获取用户信息
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// BindAccount 绑定公益站账号
func (s *UserService) BindAccount(ctx context.Context, linuxDoID, username string) (*model.BindAccountResponse, error) {
	// 获取用户
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		s.logger.WithField("linux_do_id", linuxDoID).Error("User not found for binding")
		return nil, fmt.Errorf("user not found")
	}

	// 检查是否已绑定
	if user.KyxUserID > 0 {
		// 已绑定，验证用户名是否匹配
		if user.Username != username {
			s.logger.WithFields(logrus.Fields{
				"linux_do_id":    linuxDoID,
				"current_user":   user.Username,
				"requested_user": username,
			}).Warn("Username mismatch for bound account")
			return nil, fmt.Errorf("account already bound to different username")
		}

		// 获取最新的额度信息
		_, err := s.kyxClient.GetUserByID(ctx, user.KyxUserID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get Kyx user info for bound account")
			// 即使获取失败，也返回已绑定的用户信息
			return &model.BindAccountResponse{
				User:        user,
				IsFirstBind: false,
			}, nil
		}

		s.logger.WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"kyx_user_id": user.KyxUserID,
			"username":    username,
		}).Info("Account already bound")

		return &model.BindAccountResponse{
			User:        user,
			IsFirstBind: false,
		}, nil
	}

	// 未绑定，搜索公益站用户
	kyxUser, err := s.kyxClient.SearchUser(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"username":    username,
		}).Error("Failed to search user in Kyx API")
		return nil, fmt.Errorf("failed to search user in Kyx: %w", err)
	}

	if kyxUser == nil {
		s.logger.WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"username":    username,
		}).Warn("User not found in Kyx API")
		return nil, fmt.Errorf("user not found in Kyx, please register first")
	}

	// 验证用户名是否匹配
	if kyxUser.Username != username {
		s.logger.WithFields(logrus.Fields{
			"linux_do_id":    linuxDoID,
			"kyx_username":   kyxUser.Username,
			"requested_user": username,
		}).Warn("Username mismatch")
		return nil, fmt.Errorf("username mismatch: expected %s, got %s", kyxUser.Username, username)
	}

	// 更新用户绑定信息
	user.KyxUserID = kyxUser.ID
	user.Username = kyxUser.Username
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to update user binding")
		return nil, fmt.Errorf("failed to update user binding: %w", err)
	}

	// 清除用户缓存
	_ = s.cacheService.ClearUserCache(ctx, linuxDoID)

	// 首次绑定奖励
	var bonus int64 = 0
	claimQuota, err := s.adminConfigRepo.GetClaimQuota(ctx)
	if err == nil && claimQuota > 0 {
		bonus = claimQuota
		// 添加首次绑定奖励
		if err := s.kyxClient.AddQuota(ctx, kyxUser.ID, bonus); err != nil {
			s.logger.WithError(err).Warn("Failed to add first bind bonus")
		} else {
			s.logger.WithFields(logrus.Fields{
				"linux_do_id": linuxDoID,
				"kyx_user_id": kyxUser.ID,
				"bonus":       bonus,
			}).Info("First bind bonus added")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"kyx_user_id": kyxUser.ID,
		"username":    username,
		"bonus":       bonus,
	}).Info("Account bound successfully")

	return &model.BindAccountResponse{
		User:        user,
		Bonus:       bonus,
		BonusCNY:    model.QuotaToDollar(bonus),
		IsFirstBind: true,
	}, nil
}

// GetQuotaInfo 获取用户额度信息
func (s *UserService) GetQuotaInfo(ctx context.Context, linuxDoID string) (*model.QuotaInfo, error) {
	// 先从缓存获取
	cachedQuota, err := s.cacheService.GetUserQuota(ctx, linuxDoID)
	if err == nil && cachedQuota != nil {
		s.logger.WithField("linux_do_id", linuxDoID).Debug("Quota info retrieved from cache")
		return cachedQuota, nil
	}

	// 获取用户信息
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.KyxUserID == 0 {
		return nil, fmt.Errorf("account not bound")
	}

	// 从公益站API获取额度信息
	kyxUser, err := s.kyxClient.GetUserByID(ctx, user.KyxUserID)
	if err != nil {
		s.logger.WithError(err).WithField("kyx_user_id", user.KyxUserID).Error("Failed to get Kyx user info")
		return nil, fmt.Errorf("failed to get quota info: %w", err)
	}

	// 检查今天是否已领取
	claimedToday, err := s.claimRepo.HasClaimedToday(ctx, linuxDoID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to check claim status")
		claimedToday = false
	}

	// 构建额度信息
	quotaInfo := &model.QuotaInfo{
		Username:     kyxUser.Username,
		DisplayName:  kyxUser.DisplayName,
		LinuxDoID:    kyxUser.LinuxDoID,
		Name:         kyxUser.DisplayName,
		Quota:        kyxUser.Quota,
		UsedQuota:    kyxUser.UsedQuota,
		Total:        kyxUser.Quota + kyxUser.UsedQuota,
		CanClaim:     !claimedToday,
		ClaimedToday: claimedToday,
	}

	// 缓存额度信息（5分钟）
	_ = s.cacheService.SetUserQuota(ctx, linuxDoID, quotaInfo, 5*time.Minute)

	return quotaInfo, nil
}

// GetStatistics 获取用户统计信息
func (s *UserService) GetStatistics(ctx context.Context, linuxDoID string) (*model.UserStatistics, error) {
	return s.userRepo.GetStatistics(ctx, linuxDoID)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
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
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count users")
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

// GetAllStatistics 获取所有用户的统计信息
func (s *UserService) GetAllStatistics(ctx context.Context) ([]*model.UserStatistics, error) {
	return s.userRepo.GetAllStatistics(ctx)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, linuxDoID string) error {
	// 删除用户的所有记录
	if err := s.claimRepo.DeleteByLinuxDoID(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete claim records")
	}

	if err := s.donateRepo.DeleteByLinuxDoID(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to delete donate records")
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 清除缓存
	_ = s.cacheService.ClearUserCache(ctx, linuxDoID)

	s.logger.WithField("linux_do_id", linuxDoID).Info("User deleted successfully")
	return nil
}

// UpdateUsername 更新用户名
func (s *UserService) UpdateUsername(ctx context.Context, linuxDoID, newUsername string) error {
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.Username = newUsername
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to update username")
		return fmt.Errorf("failed to update username: %w", err)
	}

	// 清除缓存
	_ = s.cacheService.ClearUserCache(ctx, linuxDoID)

	s.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"username":    newUsername,
	}).Info("Username updated")

	return nil
}

// IsAccountBound 检查账号是否已绑定
func (s *UserService) IsAccountBound(ctx context.Context, linuxDoID string) (bool, error) {
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.KyxUserID > 0, nil
}

// GetBoundKyxUserID 获取绑定的公益站用户ID
func (s *UserService) GetBoundKyxUserID(ctx context.Context, linuxDoID string) (int, error) {
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return 0, err
	}

	if user == nil {
		return 0, fmt.Errorf("user not found")
	}

	if user.KyxUserID == 0 {
		return 0, fmt.Errorf("account not bound")
	}

	return user.KyxUserID, nil
}

// RefreshQuotaCache 刷新用户额度缓存
func (s *UserService) RefreshQuotaCache(ctx context.Context, linuxDoID string) error {
	// 清除现有缓存
	if err := s.cacheService.ClearUserQuota(ctx, linuxDoID); err != nil {
		s.logger.WithError(err).Warn("Failed to clear quota cache")
	}

	// 重新获取并缓存
	_, err := s.GetQuotaInfo(ctx, linuxDoID)
	return err
}

// GetUserCount 获取用户总数
func (s *UserService) GetUserCount(ctx context.Context) (int64, error) {
	return s.userRepo.Count(ctx)
}

// GetBoundUserCount 获取已绑定用户数
func (s *UserService) GetBoundUserCount(ctx context.Context) (int64, error) {
	// 这里需要在 repository 中添加相应方法，暂时用查询所有用户然后过滤的方式
	// TODO: 优化为数据库查询
	users, err := s.userRepo.List(ctx, 10000, 0)
	if err != nil {
		return 0, err
	}

	count := int64(0)
	for _, user := range users {
		if user.KyxUserID > 0 {
			count++
		}
	}

	return count, nil
}

// SyncUserInfo 同步用户信息（从公益站）
func (s *UserService) SyncUserInfo(ctx context.Context, linuxDoID string) error {
	user, err := s.GetUser(ctx, linuxDoID)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.KyxUserID == 0 {
		return fmt.Errorf("account not bound")
	}

	// 从公益站获取最新信息
	kyxUser, err := s.kyxClient.GetUserByID(ctx, user.KyxUserID)
	if err != nil {
		return fmt.Errorf("failed to get Kyx user info: %w", err)
	}

	// 更新用户名
	if kyxUser.Username != user.Username {
		user.Username = kyxUser.Username
		if err := s.userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	// 清除缓存
	_ = s.cacheService.ClearUserCache(ctx, linuxDoID)

	s.logger.WithFields(logrus.Fields{
		"linux_do_id": linuxDoID,
		"kyx_user_id": user.KyxUserID,
		"username":    user.Username,
	}).Info("User info synced")

	return nil
}
