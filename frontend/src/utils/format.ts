/**
 * 将后端返回的时间字符串格式化为 "YYYY-MM-DD HH:mm:ss"
 * 兼容 "2026-08-13T10:30:00+08:00"（Go time.Time RFC3339）等格式。
 * 优先保留字符串中的本地时间部分，避免浏览器时区转换造成偏差。
 */
export function formatDateTime(value?: string | null): string {
  if (!value) return '-'
  const s = String(value)
  // 直接匹配 "2026-08-13T10:30:00" 或 "2026-08-13 10:30:00"
  const match = s.match(/^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}/)
  if (match) {
    return match[0].replace('T', ' ')
  }
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
