<template>
  <Layout>
    <div class="min-h-[calc(100vh-200px)] flex items-center justify-center px-6">
      <div class="max-w-4xl w-full">
        <!-- Hero Section -->
        <div class="text-center mb-16 animate-fade-in">
          <!-- Logo & Title -->
          <div class="mb-8 flex justify-center">
            <div class="w-24 h-24 rounded-2xl bg-gradient-to-br from-grok-primary via-grok-purple to-grok-primary shadow-grok-lg flex items-center justify-center text-white text-4xl font-bold animate-glow">
              K
            </div>
          </div>

          <h1 class="text-5xl md:text-6xl font-bold mb-6 gradient-text">
            KYX API 配额桥接系统
          </h1>

          <p class="text-xl text-grok-text-secondary mb-8 leading-relaxed">
            现代化的 API 配额管理平台 <br class="hidden md:block" />
            安全 · 高效 · 便捷
          </p>

          <!-- CTA Button -->
          <a-button
            v-if="!userStore.isLoggedIn"
            type="primary"
            size="large"
            @click="userStore.login"
            class="tech-button px-12 py-6 h-auto text-lg font-medium"
          >
            <span class="flex items-center gap-3">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              开始使用
            </span>
          </a-button>

          <a-button
            v-else
            type="primary"
            size="large"
            @click="goToDashboard"
            class="tech-button px-12 py-6 h-auto text-lg font-medium"
          >
            进入控制台
          </a-button>
        </div>

        <!-- Features Grid -->
        <div class="grid md:grid-cols-3 gap-6 mb-16 animate-slide-up animation-delay-200">
          <div class="grok-card p-8 text-center group cursor-pointer">
            <div class="w-16 h-16 rounded-xl bg-grok-bg-hover flex items-center justify-center mx-auto mb-4 group-hover:bg-grok-primary/20 transition-colors">
              <svg class="w-8 h-8 text-grok-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold mb-2 text-grok-text">每日领取</h3>
            <p class="text-sm text-grok-text-secondary">每天领取免费 API 配额，轻松使用</p>
          </div>

          <div class="grok-card p-8 text-center group cursor-pointer">
            <div class="w-16 h-16 rounded-xl bg-grok-bg-hover flex items-center justify-center mx-auto mb-4 group-hover:bg-grok-success/20 transition-colors">
              <svg class="w-8 h-8 text-grok-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold mb-2 text-grok-text">Keys 投喂</h3>
            <p class="text-sm text-grok-text-secondary">投喂 ModelScope Keys 获取更多配额</p>
          </div>

          <div class="grok-card p-8 text-center group cursor-pointer">
            <div class="w-16 h-16 rounded-xl bg-grok-bg-hover flex items-center justify-center mx-auto mb-4 group-hover:bg-grok-purple/20 transition-colors">
              <svg class="w-8 h-8 text-grok-purple" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold mb-2 text-grok-text">实时统计</h3>
            <p class="text-sm text-grok-text-secondary">实时查看配额使用情况和历史记录</p>
          </div>
        </div>

        <!-- Stats Banner -->
        <div class="grok-card p-8 bg-gradient-grok animate-slide-up animation-delay-400">
          <div class="flex flex-wrap items-center justify-around gap-6 text-center">
            <div>
              <div class="stat-value">1000+</div>
              <div class="stat-label">活跃用户</div>
            </div>
            <div class="w-px h-12 bg-grok-border hidden md:block"></div>
            <div>
              <div class="stat-value">50K+</div>
              <div class="stat-label">Keys 投喂</div>
            </div>
            <div class="w-px h-12 bg-grok-border hidden md:block"></div>
            <div>
              <div class="stat-value">99.9%</div>
              <div class="stat-label">服务可用性</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import Layout from '@/components/Layout.vue'

const router = useRouter()
const userStore = useUserStore()

const goToDashboard = () => {
  if (userStore.isBound) {
    router.push('/dashboard')
  } else {
    router.push('/bind')
  }
}
</script>
