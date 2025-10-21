# KyxApiQuotaBridge Frontend

Vue 3 + TypeScript + Vite + Ant Design Vue 前端应用

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite 5
- **UI 框架**: Ant Design Vue 4
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **样式**: Tailwind CSS 3
- **HTTP 客户端**: Axios
- **日期处理**: Day.js

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

## 项目结构

```
frontend/
├── public/              # 静态资源
├── src/
│   ├── api/            # API 接口定义
│   ├── assets/         # 资源文件（图片、样式等）
│   ├── layouts/        # 布局组件
│   ├── router/         # 路由配置
│   ├── stores/         # Pinia 状态管理
│   ├── types/          # TypeScript 类型定义
│   ├── views/          # 页面组件
│   ├── App.vue         # 根组件
│   └── main.ts         # 应用入口
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## 环境变量

开发环境下无需配置，Vite 已配置代理将 `/api` 请求转发到 `http://localhost:8080`。

生产环境下，前端静态文件由后端服务器提供，API 请求也由同一服务器处理。

## 路由

- **用户端**: `/user/*`
  - `/user/login` - 用户登录
  - `/user/dashboard` - 用户仪表板
  - `/user/bind` - 绑定账号
  - `/user/claim` - 领取额度
  - `/user/donate` - 投喂 Keys

- **管理端**: `/admin/*`
  - `/admin/login` - 管理员登录
  - `/admin/dashboard` - 管理仪表板
  - `/admin/config` - 系统配置
  - `/admin/keys` - Keys 管理
  - `/admin/claims` - 领取记录
  - `/admin/donates` - 投喂记录
  - `/admin/users` - 用户管理

## 代码规范

```bash
# 代码检查
npm run lint

# 代码格式化
npm run format
```

## 构建部署

前端构建产物（`dist/`）会被整合到 Docker 镜像中，由后端 Go 服务器提供静态文件服务。

详见项目根目录的 `DEPLOYMENT_GUIDE.md`。

## License

MIT
