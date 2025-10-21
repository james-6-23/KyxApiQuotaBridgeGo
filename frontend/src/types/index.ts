/**
 * KYX API Quota Bridge - TypeScript Type Definitions
 * 所有应用中使用的类型定义
 */

// ==================== 用户相关类型 ====================

/**
 * 用户信息
 */
export interface User {
  linux_do_id: string
  username: string
  kyx_username?: string
  created_at: string
  updated_at?: string
}

/**
 * 用户额度信息
 */
export interface UserQuota {
  username: string
  linux_do_id: string
  kyx_username: string
  quota: number
  balance: number
  is_bound: boolean
  can_claim_today: boolean
  last_claim_date?: string
  total_donated: number
  total_claimed: number
}

/**
 * 用户统计信息
 */
export interface UserStats {
  username: string
  linux_do_id: string
  donate_count: number
  donate_quota: number
  claim_count: number
  claim_quota: number
  total_quota: number
  created_at: string
}

// ==================== 领取和投喂记录 ====================

/**
 * 领取记录
 */
export interface ClaimRecord {
  id?: number
  linux_do_id: string
  username: string
  quota_added: number
  timestamp: string
  created_at?: string
}

/**
 * 投喂记录
 */
export interface DonateRecord {
  id?: number
  linux_do_id: string
  username: string
  keys_count: number
  valid_keys_count?: number
  invalid_keys_count?: number
  total_quota_added: number
  timestamp: string
  push_status?: 'success' | 'failed' | 'pending'
  push_message?: string
  failed_keys?: string[]
  created_at?: string
}

/**
 * 投喂的 Key 信息
 */
export interface DonatedKey {
  key: string
  linux_do_id: string
  username: string
  timestamp: string
  is_valid?: boolean
  status?: 'pending' | 'validated' | 'pushed' | 'failed'
}

// ==================== 管理员配置 ====================

/**
 * 管理员配置
 */
export interface AdminConfig {
  claim_quota: number
  session_configured: boolean
  keys_api_url?: string
  keys_authorization_configured?: boolean
  group_id?: number
  new_api_user?: string
  updated_at?: string
}

/**
 * 系统统计信息
 */
export interface SystemStats {
  total_users: number
  total_claims: number
  total_donates: number
  total_keys: number
  total_quota_distributed: number
  today_claims: number
  today_donates: number
}

// ==================== API 响应类型 ====================

/**
 * 标准 API 响应
 */
export interface ApiResponse<T = any> {
  success: boolean
  message?: string
  data?: T
  error?: string
}

/**
 * 分页参数
 */
export interface PaginationParams {
  page: number
  page_size: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

/**
 * 分页响应
 */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

/**
 * OAuth 相关
 */
export interface OAuthUrl {
  url: string
}

export interface OAuthCallback {
  code: string
  state?: string
}

// ==================== 表单类型 ====================

/**
 * 登录表单
 */
export interface LoginForm {
  password: string
}

/**
 * 绑定账号表单
 */
export interface BindAccountForm {
  kyx_username: string
}

/**
 * 投喂表单
 */
export interface DonateForm {
  keys: string[]
}

/**
 * 配置更新表单
 */
export interface ConfigUpdateForm {
  claim_quota?: number
  session?: string
  new_api_user?: string
  keys_api_url?: string
  keys_authorization?: string
  group_id?: number
}

// ==================== Key 验证相关 ====================

/**
 * Key 验证结果
 */
export interface KeyValidation {
  key: string
  valid: boolean
  message?: string
}

/**
 * 批量验证结果
 */
export interface BatchValidationResult {
  total: number
  valid: number
  invalid: number
  validations: KeyValidation[]
}

// ==================== 路由相关 ====================

/**
 * 路由元信息
 */
export interface RouteMeta {
  title?: string
  requiresAuth?: boolean
  requiresAdmin?: boolean
  icon?: string
  hidden?: boolean
  keepAlive?: boolean
}

// ==================== 组件 Props 类型 ====================

/**
 * 表格列配置
 */
export interface TableColumn {
  title: string
  dataIndex: string
  key: string
  width?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: 'left' | 'right'
  ellipsis?: boolean
  sorter?: boolean
  customRender?: (params: { text: any; record: any; index: number }) => any
}

/**
 * 统计卡片数据
 */
export interface StatCard {
  title: string
  value: number | string
  icon: string
  color: string
  trend?: {
    value: number
    isUp: boolean
  }
}

// ==================== 状态管理类型 ====================

/**
 * 认证状态
 */
export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  token?: string
  isAdmin: boolean
}

/**
 * 应用状态
 */
export interface AppState {
  loading: boolean
  sidebarCollapsed: boolean
  theme: 'light' | 'dark'
  locale: 'zh-CN' | 'en-US'
}

/**
 * 管理员状态
 */
export interface AdminState {
  config: AdminConfig | null
  stats: SystemStats | null
  users: UserStats[]
  loading: boolean
}

// ==================== 错误类型 ====================

/**
 * API 错误
 */
export interface ApiError {
  code: string
  message: string
  details?: any
}

/**
 * 表单验证错误
 */
export interface ValidationError {
  field: string
  message: string
}

// ==================== 工具类型 ====================

/**
 * 时间范围
 */
export interface DateRange {
  start: string
  end: string
}

/**
 * 筛选条件
 */
export interface FilterParams {
  keyword?: string
  status?: string
  date_range?: DateRange
  [key: string]: any
}

/**
 * 排序参数
 */
export interface SortParams {
  field: string
  order: 'ascend' | 'descend'
}

/**
 * 导出选项
 */
export interface ExportOptions {
  format: 'json' | 'csv' | 'xlsx'
  fields?: string[]
  filename?: string
}

// ==================== Vite 环境变量 ====================

/**
 * 环境变量类型
 */
export interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
  readonly VITE_API_BASE_URL: string
  readonly VITE_UPLOAD_MAX_SIZE: number
}

export interface ImportMeta {
  readonly env: ImportMetaEnv
}
