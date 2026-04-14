/**
 * 类别相关的共享工具函数和常量
 * 消除各组件中重复的 getCategoryIcon / getCategoryColor 定义
 */
import type { Category } from '../stores/accounting'

/** 分类颜色映射表（统一使用） */
export const CATEGORY_COLORS: Record<string, string> = {
  餐饮: '#f59e0b',
  交通: '#3b82f6',
  购物: '#ec4899',
  娱乐: '#8b5cf6',
  医疗: '#ef4444',
  居住: '#10b981',
  工资: '#22c55e',
  其他: '#6b7280',
}

/** 默认颜色 */
export const DEFAULT_CATEGORY_COLOR = '#6366f1'

/**
 * 根据类别名获取图标
 * @param name 类别名
 * @param categories 全量类别列表
 * @param fallback 找不到时的默认图标
 */
export function getCategoryIcon(
  name: string,
  categories: Category[],
  fallback = '📦'
): string {
  return categories.find(c => c.name === name)?.icon || fallback
}

/**
 * 根据类别名获取颜色
 * @param name 类别名
 */
export function getCategoryColor(name: string): string {
  return CATEGORY_COLORS[name] || DEFAULT_CATEGORY_COLOR
}
