/**
 * 猪猪记账 — 全局共享类型定义
 * 
 * 统一管理跨组件使用的接口类型，消除 any 滥用
 */

/** 记录项（与 stores/accounting.ts 的 Record 对齐） */
export interface AccountRecord {
  id: number
  type: string
  amount: number
  category: string
  note: string
  date: string
  created_at: string
  account_id?: number
}

/** 账户项 */
export interface AccountItem {
  id: number
  name: string
  icon: string
  balance?: number
}

/** 待办任务 */
export interface TaskItem {
  id: number
  title: string
  due_date: string
  amount: number
  completed: boolean
  daysLeft?: number
}

/** 预算预警项 */
export interface BudgetAlert {
  type: string
  message: string
  spent: number
  budget: number
  percentage: number
  remaining: number
}

/** 账户（含余额，用于 GetAllAccountsWithBalance 返回值） */
export interface AccountWithBalance {
  id: number
  name: string
  icon: string
  balance: number
  real_balance?: number
  is_default?: boolean
}

/** 提醒通知基础结构 */
export interface ReminderNotification {
  show: boolean
  title: string
  message: string
  icon: string
  timer: number | null
}

/** 任务提醒通知 */
export interface TaskReminderNotification {
  show: boolean
  message: string
  count: number
  timer: number | null
}

/** 快捷操作入口 */
export interface QuickActionDef {
  icon: string
  label: string
  color: string
  action: () => void
  route?: string
  actionType?: 'record' | 'navigate' | 'popup'
}

/** 筛选胶囊定义 */
export interface ChipDef {
  key: string
  label: string
  name?: string
  icon?: string
  active: boolean
  action: () => void
  activeStyle?: Record<string, string>
}

/** 预算信息 */
export interface BudgetInfo {
  budget_type: string
  year: number
  month: number
  budget_amount: number
  spent: number
  remaining: number
  percentage: number
}

/** AccountingStore 公共接口（供 Modal 组件使用，避免 store: any） */
export interface AccountingStoreLike {
  loadRecords(): Promise<void>
  init(): Promise<void>
  initialized: boolean
  categories?: { id?: number; name: string; icon: string; type: string; is_system?: boolean }[]
}

/** 导入结果 */
export interface ImportResult {
  successCount: number
  skipCount: number
  errorCount: number
}
