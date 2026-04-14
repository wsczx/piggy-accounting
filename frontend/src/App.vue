<template>
  <div :class="['app-container', isDark ? 'dark' : '']">
    <!-- 顶部标题栏（可拖拽区域） -->
    <header class="app-header wails-drag">
      <!-- 左侧：Logo（macOS 上需要为系统交通灯留出空间） -->
      <div class="header-left" :class="{ 'mac-padding': isMac }">
        <div class="header-logo no-drag">
          <div class="logo-icon">
            <span class="logo-emoji">🐷</span>
          </div>
          <span class="header-title">猪猪记账</span>
        </div>
      </div>

      <!-- 中间：月份选择（点击弹出日历） -->
      <div class="no-drag header-month">
        <button @click="store.changeMonth(changeMonth(store.currentMonth, -1))" class="header-btn">
          <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <button class="header-month-btn no-drag" @click="toggleDatePicker">
          <span class="header-month-text">{{ store.monthName }}</span>
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none" class="month-caret">
            <path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"
              stroke-linejoin="round" />
          </svg>
        </button>
        <button @click="store.changeMonth(changeMonth(store.currentMonth, 1))" class="header-btn">
          <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <!-- 右侧：主题切换 + 窗口控制（非 macOS 显示自定义交通灯） -->
      <div class="header-right no-drag">
        <button @click="theme.toggle()" class="header-btn header-theme-btn">
          {{ isDark ? '☀️' : '🌙' }}
        </button>
        <!-- Windows / Linux 自定义交通灯 -->
        <div v-if="!isMac" class="traffic-lights">
          <button @click="handleMinimise" class="traffic-light traffic-light-minimise" title="最小化">
            <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
              <path d="M1 4H7" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
            </svg>
          </button>
          <button @click="handleMaximise" class="traffic-light traffic-light-maximise" title="最大化">
            <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
              <rect x="1" y="1" width="6" height="6" rx="1" stroke="currentColor" stroke-width="1.2" fill="none" />
            </svg>
          </button>
          <button @click="handleClose" class="traffic-light traffic-light-close" title="关闭">
            <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
              <path d="M1 1L7 7M7 1L1 7" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </div>
    </header>

    <!-- 日期选择弹窗 -->
    <transition name="date-picker">
      <div v-if="showDatePicker" class="date-picker-overlay no-drag" @click.self="showDatePicker = false">
        <div class="date-picker-popup" :class="isDark ? 'dark' : ''">
          <!-- 年份选择行 -->
          <div class="dp-year-row">
            <button class="dp-nav-btn" @click="pickerYear--">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <span class="dp-year-text">{{ pickerYear }}年</span>
            <button class="dp-nav-btn" @click="pickerYear++">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
          <!-- 12个月格 -->
          <div class="dp-months-grid">
            <button v-for="m in 12" :key="m" class="dp-month-btn" :class="{
              'dp-month-active': isActiveMonth(pickerYear, m),
              'dp-month-current': isCurrentMonth(pickerYear, m)
            }" @click="selectMonth(pickerYear, m)">{{ m }}月</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 主内容区域 -->
    <main class="app-main">
      <router-view />
    </main>

    <!-- 底部 Tab 导航 -->
    <TabBar @new-record="showRecordPanel = true" />

    <!-- 记账弹出面板 -->
    <RecordPanel :visible="showRecordPanel" :editing-record="editingRecord"
      @close="showRecordPanel = false; editingRecord = null" @done="showRecordPanel = false; editingRecord = null" />

    <!-- 首次启动测试数据引导弹窗 -->
    <WelcomeModal v-model:show="showWelcomeModal" :is-dark="isDark" @refresh="store.refreshAll()" @data-cleared="onDataCleared" />

    <!-- 统一提醒通知（队列模式，依次展示） -->
    <transition name="reminder-toast">
      <div v-if="activeNotification" class="reminder-toast" :class="[activeNotification.type, isDark ? 'dark' : '']"
        @click="dismissActiveNotification">
        <div class="reminder-icon">{{ activeNotification.icon }}</div>
        <div class="reminder-content">
          <div class="reminder-title">{{ activeNotification.title }}</div>
          <div class="reminder-message">{{ activeNotification.message }}</div>
        </div>
        <button class="reminder-close" @click.stop="dismissActiveNotification">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from './stores/theme'
