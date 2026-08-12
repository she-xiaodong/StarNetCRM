<template>
  <div class="wecom-callback-container">
    <div class="callback-card">
      <a-spin size="large" />
      <h2>正在通过企业微信登录...</h2>
      <p>请稍候，正在验证您的身份</p>
      <p v-if="error" class="error-msg">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import request from '@/api'

const router = useRouter()
const route = useRoute()
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code) {
    error.value = '缺少授权码，请重新登录'
    setTimeout(() => router.replace('/login'), 3000)
    return
  }

  try {
    const res = await request.post('/auth/wecom', { code, state })
    const data = res.data.data

    // 存储 token
    localStorage.setItem('token', data.token)

    message.success('登录成功')
    router.replace('/dashboard')
  } catch (err: any) {
    const msg = err?.response?.data?.message || '企微登录失败'
    error.value = msg
    message.error(msg)
    setTimeout(() => router.replace('/login'), 3000)
  }
})
</script>

<style scoped>
.wecom-callback-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.callback-card {
  background: #fff;
  border-radius: 16px;
  padding: 60px 80px;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.callback-card h2 {
  margin-top: 24px;
  font-size: 20px;
  color: #333;
}

.callback-card p {
  margin-top: 8px;
  color: #999;
  font-size: 14px;
}

.error-msg {
  color: #ff4d4f !important;
  margin-top: 16px !important;
}
</style>
