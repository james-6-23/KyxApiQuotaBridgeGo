<template>
  <div class="user-dashboard">
    <!-- 欢迎信息 -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-800 mb-2">
        欢迎回来，{{ username }}！
      </h1>
      <p class="text-gray-600">
        {{ greetingMessage }}
      </p>
    </div>

    <!-- 绑定状态提醒 -->
    <a-alert
      v-if="!isBound"
      message="尚未绑定 KYX 账号"
      description="请先绑定 KYX 账号，才能领取额度和使用相关功能。"
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

    <!-- 统计卡片 -->
    <a-row :gutter="[16, 16]" class="mb-6">
      <!-- 当前余额 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="loading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">当前余额</p>
              <h3 class="text-2xl font-bold text-blue-600">
                {{ formatNumber(currentBalance) }}
              </h3>
            </div>
            <div class="stat-icon bg-blue-100">
              <WalletOutlined class="text-3xl text-blue-600" />
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 总配额 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="loading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">总配额</p>
              <h3 class="text-2xl font-bold text-green-600">
                {{ formatNumber(currentQuota) }}
              </h3>
            </div>
            <div class="stat-icon bg-green-100">
              <DatabaseOutlined class="text-3xl text-green-600" />
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 已领取 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="loading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">累计领取</p>
              <h3 class="text-2xl font-bold text-purple-600">
                {{ formatNumber(totalClaimed) }}
              </h3>
            </div>
            <div class="stat-icon bg-purple-100">
              <GiftOutlined class="text-3xl text-purple-600" />
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 已投喂 -->
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" :loading="loading">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-gray-600 text-sm mb-1">累计投喂</p>
              <h3 class="text-2xl font-bold text-orange-600">
                {{ formatNumber(totalDonated) }}
              </h3>
            </div>
            <div class="stat-icon bg-orange-100">
              <HeartOutlined class="text-3xl text-orange-600" />
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 快捷操作 -->
    <a-row :gutter="16" class="mb-6">
      <a-col :xs="24" :md="12">
        <a-card title="每日领取" class="action-card">
          <template #extra>
            <a-tag v-if="canClaimToday" color="success">
              <CheckCircleOutlined /> 可领取
            </a-tag>
            <a-tag v-else color="default">
              <ClockCircleOutlined /> 已领取
            </a-tag>
          </template>
          <div class="space-y-4">
            <p class="text-gray-600">
              {{ canClaimToday ? '今日还未领取额度，立即领取吧！' : '今日已领取额度，明天再来吧！' }}
            </p>
            <div v-if="lastClaimDate" class="text-sm text-gray-500">
              上次领取：{{ formatDate(lastClaimDate) }}
            </div>
            <a-button
              type="primary"
              block
              size="large"
              :disabled="!canClaimToday || !isBound"
              @click="goToClaim"
            >
              <template #icon>
                <GiftOutlined />
              </template>
              {{ canClaimToday ? '立即领取' : '今日已领取' }}
            </a-button>
          </div>
        </a-card>
      </a-col>

      <a-col :xs="24" :md="12">
        <a-card title="投喂 Keys" class="action-card">
          <template #extra>
            <a-tag color="processing">
              <ThunderboltOutlined /> 获取额外额度
            </a-tag>
          </template>
          <div class="space-y-4">
            <p class="text-gray-600">
              通过投喂 API Keys 获取额外配额，每个有效 Key 可获得相应额度。
            </p>
            <a-button
              type="primary"
              block
              size="large"
              :disabled="!isBound"
              @click="goToDonate"
            >
              <template #icon>
                <HeartOutlined />
              </template>
              投喂 Keys
            </a-button>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 最近记录 -->
    <a-row :gutter="16">
      <!-- 最近领取记录 -->
      <a-col :xs="24" :lg="12">
        <a-card title="最近领取记录" class="mb-6">
          <template #extra>
            <a-button type="link" size="small" @click="goToClaim">
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
                    <span class="font-medium">领取额度</span>
                    <a-tag color="success" class="ml-2">
                      +{{ item.quota_added }}
                    </a-tag>
                  </template>
                  <template #description>
                    {{ formatDate(item.timestamp) }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>

      <!-- 最近投喂记录 -->
      <a-col :xs="24" :lg="12">
        <a-card title="最近投喂记录" class="mb-6">
          <template #extra>
            <a-button type="link" size="small" @click="goToDonate">
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
                    <span class="font-medium">投喂 {{ item.keys_count }} 个 Keys</span>
                    <a-tag v-if="item.push_status === 'success'" color="success" class="ml-2">
                      成功
                    </a-tag>
                    <a-tag v-else-if="item.push_status === 'failed'" color="error" class="ml-2">
                      失败
                    </a-tag>
                    <a-tag v-else color="processing" class="ml-2">
                      处理中
                    </a-tag>
                  </template>
                  <template #description>
                    <div class="space-y-1">
                      <div>
                        有效: {{ item.valid_keys_count || 0 }} /
                        无效: {{ item.invalid_keys_count || 0 }} /
                        额度: +{{ item.total_quota_added }}
                      </div>
                      <div>{{ formatDate(item.timestamp) }}</div>
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>
    </a-row>

    <!-- 账号信息 -->
    <a-card title="账号信息" class="mb-6">
      <a-descriptions :column="{ xs: 1, sm: 2, md: 3 }">
        <a-descriptions-item label="用户名">
          {{ username }}
        </a-descriptions-item>
        <a-descriptions-item label="Linux.do ID">
          {{ linuxDoId }}
        </a-descriptions-item>
        <a-descriptions-item label="KYX 账号">
          <span v-if="isBound" class="text-green-600">
            <CheckCircleOutlined /> {{ kyxUsername }}
          </span>
          <span v-else class="text-orange-600">
            <CloseCircleOutlined /> 未绑定
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="绑定状态">
          <a-tag v-if="isBound" color="success">已绑定</a-tag>
          <a-tag v-else color="warning">未绑定</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="今日可领取">
          <a-tag v-if="canClaimToday" color="success">是</a-tag>
          <a-tag v-else color="default">否</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="上次领取">
          {{ lastClaimDate ? formatDate(lastClaimDate) : '暂无记录' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  WalletOutlined,
  DatabaseOutlined,
  GiftOutlined,
  HeartOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import type { ClaimRecord, DonateRecord } from '@/types'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

// ==================== Composables ====================
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()
const appStore = useAppStore()

// ==================== State ====================
const loading = ref(false)

// ==================== Computed ====================
const username = computed(() => authStore.username || '用户')
const linuxDoId = computed(() => authStore.linuxDoId || '')
const isBound = computed(() => userStore.isBound)
const canClaimToday = computed(() => userStore.canClaimToday)
const currentBalance = computed(() => userStore.currentBalance)
const currentQuota = computed(() => userStore.currentQuota)
const totalClaimed = computed(() => userStore.totalClaimed)
const totalDonated = computed(() => userStore.totalDonated)
const kyxUsername = computed(() => userStore.kyxUsername)
const lastClaimDate = computed(() => userStore.lastClaimDate)

// 最近的领取记录（取前5条）
const recentClaims = computed(() => {
  return userStore.claimRecords.slice(0, 5)
})

// 最近的投喂记录（取前5条）
const recentDonates = computed(() => {
  return userStore.donateRecords.slice(0, 5)
})

// 问候语
const greetingMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) {
    return '夜深了，注意休息哦 🌙'
  } else if (hour < 9) {
    return '早上好！新的一天开始了 ☀️'
  } else if (hour < 12) {
    return '上午好！工作顺利 💪'
  } else if (hour < 14) {
    return '中午好！记得休息一下 🍜'
  } else if (hour < 18) {
    return '下午好！继续加油 🚀'
  } else if (hour < 22) {
    return '晚上好！辛苦一天了 🌆'
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
 * 格式化日期
 */
const formatDate = (date: string): string => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

/**
 * 前往绑定页面
 */
const goToBind = () => {
  router.push('/user/bind')
}

/**
 * 前往领取页面
 */
const goToClaim = () => {
  router.push('/user/claim')
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
    // 并行加载所有数据
    await Promise.all([
      userStore.fetchUserQuota(),
      userStore.fetchUserStats(),
      userStore.fetchClaimRecords({ page: 1, page_size: 5 }),
      userStore.fetchDonateRecords({ page: 1, page_size: 5 })
    ])
  } catch (error: any) {
    console.error('Load data failed:', error)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

/**
 * 刷新数据
 */
const refreshData = async () => {
  await loadData()
  message.success('数据已刷新')
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 设置页面标题
  appStore.setPageTitle('仪表板')

  // 加载数据
  await loadData()

  // 如果未绑定，提示用户
  if (!isBound.value) {
    message.warning('请先绑定 KYX 账号', 3)
  }
})

// 暴露刷新方法给父组件
defineExpose({
  refreshData
})
</script>

<style scoped>
.user-dashboard {
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

/* 操作卡片样式 */
.action-card {
  height: 100%;
  transition: all 0.3s ease;
}

.action-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

/* 列表项动画 */
:deep(.ant-list-item) {
  transition: all 0.2s ease;
}

:deep(.ant-list-item:hover) {
  background-color: #f9fafb;
  padding-left: 12px;
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

/* 标签样式 */
:deep(.ant-tag) {
  border-radius: 4px;
  font-weight: 500;
}

/* 描述列表样式 */
:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  color: #6b7280;
}

:deep(.ant-descriptions-item-content) {
  color: #111827;
}
</style>
