/**
 * User API Module
 * 用户相关的 API 接口
 */

import { request } from './request'
import type {
  UserQuota,
  UserStats,
  BindAccountForm,
  ClaimRecord,
  DonateRecord,
  DonateForm,
  PaginationParams,
  PaginatedResponse
} from '@/types'

/**
 * 获取用户额度信息
 * @returns 用户额度详细信息
 */
export const getUserQuota = () => {
  return request.get<UserQuota>('/user/quota')
}

/**
 * 绑定 KYX 公益站账号
 * @param data - 绑定表单数据
 * @returns 绑定结果
 */
export const bindAccount = (data: BindAccountForm) => {
  return request.post('/user/bind', data, {
    showSuccessMsg: true,
    successMsg: '账号绑定成功'
  })
}

/**
 * 每日领取配额
 * @returns 领取结果，包含领取的额度
 */
export const claimDailyQuota = () => {
  return request.post<{ quota_added: number }>('/user/claim', null, {
    showSuccessMsg: true,
    successMsg: '配额领取成功'
  })
}

/**
 * 投喂 ModelScope Keys
 * @param data - 投喂表单数据（Keys 数组）
 * @returns 投喂结果，包含验证和处理信息
 */
export const donateKeys = (data: DonateForm) => {
  return request.post<{
    valid_count: number
    invalid_count: number
    total_quota_added: number
    message: string
  }>('/user/donate', data, {
    showSuccessMsg: true,
    successMsg: '投喂成功'
  })
}

/**
 * 获取用户的领取记录
 * @param params - 分页参数
 * @returns 用户的所有领取记录列表
 */
export const getUserClaimRecords = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<ClaimRecord>>('/user/claims', { params })
}

/**
 * 获取用户的投喂记录
 * @param params - 分页参数
 * @returns 用户的所有投喂记录列表
 */
export const getUserDonateRecords = (params?: PaginationParams) => {
  return request.get<PaginatedResponse<DonateRecord>>('/user/donates', { params })
}

/**
 * 验证单个 ModelScope Key 是否有效
 * @param key - ModelScope Key
 * @returns Key 验证结果
 */
export const validateKey = (key: string) => {
  return request.post<{ valid: boolean; message?: string }>('/user/validate-key', { key })
}

/**
 * 批量验证 ModelScope Keys
 * @param keys - ModelScope Keys 数组
 * @returns 批量验证结果
 */
export const validateKeys = (keys: string[]) => {
  return request.post<{
    results: Array<{ key: string; valid: boolean; message?: string }>
    valid_count: number
    invalid_count: number
  }>('/user/validate-keys', { keys })
}

/**
 * 获取用户统计信息
 * @returns 用户的统计数据（总领取、总投喂等）
 */
export const getUserStats = () => {
  return request.get<UserStats>('/user/statistics')
}

/**
 * 检查今天是否已经领取配额
 * @returns 今天是否可以领取
 */
export const canClaimToday = () => {
  return request.get<{ can_claim: boolean; last_claim_date?: string }>('/user/can-claim')
}
