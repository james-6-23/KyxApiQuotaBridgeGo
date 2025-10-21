<template>
  <div class="donate-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        投喂 API Keys
      </h2>
      <p class="text-gray-600">
        通过投喂有效的 API Keys 获取额外的配额奖励
      </p>
    </div>

    <!-- 未绑定提示 -->
    <a-alert
      v-if="!isBound"
      message="尚未绑定 KYX 账号"
      description="请先绑定 KYX 账号才能投喂 Keys"
      type="warning"
      show-icon
      closable
      class="mb-6"
    >
      <template #action>
        <a-button type="primary" size="small" @click="goToBind">
          立即绑定
        </a-button>
      </template>
    </a-alert>

    <!-- 投喂区域 -->
    <a-row :gutter="16" class="mb-6">
      <!-- Keys 输入区 -->
      <a-col :xs="24" :lg="14">
        <a-card title="输入 API Keys" class="donate-card">
          <template #extra>
            <a-space>
              <a-tooltip title="清空输入">
                <a-button
                  type="text"
                  size="small"
                  :disabled="!keysInput || donating"
                  @click="handleClear"
                >
                  <template #icon>
                    <DeleteOutlined />
                  </template>
                </a-button>
              </a-tooltip>
              <a-tooltip title="从剪贴板粘贴">
                <a-button
                  type="text"
                  size="small"
                  :disabled="donating"
                  @click="handlePaste"
                >
                  <template #icon>
                    <CopyOutlined />
                  </template>
                </a-button>
              </a-tooltip>
            </a-space>
          </template>

          <div class="space-y-4">
            <!-- Keys 输入框 -->
            <a-textarea
              v-model:value="keysInput"
              placeholder="请输入 API Keys，每行一个&#10;例如：&#10;sk-xxxxxxxxxxxxxxxxxxxxx&#10;sk-yyyyyyyyyyyyyyyyyyyyy&#10;sk-zzzzzzzzzzzzzzzzzzzzz"
              :rows="12"
              :disabled="donating"
              :maxlength="50000"
              show-count
              @input="handleKeysInput"
              class="keys-textarea"
            />

            <!-- 实时统计 -->
            <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
              <div class="flex items-center space-x-6">
                <div>
                  <span class="text-gray-600 text-sm">Keys 数量：</span>
                  <span class="text-lg font-semibold text-blue-600">{{ parsedKeys.length }}</span>
                </div>
                <div>
                  <span class="text-gray-600 text-sm">有效格式：</span>
                  <span class="text-lg font-semibold text-green-600">{{ validKeysCount }}</span>
                </div>
                <div>
                  <span class="text-gray-600 text-sm">无效格式：</span>
                  <span class="text-lg font-semibold text-red-600">{{ invalidKeysCount }}</span>
                </div>
              </div>
            </div>

            <!-- 错误提示 -->
            <a-alert
              v-if="invalidKeysCount > 0"
              :message="`发现 ${invalidKeysCount} 个格式不正确的 Keys`"
              description="请检查并修正格式错误的 Keys，或删除它们后再提交"
              type="warning"
              show-icon
              closable
            />

            <!-- 操作按钮 -->
            <a-space class="w-full">
              <a-button
                type="primary"
                size="large"
                :disabled="!canSubmit"
                :loading="donating"
                @click="handleSubmit"
                class="donate-button"
              >
                <template #icon>
                  <HeartOutlined />
                </template>
                投喂 {{ validKeysCount }} 个 Keys
              </a-button>
              <a-button
                size="large"
                :disabled="donating"
                @click="handlePreview"
              >
                <template #icon>
                  <EyeOutlined />
                </template>
                预览
              </a-button>
            </a-space>

            <!-- 帮助提示 -->
            <a-alert
              message="投喂说明"
              type="info"
              show-icon
            >
              <template #description>
                <ul class="text-sm space-y-1 mt-2">
                  <li>• 每个有效的 API Key 可获得相应的额度奖励</li>
                  <li>• Keys 会在后台进行验证，验证通过后才会添加额度</li>
                  <li>• 支持批量输入，每行一个 Key</li>
                  <li>• Keys 格式应为 sk- 开头的完整密钥</li>
                  <li>• 投喂后可在记录中查看处理状态</li>
                </ul>
              </template>
            </a-alert>
          </div>
        </a-card>
      </a-col>

      <!-- 统计信息 -->
      <a-col :xs="24" :lg="10">
        <a-card title="投喂统计" class="h-full">
          <div class="space-y-4">
            <!-- 统计卡片 -->
            <div class="stat-item bg-gradient-to-r from-purple-50 to-pink-50 p-4 rounded-lg">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-700 font-medium">累计投喂次数</span>
                <HeartFilled class="text-2xl text-pink-500" />
              </div>
              <div class="text-3xl font-bold text-purple-600">
                {{ userStats?.donateCount || 0 }}
              </div>
            </div>

            <div class="stat-item bg-gradient-to-r from-blue-50 to-cyan-50 p-4 rounded-lg">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-700 font-medium">累计投喂 Keys</span>
                <KeyOutlined class="text-2xl text-blue-500" />
              </div>
              <div class="text-3xl font-bold text-blue-600">
                {{ formatNumber(totalDonatedKeys) }}
              </div>
            </div>

            <div class="stat-item bg-gradient-to-r from-green-50 to-emerald-50 p-4 rounded-lg">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-700 font-medium">获得额度</span>
                <GiftFilled class="text-2xl text-green-500" />
              </div>
              <div class="text-3xl font-bold text-green-600">
                {{ formatNumber(userStats?.donateQuota || 0) }}
              </div>
            </div>

            <div class="stat-item bg-gradient-to-r from-orange-50 to-amber-50 p-4 rounded-lg">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-700 font-medium">当前余额</span>
                <WalletFilled class="text-2xl text-orange-500" />
              </div>
              <div class="text-3xl font-bold text-orange-600">
                {{ formatNumber(currentBalance) }}
              </div>
            </div>

            <a-divider class="my-4" />

            <!-- 快捷操作 -->
            <div class="space-y-2">
              <a-button block @click="goToDashboard">
                <template #icon>
                  <DashboardOutlined />
                </template>
                返回仪表板
              </a-button>
              <a-button block @click="goToClaim">
                <template #icon>
                  <GiftOutlined />
                </template>
                领取每日额度
              </a-button>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 投喂记录 -->
    <a-card title="投喂记录" class="mb-6">
      <template #extra>
        <a-space>
          <a-button
            type="text"
            :loading="loading"
            @click="refreshRecords"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
            刷新
          </a-button>
        </a-space>
      </template>

      <a-table
        :columns="columns"
        :data-source="donateRecords"
        :loading="loading"
        :pagination="paginationConfig"
        @change="handleTableChange"
        :scroll="{ x: 900 }"
        row-key="id"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'index'">
            {{ (pagination.current - 1) * pagination.pageSize + index + 1 }}
          </template>

          <template v-if="column.key === 'keys_count'">
            <a-space>
              <a-tag color="blue">{{ record.keys_count }} 个</a-tag>
              <a-tooltip v-if="record.valid_keys_count !== undefined" title="查看详情">
                <a-button type="link" size="small" @click="showDetail(record)">
                  详情
                </a-button>
              </a-tooltip>
            </a-space>
          </template>

          <template v-if="column.key === 'status'">
            <div class="space-y-1">
              <div>
                <a-tag color="success">有效: {{ record.valid_keys_count || 0 }}</a-tag>
                <a-tag color="error">无效: {{ record.invalid_keys_count || 0 }}</a-tag>
              </div>
              <div>
                <a-tag v-if="record.push_status === 'success'" color="success">
                  <CheckCircleOutlined /> 推送成功
                </a-tag>
                <a-tag v-else-if="record.push_status === 'failed'" color="error">
                  <CloseCircleOutlined /> 推送失败
                </a-tag>
                <a-tag v-else color="processing">
                  <SyncOutlined :spin="true" /> 处理中
                </a-tag>
              </div>
            </div>
          </template>

          <template v-if="column.key === 'quota'">
            <a-tag color="success" class="font-medium">
              +{{ record.total_quota_added }}
            </a-tag>
          </template>

          <template v-if="column.key === 'timestamp'">
            <div class="space-y-1">
              <div>{{ formatDate(record.timestamp) }}</div>
              <div class="text-xs text-gray-500">
                {{ formatRelativeTime(record.timestamp) }}
              </div>
            </div>
          </template>
        </template>

        <template #emptyText>
          <a-empty description="暂无投喂记录">
            <template #image>
              <InboxOutlined class="text-6xl text-gray-300" />
            </template>
          </a-empty>
        </template>
      </a-table>
    </a-card>

    <!-- 确认对话框 -->
    <a-modal
      v-model:open="confirmVisible"
      title="确认投喂"
      width="600px"
      @ok="handleConfirmDonate"
      @cancel="confirmVisible = false"
      :confirm-loading="donating"
      ok-text="确认投喂"
      cancel-text="取消"
    >
      <div class="space-y-4">
        <a-alert
          message="请仔细核对信息"
          type="info"
          show-icon
          class="mb-4"
        />

        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="Keys 总数">
            <span class="font-semibold text-blue-600">{{ parsedKeys.length }}</span>
          </a-descriptions-item>
          <a-descriptions-item label="有效格式">
            <span class="font-semibold text-green-600">{{ validKeysCount }}</span>
          </a-descriptions-item>
          <a-descriptions-item label="无效格式">
            <span class="font-semibold text-red-600">{{ invalidKeysCount }}</span>
          </a-descriptions-item>
          <a-descriptions-item label="预计获得额度">
            <span class="font-semibold text-purple-600">取决于验证结果</span>
          </a-descriptions-item>
        </a-descriptions>

        <a-alert
          v-if="invalidKeysCount > 0"
          message="注意"
          :description="`有 ${invalidKeysCount} 个 Keys 格式不正确，将被忽略。只有格式正确的 Keys 会被提交。`"
          type="warning"
          show-icon
        />

        <p class="text-sm text-gray-600">
          确认无误后点击"确认投喂"，系统将验证这些 Keys 并为有效的 Keys 添加相应额度。
        </p>
      </div>
    </a-modal>

    <!-- 预览对话框 -->
    <a-modal
      v-model:open="previewVisible"
      title="Keys 预览"
      width="700px"
      :footer="null"
    >
      <div class="space-y-4">
        <a-alert
          :message="`共 ${parsedKeys.length} 个 Keys（有效: ${validKeysCount}, 无效: ${invalidKeysCount}）`"
          type="info"
          show-icon
        />

        <a-tabs v-model:activeKey="previewTab">
          <a-tab-pane key="all" tab="全部 Keys">
            <div class="max-h-96 overflow-y-auto">
              <div
                v-for="(key, index) in parsedKeys"
                :key="index"
                class="flex items-center justify-between p-2 hover:bg-gray-50 rounded"
              >
                <div class="flex-1 font-mono text-sm truncate">
                  {{ key }}
                </div>
                <a-tag :color="isValidKey(key) ? 'success' : 'error'" class="ml-2">
                  {{ isValidKey(key) ? '有效' : '无效' }}
                </a-tag>
              </div>
            </div>
          </a-tab-pane>

          <a-tab-pane key="valid" :tab="`有效 Keys (${validKeysCount})`">
            <div class="max-h-96 overflow-y-auto">
              <div
                v-for="(key, index) in validKeys"
                :key="index"
                class="p-2 hover:bg-gray-50 rounded font-mono text-sm"
              >
                {{ key }}
              </div>
            </div>
          </a-tab-pane>

          <a-tab-pane key="invalid" :tab="`无效 Keys (${invalidKeysCount})`">
            <div class="max-h-96 overflow-y-auto">
              <div
                v-for="(key, index) in invalidKeys"
                :key="index"
                class="p-2 hover:bg-gray-50 rounded font-mono text-sm text-red-600"
              >
                {{ key }}
              </div>
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-modal>

    <!-- 详情对话框 -->
    <a-modal
      v-model:open="detailVisible"
      title="投喂详情"
      width="700px"
      :footer="null"
    >
      <div v-if="selectedRecord" class="space-y-4">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="投喂时间" :span="2">
            {{ formatDate(selectedRecord.timestamp) }}
          </a-descriptions-item>
          <a-descriptions-item label="Keys 总数">
            {{ selectedRecord.keys_count }}
          </a-descriptions-item>
          <a-descriptions-item label="有效 Keys">
            <a-tag color="success">{{ selectedRecord.valid_keys_count || 0 }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="无效 Keys">
            <a-tag color="error">{{ selectedRecord.invalid_keys_count || 0 }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="获得额度">
            <a-tag color="success">+{{ selectedRecord.total_quota_added }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="推送状态" :span="2">
            <a-tag v-if="selectedRecord.push_status === 'success'" color="success">
              推送成功
            </a-tag>
            <a-tag v-else-if="selectedRecord.push_status === 'failed'" color="error">
              推送失败
            </a-tag>
            <a-tag v-else color="processing">
              处理中
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item v-if="selectedRecord.push_message" label="推送消息" :span="2">
            {{ selectedRecord.push_message }}
          </a-descriptions-item>
        </a-descriptions>

        <div v-if="selectedRecord.failed_keys && selectedRecord.failed_keys.length > 0">
          <a-divider>失败的 Keys</a-divider>
          <div class="max-h-60 overflow-y-auto bg-gray-50 p-3 rounded">
            <div
              v-for="(key, index) in selectedRecord.failed_keys"
              :key="index"
              class="font-mono text-sm text-red-600 mb-1"
            >
              {{ key }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import {
  HeartOutlined,
  HeartFilled,
  DeleteOutlined,
  CopyOutlined,
  EyeOutlined,
  ReloadOutlined,
  DashboardOutlined,
  GiftOutlined,
  GiftFilled,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  InboxOutlined,
  KeyOutlined,
  WalletFilled
} from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { DonateRecord } from '@/types'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Composables ====================
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

// ==================== State ====================
const loading = ref(false)
const donating = ref(false)
const keysInput = ref('')
const confirmVisible = ref(false)
const previewVisible = ref(false)
const detailVisible = ref(false)
const previewTab = ref('all')
const selectedRecord = ref<DonateRecord | null>(null)

// ==================== Computed ====================
const isBound = computed(() => userStore.isBound)
const currentBalance = computed(() => userStore.currentBalance)
const donateRecords = computed(() => userStore.donateRecords)
const pagination = computed(() => userStore.donatePagination)
const userStats = computed(() => userStore.userStats)

// 解析 Keys
const parsedKeys = computed(() => {
  if (!keysInput.value.trim()) return []

  return keysInput.value
    .split('\n')
    .map(key => key.trim())
    .filter(key => key.length > 0)
    .filter((key, index, self) => self.indexOf(key) === index) // 去重
})

// 验证 Key 格式
const isValidKey = (key: string): boolean => {
  // 基本格式验证：sk- 开头，长度合理
  return /^sk-[a-zA-Z0-9]{20,}$/.test(key)
}

// 有效的 Keys
const validKeys = computed(() => {
  return parsedKeys.value.filter(key => isValidKey(key))
})

// 无效的 Keys
const invalidKeys = computed(() => {
  return parsedKeys.value.filter(key => !isValidKey(key))
})

// 有效 Keys 数量
const validKeysCount = computed(() => validKeys.value.length)

// 无效 Keys 数量
const invalidKeysCount = computed(() => invalidKeys.value.length)

// 是否可以提交
const canSubmit = computed(() => {
  return isBound.value && validKeysCount.value > 0 && !donating.value
})

// 总投喂 Keys 数
const totalDonatedKeys = computed(() => {
  return donateRecords.value.reduce((sum, record) => sum + record.keys_count, 0)
})

// 分页配置
const paginationConfig = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`,
  pageSizeOptions: ['10', '20', '50', '100']
}))

// ==================== Table Columns ====================
const columns: TableColumnsType = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    align: 'center'
  },
  {
    title: 'Keys 数量',
    dataIndex: 'keys_count',
    key: 'keys_count',
    width: 150,
    align: 'center'
  },
  {
    title: '验证状态',
    key: 'status',
    width: 200
  },
  {
    title: '获得额度',
    dataIndex: 'total_quota_added',
    key: 'quota',
    width: 120,
    align: 'center'
  },
  {
    title: '投喂时间',
    dataIndex: 'timestamp',
    key: 'timestamp',
    width: 200
  }
]

// ==================== Methods ====================

/**
 * 格式化数字
 */
const formatNumber = (num: number): string => {
  return num.toLocaleString('zh-CN')
}

/**
 * 格式化日期
 */
const formatDate = (date: string): string => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

/**
 * 格式化相对时间
 */
const formatRelativeTime = (date: string): string => {
  return dayjs(date).fromNow()
}

/**
 * 处理 Keys 输入
 */
const handleKeysInput = () => {
  // 可以在这里添加实时验证逻辑
}

/**
 * 清空输入
 */
const handleClear = () => {
  keysInput.value = ''
  message.info('已清空输入')
}

/**
 * 从剪贴板粘贴
 */
const handlePaste = async () => {
  try {
    const text = await navigator.clipboard.readText()
    if (text) {
      keysInput.value = text
      message.success('已从剪贴板粘贴')
    }
  } catch (error) {
    message.error('读取剪贴板失败，请手动粘贴')
  }
}

/**
 * 预览 Keys
 */
const handlePreview = () => {
  if (parsedKeys.value.length === 0) {
    message.warning('请先输入 Keys')
    return
  }
  previewVisible.value = true
  previewTab.value = 'all'
}

/**
 * 提交投喂
 */
const handleSubmit = () => {
  if (!isBound.value) {
    message.warning('请先绑定 KYX 账号')
    return
  }

  if (validKeysCount.value === 0) {
    message.error('没有有效的 Keys')
    return
  }

  confirmVisible.value = true
}

/**
 * 确认投喂
 */
const handleConfirmDonate = async () => {
  try {
    donating.value = true

    const success = await userStore.donate({
      keys: validKeys.value
    })

    if (success) {
      confirmVisible.value = false
      keysInput.value = ''

      message.success({
        content: `成功投喂 ${validKeysCount.value} 个 Keys！`,
        duration: 3
      })

      // 刷新记录
      await refreshRecords()
    }
  } catch (error: any) {
    console.error('Donate failed:', error)
    message.error(error.message || '投喂失败，请重试')
  } finally {
    donating.value = false
  }
}

/**
 * 显示详情
 */
const showDetail = (record: DonateRecord) => {
  selectedRecord.value = record
  detailVisible.value = true
}

/**
 * 刷新记录
 */
const refreshRecords = async () => {
  loading.value = true
  try {
    await Promise.all([
      userStore.fetchUserQuota(),
      userStore.fetchUserStats(),
      userStore.fetchDonateRecords({
        page: pagination.value.current,
        page_size: pagination.value.pageSize
      })
    ])
  } catch (error) {
    console.error('Refresh records failed:', error)
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
}

/**
 * 处理表格变化
 */
const handleTableChange: TableProps['onChange'] = (paginationObj) => {
  if (paginationObj.current && paginationObj.pageSize) {
    userStore.updateDonatePagination(paginationObj.current, paginationObj.pageSize)
  }
}

/**
 * 前往绑定页面
 */
const goToBind = () => {
  router.push('/user/bind')
}

/**
 * 前往仪表板
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
 * 加载数据
 */
const loadData = async () => {
  loading.value = true
  try {
    await Promise.all([
      userStore.fetchUserQuota(),
      userStore.fetchUserStats(),
      userStore.fetchDonateRecords({ page: 1, page_size: 10 })
    ])
  } catch (error) {
    console.error('Load data failed:', error)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('投喂 Keys')

  // 加载数据
  await loadData()
})
</script>

<style scoped>
.donate-page {
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

/* 投喂卡片 */
.donate-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-radius: 8px;
}

/* Keys 输入框 */
.keys-textarea {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
}

:deep(.keys-textarea .ant-input) {
  resize: vertical;
}

/* 投喂按钮 */
.donate-button {
  transition: all 0.3s ease;
}

.donate-button:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(236, 72, 153, 0.3);
}

/* 统计项 */
.stat-item {
  transition: all 0.3s ease;
}

.stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 表格样式 */
:deep(.ant-table) {
  font-size: 14px;
}

:deep(.ant-table-thead > tr > th) {
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
}

:deep(.ant-table-tbody > tr:hover > td) {
  background: #f9fafb;
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

/* 描述列表样式 */
:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  color: #6b7280;
}

:deep(.ant-descriptions-item-content) {
  color: #111827;
}

/* 对话框样式 */
:deep(.ant-modal-header) {
  border-bottom: 1px solid #e5e7eb;
}

/* 标签页样式 */
:deep(.ant-tabs-tab) {
  font-weight: 500;
}

/* 响应式调整 */
@media (max-width: 1024px) {
  .donate-button {
    width: 100%;
  }

  .stat-item {
    font-size: 14px;
  }
}

@media (max-width: 768px) {
  :deep(.ant-table) {
    font-size: 13px;
  }

  .keys-textarea {
    font-size: 12px;
  }

  .stat-item div:first-child {
    font-size: 13px;
  }

  .stat-item div:last-child {
    font-size: 24px;
  }
}

/* 空状态样式 */
:deep(.ant-empty) {
  padding: 40px 0;
}

/* 分页样式 */
:deep(.ant-pagination) {
  margin-top: 24px;
  text-align: center;
}

/* 输入框计数器样式 */
:deep(.ant-input-textarea-show-count::after) {
  color: #9ca3af;
  font-size: 12px;
}

/* 警告框样式 */
:deep(.ant-alert) {
  border-radius: 6px;
}

/* 滚动条样式 */
.max-h-96::-webkit-scrollbar,
.max-h-60::-webkit-scrollbar {
  width: 6px;
}

.max-h-96::-webkit-scrollbar-track,
.max-h-60::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.max-h-96::-webkit-scrollbar-thumb,
.max-h-60::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

.max-h-96::-webkit-scrollbar-thumb:hover,
.max-h-60::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 渐变背景动画 */
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

.bg-gradient-to-r {
  background-size: 200% 200%;
  animation: gradientShift 10s ease infinite;
}
</style>
