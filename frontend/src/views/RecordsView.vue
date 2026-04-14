<template>
  <div class="page-container records-page">
    <div class="rp-content">
      <!-- 头部区域 -->
      <div class="rp-header" :style="headerBg">
        <!-- 导航栏：返回 + 居中标题 -->
        <div class="rp-nav">
          <button @click="router.back()" class="rp-back-btn">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <span class="rp-title">全部账单</span>
          <div style="width:34px"></div>
        </div>

        <!-- 搜索 + 排序 -->
        <div class="rp-search-row">
          <div class="rp-search-box" :style="searchBoxBg">
            <svg width="15" height="15" fill="none" viewBox="0 0 24 24" :stroke="isDark ? '#636366' : '#9ca3af'"
              stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input v-model="searchQuery" type="text" placeholder="搜索账单..." @input="onSearchInput"
              :style="{ color: isDark ? '#f5f5f7' : '#1f2937', background: 'transparent' }" />
            <button v-if="searchQuery" @click="clearSearch" class="rp-clear-btn">×</button>
          </div>
          <div class="rp-sort-wrap">
            <div class="rp-sort-dropdown" @click.stop="showSortMenu = !showSortMenu">
              <svg width="15" height="15" fill="none" viewBox="0 0 24 24" :stroke="isDark ? '#8e8e93' : '#6b7280'"
                stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 4h13M3 8h9m-4 4h5m-4 4h5m-4 4h5" />
              </svg>
              <span class="sort-label">{{ sortLabel }}</span>
              <svg width="9" height="9" fill="none" viewBox="0 0 24 24" :stroke="isDark ? '#8e8e93' : '#9ca3af'"
                stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </div>
            <Transition name="sort-fade">
              <div v-if="showSortMenu" class="sort-dropdown-panel">
                <button v-for="(lbl, mode) in sortOptions" :key="mode" class="sort-option"
                  :class="{ active: sortMode === mode }" @click.stop="selectSort(mode as string)">{{ lbl }}</button>
              </div>
            </Transition>
          </div>
        </div>

        <!-- 筛选面板（类型 + 类别 + 统计条） -->
        <RecordFilter
          v-model:filterType="filterType"
          v-model:selectedCategoryIds="selectedCategoryIds"
          v-model:selectedTagId="selectedTagId"
          :categories="store.categories"
          :allTags="allTags"
          :allRecordsCache="store.allRecordsCache"
          :totalCount="store.searchTotal"
          :monthIncome="monthIncome"
          :monthExpense="monthExpense"
          :isDark="isDark"
          @search="doSearch"
        />
      </div>

      <!-- 列表区域 -->
      <div class="rp-body" ref="bodyEl">
        <RecordCardList
          :filteredRecords="filteredRecords"
          :hasActiveFilters="hasActiveFilters"
          :isLoading="isLoading"
          :isLoadingMore="store.searchLoading"
          :hasMore="store.searchHasMore"
          :categories="store.categories"
          :accounts="accounts"
          :sortMode="sortMode"
          :isDark="isDark"
          :deleteRec="deleteConfirmRec"
          @edit="handleEdit"
          @delete="handleDelete"
          @confirmDelete="confirmDelete"
          @cancelDelete="deleteConfirmRec = null"
        />

        <!-- 加载更多观察器 -->
        <div ref="loadMoreEl" style="height:1px"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore, type Record as AccountRecord, type TagInfo } from '../stores/accounting'
import { GetAll as GetAllAccounts } from '../../wailsjs/go/service/AccountService'
import { SearchRecords as SearchRecordsAPI } from '../../wailsjs/go/service/RecordService'
import { GetAllTags as GetAllTagsAPI } from '../../wailsjs/go/service/TagService'
import RecordFilter from '../components/RecordFilter.vue'
import RecordCardList from '../components/RecordCardList.vue'

const router = useRouter()
const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

// ========== 账户数据 ==========
interface AccountItem { id: number; name: string; icon: string }
const accounts = ref<AccountItem[]>([])

async function loadAccounts() {
  try {
    const raw = await GetAllAccounts()
    accounts.value = (raw || []).map((a: any) => ({ id: a.ID || a.id, name: a.Name || a.name, icon: a.Icon || a.icon }))
  } catch { }
}

