<template>
  <div class="balance-card" :style="cardBg">
    <!-- 装饰光斑 -->
    <div class="bc-glow bc-glow-1"></div>
    <div class="bc-glow bc-glow-2"></div>
    <!-- ===== 左侧：结余信息 ===== -->
    <div class="bc-left">
      <!-- 标题行 -->
      <div class="bc-row">
        <span class="bc-label">本月结余</span>
      </div>

      <!-- 金额 -->
      <div class="bc-amount">¥{{ formatAmount(Math.abs(stats.balance)) }}</div>

      <!-- 收入 / 支出 -->
      <div class="bc-totals">
        <div class="bc-item">
          <span class="bc-icon income"><svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="#fff"
              stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 11l5-5m0 0l5 5m-5-5v12" />
            </svg></span>
          <span>¥{{ formatAmount(stats.total_income) }}</span>
        </div>
        <div class="bc-divider"></div>
        <div class="bc-item">
          <span class="bc-icon expense"><svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="#fff"
              stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 13l-5 5m0 0l-5-5m5 5V6" />
            </svg></span>
          <span>¥{{ formatAmount(stats.total_expense) }}</span>
        </div>
      </div>

      <!-- 预算进度条 -->
      <template v-if="store.monthlyBudget">
        <div class="bc-budget-block">
          <div class="bc-budget-header">
            <span class="bc-budget-title">本月预算进度</span>
            <span class="bc-budget-used">¥{{ formatAmount(store.monthlyBudget.spent) }} / ¥{{
              formatAmount(store.monthlyBudget.budget_amount) }}</span>
          </div>
          <div class="bc-progress-track-lg">
            <div class="bc-progress-bar-lg" :style="{ width: budgetPct + '%', background: barColor }"></div>
          </div>
          <div class="bc-budget-footer">
            <span v-if="budgetPct >= 100" class="bc-budget-alert">⚠️ 已超支 ¥{{
              formatAmount(Math.abs(store.monthlyBudget.remaining)) }}</span>
            <span v-else class="bc-budget-remain">剩余 ¥{{ formatAmount(store.monthlyBudget.remaining) }} · {{
              Math.round(budgetPct) }}%</span>
          </div>
        </div>
      </template>
    </div>

    <!-- ===== 右侧：本周概览（柱状图 + TOP3 并排整体） ====== -->
    <div class="bc-right" v-if="weekDays.length > 0 || topCategories.length > 0">
      <div class="week-header" v-if="weekDays.length > 0">
        <span class="week-title">本周支出</span>
        <span class="week-total">¥{{ formatAmount(weekSummary.expense) }}</span>
      </div>
      <div class="bc-right-body">
        <!-- 柱状图 -->
        <div class="week-bars" v-if="weekDays.length > 0">
          <div v-for="d in weekDays" :key="d.label" class="week-col">
            <div class="week-bar-wrap">
              <div class="week-bar"
                :class="{ 'bar-income': d.income > d.expense, 'bar-zero': d.expense === 0 && d.income === 0 }"
                :style="{ height: barHeight(d) }"></div>
            </div>
            <span class="week-day-label" :class="{ active: d.isToday }">{{ d.label }}</span>
          </div>
        </div>

        <!-- TOP3 分类（与柱状图并排，在同一行右侧） -->
        <div class="bc-topcats" v-if="topCategories.length > 0">
          <div class="bc-topcats-title">TOP3</div>
          <div v-for="(cat, idx) in topCategories" :key="cat.category" class="bc-cat-row">
            <span class="bc-cat-name">{{ cat.category_icon }}{{ cat.category }}</span>
            <div class="bc-cat-track">
              <div class="bc-cat-bar" :style="{ width: cat.percentage + '%' }"></div>
            </div>
            <span class="bc-cat-amt">¥{{ formatAmount(cat.amount) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 未设置预算提示 -->
    <div v-else-if="!store.monthlyBudget && stats.total_expense > 0" class="bc-hint">
      尚未设置本月预算 · <span class="bc-link" @click="$emit('go-budget')">去设置</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore } from '../stores/accounting'
import { getCategoryIcon } from '../utils/category'
import { formatAmount } from '../utils/formatters'

defineEmits(['go-budget'])

const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

const cardBg = computed(() =>
  isDark.value
    ? { background: 'linear-gradient(135deg, #2d1b4e, #1a1033, #1f2937)', border: '1px solid rgba(255,255,255,0.06)' }
    : { background: 'linear-gradient(135deg, #f472b6, #a855f7, #7c3aed)', boxShadow: '0 8px 32px rgba(168,85,247,0.25)' }
)

const stats = computed(() => ({
  total_income: store.monthlyStats?.total_income ?? 0,
  total_expense: store.monthlyStats?.total_expense ?? 0,
  balance: (store.monthlyStats?.total_income ?? 0) - (store.monthlyStats?.total_expense ?? 0),
}))

/** 本周每日数据（周一到今天） */
interface DayData {
  label: string
  date: Date
  income: number
  expense: number
  isToday: boolean
}

/** 本地时区日期字符串 YYYY-MM-DD */
function toLocalDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const weekDays = computed<DayData[]>(() => {
  const days = store.dailyStats
  if (!days || days.length === 0) return []

  const now = new Date()
  const dayOfWeek = now.getDay() || 7
  const monday = new Date(now)
  monday.setDate(now.getDate() - dayOfWeek + 1)
  monday.setHours(0, 0, 0, 0)

  const result: DayData[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    if (d > now) break // 不显示未来日期

    const dateStr = toLocalDate(d)
    const found = days.find(x => x.date === dateStr)
    result.push({
      label: ['一', '二', '三', '四', '五', '六', '日'][i],
      date: d,
      income: found?.total_income ?? 0,
      expense: found?.total_expense ?? 0,
      isToday: d.toDateString() === now.toDateString(),
    })
  }
  return result
})

/** 本周汇总 */
const weekSummary = computed(() => {
  return weekDays.value.reduce(
    ({ income, expense }, d) => ({
      income: income + d.income,
      expense: expense + d.expense,
    }),
    { income: 0, expense: 0 }
  )
})

/** 柱状图高度计算 */
function barHeight(d: DayData): string {
  const maxAmt = Math.max(...weekDays.value.map(x => Math.max(x.expense, x.income)), 1)
  const val = Math.max(d.expense, d.income)
  if (val === 0) return '2px'
  // 高度范围: 4px ~ 52px
  const h = 4 + (val / maxAmt) * 48
  return `${h}px`
}

/** TOP 3 支出分类（本月，直接用 categoryStats） */
const topCategories = computed(() => {
  const cats = (store.categoryStats || [])
    .filter(c => c.amount > 0)
    .sort((a, b) => b.amount - a.amount)
    .slice(0, 3)
    .map(c => ({
      category: c.category,
      category_icon: getCategoryIcon(c.category, store.categories),
      amount: c.amount,
    }))

  if (cats.length === 0) return []
  const maxAmt = cats[0]?.amount || 1
  return cats.map(c => ({ ...c, percentage: (c.amount / maxAmt) * 100 }))
})

const budgetPct = computed(() => {
  const b = store.monthlyBudget
  if (!b || !b.budget_amount) return 0
  return Math.min(100, (b.spent / b.budget_amount) * 100)
})

// 阶段式颜色：0-60% 绿色，60-80% 黄色，80%+ 红色
const barColor = computed(() => {
  const pct = budgetPct.value
  if (pct >= 80) return '#ff453a'
  if (pct >= 60) return '#ff9f0a'
  return '#30d158'
})
</script>

<style scoped>
.balance-card {
  border-radius: 18px;
  padding: 14px 18px;
  margin-bottom: 10px;
  display: flex;
  /* 两栏：左结余 | 右(柱状图+TOP3) */
  gap: 20px;
  align-items: center;
  position: relative;
  overflow: hidden;
}

/* 装饰光斑 */
.bc-glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.bc-glow-1 {
  width: 120px;
  height: 120px;
  background: rgba(244, 114, 182, 0.15);
  top: -30px;
  right: -20px;
  filter: blur(40px);
}
.bc-glow-2 {
  width: 100px;
  height: 100px;
  background: rgba(129, 140, 248, 0.12);
  bottom: -20px;
  left: -10px;
  filter: blur(40px);
}

.bc-left {
  flex: 1;
  min-width: 0;
}

/* ====== 右侧：本周概览（柱状图 + TOP3 并排整体） ====== */
.bc-right {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  /* 标题左对齐，内容区居中 */
  align-items: flex-start;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  padding-left: 14px;
}

.bc-right-body {
  display: flex;
  align-items: center;
  /* 柱状图和TOP3垂直居中 */
  gap: 10px;
}

.bc-topcats-title {
  font-size: 9px;
  color: rgba(255, 255, 255, 0.4);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 2px;
}

/* ---- 左侧内容 ---- */
.bc-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2px;
}

