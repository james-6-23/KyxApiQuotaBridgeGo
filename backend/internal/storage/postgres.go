package storage

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
    Pool *pgxpool.Pool
}

func (p *Postgres) Close() {
    if p.Pool != nil {
        p.Pool.Close()
    }
}

func (p *Postgres) Ping(ctx context.Context) error {
    if p.Pool == nil {
        return nil
    }
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return p.Pool.Ping(ctx)
}
