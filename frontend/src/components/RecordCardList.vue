<template>
  <!-- 加载中 -->
  <div v-if="isLoading" class="rp-loading">
    <div class="loading-spinner"></div>
    <span>加载中...</span>
  </div>

  <!-- 空状态 -->
  <div v-else-if="filteredRecords.length === 0" class="rp-empty">
    <div class="empty-icon">🐷</div>
    <p class="empty-title">{{ hasActiveFilters ? '没有找到匹配的记录' : '这个月还没有记账哦' }}</p>
    <p class="empty-sub" v-if="!hasActiveFilters">点击下方按钮记一笔吧~</p>
  </div>

  <!-- 分组列表 -->
  <div v-else class="rp-groups">
    <div v-for="group in sortedGroups" :key="group.dateKey" class="rp-group">
      <!-- 粘性日期头 -->
      <div class="rp-sticky-header" :style="stickyHeaderBg">
        <span class="rp-date-label">{{ group.dateLabel }}</span>
        <div class="rp-date-stats">
          <span v-if="group.income > 0" class="rp-stat-in">+¥{{ formatAmount(group.income) }}</span>
          <span v-if="group.expense > 0" class="rp-stat-out">-¥{{ formatAmount(group.expense) }}</span>
        </div>
      </div>

      <!-- 记录卡片 -->
      <div class="rp-cards">
        <div v-for="rec in group.records" :key="rec.id" class="rp-card" :style="cardBg"
          @click="$emit('edit', rec)">
          <div class="rp-card-inner">
            <!-- 图标（圆形） -->
            <div class="rp-icon-circle" :style="iconBg(rec)">
              <span class="rp-icon-text">{{ getRecIcon(rec) }}</span>
            </div>

            <!-- 信息区 -->
            <div class="rp-info">
              <div class="rp-category-line">
                <span class="rp-cat-name">{{ rec.category }}</span>
                <span v-if="rec.note" class="rp-note">· {{ rec.note }}</span>
              </div>
              <div class="rp-meta-line">
                <span class="rp-time">{{ formatTime(rec.date) }}</span>
                <span v-if="getAccountName(getRecAccountId(rec))" class="rp-account">· {{
                  getAccountName(getRecAccountId(rec)) }}</span>
              </div>
              <!-- 标签 -->
              <div v-if="rec.tags && rec.tags.length > 0" class="rp-tags-line">
                <span v-for="tag in rec.tags" :key="tag.id" class="rp-tag-badge"
                  :style="{ background: tag.color + '18', color: tag.color, borderColor: tag.color + '40' }">
                  #{{ tag.name }}
                </span>
              </div>
            </div>

            <!-- 金额 -->
            <div class="rp-amount" :class="rec.type === 'income' ? 'amount-in' : 'amount-out'">
              {{ rec.type === 'income' ? '+' : '-' }}¥{{ formatAmount(rec.amount) }}
            </div>

            <!-- 删除按钮 -->
            <button class="rp-delete-btn" @click.stop="$emit('delete', rec)" title="删除">
              <svg width="15" height="15" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore && !isLoadingMore" class="rp-load-more" ref="loadMoreEl">
      — 下滑加载更多 —
    </div>
    <div v-else-if="isLoadingMore" class="rp-load-more">
      <span class="load-dot"></span><span class="load-dot"></span><span class="load-dot"></span>
      加载中...
    </div>
    <div v-else-if="filteredRecords.length > 0" class="rp-end">
      — 已显示全部 —
    </div>
  </div>

  <!-- 删除确认弹窗 -->
  <Teleport to="body">
    <div v-if="deleteRec" class="confirm-overlay" @click="$emit('cancelDelete')">
      <div class="confirm-box" :style="panelBg">
        <p class="confirm-title">确认删除？</p>
        <p class="confirm-desc">{{ deleteRec.category }} · ¥{{ formatAmount(deleteRec.amount) }}</p>
        <div class="confirm-actions">
          <button @click="$emit('cancelDelete')" class="confirm-cancel">取消</button>
          <button @click="$emit('confirmDelete')" class="confirm-ok">删除</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Record as AccountRecord, Category } from '../stores/accounting'
