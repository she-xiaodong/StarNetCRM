import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 路由配置
const routes: RouteRecordRaw[] = [
  // ─── 官网公开页面 ───
  {
    path: '/',
    name: 'Landing',
    component: () => import('@/views/Landing.vue'),
    meta: { title: '星络客 StarNet - 关系型CRM', requiresAuth: false },
  },
  {
    path: '/pricing',
    name: 'Pricing',
    component: () => import('@/views/Pricing.vue'),
    meta: { title: '定价方案 - 星络客', requiresAuth: false },
  },
  {
    path: '/wecom-callback',
    name: 'WecomCallback',
    component: () => import('@/views/WecomCallback.vue'),
    meta: { title: '企微登录 - 星络客', requiresAuth: false },
  },
  // ─── 工作台（需登录） ───
  {
    path: '/app',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/app/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '首页', icon: 'DashboardOutlined' },
      },
      {
        path: 'contacts',
        name: 'Contacts',
        component: () => import('@/views/Contacts.vue'),
        meta: { title: '联系人', icon: 'TeamOutlined' },
      },
      {
        path: 'graph',
        name: 'Graph',
        component: () => import('@/views/Graph.vue'),
        meta: { title: '关系图谱', icon: 'ApartmentOutlined' },
      },
      {
        path: 'path-search',
        name: 'PathSearch',
        component: () => import('@/views/PathSearch.vue'),
        meta: { title: '路径查询', icon: 'SearchOutlined' },
      },
      {
        path: 'referral',
        name: 'Referral',
        component: () => import('@/views/Referral.vue'),
        meta: { title: '引荐管理', icon: 'SendOutlined' },
      },
      {
        path: 'analytics',
        name: 'Analytics',
        component: () => import('@/views/Analytics.vue'),
        meta: { title: '人脉分析', icon: 'BarChartOutlined' },
      },
      {
        path: 'tags',
        name: 'Tags',
        component: () => import('@/views/Tags.vue'),
        meta: { title: '标签库', icon: 'TagsOutlined' },
      },
    ],
  },
  // ─── 管理后台（独立地址，需 admin 角色） ───
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/admin/overview',
    meta: { title: '管理后台', requireAdmin: true },
    children: [
      {
        path: 'overview',
        name: 'Admin',
        component: () => import('@/views/Admin.vue'),
        meta: { title: '管理后台', requireAdmin: true },
      },
    ],
  },
  // ─── 旧版登录页（保留兼容） ───
  {
    path: '/login',
    redirect: '/',
  },
  // ─── 旧版工作台路径（保留兼容，重定向到新路径） ───
  {
    path: '/dashboard',
    redirect: '/app/dashboard',
  },
  {
    path: '/contacts',
    redirect: '/app/contacts',
  },
  {
    path: '/graph',
    redirect: '/app/graph',
  },
  {
    path: '/path-search',
    redirect: '/app/path-search',
  },
  {
    path: '/referral',
    redirect: '/app/referral',
  },
  {
    path: '/analytics',
    redirect: '/app/analytics',
  },
  {
    path: '/tags',
    redirect: '/app/tags',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  // 设置页面标题
  document.title = (to.meta.title as string) || '星络客 StarNet'

  const token = localStorage.getItem('token')
  const publicPaths = ['/', '/pricing', '/wecom-callback']

  if (publicPaths.includes(to.path)) {
    // 公开页面永远放行
    next()
  } else if (!token) {
    // 需登录但无token → 回首页
    next('/')
  } else if (to.meta.requireAdmin) {
    // 管理后台：需要 admin 角色
    const { useAuthStore } = await import('@/stores/auth')
    const auth = useAuthStore()
    if (!auth.userInfo) {
      try {
        await auth.fetchUserInfo()
      } catch {
        // token 失效，交给请求拦截处理
      }
    }
    if (auth.isAdmin) {
      next()
    } else {
      next('/app/dashboard')
    }
  } else {
    next()
  }
})

export default router
