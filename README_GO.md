# KYX API Quota Bridge - Go 版本

这是 KYX API Quota Bridge 的 Go 语言实现版本，使用 Gin Web 框架和 SQLite 数据库。

## 功能特性

- ✅ Linux Do OAuth2 登录
- ✅ 公益站账号绑定（精确匹配 + Linux Do ID 验证）
- ✅ 每日额度领取
- ✅ ModelScope Key 投喂
- ✅ Keys 自动推送到分组
- ✅ 投喂记录推送状态追踪
- ✅ 失败 Keys 一键重新推送
- ✅ 管理员重新绑定用户账号
- ✅ 完整的管理员后台

## 技术栈

- **框架**: Gin Web Framework
- **数据库**: SQLite (modernc.org/sqlite)
- **认证**: OAuth2 (golang.org/x/oauth2)
- **UUID**: google/uuid
- **JWT**: golang-jwt/jwt/v5

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- GCC (用于编译 SQLite)

### 安装依赖

```bash
go mod download
```

### 配置

复制 `.env.example` 为 `.env` 并填写配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入您的配置：

```env
LINUX_DO_CLIENT_ID=your_client_id
LINUX_DO_CLIENT_SECRET=your_client_secret
LINUX_DO_REDIRECT_URI=http://localhost:8080/api/auth/callback
ADMIN_PASSWORD=your_admin_password
```

### 运行

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 构建

```bash
go build -o kyx-api-quota-bridge
./kyx-api-quota-bridge
```

## Docker 部署

### 构建镜像

```bash
docker build -t kyx-api-quota-bridge:go .
```

### 运行容器

```bash
docker run -d \
  --name kyx-api-quota-bridge \
  -p 8080:8080 \
  -v $(pwd)/data:/root \
  -e LINUX_DO_CLIENT_ID=your_client_id \
  -e LINUX_DO_CLIENT_SECRET=your_client_secret \
  -e LINUX_DO_REDIRECT_URI=https://yourdomain.com/api/auth/callback \
  -e ADMIN_PASSWORD=your_admin_password \
  kyx-api-quota-bridge:go
```

## 项目结构

```
.
├── main.go                     # 主入口
├── internal/
│   ├── config/                 # 配置管理
│   │   └── config.go
│   ├── models/                 # 数据模型
│   │   └── models.go
│   ├── store/                  # 数据库层
│   │   └── db.go
│   └── api/                    # API 层
│       ├── server.go           # 服务器设置
│       ├── handlers_user.go    # 用户处理器
│       ├── handlers_admin.go   # 管理员处理器
│       ├── handlers_claim.go   # 领取处理器
│       ├── handlers_donate.go  # 投喂处理器
│       └── utils.go            # 工具函数
├── go.mod                      # Go 模块定义
├── go.sum                      # 依赖锁定
├── Dockerfile                  # Docker 构建文件
├── .env.example                # 环境变量示例
└── README_GO.md               # 本文档
```

## API 文档

### 用户 API

#### 认证
- `GET /api/auth/login` - 登录
- `GET /api/auth/callback` - OAuth 回调
- `POST /api/auth/bind` - 绑定公益站账号
- `POST /api/auth/logout` - 登出

#### 用户信息
- `GET /api/user/quota` - 获取用户额度
- `GET /api/user/records/claim` - 获取领取记录
- `GET /api/user/records/donate` - 获取投喂记录

#### 功能
- `POST /api/claim/daily` - 每日领取额度
- `POST /api/donate/validate` - 投喂 Keys
- `POST /api/test/key` - 测试单个 Key

### 管理员 API

#### 配置管理
- `GET /api/admin/config` - 获取配置
- `PUT /api/admin/config/quota` - 更新领取额度
- `PUT /api/admin/config/session` - 更新 Session
- `PUT /api/admin/config/new-api-user` - 更新 new-api-user
- `PUT /api/admin/config/keys-api-url` - 更新 Keys API URL
- `PUT /api/admin/config/keys-authorization` - 更新授权令牌
- `PUT /api/admin/config/group-id` - 更新分组 ID

#### Keys 管理
- `GET /api/admin/keys/export` - 导出所有 Keys
- `POST /api/admin/keys/test` - 批量测试 Keys
- `POST /api/admin/keys/delete` - 删除 Keys

#### 记录查询
- `GET /api/admin/records/claim` - 获取所有领取记录
- `GET /api/admin/records/donate` - 获取所有投喂记录

#### 用户管理
- `GET /api/admin/users` - 获取所有用户
- `POST /api/admin/rebind-user` - 重新绑定用户账号

#### 其他
- `POST /api/admin/retry-push` - 重新推送失败的 Keys

## 与 Deno 版本的差异

### 优势
- ✅ 更好的性能
- ✅ 更小的内存占用
- ✅ 原生支持并发
- ✅ 成熟的生态系统
- ✅ 更容易部署（单一二进制文件）

### 功能一致性
- ✅ 所有核心功能完全一致
- ✅ API 接口兼容
- ✅ 数据库结构相同
- ✅ 前端页面可直接复用

## 性能优化

1. **数据库连接池**: 自动管理
2. **并发处理**: Go 原生协程支持
3. **HTTP 客户端**: 复用连接
4. **静态文件**: 嵌入式文件系统

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

