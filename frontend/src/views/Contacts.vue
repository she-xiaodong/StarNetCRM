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
        <template v-if="column.key === 'referrer_path'">
          <a-tag v-if="!record.referrer_id" color="green" size="small">直接人脉</a-tag>
          <a-tooltip v-else :title="record.referrer_path_text">
            <span class="referrer-path-text">{{ record.referrer_path_text || '—' }}</span>
          </a-tooltip>
        </template>
        <template v-if="column.key === 'relation'">
          <div v-if="record._relation" class="relation-cell">
            <a-tooltip :title="strengthDesc(record._relation.strength)">
              <span class="stars">{{ renderStars(record._relation.strength) }}</span>
              <span class="strength-num">{{ record._relation.strength }}</span>
            </a-tooltip>
            <a-tag
              v-for="(tag, idx) in relationTagList(record._relation)"
              :key="idx"
              color="purple"
              size="small"
              style="margin-left: 4px; margin-bottom: 2px"
            >
              {{ tag }}
            </a-tag>
          </div>
          <span v-else style="color: #ccc">-</span>
        </template>
        <template v-if="column.key === 'created_at'">
          {{ formatDateTime(record.created_at) }}
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
          <a-col :span="12">
            <a-form-item label="标签">
              <a-select
                v-model:value="form.tags"
                mode="multiple"
                placeholder="选择标签"
                :options="tagOptions"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="引荐人">
              <a-select
                v-model:value="form.referrer_id"
                placeholder="选择引荐人（留空=直接人脉）"
                allow-clear
                show-search
                option-filter-prop="label"
                :options="referrerOptions"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="备注">
              <a-textarea v-model:value="form.note" :rows="3" placeholder="请输入备注信息" />
            </a-form-item>
          </a-col>
        </a-row>

        <!-- 关系信息区块：亲密度 + 关系标签 都属于"我↔TA"的关系 -->
        <a-divider plain style="margin: 8px 0 16px">关系信息（亲密度与标签属于你与TA的关系）</a-divider>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="关系类型">
              <a-select v-model:value="form.relation_type" placeholder="选择与TA的关系" allow-clear>
                <a-select-option v-for="(label, val) in RELATION_TYPES" :key="val" :value="val">
                  {{ label }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="关系标签">
              <a-select
                v-model:value="form.relation_tags"
                mode="multiple"
                placeholder="打在关系上的标签"
                :options="relationTagOptions"
              />
            </a-form-item>
          </a-col>
          <a-col :span="24">
            <a-form-item label="亲密度">
              <a-slider
                v-model:value="form.relation_strength"
                :min="1"
                :max="10"
                :marks="STRENGTH_MARKS"
              />
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
import { getContacts, createContact, updateContact, deleteContact, getTags, getRelations } from '@/api'
import { formatDateTime } from '@/utils/format'

// 关系类型（后端 type 字段取值 + 中文名）
const RELATION_TYPES: Record<string, string> = {
  colleague: '同事',
  manager: '上下级',
  customer: '客户',
  partner: '合作伙伴',
  alumni: '校友',
  friend: '朋友',
  referral: '引荐',
  custom: '自定义',
}
// 亲密度滑杆刻度文案
const STRENGTH_MARKS = {
  1: '点头之交',
  3: '一般',
  5: '熟识',
  8: '亲密',
  10: '挚友',
}
// 关系标签候选（打在关系上，不继承联系人标签）
const RELATION_TAG_OPTIONS = [
  '深交', '生意往来', '合作过', '可引荐', '潜在客户', '老客户',
  '前同事', '家人', '邻居', '同乡', '同好', '导师', '学生', 'VIP',
]
const relationTagOptions = RELATION_TAG_OPTIONS.map((name) => ({ label: name, value: name }))

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
  { title: '引荐路径', key: 'referrer_path', width: 240 },
  { title: '标签', key: 'tags', width: 160 },
  { title: '关系', key: 'relation', width: 220 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at' },
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
  referrer_id: '' as string,
  note: '',
  relation_type: '' as string,
  relation_tags: [] as string[],
  relation_strength: 5 as number,
})

