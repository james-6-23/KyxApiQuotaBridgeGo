package service

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "sort"
    "strings"
    "time"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/kyx"
    "github.com/kyx-api-quota-bridge/backend/internal/integrations/modelscope"
    "github.com/kyx-api-quota-bridge/backend/internal/repository"
)

type DonationService struct {
    cfg        *config.Config
    users      repository.UserRepository
    donations  repository.DonateRepository
    keys       repository.DonatedKeyRepository
    kyx        *kyx.Client
    modelscope *modelscope.Client
    adminCfg   repository.AdminConfigRepository
}

type ValidateKeysResult struct {
    ValidKeys        []string
    InvalidKeys      []string
    DuplicateRemoved int
    AlreadyExists    int
    TotalQuota       int64
    Message          string
}

func (s *DonationService) ValidateAndDonate(ctx context.Context, session domain.Session, keys []string) (*ValidateKeysResult, error) {
    user, err := s.users.FindByLinuxDoID(ctx, session.LinuxDoID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not bound")
    }

    cleaned := make([]string, 0, len(keys))
    seen := map[string]struct{}{}
    duplicates := 0
    for _, key := range keys {
        key = strings.TrimSpace(key)
        if key == "" {
            continue
        }
        if _, ok := seen[key]; ok {
            duplicates++
            continue
        }
        seen[key] = struct{}{}
        cleaned = append(cleaned, key)
    }

    var (
        validKeys   []string
        invalidKeys []string
        existsCount int
    )

    for _, key := range cleaned {
        exists, err := s.keys.Exists(ctx, key)
        if err != nil {
            return nil, err
        }
        if exists {
            existsCount++
            continue
        }
        ok, err := s.modelscope.ValidateKey(ctx, key)
        if err != nil {
            return nil, err
        }
        if ok {
            validKeys = append(validKeys, key)
        } else {
            invalidKeys = append(invalidKeys, key)
        }
    }

    if len(validKeys) == 0 {
        return &ValidateKeysResult{
            InvalidKeys:      invalidKeys,
            DuplicateRemoved: duplicates,
            AlreadyExists:    existsCount,
            Message:          "没有新的有效 Key",
        }, errors.New("no valid keys")
    }

    adminCfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return nil, err
    }

    kyxUser, err := s.findSnapshot(ctx, user.Username)
    if err != nil {
        return nil, err
    }

    totalQuota := int64(len(validKeys)) * s.cfg.KYX.DonateQuotaPerKey
    updateRes, err := s.kyx.UpdateQuota(ctx, kyx.UpdateQuotaRequest{
        ID:       user.KYXUserID,
        Quota:    kyxUser.Quota + totalQuota,
        Username: kyxUser.Username,
        Group:    kyxUser.Group,
    })
    if err != nil {
        return nil, err
    }
    if !updateRes.Success {
        return nil, fmt.Errorf("quota update failed: %s", updateRes.Message)
    }

    record := domain.DonateRecord{
        LinuxDoID:       session.LinuxDoID,
        Username:        user.Username,
        KeysCount:       len(validKeys),
        TotalQuotaAdded: totalQuota,
        PushStatus:      "pending",
        CreatedAt:       time.Now().UTC(),
    }

    id, err := s.donations.Insert(ctx, record)
    if err != nil {
        return nil, err
    }
    record.ID = id

    donKeys := make([]domain.DonatedKey, 0, len(validKeys))
    for _, key := range validKeys {
        donKeys = append(donKeys, domain.DonatedKey{
            KeyValue:       key,
            LinuxDoID:      session.LinuxDoID,
            Username:       user.Username,
            DonateRecordID: id,
            CreatedAt:      time.Now().UTC(),
            Used:           true,
        })
    }
    if err := s.keys.InsertMany(ctx, donKeys); err != nil {
        return nil, err
    }

    pushStatus := "success"
    pushMessage := "推送成功"
    failed := []string{}

    if adminCfg.KeysAuthorization != "" {
        if err := s.pushKeys(ctx, validKeys, adminCfg); err != nil {
            pushStatus = "failed"
            pushMessage = err.Error()
            failed = validKeys
        }
    } else {
        pushStatus = "failed"
        pushMessage = "未配置推送授权"
        failed = validKeys
    }

    record.PushStatus = pushStatus
    record.PushMessage = pushMessage
    record.FailedKeys = failed
    if err := s.donations.Update(ctx, record); err != nil {
        return nil, err
    }

    return &ValidateKeysResult{
        ValidKeys:        validKeys,
        InvalidKeys:      invalidKeys,
        DuplicateRemoved: duplicates,
        AlreadyExists:    existsCount,
        TotalQuota:       totalQuota,
        Message:          fmt.Sprintf("成功投喂 %d 个新 Key", len(validKeys)),
    }, nil
}

func (s *DonationService) pushKeys(ctx context.Context, keys []string, cfg *domain.AdminConfig) error {
    payload := map[string]any{
        "group_id":  cfg.GroupID,
        "keys_text": strings.Join(keys, "\n"),
    }
    body, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.KeysAPIURL, bytes.NewReader(body))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.KeysAuthorization))

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return fmt.Errorf("推送失败: %s", res.Status)
    }
    return nil
}

func (s *DonationService) findSnapshot(ctx context.Context, username string) (*kyx.User, error) {
    res, err := s.kyx.SearchUser(ctx, username, 1, 100)
    if err != nil {
        return nil, err
    }
    if !res.Success {
        return nil, fmt.Errorf(res.Message)
    }
    for _, user := range res.Data.Items {
        if strings.EqualFold(user.Username, username) {
            return &user, nil
        }
    }
    return nil, fmt.Errorf("user not found")
}

func (s *DonationService) RetryPush(ctx context.Context, linuxDoID string, timestamp int64) error {
    record, err := s.donations.FindByLinuxDoIDAndTimestamp(ctx, linuxDoID, timestamp)
    if err != nil {
        return err
    }
    if record == nil || len(record.FailedKeys) == 0 {
        return errors.New("no failed keys")
    }

    adminCfg, err := s.adminCfg.Get(ctx)
    if err != nil {
        return err
    }
    if adminCfg.KeysAuthorization == "" {
        return errors.New("keys authorization not configured")
    }

    if err := s.pushKeys(ctx, record.FailedKeys, adminCfg); err != nil {
        return err
    }

    record.PushStatus = "success"
    record.PushMessage = "推送成功"
    record.FailedKeys = nil
    return s.donations.Update(ctx, *record)
}

func (s *DonationService) DeleteKeys(ctx context.Context, keys []string) error {
    if len(keys) == 0 {
        return nil
    }
    return s.keys.Delete(ctx, keys)
}

func (s *DonationService) ExportKeys(ctx context.Context) ([]domain.DonatedKey, error) {
    keys, err := s.keys.ListAll(ctx)
    if err != nil {
        return nil, err
    }
    sort.Slice(keys, func(i, j int) bool {
        return keys[i].CreatedAt.After(keys[j].CreatedAt)
    })
    return keys, nil
}
