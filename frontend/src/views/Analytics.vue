<template>
  <div class="analytics-page">
    <h3 class="page-title">人脉分析</h3>

    <!-- 核心指标 -->
    <a-row :gutter="16" class="metrics-row">
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="活跃关系链数"
            :value="356"
            suffix="条"
            value-style="color: #2B5FD7"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="本周新增联系人"
            :value="12"
            suffix="人"
            value-style="color: #52C41A"
          />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card size="small">
          <a-statistic
            title="平均路径长度"
            :value="2.8"
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
            :value="5"
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
                  :style="{ height: (item.count / 50) * 100 + '%' }"
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
              :percent="(record.degree / 50) * 100"
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
import { ref } from 'vue'

const relationDistribution = ref([
  { type: 'colleague', label: '同事', count: 120, percent: 34, color: '#4A90D9' },
  { type: 'customer', label: '客户', count: 98, percent: 27, color: '#52C41A' },
  { type: 'partner', label: '合作伙伴', count: 65, percent: 18, color: '#FA8C16' },
  { type: 'alumni', label: '校友', count: 42, percent: 12, color: '#722ED1' },
  { type: 'friend', label: '朋友/其他', count: 31, percent: 9, color: '#8C8C8C' },
])

const growthData = ref([
  { month: '1月', count: 28 },
  { month: '2月', count: 32 },
  { month: '3月', count: 35 },
  { month: '4月', count: 40 },
  { month: '5月', count: 38 },
  { month: '6月', count: 45 },
  { month: '7月', count: 50 },
  { month: '8月', count: 48 },
])

const superConnectorColumns = [
  { title: '姓名', key: 'name' },
  { title: '公司', dataIndex: 'company' },
  { title: '关系数', key: 'degree', width: 250 },
  { title: '最有价值的连接', dataIndex: 'topConnection' },
]

const superConnectors = ref([
  { id: '1', name: '张三', company: '阿里巴巴', degree: 48, topConnection: '腾讯/百度/华为高管层' },
  { id: '2', name: '王五', company: '字节跳动', degree: 35, topConnection: '投资圈/创业圈核心人物' },
  { id: '3', name: '李四', company: '腾讯科技', degree: 28, topConnection: '互联网产品/技术圈' },
])
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
