package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

// DonateRepository 投喂记录仓库
type DonateRepository struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewDonateRepository 创建投喂记录仓库
func NewDonateRepository(db *database.DB, logger *logrus.Logger) *DonateRepository {
	return &DonateRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建投喂记录
func (r *DonateRepository) Create(ctx context.Context, record *model.DonateRecord) error {
	query := `
		INSERT INTO donate_records (
			linux_do_id, username, keys_count, total_quota_added,
			push_status, push_message, failed_keys, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		record.LinuxDoID,
		record.Username,
		record.KeysCount,
		record.TotalQuotaAdded,
		record.PushStatus,
		record.PushMessage,
		record.FailedKeys,
		now,
	).Scan(&record.ID, &record.CreatedAt)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id":       record.LinuxDoID,
			"username":          record.Username,
			"keys_count":        record.KeysCount,
			"total_quota_added": record.TotalQuotaAdded,
		}).Error("Failed to create donate record")
		return fmt.Errorf("failed to create donate record: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"record_id":         record.ID,
		"linux_do_id":       record.LinuxDoID,
		"username":          record.Username,
		"keys_count":        record.KeysCount,
		"total_quota_added": record.TotalQuotaAdded,
		"push_status":       record.PushStatus,
	}).Info("Donate record created successfully")

	return nil
}

// GetByID 根据ID获取投喂记录
func (r *DonateRepository) GetByID(ctx context.Context, id int) (*model.DonateRecord, error) {
	var record model.DonateRecord
	query := `
		SELECT id, linux_do_id, username, keys_count, total_quota_added,
			   push_status, push_message, failed_keys, created_at
		FROM donate_records
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &record, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).WithField("id", id).Error("Failed to get donate record by ID")
		return nil, fmt.Errorf("failed to get donate record by id: %w", err)
	}

	return &record, nil
}

// GetByLinuxDoID 获取用户的投喂记录
func (r *DonateRepository) GetByLinuxDoID(ctx context.Context, linuxDoID string, limit, offset int) ([]*model.DonateRecord, error) {
	query := `
		SELECT id, linux_do_id, username, keys_count, total_quota_added,
			   push_status, push_message, failed_keys, created_at
		FROM donate_records
		WHERE linux_do_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var records []*model.DonateRecord
	err := r.db.SelectContext(ctx, &records, query, linuxDoID, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get donate records by LinuxDoID")
		return nil, fmt.Errorf("failed to get donate records: %w", err)
	}

	return records, nil
}

// List 获取投喂记录列表（分页）
func (r *DonateRepository) List(ctx context.Context, limit, offset int) ([]*model.DonateRecord, error) {
	query := `
		SELECT id, linux_do_id, username, keys_count, total_quota_added,
			   push_status, push_message, failed_keys, created_at
		FROM donate_records
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var records []*model.DonateRecord
	err := r.db.SelectContext(ctx, &records, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list donate records")
		return nil, fmt.Errorf("failed to list donate records: %w", err)
	}

	return records, nil
}

// Count 获取投喂记录总数
func (r *DonateRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM donate_records`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count donate records")
		return 0, fmt.Errorf("failed to count donate records: %w", err)
	}

	return count, nil
}

// CountByLinuxDoID 获取用户的投喂记录总数
func (r *DonateRepository) CountByLinuxDoID(ctx context.Context, linuxDoID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM donate_records WHERE linux_do_id = $1`

	err := r.db.GetContext(ctx, &count, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count user donate records")
		return 0, fmt.Errorf("failed to count user donate records: %w", err)
	}

	return count, nil
}

// GetTotalKeys 获取用户投喂的总Key数量
func (r *DonateRepository) GetTotalKeys(ctx context.Context, linuxDoID string) (int64, error) {
	var total sql.NullInt64
	query := `
		SELECT COALESCE(SUM(keys_count), 0)
		FROM donate_records
		WHERE linux_do_id = $1
	`

	err := r.db.GetContext(ctx, &total, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total donated keys")
		return 0, fmt.Errorf("failed to get total donated keys: %w", err)
	}

	return total.Int64, nil
}

