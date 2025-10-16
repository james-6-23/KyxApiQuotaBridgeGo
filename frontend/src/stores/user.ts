import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi, authApi, type UserInfo } from '@/api'

export const useUserStore = defineStore('user', () => {
  // 状态
  const userInfo = ref<UserInfo | null>(null)
  const isInitialized = ref(false)
  const isAdmin = ref(false)

  // 计算属性
  const isLoggedIn = computed(() => !!userInfo.value)
  const isBound = computed(() => isLoggedIn.value && !!userInfo.value?.username)

  // 检查认证状态
  const checkAuth = async () => {
    try {
      const response = await userApi.getQuota()
      if (response.success && response.data) {
        userInfo.value = response.data
      }
    } catch (error) {
      // 未登录或会话过期
      userInfo.value = null
    } finally {
      isInitialized.value = true
    }
  }

  // 刷新用户信息
  const refreshUserInfo = async () => {
    if (!isLoggedIn.value) return

    try {
      const response = await userApi.getQuota()
      if (response.success && response.data) {
        userInfo.value = response.data
      }
    } catch (error) {
      console.error('Failed to refresh user info:', error)
    }
  }

  // 登录 (OAuth 重定向)
  const login = () => {
    const clientId = import.meta.env.VITE_OAUTH_CLIENT_ID || 'your-client-id'
    const redirectUri = encodeURIComponent(window.location.origin + '/oauth/callback')
    const oauthUrl = `https://connect.linux.do/oauth/authorize?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=read`
    window.location.href = oauthUrl
  }

  // 退出登录
  const logout = async () => {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      userInfo.value = null
      isAdmin.value = false
    }
  }

  // 设置管理员状态
  const setAdminStatus = (status: boolean) => {
    isAdmin.value = status
  }

  return {
    userInfo,
    isInitialized,
    isAdmin,
    isLoggedIn,
    isBound,
    checkAuth,
    refreshUserInfo,
    login,
    logout,
    setAdminStatus,
  }
})