import { useAccountingStore, type Record } from './stores/accounting'
import { changeMonth } from './utils/formatters'
import TabBar from './components/TabBar.vue'
import RecordPanel from './components/RecordPanel.vue'
import WelcomeModal from './components/modals/WelcomeModal.vue'
import { GetPlatform, HasTestData } from '../wailsjs/go/main/App'
import { createLogger } from './utils/logger'
import { EventsOn } from '../wailsjs/runtime/runtime'
import {
  InitializeNotifications,
  RequestNotificationAuthorization,
  IsNotificationAvailable,
  SendNotification
} from '../wailsjs/runtime/runtime'
import { GetReminderSettings } from '../wailsjs/go/service/ReminderService'

const router = useRouter()

const log = createLogger('App')

const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

const showRecordPanel = ref(false)
const editingRecord = ref<Record | null>(null)
const isMac = ref(false)
const showWelcomeModal = ref(false)

// ===== 提醒通知队列 =====
interface NotificationItem {
  icon: string
  title: string
  message: string
  type: 'daily' | 'task' | 'budget' | 'weekly'
  action?: string
  autoHideMs: number
}

const notificationQueue = ref<NotificationItem[]>([])
const isShowingNotification = ref(false)
const autoHideTimer = ref<number | null>(null)

const activeNotification = computed(() => {
  if (!isShowingNotification.value || notificationQueue.value.length === 0) return null
  return notificationQueue.value[0]
})

function enqueueNotification(item: NotificationItem) {
  notificationQueue.value.push(item)
  if (!isShowingNotification.value) {
    showNextNotification()
  }
}

function dismissActiveNotification() {
  const current = notificationQueue.value.shift()
  // 任务提醒不再跳转独立页面，仅关闭通知
  if (autoHideTimer.value) {
    clearTimeout(autoHideTimer.value)
    autoHideTimer.value = null
  }
  // 下一条延迟 300ms 显示，留出过渡动画时间
  setTimeout(() => showNextNotification(), 300)
}

function showNextNotification() {
  if (notificationQueue.value.length === 0) {
    isShowingNotification.value = false
    return
  }
  isShowingNotification.value = true
  const item = notificationQueue.value[0]
  autoHideTimer.value = window.setTimeout(() => {
    dismissActiveNotification()
  }, item.autoHideMs)
}

// 发送系统原生通知（需要用户在提醒设置中开启，且 App 已获得系统权限）
async function sendSystemNotification(item: NotificationItem) {
  try {
    if (!systemNotificationReady) return
    // 实时读取后端设置（避免缓存不同步）
    const settings = await GetReminderSettings()
    if (!settings?.system_notification_enabled) return
    await SendNotification({
      id: `piggy-${item.type}-${Date.now()}`,
      title: item.title,
      body: item.message,
    })
    log.info('系统通知已发送', 'title:', item.title)
  } catch (e) {
    log.warn('发送系统通知失败', e)
  }
}

// ===== 事件监听 =====

// 取消事件监听的函数
let unsubscribeReminder: (() => void) | null = null
let unsubscribeTaskReminder: (() => void) | null = null
let unsubscribeBudgetAlert: (() => void) | null = null
let unsubscribeWeeklySummary: (() => void) | null = null

// 系统通知是否可用（权限 + 服务就绪）
let systemNotificationReady = false

// 日期选择器
const showDatePicker = ref(false)
const pickerYear = ref(new Date().getFullYear())

