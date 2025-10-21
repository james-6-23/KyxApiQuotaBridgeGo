/**
 * Admin Store
 * 管理员状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  getAdminConfig,
  updateConfig as updateAdminConfigAPI,
  getSystemStats,
  getAllUsers,
  getAllClaimRecords,
  getAllDonateRecords,
  deleteUser,
  testKeys as validateKeysAPI,
  pushKeys as pushKeysAPI
} from '@/api/admin'
import type {
  AdminConfig,
  SystemStats,
  UserStats,
  ClaimRecord,
  DonateRecord,
  DonatedKey,
  ConfigUpdateForm,
  PaginationParams,
  BatchValidationResult
} from '@/types'

export const useAdminStore = defineStore('admin', () => {
  // ==================== State ====================

  const config = ref<AdminConfig | null>(null)
  const stats = ref<SystemStats | null>(null)
  const users = ref<UserStats[]>([])
  const claimRecords = ref<ClaimRecord[]>([])
  const donateRecords = ref<DonateRecord[]>([])
  const keys = ref<DonatedKey[]>([])
  const loading = ref(false)
  const configLoading = ref(false)
  const statsLoading = ref(false)

  // 分页信息
  const userPagination = ref({
    current: 1,
    pageSize: 20,
    total: 0
  })

  const claimPagination = ref({
    current: 1,
    pageSize: 20,
    total: 0
  })

  const donatePagination = ref({
    current: 1,
    pageSize: 20,
    total: 0
  })

  const keyPagination = ref({
    current: 1,
    pageSize: 20,
    total: 0
  })

  // ==================== Getters ====================

  const isSessionConfigured = computed(() => config.value?.session_configured || false)

  const isKeysApiConfigured = computed(() => config.value?.keys_authorization_configured || false)

  const claimQuota = computed(() => config.value?.claim_quota || 0)

  const keysApiUrl = computed(() => config.value?.keys_api_url || '')

  const groupId = computed(() => config.value?.group_id || 0)

  const newApiUser = computed(() => config.value?.new_api_user || '')

  const totalUsers = computed(() => stats.value?.total_users || 0)

  const totalClaims = computed(() => stats.value?.total_claims || 0)

  const totalDonates = computed(() => stats.value?.total_donates || 0)

  const totalKeys = computed(() => stats.value?.total_keys || 0)

  const totalQuotaDistributed = computed(() => stats.value?.total_quota_distributed || 0)

  const todayClaims = computed(() => stats.value?.today_claims || 0)

  const todayDonates = computed(() => stats.value?.today_donates || 0)

  // 系统健康状态
  const systemHealth = computed(() => {
    if (!config.value) return 'unknown'

    const configured = config.value.session_configured && config.value.keys_authorization_configured
    return configured ? 'healthy' : 'warning'
  })

  // ==================== Actions ====================

  /**
   * 获取管理员配置
   */
  const fetchConfig = async (): Promise<boolean> => {
    try {
      configLoading.value = true

      const { data } = await getAdminConfig()

      if (data.success && data.data) {
        config.value = data.data
        return true
      } else {
        message.error(data.message || '获取配置失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch admin config failed:', error)
      message.error(error.message || '获取配置失败')
      return false
    } finally {
      configLoading.value = false
    }
  }

  /**
   * 更新管理员配置
   */
  const updateConfig = async (form: ConfigUpdateForm): Promise<boolean> => {
    try {
      configLoading.value = true

      const { data } = await updateAdminConfigAPI(form)

      if (data.success) {
        message.success(data.message || '配置更新成功')

        // 重新获取配置
        await fetchConfig()

        return true
      } else {
        message.error(data.message || '配置更新失败')
        return false
      }
    } catch (error: any) {
      console.error('Update admin config failed:', error)
      message.error(error.message || '配置更新失败')
      return false
    } finally {
      configLoading.value = false
    }
  }

  /**
   * 获取系统统计信息
   */
  const fetchStats = async (): Promise<boolean> => {
    try {
      statsLoading.value = true

      const { data } = await getSystemStats()

      if (data.success && data.data) {
        stats.value = data.data
        return true
      } else {
        message.error(data.message || '获取统计信息失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch system stats failed:', error)
      message.error(error.message || '获取统计信息失败')
      return false
    } finally {
      statsLoading.value = false
    }
  }

  /**
   * 获取所有用户
   */
  const fetchUsers = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: userPagination.value.current,
        page_size: userPagination.value.pageSize
      }

      const { data } = await getAllUsers(pagination)

      if (data.success && data.data) {
        users.value = data.data.items || []
        userPagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 20,
          total: data.data.total || 0
        }
        return true
      } else {
        message.error(data.message || '获取用户列表失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch users failed:', error)
      message.error(error.message || '获取用户列表失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取所有领取记录
   */
  const fetchClaimRecords = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: claimPagination.value.current,
        page_size: claimPagination.value.pageSize
      }

      const { data } = await getAllClaimRecords(pagination)

      if (data.success && data.data) {
        claimRecords.value = data.data.items || []
        claimPagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 20,
          total: data.data.total || 0
        }
        return true
      } else {
        message.error(data.message || '获取领取记录失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch claim records failed:', error)
      message.error(error.message || '获取领取记录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取所有投喂记录
   */
  const fetchDonateRecords = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: donatePagination.value.current,
        page_size: donatePagination.value.pageSize
      }

      const { data } = await getAllDonateRecords(pagination)

      if (data.success && data.data) {
        donateRecords.value = data.data.items || []
        donatePagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 20,
          total: data.data.total || 0
        }
        return true
      } else {
        message.error(data.message || '获取投喂记录失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch donate records failed:', error)
      message.error(error.message || '获取投喂记录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取所有 Keys
   */
  const fetchKeys = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: keyPagination.value.current,
        page_size: keyPagination.value.pageSize
      }

      const { data } = await getAllKeys(pagination)

      if (data.success && data.data) {
        keys.value = data.data.items || []
        keyPagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 20,
          total: data.data.total || 0
        }
        return true
      } else {
        message.error(data.message || '获取 Keys 失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch keys failed:', error)
      message.error(error.message || '获取 Keys 失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除用户
   */
  const removeUser = async (linuxDoId: string): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await deleteUser(linuxDoId)

      if (data.success) {
        message.success(data.message || '删除用户成功')

        // 刷新用户列表
        await fetchUsers()
        await fetchStats()

        return true
      } else {
        message.error(data.message || '删除用户失败')
        return false
      }
    } catch (error: any) {
      console.error('Delete user failed:', error)
      message.error(error.message || '删除用户失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除领取记录
   */
  const removeClaimRecord = async (id: number): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await deleteClaimRecord(id)

      if (data.success) {
        message.success(data.message || '删除记录成功')

        // 刷新领取记录
        await fetchClaimRecords()
        await fetchStats()

        return true
      } else {
        message.error(data.message || '删除记录失败')
        return false
      }
    } catch (error: any) {
      console.error('Delete claim record failed:', error)
      message.error(error.message || '删除记录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除投喂记录
   */
  const removeDonateRecord = async (id: number): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await deleteDonateRecord(id)

      if (data.success) {
        message.success(data.message || '删除记录成功')

        // 刷新投喂记录
        await fetchDonateRecords()
        await fetchStats()

        return true
      } else {
        message.error(data.message || '删除记录失败')
        return false
      }
    } catch (error: any) {
      console.error('Delete donate record failed:', error)
      message.error(error.message || '删除记录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 验证 Keys
   */
  const verifyKeys = async (keyList: string[]): Promise<BatchValidationResult | null> => {
    try {
      loading.value = true

      const { data } = await validateKeysAPI(keyList)

      if (data.success && data.data) {
        message.success('Keys 验证完成')
        return data.data
      } else {
        message.error(data.message || 'Keys 验证失败')
        return null
      }
    } catch (error: any) {
      console.error('Validate keys failed:', error)
      message.error(error.message || 'Keys 验证失败')
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 推送 Keys 到 KYX
   */
  const pushKeys = async (): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await pushKeysAPI(selectedKeys.value)

      if (data.success) {
        message.success(data.message || '推送成功')

        // 刷新 Keys 列表和统计信息
        await fetchKeys()
        await fetchStats()

        return true
      } else {
        message.error(data.message || '推送失败')
        return false
      }
    } catch (error: any) {
      console.error('Push keys failed:', error)
      message.error(error.message || '推送失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新所有管理员数据
   */
  const refreshAllData = async (): Promise<void> => {
    await Promise.all([
      fetchConfig(),
      fetchStats(),
      fetchUsers(),
      fetchClaimRecords(),
      fetchDonateRecords(),
      fetchKeys()
    ])
  }

  /**
   * 刷新仪表板数据（配置和统计）
   */
  const refreshDashboardData = async (): Promise<void> => {
    await Promise.all([
      fetchConfig(),
      fetchStats()
    ])
  }

  /**
   * 清空管理员数据
   */
  const clearAdminData = () => {
    config.value = null
    stats.value = null
    users.value = []
    claimRecords.value = []
    donateRecords.value = []
    keys.value = []

    userPagination.value = { current: 1, pageSize: 20, total: 0 }
    claimPagination.value = { current: 1, pageSize: 20, total: 0 }
    donatePagination.value = { current: 1, pageSize: 20, total: 0 }
    keyPagination.value = { current: 1, pageSize: 20, total: 0 }
  }

  /**
   * 更新用户分页
   */
  const updateUserPagination = async (page: number, pageSize: number): Promise<void> => {
    userPagination.value.current = page
    userPagination.value.pageSize = pageSize
    await fetchUsers({ page, page_size: pageSize })
  }

  /**
   * 更新领取记录分页
   */
  const updateClaimPagination = async (page: number, pageSize: number): Promise<void> => {
    claimPagination.value.current = page
    claimPagination.value.pageSize = pageSize
    await fetchClaimRecords({ page, page_size: pageSize })
  }

  /**
   * 更新投喂记录分页
   */
  const updateDonatePagination = async (page: number, pageSize: number): Promise<void> => {
    donatePagination.value.current = page
    donatePagination.value.pageSize = pageSize
    await fetchDonateRecords({ page, page_size: pageSize })
  }

  /**
   * 更新 Key 分页
   */
  const updateKeyPagination = async (page: number, pageSize: number): Promise<void> => {
    keyPagination.value.current = page
    keyPagination.value.pageSize = pageSize
    await fetchKeys({ page, page_size: pageSize })
  }

  // ==================== Return ====================

  return {
    // State
    config,
    stats,
    users,
    claimRecords,
    donateRecords,
    keys,
    loading,
    configLoading,
    statsLoading,
    userPagination,
    claimPagination,
    donatePagination,
    keyPagination,

    // Getters
    isSessionConfigured,
    isKeysApiConfigured,
    claimQuota,
    keysApiUrl,
    groupId,
    newApiUser,
    totalUsers,
    totalClaims,
    totalDonates,
    totalKeys,
    totalQuotaDistributed,
    todayClaims,
    todayDonates,
    systemHealth,

    // Actions
    fetchConfig,
    updateConfig,
    fetchStats,
    fetchUsers,
    fetchClaimRecords,
    fetchDonateRecords,
    fetchKeys,
    removeUser,
    removeClaimRecord,
    removeDonateRecord,
    verifyKeys,
    pushKeys,
    refreshAllData,
    refreshDashboardData,
    clearAdminData,
    updateUserPagination,
    updateClaimPagination,
    updateDonatePagination,
    updateKeyPagination
  }
})
