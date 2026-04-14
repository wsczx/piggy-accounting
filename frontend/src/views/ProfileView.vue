<template>
  <div class="page-container">
    <div class="page-content">
      <!-- 用户信息卡片（品牌粉紫渐变，与 Logo 一致） -->
      <div class="profile-card" :style="isDark
        ? { background: 'linear-gradient(135deg, #2d1f3d, #1a1025)', border: '1px solid rgba(255,255,255,0.05)' }
        : { background: 'linear-gradient(135deg, #f472b6, #a855f7)', boxShadow: '0 8px 32px rgba(168,85,247,0.3)' }">
        <div class="pc-top">
          <div class="pc-logo">🐷</div>
          <div class="pc-title">猪猪记账</div>
          <div class="pc-sub">管好每一笔，生活更轻松</div>
        </div>
        <div class="pc-asset">
          <span class="pc-asset-label">总资产</span>
          <span class="pc-asset-val">¥{{ formatAmount(totalAssets) }}</span>
        </div>
      </div>

      <!-- 功能入口 -->
      <div class="section section-features">
        <div class="section-title">功能</div>
        <div class="feature-grid">
          <button v-for="item in features" :key="item.label" @click="item.action" class="feature-item">
            <div class="feature-icon" :style="{ background: item.color + '15' }">{{ item.icon }}</div>
            <span class="feature-label">{{ item.label }}</span>
          </button>
        </div>
      </div>

      <!-- 版本信息 -->
      <div class="version">猪猪记账 v1.0.0</div>
    </div>

    <!-- ===== 所有弹窗组件 ===== -->
    <DataManageModal v-model:show="showDataManageModal" :isDark="isDark" :store="store" @toast="handleToast"
      @refresh="handleRefresh" @fullRefresh="handleFullRefresh" />
    <CategoryModal v-model:show="showCategoryModal" :isDark="isDark" :expense-categories="store.expenseCategories"
      :income-categories="store.incomeCategories" :store="store" @toast="handleToast" @refresh="handleRefresh" />

    <ReminderModal v-model:show="showReminderModal" :isDark="isDark" :budget-alerts="budgetAlerts" :store="store"
      @toast="handleToast" @refresh="handleRefresh" />
    <BudgetModal v-model:show="showBudgetModal" :isDark="isDark" :monthly-budget="monthlyBudget"
      :yearly-budget="yearlyBudget" :store="store" @toast="handleToast" @refresh="handleRefresh" />
    <RecurringModal v-model:show="showRecurringModal" :isDark="isDark" :accounts="accounts" :store="store"
      @toast="handleToast" @refresh="handleRefresh" />
    <AccountModal v-model:show="showAccountModal" :isDark="isDark" :accounts="accounts" :store="store"
      @toast="handleToast" @refresh="handleRefresh" />
    <TransferModal v-model:show="showTransferModal" :isDark="isDark" :accounts="accounts" @toast="handleToast"
      @refresh="handleRefresh" />

    <!-- 简单弹窗（无后端依赖或极简） -->
    <SecuritySettingsModal v-model:show="showSecuritySettings" :isDark="isDark" />
    <HelpFeedbackModal v-model:show="showHelpFeedback" :isDark="isDark" @toast="handleToast" />
    <AboutModal v-model:show="showAboutModal" :isDark="isDark" />
    <LedgerModal v-model:show="showLedgerModal" :isDark="isDark" :store="store" @toast="handleToast"
      @refresh="handleFullRefresh" />
    <TagModal v-model:show="showTagModal" :isDark="isDark" :store="store" @toast="handleToast"
      @refresh="handleRefresh" />

    <!-- 通用确认对话框 & 全局 Toast -->
    <div v-if="confirmDialog.show" class="modal-overlay modal-overlay-high">
      <div class="modal modal-confirm">
        <div class="confirm-icon">{{ confirmDialog.icon || '⚠️' }}</div>
        <div class="confirm-title">{{ confirmDialog.title }}</div>
        <div v-if="confirmDialog.message" class="confirm-message">{{ confirmDialog.message }}</div>
        <div class="confirm-actions-row">
          <button @click="confirmDialog.onCancel()" class="cancel-btn">取消</button>
          <button @click="confirmDialog.onConfirm()" class="confirm-btn"
            :class="{ 'btn-danger': confirmDialog.dangerColor }">{{ confirmDialog.confirmText || '确定' }}</button>
        </div>
      </div>
    </div>

    <!-- 全局 Toast -->
    <Transition name="toast-fade">
      <div v-if="toast" class="global-toast" :class="toast.type === 'success' ? 'toast-success' : 'toast-error'">
        {{ toast.msg }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore } from '../stores/accounting'
import { formatAmount } from '../utils/formatters'
import type { BudgetAlert } from '../types'

/** AccountInfo — 与各 Modal 组件的接口对齐 */
interface AccountInfo {
  id: number
  name: string
  icon: string
  balance: number
  real_balance: number
  is_default: boolean
}

