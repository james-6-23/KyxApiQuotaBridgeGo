/**
 * Authentication API Module
 * 认证相关的 API 接口
 */

import { request } from './request'
import type { OAuthUrl, User, LoginForm } from '@/types'

/**
 * 获取 OAuth 登录 URL
 * @returns OAuth 登录 URL
 */
export const getOAuthUrl = () => {
  return request.get<OAuthUrl>('/auth/url', {
    skipAuth: true
  })
}

/**
 * OAuth 回调处理
 * @param code - OAuth 授权码
 * @param state - OAuth 状态参数（可选）
 * @returns 用户信息
 */
export const handleOAuthCallback = (code: string, state?: string) => {
  return request.get<User>('/auth/callback', {
    params: { code, state },
    skipAuth: true
  })
}

/**
 * 检查当前登录状态
 * @returns 用户信息（如果已登录）
 */
export const checkAuth = () => {
  return request.get<User>('/auth/check', {
    skipErrorHandler: true
  })
}

/**
 * 用户登出
 * @returns 登出结果
 */
export const logout = () => {
  return request.post('/auth/logout', null, {
    showSuccessMsg: true,
    successMsg: '已退出登录'
  })
}

/**
 * 管理员登录
 * @param data - 登录表单数据
 * @returns 包含 Token 的响应
 */
export const adminLogin = (data: LoginForm) => {
  return request.post<{ token: string; user?: any }>('/admin/login', data, {
    skipAuth: true,
    showSuccessMsg: true,
    successMsg: '登录成功'
  })
}

/**
 * 刷新 Token（如果后端支持）
 * @returns 新的 Token
 */
export const refreshToken = () => {
  return request.post<{ token: string }>('/auth/refresh', null)
}

/**
 * 验证 Token 是否有效
 * @returns Token 验证结果
 */
export const validateToken = () => {
  return request.get<{ valid: boolean }>('/auth/validate')
}

/**
 * 获取当前用户信息
 * @returns 用户详细信息
 */
export const getCurrentUser = () => {
  return request.get<User>('/auth/user')
}
