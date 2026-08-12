<template>
  <div class="login-page">
    <div class="login-bg">
      <div class="bg-particles">
        <div
          v-for="i in 30"
          :key="i"
          class="particle"
          :style="particleStyle(i)"
        />
      </div>
    </div>

    <div class="login-container">
      <!-- 品牌区 -->
      <div class="brand-section">
        <div class="brand-icon">⭐</div>
        <h1 class="brand-name">星络客 StarNet</h1>
        <p class="brand-desc">星络相连，六度通达</p>
      </div>

      <!-- 登录表单 -->
      <div class="form-section">
        <a-tabs v-model:activeKey="activeTab" centered size="large">
          <a-tab-pane key="login" tab="登录">
            <a-form
              :model="loginForm"
              layout="vertical"
              @finish="handleLogin"
              class="login-form"
            >
              <a-form-item
                name="username"
                :rules="[{ required: true, message: '请输入用户名' }]"
              >
                <a-input
                  v-model:value="loginForm.username"
                  size="large"
                  placeholder="用户名"
                  autocomplete="username"
                >
                  <template #prefix><UserOutlined /></template>
                </a-input>
              </a-form-item>

              <a-form-item
                name="password"
                :rules="[{ required: true, message: '请输入密码' }]"
              >
                <a-input-password
                  v-model:value="loginForm.password"
                  size="large"
                  placeholder="密码"
                  autocomplete="current-password"
                  @keydown.enter="handleLogin"
                >
                  <template #prefix><LockOutlined /></template>
                </a-input-password>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  size="large"
                  block
                  html-type="submit"
                  :loading="loading"
                  class="login-btn"
                >
                  登 录
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>

          <a-tab-pane key="register" tab="注册">
            <a-form
              :model="registerForm"
              layout="vertical"
              @finish="handleRegister"
              class="login-form"
            >
              <a-form-item
                name="name"
                :rules="[{ required: true, message: '请输入姓名' }]"
              >
                <a-input
                  v-model:value="registerForm.name"
                  size="large"
                  placeholder="姓名"
                >
                  <template #prefix><SmileOutlined /></template>
                </a-input>
              </a-form-item>

              <a-form-item
                name="username"
                :rules="[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少3个字符' },
                ]"
              >
                <a-input
                  v-model:value="registerForm.username"
                  size="large"
                  placeholder="用户名"
                >
                  <template #prefix><UserOutlined /></template>
                </a-input>
              </a-form-item>

              <a-form-item
                name="password"
                :rules="[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6位' },
                ]"
              >
                <a-input-password
                  v-model:value="registerForm.password"
                  size="large"
                  placeholder="密码"
                >
                  <template #prefix><LockOutlined /></template>
                </a-input-password>
              </a-form-item>

              <a-form-item
                name="confirmPassword"
                :rules="[
                  { required: true, message: '请确认密码' },
                  { validator: validateConfirmPassword },
                ]"
              >
                <a-input-password
                  v-model:value="registerForm.confirmPassword"
                  size="large"
                  placeholder="确认密码"
                >
                  <template #prefix><CheckCircleOutlined /></template>
                </a-input-password>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  size="large"
                  block
                  html-type="submit"
                  :loading="loading"
                  class="login-btn"
                >
                  注 册
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>
        </a-tabs>

        <!-- 企微登录按钮 -->
        <a-divider>
          <span style="color: #999; font-size: 13px">其他方式</span>
        </a-divider>
        <a-button block size="large" class="wecom-btn" @click="handleWecomLogin">
          <template #icon>
            <span style="color: #07C160; font-weight: bold">We</span>
          </template>
          企业微信登录
        </a-button>
      </div>
    </div>

    <!-- 页脚 -->
    <div class="login-footer">
      <span>&copy; 2024 星络客 StarNet. All Rights Reserved.</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useAuthStore } from '@/stores/auth'
import { getWecomOAuthConfig } from '@/api'
import {
  UserOutlined,
  LockOutlined,
  SmileOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const activeTab = ref('login')
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
})

const registerForm = reactive({
  name: '',
  username: '',
  password: '',
  confirmPassword: '',
})

