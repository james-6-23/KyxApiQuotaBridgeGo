<template>
  <div class="donates-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        投喂记录管理
      </h2>
      <p class="text-gray-600">
        查看和管理所有用户的投喂 Keys 记录
      </p>
    </div>

    <!-- 搜索和筛选 -->
    <a-card class="mb-6">
      <a-row :gutter="16">
        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-input
            v-model:value="searchKeyword"
            placeholder="搜索用户名或 ID"
            allow-clear
            @pressEnter="handleSearch"
          >
            <template #prefix>
              <SearchOutlined class="text-gray-400" />
            </template>
          </a-input>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-range-picker
            v-model:value="dateRange"
            :placeholder="['开始日期', '结束日期']"
            class="w-full"
            @change="handleSearch"
          />
        </a-col>

        <a-col :xs="24" :sm="12" :md="4" :lg="4">
          <a-button block @click="handleSearch">
            <template #icon>
              <SearchOutlined />
            </template>
            搜索
          </a-button>
        </a-col>

        <a-col :xs="24" :sm="12" :md="4" :lg="4">
          <a-button block @click="handleReset">
            <template #icon>
              <ReloadOutlined />
            </template>
            重置
          </a-button>
        </a-col>
      </a-row>
    </a-card>

    <!-- 统计信息 -->
    <a-row :gutter="16" class="mb-6">
      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="总投喂次数"
            :value="totalDonations"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <HeartOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="总投喂 Keys"
            :value="totalKeys"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <KeyOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="今日投喂"
            :value="todayDonations"
            :value-style="{ color: '#faad14' }"
          >
            <template #prefix>
              <ClockCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="活跃用户"
            :value="activeUsers"
            :value-style="{ color: '#722ed1' }"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 投喂记录表格 -->
    <a-card title="投喂记录">
      <template #extra>
        <a-space>
          <a-tooltip title="导出数据">
            <a-button type="text" size="small" @click="handleExport">
              <template #icon>
                <ExportOutlined />
              </template>
            </a-button>
          </a-tooltip>
          <a-tooltip title="刷新">
            <a-button
              type="text"
              size="small"
              :loading="loading"
              @click="refreshData"
            >
              <template #icon>
                <ReloadOutlined />
              </template>
            </a-button>
          </a-tooltip>
        </a-space>
      </template>

      <a-table
        :columns="columns"
        :data-source="donationRecords"
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

          <template v-if="column.key === 'user'">
            <div class="space-y-1">
              <div class="font-medium">{{ record.username || '匿名用户' }}</div>
              <div class="text-xs text-gray-500">ID: {{ record.user_id || 'N/A' }}</div>
            </div>
          </template>

          <template v-if="column.key === 'key_count'">
            <a-tag color="success" class="font-medium">
              {{ record.key_count }} 个
            </a-tag>
          </template>

          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </a-tag>
          </template>

          <template v-if="column.key === 'timestamp'">
            <div class="space-y-1">
              <div class="text-sm">{{ formatDate(record.timestamp) }}</div>
              <div class="text-xs text-gray-500">{{ formatRelativeTime(record.timestamp) }}</div>
            </div>
          </template>

          <template v-if="column.key === 'actions'">
            <a-space>
              <a-tooltip title="查看详情">
                <a-button
                  type="link"
                  size="small"
                  @click="showDetail(record)"
                >
                  详情
                </a-button>
              </a-tooltip>
              <a-popconfirm
                title="确定要删除这条记录吗？"
                ok-text="确定"
                cancel-text="取消"
                @confirm="handleDelete(record.id)"
              >
                <a-button
                  type="link"
                  size="small"
                  danger
                >
                  删除
                </a-button>
              </a-popconfirm>
            </a-space>
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

    <!-- 详情对话框 -->
    <a-modal
      v-model:open="detailVisible"
      title="投喂记录详情"
      width="600px"
      :footer="null"
    >
      <div v-if="selectedRecord" class="space-y-4">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="记录 ID">
            {{ selectedRecord.id }}
          </a-descriptions-item>
          <a-descriptions-item label="用户名">
            {{ selectedRecord.username || '匿名用户' }}
          </a-descriptions-item>
          <a-descriptions-item label="用户 ID">
            {{ selectedRecord.user_id || 'N/A' }}
          </a-descriptions-item>
          <a-descriptions-item label="投喂 Keys">
            <a-tag color="success" class="font-medium">
              {{ selectedRecord.key_count }} 个
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="getStatusColor(selectedRecord.status)">
              {{ getStatusText(selectedRecord.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="投喂时间">
            {{ formatDate(selectedRecord.timestamp) }}
          </a-descriptions-item>
          <a-descriptions-item v-if="selectedRecord.remark" label="备注">
            {{ selectedRecord.remark }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import type { Dayjs } from 'dayjs'
import {
  SearchOutlined,
  ReloadOutlined,
  HeartOutlined,
  KeyOutlined,
  ClockCircleOutlined,
  UserOutlined,
  ExportOutlined,
  InboxOutlined
} from '@ant-design/icons-vue'
import { useAppStore } from '@/stores/app'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Types ====================
interface DonationRecord {
  id?: number
  user_id?: number
  username?: string
  key_count: number
  status: string
  timestamp: string
  remark?: string
  created_at?: string
}

interface Pagination {
  current: number
  pageSize: number
  total: number
}

// ==================== Composables ====================
const appStore = useAppStore()

// ==================== State ====================
const loading = ref(false)
const searchKeyword = ref('')
const dateRange = ref<[Dayjs, Dayjs]>()
const detailVisible = ref(false)
const selectedRecord = ref<DonationRecord | null>(null)
const donationRecords = ref<DonationRecord[]>([])
const pagination = ref<Pagination>({
  current: 1,
  pageSize: 10,
  total: 0
})

// ==================== Computed ====================
const totalDonations = computed(() => pagination.value.total)
const totalKeys = computed(() => {
  return donationRecords.value.reduce((sum, record) => sum + record.key_count, 0)
})
const todayDonations = computed(() => {
  const today = dayjs().startOf('day')
  return donationRecords.value.filter(record =>
    dayjs(record.timestamp).isAfter(today)
  ).length
})
const activeUsers = computed(() => {
  const uniqueUsers = new Set(donationRecords.value.map(r => r.user_id))
  return uniqueUsers.size
})

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
    title: '用户',
    key: 'user',
    width: 180
  },
  {
    title: 'Keys 数量',
    key: 'key_count',
    width: 120,
    align: 'center',
    sorter: (a, b) => (a as DonationRecord).key_count - (b as DonationRecord).key_count
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    align: 'center'
  },
  {
    title: '投喂时间',
    key: 'timestamp',
    width: 200,
    sorter: (a, b) => dayjs((a as DonationRecord).timestamp).unix() - dayjs((b as DonationRecord).timestamp).unix()
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    align: 'center',
    fixed: 'right'
  }
]

// ==================== Methods ====================

/**
 * 获取状态颜色
 */
const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    'active': 'success',
    'pending': 'processing',
    'inactive': 'default',
    'expired': 'error'
  }
  return colorMap[status] || 'default'
}

