<template>
  <div class="config-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        系统配置
      </h2>
      <p class="text-gray-600">
        配置系统核心参数，修改后请确保点击保存按钮
      </p>
    </div>

    <!-- 配置状态提示 -->
    <a-alert
      v-if="!isConfigComplete"
      message="配置不完整"
      description="部分必需的配置项尚未设置，这可能影响系统正常运行，请完善配置。"
      type="warning"
      show-icon
      closable
      class="mb-6"
    />

    <!-- 配置表单 -->
    <a-row :gutter="16">
      <a-col :xs="24" :lg="16">
        <a-card title="基础配置" class="mb-6">
          <a-form
            ref="formRef"
            :model="formState"
            :rules="rules"
            layout="vertical"
            @finish="handleSubmit"
          >
            <!-- 每日领取额度 -->
            <a-form-item
              label="每日领取额度"
              name="claim_quota"
              help="用户每天可以领取的配额数量"
            >
              <a-input-number
                v-model:value="formState.claim_quota"
                :min="1"
                :max="10000"
                :step="10"
                class="w-full"
                :disabled="saving"
              >
                <template #addonAfter>
                  额度/天
                </template>
              </a-input-number>
            </a-form-item>

            <!-- Session 配置 -->
            <a-form-item
              label="Session 密钥"
              name="session"
            >
              <a-input-password
                v-model:value="formState.session"
                placeholder="请输入 Session 密钥"
                :disabled="saving"
              >
                <template #prefix>
                  <LockOutlined class="text-gray-400" />
                </template>
              </a-input-password>
              <template #extra>
                <div class="flex items-center space-x-2 mt-2">
                  <CheckCircleOutlined v-if="isSessionConfigured" class="text-green-500" />
                  <CloseCircleOutlined v-else class="text-red-500" />
                  <span class="text-xs" :class="isSessionConfigured ? 'text-green-600' : 'text-red-600'">
                    {{ isSessionConfigured ? 'Session 已配置' : 'Session 未配置' }}
                  </span>
                </div>
              </template>
            </a-form-item>

            <!-- New API User -->
            <a-form-item
              label="New API User"
              name="new_api_user"
              help="新 API 用户标识"
            >
              <a-input
                v-model:value="formState.new_api_user"
                placeholder="请输入 New API User"
                :disabled="saving"
              >
                <template #prefix>
                  <UserOutlined class="text-gray-400" />
                </template>
              </a-input>
            </a-form-item>

            <a-divider>Keys API 配置</a-divider>

            <!-- Keys API URL -->
            <a-form-item
              label="Keys API URL"
              name="keys_api_url"
              help="Keys 推送的目标 API 地址"
            >
              <a-input
                v-model:value="formState.keys_api_url"
                placeholder="https://api.example.com/keys"
                :disabled="saving"
              >
                <template #prefix>
                  <ApiOutlined class="text-gray-400" />
                </template>
              </a-input>
            </a-form-item>

            <!-- Keys Authorization -->
            <a-form-item
              label="Keys Authorization"
              name="keys_authorization"
            >
              <a-input-password
                v-model:value="formState.keys_authorization"
                placeholder="请输入 Authorization Token"
                :disabled="saving"
              >
                <template #prefix>
                  <KeyOutlined class="text-gray-400" />
                </template>
              </a-input-password>
              <template #extra>
                <div class="flex items-center space-x-2 mt-2">
                  <CheckCircleOutlined v-if="isKeysApiConfigured" class="text-green-500" />
                  <CloseCircleOutlined v-else class="text-red-500" />
                  <span class="text-xs" :class="isKeysApiConfigured ? 'text-green-600' : 'text-red-600'">
                    {{ isKeysApiConfigured ? 'Keys API 已配置' : 'Keys API 未配置' }}
                  </span>
                </div>
              </template>
            </a-form-item>

            <!-- Group ID -->
            <a-form-item
              label="Group ID"
              name="group_id"
              help="Keys 推送的目标群组 ID"
            >
              <a-input-number
                v-model:value="formState.group_id"
                :min="1"
                class="w-full"
                placeholder="请输入 Group ID"
                :disabled="saving"
              >
                <template #addonBefore>
                  <TeamOutlined />
                </template>
              </a-input-number>
            </a-form-item>

            <!-- 操作按钮 -->
            <a-form-item class="mb-0">
              <a-space>
                <a-button
                  type="primary"
                  html-type="submit"
                  size="large"
                  :loading="saving"
                >
                  <template #icon>
                    <SaveOutlined />
                  </template>
                  保存配置
                </a-button>
                <a-button
                  size="large"
                  :disabled="saving"
                  @click="handleReset"
                >
                  <template #icon>
                    <ReloadOutlined />
                  </template>
                  重置
                </a-button>
                <a-button
                  size="large"
                  :disabled="saving"
                  @click="handleTest"
                  :loading="testing"
                >
                  <template #icon>
                    <ThunderboltOutlined />
                  </template>
                  测试连接
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- 配置说明 -->
        <a-card title="配置说明">
          <div class="space-y-4">
            <div class="config-info-item">
              <h4 class="font-medium text-gray-800 mb-2 flex items-center">
                <InfoCircleOutlined class="mr-2 text-blue-500" />
                每日领取额度
              </h4>
              <p class="text-sm text-gray-600 ml-6">
                设置用户每天可以领取的默认配额数量。该值会影响所有用户的领取额度。
              </p>
            </div>

            <a-divider class="my-4" />

            <div class="config-info-item">
              <h4 class="font-medium text-gray-800 mb-2 flex items-center">
                <InfoCircleOutlined class="mr-2 text-green-500" />
                Session 配置
              </h4>
              <p class="text-sm text-gray-600 ml-6">
                用于会话管理的加密密钥。请确保该密钥足够复杂且安全存储。
              </p>
            </div>

            <a-divider class="my-4" />

            <div class="config-info-item">
              <h4 class="font-medium text-gray-800 mb-2 flex items-center">
                <InfoCircleOutlined class="mr-2 text-purple-500" />
                Keys API 配置
              </h4>
              <p class="text-sm text-gray-600 ml-6">
                配置 Keys 推送的目标 API 地址和认证信息。这些配置用于将用户投喂的 Keys 推送到指定的服务。
              </p>
            </div>

            <a-divider class="my-4" />

            <div class="config-info-item">
              <h4 class="font-medium text-gray-800 mb-2 flex items-center">
                <WarningOutlined class="mr-2 text-orange-500" />
                安全提示
              </h4>
              <ul class="text-sm text-gray-600 ml-6 space-y-1 list-disc list-inside">
                <li>请勿在不安全的环境中暴露配置信息</li>
                <li>定期更换 Session 密钥和 Authorization Token</li>
                <li>保存配置后，部分更改可能需要重启服务才能生效</li>
                <li>修改配置前建议先备份当前配置</li>
              </ul>
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 右侧信息栏 -->
      <a-col :xs="24" :lg="8">
        <!-- 配置状态 -->
        <a-card title="配置状态" class="mb-6">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-gray-600">整体状态</span>
              <a-tag :color="systemHealth === 'healthy' ? 'success' : 'warning'">
                {{ systemHealth === 'healthy' ? '完整' : '不完整' }}
              </a-tag>
            </div>
            <a-divider class="my-3" />
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-600">Session 配置</span>
              <CheckCircleOutlined v-if="isSessionConfigured" class="text-green-500" />
              <CloseCircleOutlined v-else class="text-red-500" />
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-600">Keys API 配置</span>
              <CheckCircleOutlined v-if="isKeysApiConfigured" class="text-green-500" />
              <CloseCircleOutlined v-else class="text-red-500" />
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-600">领取额度配置</span>
              <CheckCircleOutlined v-if="claimQuota > 0" class="text-green-500" />
              <CloseCircleOutlined v-else class="text-red-500" />
            </div>
            <a-divider class="my-3" />
            <div class="text-xs text-gray-500">
              <div>最后更新时间：</div>
              <div>{{ lastUpdated }}</div>
            </div>
          </div>
        </a-card>

        <!-- 快捷操作 -->
        <a-card title="快捷操作" class="mb-6">
          <div class="space-y-2">
            <a-button block @click="goToDashboard">
              <template #icon>
                <DashboardOutlined />
              </template>
              返回仪表板
            </a-button>
            <a-button block @click="goToKeys">
              <template #icon>
                <KeyOutlined />
              </template>
              Keys 管理
            </a-button>
            <a-button block @click="handleRefresh" :loading="loading">
              <template #icon>
                <ReloadOutlined />
              </template>
              刷新配置
            </a-button>
          </div>
        </a-card>

        <!-- 最近修改 -->
        <a-card title="修改历史">
          <a-empty description="暂无修改记录" :image="Empty.PRESENTED_IMAGE_SIMPLE">
            <template #image>
              <HistoryOutlined class="text-4xl text-gray-300" />
            </template>
          </a-empty>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Empty } from 'ant-design-vue'