.bc-label {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
  font-weight: 500;
}

.bc-amount {
  font-size: 30px;
  font-weight: 800;
  color: #fff;
  letter-spacing: -0.5px;
  margin-bottom: 6px;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.bc-totals {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.bc-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.85);
  font-weight: 500;
  white-space: nowrap;
}

.bc-icon {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.18);
}

.bc-icon.income {
  background: rgba(48, 209, 88, 0.25);
}

.bc-icon.expense {
  background: rgba(255, 69, 58, 0.25);
}

.bc-divider {
  width: 1px;
  height: 14px;
  background: rgba(255, 255, 255, 0.15);
  flex-shrink: 0;
}

/* 预算区块 */
.bc-budget-block {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.bc-budget-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}

.bc-budget-title {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.65);
  font-weight: 500;
}

.bc-budget-used {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.5);
}

.bc-progress-track-lg {
  width: 100%;
  height: 7px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.13);
  overflow: hidden;
}

.bc-progress-bar-lg {
  height: 100%;
  border-radius: 4px;
  transition: width 0.4s;
  position: relative;
  overflow: hidden;
}
.bc-progress-bar-lg::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
  background-size: 200% 100%;
  animation: shimmer 2s ease-in-out infinite;
}

.bc-budget-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}

.bc-budget-remain {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.45);
}

