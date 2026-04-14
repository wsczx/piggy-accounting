<template>
  <div class="page-container">
    <!-- 加载状态 -->
    <div v-if="store.loading" class="loading-state">
      <div class="loading-center">
        <div class="loading-pig">🐷</div>
        <p class="loading-text">加载中...</p>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="store.error" class="error-state">
      <div class="error-center">
        <div class="error-emoji">😿</div>
        <p class="error-text">{{ store.error }}</p>
        <button @click="store.init()" class="retry-btn">重试</button>
      </div>
    </div>

    <!-- 主内容：左右两栏布局（HomeView 覆盖 page-content 的 480px 限制） -->
    <div v-else class="page-content home-layout">
      <!-- 左栏：余额 + 最近账单（含内嵌快捷） -->
      <div class="home-left">
        <BalanceCard />

        <section class="records-section" style="animation-delay:0.1s;">
          <header class="sec-header">
            <h3 class="sec-title">最近账单</h3>
            <QuickActions @action-click="handleQuickAction" />
          </header>

          <!-- <div class="search-box">
            <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" style="color:#9ca3af;flex-shrink:0"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
            <input v-model="searchQuery" type="text" placeholder="搜索账单..." />
          </div> -->

          <div class="records-list">
            <div v-if="displayRecords.length === 0" class="empty-state">
              <span style="font-size:32px;">📝</span>
              <p>暂无账单记录</p>
            </div>
            <template v-else>
              <div v-for="r in displayRecords" :key="r.id" class="record-item" @click="handleEditRecord(r)">
                <div class="rec-left">
                  <span class="rec-icon" :style="{ background: getCategoryColor(r.category) + '22' }">
                    {{ _getCategoryIcon(r.category) }}
                  </span>
                  <div class="rec-info">
                    <div class="rec-name">{{ r.category }}<span v-if="r.note" class="rec-note">· {{ r.note }}</span>
                    </div>
                    <div class="rec-meta">
                      <span class="rec-date">{{ formatDate(r.date) }}</span>
                      <span v-for="tag in (r.tags || [])" :key="tag.id" class="rec-tag-badge"
                        :style="{ background: tag.color + '22', color: tag.color, borderColor: tag.color + '44' }">
                        {{ tag.name }}
                      </span>
                    </div>
                  </div>
                </div>
                <span class="rec-amount" :class="r.type">
                  {{ r.type === 'income' ? '+' : '-' }}¥{{ formatAmount(r.amount) }}
                </span>
              </div>
            </template>
          </div>

          <footer v-if="filteredRecords.length > 7" class="view-all">
            <button @click="router.push('/records')">查看全部账单 →</button>
          </footer>
        </section>
      </div>

      <!-- 右栏：任务面板 -->
      <aside class="home-right">
        <div class="task-panel">
          <header class="tp-header">
            <h3 class="tp-title">待办任务</h3>
            <button @click="openAddTask" class="icon-btn primary">＋</button>
          </header>

          <div class="tp-stats">
            <div class="tp-stat tp-stat-pending"><b>{{ pendingCount }}</b><span>待处理</span></div>
            <div class="tp-stat tp-stat-today"><b>{{ todayCount }}</b><span>今日到期</span></div>
            <div class="tp-stat tp-stat-overdue"><b>{{ overdueCount }}</b><span>已逾期</span></div>
          </div>

          <div class="tp-list">
            <div v-if="tasks.length === 0" class="empty-state">
              <span style="font-size:32px;">✅</span>
              <p>暂无待办任务</p>
            </div>
            <template v-else>
              <template v-for="g in taskGroups" :key="g.label">
                <div v-if="g.tasks.length > 0" class="tg-group">
                  <div class="tg-label" :style="{ color: g.color }">{{ g.label }}</div>
                  <div v-for="t in g.tasks" :key="t.id" class="tg-row" @click="openEditTask(t)">
                    <span class="tg-check" @click.stop="toggleTask(t.id)">
                      <i v-if="t.completed">✓</i>
                    </span>
                    <div class="tg-body">
                      <div class="tg-name" :class="{ done: t.completed }">{{ t.title }}</div>
                      <div class="tg-meta">
                        <span :class="dateClass(t)">{{ dateLabel(t) }}</span>
                        <span v-if="t.amount" class="tg-amt">¥{{ formatAmount(t.amount) }}</span>
                      </div>
                    </div>
                    <button @click.stop="confirmDelete(t)" class="tg-del">🗑️</button>
                  </div>
                </div>
              </template>
            </template>
          </div>

          <footer v-if="tasks.length > 5" class="tp-view-all">
            <button @click="router.push('/tasks')">查看全部任务 →</button>
          </footer>
        </div>
      </aside>
    </div>

    <!-- 通用确认弹窗 -->
    <ConfirmModal v-model="showDeleteConfirm" icon="🗑️" title="删除任务" :message="`确定要删除「${deletingTitle}」吗？`" dangerous
      @confirm="doDelete" />

    <TaskForm v-model="showTaskForm" :initial-title="editFormData?.title" :initial-due-date="editFormData?.due_date"
      :initial-amount="editFormData?.amount" @save="onTaskSave"
      @toast="(msg: string) => log.error(msg)" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore, type Record } from '../stores/accounting'