/**
 * 获取状态文本
 */
const getStatusText = (status: string): string => {
  const textMap: Record<string, string> = {
    'active': '已激活',
    'pending': '处理中',
    'inactive': '未激活',
    'expired': '已过期'
  }
  return textMap[status] || status
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
 * 搜索
 */
const handleSearch = () => {
  loadData()
}

/**
 * 重置
 */
const handleReset = () => {
  searchKeyword.value = ''
  dateRange.value = undefined
  loadData()
}

/**
 * 导出数据
 */
const handleExport = () => {
  message.info('导出功能开发中...')
}

/**
 * 显示详情
 */
const showDetail = (record: DonationRecord) => {
  selectedRecord.value = record
  detailVisible.value = true
}

/**
 * 删除记录
 */
const handleDelete = async (id?: number) => {
  if (!id) {
    message.error('记录 ID 不存在')
    return
  }

  try {
    // TODO: 调用删除 API
    message.success('删除成功')
    await loadData()
  } catch (error) {
    console.error('Delete donation record failed:', error)
    message.error('删除失败')
  }
}

/**
 * 刷新数据
 */
const refreshData = async () => {
  await loadData()
  message.success('数据已刷新')
}

/**
 * 处理表格变化
 */
const handleTableChange: TableProps['onChange'] = (paginationObj) => {
  if (paginationObj.current && paginationObj.pageSize) {
    pagination.value.current = paginationObj.current
    pagination.value.pageSize = paginationObj.pageSize
    loadData()
  }
}

/**
 * 加载数据
 */
const loadData = async () => {
  loading.value = true
  try {
    // TODO: 从 API 加载数据
    // 暂时使用模拟数据
    donationRecords.value = []
    pagination.value.total = 0
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
  appStore.setPageTitle('投喂记录')

  // 加载数据
  await loadData()
})
</script>

<style scoped>
.donates-page {
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

/* 统计卡片样式 */
:deep(.ant-statistic-title) {
  font-size: 13px;
  color: #6b7280;
}

:deep(.ant-statistic-content) {
  font-size: 20px;
  font-weight: 600;
}

/* 标签样式 */
:deep(.ant-tag) {
  border-radius: 4px;
  font-weight: 500;
}

/* 卡片标题 */
:deep(.ant-card-head-title) {
  font-weight: 600;
  color: #111827;
}

/* 日期选择器 */
:deep(.ant-picker) {
  width: 100%;
}

/* 响应式调整 */
@media (max-width: 768px) {
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
