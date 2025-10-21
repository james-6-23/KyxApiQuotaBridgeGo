<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full">
      <!-- Logo 和标题 -->
      <div class="text-center mb-8">
        <div class="flex justify-center mb-4">
          <div class="w-16 h-16 bg-blue-600 rounded-full flex items-center justify-center shadow-lg">
            <ApiOutlined class="text-3xl text-white" />
          </div>
        </div>
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          KYX API Quota Bridge
        </h1>
        <p class="text-gray-600">
          API 配额管理系统
        </p>
      </div>

      <!-- 登录卡片 -->
      <a-card class="shadow-xl rounded-lg" :loading="loading">
        <div class="space-y-6">
          <!-- 欢迎信息 -->
          <div class="text-center">
            <h2 class="text-2xl font-semibold text-gray-800 mb-2">
              欢迎登录
            </h2>
            <p class="text-sm text-gray-600">
              使用 Linux.do 账号快速登录
            </p>
          </div>

          <!-- 登录按钮 -->
          <div class="space-y-4">
            <a-button
              type="primary"
              size="large"
              block
              :loading="oauthLoading"
              @click="handleOAuthLogin"
              class="h-12 text-base font-medium"
            >
              <template #icon>
                <LoginOutlined />
              </template>
              使用 Linux.do 账号登录
            </a-button>

            <!-- 错误提示 -->
            <a-alert
              v-if="errorMessage"
              :message="errorMessage"
              type="error"
              closable
              show-icon
              @close="errorMessage = ''"
            />
          </div>

          <!-- 功能说明 -->
          <div class="border-t pt-6">
            <h3 class="text-sm font-medium text-gray-700 mb-3">
              系统功能
            </h3>
            <div class="space-y-2">
              <div class="flex items-start space-x-2 text-sm text-gray-600">
                <CheckCircleOutlined class="text-green-500 mt-0.5" />
                <span>绑定 KYX 账号，统一管理 API 配额</span>
              </div>
              <div class="flex items-start space-x-2 text-sm text-gray-600">
                <CheckCircleOutlined class="text-green-500 mt-0.5" />
                <span>每日领取免费 API 配额</span>
              </div>
              <div class="flex items-start space-x-2 text-sm text-gray-600">
                <CheckCircleOutlined class="text-green-500 mt-0.5" />
                <span>投喂 API Keys 获取额外配额</span>
              </div>
            </div>
          </div>

          <!-- 提示信息 -->
          <div class="bg-blue-50 rounded-lg p-4">
            <div class="flex">
              <InfoCircleOutlined class="text-blue-600 text-lg mt-0.5 mr-3" />
              <div class="text-sm text-blue-800">
                <p class="font-medium mb-1">首次登录提示</p>
                <p>点击登录后将跳转到 Linux.do 进行授权，授权成功后自动返回系统。</p>
              </div>
            </div>
          </div>
        </div>
      </a-card>

      <!-- 底部链接 -->
      <div class="mt-8 text-center">
        <div class="space-x-4 text-sm">
          <a
            href="#"
            class="text-gray-600 hover:text-blue-600 transition-colors"
            @click.prevent="showAboutDialog"
          >
            关于我们
          </a>
          <span class="text-gray-400">|</span>
          <router-link
            to="/admin/login"
            class="text-gray-600 hover:text-blue-600 transition-colors"
          >
            管理员登录
          </router-link>
        </div>
        <p class="text-gray-500 text-xs mt-4">
          © {{ currentYear }} KYX API Quota Bridge. All rights reserved.
        </p>
      </div>
    </div>

    <!-- 关于对话框 -->
    <a-modal
      v-model:open="aboutVisible"
      title="关于 KYX API Quota Bridge"
      :footer="null"
      width="500px"
    >
      <div class="space-y-4">
        <p class="text-gray-700">
          KYX API Quota Bridge 是一个 API 配额管理系统，帮助用户统一管理和分配 API 使用额度。
        </p>
        <div class="border-t pt-4">
          <h4 class="font-medium text-gray-900 mb-2">主要功能</h4>
          <ul class="list-disc list-inside space-y-1 text-sm text-gray-600">
            <li>通过 Linux.do OAuth 安全登录</li>
            <li>绑定 KYX 账号进行额度管理</li>
            <li>每日领取免费 API 配额</li>
            <li>通过投喂 Keys 获取额外配额</li>
            <li>查看领取和投喂历史记录</li>
          </ul>
        </div>
        <div class="border-t pt-4">
          <h4 class="font-medium text-gray-900 mb-2">技术栈</h4>
          <div class="flex flex-wrap gap-2">
            <a-tag color="blue">Vue 3</a-tag>
            <a-tag color="green">TypeScript</a-tag>
            <a-tag color="purple">Pinia</a-tag>
            <a-tag color="cyan">Ant Design Vue</a-tag>
            <a-tag color="orange">Tailwind CSS</a-tag>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  ApiOutlined,
  LoginOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { getOAuthUrl } from '@/api/auth'

