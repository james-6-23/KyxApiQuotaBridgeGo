package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

// KeyRepository 已使用的Key仓库
type KeyRepository struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewKeyRepository 创建Key仓库
func NewKeyRepository(db *database.DB, logger *logrus.Logger) *KeyRepository {
	return &KeyRepository{
		db:     db,
		logger: logger,
	}
}

// HashKey 计算Key的SHA256哈希值
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Add 添加已使用的Key
func (r *KeyRepository) Add(ctx context.Context, key *model.UsedKey) error {
	query := `
		INSERT INTO used_keys (key_hash, full_key, linux_do_id, username, used_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key_hash) DO NOTHING
	`

	now := time.Now()
	if key.UsedAt.IsZero() {
		key.UsedAt = now
	}

	// 如果没有提供哈希值，则计算
	if key.KeyHash == "" {
		key.KeyHash = HashKey(key.FullKey)
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		key.KeyHash,
		key.FullKey,
		key.LinuxDoID,
		key.Username,
		key.UsedAt,
	)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": key.LinuxDoID,
			"username":    key.Username,
		}).Error("Failed to add used key")
		return fmt.Errorf("failed to add used key: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Key 已存在
		r.logger.WithField("key_hash", key.KeyHash).Debug("Key already exists")
		return nil
	}

	r.logger.WithFields(logrus.Fields{
		"key_hash":    key.KeyHash,
		"linux_do_id": key.LinuxDoID,
		"username":    key.Username,
	}).Info("Used key added successfully")

	return nil
}

// AddBatch 批量添加已使用的Key
func (r *KeyRepository) AddBatch(ctx context.Context, keys []*model.UsedKey) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	query := `
		INSERT INTO used_keys (key_hash, full_key, linux_do_id, username, used_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key_hash) DO NOTHING
	`

	now := time.Now()
	addedCount := 0

	// 使用事务批量插入
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.WithError(err).Error("Failed to begin transaction for batch add keys")
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, key := range keys {
		if key.UsedAt.IsZero() {
			key.UsedAt = now
		}
		if key.KeyHash == "" {
			key.KeyHash = HashKey(key.FullKey)
		}

		result, err := tx.ExecContext(
			ctx,
			query,
			key.KeyHash,
			key.FullKey,
			key.LinuxDoID,
			key.Username,
			key.UsedAt,
		)

		if err != nil {
			r.logger.WithError(err).WithField("key_hash", key.KeyHash).Warn("Failed to add key in batch")
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			addedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.WithError(err).Error("Failed to commit batch add keys transaction")
		return addedCount, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"total":       len(keys),
		"added_count": addedCount,
	}).Info("Batch add keys completed")

	return addedCount, nil
}

// Exists 检查Key是否已被使用
func (r *KeyRepository) Exists(ctx context.Context, keyHash string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM used_keys WHERE key_hash = $1)`

	err := r.db.GetContext(ctx, &exists, query, keyHash)
	if err != nil {
		r.logger.WithError(err).WithField("key_hash", keyHash).Error("Failed to check key existence")
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return exists, nil
}

// ExistsByKey 检查完整Key是否已被使用
func (r *KeyRepository) ExistsByKey(ctx context.Context, fullKey string) (bool, error) {
	keyHash := HashKey(fullKey)
	return r.Exists(ctx, keyHash)
}

// ExistsBatch 批量检查Key是否已被使用
func (r *KeyRepository) ExistsBatch(ctx context.Context, keyHashes []string) (map[string]bool, error) {
	if len(keyHashes) == 0 {
		return make(map[string]bool), nil
	}

	query := `SELECT key_hash FROM used_keys WHERE key_hash = ANY($1)`

	var existingHashes []string
	err := r.db.SelectContext(ctx, &existingHashes, query, keyHashes)
	if err != nil {
		r.logger.WithError(err).Error("Failed to batch check key existence")
		return nil, fmt.Errorf("failed to batch check key existence: %w", err)
	}

	result := make(map[string]bool)
	for _, hash := range keyHashes {
		result[hash] = false
	}
	for _, hash := range existingHashes {
		result[hash] = true
	}

	return result, nil
}

// GetByHash 根据哈希值获取Key信息
func (r *KeyRepository) GetByHash(ctx context.Context, keyHash string) (*model.UsedKey, error) {
	var key model.UsedKey
	query := `
		SELECT key_hash, full_key, linux_do_id, username, used_at
		FROM used_keys
		WHERE key_hash = $1
	`

	err := r.db.GetContext(ctx, &key, query, keyHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("key_hash", keyHash).Error("Failed to get key by hash")
		return nil, fmt.Errorf("failed to get key by hash: %w", err)
	}

	return &key, nil
}

// GetByLinuxDoID 获取用户使用的Key列表
func (r *KeyRepository) GetByLinuxDoID(ctx context.Context, linuxDoID string, limit, offset int) ([]*model.UsedKey, error) {
	query := `
		SELECT key_hash, full_key, linux_do_id, username, used_at
		FROM used_keys
		WHERE linux_do_id = $1
		ORDER BY used_at DESC
		LIMIT $2 OFFSET $3
	`

	var keys []*model.UsedKey
	err := r.db.SelectContext(ctx, &keys, query, linuxDoID, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get keys by LinuxDoID")
		return nil, fmt.Errorf("failed to get keys by linux_do_id: %w", err)
	}

	return keys, nil
}

// List 获取已使用的Key列表（分页）
func (r *KeyRepository) List(ctx context.Context, limit, offset int) ([]*model.UsedKey, error) {
	query := `
		SELECT key_hash, full_key, linux_do_id, username, used_at
		FROM used_keys
		ORDER BY used_at DESC
		LIMIT $1 OFFSET $2
	`

	var keys []*model.UsedKey
	err := r.db.SelectContext(ctx, &keys, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list used keys")
		return nil, fmt.Errorf("failed to list used keys: %w", err)
	}

	return keys, nil
}

// Count 获取已使用的Key总数
func (r *KeyRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM used_keys`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count used keys")
		return 0, fmt.Errorf("failed to count used keys: %w", err)
	}

	return count, nil
}

