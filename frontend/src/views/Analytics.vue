<template>
  <div class="analytics-page">
    <h3 class="page-title">人脉分析</h3>

    <!-- 核心指标 -->
    <a-row :gutter="16" class="metrics-row">
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="活跃关系链数"
            :value="metrics.active_relations"
            suffix="条"
            value-style="color: #2B5FD7"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="本周新增联系人"
            :value="metrics.week_new_contacts"
            suffix="人"
            value-style="color: #52C41A"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="平均路径长度"
            :value="metrics.avg_path_length"
            precision="1"
            suffix="度"
            value-style="color: #FA8C16"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="孤立节点"
            :value="metrics.isolated_nodes"
            suffix="人"
            value-style="color: #FF6B6B"
          />
        </a-card>
      </a-col>
    </a-row>

    <!-- 图表区域 -->
    <a-row :gutter="16" class="charts-row">
      <!-- 关系类型分布 -->
      <a-col :span="12">
        <a-card title="关系类型分布" size="small">
          <div class="chart-placeholder">
            <div class="bar-item" v-for="item in relationDistribution" :key="item.type">
              <span class="bar-label">{{ item.label }}</span>
              <div class="bar-track">
                <div class="bar-fill" :style="{ width: item.percent + '%', background: item.color }"></div>
              </div>
              <span class="bar-value">{{ item.count }} ({{ item.percent }}%)</span>
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 人脉增长趋势 -->
      <a-col :span="12">
        <a-card title="人脉增长趋势" size="small">
          <div class="chart-placeholder trend-chart">
            <div class="trend-bars">
              <div
                v-for="item in growthData"
                :key="item.month"
                class="trend-column"
              >
                <div
                  class="trend-bar"
                  :style="{ height: barHeight(item.count) + '%' }"
                ></div>
                <span class="trend-label">{{ item.month }}</span>
                <span class="trend-value">{{ item.count }}</span>
              </div>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- 超级连接者 -->
    <a-card title="超级连接者" size="small" class="connectors-card">
      <a-table
        :columns="superConnectorColumns"
        :data-source="superConnectors"
        row-key="id"
        size="small"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <a-avatar size="small" :style="{ backgroundColor: '#2B5FD7' }">
              {{ record.name.charAt(0) }}
            </a-avatar>
            <span style="margin-left: 8px">{{ record.name }}</span>
          </template>
          <template v-if="column.key === 'degree'">
            <a-progress
              :percent="degreePercent(record.degree)"
              :size="20"
              :showInfo="false"
            />
            <span style="margin-left: 8px; font-size: 13px">{{ record.degree }} 条关系</span>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAnalytics } from '../api'

const relationColorPalette = ['#4A90D9', '#52C41A', '#FA8C16', '#722ED1', '#8C8C8C', '#2B5FD7', '#EB2F96', '#13C2C2']

const metrics = ref({
  active_relations: 0,
  week_new_contacts: 0,
  avg_path_length: 0,
  isolated_nodes: 0,
})

const relationDistribution = ref<{ type: string; label: string; count: number; percent: number; color: string }[]>([])
const growthData = ref<{ month: string; count: number }[]>([])

const superConnectorColumns = [
  { title: '姓名', key: 'name' },
  { title: '公司', dataIndex: 'company' },
  { title: '关系数', key: 'degree', width: 250 },
  { title: '最有价值的连接', dataIndex: 'top_connection' },
]

const superConnectors = ref<{ id: string; name: string; company?: string; degree: number; top_connection?: string }[]>([])

const maxDegree = ref(0)
const maxTrend = ref(0)

function barHeight(count: number) {
  if (!maxTrend.value || !count) return 4
  return Math.max(4, Math.round((count / maxTrend.value) * 100))
}

function degreePercent(degree: number) {
  if (!maxDegree.value || !degree) return 0
  return Math.round((degree / maxDegree.value) * 100)
}

function parseTags(tags?: string): string[] {
  if (!tags) return []
  try {
    const arr = JSON.parse(tags)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function loadAnalytics() {
  getAnalytics().then((res: any) => {
    const data = res.data?.data
    if (!data) return
    metrics.value = {
      active_relations: data.active_relations ?? 0,
      week_new_contacts: data.week_new_contacts ?? 0,
      avg_path_length: data.avg_path_length ?? 0,
      isolated_nodes: data.isolated_nodes ?? 0,
    }
    relationDistribution.value = (data.relation_distribution || []).map(
      (item: any, idx: number) => ({
        type: item.type,
        label: item.label,
        count: item.count,
        percent: item.percent,
        color: relationColorPalette[idx % relationColorPalette.length],
      }),
    )
    growthData.value = data.growth_trend || []
    maxTrend.value = Math.max(1, ...growthData.value.map((i) => i.count))
    superConnectors.value = data.super_connectors || []
    maxDegree.value = Math.max(1, ...superConnectors.value.map((i) => i.degree ?? 0))
  })
}

onMounted(loadAnalytics)
</script>

<style lang="less" scoped>
.analytics-page {
  .page-title {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 20px;
  }

  .metrics-row {
    margin-bottom: 16px;
  }

  .charts-row {
    margin-bottom: 16px;
  }

  .chart-placeholder {
    .bar-item {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;

      .bar-label {
        width: 80px;
        font-size: 13px;
        color: #666;
        text-align: right;
      }

      .bar-track {
        flex: 1;
        height: 20px;
        background: #f0f0f0;
        border-radius: 10px;
        overflow: hidden;

        .bar-fill {
          height: 100%;
          border-radius: 10px;
          transition: width 0.6s ease;
        }
      }

      .bar-value {
        width: 80px;
        font-size: 13px;
        color: #333;
      }
    }
  }

  .trend-chart {
    .trend-bars {
      display: flex;
      align-items: flex-end;
      gap: 16px;
      height: 180px;
      padding: 8px 0;

      .trend-column {
        display: flex;
        flex-direction: column;
        align-items: center;
        flex: 1;

        .trend-bar {
          width: 100%;
          max-width: 40px;
          background: linear-gradient(to top, #2B5FD7, #4A90D9);
          border-radius: 4px 4px 0 0;
          min-height: 8px;
          transition: height 0.6s ease;
        }

        .trend-label {
          font-size: 12px;
          color: #999;
          margin-top: 4px;
        }

        .trend-value {
          font-size: 12px;
          color: #333;
          font-weight: 600;
        }
      }
    }
  }

  .connectors-card {
    border-radius: 12px;
  }
}
</style>
