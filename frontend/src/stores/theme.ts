import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  // 从 localStorage 读取用户偏好，默认为浅色模式
  const theme = ref<Theme>((localStorage.getItem('theme') as Theme) || 'light')

  // 切换主题
  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  // 设置主题
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
  }

  // 监听主题变化，更新 DOM 和 localStorage
  watch(
    theme,
    (newTheme) => {
      // 更新 HTML class
      if (newTheme === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }

      // 保存到 localStorage
      localStorage.setItem('theme', newTheme)
    },
    { immediate: true }
  )

  return {
    theme,
    toggleTheme,
    setTheme,
  }
})
