# ========================================
# KyxApiQuotaBridge Makefile
# ========================================
# Go项目构建和部署自动化脚本
# ========================================

# 项目配置
PROJECT_NAME := kyx-quota-bridge
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 配置
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# 目录配置
CMD_DIR := ./cmd/server
BUILD_DIR := ./build
BIN_DIR := $(BUILD_DIR)/bin
COVERAGE_DIR := $(BUILD_DIR)/coverage

# 二进制文件
BINARY_NAME := $(PROJECT_NAME)
BINARY_UNIX := $(BINARY_NAME)_unix
BINARY_WINDOWS := $(BINARY_NAME).exe

# Docker 配置
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)
DOCKER_COMPOSE := docker-compose

# 链接标志
LDFLAGS := -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GitCommit=$(GIT_COMMIT) \
	-w -s"

# 颜色输出
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# ========================================
# 默认目标
# ========================================

.PHONY: all
all: clean deps build test ## 执行完整构建流程（清理、依赖、构建、测试）

.DEFAULT_GOAL := help

# ========================================
# 帮助信息
# ========================================

.PHONY: help
help: ## 显示帮助信息
	@echo "$(BLUE)KyxApiQuotaBridge Makefile 命令列表$(NC)"
	@echo "========================================"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-18s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""

# ========================================
##@ 开发环境
# ========================================

.PHONY: init
init: ## 初始化开发环境
	@echo "$(BLUE)初始化开发环境...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@mkdir -p $(BUILD_DIR) $(BIN_DIR) $(COVERAGE_DIR)
	@mkdir -p logs data/postgres data/redis backups
	@cp -n .env.example .env 2>/dev/null || true
	@echo "$(GREEN)✓ 开发环境初始化完成$(NC)"

.PHONY: deps
deps: ## 下载Go依赖
	@echo "$(BLUE)下载依赖...$(NC)"
	@$(GOGET) -v ./...
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)✓ 依赖下载完成$(NC)"

.PHONY: deps-update
deps-update: ## 更新所有依赖
	@echo "$(BLUE)更新依赖...$(NC)"
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)✓ 依赖更新完成$(NC)"

.PHONY: deps-verify
deps-verify: ## 验证依赖
	@echo "$(BLUE)验证依赖...$(NC)"
	@$(GOMOD) verify
	@echo "$(GREEN)✓ 依赖验证通过$(NC)"

# ========================================
##@ 构建
# ========================================

.PHONY: build
build: ## 构建应用（当前平台）
	@echo "$(BLUE)构建应用...$(NC)"
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(GREEN)✓ 构建完成: $(BIN_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-linux
build-linux: ## 构建Linux版本
	@echo "$(BLUE)构建Linux版本...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_UNIX) $(CMD_DIR)
	@echo "$(GREEN)✓ Linux版本构建完成$(NC)"

.PHONY: build-windows
build-windows: ## 构建Windows版本
	@echo "$(BLUE)构建Windows版本...$(NC)"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_WINDOWS) $(CMD_DIR)
	@echo "$(GREEN)✓ Windows版本构建完成$(NC)"

.PHONY: build-all
build-all: build-linux build-windows ## 构建所有平台版本
	@echo "$(GREEN)✓ 所有平台构建完成$(NC)"

.PHONY: install
install: ## 安装到 $GOPATH/bin
	@echo "$(BLUE)安装应用...$(NC)"
	@$(GOCMD) install $(LDFLAGS) $(CMD_DIR)
	@echo "$(GREEN)✓ 安装完成$(NC)"

# ========================================
##@ 运行和开发
# ========================================

.PHONY: run
run: ## 运行应用
	@echo "$(BLUE)启动应用...$(NC)"
	@$(GOCMD) run $(CMD_DIR)/main.go

.PHONY: dev
dev: ## 开发模式运行（带热重载）
	@echo "$(BLUE)开发模式启动...$(NC)"
	@which air > /dev/null || $(GOGET) -u github.com/cosmtrek/air
	@air

.PHONY: watch
watch: dev ## 同 dev，监听文件变化自动重启

# ========================================
##@ 测试
# ========================================

.PHONY: test
test: ## 运行所有测试
	@echo "$(BLUE)运行测试...$(NC)"
	@$(GOTEST) -v -race ./...
	@echo "$(GREEN)✓ 测试完成$(NC)"

.PHONY: test-short
test-short: ## 运行短测试（跳过耗时测试）
	@echo "$(BLUE)运行短测试...$(NC)"
	@$(GOTEST) -v -short ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)生成测试覆盖率...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@echo "$(GREEN)✓ 覆盖率报告: $(COVERAGE_DIR)/coverage.html$(NC)"

.PHONY: test-integration
test-integration: ## 运行集成测试
	@echo "$(BLUE)运行集成测试...$(NC)"
	@$(GOTEST) -v -tags=integration ./...

.PHONY: bench
bench: ## 运行性能测试
	@echo "$(BLUE)运行性能测试...$(NC)"
	@$(GOTEST) -bench=. -benchmem ./...

# ========================================
##@ 代码质量
# ========================================

.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(NC)"
	@$(GOFMT) -s -w .
	@echo "$(GREEN)✓ 代码格式化完成$(NC)"

.PHONY: fmt-check
fmt-check: ## 检查代码格式
	@echo "$(BLUE)检查代码格式...$(NC)"
	@test -z "$$($(GOFMT) -l .)" || (echo "$(RED)代码格式不正确，请运行 make fmt$(NC)" && exit 1)
	@echo "$(GREEN)✓ 代码格式正确$(NC)"

