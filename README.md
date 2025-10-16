# KYX API 配额桥接系统

一个基于 Go + Vue.js + Element Plus 构建的 API 配额管理和分配平台，支持用户通过 Linux Do OAuth2 登录，绑定公益站账号，领取每日配额，以及投喂 ModelScope Keys。

## ✨ 功能特性

### 用户功能
- ✅ Linux Do OAuth2 登录认证
- ✅ 公益站账号绑定（精确匹配 + Linux Do ID 验证）
- ✅ 每日额度领取
- ✅ ModelScope Key 投喂（捐赠）
- ✅ Keys 自动验证和推送到指定分组
- ✅ 投喂记录推送状态追踪
- ✅ 查看领取/投喂记录

### 管理员功能
- ✅ 完整的管理员后台
- ✅ 系统配置管理（额度设置、Session 配置等）
- ✅ Keys 管理（导出、批量测试、删除）
- ✅ 用户管理（列表查看、重新绑定账号）
- ✅ 记录查询（领取记录、投喂记录）
- ✅ 失败 Keys 一键重新推送

## 🎨 技术栈

### 后端
- **语言**: Go 1.24+
- **Web 框架**: Gin v1.10.0
- **数据库**: SQLite (modernc.org/sqlite)
- **认证**: OAuth2 + JWT
- **UUID**: google/uuid

### 前端
- **框架**: Vue.js 3.4.0
- **UI 组件库**: Element Plus 2.8.0
- **图标**: Element Plus Icons
- **样式**: 原生 CSS + 渐变背景

### 部署
- **容器化**: Docker
- **单一二进制**: 支持直接编译运行

## 📁 项目结构

```
KyxApiQuotaBridgeGo/
├── main.go                         # 程序入口
├── go.mod                          # Go 模块定义
├── Dockerfile                      # Docker 构建文件
├── README.md                       # 项目文档
│
├── internal/                       # 内部包
│   ├── config/                     # 配置管理
│   │   └── config.go              # 配置加载器
│   │
│   ├── models/                     # 数据模型
│   │   └── models.go              # 所有数据结构定义
│   │
│   ├── store/                      # 数据存储层
│   │   └── db.go                  # SQLite 数据库操作
│   │
│   └── api/                        # API 服务层
│       ├── server.go              # 服务器设置和路由
│       ├── handlers_user.go       # 用户相关处理器
│       ├── handlers_admin.go      # 管理员处理器
│       ├── handlers_claim.go      # 领取额度处理器
│       ├── handlers_donate.go     # 投喂 Keys 处理器
│       ├── templates.go           # 模板渲染（已废弃）
│       └── utils.go               # 工具函数
│
└── web/                           # 前端文件
    ├── index.html                 # 用户主页
    ├── app.js                     # 用户页面逻辑
    ├── admin.html                 # 管理员后台
    └── admin.js                   # 管理员页面逻辑
```

## 🚀 快速开始

### 环境要求

- Go 1.24+ （注意：已适配 Go 1.24.6 版本）
- Git

### 方式1：直接运行（开发环境）

```bash
# 1. 克隆项目
git clone https://github.com/james-6-23/KyxApiQuotaBridgeGo.git
cd KyxApiQuotaBridgeGo

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填入你的配置

# 3. 下载依赖
go mod download

# 4. 运行项目
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 方式2：构建运行

```bash
# 1. 构建
go build -o kyx-api-quota-bridge

# 2. 运行
./kyx-api-quota-bridge
```

### 方式3：Docker 部署（推荐生产环境）

```bash
# 1. 构建镜像
docker build -t kyx-api-quota-bridge:latest .

# 2. 运行容器
docker run -d \
  --name kyx-api-quota-bridge \
  -p 8080:8080 \
  -v $(pwd)/data:/root \
  -e LINUX_DO_CLIENT_ID=your_client_id \
  -e LINUX_DO_CLIENT_SECRET=your_client_secret \
  -e LINUX_DO_REDIRECT_URI=https://yourdomain.com/oauth/callback \
  -e ADMIN_PASSWORD=your_admin_password \
  kyx-api-quota-bridge:latest
```

## ⚙️ 环境变量配置

创建 `.env` 文件：

```env
# Linux Do OAuth2 配置（必需）
LINUX_DO_CLIENT_ID=your_client_id
LINUX_DO_CLIENT_SECRET=your_client_secret
LINUX_DO_REDIRECT_URI=http://localhost:8080/oauth/callback

# 管理员密码（必需）
ADMIN_PASSWORD=your_secure_password

# 数据库路径（可选，默认：./data.db）
DATABASE_PATH=./data.db

# 服务器端口（可选，默认：8080）
PORT=8080
```

### 获取 Linux Do OAuth2 凭证

1. 访问 [Linux Do 开发者设置](https://linux.do/my/preferences/applications)
2. 创建新应用
3. 设置回调 URL：`http://localhost:8080/oauth/callback`（或你的域名）
4. 复制 Client ID 和 Client Secret

## 📋 API 文档

### 用户 API

