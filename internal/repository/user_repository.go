package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

// UserRepository 用户仓库
type UserRepository struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *database.DB, logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("id", id).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByLinuxDoID 根据LinuxDoID获取用户
func (r *UserRepository) GetByLinuxDoID(ctx context.Context, linuxDoID string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at
		FROM users
		WHERE linux_do_id = $1
	`

	err := r.db.GetContext(ctx, &user, query, linuxDoID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user by LinuxDoID")
		return nil, fmt.Errorf("failed to get user by linux_do_id: %w", err)
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("username", username).Error("Failed to get user by username")
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (linux_do_id, username, kyx_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.LinuxDoID,
		user.Username,
		user.KyxUserID,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": user.LinuxDoID,
			"username":    user.Username,
		}).Error("Failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"linux_do_id": user.LinuxDoID,
		"username":    user.Username,
	}).Info("User created successfully")

	return nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, kyx_user_id = $2, updated_at = $3
		WHERE linux_do_id = $4
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.KyxUserID,
		now,
		user.LinuxDoID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("user not found")
	}
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", user.LinuxDoID).Error("Failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"linux_do_id": user.LinuxDoID,
		"username":    user.Username,
	}).Info("User updated successfully")

	return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, linuxDoID string) error {
	query := `DELETE FROM users WHERE linux_do_id = $1`

	result, err := r.db.ExecContext(ctx, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.WithField("linux_do_id", linuxDoID).Info("User deleted successfully")
	return nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	query := `
		SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*model.User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count 获取用户总数
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count users")
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Exists 检查用户是否存在
func (r *UserRepository) Exists(ctx context.Context, linuxDoID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE linux_do_id = $1)`

	err := r.db.GetContext(ctx, &exists, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to check user existence")
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// GetStatistics 获取用户统计信息
func (r *UserRepository) GetStatistics(ctx context.Context, linuxDoID string) (*model.UserStatistics, error) {
	var stats model.UserStatistics
	query := `
		SELECT
			linux_do_id,
			username,
			register_time,
			total_claims,
			total_claim_quota,
			total_donates,
			total_keys_donated,
			total_donate_quota,
			total_quota
		FROM user_statistics
		WHERE linux_do_id = $1
	`

	err := r.db.GetContext(ctx, &stats, query, linuxDoID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get user statistics")
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	return &stats, nil
}

// GetAllStatistics 获取所有用户统计信息
func (r *UserRepository) GetAllStatistics(ctx context.Context) ([]*model.UserStatistics, error) {
	query := `
		SELECT
			linux_do_id,
			username,
			register_time,
			total_claims,
			total_claim_quota,
			total_donates,
			total_keys_donated,
			total_donate_quota,
			total_quota
		FROM user_statistics
		ORDER BY total_quota DESC
	`

	var stats []*model.UserStatistics
	err := r.db.SelectContext(ctx, &stats, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get all user statistics")
		return nil, fmt.Errorf("failed to get all user statistics: %w", err)
	}

	return stats, nil
}

// Transaction 执行事务操作
func (r *UserRepository) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return r.db.Transaction(ctx, fn)
}
