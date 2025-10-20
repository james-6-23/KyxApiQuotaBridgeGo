package postgres

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/jackc/pgx/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type DonateRepository struct {
    db *storage.Postgres
}

func NewDonateRepository(db *storage.Postgres) *DonateRepository {
    return &DonateRepository{db: db}
}

func (r *DonateRepository) Insert(ctx context.Context, record domain.DonateRecord) (int64, error) {
    query := `
        INSERT INTO donate_records (linux_do_id, username, keys_count, total_quota_added, push_status, push_message, failed_keys, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id`

    if record.CreatedAt.IsZero() {
        record.CreatedAt = time.Now().UTC()
    }

    failedKeys, _ := json.Marshal(record.FailedKeys)

    row := r.db.Pool.QueryRow(ctx, query,
        record.LinuxDoID,
        record.Username,
        record.KeysCount,
        record.TotalQuotaAdded,
        record.PushStatus,
        record.PushMessage,
        failedKeys,
        record.CreatedAt,
    )

    var id int64
    if err := row.Scan(&id); err != nil {
        return 0, err
    }
    return id, nil
}

func (r *DonateRepository) Update(ctx context.Context, record domain.DonateRecord) error {
    query := `
        UPDATE donate_records SET push_status=$1, push_message=$2, failed_keys=$3 WHERE id=$4`

    failedKeys, _ := json.Marshal(record.FailedKeys)

    ct, err := r.db.Pool.Exec(ctx, query, record.PushStatus, record.PushMessage, failedKeys, record.ID)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return errors.New("record not found")
    }
    return nil
}

func (r *DonateRepository) ListAll(ctx context.Context) ([]domain.DonateRecord, error) {
    query := `SELECT id, linux_do_id, username, keys_count, total_quota_added, push_status, push_message, failed_keys, created_at FROM donate_records ORDER BY created_at DESC`
    rows, err := r.db.Pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []domain.DonateRecord
    for rows.Next() {
        var record domain.DonateRecord
        var failedKeys []byte
        if err := rows.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.KeysCount, &record.TotalQuotaAdded, &record.PushStatus, &record.PushMessage, &failedKeys, &record.CreatedAt); err != nil {
            return nil, err
        }
        if len(failedKeys) > 0 {
            _ = json.Unmarshal(failedKeys, &record.FailedKeys)
        }
        records = append(records, record)
    }
    return records, rows.Err()
}

func (r *DonateRepository) ListByUser(ctx context.Context, linuxDoID string) ([]domain.DonateRecord, error) {
    query := `SELECT id, linux_do_id, username, keys_count, total_quota_added, push_status, push_message, failed_keys, created_at FROM donate_records WHERE linux_do_id = $1 ORDER BY created_at DESC`
    rows, err := r.db.Pool.Query(ctx, query, linuxDoID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []domain.DonateRecord
    for rows.Next() {
        var record domain.DonateRecord
        var failedKeys []byte
        if err := rows.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.KeysCount, &record.TotalQuotaAdded, &record.PushStatus, &record.PushMessage, &failedKeys, &record.CreatedAt); err != nil {
            return nil, err
        }
        if len(failedKeys) > 0 {
            _ = json.Unmarshal(failedKeys, &record.FailedKeys)
        }
        records = append(records, record)
    }
    return records, rows.Err()
}

func (r *DonateRepository) FindByLinuxDoIDAndTimestamp(ctx context.Context, linuxDoID string, createdAt int64) (*domain.DonateRecord, error) {
    query := `SELECT id, linux_do_id, username, keys_count, total_quota_added, push_status, push_message, failed_keys, created_at FROM donate_records WHERE linux_do_id = $1 AND EXTRACT(EPOCH FROM created_at) = $2`

    row := r.db.Pool.QueryRow(ctx, query, linuxDoID, createdAt)

    var record domain.DonateRecord
    var failedKeys []byte
    if err := row.Scan(&record.ID, &record.LinuxDoID, &record.Username, &record.KeysCount, &record.TotalQuotaAdded, &record.PushStatus, &record.PushMessage, &failedKeys, &record.CreatedAt); err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    if len(failedKeys) > 0 {
        _ = json.Unmarshal(failedKeys, &record.FailedKeys)
    }
    return &record, nil
}
