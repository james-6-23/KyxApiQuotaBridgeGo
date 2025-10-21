package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/cache"
)

// CacheService 缓存服务
type CacheService struct {
	cache  *cache.Redis
	logger *logrus.Logger
}

// NewCacheService 创建缓存服务
func NewCacheService(cache *cache.Redis, logger *logrus.Logger) *CacheService {
	return &CacheService{
		cache:  cache,
		logger: logger,
	}
}

// 缓存键生成方法

// UserKey 生成用户缓存键
func (s *CacheService) UserKey(linuxDoID string) string {
	return model.CacheKeyUser + linuxDoID
}

// UserQuotaKey 生成用户额度缓存键
func (s *CacheService) UserQuotaKey(linuxDoID string) string {
	return model.CacheKeyUserQuota + linuxDoID
}

// ClaimTodayKey 生成今日领取缓存键
func (s *CacheService) ClaimTodayKey(linuxDoID string) string {
	today := time.Now().Format("2006-01-02")
	return model.CacheKeyClaimToday + linuxDoID + ":" + today
}

// DonateCountKey 生成投喂计数缓存键
func (s *CacheService) DonateCountKey(linuxDoID string) string {
	today := time.Now().Format("2006-01-02")
	return model.CacheKeyDonateCount + linuxDoID + ":" + today
}

// SessionKey 生成会话缓存键
func (s *CacheService) SessionKey(sessionID string) string {
	return model.CacheKeySession + sessionID
}

// RateLimitLoginKey 生成登录限流缓存键
func (s *CacheService) RateLimitLoginKey(ip string) string {
	return model.RateLimitLogin + ip
}

// RateLimitDonateKey 生成投喂限流缓存键
func (s *CacheService) RateLimitDonateKey(linuxDoID string) string {
	return model.RateLimitDonate + linuxDoID
}

// RateLimitAPIKey 生成API限流缓存键
func (s *CacheService) RateLimitAPIKey(identifier string) string {
	return model.RateLimitAPI + identifier
}

// 基础缓存操作

// Set 设置缓存
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// Get 获取缓存
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.cache.Get(ctx, key)
}

// GetJSON 获取JSON缓存
func (s *CacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return s.cache.GetJSON(ctx, key, dest)
}

// SetJSON 设置JSON缓存
func (s *CacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.cache.SetJSON(ctx, key, value, ttl)
}

// Del 删除缓存
func (s *CacheService) Del(ctx context.Context, keys ...string) error {
	return s.cache.Del(ctx, keys...)
}

// Exists 检查缓存是否存在
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

// Expire 设置缓存过期时间
func (s *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.cache.Expire(ctx, key, ttl)
}

// TTL 获取缓存剩余时间
func (s *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.cache.TTL(ctx, key)
}

// Incr 自增计数器
func (s *CacheService) Incr(ctx context.Context, key string) (int64, error) {
	return s.cache.Incr(ctx, key)
}

// IncrBy 自增指定值
func (s *CacheService) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.cache.IncrBy(ctx, key, value)
}

// Decr 自减计数器
func (s *CacheService) Decr(ctx context.Context, key string) (int64, error) {
	return s.cache.Decr(ctx, key)
}

// DecrBy 自减指定值
func (s *CacheService) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.cache.DecrBy(ctx, key, value)
}

// 用户相关缓存

