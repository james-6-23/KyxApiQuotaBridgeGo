package config

import (
    "fmt"
    "os"
    "time"

    "github.com/caarlos0/env/v10"
)

type Config struct {
    Environment string        `env:"APP_ENV" envDefault:"development"`
    HTTPPort     int           `env:"HTTP_PORT" envDefault:"8080"`
    ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`

    Postgres PostgresConfig
    OAuth    OAuthConfig
    KYX      KYXConfig
    ModelScope ModelScopeConfig
    Cache    CacheConfig
}

type PostgresConfig struct {
    DSN             string        `env:"PG_DSN,required"`
    MaxOpenConns    int           `env:"PG_MAX_OPEN_CONNS" envDefault:"20"`
    MaxIdleConns    int           `env:"PG_MAX_IDLE_CONNS" envDefault:"5"`
    ConnMaxLifetime time.Duration `env:"PG_CONN_MAX_LIFETIME" envDefault:"1h"`
}

type OAuthConfig struct {
    ClientID     string `env:"LINUXDO_CLIENT_ID,required"`
    ClientSecret string `env:"LINUXDO_CLIENT_SECRET,required"`
    RedirectURI  string `env:"LINUXDO_REDIRECT_URI,required"`
    AuthURL      string `env:"LINUXDO_AUTH_URL" envDefault:"https://connect.linux.do/oauth2/authorize"`
    TokenURL     string `env:"LINUXDO_TOKEN_URL" envDefault:"https://connect.linux.do/oauth2/token"`
    UserInfoURL  string `env:"LINUXDO_USERINFO_URL" envDefault:"https://connect.linux.do/api/user"`
}

type KYXConfig struct {
    APIBaseURL string `env:"KYX_API_BASE" envDefault:"https://api.kkyyxx.xyz"`
    Session    string `env:"KYX_SESSION"`
    NewAPIUser string `env:"KYX_NEW_API_USER" envDefault:"1"`
    KeysAPIURL string `env:"KYX_KEYS_API_URL" envDefault:"https://gpt-load.kyx03.de/api/keys/add-async"`
    KeysAuthorization string `env:"KYX_KEYS_AUTHORIZATION"`
    GroupID    int    `env:"KYX_GROUP_ID" envDefault:"26"`
    DefaultClaimQuota int64 `env:"KYX_DEFAULT_CLAIM_QUOTA" envDefault:"20000000"`
    DonateQuotaPerKey int64 `env:"KYX_DONATE_QUOTA_PER_KEY" envDefault:"100000000"`
    MinQuotaThreshold int64 `env:"KYX_MIN_QUOTA_THRESHOLD" envDefault:"10000000"`
    FirstBindBonus    int64 `env:"KYX_FIRST_BIND_BONUS" envDefault:"100000000"`
}

type ModelScopeConfig struct {
    APIBaseURL string `env:"MODELSCOPE_API_BASE" envDefault:"https://api-inference.modelscope.cn/v1"`
}

type CacheConfig struct {
    Enabled bool          `env:"CACHE_ENABLED" envDefault:"false"`
    RedisURL string       `env:"REDIS_URL"`
    TTL      time.Duration `env:"CACHE_TTL" envDefault:"5m"`
}

func Provide() (*Config, error) {
    var cfg Config
    opts := env.Options{OnSet: func(tag string, value any, isDefault bool) {
        _ = tag
        _ = value
        _ = isDefault
    }}
    if err := env.ParseWithOptions(&cfg, opts); err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }

    if cfg.Postgres.DSN == "" {
        return nil, fmt.Errorf("PG_DSN must be set")
    }
    if cfg.OAuth.ClientID == "" || cfg.OAuth.ClientSecret == "" || cfg.OAuth.RedirectURI == "" {
        return nil, fmt.Errorf("LinuxDo OAuth credentials are required")
    }
    return &cfg, nil
}

func (c *Config) IsProduction() bool {
    return c.Environment == "production"
}

func MustProvide() *Config {
    cfg, err := Provide()
    if err != nil {
        panic(err)
    }
    return cfg
}

func init() {
    _ = os.Setenv("TZ", "UTC")
}
