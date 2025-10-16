# KYX API 配额桥接系统 - 前端

基于 Vue 3 + Vite + Tailwind CSS + Ant Design Vue 构建的现代化前端应用。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite 5
- **UI 库**: Ant Design Vue 4
- **CSS 框架**: Tailwind CSS 3
- **状态管理**: Pinia 2
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios
- **语言**: TypeScript

## 设计风格

采用 Grok (X AI) 风格的现代极简主义设计:
- 深色主题为主
- 科技感强
- 流畅的动画效果
- 玻璃态效果
- 渐变色彩

## 快速开始

### 安装依赖

```bash
npm install
```

### 配置环境变量

复制 `.env.example` 为 `.env` 并填写相关配置:

```bash
cp .env.example .env
```

### 开发模式

```bash
npm run dev
```

应用将在 http://localhost:3000 启动

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `../web` 目录，可以直接被 Go 后端服务

### 预览生产版本

```bash
npm run preview
```

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 接口定义
│   ├── assets/           # 静态资源
│   ├── components/       # 组件
│   │   └── Layout.vue    # 布局组件
│   ├── router/           # 路由配置
│   ├── stores/           # Pinia 状态管理
│   │   └── user.ts       # 用户状态
│   ├── styles/           # 样式文件
│   │   └── main.css      # 主样式 (Tailwind CSS)
│   ├── utils/            # 工具函数
│   │   └── request.ts    # HTTP 请求封装
│   ├── views/            # 页面组件
│   │   ├── Home.vue      # 首页
│   │   ├── Dashboard.vue # 用户仪表板
│   │   ├── Admin.vue     # 管理员后台
│   │   ├── Bind.vue      # 绑定账号
│   │   └── NotFound.vue  # 404 页面
│   ├── App.vue           # 根组件
│   ├── main.ts           # 入口文件
│   └── vite-env.d.ts     # TypeScript 类型定义
├── index.html            # HTML 模板
├── package.json          # 依赖配置
├── tsconfig.json         # TypeScript 配置
├── vite.config.ts        # Vite 配置
├── tailwind.config.js    # Tailwind CSS 配置
└── postcss.config.js     # PostCSS 配置
```

## 主要功能

### 用户功能
- OAuth 登录 (Linux Do)
- 绑定公益站账号
- 查看配额信息
- 每日领取配额
- 投喂 ModelScope Keys
- 测试 Key 有效性
- 查看领取/投喂记录

### 管理员功能
- 管理员登录
- 系统配置管理
- Keys 管理 (导出/测试/删除)
- 用户管理
- 记录查询

## 自定义主题

项目使用 Tailwind CSS 的 Grok 主题配置，可在 `tailwind.config.js` 中自定义颜色:

```js
colors: {
  grok: {
    bg: '#000000',
    primary: '#1d9bf0',
    // ...更多颜色
  }
}
```

## 开发建议

1. 使用 Vue DevTools 进行调试
2. 遵循 Vue 3 Composition API 最佳实践
3. 使用 TypeScript 提供类型安全
4. 保持组件的单一职责
5. 使用 Tailwind CSS 的 utility-first 方式编写样式

## 浏览器支持

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## License

MIT
