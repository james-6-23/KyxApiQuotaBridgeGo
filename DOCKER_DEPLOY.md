# Docker 部署指南

本文档介绍如何使用 Docker 和 Docker Compose 部署 KYX API Quota Bridge 项目。

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- Git

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/james-6-23/KyxApiQuotaBridgeGo.git
cd KyxApiQuotaBridgeGo
```

### 2. 配置环境变量

复制示例环境文件并编辑：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置必要的环境变量：

```env
# 应用端口
APP_PORT=8080

# 管理员账户
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_secure_password

# JWT 密钥（请生成一个强随机字符串）
JWT_SECRET=your_jwt_secret_here

# Claude API
CLAUDE_API_URL=https://api.anthropic.com

# OAuth 配置（可选）
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

### 3. 使用 Docker Compose 启动

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止服务并删除数据卷
docker-compose down -v
```

### 4. 访问应用

应用启动后，访问：
- 前端界面：http://localhost:8080
- API 文档：http://localhost:8080/api/docs

## 单独使用 Docker

### 构建镜像

```bash
docker build -t kyx-api-quota-bridge:latest .
```

### 运行容器

```bash
docker run -d \
  --name kyx-api-quota-bridge \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e DATABASE_PATH=/app/data/data.db \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=admin123 \
  -e JWT_SECRET=your_jwt_secret \
  kyx-api-quota-bridge:latest
```

## GitHub Actions 自动构建

项目配置了 GitHub Actions，可以自动构建和推送 Docker 镜像到 GitHub Container Registry。

### 工作流说明

1. **CI 工作流** (`.github/workflows/ci.yml`)
   - 自动运行测试
   - 执行代码质量检查
   - 验证 Docker 构建

2. **Docker 构建工作流** (`.github/workflows/docker-build.yml`)
   - 推送到 `main` 分支时自动构建
   - 支持多平台（amd64, arm64）
   - 自动推送到 GHCR

### 使用 GitHub Container Registry 镜像

```bash
# 拉取镜像
docker pull ghcr.io/james-6-23/kyxapiquotabridgego:latest

# 运行
docker run -d \
  --name kyx-api-quota-bridge \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ghcr.io/james-6-23/kyxapiquotabridgego:latest
```

## 数据持久化

数据库文件默认存储在 `./data` 目录下。确保该目录有适当的权限：

```bash
mkdir -p data
chmod 755 data
```

## 健康检查

容器包含健康检查端点：

```bash
# 检查容器健康状态
docker inspect --format='{{.State.Health.Status}}' kyx-api-quota-bridge

# 手动测试健康端点
curl http://localhost:8080/api/health
```

## 日志管理

查看容器日志：

```bash
# 实时查看日志
docker-compose logs -f app

# 查看最近 100 行日志
docker-compose logs --tail=100 app
```

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build

# 或使用预构建镜像
docker-compose pull
docker-compose up -d
```

## 故障排查

### 容器无法启动

1. 检查环境变量是否正确设置
2. 检查端口 8080 是否被占用
3. 查看容器日志：`docker-compose logs app`

### 数据库问题

1. 确保 `data` 目录存在且有写入权限
2. 检查 `DATABASE_PATH` 环境变量

### 前端无法访问

1. 确认前端文件已正确构建到 `web` 目录
2. 检查后端是否正确提供静态文件服务

## 生产环境建议

1. **安全性**
   - 修改默认管理员密码
   - 使用强 JWT 密钥
   - 配置 HTTPS（建议使用 Nginx 或 Traefik 反向代理）
   - 限制数据库文件访问权限

2. **性能**
   - 根据负载调整容器资源限制
   - 配置日志轮转
   - 使用专用数据卷而非 bind mount

3. **监控**
   - 配置健康检查告警
   - 集成日志聚合系统
   - 监控容器资源使用

## 参考链接

- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [项目 README](./README.md)
