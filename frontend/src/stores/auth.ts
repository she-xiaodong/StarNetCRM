import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/api'

export interface UserInfo {
  user_id: string
  tenant_id: string
  name: string
  role: string
  phone: string
  email: string
  department: string
  avatar: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const userName = computed(() => userInfo.value?.name || '')

  // 登录
  async function login(username: string, password: string) {
    const res = await request.post('/auth/login', { username, password })
    const data = res.data.data
    token.value = data.token
    localStorage.setItem('token', data.token)
    await fetchUserInfo()
    return data
  }

  // 注册
  async function register(params: {
    username: string
    password: string
    name: string
    phone?: string
    email?: string
  }) {
    const res = await request.post('/auth/register', params)
    const data = res.data.data
    token.value = data.token
    localStorage.setItem('token', data.token)
    await fetchUserInfo()
    return data
  }

  // 获取用户信息
  async function fetchUserInfo() {
    const res = await request.get('/user/me')
    userInfo.value = res.data.data
  }

  // 登出
  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isAdmin,
    userName,
    login,
    register,
    fetchUserInfo,
    logout,
  }
})
