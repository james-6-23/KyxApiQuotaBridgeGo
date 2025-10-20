package domain

import "time"

type User struct {
    ID              int64
    LinuxDoID       string
    Username        string
    KYXUserID       int64
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type ClaimRecord struct {
    ID         int64
    LinuxDoID  string
    Username   string
    QuotaAdded int64
    Date       time.Time
    CreatedAt  time.Time
}

type DonateRecord struct {
    ID               int64
    LinuxDoID        string
    Username         string
    KeysCount        int
    TotalQuotaAdded  int64
    PushStatus       string
    PushMessage      string
    FailedKeys       []string
    CreatedAt        time.Time
}

type DonatedKey struct {
    KeyValue   string
    LinuxDoID  string
    Username   string
    DonateRecordID int64
    CreatedAt  time.Time
    Used       bool
}

type AdminConfig struct {
    Session           string
    NewAPIUser        string
    ClaimQuota        int64
    KeysAPIURL        string
    KeysAuthorization string
    GroupID           int
    UpdatedAt         time.Time
}

type Session struct {
    ID         string
    LinuxDoID  string
    Username   string
    DisplayName string
    AvatarURL  string
    ExpiresAt  time.Time
    IsAdmin    bool
}

type QuotaSnapshot struct {
    Username      string
    DisplayName   string
    LinuxDoID     string
    AvatarURL     string
    Name          string
    Quota         int64
    UsedQuota     int64
    Total         int64
    CanClaim      bool
    ClaimedToday  bool
}
