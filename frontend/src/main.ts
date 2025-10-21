/**
 * Vue 3 Application Entry Point
 * 应用程序入口文件
 */

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import App from './App.vue'
import router from './router'

// Ant Design Vue
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

// Tailwind CSS & Custom Styles
import './assets/styles/tailwind.css'

// NProgress
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

// NProgress Configuration
NProgress.configure({
  showSpinner: false,
  trickleSpeed: 200,
  minimum: 0.3,
  easing: 'ease',
  speed: 500
})

// Create Vue App
const app = createApp(App)

// Create Pinia Store with Persistence
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

// Use Plugins
app.use(pinia)
app.use(router)
app.use(Antd)

// Global Properties
app.config.globalProperties.$appName = import.meta.env.VITE_APP_TITLE || 'KYX API Quota Bridge'

// Error Handler
app.config.errorHandler = (err, instance, info) => {
  console.error('Global Error:', err)
  console.error('Component:', instance)
  console.error('Error Info:', info)

  // 可以在这里添加错误上报逻辑
  // 例如：上报到 Sentry 等错误监控服务
}

// Warning Handler (Development Only)
if (import.meta.env.DEV) {
  app.config.warnHandler = (msg, instance, trace) => {
    console.warn('Warning:', msg)
    console.warn('Trace:', trace)
  }
}

// Performance Monitoring (Development Only)
if (import.meta.env.DEV) {
  app.config.performance = true
}

// Mount App
app.mount('#app')

// Log App Info
console.log(`
╔═══════════════════════════════════════════════╗
║                                               ║
║   🚀 KYX API Quota Bridge                    ║
║   📦 Version: 1.0.0                          ║
║   🎯 Environment: ${import.meta.env.MODE.toUpperCase().padEnd(25)}    ║
║   ⚡ Powered by Vue 3 + Vite                 ║
║                                               ║
╚═══════════════════════════════════════════════╝
`)

// Hot Module Replacement
if (import.meta.hot) {
  import.meta.hot.accept()
}
