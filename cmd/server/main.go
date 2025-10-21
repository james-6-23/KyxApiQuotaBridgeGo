package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/kyx-quota-bridge/internal/config"
	"github.com/yourusername/kyx-quota-bridge/internal/handler"
	"github.com/yourusername/kyx-quota-bridge/internal/middleware"
	"github.com/yourusername/kyx-quota-bridge/internal/repository"
	"github.com/yourusername/kyx-quota-bridge/internal/service"
	"github.com/yourusername/kyx-quota-bridge/pkg/cache"
	"github.com/yourusername/kyx-quota-bridge/pkg/database"
)

var (
	// Version 版本号 (可通过 -ldflags 注入)
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 1. 初始化日志
	logger := initLogger()
	logger.WithFields(logrus.Fields{
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
	}).Info("Starting Kyx Quota Bridge")

	// 2. 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}
	logger.Info("Configuration loaded successfully")

	// 根据配置设置日志级别
	setLogLevel(logger, cfg.Log.Level)

	// 3. 连接数据库
	db, err := database.New(&database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute,
	}, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()
	logger.Info("Database connected successfully")

	// 4. 连接Redis
	redisClient, err := cache.New(&cache.Config{
		Host:       cfg.Redis.Host,
		Port:       cfg.Redis.Port,
		Password:   cfg.Redis.Password,
		DB:         cfg.Redis.DB,
		PoolSize:   cfg.Redis.PoolSize,
		MaxRetries: cfg.Redis.MaxRetries,
	}, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()
	logger.Info("Redis connected successfully")

	// 5. 初始化仓库层
	userRepo := repository.NewUserRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, redisClient, logger)
	claimRepo := repository.NewClaimRepository(db, logger)
	donateRepo := repository.NewDonateRepository(db, logger)
	keyRepo := repository.NewKeyRepository(db, logger)
	adminConfigRepo := repository.NewAdminConfigRepository(db, redisClient, logger)
	logger.Info("Repositories initialized")

	// 6. 初始化服务层
	cacheService := service.NewCacheService(redisClient, logger)

	// KyxClient
	kyxClient := service.NewKyxClient(service.KyxClientConfig{
		BaseURL: cfg.Kyx.APIBase,
		Session: "", // 从数据库加载
		Timeout: 30 * time.Second,
	}, logger)

	// LinuxDoClient
	linuxDoClient := service.NewLinuxDoClient(service.LinuxDoClientConfig{
		ClientID:     cfg.LinuxDo.ClientID,
		ClientSecret: cfg.LinuxDo.ClientSecret,
		RedirectURI:  cfg.LinuxDo.RedirectURI,
		BaseURL:      "https://connect.linux.do",
		Timeout:      30 * time.Second,
	}, logger)

	// AuthService
	authService := service.NewAuthService(
		linuxDoClient,
		sessionRepo,
		userRepo,
		cacheService,
		service.AuthServiceConfig{
			JWTSecret:      cfg.Admin.JWTSecret,
			AdminPassword:  cfg.Admin.Password,
			SessionTimeout: time.Duration(cfg.Admin.SessionExpire) * time.Hour,
		},
		logger,
	)

	// UserService
	userService := service.NewUserService(
		userRepo,
		claimRepo,
		donateRepo,
		adminConfigRepo,
		kyxClient,
		linuxDoClient,
		cacheService,
		logger,
	)

	// QuotaService
	quotaService := service.NewQuotaService(
		claimRepo,
		userRepo,
		adminConfigRepo,
		kyxClient,
		cacheService,
		logger,
	)

	// DonateService
	donateService := service.NewDonateService(
		donateRepo,
		keyRepo,
		userRepo,
		adminConfigRepo,
		kyxClient,
		cacheService,
		logger,
	)

	// AdminService
	adminService := service.NewAdminService(
		adminConfigRepo,
		userRepo,
		claimRepo,
		donateRepo,
		keyRepo,
		sessionRepo,
		kyxClient,
		cacheService,
		logger,
	)

	logger.Info("Services initialized")

	// 7. 初始化处理器层
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, quotaService, donateService, logger)
	adminHandler := handler.NewAdminHandler(adminService, userService, quotaService, donateService, logger)
	logger.Info("Handlers initialized")

	// 8. 初始化中间件
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	corsMiddleware := middleware.DefaultCORS(logger)
	loggerMiddleware := middleware.NewLoggerMiddleware(logger)
	recoveryMiddleware := middleware.DefaultRecovery(logger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(cacheService, logger)
	logger.Info("Middlewares initialized")

	// 9. 初始化管理员配置（如果不存在）
	if err := adminService.InitializeDefaultConfig(context.Background()); err != nil {
		logger.WithError(err).Warn("Failed to initialize default config")
	}

	// 10. 设置Gin模式
	if cfg.Server.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 11. 创建路由
	router := setupRouter(
		cfg,
		logger,
		authHandler,
		userHandler,
		adminHandler,
		authMiddleware,
		corsMiddleware,
		loggerMiddleware,
		recoveryMiddleware,
		rateLimitMiddleware,
	)

	// 12. 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 13. 启动服务器（在goroutine中）
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// 14. 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 15. 优雅关闭，超时时间30秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited successfully")
}