.PHONY: lint
lint: ## 运行代码检查
	@echo "$(BLUE)运行代码检查...$(NC)"
	@which $(GOLINT) > /dev/null || (echo "$(YELLOW)安装 golangci-lint...$(NC)" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin)
	@$(GOLINT) run ./...
	@echo "$(GREEN)✓ 代码检查完成$(NC)"

.PHONY: vet
vet: ## 运行 go vet
	@echo "$(BLUE)运行 go vet...$(NC)"
	@$(GOCMD) vet ./...
	@echo "$(GREEN)✓ go vet 完成$(NC)"

.PHONY: check
check: fmt-check vet lint test ## 运行所有检查（格式、vet、lint、测试）
	@echo "$(GREEN)✓ 所有检查通过$(NC)"

# ========================================
##@ Docker
# ========================================

.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "$(BLUE)构建Docker镜像...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile .
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)✓ Docker镜像构建完成$(NC)"

.PHONY: docker-up
docker-up: ## 启动Docker服务
	@echo "$(BLUE)启动Docker服务...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✓ 服务已启动$(NC)"
	@$(DOCKER_COMPOSE) ps

.PHONY: docker-down
docker-down: ## 停止Docker服务
	@echo "$(BLUE)停止Docker服务...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✓ 服务已停止$(NC)"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## 重启Docker服务

.PHONY: docker-logs
docker-logs: ## 查看Docker日志
	@$(DOCKER_COMPOSE) logs -f

.PHONY: docker-ps
docker-ps: ## 查看Docker容器状态
	@$(DOCKER_COMPOSE) ps

.PHONY: docker-clean
docker-clean: ## 清理Docker资源
	@echo "$(YELLOW)清理Docker资源...$(NC)"
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@docker system prune -f
	@echo "$(GREEN)✓ Docker资源清理完成$(NC)"

# ========================================
##@ 数据库
# ========================================

.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	@echo "$(BLUE)运行数据库迁移...$(NC)"
	@$(DOCKER_COMPOSE) exec -T postgres psql -U kyxuser kyxquota < migrations/001_init.sql
	@echo "$(GREEN)✓ 数据库迁移完成$(NC)"

.PHONY: db-reset
db-reset: ## 重置数据库（危险操作）
	@echo "$(RED)警告: 这将删除所有数据！$(NC)"
	@read -p "确认继续? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(BLUE)重置数据库...$(NC)"; \
		$(DOCKER_COMPOSE) exec postgres psql -U kyxuser -c "DROP DATABASE IF EXISTS kyxquota;"; \
		$(DOCKER_COMPOSE) exec postgres psql -U kyxuser -c "CREATE DATABASE kyxquota;"; \
		$(MAKE) db-migrate; \
		echo "$(GREEN)✓ 数据库重置完成$(NC)"; \
	fi

.PHONY: db-backup
db-backup: ## 备份数据库
	@echo "$(BLUE)备份数据库...$(NC)"
	@mkdir -p backups
	@$(DOCKER_COMPOSE) exec -T postgres pg_dump -U kyxuser kyxquota > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✓ 数据库备份完成$(NC)"

.PHONY: db-shell
db-shell: ## 进入数据库Shell
	@$(DOCKER_COMPOSE) exec postgres psql -U kyxuser kyxquota

# ========================================
##@ 清理
# ========================================

.PHONY: clean
clean: ## 清理构建文件
	@echo "$(BLUE)清理构建文件...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME) $(BINARY_UNIX) $(BINARY_WINDOWS)
	@echo "$(GREEN)✓ 清理完成$(NC)"

.PHONY: clean-all
clean-all: clean docker-clean ## 清理所有文件（包括Docker）
	@echo "$(BLUE)清理所有文件...$(NC)"
	@rm -rf logs/* data/postgres/* data/redis/*
	@echo "$(GREEN)✓ 所有文件清理完成$(NC)"

# ========================================
##@ 发布
# ========================================

.PHONY: release
release: clean check build-all ## 创建发布版本
	@echo "$(BLUE)创建发布版本 $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@tar -czf $(BUILD_DIR)/release/$(PROJECT_NAME)_$(VERSION)_linux_amd64.tar.gz -C $(BIN_DIR) $(BINARY_UNIX)
	@zip -j $(BUILD_DIR)/release/$(PROJECT_NAME)_$(VERSION)_windows_amd64.zip $(BIN_DIR)/$(BINARY_WINDOWS)
	@echo "$(GREEN)✓ 发布版本创建完成: $(BUILD_DIR)/release/$(NC)"
	@ls -lh $(BUILD_DIR)/release/

.PHONY: tag
tag: ## 创建Git标签
	@echo "当前版本: $(VERSION)"
	@read -p "输入新版本号: " version; \
	git tag -a v$$version -m "Release v$$version"; \
	git push origin v$$version; \
	echo "$(GREEN)✓ 标签 v$$version 已创建并推送$(NC)"

# ========================================
##@ 工具
# ========================================

.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "$(BLUE)安装开发工具...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ 开发工具安装完成$(NC)"

.PHONY: version
version: ## 显示版本信息
	@echo "项目: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

.PHONY: info
info: ## 显示项目信息
	@echo "$(BLUE)项目信息$(NC)"
	@echo "========================================"
	@echo "项目名称: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go版本: $$(go version)"
	@echo "Git分支: $$(git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "构建目录: $(BUILD_DIR)"
	@echo "========================================"

# ========================================
# 特殊目标
# ========================================

.PHONY: .FORCE
.FORCE:

# 创建必要的目录
$(BUILD_DIR) $(BIN_DIR) $(COVERAGE_DIR):
	@mkdir -p $@
