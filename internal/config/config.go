package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	LinuxDo  LinuxDoConfig
	Kyx      KyxConfig
	Admin    AdminConfig
	Log      LogConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	EnableCORS   bool          `mapstructure:"enable_cors"`
	CORSOrigins  []string      `mapstructure:"cors_origins"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // minutes
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	PoolSize   int    `mapstructure:"pool_size"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// LinuxDoConfig Linux Do OAuth2配置
type LinuxDoConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri"`
	AuthURL      string `mapstructure:"auth_url"`
	TokenURL     string `mapstructure:"token_url"`
	UserInfoURL  string `mapstructure:"user_info_url"`
}

// KyxConfig 公益站API配置
type KyxConfig struct {
	APIBase             string `mapstructure:"api_base"`
	ModelScopeAPIBase   string `mapstructure:"modelscope_api_base"`
	DefaultClaimQuota   int64  `mapstructure:"default_claim_quota"`
	DonateQuotaPerKey   int64  `mapstructure:"donate_quota_per_key"`
	MinQuotaThreshold   int64  `mapstructure:"min_quota_threshold"`
	MaxDonateKeysPerDay int    `mapstructure:"max_donate_keys_per_day"`
	FirstBindBonusQuota int64  `mapstructure:"first_bind_bonus_quota"`
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Password      string `mapstructure:"password"`
	JWTSecret     string `mapstructure:"jwt_secret"`
	SessionExpire int    `mapstructure:"session_expire"` // hours
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
	Output string `mapstructure:"output"` // stdout, file
	Path   string `mapstructure:"path"`   // log file path
}

var globalConfig *Config

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在不算错误，可能使用环境变量
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	// 初始化 viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置环境变量前缀
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件（如果存在）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file error: %w", err)
		}
		// 配置文件不存在，使用默认值和环境变量
		fmt.Println("Config file not found, using defaults and environment variables")
	}

	// 从环境变量覆盖配置
	bindEnvVariables()

	config := &Config{}

	// 解析服务器配置
	config.Server = ServerConfig{
		Port:         viper.GetString("SERVER_PORT"),
		Mode:         viper.GetString("SERVER_MODE"),
		ReadTimeout:  viper.GetDuration("SERVER_READ_TIMEOUT") * time.Second,
		WriteTimeout: viper.GetDuration("SERVER_WRITE_TIMEOUT") * time.Second,
		EnableCORS:   viper.GetBool("ENABLE_CORS"),
		CORSOrigins:  viper.GetStringSlice("CORS_ORIGINS"),
	}

	// 解析数据库配置
	config.Database = DatabaseConfig{
		Host:            viper.GetString("DB_HOST"),
		Port:            viper.GetInt("DB_PORT"),
		User:            viper.GetString("DB_USER"),
		Password:        viper.GetString("DB_PASSWORD"),
		DBName:          viper.GetString("DB_NAME"),
		SSLMode:         viper.GetString("DB_SSLMODE"),
		MaxOpenConns:    viper.GetInt("DB_MAX_CONNS"),
		MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetInt("DB_CONN_MAX_LIFETIME"),
	}

	// 解析Redis配置
	config.Redis = RedisConfig{
		Host:       viper.GetString("REDIS_HOST"),
		Port:       viper.GetInt("REDIS_PORT"),
		Password:   viper.GetString("REDIS_PASSWORD"),
		DB:         viper.GetInt("REDIS_DB"),
		PoolSize:   viper.GetInt("REDIS_POOL_SIZE"),
		MaxRetries: viper.GetInt("REDIS_MAX_RETRIES"),
	}

	// 解析Linux Do配置
	config.LinuxDo = LinuxDoConfig{
		ClientID:     viper.GetString("LINUX_DO_CLIENT_ID"),
		ClientSecret: viper.GetString("LINUX_DO_CLIENT_SECRET"),
		RedirectURI:  viper.GetString("LINUX_DO_REDIRECT_URI"),
		AuthURL:      viper.GetString("LINUX_DO_AUTH_URL"),
		TokenURL:     viper.GetString("LINUX_DO_TOKEN_URL"),
		UserInfoURL:  viper.GetString("LINUX_DO_USER_INFO_URL"),
	}

	// 解析公益站配置
	config.Kyx = KyxConfig{
		APIBase:             viper.GetString("KYX_API_BASE"),
		ModelScopeAPIBase:   viper.GetString("MODELSCOPE_API_BASE"),
		DefaultClaimQuota:   viper.GetInt64("DEFAULT_CLAIM_QUOTA"),
		DonateQuotaPerKey:   viper.GetInt64("DONATE_QUOTA_PER_KEY"),
		MinQuotaThreshold:   viper.GetInt64("MIN_QUOTA_THRESHOLD"),
		MaxDonateKeysPerDay: viper.GetInt("MAX_DONATE_KEYS_PER_DAY"),
		FirstBindBonusQuota: viper.GetInt64("FIRST_BIND_BONUS_QUOTA"),
	}

	// 解析管理员配置
	config.Admin = AdminConfig{
		Password:      viper.GetString("ADMIN_PASSWORD"),
		JWTSecret:     viper.GetString("JWT_SECRET"),
		SessionExpire: viper.GetInt("SESSION_EXPIRE_HOURS"),
	}

	// 解析日志配置
	config.Log = LogConfig{
		Level:  viper.GetString("LOG_LEVEL"),
		Format: viper.GetString("LOG_FORMAT"),
		Output: viper.GetString("LOG_OUTPUT"),
		Path:   viper.GetString("LOG_PATH"),
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	globalConfig = config
	return config, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// 服务器默认值
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "release")
	viper.SetDefault("SERVER_READ_TIMEOUT", 30)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 30)
	viper.SetDefault("ENABLE_CORS", true)
	viper.SetDefault("CORS_ORIGINS", []string{"*"})

	// 数据库默认值
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "kyxuser")
	viper.SetDefault("DB_NAME", "kyxquota")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_CONNS", 100)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 30)

	// Redis默认值
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_POOL_SIZE", 50)
	viper.SetDefault("REDIS_MAX_RETRIES", 3)

	// Linux Do默认值
	viper.SetDefault("LINUX_DO_AUTH_URL", "https://connect.linux.do/oauth2/authorize")
	viper.SetDefault("LINUX_DO_TOKEN_URL", "https://connect.linux.do/oauth2/token")
	viper.SetDefault("LINUX_DO_USER_INFO_URL", "https://connect.linux.do/api/user")

	// 公益站默认值
	viper.SetDefault("KYX_API_BASE", "https://api.kkyyxx.xyz")
	viper.SetDefault("MODELSCOPE_API_BASE", "https://api-inference.modelscope.cn/v1")
	viper.SetDefault("DEFAULT_CLAIM_QUOTA", 20000000)
	viper.SetDefault("DONATE_QUOTA_PER_KEY", 25000000)
	viper.SetDefault("MIN_QUOTA_THRESHOLD", 10000000)
	viper.SetDefault("MAX_DONATE_KEYS_PER_DAY", 5)
	viper.SetDefault("FIRST_BIND_BONUS_QUOTA", 50000000)

	// 管理员默认值
	viper.SetDefault("ADMIN_PASSWORD", "admin123")
	viper.SetDefault("JWT_SECRET", "your-secret-key-please-change-in-production")
	viper.SetDefault("SESSION_EXPIRE_HOURS", 168) // 7 days

	// 日志默认值
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_OUTPUT", "stdout")
	viper.SetDefault("LOG_PATH", "./logs/app.log")
}

