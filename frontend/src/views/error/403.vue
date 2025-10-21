<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full text-center">
      <!-- 403 图标/数字 -->
      <div class="mb-8">
        <div class="text-9xl font-bold text-red-500 mb-4">403</div>
        <StopOutlined class="text-6xl text-gray-400" />
      </div>

      <!-- 错误信息 -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-4">禁止访问</h1>
        <p class="text-lg text-gray-600 mb-2">
          抱歉，您没有权限访问此页面
        </p>
        <p class="text-sm text-gray-500">
          该页面需要特定的权限才能访问，请联系管理员或使用有权限的账号登录
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

      <!-- 额外操作 -->
      <div class="mt-8">
        <a-button
          type="link"
          size="large"
          @click="handleLogout"
          class="text-gray-600"
        >
          <template #icon>
            <LogoutOutlined />
          </template>
          切换账号登录
        </a-button>
      </div>

      <!-- 额外信息 -->
      <div class="mt-12 text-sm text-gray-500">
        <p>如需申请权限，请联系系统管理员</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  StopOutlined,
  ArrowLeftOutlined,
  HomeOutlined,
  LogoutOutlined
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

/**
 * 退出登录并切换账号
 */
const handleLogout = async () => {
  try {
    await authStore.logout(false)
    message.info('请使用有权限的账号登录')

    // 根据当前路径判断跳转到哪个登录页
    const currentPath = router.currentRoute.value.path
    if (currentPath.startsWith('/admin')) {
      router.push('/admin/login')
    } else {
      router.push('/user/login')
    }
  } catch (error) {
    message.error('退出登录失败')
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

/* 403 数字动画 */
.text-9xl {
  animation: shake 0.8s ease-in-out;
}

@keyframes shake {
  0%, 100% {
    transform: translateX(0);
  }
  10%, 30%, 50%, 70%, 90% {
    transform: translateX(-5px);
  }
  20%, 40%, 60%, 80% {
    transform: translateX(5px);
  }
}

/* 图标脉动效果 */
.anticon {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}
</style>
