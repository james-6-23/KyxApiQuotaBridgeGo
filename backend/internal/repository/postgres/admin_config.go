package postgres

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type AdminConfigRepository struct {
    db *storage.Postgres
}

func NewAdminConfigRepository(db *storage.Postgres) *AdminConfigRepository {
    return &AdminConfigRepository{db: db}
}

func (r *AdminConfigRepository) Get(ctx context.Context) (*domain.AdminConfig, error) {
    query := `SELECT session, new_api_user, claim_quota, keys_api_url, keys_authorization, group_id, updated_at FROM admin_config WHERE id = 1`
    row := r.db.Pool.QueryRow(ctx, query)

    var cfg domain.AdminConfig
    if err := row.Scan(&cfg.Session, &cfg.NewAPIUser, &cfg.ClaimQuota, &cfg.KeysAPIURL, &cfg.KeysAuthorization, &cfg.GroupID, &cfg.UpdatedAt); err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &cfg, nil
}

func (r *AdminConfigRepository) Update(ctx context.Context, cfg domain.AdminConfig) error {
    query := `
        INSERT INTO admin_config (id, session, new_api_user, claim_quota, keys_api_url, keys_authorization, group_id, updated_at)
        VALUES (1, $1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            session = EXCLUDED.session,
            new_api_user = EXCLUDED.new_api_user,
            claim_quota = EXCLUDED.claim_quota,
            keys_api_url = EXCLUDED.keys_api_url,
            keys_authorization = EXCLUDED.keys_authorization,
            group_id = EXCLUDED.group_id,
            updated_at = EXCLUDED.updated_at`

    if cfg.UpdatedAt.IsZero() {
        cfg.UpdatedAt = time.Now().UTC()
    }

    _, err := r.db.Pool.Exec(ctx, query, cfg.Session, cfg.NewAPIUser, cfg.ClaimQuota, cfg.KeysAPIURL, cfg.KeysAuthorization, cfg.GroupID, cfg.UpdatedAt)
    return err
}