// ========== 排序 ==========
const showSortMenu = ref(false)
const sortMode = ref('time_desc')
const sortOptions: Record<string, string> = { time_desc: '最新', time_asc: '最早', amount_desc: '金额↓', amount_asc: '金额↑' }
const sortLabel = computed(() => sortOptions[sortMode.value] || '最新')

function selectSort(mode: string) {
  sortMode.value = mode
  showSortMenu.value = false
}

// ========== 搜索 & 筛选 ==========
const searchQuery = ref('')
const isLoading = ref(false)
const filterType = ref<'all' | 'income' | 'expense'>('all')
const selectedCategoryIds = ref<number[]>([])
const selectedTagId = ref(0)
const allTags = ref<TagInfo[]>([])

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => doSearch(), 300)
}

function clearSearch() {
  searchQuery.value = ''
  doSearch()
}

async function doSearch() {
  isLoading.value = true
  try {
    await store.searchRecords({
      keyword: searchQuery.value,
      type: filterType.value === 'all' ? '' : filterType.value,
      category: '',
      tag_id: selectedTagId.value,
    }, true)
  } finally {
    isLoading.value = false
  }
}

// ========== 数据处理 ==========
const filteredRecords = computed(() => {
  let list = [...store.records]

  if (filterType.value !== 'all') {
    list = list.filter(r => r.type === filterType.value)
  }

  if (selectedCategoryIds.value.length > 0) {
    list = list.filter(r => {
      const cat = store.categories.find(c => c.name === r.category)
      return cat && selectedCategoryIds.value.includes(cat.id)
    })
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(r =>
      r.category.toLowerCase().includes(q) ||
      (r.note || '').toLowerCase().includes(q)
    )
  }

  switch (sortMode.value) {
    case 'time_asc':
      list.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()); break
    case 'amount_desc':
      list.sort((a, b) => b.amount - a.amount); break
    case 'amount_asc':
      list.sort((a, b) => a.amount - b.amount); break
    default:
      list.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()); break
  }

  return list
})

const hasActiveFilters = computed(() =>
  filterType.value !== 'all' || selectedCategoryIds.value.length > 0 || selectedTagId.value > 0 || !!searchQuery.value.trim()
)

// 收支汇总：基于全量月度缓存
const monthIncome = computed(() => store.allRecordsCache.filter((r: AccountRecord) => r.type === 'income').reduce((s, r) => s + r.amount, 0))
const monthExpense = computed(() => store.allRecordsCache.filter((r: AccountRecord) => r.type === 'expense').reduce((s, r) => s + r.amount, 0))

// ========== 删除 ==========
const deleteConfirmRec = ref<AccountRecord | null>(null)

function handleEdit(rec: AccountRecord) {
  window.dispatchEvent(new CustomEvent('piggy:edit-record', { detail: rec }))
}

function handleDelete(rec: AccountRecord) {
  deleteConfirmRec.value = rec
}

async function confirmDelete() {
  if (!deleteConfirmRec.value) return
  const id = deleteConfirmRec.value.id
  deleteConfirmRec.value = null
  try { await store.deleteRecord(id); doSearch() } catch { }
}

// ========== 无限滚动 ==========
const loadMoreEl = ref<HTMLElement | null>(null)
const bodyEl = ref<HTMLElement | null>(null)

let observer: IntersectionObserver | null = null

onMounted(async () => {
  await loadAccounts()
  // 加载标签列表
  try {
    const tags = await GetAllTagsAPI()
    allTags.value = (tags || []).map((t: any) => ({ id: t.id, name: t.name, color: t.color }))
  } catch (e) {
    console.warn('加载标签列表失败', e)
  }
  // 加载全量记录缓存（胶囊筛选依赖，每次进入页面都刷新）
  try {
    const month = store.currentMonth
    const cacheResult = await SearchRecordsAPI({
      start_date: month + '-01',
      end_date: month + '-31',
      type: '', category: '', category_ids: '', keyword: '',
      account_id: 0, tag_id: 0, page: 1, limit: 9999,
    })
    store.fillAllRecordsCache(cacheResult?.records ?? [])
  } catch (e) {
    console.warn('加载全量记录缓存失败', e)
  }
  doSearch()

  observer = new IntersectionObserver(
    entries => {
      if (entries[0]?.isIntersecting && store.searchHasMore && !store.searchLoading) {
        store.loadMoreRecords()
      }
    },
    { rootMargin: '100px' }
  )
  nextTick(() => {
    if (loadMoreEl.value) observer?.observe(loadMoreEl.value)
  })
})

