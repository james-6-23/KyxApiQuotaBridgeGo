import { request } from '@/utils/request'

// 用户信息类型
export interface UserInfo {
  username: string
  display_name: string
  linux_do_id: string
  avatar_url: string
  name: string
  quota: number
  used_quota: number
  total: number
  can_claim: boolean
  claimed_today: boolean
}

// 绑定请求
export interface BindRequest {
  username: string
}

// 领取响应
export interface ClaimResponse {
  quota_added: number
  new_quota: number
}

// Key 验证请求
export interface KeyTestRequest {
  key: string
}

// Key 验证响应
export interface KeyTestResponse {
  key: string
  valid: boolean
}

// 投喂请求
export interface DonateRequest {
  keys: string[]
}

// 投喂响应
export interface DonateResponse {
  valid_keys: number
  already_exists: number
  duplicate_removed: number
  quota_added: number
  push_status: string
  results: KeyTestResponse[]
}

// 领取记录
export interface ClaimRecord {
  linux_do_id: string
  username: string
  quota_added: number
  timestamp: number
  date: string
}

// 投喂记录
export interface DonateRecord {
  linux_do_id: string
  username: string
  keys_count: number
  total_quota_added: number
  timestamp: number
  push_status: string
  push_message: string
  failed_keys: string[] | null
}

// 认证 API
export const authApi = {
  // 绑定账号
  bind: (data: BindRequest) => request.post('/api/auth/bind', data),

  // 退出登录
  logout: () => request.post('/api/auth/logout'),
}

// 用户 API
export const userApi = {
  // 获取用户额度信息
  getQuota: () => request.get<UserInfo>('/api/user/quota'),

  // 获取领取记录
  getClaimRecords: () => request.get<ClaimRecord[]>('/api/user/records/claim'),

  // 获取投喂记录
  getDonateRecords: () => request.get<DonateRecord[]>('/api/user/records/donate'),
}

// 领取 API
export const claimApi = {
  // 每日领取
  daily: () => request.post<ClaimResponse>('/api/claim/daily', {}),
}

// 测试 API
export const testApi = {
  // 测试单个 Key
  testKey: (data: KeyTestRequest) => request.post<KeyTestResponse>('/api/test/key', data),
}

// 投喂 API
export const donateApi = {
  // 投喂 Keys
  validate: (data: DonateRequest) => request.post<DonateResponse>('/api/donate/validate', data),
}

// 管理员配置类型
export interface AdminConfig {
  claim_quota: number
  session_configured: boolean
  keys_api_url: string
  keys_authorization_configured: boolean
  group_id: number
  updated_at: number
}

// 管理员 API
export const adminApi = {
  // 登录
  login: (password: string) => request.post('/api/admin/login', { password }),

  // 获取配置
  getConfig: () => request.get<AdminConfig>('/api/admin/config'),

  // 更新领取额度
  updateQuota: (claim_quota: number) => request.put('/api/admin/config/quota', { claim_quota }),

  // 更新 Session
  updateSession: (session: string) => request.put('/api/admin/config/session', { session }),

  // 更新 New API User
  updateNewApiUser: (new_api_user: string) => request.put('/api/admin/config/new-api-user', { new_api_user }),

  // 更新 Keys API URL
  updateKeysApiUrl: (keys_api_url: string) => request.put('/api/admin/config/keys-api-url', { keys_api_url }),

  // 更新 Keys Authorization
  updateKeysAuthorization: (keys_authorization: string) => request.put('/api/admin/config/keys-authorization', { keys_authorization }),

  // 更新 Group ID
  updateGroupId: (group_id: number) => request.put('/api/admin/config/group-id', { group_id }),

  // 导出所有 Keys
  exportKeys: () => request.get<string[]>('/api/admin/keys/export'),

  // 批量测试 Keys
  testKeys: (keys: string[]) => request.post<KeyTestResponse[]>('/api/admin/keys/test', { keys }),

  // 删除 Keys
  deleteKeys: (keys: string[]) => request.post('/api/admin/keys/delete', { keys }),

  // 获取所有领取记录
  getClaimRecords: () => request.get<ClaimRecord[]>('/api/admin/records/claim'),

  // 获取所有投喂记录
  getDonateRecords: () => request.get<DonateRecord[]>('/api/admin/records/donate'),

  // 获取所有用户
  getUsers: () => request.get('/api/admin/users'),

  // 重新绑定用户
  rebindUser: (linux_do_id: string, new_username: string) =>
    request.post('/api/admin/rebind-user', { linux_do_id, new_username }),

  // 重新推送失败的 Keys
  retryPush: (linux_do_id: string, timestamp: number) =>
    request.post('/api/admin/retry-push', { linux_do_id, timestamp }),
}
