<template>
  <div class="referral-page">
    <div class="page-header">
      <h3>引荐管理</h3>
      <a-button type="primary" @click="showCreateModal = true">
        <template #icon><PlusOutlined /></template>
        发起引荐
      </a-button>
    </div>

    <!-- 统计卡片 -->
    <a-row :gutter="16" class="stats-row">
      <a-col :span="6">
        <a-statistic title="全部" :value="12" />
      </a-col>
      <a-col :span="6">
        <a-statistic title="草稿" :value="3" value-style="color: #999" />
      </a-col>
      <a-col :span="6">
        <a-statistic title="已发送" :value="5" value-style="color: #1890FF" />
      </a-col>
      <a-col :span="6">
        <a-statistic title="已成功" :value="4" value-style="color: #52C41A" />
      </a-col>
    </a-row>

    <!-- 引荐列表 -->
    <a-card class="list-card">
      <a-tabs v-model:activeKey="activeTab">
        <a-tab-pane key="all" tab="全部" />
        <a-tab-pane key="draft" tab="草稿" />
        <a-tab-pane key="sent" tab="已发送" />
        <a-tab-pane key="accepted" tab="已接受" />
        <a-tab-pane key="connected" tab="已连接" />
      </a-tabs>

      <a-table
        :columns="columns"
        :data-source="referrals"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'path'">
            <div class="path-preview">
              {{ record.fromName }}
              <span style="color: #2B5FD7; margin: 0 4px">→</span>
              {{ record.middleName }}
              <span style="color: #2B5FD7; margin: 0 4px">→</span>
              {{ record.targetName }}
            </div>
          </template>
          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </a-tag>
          </template>
          <template v-if="column.key === 'action'">
            <a-space>
              <a-button
                v-if="record.status === 'draft'"
                type="link"
                size="small"
                @click="handleSend(record)"
              >
                发送
              </a-button>
              <a-button type="link" size="small" @click="handleView(record)">
                详情
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'

const activeTab = ref('all')
const showCreateModal = ref(false)

const columns = [
  { title: '引荐路径', key: 'path' },
  { title: '引荐理由', dataIndex: 'reason', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', width: 120 },
  { title: '操作', key: 'action', width: 140 },
]

const referrals = ref([
  {
    id: '1',
    fromName: '我',
    middleName: '张三',
    targetName: '赵六',
    reason: '寻求融资合伙人，赵六在百度有丰富的AI领域投资经验',
    status: 'draft',
    createdAt: '2024-03-15',
  },
  {
    id: '2',
    fromName: '我',
    middleName: '李四',
    targetName: '周八',
    reason: '拓展腾讯云业务合作',
    status: 'sent',
    createdAt: '2024-03-10',
  },
  {
    id: '3',
    fromName: '我',
    middleName: '王五',
    targetName: '郑十',
    reason: '商务合作对接',
    status: 'connected',
    createdAt: '2024-03-05',
  },
])

function getStatusColor(status: string) {
  const map: Record<string, string> = {
    draft: 'default',
    sent: 'processing',
    accepted: 'blue',
    rejected: 'red',
    connected: 'success',
  }
  return map[status] || 'default'
}

function getStatusText(status: string) {
  const map: Record<string, string> = {
    draft: '草稿',
    sent: '已发送',
    accepted: '已接受',
    rejected: '已拒绝',
    connected: '已连接',
  }
  return map[status] || status
}

function handleSend(record: any) {
  message.success(`已发送对 ${record.targetName} 的引荐`)
}

function handleView(record: any) {
  message.info(`查看引荐详情: ${record.id}`)
}
</script>

<style lang="less" scoped>
.referral-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h3 {
      font-size: 20px;
      font-weight: 600;
    }
  }

  .stats-row {
    margin-bottom: 16px;
  }

  .list-card {
    border-radius: 12px;

    .path-preview {
      font-size: 13px;
      color: #333;
    }
  }
}
</style>
