package storage

import (
    "context"
    "fmt"

    "github.com/cenkalti/backoff/v4"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
)

func Provide(cfg *config.Config) (*Postgres, error) {
    var pool *pgxpool.Pool
    operation := func() error {
        p, err := pgxpool.New(context.Background(), cfg.Postgres.DSN)
        if err != nil {
            return err
        }
        if err := p.Ping(context.Background()); err != nil {
            p.Close()
            return err
        }
        pool = p
        return nil
    }

    b := backoff.NewExponentialBackOff()
    b.MaxElapsedTime = cfg.Postgres.ConnMaxLifetime

    if err := backoff.Retry(operation, b); err != nil {
        return nil, fmt.Errorf("connect postgres: %w", err)
    }

    pool.Config().MaxConns = int32(cfg.Postgres.MaxOpenConns)
    pool.Config().MinConns = int32(cfg.Postgres.MaxIdleConns)

    return &Postgres{Pool: pool}, nil
}
