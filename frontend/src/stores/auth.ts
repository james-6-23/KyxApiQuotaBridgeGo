/**
 * Auth Store
 * 认证状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  checkAuth,
  logout as logoutApi,
  adminLogin as adminLoginApi,
  handleOAuthCallback
} from '@/api/auth'
import type { User, LoginForm } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const router = useRouter()

  // ==================== State ====================

  const user = ref<User | null>(null)
  const token = ref<string>('')
  const loading = ref(false)
  const isAdmin = ref(false)

  // ==================== Getters ====================

  const isAuthenticated = computed(() => !!user.value && !!token.value)

  const username = computed(() => user.value?.username || '')

  const linuxDoId = computed(() => user.value?.linux_do_id || '')

  const userInfo = computed(() => user.value)

  // ==================== Actions ====================

  /**
   * 设置用户信息
   */
  const setUser = (userData: User) => {
    user.value = userData
    // 同步到 localStorage
    localStorage.setItem('user', JSON.stringify(userData))
  }

  /**
   * 设置 Token
   */
  const setToken = (tokenValue: string) => {
    token.value = tokenValue
    localStorage.setItem('token', tokenValue)
  }

  /**
   * 设置管理员状态
   */
  const setAdminStatus = (status: boolean) => {
    isAdmin.value = status
    localStorage.setItem('isAdmin', String(status))
  }

  /**
   * 检查用户登录状态
   * 通过调用后端API验证Cookie中的session
   */
  const checkAuthStatus = async (): Promise<boolean> => {
    // 如果已经有用户信息，直接返回
    if (user.value) {
      return true
    }

    try {
      loading.value = true

      // 调用后端API验证Cookie中的session
      const { data } = await checkAuth()

      if (data.success && data.data && data.data.authenticated) {
        // 登录有效，保存用户信息
        if (data.data.user) {
          user.value = data.data.user
          // 设置一个假的token标记(因为实际token在HttpOnly Cookie中)
          token.value = 'session-cookie'
          return true
        }
      }

      // Session无效或已过期
      clearAuth()
      return false
    } catch (error) {
      console.error('Check auth status failed:', error)
      clearAuth()
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 处理 OAuth 回调
   */
  const handleOAuthLogin = async (code: string, state?: string): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await handleOAuthCallback(code, state)

      if (data.success && data.data) {
        // 设置用户信息
        setUser(data.data.user || data.data)
        // 设置token标记(实际token在HttpOnly Cookie中)
        setToken('session-cookie')
        setAdminStatus(false)
        message.success('登录成功')
        return true
      } else {
        message.error(data.message || '登录失败')
        return false
      }
    } catch (error: any) {
      console.error('OAuth login failed:', error)
      message.error(error.message || '登录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 管理员登录
   */
  const adminLogin = async (loginForm: LoginForm): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await adminLoginApi(loginForm)

      if (data.success && data.data) {
        // 设置 Token
        if (data.data.token) {
          setToken(data.data.token)
        }

        // 设置用户信息（如果有）
        if (data.data.user) {
          setUser(data.data.user)
        } else {
          // 如果没有返回用户信息，创建一个管理员用户对象
          setUser({
            username: 'admin',
            linux_do_id: 'admin',
            created_at: new Date().toISOString()
          })
        }

        setAdminStatus(true)
        message.success('管理员登录成功')
        return true
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
   * 登出
   */
  const logout = async (redirect = true) => {
    try {
      loading.value = true

      // 调用后端登出接口
      await logoutApi()

      message.success('已退出登录')
    } catch (error) {
      console.error('Logout failed:', error)
      // 即使后端登出失败，也要清除本地数据
    } finally {
      // 清除本地状态
      clearAuth()
      loading.value = false

      // 重定向到登录页
      if (redirect) {
        if (isAdmin.value) {
          router.push('/admin/login')
        } else {
          router.push('/user/login')
        }
      }
    }
  }

  /**
   * 清除认证信息
   */
  const clearAuth = () => {
    user.value = null
    token.value = ''
    isAdmin.value = false

    // 清除 localStorage
    localStorage.removeItem('user')
    localStorage.removeItem('token')
    localStorage.removeItem('isAdmin')
  }

  /**
   * 刷新用户信息
   */
  const refreshUserInfo = async (): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await checkAuth()

      if (data.success && data.data) {
        setUser(data.data)
        return true
      }

      return false
    } catch (error) {
      console.error('Refresh user info failed:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 初始化认证状态（从 localStorage 恢复）
   */
  const initAuth = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    const savedIsAdmin = localStorage.getItem('isAdmin')

    if (savedToken) {
      token.value = savedToken
    }

    if (savedUser) {
      try {
        user.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('Parse saved user failed:', error)
        localStorage.removeItem('user')
      }
    }

    if (savedIsAdmin) {
      isAdmin.value = savedIsAdmin === 'true'
    }
  }

  // ==================== Return ====================

  return {
    // State
    user,
    token,
    loading,
    isAdmin,

    // Getters
    isAuthenticated,
    username,
    linuxDoId,
    userInfo,

    // Actions
    setUser,
    setToken,
    setAdminStatus,
    checkAuthStatus,
    handleOAuthLogin,
    adminLogin,
    logout,
    clearAuth,
    refreshUserInfo,
    initAuth
  }
}, {
  // 持久化配置
  persist: {
    key: 'auth-store',
    storage: localStorage,
    paths: ['user', 'token', 'isAdmin']
  }
})