// 背景粒子样式
function particleStyle(i: number) {
  return {
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    animationDelay: `${Math.random() * 3}s`,
    animationDuration: `${3 + Math.random() * 4}s`,
    width: `${2 + Math.random() * 3}px`,
    height: `${2 + Math.random() * 3}px`,
  }
}

function validateConfirmPassword(_rule: any, value: string) {
  if (value !== registerForm.password) {
    return Promise.reject('两次密码不一致')
  }
  return Promise.resolve()
}

async function handleLogin() {
  if (!loginForm.username || !loginForm.password) return

  loading.value = true
  try {
    await authStore.login(loginForm.username, loginForm.password)
    message.success('登录成功')
    router.push('/dashboard')
  } catch (err: any) {
    message.error(err.response?.data?.message || '用户名或密码错误')
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!registerForm.name || !registerForm.username || !registerForm.password) return

  loading.value = true
  try {
    await authStore.register({
      name: registerForm.name,
      username: registerForm.username,
      password: registerForm.password,
    })
    message.success('注册成功')
    router.push('/dashboard')
  } catch (err: any) {
    message.error(err.response?.data?.message || '注册失败')
  } finally {
    loading.value = false
  }
}

const wecomConfig = ref<{ corp_id: string; redirect_uri: string } | null>(null)

onMounted(async () => {
  try {
    const res = await getWecomOAuthConfig()
    wecomConfig.value = res.data.data
  } catch {
    // 企微配置未就绪，按钮仍可点击但跳转会失败
  }
})

function handleWecomLogin() {
  const cfg = wecomConfig.value
  if (!cfg?.corp_id || !cfg?.redirect_uri) {
    message.warning('企业微信登录尚未配置，请联系管理员')
    return
  }
  const redirectUri = encodeURIComponent(cfg.redirect_uri)
  const state = Math.random().toString(36).substring(2, 10)
  const wxUrl = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${cfg.corp_id}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_base&state=${state}#wechat_redirect`
  window.location.href = wxUrl
}
</script>

<style lang="less" scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #1a237e 0%, #2B5FD7 40%, #4A90D9 70%, #87CEEB 100%);
}

.login-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;

  .particle {
    position: absolute;
    background: rgba(255, 255, 255, 0.3);
    border-radius: 50%;
    animation: float linear infinite;
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) scale(1);
    opacity: 0;
  }
  20% {
    opacity: 0.4;
  }
  80% {
    opacity: 0.1;
  }
  100% {
    transform: translateY(-100vh) scale(0.5);
    opacity: 0;
  }
}

.login-container {
  position: relative;
  z-index: 10;
  display: flex;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
}

.brand-section {
  width: 320px;
  padding: 48px 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #2B5FD7 0%, #4A90D9 100%);

  .brand-icon {
    font-size: 56px;
    margin-bottom: 16px;
  }

  .brand-name {
    color: #fff;
    font-size: 24px;
    font-weight: 700;
    margin-bottom: 8px;
    letter-spacing: 4px;
  }

  .brand-desc {
    color: rgba(255, 255, 255, 0.8);
    font-size: 14px;
    letter-spacing: 2px;
  }
}

.form-section {
  width: 400px;
  padding: 40px 32px;

  .login-form {
    margin-top: 8px;
  }

  .login-btn {
    height: 44px;
    font-size: 16px;
    letter-spacing: 4px;
    background: #2B5FD7;
    border-color: #2B5FD7;
    border-radius: 8px;

    &:hover {
      background: #4A90D9 !important;
      border-color: #4A90D9 !important;
    }
  }

  .wecom-btn {
    border-radius: 8px;
    border-color: #07C160;
    color: #07C160;

    &:hover {
      border-color: #07C160 !important;
      color: #07C160 !important;
      background: rgba(7, 193, 96, 0.05);
    }
  }
}

.login-footer {
  position: absolute;
  bottom: 24px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  z-index: 10;
}

// 响应式
@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    margin: 0 16px;
  }

  .brand-section {
    width: 100%;
    padding: 32px;
  }

  .form-section {
    width: 100%;
    padding: 24px;
  }
}
</style>