onMounted(async () => {
  log.info('App mounted, 开始初始化')

  // 获取平台
  try {
    const platform = await GetPlatform()
    isMac.value = platform === 'darwin'
    log.info('平台检测完成', 'platform:', platform, 'isMac:', isMac.value)
  } catch (e) {
    log.warn('平台检测失败（开发模式？）', e)
    isMac.value = false
  }

  // 检测是否为新创建的账本（含测试数据），弹出引导提示
  try {
    const hasTestData = await HasTestData()
    if (hasTestData) {
      showWelcomeModal.value = true
      log.info('新账本含测试数据，弹出引导提示')
    }
  } catch (e) {
    log.warn('检测测试数据状态失败', e)
  }

  // 初始化系统通知服务
  try {
    const available = await IsNotificationAvailable()
    if (available) {
      await InitializeNotifications()
      // macOS 需要请求通知权限
      const authorized = await RequestNotificationAuthorization()
      systemNotificationReady = authorized
      log.info('系统通知初始化', 'available:', available, 'authorized:', authorized)
    } else {
      log.info('系统通知不可用（开发模式？）')
    }
  } catch (e) {
    log.warn('系统通知初始化失败（开发模式下预期行为）', e)
  }

  // 禁用右键菜单（macOS 和 Windows 都禁用）
  document.addEventListener('contextmenu', (e) => e.preventDefault())

  // 监听每日提醒事件
  unsubscribeReminder = EventsOn('daily_reminder', (data: { time: string, message: string }) => {
    log.info('收到每日提醒', data)
    const item: NotificationItem = {
      icon: '🔔', title: '记账提醒', message: data.message || '记得记账哦',
      type: 'daily', autoHideMs: 8000
    }
    enqueueNotification(item)
    sendSystemNotification(item)
  })

  // 监听任务提醒事件
  unsubscribeTaskReminder = EventsOn('task_reminder', (data: { count: number, tasks: { title: string }[] }) => {
    log.info('收到任务提醒', data)
    let message = `您有 ${data.count} 个待办任务即将到期`
    if (data.tasks?.length > 0) {
      const names = data.tasks.slice(0, 3).map(t => t.title).join('、')
      message = `${names}${data.tasks.length > 3 ? ' 等' : ''} 即将到期`
    }
    const item: NotificationItem = {
      icon: '✅', title: '待办任务提醒', message,
      type: 'task', autoHideMs: 10000
    }
    enqueueNotification(item)
    sendSystemNotification(item)
  })

  // 监听预算预警事件
  unsubscribeBudgetAlert = EventsOn('budget_alert', (data: { type: string, percentage: number, message: string }) => {
    log.info('收到预算预警', data)
    const label = data.type === 'monthly' ? '月预算' : '年预算'
    const item: NotificationItem = {
      icon: '💰', title: `⚠️ ${label}预警`,
      message: data.message || `${label}已使用 ${data.percentage.toFixed(1)}%`,
      type: 'budget', autoHideMs: 10000
    }
    enqueueNotification(item)
    sendSystemNotification(item)
  })

  // 监听周汇总事件
  unsubscribeWeeklySummary = EventsOn('weekly_summary', (data: { message: string, total_expense: number, balance: number, top_category: string }) => {
    log.info('收到周汇总', data)
    const msg = `${data.message || '周消费汇总'}\n支出 ¥${data.total_expense.toFixed(0)} · 结余 ¥${data.balance.toFixed(0)}${data.top_category ? ` · 最高: ${data.top_category}` : ''}`
    const item: NotificationItem = {
      icon: '📊', title: '周消费汇总', message: msg,
      type: 'weekly', autoHideMs: 12000
    }
    enqueueNotification(item)
    sendSystemNotification(item)
  })

  // 监听来自子组件的编辑/新增记录请求（替代 window hack）
  window.addEventListener('piggy:edit-record', ((e: CustomEvent) => {
    editRecord(e.detail)
  }) as EventListener)

  window.addEventListener('piggy:add-record', (() => {
    addRecord()
  }) as EventListener)

  log.info('App 初始化完成')
})

onUnmounted(() => {
  // 取消 Wails 事件监听
  if (unsubscribeReminder) {
    unsubscribeReminder()
  }
  if (unsubscribeTaskReminder) {
    unsubscribeTaskReminder()
  }
  if (unsubscribeBudgetAlert) {
    unsubscribeBudgetAlert()
  }
  if (unsubscribeWeeklySummary) {
    unsubscribeWeeklySummary()
  }
  // 取消自定义 DOM 事件监听
  window.removeEventListener('piggy:edit-record', editRecord as unknown as EventListener)
  window.removeEventListener('piggy:add-record', addRecord as unknown as EventListener)
})

// 日期选择器
function toggleDatePicker() {
  if (!showDatePicker.value) {
    // 打开时同步到当前月份
    const [y] = store.currentMonth.split('-')
    pickerYear.value = parseInt(y)
  }
  showDatePicker.value = !showDatePicker.value
}