import { formatAmount, formatDate } from '../utils/formatters'
import { getCategoryIcon, getCategoryColor } from '../utils/category'
import { calcDaysLeft } from '../utils/date'
import type { TaskItem } from '../types'
import { GetAll as GetAllTasks, Create as CreateTask, ToggleComplete as ToggleTask, Delete as DeleteTask, Update as UpdateTask } from '../../wailsjs/go/service/TaskService'

// 组件
import BalanceCard from '../components/BalanceCard.vue'
import QuickActions from '../components/QuickActions.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import TaskForm from '../components/TaskForm.vue'
import type { QuickAction } from '../components/QuickActions.vue'
import { createLogger } from '../utils/logger'

const log = createLogger('Home')

const router = useRouter()
const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

// ========== 记录列表 ==========
const searchQuery = ref('')

const filteredRecords = computed(() => {
  let list = store.records
  const q = searchQuery.value.trim().toLowerCase()
  if (q) list = list.filter(r => r.note?.toLowerCase().includes(q) || r.category.toLowerCase().includes(q))
  return list
})

const displayRecords = computed(() => filteredRecords.value.slice(0, 7))

function handleEditRecord(record: Record) {
  // 通过 provide/inject 或事件总线打开编辑（当前使用 window hack 兼容）
  const evt = new CustomEvent('piggy:edit-record', { detail: record })
  window.dispatchEvent(evt)
}

// 类别图标 / 颜色（使用共享工具函数）
function _getCategoryIcon(name: string): string {
  return getCategoryIcon(name, store.categories)
}


// ========== 任务面板（TaskItem 从 ../types 导入） ==========

const tasks = ref<TaskItem[]>([])
const showTaskForm = ref(false)
const showDeleteConfirm = ref(false)
const editFormData = ref<Partial<TaskItem> | null>(null)
const deletingId = ref<number | null>(null)
const deletingTitle = ref('')

/** 计算距离到期的天数（复用共享工具） */
function calcDays(dueDate: string): number {
  return calcDaysLeft(dueDate)
}

const enriched = computed(() => tasks.value.map(t => ({ ...t, daysLeft: calcDays(t.due_date) })))
const pendingCount = computed(() => enriched.value.filter(t => !t.completed).length)
const todayCount = computed(() => enriched.value.filter(t => !t.completed && t.daysLeft === 0).length)
const overdueCount = computed(() => enriched.value.filter(t => !t.completed && t.daysLeft < 0).length)

const taskGroups = computed(() => [
  { label: '已逾期', color: '#ff453a', tasks: enriched.value.filter(t => !t.completed && t.daysLeft < 0) },
  { label: '今日到期', color: '#ff9f0a', tasks: enriched.value.filter(t => !t.completed && t.daysLeft === 0) },
  { label: '即将到期', color: '#30d158', tasks: enriched.value.filter(t => !t.completed && t.daysLeft > 0) },
  { label: '已完成', color: '#8e8e93', tasks: enriched.value.filter(t => t.completed) },
])

