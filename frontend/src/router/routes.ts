/**
 * Router Routes Configuration
 * 路由配置
 */

import type { RouteRecordRaw } from 'vue-router'

/**
 * 路由配置
 */
const routes: RouteRecordRaw[] = [
  // ==================== 根路由 ====================
  {
    path: '/',
    redirect: '/user/dashboard',
    meta: {
      hidden: true
    }
  },

  // ==================== 用户端路由 ====================
  {
    path: '/user',
    name: 'User',
    component: () => import('@/layouts/UserLayout.vue'),
    redirect: '/user/dashboard',
    meta: {
      title: '用户中心',
      requiresAuth: true,
      requiresAdmin: false
    },
    children: [
      {
        path: 'dashboard',
        name: 'UserDashboard',
        component: () => import('@/views/user/Dashboard.vue'),
        meta: {
          title: '仪表板',
          icon: 'DashboardOutlined',
          requiresAuth: true
        }
      },
      {
        path: 'bind',
        name: 'UserBind',
        component: () => import('@/views/user/Bind.vue'),
        meta: {
          title: '绑定账号',
          icon: 'LinkOutlined',
          requiresAuth: true
        }
      },
      {
        path: 'claim',
        name: 'UserClaim',
        component: () => import('@/views/user/Claim.vue'),
        meta: {
          title: '领取额度',
          icon: 'GiftOutlined',
          requiresAuth: true
        }
      },
      {
        path: 'donate',
        name: 'UserDonate',
        component: () => import('@/views/user/Donate.vue'),
        meta: {
          title: '投喂 Keys',
          icon: 'HeartOutlined',
          requiresAuth: true
        }
      }
    ]
  },

  // ==================== 用户登录路由 ====================
  {
    path: '/user/login',
    name: 'UserLogin',
    component: () => import('@/views/user/Login.vue'),
    meta: {
      title: '用户登录',
      requiresAuth: false,
      hidden: true
    }
  },

  // ==================== OAuth 回调路由 ====================
  {
    path: '/user/oauth/callback',
    name: 'OAuthCallback',
    component: () => import('@/views/user/OAuthCallback.vue'),
    meta: {
      title: 'OAuth 回调',
      requiresAuth: false,
      hidden: true
    }
  },

  // ==================== 管理员端路由 ====================
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/admin/dashboard',
    meta: {
      title: '管理后台',
      requiresAuth: true,
      requiresAdmin: true
    },
    children: [
      {
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: {
          title: '仪表板',
          icon: 'DashboardOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'config',
        name: 'AdminConfig',
        component: () => import('@/views/admin/Config.vue'),
        meta: {
          title: '系统配置',
          icon: 'SettingOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'keys',
        name: 'AdminKeys',
        component: () => import('@/views/admin/Keys.vue'),
        meta: {
          title: 'Keys 管理',
          icon: 'KeyOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'claims',
        name: 'AdminClaims',
        component: () => import('@/views/admin/Claims.vue'),
        meta: {
          title: '领取记录',
          icon: 'GiftOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'donates',
        name: 'AdminDonates',
        component: () => import('@/views/admin/Donates.vue'),
        meta: {
          title: '投喂记录',
          icon: 'HeartOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue'),
        meta: {
          title: '用户管理',
          icon: 'UserOutlined',
          requiresAuth: true,
          requiresAdmin: true
        }
      }
    ]
  },

  // ==================== 管理员登录路由 ====================
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue'),
    meta: {
      title: '管理员登录',
      requiresAuth: false,
      hidden: true
    }
  },

  // ==================== 错误页面路由 ====================
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: {
      title: '403 - 禁止访问',
      requiresAuth: false,
      hidden: true
    }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: {
      title: '404 - 页面未找到',
      requiresAuth: false,
      hidden: true
    }
  },
  {
    path: '/500',
    name: 'ServerError',
    component: () => import('@/views/error/500.vue'),
    meta: {
      title: '500 - 服务器错误',
      requiresAuth: false,
      hidden: true
    }
  },

  // ==================== 404 兜底路由 ====================
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
    meta: {
      hidden: true
    }
  }
]

export default routes
