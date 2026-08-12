<template>
  <div class="graph-page" ref="wrapperRef">
    <div class="graph-toolbar">
      <h3>人脉关系图谱</h3>
      <a-space>
        <a-button size="small" @click="handleRefresh" :loading="loading">刷新</a-button>
        <a-button size="small" @click="handleResetView">重置视图</a-button>
      </a-space>
    </div>

    <div v-if="nodata" class="no-data-hint">
      暂无关系数据，请先 <router-link to="/contacts">添加联系人</router-link>
    </div>

    <div ref="chartRef" class="chart-container" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { getFirstDegree } from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const chartRef = ref<HTMLDivElement>()
const wrapperRef = ref<HTMLDivElement>()
const loading = ref(false)
const nodata = ref(false)
let chart: echarts.ECharts | null = null

const COLORS = {
  currentUser: '#2B5FD7',
  contacts: '#52C41A',
}

async function fetchGraph() {
  loading.value = true
  nodata.value = false
  try {
    const res = await getFirstDegree()
    const data = res.data.data || {}
    const nodes: any[] = data.nodes || []
    const edges: any[] = data.edges || []

    if (nodes.length === 0 || edges.length === 0) {
      nodata.value = true
      renderGraph([], [])
      return
    }

    renderGraph(nodes, edges)
  } catch {
    nodata.value = true
  } finally {
    loading.value = false
  }
}

function renderGraph(apiNodes: any[], apiEdges: any[]) {
  const userInfo = auth.userInfo
  const myUserId = userInfo?.id || userInfo?.user_id || 'me'
  const myUserName = userInfo?.name || userInfo?.username || '我'

  // 构建节点map：id → properties
  const nodeMap = new Map<string, any>()
  for (const n of apiNodes) {
    const props = n.properties || {}
    nodeMap.set(n.id, {
      id: n.id,
      name: props.name || n.id,
      company: props.company || '',
      title: props.title || '',
      department: props.department || '',
    })
  }

  // 构建 ECharts 节点：中心节点放第一个
  const chartNodes: any[] = []

  // 中心节点（当前用户）
  const isMeInGraph = nodeMap.has(myUserId)
  if (isMeInGraph || nodeMap.size > 0) {
    chartNodes.push({
      id: myUserId,
      name: myUserName,
      symbolSize: 50,
      category: 0,
      fixed: false,
      itemStyle: {
        color: COLORS.currentUser,
        shadowBlur: 12,
        shadowColor: COLORS.currentUser,
      },
      label: { show: true, formatter: `【我】\n${myUserName}`, fontSize: 13, fontWeight: 'bold' },
    })
  }

  // 联系人节点
  for (const [, info] of nodeMap) {
    if (info.id === myUserId) continue
    chartNodes.push({
      id: info.id,
      name: info.name,
      symbolSize: 28,
      category: 1,
      itemStyle: { color: COLORS.contacts },
      label: { show: true, fontSize: 11 },
      company: info.company,
      title: info.title,
    })
  }

  // 构建 ECharts 边
  const chartEdges = apiEdges.map((e: any) => ({
    source: e.source,
    target: e.target,
    label: { show: false },
    lineStyle: {
      color: COLORS.contacts,
      width: 1.5,
      curveness: 0.2,
    },
  }))

  if (!chart) {
    chart = echarts.init(chartRef.value!)
  }

  const nodata = chartNodes.length <= 1

  chart.setOption({
    backgroundColor: '#fff',
    tooltip: {
      formatter: (params: any) => {
        if (params.dataType === 'node' && params.data.name) {
          const d = params.data
          let tip = `<b>${d.name}</b>`
          if (d.company) tip += `<br/>公司: ${d.company}`
          if (d.title) tip += `<br/>职位: ${d.title}`
          return tip
        }
        return ''
      },
    },
    legend: nodata
      ? undefined
      : {
          data: [
            { name: '我', icon: 'circle' },
            { name: '联系人', icon: 'circle' },
          ],
          bottom: 10,
        },
    series: [
      {
        type: 'graph',
        layout: 'force',
        roam: true,
        draggable: true,
        force: {
          repulsion: 260,
          gravity: 0.08,
          edgeLength: [120, 260],
          layoutAnimation: true,
        },
        categories: [
          { name: '我', itemStyle: { color: COLORS.currentUser } },
          { name: '联系人', itemStyle: { color: COLORS.contacts } },
        ],
        data: chartNodes,
        links: chartEdges,
        emphasis: {
          focus: 'adjacency',
          lineStyle: { width: 3 },
        },
        scaleLimit: { min: 0.3, max: 5 },
        lineStyle: {
          opacity: 0.6,
          curveness: 0.2,
        },
      },
    ],
  })

  chart.resize()
}

function handleRefresh() {
  fetchGraph()
}

function handleResetView() {
  chart?.dispatchAction({ type: 'restore' })
}

let resizeTimer: number | null = null
function onResize() {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = window.setTimeout(() => chart?.resize(), 150)
}

onMounted(() => {
  fetchGraph()
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  chart?.dispose()
})
</script>

<style lang="less" scoped>
.graph-page {
  height: calc(100vh - 140px);
  display: flex;
  flex-direction: column;

  .graph-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    flex-shrink: 0;
    h3 {
      margin: 0;
    }
  }

  .chart-container {
    flex: 1;
    min-height: 0;
    border: 1px solid #f0f0f0;
    border-radius: 8px;
    overflow: hidden;
  }

  .no-data-hint {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 15px;
    a {
      color: #2b5fd7;
    }
  }
}
</style>