.bc-budget-alert {
  font-size: 10px;
  color: #ff453a;
  font-weight: 600;
}

/* ====== 右侧：本周柱状图 ====== */
.week-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 8px;
}

.week-title {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.6);
  font-weight: 500;
}

.week-total {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
}

.week-bars {
  display: flex;
  align-items: flex-end;
  gap: 5px;
  height: 60px;
  padding-top: 4px;
}

.week-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  min-width: 0;
}

.week-bar-wrap {
  width: 100%;
  height: 52px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.week-bar {
  width: 80%;
  border-radius: 3px 3px 2px 2px;
  min-height: 2px;
  /* 默认：支出 → 半透明白色偏红 */
  background: linear-gradient(to top, rgba(255, 120, 120, 0.35), rgba(255, 180, 160, 0.18));
  transition: height 0.4s ease;
}

.week-bar.bar-income {
  /* 收入 > 支出 → 半透明绿色 */
  background: linear-gradient(to top, rgba(120, 230, 160, 0.40), rgba(160, 245, 190, 0.18));
}

.week-bar.bar-zero {
  background: rgba(255, 255, 255, 0.06);
  border-radius: 2px;
}

.week-day-label {
  font-size: 9px;
  color: rgba(255, 255, 255, 0.35);
  line-height: 1;
}

.week-day-label.active {
  color: rgba(255, 255, 255, 0.75);
  font-weight: 600;
}

/* ====== TOP3 分类（与柱状图并排） ====== */
.bc-topcats {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bc-cat-row {
  display: flex;
  align-items: center;
  gap: 3px;
}

.bc-cat-name {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.7);
  flex-shrink: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 52px;
}

.bc-cat-track {
  flex: 1;
  min-width: 0;
  height: 4px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
}

.bc-cat-bar {
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0.48));
  transition: width 0.4s;
  min-width: 2px;
}

.bc-cat-amt {
  font-size: 9px;
  color: rgba(255, 255, 255, 0.5);
  width: 42px;
  text-align: right;
  flex-shrink: 0;
}

/* 提示文字 */
.bc-hint {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.32);
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.bc-link {
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-style: dotted;
}
</style>

<style>
@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
</style>
