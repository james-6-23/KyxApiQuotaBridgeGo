package config

import "os"

// Config 应用配置
type Config struct {
	// Linux Do OAuth2
	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoRedirectURI  string

	// 管理员密码
	AdminPassword string

	// API 端点
	LinuxDoAuthURL     string
	LinuxDoTokenURL    string
	LinuxDoUserInfoURL string
	KyxAPIBase         string
	ModelScopeAPIBase  string

	// 默认配置
	DefaultClaimQuota int64
	DonateQuotaPerKey int64
	MinQuotaThreshold int64

	// 数据库路径
	DatabasePath string
}

// Load 加载配置
func Load() *Config {
	return &Config{
		// Linux Do OAuth2
		LinuxDoClientID:     getEnv("LINUX_DO_CLIENT_ID", ""),
		LinuxDoClientSecret: getEnv("LINUX_DO_CLIENT_SECRET", ""),
		LinuxDoRedirectURI:  getEnv("LINUX_DO_REDIRECT_URI", ""),

		// 管理员密码
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		// API 端点
		LinuxDoAuthURL:     "https://connect.linux.do/oauth2/authorize",
		LinuxDoTokenURL:    "https://connect.linux.do/oauth2/token",
		LinuxDoUserInfoURL: "https://connect.linux.do/api/user",
		KyxAPIBase:         "https://api.kkyyxx.xyz",
		ModelScopeAPIBase:  "https://api-inference.modelscope.cn/v1",

		// 默认配置
		DefaultClaimQuota: 20000000,
		DonateQuotaPerKey: 100000000,
		MinQuotaThreshold: 10000000,

		// 数据库路径
		DatabasePath: getEnv("DATABASE_PATH", "./data.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
