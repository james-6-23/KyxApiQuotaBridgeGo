<template>
  <div class="users-page">
    <!-- 页面头部 -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-gray-800 mb-2">
        用户管理
      </h2>
      <p class="text-gray-600">
        查看和管理系统所有用户信息
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

        <a-col :xs="24" :sm="12" :md="6" :lg="4">
          <a-select
            v-model:value="roleFilter"
            placeholder="角色筛选"
            allow-clear
            class="w-full"
            @change="handleSearch"
          >
            <a-select-option value="user">普通用户</a-select-option>
            <a-select-option value="admin">管理员</a-select-option>
          </a-select>
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
            title="总用户数"
            :value="totalUsers"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="管理员"
            :value="adminCount"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <CrownOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <a-col :xs="12" :sm="6">
        <a-card size="small">
          <a-statistic
            title="今日新增"
            :value="todayNewUsers"
            :value-style="{ color: '#faad14' }"
          >
            <template #prefix>
              <UserAddOutlined />
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
              <RiseOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- 用户表格 -->
    <a-card title="用户列表">
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
        :data-source="userList"
        :loading="loading"
        :pagination="paginationConfig"
        @change="handleTableChange"
        :scroll="{ x: 1000 }"
        row-key="id"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'index'">
            {{ (pagination.current - 1) * pagination.pageSize + index + 1 }}
          </template>

          <template v-if="column.key === 'user'">
            <div class="space-y-1">
              <div class="font-medium">{{ record.username }}</div>
              <div class="text-xs text-gray-500">ID: {{ record.id }}</div>
            </div>
          </template>

          <template v-if="column.key === 'role'">
            <a-tag :color="record.is_admin ? 'blue' : 'default'">
              {{ record.is_admin ? '管理员' : '普通用户' }}
            </a-tag>
          </template>

          <template v-if="column.key === 'quota'">
            <a-tag color="success" class="font-medium">
              {{ record.quota || 0 }}
            </a-tag>
          </template>

          <template v-if="column.key === 'status'">
            <a-badge
              :status="record.is_active ? 'success' : 'default'"
              :text="record.is_active ? '活跃' : '未激活'"
            />
          </template>

          <template v-if="column.key === 'last_login'">
            <div v-if="record.last_login" class="space-y-1">
              <div class="text-sm">{{ formatDate(record.last_login) }}</div>
              <div class="text-xs text-gray-500">{{ formatRelativeTime(record.last_login) }}</div>
            </div>
            <span v-else class="text-gray-400">从未登录</span>
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
              <a-dropdown>
                <a-button type="link" size="small">
                  更多
                  <DownOutlined />
                </a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item key="edit" @click="handleEdit(record)">
                      <EditOutlined />
                      编辑
                    </a-menu-item>
                    <a-menu-item
                      key="toggle-admin"
                      @click="toggleAdmin(record)"
                    >
                      <CrownOutlined />
                      {{ record.is_admin ? '取消管理员' : '设为管理员' }}
                    </a-menu-item>
                    <a-menu-divider />
                    <a-menu-item key="delete" danger @click="handleDelete(record.id)">
                      <DeleteOutlined />
                      删除
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </template>
        </template>

        <template #emptyText>
          <a-empty description="暂无用户数据">
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
      title="用户详情"
      width="600px"
      :footer="null"
    >
      <div v-if="selectedUser" class="space-y-4">
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="用户 ID">
            {{ selectedUser.id }}
          </a-descriptions-item>
          <a-descriptions-item label="用户名">
            {{ selectedUser.username }}
          </a-descriptions-item>
          <a-descriptions-item label="Linux.do ID">
            {{ selectedUser.linux_do_id || 'N/A' }}
          </a-descriptions-item>
          <a-descriptions-item label="角色">
            <a-tag :color="selectedUser.is_admin ? 'blue' : 'default'">
              {{ selectedUser.is_admin ? '管理员' : '普通用户' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="剩余额度">
            <a-tag color="success" class="font-medium">
              {{ selectedUser.quota || 0 }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-badge
              :status="selectedUser.is_active ? 'success' : 'default'"
              :text="selectedUser.is_active ? '活跃' : '未激活'"
            />
          </a-descriptions-item>
          <a-descriptions-item label="注册时间">
            {{ formatDate(selectedUser.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="最后登录">
            {{ selectedUser.last_login ? formatDate(selectedUser.last_login) : '从未登录' }}
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
import {
  SearchOutlined,
  ReloadOutlined,
  UserOutlined,
  CrownOutlined,
  UserAddOutlined,
  RiseOutlined,
  ExportOutlined,
  InboxOutlined,
  DownOutlined,
  EditOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import { useAppStore } from '@/stores/app'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Types ====================
interface User {
  id: number
  username: string
  linux_do_id?: string
  is_admin: boolean
  quota?: number
  is_active: boolean
  last_login?: string
  created_at: string
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
const roleFilter = ref<string>()
const detailVisible = ref(false)
const selectedUser = ref<User | null>(null)
const userList = ref<User[]>([])
const pagination = ref<Pagination>({
  current: 1,
  pageSize: 10,
  total: 0
})

// ==================== Computed ====================
const totalUsers = computed(() => pagination.value.total)
const adminCount = computed(() => {
  return userList.value.filter(u => u.is_admin).length
})
const todayNewUsers = computed(() => {
  const today = dayjs().startOf('day')
  return userList.value.filter(u =>
    dayjs(u.created_at).isAfter(today)
  ).length
})
const activeUsers = computed(() => {
  return userList.value.filter(u => u.is_active).length
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
    title: '角色',
    key: 'role',
    width: 100,
    align: 'center'
  },
  {
    title: '剩余额度',
    key: 'quota',
    width: 120,
    align: 'center',
    sorter: (a, b) => ((a as User).quota || 0) - ((b as User).quota || 0)
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    align: 'center'
  },
  {
    title: '最后登录',
    key: 'last_login',
    width: 200,
    sorter: (a, b) => {
      const aTime = (a as User).last_login
      const bTime = (b as User).last_login
      if (!aTime) return 1
      if (!bTime) return -1
      return dayjs(aTime).unix() - dayjs(bTime).unix()
    }
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
  roleFilter.value = undefined
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
const showDetail = (user: User) => {
  selectedUser.value = user
  detailVisible.value = true
}

/**
 * 编辑用户
 */
const handleEdit = (user: User) => {
  message.info(`编辑用户: ${user.username}`)
  // TODO: 实现编辑功能
}

/**
 * 切换管理员权限
 */
const toggleAdmin = async (user: User) => {
  try {
    // TODO: 调用 API
    message.success(`已${user.is_admin ? '取消' : '设置'}管理员权限`)
    await loadData()
  } catch (error) {
    console.error('Toggle admin failed:', error)
    message.error('操作失败')
  }
}

/**
 * 删除用户
 */
const handleDelete = async (_id: number) => {
  try {
    // TODO: 调用删除 API
    message.success('删除成功')
    await loadData()
  } catch (error) {
    console.error('Delete user failed:', error)
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
    userList.value = []
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
  appStore.setPageTitle('用户管理')

  // 加载数据
  await loadData()
})
</script>

<style scoped>
.users-page {
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
