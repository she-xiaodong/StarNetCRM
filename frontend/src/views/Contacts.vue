<template>
  <div class="contacts-page">
    <!-- 工具栏 -->
    <div class="toolbar">
      <a-space>
        <a-input-search
          v-model:value="keyword"
          placeholder="搜索姓名、公司、职位..."
          style="width: 280px"
          @search="handleSearch"
        />
      </a-space>
      <a-space>
        <a-button type="primary" @click="openCreate">
          <template #icon><PlusOutlined /></template>
          添加联系人
        </a-button>
      </a-space>
    </div>

    <!-- 联系人表格 -->
    <a-table
      :columns="columns"
      :data-source="contacts"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <div class="contact-name-cell">
            <a-avatar size="small" :style="{ backgroundColor: '#2B5FD7' }">
              {{ record.name.charAt(0) }}
            </a-avatar>
            <span style="margin-left: 8px; font-weight: 500">{{ record.name }}</span>
          </div>
        </template>
        <template v-if="column.key === 'tags'">
          <a-tag
            v-for="(tag, idx) in parseTags(record.tags)"
            :key="idx"
            color="blue"
            size="small"
            style="margin-bottom: 2px"
          >
            {{ tag }}
          </a-tag>
          <span v-if="!parseTags(record.tags).length" style="color: #ccc">-</span>
        </template>
        <template v-if="column.key === 'action'">
          <a-space>
            <a @click="handleEdit(record)">编辑</a>
            <a-popconfirm
              title="确定删除此联系人？"
              @confirm="handleDelete(record)"
            >
              <a class="text-danger">删除</a>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 创建/编辑弹窗 -->
    <a-modal
      v-model:open="showModal"
      :title="editingContact ? '编辑联系人' : '添加联系人'"
      width="600px"
      :confirm-loading="saving"
      @ok="handleSave"
      @cancel="resetForm"
    >
      <a-form :model="form" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="姓名" required>
              <a-input v-model:value="form.name" placeholder="请输入姓名" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="公司">
              <a-input v-model:value="form.company" placeholder="请输入公司" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="职位">
              <a-input v-model:value="form.title" placeholder="请输入职位" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="部门">
              <a-input v-model:value="form.department" placeholder="请输入部门" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="手机号">
              <a-input v-model:value="form.phone" placeholder="请输入手机号" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="邮箱">
              <a-input v-model:value="form.email" placeholder="请输入邮箱" />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="标签">
              <a-select
                v-model:value="form.tags"
                mode="multiple"
                placeholder="选择标签"
                :options="tagOptions"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="备注">
              <a-textarea v-model:value="form.note" :rows="3" placeholder="请输入备注信息" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { getContacts, createContact, updateContact, deleteContact, getTags } from '@/api'

const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const showModal = ref(false)
const editingContact = ref<any>(null)
const tagOptions = ref<{ label: string; value: string }[]>([])

const columns = [
  { title: '姓名', key: 'name' },
  { title: '公司', dataIndex: 'company' },
  { title: '职位', dataIndex: 'title' },
  { title: '部门', dataIndex: 'department' },
  { title: '手机号', dataIndex: 'phone' },
  { title: '标签', key: 'tags', width: 160 },
  { title: '创建时间', dataIndex: 'created_at' },
  { title: '操作', key: 'action', width: 140 },
]

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const contacts = ref<any[]>([])

const form = reactive({
  name: '',
  company: '',
  title: '',
  department: '',
  phone: '',
  email: '',
  tags: [] as string[],
  note: '',
})

// 解析 tags 字段（可能是 JSON 字符串或数组）
function parseTags(tags: any): string[] {
  if (!tags) return []
  if (Array.isArray(tags)) return tags
  if (typeof tags === 'string') {
    try { return JSON.parse(tags) } catch { return tags ? [tags] : [] }
  }
  return []
}

async function fetchContacts(page = 1) {
  loading.value = true
  try {
    const res = await getContacts({
      keyword: keyword.value || undefined,
      page,
      page_size: pagination.pageSize,
    })
    const data = res.data.data
    if (data.list) {
      contacts.value = data.list
      pagination.total = data.total
    } else if (Array.isArray(data)) {
      contacts.value = data
      pagination.total = data.length
    } else {
      contacts.value = []
      pagination.total = 0
    }
  } catch {
    contacts.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

async function fetchTags() {
  try {
    const res = await getTags()
    const list = res.data.data
    if (Array.isArray(list)) {
      tagOptions.value = list.map((t: any) => ({ label: t.name, value: t.name }))
    }
  } catch {
    // 标签加载失败不影响主流程
  }
}

function handleSearch() {
  pagination.current = 1
  fetchContacts(1)
}

function handleTableChange(pag: any) {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchContacts(pag.current)
}

function openCreate() {
  editingContact.value = null
  Object.assign(form, {
    name: '', company: '', title: '', department: '',
    phone: '', email: '', tags: [], note: '',
  })
  showModal.value = true
}

function handleEdit(record: any) {
  editingContact.value = record
  Object.assign(form, {
    name: record.name || '',
    company: record.company || '',
    title: record.title || '',
    department: record.department || '',
    phone: record.phone || '',
    email: record.email || '',
    tags: parseTags(record.tags),
    note: record.note || '',
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning('请输入姓名')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      company: form.company,
      title: form.title,
      department: form.department,
      phone: form.phone,
      email: form.email,
      tags: form.tags,
      note: form.note,
    }
    if (editingContact.value) {
      await updateContact(editingContact.value.id, payload)
      message.success('更新成功')
    } else {
      await createContact(payload)
      message.success('添加成功')
    }
    showModal.value = false
    fetchContacts(pagination.current)
  } catch {
    // 错误已在拦截器处理
  } finally {
    saving.value = false
  }
}

async function handleDelete(record: any) {
  try {
    await deleteContact(record.id)
    message.success(`已删除联系人: ${record.name}`)
    fetchContacts(pagination.current)
  } catch {
    // 错误已在拦截器处理
  }
}

function resetForm() {
  editingContact.value = null
}

onMounted(() => {
  fetchContacts()
  fetchTags()
})
</script>

<style lang="less" scoped>
.contacts-page {
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 12px;
  }
  .contact-name-cell {
    display: flex;
    align-items: center;
  }
  .text-danger {
    color: #ff4d4f;
  }
}
</style>
