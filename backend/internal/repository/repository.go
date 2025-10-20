package repository

import "context"

import "github.com/kyx-api-quota-bridge/backend/internal/domain"

type UserRepository interface {
    Upsert(ctx context.Context, user domain.User) error
    FindByLinuxDoID(ctx context.Context, linuxDoID string) (*domain.User, error)
    List(ctx context.Context) ([]domain.User, error)
}

type ClaimRepository interface {
    FindToday(ctx context.Context, linuxDoID string) (*domain.ClaimRecord, error)
    Insert(ctx context.Context, record domain.ClaimRecord) error
    ListAll(ctx context.Context) ([]domain.ClaimRecord, error)
    ListByUser(ctx context.Context, linuxDoID string) ([]domain.ClaimRecord, error)
}

type DonateRepository interface {
    Insert(ctx context.Context, record domain.DonateRecord) (int64, error)
    Update(ctx context.Context, record domain.DonateRecord) error
    ListAll(ctx context.Context) ([]domain.DonateRecord, error)
    ListByUser(ctx context.Context, linuxDoID string) ([]domain.DonateRecord, error)
    FindByLinuxDoIDAndTimestamp(ctx context.Context, linuxDoID string, createdAt int64) (*domain.DonateRecord, error)
}

type DonatedKeyRepository interface {
    MarkUsed(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    InsertMany(ctx context.Context, keys []domain.DonatedKey) error
    ListAll(ctx context.Context) ([]domain.DonatedKey, error)
    Delete(ctx context.Context, keys []string) error
}

type AdminConfigRepository interface {
    Get(ctx context.Context) (*domain.AdminConfig, error)
    Update(ctx context.Context, cfg domain.AdminConfig) error
}
