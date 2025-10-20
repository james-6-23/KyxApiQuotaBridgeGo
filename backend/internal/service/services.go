package service

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/kyx"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/linuxdo"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/modelscope"
    "github.com/kyx-api-quota-bridge/backend/internal/repository"
)

type Services struct {
    Users      *UserService
    Claims     *ClaimService
    Donations  *DonationService
    Admin      *AdminService
    Sessions   *SessionService
}

func Provide(
    cfg *config.Config,
    userRepo repository.UserRepository,
    claimRepo repository.ClaimRepository,
    donateRepo repository.DonateRepository,
    keyRepo repository.DonatedKeyRepository,
    adminRepo repository.AdminConfigRepository,
) (*Services, error) {
    if cfg == nil {
        return nil, errors.New("config is required")
    }

    kyxClient := kyx.NewClient(cfg.KYX)
    linuxdoClient := linuxdo.NewClient(cfg.OAuth)
    modelscopeClient := modelscope.NewClient(cfg.ModelScope)

    users := &UserService{
        cfg:      cfg,
        users:    userRepo,
        kyx:      kyxClient,
        linuxdo:  linuxdoClient,
        adminCfg: adminRepo,
        claims:   claimRepo,
    }

    claims := &ClaimService{
        cfg:        cfg,
        users:      userRepo,
        claims:     claimRepo,
        kyx:        kyxClient,
        adminCfg:   adminRepo,
    }

    donations := &DonationService{
        cfg:        cfg,
        users:      userRepo,
        donations:  donateRepo,
        keys:       keyRepo,
        kyx:        kyxClient,
        modelscope: modelscopeClient,
        adminCfg:   adminRepo,
    }

    admin := &AdminService{
        cfg:      cfg,
        users:    userRepo,
        claims:   claimRepo,
        donations: donateRepo,
        keys:     keyRepo,
        adminCfg: adminRepo,
        kyx:      kyxClient,
    }

    sessions := &SessionService{
        ttl:       cfg.Cache.TTL,
        store:     NewMemorySessionStore(),
    }

    return &Services{
        Users:     users,
        Claims:    claims,
        Donations: donations,
        Admin:     admin,
        Sessions:  sessions,
    }, nil
}

type SessionStore interface {
    Save(ctx context.Context, session domain.Session) error
    Get(ctx context.Context, id string) (*domain.Session, error)
    Delete(ctx context.Context, id string) error
}

type SessionService struct {
    store SessionStore
    ttl   time.Duration
}

func (s *SessionService) Create(ctx context.Context, session domain.Session) (string, error) {
    if session.ID == "" {
        session.ID = uuid.NewString()
    }
    if session.ExpiresAt.IsZero() {
        session.ExpiresAt = time.Now().Add(s.ttl)
    }
    if err := s.store.Save(ctx, session); err != nil {
        return "", err
    }
    return session.ID, nil
}

func (s *SessionService) Get(ctx context.Context, id string) (*domain.Session, error) {
    session, err := s.store.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    if session == nil || session.ExpiresAt.Before(time.Now()) {
        return nil, nil
    }
    return session, nil
}

func (s *SessionService) Delete(ctx context.Context, id string) error {
    return s.store.Delete(ctx, id)
}
