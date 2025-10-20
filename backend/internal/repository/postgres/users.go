package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type UserRepository struct {
    db *storage.Postgres
}

func NewUserRepository(db *storage.Postgres) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Upsert(ctx context.Context, user domain.User) error {
    query := `
        INSERT INTO users (linux_do_id, username, kyx_user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (linux_do_id) DO UPDATE
        SET username = EXCLUDED.username,
            kyx_user_id = EXCLUDED.kyx_user_id,
            updated_at = EXCLUDED.updated_at`

    now := time.Now().UTC()
    if user.CreatedAt.IsZero() {
        user.CreatedAt = now
    }
    user.UpdatedAt = now

    _, err := r.db.Pool.Exec(ctx, query, user.LinuxDoID, user.Username, user.KYXUserID, user.CreatedAt, user.UpdatedAt)
    return err
}

func (r *UserRepository) FindByLinuxDoID(ctx context.Context, linuxDoID string) (*domain.User, error) {
    query := `SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at FROM users WHERE linux_do_id = $1`

    row := r.db.Pool.QueryRow(ctx, query, linuxDoID)

    var user domain.User
    if err := row.Scan(&user.ID, &user.LinuxDoID, &user.Username, &user.KYXUserID, &user.CreatedAt, &user.UpdatedAt); err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
    query := `SELECT id, linux_do_id, username, kyx_user_id, created_at, updated_at FROM users ORDER BY created_at DESC`

    rows, err := r.db.Pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("list users: %w", err)
    }
    defer rows.Close()

    var users []domain.User
    for rows.Next() {
        var user domain.User
        if err := rows.Scan(&user.ID, &user.LinuxDoID, &user.Username, &user.KYXUserID, &user.CreatedAt, &user.UpdatedAt); err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    return users, rows.Err()
}
