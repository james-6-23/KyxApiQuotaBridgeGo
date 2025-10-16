# 项目结构说明

本文档描述 KYX API Quota Bridge 项目的目录结构和组织方式。

## 目录结构

```
KyxApiQuotaBridgeGo/
├── .github/                    # GitHub 配置
│   └── workflows/             # GitHub Actions 工作流
│       ├── ci.yml            # 持续集成工作流
│       └── docker-build.yml  # Docker 镜像构建工作流
│
├── frontend/                   # 前端应用（Vue 3）
│   ├── public/                # 静态资源
│   ├── src/                   # 源代码
│   │   ├── assets/           # 资源文件
│   │   ├── components/       # Vue 组件
│   │   ├── router/           # 路由配置
│   │   ├── stores/           # Pinia 状态管理
│   │   ├── views/            # 页面视图
│   │   ├── App.vue           # 根组件
│   │   └── main.ts           # 入口文件
│   ├── index.html            # HTML 模板
│   ├── package.json          # 前端依赖
│   ├── tsconfig.json         # TypeScript 配置
│   ├── vite.config.ts        # Vite 构建配置
│   └── tailwind.config.js    # Tailwind CSS 配置
│
├── internal/                   # Go 后端内部包
│   ├── api/                  # API 处理器
│   │   ├── handlers/         # HTTP 处理函数
│   │   ├── middleware/       # 中间件
│   │   ├── routes.go         # 路由定义
│   │   └── server.go         # 服务器配置
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── models/               # 数据模型
│   │   └── models.go
│   └── store/                # 数据存储层
│       └── db.go             # 数据库操作
│
├── web/                        # 前端构建输出（由 Vite 生成）
│   ├── assets/               # 编译后的资源
│   └── index.html            # 编译后的 HTML
│
├── data/                       # 数据目录（运行时创建）
│   └── data.db               # SQLite 数据库文件
│
├── .dockerignore              # Docker 构建忽略文件
├── .env.example               # 环境变量示例
├── .gitignore                 # Git 忽略文件
├── docker-compose.yml         # Docker Compose 配置
├── Dockerfile                 # Docker 镜像构建文件
├── go.mod                     # Go 模块依赖
├── go.sum                     # Go 依赖校验和
├── main.go                    # Go 应用入口
├── API.md                     # API 文档
├── DOCKER_DEPLOY.md           # Docker 部署指南
├── PROJECT_STRUCTURE.md       # 本文档
├── QUICKSTART.md              # 快速开始指南
└── README.md                  # 项目说明文档
```

## 核心组件说明

### 后端 (Go)

#### main.go
应用程序入口点，负责：
- 加载配置
- 初始化数据库
- 创建并启动 API 服务器

#### internal/api
API 层，包含：
- **handlers/**: HTTP 请求处理器
- **middleware/**: 认证、日志等中间件
- **routes.go**: 路由配置
- **server.go**: Gin 服务器初始化和配置

#### internal/config
配置管理，处理：
- 环境变量读取
- 配置文件解析
- 默认值设置

#### internal/models
数据模型定义：
- 用户模型
- API 密钥模型
- 配额管理模型
- 请求日志模型

#### internal/store
数据存储层，使用 SQLite：
- 数据库连接管理
- CRUD 操作
- 数据迁移

### 前端 (Vue 3)

#### src/views
页面视图：
- 登录页面
- 仪表板
- API 密钥管理
- 用户管理
- 配额管理

#### src/components
可复用组件：
- 通用 UI 组件
- 业务组件
- 布局组件

#### src/stores (Pinia)
状态管理：
- 用户状态
- 主题设置
- API 配置

#### src/router
Vue Router 配置：
- 路由定义
- 路由守卫
- 懒加载配置

## 构建流程

### 开发环境

**后端：**
```bash
go run main.go
```

**前端：**
```bash
cd frontend
npm run dev
```

### 生产构建

**前端构建：**
```bash
cd frontend
npm run build  # 输出到 ../web
```

**后端构建：**
```bash
go build -o kyx-api-quota-bridge .
```

### Docker 构建

Docker 构建使用多阶段构建：

1. **阶段 1**: 构建前端（Node.js）
   - 安装依赖
   - 构建 Vue 应用
   - 输出到 `web` 目录

2. **阶段 2**: 构建后端（Go）
   - 下载 Go 依赖
   - 编译 Go 应用
   - 复制前端构建产物

3. **阶段 3**: 运行环境（Alpine）
   - 复制二进制文件
   - 配置运行环境
   - 设置健康检查

## 配置文件

### .env
运行时环境变量：
- 数据库路径
- 管理员凭据
- JWT 密钥
- OAuth 配置

### .dockerignore
Docker 构建时忽略的文件：
- 开发工具配置
- 临时文件
- 构建产物
- 测试文件

### .gitignore
Git 版本控制忽略的文件：
- 二进制文件
- 依赖目录
- 环境配置
- 数据库文件
- 构建产物

## 数据持久化

### 数据库
- **类型**: SQLite
- **位置**: `data/data.db`
- **备份**: 建议定期备份 `data` 目录

### Docker 卷
在 Docker 部署中，`data` 目录通过卷挂载实现持久化：
```yaml
volumes:
  - ./data:/app/data
```

## API 端点

详细的 API 文档请参考 `API.md`。

主要端点：
- `/api/auth/*` - 认证相关
- `/api/users/*` - 用户管理
- `/api/keys/*` - API 密钥管理
- `/api/quota/*` - 配额管理
- `/api/health` - 健康检查

## 开发工作流

1. **克隆项目**
   ```bash
   git clone https://github.com/james-6-23/KyxApiQuotaBridgeGo.git
   cd KyxApiQuotaBridgeGo
   ```

2. **安装依赖**
   ```bash
   # 后端
   go mod download

   # 前端
   cd frontend && npm install
   ```

3. **运行开发服务器**
   ```bash
   # 终端 1: 后端
   go run main.go

   # 终端 2: 前端
   cd frontend && npm run dev
   ```

4. **构建生产版本**
   ```bash
   # 使用 Docker
   docker-compose up --build
   ```

## 部署

详细的部署指南请参考：
- Docker 部署：`DOCKER_DEPLOY.md`
- 快速开始：`QUICKSTART.md`

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 技术栈

**后端：**
- Go 1.24
- Gin Web Framework
- SQLite (modernc.org/sqlite)
- JWT 认证

**前端：**
- Vue 3
- TypeScript
- Vite
- Ant Design Vue
- Tailwind CSS
- Pinia (状态管理)
- Vue Router

**DevOps：**
- Docker
- Docker Compose
- GitHub Actions
- GitHub Container Registry

## 许可证

详见 LICENSE 文件。
