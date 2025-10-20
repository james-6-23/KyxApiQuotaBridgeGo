package service

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
)

type RedisSessionStore struct {
    client *redis.Client
    prefix string
    ttl    time.Duration
}

func NewRedisSessionStore(client *redis.Client, ttl time.Duration) *RedisSessionStore {
    return &RedisSessionStore{client: client, prefix: "sessions:", ttl: ttl}
}

func (r *RedisSessionStore) key(id string) string {
    return r.prefix + id
}

func (r *RedisSessionStore) Save(ctx context.Context, session domain.Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, r.key(session.ID), data, r.ttl).Err()
}

func (r *RedisSessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
    result, err := r.client.Get(ctx, r.key(id)).Bytes()
    if err == redis.Nil {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    var session domain.Session
    if err := json.Unmarshal(result, &session); err != nil {
        return nil, err
    }
    return &session, nil
}

func (r *RedisSessionStore) Delete(ctx context.Context, id string) error {
    return r.client.Del(ctx, r.key(id)).Err()
}