import { formatAmount } from '../utils/formatters'
import { getCategoryIcon } from '../utils/category'

const props = defineProps<{
  filteredRecords: AccountRecord[]
  hasActiveFilters: boolean
  isLoading: boolean
  isLoadingMore: boolean
  hasMore: boolean
  categories: Category[]
  accounts: { id: number; name: string; icon: string }[]
  sortMode: string
  isDark: boolean
  deleteRec: AccountRecord | null
}>()

defineEmits<{
  (e: 'edit', rec: AccountRecord): void
  (e: 'delete', rec: AccountRecord): void
  (e: 'confirmDelete'): void
  (e: 'cancelDelete'): void
  (e: 'loadMore'): void
}>()

// 样式计算属性
const cardBg = computed(() => ({ background: props.isDark ? '#1c1c1e' : '#fff' }))
const stickyHeaderBg = computed(() => ({ background: props.isDark ? '#1c1c1e' : '#fafafa' }))
const panelBg = computed(() => ({ background: props.isDark ? '#2c2c2e' : '#fff' }))

function getAccountName(accountId: number | undefined): string {
  if (!accountId) return ''
  const acc = props.accounts.find(a => a.id === accountId)
  return acc ? acc.name : ''
}

function getRecAccountId(rec: any): number | undefined {
  return (rec as any).account_id
}

function getRecIcon(rec: AccountRecord): string {
  return getCategoryIcon(rec.category, props.categories, rec.type === 'income' ? '💰' : '📦')
}

function iconBg(rec: AccountRecord) {
  if (rec.type === 'income') return { background: props.isDark ? 'rgba(48,209,88,0.15)' : 'rgba(48,209,88,0.10)' }
  return { background: props.isDark ? 'rgba(255,69,58,0.15)' : 'rgba(255,69,58,0.10)' }
}

// 按日期分组
interface DayGroup {
  dateKey: string
  dateLabel: string
  records: AccountRecord[]
  income: number
  expense: number
}

const sortedGroups = computed<DayGroup[]>(() => {
  const map = new Map<string, DayGroup>()
  for (const rec of props.filteredRecords) {
    const ds = (rec.date || '').split('T')[0]
    if (!ds) continue
    const existing = map.get(ds)
    if (existing) {
      existing.records.push(rec)
      if (rec.type === 'income') {
        existing.income += rec.amount
      } else {
        existing.expense += rec.amount
      }
    } else {
      map.set(ds, {
        dateKey: ds,
        dateLabel: formatDateLabel(ds),
        records: [rec],
        income: rec.type === 'income' ? rec.amount : 0,
        expense: rec.type === 'expense' ? rec.amount : 0,
      })
    }
  }
  const asc = (props.sortMode === 'time_asc') ? -1 : 1
  return Array.from(map.values()).sort((a, b) => asc * b.dateKey.localeCompare(a.dateKey))
})

function formatDateLabel(dateStr: string): string {
  const today = new Date()
  const tds = today.toISOString().split('T')[0]
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  const yds = yesterday.toISOString().split('T')[0]
  if (dateStr === tds) return '今天'
  if (dateStr === yds) return '昨天'
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

function formatTime(dateStr: string): string {
  try {
    const d = new Date(dateStr)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  } catch { return '' }
}

// 无限滚动
const loadMoreEl = ref<HTMLElement | null>(null)

watch(loadMoreEl, (el) => {
  if (!el) return
  const observer = new IntersectionObserver(
    entries => {
      if (entries[0]?.isIntersecting && props.hasMore && !props.isLoadingMore) {
        // 触发加载更多
      }
    },
    { rootMargin: '100px' }
  )
  observer.observe(el)
})
</script>

<style scoped>
/* 加载 / 空 */
.rp-loading,
.rp-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 10px;
}

.loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid v-bind('isDark ? "rgba(255,255,255,0.08)" : "#e5e7eb"');
  border-top-color: #a855f1;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-icon {
  font-size: 48px;
  opacity: 0.55;
}