| 方法 | 路径 | 说明 | 认证 |
|-----|------|-----|------|
| GET | `/` | 用户主页 | - |
| GET | `/oauth/callback` | OAuth 回调 | - |
| POST | `/api/auth/bind` | 绑定账号 | Session |
| POST | `/api/auth/logout` | 退出登录 | Session |
| GET | `/api/user/quota` | 获取用户额度 | Session |
| GET | `/api/user/records/claim` | 获取领取记录 | Session |
| GET | `/api/user/records/donate` | 获取投喂记录 | Session |
| POST | `/api/claim/daily` | 每日领取额度 | Session |
| POST | `/api/donate/validate` | 投喂 Keys | Session |
| POST | `/api/test/key` | 测试单个 Key | Session |

### 管理员 API

| 方法 | 路径 | 说明 | 认证 |
|-----|------|-----|------|
| GET | `/admin.html` | 管理员后台 | - |
| POST | `/api/admin/login` | 管理员登录 | - |
| GET | `/api/admin/config` | 获取配置 | Admin Session |
| PUT | `/api/admin/config/quota` | 更新领取额度 | Admin Session |
| PUT | `/api/admin/config/session` | 更新 Session | Admin Session |
| PUT | `/api/admin/config/new-api-user` | 更新 new-api-user | Admin Session |
| PUT | `/api/admin/config/keys-api-url` | 更新 Keys API URL | Admin Session |
| PUT | `/api/admin/config/keys-authorization` | 更新授权令牌 | Admin Session |
| PUT | `/api/admin/config/group-id` | 更新分组 ID | Admin Session |
| GET | `/api/admin/keys/export` | 导出所有 Keys | Admin Session |
| POST | `/api/admin/keys/test` | 批量测试 Keys | Admin Session |
| POST | `/api/admin/keys/delete` | 删除 Keys | Admin Session |
| GET | `/api/admin/records/claim` | 获取所有领取记录 | Admin Session |
| GET | `/api/admin/records/donate` | 获取所有投喂记录 | Admin Session |
| GET | `/api/admin/users` | 获取所有用户 | Admin Session |
| POST | `/api/admin/rebind-user` | 重新绑定用户 | Admin Session |
| POST | `/api/admin/retry-push` | 重新推送失败 Keys | Admin Session |

## 🎯 使用流程

### 用户使用流程

1. **访问首页** → 点击"Linux Do 登录"
2. **OAuth 认证** → 授权应用访问你的 Linux Do 账号
3. **绑定账号** → 输入公益站用户名和 Session
4. **领取额度** → 每日可领取配额
5. **投喂 Keys** → 提交 ModelScope Keys 贡献社区

### 管理员使用流程

1. **访问后台** → `/admin.html`
2. **输入密码** → 使用环境变量中设置的管理员密码
3. **配置系统** → 设置每日领取额度、Session 等
4. **管理 Keys** → 导出、测试、删除 Keys
5. **管理用户** → 查看用户列表、重新绑定账号
6. **查看记录** → 查看所有领取和投喂记录

## 🔐 安全性

- ✅ 密码不明文存储
- ✅ Session 自动过期（24小时）
- ✅ HTTP-only Cookie
- ✅ SQL 注入防护（参数化查询）
- ✅ CORS 配置
- ✅ 管理员权限验证
- ✅ OAuth2 标准认证流程

## 🐛 调试

```bash
# 查看日志
docker logs -f kyx-api-quota-bridge

# 进入容器
docker exec -it kyx-api-quota-bridge sh

# 查看数据库
sqlite3 data.db
.tables
SELECT * FROM users;
```

## 📊 性能特点

| 特性 | 指标 |
|------|-----|
| 启动时间 | ~10ms |
| 内存占用 | ~15MB |
| 并发性能 | 优秀（goroutines） |
| 部署方式 | 单一二进制文件 |
| 数据库 | SQLite（无额外依赖） |

## 🔄 更新部署

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建
go build -o kyx-api-quota-bridge

# 3. 重启服务
# Docker:
docker restart kyx-api-quota-bridge

# 或直接运行:
./kyx-api-quota-bridge
```

## 📝 开发说明

### 添加新功能

1. 在 `internal/api/` 中添加新的处理器
2. 在 `internal/api/server.go` 中注册路由
3. 如需数据库支持，在 `internal/store/db.go` 中添加方法
4. 更新前端页面（`web/` 目录）

### 修改前端

- 用户页面：编辑 `web/index.html` 和 `web/app.js`
- 管理员页面：编辑 `web/admin.html` 和 `web/admin.js`
- 使用 Vue 3 Composition API
- UI 组件使用 Element Plus

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- Vue.js 社区
- Element Plus 团队
- Go 社区
- Gin 框架
- modernc.org SQLite 实现

## 📞 支持

如有问题，请：
1. 查看本文档
2. 提交 [Issue](https://github.com/james-6-23/KyxApiQuotaBridgeGo/issues)
3. 参与 [Discussions](https://github.com/james-6-23/KyxApiQuotaBridgeGo/discussions)

---

**注意事项**：
1. 首次运行会自动创建数据库和表结构
2. 确保 `.env` 文件配置正确
3. 生产环境建议使用 Docker 部署
4. 定期备份 `data.db` 数据库文件
5. Go 版本必须 >= 1.24

祝使用愉快！🎉