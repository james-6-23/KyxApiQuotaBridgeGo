/**
 * Admin API Module
 * 管理员相关的 API 接口
 */

import { request } from './request'
import type {
  AdminConfig,
  ConfigUpdateForm,
  UserStats,
  ClaimRecord,
  DonateRecord,
  DonatedKey,
  SystemStats,
  PaginationParams,
  PaginatedResponse
} from '@/types'

// ==================== 配置管理 ====================

/**
 * 获取系统配置
 * @returns 系统配置信息
 */
export const getAdminConfig = () => {
  return request.get<AdminConfig>('/admin/config')
}

/**
 * 更新每日领取额度
 * @param claim_quota - 新的领取额度
 * @returns 更新结果
 */
export const updateQuota = (claim_quota: number) => {
  return request.put('/admin/config/quota', { claim_quota }, {
    showSuccessMsg: true,
    successMsg: '领取额度已更新'
  })
}

/**
 * 更新 API Session
 * @param session - 新的 Session 值
 * @returns 更新结果
 */
export const updateSession = (session: string) => {
  return request.put('/admin/config/session', { session }, {
    showSuccessMsg: true,
    successMsg: 'Session 已更新'
  })
}

/**
 * 更新 new-api-user 配置
 * @param new_api_user - 新的 new-api-user 值
 * @returns 更新结果
 */
export const updateNewApiUser = (new_api_user: string) => {
  return request.put('/admin/config/new-api-user', { new_api_user }, {
    showSuccessMsg: true,
    successMsg: 'new-api-user 已更新'
  })
}

/**
 * 更新 Keys API URL
 * @param keys_api_url - Keys 推送 API 地址
 * @returns 更新结果
 */
export const updateKeysApiUrl = (keys_api_url: string) => {
  return request.put('/admin/config/keys-api-url', { keys_api_url }, {
    showSuccessMsg: true,
    successMsg: 'Keys API URL 已更新'
  })
}

/**
 * 更新 Keys Authorization Token
 * @param keys_authorization - Keys 推送授权令牌
 * @returns 更新结果
 */
export const updateKeysAuthorization = (keys_authorization: string) => {
  return request.put('/admin/config/keys-authorization', { keys_authorization }, {
    showSuccessMsg: true,
    successMsg: 'Keys Authorization 已更新'
  })
}

/**
 * 更新 Group ID
 * @param group_id - Keys 推送的目标分组 ID
 * @returns 更新结果
 */
export const updateGroupId = (group_id: number) => {
  return request.put('/admin/config/group-id', { group_id }, {
    showSuccessMsg: true,
    successMsg: 'Group ID 已更新'
  })
}

/**
 * 批量更新配置
 * @param config - 配置更新对象
 * @returns 更新结果
 */
export const updateConfig = (config: Partial<ConfigUpdateForm>) => {
  return request.put('/admin/config', config, {
    showSuccessMsg: true,
    successMsg: '配置已更新'
  })
}

// ==================== Keys 管理 ====================

/**
 * 导出所有 Keys
 * @returns 所有已投喂的 Keys 列表
 */
export const exportKeys = () => {
  return request.get<DonatedKey[]>('/admin/keys/export')
}

/**
 * 获取所有 Keys（分页）
 * @param params - 分页参数
 * @returns 所有已投喂的 Keys 列表（分页）
 */
export const getAllKeys = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<DonatedKey>>('/admin/keys', { params })
}

/**
 * 删除指定的 Keys
 * @param keys - 要删除的 Keys 数组
 * @returns 删除结果
 */
export const deleteKeys = (keys: string[]) => {
  return request.post('/admin/keys/delete', { keys }, {
    showSuccessMsg: true,
    successMsg: `成功删除 ${keys.length} 个 Key`
  })
}

/**
 * 测试 Keys 有效性
 * @param keys - 要测试的 Keys 数组
 * @returns 测试结果
 */
export const testKeys = (keys: string[]) => {
  return request.post<{
    total: number
    valid: number
    invalid: number
    validations: Array<{ key: string; valid: boolean; message?: string }>
  }>('/admin/keys/test', { keys })
}

