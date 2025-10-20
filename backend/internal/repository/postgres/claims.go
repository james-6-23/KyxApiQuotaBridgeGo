package postgres

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type ClaimRepository struct {
    db *storage.Postgres
}

func NewClaimRepository(db *storage.Postgres) *ClaimRepository {
    return &ClaimRepository{db: db}
}

func (r *ClaimRepository) FindToday(ctx context.Context, linuxDoID string) (*domain.ClaimRecord, error) {
    query := `
        SELECT id, linux_do_id, username, quota_added, date, created_at
        FROM claim_records
        WHERE linux_do_id = $1 AND date = CURRENT_DATE`

    row := r.db.Pool.QueryRow(ctx, query, linuxDoID)

    var record domain.ClaimRecord
    if err := row.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.QuotaAdded, &record.Date, &record.CreatedAt); err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &record, nil
}

func (r *ClaimRepository) Insert(ctx context.Context, record domain.ClaimRecord) error {
    query := `
        INSERT INTO claim_records (linux_do_id, username, quota_added, date, created_at)
        VALUES ($1, $2, $3, $4, $5)`

    if record.Date.IsZero() {
        record.Date = time.Now().UTC()
    }
    if record.CreatedAt.IsZero() {
        record.CreatedAt = time.Now().UTC()
    }

    _, err := r.db.Pool.Exec(ctx, query, record.LinuxDoID, record.Username, record.QuotaAdded, record.Date, record.CreatedAt)
    return err
}

func (r *ClaimRepository) ListAll(ctx context.Context) ([]domain.ClaimRecord, error) {
    query := `SELECT id, linux_do_id, username, quota_added, date, created_at FROM claim_records ORDER BY created_at DESC`

    rows, err := r.db.Pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []domain.ClaimRecord
    for rows.Next() {
        var record domain.ClaimRecord
        if err := rows.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.QuotaAdded, &record.Date, &record.CreatedAt); err != nil {
            return nil, err
        }
        records = append(records, record)
    }
    return records, rows.Err()
}

func (r *ClaimRepository) ListByUser(ctx context.Context, linuxDoID string) ([]domain.ClaimRecord, error) {
    query := `SELECT id, linux_do_id, username, quota_added, date, created_at FROM claim_records WHERE linux_do_id = $1 ORDER BY created_at DESC`

    rows, err := r.db.Pool.Query(ctx, query, linuxDoID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []domain.ClaimRecord
    for rows.Next() {
        var record domain.ClaimRecord
        if err := rows.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.QuotaAdded, &record.Date, &record.CreatedAt); err != nil {
            return nil, err
        }
        records = append(records, record)
    }
    return records, rows.Err()
}
