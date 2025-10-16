<template>
  <div class="min-h-screen bg-light-bg dark:bg-dark-bg flex items-center justify-center px-6">
    <div class="text-center">
      <div class="text-9xl font-bold gradient-text mb-4">404</div>
      <h1 class="text-3xl font-bold text-light-text dark:text-dark-text mb-4">页面不存在</h1>
      <p class="text-light-text-secondary dark:text-dark-text-secondary mb-8">
        抱歉，您访问的页面不存在或已被移除
      </p>
      <a-button type="primary" size="large" @click="$router.push('/')" class="tech-button">
        返回首页
      </a-button>
    </div>
  </div>
</template>
