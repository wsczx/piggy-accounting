<template>
  <!-- 筛选面板卡片 -->
  <div class="rp-filter-panel">
    <!-- 第一行：类型分段 -->
    <div class="rp-segment-row">
      <div class="rp-segment" role="tablist">
        <button v-for="f in filterChips" :key="f.key" class="rp-seg-btn"
          :class="{ active: f.active, 'data-exp': f.key === 'exp', 'data-inc': f.key === 'inc' }" role="tab"
          @click="onChipClick(f)">{{ f.label }}</button>
      </div>
    </div>
    <!-- 第二行：标签筛选 -->
    <div v-if="allTags.length > 0" class="rp-tag-row">
      <button class="rp-chip rp-tag-chip"
        :class="{ active: selectedTagId === 0 }"
        @click="toggleTag(0)">全部</button>
      <button v-for="tag in allTags" :key="'tag-' + tag.id" class="rp-chip rp-tag-chip"
        :class="{ active: selectedTagId === tag.id }"
        :style="selectedTagId === tag.id ? { '--tag-active-color': tag.color } : {}"
        @click="toggleTag(tag.id)">
        <span class="rp-tag-dot" :style="{ background: tag.color }"></span>{{ tag.name }}
      </button>
    </div>
    <!-- 第三行：类别多选标签横滑 -->
    <div v-if="categoryChips.length > 0" class="rp-cat-row">
      <button v-for="c in categoryChips" :key="'cat-' + c.id" class="rp-chip rp-cat-chip"
        :class="{ active: selectedCategoryIds.includes(c.id) }"
        @click="toggleCategoryChip(c.id)">{{ c.icon }} {{ c.name }}</button>
    </div>
  </div>

  <!-- 月度迷你统计条 -->
  <div v-if="totalCount > 0" class="rp-summary-bar">
    <div class="rp-sum-left">
      <span class="rp-sum-count">{{ totalCount }} 笔记录</span>
    </div>
    <div class="rp-sum-right">
      <span class="rp-sum-in">收入 <b>+¥{{ formatAmount(monthIncome) }}</b></span>
      <span class="rp-sum-divider"></span>
      <span class="rp-sum-out">支出 <b>-¥{{ formatAmount(monthExpense) }}</b></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Category, TagInfo } from '../stores/accounting'
import { formatAmount } from '../utils/formatters'

interface ChipDef {
  key: string
  label: string
  active: boolean
  action: () => void
}

const props = defineProps<{
  filterType: 'all' | 'income' | 'expense'
  selectedCategoryIds: number[]
  selectedTagId: number
  categories: Category[]
  allTags: TagInfo[]
  allRecordsCache: any[]
  totalCount: number
  monthIncome: number
  monthExpense: number
  isDark: boolean
}>()

const emit = defineEmits<{
  (e: 'update:filterType', val: 'all' | 'income' | 'expense'): void
  (e: 'update:selectedCategoryIds', val: number[]): void
  (e: 'update:selectedTagId', val: number): void
  (e: 'search'): void
}>()

// 类型分段胶囊
const filterChips = computed<ChipDef[]>(() => [
  { key: 'all', label: '全部', active: props.filterType === 'all', action: () => setFilter('all') },
  { key: 'exp', label: '↓ 支出', active: props.filterType === 'expense', action: () => setFilter('expense') },
  { key: 'inc', label: '↑ 收入', active: props.filterType === 'income', action: () => setFilter('income') },
])

// 类别胶囊：基于全量月度缓存，按当前分段类型过滤
const categoryChips = computed(() => {
  // 根据分段类型筛选记录，再取用到的类别名
  const cacheRecords = props.filterType === 'all'
    ? props.allRecordsCache
    : props.allRecordsCache.filter((r: any) => r.type === props.filterType)
  const usedNames = new Set(cacheRecords.map((r: any) => r.category))
  // 匹配对应类型的类别
  const baseCats = props.filterType === 'all'
    ? props.categories
    : props.categories.filter(c => c.type === props.filterType)
  return baseCats.filter(c => usedNames.has(c.name))
})

function setFilter(type: 'all' | 'income' | 'expense') {
  emit('update:filterType', type)
  emit('update:selectedCategoryIds', [])
  emit('search')
}

function onChipClick(f: ChipDef) {
  f.action()
}

function toggleCategoryChip(catId: number) {
  const ids = [...props.selectedCategoryIds]
  const idx = ids.indexOf(catId)
  if (idx > -1) {
    ids.splice(idx, 1)
  } else {
    ids.push(catId)
  }
  emit('update:selectedCategoryIds', ids)
  emit('search')
}

function toggleTag(tagId: number) {
  emit('update:selectedTagId', tagId)
  emit('search')
}
</script>

