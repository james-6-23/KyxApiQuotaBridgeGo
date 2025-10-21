<template>
  <div class="admin-dashboard">
    <!-- 欢迎信息 -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-800 mb-2">
        管理员仪表板
      </h1>
      <p class="text-gray-600">
        {{ greetingMessage }}
      </p>
    </div>

    <!-- 系统健康状态 -->
    <a-alert
      v-if="systemHealth === 'warning'"
      message="系统配置不完整"
      description="部分系统配置未完成，可能影响功能正常使用，请前往系统配置页面完善。"
      type="warning"
      show-icon
      closable
      class="mb-6"
    >
      <template #action>
        <a-button type="primary" size="small" @click="goToConfig">
          前往配置
        </a-button>
      </template>
    </a-alert>

    <!-- 统计卡片 -->
    <a-row :gutter="[16, 16]" class="mb-6">
      <!-- 总用户数 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="statsLoading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">总用户数</p>
              <h3 class="text-2xl font-bold text-blue-600">
                {{ formatNumber(totalUsers) }}
              </h3>
            </div>
            <div class="stat-icon bg-blue-100">
              <UserOutlined class="text-3xl text-blue-600" />
            </div>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-100">
            <a-button type="link" size="small" @click="goToUsers" class="p-0">
              查看详情 <RightOutlined />
            </a-button>
          </div>
        </a-card>
      </a-col>

      <!-- 总领取次数 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="statsLoading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">总领取次数</p>
              <h3 class="text-2xl font-bold text-green-600">
                {{ formatNumber(totalClaims) }}
              </h3>
              <p class="text-xs text-gray-500 mt-1">
                今日: {{ todayClaims }}
              </p>
            </div>
            <div class="stat-icon bg-green-100">
              <GiftOutlined class="text-3xl text-green-600" />
            </div>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-100">
            <a-button type="link" size="small" @click="goToClaims" class="p-0">
              查看记录 <RightOutlined />
            </a-button>
          </div>
        </a-card>
      </a-col>

      <!-- 总投喂次数 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="statsLoading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">总投喂次数</p>
              <h3 class="text-2xl font-bold text-purple-600">
                {{ formatNumber(totalDonates) }}
              </h3>
              <p class="text-xs text-gray-500 mt-1">
                今日: {{ todayDonates }}
              </p>
            </div>
            <div class="stat-icon bg-purple-100">
              <HeartOutlined class="text-3xl text-purple-600" />
            </div>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-100">
            <a-button type="link" size="small" @click="goToDonates" class="p-0">
              查看记录 <RightOutlined />
            </a-button>
          </div>
        </a-card>
      </a-col>

      <!-- 总 Keys 数量 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="statsLoading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">总 Keys 数量</p>
              <h3 class="text-2xl font-bold text-orange-600">
                {{ formatNumber(totalKeys) }}
              </h3>
            </div>
            <div class="stat-icon bg-orange-100">
              <KeyOutlined class="text-3xl text-orange-600" />
            </div>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-100">
            <a-button type="link" size="small" @click="goToKeys" class="p-0">
              管理 Keys <RightOutlined />
            </a-button>
          </div>
        </a-card>
      </a-col>

      <!-- 总配额分配 -->
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card class="stat-card h-full" :loading="statsLoading">
          <div class="flex items-center justify-between mb-3">
            <div>
              <p class="text-gray-600 text-sm mb-1">总配额分配</p>
              <h3 class="text-3xl font-bold text-indigo-600">
                {{ formatNumber(totalQuotaDistributed) }}
              </h3>
            </div>
            <div class="stat-icon bg-indigo-100">
              <DatabaseOutlined class="text-3xl text-indigo-600" />
            </div>
          </div>
          <a-progress
            :percent="getQuotaProgress()"
            :show-info="false"
            stroke-color="#6366f1"
          />
        </a-card>
      </a-col>

      <!-- 系统配置状态 -->
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card class="stat-card h-full" :loading="configLoading">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-gray-700 font-medium">系统配置状态</span>
              <a-tag :color="systemHealth === 'healthy' ? 'success' : 'warning'">
                {{ systemHealth === 'healthy' ? '正常' : '不完整' }}
              </a-tag>
            </div>
            <div class="space-y-2">
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
                <span class="text-gray-600">每日领取额度</span>
                <span class="text-blue-600 font-medium">{{ claimQuota }}</span>
              </div>
            </div>
            <a-button type="primary" ghost block @click="goToConfig">
              <template #icon>
                <SettingOutlined />
              </template>
              系统配置
            </a-button>
          </div>
        </a-card>
      </a-col>

      <!-- 快捷操作 -->
      <a-col :xs="24" :sm="12" :lg="8">
        <a-card title="快捷操作" class="h-full">
          <div class="space-y-2">
            <a-button block @click="goToKeys">
              <template #icon>
                <KeyOutlined />
              </template>
              Keys 管理
            </a-button>
            <a-button block @click="goToUsers">
              <template #icon>
                <UserOutlined />
              </template>
              用户管理
            </a-button>
            <a-button block @click="refreshAllData" :loading="loading">
              <template #icon>
                <ReloadOutlined />
              </template>
              刷新数据
            </a-button>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 数据趋势 -->
    <a-row :gutter="16" class="mb-6">
      <a-col :xs="24" :lg="12">
        <a-card title="领取趋势" class="mb-6 lg:mb-0">
          <template #extra>
            <a-radio-group v-model:value="claimChartPeriod" size="small">
              <a-radio-button value="7d">7天</a-radio-button>
              <a-radio-button value="30d">30天</a-radio-button>
            </a-radio-group>
          </template>
          <div class="h-64 flex items-center justify-center text-gray-400">
            <div class="text-center">
              <BarChartOutlined class="text-6xl mb-2" />
              <p>图表功能开发中...</p>
            </div>
          </div>
        </a-card>
      </a-col>

      <a-col :xs="24" :lg="12">
        <a-card title="投喂趋势">
          <template #extra>
            <a-radio-group v-model:value="donateChartPeriod" size="small">
              <a-radio-button value="7d">7天</a-radio-button>
              <a-radio-button value="30d">30天</a-radio-button>
            </a-radio-group>
          </template>
          <div class="h-64 flex items-center justify-center text-gray-400">
            <div class="text-center">
              <LineChartOutlined class="text-6xl mb-2" />
              <p>图表功能开发中...</p>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 最近活动 -->
    <a-row :gutter="16">
      <!-- 最近领取 -->
      <a-col :xs="24" :lg="12">
        <a-card title="最近领取" class="mb-6 lg:mb-0">
          <template #extra>
            <a-button type="link" size="small" @click="goToClaims">
              查看全部
            </a-button>
          </template>
          <a-list
            :loading="loading"
            :data-source="recentClaims"
            :locale="{ emptyText: '暂无领取记录' }"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <span class="font-medium">{{ item.username }}</span>
                    <a-tag color="success" class="ml-2">
                      +{{ item.quota_added }}
                    </a-tag>
                  </template>
                  <template #description>
                    <div class="space-y-1">
                      <div class="text-xs">Linux.do ID: {{ item.linux_do_id }}</div>
                      <div class="text-xs">{{ formatRelativeTime(item.timestamp) }}</div>
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>

      <!-- 最近投喂 -->
      <a-col :xs="24" :lg="12">
        <a-card title="最近投喂">
          <template #extra>
            <a-button type="link" size="small" @click="goToDonates">
              查看全部
            </a-button>
          </template>
          <a-list
            :loading="loading"
            :data-source="recentDonates"
            :locale="{ emptyText: '暂无投喂记录' }"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <span class="font-medium">{{ item.username }}</span>
                    <a-tag color="blue" class="ml-2">
                      {{ item.keys_count }} Keys
                    </a-tag>
                  </template>
                  <template #description>
                    <div class="space-y-1">
                      <div class="text-xs">
                        有效: {{ item.valid_keys_count || 0 }} /
                        无效: {{ item.invalid_keys_count || 0 }} /
                        额度: +{{ item.total_quota_added }}
                      </div>
                      <div class="text-xs">{{ formatRelativeTime(item.timestamp) }}</div>
                    </div>
                  </template>
                </a-list-item-meta>
                <template #actions>
                  <a-tag v-if="item.push_status === 'success'" color="success" class="m-0">
                    成功
                  </a-tag>
                  <a-tag v-else-if="item.push_status === 'failed'" color="error" class="m-0">
                    失败
                  </a-tag>
                  <a-tag v-else color="processing" class="m-0">
                    处理中
                  </a-tag>
                </template>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  UserOutlined,
  GiftOutlined,
  HeartOutlined,
  KeyOutlined,
  DatabaseOutlined,
  SettingOutlined,
  ReloadOutlined,
  RightOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  BarChartOutlined,
  LineChartOutlined
} from '@ant-design/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { useAppStore } from '@/stores/app'
import type { ClaimRecord, DonateRecord } from '@/types'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Composables ====================
const router = useRouter()
const adminStore = useAdminStore()
const appStore = useAppStore()

