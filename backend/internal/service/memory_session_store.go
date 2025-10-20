package service

import (
    "context"
    "sync"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
)

type MemorySessionStore struct {
    mu       sync.RWMutex
    sessions map[string]domain.Session
}

func NewMemorySessionStore() *MemorySessionStore {
    return &MemorySessionStore{sessions: make(map[string]domain.Session)}
}

func (m *MemorySessionStore) Save(ctx context.Context, session domain.Session) error {
    _ = ctx
    m.mu.Lock()
    defer m.mu.Unlock()
    m.sessions[session.ID] = session
    return nil
}

func (m *MemorySessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
    _ = ctx
    m.mu.RLock()
    defer m.mu.RUnlock()
    session, ok := m.sessions[id]
    if !ok {
        return nil, nil
    }
    return &session, nil
}

func (m *MemorySessionStore) Delete(ctx context.Context, id string) error {
    _ = ctx
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.sessions, id)
    return nil
}
