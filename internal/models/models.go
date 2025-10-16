package models

// User 用户模型
type User struct {
	LinuxDoID string `json:"linux_do_id"`
	Username  string `json:"username"`
	KyxUserID int64  `json:"kyx_user_id"`
	CreatedAt int64  `json:"created_at"`
}

// Session 会话模型
type Session struct {
	SessionID string `json:"session_id"`
	LinuxDoID string `json:"linux_do_id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Admin     bool   `json:"admin"`
	ExpiresAt int64  `json:"expires_at"`
}

// ClaimRecord 领取记录
type ClaimRecord struct {
	LinuxDoID  string `json:"linux_do_id"`
	Username   string `json:"username"`
	QuotaAdded int64  `json:"quota_added"`
	Timestamp  int64  `json:"timestamp"`
	Date       string `json:"date"` // YYYY-MM-DD
}

// DonateRecord 投喂记录
type DonateRecord struct {
	LinuxDoID       string   `json:"linux_do_id"`
	Username        string   `json:"username"`
	KeysCount       int      `json:"keys_count"`
	TotalQuotaAdded int64    `json:"total_quota_added"`
	Timestamp       int64    `json:"timestamp"`
	PushStatus      string   `json:"push_status,omitempty"` // success | failed
	PushMessage     string   `json:"push_message,omitempty"`
	FailedKeys      []string `json:"failed_keys,omitempty"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Session           string `json:"session"`
	NewAPIUser        string `json:"new_api_user"`
	ClaimQuota        int64  `json:"claim_quota"`
	KeysAPIURL        string `json:"keys_api_url"`
	KeysAuthorization string `json:"keys_authorization"`
	GroupID           int    `json:"group_id"`
	UpdatedAt         int64  `json:"updated_at"`
}

// DonatedKey 投喂的 Key
type DonatedKey struct {
	Key       string `json:"key"`
	LinuxDoID string `json:"linux_do_id"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

// KyxUser 公益站用户信息
type KyxUser struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	LinuxDoID   string `json:"linux_do_id"`
	Quota       int64  `json:"quota"`
	UsedQuota   int64  `json:"used_quota"`
	Group       string `json:"group"`
}

// KyxSearchResult 公益站搜索结果
type KyxSearchResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		Items []KyxUser `json:"items"`
	} `json:"data"`
}
