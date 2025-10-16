<template>
  <Layout>
    <div class="max-w-7xl mx-auto px-6 py-8">
      <!-- 欢迎标题 -->
      <div class="mb-8 animate-fade-in">
        <h1 class="text-3xl font-bold text-grok-text mb-2">
          欢迎回来, {{ userStore.userInfo?.display_name }} 👋
        </h1>
        <p class="text-grok-text-secondary">
          这是您的配额仪表板
        </p>
      </div>

      <!-- 统计卡片 -->
      <div class="grid md:grid-cols-3 gap-6 mb-8 animate-slide-up">
        <!-- 可用额度 -->
        <div class="stat-card">
          <div class="stat-label">可用额度</div>
          <div class="stat-value">
            {{ formatQuota(userStore.userInfo?.quota || 0) }}
          </div>
          <div class="mt-4">
            <a-progress
              :percent="quotaPercentage"
              :stroke-color="{
                '0%': '#1d9bf0',
                '100%': '#7856ff',
              }"
              :show-info="false"
            />
            <div class="text-xs text-grok-text-tertiary mt-2">
              已使用 {{ formatQuota(userStore.userInfo?.used_quota || 0)}} /
              总计 {{ formatQuota(userStore.userInfo?.total || 0) }}
            </div>
          </div>
        </div>

        <!-- 今日领取 -->
        <div class="stat-card">
          <div class="stat-label">今日领取</div>
          <div class="stat-value">
            {{ userStore.userInfo?.claimed_today ? '已领取' : '未领取' }}
          </div>
          <div class="mt-4">
            <a-tag
              :color="userStore.userInfo?.can_claim ? 'success' : 'default'"
              class="tech-tag"
            >
              {{ userStore.userInfo?.can_claim ? '可以领取' : '明日再来' }}
            </a-tag>
          </div>
        </div>

        <!-- 累计贡献 -->
        <div class="stat-card">
          <div class="stat-label">累计贡献</div>
          <div class="stat-value text-grok-success">
            {{ donateCount }}
          </div>
          <div class="text-xs text-grok-text-tertiary mt-4">
            感谢您的贡献！
          </div>
        </div>
      </div>

      <!-- 功能标签页 -->
      <div class="grok-card p-6 animate-slide-up animation-delay-200">
        <a-tabs v-model:activeKey="activeTab">
          <!-- 领取额度 -->
          <a-tab-pane key="claim" tab="领取额度">
            <div class="space-y-6">
              <a-button
                type="primary"
                size="large"
                :disabled="!userStore.userInfo?.can_claim"
                :loading="claiming"
                @click="handleClaim"
                block
                class="tech-button h-14 text-lg"
              >
                {{ userStore.userInfo?.can_claim ? '领取今日额度' : '今日已领取' }}
              </a-button>

              <div class="grok-divider"></div>

              <div>
                <h3 class="text-lg font-semibold text-grok-text mb-4">领取记录</h3>
                <a-table
                  :dataSource="claimRecords"
                  :columns="claimColumns"
                  :pagination="{ pageSize: 10 }"
                  :loading="loadingClaim"
                  row-key="timestamp"
                />
              </div>
            </div>
          </a-tab-pane>

          <!-- 投喂 Keys -->
          <a-tab-pane key="donate" tab="投喂 Keys">
            <div class="space-y-6">
              <a-form @finish="handleDonate" layout="vertical">
                <a-form-item
                  label="ModelScope Keys"
                  name="keys"
                  :rules="[{ required: true, message: '请输入至少一个 Key' }]"
                >
                  <a-textarea
                    v-model:value="donateForm.keys"
                    :rows="8"
                    placeholder="请输入 ModelScope Keys，每行一个&#10;sk-xxx...&#10;sk-yyy...&#10;sk-zzz..."
                    class="tech-input font-mono"
                  />
                </a-form-item>

                <a-form-item>
                  <a-button
                    type="primary"
                    html-type="submit"
                    size="large"
                    :loading="donating"
                    block
                    class="tech-button h-14 text-lg"
                  >
                    提交 Keys
                  </a-button>
                </a-form-item>
              </a-form>

              <div class="grok-divider"></div>

              <div>
                <h3 class="text-lg font-semibold text-grok-text mb-4">投喂记录</h3>
                <a-table
                  :dataSource="donateRecords"
                  :columns="donateColumns"
                  :pagination="{ pageSize: 10 }"
                  :loading="loadingDonate"
                  row-key="timestamp"
                />
              </div>
            </div>
          </a-tab-pane>

          <!-- 测试 Key -->
          <a-tab-pane key="test" tab="测试 Key">
            <div class="max-w-2xl mx-auto space-y-6">
              <a-form @finish="handleTest" layout="vertical">
                <a-form-item
                  label="ModelScope Key"
                  name="key"
                  :rules="[{ required: true, message: '请输入 Key' }]"
                >
                  <a-input
                    v-model:value="testForm.key"
                    placeholder="sk-xxx..."
                    size="large"
                    class="tech-input font-mono"
                  />
                </a-form-item>

                <a-form-item>
                  <a-button
                    type="primary"
                    html-type="submit"
                    size="large"
                    :loading="testing"
                    block
                    class="tech-button"
                  >
                    测试 Key
                  </a-button>
                </a-form-item>
              </a-form>

              <a-alert
                v-if="testResult"
                :type="testResult.valid ? 'success' : 'error'"
                :message="testResult.valid ? 'Key 有效 ✓' : 'Key 无效 ✗'"
                :description="`Key: ${testResult.key}`"
                show-icon
                class="grok-card"
              />
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { useUserStore } from '@/stores/user'
import { claimApi, donateApi, testApi, userApi, type ClaimRecord, type DonateRecord, type KeyTestResponse } from '@/api'
import Layout from '@/components/Layout.vue'

