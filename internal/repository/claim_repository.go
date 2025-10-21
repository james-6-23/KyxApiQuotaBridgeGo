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

// ClaimRepository 领取记录仓库
type ClaimRepository struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewClaimRepository 创建领取记录仓库
func NewClaimRepository(db *database.DB, logger *logrus.Logger) *ClaimRepository {
	return &ClaimRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建领取记录
func (r *ClaimRepository) Create(ctx context.Context, record *model.ClaimRecord) error {
	query := `
		INSERT INTO claim_records (linux_do_id, username, quota_added, claim_date, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	now := time.Now()
	claimDate := now.Format("2006-01-02")

	err := r.db.QueryRowContext(
		ctx,
		query,
		record.LinuxDoID,
		record.Username,
		record.QuotaAdded,
		claimDate,
		now,
	).Scan(&record.ID, &record.CreatedAt)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": record.LinuxDoID,
			"username":    record.Username,
			"quota_added": record.QuotaAdded,
		}).Error("Failed to create claim record")
		return fmt.Errorf("failed to create claim record: %w", err)
	}

	record.ClaimDate = claimDate

	r.logger.WithFields(logrus.Fields{
		"record_id":   record.ID,
		"linux_do_id": record.LinuxDoID,
		"username":    record.Username,
		"quota_added": record.QuotaAdded,
		"claim_date":  claimDate,
	}).Info("Claim record created successfully")

	return nil
}

// HasClaimedToday 检查用户今天是否已领取
func (r *ClaimRepository) HasClaimedToday(ctx context.Context, linuxDoID string) (bool, error) {
	today := time.Now().Format("2006-01-02")
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM claim_records
			WHERE linux_do_id = $1 AND claim_date = $2
		)
	`

	err := r.db.GetContext(ctx, &exists, query, linuxDoID, today)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"linux_do_id": linuxDoID,
			"date":        today,
		}).Error("Failed to check if claimed today")
		return false, fmt.Errorf("failed to check if claimed today: %w", err)
	}

	return exists, nil
}

// GetByLinuxDoID 获取用户的领取记录
func (r *ClaimRepository) GetByLinuxDoID(ctx context.Context, linuxDoID string, limit, offset int) ([]*model.ClaimRecord, error) {
	query := `
		SELECT id, linux_do_id, username, quota_added, claim_date, created_at
		FROM claim_records
		WHERE linux_do_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var records []*model.ClaimRecord
	err := r.db.SelectContext(ctx, &records, query, linuxDoID, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get claim records by LinuxDoID")
		return nil, fmt.Errorf("failed to get claim records: %w", err)
	}

	return records, nil
}

// GetByDate 获取指定日期的领取记录
func (r *ClaimRepository) GetByDate(ctx context.Context, date string, limit, offset int) ([]*model.ClaimRecord, error) {
	query := `
		SELECT id, linux_do_id, username, quota_added, claim_date, created_at
		FROM claim_records
		WHERE claim_date = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var records []*model.ClaimRecord
	err := r.db.SelectContext(ctx, &records, query, date, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("date", date).Error("Failed to get claim records by date")
		return nil, fmt.Errorf("failed to get claim records by date: %w", err)
	}

	return records, nil
}

// List 获取领取记录列表（分页）
func (r *ClaimRepository) List(ctx context.Context, limit, offset int) ([]*model.ClaimRecord, error) {
	query := `
		SELECT id, linux_do_id, username, quota_added, claim_date, created_at
		FROM claim_records
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var records []*model.ClaimRecord
	err := r.db.SelectContext(ctx, &records, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list claim records")
		return nil, fmt.Errorf("failed to list claim records: %w", err)
	}

	return records, nil
}

// Count 获取领取记录总数
func (r *ClaimRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM claim_records`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to count claim records")
		return 0, fmt.Errorf("failed to count claim records: %w", err)
	}

	return count, nil
}

// CountByLinuxDoID 获取用户的领取记录总数
func (r *ClaimRepository) CountByLinuxDoID(ctx context.Context, linuxDoID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM claim_records WHERE linux_do_id = $1`

	err := r.db.GetContext(ctx, &count, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to count user claim records")
		return 0, fmt.Errorf("failed to count user claim records: %w", err)
	}

	return count, nil
}

// GetTotalQuota 获取用户总领取额度
func (r *ClaimRepository) GetTotalQuota(ctx context.Context, linuxDoID string) (int64, error) {
	var total sql.NullInt64
	query := `
		SELECT COALESCE(SUM(quota_added), 0)
		FROM claim_records
		WHERE linux_do_id = $1
	`

	err := r.db.GetContext(ctx, &total, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to get total claimed quota")
		return 0, fmt.Errorf("failed to get total claimed quota: %w", err)
	}

	return total.Int64, nil
}

// GetTodayStats 获取今日领取统计
func (r *ClaimRepository) GetTodayStats(ctx context.Context) (count int64, totalQuota int64, err error) {
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(quota_added), 0) as total_quota
		FROM claim_records
		WHERE claim_date = $1
	`

	var result struct {
		Count      int64 `db:"count"`
		TotalQuota int64 `db:"total_quota"`
	}

	err = r.db.GetContext(ctx, &result, query, today)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get today's claim stats")
		return 0, 0, fmt.Errorf("failed to get today's claim stats: %w", err)
	}

	return result.Count, result.TotalQuota, nil
}

// GetDateRangeStats 获取日期范围内的领取统计
func (r *ClaimRepository) GetDateRangeStats(ctx context.Context, startDate, endDate string) (count int64, totalQuota int64, err error) {
	query := `
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(quota_added), 0) as total_quota
		FROM claim_records
		WHERE claim_date BETWEEN $1 AND $2
	`

	var result struct {
		Count      int64 `db:"count"`
		TotalQuota int64 `db:"total_quota"`
	}

	err = r.db.GetContext(ctx, &result, query, startDate, endDate)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error("Failed to get date range claim stats")
		return 0, 0, fmt.Errorf("failed to get date range claim stats: %w", err)
	}

	return result.Count, result.TotalQuota, nil
}

// Delete 删除领取记录
func (r *ClaimRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM claim_records WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).WithField("id", id).Error("Failed to delete claim record")
		return fmt.Errorf("failed to delete claim record: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("claim record not found")
	}

	r.logger.WithField("id", id).Info("Claim record deleted successfully")
	return nil
}

// DeleteByLinuxDoID 删除用户的所有领取记录
func (r *ClaimRepository) DeleteByLinuxDoID(ctx context.Context, linuxDoID string) error {
	query := `DELETE FROM claim_records WHERE linux_do_id = $1`

	result, err := r.db.ExecContext(ctx, query, linuxDoID)
	if err != nil {
		r.logger.WithError(err).WithField("linux_do_id", linuxDoID).Error("Failed to delete user claim records")
		return fmt.Errorf("failed to delete user claim records: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"linux_do_id":   linuxDoID,
		"rows_affected": rowsAffected,
	}).Info("User claim records deleted successfully")

	return nil
}