<style scoped>
/* ====== 筛选面板（粉紫品牌色卡片） ====== */
.rp-filter-panel {
  border-radius: 12px;
  padding: 10px 12px;
  background: v-bind('isDark ? "rgba(244,114,182,0.05)" : "linear-gradient(135deg, rgba(244,114,182,0.06), rgba(168,85,247,0.06))"');
  border: 1px solid v-bind('isDark ? "rgba(168,85,247,0.10)" : "linear-gradient(135deg, rgba(244,114,182,0.15), rgba(168,85,247,0.15))"');
}

/* 分段控制器行 */
.rp-segment-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* iOS 风格分段控制 */
.rp-segment {
  flex: 1;
  display: flex;
  background: v-bind('isDark ? "rgba(255,255,255,0.06)" : "rgba(168,85,247,0.08)"');
  border-radius: 10px;
  padding: 2px;
}

.rp-seg-btn {
  flex: 1;
  position: relative;
  padding: 6px 0;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  z-index: 1;
  transition: all 0.2s ease;
  background: transparent;
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}

.rp-seg-btn.active {
  background: v-bind('isDark ? "#2c2c2e" : "#fff"');
  color: v-bind('isDark ? "#f5f5f7" : "#111827"');
  box-shadow: 0 1px 4px rgba(168, 85, 247, 0.12);
}

.rp-seg-btn.active.data-exp {
  color: #ff453a;
}

.rp-seg-btn.active.data-inc {
  color: #30d158;
}

.rp-seg-btn:hover:not(.active) {
  color: v-bind('isDark ? "#b4b4b4" : "#4b5563"');
}

/* 标签筛选行 */
.rp-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid v-bind('isDark ? "rgba(255,255,255,0.06)" : "rgba(0,0,0,0.05)"');
}

.rp-tag-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: 3px;
  flex-shrink: 0;
}

.rp-tag-chip.active {
  background: v-bind('isDark ? "rgba(99,102,241,0.15)" : "rgba(99,102,241,0.08)"') !important;
  color: var(--tag-active-color, #6366f1) !important;
  border-color: v-bind('isDark ? "rgba(99,102,241,0.30)" : "rgba(99,102,241,0.20)"') !important;
}

/* 类别标签行（横向滚动） */
.rp-cat-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 0;
  margin-top: 6px;
  padding-top: 6px;
}

.rp-cat-row::-webkit-scrollbar {
  display: none;
}

.rp-chip {
  flex-shrink: 0;
  padding: 4px 12px;
  border-radius: 14px;
  border: 1px solid transparent;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.16s ease;
  white-space: nowrap;
  background: v-bind('isDark ? "rgba(255,255,255,0.05)" : "#f3f4f6"');
  color: v-bind('isDark ? "#98989d" : "#6b7280"');
}

.rp-chip:hover {
  background: v-bind('isDark ? "rgba(255,255,255,0.09)" : "#e5e7eb"');
}

/* 类别胶囊 active 态 */
.rp-cat-chip.active {
  background: v-bind('isDark ? "rgba(168,85,247,0.15)" : "rgba(168,85,247,0.08)"') !important;
  color: #a855f7 !important;
  border-color: v-bind('isDark ? "rgba(168,85,247,0.30)" : "rgba(168,85,247,0.20)"') !important;
}

/* 迷你统计条 — 粉紫品牌色卡片风格 */
.rp-summary-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
  padding: 9px 14px;
  border-radius: 12px;
  background: v-bind('isDark ? "rgba(244,114,182,0.06)" : "linear-gradient(135deg, rgba(244,114,182,0.05), rgba(168,85,247,0.05))"');
  border: 1px solid v-bind('isDark ? "rgba(168,85,247,0.10)" : "linear-gradient(135deg, rgba(244,114,182,0.12), rgba(168,85,247,0.12))"');
}

.rp-sum-left {
  display: flex;
  align-items: center;
  gap: 4px;
}

.rp-sum-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rp-sum-count {
  font-size: 12px;
  font-weight: 600;
  color: v-bind('isDark ? "#c084fc" : "#a855f7"');
}

.rp-sum-in,
.rp-sum-out {
  font-size: 11.5px;
  color: v-bind('isDark ? "#98989d" : "#6b7280"');
  display: flex;
  align-items: center;
  gap: 3px;
}

.rp-sum-in b {
  color: #30d158;
  font-weight: 650;
}

.rp-sum-out b {
  color: #ff453a;
  font-weight: 650;
}

.rp-sum-divider {
  width: 1px;
  height: 12px;
  background: v-bind('isDark ? "rgba(255,255,255,0.10)" : "rgba(0,0,0,0.08)"');
}
</style>
