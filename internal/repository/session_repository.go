package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/model"
	"github.com/yourusername/kyx-quota-bridge/pkg/cache"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

// SessionRepository 会话仓库
type SessionRepository struct {
	db     *database.DB
	cache  *cache.Redis
	logger *logrus.Logger
}

// NewSessionRepository 创建会话仓库
func NewSessionRepository(db *database.DB, cache *cache.Redis, logger *logrus.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// Create 创建会话
func (r *SessionRepository) Create(ctx context.Context, session *model.Session, ttl time.Duration) error {
	// 保存到 Redis
	key := model.CacheKeySession + session.SessionID
	if err := r.cache.SetJSON(ctx, key, session.Data, ttl); err != nil {
		r.logger.WithError(err).WithField("session_id", session.SessionID).Error("Failed to create session in Redis")
		return fmt.Errorf("failed to create session: %w", err)
	}

	// 异步保存到数据库（可选，作为备份）
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dataJSON, _ := json.Marshal(session.Data)
		query := `
			INSERT INTO sessions (session_id, data, expires_at, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (session_id) DO UPDATE SET
				data = EXCLUDED.data,
				expires_at = EXCLUDED.expires_at
		`

		if _, err := r.db.ExecContext(bgCtx, query, session.SessionID, dataJSON, session.ExpiresAt, time.Now()); err != nil {
			r.logger.WithError(err).WithField("session_id", session.SessionID).Warn("Failed to backup session to database")
		}
	}()

	r.logger.WithField("session_id", session.SessionID).Debug("Session created successfully")
	return nil
}

// Get 获取会话
func (r *SessionRepository) Get(ctx context.Context, sessionID string) (*model.Session, error) {
	// 先从 Redis 获取
	key := model.CacheKeySession + sessionID
	var data model.JSONMap
	err := r.cache.GetJSON(ctx, key, &data)

	if err == nil && data != nil {
		// Redis 中存在
		ttl, _ := r.cache.TTL(ctx, key)
		return &model.Session{
			SessionID: sessionID,
			Data:      data,
			ExpiresAt: time.Now().Add(ttl),
			CreatedAt: time.Now(),
		}, nil
	}

	// Redis 中不存在，尝试从数据库获取
	var session model.Session
	query := `
		SELECT session_id, data, expires_at, created_at
		FROM sessions
		WHERE session_id = $1 AND expires_at > $2
	`

	err = r.db.QueryRowContext(ctx, query, sessionID, time.Now()).Scan(
		&session.SessionID,
		&session.Data,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, nil // 会话不存在或已过期
	}

	// 恢复到 Redis
	ttl := time.Until(session.ExpiresAt)
	if ttl > 0 {
		_ = r.cache.SetJSON(ctx, key, session.Data, ttl)
	}

	return &session, nil
}

// Update 更新会话
func (r *SessionRepository) Update(ctx context.Context, sessionID string, data model.JSONMap, ttl time.Duration) error {
	key := model.CacheKeySession + sessionID

	// 更新 Redis
	if err := r.cache.SetJSON(ctx, key, data, ttl); err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to update session in Redis")
		return fmt.Errorf("failed to update session: %w", err)
	}

	// 异步更新数据库
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dataJSON, _ := json.Marshal(data)
		query := `
			UPDATE sessions
			SET data = $1, expires_at = $2
			WHERE session_id = $3
		`

		if _, err := r.db.ExecContext(bgCtx, query, dataJSON, time.Now().Add(ttl), sessionID); err != nil {
			r.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to update session in database")
		}
	}()

	return nil
}

// Delete 删除会话
func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	// 从 Redis 删除
	key := model.CacheKeySession + sessionID
	if err := r.cache.Del(ctx, key); err != nil {
		r.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to delete session from Redis")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// 异步从数据库删除
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := `DELETE FROM sessions WHERE session_id = $1`
		if _, err := r.db.ExecContext(bgCtx, query, sessionID); err != nil {
			r.logger.WithError(err).WithField("session_id", sessionID).Warn("Failed to delete session from database")
		}
	}()

	r.logger.WithField("session_id", sessionID).Debug("Session deleted successfully")
	return nil
}

// Exists 检查会话是否存在
func (r *SessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := model.CacheKeySession + sessionID
	return r.cache.Exists(ctx, key)
}

// CleanExpired 清理过期会话（从数据库）
func (r *SessionRepository) CleanExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		r.logger.WithError(err).Error("Failed to clean expired sessions")
		return 0, fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		r.logger.WithField("count", count).Info("Expired sessions cleaned")
	}

	return count, nil
}

// GetSessionData 获取会话中的特定数据
func (r *SessionRepository) GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error) {
	session, err := r.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	if value, ok := session.Data[key]; ok {
		return value, nil
	}

	return nil, nil
}

// SetSessionData 设置会话中的特定数据
func (r *SessionRepository) SetSessionData(ctx context.Context, sessionID string, key string, value interface{}, ttl time.Duration) error {
	session, err := r.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		// 创建新会话
		session = &model.Session{
			SessionID: sessionID,
			Data:      make(model.JSONMap),
			ExpiresAt: time.Now().Add(ttl),
			CreatedAt: time.Now(),
		}
	}

	session.Data[key] = value
	return r.Update(ctx, sessionID, session.Data, ttl)
}
