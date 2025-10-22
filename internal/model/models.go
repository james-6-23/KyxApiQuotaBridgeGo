package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User 用户模型
type User struct {
	ID        int       `json:"id" db:"id"`
	LinuxDoID string    `json:"linux_do_id" db:"linux_do_id"`
	Username  string    `json:"username" db:"username"`
	KyxUserID int       `json:"kyx_user_id" db:"kyx_user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ClaimRecord 领取记录模型
type ClaimRecord struct {
	ID         int       `json:"id" db:"id"`
	LinuxDoID  string    `json:"linux_do_id" db:"linux_do_id"`
	Username   string    `json:"username" db:"username"`
	QuotaAdded int64     `json:"quota_added" db:"quota_added"`
	ClaimDate  string    `json:"claim_date" db:"claim_date"` // YYYY-MM-DD
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// DonateRecord 投喂记录模型
type DonateRecord struct {
	ID              int       `json:"id" db:"id"`
	LinuxDoID       string    `json:"linux_do_id" db:"linux_do_id"`
	Username        string    `json:"username" db:"username"`
	KeysCount       int       `json:"keys_count" db:"keys_count"`
	TotalQuotaAdded int64     `json:"total_quota_added" db:"total_quota_added"`
	PushStatus      string    `json:"push_status" db:"push_status"` // success, failed
	PushMessage     string    `json:"push_message" db:"push_message"`
	FailedKeys      JSONArray `json:"failed_keys" db:"failed_keys"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// UsedKey 已使用的Key模型
type UsedKey struct {
	KeyHash   string    `json:"key_hash" db:"key_hash"`
	FullKey   string    `json:"full_key" db:"full_key"`
	LinuxDoID string    `json:"linux_do_id" db:"linux_do_id"`
	Username  string    `json:"username" db:"username"`
	UsedAt    time.Time `json:"used_at" db:"used_at"`
}

// AdminConfig 管理员配置模型
type AdminConfig struct {
	ID                int       `json:"id" db:"id"`
	Session           string    `json:"session" db:"session"`
	NewAPIUser        string    `json:"new_api_user" db:"new_api_user"`
	ClaimQuota        int64     `json:"claim_quota" db:"claim_quota"`
	KeysAPIURL        string    `json:"keys_api_url" db:"keys_api_url"`
	KeysAuthorization string    `json:"keys_authorization" db:"keys_authorization"`
	GroupID           int       `json:"group_id" db:"group_id"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Session 会话模型
type Session struct {
	SessionID string    `json:"session_id" db:"session_id"`
	Data      JSONMap   `json:"data" db:"data"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserStatistics 用户统计模型（从视图读取）
type UserStatistics struct {
	LinuxDoID        string    `json:"linux_do_id" db:"linux_do_id"`
	Username         string    `json:"username" db:"username"`
	RegisterTime     time.Time `json:"register_time" db:"register_time"`
	TotalClaims      int       `json:"total_claims" db:"total_claims"`
	TotalClaimQuota  int64     `json:"total_claim_quota" db:"total_claim_quota"`
	TotalDonates     int       `json:"total_donates" db:"total_donates"`
	TotalKeysDonated int       `json:"total_keys_donated" db:"total_keys_donated"`
	TotalDonateQuota int64     `json:"total_donate_quota" db:"total_donate_quota"`
	TotalQuota       int64     `json:"total_quota" db:"total_quota"`
}

// JSONArray 自定义类型用于处理 PostgreSQL JSONB 数组
type JSONArray []string

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

// JSONMap 自定义类型用于处理 PostgreSQL JSONB 对象
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

// ========== API 请求/响应结构 ==========

// BindAccountRequest 绑定账号请求
type BindAccountRequest struct {
	Username string `json:"username" binding:"required"`
}

// BindAccountResponse 绑定账号响应
type BindAccountResponse struct {
	User        *User   `json:"user"`
	Bonus       int64   `json:"bonus,omitempty"`
	BonusCNY    float64 `json:"bonus_cny,omitempty"`
	IsFirstBind bool    `json:"is_first_bind"`
}

// QuotaInfo 额度信息
type QuotaInfo struct {
	Username     string `json:"username"`
	DisplayName  string `json:"display_name,omitempty"`
	LinuxDoID    string `json:"linux_do_id"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	Name         string `json:"name,omitempty"`
	Quota        int64  `json:"quota"`
	UsedQuota    int64  `json:"used_quota"`
	Total        int64  `json:"total"`
	CanClaim     bool   `json:"can_claim"`
	ClaimedToday bool   `json:"claimed_today"`
}

// DonateRequest 投喂请求
type DonateRequest struct {
	Keys []string `json:"keys" binding:"required,min=1"`
}

// DonateResponse 投喂响应
type DonateResponse struct {
	ValidKeys        int                   `json:"valid_keys"`
	AlreadyExists    int                   `json:"already_exists"`
	DuplicateRemoved int                   `json:"duplicate_removed"`
	QuotaAdded       int64                 `json:"quota_added"`
	Results          []KeyValidationResult `json:"results,omitempty"`
}

// KeyValidationResult Key验证结果
type KeyValidationResult struct {
	Key    string `json:"key"`
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// AdminConfigResponse 管理员配置响应
type AdminConfigResponse struct {
	ClaimQuota                  int64  `json:"claim_quota"`
	SessionConfigured           bool   `json:"session_configured"`
	KeysAPIURL                  string `json:"keys_api_url"`
	KeysAuthorizationConfigured bool   `json:"keys_authorization_configured"`
	GroupID                     int    `json:"group_id"`
	UpdatedAt                   int64  `json:"updated_at"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ClaimQuota        *int64  `json:"claim_quota,omitempty"`
	Session           *string `json:"session,omitempty"`
	NewAPIUser        *string `json:"new_api_user,omitempty"`
	KeysAPIURL        *string `json:"keys_api_url,omitempty"`
	KeysAuthorization *string `json:"keys_authorization,omitempty"`
	GroupID           *int    `json:"group_id,omitempty"`
}

// ========== 外部API结构 ==========

// KyxUser 公益站用户信息
type KyxUser struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	LinuxDoID   string `json:"linux_do_id"`
	Quota       int64  `json:"quota"`
	UsedQuota   int64  `json:"used_quota"`
	Group       string `json:"group"`
}

// KyxSearchResponse 公益站搜索响应
type KyxSearchResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    []KyxUser `json:"data,omitempty"`
}