// Modal 组件
import DataManageModal from '../components/modals/DataManageModal.vue'
import CategoryModal from '../components/modals/CategoryModal.vue'

import ReminderModal from '../components/modals/ReminderModal.vue'
import BudgetModal from '../components/modals/BudgetModal.vue'
import RecurringModal from '../components/modals/RecurringModal.vue'
import AccountModal from '../components/modals/AccountModal.vue'
import TransferModal from '../components/modals/TransferModal.vue'
import SecuritySettingsModal from '../components/modals/SecuritySettingsModal.vue'
import HelpFeedbackModal from '../components/modals/HelpFeedbackModal.vue'
import AboutModal from '../components/modals/AboutModal.vue'
import LedgerModal from '../components/modals/LedgerModal.vue'
import TagModal from '../components/modals/TagModal.vue'

// Wails API（仅 ProfileView 自身需要的）
import { GetTotalAssets } from '../../wailsjs/go/service/AccountService'
import { CheckBudgetAlerts, GetAllReminders, GetReminderSettings } from '../../wailsjs/go/service/ReminderService'
import { createLogger } from '../utils/logger'

const log = createLogger('Profile')

const router = useRouter()
const route = useRoute()
const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

// ========== 核心状态 ==========
const totalAssets = ref(0)

async function loadTotalAssets() {
  try { totalAssets.value = await GetTotalAssets() } catch (e) { log.error('获取总资产失败', e) }
}

onMounted(() => {
  loadTotalAssets()
  handleOpenParam(route.query.open as string)
})

watch(() => route.query.open, (val) => {
  if (val) handleOpenParam(val as string)
}, { immediate: false })

// ========== 弹窗开关 ==========
const showDataManageModal = ref(false)
const showCategoryModal = ref(false)

const showReminderModal = ref(false)
const showBudgetModal = ref(false)
const showRecurringModal = ref(false)
const showAccountModal = ref(false)
const showTransferModal = ref(false)
const showSecuritySettings = ref(false)
const showHelpFeedback = ref(false)
const showAboutModal = ref(false)
const showLedgerModal = ref(false)
const showTagModal = ref(false)

// 账户数据（供 AccountModal/TransferModal/RecurringModal 共享）
const accounts = ref<AccountInfo[]>([])

async function loadAccounts() {
  try {
    const accountService = await import('../../wailsjs/go/service/AccountService')
    accounts.value = await accountService.GetAllWithBalance()
  } catch (e) { log.error('加载账户失败', e) }
}

// 预算预警数据
const budgetAlerts = ref<BudgetAlert[]>([])

async function checkAlerts() {
  try { budgetAlerts.value = (await CheckBudgetAlerts()) || [] } catch (e) { log.error('检查预算预警失败', e) }
}

/** 根据 ?open=xxx 参数自动打开对应弹窗 */
function handleOpenParam(open: string | undefined) {
  if (!open) return
  setTimeout(() => {
    switch (open) {
      case 'account': loadAccounts(); showAccountModal.value = true; break
      case 'transfer': loadAccounts(); showTransferModal.value = true; break
      case 'recurring': loadAccounts(); showRecurringModal.value = true; break
      case 'import': showDataManageModal.value = true; break
      case 'export': showDataManageModal.value = true; break
      case 'backup': showDataManageModal.value = true; break
      case 'category': showCategoryModal.value = true; break

      case 'ledger': showLedgerModal.value = true; break
      case 'reminder':
        checkAlerts(); showReminderModal.value = true; break
    }
  }, 300)
}

// ========== 功能入口配置 ==========
const monthlyBudget = computed(() => store.monthlyBudget)
const yearlyBudget = computed(() => store.yearlyBudget)

const features = [
  { icon: '🏦', label: '账户', color: '#6366f1', action: () => { loadAccounts(); showAccountModal.value = true } },
  { icon: '🔄', label: '转账', color: '#f59e0b', action: () => { loadAccounts(); showTransferModal.value = true } },
  { icon: '⏰', label: '周期', color: '#10b981', action: () => { loadAccounts(); showRecurringModal.value = true } },
  { icon: '📂', label: '分类', color: '#14b8a6', action: () => { showCategoryModal.value = true } },
  { icon: '📒', label: '多账本', color: '#84cc16', action: () => { showLedgerModal.value = true } },
  { icon: '💰', label: '预算设置', color: '#ef4444', action: () => showBudgetModal.value = true },
  { icon: '🔔', label: '提醒设置', color: '#f97316', action: () => { checkAlerts(); showReminderModal.value = true } },
  { icon: '🔒', label: '安全设置', color: '#64748b', action: () => showSecuritySettings.value = true },
  { icon: '💾', label: '数据管理', color: '#a855f7', action: () => { showDataManageModal.value = true } },
  { icon: '🏷️', label: '标签管理', color: '#8b5cf6', action: () => { showTagModal.value = true } },
  { icon: '❓', label: '帮助反馈', color: '#0ea5e9', action: () => { showHelpFeedback.value = true } },
  { icon: 'ℹ️', label: '关于', color: '#6b7280', action: () => { showAboutModal.value = true } },
]

