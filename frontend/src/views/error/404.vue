<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full text-center">
      <!-- 404 图标/数字 -->
      <div class="mb-8">
        <div class="text-9xl font-bold text-blue-500 mb-4">404</div>
        <FileSearchOutlined class="text-6xl text-gray-400" />
      </div>

      <!-- 错误信息 -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-4">页面未找到</h1>
        <p class="text-lg text-gray-600 mb-2">
          抱歉，您访问的页面不存在或已被删除
        </p>
        <p class="text-sm text-gray-500">
          请检查 URL 是否正确，或返回首页继续浏览
        </p>
      </div>

      <!-- 操作按钮 -->
      <div class="flex flex-col sm:flex-row gap-4 justify-center">
        <a-button
          type="default"
          size="large"
          @click="handleGoBack"
          class="flex items-center justify-center"
        >
          <template #icon>
            <ArrowLeftOutlined />
          </template>
          返回上一页
        </a-button>

        <a-button
          type="primary"
          size="large"
          @click="handleGoHome"
          class="flex items-center justify-center"
        >
          <template #icon>
            <HomeOutlined />
          </template>
          返回首页
        </a-button>
      </div>

      <!-- 额外信息 -->
      <div class="mt-12 text-sm text-gray-500">
        <p>如果您认为这是一个错误，请联系管理员</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import {
  FileSearchOutlined,
  ArrowLeftOutlined,
  HomeOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

/**
 * 返回上一页
 */
const handleGoBack = () => {
  // 如果有历史记录，返回上一页
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    // 否则返回首页
    handleGoHome()
  }
}

/**
 * 返回首页
 */
const handleGoHome = () => {
  // 根据用户类型返回对应的首页
  if (authStore.isAuthenticated) {
    if (authStore.isAdmin) {
      router.push('/admin/dashboard')
    } else {
      router.push('/user/dashboard')
    }
  } else {
    router.push('/user/login')
  }
}
</script>

<style scoped>
/* 页面动画 */
.min-h-screen > div {
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

/* 404 数字动画 */
.text-9xl {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.8;
  }
}
</style>
