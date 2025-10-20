package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/kyx"
    "github.com/kyx-api-quota-bridge/backend/internal/repository"
)

type AdminService struct {
    cfg       *config.Config
    users     repository.UserRepository
    claims    repository.ClaimRepository
    donations repository.DonateRepository
    keys      repository.DonatedKeyRepository
    adminCfg  repository.AdminConfigRepository
    kyx       *kyx.Client
}

func (s *AdminService) Authenticate(password string) bool {
    return password == s.cfg.Cache.RedisURL // placeholder, replace with secure config
}

func (s *AdminService) GetConfig(ctx context.Context) (*domain.AdminConfig, error) {
    cfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return nil, err
    }
    if cfg == nil {
        return &domain.AdminConfig{}, nil
    }
    return cfg, nil
}

func (s *AdminService) UpdateConfig(ctx context.Context, update domain.AdminConfig) error {
    return s.adminCfg.Update(ctx, update)
}

func (s *AdminService) ListClaims(ctx context.Context) ([]domain.ClaimRecord, error) {
    return s.claims.ListAll(ctx)
}

func (s *AdminService) ListDonations(ctx context.Context) ([]domain.DonateRecord, error) {
    return s.donations.ListAll(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context) ([]domain.UserStats, error) {
    users, err := s.users.List(ctx)
    if err != nil {
        return nil, err
    }
    claims, err := s.claims.ListAll(ctx)
    if err != nil {
        return nil, err
    }
    donations, err := s.donations.ListAll(ctx)
    if err != nil {
        return nil, err
    }

    stats := make(map[string]*domain.UserStats)
    for _, user := range users {
        stats[user.LinuxDoID] = &domain.UserStats{
            Username:  user.Username,
            LinuxDoID: user.LinuxDoID,
            CreatedAt: user.CreatedAt,
        }
    }

    for _, claim := range claims {
        st := stats[claim.LinuxDoID]
        if st == nil {
            continue
        }
        st.ClaimCount++
        st.ClaimQuota += claim.QuotaAdded
    }

    for _, donate := range donations {
        st := stats[donate.LinuxDoID]
        if st == nil {
            continue
        }
        st.DonateCount += donate.KeysCount
        st.DonateQuota += donate.TotalQuotaAdded
    }

    result := make([]domain.UserStats, 0, len(stats))
    for _, st := range stats {
        st.TotalQuota = st.ClaimQuota + st.DonateQuota
        result = append(result, *st)
    }

    sort.Slice(result, func(i, j int) bool {
        return result[i].TotalQuota > result[j].TotalQuota
    })

    return result, nil
}

func (s *AdminService) ExportUsers(ctx context.Context) ([]byte, error) {
    users, err := s.users.List(ctx)
    if err != nil {
        return nil, err
    }
    claims, err := s.claims.ListAll(ctx)
    if err != nil {
        return nil, err
    }
    donations, err := s.donations.ListAll(ctx)
    if err != nil {
        return nil, err
    }

    data := domain.UserExport{
        ExportTime: time.Now().UTC(),
        TotalUsers: len(users),
        Users:      make([]domain.UserExportItem, 0, len(users)),
    }

    claimByUser := make(map[string][]domain.ClaimRecord)
    for _, claim := range claims {
        claimByUser[claim.LinuxDoID] = append(claimByUser[claim.LinuxDoID], claim)
    }

    donateByUser := make(map[string][]domain.DonateRecord)
    for _, donate := range donations {
        donateByUser[donate.LinuxDoID] = append(donateByUser[donate.LinuxDoID], donate)
    }

    for _, user := range users {
        claims := claimByUser[user.LinuxDoID]
        donations := donateByUser[user.LinuxDoID]

        claimQuota := int64(0)
        for _, c := range claims {
            claimQuota += c.QuotaAdded
        }

        donateQuota := int64(0)
        donateCount := 0
        for _, d := range donations {
            donateQuota += d.TotalQuotaAdded
            donateCount += d.KeysCount
        }

        item := domain.UserExportItem{
            Username:   user.Username,
            LinuxDoID:  user.LinuxDoID,
            KYXUserID:  user.KYXUserID,
            CreatedAt:  user.CreatedAt,
            Statistics: domain.UserStatistics{
                ClaimCount:      len(claims),
                ClaimQuota:      claimQuota,
                DonateCount:     donateCount,
                DonateQuota:     donateQuota,
                TotalQuota:      claimQuota + donateQuota,
            },
        }
        data.Users = append(data.Users, item)
    }

    for i := range data.Users {
        data.Users[i].Statistics.ClaimQuotaCNY = float64(data.Users[i].Statistics.ClaimQuota) * 0.02
        data.Users[i].Statistics.DonateQuotaCNY = float64(data.Users[i].Statistics.DonateQuota) * 0.02
        data.Users[i].Statistics.TotalQuotaCNY = float64(data.Users[i].Statistics.TotalQuota) * 0.02
    }

    sort.Slice(data.Users, func(i, j int) bool {
        return data.Users[i].Statistics.TotalQuota > data.Users[j].Statistics.TotalQuota
    })

    return json.MarshalIndent(data, "", "  ")
}

func (s *AdminService) ImportUsers(ctx context.Context, payload domain.UserImportPayload) (int, int, error) {
    imported := 0
    skipped := 0

    for _, user := range payload.Users {
        existing, err := s.users.FindByLinuxDoID(ctx, user.LinuxDoID)
        if err != nil {
            return imported, skipped, err
        }
        if existing != nil {
            skipped++
            continue
        }

        if err := s.users.Upsert(ctx, domain.User{
            LinuxDoID: user.LinuxDoID,
            Username:  user.Username,
            KYXUserID: user.KYXUserID,
            CreatedAt: user.CreatedAt,
            UpdatedAt: time.Now().UTC(),
        }); err != nil {
            return imported, skipped, err
        }
        imported++
    }

    return imported, skipped, nil
}

func (s *AdminService) RebindUser(ctx context.Context, linuxDoID, newUsername string) error {
    user, err := s.users.FindByLinuxDoID(ctx, linuxDoID)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }

    snapshot, err := s.kyx.SearchUser(ctx, newUsername, 1, 100)
    if err != nil {
        return err
    }
    if !snapshot.Success {
        return fmt.Errorf(snapshot.Message)
    }

    var match *kyx.User
    for _, item := range snapshot.Data.Items {
        if strings.EqualFold(item.Username, newUsername) {
            match = &item
            break
        }
    }
    if match == nil {
        return errors.New("user not found")
    }
    if match.LinuxDoID != linuxDoID {
        return fmt.Errorf("linuxdo id mismatch: %s", match.LinuxDoID)
    }

    return s.users.Upsert(ctx, domain.User{
        LinuxDoID: linuxDoID,
        Username:  match.Username,
        KYXUserID: match.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: time.Now().UTC(),
    })
}