// ==================== State ====================
const loading = ref(false)
const statsLoading = ref(false)
const configLoading = ref(false)
const claimChartPeriod = ref('7d')
const donateChartPeriod = ref('30d')

// ==================== Computed ====================
const totalUsers = computed(() => adminStore.totalUsers)
const totalClaims = computed(() => adminStore.totalClaims)
const totalDonates = computed(() => adminStore.totalDonates)
const totalKeys = computed(() => adminStore.totalKeys)
const totalQuotaDistributed = computed(() => adminStore.totalQuotaDistributed)
const todayClaims = computed(() => adminStore.todayClaims)
const todayDonates = computed(() => adminStore.todayDonates)
const systemHealth = computed(() => adminStore.systemHealth)
const isSessionConfigured = computed(() => adminStore.isSessionConfigured)
const isKeysApiConfigured = computed(() => adminStore.isKeysApiConfigured)
const claimQuota = computed(() => adminStore.claimQuota)

// 最近的领取记录（取前5条）
const recentClaims = computed(() => {
  return adminStore.claimRecords.slice(0, 5)
})

// 最近的投喂记录（取前5条）
const recentDonates = computed(() => {
  return adminStore.donateRecords.slice(0, 5)
})

// 问候语
const greetingMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) {
    return '深夜了，注意休息 🌙'
  } else if (hour < 9) {
    return '早上好！新的一天开始了 ☀️'
  } else if (hour < 12) {
    return '上午好！'
  } else if (hour < 14) {
    return '中午好！'
  } else if (hour < 18) {
    return '下午好！'
  } else if (hour < 22) {
    return '晚上好！'
  } else {
    return '夜深了，早点休息 🌙'
  }
})

