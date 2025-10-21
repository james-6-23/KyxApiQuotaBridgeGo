<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full">
      <!-- Logo 和标题 -->
      <div class="text-center mb-8">
        <div class="flex justify-center mb-4">
          <div class="w-16 h-16 bg-red-600 rounded-full flex items-center justify-center shadow-lg">
            <CrownOutlined class="text-3xl text-white" />
          </div>
        </div>
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          管理员登录
        </h1>
        <p class="text-gray-600">
          KYX API Quota Bridge 管理后台
        </p>
      </div>

      <!-- 登录卡片 -->
      <a-card class="shadow-xl rounded-lg">
        <div class="space-y-6">
          <!-- 欢迎信息 -->
          <div class="text-center">
            <h2 class="text-2xl font-semibold text-gray-800 mb-2">
              管理员入口
            </h2>
            <p class="text-sm text-gray-600">
              请输入管理员密码以继续
            </p>
          </div>

          <!-- 登录表单 -->
          <a-form
            ref="formRef"
            :model="formState"
            :rules="rules"
            layout="vertical"
            @finish="handleSubmit"
          >
            <a-form-item
              name="password"
              :validate-status="validateStatus"
              :help="helpMessage"
            >
              <a-input-password
                v-model:value="formState.password"
                size="large"
                placeholder="请输入管理员密码"
                :disabled="loading"
                @pressEnter="handleSubmit"
              >
                <template #prefix>
                  <LockOutlined class="text-gray-400" />
                </template>
              </a-input-password>
            </a-form-item>

            <!-- 记住密码 -->
            <a-form-item class="mb-4">
              <a-checkbox v-model:checked="rememberPassword">
                记住密码（7天）
              </a-checkbox>
            </a-form-item>

            <!-- 错误提示 -->
            <a-alert
              v-if="errorMessage"
              :message="errorMessage"
              type="error"
              closable
              show-icon
              class="mb-4"
              @close="errorMessage = ''"
            />

            <!-- 登录按钮 -->
            <a-form-item class="mb-0">
              <a-button
                type="primary"
                html-type="submit"
                size="large"
                block
                :loading="loading"
                class="h-12 text-base font-medium"
              >
                <template #icon>
                  <LoginOutlined />
                </template>
                登录
              </a-button>
            </a-form-item>
          </a-form>

          <!-- 安全提示 -->
          <div class="bg-red-50 rounded-lg p-4">
            <div class="flex">
              <WarningOutlined class="text-red-600 text-lg mt-0.5 mr-3" />
              <div class="text-sm text-red-800">
                <p class="font-medium mb-1">安全提示</p>
                <ul class="list-disc list-inside space-y-1">
                  <li>请勿在公共设备上使用管理员账号</li>
                  <li>请定期更换管理员密码</li>
                  <li>请妥善保管您的登录凭证</li>
                </ul>
              </div>
            </div>
          </div>

          <!-- 功能说明 -->
          <div class="border-t pt-6">
            <h3 class="text-sm font-medium text-gray-700 mb-3">
              管理功能
            </h3>
            <div class="grid grid-cols-2 gap-3">
              <div class="flex items-center space-x-2 text-sm text-gray-600">
                <SettingOutlined class="text-blue-500" />
                <span>系统配置</span>
              </div>
              <div class="flex items-center space-x-2 text-sm text-gray-600">
                <KeyOutlined class="text-green-500" />
                <span>Keys 管理</span>
              </div>
              <div class="flex items-center space-x-2 text-sm text-gray-600">
                <UserOutlined class="text-purple-500" />
                <span>用户管理</span>
              </div>
              <div class="flex items-center space-x-2 text-sm text-gray-600">
                <BarChartOutlined class="text-orange-500" />
                <span>数据统计</span>
              </div>
            </div>
          </div>
        </div>
      </a-card>

      <!-- 底部链接 -->
      <div class="mt-8 text-center">
        <div class="space-x-4 text-sm">
          <router-link
            to="/user/login"
            class="text-gray-600 hover:text-blue-600 transition-colors"
          >
            返回用户登录
          </router-link>
          <span class="text-gray-400">|</span>
          <a
            href="#"
            class="text-gray-600 hover:text-blue-600 transition-colors"
            @click.prevent="showHelpDialog"
          >
            忘记密码？
          </a>
        </div>
        <p class="text-gray-500 text-xs mt-4">
          © {{ currentYear }} KYX API Quota Bridge. All rights reserved.
        </p>
      </div>
    </div>

    <!-- 帮助对话框 -->
    <a-modal
      v-model:open="helpVisible"
      title="密码找回帮助"
      :footer="null"
      width="500px"
    >
      <div class="space-y-4">
        <a-alert
          message="重要提示"
          description="管理员密码需要通过服务器配置文件修改，无法在线重置。"
          type="warning"
          show-icon
        />

        <div class="space-y-3 text-sm text-gray-600">
          <p class="font-medium text-gray-800">如需修改管理员密码，请按以下步骤操作：</p>
          <ol class="list-decimal list-inside space-y-2 pl-2">
            <li>登录服务器，找到配置文件</li>
            <li>修改 <code class="bg-gray-100 px-2 py-1 rounded text-xs">ADMIN_PASSWORD</code> 环境变量</li>
            <li>重启服务使配置生效</li>
          </ol>
        </div>

        <a-divider />

        <div class="text-sm text-gray-600">
          <p class="font-medium text-gray-800 mb-2">需要技术支持？</p>
          <p>请联系系统管理员或查看项目文档获取详细帮助。</p>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue'
