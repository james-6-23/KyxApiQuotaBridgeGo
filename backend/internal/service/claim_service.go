package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/kyx"
    "github.com/kyx-api-quota-bridge/backend/internal/repository"
)

type ClaimService struct {
    cfg      *config.Config
    users    repository.UserRepository
    claims   repository.ClaimRepository
    kyx      *kyx.Client
    adminCfg repository.AdminConfigRepository
}

type ClaimResult struct {
    QuotaAdded int64
    Message    string
}

func (s *ClaimService) ClaimDaily(ctx context.Context, session domain.Session) (*ClaimResult, error) {
    user, err := s.users.FindByLinuxDoID(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not bound")
    }

    already, err := s.claims.FindToday(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }
    if already != nil {
        return nil, errors.New("already claimed today")
    }

    adminCfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return nil, err
    }
    if adminCfg == nil {
        return nil, errors.New("admin config missing")
    }

    kyxUser, err := s.findUserSnapshot(ctx, user.Username)
    if err != nil {
        return nil, err
    }

    if kyxUser.Quota >= s.cfg.KYX.MinQuotaThreshold {
        return nil, errors.New("quota above threshold")
    }

    newQuota := kyxUser.Quota + adminCfg.ClaimQuota
    updateRes, err := s.kyx.UpdateQuota(ctx, kyx.UpdateQuotaRequest{
        ID:       user.KYXUserID,
        Quota:    newQuota,
        Username: kyxUser.Username,
        Group:    kyxUser.Group,
    })
    if err != nil {
        return nil, err
    }
    if !updateRes.Success {
        return nil, fmt.Errorf("update quota failed: %s", updateRes.Message)
    }

    record := domain.ClaimRecord{
        LinuxDoID:  session.LinuxDoID,
        Username:   user.Username,
        QuotaAdded: adminCfg.ClaimQuota,
        Date:       time.Now().UTC(),
        CreatedAt:  time.Now().UTC(),
    }
    if err := s.claims.Insert(ctx, record); err != nil {
        return nil, err
    }

    return &ClaimResult{
        QuotaAdded: adminCfg.ClaimQuota,
        Message:    fmt.Sprintf("成功添加额度 %.2f￥", float64(adminCfg.ClaimQuota)*0.02),
    }, nil
}

func (s *ClaimService) findUserSnapshot(ctx context.Context, username string) (*kyx.User, error) {
    const pageSize = 100
    for page := 1; page <= 20; page++ {
        res, err := s.kyx.SearchUser(ctx, username, page, pageSize)
        if err != nil {
            return nil, err
        }
        if !res.Success {
            return nil, fmt.Errorf(res.Message)
        }
        for _, item := range res.Data.Items {
            if item.Username == username {
                return &item, nil
            }
        }
        if page*pageSize >= res.Data.Total {
            break
        }
    }
    return nil, fmt.Errorf("user not found")
}
