<template>
  <a-layout class="main-layout">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      class="layout-sider"
      :width="220"
    >
      <!-- Logo区域 -->
      <div class="logo-area">
        <div class="logo-icon">⭐</div>
        <span v-show="!collapsed" class="logo-text">星络客</span>
      </div>

      <!-- 导航菜单 -->
      <a-menu
        v-model:selectedKeys="selectedKeys"
        mode="inline"
        theme="dark"
        @click="handleMenuClick"
      >
        <a-menu-item key="dashboard">
          <DashboardOutlined />
          <span>首页</span>
        </a-menu-item>
        <a-menu-item key="contacts">
          <TeamOutlined />
          <span>联系人</span>
        </a-menu-item>
        <a-menu-item key="graph">
          <ApartmentOutlined />
          <span>关系图谱</span>
        </a-menu-item>
        <a-menu-item key="path-search">
          <SearchOutlined />
          <span>路径查询</span>
        </a-menu-item>
        <a-menu-item key="referral">
          <SendOutlined />
          <span>引荐管理</span>
        </a-menu-item>
        <a-menu-item key="analytics">
          <BarChartOutlined />
          <span>人脉分析</span>
        </a-menu-item>
        <a-menu-item key="tags">
          <TagsOutlined />
          <span>标签库</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <!-- 主内容区 -->
    <a-layout>
      <!-- 顶部栏 -->
      <a-layout-header class="layout-header">
        <div class="header-left">
          <MenuFoldOutlined
            v-if="!collapsed"
            class="trigger"
            @click="collapsed = true"
          />
          <MenuUnfoldOutlined
            v-else
            class="trigger"
            @click="collapsed = false"
          />
          <a-breadcrumb class="header-breadcrumb">
            <a-breadcrumb-item>首页</a-breadcrumb-item>
            <a-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</a-breadcrumb-item>
          </a-breadcrumb>
        </div>
        <div class="header-right">
          <a-dropdown>
            <div class="user-info">
              <a-avatar size="small" style="background-color: #2B5FD7">
                {{ userName.charAt(0) }}
              </a-avatar>
              <span class="user-name">{{ userName }}</span>
            </div>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile">
                  <UserOutlined />
                  <span>个人信息</span>
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  <span>退出登录</span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 内容区 -->
      <a-layout-content class="layout-content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  DashboardOutlined,
  TeamOutlined,
  ApartmentOutlined,
  SearchOutlined,
  SendOutlined,
  BarChartOutlined,
  TagsOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const collapsed = ref(false)
const selectedKeys = ref<string[]>(['dashboard'])

const userName = computed(() => authStore.userName)
const currentTitle = computed(() => route.meta?.title as string || '')

// 菜单点击路由跳转
function handleMenuClick({ key }: { key: string }) {
  router.push(`/app/${key}`)
}

// 登出
function handleLogout() {
  authStore.logout()
  router.push('/')
}

onMounted(async () => {
  // 尝试获取用户信息
  try {
    await authStore.fetchUserInfo()
  } catch {
    // 忽略错误
  }

  // 同步激活菜单
  const path = route.path.replace('/app/', '')
  if (path) {
    selectedKeys.value = [path]
  }
})
</script>

<style lang="less" scoped>
.main-layout {
  height: 100%;
}

.layout-sider {
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.06);

  .logo-area {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    .logo-icon {
      font-size: 28px;
      margin-right: 8px;
    }

    .logo-text {
      color: #fff;
      font-size: 18px;
      font-weight: 600;
      letter-spacing: 2px;
    }
  }
}

.layout-header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  z-index: 10;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .trigger {
      font-size: 18px;
      cursor: pointer;
      color: #666;
      transition: color 0.3s;

      &:hover {
        color: #2B5FD7;
      }
    }
  }

  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;

      .user-name {
        font-size: 14px;
        color: #333;
      }
    }
  }
}

.layout-content {
  margin: 16px;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  min-height: calc(100vh - 96px);
  overflow-y: auto;
}
</style>
