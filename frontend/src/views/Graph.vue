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
      暂无关系数据，请先 <router-link to="/app/contacts">添加联系人</router-link>
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
let revealTimers: number[] = []
let centerNodeRef: any = null
let currentVisibleNodes: any[] = []
let currentVisibleEdges: any[] = []

// 引荐层级颜色：0=我 1=一级 2=二级 3=三级 4=四级及以上
const LEVEL_COLORS = ['#2B5FD7', '#52C41A', '#FA8C16', '#722ED1', '#EB2F96']
const LEVEL_NAMES = ['我', '一级人脉', '二级人脉', '三级人脉', '四级+']
const LEVEL_SIZES = [52, 34, 28, 24, 22]
// 分批出现的间隔(ms)：中心 → 一级 → 二级 → 三级 → 四级
const REVEAL_INTERVAL = 600

function getLevel(props: any): number {
  const lv = Number(props?.level)
  if (Number.isNaN(lv) || lv < 0) return 1
  return Math.min(4, lv)
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

  // 清理上一轮的入场定时器
  revealTimers.forEach((t) => clearTimeout(t))
  revealTimers = []

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
      level: getLevel(props),
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
      symbolSize: LEVEL_SIZES[0],
      category: 0,
      fixed: false,
      itemStyle: {
        color: LEVEL_COLORS[0],
        shadowBlur: 12,
        shadowColor: LEVEL_COLORS[0],
      },
      label: { show: true, formatter: `【我】\n${myUserName}`, fontSize: 13, fontWeight: 'bold' },
    })
  }

  // 联系人节点：按引荐层级分级着色、调整大小。
  // 标签策略：一级人脉显示名字；二级及以上默认隐藏，hover/点击时显示，避免大图文字互相遮挡
  for (const [, info] of nodeMap) {
    if (info.id === myUserId) continue
    const lv = info.level
    chartNodes.push({
      id: info.id,
      name: info.name,
      symbolSize: LEVEL_SIZES[lv],
      category: lv,
      itemStyle: { color: LEVEL_COLORS[lv] },
      label: lv <= 1 ? { show: true, fontSize: 10 } : { show: false },
      emphasis: {
        itemStyle: { borderColor: '#333', borderWidth: 2 },
        label: { show: true, fontSize: 12, fontWeight: 'bold' },
      },
      company: info.company,
      title: info.title,
      levelName: LEVEL_NAMES[lv],
    })
  }

  // 构建 ECharts 边：默认不显示文字，hover 时显示"引荐"关系，避免满屏文字遮挡节点
  const chartEdges = apiEdges.map((e: any) => {
    const isReferral = e.label === '引荐' || e.type === 'referral'
    return {
      source: e.source,
      target: e.target,
      label: { show: false },
      emphasis: {
        label: isReferral
          ? { show: true, formatter: '引荐', fontSize: 10, color: '#666', backgroundColor: '#fff', padding: [2, 4] }
          : { show: false },
      },
      lineStyle: {
        color: '#bfbfbf',
        width: 1.5,
        curveness: 0.2,
      },
    }
  })

  if (!chart) {
    chart = echarts.init(chartRef.value!)
  }
  chart.resize()
  const hasMulti = chartNodes.length > 1
  const cx = chart.getWidth() / 2
  const cy = chart.getHeight() / 2

  // 中心节点固定在画布正中央
  centerNodeRef = chartNodes[0] || null
  if (centerNodeRef) {
    centerNodeRef.fixed = true
    centerNodeRef.x = cx
    centerNodeRef.y = cy
  }

  // 节点按层级分组（不含中心）
  const byLevel: Record<number, any[]> = {}
  for (const n of chartNodes) {
    if (n.category === 0) continue
    ;(byLevel[n.category] = byLevel[n.category] || []).push(n)
  }

  // 边按两端节点中较高的层级归组，随该层节点一起出现
  const idLevel = new Map<string, number>()
  for (const n of chartNodes) idLevel.set(n.id, n.category)
  const edgesByLevel: Record<number, any[]> = {}
  for (const e of chartEdges) {
    const sId = typeof e.source === 'string' ? e.source : e.source?.id
    const tId = typeof e.target === 'string' ? e.target : e.target?.id
    const lv = Math.max(idLevel.get(sId) ?? 0, idLevel.get(tId) ?? 0)
    ;(edgesByLevel[lv] = edgesByLevel[lv] || []).push(e)
  }

  // 基础图表配置
  const baseOption: any = {
    backgroundColor: '#fff',
    tooltip: {
      formatter: (params: any) => {
        if (params.dataType === 'node' && params.data.name) {
          const d = params.data
          let tip = `<b>${d.name}</b>`
          if (d.levelName) tip += `<br/>层级: ${d.levelName}`
          if (d.company) tip += `<br/>公司: ${d.company}`
          if (d.title) tip += `<br/>职位: ${d.title}`
          return tip
        }
        if (params.dataType === 'edge' && params.data.label?.formatter) {
          return `引荐关系`
        }
        return ''
      },
    },
    legend: hasMulti
      ? {
          data: LEVEL_NAMES.map((name) => ({ name, icon: 'circle' })),
          bottom: 10,
        }
      : undefined,
    series: [
      {
        type: 'graph',
        layout: 'force',
        roam: true,
        draggable: true,
        force: {
          // 节点多时适当增大斥力、缩短边长，让大图能够完整舒展开而不堆积
          repulsion: chartNodes.length > 200 ? 340 : 260,
          gravity: 0.05,
          edgeLength: [90, 200],
          layoutAnimation: true,
        },
        categories: LEVEL_NAMES.map((name, i) => ({
          name,
          itemStyle: { color: LEVEL_COLORS[i] },
        })),
        emphasis: {
          focus: 'adjacency',
          lineStyle: { width: 3 },
        },
        scaleLimit: { min: 0.3, max: 5 },
        lineStyle: {
          opacity: 0.6,
          curveness: 0.2,
        },
        animationDurationUpdate: 700,
        animationEasingUpdate: 'cubicOut',
      },
    ],
  }

  // 首次渲染：只显示中心节点，随后按层级分批浮现
  const centerOnly: any[] = centerNodeRef ? [centerNodeRef] : []
  baseOption.series[0].data = centerOnly
  baseOption.series[0].links = []
  chart.setOption(baseOption)

  // 分批：一级 → 二级 → 三级 → 四级，每次 setOption 增量出现。
  // 新节点预置于中心外侧的圆环上（层级越深半径越大），避免 300+ 节点全部从中心点爆开而挤成一团
  const RING_RADIUS: Record<number, number> = { 1: 240, 2: 360, 3: 480, 4: 600 }
  const RING_SEED: Record<number, number> = { 1: 0, 2: Math.PI / 4, 3: Math.PI / 8, 4: Math.PI / 3 }
  const LEVEL_ORDER = [1, 2, 3, 4]
  currentVisibleNodes = [...centerOnly]
  currentVisibleEdges = []
  let delay = 0
  for (const lv of LEVEL_ORDER) {
    const nodes = byLevel[lv] || []
    if (!nodes.length) continue
    delay += REVEAL_INTERVAL
    const timer = window.setTimeout(() => {
      const radius = RING_RADIUS[lv] || 320
      const seed = RING_SEED[lv] || 0
      nodes.forEach((n, i) => {
        const angle = seed + (i / nodes.length) * Math.PI * 2
        n.x = cx + Math.cos(angle) * radius
        n.y = cy + Math.sin(angle) * radius
      })
      currentVisibleNodes.push(...nodes)
      const lvEdges = edgesByLevel[lv] || []
      for (const e of lvEdges) {
        if (!currentVisibleEdges.includes(e)) currentVisibleEdges.push(e)
      }
      chart?.setOption({
        series: [
          {
            data: currentVisibleNodes,
            links: currentVisibleEdges,
          },
        ],
      })
    }, delay)
    revealTimers.push(timer)
  }

  chart.resize()

  // 所有层级浮现完成后，等力导向布局基本稳定，再按节点实际坐标自动缩放，
  // 保证 300+ 节点的大图完整落在画布内（避免上下/左右溢出被裁掉）
  const fitDelay = delay + 3200
  const fitTimer = window.setTimeout(() => {
    fitView()
  }, fitDelay)
  revealTimers.push(fitTimer)
}