// bindEnvVariables 绑定环境变量
func bindEnvVariables() {
	// 服务器
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("SERVER_MODE")
	viper.BindEnv("SERVER_READ_TIMEOUT")
	viper.BindEnv("SERVER_WRITE_TIMEOUT")
	viper.BindEnv("ENABLE_CORS")
	viper.BindEnv("CORS_ORIGINS")

	// 数据库
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_SSLMODE")
	viper.BindEnv("DB_MAX_CONNS")
	viper.BindEnv("DB_MAX_IDLE_CONNS")
	viper.BindEnv("DB_CONN_MAX_LIFETIME")

	// Redis
	viper.BindEnv("REDIS_HOST")
	viper.BindEnv("REDIS_PORT")
	viper.BindEnv("REDIS_PASSWORD")
	viper.BindEnv("REDIS_DB")
	viper.BindEnv("REDIS_POOL_SIZE")
	viper.BindEnv("REDIS_MAX_RETRIES")

	// Linux Do
	viper.BindEnv("LINUX_DO_CLIENT_ID")
	viper.BindEnv("LINUX_DO_CLIENT_SECRET")
	viper.BindEnv("LINUX_DO_REDIRECT_URI")
	viper.BindEnv("LINUX_DO_AUTH_URL")
	viper.BindEnv("LINUX_DO_TOKEN_URL")
	viper.BindEnv("LINUX_DO_USER_INFO_URL")

	// 公益站
	viper.BindEnv("KYX_API_BASE")
	viper.BindEnv("MODELSCOPE_API_BASE")
	viper.BindEnv("DEFAULT_CLAIM_QUOTA")
	viper.BindEnv("DONATE_QUOTA_PER_KEY")
	viper.BindEnv("MIN_QUOTA_THRESHOLD")
	viper.BindEnv("MAX_DONATE_KEYS_PER_DAY")
	viper.BindEnv("FIRST_BIND_BONUS_QUOTA")

	// 管理员
	viper.BindEnv("ADMIN_PASSWORD")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("SESSION_EXPIRE_HOURS")

	// 日志
	viper.BindEnv("LOG_LEVEL")
	viper.BindEnv("LOG_FORMAT")
	viper.BindEnv("LOG_OUTPUT")
	viper.BindEnv("LOG_PATH")
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证必填项
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Database.Host == "" || c.Database.User == "" || c.Database.DBName == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	if c.Database.Password == "" {
		fmt.Println("Warning: database password is empty")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if c.LinuxDo.ClientID == "" || c.LinuxDo.ClientSecret == "" {
		return fmt.Errorf("Linux Do OAuth2 credentials are required")
	}

	if c.LinuxDo.RedirectURI == "" {
		return fmt.Errorf("Linux Do redirect URI is required")
	}

	if c.Admin.Password == "admin123" {
		fmt.Println("Warning: using default admin password, please change it in production")
	}

	if c.Admin.JWTSecret == "your-secret-key-please-change-in-production" {
		fmt.Println("Warning: using default JWT secret, please change it in production")
	}

	// 验证服务器模式
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("invalid server mode: %s (must be 'debug' or 'release')", c.Server.Mode)
	}

	// 验证日志级别
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}

	return nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, please call Load() first")
	}
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment 是否开发模式
func (c *ServerConfig) IsDevelopment() bool {
	return c.Mode == "debug"
}

// IsProduction 是否生产模式
func (c *ServerConfig) IsProduction() bool {
	return c.Mode == "release"
}

// QuotaToDollar 额度转美元
func QuotaToDollar(quota int64) float64 {
	return float64(quota) / 500000.0
}

// DollarToQuota 美元转额度
func DollarToQuota(dollar float64) int64 {
	return int64(dollar * 500000)
}
