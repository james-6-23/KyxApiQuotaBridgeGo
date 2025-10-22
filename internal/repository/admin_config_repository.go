package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/cache"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

// AdminConfigRepository 管理员配置仓库
type AdminConfigRepository struct {
	db     *database.DB
	cache  *cache.Redis
	logger *logrus.Logger
}

// NewAdminConfigRepository 创建管理员配置仓库
func NewAdminConfigRepository(db *database.DB, cache *cache.Redis, logger *logrus.Logger) *AdminConfigRepository {
	return &AdminConfigRepository{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// Get 获取管理员配置（单例模式）
func (r *AdminConfigRepository) Get(ctx context.Context) (*model.AdminConfig, error) {
	// 先从缓存获取
	var config model.AdminConfig
	err := r.cache.GetJSON(ctx, model.CacheKeyAdminConfig, &config)
	if err == nil && config.ID > 0 {
		r.logger.Debug("Admin config retrieved from cache")
		return &config, nil
	}
	// 缓存未命中或出错，继续从数据库查询
	if err != nil {
		r.logger.WithError(err).Debug("Failed to get admin config from cache, falling back to database")
	}

	// 从数据库获取
	query := `
		SELECT id, session, new_api_user, claim_quota, keys_api_url,
		       keys_authorization, group_id, updated_at
		FROM admin_config
		ORDER BY id DESC
		LIMIT 1
	`

	err = r.db.GetContext(ctx, &config, query)
	if err == sql.ErrNoRows {
		// 配置不存在，返回空配置
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to get admin config")
		return nil, fmt.Errorf("failed to get admin config: %w", err)
	}

	// 缓存配置（1小时）
	_ = r.cache.SetJSON(ctx, model.CacheKeyAdminConfig, &config, time.Hour)

	r.logger.WithField("config_id", config.ID).Debug("Admin config retrieved from database")
	return &config, nil
}

// Create 创建管理员配置
func (r *AdminConfigRepository) Create(ctx context.Context, config *model.AdminConfig) error {
	query := `
		INSERT INTO admin_config (
			session, new_api_user, claim_quota, keys_api_url,
			keys_authorization, group_id, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		config.Session,
		config.NewAPIUser,
		config.ClaimQuota,
		config.KeysAPIURL,
		config.KeysAuthorization,
		config.GroupID,
		now,
	).Scan(&config.ID, &config.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create admin config")
		return fmt.Errorf("failed to create admin config: %w", err)
	}

	// 清除缓存
	_ = r.cache.Del(ctx, model.CacheKeyAdminConfig)

	r.logger.WithField("config_id", config.ID).Info("Admin config created successfully")
	return nil
}

// Update 更新管理员配置
func (r *AdminConfigRepository) Update(ctx context.Context, config *model.AdminConfig) error {
	// 先获取当前配置
	currentConfig, err := r.Get(ctx)
	if err != nil {
		return err
	}

	if currentConfig == nil {
		// 配置不存在，创建新配置
		return r.Create(ctx, config)
	}

	// 构建动态更新语句
	query := `
		UPDATE admin_config
		SET session = $1,
		    new_api_user = $2,
		    claim_quota = $3,
		    keys_api_url = $4,
		    keys_authorization = $5,
		    group_id = $6,
		    updated_at = $7
		WHERE id = $8
		RETURNING id, updated_at
	`

	now := time.Now()
	err = r.db.QueryRowContext(
		ctx,
		query,
		config.Session,
		config.NewAPIUser,
		config.ClaimQuota,
		config.KeysAPIURL,
		config.KeysAuthorization,
		config.GroupID,
		now,
		currentConfig.ID,
	).Scan(&config.ID, &config.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).WithField("config_id", currentConfig.ID).Error("Failed to update admin config")
		return fmt.Errorf("failed to update admin config: %w", err)
	}

	// 清除缓存
	_ = r.cache.Del(ctx, model.CacheKeyAdminConfig)

	r.logger.WithField("config_id", config.ID).Info("Admin config updated successfully")
	return nil
}

// UpdatePartial 部分更新管理员配置
func (r *AdminConfigRepository) UpdatePartial(ctx context.Context, updates map[string]interface{}) error {
	// 先获取当前配置
	currentConfig, err := r.Get(ctx)
	if err != nil {
		return err
	}

	if currentConfig == nil {
		return fmt.Errorf("admin config not found, please create first")
	}

	// 构建动态更新语句
	query := "UPDATE admin_config SET updated_at = $1"
	args := []interface{}{time.Now()}
	paramIndex := 2

	if val, ok := updates["session"]; ok {
		query += fmt.Sprintf(", session = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}
	if val, ok := updates["new_api_user"]; ok {
		query += fmt.Sprintf(", new_api_user = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}
	if val, ok := updates["claim_quota"]; ok {
		query += fmt.Sprintf(", claim_quota = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}
	if val, ok := updates["keys_api_url"]; ok {
		query += fmt.Sprintf(", keys_api_url = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}
	if val, ok := updates["keys_authorization"]; ok {
		query += fmt.Sprintf(", keys_authorization = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}
	if val, ok := updates["group_id"]; ok {
		query += fmt.Sprintf(", group_id = $%d", paramIndex)
		args = append(args, val)
		paramIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", paramIndex)
	args = append(args, currentConfig.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).WithField("config_id", currentConfig.ID).Error("Failed to partial update admin config")
		return fmt.Errorf("failed to partial update admin config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("admin config not found")
	}

	// 清除缓存
	_ = r.cache.Del(ctx, model.CacheKeyAdminConfig)

	r.logger.WithFields(logrus.Fields{
		"config_id": currentConfig.ID,
		"updates":   updates,
	}).Info("Admin config partially updated successfully")

	return nil
}

// Delete 删除管理员配置
func (r *AdminConfigRepository) Delete(ctx context.Context) error {
	query := `DELETE FROM admin_config`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete admin config")
		return fmt.Errorf("failed to delete admin config: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("admin config not found")
	}

	// 清除缓存
	_ = r.cache.Del(ctx, model.CacheKeyAdminConfig)

	r.logger.Info("Admin config deleted successfully")
	return nil
}

// Exists 检查配置是否存在
func (r *AdminConfigRepository) Exists(ctx context.Context) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM admin_config LIMIT 1)`

	err := r.db.GetContext(ctx, &exists, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to check admin config existence")
		return false, fmt.Errorf("failed to check admin config existence: %w", err)
	}

	return exists, nil
}

// GetClaimQuota 获取领取额度配置
func (r *AdminConfigRepository) GetClaimQuota(ctx context.Context) (int64, error) {
	config, err := r.Get(ctx)
	if err != nil {
		return 0, err
	}
	if config == nil {
		// 返回默认值
		return 500000, nil // 默认 $1
	}
	return config.ClaimQuota, nil
}

// UpdateClaimQuota 更新领取额度配置
func (r *AdminConfigRepository) UpdateClaimQuota(ctx context.Context, quota int64) error {
	return r.UpdatePartial(ctx, map[string]interface{}{
		"claim_quota": quota,
	})
}

// GetSession 获取公益站Session
func (r *AdminConfigRepository) GetSession(ctx context.Context) (string, error) {
	config, err := r.Get(ctx)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}
	return config.Session, nil
}

// UpdateSession 更新公益站Session
func (r *AdminConfigRepository) UpdateSession(ctx context.Context, session string) error {
	return r.UpdatePartial(ctx, map[string]interface{}{
		"session": session,
	})
}

// GetKeysAPIConfig 获取Keys API配置
func (r *AdminConfigRepository) GetKeysAPIConfig(ctx context.Context) (url string, authorization string, err error) {
	config, err := r.Get(ctx)
	if err != nil {
		return "", "", err
	}
	if config == nil {
		return "", "", nil
	}
	return config.KeysAPIURL, config.KeysAuthorization, nil
}

// UpdateKeysAPIConfig 更新Keys API配置
func (r *AdminConfigRepository) UpdateKeysAPIConfig(ctx context.Context, url, authorization string) error {
	return r.UpdatePartial(ctx, map[string]interface{}{
		"keys_api_url":       url,
		"keys_authorization": authorization,
	})
}

// ClearCache 清除配置缓存
func (r *AdminConfigRepository) ClearCache(ctx context.Context) error {
	err := r.cache.Del(ctx, model.CacheKeyAdminConfig)
	if err != nil {
		r.logger.WithError(err).Warn("Failed to clear admin config cache")
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	r.logger.Debug("Admin config cache cleared")
	return nil
}

// InitializeDefault 初始化默认配置
func (r *AdminConfigRepository) InitializeDefault(ctx context.Context) error {
	// 检查是否已存在配置
	exists, err := r.Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		r.logger.Info("Admin config already exists, skipping initialization")
		return nil
	}

	// 创建默认配置
	defaultConfig := &model.AdminConfig{
		Session:           "",
		NewAPIUser:        "",
		ClaimQuota:        500000, // 默认 $1
		KeysAPIURL:        "",
		KeysAuthorization: "",
		GroupID:           1,
	}

	err = r.Create(ctx, defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize default config: %w", err)
	}

	r.logger.Info("Default admin config initialized successfully")
	return nil
}