// ==================== Methods ====================

/**
 * 格式化数字
 */
const formatNumber = (num: number): string => {
  return num.toLocaleString('zh-CN')
}

/**
 * 格式化相对时间
 */
const formatRelativeTime = (date: string): string => {
  return dayjs(date).fromNow()
}

/**
 * 获取配额进度百分比
 */
const getQuotaProgress = (): number => {
  const quota = totalQuotaDistributed.value
  return Math.min((quota / 100000) * 100, 100)
}

/**
 * 前往配置页面
 */
const goToConfig = () => {
  router.push('/admin/config')
}

/**
 * 前往 Keys 管理
 */
const goToKeys = () => {
  router.push('/admin/keys')
}

/**
 * 前往用户管理
 */
const goToUsers = () => {
  router.push('/admin/users')
}

/**
 * 前往领取记录
 */
const goToClaims = () => {
  router.push('/admin/claims')
}

/**
 * 前往投喂记录
 */
const goToDonates = () => {
  router.push('/admin/donates')
}

/**
 * 刷新所有数据
 */
const refreshAllData = async () => {
  loading.value = true
  try {
    await adminStore.refreshDashboardData()
    message.success('数据已刷新')
  } catch (error) {
    console.error('Refresh data failed:', error)
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
}

/**
 * 加载数据
 */
const loadData = async () => {
  loading.value = true
  statsLoading.value = true
  configLoading.value = true

  try {
    // 并行加载所有数据
    await Promise.all([
      adminStore.fetchStats(),
      adminStore.fetchConfig(),
      adminStore.fetchClaimRecords({ page: 1, page_size: 5 }),
      adminStore.fetchDonateRecords({ page: 1, page_size: 5 })
    ])
  } catch (error) {
    console.error('Load data failed:', error)
    message.error('加载数据失败')
  } finally {
    loading.value = false
    statsLoading.value = false
    configLoading.value = false
  }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('仪表板')

  // 加载数据
  await loadData()
})

// 暴露刷新方法给父组件
defineExpose({
  refreshAllData
})
</script>

<style scoped>
.admin-dashboard {
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

/* 统计卡片样式 */
.stat-card {
  transition: all 0.3s ease;
  cursor: pointer;
  height: 100%;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.stat-card:hover .stat-icon {
  transform: scale(1.1) rotate(5deg);
}

/* 列表项动画 */
:deep(.ant-list-item) {
  transition: all 0.2s ease;
}

:deep(.ant-list-item:hover) {
  background-color: #f9fafb;
  padding-left: 12px;
}

/* 进度条样式 */
:deep(.ant-progress-line) {
  height: 8px;
  border-radius: 4px;
}

:deep(.ant-progress-inner) {
  background: #e5e7eb;
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
  .stat-card h3 {
    font-size: 1.5rem;
  }

  .stat-icon {
    width: 50px;
    height: 50px;
  }

  .stat-icon .anticon {
    font-size: 24px !important;
  }
}

/* 加载状态优化 */
:deep(.ant-card-loading-content) {
  padding: 16px 0;
}
</style>
