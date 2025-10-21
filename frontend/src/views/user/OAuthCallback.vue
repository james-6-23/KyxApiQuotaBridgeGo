<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="max-w-md w-full px-4">
      <a-card class="shadow-xl rounded-lg">
        <!-- 加载状态 -->
        <div v-if="processing" class="text-center py-8">
          <a-spin size="large" class="mb-6" />
          <h2 class="text-xl font-semibold text-gray-800 mb-2">
            正在处理登录...
          </h2>
          <p class="text-gray-600 text-sm">
            {{ statusMessage }}
          </p>
          <div class="mt-6">
            <a-progress :percent="progress" :show-info="false" />
          </div>
        </div>

        <!-- 成功状态 -->
        <div v-else-if="success" class="text-center py-8">
          <div class="mb-6">
            <CheckCircleOutlined class="text-6xl text-green-500" />
          </div>
          <h2 class="text-xl font-semibold text-gray-800 mb-2">
            登录成功！
          </h2>
          <p class="text-gray-600 text-sm mb-6">
            即将跳转到仪表板...
          </p>
          <a-progress :percent="100" status="success" :show-info="false" />
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="text-center py-8">
          <div class="mb-6">
            <CloseCircleOutlined class="text-6xl text-red-500" />
          </div>
          <h2 class="text-xl font-semibold text-gray-800 mb-2">
            登录失败
          </h2>
          <a-alert
            :message="errorMessage"
            type="error"
            show-icon
            class="text-left mb-6"
          />
          <div class="space-y-3">
            <a-button
              type="primary"
              block
              size="large"
              @click="handleRetry"
            >
              <template #icon>
                <ReloadOutlined />
              </template>
              重新登录
            </a-button>
            <a-button
              block
              size="large"
              @click="handleGoHome"
            >
              返回首页
            </a-button>
          </div>
        </div>

        <!-- 无效访问 -->
        <div v-else class="text-center py-8">
          <div class="mb-6">
            <WarningOutlined class="text-6xl text-orange-500" />
          </div>
          <h2 class="text-xl font-semibold text-gray-800 mb-2">
            无效的访问
          </h2>
          <p class="text-gray-600 text-sm mb-6">
            缺少必要的授权参数，请重新登录
          </p>
          <a-button
            type="primary"
            block
            size="large"
            @click="handleGoLogin"
          >
            前往登录
          </a-button>
        </div>
      </a-card>

      <!-- 调试信息（仅开发环境） -->
      <div v-if="isDev && debugInfo" class="mt-4">
        <a-collapse ghost>
          <a-collapse-panel key="1" header="调试信息">
            <pre class="text-xs bg-gray-100 p-3 rounded overflow-auto">{{ debugInfo }}</pre>
          </a-collapse-panel>
        </a-collapse>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

// ==================== Composables ====================
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================
const processing = ref(false)
const success = ref(false)
const error = ref(false)
const errorMessage = ref('')
const statusMessage = ref('正在验证授权信息...')
const progress = ref(0)
const debugInfo = ref<any>(null)

// ==================== Computed ====================
const isDev = computed(() => import.meta.env.DEV)

// ==================== Methods ====================

/**
 * 处理 OAuth 回调
 */
const handleOAuthCallback = async () => {
  // 获取 URL 参数
  const code = route.query.code as string
  const state = route.query.state as string
  const errorParam = route.query.error as string
  const errorDescription = route.query.error_description as string

  // 调试信息
  if (isDev.value) {
    debugInfo.value = {
      code: code ? `${code.substring(0, 10)}...` : 'null',
      state: state || 'null',
      error: errorParam || 'null',
      errorDescription: errorDescription || 'null',
      timestamp: new Date().toISOString()
    }
  }

  // 检查是否有错误参数
  if (errorParam) {
    error.value = true
    errorMessage.value = errorDescription || `授权失败: ${errorParam}`
    message.error(errorMessage.value)
    return
  }

  // 检查必要参数
  if (!code) {
    error.value = true
    errorMessage.value = '缺少授权码，请重新登录'
    message.error(errorMessage.value)
    return
  }

  try {
    processing.value = true
    progress.value = 20

    // 更新状态消息
    statusMessage.value = '正在验证授权码...'
    await sleep(500)
    progress.value = 40

    // 调用 Store 处理 OAuth 登录
    const loginSuccess = await authStore.handleOAuthLogin(code, state)

    progress.value = 70

    if (loginSuccess) {
      // 登录成功
      statusMessage.value = '登录成功，正在加载用户信息...'
      progress.value = 90

      // 短暂延迟以显示成功状态
      await sleep(500)
      progress.value = 100
      success.value = true

      // 等待一下再跳转
      await sleep(1000)

      // 获取重定向路径
      const redirect = route.query.redirect as string
      const targetPath = redirect || '/user/dashboard'

      // 跳转到目标页面
      router.replace(targetPath)
    } else {
      // 登录失败
      error.value = true
      errorMessage.value = '登录验证失败，请重试'
    }
  } catch (err: any) {
    console.error('OAuth callback error:', err)
    error.value = true
    errorMessage.value = err.message || '登录过程中发生错误，请重试'
    message.error(errorMessage.value)
  } finally {
    processing.value = false
  }
}

/**
 * 重试登录
 */
const handleRetry = () => {
  router.replace('/user/login')
}

/**
 * 返回首页
 */
const handleGoHome = () => {
  router.replace('/')
}

/**
 * 前往登录
 */
const handleGoLogin = () => {
  router.replace('/user/login')
}

/**
 * 延迟函数
 */
const sleep = (ms: number) => {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * 检查是否已登录
 */
const checkExistingAuth = () => {
  if (authStore.isAuthenticated) {
    // 如果已经登录，直接跳转
    const redirect = route.query.redirect as string
    router.replace(redirect || '/user/dashboard')
    return true
  }
  return false
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('OAuth 登录')

  // 检查是否已登录
  if (checkExistingAuth()) {
    return
  }

  // 检查是否有必要的参数
  const code = route.query.code as string
  const errorParam = route.query.error as string

  if (code || errorParam) {
    // 有授权参数，处理 OAuth 回调
    await handleOAuthCallback()
  } else {
    // 没有参数，显示无效访问
    // 状态已经是默认值，不需要设置
  }
})
</script>

<style scoped>
/* 渐变背景 */
.bg-gradient-to-br {
  background: linear-gradient(to bottom right, #eff6ff, #e0e7ff);
}

/* 卡片动画 */
.ant-card {
  animation: fadeInScale 0.5s ease-out;
}

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* 图标动画 */
.anticon {
  animation: iconFadeIn 0.6s ease-out;
}

@keyframes iconFadeIn {
  from {
    opacity: 0;
    transform: scale(0.5);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* 成功图标脉动 */
.text-green-500 {
  animation: successPulse 1.5s ease-in-out infinite;
}

@keyframes successPulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

/* 错误图标抖动 */
.text-red-500 {
  animation: errorShake 0.5s ease-in-out;
}

@keyframes errorShake {
  0%, 100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-10px);
  }
  75% {
    transform: translateX(10px);
  }
}

/* 加载进度条 */
:deep(.ant-progress-line) {
  transition: all 0.3s ease;
}

/* 调试信息 */
pre {
  max-height: 300px;
}

/* 响应式 */
@media (max-width: 640px) {
  .max-w-md {
    width: 100%;
    padding: 0 1rem;
  }
}
</style>
