<template>
  <div class="path-search-page">
    <div class="search-header">
      <h3>六度人脉路径搜索</h3>
      <p class="search-desc">
        基于"六度分隔理论"，在您的人脉网络中寻找两人之间的最短关系路径
      </p>
    </div>

    <!-- 搜索区域 -->
    <a-card class="search-card">
      <a-form layout="inline" class="search-form">
        <a-form-item label="起点人物">
          <a-select
            v-model:value="searchForm.startPersonId"
            placeholder="选择起点联系人"
            style="width: 220px"
            showSearch
          >
            <a-select-option value="me">我自己</a-select-option>
            <a-select-option value="zhangsan">张三</a-select-option>
            <a-select-option value="lisi">李四</a-select-option>
            <a-select-option value="wangwu">王五</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item class="path-arrow">➡️</a-form-item>

        <a-form-item label="目标人物">
          <a-select
            v-model:value="searchForm.endPersonId"
            placeholder="选择目标联系人"
            style="width: 220px"
            showSearch
          >
            <a-select-option value="zhangsan">张三</a-select-option>
            <a-select-option value="zhaoliu">赵六</a-select-option>
            <a-select-option value="zhouba">周八</a-select-option>
            <a-select-option value="wujiu">吴九</a-select-option>
            <a-select-option value="zhengshi">郑十</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="最大深度">
          <a-select v-model:value="searchForm.maxDepth" style="width: 100px">
            <a-select-option :value="3">3度</a-select-option>
            <a-select-option :value="4">4度</a-select-option>
            <a-select-option :value="5">5度</a-select-option>
            <a-select-option :value="6">6度</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item>
          <a-button type="primary" :loading="searching" @click="handleSearch">
            <template #icon><SearchOutlined /></template>
            搜索路径
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 搜索结果 -->
    <div v-if="searchResult" class="result-section">
      <a-alert
        v-if="searchResult.found === false"
        type="info"
        show-icon
        :message="searchResult.message"
        class="result-alert"
      />

      <template v-if="searchResult.found !== false && searchResult.path">
        <!-- 路径可视化 -->
        <a-card title="关系路径" class="path-card">
          <div class="path-chain">
            <div
              v-for="(node, index) in searchResult.path"
              :key="index"
              class="path-item"
            >
              <!-- 人物节点 -->
              <div class="path-person">
                <a-avatar
                  size="large"
                  :style="{
                    backgroundColor: index === 0 ? '#2B5FD7' :
                      index === searchResult.path.length - 1 ? '#F5A623' : '#4A90D9'
                  }"
                >
                  {{ node.person?.name?.charAt(0) || '?' }}
                </a-avatar>
                <div class="person-info">
                  <div class="person-name">
                    {{ node.person?.name || '未知' }}
                    <a-tag v-if="index === 0" color="blue" size="small">起点</a-tag>
                    <a-tag v-if="index === searchResult.path.length - 1" color="gold" size="small">目标</a-tag>
                  </div>
                  <div class="person-company">{{ node.person?.company || '' }}</div>
                </div>
              </div>

              <!-- 关系连线 -->
              <div v-if="index < searchResult.path.length - 1" class="path-connector">
                <div class="connector-line"></div>
                <a-tag color="processing" class="connector-tag">
                  {{ getRelationTypeName(node.relation?.type) }}
                  <span class="strength">亲密度: {{ node.relation?.strength || '5' }}/10</span>
                </a-tag>
                <div class="connector-line"></div>
              </div>
            </div>
          </div>
        </a-card>

        <!-- 路径摘要 -->
        <a-card title="路径摘要" class="summary-card">
          <a-descriptions :column="3" size="small">
            <a-descriptions-item label="路径长度">
              <a-tag :color="getLengthColor(searchResult.length)">
                {{ searchResult.length }} 度分隔
              </a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="整体亲密度">
              <a-rate :value="getStrengthRate(searchResult.strength)" disabled :count="3" />
            </a-descriptions-item>
            <a-descriptions-item label="引荐难度">
              <a-tag :color="searchResult.length <= 3 ? 'green' : 'orange'">
                {{ searchResult.length <= 3 ? '容易' : '中等' }}
              </a-tag>
            </a-descriptions-item>
          </a-descriptions>

          <a-divider style="margin: 12px 0" />

          <div class="path-actions">
            <span>你想通过这条路径进行引荐吗？</span>
            <a-button type="primary" size="small" @click="handleCreateReferral">
              <template #icon><SendOutlined /></template>
              发起引荐
            </a-button>
          </div>
        </a-card>
      </template>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <SearchOutlined style="font-size: 64px; color: #e8e8e8" />
      <p>输入起点和终点，探索你们之间的人脉路径</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { SearchOutlined, SendOutlined } from '@ant-design/icons-vue'

