# KyxApiQuotaBridge

> 公益站额度自助领取系统 - 基于 Go + Vue 3 的前后端分离架构

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Vue Version](https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📖 项目简介

一个基于 Go 语言和 Vue 3 开发的额度管理桥接服务，用于连接 Linux.do 论坛和公益站 API，提供用户认证、额度领取、Key 投喂等功能。

### ✨ 核心特性

#### 用户功能
- ✅ **Linux.do OAuth 登录** - 快速安全的第三方登录
- ✅ **账号绑定** - 自动绑定 Linux.do 账号与公益站账号
- ✅ **每日领取** - 每天领取固定额度奖励
- ✅ **Keys 投喂** - 投喂 ModelScope Keys 获得额度奖励
- ✅ **历史记录** - 查看领取和投喂历史

#### 管理功能
- ✅ **系统配置** - 动态配置领取额度、Session 等
- ✅ **用户管理** - 查看和管理所有用户
- ✅ **Keys 管理** - 批量管理投喂的 Keys
- ✅ **数据统计** - 实时统计和数据分析
- ✅ **记录查询** - 查询所有领取和投喂记录

#### 技术亮点
- ✅ **前后端一体化部署** - 单个 Docker 镜像包含前后端
- ✅ **Redis 缓存** - 高性能缓存优化
- ✅ **多维限流** - 防止接口滥用
- ✅ **健康检查** - 完善的服务监控
- ✅ **优雅停机** - 平滑关闭和信号处理
- ✅ **自动备份** - 可选的数据库自动备份
- ✅ **CI/CD** - GitHub Actions 自动构建部署

---

## 🛠️ 技术栈

### 后端
| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 编程语言 |
| Gin | Latest | Web 框架 |
| PostgreSQL | 15+ | 主数据库 |
| Redis | 7+ | 缓存数据库 |
| Logrus | Latest | 日志库 |

### 前端
| 技术 | 版本 | 说明 |
|------|------|------|
| Vue 3 | 3.4+ | 前端框架 |
| Vite | 5+ | 构建工具 |
| TypeScript | 5+ | 类型支持 |
| Ant Design Vue | 4+ | UI 组件库 |
| Tailwind CSS | 3+ | CSS 框架 |
| Pinia | 2+ | 状态管理 |

### 部署
| 技术 | 说明 |
|------|------|
| Docker | 容器化 |
| Docker Compose | 服务编排 |
| GitHub Actions | CI/CD |
| Nginx | 反向代理（可选） |

---

## 📁 项目结构

```
KyxApiQuotaBridgeGo/
├── cmd/
│   └── server/
│       └── main.go              # 后端入口
├── internal/
│   ├── config/                  # 配置管理
│   ├── handler/                 # HTTP 处理器
│   ├── middleware/              # 中间件
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   └── service/                 # 业务逻辑层
├── pkg/
│   ├── cache/                   # Redis 封装
│   └── database/                # 数据库封装
├── frontend/                    # Vue 3 前端
│   ├── src/
│   │   ├── api/                # API 接口
│   │   ├── views/              # 页面组件
│   │   ├── stores/             # 状态管理
│   │   └── router/             # 路由配置
│   ├── package.json
│   └── vite.config.ts
├── migrations/                  # 数据库迁移
├── .github/
│   └── workflows/               # GitHub Actions
├── Dockerfile                   # 前后端一体化构建
├── docker-compose.yml           # 服务编排
├── .env.example                 # 环境变量模板
├── deploy.sh                    # 一键部署脚本
├── go.mod
├── Makefile
└── README.md
```

---

## 🚀 快速开始

### 前置要求

- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Git**

### 方式 1: 一键部署（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/yourusername/KyxApiQuotaBridgeGo.git
cd KyxApiQuotaBridgeGo

# 2. 运行一键部署脚本
chmod +x deploy.sh
./deploy.sh

# 3. 按提示编辑 .env 文件配置必需项
# 完成！访问 http://localhost:8080
```

### 方式 2: 手动部署

```bash
# 1. 复制环境变量配置
cp .env.example .env

# 2. 编辑配置文件（必须修改以下项）
nano .env

# 必填项：
# - GITHUB_USERNAME (你的 GitHub 用户名)
# - DB_PASSWORD (数据库密码)
# - REDIS_PASSWORD (Redis 密码)
# - ADMIN_PASSWORD (管理员密码)
# - JWT_SECRET (JWT 密钥)
# - LINUX_DO_CLIENT_ID (OAuth 客户端 ID)
# - LINUX_DO_CLIENT_SECRET (OAuth 客户端密钥)
# - LINUX_DO_REDIRECT_URI (OAuth 回调地址)

# 3. 生成安全密码（可选）
openssl rand -base64 32  # 数据库密码
openssl rand -base64 32  # Redis 密码
openssl rand -base64 24  # 管理员密码
openssl rand -base64 64  # JWT 密钥

# 4. 拉取镜像并启动服务
docker-compose pull
docker-compose up -d

# 5. 查看服务状态
docker-compose ps

# 6. 查看日志
docker-compose logs -f
```

### 方式 3: 本地开发

```bash
# 前端开发
cd frontend
npm install
npm run dev  # http://localhost:3000

# 后端开发（另一个终端）
cd ..
go mod download
go run cmd/server/main.go  # http://localhost:8080
```

---

## ⚙️ 配置说明

### 环境变量配置

#### 必需配置

```bash
# GitHub 用户名（用于拉取 Docker 镜像）
GITHUB_USERNAME=yourusername

# 数据库配置
DB_PASSWORD=your_database_password

# Redis 配置
REDIS_PASSWORD=your_redis_password

# 管理员配置
ADMIN_PASSWORD=your_admin_password
JWT_SECRET=your_jwt_secret

# Linux.do OAuth2 配置（从 https://connect.linux.do 获取）
LINUX_DO_CLIENT_ID=your_client_id
LINUX_DO_CLIENT_SECRET=your_client_secret
LINUX_DO_REDIRECT_URI=https://yourdomain.com/api/auth/callback
```

#### 可选配置

```bash
# 应用端口（默认 8080）
APP_PORT=8080

# 服务器模式（release/debug）
SERVER_MODE=release

# 日志级别（debug/info/warn/error）
LOG_LEVEL=info

# 备份配置
BACKUP_SCHEDULE=@daily      # 备份计划
BACKUP_KEEP_DAYS=7          # 保留天数备份
BACKUP_KEEP_WEEKS=4         # 保留周数备份
BACKUP_KEEP_MONTHS=6        # 保留月数备份
```

### OAuth2 配置

1. 访问 [Linux.do 开发者设置](https://connect.linux.do)
2. 创建新的 OAuth2 应用
3. 设置回调 URL: `https://yourdomain.com/api/auth/callback`
4. 获取 Client ID 和 Client Secret
5. 更新 `.env` 文件

---

## 🐳 Docker 部署详解

### 镜像说明

项目使用多阶段构建，将前后端整合到单个 Docker 镜像：

```dockerfile
# 阶段 1: 前端构建（Node.js 18）
# 构建 Vue 3 前端，产物: /frontend/dist

# 阶段 2: 后端构建（Go 1.21）
# 构建 Go 后端，产物: /build/kyx-quota-bridge

# 阶段 3: 最终镜像（Alpine 3.18）
# 整合前后端，体积小，安全高效
```

### 服务架构

```
┌─────────────────────────────────────────────────────┐
│                Docker Compose 网络                   │
│                                                       │
│  ┌─────────────────┐         ┌──────────────────┐  │
│  │   Nginx (可选)  │         │   应用容器        │  │
│  │   反向代理       │────────▶│   前端+后端       │  │
│  │   :80/:443      │         │   :8080          │  │
│  └─────────────────┘         └────────┬─────────┘  │
│                                        │             │
│                       ┌────────────────┴───────┐    │
│                       │                        │    │
│              ┌────────▼───────┐      ┌────────▼───┐│
│              │   PostgreSQL   │      │   Redis    ││
│              │   数据库        │      │   缓存      ││
│              │   (内部)        │      │   (内部)    ││
│              └────────────────┘      └────────────┘│
└─────────────────────────────────────────────────────┘
```

### 服务管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose stop

# 重启服务
docker-compose restart

# 停止并删除容器
docker-compose down

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f app

# 进入容器
docker-compose exec app sh

# 更新服务
docker-compose pull
docker-compose up -d
```

### 数据备份

#### 启用自动备份

```bash
# 启动备份服务
docker-compose --profile backup up -d backup

# 查看备份日志
docker-compose logs backup
```

#### 手动备份

```bash
# 备份数据库
docker-compose exec -T postgres pg_dump -U kyxuser kyxquota > backup_$(date +%Y%m%d).sql

# 备份 Redis（如需要）
docker-compose exec redis redis-cli --no-auth-warning -a "$REDIS_PASSWORD" SAVE
```

#### 恢复数据

```bash
# 恢复数据库
cat backup_20240101.sql | docker-compose exec -T postgres psql -U kyxuser kyxquota
```

---

## 🌐 生产部署

### 使用 Nginx 反向代理

#### 1. 安装 Nginx

```bash
# Ubuntu/Debian
sudo apt install nginx -y

# CentOS/RHEL
sudo yum install nginx -y
```

#### 2. 配置 Nginx

创建配置文件 `/etc/nginx/sites-available/kyx-quota-bridge`:

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # 限制请求大小
    client_max_body_size 10M;

    # 代理到应用
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # 日志
    access_log /var/log/nginx/kyx-access.log;
    error_log /var/log/nginx/kyx-error.log;
}
```

启用站点：

```bash
sudo ln -s /etc/nginx/sites-available/kyx-quota-bridge /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

#### 3. 配置 SSL（Let's Encrypt）

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx -y

# 获取 SSL 证书
sudo certbot --nginx -d yourdomain.com

# 自动续期测试
sudo certbot renew --dry-run
```

### 防火墙配置

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --reload
```

---

## 🤖 GitHub Actions CI/CD

### 自动构建流程

GitHub Actions 会在代码推送时自动构建和推送 Docker 镜像到 GitHub Container Registry (GHCR)，**不会自动部署到服务器**。

#### 触发条件

- ✅ 推送到 `main` 或 `develop` 分支
- ✅ 推送 `v*.*.*` 格式的 tag
- ✅ 针对 `main` 分支的 Pull Request
- ✅ 手动触发：GitHub Actions → Run workflow

#### 工作流程

推送代码后，GitHub Actions 会自动完成：

1. ✅ **前端构建测试** - 编译 Vue 3 前端
2. ✅ **后端测试** - 运行 Go 测试和 lint
3. ✅ **构建 Docker 镜像** - 多架构构建（amd64/arm64）
4. ✅ **推送到 GHCR** - 推送镜像到 `ghcr.io/<你的用户名>/kyx-quota-bridge`

#### 镜像标签规则

| 推送类型 | 生成的标签 | 示例 |
|---------|-----------|------|
| main 分支 | `latest`, `main` | `ghcr.io/user/app:latest` |
| develop 分支 | `develop` | `ghcr.io/user/app:develop` |
| Tag 推送 | `v1.2.3`, `v1.2`, `v1` | `ghcr.io/user/app:v1.2.3` |
| 特定提交 | `main-sha256abc` | `ghcr.io/user/app:main-sha256abc` |

### 手动部署到服务器

镜像构建完成后，在服务器上执行以下命令部署：

```bash
# 1. SSH 登录到服务器
ssh user@your-server

# 2. 进入项目目录
cd /path/to/kyx-quota-bridge

# 3. 拉取最新镜像
docker-compose pull

# 4. 重启服务
docker-compose up -d

# 5. 查看状态
docker-compose ps

# 6. 查看日志
docker-compose logs -f
```

#### 自动化部署脚本（可选）

在服务器上创建 `update.sh` 脚本：

```bash
#!/bin/bash
cd /path/to/kyx-quota-bridge

echo "📦 拉取最新镜像..."
docker-compose pull

echo "🔄 重启服务..."
docker-compose up -d

echo "⏳ 等待服务启动..."
sleep 10

echo "🏥 健康检查..."
curl -f http://localhost:8080/health && echo "✅ 部署成功！" || echo "❌ 部署失败！"

echo "📊 服务状态："
docker-compose ps
```

使用方法：
```bash
chmod +x update.sh
./update.sh  # 一键更新部署
```

---

## 📊 监控和维护

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost:8080/health

# 预期输出：
# {"status":"healthy","version":"x.x.x","timestamp":1234567890}

# 检查版本信息
curl http://localhost:8080/version
```

### 日志管理

```bash
# 实时查看日志
docker-compose logs -f

# 查看最近 100 行日志
docker-compose logs --tail=100

# 查看特定服务日志
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f redis

# 导出日志
docker-compose logs > logs.txt
```

### 资源监控

```bash
# 查看容器资源使用
docker stats

# 查看磁盘使用
docker system df

# 查看数据卷
docker volume ls
```

### 清理优化

```bash
# 清理未使用的 Docker 资源
docker system prune -a

# 清理日志文件（30天前）
find logs/ -name "*.log" -mtime +30 -delete

# 清理旧备份（90天前）
find backups/ -name "*.sql" -mtime +90 -delete
```

---

## 🔧 故障排查

### 常见问题

#### 1. 容器无法启动

**症状**: `docker-compose up -d` 失败

**解决方案**:
```bash
# 查看详细日志
docker-compose logs app

# 检查环境变量
cat .env

# 检查端口占用
lsof -i :8080
```

#### 2. 前端无法访问

**症状**: 浏览器无法打开应用

**解决方案**:
- 检查防火墙规则
- 确认端口映射正确
- 查看 Nginx 配置（如使用）

#### 3. 数据库连接失败

**症状**: `Failed to connect to database`

**解决方案**:
```bash
# 检查 PostgreSQL 状态
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres

# 进入数据库容器
docker-compose exec postgres psql -U kyxuser -d kyxquota
```

#### 4. Redis 连接失败

**症状**: `Failed to connect to Redis`

**解决方案**:
```bash
# 检查 Redis 状态
docker-compose ps redis

# 测试连接
docker-compose exec redis redis-cli -a "$REDIS_PASSWORD" ping
```

#### 5. OAuth 登录失败

**症状**: 登录跳转后报错

**解决方案**:
- 确认 `LINUX_DO_REDIRECT_URI` 正确
- 检查 Linux.do 应用配置
- 验证客户端 ID 和密钥

---

## 🔒 安全建议

### 1. 密码安全

```bash
# 使用强随机密码
openssl rand -base64 32

# 定期更换密码
# 不要在代码中硬编码密码
# 使用环境变量管理敏感信息
```

### 2. 网络安全

- ✅ 使用 HTTPS（配置 SSL 证书）
- ✅ 启用防火墙
- ✅ 仅开放必要端口
- ✅ 数据库和 Redis 不暴露到公网

### 3. 访问控制

- ✅ 使用强管理员密码
- ✅ 限制管理员 IP（可选）
- ✅ 定期审查用户权限

### 4. 数据保护

- ✅ 定期备份数据
- ✅ 测试备份恢复
- ✅ 异地存储备份

### 5. 更新维护

- ✅ 定期更新依赖
- ✅ 关注安全公告
- ✅ 及时修复漏洞

---

## 📚 API 文档

### 认证相关

```http
# 获取 OAuth 授权 URL
GET /api/auth/url

# OAuth 回调
GET /api/auth/callback?code=xxx&state=xxx

# 检查认证状态
GET /api/auth/check

# 用户登出
POST /api/auth/logout

# 管理员登录
POST /api/auth/admin/login
Content-Type: application/json
{
  "password": "admin_password"
}
```

### 用户相关

```http
# 绑定账号
POST /api/user/bind
Authorization: Cookie
Content-Type: application/json
{
  "username": "your_username"
}

# 获取额度信息
GET /api/user/quota

# 领取每日额度
POST /api/user/claim

# 投喂 Keys
POST /api/user/donate
Content-Type: application/json
{
  "keys": ["sk-xxx", "sk-yyy"]
}

# 获取领取历史
GET /api/user/claims?page=1&page_size=20

# 获取投喂历史
GET /api/user/donates?page=1&page_size=20
```

### 管理员相关

```http
# 获取系统配置
GET /api/admin/config
Authorization: Bearer <token>

# 更新系统配置
PUT /api/admin/config
Authorization: Bearer <token>
Content-Type: application/json
{
  "claim_quota": 500000,
  "session": "your_session"
}

# 获取系统统计
GET /api/admin/stats

# 获取用户列表
GET /api/admin/users?page=1&page_size=20

# 删除用户
DELETE /api/admin/users/:linux_do_id
```

---

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### Commit 规范

```bash
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
perf: 性能优化
test: 测试相关
chore: 构建/工具链相关
```

---

## 📄 许可证

本项目采用 [MIT License](LICENSE)

---

## 🙏 致谢

- [Go](https://go.dev/) - 编程语言
- [Vue.js](https://vuejs.org/) - 前端框架
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [PostgreSQL](https://www.postgresql.org/) - 数据库
- [Redis](https://redis.io/) - 缓存
- [Ant Design Vue](https://antdv.com/) - UI 组件库
- Linux.do 社区

---

## 📞 联系方式

- **项目地址**: [GitHub](https://github.com/yourusername/KyxApiQuotaBridgeGo)
- **问题反馈**: [Issues](https://github.com/yourusername/KyxApiQuotaBridgeGo/issues)
- **讨论交流**: [Discussions](https://github.com/yourusername/KyxApiQuotaBridgeGo/discussions)

---

## 📝 更新日志

### v1.0.0 (2024)

- ✅ 前后端分离架构重构
- ✅ Vue 3 + TypeScript 前端
- ✅ Go 1.21 后端
- ✅ Docker 一体化部署
- ✅ GitHub Actions CI/CD
- ✅ 完善的文档和部署指南

---

**⚠️ 免责声明**: 本项目仅供学习和研究使用，请遵守相关法律法规和服务条款。

---

<p align="center">
  Made with ❤️ by KyxApiQuotaBridge Team
</p>
