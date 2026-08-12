<template>
  <div class="dashboard">
    <!-- 欢迎区 -->
    <div class="welcome-section">
      <h2 class="welcome-title">👋 欢迎回来，{{ authStore.userName }}</h2>
      <p class="welcome-sub">今天是你的人脉增长日</p>
    </div>

    <!-- 统计卡片 -->
    <a-row :gutter="16" class="stats-row">
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon contacts-icon">
              <TeamOutlined />
            </div>
            <div class="stat-info">
              <span class="stat-value">128</span>
              <span class="stat-label">联系人数</span>
            </div>
          </div>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon relation-icon">
              <ApartmentOutlined />
            </div>
            <div class="stat-info">
              <span class="stat-value">356</span>
              <span class="stat-label">关系连结数</span>
            </div>
          </div>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon path-icon">
              <SendOutlined />
            </div>
            <div class="stat-info">
              <span class="stat-value">12</span>
              <span class="stat-label">引荐进行中</span>
            </div>
          </div>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card" hoverable>
          <div class="stat-content">
            <div class="stat-icon score-icon">
              <TrophyOutlined />
            </div>
            <div class="stat-info">
              <span class="stat-value">85</span>
              <span class="stat-label">人脉评分</span>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 下半部分 -->
    <a-row :gutter="16" class="bottom-row">
      <!-- 最近联系人 -->
      <a-col :xs="24" :lg="12">
        <a-card title="最近添加的联系人" size="small">
          <template #extra>
            <a @click="$router.push('/contacts')">查看全部</a>
          </template>
          <a-list :data-source="recentContacts" size="small">
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta :title="item.name" :description="item.company">
                  <template #avatar>
                    <a-avatar :style="{ backgroundColor: '#2B5FD7' }">
                      {{ item.name.charAt(0) }}
                    </a-avatar>
                  </template>
                </a-list-item-meta>
                <template #actions>
                  <a-tag color="blue">{{ item.relation }}</a-tag>
                </template>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>

      <!-- 快速操作 -->
      <a-col :xs="24" :lg="12">
        <a-card title="快速操作" size="small">
          <a-row :gutter="[16, 16]">
            <a-col :span="8">
              <div class="quick-action" @click="$router.push('/contacts')">
                <div class="action-icon" style="background: #e6f0ff">
                  <TeamOutlined style="color: #2B5FD7; font-size: 24px" />
                </div>
                <span>添加联系人</span>
              </div>
            </a-col>
            <a-col :span="8">
              <div class="quick-action" @click="$router.push('/graph')">
                <div class="action-icon" style="background: #e6ffe6">
                  <ApartmentOutlined style="color: #00C9A7; font-size: 24px" />
                </div>
                <span>查看图谱</span>
              </div>
            </a-col>
            <a-col :span="8">
              <div class="quick-action" @click="$router.push('/path-search')">
                <div class="action-icon" style="background: #fff7e6">
                  <SearchOutlined style="color: #F5A623; font-size: 24px" />
                </div>
                <span>路径搜索</span>
              </div>
            </a-col>
            <a-col :span="8">
              <div class="quick-action" @click="$router.push('/referral')">
                <div class="action-icon" style="background: #ffe6f0">
                  <SendOutlined style="color: #FF6B6B; font-size: 24px" />
                </div>
                <span>发起引荐</span>
              </div>
            </a-col>
            <a-col :span="8">
              <div class="quick-action">
                <div class="action-icon" style="background: #f0e6ff">
                  <InboxOutlined style="color: #722ED1; font-size: 24px" />
                </div>
                <span>导入数据</span>
              </div>
            </a-col>
            <a-col :span="8">
              <div class="quick-action" @click="$router.push('/analytics')">
                <div class="action-icon" style="background: #e6f4ff">
                  <BarChartOutlined style="color: #1890FF; font-size: 24px" />
                </div>
                <span>人脉分析</span>
              </div>
            </a-col>
          </a-row>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import {
  TeamOutlined,
  ApartmentOutlined,
  SendOutlined,
  TrophyOutlined,
  SearchOutlined,
  InboxOutlined,
  BarChartOutlined,
} from '@ant-design/icons-vue'

const authStore = useAuthStore()

const recentContacts = ref([
  { name: '张三', company: '阿里巴巴', relation: '同事' },
  { name: '李四', company: '腾讯科技', relation: '合作伙伴' },
  { name: '王五', company: '字节跳动', relation: '客户' },
  { name: '赵六', company: '百度集团', relation: '校友' },
])
</script>

<style lang="less" scoped>
.dashboard {
  .welcome-section {
    margin-bottom: 24px;

    .welcome-title {
      font-size: 22px;
      font-weight: 600;
      color: #1A1A1A;
      margin-bottom: 4px;
    }

    .welcome-sub {
      color: #999;
      font-size: 14px;
    }
  }

  .stats-row {
    margin-bottom: 16px;

    .stat-card {
      border-radius: 12px;
      transition: all 0.3s;

      &:hover {
        transform: translateY(-2px);
      }

      .stat-content {
        display: flex;
        align-items: center;
        gap: 16px;

        .stat-icon {
          width: 52px;
          height: 52px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 24px;
          color: #fff;
        }

        .contacts-icon { background: linear-gradient(135deg, #2B5FD7, #4A90D9); }
        .relation-icon { background: linear-gradient(135deg, #00C9A7, #52C41A); }
        .path-icon { background: linear-gradient(135deg, #F5A623, #FA8C16); }
        .score-icon { background: linear-gradient(135deg, #722ED1, #2B5FD7); }

        .stat-info {
          display: flex;
          flex-direction: column;

          .stat-value {
            font-size: 28px;
            font-weight: 700;
            color: #1A1A1A;
          }

          .stat-label {
            font-size: 13px;
            color: #999;
          }
        }
      }
    }
  }

  .bottom-row {
    .quick-action {
      display: flex;
      flex-direction: column;
      align-items: center;
      cursor: pointer;
      padding: 12px;
      border-radius: 8px;
      transition: all 0.3s;

      &:hover {
        background: #f5f5f5;
      }

      .action-icon {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 8px;
      }

      span {
        font-size: 13px;
        color: #666;
      }
    }
  }
}
</style>
