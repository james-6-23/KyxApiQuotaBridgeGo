/**
 * Router Guards
 * 路由守卫配置
 */

import type { Router } from 'vue-router'
import NProgress from 'nprogress'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { message } from 'ant-design-vue'

/**
 * 设置路由守卫
 */
export function setupRouterGuards(router: Router) {
  // 创建前置守卫
  router.beforeEach(async (to, _from, next) => {
    // 启动进度条
    NProgress.start()

    const authStore = useAuthStore()
    const appStore = useAppStore()

    // 设置页面标题
    const pageTitle = (to.meta.title as string) || ''
    appStore.setPageTitle(pageTitle)

    // 公开路由白名单
    const whiteList = [
      '/user/login',
      '/user/oauth/callback',
      '/admin/login',
      '/404',
      '/403',
      '/500'
    ]

    // 如果是白名单路由，直接放行
    if (whiteList.includes(to.path)) {
      next()
      return
    }

    // 处理根路径重定向
    if (to.path === '/') {
      // 如果已登录，根据角色重定向
      if (authStore.isAuthenticated) {
        if (authStore.isAdmin) {
          next('/admin/dashboard')
        } else {
          next('/user/dashboard')
        }
      } else {
        // 未登录，重定向到用户登录页
        next('/user/login')
      }
      return
    }

    // 检查是否需要认证
    const requiresAuth = to.meta.requiresAuth !== false // 默认需要认证

    if (requiresAuth) {
      // 检查是否已登录
      if (!authStore.isAuthenticated) {
        // 尝试从 localStorage 恢复认证状态
        const restored = await authStore.checkAuthStatus()

        if (!restored) {
          message.warning('请先登录')

          // 根据目标路由判断应该跳转到哪个登录页
          if (to.path.startsWith('/admin')) {
            next({
              path: '/admin/login',
              query: { redirect: to.fullPath }
            })
          } else {
            next({
              path: '/user/login',
              query: { redirect: to.fullPath }
            })
          }
          return
        }
      }

      // 检查管理员权限
      if (to.meta.requiresAdmin) {
        if (!authStore.isAdmin) {
          message.error('您没有权限访问此页面')
          next('/403')
          return
        }
      }

      // 如果是管理员，但访问的是用户页面，重定向到管理员页面
      if (authStore.isAdmin && to.path.startsWith('/user') && !to.path.startsWith('/user/oauth')) {
        next('/admin/dashboard')
        return
      }

      // 如果是普通用户，但访问的是管理员页面，拒绝访问
      if (!authStore.isAdmin && to.path.startsWith('/admin')) {
        message.error('您没有权限访问管理员页面')
        next('/403')
        return
      }
    }

    // 通过所有检查，放行
    next()
  })

  // 创建后置守卫
  router.afterEach((to) => {
    // 关闭进度条
    NProgress.done()

    // 更新面包屑导航（可选）
    const appStore = useAppStore()
    const breadcrumbs = generateBreadcrumbs(to)
    appStore.setBreadcrumbs(breadcrumbs)

    // 在移动端导航后自动收起侧边栏
    if (appStore.isMobile && !appStore.isSidebarCollapsed) {
      appStore.setSidebarCollapsed(true)
    }
  })

  // 创建错误守卫
  router.onError((error) => {
    console.error('Router error:', error)
    message.error('页面加载失败，请刷新重试')
    NProgress.done()
  })
}

/**
 * 生成面包屑导航
 */
function generateBreadcrumbs(route: any): Array<{ name: string; path?: string }> {
  const breadcrumbs: Array<{ name: string; path?: string }> = []

  // 添加首页
  if (route.path.startsWith('/admin')) {
    breadcrumbs.push({ name: '管理后台', path: '/admin/dashboard' })
  } else if (route.path.startsWith('/user')) {
    breadcrumbs.push({ name: '首页', path: '/user/dashboard' })
  }

  // 如果有 matched 路由，遍历生成面包屑
  if (route.matched && route.matched.length > 0) {
    route.matched.forEach((matchedRoute: any) => {
      // 跳过根路由和重定向路由
      if (!matchedRoute.meta || !matchedRoute.meta.title || matchedRoute.redirect) {
        return
      }

      // 不添加重复的面包屑
      const exists = breadcrumbs.some(item => item.path === matchedRoute.path)
      if (!exists && matchedRoute.meta.title) {
        breadcrumbs.push({
          name: matchedRoute.meta.title,
          path: matchedRoute.path
        })
      }
    })
  }

  // 如果最后一个面包屑就是当前页面，移除其路径（不可点击）
  if (breadcrumbs.length > 0) {
    const lastBreadcrumb = breadcrumbs[breadcrumbs.length - 1]
    if (lastBreadcrumb.path === route.path) {
      delete lastBreadcrumb.path
    }
  }

  return breadcrumbs
}

/**
 * 路由守卫钩子
 */
export function useRouterGuards() {
  const authStore = useAuthStore()

  /**
   * 检查当前路由是否需要认证
   */
  const requiresAuth = (route: any): boolean => {
    return route.meta?.requiresAuth !== false
  }

  /**
   * 检查当前路由是否需要管理员权限
   */
  const requiresAdmin = (route: any): boolean => {
    return route.meta?.requiresAdmin === true
  }

  /**
   * 检查用户是否有权限访问路由
   */
  const canAccess = (route: any): boolean => {
    // 不需要认证的路由，直接允许
    if (!requiresAuth(route)) {
      return true
    }

    // 需要认证但未登录
    if (!authStore.isAuthenticated) {
      return false
    }

    // 需要管理员权限但用户不是管理员
    if (requiresAdmin(route) && !authStore.isAdmin) {
      return false
    }

    return true
  }

  /**
   * 获取登录重定向路径
   */
  const getLoginPath = (targetPath: string): string => {
    if (targetPath.startsWith('/admin')) {
      return '/admin/login'
    }
    return '/user/login'
  }

  /**
   * 获取登录后的重定向路径
   */
  const getRedirectPath = (): string => {
    if (authStore.isAdmin) {
      return '/admin/dashboard'
    }
    return '/user/dashboard'
  }

  return {
    requiresAuth,
    requiresAdmin,
    canAccess,
    getLoginPath,
    getRedirectPath
  }
}
