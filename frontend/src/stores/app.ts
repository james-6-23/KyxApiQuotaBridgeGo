/**
 * App Store
 * 应用全局状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
  // ==================== State ====================

  const loading = ref(false)
  const sidebarCollapsed = ref(false)
  const theme = ref<'light' | 'dark'>('light')
  const locale = ref<'zh-CN' | 'en-US'>('zh-CN')
  const pageTitle = ref('')
  const breadcrumbs = ref<Array<{ name: string; path?: string }>>([])

  // 全局加载队列（用于多个请求同时进行时的加载状态管理）
  const loadingQueue = ref<Set<string>>(new Set())

  // 移动端检测
  const isMobile = ref(false)

  // 窗口尺寸
  const windowWidth = ref(window.innerWidth)
  const windowHeight = ref(window.innerHeight)

  // ==================== Getters ====================

  const isLoading = computed(() => loading.value || loadingQueue.value.size > 0)

  const isDarkMode = computed(() => theme.value === 'dark')

  const isLightMode = computed(() => theme.value === 'light')

  const currentLocale = computed(() => locale.value)

  const isSidebarCollapsed = computed(() => sidebarCollapsed.value)

  // 响应式布局
  const isSmallScreen = computed(() => windowWidth.value < 768)
  const isMediumScreen = computed(() => windowWidth.value >= 768 && windowWidth.value < 1024)
  const isLargeScreen = computed(() => windowWidth.value >= 1024)

  // ==================== Actions ====================

  /**
   * 设置加载状态
   */
  const setLoading = (value: boolean) => {
    loading.value = value
  }

  /**
   * 添加加载任务
   */
  const addLoadingTask = (taskId: string) => {
    loadingQueue.value.add(taskId)
  }

  /**
   * 移除加载任务
   */
  const removeLoadingTask = (taskId: string) => {
    loadingQueue.value.delete(taskId)
  }

  /**
   * 清空所有加载任务
   */
  const clearLoadingTasks = () => {
    loadingQueue.value.clear()
  }

  /**
   * 切换侧边栏状态
   */
  const toggleSidebar = () => {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed.value))
  }

  /**
   * 设置侧边栏状态
   */
  const setSidebarCollapsed = (value: boolean) => {
    sidebarCollapsed.value = value
    localStorage.setItem('sidebarCollapsed', String(value))
  }

  /**
   * 切换主题
   */
  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    applyTheme(theme.value)
    localStorage.setItem('theme', theme.value)
  }

  /**
   * 设置主题
   */
  const setTheme = (value: 'light' | 'dark') => {
    theme.value = value
    applyTheme(value)
    localStorage.setItem('theme', value)
  }

  /**
   * 应用主题到 DOM
   */
  const applyTheme = (themeValue: 'light' | 'dark') => {
    const root = document.documentElement
    if (themeValue === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }

  /**
   * 切换语言
   */
  const toggleLocale = () => {
    locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
    localStorage.setItem('locale', locale.value)
  }

  /**
   * 设置语言
   */
  const setLocale = (value: 'zh-CN' | 'en-US') => {
    locale.value = value
    localStorage.setItem('locale', value)
  }

  /**
   * 设置页面标题
   */
  const setPageTitle = (title: string) => {
    pageTitle.value = title
    document.title = title ? `${title} - KYX API Quota Bridge` : 'KYX API Quota Bridge'
  }

  /**
   * 设置面包屑导航
   */
  const setBreadcrumbs = (items: Array<{ name: string; path?: string }>) => {
    breadcrumbs.value = items
  }

  /**
   * 更新移动端状态
   */
  const updateMobileStatus = () => {
    isMobile.value = window.innerWidth < 768

    // 在移动端自动收起侧边栏
    if (isMobile.value && !sidebarCollapsed.value) {
      sidebarCollapsed.value = true
    }
  }

  /**
   * 更新窗口尺寸
   */
  const updateWindowSize = () => {
    windowWidth.value = window.innerWidth
    windowHeight.value = window.innerHeight
    updateMobileStatus()
  }

  /**
   * 初始化应用设置
   */
  const initApp = () => {
    // 恢复主题设置
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null
    if (savedTheme) {
      theme.value = savedTheme
      applyTheme(savedTheme)
    }

    // 恢复语言设置
    const savedLocale = localStorage.getItem('locale') as 'zh-CN' | 'en-US' | null
    if (savedLocale) {
      locale.value = savedLocale
    }

    // 恢复侧边栏状态
    const savedSidebarCollapsed = localStorage.getItem('sidebarCollapsed')
    if (savedSidebarCollapsed !== null) {
      sidebarCollapsed.value = savedSidebarCollapsed === 'true'
    }

    // 检测移动端
    updateMobileStatus()

    // 监听窗口大小变化
    window.addEventListener('resize', updateWindowSize)
  }

  /**
   * 重置应用设置
   */
  const resetAppSettings = () => {
    theme.value = 'light'
    locale.value = 'zh-CN'
    sidebarCollapsed.value = false
    pageTitle.value = ''
    breadcrumbs.value = []

    applyTheme('light')

    localStorage.removeItem('theme')
    localStorage.removeItem('locale')
    localStorage.removeItem('sidebarCollapsed')

    setPageTitle('')
  }

  /**
   * 显示全局加载
   */
  const showLoading = (taskId?: string) => {
    if (taskId) {
      addLoadingTask(taskId)
    } else {
      loading.value = true
    }
  }

  /**
   * 隐藏全局加载
   */
  const hideLoading = (taskId?: string) => {
    if (taskId) {
      removeLoadingTask(taskId)
    } else {
      loading.value = false
    }
  }

  /**
   * 执行带加载状态的异步任务
   */
  const withLoading = async <T>(
    task: () => Promise<T>,
    taskId?: string
  ): Promise<T> => {
    const id = taskId || `task_${Date.now()}_${Math.random()}`

    try {
      addLoadingTask(id)
      return await task()
    } finally {
      removeLoadingTask(id)
    }
  }

  // ==================== Return ====================

  return {
    // State
    loading,
    sidebarCollapsed,
    theme,
    locale,
    pageTitle,
    breadcrumbs,
    loadingQueue,
    isMobile,
    windowWidth,
    windowHeight,

    // Getters
    isLoading,
    isDarkMode,
    isLightMode,
    currentLocale,
    isSidebarCollapsed,
    isSmallScreen,
    isMediumScreen,
    isLargeScreen,

    // Actions
    setLoading,
    addLoadingTask,
    removeLoadingTask,
    clearLoadingTasks,
    toggleSidebar,
    setSidebarCollapsed,
    toggleTheme,
    setTheme,
    applyTheme,
    toggleLocale,
    setLocale,
    setPageTitle,
    setBreadcrumbs,
    updateMobileStatus,
    updateWindowSize,
    initApp,
    resetAppSettings,
    showLoading,
    hideLoading,
    withLoading
  }
}, {
  // 持久化配置
  persist: {
    key: 'app-store',
    storage: localStorage,
    paths: ['theme', 'locale', 'sidebarCollapsed']
  }
})
