# 主题切换功能说明

## 功能概述

前端应用现在完全支持浅色/深色主题切换，用户可以通过导航栏中的主题切换按钮在两种主题之间自由切换。

## 已实现的功能

### 1. 主题管理 (Theme Store)
- 位置: `src/stores/theme.ts`
- 功能:
  - 自动从 localStorage 读取用户主题偏好
  - 默认主题: 浅色模式 (light)
  - 主题切换时自动保存到 localStorage
  - 自动更新 HTML 根元素的 class 属性

### 2. Ant Design Vue 主题集成
- 位置: `src/App.vue`
- 功能:
  - 根据当前主题动态切换 Ant Design 的主题算法
  - 深色模式使用 `theme.darkAlgorithm`
  - 浅色模式使用 `theme.defaultAlgorithm`
  - 自定义颜色令牌适配两种主题

### 3. Tailwind CSS 主题配置
- 位置: `tailwind.config.js`
- 功能:
  - 定义了完整的浅色/深色主题颜色系统
  - 包含 `light-*` 和 `dark-*` 颜色变量
  - 兼容性别名 `grok-*` (用于现有代码)
  - 暗色模式使用 `class` 策略

### 4. 全局样式适配
- 位置: `src/styles/main.css`
- 功能:
  - 所有自定义组件类支持主题切换
  - Ant Design 组件样式覆盖支持主题
  - 平滑的颜色过渡动画 (0.3s)

### 5. 主题切换 UI
- 位置: `src/components/Layout.vue`
- 功能:
  - 导航栏中的主题切换按钮
  - 浅色模式显示太阳图标
  - 深色模式显示月亮图标
  - 提供悬停提示文本

## 主题颜色系统

### 浅色主题
```
背景色:
- bg: #ffffff
- bg-secondary: #f8f9fa
- bg-tertiary: #f1f3f5
- bg-hover: #e9ecef

边框:
- border: #dee2e6
- border-hover: #ced4da

文字:
- text: #212529
- text-secondary: #495057
- text-tertiary: #6c757d
```

### 深色主题
```
背景色:
- bg: #000000
- bg-secondary: #0a0a0a
- bg-tertiary: #141414
- bg-hover: #1a1a1a

边框:
- border: #262626
- border-hover: #383838

文字:
- text: #e5e5e5
- text-secondary: #a3a3a3
- text-tertiary: #737373
```

### 主题色 (通用)
```
- primary: #1d9bf0 (蓝色)
- purple: #7856ff (紫色)
- success: #00ba7c (绿色)
- warning: #f9a825 (黄色)
- error: #f44336 (红色)
```

## 使用方法

### 在组件中使用主题

```vue
<template>
  <!-- 使用 Tailwind 类名 -->
  <div class="bg-light-bg dark:bg-dark-bg text-light-text dark:text-dark-text">
    内容
  </div>

  <!-- 使用主题卡片 -->
  <div class="theme-card">
    卡片内容
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

// 切换主题
themeStore.toggleTheme()

// 设置特定主题
themeStore.setTheme('dark')
themeStore.setTheme('light')

// 获取当前主题
console.log(themeStore.theme) // 'light' | 'dark'
</script>
```

### 可用的 CSS 类

#### 主题相关类
- `.theme-card` / `.grok-card` - 主题卡片
- `.glass-effect` - 玻璃态效果
- `.tech-button` - 科技感按钮
- `.tech-input` - 输入框样式
- `.gradient-text` - 渐变文字

#### 背景色类
- `bg-light-bg` / `dark:bg-dark-bg`
- `bg-light-bg-secondary` / `dark:bg-dark-bg-secondary`
- `bg-light-bg-tertiary` / `dark:bg-dark-bg-tertiary`

#### 文字颜色类
- `text-light-text` / `dark:text-dark-text`
- `text-light-text-secondary` / `dark:text-dark-text-secondary`
- `text-light-text-tertiary` / `dark:text-dark-text-tertiary`

#### 边框颜色类
- `border-light-border` / `dark:border-dark-border`
- `border-light-border-hover` / `dark:border-dark-border-hover`

## 已更新的页面

所有视图页面都已更新以支持主题切换:
- ✅ `Home.vue` - 首页
- ✅ `Bind.vue` - 绑定页面
- ✅ `Dashboard.vue` - 仪表板
- ✅ `Admin.vue` - 管理员登录
- ✅ `NotFound.vue` - 404 页面
- ✅ `Layout.vue` - 布局组件

## 测试步骤

1. 启动开发服务器:
   ```bash
   cd frontend
   npm run dev
   ```

2. 访问 http://localhost:3000

3. 测试主题切换:
   - 点击导航栏中的主题切换按钮
   - 观察页面颜色变化
   - 刷新页面，确认主题偏好被保存
   - 测试所有页面的主题一致性

4. 检查 Ant Design 组件:
   - 按钮、输入框、表格等组件颜色正确
   - 悬停、聚焦状态正常
   - 模态框、下拉菜单等浮层组件颜色正确

## 注意事项

1. 主题偏好保存在浏览器的 localStorage 中
2. 默认主题为浅色模式
3. 所有颜色变化都有 0.3s 的平滑过渡
4. 主题切换时 Ant Design 组件会自动更新
5. `grok-*` 类名是为了兼容旧代码，新代码建议使用 `light-*` 和 `dark:dark-*` 模式

## 未来改进建议

1. 添加系统主题自动检测 (prefers-color-scheme)
2. 添加更多主题选项 (如高对比度主题)
3. 为特定组件添加自定义主题变量
4. 考虑添加主题过渡动画优化