const userStore = useUserStore()
const activeTab = ref('claim')

// 加载状态
const claiming = ref(false)
const donating = ref(false)
const testing = ref(false)
const loadingClaim = ref(false)
const loadingDonate = ref(false)

// 表单数据
const donateForm = ref({ keys: '' })
const testForm = ref({ key: '' })

// 记录数据
const claimRecords = ref<ClaimRecord[]>([])
const donateRecords = ref<DonateRecord[]>([])
const donateCount = ref(0)

// 测试结果
const testResult = ref<KeyTestResponse | null>(null)

// 计算额度百分比
const quotaPercentage = computed(() => {
  const total = userStore.userInfo?.total || 0
  const used = userStore.userInfo?.used_quota || 0
  if (total === 0) return 0
  return Math.round(((total - used) / total) * 100)
})

// 格式化额度显示
const formatQuota = (quota: number) => {
  return (quota / 50).toFixed(2) + ' ¥'
}

// 表格列定义
const claimColumns = [
  {
    title: '时间',
    dataIndex: 'date',
    key: 'date',
  },
  {
    title: '额度',
    dataIndex: 'quota_added',
    key: 'quota_added',
    customRender: ({ text }: any) => formatQuota(text),
  },
]

const donateColumns = [
  {
    title: '时间',
    dataIndex: 'timestamp',
    key: 'timestamp',
    customRender: ({ text }: any) => dayjs(text * 1000).format('YYYY-MM-DD HH:mm:ss'),
  },
  {
    title: 'Keys 数量',
    dataIndex: 'keys_count',
    key: 'keys_count',
  },
  {
    title: '增加额度',
    dataIndex: 'total_quota_added',
    key: 'total_quota_added',
    customRender: ({ text }: any) => formatQuota(text),
  },
  {
    title: '推送状态',
    dataIndex: 'push_status',
    key: 'push_status',
    customRender: ({ text }: any) => {
      return text === 'success' ? '✓ 成功' : '✗ 失败'
    },
  },
]

// 领取额度
const handleClaim = async () => {
  claiming.value = true
  try {
    const response = await claimApi.daily()
    if (response.success) {
      message.success(response.message || '领取成功！')
      await userStore.refreshUserInfo()
      await loadClaimRecords()
    }
  } catch (error) {
    // 错误已在拦截器中处理
  } finally {
    claiming.value = false
  }
}

// 投喂 Keys
const handleDonate = async () => {
  const keys = donateForm.value.keys
    .split('\n')
    .map((k) => k.trim())
    .filter((k) => k.length > 0)

  if (keys.length === 0) {
    message.error('请输入至少一个 Key')
    return
  }

  donating.value = true
  try {
    const response = await donateApi.validate({ keys })
    if (response.success) {
      message.success(response.message || '投喂成功！')
      donateForm.value.keys = ''
      await userStore.refreshUserInfo()
      await loadDonateRecords()
    }
  } catch (error) {
    // 错误已在拦截器中处理
  } finally {
    donating.value = false
  }
}

// 测试 Key
const handleTest = async () => {
  testing.value = true
  testResult.value = null
  try {
    const response = await testApi.testKey({ key: testForm.value.key })
    if (response.success && response.data) {
      testResult.value = response.data
    }
  } catch (error) {
    // 错误已在拦截器中处理
  } finally {
    testing.value = false
  }
}

// 加载领取记录
const loadClaimRecords = async () => {
  loadingClaim.value = true
  try {
    const response = await userApi.getClaimRecords()
    if (response.success && response.data) {
      claimRecords.value = response.data
    }
  } catch (error) {
    console.error('Failed to load claim records:', error)
  } finally {
    loadingClaim.value = false
  }
}

// 加载投喂记录
const loadDonateRecords = async () => {
  loadingDonate.value = true
  try {
    const response = await userApi.getDonateRecords()
    if (response.success && response.data) {
      donateRecords.value = response.data
      donateCount.value = response.data.length
    }
  } catch (error) {
    console.error('Failed to load donate records:', error)
  } finally {
    loadingDonate.value = false
  }
}

onMounted(() => {
  loadClaimRecords()
  loadDonateRecords()
})
</script>