/**
 * 推送 Keys 到指定分组
 * @param keys - 要推送的 Keys 数组
 * @param group_id - 目标分组 ID（可选）
 * @returns 推送结果
 */
export const pushKeys = (keys: string[], group_id?: number) => {
  return request.post('/admin/keys/push', { keys, group_id }, {
    showSuccessMsg: true,
    successMsg: 'Keys 推送成功'
  })
}

/**
 * 获取 Keys 统计信息
 * @returns Keys 统计数据
 */
export const getKeysStats = () => {
  return request.get<{
    total_keys: number
    valid_keys: number
    invalid_keys: number
    today_donated: number
  }>('/admin/keys/stats')
}

// ==================== 记录管理 ====================

/**
 * 获取所有领取记录
 * @param params - 分页参数
 * @returns 所有用户的领取记录列表
 */
export const getAllClaimRecords = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<ClaimRecord>>('/admin/records/claim', { params })
}

/**
 * 获取所有投喂记录
 * @param params - 分页参数
 * @returns 所有用户的投喂记录列表
 */
export const getAllDonateRecords = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<DonateRecord>>('/admin/records/donate', { params })
}

/**
 * 删除领取记录
 * @param id - 领取记录 ID
 * @returns 删除结果
 */
export const deleteClaimRecord = (id: number) => {
  return request.delete(`/admin/records/claim/${id}`, {
    showSuccessMsg: true,
    successMsg: '领取记录已删除'
  })
}

/**
 * 删除投喂记录
 * @param id - 投喂记录 ID
 * @returns 删除结果
 */
export const deleteDonateRecord = (id: number) => {
  return request.delete(`/admin/records/donate/${id}`, {
    showSuccessMsg: true,
    successMsg: '投喂记录已删除'
  })
}

/**
 * 重新推送失败的 Keys
 * @param linux_do_id - 用户的 Linux Do ID
 * @param timestamp - 投喂记录的时间戳
 * @returns 重新推送结果
 */
export const retryPushKeys = (linux_do_id: string, timestamp: number) => {
  return request.post('/admin/retry-push', { linux_do_id, timestamp }, {
    showSuccessMsg: true,
    successMsg: 'Keys 重新推送成功'
  })
}

/**
 * 获取指定时间范围的领取记录
 * @param start_date - 开始日期
 * @param end_date - 结束日期
 * @returns 时间范围内的领取记录
 */
export const getClaimRecordsByDateRange = (start_date: string, end_date: string) => {
  return request.get<ClaimRecord[]>('/admin/records/claim/range', {
    params: { start_date, end_date }
  })
}

/**
 * 获取指定时间范围的投喂记录
 * @param start_date - 开始日期
 * @param end_date - 结束日期
 * @returns 时间范围内的投喂记录
 */
export const getDonateRecordsByDateRange = (start_date: string, end_date: string) => {
  return request.get<DonateRecord[]>('/admin/records/donate/range', {
    params: { start_date, end_date }
  })
}

// ==================== 用户管理 ====================

/**
 * 获取所有用户列表
 * @param params - 分页参数
 * @returns 所有用户的统计信息列表
 */
export const getAllUsers = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<UserStats>>('/admin/users', { params })
}

/**
 * 重新绑定用户账号
 * @param linux_do_id - 用户的 Linux Do ID
 * @param new_username - 新的 KYX 用户名
 * @returns 重新绑定结果
 */
export const rebindUser = (linux_do_id: string, new_username: string) => {
  return request.post('/admin/rebind-user', { linux_do_id, new_username }, {
    showSuccessMsg: true,
    successMsg: '用户账号重新绑定成功'
  })
}

/**
 * 导出用户数据（JSON 格式）
 * @returns 用户数据文件
 */
export const exportUsers = () => {
  const filename = `users_export_${new Date().toISOString().split('T')[0]}.json`
  return request.download('/admin/export/users', filename)
}

/**
 * 导出用户数据（CSV 格式）
 * @returns 用户数据 CSV 文件
 */
