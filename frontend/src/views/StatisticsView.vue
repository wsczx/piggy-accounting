<template>
  <div class="page-container stats-dashboard">
    <!-- 加载 -->
    <div v-if="store.loading" class="loading-state">
      <div style="text-align:center;">
        <div style="font-size:48px; margin-bottom:12px;">🐷</div>
        <p class="loading-text">加载中...</p>
      </div>
    </div>

    <div v-else class="dashboard-content">
      <!-- 顶部：周期切换 + 概览卡片 -->
      <div class="dashboard-header">
        <!-- 周期切换 -->
        <div class="period-switcher">
          <div class="switcher-bg">
            <button
              @click="statPeriod = 'monthly'"
              class="switcher-btn"
              :class="{ active: statPeriod === 'monthly', 'switcher-btn-active': statPeriod === 'monthly' }"
            >
              月度统计
            </button>
            <button
              @click="statPeriod = 'yearly'"
              class="switcher-btn"
              :class="{ active: statPeriod === 'yearly', 'switcher-btn-active': statPeriod === 'yearly' }"
            >
              年度统计
            </button>
          </div>
        </div>

        <!-- 概览卡片 -->
        <div class="overview-cards">
          <div class="overview-card income-card" :style="{ background: incomeCardBg }">
            <div class="card-icon">📈</div>
            <div class="card-label card-label-income">{{ statPeriod === 'monthly' ? '本月' : '本年' }}收入</div>
            <div class="card-value card-value-income">¥{{ formatAmount(currentStats.total_income) }}</div>
          </div>
          <div class="overview-card expense-card" :style="{animationDelay: '50ms', background: expenseCardBg }">
            <div class="card-icon">📉</div>
            <div class="card-label card-label-expense">{{ statPeriod === 'monthly' ? '本月' : '本年' }}支出</div>
            <div class="card-value card-value-expense">¥{{ formatAmount(currentStats.total_expense) }}</div>
          </div>
          <div class="overview-card balance-card" :style="{animationDelay: '100ms', background: balanceCardBg }">
            <div class="card-icon">💰</div>
            <div class="card-label card-label-balance">结余</div>
            <div class="card-value card-value-balance">¥{{ formatAmount(Math.abs(currentStats.balance)) }}</div>
          </div>
        </div>
      </div>

      <!-- 中间：双列布局（预算 + 分类） -->
      <div class="dashboard-middle">
        <!-- 左侧：预算使用情况 -->
        <div class="dashboard-card budget-card" :style="{animationDelay: '150ms'}">
          <div class="card-header">
            <h3>预算使用情况</h3>
            <span class="budget-period-tag">{{ statPeriod === 'monthly' ? '本月' : '本年' }}</span>
          </div>

          <div class="budget-content">
            <template v-if="currentBudget">
              <div class="budget-ring-center">
                <PieChart
                  :data="budgetPieData"
                  :size="130"
                  :stroke-width="11"
                  :center-label="statPeriod === 'monthly' ? '本月' : '本年'"
                  :center-value="`${Math.round(currentBudget.percentage)}%`"
                  :is-dark="isDark"
                />
              </div>
              <div class="budget-info-center">
                <div class="budget-row-big">
                  <span class="budget-row-label">预算</span>
                  <span class="budget-num">¥{{ formatAmount(currentBudget.budget_amount) }}</span>
                </div>
                <div class="budget-row-big">
                  <span class="budget-row-label">已用</span>
                  <span class="budget-num budget-num-expense">¥{{ formatAmount(currentBudget.spent) }}</span>
                </div>
                <div class="budget-row-big">
                  <span class="budget-row-label">{{ currentBudget.remaining >= 0 ? '剩余' : '超支' }}</span>
                  <span class="budget-num" :style="{ color: budgetStatusColor }">{{ currentBudget.remaining >= 0 ? '' : '-' }}¥{{ formatAmount(Math.abs(currentBudget.remaining)) }}</span>
                </div>
              </div>
              <div class="budget-progress-wrap">
                <div class="budget-progress-track">
                  <div
                    class="budget-progress-fill"
                    :style="{width: Math.min(currentBudget.percentage, 100) + '%', background: budgetStatusColor}"
                  ></div>
                </div>
              </div>
            </template>
            <template v-else>
              <div class="budget-empty">
                <div class="empty-icon">💰</div>
                <p class="budget-empty-text">尚未设置{{ statPeriod === 'monthly' ? '月度' : '年度' }}预算</p>
                <p class="budget-hint">前往「我的」→「预算设置」</p>
              </div>
            </template>
          </div>
        </div>

        <!-- 右侧：分类统计（柱状图排行） -->
        <div class="dashboard-card category-card" :style="{animationDelay: '200ms'}">
          <div class="card-header">
            <h3>分类{{ statType === 'income' ? '收入' : '支出' }}排行</h3>
            <div class="type-tabs">
              <button
                @click="statType = 'expense'; loadStats()"
                :class="{ active: statType === 'expense' }"
                class="type-tab-expense-active"
              >支出</button>
              <button
                @click="statType = 'income'; loadStats()"
                :class="{ active: statType === 'income' }"
                class="type-tab-income-active"
              >收入</button>
            </div>
          </div>

          <div class="category-content-bar">
            <template v-if="currentCategoryStats.length > 0">
              <div class="bar-chart-list">
                <div
                  v-for="(item, idx) in currentCategoryStats"
                  :key="item.category"
                  class="bar-row"
                  :style="{animationDelay: idx * 40 + 'ms'}"
                >
                  <div class="bar-label">
                    <span class="bar-rank" :class="{ 'top3': idx < 3 }">{{ idx + 1 }}</span>
                    <span class="bar-icon">{{ item.category_icon }}</span>
                    <span class="bar-name">{{ item.category }}</span>
                  </div>
                  <div class="bar-graphic">
                    <div class="bar-track">
                      <div
                        class="bar-fill"
                        :style="{
                          width: (item.amount / maxCategoryAmount * 100) + '%',
                          background: getCategoryColor(idx)
                        }"
                      ></div>
                    </div>
                    <div class="bar-meta">
                      <span class="bar-amount">¥{{ formatAmount(item.amount) }}</span>
                      <span class="bar-percent">{{ formatPercent(item.percentage) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </template>
            <template v-else>
              <div class="category-empty">
                <div class="empty-icon">📭</div>
                <p class="category-empty-text">暂无{{ statType === 'income' ? '收入' : '支出' }}记录</p>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 底部：趋势图表 -->
      <div class="dashboard-card trend-card" :style="{animationDelay: '250ms'}">
        <div class="card-header">
          <h3>{{ statPeriod === 'monthly' ? '日度' : '月度' }}趋势</h3>
          <div class="trend-legend">
            <span class="legend-dot legend-dot-income"></span>
            <span class="legend-text">收入</span>
            <span class="legend-dot legend-dot-expense"></span>
            <span class="legend-text">支出</span>
          </div>
        </div>
        <div class="trend-chart-wrapper">
          <LineChart
            :data="trendData"
            :width="trendChartWidth"
            :height="140"
            :is-dark="isDark"
            :show-points="true"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore } from '../stores/accounting'
import { formatAmount, formatPercent } from '../utils/formatters'
import PieChart from '../components/PieChart.vue'
import LineChart from '../components/LineChart.vue'

const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

const statType = ref<'expense' | 'income'>('expense')
const statPeriod = ref<'monthly' | 'yearly'>('monthly')
const trendChartWidth = ref(600)

// ====== Computed: 动态主题色 ======
// 概览卡片背景 (暗色半透明 vs 亮色渐变)
const incomeCardBg = computed(() =>
  isDark.value ? 'rgba(48,209,88,0.12)' : 'linear-gradient(135deg, #30d158 0%, #28c840 100%)'
)
const expenseCardBg = computed(() =>
  isDark.value ? 'rgba(255,69,58,0.12)' : 'linear-gradient(135deg, #ff453a 0%, #ff3b30 100%)'
)
const balanceCardBg = computed(() =>
  isDark.value ? 'rgba(168,85,247,0.12)' : 'linear-gradient(135deg, #f472b6 0%, #a855f7 100%)'
)

// 当前统计数据
const currentStats = computed(() => {
  return statPeriod.value === 'monthly' ? store.monthlyStats : store.yearlyStats
})

// 当前预算 — 跟随 statPeriod
const currentBudget = computed(() => {
  return statPeriod.value === 'monthly' ? store.monthlyBudget : store.yearlyBudget
})

// 当前分类统计
const currentCategoryStats = computed(() => {
  return statPeriod.value === 'monthly' ? store.categoryStats : store.yearlyCategoryStats
})

// 分类总计
const categoryTotal = computed(() => {
  return currentCategoryStats.value.reduce((sum, item) => sum + item.amount, 0)
})

// 柱状图最大值基准
const maxCategoryAmount = computed(() => {
  if (currentCategoryStats.value.length === 0) return 1
  return Math.max(...currentCategoryStats.value.map(item => item.amount))
})

// 预算环形图数据
const budgetPieData = computed(() => {
  if (!currentBudget.value) return []
  const used = currentBudget.value.spent
  const remaining = Math.max(currentBudget.value.remaining, 0)
  const over = currentBudget.value.remaining < 0 ? Math.abs(currentBudget.value.remaining) : 0

  const data = []
  if (used > 0) data.push({ label: '已用', value: used, color: '#ff453a' })
  if (remaining > 0) data.push({ label: '剩余', value: remaining, color: '#30d158' })
  if (over > 0) data.push({ label: '超支', value: over, color: '#ff9f0a' })
  return data
})

// 分类颜色（品牌粉紫渐变色系）
const categoryColors = ['#f472b6', '#ec4899', '#a855f7', '#8b5cf6', '#6366f1', '#c084fc', '#d946ef']

const getCategoryColor = (index: number) => categoryColors[index % categoryColors.length]

// 分类饼图数据
const categoryPieData = computed(() => {
  return currentCategoryStats.value.slice(0, 6).map((item, index) => ({
    label: item.category,
    value: item.amount,
    color: getCategoryColor(index)
  }))
})

// 趋势图数据
const trendData = computed(() => {
  if (statPeriod.value === 'monthly') {
    return store.dailyStats.map(d => ({
      label: d.date.slice(8) + '日',
      income: d.total_income,
      expense: d.total_expense
    }))
  } else {
    return store.monthlyTrend.map(m => ({
      label: m.month.slice(5) + '月',
      income: m.total_income,
      expense: m.total_expense
    }))
  }
})

// 预算状态颜色
const budgetStatusColor = computed(() => {
  if (!currentBudget.value) return '#30d158'
  if (currentBudget.value.percentage >= 100) return '#ff453a'
  if (currentBudget.value.percentage >= 80) return '#ff9f0a'
  return '#30d158'
})

onMounted(() => {
  updateChartWidth()
  window.addEventListener('resize', updateChartWidth)
  if (!store.initialized) {
    store.init().then(() => loadAllStats())
  } else {
    loadAllStats()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', updateChartWidth)
})

function updateChartWidth() {
  const container = document.querySelector('.trend-chart-wrapper')
  if (container) trendChartWidth.value = container.clientWidth
}

function loadAllStats() {
  loadStats()
  const year = store.currentMonth.split('-')[0]
  store.loadYearlyStats(year)
  store.loadYearlyCategoryStats(year, statType.value)
  store.loadMonthlyTrend(year)
}

function loadStats() {
  if (statPeriod.value === 'monthly') {
    store.loadCategoryStats(store.currentMonth, statType.value)
  } else {
    const year = store.currentMonth.split('-')[0]
    store.loadYearlyCategoryStats(year, statType.value)
  }
}

watch(statPeriod, () => { loadStats() })

// 监听月份切换，重新加载统计数据
watch(() => store.currentMonth, () => { loadAllStats() })
</script>

<style scoped>
/* 仪表盘容器 */
.stats-dashboard { height: 100%; overflow: hidden; }

.loading-state {
  display: flex; align-items: center; justify-content: center;
  height: 100%;
}
.loading-text {
  font-size: 14px;
  color: var(--text-secondary);
}

.dashboard-content {
  height: 100%;
  display: flex; flex-direction: column;
  padding: 12px 16px; box-sizing: border-box;
  gap: 12px;
}

/* ====== 顶部区域 ====== */
.dashboard-header { flex-shrink: 0; }

.period-switcher { display: flex; justify-content: center; margin-bottom: 12px; }

.switcher-bg {
  display: flex; gap: 4px; padding: 3px;
  border-radius: var(--radius-lg);
  background: var(--bg-input);
}

.switcher-btn {
  padding: 6px 16px; border-radius: 9px; border: none;
  cursor: pointer; font-size: 13px; font-weight: 600;
  transition: all 0.2s; background: transparent;
  color: var(--text-secondary);
}
.switcher-btn.active { box-shadow: 0 1px 4px rgba(0,0,0,0.1); }
.switcher-btn-active {
  background: var(--card-bg);
  color: var(--text-primary);
}

/* 概览卡片 */
.overview-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }

.overview-card {
  border-radius: 14px; padding: 14px 12px;
  display: flex; flex-direction: column; align-items: center; text-align: center;
  transition: transform 0.2s, box-shadow 0.2s;
  background: var(--card-bg);
}
.overview-card:hover { transform: translateY(-2px); }

.card-icon { font-size: 20px; margin-bottom: 4px; }

/* 收入/支出/结余 — 亮色模式白色文字（彩色渐变底），暗色模式彩色文字 */
.card-label { font-size: 11px; margin-bottom: 2px; opacity: 0.9; }
.card-label-income,
.card-label-expense,
.card-label-balance { color: #fff; }

.card-value { font-size: 16px; font-weight: 700; font-variant-numeric: tabular-nums; }
.card-value-income { color: #fff; }
.card-value-expense { color: #fff; }
.card-value-balance { color: #fff; }

/* 暗色模式下恢复彩色文字 */
[data-theme="dark"] .card-label-income,
[data-theme="dark"] .card-value-income { color: var(--income-color); }
[data-theme="dark"] .card-label-expense,
[data-theme="dark"] .card-value-expense { color: var(--expense-color); }
[data-theme="dark"] .card-label-balance,
[data-theme="dark"] .card-value-balance { color: var(--accent-color); }

/* ====== 中间双列布局 ====== */
.dashboard-middle {
  display: grid; grid-template-columns: 220px 1fr; gap: 12px;
  flex: 1; min-height: 0;
}

/* 卡片通用 — 使用 token */
.dashboard-card {
  border-radius: 16px; padding: 14px;
  display: flex; flex-direction: column; overflow: hidden;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}

.budget-card { min-height: 260px; }

.card-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px; flex-shrink: 0;
}
.card-header h3 {
  font-size: 14px; font-weight: 700; margin: 0;
  color: var(--text-primary);
}

/* 预算周期标签 */
.budget-period-tag {
  font-size: 11px; font-weight: 500; padding: 2px 8px;
  border-radius: 6px;
  color: var(--text-secondary);
  background: var(--bg-input);
}

/* ====== 预算卡片 — 居中加大布局 ====== */
.budget-content {
  flex: 1;
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 10px;
  overflow: hidden; padding-top: 4px; min-height: 0;
}

.budget-ring-center {
  flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
}

.budget-info-center {
  width: 100%; display: flex; flex-direction: column; gap: 6px;
  flex-shrink: 0;
}

.budget-row-big {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 13px; padding: 0 4px;
}
.budget-row-label {
  color: var(--text-secondary);
}

.budget-num {
  font-weight: 700; font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}
.budget-num-expense { color: var(--expense-color); }

.budget-progress-wrap {
  width: 100%; flex-shrink: 0; padding: 0 4px;
}

.budget-progress-track {
  height: 6px; background: rgba(128,128,128,0.1);
  border-radius: 3px; overflow: hidden;
}

.budget-progress-fill {
  height: 100%; border-radius: 3px;
  transition: width 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.budget-empty {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 8px; text-align: center;
}

.empty-icon { font-size: 32px; opacity: 0.6; }
.budget-empty-text { font-size: 12px; margin: 0; color: var(--text-secondary); }
.budget-hint { margin-top: -4px; font-size: 11px; color: var(--text-secondary); opacity: 0.7; }

/* ====== 类型 Tab ====== */
.type-tabs {
  display: flex; gap: 2px; padding: 2px; border-radius: 8px;
  background: var(--bg-input);
}

.type-tabs button {
  padding: 4px 10px; border-radius: 6px; border: none;
  cursor: pointer; font-size: 11px; font-weight: 600;
  transition: all 0.2s; background: transparent;
  color: var(--text-secondary);
}
.type-tab-expense-active.active { background: var(--expense-color); color: #fff; }
.type-tab-income-active.active { background: var(--income-color); color: #fff; }

/* ====== 分类卡片 — 柱状图布局 ====== */
.category-content-bar {
  flex: 1; display: flex; flex-direction: column; gap: 8px;
  overflow-y: auto; min-height: 0;
}

.bar-chart-list { display: flex; flex-direction: column; gap: 10px; }

.bar-row {
  display: flex; flex-direction: column; gap: 4px;
  animation: fadeSlideIn 0.3s ease-out forwards; opacity: 0;
}

@keyframes fadeSlideIn {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}

.bar-label { display: flex; align-items: center; gap: 6px; }

.bar-rank {
  font-size: 11px; font-weight: 700;
  width: 16px; text-align: center; flex-shrink: 0;
  color: var(--text-muted);
}
.bar-rank.top3 {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  color: #fff; border-radius: 4px; line-height: 18px;
}

.bar-icon { font-size: 14px; width: 20px; text-align: center; flex-shrink: 0; }

.bar-name {
  font-size: 12px; font-weight: 500;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  color: var(--text-primary);
}

.bar-graphic { display: flex; align-items: center; gap: 8px; padding-left: 22px; }

.bar-track {
  flex: 1; height: 8px; background: rgba(128,128,128,0.1);
  border-radius: 4px; overflow: hidden; min-width: 0;
}

.bar-fill {
  height: 100%; border-radius: 4px;
  transition: width 0.5s cubic-bezier(0.22, 1, 0.36, 1); min-width: 2px;
}

.bar-meta { display: flex; flex-direction: column; align-items: flex-end; gap: 0; flex-shrink: 0; min-width: 70px; }

.bar-amount {
  font-size: 11px; font-variant-numeric: tabular-nums;
  color: var(--text-primary);
  font-weight: 600;
}
.bar-percent {
  font-size: 10px;
  color: var(--text-secondary);
}

.category-empty {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 8px;
}
.category-empty-text {
  font-size: 12px; margin: 0;
  color: var(--text-secondary);
}

/* ====== 趋势卡片 ====== */
.trend-card { flex-shrink: 0; height: 190px; }

.trend-legend { display: flex; align-items: center; gap: 6px; font-size: 11px; }

.legend-dot { width: 8px; height: 8px; border-radius: 50%; }
.legend-dot-income { background: var(--income-color); }
.legend-dot-expense { background: var(--expense-color); }
.legend-text { color: var(--text-secondary); }

.trend-chart-wrapper { flex: 1; min-height: 0; display: flex; align-items: center; }

/* ====== 滚动条 ====== */
.category-content-bar::-webkit-scrollbar { width: 3px; }
.category-content-bar::-webkit-scrollbar-track { background: transparent; }
.category-content-bar::-webkit-scrollbar-thumb { background: rgba(128,128,128,0.2); border-radius: 10px; }

/* ====== 响应式 ====== */
@media (max-width: 600px) {
  .dashboard-middle { grid-template-columns: 1fr; grid-template-rows: auto auto; }
  .overview-cards { grid-template-columns: 1fr; }
  .category-content-bar { overflow-y: visible; }
}
</style>