.empty-title {
  font-size: 14px;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

.empty-sub {
  font-size: 12px;
  color: v-bind('isDark ? "#3a3a3c" : "#d1d5db"');
  margin-top: 2px;
}

/* 分组 */
.rp-group {
  margin-bottom: 8px;
}

.rp-sticky-header {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 7px 12px;
  border-radius: 8px 8px 0 0;
  font-size: 12px;
}

.rp-date-label {
  font-weight: 600;
  color: v-bind('isDark ? "#e5e5ea" : "#374151"');
}

.rp-date-stats {
  display: flex;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
}

.rp-stat-in {
  color: #30d158;
}

.rp-stat-out {
  color: #ff453a;
}

/* 记录卡片 */
.rp-cards {
  border-radius: 0 0 10px 10px;
  overflow: hidden;
}

.rp-card {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.rp-card:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-hover);
}

.rp-card:last-child {
  border-bottom: none;
}

.rp-card-inner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  position: relative;
  z-index: 2;
  cursor: pointer;
  transition: background 0.15s;
}

.rp-card-inner:hover {
  background: v-bind('isDark ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.015)"');
}

/* 删除按钮 */
.rp-delete-btn {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
  background: transparent;
  transition: all 0.15s;
}

.rp-delete-btn:hover {
  color: #ff453a;
  background: rgba(255, 69, 58, 0.08);
}

/* 圆形图标 */
.rp-icon-circle {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.rp-icon-text {
  font-size: 18px;
}

/* 信息区（紧凑布局） */
.rp-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rp-category-line {
  display: flex;
  align-items: center;
  gap: 4px;
  overflow: hidden;
}

.rp-cat-name {
  font-size: 14px;
  font-weight: 500;
  color: v-bind('isDark ? "#e5e5ea" : "#1f2937"');
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rp-note {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

.rp-meta-line {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

.rp-account {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 80px;
}

/* 标签行 */
.rp-tags-line {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-top: 2px;
}

.rp-tag-badge {
  display: inline-block;
  padding: 1px 7px;
  border-radius: 6px;
  font-size: 10.5px;
  font-weight: 500;
  border: 1px solid;
  white-space: nowrap;
  line-height: 1.4;
}

/* 金额 */
.rp-amount {
  font-size: 15px;
  font-weight: 600;
  flex-shrink: 0;
  margin-right: 2px;
  font-variant-numeric: tabular-nums;
}

.amount-in {
  color: #30d158;
}

.amount-out {
  color: #ff453a;
}

/* 加载更多 */
.rp-load-more,
.rp-end {
  text-align: center;
  padding: 16px 0;
  font-size: 12px;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

.load-dot {
  display: inline-block;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #a855f7;
  animation: loadPulse 1s ease-in-out infinite;
  margin: 0 2px;
}

.load-dot:nth-child(2) {
  animation-delay: 0.15s;
}

.load-dot:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes loadPulse {
  0%,
  80%,
  100% {
    opacity: 0.3;
    transform: scale(0.8);
  }
  40% {
    opacity: 1;
    transform: scale(1);
  }
}

/* 删除确认 */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 999;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-box {
  width: 260px;
  border-radius: 16px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.2);
}

.confirm-title {
  font-size: 16px;
  font-weight: 600;
  color: v-bind('isDark ? "#f5f5f7" : "#1f2937"');
  margin-bottom: 6px;
}

.confirm-desc {
  font-size: 13px;
  color: v-bind('isDark ? "#98989d" : "#6b7280"');
  margin-bottom: 16px;
}

.confirm-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.confirm-cancel,
.confirm-ok {
  flex: 1;
  padding: 9px 0;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.confirm-cancel {
  background: v-bind('isDark ? "rgba(255,255,255,0.08)" : "#f3f4f6"');
  color: v-bind('isDark ? "#f5f5f7" : "#374151"');
}

.confirm-ok {
  background: #ff453a;
  color: #fff;
}
</style>
