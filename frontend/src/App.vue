<template>
  <ConfigProvider :theme="antdTheme">
    <div class="app-container">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
  </ConfigProvider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ConfigProvider, theme } from 'ant-design-vue'
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

// Ant Design Vue 主题配置 - 根据当前主题动态切换
const antdTheme = computed(() => ({
  algorithm: themeStore.theme === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
  token: {
    colorPrimary: '#1d9bf0',
    colorSuccess: '#00ba7c',
    colorWarning: '#f9a825',
    colorError: '#f44336',
    colorInfo: '#7856ff',
    // 深色模式的背景色
    ...(themeStore.theme === 'dark' ? {
      colorBgBase: '#000000',
      colorBgContainer: '#0a0a0a',
      colorBgElevated: '#141414',
      colorBorder: '#262626',
      colorBorderSecondary: '#1a1a1a',
      colorText: '#e5e5e5',
      colorTextSecondary: '#a3a3a3',
      colorTextTertiary: '#737373',
    } : {
      // 浅色模式的背景色
      colorBgBase: '#ffffff',
      colorBgContainer: '#f8f9fa',
      colorBgElevated: '#ffffff',
      colorBorder: '#dee2e6',
      colorBorderSecondary: '#e9ecef',
      colorText: '#212529',
      colorTextSecondary: '#495057',
      colorTextTertiary: '#6c757d',
    }),
    borderRadius: 12,
    fontSize: 14,
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  },
  components: {
    Button: {
      controlHeight: 40,
      paddingContentHorizontal: 24,
    },
    Input: {
      controlHeight: 44,
      paddingBlock: 12,
    },
    Card: {
      borderRadiusLG: 16,
    },
    Table: {
      borderRadius: 12,
      headerBg: themeStore.theme === 'dark' ? '#141414' : '#f1f3f5',
    },
    Modal: {
      borderRadiusLG: 16,
    },
  },
}))
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  width: 100%;
}

/* 路由切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
