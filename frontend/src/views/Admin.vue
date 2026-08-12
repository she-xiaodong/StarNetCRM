<template>
  <div class="admin-page">
    <a-page-header title="管理后台" sub-title="客户管理 · 账号开通 · 订阅管理" />

    <!-- ─── 统计概览 ─── -->
    <a-row :gutter="16" class="stats-row">
      <a-col :span="6">
        <a-card :bordered="false"><a-statistic title="租户总数" :value="stats.totalTenants" /></a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false"><a-statistic title="活跃租户" :value="stats.activeTenants" :value-style="{ color: '#00C9A7' }" /></a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false"><a-statistic title="总用户数" :value="stats.totalUsers" /></a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false"><a-statistic title="本月新增" :value="stats.newThisMonth" :value-style="{ color: '#2B5FD7' }" /></a-card>
      </a-col>
    </a-row>

    <!-- ─── 租户列表 ─── -->
    <a-card title="客户列表" :bordered="false" style="margin-top: 16px">
      <template #extra>
        <a-space>
          <a-input-search v-model:value="searchKeyword" placeholder="搜索客户..." style="width: 240px" @search="fetchTenants" />
          <a-button type="primary" @click="showCreate = true">
            <plus-outlined /> 开通账号
          </a-button>
        </a-space>
      </template>

      <a-table
        :columns="columns"
        :data-source="tenants"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'green' : record.status === 'trial' ? 'blue' : 'default'">
              {{ record.status === 'active' ? '使用中' : record.status === 'trial' ? '试用中' : '已过期' }}
            </a-tag>
          </template>
          <template v-if="column.key === 'plan'">
            <a-tag :color="record.plan === 'enterprise' ? 'purple' : record.plan === 'pro' ? 'blue' : 'default'">
              {{ record.plan === 'enterprise' ? '企业版' : record.plan === 'pro' ? '专业版' : '免费版' }}
            </a-tag>
          </template>
          <template v-if="column.key === 'user_count'">
            {{ record.user_count }} 人
          </template>
          <template v-if="column.key === 'action'">
            <a-space>
              <a @click="viewTenant(record)">详情</a>
              <a @click="editTenant(record)">编辑</a>
              <a-popconfirm title="确定删除该客户吗？" @confirm="deleteTenant(record.id)">
                <a style="color: #FF6B6B">删除</a>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- ─── 开通账号弹窗 ─── -->
    <a-modal v-model:open="showCreate" title="开通客户账号" :footer="null" :width="520" destroyOnClose>
      <a-form :model="createForm" layout="vertical" @finish="handleCreate">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="公司名称" name="company_name" :rules="[{ required: true, message: '请输入' }]">
              <a-input v-model:value="createForm.company_name" placeholder="公司名称" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="管理员姓名" name="admin_name" :rules="[{ required: true, message: '请输入' }]">
              <a-input v-model:value="createForm.admin_name" placeholder="管理员姓名" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="管理员用户名" name="admin_username" :rules="[{ required: true, message: '请输入' }]">
          <a-input v-model:value="createForm.admin_username" placeholder="用于登录的用户名" />
        </a-form-item>
        <a-form-item label="管理员密码" name="admin_password" :rules="[{ required: true, min: 6, message: '密码至少6位' }]">
          <a-input-password v-model:value="createForm.admin_password" placeholder="设置登录密码" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="手机号">
              <a-input v-model:value="createForm.admin_phone" placeholder="选填" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="邮箱">
              <a-input v-model:value="createForm.admin_email" placeholder="选填" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="订阅方案" name="plan" :rules="[{ required: true, message: '请选择' }]">
          <a-select v-model:value="createForm.plan" placeholder="选择方案">
            <a-select-option value="free">免费版</a-select-option>
            <a-select-option value="pro">专业版 (¥29/人/月)</a-select-option>
            <a-select-option value="enterprise">企业版 (¥59/人/月)</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="试用天数" v-if="createForm.plan !== 'free'">
          <a-input-number v-model:value="createForm.trial_days" :min="0" :max="90" style="width: 100%" placeholder="0表示无试用期" />
        </a-form-item>
        <a-form-item label="团队人数上限">
          <a-input-number v-model:value="createForm.max_users" :min="1" :max="10000" style="width: 100%" placeholder="默认为5" />
        </a-form-item>
        <a-button type="primary" html-type="submit" size="large" block :loading="creating">
          确认开通
        </a-button>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import request from '@/api'

// ─── 统计 ───
const stats = reactive({
  totalTenants: 0,
  activeTenants: 0,
  totalUsers: 0,
  newThisMonth: 0,
})

// ─── 租户列表 ───
const tenants = ref<any[]>([])
const loading = ref(false)
const searchKeyword = ref('')
const pagination = reactive({ current: 1, pageSize: 10, total: 0 })

const columns = [
  { title: '公司名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '管理员', dataIndex: 'admin_name', key: 'admin' },
  { title: '方案', key: 'plan', width: 100 },
  { title: '状态', key: 'status', width: 100 },
  { title: '用户数', key: 'user_count', width: 80 },
  { title: '到期时间', dataIndex: 'expires_at', key: 'expires', width: 120 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' },
]

async function fetchTenants() {
  loading.value = true
  try {
    const res = await request.get('/admin/tenants', {
      params: {
        keyword: searchKeyword.value,
        page: pagination.current,
        page_size: pagination.pageSize,
      },
    })
    const data = res.data.data
    tenants.value = data.list || []
    pagination.total = data.total || 0
  } catch {
    tenants.value = []
  } finally {
    loading.value = false
  }
}

function handleTableChange(pag: any) {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchTenants()
}

// ─── 开通账号 ───
const showCreate = ref(false)
const creating = ref(false)
const createForm = reactive({
  company_name: '',
  admin_name: '',
  admin_username: '',
  admin_password: '',
  admin_phone: '',
  admin_email: '',
  plan: 'pro',
  trial_days: 14,
  max_users: 5,
})

async function handleCreate() {
  creating.value = true
  try {
    await request.post('/admin/tenants', createForm)
    message.success('账号开通成功')
    showCreate.value = false
    fetchTenants()
    // 重置表单
    Object.assign(createForm, {
      company_name: '', admin_name: '', admin_username: '',
      admin_password: '', admin_phone: '', admin_email: '',
      plan: 'pro', trial_days: 14, max_users: 5,
    })
  } catch { /* handled */ }
  finally { creating.value = false }
}

function viewTenant(record: any) { message.info('查看详情: ' + record.name) }
function editTenant(record: any) { message.info('编辑: ' + record.name) }
async function deleteTenant(id: string) {
  try {
    await request.delete(`/admin/tenants/${id}`)
    message.success('已删除')
    fetchTenants()
  } catch { /* handled */ }
}

async function fetchStats() {
  try {
    const res = await request.get('/admin/stats')
    Object.assign(stats, res.data.data)
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchTenants()
  fetchStats()
})
</script>

<style scoped lang="less">
.admin-page {
  padding: 0;
}
.stats-row {
  margin-top: 16px;
}
</style>
