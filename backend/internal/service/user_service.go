package service

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/kyx"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/linuxdo"
    "github.com/kyx-api-quota-bridge/backend/internal/repository"
)

type UserService struct {
    cfg      *config.Config
    users    repository.UserRepository
    kyx      *kyx.Client
    linuxdo  *linuxdo.Client
    adminCfg repository.AdminConfigRepository
    claims   repository.ClaimRepository
}

type BindResult struct {
    User       domain.User
    BonusQuota int64
    BonusCNY   float64
    Message    string
}

func (s *UserService) Bind(ctx context.Context, session domain.Session, username string) (*BindResult, error) {
    if strings.TrimSpace(username) == "" {
        return nil, errors.New("username cannot be empty")
    }

    adminCfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return nil, fmt.Errorf("load admin config: %w", err)
    }
    if adminCfg == nil || adminCfg.Session == "" {
        return nil, errors.New("system not configured")
    }

    kyxUser, err := s.findExactUser(ctx, username, adminCfg)
    if err != nil {
        return nil, err
    }

    if kyxUser.LinuxDoID != session.LinuxDoID {
        return nil, fmt.Errorf("linuxdo id mismatch: %s != %s", kyxUser.LinuxDoID, session.LinuxDoID)
    }

    existing, err := s.users.FindByLinuxDoID(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }

    user := domain.User{
        LinuxDoID: session.LinuxDoID,
        Username:  kyxUser.Username,
        KYXUserID: kyxUser.ID,
    }
    if existing != nil {
        user.CreatedAt = existing.CreatedAt
    }

    if err := s.users.Upsert(ctx, user); err != nil {
        return nil, err
    }

    result := &BindResult{User: user}

    if existing == nil {
        bonus := s.cfg.KYX.FirstBindBonus
        updateRes, err := s.kyx.UpdateQuota(ctx, kyx.UpdateQuotaRequest{
            ID:       user.KYXUserID,
            Quota:    kyxUser.Quota + bonus,
            Username: kyxUser.Username,
            Group:    kyxUser.Group,
        })
        if err != nil {
            return nil, err
        }
        if !updateRes.Success {
            result.Message = "绑定成功，奖励发放失败"
            return result, nil
        }
        result.BonusQuota = bonus
        result.BonusCNY = float64(bonus) * 0.02
        result.Message = "绑定成功，奖励已发放"
    } else {
        result.Message = "重新绑定成功"
    }

    return result, nil
}

func (s *UserService) findExactUser(ctx context.Context, username string, adminCfg *domain.AdminConfig) (*kyx.User, error) {
    const pageSize = 100
    for page := 1; page <= 20; page++ {
        res, err := s.kyx.SearchUser(ctx, username, page, pageSize)
        if err != nil {
            return nil, fmt.Errorf("search user: %w", err)
        }
        if !res.Success {
            return nil, fmt.Errorf(res.Message)
        }
        for _, user := range res.Data.Items {
            if strings.EqualFold(user.Username, username) {
                return &user, nil
            }
        }
        if page*pageSize >= res.Data.Total {
            break
        }
    }
    return nil, fmt.Errorf("user not found")
}

func (s *UserService) RefreshQuota(ctx context.Context, session domain.Session) (*domain.QuotaSnapshot, error) {
    user, err := s.users.FindByLinuxDoID(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not bound")
    }

    adminCfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return nil, err
    }

    kyxUser, err := s.findExactUser(ctx, user.Username, adminCfg)
    if err != nil {
        return nil, err
    }

    todayClaim, err := s.claims.FindToday(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }
    canClaim := kyxUser.Quota < s.cfg.KYX.MinQuotaThreshold && todayClaim == nil

    return &domain.QuotaSnapshot{
        Username:     kyxUser.Username,
        DisplayName:  kyxUser.DisplayName,
        LinuxDoID:    session.LinuxDoID,
        AvatarURL:    session.AvatarURL,
        Name:         session.DisplayName,
        Quota:        kyxUser.Quota,
        UsedQuota:    kyxUser.UsedQuota,
        Total:        kyxUser.Quota + kyxUser.UsedQuota,
        CanClaim:     canClaim,
        ClaimedToday: todayClaim != nil,
    }, nil
}
