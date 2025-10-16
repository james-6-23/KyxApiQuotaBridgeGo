<template>
  <Layout>
    <div class="min-h-[calc(100vh-200px)] flex items-center justify-center px-6">
      <div class="max-w-md w-full animate-fade-in">
        <div class="grok-card p-8">
          <div class="text-center mb-8">
            <h2 class="text-2xl font-bold text-grok-text mb-2">绑定公益站账号</h2>
            <p class="text-sm text-grok-text-secondary">
              绑定后才能领取额度和投喂 Keys
            </p>
          </div>

          <a-form @finish="handleBind" layout="vertical">
            <a-form-item
              label="公益站用户名"
              name="username"
              :rules="[{ required: true, message: '请输入公益站用户名' }]"
            >
              <a-input
                v-model:value="form.username"
                placeholder="请输入公益站用户名"
                size="large"
                class="tech-input"
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
                绑定账号
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user'
import Layout from '@/components/Layout.vue'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  username: '',
})
const loading = ref(false)

const handleBind = async () => {
  loading.value = true
  try {
    const response = await authApi.bind({ username: form.value.username })
    if (response.success) {
      message.success('绑定成功！')
      await userStore.refreshUserInfo()
      router.push('/dashboard')
    }
  } catch (error) {
    // 错误已在拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>