function isActiveMonth(year: number, month: number) {
  const m = String(month).padStart(2, '0')
  return store.currentMonth === `${year}-${m}`
}

function isCurrentMonth(year: number, month: number) {
  const now = new Date()
  return now.getFullYear() === year && (now.getMonth() + 1) === month
}

function selectMonth(year: number, month: number) {
  const m = String(month).padStart(2, '0')
  store.changeMonth(`${year}-${m}`)
  showDatePicker.value = false
}

function editRecord(record: Record) {
  editingRecord.value = record
  showRecordPanel.value = true
}

function addRecord() {
  editingRecord.value = null
  showRecordPanel.value = true
}

function quickAddRecord(cat: { name: string; icon: string; type: string }) {
  editingRecord.value = null
  // 这里可以预设分类，等 RecordPanel 支持预填
  showRecordPanel.value = true
}

// 测试数据清空后，广播事件通知子组件刷新
function onDataCleared() {
  window.dispatchEvent(new CustomEvent('piggy:data-cleared'))
  log.info('已广播数据清空事件')
}

// 注意：piggy:edit-record 和 piggy:add-record 事件监听
// 已在上方主 onMounted 块（第258-264行）中注册，此处删除重复注册以避免双次触发

// 窗口控制（Windows 自定义标题栏用）
async function handleMinimise() {
  const { WindowMinimise } = await import('../wailsjs/runtime/runtime')
  WindowMinimise()
}
async function handleMaximise() {
  const { WindowToggleMaximise } = await import('../wailsjs/runtime/runtime')
  WindowToggleMaximise()
}
async function handleClose() {
  const { Quit } = await import('../wailsjs/runtime/runtime')
  Quit()
}
</script>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  gap: 0;
  padding-left: 16px;
}

/* macOS：系统交通灯占据左侧约 70px，Logo 需要让出空间 */
.header-left.mac-padding {
  padding-left: 76px;
}

/* 自定义交通灯按钮（Windows） */
.traffic-lights {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 6px;
  padding-right: 14px;
}

.traffic-light {
  width: 13px;
  height: 13px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: all 0.15s;
  background: transparent;
}

.traffic-light svg {
  opacity: 0;
  transition: opacity 0.15s;
}

/* 鼠标悬停标题栏时显示图标 */
.app-header:hover .traffic-light svg {
  opacity: 1;
}

/* 关闭 - 红色 */
.traffic-light-close {
  background: #ff5f57;
  box-shadow: 0 0 0 0.5px rgba(255, 95, 87, 0.4);
}

.traffic-light-close:hover {
  background: #ff3b30;
}

.traffic-light-close svg {
  color: #4a0002;
}

/* 最小化 - 黄色 */
.traffic-light-minimise {
  background: #febc2e;
  box-shadow: 0 0 0 0.5px rgba(254, 188, 46, 0.4);
}

.traffic-light-minimise:hover {
  background: #f5a623;
}

.traffic-light-minimise svg {
  color: #5a3e00;
}

/* 最大化 - 绿色 */
.traffic-light-maximise {
  background: #28c840;
  box-shadow: 0 0 0 0.5px rgba(40, 200, 64, 0.4);
}

.traffic-light-maximise:hover {
  background: #1db954;
}

.traffic-light-maximise svg {
  color: #006500;
}

/* 暗色模式下交通灯保持原色 */
.dark .traffic-light-close {
  background: #ff5f57;
}

.dark .traffic-light-minimise {
  background: #febc2e;
}

.dark .traffic-light-maximise {
  background: #28c840;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 2px;
}

.header-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 4px;
}

.logo-icon {
  width: 26px;
  height: 26px;
  border-radius: 8px;
  background: linear-gradient(135deg, #f472b6, #a855f7);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(168, 85, 247, 0.3);
}
.logo-emoji { font-size: 13px; }

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
  letter-spacing: -0.2px;
}

.dark .header-title {
  color: #f5f5f7;
}