// CountByLinuxDoID 获取用户使用的Key总数
func (r *KeyRepository) CountByLinuxDoID(ctx context.Context, linuxDoID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM used_keys WHERE linux_do_id = $1`

	err := r.db.GetContext(ctx, &count, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count user used keys")
		return 0, fmt.Errorf("failed to count user used keys: %w", err)
	}

	return count, nil
}

// GetTodayCount 获取今天使用的Key总数
func (r *KeyRepository) GetTodayCount(ctx context.Context) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var count int64
	query := `
		SELECT COUNT(*)
		FROM used_keys
		WHERE used_at >= $1 AND used_at < $2
	`

	err := r.db.GetContext(ctx, &count, query, today, tomorrow)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get today's used keys count")
		return 0, fmt.Errorf("failed to get today's used keys count: %w", err)
	}

	return count, nil
}

// GetDateRangeCount 获取日期范围内使用的Key总数
func (r *KeyRepository) GetDateRangeCount(ctx context.Context, startDate, endDate time.Time) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM used_keys
		WHERE used_at >= $1 AND used_at < $2
	`

	err := r.db.GetContext(ctx, &count, query, startDate, endDate)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error("Failed to get date range used keys count")
		return 0, fmt.Errorf("failed to get date range used keys count: %w", err)
	}

	return count, nil
}

// Delete 删除Key记录
func (r *KeyRepository) Delete(ctx context.Context, keyHash string) error {
	query := `DELETE FROM used_keys WHERE key_hash = $1`

	result, err := r.db.ExecContext(ctx, query, keyHash)
	if err != nil {
		r.logger.WithError(err).WithField("key_hash", keyHash).Error("Failed to delete used key")
		return fmt.Errorf("failed to delete used key: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("key not found")
	}

	r.logger.WithField("key_hash", keyHash).Info("Used key deleted successfully")
	return nil
}

// DeleteByLinuxDoID 删除用户的所有Key记录
func (r *KeyRepository) DeleteByLinuxDoID(ctx context.Context, linuxDoID string) error {
	query := `DELETE FROM used_keys WHERE linux_do_id = $1`

	result, err := r.db.ExecContext(ctx, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to delete user used keys")
		return fmt.Errorf("failed to delete user used keys: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"linux_do_id":   linuxDoID,
		"rows_affected": rowsAffected,
	}).Info("User used keys deleted successfully")

	return nil
}

// DeleteOlderThan 删除早于指定时间的Key记录
func (r *KeyRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM used_keys WHERE used_at < $1`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		r.logger.WithError(err).WithField("older_than", olderThan).Error("Failed to delete old used keys")
		return 0, fmt.Errorf("failed to delete old used keys: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.WithFields(logrus.Fields{
			"older_than":    olderThan,
			"rows_affected": rowsAffected,
		}).Info("Old used keys deleted successfully")
	}

	return rowsAffected, nil
}

// GetRecentKeys 获取最近使用的Key列表
func (r *KeyRepository) GetRecentKeys(ctx context.Context, duration time.Duration, limit int) ([]*model.UsedKey, error) {
	since := time.Now().Add(-duration)
	query := `
		SELECT key_hash, full_key, linux_do_id, username, used_at
		FROM used_keys
		WHERE used_at >= $1
		ORDER BY used_at DESC
		LIMIT $2
	`

	var keys []*model.UsedKey
	err := r.db.SelectContext(ctx, &keys, query, since, limit)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"duration": duration,
			"limit":    limit,
		}).Error("Failed to get recent used keys")
		return nil, fmt.Errorf("failed to get recent used keys: %w", err)
	}

	return keys, nil
}

// GetKeysByUsername 根据用户名获取Key列表
func (r *KeyRepository) GetKeysByUsername(ctx context.Context, username string, limit, offset int) ([]*model.UsedKey, error) {
	query := `
		SELECT key_hash, full_key, linux_do_id, username, used_at
		FROM used_keys
		WHERE username = $1
		ORDER BY used_at DESC
		LIMIT $2 OFFSET $3
	`

	var keys []*model.UsedKey
	err := r.db.SelectContext(ctx, &keys, query, username, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("username", username).Error("Failed to get keys by username")
		return nil, fmt.Errorf("failed to get keys by username: %w", err)
	}

	return keys, nil
}

// GetUniqueUsers 获取使用过Key的唯一用户数
func (r *KeyRepository) GetUniqueUsers(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(DISTINCT linux_do_id) FROM used_keys`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get unique users count")
		return 0, fmt.Errorf("failed to get unique users count: %w", err)
	}

	return count, nil
}