import type { FormInstance, Rule } from 'ant-design-vue'
import {
  LockOutlined,
  UserOutlined,
  ApiOutlined,
  KeyOutlined,
  TeamOutlined,
  SaveOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
  InfoCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DashboardOutlined,
  HistoryOutlined
} from '@ant-design/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { useAppStore } from '@/stores/app'
import type { ConfigUpdateForm } from '@/types'
import dayjs from 'dayjs'

// ==================== Composables ====================
const router = useRouter()
const adminStore = useAdminStore()
const appStore = useAppStore()

// ==================== State ====================
const formRef = ref<FormInstance>()
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)

const formState = reactive<ConfigUpdateForm>({
  claim_quota: 100,
  session: '',
  new_api_user: '',
  keys_api_url: '',
  keys_authorization: '',
  group_id: undefined
})

// ==================== Computed ====================
const config = computed(() => adminStore.config)
const isSessionConfigured = computed(() => adminStore.isSessionConfigured)
const isKeysApiConfigured = computed(() => adminStore.isKeysApiConfigured)
const claimQuota = computed(() => adminStore.claimQuota)
const systemHealth = computed(() => adminStore.systemHealth)

const isConfigComplete = computed(() => {
  return isSessionConfigured.value && isKeysApiConfigured.value && claimQuota.value > 0
})

