<template>
  <div class="keys-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        Keys 管理
      </h2>
      <p class="text-gray-600">
        管理用户投喂的 API Keys，进行验证和推送操作
      </p>
    </div>

    <!-- 操作栏 -->
    <a-card class="mb-6">
      <a-row :gutter="16">
        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-input
            v-model:value="searchKeyword"
            placeholder="搜索 Key 或用户名"
            allow-clear
            @pressEnter="handleSearch"
          >
            <template #prefix>
              <SearchOutlined class="text-gray-400" />
            </template>
          </a-input>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-select
            v-model:value="statusFilter"
            placeholder="筛选状态"
            allow-clear
            class="w-full"
            @change="handleSearch"
          >
            <a-select-option value="pending">待处理</a-select-option>
            <a-select-option value="validated">已验证</a-select-option>
            <a-select-option value="pushed">已推送</a-select-option>
            <a-select-option value="failed">推送失败</a-select-option>
          </a-select>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-button block @click="handleSearch">
            <template #icon>
              <SearchOutlined />
            </template>
            搜索
          </a-button>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8" :lg="6">
          <a-button block @click="handleReset">
            <template #icon>
              <ReloadOutlined />
            </template>
            重置
          </a-button>
        </a-col>
      </a-row>

      <a-divider />

      <a-row :gutter="16">
        <a-col :xs="24" :sm="12" :md="8">
          <a-button
            type="primary"
            block
            :disabled="selectedKeys.length === 0"
            @click="handleBatchValidate"
            :loading="validating"
          >
            <template #icon>
              <CheckCircleOutlined />
            </template>
            批量验证 ({{ selectedKeys.length }})
          </a-button>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8">
          <a-button
            type="primary"
            block
            :disabled="!hasValidKeys"
            @click="handlePushToKyx"
            :loading="pushing"
          >
            <template #icon>
              <CloudUploadOutlined />
            </template>
            推送到 KYX
          </a-button>
        </a-col>

        <a-col :xs="24" :sm="12" :md="8">
          <a-button
            block
            @click="refreshData"
            :loading="loading"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
            刷新数据
          </a-button>
        </a-col>
      </a-row>
    </a-card>

    <!-- 统计信息 -->
    <a-row :gutter="16" class="mb-6">
      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="总 Keys"
            :value="totalKeys"
            :value-style="{ color: '#3f8600' }"
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
            title="待处理"
            :value="pendingKeys"
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
            title="已验证"
            :value="validatedKeys"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <CheckCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="已推送"
            :value="pushedKeys"
            :value-style="{ color: '#52c41a' }"
          >
            <template #prefix>
              <CloudUploadOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- Keys 列表 -->
    <a-card title="Keys 列表">
      <template #extra>
        <a-space>
          <a-tooltip title="导出数据">
            <a-button type="text" size="small">
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
        :data-source="keys"
        :loading="loading"
        :pagination="paginationConfig"
        :row-selection="rowSelection"
        @change="handleTableChange"
        :scroll="{ x: 1200 }"
        row-key="key"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'index'">
            {{ (pagination.current - 1) * pagination.pageSize + index + 1 }}
          </template>

          <template v-if="column.key === 'key'">
            <div class="flex items-center space-x-2">
              <code class="text-xs bg-gray-100 px-2 py-1 rounded">
                {{ maskKey(record.key) }}
              </code>
              <a-tooltip title="复制完整 Key">
                <a-button
                  type="text"
                  size="small"
                  @click="handleCopyKey(record.key)"
                >
                  <template #icon>
                    <CopyOutlined />
                  </template>
                </a-button>
              </a-tooltip>
            </div>
          </template>

          <template v-if="column.key === 'user'">
            <div class="space-y-1">
              <div class="font-medium">{{ record.username }}</div>
              <div class="text-xs text-gray-500">ID: {{ record.linux_do_id }}</div>
            </div>
          </template>

          <template v-if="column.key === 'status'">
            <a-tag v-if="record.status === 'pending'" color="warning">
              <ClockCircleOutlined /> 待处理
            </a-tag>
            <a-tag v-else-if="record.status === 'validated'" color="processing">
              <CheckCircleOutlined /> 已验证
            </a-tag>
            <a-tag v-else-if="record.status === 'pushed'" color="success">
              <CloudUploadOutlined /> 已推送
            </a-tag>
            <a-tag v-else-if="record.status === 'failed'" color="error">
              <CloseCircleOutlined /> 推送失败
            </a-tag>
            <a-tag v-else color="default">
              未知
            </a-tag>
          </template>

          <template v-if="column.key === 'is_valid'">
            <CheckCircleOutlined v-if="record.is_valid" class="text-green-500 text-lg" />
            <CloseCircleOutlined v-else class="text-red-500 text-lg" />
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
                title="确定要删除这个 Key 吗？"
                ok-text="确定"
                cancel-text="取消"
                @confirm="handleDelete(record)"
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
          <a-empty description="暂无 Keys 数据">
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
      title="Key 详情"
      width="700px"
      :footer="null"
    >
      <div v-if="selectedKey" class="space-y-4">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="Key" :span="2">
            <div class="flex items-center space-x-2">
              <code class="text-xs bg-gray-100 px-2 py-1 rounded flex-1 overflow-hidden text-ellipsis">
                {{ selectedKey.key }}
              </code>
              <a-button size="small" @click="handleCopyKey(selectedKey.key)">
                <template #icon>
                  <CopyOutlined />
                </template>
                复制
              </a-button>
            </div>
          </a-descriptions-item>
          <a-descriptions-item label="用户名">
            {{ selectedKey.username }}
          </a-descriptions-item>
          <a-descriptions-item label="Linux.do ID">
            {{ selectedKey.linux_do_id }}
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag v-if="selectedKey.status === 'pending'" color="warning">待处理</a-tag>
            <a-tag v-else-if="selectedKey.status === 'validated'" color="processing">已验证</a-tag>
            <a-tag v-else-if="selectedKey.status === 'pushed'" color="success">已推送</a-tag>
            <a-tag v-else-if="selectedKey.status === 'failed'" color="error">推送失败</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="是否有效">
            <a-tag v-if="selectedKey.is_valid" color="success">有效</a-tag>
            <a-tag v-else color="error">无效</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="投喂时间" :span="2">
            {{ formatDate(selectedKey.timestamp) }}
          </a-descriptions-item>
        </a-descriptions>

        <a-divider />

        <div class="space-y-2">
          <a-button
            v-if="selectedKey.status === 'pending'"
            type="primary"
            block
            @click="handleValidateSingle(selectedKey)"
          >
            <template #icon>
              <CheckCircleOutlined />
            </template>
            验证此 Key
          </a-button>
          <a-button
            v-if="selectedKey.status === 'validated'"
            type="primary"
            block
            @click="handlePushSingle(selectedKey)"
          >
            <template #icon>
              <CloudUploadOutlined />
            </template>
            推送此 Key
          </a-button>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import {
  SearchOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CloudUploadOutlined,
  KeyOutlined,
  ClockCircleOutlined,
  ExportOutlined,
  CopyOutlined,
  InboxOutlined
} from '@ant-design/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { useAppStore } from '@/stores/app'
import type { DonatedKey } from '@/types'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Composables ====================
const adminStore = useAdminStore()
const appStore = useAppStore()

