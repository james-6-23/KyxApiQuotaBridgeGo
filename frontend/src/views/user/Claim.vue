<template>
  <div class="claim-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        领取每日额度
      </h2>
      <p class="text-gray-600">
        每天可以领取一次免费额度，领取后将自动添加到您的账户余额中
      </p>
    </div>

    <!-- 未绑定提示 -->
    <a-alert
      v-if="!isBound"
      message="尚未绑定 KYX 账号"
      description="请先绑定 KYX 账号才能领取额度"
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

    <!-- 领取卡片 -->
    <a-row :gutter="16" class="mb-6">
      <a-col :xs="24" :lg="12">
        <a-card class="claim-card" :loading="loading">
          <div class="text-center py-6">
            <!-- 状态图标 -->
            <div class="mb-6">
              <div v-if="canClaimToday" class="status-icon success">
                <GiftOutlined class="text-6xl" />
              </div>
              <div v-else class="status-icon disabled">
                <ClockCircleOutlined class="text-6xl" />
              </div>
            </div>

            <!-- 状态信息 -->
            <h3 class="text-2xl font-bold mb-2" :class="canClaimToday ? 'text-green-600' : 'text-gray-500'">
              {{ canClaimToday ? '可以领取' : '今日已领取' }}
            </h3>
            <p class="text-gray-600 mb-6">
              {{ statusMessage }}
            </p>

            <!-- 领取按钮 -->
            <a-button
              type="primary"
              size="large"
              :disabled="!canClaimToday || !isBound || claiming"
              :loading="claiming"
              @click="handleClaim"
              class="claim-button"
            >
              <template #icon>
                <GiftOutlined />
              </template>
              {{ canClaimToday ? '立即领取' : '今日已领取' }}
            </a-button>

            <!-- 额外信息 -->
            <div class="mt-6 pt-6 border-t">
              <a-descriptions :column="1" size="small">
                <a-descriptions-item label="每日可领取额度">
                  <span class="text-lg font-semibold text-blue-600">
                    {{ claimQuota || '加载中...' }}
                  </span>
                </a-descriptions-item>
                <a-descriptions-item v-if="lastClaimDate" label="上次领取时间">
                  {{ formatDate(lastClaimDate) }}
                </a-descriptions-item>
                <a-descriptions-item v-if="nextClaimTime" label="下次可领取时间">
                  {{ nextClaimTime }}
                </a-descriptions-item>
              </a-descriptions>
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 统计信息 -->
      <a-col :xs="24" :lg="12">
        <a-card title="领取统计" class="h-full">
          <div class="space-y-4">
            <div class="stat-item">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-600">累计领取次数</span>
                <span class="text-2xl font-bold text-blue-600">
                  {{ userStats?.claim_count || 0 }}
                </span>
              </div>
              <a-progress
                :percent="getClaimProgress()"
                :show-info="false"
                stroke-color="#3b82f6"
              />
            </div>

            <div class="stat-item">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-600">累计领取额度</span>
                <span class="text-2xl font-bold text-green-600">
                  {{ formatNumber(userStats?.claim_quota || 0) }}
                </span>
              </div>
              <a-progress
                :percent="getQuotaProgress()"
                :show-info="false"
                stroke-color="#22c55e"
              />
            </div>

            <div class="stat-item">
              <div class="flex items-center justify-between mb-2">
                <span class="text-gray-600">当前余额</span>
                <span class="text-2xl font-bold text-purple-600">
                  {{ formatNumber(currentBalance) }}
                </span>
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
              <a-button block @click="goToDonate">
                <template #icon>
                  <HeartOutlined />
                </template>
                投喂 Keys 获取更多额度
              </a-button>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 领取记录 -->
    <a-card title="领取记录" class="mb-6">
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
        :data-source="claimRecords"
        :loading="loading"
        :pagination="paginationConfig"
        @change="handleTableChange"
        :scroll="{ x: 600 }"
        row-key="id"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'index'">
            {{ (pagination.current - 1) * pagination.pageSize + index + 1 }}
          </template>

          <template v-if="column.key === 'quota_added'">
            <a-tag color="success" class="font-medium">
              +{{ record.quota_added }}
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

          <template v-if="column.key === 'username'">
            <a-tag>{{ record.username }}</a-tag>
          </template>
        </template>

        <template #emptyText>
          <a-empty description="暂无领取记录">
            <template #image>
              <InboxOutlined class="text-6xl text-gray-300" />
            </template>
          </a-empty>
        </template>
      </a-table>
    </a-card>

    <!-- 领取规则说明 -->
    <a-card title="领取规则说明">
      <div class="space-y-3 text-sm text-gray-600">
        <div class="flex items-start space-x-2">
          <CheckCircleOutlined class="text-green-500 mt-0.5" />
          <span>每个账号每天可以领取一次额度</span>
        </div>
        <div class="flex items-start space-x-2">
          <CheckCircleOutlined class="text-green-500 mt-0.5" />
          <span>领取时间以系统时间为准，每日 00:00 重置</span>
        </div>
        <div class="flex items-start space-x-2">
          <CheckCircleOutlined class="text-green-500 mt-0.5" />
          <span>领取的额度将自动添加到账户余额中</span>
        </div>
        <div class="flex items-start space-x-2">
          <CheckCircleOutlined class="text-green-500 mt-0.5" />
          <span>必须先绑定 KYX 账号才能领取额度</span>
        </div>
        <div class="flex items-start space-x-2">
          <InfoCircleOutlined class="text-blue-500 mt-0.5" />
          <span>如需更多额度，可以通过投喂 API Keys 获取</span>
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import {
  GiftOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  DashboardOutlined,
  HeartOutlined,
  InboxOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { ClaimRecord } from '@/types'
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
const claiming = ref(false)
const claimQuota = ref(0)

// ==================== Computed ====================
const isBound = computed(() => userStore.isBound)
const canClaimToday = computed(() => userStore.canClaimToday)
const currentBalance = computed(() => userStore.currentBalance)
const lastClaimDate = computed(() => userStore.lastClaimDate)
const claimRecords = computed(() => userStore.claimRecords)
const pagination = computed(() => userStore.claimPagination)
const userStats = computed(() => userStore.userStats)

const statusMessage = computed(() => {
  if (!isBound.value) {
    return '请先绑定 KYX 账号'
  }
  if (canClaimToday.value) {
    return `今日可领取 ${claimQuota.value} 额度`
  }
  return '明天再来领取吧'
})

const nextClaimTime = computed(() => {
  if (canClaimToday.value) return null
  const tomorrow = dayjs().add(1, 'day').startOf('day')
  return tomorrow.format('YYYY-MM-DD 00:00:00')
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
    title: '用户名',
    dataIndex: 'username',
    key: 'username',
    width: 150
  },
  {
    title: '领取额度',
    dataIndex: 'quota_added',
    key: 'quota_added',
    width: 120,
    align: 'center'
  },
  {
    title: '领取时间',
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
 * 获取领取进度百分比
 */
const getClaimProgress = (): number => {
  const count = userStats.value?.claim_count || 0
  return Math.min((count / 30) * 100, 100)
}

/**
 * 获取额度进度百分比
 */
const getQuotaProgress = (): number => {
  const quota = userStats.value?.claim_quota || 0
  return Math.min((quota / 10000) * 100, 100)
}

/**
 * 处理领取
 */
const handleClaim = async () => {
  if (!isBound.value) {
    message.warning('请先绑定 KYX 账号')
    return
  }

  if (!canClaimToday.value) {
    message.info('今日已领取，明天再来吧')
    return
  }

  try {
    claiming.value = true

    const success = await userStore.claimQuota()

    if (success) {
      message.success({
        content: `领取成功！获得 ${claimQuota.value} 额度`,
        duration: 3
      })

      // 刷新记录
      await refreshRecords()
    }
  } catch (error: any) {
    console.error('Claim failed:', error)
    message.error(error.message || '领取失败，请重试')
  } finally {
    claiming.value = false
  }
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
      userStore.fetchClaimRecords({
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
    userStore.updateClaimPagination(paginationObj.current, paginationObj.pageSize)
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
 * 前往投喂页面
 */
const goToDonate = () => {
  router.push('/user/donate')
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
      userStore.fetchClaimRecords({ page: 1, page_size: 10 })
    ])

    // 获取系统配置的领取额度（这里暂时硬编码，实际应该从配置获取）
    claimQuota.value = 100 // TODO: 从管理员配置获取
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
  appStore.setPageTitle('领取额度')

  // 加载数据
  await loadData()
})
</script>

<style scoped>
.claim-page {
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

/* 领取卡片 */
.claim-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-radius: 8px;
  height: 100%;
}

/* 状态图标 */
.status-icon {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  animation: iconPulse 2s ease-in-out infinite;
}

.status-icon.success {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
  color: white;
  box-shadow: 0 8px 24px rgba(16, 185, 129, 0.3);
}

.status-icon.disabled {
  background: linear-gradient(135deg, #9ca3af 0%, #d1d5db 100%);
  color: white;
  box-shadow: 0 8px 24px rgba(156, 163, 175, 0.2);
}

@keyframes iconPulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

/* 领取按钮 */
.claim-button {
  min-width: 180px;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 24px;
  transition: all 0.3s ease;
}

.claim-button:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(59, 130, 246, 0.3);
}

/* 统计项 */
.stat-item {
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.stat-item:hover {
  background: #f3f4f6;
  transform: translateX(4px);
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

/* 进度条样式 */
:deep(.ant-progress-line) {
  height: 8px;
  border-radius: 4px;
}

:deep(.ant-progress-inner) {
  background: #e5e7eb;
}

/* 描述列表样式 */
:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  color: #6b7280;
}

:deep(.ant-descriptions-item-content) {
  color: #111827;
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

/* 响应式调整 */
@media (max-width: 1024px) {
  .status-icon {
    width: 100px;
    height: 100px;
  }

  .status-icon .anticon {
    font-size: 48px !important;
  }

  .claim-button {
    min-width: 150px;
    height: 44px;
    font-size: 15px;
  }
}

@media (max-width: 768px) {
  .status-icon {
    width: 80px;
    height: 80px;
  }

  .status-icon .anticon {
    font-size: 36px !important;
  }

  .claim-button {
    width: 100%;
    min-width: auto;
  }

  :deep(.ant-table) {
    font-size: 13px;
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
</style>
