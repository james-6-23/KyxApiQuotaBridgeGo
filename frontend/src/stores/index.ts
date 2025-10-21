/**
 * Stores Index
 * 统一导出所有 Store
 */

export { useAuthStore } from './auth'
export { useUserStore } from './user'
export { useAdminStore } from './admin'
export { useAppStore } from './app'

// 类型导出
export type { AuthState, AppState, AdminState } from '@/types'
