package main

import (
	"log"
	"os"

	"github.com/kyx-api-quota-bridge/internal/api"
	"github.com/kyx-api-quota-bridge/internal/config"
	"github.com/kyx-api-quota-bridge/internal/store"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := store.NewDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 创建 API 服务器
	server := api.NewServer(cfg, db)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