// ==================== Composables ====================
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// ==================== State ====================
const loading = ref(false)
const oauthLoading = ref(false)
const errorMessage = ref('')
const aboutVisible = ref(false)

// ==================== Computed ====================
const currentYear = computed(() => new Date().getFullYear())

// ==================== Methods ====================

/**
 * 处理 OAuth 登录
 */
const handleOAuthLogin = async () => {
  try {
    oauthLoading.value = true
    errorMessage.value = ''

    // 获取 OAuth 授权 URL
    const { data } = await getOAuthUrl()

    if (data.success && data.data && data.data.url) {
      // 跳转到 OAuth 授权页面
      window.location.href = data.data.url
    } else {
      errorMessage.value = data.message || '获取登录链接失败，请稍后重试'
      message.error(errorMessage.value)
    }
  } catch (error: any) {
    console.error('OAuth login failed:', error)
    errorMessage.value = error.message || '登录失败，请检查网络连接后重试'
    message.error(errorMessage.value)
  } finally {
    oauthLoading.value = false
  }
}

/**
 * 显示关于对话框
 */
const showAboutDialog = () => {
  aboutVisible.value = true
}

/**
 * 检查登录状态
 */
const checkLoginStatus = async () => {
  // 如果已经登录，重定向到仪表板
  if (authStore.isAuthenticated) {
    const redirect = route.query.redirect as string
    if (redirect) {
      router.replace(redirect)
    } else {
      router.replace('/user/dashboard')
    }
    return
  }

  // 尝试从 localStorage 恢复登录状态
  const token = localStorage.getItem('token')
  if (token) {
    loading.value = true
    const restored = await authStore.checkAuthStatus()
    loading.value = false

    if (restored) {
      const redirect = route.query.redirect as string
      if (redirect) {
        router.replace(redirect)
      } else {
        router.replace('/user/dashboard')
      }
    }
  }
}

// ==================== Lifecycle ====================

onMounted(() => {
  // 检查登录状态
  checkLoginStatus()

  // 如果 URL 中有错误信息，显示出来
  if (route.query.error) {
    errorMessage.value = route.query.error as string
    message.error(errorMessage.value)
  }
})
</script>

<style scoped>
/* 渐变背景动画 */
.bg-gradient-to-br {
  animation: gradientShift 15s ease infinite;
  background-size: 200% 200%;
}

@keyframes gradientShift {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

/* 卡片入场动画 */
.ant-card {
  animation: fadeInUp 0.6s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 按钮悬停效果增强 */
:deep(.ant-btn-primary) {
  transition: all 0.3s ease;
}

:deep(.ant-btn-primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.4);
}

/* Logo 脉动动画 */
.w-16.h-16 {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(37, 99, 235, 0.7);
  }
  50% {
    box-shadow: 0 0 0 10px rgba(37, 99, 235, 0);
  }
}

/* 响应式调整 */
@media (max-width: 640px) {
  .max-w-md {
    padding: 0 1rem;
  }
}
</style>
