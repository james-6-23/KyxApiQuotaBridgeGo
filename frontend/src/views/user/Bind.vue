<template>
  <div class="bind-account-page">
    <!-- 已绑定状态 -->
    <div v-if="isBound">
      <a-result
        status="success"
        title="账号已绑定"
        :sub-title="`您的 KYX 账号 ${kyxUsername} 已成功绑定`"
      >
        <template #icon>
          <CheckCircleOutlined class="text-green-500" />
        </template>
        <template #extra>
          <a-space direction="vertical" :size="16" class="w-full">
            <a-card class="text-left">
              <a-descriptions :column="1" bordered>
                <a-descriptions-item label="Linux.do 用户名">
                  {{ username }}
                </a-descriptions-item>
                <a-descriptions-item label="Linux.do ID">
                  {{ linuxDoId }}
                </a-descriptions-item>
                <a-descriptions-item label="KYX 用户名">
                  <a-tag color="success">{{ kyxUsername }}</a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="绑定时间">
                  {{ quota?.created_at ? formatDate(quota.created_at) : '未知' }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <a-space>
              <a-button type="primary" size="large" @click="goToDashboard">
                <template #icon>
                  <DashboardOutlined />
                </template>
                返回仪表板
              </a-button>
              <a-button size="large" @click="goToClaim">
                <template #icon>
                  <GiftOutlined />
                </template>
                领取额度
              </a-button>
            </a-space>

            <!-- 解绑提示 -->
            <a-alert
              message="解绑说明"
              description="如需更换绑定的 KYX 账号，请联系管理员进行解绑操作。"
              type="info"
              show-icon
            />
          </a-space>
        </template>
      </a-result>
    </div>

    <!-- 未绑定 - 显示绑定表单 -->
    <div v-else>
      <a-row justify="center">
        <a-col :xs="24" :sm="20" :md="16" :lg="12" :xl="10">
          <a-card class="bind-card">
            <!-- 卡片头部 -->
            <div class="text-center mb-6">
              <div class="mb-4">
                <LinkOutlined class="text-5xl text-blue-500" />
              </div>
              <h2 class="text-2xl font-bold text-gray-800 mb-2">
                绑定 KYX 账号
              </h2>
              <p class="text-gray-600">
                绑定您的 KYX 账号以使用配额管理功能
              </p>
            </div>

            <!-- 绑定表单 -->
            <a-form
              ref="formRef"
              :model="formState"
              :rules="rules"
              layout="vertical"
              @finish="handleSubmit"
            >
              <a-form-item
                label="KYX 用户名"
                name="kyx_username"
                :validate-status="validateStatus"
                :help="helpMessage"
              >
                <a-input
                  v-model:value="formState.kyx_username"
                  size="large"
                  placeholder="请输入您的 KYX 用户名"
                  :disabled="binding"
                  @blur="handleBlur"
                >
                  <template #prefix>
                    <UserOutlined class="text-gray-400" />
                  </template>
                </a-input>
              </a-form-item>

              <!-- 说明信息 -->
              <a-alert
                message="绑定说明"
                type="info"
                show-icon
                class="mb-6"
              >
                <template #description>
                  <ul class="text-sm space-y-1 mt-2">
                    <li>• 请确保填写的 KYX 用户名准确无误</li>
                    <li>• 绑定后将用于 API 配额的统一管理</li>
                    <li>• 每个 Linux.do 账号只能绑定一个 KYX 账号</li>
                    <li>• 绑定成功后如需更换，请联系管理员</li>
                  </ul>
                </template>
              </a-alert>

              <!-- 提交按钮 -->
              <a-form-item class="mb-0">
                <a-space direction="vertical" :size="12" class="w-full">
                  <a-button
                    type="primary"
                    html-type="submit"
                    size="large"
                    block
                    :loading="binding"
                  >
                    <template #icon>
                      <LinkOutlined />
                    </template>
                    确认绑定
                  </a-button>

                  <a-button
                    size="large"
                    block
                    @click="goToDashboard"
                    :disabled="binding"
                  >
                    取消
                  </a-button>
                </a-space>
              </a-form-item>
            </a-form>

            <!-- 帮助信息 -->
            <div class="mt-6 pt-6 border-t">
              <h4 class="text-sm font-medium text-gray-700 mb-3">
                <QuestionCircleOutlined class="mr-1" />
                常见问题
              </h4>
              <a-collapse ghost accordion>
                <a-collapse-panel key="1" header="什么是 KYX 账号？">
                  <p class="text-sm text-gray-600">
                    KYX 账号是用于 API 配额管理系统的账号标识，通过绑定可以统一管理您的 API 使用额度。
                  </p>
                </a-collapse-panel>
                <a-collapse-panel key="2" header="如何获取 KYX 用户名？">
                  <p class="text-sm text-gray-600">
                    请访问 KYX 平台查看您的账号信息，或联系管理员获取帮助。
                  </p>
                </a-collapse-panel>
                <a-collapse-panel key="3" header="绑定后可以更改吗？">
                  <p class="text-sm text-gray-600">
                    绑定后如需更换 KYX 账号，请联系系统管理员进行解绑和重新绑定操作。
                  </p>
                </a-collapse-panel>
                <a-collapse-panel key="4" header="绑定失败怎么办？">
                  <p class="text-sm text-gray-600">
                    请检查用户名是否正确，如果问题持续存在，请联系管理员寻求帮助。
                  </p>
                </a-collapse-panel>
              </a-collapse>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- 确认对话框 -->
    <a-modal
      v-model:open="confirmVisible"
      title="确认绑定"
      @ok="handleConfirmBind"
      @cancel="confirmVisible = false"
      :confirm-loading="binding"
    >
      <div class="space-y-4">
        <a-alert
          message="请仔细核对信息"
          type="warning"
          show-icon
          class="mb-4"
        />
        <p class="text-gray-700">
          您即将绑定以下 KYX 账号：
        </p>
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="Linux.do 用户名">
            {{ username }}
          </a-descriptions-item>
          <a-descriptions-item label="KYX 用户名">
            <strong class="text-blue-600">{{ formState.kyx_username }}</strong>
          </a-descriptions-item>
        </a-descriptions>
        <p class="text-sm text-gray-600">
          请确认信息无误后点击"确定"完成绑定。绑定成功后，如需更换请联系管理员。
        </p>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue'
import type { Rule } from 'ant-design-vue/es/form'
import {
  LinkOutlined,
  UserOutlined,
  CheckCircleOutlined,
  DashboardOutlined,
  GiftOutlined,
  QuestionCircleOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { BindAccountForm } from '@/types'
import dayjs from 'dayjs'

// ==================== Composables ====================
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()
const appStore = useAppStore()

// ==================== State ====================
const formRef = ref<FormInstance>()
const binding = ref(false)
const confirmVisible = ref(false)
const validateStatus = ref<'' | 'success' | 'warning' | 'error' | 'validating'>('')
const helpMessage = ref('')

const formState = reactive<BindAccountForm>({
  kyx_username: ''
})

// ==================== Computed ====================
const username = computed(() => authStore.username)
const linuxDoId = computed(() => authStore.linuxDoId)
const isBound = computed(() => userStore.isBound)
const kyxUsername = computed(() => userStore.kyxUsername)
const quota = computed(() => userStore.quota)

// ==================== Form Rules ====================
const rules: Record<string, Rule[]> = {
  kyx_username: [
    { required: true, message: '请输入 KYX 用户名', trigger: 'blur' },
    {
      min: 2,
      max: 50,
      message: '用户名长度应在 2-50 个字符之间',
      trigger: 'blur'
    },
    {
      pattern: /^[a-zA-Z0-9_-]+$/,
      message: '用户名只能包含字母、数字、下划线和连字符',
      trigger: 'blur'
    },
    {
      validator: async (_rule: Rule, value: string) => {
        if (!value) {
          return Promise.resolve()
        }

        // 检查是否包含空格
        if (value.includes(' ')) {
          return Promise.reject('用户名不能包含空格')
        }

        // 检查首尾是否有特殊字符
        if (value.startsWith('_') || value.startsWith('-') ||
            value.endsWith('_') || value.endsWith('-')) {
          return Promise.reject('用户名不能以下划线或连字符开头或结尾')
        }

        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ]
}

// ==================== Methods ====================

/**
 * 格式化日期
 */
const formatDate = (date: string): string => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

/**
 * 处理输入框失焦
 */
const handleBlur = () => {
  const value = formState.kyx_username.trim()
  if (value) {
    // 实时验证提示
    if (value.length < 2) {
      validateStatus.value = 'error'
      helpMessage.value = '用户名至少 2 个字符'
    } else if (value.length > 50) {
      validateStatus.value = 'error'
      helpMessage.value = '用户名最多 50 个字符'
    } else if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
      validateStatus.value = 'error'
      helpMessage.value = '用户名只能包含字母、数字、下划线和连字符'
    } else {
      validateStatus.value = 'success'
      helpMessage.value = '用户名格式正确'
    }
  } else {
    validateStatus.value = ''
    helpMessage.value = ''
  }
}

/**
 * 表单提交
 */
const handleSubmit = async () => {
  try {
    // 验证表单
    await formRef.value?.validate()

    // 去除首尾空格
    formState.kyx_username = formState.kyx_username.trim()

    // 显示确认对话框
    confirmVisible.value = true
  } catch (error) {
    console.error('Form validation failed:', error)
  }
}

/**
 * 确认绑定
 */
const handleConfirmBind = async () => {
  try {
    binding.value = true

    // 调用 Store 绑定方法
    const success = await userStore.bind({
      kyx_username: formState.kyx_username
    })

    if (success) {
      confirmVisible.value = false
      message.success('绑定成功！')

      // 短暂延迟后跳转到仪表板
      setTimeout(() => {
        router.push('/user/dashboard')
      }, 1500)
    }
  } catch (error: any) {
    console.error('Bind account failed:', error)
    message.error(error.message || '绑定失败，请重试')
  } finally {
    binding.value = false
  }
}

/**
 * 返回仪表板
 */
const goToDashboard = () => {
  router.push('/user/dashboard')
}

/**
 * 前往领取页面
 */
const goToClaim = () => {
  router.push('/user/claim')
}

/**
 * 加载用户数据
 */
const loadData = async () => {
  try {
    await userStore.fetchUserQuota()
  } catch (error) {
    console.error('Load data failed:', error)
  }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('绑定账号')

  // 加载用户数据
  await loadData()
})
</script>

<style scoped>
.bind-account-page {
  animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 绑定卡片样式 */
.bind-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-radius: 8px;
}

/* 图标动画 */
.text-5xl {
  animation: iconPulse 2s ease-in-out infinite;
}

@keyframes iconPulse {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.05);
    opacity: 0.9;
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
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

/* 按钮样式 */
:deep(.ant-btn-primary) {
  border-radius: 6px;
  font-weight: 500;
  transition: all 0.3s ease;
}

:deep(.ant-btn-primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

/* 成功结果样式 */
:deep(.ant-result-icon) {
  margin-bottom: 24px;
}

:deep(.ant-result-icon .anticon) {
  font-size: 72px;
}

/* 描述列表样式 */
:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  background-color: #f9fafb;
}

/* 折叠面板样式 */
:deep(.ant-collapse-ghost .ant-collapse-item) {
  border-bottom: 1px solid #e5e7eb;
}

:deep(.ant-collapse-header) {
  padding: 12px 0 !important;
  color: #374151 !important;
  font-weight: 500;
}

:deep(.ant-collapse-content-box) {
  padding: 12px 0 !important;
}

/* 警告框样式 */
:deep(.ant-alert) {
  border-radius: 6px;
}

/* 标签样式 */
:deep(.ant-tag) {
  border-radius: 4px;
  padding: 4px 12px;
  font-weight: 500;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .bind-card {
    margin: 0;
  }

  :deep(.ant-result-icon .anticon) {
    font-size: 56px;
  }

  .text-5xl {
    font-size: 3rem;
  }
}

/* 加载状态 */
:deep(.ant-form-item-has-feedback .ant-input) {
  padding-right: 30px;
}

/* 验证状态颜色 */
:deep(.ant-form-item-has-success .ant-input-affix-wrapper) {
  border-color: #52c41a;
}

:deep(.ant-form-item-has-error .ant-input-affix-wrapper) {
  border-color: #ff4d4f;
}

:deep(.ant-form-item-has-warning .ant-input-affix-wrapper) {
  border-color: #faad14;
}
</style>
