/**
 * 日期相关的共享工具函数
 * 消除各组件中重复的天数计算逻辑
 */

/**
 * 计算距离到期的天数（正数=未来，负数=已逾期，0=今天）
 * @param dueDate 到期日期字符串 (YYYY-MM-DD)
 */
export function calcDaysLeft(dueDate: string): number {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const due = new Date(dueDate)
  due.setHours(0, 0, 0, 0)
  if (isNaN(due.getTime())) return 0
  return Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
}