function dateLabel(t: TaskItem & { daysLeft: number }): string {
  if (t.completed) return '已完成'
  if (t.daysLeft < 0) return `逾期 ${Math.abs(t.daysLeft)} 天`
  if (t.daysLeft === 0) return '今天'
  return `还有 ${t.daysLeft} 天`
}

function dateClass(t: TaskItem & { daysLeft: number }): string {
  if (t.completed) return 'tg-date-done'
  if (t.daysLeft < 0) return 'tg-date-overdue'
  if (t.daysLeft === 0) return 'tg-date-today'
  return ''
}

async function loadTasks() {
  try { tasks.value = await GetAllTasks() }
  catch (e) { log.error('加载任务失败', e) }
}

async function toggleTask(id: number) {
  try { await ToggleTask(id); await loadTasks() }
  catch (e) { log.error('切换任务状态失败', e) }
}

// ---- 任务弹窗操作 ----
function openAddTask() {
  editFormData.value = null
  showTaskForm.value = true
}
function openEditTask(t: TaskItem) {
  editFormData.value = { ...t }
  showTaskForm.value = true
}
function confirmDelete(t: TaskItem) {
  deletingId.value = t.id
  deletingTitle.value = t.title
  showDeleteConfirm.value = true
}
async function doDelete() {
  if (!deletingId.value) return
  try {
    await DeleteTask(deletingId.value)
    await loadTasks()
  } catch (e) {
    log.error('删除任务失败', e)
    // TODO: 统一 toast 通知
  } finally {
    showDeleteConfirm.value = false
    deletingId.value = null
    deletingTitle.value = ''
  }
}
async function onTaskSave(data: { title: string; dueDate: string; amount: number }) {
  try {
    if (editFormData.value?.id) {
      await UpdateTask(editFormData.value.id, data.title, data.dueDate, data.amount)
    } else {
      await CreateTask(data.title, data.dueDate, data.amount)
    }
    showTaskForm.value = false
    editFormData.value = null
    await loadTasks()
  } catch (e) {
    log.error('保存任务失败', e)
    // TODO: 统一 toast 通知
  }
}

// ========== 初始化 & 路由 ==========
onMounted(() => {
  store.init()
  loadTasks()
  window.addEventListener('piggy:data-cleared', onDataCleared)
})

function onDataCleared() {
  store.init()
  loadTasks()
}

onUnmounted(() => {
  window.removeEventListener('piggy:data-cleared', onDataCleared)
})

function handleQuickAction(a: QuickAction) {
  switch (a.action) {
    case 'record': {
      // 打开记账面板
      const evt = new CustomEvent('piggy:add-record')
      window.dispatchEvent(evt)
      break
    }
    case 'navigate':
      if (a.route) router.push(a.route)
      break
    case 'popup':
      if (a.route) router.push(a.route)
      break
    default:
      if (a.route) router.push(a.route)
      break
  }
}
</script>

<style scoped>
/* ====== 布局：HomeView 是唯一左右两栏页面，覆盖 page-content 的 480px 限制 ====== */
.home-layout {
  max-width: none !important;
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 16px;
  height: 100%;
  overflow: hidden;
  padding-left: 20px;
  padding-right: 20px;
  box-sizing: border-box;
}

.home-left {
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  min-height: 0;
}

.home-right {
  overflow-y: auto;
  min-height: 0;
}

/* ====== 全局状态 ====== */
.loading-state,
.error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.loading-center {
  text-align: center;
}

.loading-pig {
  font-size: 48px;
  margin-bottom: 12px;
}

.loading-text {
  font-size: 14px;
  color: var(--text-secondary);
}

.error-center {
  text-align: center;
  padding: 20px;
}

.error-emoji {
  font-size: 48px;
  margin-bottom: 12px;
}

.error-text {
  font-size: 14px;
  color: var(--danger-color);
  margin-bottom: 12px;
}

.retry-btn {
  padding: 8px 20px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #fff;
  background: var(--accent-color);
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
}

.empty-state p {
  color: var(--text-muted, #9ca3af);
  margin-top: 8px;
}

/* ====== 区块标题栏 ====== */
.sec-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.sec-title {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary, #1f2937);
}

.filter-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 10px;
  border: none;
  font-size: 12px;
  cursor: pointer;
  color: #fff;
  background: var(--accent-color);
}