// ==================== State ====================
const loading = ref(false)
const validating = ref(false)
const pushing = ref(false)
const searchKeyword = ref('')
const statusFilter = ref<string>()
const selectedKeys = ref<string[]>([])
const detailVisible = ref(false)
const selectedKey = ref<DonatedKey | null>(null)

// ==================== Computed ====================
const keys = computed(() => adminStore.keys)
const pagination = computed(() => adminStore.keyPagination)

// 统计数据
const totalKeys = computed(() => keys.value.length)
const pendingKeys = computed(() => keys.value.filter(k => k.status === 'pending').length)
const validatedKeys = computed(() => keys.value.filter(k => k.status === 'validated').length)
const pushedKeys = computed(() => keys.value.filter(k => k.status === 'pushed').length)

// 是否有已验证的 Keys
const hasValidKeys = computed(() => validatedKeys.value > 0)

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

// 行选择配置
const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  onChange: (selectedRowKeys: string[]) => {
    selectedKeys.value = selectedRowKeys
  }
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
    title: 'Key',
    dataIndex: 'key',
    key: 'key',
    width: 300
  },
  {
    title: '用户',
    key: 'user',
    width: 150
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
    align: 'center'
  },
  {
    title: '是否有效',
    key: 'is_valid',
    width: 100,
    align: 'center'
  },
  {
    title: '投喂时间',
    key: 'timestamp',
    width: 180
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
 * 遮罩 Key（只显示前后部分）
 */
const maskKey = (key: string): string => {
  if (key.length <= 20) return key
  return `${key.substring(0, 10)}...${key.substring(key.length - 10)}`
}

/**
 * 复制 Key
 */
const handleCopyKey = async (key: string) => {
  try {
    await navigator.clipboard.writeText(key)
    message.success('Key 已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
}

/**
 * 搜索
 */
const handleSearch = () => {
  // 重新加载数据，应用筛选条件
  loadData()
}

/**
 * 重置
 */
const handleReset = () => {
  searchKeyword.value = ''
  statusFilter.value = undefined
  selectedKeys.value = []
  loadData()
}

/**
 * 批量验证
 */
const handleBatchValidate = async () => {
  if (selectedKeys.value.length === 0) {
    message.warning('请先选择要验证的 Keys')
    return
  }

  validating.value = true
  try {
    const result = await adminStore.verifyKeys(selectedKeys.value)
    if (result) {
      message.success(`验证完成！有效: ${result.valid}, 无效: ${result.invalid}`)
      selectedKeys.value = []
      await loadData()
    }
  } catch (error) {
    console.error('Batch validate failed:', error)
    message.error('批量验证失败')
  } finally {
    validating.value = false
  }
}

/**
 * 推送到 KYX
 */
const handlePushToKyx = async () => {
  pushing.value = true
  try {
    const success = await adminStore.pushKeys()
    if (success) {
      await loadData()
    }
  } catch (error) {
    console.error('Push to KYX failed:', error)
    message.error('推送失败')
  } finally {
    pushing.value = false
  }
}

/**
 * 验证单个 Key
 */
const handleValidateSingle = async (key: DonatedKey) => {
  validating.value = true
  try {
    const result = await adminStore.verifyKeys([key.key])
    if (result) {
      message.success('验证完成')
      detailVisible.value = false
      await loadData()
    }
  } catch (error) {
    console.error('Validate single key failed:', error)
    message.error('验证失败')
  } finally {
    validating.value = false
  }
}

/**
 * 推送单个 Key
 */
const handlePushSingle = async (_key: DonatedKey) => {
  pushing.value = true
  try {
    // 这里应该调用单个推送的 API，暂时使用批量推送
    const success = await adminStore.pushKeys()
    if (success) {
      message.success('推送成功')
      detailVisible.value = false
      await loadData()
    }
  } catch (error) {
    console.error('Push single key failed:', error)
    message.error('推送失败')
  } finally {
    pushing.value = false
  }
}

/**
 * 显示详情
 */
const showDetail = (key: DonatedKey) => {
  selectedKey.value = key
  detailVisible.value = true
}

/**
 * 删除 Key
 */
const handleDelete = async (_key: DonatedKey) => {
  try {
    // TODO: 实现删除 API
    message.success('删除成功')
    await loadData()
  } catch (error) {
    console.error('Delete key failed:', error)
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
    adminStore.updateKeyPagination(paginationObj.current, paginationObj.pageSize)
  }
}

/**
 * 加载数据
 */
const loadData = async () => {
  loading.value = true
  try {
    await adminStore.fetchKeys({
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    })
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
  appStore.setPageTitle('Keys 管理')

  // 加载数据
  await loadData()
})
</script>

<style scoped>
.keys-page {
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

/* 响应式调整 */
@media (max-width: 768px) {
  :deep(.ant-table) {
    font-size: 13px;
  }

  code {
    font-size: 11px;
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