// GetTotalQuota 获取用户投喂的总额度
func (r *DonateRepository) GetTotalQuota(ctx context.Context, linuxDoID string) (int64, error) {
	var total sql.NullInt64
	query := `
		SELECT COALESCE(SUM(total_quota_added), 0)
		FROM donate_records
		WHERE linux_do_id = $1
	`

	err := r.db.GetContext(ctx, &total, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total donated quota")
		return 0, fmt.Errorf("failed to get total donated quota: %w", err)
	}

	return total.Int64, nil
}

// GetTodayStats 获取今日投喂统计
func (r *DonateRepository) GetTodayStats(ctx context.Context) (count int64, totalKeys int64, totalQuota int64, err error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	query := `
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(keys_count), 0) as total_keys,
			COALESCE(SUM(total_quota_added), 0) as total_quota
		FROM donate_records
		WHERE created_at >= $1 AND created_at < $2
	`

	var result struct {
		Count      int64 `db:"count"`
		TotalKeys  int64 `db:"total_keys"`
		TotalQuota int64 `db:"total_quota"`
	}

	err = r.db.GetContext(ctx, &result, query, today, tomorrow)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get today's donate stats")
		return 0, 0, 0, fmt.Errorf("failed to get today's donate stats: %w", err)
	}

	return result.Count, result.TotalKeys, result.TotalQuota, nil
}

// GetDateRangeStats 获取日期范围内的投喂统计
func (r *DonateRepository) GetDateRangeStats(ctx context.Context, startDate, endDate time.Time) (count int64, totalKeys int64, totalQuota int64, err error) {
	query := `
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(keys_count), 0) as total_keys,
			COALESCE(SUM(total_quota_added), 0) as total_quota
		FROM donate_records
		WHERE created_at >= $1 AND created_at < $2
	`

	var result struct {
		Count      int64 `db:"count"`
		TotalKeys  int64 `db:"total_keys"`
		TotalQuota int64 `db:"total_quota"`
	}

	err = r.db.GetContext(ctx, &result, query, startDate, endDate)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error("Failed to get date range donate stats")
		return 0, 0, 0, fmt.Errorf("failed to get date range donate stats: %w", err)
	}

	return result.Count, result.TotalKeys, result.TotalQuota, nil
}

// GetSuccessRate 获取投喂成功率
func (r *DonateRepository) GetSuccessRate(ctx context.Context, linuxDoID string) (successRate float64, err error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN push_status = 'success' THEN 1 END) as success
		FROM donate_records
		WHERE linux_do_id = $1
	`

	var result struct {
		Total   int64 `db:"total"`
		Success int64 `db:"success"`
	}

	err = r.db.GetContext(ctx, &result, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get donate success rate")
		return 0, fmt.Errorf("failed to get donate success rate: %w", err)
	}

	if result.Total == 0 {
		return 0, nil
	}

	successRate = float64(result.Success) / float64(result.Total) * 100
	return successRate, nil
}

// GetFailedRecords 获取失败的投喂记录
func (r *DonateRepository) GetFailedRecords(ctx context.Context, limit, offset int) ([]*model.DonateRecord, error) {
	query := `
		SELECT id, linux_do_id, username, keys_count, total_quota_added,
			   push_status, push_message, failed_keys, created_at
		FROM donate_records
		WHERE push_status = 'failed'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var records []*model.DonateRecord
	err := r.db.SelectContext(ctx, &records, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get failed donate records")
		return nil, fmt.Errorf("failed to get failed donate records: %w", err)
	}

	return records, nil
}

// CountFailed 获取失败的投喂记录数
func (r *DonateRepository) CountFailed(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM donate_records WHERE push_status = 'failed'`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count failed donate records")
		return 0, fmt.Errorf("failed to count failed donate records: %w", err)
	}

	return count, nil
}

// Delete 删除投喂记录
func (r *DonateRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM donate_records WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("id", id).Error("Failed to delete donate record")
		return fmt.Errorf("failed to delete donate record: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("donate record not found")
	}

	r.logger.WithField("id", id).Info("Donate record deleted successfully")
	return nil
}

// DeleteByLinuxDoID 删除用户的所有投喂记录
func (r *DonateRepository) DeleteByLinuxDoID(ctx context.Context, linuxDoID string) error {
	query := `DELETE FROM donate_records WHERE linux_do_id = $1`

	result, err := r.db.ExecContext(ctx, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to delete user donate records")
		return fmt.Errorf("failed to delete user donate records: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"linux_do_id":   linuxDoID,
		"rows_affected": rowsAffected,
	}).Info("User donate records deleted successfully")

	return nil
}
