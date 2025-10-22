/**
 * Axios HTTP Client Configuration
 * 统一的 HTTP 请求封装，包含请求/响应拦截器
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { message } from 'ant-design-vue'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import type { ApiResponse } from '@/types'

// NProgress 配置
NProgress.configure({
  showSpinner: false,
  trickleSpeed: 200,
  minimum: 0.3
})

// 消息去重控制
const messageCache = new Set<string>()
const MESSAGE_DURATION = 3000 // 消息显示时间
const MESSAGE_CACHE_DURATION = 5000 // 去重缓存时间

/**
 * 显示消息（带去重）
 */
function showMessage(type: 'success' | 'error' | 'warning' | 'info', content: string) {
  // 如果消息已在缓存中，不再显示
  if (messageCache.has(content)) {
    return
  }

  // 添加到缓存
  messageCache.add(content)

  // 显示消息
  message[type](content, MESSAGE_DURATION / 1000)

  // 清除缓存
  setTimeout(() => {
    messageCache.delete(content)
  }, MESSAGE_CACHE_DURATION)
}

// 配置 Ant Design Vue message 最大显示数量
message.config({
  maxCount: 3, // 最多同时显示3个消息
  duration: MESSAGE_DURATION / 1000
})

/**
 * 请求配置接口
 */
interface RequestConfig extends AxiosRequestConfig {
  skipAuth?: boolean // 跳过认证
  skipErrorHandler?: boolean // 跳过错误处理
  showLoading?: boolean // 显示加载进度
  showSuccessMsg?: boolean // 显示成功消息
  successMsg?: string // 自定义成功消息
}

/**
 * 创建 Axios 实例
 */
const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json;charset=UTF-8'
  },
  withCredentials: true // 允许携带 cookie
})

/**
 * 请求拦截器
 */
service.interceptors.request.use(
  (config: any) => {
    // 显示加载进度条
    if (config.showLoading !== false) {
      NProgress.start()
    }

    // 添加认证 Token
    if (!config.skipAuth) {
      const token = localStorage.getItem('token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }

    // 添加时间戳防止缓存
    if (config.method === 'get') {
      config.params = {
        ...config.params,
        _t: Date.now()
      }
    }

    return config
  },
  (error: AxiosError) => {
    NProgress.done()
    console.error('Request Error:', error)
    return Promise.reject(error)
  }
)

/**
 * 响应拦截器
 */
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    NProgress.done()

    const config = response.config as RequestConfig
    const res = response.data

    // 如果返回的是文件流（下载）
    if (response.config.responseType === 'blob') {
      return response
    }

    // API 返回成功
    if (res.success) {
      // 显示成功消息
      if (config.showSuccessMsg) {
        showMessage('success', config.successMsg || res.message || '操作成功')
      }
      return response
    }

    // API 返回失败
    if (!config.skipErrorHandler) {
      handleApiError(res)
    }

    return Promise.reject(new Error(res.message || '请求失败'))
  },
  (error: AxiosError<ApiResponse>) => {
    NProgress.done()

    const config = error.config as RequestConfig

    // 跳过错误处理
    if (config?.skipErrorHandler) {
      return Promise.reject(error)
    }

    // 处理不同的错误情况
    if (error.response) {
      handleHttpError(error.response)
    } else if (error.request) {
      // 请求已发出但没有收到响应
      showMessage('error', '网络连接失败，请检查您的网络')
    } else {
      // 请求配置出错
      showMessage('error', error.message || '请求配置错误')
    }

    return Promise.reject(error)
  }
)

/**
 * 处理 API 业务错误
 */
function handleApiError(res: ApiResponse) {
  const errorMsg = res.message || res.error || '操作失败'

  // 根据错误类型显示不同的消息
  if (errorMsg.includes('未登录') || errorMsg.includes('登录')) {
    showMessage('warning', '请先登录')
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    // 跳转到登录页
    setTimeout(() => {
      window.location.href = '/login'
    }, 1500)
  } else if (errorMsg.includes('权限') || errorMsg.includes('无权')) {
    showMessage('error', '您没有权限执行此操作')
  } else if (errorMsg.includes('已存在')) {
    showMessage('warning', errorMsg)
  } else {
    showMessage('error', errorMsg)
  }
}

/**
 * 处理 HTTP 状态码错误
 */
function handleHttpError(response: AxiosResponse<ApiResponse>) {
  const status = response.status
  const data = response.data

  switch (status) {
    case 400:
      showMessage('error', data?.message || '请求参数错误')
      break
    case 401:
      showMessage('warning', '登录已过期，请重新登录')
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      setTimeout(() => {
        window.location.href = '/login'
      }, 1500)
      break
    case 403:
      showMessage('error', '您没有权限访问此资源')
      break
    case 404:
      showMessage('error', '请求的资源不存在')
      break
    case 408:
      showMessage('error', '请求超时')
      break
    case 429:
      showMessage('warning', '请求过于频繁，请稍后再试')
      break
    case 500:
      showMessage('error', '服务器内部错误')
      break
    case 502:
      showMessage('error', '网关错误')
      break
    case 503:
      showMessage('error', '服务暂时不可用')
      break
    case 504:
      showMessage('error', '网关超时')
      break
    default:
      showMessage('error', data?.message || `请求失败 (${status})`)
  }
}

/**
 * 统一请求方法
 */
class Request {
  /**
   * GET 请求
   */
  get<T = any>(url: string, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.get(url, config)
  }

  /**
   * POST 请求
   */
  post<T = any>(url: string, data?: any, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.post(url, data, config)
  }

  /**
   * PUT 请求
   */
  put<T = any>(url: string, data?: any, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.put(url, data, config)
  }

  /**
   * DELETE 请求
   */
  delete<T = any>(url: string, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.delete(url, config)
  }

  /**
   * PATCH 请求
   */
  patch<T = any>(url: string, data?: any, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.patch(url, data, config)
  }

  /**
   * 文件上传
   */
  upload<T = any>(url: string, formData: FormData, config?: RequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return service.post(url, formData, {
      ...config,
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }

  /**
   * 文件下载
   */
  download(url: string, filename: string, config?: RequestConfig): Promise<void> {
    return service.get(url, {
      ...config,
      responseType: 'blob'
    }).then(response => {
      const blob = new Blob([response.data])
      const link = document.createElement('a')
      link.href = window.URL.createObjectURL(blob)
      link.download = filename
      link.click()
      window.URL.revokeObjectURL(link.href)
      showMessage('success', '下载成功')
    }).catch(error => {
      showMessage('error', '下载失败')
      throw error
    })
  }

  /**
   * 批量请求
   */
  all<T = any>(requests: Promise<any>[]): Promise<T[]> {
    return Promise.all(requests)
  }
}

// 导出请求实例
export const request = new Request()

// 导出 axios 实例（用于特殊情况）
export default service

// 导出类型
export type { RequestConfig }