onUnmounted(() => {
  observer?.disconnect()
  if (debounceTimer) clearTimeout(debounceTimer)
})

watch(loadMoreEl, el => {
  if (el && observer) { observer.disconnect(); observer.observe(el) }
})

// 清除排序菜单
watch(showSortMenu, (val) => {
  if (val) {
    const handler = () => { showSortMenu.value = false }
    setTimeout(() => document.addEventListener('click', handler, { once: true }), 0)
  }
})

// ========== 样式计算属性 ==========
const headerBg = computed(() => ({ background: isDark.value ? '#1c1c1e' : '#fff' }))
const searchBoxBg = computed(() => ({ background: isDark.value ? '#2c2c2e' : '#f3f4f6' }))
</script>

<style scoped>
/* ====== 页面布局 ====== */
.records-page {
  height: 100%;
  overflow: hidden;
}

.records-page :deep(.page-content) {
  max-width: none;
  padding: 16px 20px;
  overflow: hidden;
}

.rp-content {
  width: 100%;
  max-width: 580px;
  margin: 0 auto;
  border-radius: 16px;
  background: v-bind('isDark ? "#1c1c1e" : "#fff"');
  box-shadow: v-bind('isDark ? "none" : "0 2px 12px rgba(0,0,0,0.06)"');
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: calc(100vh - var(--header-height, 56px) - 32px);
}

.rp-header {
  flex-shrink: 0;
  padding: 14px 16px 10px;
  border-radius: 16px 16px 0 0;
  overflow: visible;
}

.rp-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.rp-back-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: v-bind('isDark ? "#f5f5f7" : "#1f2937"');
  transition: background 0.15s;
}

.rp-back-btn:hover {
  background: v-bind('isDark ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.05)"');
}

.rp-title {
  font-size: 17px;
  font-weight: 600;
  flex-shrink: 0;
  color: v-bind('isDark ? "#f5f5f7" : "#111827"');
}

/* 搜索行 */
.rp-search-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.rp-search-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 9px 12px;
  border-radius: 12px;
}

.rp-search-box input {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  font-size: 14px;
  min-width: 0;
}

.rp-clear-btn {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: none;
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  flex-shrink: 0;
  color: v-bind('isDark ? "#98989d" : "#9ca3af"');
  background: v-bind('isDark ? "rgba(255,255,255,0.1)" : "#e5e7eb"');
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 排序 */
.rp-sort-wrap {
  position: relative;
  flex-shrink: 0;
}

.rp-sort-dropdown {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 9px 12px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  transition: opacity 0.15s;
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
  background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"');
}

.rp-sort-dropdown:hover {
  opacity: 0.7;
}

.sort-label {
  font-size: 13px;
  font-weight: 500;
}

.sort-dropdown-panel {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 110px;
  border-radius: 12px;
  padding: 4px;
  z-index: 100;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"');
}

.sort-option {
  display: block;
  width: 100%;
  text-align: left;
  padding: 9px 14px;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  background: transparent;
  color: v-bind('isDark ? "#e5e5ea" : "#374151"');
  transition: all 0.15s ease;
}

.sort-option:hover {
  background: v-bind('isDark ? "rgba(168,85,247,0.15)" : "rgba(168,85,247,0.08)"');
}

.sort-option.active {
  background: v-bind('isDark ? "rgba(168,85,247,0.2)" : "rgba(168,85,247,0.12)"');
  color: #a855f7;
  font-weight: 600;
}

/* 下拉动画 */
.sort-fade-enter-active {
  transition: all 0.14s ease-out;
}

.sort-fade-leave-active {
  transition: all 0.10s ease-in;
}

.sort-fade-enter-from {
  opacity: 0;
  transform: translateY(-4px);
}

.sort-fade-leave-to {
  opacity: 0;
  transform: translateY(-3px);
}

/* 列表体 */
.rp-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 4px 12px;
  -webkit-overflow-scrolling: touch;
  display: flex;
  flex-direction: column;
}

.rp-body::-webkit-scrollbar {
  display: none;
}
</style>
