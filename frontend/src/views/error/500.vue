<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full text-center">
      <!-- 500 图标/数字 -->
      <div class="mb-8">
        <div class="text-9xl font-bold text-orange-500 mb-4">500</div>
        <WarningOutlined class="text-6xl text-gray-400" />
      </div>

      <!-- 错误信息 -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-4">服务器错误</h1>
        <p class="text-lg text-gray-600 mb-2">
          抱歉，服务器遇到了一些问题
        </p>
        <p class="text-sm text-gray-500">
          我们正在努力修复，请稍后再试或刷新页面重新加载
        </p>
      </div>

      <!-- 错误详情（如果有） -->
      <div v-if="errorMessage" class="mb-6">
        <a-alert
          type="error"
          :message="errorMessage"
          show-icon
          closable
          class="text-left"
        />
      </div>

      <!-- 操作按钮 -->
      <div class="flex flex-col sm:flex-row gap-4 justify-center">
        <a-button
          type="default"
          size="large"
          @click="handleRefresh"
          :loading="refreshing"
          class="flex items-center justify-center"
        >
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新页面
        </a-button>

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
      <div class="mt-12 space-y-4">
        <div class="text-sm text-gray-500">
          <p>如果问题持续存在，请联系技术支持</p>
        </div>

        <!-- 故障排查建议 -->
        <a-collapse
          ghost
          class="mt-4"
        >
          <a-collapse-panel key="1" header="故障排查建议">
            <ul class="text-left text-sm text-gray-600 space-y-2">
              <li>• 尝试刷新页面</li>
              <li>• 检查网络连接是否正常</li>
              <li>• 清除浏览器缓存后重试</li>
              <li>• 稍后再试（服务器可能正在维护）</li>
              <li>• 联系管理员获取帮助</li>
            </ul>
          </a-collapse-panel>
        </a-collapse>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  WarningOutlined,
  ArrowLeftOutlined,
  HomeOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 状态
const refreshing = ref(false)
const errorMessage = ref<string>('')

/**
 * 刷新页面
 */
const handleRefresh = async () => {
  refreshing.value = true

  try {
    // 等待一段时间后刷新
    await new Promise(resolve => setTimeout(resolve, 500))

    // 重新加载页面
    window.location.reload()
  } catch (error) {
    message.error('刷新失败')
    refreshing.value = false
  }
}

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

// 挂载时从路由参数获取错误信息
onMounted(() => {
  if (route.query.message) {
    errorMessage.value = route.query.message as string
  }
})
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

/* 500 数字动画 */
.text-9xl {
  animation: wobble 1s ease-in-out;
}

@keyframes wobble {
  0%, 100% {
    transform: rotate(0deg);
  }
  25% {
    transform: rotate(-5deg);
  }
  75% {
    transform: rotate(5deg);
  }
}

/* 图标脉动效果 */
.anticon {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.7;
    transform: scale(1.1);
  }
}

/* 折叠面板样式优化 */
:deep(.ant-collapse-ghost) {
  background-color: transparent;
}

:deep(.ant-collapse-ghost .ant-collapse-item) {
  border-bottom: none;
}

:deep(.ant-collapse-header) {
  color: #1890ff !important;
  font-weight: 500;
}
</style>
