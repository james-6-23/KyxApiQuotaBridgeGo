<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ConfigProvider } from 'ant-design-vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

// 设置 dayjs 中文
dayjs.locale('zh-cn')

const router = useRouter()
const route = useRoute()

// Ant Design Vue 主题配置
const theme = computed(() => ({
  token: {
    colorPrimary: '#1890ff',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#f5222d',
    colorInfo: '#1890ff',
    colorLink: '#1890ff',
    fontSize: 14,
    borderRadius: 4,
    wireframe: false,
  },
  algorithm: undefined, // 可以设置为 theme.darkAlgorithm 实现暗色模式
}))

// 监听路由变化
router.beforeEach((to, from, next) => {
  // 设置页面标题
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} - KYX API Quota Bridge`
  } else {
    document.title = 'KYX API Quota Bridge'
  }

  next()
})
</script>

<template>
  <ConfigProvider :locale="zhCN" :theme="theme">
    <div id="app" class="min-h-screen bg-gray-50">
      <!-- Router View -->
      <router-view v-slot="{ Component, route }">
        <transition
          name="fade"
          mode="out-in"
          appear
        >
          <component :is="Component" :key="route.path" />
        </transition>
      </router-view>
    </div>
  </ConfigProvider>
</template>

<style scoped>
/* Page Transition Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>

<style>
/* Global Styles */
#app {
  width: 100%;
  min-height: 100vh;
}

/* Ant Design Vue Custom Overrides */
.ant-btn {
  transition: all 0.3s cubic-bezier(0.645, 0.045, 0.355, 1);
}

.ant-card {
  transition: all 0.3s cubic-bezier(0.645, 0.045, 0.355, 1);
}

.ant-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* Loading State */
.ant-spin-container {
  transition: opacity 0.3s;
}

/* Modal Animations */
.ant-modal {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Message Positioning */
.ant-message {
  z-index: 9999;
}

/* Notification Positioning */
.ant-notification {
  z-index: 9999;
}

/* Custom Scrollbar for Content Areas */
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: #bfbfbf #f0f0f0;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: #f0f0f0;
  border-radius: 4px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #bfbfbf;
  border-radius: 4px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #999;
}

/* Print Styles */
@media print {
  .no-print {
    display: none !important;
  }
}

/* Responsive Typography */
@media (max-width: 768px) {
  html {
    font-size: 14px;
  }
}

@media (max-width: 480px) {
  html {
    font-size: 13px;
  }
}

/* Focus Visible Styles */
*:focus-visible {
  outline: 2px solid #1890ff;
  outline-offset: 2px;
}

/* Reduce Motion for Accessibility */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* High Contrast Mode Support */
@media (prefers-contrast: high) {
  .ant-btn {
    border-width: 2px;
  }
}

/* Dark Mode Support (Future Implementation) */
@media (prefers-color-scheme: dark) {
  /* Dark mode styles will be added here */
}

/* Selection Colors */
::selection {
  background-color: #bae7ff;
  color: #003a8c;
}

::-moz-selection {
  background-color: #bae7ff;
  color: #003a8c;
}
</style>