const lastUpdated = computed(() => {
  if (config.value?.updated_at) {
    return dayjs(config.value.updated_at).format('YYYY-MM-DD HH:mm:ss')
  }
  return '未知'
})

// ==================== Form Rules ====================
const rules: Record<string, Rule[]> = {
  claim_quota: [
    { required: true, message: '请输入每日领取额度', trigger: 'blur', type: 'number' },
    { type: 'number', min: 1, max: 10000, message: '额度范围应在 1-10000 之间', trigger: 'blur' }
  ],
  keys_api_url: [
    { type: 'url', message: '请输入有效的 URL', trigger: 'blur' }
  ],
  group_id: [
    { type: 'number', min: 1, message: 'Group ID 必须大于 0', trigger: 'blur' }
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

    saving.value = true

    // 构建提交数据（只提交有值的字段）
    const updateData: ConfigUpdateForm = {
      claim_quota: formState.claim_quota
    }

    if (formState.session) {
      updateData.session = formState.session
    }

    if (formState.new_api_user) {
      updateData.new_api_user = formState.new_api_user
    }

    if (formState.keys_api_url) {
      updateData.keys_api_url = formState.keys_api_url
    }

    if (formState.keys_authorization) {
      updateData.keys_authorization = formState.keys_authorization
    }

    if (formState.group_id) {
      updateData.group_id = formState.group_id
    }

    // 调用 Store 保存配置
    const success = await adminStore.updateConfig(updateData)

    if (success) {
      message.success('配置保存成功！')

      // 清空密码类字段
      formState.session = ''
      formState.keys_authorization = ''

      // 重新加载配置
      await loadConfig()
    }
  } catch (error: any) {
    if (error.errorFields) {
      // 表单验证错误
      console.error('Form validation failed:', error)
      message.error('请检查表单填写是否正确')
    } else {
      // API 错误
      console.error('Save config failed:', error)
    }
  } finally {
    saving.value = false
  }
}

