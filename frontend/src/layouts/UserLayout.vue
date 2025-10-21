<template>
  <a-layout class="min-h-screen">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      :width="240"
      :collapsed-width="80"
      class="shadow-lg"
      :class="{ 'fixed left-0 top-0 bottom-0 z-50': isMobile }"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center justify-center border-b border-gray-700">
        <div v-if="!collapsed" class="text-white text-lg font-bold flex items-center space-x-2">
          <ApiOutlined class="text-2xl" />
          <span>KYX Quota</span>
        </div>
        <ApiOutlined v-else class="text-white text-2xl" />
      </div>

      <!-- 菜单 -->
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        class="border-r-0"
        @click="handleMenuClick"
      >
        <a-menu-item key="/user/dashboard">
          <template #icon>
            <DashboardOutlined />
          </template>
          <span>仪表板</span>
        </a-menu-item>

        <a-menu-item key="/user/bind">
          <template #icon>
            <LinkOutlined />
          </template>
          <span>绑定账号</span>
        </a-menu-item>

        <a-menu-item key="/user/claim">
          <template #icon>
            <GiftOutlined />
          </template>
          <span>领取额度</span>
        </a-menu-item>

        <a-menu-item key="/user/donate">
          <template #icon>
            <HeartOutlined />
          </template>
          <span>投喂 Keys</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobile && !collapsed"
      class="fixed inset-0 bg-black bg-opacity-50 z-40"
      @click="collapsed = true"
    ></div>

    <!-- 主内容区 -->
    <a-layout :class="{ 'ml-0': isMobile, 'ml-60': !isMobile && !collapsed, 'ml-20': !isMobile && collapsed }">
      <!-- 顶部导航栏 -->
      <a-layout-header class="bg-white shadow-sm px-6 flex items-center justify-between fixed top-0 right-0 z-30"
        :class="{ 'left-0': isMobile, 'left-60': !isMobile && !collapsed, 'left-20': !isMobile && collapsed }"
      >
        <!-- 左侧：折叠按钮和面包屑 -->
        <div class="flex items-center space-x-4">
          <a-button
            type="text"
            @click="toggleSidebar"
            class="text-gray-600 hover:text-blue-600"
          >
            <template #icon>
              <MenuUnfoldOutlined v-if="collapsed" />
              <MenuFoldOutlined v-else />
            </template>
          </a-button>

          <!-- 面包屑 -->
          <a-breadcrumb v-if="!isSmallScreen">
            <a-breadcrumb-item
              v-for="(breadcrumb, index) in breadcrumbs"
              :key="index"
            >
              <router-link v-if="breadcrumb.path" :to="breadcrumb.path">
                {{ breadcrumb.name }}
              </router-link>
              <span v-else>{{ breadcrumb.name }}</span>
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>

        <!-- 右侧：用户信息和操作 -->
        <div class="flex items-center space-x-4">
          <!-- 刷新按钮 -->
          <a-tooltip title="刷新数据">
            <a-button
              type="text"
              :loading="loading"
              @click="handleRefresh"
              class="text-gray-600 hover:text-blue-600"
            >
              <template #icon>
                <ReloadOutlined />
              </template>
            </a-button>
          </a-tooltip>

          <!-- 主题切换 -->
          <a-tooltip :title="isDarkMode ? '切换到浅色模式' : '切换到深色模式'">
            <a-button
              type="text"
              @click="toggleTheme"
              class="text-gray-600 hover:text-blue-600"
            >
              <template #icon>
                <BulbOutlined v-if="isDarkMode" />
                <BulbFilled v-else />
              </template>
            </a-button>
          </a-tooltip>

          <!-- 用户下拉菜单 -->
          <a-dropdown placement="bottomRight">
            <div class="flex items-center space-x-2 cursor-pointer hover:bg-gray-100 px-3 py-2 rounded-lg transition-colors">
              <a-avatar :size="32" class="bg-blue-500">
                <template #icon>
                  <UserOutlined />
                </template>
              </a-avatar>
              <span v-if="!isSmallScreen" class="text-gray-700 font-medium">{{ username }}</span>
              <DownOutlined class="text-gray-500 text-xs" />
            </div>

            <template #overlay>
              <a-menu>
                <a-menu-item key="profile" disabled>
                  <UserOutlined />
                  <span class="ml-2">{{ username }}</span>
                </a-menu-item>
                <a-menu-item key="linuxdo" disabled>
                  <IdcardOutlined />
                  <span class="ml-2">ID: {{ linuxDoId }}</span>
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  <span class="ml-2">退出登录</span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 内容区域 -->
      <a-layout-content class="mt-16 p-6 min-h-[calc(100vh-64px)]">
        <!-- 页面标题 -->
        <div v-if="pageTitle" class="mb-6">
          <h1 class="text-2xl font-bold text-gray-800">{{ pageTitle }}</h1>
        </div>

        <!-- 路由视图 -->
        <div class="bg-white rounded-lg shadow-sm p-6">
          <router-view v-slot="{ Component, route }">
            <transition name="fade" mode="out-in">
              <keep-alive :include="keepAliveComponents">
                <component :is="Component" :key="route.path" />
              </keep-alive>
            </transition>
          </router-view>
        </div>
      </a-layout-content>

      <!-- 页脚 -->
      <a-layout-footer class="text-center text-gray-600">
        <div class="space-y-2">
          <div>KYX API Quota Bridge © {{ currentYear }}</div>
          <div class="text-sm text-gray-500">
            Powered by Vue 3 + Ant Design Vue
          </div>
        </div>
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  LinkOutlined,
  GiftOutlined,
  HeartOutlined,
  UserOutlined,
  LogoutOutlined,
  DownOutlined,
  ReloadOutlined,
  BulbOutlined,
  BulbFilled,
  ApiOutlined,
  IdcardOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

