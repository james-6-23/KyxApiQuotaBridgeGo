/**
 * API Index Module
 * 统一导出所有 API 模块
 */

// Export HTTP Client
export { request } from './request'
export type { RequestConfig } from './request'

// Export Auth API
export * from './auth'

// Export User API
export * from './user'

// Export Admin API
export * from './admin'

// Re-export Types
export type {
  ApiResponse,
  User,
  UserQuota,
  UserStats,
  ClaimRecord,
  DonateRecord,
  DonatedKey,
  AdminConfig,
  SystemStats,
  OAuthUrl,
  LoginForm,
  BindAccountForm,
  DonateForm,
  ConfigUpdateForm,
  KeyValidation,
  BatchValidationResult,
  PaginationParams,
  PaginatedResponse
} from '@/types'

/**
 * API 模块说明：
 *
 * 认证相关：
 * - getOAuthUrl() - 获取 OAuth 登录 URL
 * - handleOAuthCallback() - 处理 OAuth 回调
 * - checkAuth() - 检查登录状态
 * - logout() - 登出
 * - adminLogin() - 管理员登录
 *
 * 用户相关：
 * - getUserQuota() - 获取用户额度
 * - bindAccount() - 绑定账号
 * - claimDailyQuota() - 每日领取
 * - donateKeys() - 投喂 Keys
 * - getUserClaimRecords() - 获取领取记录
 * - getUserDonateRecords() - 获取投喂记录
 *
 * 管理员相关：
 * - getAdminConfig() - 获取配置
 * - updateQuota() - 更新额度
 * - updateSession() - 更新 Session
 * - exportKeys() - 导出 Keys
 * - deleteKeys() - 删除 Keys
 * - getAllClaimRecords() - 获取所有领取记录
 * - getAllDonateRecords() - 获取所有投喂记录
 * - getAllUsers() - 获取用户列表
 * - rebindUser() - 重新绑定用户
 * - getSystemStats() - 获取系统统计
 * ... 更多管理员 API
 */
