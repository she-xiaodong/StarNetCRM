import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { message } from 'ant-design-vue'
import router from '@/router'

// ─── API 基础配置 ───
const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// 响应数据结构
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface PageData<T = any> {
  total: number
  page: number
  page_size: number
  list: T[]
}

// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data
    if (res.code !== 0) {
      message.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message))
    }
    return response
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 401:
          localStorage.removeItem('token')
          message.error('登录已过期，请重新登录')
          router.push('/login')
          break
        case 403:
          message.error('无权限访问')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器内部错误')
          break
        default:
          message.error(data?.message || '请求失败')
      }
    } else {
      message.error('网络连接异常')
    }
    return Promise.reject(error)
  }
)

// ─── 企微登录 ───
export function wecomLogin(code: string, state?: string) {
  return request.post('/auth/wecom', { code, state })
}

// ─── 企微OAuth配置 ───
export function getWecomOAuthConfig() {
  return request.get('/auth/wecom/config')
}

// ─── 联系人 CRUD ───
export function getContacts(params?: { keyword?: string; page?: number; page_size?: number }) {
  return request.get('/contacts', { params })
}
export function getContact(id: string) {
  return request.get(`/contacts/${id}`)
}
export function createContact(data: {
  name: string; company?: string; title?: string; department?: string
  phone?: string; email?: string; tags?: string[]; note?: string
}) {
  return request.post('/contacts', data)
}
export function updateContact(id: string, data: Record<string, any>) {
  return request.put(`/contacts/${id}`, data)
}
export function deleteContact(id: string) {
  return request.delete(`/contacts/${id}`)
}

// ─── 图谱 ───
export function getFirstDegree(personId?: string) {
  return request.get('/graph/first-degree', { params: personId ? { person_id: personId } : {} })
}
export function createRelation(data: {
  from_id?: string; to_id: string; type?: string; source?: string; strength?: number
}) {
  return request.post('/graph/relations', data)
}
export function searchPath(startId: string, endId: string) {
  return request.post('/graph/search-path', { start_id: startId, end_id: endId })
}

// ─── 标签管理 ───
export function getTags() {
  return request.get('/tags')
}
export function createTag(data: { name: string; color?: string; type?: string }) {
  return request.post('/tags', data)
}
export function updateTag(id: string, data: { name?: string; color?: string; type?: string }) {
  return request.put(`/tags/${id}`, data)
}
export function deleteTag(id: string) {
  return request.delete(`/tags/${id}`)
}

// ─── 首页统计 ───
export interface DashboardStats {
  total_contacts: number
  total_relations: number
  active_referrals: number
  network_score: number
  recent_contacts: {
    id: string
    name: string
    company?: string
    title?: string
    tags?: string
    created_at?: string
  }[]
}
export function getDashboardStats() {
  return request.get<DashboardStats>('/stats/dashboard')
}

// ─── 人脉分析 ───
export interface AnalyticsStats {
  active_relations: number
  week_new_contacts: number
  avg_path_length: number
  isolated_nodes: number
  relation_distribution: {
    type: string
    label: string
    count: number
    percent: number
  }[]
  growth_trend: {
    month: string
    count: number
  }[]
  super_connectors: {
    id: string
    name: string
    company?: string
    degree: number
    top_connection?: string
  }[]
}
export function getAnalytics() {
  return request.get<AnalyticsStats>('/stats/analytics')
}

export default request
