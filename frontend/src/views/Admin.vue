<template>
  <div class="min-h-screen bg-light-bg dark:bg-dark-bg">
    <div class="flex items-center justify-center min-h-screen px-6">
      <div class="max-w-md w-full theme-card p-8 animate-fade-in">
        <div class="text-center mb-8">
          <div class="w-16 h-16 rounded-xl bg-gradient-to-br from-primary to-purple flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <h2 class="text-2xl font-bold text-light-text dark:text-dark-text">管理员登录</h2>
        </div>

        <a-form @finish="handleLogin" layout="vertical">
          <a-form-item
            name="password"
            :rules="[{ required: true, message: '请输入管理员密码' }]"
          >
            <a-input-password
              v-model:value="form.password"
              placeholder="请输入管理员密码"
              size="large"
              class="tech-input"
              @keyup.enter="handleLogin"
            />
          </a-form-item>

          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              :loading="loading"
              block
              class="tech-button h-12"
            >
              登录
            </a-button>
          </a-form-item>
        </a-form>

        <div class="text-center">
          <a-button type="link" @click="$router.push('/')">
            返回首页
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { adminApi } from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  password: '',
})
const loading = ref(false)

const handleLogin = async () => {
  loading.value = true
  try {
    const response = await adminApi.login(form.value.password)
    if (response.success) {
      message.success('登录成功')
      userStore.setAdminStatus(true)
      router.push('/admin')
    }
  } catch (error) {
    // 错误已在拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>