const router = useRouter()

const searching = ref(false)
const searchResult = ref<any>(null)

const searchForm = reactive({
  startPersonId: 'me' as string | undefined,
  endPersonId: undefined as string | undefined,
  maxDepth: 6,
})

const relationTypeMap: Record<string, string> = {
  colleague: '同事',
  customer: '客户',
  partner: '合作伙伴',
  alumni: '校友',
  friend: '朋友',
}

function getRelationTypeName(type?: string) {
  return relationTypeMap[type || ''] || '关联'
}

function getLengthColor(length: number) {
  if (length <= 2) return 'green'
  if (length <= 4) return 'blue'
  return 'orange'
}

function getStrengthRate(strength: string) {
  if (strength === 'strong') return 3
  if (strength === 'medium') return 2
  return 1
}

async function handleSearch() {
  if (!searchForm.startPersonId || !searchForm.endPersonId) {
    message.warning('请选择起点和终点')
    return
  }

  searching.value = true
  // 模拟搜索
  await new Promise((r) => setTimeout(r, 800))

  // 模拟结果
  searchResult.value = {
    found: true,
    length: 3,
    strength: 'medium',
    path: [
      {
        person: { name: '我', company: '新杰汇智', title: '产品经理' },
        degree: 0,
        relation: { type: 'colleague', strength: 9 },
      },
      {
        person: { name: '张三', company: '阿里巴巴', title: '高级工程师' },
        degree: 1,
        relation: { type: 'alumni', strength: 5 },
      },
      {
        person: { name: '赵六', company: '百度集团', title: '总监' },
        degree: 2,
        relation: { type: 'partner', strength: 6 },
      },
      {
        person: { name: '郑十', company: '美团', title: '副总裁' },
        degree: 3,
      },
    ],
  }

  searching.value = false
}

function handleCreateReferral() {
  router.push('/referral')
}
</script>

<style lang="less" scoped>
.path-search-page {
  .search-header {
    margin-bottom: 20px;

    h3 {
      font-size: 20px;
      font-weight: 600;
      margin-bottom: 4px;
    }

    .search-desc {
      color: #999;
      font-size: 13px;
    }
  }

  .search-card {
    border-radius: 12px;
    margin-bottom: 24px;

    .search-form {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;
    }
  }

  .result-section {
    .result-alert {
      margin-bottom: 16px;
    }

    .path-card {
      margin-bottom: 16px;
      border-radius: 12px;
    }

    .path-chain {
      display: flex;
      flex-direction: column;
      align-items: center;
    }

    .path-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      width: 100%;
    }

    .path-person {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 20px;
      background: #fafafa;
      border-radius: 12px;
      width: 320px;
      box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);

      .person-info {
        .person-name {
          font-weight: 600;
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .person-company {
          font-size: 12px;
          color: #999;
          margin-top: 2px;
        }
      }
    }

    .path-connector {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 4px 0;

      .connector-line {
        width: 2px;
        height: 20px;
        background: #d9d9d9;
      }

      .connector-tag {
        margin: 2px 0;

        .strength {
          font-size: 11px;
          color: #999;
          margin-left: 4px;
        }
      }
    }

    .summary-card {
      border-radius: 12px;

      .path-actions {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 14px;
        color: #666;
      }
    }
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 0;
    gap: 16px;

    p {
      color: #999;
      font-size: 14px;
    }
  }
}
</style>
