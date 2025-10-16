# 🚀 KYX API 配额桥接系统 - 快速开始

> 现代化的 API 配额管理平台 | Vue 3 + Go + Grok 风格设计

## 📋 前置要求

- Node.js >= 18.0
- Go >= 1.21
- npm 或 yarn

## 🛠️ 安装步骤

### 1. 克隆项目

```bash
git clone https://github.com/james-6-23/KyxApiQuotaBridgeGo.git
cd KyxApiQuotaBridgeGo
```

### 2. 安装前端依赖

```bash
cd frontend
npm install
```

### 3. 配置环境变量

#### 前端配置

```bash
cd frontend
cp .env.example .env
```

编辑 `.env` 文件:

```env
VITE_OAUTH_CLIENT_ID=你的-linux-do-client-id
VITE_OAUTH_REDIRECT_URI=http://localhost:3000/oauth/callback
VITE_API_BASE_URL=http://localhost:8080
```

#### 后端配置

在项目根目录创建 `.env` 文件:

```env
# Linux Do OAuth 配置
OAUTH_CLIENT_ID=你的-client-id
OAUTH_CLIENT_SECRET=你的-client-secret
OAUTH_REDIRECT_URI=http://localhost:8080/oauth/callback

# 管理员密码
ADMIN_PASSWORD=your-admin-password

# 公益站配置
# (首次运行后可在管理后台配置)
```

#### 如何获取 Linux Do OAuth2 凭证？

1. 访问 https://linux.do/my/preferences/applications
2. 点击"新建应用"
3. 填写应用信息：
   - 应用名称：KYX API Bridge
   - 回调 URL：`http://localhost:8080/oauth/callback`
4. 保存后复制 Client ID 和 Client Secret

### 4. 安装后端依赖

```bash
# 在项目根目录
go mod download
```

## 🎯 开发模式

### 方式一: 前后端分离开发 (推荐)

**启动后端 (终端1)**
```bash
go run main.go
# 后端运行在 http://localhost:8080
```

**启动前端 (终端2)**
```bash
cd frontend
npm run dev
# 前端运行在 http://localhost:3000
```

访问 http://localhost:3000 查看应用，你将看到：
- 🎨 Grok 风格深色主题界面
- ✨ 流畅的动画效果
- 🔐 现代化的登录界面
- 📱 完美的响应式设计

### 方式二: 前后端一体化

```bash
# 1. 构建前端
cd frontend
npm run build

# 2. 启动后端（会自动服务前端静态文件）
cd ..
go run main.go

# 访问 http://localhost:8080
```

## 功能测试清单

### 用户功能测试

1. **登录测试**
   - [ ] 点击"Linux Do 登录"按钮
   - [ ] 完成 OAuth 授权
   - [ ] 成功返回并显示用户名

2. **绑定账号测试**
   - [ ] 输入公益站用户名
   - [ ] 输入 Session
   - [ ] 点击"绑定账号"
   - [ ] 绑定成功后进入仪表板

3. **领取额度测试**
   - [ ] 查看可用额度显示
   - [ ] 点击"领取今日额度"
   - [ ] 成功领取并更新额度

4. **投喂 Keys 测试**
   - [ ] 切换到"投喂 Keys"标签
   - [ ] 输入测试 Keys（每行一个）
   - [ ] 点击"提交 Keys"
   - [ ] 查看投喂结果

5. **测试 Key 功能**
   - [ ] 切换到"测试 Key"标签
   - [ ] 输入单个 Key
   - [ ] 点击"测试 Key"
   - [ ] 查看验证结果

### 管理员功能测试

1. **登录测试**
   - [ ] 访问 `/admin.html`
   - [ ] 输入管理员密码
   - [ ] 成功进入后台

2. **系统配置测试**
   - [ ] 修改每日领取额度
   - [ ] 更新 Session 配置
   - [ ] 更新 Keys API 配置

3. **Keys 管理测试**
   - [ ] 导出所有 Keys
   - [ ] 批量测试 Keys
   - [ ] 删除指定 Keys

4. **用户管理测试**
   - [ ] 查看用户列表
   - [ ] 重新绑定用户账号

5. **记录查询测试**
   - [ ] 查看领取记录
   - [ ] 查看投喂记录

## 常见问题

### Q1: 启动时提示端口被占用？

```bash
# 修改端口
# 在 .env 中添加：
PORT=8081
```

### Q2: OAuth 回调失败？

检查以下几点：
1. `.env` 中的 `LINUX_DO_REDIRECT_URI` 是否正确
2. Linux Do 应用设置中的回调 URL 是否匹配
3. 是否使用 `http://` 而不是 `https://`（本地开发）

### Q3: 前端页面无法加载？

确保 `web/` 目录下的文件完整：
- `index.html`
- `app.js`
- `admin.html`
- `admin.js`

### Q4: 管理员登录失败？

检查 `.env` 文件中的 `ADMIN_PASSWORD` 是否正确设置。

### Q5: 数据库错误？

删除 `data.db` 文件，重新启动程序：

```bash
rm data.db
go run main.go
```

## 开发模式

启用 Gin 的调试模式以查看详细日志：

```bash
# 在 internal/api/server.go 中修改：
# gin.SetMode(gin.ReleaseMode)  改为
gin.SetMode(gin.DebugMode)
```

## 生产部署

参考 [README.md](README.md) 中的 Docker 部署部分。

## 下一步

- 📖 阅读完整文档：[README.md](README.md)
- 🔧 了解 API 接口：查看 API 文档章节
- 🎨 自定义前端：修改 `web/` 目录下的文件
- 🚀 部署到服务器：使用 Docker 容器化部署

## 获取帮助

- 📝 提交 Issue
- 💬 参与 Discussions
- 📧 联系维护者

---

祝你使用愉快！🎉