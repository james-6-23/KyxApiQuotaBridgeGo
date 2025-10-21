/**
 * User Store
 * 用户状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  getUserQuota,
  getUserStats,
  bindKyxAccount,
  claimDailyQuota,
  donateKeys,
  getUserClaimRecords,
  getUserDonateRecords
} from '@/api/user'
import type {
  UserQuota,
  UserStats,
  ClaimRecord,
  DonateRecord,
  BindAccountForm,
  DonateForm,
  PaginationParams
} from '@/types'

export const useUserStore = defineStore('user', () => {
  // ==================== State ====================

  const quota = ref<UserQuota | null>(null)
  const stats = ref<UserStats | null>(null)
  const claimRecords = ref<ClaimRecord[]>([])
  const donateRecords = ref<DonateRecord[]>([])
  const loading = ref(false)
  const claimLoading = ref(false)
  const donateLoading = ref(false)

  // 分页信息
  const claimPagination = ref({
    current: 1,
    pageSize: 10,
    total: 0
  })

  const donatePagination = ref({
    current: 1,
    pageSize: 10,
    total: 0
  })

  // ==================== Getters ====================

  const isBound = computed(() => quota.value?.is_bound || false)

  const canClaimToday = computed(() => quota.value?.can_claim_today || false)

  const currentBalance = computed(() => quota.value?.balance || 0)

  const currentQuota = computed(() => quota.value?.quota || 0)

  const totalDonated = computed(() => quota.value?.total_donated || 0)

  const totalClaimed = computed(() => quota.value?.total_claimed || 0)

  const kyxUsername = computed(() => quota.value?.kyx_username || '')

  const lastClaimDate = computed(() => quota.value?.last_claim_date || '')

  // 用户统计
  const userStats = computed(() => {
    if (!stats.value) return null

    return {
      donateCount: stats.value.donate_count,
      donateQuota: stats.value.donate_quota,
      claimCount: stats.value.claim_count,
      claimQuota: stats.value.claim_quota,
      totalQuota: stats.value.total_quota,
      createdAt: stats.value.created_at
    }
  })

  // ==================== Actions ====================

  /**
   * 获取用户额度信息
   */
  const fetchUserQuota = async (): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await getUserQuota()

      if (data.success && data.data) {
        quota.value = data.data
        return true
      } else {
        message.error(data.message || '获取额度信息失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch user quota failed:', error)
      message.error(error.message || '获取额度信息失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取用户统计信息
   */
  const fetchUserStats = async (): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await getUserStats()

      if (data.success && data.data) {
        stats.value = data.data
        return true
      } else {
        message.error(data.message || '获取统计信息失败')
        return false
      }
    } catch (error: any) {
      console.error('Fetch user stats failed:', error)
      message.error(error.message || '获取统计信息失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 绑定 KYX 账号
   */
  const bindAccount = async (form: BindAccountForm): Promise<boolean> => {
    try {
      loading.value = true

      const { data } = await bindKyxAccount(form)

      if (data.success) {
        message.success(data.message || '绑定成功')

        // 重新获取额度信息
        await fetchUserQuota()

        return true
      } else {
        message.error(data.message || '绑定失败')
        return false
      }
    } catch (error: any) {
      console.error('Bind account failed:', error)
      message.error(error.message || '绑定失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 领取每日额度
   */
  const claimQuota = async (): Promise<boolean> => {
    try {
      claimLoading.value = true

      const { data } = await claimDailyQuota()

      if (data.success) {
        message.success(data.message || '领取成功')

        // 重新获取额度信息
        await fetchUserQuota()

        // 刷新领取记录
        await fetchClaimRecords({ page: 1, page_size: claimPagination.value.pageSize })

        return true
      } else {
        message.error(data.message || '领取失败')
        return false
      }
    } catch (error: any) {
      console.error('Claim quota failed:', error)
      message.error(error.message || '领取失败')
      return false
    } finally {
      claimLoading.value = false
    }
  }

  /**
   * 投喂 Keys
   */
  const donate = async (form: DonateForm): Promise<boolean> => {
    try {
      donateLoading.value = true

      const { data } = await donateKeys(form)

      if (data.success) {
        message.success(data.message || '投喂成功')

        // 重新获取额度信息和统计信息
        await fetchUserQuota()
        await fetchUserStats()

        // 刷新投喂记录
        await fetchDonateRecords({ page: 1, page_size: donatePagination.value.pageSize })

        return true
      } else {
        message.error(data.message || '投喂失败')
        return false
      }
    } catch (error: any) {
      console.error('Donate keys failed:', error)
      message.error(error.message || '投喂失败')
      return false
    } finally {
      donateLoading.value = false
    }
  }

  /**
   * 获取领取记录
   */
  const fetchClaimRecords = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: claimPagination.value.current,
        page_size: claimPagination.value.pageSize
      }

      const { data } = await getUserClaimRecords(pagination)

      if (data.success && data.data) {
        claimRecords.value = data.data.items || []
        claimPagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 10,
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
   * 获取投喂记录
   */
  const fetchDonateRecords = async (params?: PaginationParams): Promise<boolean> => {
    try {
      loading.value = true

      const pagination: PaginationParams = params || {
        page: donatePagination.value.current,
        page_size: donatePagination.value.pageSize
      }

      const { data } = await getUserDonateRecords(pagination)

      if (data.success && data.data) {
        donateRecords.value = data.data.items || []
        donatePagination.value = {
          current: data.data.page || 1,
          pageSize: data.data.page_size || 10,
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
   * 刷新所有用户数据
   */
  const refreshAllData = async (): Promise<void> => {
    await Promise.all([
      fetchUserQuota(),
      fetchUserStats(),
      fetchClaimRecords(),
      fetchDonateRecords()
    ])
  }

  /**
   * 清空用户数据
   */
  const clearUserData = () => {
    quota.value = null
    stats.value = null
    claimRecords.value = []
    donateRecords.value = []
    claimPagination.value = {
      current: 1,
      pageSize: 10,
      total: 0
    }
    donatePagination.value = {
      current: 1,
      pageSize: 10,
      total: 0
    }
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

  // ==================== Return ====================

  return {
    // State
    quota,
    stats,
    claimRecords,
    donateRecords,
    loading,
    claimLoading,
    donateLoading,
    claimPagination,
    donatePagination,

    // Getters
    isBound,
    canClaimToday,
    currentBalance,
    currentQuota,
    totalDonated,
    totalClaimed,
    kyxUsername,
    lastClaimDate,
    userStats,

    // Actions
    fetchUserQuota,
    fetchUserStats,
    bindAccount,
    claimQuota,
    donate,
    fetchClaimRecords,
    fetchDonateRecords,
    refreshAllData,
    clearUserData,
    updateClaimPagination,
    updateDonatePagination
  }
})
