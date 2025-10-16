<template>
  <div class="min-h-screen bg-light-bg dark:bg-dark-bg grid-background">
    <!-- 导航栏 -->
    <header class="glass-effect sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-purple flex items-center justify-center text-white font-bold text-xl shadow-lg">
            K
          </div>
          <span class="text-xl font-bold gradient-text">KYX API Bridge</span>
        </div>

        <!-- 主题切换 + 用户菜单 -->
        <div class="flex items-center gap-4">
          <!-- 主题切换按钮 -->
          <a-button
            @click="themeStore.toggleTheme()"
            class="tech-button !p-2 !w-10 !h-10 flex items-center justify-center"
            :title="themeStore.theme === 'light' ? '切换到深色模式' : '切换到浅色模式'"
          >
            <!-- 太阳图标 (浅色模式) -->
            <svg v-if="themeStore.theme === 'light'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <!-- 月亮图标 (深色模式) -->
            <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          </a-button>

          <!-- 登录按钮 -->
          <a-button v-if="!userStore.isLoggedIn" type="primary" size="large" @click="userStore.login" class="tech-button">
            <span class="flex items-center gap-2">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
              </svg>
              Linux Do 登录
            </span>
          </a-button>

          <!-- 用户下拉菜单 -->
          <a-dropdown v-else>
            <a-button class="tech-button flex items-center gap-3">
              <a-avatar :src="userStore.userInfo?.avatar_url" :size="32" />
              <span>{{ userStore.userInfo?.display_name }}</span>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </a-button>
            <template #overlay>
              <a-menu class="theme-card">
                <a-menu-item v-if="userStore.isAdmin" @click="$router.push('/admin')">
                  <span class="flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    管理后台
                  </span>
                </a-menu-item>
                <a-menu-divider v-if="userStore.isAdmin" />
                <a-menu-item @click="handleLogout">
                  <span class="flex items-center gap-2 text-error">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                    </svg>
                    退出登录
                  </span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </div>
    </header>

    <!-- 内容区 -->
    <main>
      <slot />
    </main>

    <!-- 页脚 -->
    <footer class="py-8 text-center text-light-text-tertiary dark:text-dark-text-tertiary">
      <p class="text-sm">
        © 2024 KYX API Bridge • Built with Vue 3 + Vite + Tailwind CSS
      </p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const themeStore = useThemeStore()
const router = useRouter()

const handleLogout = async () => {
  await userStore.logout()
  message.success('已退出登录')
  router.push('/')
}
</script>