// ==================== Stores ====================
const authStore = useAuthStore()
const userStore = useUserStore()
const appStore = useAppStore()
const router = useRouter()
const route = useRoute()

// ==================== State ====================
const collapsed = ref(false)
const selectedKeys = ref<string[]>([route.path])

// ==================== Computed ====================
const username = computed(() => authStore.username || '用户')
const linuxDoId = computed(() => authStore.linuxDoId || '')
const loading = computed(() => userStore.loading || appStore.loading)
const pageTitle = computed(() => appStore.pageTitle)
const breadcrumbs = computed(() => appStore.breadcrumbs)
const isDarkMode = computed(() => appStore.isDarkMode)
const isMobile = computed(() => appStore.isMobile)
const isSmallScreen = computed(() => appStore.isSmallScreen)
const currentYear = computed(() => new Date().getFullYear())

// 需要缓存的组件
const keepAliveComponents = ['UserDashboard', 'UserBind', 'UserClaim', 'UserDonate']

// ==================== Methods ====================

/**
 * 切换侧边栏
 */
const toggleSidebar = () => {
  collapsed.value = !collapsed.value
  appStore.setSidebarCollapsed(collapsed.value)
}

/**
 * 切换主题
 */
const toggleTheme = () => {
  appStore.toggleTheme()
}

/**
 * 菜单点击事件
 */
const handleMenuClick = ({ key }: { key: string }) => {
  router.push(key)

  // 在移动端点击菜单后自动收起侧边栏
  if (isMobile.value) {
    collapsed.value = true
  }
}

/**
 * 刷新数据
 */
const handleRefresh = async () => {
  try {
    await userStore.refreshAllData()
    message.success('数据已刷新')
  } catch (error) {
    message.error('刷新失败')
  }
}

/**
 * 退出登录
 */
const handleLogout = async () => {
  try {
    await authStore.logout()
    message.success('已退出登录')
    router.push('/user/login')
  } catch (error) {
    message.error('退出登录失败')
  }
}

// ==================== Watch ====================

// 监听路由变化，更新选中的菜单项
watch(
  () => route.path,
  (newPath) => {
    selectedKeys.value = [newPath]
  },
  { immediate: true }
)

// 监听侧边栏状态
watch(
  () => appStore.isSidebarCollapsed,
  (newValue) => {
    collapsed.value = newValue
  },
  { immediate: true }
)

// ==================== Lifecycle ====================

onMounted(async () => {
  // 初始化应用状态
  appStore.initApp()

  // 获取用户数据
  if (authStore.isAuthenticated) {
    await userStore.fetchUserQuota()
  }
})
</script>

<style scoped>
/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 侧边栏样式 */
:deep(.ant-layout-sider) {
  background: #001529;
}

:deep(.ant-menu-dark) {
  background: #001529;
}

:deep(.ant-menu-dark .ant-menu-item-selected) {
  background-color: #1890ff;
}

/* 头部固定时的内容区域调整 */
:deep(.ant-layout-header) {
  height: 64px;
  line-height: 64px;
  padding-left: 24px;
  padding-right: 24px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  :deep(.ant-layout-header) {
    padding-left: 16px;
    padding-right: 16px;
  }

  :deep(.ant-layout-content) {
    padding: 16px;
  }
}

/* 移动端侧边栏覆盖在内容上 */
@media (max-width: 768px) {
  :deep(.ant-layout-sider) {
    position: fixed !important;
    height: 100vh;
    left: 0;
    top: 0;
    z-index: 1001;
  }
}
</style>
