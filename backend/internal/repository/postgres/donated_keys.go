package postgres

import (
    "context"

    "github.com/jackc/pgx/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type DonatedKeyRepository struct {
    db *storage.Postgres
}

func NewDonatedKeyRepository(db *storage.Postgres) *DonatedKeyRepository {
    return &DonatedKeyRepository{db: db}
}

func (r *DonatedKeyRepository) MarkUsed(ctx context.Context, key string) error {
    query := `UPDATE donated_keys SET is_used = true WHERE key_value = $1`
    _, err := r.db.Pool.Exec(ctx, query, key)
    return err
}

func (r *DonatedKeyRepository) Exists(ctx context.Context, key string) (bool, error) {
    query := `SELECT 1 FROM donated_keys WHERE key_value = $1`
    row := r.db.Pool.QueryRow(ctx, query, key)
    var exists int
    if err := row.Scan(&exists); err != nil {
        return false, nil
    }
    return true, nil
}

func (r *DonatedKeyRepository) InsertMany(ctx context.Context, keys []domain.DonatedKey) error {
    batch := &pgx.Batch{}
    query := `INSERT INTO donated_keys (key_value, linux_do_id, username, donate_record_id, created_at, is_used) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (key_value) DO NOTHING`
    for _, key := range keys {
        batch.Queue(query, key.KeyValue, key.LinuxDoID, key.Username, key.DonateRecordID, key.CreatedAt, key.Used)
    }
    br := r.db.Pool.SendBatch(ctx, batch)
    if err := br.Close(); err != nil {
        return err
    }
    return nil
}

func (r *DonatedKeyRepository) ListAll(ctx context.Context) ([]domain.DonatedKey, error) {
    query := `SELECT key_value, linux_do_id, username, donate_record_id, created_at, is_used FROM donated_keys ORDER BY created_at DESC`
    rows, err := r.db.Pool.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var keys []domain.DonatedKey
    for rows.Next() {
        var key domain.DonatedKey
        if err := rows.Scan(&key.KeyValue, &key.LinuxDoID, &key.Username, &key.DonateRecordID, &key.CreatedAt, &key.Used); err != nil {
            return nil, err
        }
        keys = append(keys, key)
    }
    return keys, rows.Err()
}

func (r *DonatedKeyRepository) Delete(ctx context.Context, keys []string) error {
    query := `DELETE FROM donated_keys WHERE key_value = ANY($1)`
    _, err := r.db.Pool.Exec(ctx, query, keys)
    return err
}