// ========== 通用 Toast / Confirm ==========
const toast = ref<{ msg: string; type: 'success' | 'error' } | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

function handleToast(msg: string, type: 'success' | 'error') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { msg, type }
  toastTimer = setTimeout(() => { toast.value = null }, 2500)
}

const confirmDialog = ref({
  show: false, icon: '⚠️', title: '', message: '', confirmText: '确定',
  dangerColor: false, onConfirm: () => { }, onCancel: () => { },
})

function showConfirm(options: { icon?: string; title: string; message?: string; confirmText?: string; dangerColor?: boolean }): Promise<boolean> {
  return new Promise((resolve) => {
    confirmDialog.value = {
      show: true, icon: options.icon || '⚠️', title: options.title,
      message: options.message || '', confirmText: options.confirmText || '确定',
      dangerColor: options.dangerColor ?? false,
      onConfirm: () => { confirmDialog.value.show = false; resolve(true) },
      onCancel: () => { confirmDialog.value.show = false; resolve(false) },
    }
  })
}

// ========== 刷新回调 ==========
function handleRefresh() {
  // 基础刷新：重新加载当月数据
  const month = store.currentMonth
  if (month) {
    store.loadRecords(month + '-01', month + '-31', true).catch(() => { })
  } else {
    store.loadRecords().catch(() => { })
  }
  // 同时刷新预算预警数据（提醒设置中的阈值/开关变化后需要实时更新预警卡片）
  checkAlerts()
}

async function handleFullRefresh() {
  // 完整刷新：切换账本/删除等需要全局刷新的场景
  await store.init()
  await loadAccounts()
  await loadTotalAssets()
}
</script>

<style scoped>
/* 页面布局 */
.profile-card {
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 16px;
  text-align: center;
}

.pc-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 14px;
}

.pc-logo {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
}

.pc-title {
  font-size: 17px;
  font-weight: 700;
  color: #fff;
  margin-top: 8px;
}

.pc-sub {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.65);
  margin-top: 3px;
}

.pc-asset {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 8px;
  margin-top: 6px;
}

.pc-asset-label {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.pc-asset-val {
  font-size: 12px;
  font-weight: 700;
  color: #fff;
}

/* 功能入口 */
.section {
  margin-bottom: 16px;
}

.section-title {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 10px;
  padding-left: 4px;
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.feature-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 14px 0;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
  background: v-bind('isDark ? "#1c1c1e" : "#fff"');
  box-shadow: v-bind('isDark ? "0 1px 3px rgba(0,0,0,0.2)" : "0 1px 4px rgba(0,0,0,0.04)"');
}

.feature-item:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: var(--shadow-hover);
}

.feature-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.feature-label {
  font-size: 11px;
  font-weight: 500;
  color: v-bind('isDark ? "#fff" : "#1f2937"');
}

.version {
  text-align: center;
  font-size: 12px;
  padding: 20px 0;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

/* 弹窗基础样式 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}

.modal-overlay-high {
  z-index: 300;
}

.modal {
  width: 100%;
  max-width: 400px;
  border-radius: 20px;
  padding: 20px;
  background: v-bind('isDark ? "#1c1c1e" : "#fff"');
}

/* 确认弹窗（替代内联 style） */
.modal-confirm {
  max-width: 340px;
  padding: 24px;
  text-align: center;
}

.confirm-icon {
  font-size: 36px;
  margin-bottom: 12px;
}

.confirm-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 8px;
  line-height: 1.5;
  color: v-bind('isDark ? "#f5f5f7" : "#1f2937"');
}

.confirm-message {
  font-size: 13px;
  margin-bottom: 20px;
  line-height: 1.6;
  white-space: pre-line;
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}

.confirm-actions-row {
  display: flex;
  gap: 10px;
}

.cancel-btn,
.confirm-btn {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.cancel-btn {
  background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"');
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}

.confirm-btn {
  background: var(--accent-color);
  color: #fff;
}

.confirm-btn.btn-danger {
  background: #ff453a;
}

/* 全局 Toast（替代内联 style） */
.global-toast {
  position: fixed;
  bottom: 80px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  z-index: 500;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.toast-success {
  background: rgba(52, 199, 89, 0.95);
  color: #fff;
}

.toast-error {
  background: rgba(255, 69, 58, 0.95);
  color: #fff;
}

@keyframes toastSlideUp {
  from {
    opacity: 0;
    transform: translate(-50%, 20px);
  }

  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

.toast-fade-enter-active {
  animation: toastSlideUp 0.3s ease;
}

.toast-fade-leave-active {
  animation: toastSlideUp 0.2s ease reverse;
}
</style>