import type { Rule } from 'ant-design-vue/es/form'
import {
  CrownOutlined,
  LockOutlined,
  LoginOutlined,
  WarningOutlined,
  SettingOutlined,
  KeyOutlined,
  UserOutlined,
  BarChartOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { LoginForm } from '@/types'

// ==================== Composables ====================
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// ==================== State ====================
const formRef = ref<FormInstance>()
const loading = ref(false)
const errorMessage = ref('')
const helpVisible = ref(false)
const rememberPassword = ref(false)
const validateStatus = ref<'' | 'success' | 'warning' | 'error' | 'validating'>('')
const helpMessage = ref('')

const formState = reactive<LoginForm>({
  password: ''
})

// ==================== Computed ====================
const currentYear = computed(() => new Date().getFullYear())

// ==================== Form Rules ====================
const rules: Record<string, Rule[]> = {
  password: [
    { required: true, message: '请输入管理员密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' }
  ]
}

// ==================== Methods ====================

/**
 * 表单提交
 */
const handleSubmit = async () => {
  try {
    // 验证表单
    await formRef.value?.validate()

    loading.value = true
    errorMessage.value = ''

    // 调用登录 API
    const success = await authStore.adminLogin(formState)

    if (success) {
      // 如果勾选记住密码，保存到 localStorage
      if (rememberPassword.value) {
        localStorage.setItem('admin_remember', 'true')
        localStorage.setItem('admin_password_hint', formState.password.substring(0, 2) + '***')
      }

      message.success('登录成功！')

      // 短暂延迟后跳转
      setTimeout(() => {
        const redirect = route.query.redirect as string
        if (redirect) {
          router.replace(redirect)
        } else {
          router.replace('/admin/dashboard')
        }
      }, 500)
    }
  } catch (error: any) {
    if (error.errorFields) {
      // 表单验证错误
      console.error('Form validation failed:', error)
    } else {
      // API 错误
      console.error('Login failed:', error)
      errorMessage.value = error.message || '登录失败，请检查密码是否正确'
    }
  } finally {
    loading.value = false
  }
}

/**
 * 显示帮助对话框
 */
const showHelpDialog = () => {
  helpVisible.value = true
}

/**
 * 检查登录状态
 */
const checkLoginStatus = async () => {
  // 如果已经登录且是管理员，重定向到仪表板
  if (authStore.isAuthenticated && authStore.isAdmin) {
    const redirect = route.query.redirect as string
    if (redirect) {
      router.replace(redirect)
    } else {
      router.replace('/admin/dashboard')
    }
    return
  }

  // 检查是否记住密码
  const remembered = localStorage.getItem('admin_remember')
  if (remembered === 'true') {
    rememberPassword.value = true
    const hint = localStorage.getItem('admin_password_hint')
    if (hint) {
      helpMessage.value = `密码提示: ${hint}`
      validateStatus.value = 'warning'
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

  // 自动聚焦到密码输入框
  setTimeout(() => {
    const passwordInput = document.querySelector('input[type="password"]') as HTMLInputElement
    if (passwordInput) {
      passwordInput.focus()
    }
  }, 300)
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
  box-shadow: 0 4px 12px rgba(220, 38, 38, 0.4);
}

/* Logo 脉动动画 */
.w-16.h-16 {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(220, 38, 38, 0.7);
  }
  50% {
    box-shadow: 0 0 0 10px rgba(220, 38, 38, 0);
  }
}

/* 表单项样式优化 */
:deep(.ant-form-item-label > label) {
  font-weight: 500;
  color: #374151;
}

/* 输入框样式 */
:deep(.ant-input-affix-wrapper) {
  border-radius: 6px;
}

:deep(.ant-input-affix-wrapper:hover),
:deep(.ant-input-affix-wrapper:focus) {
  border-color: #dc2626;
  box-shadow: 0 0 0 2px rgba(220, 38, 38, 0.1);
}

/* 密码输入框样式 */
:deep(.ant-input-password) {
  border-radius: 6px;
}

/* 复选框样式 */
:deep(.ant-checkbox-wrapper) {
  color: #6b7280;
}

:deep(.ant-checkbox-checked .ant-checkbox-inner) {
  background-color: #dc2626;
  border-color: #dc2626;
}

/* 功能图标样式 */
.text-blue-500,
.text-green-500,
.text-purple-500,
.text-orange-500 {
  font-size: 16px;
}

/* 代码样式 */
code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

/* 警告框样式 */
:deep(.ant-alert) {
  border-radius: 6px;
}

/* 响应式调整 */
@media (max-width: 640px) {
  .max-w-md {
    padding: 0 1rem;
  }

  .grid-cols-2 {
    grid-template-columns: 1fr;
  }

  .text-3xl {
    font-size: 1.75rem;
  }
}

/* 帮助对话框样式 */
:deep(.ant-modal-header) {
  border-bottom: 1px solid #e5e7eb;
}

:deep(.ant-modal-body) {
  padding-top: 24px;
}
</style>