// LinuxDoUserInfo Linux Do 用户信息
type LinuxDoUserInfo struct {
	ID        int    `json:"id"`        // Linux.do API 返回的是数字类型
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// LinuxDoTokenResponse Linux Do Token响应
type LinuxDoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// ========== 通用响应结构 ==========

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string `form:"sort_by" binding:"omitempty"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
}

// PaginationResult 分页结果
type PaginationResult struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
	HasMore    bool        `json:"has_more"`
	Data       interface{} `json:"data"`
}

// ========== 缓存键常量 ==========

const (
	// 缓存键前缀
	CacheKeyUser        = "user:"
	CacheKeyUserQuota   = "user:quota:"
	CacheKeyClaimToday  = "claim:today:"
	CacheKeyDonateCount = "donate:count:"
	CacheKeySession     = "session:"
	CacheKeyAdminConfig = "admin:config"
	CacheKeyKeysBloom   = "keys:bloom"

	// 限流键前缀
	RateLimitLogin  = "ratelimit:login:"
	RateLimitDonate = "ratelimit:donate:"
	RateLimitAPI    = "ratelimit:api:"
)

// ========== 辅助函数 ==========

// NewResponse 创建成功响应
func NewResponse(data interface{}, message string) *Response {
	return &Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string, err error) *ErrorResponse {
	resp := &ErrorResponse{
		Success: false,
		Message: message,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

// QuotaToDollar 额度转美元
func QuotaToDollar(quota int64) float64 {
	return float64(quota) / 500000.0
}

// DollarToQuota 美元转额度
func DollarToQuota(dollar float64) int64 {
	return int64(dollar * 500000)
}

// FormatQuota 格式化额度为美元字符串
func FormatQuota(quota int64) string {
	return "$" + fmt.Sprintf("%.2f", QuotaToDollar(quota))
}

// TableName methods for GORM compatibility (if needed in future)

func (User) TableName() string {
	return "users"
}

func (ClaimRecord) TableName() string {
	return "claim_records"
}

func (DonateRecord) TableName() string {
	return "donate_records"
}

func (UsedKey) TableName() string {
	return "used_keys"
}

func (AdminConfig) TableName() string {
	return "admin_config"
}

func (Session) TableName() string {
	return "sessions"
}