// 引荐人候选（全量联系人，排除自己）
const referrerOptions = ref<{ label: string; value: string }[]>([])

async function fetchReferrerCandidates() {
  try {
    const res = await getContacts({ page: 1, page_size: 500 })
    const list = res.data.data.list || []
    referrerOptions.value = list
      .filter((c: any) => !editingContact.value || c.id !== editingContact.value.id)
      .map((c: any) => ({ label: `${c.name}${c.company ? '（' + c.company + '）' : ''}`, value: c.person_id || c.id }))
  } catch {
    referrerOptions.value = []
  }
}

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
    let list: any[] = []
    if (data.list) {
      list = data.list
      pagination.total = data.total
    } else if (Array.isArray(data)) {
      list = data
      pagination.total = data.length
    }
    contacts.value = list
    attachRelations(list)
  } catch {
    contacts.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

// 拉取当前用户的所有关系，附加到联系人行（record._relation）
async function attachRelations(list: any[]) {
  if (!list.length) return
  try {
    const res = await getRelations({ page: 1, page_size: 1000 })
    const relList: any[] = res.data.data?.list || []
    const relMap: Record<string, any> = {}
    for (const rel of relList) {
      relMap[rel.to_person_id] = rel
    }
    for (const item of list) {
      const personId = item.person_id || item.id
      if (personId && relMap[personId]) {
        item._relation = relMap[personId]
      } else {
        item._relation = null
      }
    }
  } catch {
    // 关系加载失败不影响列表
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
    phone: '', email: '', tags: [], referrer_id: '', note: '',
    relation_type: '', relation_tags: [], relation_strength: 5,
  })
  fetchReferrerCandidates()
  showModal.value = true
}

function handleEdit(record: any) {
  editingContact.value = record
  const rel = record._relation || {}
  Object.assign(form, {
    name: record.name || '',
    company: record.company || '',
    title: record.title || '',
    department: record.department || '',
    phone: record.phone || '',
    email: record.email || '',
    tags: parseTags(record.tags),
    referrer_id: record.referrer_id || '',
    note: record.note || '',
    relation_type: rel.type || record.relation_type || '',
    relation_tags: rel.tags || [],
    relation_strength: rel.strength || 5,
  })
  fetchReferrerCandidates()
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning('请输入姓名')
    return
  }
  saving.value = true
  try {
    const payload: Record<string, any> = {
      name: form.name.trim(),
      company: form.company,
      title: form.title,
      department: form.department,
      phone: form.phone,
      email: form.email,
      tags: form.tags,
      referrer_id: form.referrer_id || '',
      note: form.note,
    }
    // 仅当用户填写了关系信息才传给后端（避免误建"朋友/5"默认关系）
    if (form.relation_type || form.relation_tags.length) {
      payload.relation_type = form.relation_type || 'friend'
      payload.relation_tags = form.relation_tags
      payload.relation_strength = form.relation_strength
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

// 亲密度 → 星级展示
function renderStars(strength: number): string {
  const s = Math.max(1, Math.min(10, Number(strength) || 5))
  const full = Math.round(s / 2)
  return '★'.repeat(full) + '☆'.repeat(5 - full)
}
// 亲密度等级描述
function strengthDesc(strength: number): string {
  const map: Record<number, string> = {
    1: '点头之交', 2: '泛泛之交', 3: '一般', 4: '较熟', 5: '熟识',
    6: '良好', 7: '亲近', 8: '亲密', 9: '非常亲密', 10: '挚友',
  }
  const s = Math.max(1, Math.min(10, Number(strength) || 5))
  return map[s] || `亲密度 ${s}`
}
// 关系标签列表
function relationTagList(rel: any): string[] {
  if (!rel) return []
  return Array.isArray(rel.tags) ? rel.tags : []
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
  .referrer-path-text {
    color: #2b5fd7;
    font-size: 13px;
  }
  .text-danger {
    color: #ff4d4f;
  }
}
</style>
