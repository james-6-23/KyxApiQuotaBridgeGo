/**
 * Admin Auth Store
 * 管理员认证状态管理（独立于用户认证）
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { adminLogin as adminLoginApi } from '@/api/auth'
import type { LoginForm } from '@/types'

export const useAdminAuthStore = defineStore('adminAuth', () => {
  const router = useRouter()

  // ==================== State ====================

  const adminToken = ref<string>('')
  const loading = ref(false)

  // ==================== Getters ====================

  const isAdminAuthenticated = computed(() => !!adminToken.value)

  // ==================== Actions ====================

  /**
   * 设置管理员 Token
   */
  const setAdminToken = (token: string) => {
    adminToken.value = token
    localStorage.setItem('adminToken', token)
  }

  /**
   * 管理员登录
   */
  const adminLogin = async (loginForm: LoginForm): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await adminLoginApi(loginForm)

      if (data.success && data.data) {
        // 设置管理员 Token
        if (data.data.token) {
          setAdminToken(data.data.token)
          message.success('管理员登录成功')
          return true
        } else {
          message.error('登录失败：未返回Token')
          return false
        }
      } else {
        message.error(data.message || '登录失败')
        return false
      }
    } catch (error: any) {
      console.error('Admin login failed:', error)
      message.error(error.message || '登录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 管理员登出
   */
  const adminLogout = (redirect = true) => {
    adminToken.value = ''
    localStorage.removeItem('adminToken')
    message.success('已退出管理员登录')

    if (redirect) {
      router.push('/admin/login')
    }
  }

  /**
   * 清除管理员认证信息
   */
  const clearAdminAuth = () => {
    adminToken.value = ''
    localStorage.removeItem('adminToken')
  }

  /**
   * 初始化管理员认证状态（从 localStorage 恢复）
   */
  const initAdminAuth = () => {
    const savedToken = localStorage.getItem('adminToken')
    if (savedToken) {
      adminToken.value = savedToken
    }
  }

  // ==================== Return ====================

  return {
    // State
    adminToken,
    loading,

    // Getters
    isAdminAuthenticated,

    // Actions
    setAdminToken,
    adminLogin,
    adminLogout,
    clearAdminAuth,
    initAdminAuth
  }
}, {
  // 持久化配置
  persist: {
    key: 'admin-auth-store',
    storage: localStorage,
    paths: ['adminToken']
  }
})
