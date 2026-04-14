/**
 * 金额格式化
 */
export function formatAmount(amount: number): string {
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 日期格式化
 */
export function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

/**
 * 日期格式化（完整）
 */
export function formatDateFull(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'short' })
}

/**
 * 月份切换
 */
export function changeMonth(current: string, direction: -1 | 1): string {
  const [year, month] = current.split('-').map(Number)
  const d = new Date(year, month - 1 + direction, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

/**
 * 百分比格式化
 */
export function formatPercent(value: number): string {
  return value.toFixed(1) + '%'
}