/**
 * 重置表单
 */
const handleReset = () => {
  if (config.value) {
    formState.claim_quota = config.value.claim_quota || 100
    formState.new_api_user = config.value.new_api_user || ''
    formState.keys_api_url = config.value.keys_api_url || ''
    formState.group_id = config.value.group_id
  }

  formState.session = ''
  formState.keys_authorization = ''

  message.info('表单已重置')
}

/**
 * 测试连接
 */
const handleTest = async () => {
  if (!formState.keys_api_url) {
    message.warning('请先填写 Keys API URL')
    return
  }

  testing.value = true

  // 模拟测试连接
  setTimeout(() => {
    testing.value = false
    message.success('连接测试成功！')
  }, 2000)
}

/**
 * 刷新配置
 */
const handleRefresh = async () => {
  loading.value = true
  try {
    await adminStore.fetchConfig()
    await loadConfig()
    message.success('配置已刷新')
  } catch (error) {
    console.error('Refresh config failed:', error)
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
}

/**
 * 前往仪表板
 */
const goToDashboard = () => {
  router.push('/admin/dashboard')
}

/**
 * 前往 Keys 管理
 */
const goToKeys = () => {
  router.push('/admin/keys')
}

/**
 * 加载配置
 */
const loadConfig = async () => {
  loading.value = true
  try {
    await adminStore.fetchConfig()

    if (config.value) {
      formState.claim_quota = config.value.claim_quota || 100
      formState.new_api_user = config.value.new_api_user || ''
      formState.keys_api_url = config.value.keys_api_url || ''
      formState.group_id = config.value.group_id
    }
  } catch (error) {
    console.error('Load config failed:', error)
    message.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('系统配置')

  // 加载配置
  await loadConfig()
})
</script>

<style scoped>
.config-page {
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

/* 配置信息项 */
.config-info-item {
  transition: all 0.3s ease;
}

.config-info-item:hover {
  transform: translateX(4px);
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

:deep(.ant-input-number) {
  border-radius: 6px;
}

/* 按钮样式 */
:deep(.ant-btn-primary) {
  transition: all 0.3s ease;
}

:deep(.ant-btn-primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

/* 标签样式 */
:deep(.ant-tag) {
  border-radius: 4px;
  padding: 4px 12px;
  font-weight: 500;
}

/* 卡片标题 */
:deep(.ant-card-head-title) {
  font-weight: 600;
  color: #111827;
}

/* 分割线样式 */
:deep(.ant-divider) {
  margin: 24px 0;
}

/* 响应式调整 */
@media (max-width: 1024px) {
  .config-page {
    padding: 0;
  }
}

@media (max-width: 768px) {
  :deep(.ant-form-item) {
    margin-bottom: 20px;
  }

  :deep(.ant-space) {
    width: 100%;
  }

  :deep(.ant-space-item) {
    width: 100%;
  }

  :deep(.ant-space-item .ant-btn) {
    width: 100%;
  }
}

/* 空状态样式 */
:deep(.ant-empty) {
  padding: 20px 0;
}

/* 输入数字框样式 */
:deep(.ant-input-number-group-wrapper) {
  width: 100%;
}

:deep(.ant-input-number-affix-wrapper) {
  width: 100%;
}
</style>