// GetUserQuota 获取用户额度缓存
func (s *CacheService) GetUserQuota(ctx context.Context, linuxDoID string) (*model.QuotaInfo, error) {
	key := s.UserQuotaKey(linuxDoID)
	var quota model.QuotaInfo
	err := s.GetJSON(ctx, key, &quota)
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

// SetUserQuota 设置用户额度缓存
func (s *CacheService) SetUserQuota(ctx context.Context, linuxDoID string, quota *model.QuotaInfo, ttl time.Duration) error {
	key := s.UserQuotaKey(linuxDoID)
	return s.SetJSON(ctx, key, quota, ttl)
}

// ClearUserQuota 清除用户额度缓存
func (s *CacheService) ClearUserQuota(ctx context.Context, linuxDoID string) error {
	key := s.UserQuotaKey(linuxDoID)
	return s.Del(ctx, key)
}

// HasClaimedToday 检查用户今天是否已领取（缓存）
func (s *CacheService) HasClaimedToday(ctx context.Context, linuxDoID string) (bool, error) {
	key := s.ClaimTodayKey(linuxDoID)
	return s.Exists(ctx, key)
}

// MarkClaimedToday 标记用户今天已领取
func (s *CacheService) MarkClaimedToday(ctx context.Context, linuxDoID string) error {
	key := s.ClaimTodayKey(linuxDoID)
	// 缓存到明天凌晨
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	ttl := tomorrow.Sub(now)
	return s.Set(ctx, key, "1", ttl)
}

// GetDonateCount 获取今日投喂次数
func (s *CacheService) GetDonateCount(ctx context.Context, linuxDoID string) (int64, error) {
	key := s.DonateCountKey(linuxDoID)
	val, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if val == "" {
		return 0, nil
	}
	var count int64
	_, err = fmt.Sscanf(val, "%d", &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IncrDonateCount 增加今日投喂次数
func (s *CacheService) IncrDonateCount(ctx context.Context, linuxDoID string) (int64, error) {
	key := s.DonateCountKey(linuxDoID)
	count, err := s.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	// 设置过期时间到明天凌晨
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	ttl := tomorrow.Sub(now)
	_ = s.Expire(ctx, key, ttl)

	return count, nil
}

// Bloom Filter 相关

// BloomFilterAdd 向布隆过滤器添加Key
func (s *CacheService) BloomFilterAdd(ctx context.Context, keyHash string) error {
	return s.cache.BloomAdd(ctx, model.CacheKeyKeysBloom, keyHash)
}

// BloomFilterExists 检查Key是否可能存在于布隆过滤器
func (s *CacheService) BloomFilterExists(ctx context.Context, keyHash string) (bool, error) {
	return s.cache.BloomExists(ctx, model.CacheKeyKeysBloom, keyHash)
}

// BloomFilterAddBatch 批量向布隆过滤器添加Key
func (s *CacheService) BloomFilterAddBatch(ctx context.Context, keyHashes []string) error {
	for _, keyHash := range keyHashes {
		if err := s.BloomFilterAdd(ctx, keyHash); err != nil {
			s.logger.WithError(err).WithField("key_hash", keyHash).Warn("Failed to add key to bloom filter")
		}
	}
	return nil
}

// BloomFilterExistsBatch 批量检查Key是否可能存在于布隆过滤器
func (s *CacheService) BloomFilterExistsBatch(ctx context.Context, keyHashes []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, keyHash := range keyHashes {
		exists, err := s.BloomFilterExists(ctx, keyHash)
		if err != nil {
			s.logger.WithError(err).WithField("key_hash", keyHash).Warn("Failed to check bloom filter")
			result[keyHash] = false
			continue
		}
		result[keyHash] = exists
	}
	return result, nil
}

// 限流相关

// CheckRateLimit 检查限流
func (s *CacheService) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	count, err := s.Incr(ctx, key)
	if err != nil {
		return false, err
	}

	if count == 1 {
		// 第一次访问，设置过期时间
		_ = s.Expire(ctx, key, window)
	}

	if count > limit {
		return false, nil
	}

	return true, nil
}

// GetRateLimitRemaining 获取限流剩余次数
func (s *CacheService) GetRateLimitRemaining(ctx context.Context, key string, limit int64) (int64, error) {
	val, err := s.Get(ctx, key)
	if err != nil {
		return limit, nil
	}
	if val == "" {
		return limit, nil
	}

	var count int64
	_, err = fmt.Sscanf(val, "%d", &count)
	if err != nil {
		return limit, nil
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// ResetRateLimit 重置限流
func (s *CacheService) ResetRateLimit(ctx context.Context, key string) error {
	return s.Del(ctx, key)
}

// 批量操作

// DeletePattern 删除匹配模式的所有键
func (s *CacheService) DeletePattern(ctx context.Context, pattern string) (int64, error) {
	keys, err := s.cache.Keys(ctx, pattern)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, nil
	}

	err = s.Del(ctx, keys...)
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

// ClearUserCache 清除用户相关的所有缓存
func (s *CacheService) ClearUserCache(ctx context.Context, linuxDoID string) error {
	keys := []string{
		s.UserKey(linuxDoID),
		s.UserQuotaKey(linuxDoID),
		s.ClaimTodayKey(linuxDoID),
		s.DonateCountKey(linuxDoID),
	}

	return s.Del(ctx, keys...)
}

// ClearAllUserCaches 清除所有用户缓存
func (s *CacheService) ClearAllUserCaches(ctx context.Context) error {
	patterns := []string{
		model.CacheKeyUser + "*",
		model.CacheKeyUserQuota + "*",
		model.CacheKeyClaimToday + "*",
		model.CacheKeyDonateCount + "*",
	}

	var totalDeleted int64
	for _, pattern := range patterns {
		deleted, err := s.DeletePattern(ctx, pattern)
		if err != nil {
			s.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to delete cache pattern")
			continue
		}
		totalDeleted += deleted
	}

	s.logger.WithField("total_deleted", totalDeleted).Info("All user caches cleared")
	return nil
}

// Ping 检查Redis连接
func (s *CacheService) Ping(ctx context.Context) error {
	return s.cache.Ping(ctx)
}

// Close 关闭Redis连接
func (s *CacheService) Close() error {
	return s.cache.Close()
}

// Stats 获取缓存统计信息
func (s *CacheService) Stats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取各类缓存的数量
	userKeys, _ := s.cache.Keys(ctx, model.CacheKeyUser+"*")
	stats["user_cache_count"] = len(userKeys)

	quotaKeys, _ := s.cache.Keys(ctx, model.CacheKeyUserQuota+"*")
	stats["quota_cache_count"] = len(quotaKeys)

	claimKeys, _ := s.cache.Keys(ctx, model.CacheKeyClaimToday+"*")
	stats["claim_cache_count"] = len(claimKeys)

	sessionKeys, _ := s.cache.Keys(ctx, model.CacheKeySession+"*")
	stats["session_cache_count"] = len(sessionKeys)

	return stats, nil
}
