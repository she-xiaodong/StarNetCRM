// ─── 星络客 (StarNet) 设计令牌 ───
// 色彩体系（需求文档 6.2 节）
export const COLORS = {
  // 品牌主色
  primary: '#2B5FD7',      // 星络蓝 — 按钮、标题、导航
  success: '#00C9A7',      // 星络绿 — 正向操作、成功状态、活跃关系
  gold: '#F5A623',         // 星络金 — 高亮路径、重要节点、VIP标签
  danger: '#FF6B6B',       // 星络红 — 警告、删除、关系失效
  bg: '#F0F2F5',           // 背景灰 — 页面背景、卡片背景

  // 图谱节点颜色（按关系类型）
  graph: {
    colleague: '#4A90D9',  // 蓝色系 — 同事
    customer: '#52C41A',   // 绿色系 — 客户
    partner: '#FA8C16',    // 橙色系 — 合作伙伴
    alumni: '#722ED1',     // 紫色系 — 校友/同乡
    friend: '#8C8C8C',     // 灰色系 — 朋友/其他
    center: '#2B5FD7',     // 中心节点
    highlight: '#F5A623',  // 路径高亮
  },

  // 文字色
  text: {
    primary: '#1A1A1A',
    secondary: '#666666',
    disabled: '#BFBFBF',
  },

  // 功能色
  white: '#FFFFFF',
  border: '#E8E8E8',
}

// 间距
export const SPACING = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
}

// 字号
export const FONT_SIZE = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 18,
  xl: 20,
  xxl: 24,
  title: 28,
}