// initLogger 初始化日志
func initLogger() *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	logger.SetLevel(logrus.InfoLevel)

	// 设置日志格式
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// 设置输出
	logger.SetOutput(os.Stdout)

	return logger
}

// setLogLevel 设置日志级别
func setLogLevel(logger *logrus.Logger, level string) {
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

// setupRouter 设置路由
func setupRouter(
	cfg *config.Config,
	logger *logrus.Logger,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	adminHandler *handler.AdminHandler,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	loggerMiddleware *middleware.LoggerMiddleware,
	recoveryMiddleware *middleware.RecoveryMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) *gin.Engine {
	router := gin.New()

	// 全局中间件
	router.Use(recoveryMiddleware.Handler())
	router.Use(loggerMiddleware.Handler())
	router.Use(corsMiddleware.Handler())

	// 健康检查（无认证）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"version":    Version,
			"build_time": BuildTime,
			"timestamp":  time.Now().Unix(),
		})
	})

	// 版本信息
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    Version,
			"build_time": BuildTime,
			"git_commit": GitCommit,
		})
	})

	// API 路由组
	api := router.Group("/api")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.GET("/url", authHandler.GetAuthURL)
			auth.GET("/callback", authHandler.HandleCallback)
			auth.GET("/check", authHandler.CheckAuth)
			auth.POST("/admin/login", rateLimitMiddleware.LoginRateLimit(), authHandler.AdminLogin)
		}

		// 需要用户认证的路由
		authenticated := api.Group("")
		authenticated.Use(authMiddleware.RequireAuth())
		{
			// 认证相关
			authenticated.POST("/auth/logout", authHandler.Logout)
			authenticated.POST("/auth/refresh", authHandler.RefreshSession)
			authenticated.GET("/auth/me", authHandler.GetCurrentUser)

			// 用户相关
			authenticated.POST("/user/bind", userHandler.BindAccount)
			authenticated.GET("/user/bind/status", userHandler.CheckBindStatus)
			authenticated.GET("/user/quota", userHandler.GetQuota)
			authenticated.GET("/user/profile", userHandler.GetProfile)
			authenticated.GET("/user/statistics", userHandler.GetStatistics)

			// 领取记录
			authenticated.GET("/user/claims", userHandler.GetClaimHistory)
			authenticated.POST("/user/claim", userHandler.ClaimQuota)

			// 投喂记录
			authenticated.GET("/user/donates", userHandler.GetDonateHistory)
			authenticated.POST("/user/donate", rateLimitMiddleware.DonateRateLimit(), userHandler.DonateKeys)
		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin())
		{
			// 配置管理
			admin.GET("/config", adminHandler.GetConfig)
			admin.PUT("/config", adminHandler.UpdateConfig)

			// 统计信息
			admin.GET("/stats", adminHandler.GetSystemStats)
			admin.GET("/dashboard", adminHandler.GetDashboard)

			// 用户管理
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/statistics", adminHandler.GetAllStatistics)
			admin.DELETE("/users/:linux_do_id", adminHandler.DeleteUser)

			// 记录管理
			admin.GET("/claims", adminHandler.ListAllClaims)
			admin.GET("/donates", adminHandler.ListAllDonates)
			admin.GET("/activity", adminHandler.GetRecentActivity)

			// 维护操作
			admin.POST("/maintenance/sessions", adminHandler.CleanExpiredSessions)
			admin.POST("/maintenance/keys", adminHandler.CleanOldKeys)
			admin.POST("/cache/clear", adminHandler.ClearCache)

			// 测试工具
			admin.GET("/test/kyx", adminHandler.TestKyxConnection)
			admin.GET("/test/session", adminHandler.ValidateKyxSession)

			// 健康状态
			admin.GET("/health", adminHandler.GetHealthStatus)

			// 数据导出
			admin.GET("/export", adminHandler.ExportData)
		}
	}

	// 前端静态文件服务
	// 1. 提供静态资源（CSS, JS, 图片等）
	router.Static("/assets", "./web/assets")

	// 2. 提供 favicon 和其他根目录文件
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// 3. 为所有非 API 路由返回 index.html（支持前端路由）
	router.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回 404 JSON
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "API endpoint not found",
				"path":    c.Request.URL.Path,
			})
			return
		}

		// 否则返回前端 index.html，让前端路由处理
		c.File("./web/index.html")
	})

	// 打印所有注册的路由（仅在debug模式）
	if cfg.Server.IsDevelopment() {
		routes := router.Routes()
		logger.WithField("count", len(routes)).Info("Registered routes:")
		for _, route := range routes {
			logger.Debugf("  %s %s", route.Method, route.Path)
		}
	}

	return router
}