export const exportUsersCSV = () => {
  const filename = `users_export_${new Date().toISOString().split('T')[0]}.csv`
  return request.download('/admin/export/users/csv', filename)
}

/**
 * 获取指定用户的详细信息
 * @param linux_do_id - 用户的 Linux Do ID
 * @returns 用户详细信息
 */
export const getUserDetail = (linux_do_id: string) => {
  return request.get<{
    user: UserStats
    claim_records: ClaimRecord[]
    donate_records: DonateRecord[]
  }>(`/admin/users/${linux_do_id}`)
}

/**
 * 删除用户（危险操作）
 * @param linux_do_id - 用户的 Linux Do ID
 * @returns 删除结果
 */
export const deleteUser = (linux_do_id: string) => {
  return request.delete(`/admin/users/${linux_do_id}`, {
    showSuccessMsg: true,
    successMsg: '用户已删除'
  })
}

// ==================== 统计数据 ====================

/**
 * 获取系统统计信息
 * @returns 系统整体统计数据
 */
export const getSystemStats = () => {
  return request.get<SystemStats>('/admin/stats')
}

/**
 * 获取今日统计数据
 * @returns 今日的统计信息
 */
export const getTodayStats = () => {
  return request.get<{
    today_claims: number
    today_donates: number
    today_new_users: number
    today_quota_distributed: number
  }>('/admin/stats/today')
}

/**
 * 获取最近 7 天的趋势数据
 * @returns 最近 7 天的统计趋势
 */
export const getWeeklyTrends = () => {
  return request.get<Array<{
    date: string
    claims: number
    donates: number
    new_users: number
  }>>('/admin/stats/weekly')
}

/**
 * 获取最近 30 天的趋势数据
 * @returns 最近 30 天的统计趋势
 */
export const getMonthlyTrends = () => {
  return request.get<Array<{
    date: string
    claims: number
    donates: number
    new_users: number
  }>>('/admin/stats/monthly')
}

/**
 * 获取排行榜数据
 * @returns 用户排行榜（投喂榜、领取榜）
 */
export const getLeaderboard = () => {
  return request.get<{
    top_donators: Array<{ username: string; total_donated: number }>
    top_claimers: Array<{ username: string; total_claimed: number }>
  }>('/admin/stats/leaderboard')
}

// ==================== 日志管理 ====================

/**
 * 获取操作日志
 * @param limit - 返回的日志数量限制
 * @returns 操作日志列表
 */
export const getOperationLogs = (limit = 100) => {
  return request.get<Array<{
    id: number
    admin_user: string
    action: string
    details: string
    timestamp: string
  }>>('/admin/logs', {
    params: { limit }
  })
}

/**
 * 获取错误日志
 * @param limit - 返回的日志数量限制
 * @returns 错误日志列表
 */
export const getErrorLogs = (limit = 100) => {
  return request.get<Array<{
    id: number
    error_type: string
    error_message: string
    stack_trace?: string
    timestamp: string
  }>>('/admin/logs/errors', {
    params: { limit }
  })
}

// ==================== 系统维护 ====================

/**
 * 清理过期的 Session
 * @returns 清理结果
 */
export const cleanupSessions = () => {
  return request.post('/admin/maintenance/cleanup-sessions', null, {
    showSuccessMsg: true,
    successMsg: 'Session 清理完成'
  })
}

/**
 * 清理缓存
 * @returns 清理结果
 */
export const clearCache = () => {
  return request.post('/admin/maintenance/clear-cache', null, {
    showSuccessMsg: true,
    successMsg: '缓存已清理'
  })
}

/**
 * 获取系统健康状态
 * @returns 系统健康检查结果
 */
export const getSystemHealth = () => {
  return request.get<{
    status: 'healthy' | 'degraded' | 'unhealthy'
    database: boolean
    redis: boolean
    api: boolean
    uptime: number
  }>('/admin/health')
}

/**
 * 备份数据库
 * @returns 备份结果
 */
export const backupDatabase = () => {
  return request.post('/admin/maintenance/backup', null, {
    showSuccessMsg: true,
    successMsg: '数据库备份已创建'
  })
}
