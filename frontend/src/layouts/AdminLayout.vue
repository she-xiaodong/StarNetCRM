<template>
  <div class="admin-layout">
    <!-- 顶部导航 -->
    <header class="admin-header">
      <div class="header-left" @click="goHome">
        <span class="logo-icon">✦</span>
        <span class="logo-text">星络客 · 管理后台</span>
        <span class="logo-sub">总后台 · 客户管理</span>
      </div>
      <div class="header-right">
        <a-button type="link" class="btn-home" @click="goHome">
          <home-outlined /> 返回官网
        </a-button>
        <a-divider type="vertical" style="border-color: rgba(255,255,255,0.2)" />
        <a-dropdown>
          <div class="user-info">
            <a-avatar size="small" style="background-color: #2B5FD7">{{ userName.charAt(0) }}</a-avatar>
            <span class="user-name">{{ userName }}</span>
          </div>
          <template #overlay>
            <a-menu>
              <a-menu-item key="logout" @click="handleLogout">
                <logout-outlined /> 退出登录
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="admin-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { HomeOutlined, LogoutOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const userName = computed(() => authStore.userName)

function goHome() {
  router.push('/')
}

function handleLogout() {
  authStore.logout()
  router.push('/')
}

onMounted(async () => {
  try {
    await authStore.fetchUserInfo()
  } catch {
    // ignore
  }
})
</script>

<style scoped lang="less">
.admin-layout {
  min-height: 100vh;
  background: #f0f2f5;
  display: flex;
  flex-direction: column;
}

.admin-header {
  height: 56px;
  background: #1a1a2e;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;

    .logo-icon {
      font-size: 22px;
      color: #4A7AE8;
    }

    .logo-text {
      font-size: 16px;
      font-weight: 600;
      letter-spacing: 1px;
    }

    .logo-sub {
      font-size: 12px;
      color: rgba(255, 255, 255, 0.5);
      margin-left: 4px;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;

    .btn-home {
      color: rgba(255, 255, 255, 0.75);

      &:hover {
        color: #fff;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;

      .user-name {
        font-size: 14px;
      }
    }
  }
}

.admin-content {
  flex: 1;
  padding: 24px;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
}
</style>