/* ====== 搜索框 ====== */
.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 12px;
  margin-bottom: 10px;
  background: var(--bg-input);
  flex-shrink: 0;
}

.search-box input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 14px;
  outline: none;
  color: var(--text-primary, #1f2937);
}

/* ====== 记录列表 ====== */
.records-section {
  border-radius: 16px;
  padding: 12px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}

.records-list {
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  min-height: 0;
}

.record-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  margin: 0 -12px;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease, background 0.15s ease;
  border-radius: 12px;
}

.record-item:last-child {
  border-bottom: none;
}

.record-item:hover {
  background: var(--bg-secondary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-hover);
}

.rec-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.rec-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.rec-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rec-name {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary, #1f2937);
}

.rec-note {
  font-size: 12px;
  opacity: 0.55;
}

.rec-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
  margin-top: 1px;
}

.rec-date {
  font-size: 11px;
  color: var(--text-muted, #9ca3af);
  flex-shrink: 0;
}

.rec-tag-badge {
  display: inline-flex;
  align-items: center;
  padding: 0 5px;
  height: 16px;
  border-radius: 4px;
  border: 1px solid transparent;
  font-size: 10px;
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
}

.rec-amount {
  font-size: 15px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
  margin-left: 8px;
}

.rec-amount.income {
  color: #30d158;
}

.rec-amount.expense {
  color: #ff453a;
}

.view-all {
  text-align: center;
  padding-top: 10px;
  border-top: 1px solid rgba(128, 128, 128, 0.08);
  flex-shrink: 0;
}

.view-all button {
  padding: 8px 16px;
  border: none;
  background: transparent;
  font-size: 13px;
  color: var(--accent-color);
  cursor: pointer;
}

/* ====== 任务面板 ====== */
.task-panel {
  border-radius: 16px;
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}

.tp-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.tp-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary, #1f2937);
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  transition: transform 0.15s;
}

.icon-btn.primary {
  background: var(--accent-color);
  color: #fff;
}

.icon-btn:hover {
  transform: scale(1.07);
}

.tp-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.tp-stat {
  border-radius: 12px;
  padding: 12px 4px;
  text-align: center;
  background: var(--bg-input);
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tp-stat b {
  font-size: 20px;
  font-weight: 700;
}

.tp-stat span {
  font-size: 11px;
  color: var(--text-muted, #6b7280);
}

/* 统计数字颜色（替代内联 style） */
.tp-stat-pending b { color: #6366f1; }
.tp-stat-today b { color: #30d158; }
.tp-stat-overdue b { color: #ff453a; }

.tp-list {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.tg-group {
  margin-bottom: 12px;
}

.tg-group:last-child {
  margin-bottom: 0;
}

.tg-label {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 6px;
  padding-left: 4px;
}

.tg-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: 12px;
  margin-bottom: 6px;
  transition: background 0.15s;
}

.tg-row:hover {
  background: rgba(128, 128, 128, 0.04);
}

.tg-check {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #6366f1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  font-size: 0;
}

.tg-check i {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #6366f1;
  color: #fff;
  font-size: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tg-body {
  flex: 1;
  min-width: 0;
  cursor: pointer;
}

.tg-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary, #1f2937);
}

.tg-name.done {
  text-decoration: line-through;
  opacity: 0.5;
}

.tg-meta {
  display: flex;
  gap: 6px;
  font-size: 11px;
}

.tg-meta span:first-child {
  color: var(--text-muted, #8e8e93);
}

/* Task date status classes (replaces inline :style) */
.tg-date-done { color: var(--status-done); }
.tg-date-overdue { color: var(--status-overdue); }
.tg-date-today { color: var(--status-today); }

.tg-amt {
  color: #6366f1;
  font-weight: 500;
}

.tg-del {
  padding: 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  opacity: 0.6;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.tg-del:hover {
  opacity: 1;
}

.tp-view-all {
  text-align: center;
  padding-top: 10px;
  border-top: 1px solid rgba(128, 128, 128, 0.08);
  flex-shrink: 0;
}

.tp-view-all button {
  padding: 8px 16px;
  border: none;
  background: transparent;
  font-size: 13px;
  color: var(--accent-color);
  cursor: pointer;
}
</style>
