/**
 * Vue Router Configuration
 * Vue Router 配置入口
 */

import { createRouter, createWebHistory } from 'vue-router'
import type { App } from 'vue'
import routes from './routes'
import { setupRouterGuards } from './guards'

/**
 * 创建 Router 实例
 */
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    // 如果有保存的位置（浏览器前进/后退）
    if (savedPosition) {
      return savedPosition
    }

    // 如果有锚点
    if (to.hash) {
      return {
        el: to.hash,
        behavior: 'smooth'
      }
    }

    // 默认滚动到顶部
    return { top: 0, behavior: 'smooth' }
  }
})

/**
 * 配置路由守卫
 */
setupRouterGuards(router)

/**
 * 安装 Router
 */
export function setupRouter(app: App) {
  app.use(router)
}

/**
 * 重置路由
 */
export function resetRouter() {
  const newRouter = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
  })

  // 替换 matcher
  ;(router as any).matcher = (newRouter as any).matcher
}

export default router
