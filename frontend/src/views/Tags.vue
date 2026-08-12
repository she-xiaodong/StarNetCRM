<template>
  <div class="tags-page">
    <div class="page-header">
      <h3>标签库管理</h3>
      <a-button type="primary" @click="openCreate">
        <template #icon><PlusOutlined /></template>
        新建标签
      </a-button>
    </div>

    <a-alert
      message="标签用于描述人脉关系类型（如：同学、同事、投资人、合作伙伴等），创建联系人时可选择关联标签。"
      type="info"
      show-icon
      style="margin-bottom: 16px"
    />

    <a-table :columns="columns" :data-source="tags" :loading="loading" row-key="id" size="middle">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <a-tag :color="record.color || '#2B5FD7'">{{ record.name }}</a-tag>
        </template>
        <template v-if="column.key === 'action'">
          <a-space>
            <a @click="handleEdit(record)">编辑</a>
            <a-popconfirm title="确定删除此标签？关联的联系人不会受影响。" @confirm="doDelete(record.id)">
              <a class="text-danger">删除</a>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 弹窗 -->
    <a-modal
      v-model:open="showModal"
      :title="editing ? '编辑标签' : '新建标签'"
      :confirm-loading="saving"
      @ok="handleSave"
    >
      <a-form layout="vertical">
        <a-form-item label="标签名称" required>
          <a-input v-model:value="form.name" placeholder="如：同学、同事、投资人" />
        </a-form-item>
        <a-form-item label="类型">
          <a-select v-model:value="form.type" placeholder="选择类型">
            <a-select-option value="relationship">关系标签</a-select-option>
            <a-select-option value="attribute">属性标签</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="颜色">
          <div class="color-picker-row">
            <span
              v-for="c in presetColors"
              :key="c"
              class="color-swatch"
              :class="{ active: form.color === c }"
              :style="{ backgroundColor: c }"
              @click="form.color = c"
            />
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { getTags, createTag, updateTag, deleteTag } from '@/api'

const loading = ref(false)
const saving = ref(false)
const showModal = ref(false)
const editing = ref<any>(null)
const tags = ref<any[]>([])

const presetColors = ['#2B5FD7', '#52C41A', '#FA8C16', '#F5222D', '#722ED1', '#13C2C2', '#EB2F96', '#A0D911']

const columns = [
  { title: '标签名称', key: 'name', width: 200 },
  { title: '类型', dataIndex: 'type', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 140 },
]

const form = reactive({ name: '', color: '#2B5FD7', type: 'relationship' })

async function fetchList() {
  loading.value = true
  try {
    const res = await getTags()
    tags.value = res.data.data || []
  } catch {
    tags.value = []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', color: '#2B5FD7', type: 'relationship' })
  showModal.value = true
}

function handleEdit(record: any) {
  editing.value = record
  Object.assign(form, {
    name: record.name || '',
    color: record.color || '#2B5FD7',
    type: record.type || 'relationship',
  })
  showModal.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    message.warning('请输入标签名称')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateTag(editing.value.id, { ...form })
      message.success('更新成功')
    } else {
      await createTag({ ...form })
      message.success('创建成功')
    }
    showModal.value = false
    fetchList()
  } catch {
    // 已在拦截器处理
  } finally {
    saving.value = false
  }
}

async function doDelete(id: string) {
  try {
    await deleteTag(id)
    message.success('已删除')
    fetchList()
  } catch {
    // 已处理
  }
}

onMounted(fetchList)
</script>

<style lang="less" scoped>
.tags-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    h3 {
      margin: 0;
    }
  }
  .color-picker-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .color-swatch {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    cursor: pointer;
    border: 2px solid transparent;
    transition: all .15s;
    &.active {
      border-color: #333;
      transform: scale(1.15);
    }
    &:hover {
      opacity: 0.8;
    }
  }
  .text-danger {
    color: #ff4d4f;
  }
}
</style>