// 根据节点实际布局坐标计算包围盒，自动缩放到画布内完整显示
function fitView() {
  if (!chart || chart.isDisposed()) return
  try {
    const seriesModel = chart.getModel().getSeriesByIndex(0)
    const graph = seriesModel?.getGraph?.()
    if (!graph) return
    const nodes = graph.getNodes()
    if (nodes.length < 2) return
    let minX = Infinity
    let minY = Infinity
    let maxX = -Infinity
    let maxY = -Infinity
    for (const n of nodes) {
      const p = n.getLayout?.()
      const x = p?.x
      const y = p?.y
      if (x == null || y == null || Number.isNaN(x) || Number.isNaN(y)) continue
      minX = Math.min(minX, x)
      maxX = Math.max(maxX, x)
      minY = Math.min(minY, y)
      maxY = Math.max(maxY, y)
    }
    if (maxX <= minX || maxY <= minY) return
    const W = chart.getWidth()
    const H = chart.getHeight()
    const bw = maxX - minX
    const bh = maxY - minY
    // 图已完整落在画布内则无需缩放
    if (bw < W * 0.92 && bh < H * 0.92) return
    const zoom = Math.min(W / (bw + 120), H / (bh + 120))
    chart.setOption({
      series: [
        {
          center: [(minX + maxX) / 2, (minY + maxY) / 2],
          zoom: Math.max(0.3, Math.min(1.2, zoom)),
        },
      ],
    })
  } catch {
    // fitView 为增强功能，失败不影响主流程
  }
}

function handleRefresh() {
  fetchGraph()
}

function handleResetView() {
  chart?.dispatchAction({ type: 'restore' })
  // restore 后重新把中心节点放回画布中央
  if (chart && centerNodeRef) {
    centerNodeRef.x = chart.getWidth() / 2
    centerNodeRef.y = chart.getHeight() / 2
    chart.setOption({
      series: [
        {
          data: currentVisibleNodes,
          links: currentVisibleEdges,
        },
      ],
    })
  }
}

let resizeTimer: number | null = null
function onResize() {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = window.setTimeout(() => {
    chart?.resize()
    // resize 后中心节点保持居中
    if (chart && centerNodeRef) {
      centerNodeRef.x = chart.getWidth() / 2
      centerNodeRef.y = chart.getHeight() / 2
    }
  }, 150)
}

onMounted(() => {
  fetchGraph()
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  revealTimers.forEach((t) => clearTimeout(t))
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
