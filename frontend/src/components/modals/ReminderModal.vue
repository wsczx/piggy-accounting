<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal modal-main">
      <div class="modal-header">
        <span class="modal-title">提醒设置</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 预算预警通知 -->
      <div v-if="budgetAlerts.length > 0" class="alert-section">
        <div class="alert-title">⚠️ 预算预警</div>
        <div v-for="alert in budgetAlerts" :key="alert.type" class="alert-item">
          <div class="alert-item-header">
            <span class="alert-type-label">{{ alert.type === 'monthly' ? '月预算' : '年预算' }}</span>
            <span class="alert-percentage">{{ alert.percentage.toFixed(1) }}%</span>
          </div>
          <div class="alert-progress">
            <div class="alert-progress-bar"
              :style="{ width: Math.min(alert.percentage, 100) + '%', background: alert.percentage >= 100 ? '#ff453a' : alert.percentage >= 80 ? '#ff9f0a' : '#30d158' }" />
          </div>
          <div class="alert-meta-row">
            <span class="alert-meta-spent">已花 ¥{{ alert.spent.toFixed(0) }}</span>
            <span class="alert-meta-remaining" :class="{ 'text-danger': alert.remaining < 0 }">
              {{ alert.remaining >= 0 ? '剩余 ¥' + alert.remaining.toFixed(0) : '超支 ¥' + Math.abs(alert.remaining).toFixed(0) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 全局通知设置 -->
      <div class="settings-section">
        <div class="section-title">通知方式</div>

        <!-- 弹窗提醒开关 -->
        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-icon">💬</span>
            <div>
              <div class="setting-name">弹窗提醒</div>
              <div class="setting-desc">在应用内显示提醒弹窗</div>
            </div>
          </div>
          <div @click="togglePopupEnabled"
            class="toggle-switch" :class="{ on: reminderSettings.popup_enabled }">
            <div class="toggle-thumb"></div>
          </div>
        </div>

        <!-- 系统通知开关 -->
        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-icon">🔔</span>
            <div>
              <div class="setting-name">系统通知</div>
              <div class="setting-desc">发送系统桌面通知，应用在后台也能看到</div>
            </div>
          </div>
          <div @click="toggleSystemNotification"
            class="toggle-switch" :class="{ on: reminderSettings.system_notification_enabled }">
            <div class="toggle-thumb"></div>
          </div>
        </div>

        <!-- Webhook 设置 -->
        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-icon">🔗</span>
            <div>
              <div class="setting-name">Webhook 通知</div>
              <div class="setting-desc">发送 HTTP 请求到指定地址</div>
            </div>
          </div>
          <div @click="toggleWebhookEnabled"
            class="toggle-switch" :class="{ on: reminderSettings.webhook_enabled }">
            <div class="toggle-thumb"></div>
          </div>
        </div>

        <!-- Webhook URL 输入 -->
        <div v-if="reminderSettings.webhook_enabled" class="webhook-url-section">
          <label class="form-label">Webhook 地址</label>
          <div class="url-input-row">
            <input v-model="reminderSettings.webhook_url" type="text"
              placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" @blur="saveReminderSettings"
              class="url-input" :class="{ 'warn-border': webhookUrlWarning }" />
            <button @click="testWebhook" :disabled="webhookTesting" class="test-btn">{{ webhookTesting ? '测试中...' : '测试' }}</button>
          </div>
          <div v-if="webhookUrlWarning" class="url-warning"><span>⚠️</span><span>{{ webhookUrlWarning }}</span></div>
          <div v-else class="url-hint">支持企业微信 / 飞书 / 钉钉群机器人，以及任意 HTTP POST 地址</div>
        </div>
      </div>

      <!-- 提醒设置列表 -->
      <div class="reminders-list">
        <div v-for="r in reminders" :key="r.id" class="reminder-card">
          <div class="reminder-card-header">
            <div class="reminder-info">
              <span class="reminder-icon">{{ getReminderIcon(r.type) }}</span>
              <div>
                <div class="reminder-name">{{ getReminderLabel(r.type, r.budget_type) }}</div>
                <div v-if="r.type === 'budget_warning'" class="reminder-detail">超过 {{ r.threshold }}% 时提醒</div>
                <div v-if="r.type === 'daily_reminder' && r.reminder_time" class="reminder-detail">每日 {{ r.reminder_time }} 提醒</div>
                <div v-if="r.type === 'task_reminder'" class="reminder-detail">到期前 {{ reminderSettings.task_reminder_days || 1 }} 天提醒</div>
              </div>
            </div>
            <div @click="toggleReminder(r)" class="toggle-switch toggle-sm" :class="{ on: r.enabled }">
              <div class="toggle-thumb"></div>
            </div>
          </div>

          <!-- 预算预警阈值滑块 -->
          <div v-if="r.type === 'budget_warning' && r.enabled" class="slider-section">
            <div class="slider-header">
              <span>预警阈值</span>
              <span class="slider-value" :class="{ 'val-danger': r.threshold >= 100, 'val-warn': r.threshold >= 80 && r.threshold < 100 }">{{ r.threshold }}%</span>
            </div>
            <input type="range" min="10" max="100" step="5" v-model.number="r.threshold"
              @change="(e) => updateReminderThreshold(r, Number((e.target as HTMLInputElement).value))"
              class="range-input" />
          </div>

          <!-- 每日提醒时间 -->
          <div v-if="r.type === 'daily_reminder' && r.enabled" class="time-btn-section">
            <label class="form-label">提醒时间</label>
            <div @click="openTimePicker(r)" class="time-picker-trigger">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" />
              </svg>
              <span>{{ r.reminder_time || '20:00' }}</span>
            </div>
          </div>

          <!-- 任务提醒提前天数 -->
          <div v-if="r.type === 'task_reminder' && r.enabled" class="slider-section">
            <div class="slider-header"><span>提前提醒天数</span><span class="slider-value">{{ reminderSettings.task_reminder_days || 1 }} 天</span></div>
            <input type="range" min="1" max="7" step="1" :value="reminderSettings.task_reminder_days || 1"
              @change="(e) => updateTaskReminderDays(Number((e.target as HTMLInputElement).value))"
              class="range-input" />
          </div>
        </div>
      </div>
    </div>

    <!-- 内嵌时间选择器 -->
    <div v-if="showTimePicker" class="modal-overlay" @click.self="showTimePicker = false">
      <div class="modal time-picker-modal" style="width:300px;">
        <div class="modal-header">
          <span class="modal-title">选择时间</span>
          <button @click="showTimePicker = false" class="close-btn">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>
        <div class="tp-body">
          <div class="tp-section">
            <label class="form-label">快速输入 (HH:MM)</label>
            <input v-model="timeInputValue" type="text" placeholder="20:00" maxlength="5" @input="onTimeInput" @keydown.enter="confirmTimePicker" class="tp-quick-input" />
          </div>
          <div class="tp-divider">
            <div class="divider-line"></div><span class="divider-text">或手动调整</span><div class="divider-line"></div>
          </div>
          <div class="tp-picker">
            <div class="tp-col">
              <button @click="adjustHour(1)" class="tp-arrow-btn"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg></button>
              <input v-model="selectedHour" type="text" maxlength="2" @input="onHourInput" @blur="validateHour" class="tp-num-input" />
              <button @click="adjustHour(-1)" class="tp-arrow-btn"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg></button>
            </div>
            <div class="tp-sep">:</div>
            <div class="tp-col">
              <button @click="adjustMinute(1)" class="tp-arrow-btn"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg></button>
              <input v-model="selectedMinute" type="text" maxlength="2" @input="onMinuteInput" @blur="validateMinute" class="tp-num-input" />
              <button @click="adjustMinute(-1)" class="tp-arrow-btn"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg></button>
            </div>
          </div>
          <div class="tp-section tp-mt">
            <label class="form-label">常用时间</label>
            <div class="quick-times">
              <button v-for="time in ['08:00', '09:00', '12:00', '14:00', '18:00', '20:00', '21:00', '22:00']" :key="time" @click="selectQuickTime(time)" class="quick-time-btn">{{ time }}</button>
            </div>
          </div>
          <div class="tp-actions">
            <button @click="showTimePicker = false" class="tp-cancel-btn">取消</button>
            <button @click="confirmTimePicker" class="tp-confirm-btn">确定</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { GetAllReminders, UpdateReminder, CheckBudgetAlerts, GetReminderSettings, UpdateReminderSettings } from '../../../wailsjs/go/service/ReminderService'
import { TestWebhook } from '../../../wailsjs/go/main/App'
import { createLogger } from '../../utils/logger'
import type { AccountingStoreLike } from '../../types'

const log = createLogger('Reminder')

interface ReminderItem {
  id: number; type: string; budget_type: string
  threshold: number; reminder_time: string; enabled: boolean
  message: string
}
interface BudgetAlert {
  type: string; percentage: number; budget: number; spent: number; remaining: number; message: string
}

const props = defineProps<{
  show: boolean
  isDark: boolean
  budgetAlerts: BudgetAlert[]
  store: AccountingStoreLike
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const reminders = ref<ReminderItem[]>([])
const reminderSettings = ref({ webhook_url: '', webhook_enabled: false, popup_enabled: true, system_notification_enabled: false, task_reminder_days: 1 })
const webhookTesting = ref(false)

const webhookUrlWarning = computed(() => {
  const url = reminderSettings.value.webhook_url?.trim() || ''
  if (!url) return ''
  if (url.includes('work.weixin.qq.com/wework_admin')) return '这是企业微信管理页地址，不是 Webhook 地址。请在群聊中添加机器人，然后复制机器人详情页的 Webhook 地址（格式：https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx）'
  if (url.includes('work.weixin.qq.com') && !url.includes('qyapi')) return '请使用企业微信群机器人的 Webhook 地址（格式：https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx），而不是管理页面地址'
  return ''
})

// 时间选择器状态
const showTimePicker = ref(false)
const editingReminder = ref<ReminderItem | null>(null)
const selectedHour = ref('20')
const selectedMinute = ref('00')
const timeInputValue = ref('20:00')

watch(() => props.show, async (val) => { if (val) { await Promise.all([loadReminders(), loadReminderSettings()]) } })

function showToast(msg: string, type: 'success' | 'error') { emit('toast', msg, type) }
async function loadReminders() { try { reminders.value = await GetAllReminders() } catch (e) { log.error('加载提醒设置失败', e) } }

async function loadReminderSettings() {
  try {
    const settings = await GetReminderSettings()
    if (settings) { reminderSettings.value = { webhook_url: settings.webhook_url || '', webhook_enabled: settings.webhook_enabled || false, popup_enabled: settings.popup_enabled !== false, system_notification_enabled: settings.system_notification_enabled || false, task_reminder_days: settings.task_reminder_days || 1 } }
  } catch (e) { log.error('加载提醒设置失败(2)', e) }
}

async function saveReminderSettings() {
  try { await UpdateReminderSettings(reminderSettings.value.webhook_url, reminderSettings.value.webhook_enabled, reminderSettings.value.popup_enabled, reminderSettings.value.system_notification_enabled, reminderSettings.value.task_reminder_days) }
  catch (e) { log.error('保存提醒设置失败', e) }
}

async function togglePopupEnabled() { reminderSettings.value.popup_enabled = !reminderSettings.value.popup_enabled; await saveReminderSettings(); showToast(reminderSettings.value.popup_enabled ? '已开启弹窗提醒' : '已关闭弹窗提醒', 'success') }
async function toggleSystemNotification() { reminderSettings.value.system_notification_enabled = !reminderSettings.value.system_notification_enabled; await saveReminderSettings(); showToast(reminderSettings.value.system_notification_enabled ? '已开启系统通知' : '已关闭系统通知', 'success') }
async function toggleWebhookEnabled() { reminderSettings.value.webhook_enabled = !reminderSettings.value.webhook_enabled; await saveReminderSettings(); showToast(reminderSettings.value.webhook_enabled ? '已开启 Webhook' : '已关闭 Webhook', 'success') }

async function testWebhook() {
  const url = reminderSettings.value.webhook_url?.trim()
  if (!url) { showToast('请先填写 Webhook 地址', 'error'); return }
  await saveReminderSettings(); webhookTesting.value = true
  try { await TestWebhook(url); showToast('✅ Webhook 测试成功！', 'success') }
  catch (e) { const msg = String(e).replace(/^Error: /, ''); showToast('Webhook 测试失败: ' + msg, 'error') }
  finally { webhookTesting.value = false }
}

async function toggleReminder(reminder: ReminderItem) {
  try { const newEnabled = !reminder.enabled; await UpdateReminder(reminder.id, newEnabled, reminder.threshold, reminder.message, reminder.reminder_time, reminderSettings.value.task_reminder_days || 1); reminder.enabled = newEnabled; showToast(newEnabled ? '已开启提醒' : '已关闭提醒', 'success'); await checkAlerts(); emit('refresh') }
  catch (e) { showToast('更新失败: ' + e, 'error') }
}

async function updateReminderThreshold(reminder: ReminderItem, threshold: number) {
  try { await UpdateReminder(reminder.id, reminder.enabled, threshold, reminder.message, reminder.reminder_time, reminderSettings.value.task_reminder_days || 1); reminder.threshold = threshold; showToast(`预警阈值已设置为 ${threshold}%`, 'success'); await checkAlerts(); emit('refresh') }
  catch (e) { showToast('更新失败: ' + e, 'error') }
}

async function updateTime(reminder: ReminderItem, time: string) {
  try { await UpdateReminder(reminder.id, reminder.enabled, reminder.threshold, reminder.message, time, reminderSettings.value.task_reminder_days || 1); reminder.reminder_time = time; showToast(`提醒时间已设置为 ${time}`, 'success') }
  catch (e) { showToast('更新失败: ' + e, 'error') }
}

async function updateTaskReminderDays(days: number) {
  reminderSettings.value.task_reminder_days = days
  await saveReminderSettings()
  showToast(`提前提醒天数已设置为 ${days} 天`, 'success')
}

async function checkAlerts() {
  try { await CheckBudgetAlerts(); emit('refresh') }
  catch (e) { log.error('检查预算预警失败', e) }
}

function getReminderLabel(type: string, budgetType: string) {
  if (type === 'budget_warning') return budgetType === 'monthly' ? '月预算预警' : '年预算预警'
  if (type === 'daily_reminder') return '每日提醒'; if (type === 'weekly_summary') return '周消费汇总'
  if (type === 'task_reminder') return '待办任务提醒'; return type
}
function getReminderIcon(type: string) {
  if (type === 'budget_warning') return '💰'; if (type === 'daily_reminder') return '📅'
  if (type === 'weekly_summary') return '📊'; if (type === 'task_reminder') return '✅'; return '🔔'
}

// 时间选择器
function openTimePicker(reminder: ReminderItem) {
  editingReminder.value = reminder; const time = reminder.reminder_time || '20:00'; const [h, m] = time.split(':')
  selectedHour.value = h || '20'; selectedMinute.value = m || '00'; timeInputValue.value = `${selectedHour.value}:${selectedMinute.value}`; showTimePicker.value = true
}
function onTimeInput(e: Event) {
  const cleaned = (e.target as HTMLInputElement).value.replace(/[^0-9:]/g, '')
  timeInputValue.value = cleaned
  const match = cleaned.match(/^(\d{1,2}):?(\d{0,2})$/)
  if (match) { const h = parseInt(match[1]); const m = parseInt(match[2] || '0'); if (h >= 0 && h <= 23) selectedHour.value = String(h).padStart(2, '0'); if (m >= 0 && m <= 59) selectedMinute.value = String(m).padStart(2, '0') }
}
function onHourInput(e: Event) { let h = parseInt((e.target as HTMLInputElement).value.replace(/[^0-9]/g, '')) || 0; if (h > 23) h = 23; selectedHour.value = String(h).padStart(2, '0'); updateTimeInput() }
function validateHour() { let h = parseInt(selectedHour.value) || 0; if (h > 23) h = 23; selectedHour.value = String(h).padStart(2, '0'); updateTimeInput() }
function adjustHour(delta: number) { let h = parseInt(selectedHour.value) || 0; h = (h + delta + 24) % 24; selectedHour.value = String(h).padStart(2, '0'); updateTimeInput() }
function onMinuteInput(e: Event) { let m = parseInt((e.target as HTMLInputElement).value.replace(/[^0-9]/g, '')) || 0; if (m > 59) m = 59; selectedMinute.value = String(m).padStart(2, '0'); updateTimeInput() }
function validateMinute() { let m = parseInt(selectedMinute.value) || 0; if (m > 59) m = 59; selectedMinute.value = String(m).padStart(2, '0'); updateTimeInput() }
function adjustMinute(delta: number) { let m = parseInt(selectedMinute.value) || 0; m = (m + delta + 60) % 60; selectedMinute.value = String(m).padStart(2, '0'); updateTimeInput() }
function updateTimeInput() { timeInputValue.value = `${selectedHour.value}:${selectedMinute.value}` }
function selectQuickTime(time: string) { const [h, m] = time.split(':'); selectedHour.value = h; selectedMinute.value = m; timeInputValue.value = time }

async function confirmTimePicker() {
  if (!editingReminder.value) return
  await updateTime(editingReminder.value, `${selectedHour.value}:${selectedMinute.value}`)
  showTimePicker.value = false; editingReminder.value = null
}
</script>

<style scoped>
/* ====== 弹窗基础 ====== */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
.modal { width: 100%; max-width: 400px; border-radius: var(--radius-xl); padding: 20px; background: var(--card-bg); max-height: 80vh; overflow-y: auto; }
.time-picker-modal { max-width: none; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }

.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: var(--text-secondary); transition: background .15s; }
.close-btn:hover { background: var(--bg-input); }

/* ====== 预算预警区域 ====== */
.alert-section { margin-bottom: 16px; padding: 12px; border-radius: var(--radius-lg); background: rgba(255,69,58,0.06); border: 1px solid rgba(255,69,58,0.15); }
.alert-title { font-size: 13px; font-weight: 600; color: #ff453a; margin-bottom: 10px; }

.alert-item { padding: 10px; border-radius: var(--radius-md); background: var(--card-bg); margin-bottom: 8px; border: 1px solid var(--border-color-light); }
.alert-item:last-child { margin-bottom: 0; }
.alert-item-header { display: flex; justify-content: space-between; align-items: center; }
.alert-type-label { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.alert-percentage { font-size: 13px; font-weight: 700; color: #ff453a; }

.alert-progress { width: 100%; height: 6px; border-radius: 3px; margin-top: 8px; background: var(--bg-input); }
.alert-progress-bar { height: 100%; border-radius: 3px; transition: width 0.3s; }

.alert-meta-row { display: flex; justify-content: space-between; margin-top: 6px; }
.alert-meta-spent { font-size: 12px; color: var(--text-secondary); }
.alert-meta-remaining { font-size: 12px; color: var(--text-secondary); }
.text-danger { color: #ff453a; }

/* ====== 设置区块 ====== */
.settings-section { margin-bottom: 20px; padding: 14px 12px; border-radius: var(--radius-lg); background: var(--bg-input); }
.section-title { font-size: 14px; font-weight: 600; margin-bottom: 12px; color: var(--text-primary); }

.setting-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.setting-row:last-child { margin-bottom: 0; }
.setting-info { display: flex; align-items: center; gap: 10px; }
.setting-icon { font-size: 18px; }
.setting-name { font-size: 14px; color: var(--text-primary); }
.setting-desc { font-size: 12px; color: var(--text-secondary); }

/* Toggle Switch */
.toggle-switch { width: 48px; height: 28px; border-radius: 14px; padding: 2px; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; background: var(--switch-track); flex-shrink: 0; }
.toggle-switch.on { background: var(--accent-color); }
.toggle-switch.toggle-sm { width: 44px; height: 26px; border-radius: 13px; }
.toggle-switch.toggle-sm .toggle-thumb { width: 22px; height: 22px; }
.toggle-switch.toggle-sm.on .toggle-thumb { transform: translateX(18px) !important; }
.toggle-thumb { width: 24px; height: 24px; border-radius: 50%; background: #fff; transition: all 0.2s; box-shadow: 0 1px 3px rgba(0,0,0,0.2); flex-shrink: 0; }

/* Webhook URL */
.webhook-url-section { margin-top: 12px; }
.form-label { font-size: 12px; margin-bottom: 6px; color: var(--text-secondary); display: block; }

.url-input-row { display: flex; gap: 8px; }
.url-input { flex: 1; padding: 10px 12px; border-radius: var(--radius-md); font-size: 14px; border: none; background: var(--card-bg); color: var(--text-primary); outline: none; box-sizing: border-box; border: 1.5px solid transparent; }
.url-input.warn-border { border-color: #f59e0b; }
.test-btn { padding: 10px 16px; border-radius: var(--radius-md); border: none; cursor: pointer; font-size: 13px; font-weight: 600; white-space: nowrap; background: var(--accent-color); color: #fff; transition: opacity 0.15s; flex-shrink: 0; }
.test-btn:hover { opacity: 0.85; }
.test-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.url-warning { font-size: 11px; margin-top: 6px; color: #f59e0b; display: flex; align-items: flex-start; gap: 4px; }
.url-hint { font-size: 11px; margin-top: 6px; color: var(--text-muted); }

/* ====== 提醒卡片列表 ====== */
.reminders-list { display: flex; flex-direction: column; gap: 8px; }

.reminder-card { display: flex; flex-direction: column; gap: 10px; padding: 14px 12px; border-radius: var(--radius-lg); background: var(--bg-input); }

.reminder-card-header { display: flex; align-items: center; justify-content: space-between; }
.reminder-info { display: flex; align-items: center; gap: 10px; min-width: 0; }
.reminder-icon { font-size: 20px; }
.reminder-name { font-size: 14px; font-weight: 500; color: var(--text-primary); }
.reminder-detail { font-size: 12px; color: var(--text-secondary); }

.slider-section { padding: 0 4px; }
.slider-header { display: flex; justify-content: space-between; font-size: 12px; margin-bottom: 6px; color: var(--text-secondary); }
.slider-value { font-weight: 600; color: var(--text-primary); }
.val-danger { color: #ff453a; }
.val-warn { color: #ff9f0a; }
.range-input { width: 100%; accent-color: var(--accent-color); cursor: pointer; }

.time-btn-section { padding: 0 4px; }
.time-picker-trigger { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-radius: var(--radius-md); cursor: pointer; width: fit-content; background: var(--card-bg); color: var(--text-primary); transition: opacity 0.15s; }
.time-picker-trigger:hover { opacity: 0.8; }
.time-picker-trigger span { font-size: 14px; font-weight: 500; }
.time-picker-trigger svg { color: var(--accent-color); }

/* ====== 时间选择器内部弹窗 ====== */
.tp-body { padding: 20px; }
.tp-section { margin-bottom: 20px; }
.tp-mt { margin-top: 20px; }

.tp-quick-input { width: 100%; padding: 12px 14px; border-radius: var(--radius-md); font-size: 18px; font-weight: 600; text-align: center; border: none; font-variant-numeric: tabular-nums; background: var(--bg-input); color: var(--text-primary); outline: none; box-sizing: border-box; }

.tp-divider { display: flex; align-items: center; margin-bottom: 20px; }
.divider-line { flex: 1; height: 1px; background: var(--border-color-light); }
.divider-text { margin: 0 12px; font-size: 12px; color: var(--text-secondary); }

.tp-picker { display: flex; align-items: center; justify-content: center; gap: 12px; }
.tp-col { display: flex; flex-direction: column; align-items: center; gap: 8px; }

.tp-arrow-btn { width: 60px; height: 36px; border-radius: 8px; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; background: var(--bg-input); color: var(--text-primary); transition: opacity 0.15s; }
.tp-arrow-btn:hover { opacity: 0.7; }
.tp-num-input { width: 60px; padding: 8px; border-radius: 8px; font-size: 28px; font-weight: 600; text-align: center; border: none; font-variant-numeric: tabular-nums; background: var(--card-bg); color: var(--text-primary); outline: none; text-align: center; }
.tp-sep { font-size: 32px; font-weight: 600; padding-bottom: 4px; color: var(--text-primary); }

.quick-times { display: flex; flex-wrap: wrap; gap: 8px; }
.quick-time-btn { padding: 6px 12px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; background: var(--bg-input); color: var(--text-primary); transition: opacity 0.15s; }
.quick-time-btn:hover { opacity: 0.7; }

.tp-actions { margin-top: 20px; display: flex; gap: 10px; }
.tp-cancel-btn { flex: 1; padding: 12px; border-radius: var(--radius-md); border: none; cursor: pointer; font-size: 14px; background: var(--bg-input); color: var(--text-primary); }
.tp-confirm-btn { flex: 1; padding: 12px; border-radius: var(--radius-md); border: none; cursor: pointer; font-size: 14px; background: var(--accent-color); color: #fff; }
</style>