.header-month {
  display: flex;
  align-items: center;
  gap: 2px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

/* 月份按钮（可点击弹出日历） */
.header-month-btn {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 4px 8px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  background: transparent;
  transition: background 0.15s;
}

.header-month-btn:hover {
  background: rgba(0, 0, 0, 0.06);
}

.dark .header-month-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.header-month-text {
  font-size: 13px;
  font-weight: 500;
  min-width: 76px;
  text-align: center;
  color: #374151;
}

.dark .header-month-text {
  color: #e5e5ea;
}

.month-caret {
  color: #9ca3af;
  transition: transform 0.15s;
  flex-shrink: 0;
}

.header-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  cursor: pointer;
  background: transparent;
  transition: all 0.15s;
  color: #9ca3af;
  font-size: 14px;
  padding: 0;
}

.header-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #374151;
}

.dark .header-btn {
  color: #98989d;
}

.dark .header-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #f5f5f7;
}

.header-theme-btn {
  font-size: 15px;
}

/* 让 header 使用 relative 定位，以便月份居中 */
.app-header {
  position: relative;
}

/* ===== 日期选择弹窗 ===== */
.date-picker-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 54px;
}

.date-picker-popup {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  width: 240px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15), 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.date-picker-popup.dark {
  background: #2c2c2e;
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.dp-year-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.dp-year-text {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
}

.date-picker-popup.dark .dp-year-text {
  color: #f5f5f7;
}

.dp-nav-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  transition: background 0.15s;
  padding: 0;
}

.dp-nav-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  color: #1d1d1f;
}

.date-picker-popup.dark .dp-nav-btn {
  color: #98989d;
}

.date-picker-popup.dark .dp-nav-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #f5f5f7;
}

.dp-months-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
}

.dp-month-btn {
  padding: 8px 0;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  background: transparent;
  color: #374151;
  transition: all 0.15s;
  text-align: center;
}

.dp-month-btn:hover {
  background: rgba(99, 102, 241, 0.08);
  color: #6366f1;
}

.date-picker-popup.dark .dp-month-btn {
  color: #d1d5db;
}

.date-picker-popup.dark .dp-month-btn:hover {
  background: rgba(129, 140, 248, 0.12);
  color: #818cf8;
}

/* 今天所在月份 - 下划线标记 */
.dp-month-btn.dp-month-current {
  color: #6366f1;
  font-weight: 600;
}

.date-picker-popup.dark .dp-month-btn.dp-month-current {
  color: #818cf8;
}

/* 当前选中月份 - 高亮背景 */
.dp-month-btn.dp-month-active {
  background: linear-gradient(135deg, #f472b6, #a855f7);
  color: #fff !important;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(168, 85, 247, 0.3);
}

/* 日期选择器出现动画 */
.date-picker-enter-active {
  animation: dpIn 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.date-picker-leave-active {
  animation: dpIn 0.15s ease-in reverse;
}

@keyframes dpIn {
  from {
    opacity: 0;
    transform: translateY(-8px) scale(0.96);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* ===== 每日提醒通知 ===== */
.reminder-toast {
  position: fixed;
  top: 60px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15), 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(0, 0, 0, 0.06);
  min-width: 280px;
  max-width: 90vw;
  cursor: pointer;
}

.reminder-toast.dark {
  background: #2c2c2e;
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.reminder-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.reminder-content {
  flex: 1;
  min-width: 0;
}

.reminder-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
  margin-bottom: 2px;
}

.reminder-toast.dark .reminder-title {
  color: #f5f5f7;
}

.reminder-message {
  font-size: 13px;
  color: #6b7280;
  white-space: pre-line;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reminder-toast.dark .reminder-message {
  color: #98989d;
}

.reminder-close {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  transition: all 0.15s;
  flex-shrink: 0;
}

.reminder-close:hover {
  background: rgba(0, 0, 0, 0.06);
  color: #6b7280;
}

.reminder-toast.dark .reminder-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #d1d5db;
}

/* 提醒通知动画 */
.reminder-toast-enter-active {
  animation: reminderIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.reminder-toast-leave-active {
  animation: reminderOut 0.2s ease-in;
}

@keyframes reminderIn {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px) scale(0.9);
  }

  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0) scale(1);
  }
}

@keyframes reminderOut {
  from {
    opacity: 1;
    transform: translateX(-50%) translateY(0) scale(1);
  }

  to {
    opacity: 0;
    transform: translateX(-50%) translateY(-10px) scale(0.95);
  }
}
</style>
