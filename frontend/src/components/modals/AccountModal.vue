<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">
          {{ formMode === 'list' ? '账户管理' : (formMode === 'edit' ? '编辑账户' : '新建账户') }}
        </span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 列表视图 -->
      <template v-if="formMode === 'list'">
        <div class="account-list">
          <div v-for="acc in accounts" :key="acc.id" class="account-item">
            <div class="account-info">
              <div class="account-icon">{{ acc.icon }}</div>
              <div class="account-detail">
                <div class="account-name-row">
                  <span class="account-name">{{ acc.name }}</span>
                  <span v-if="acc.is_default" class="badge-default">默认</span>
                </div>
                <div class="account-balance">
                  ¥{{ acc.real_balance.toFixed(2) }}
                  <span v-if="acc.balance !== 0" class="account-init">(初始 {{ acc.balance >= 0 ? '+' : '' }}{{ acc.balance.toFixed(2) }})</span>
                </div>
              </div>
            </div>
            <div class="account-actions">
              <button v-if="!acc.is_default" @click.stop="handleSetDefault(acc.id)" class="btn-sm btn-ghost">设默认</button>
              <button @click.stop="startEdit(acc)" class="btn-sm btn-ghost">编辑</button>
              <button @click.stop="handleDelete(acc.id)" class="btn-sm btn-danger">删除</button>
            </div>
          </div>
        </div>

        <!-- 总资产 -->
        <div class="total-assets">
          <div class="total-label">总资产</div>
          <div class="total-value">¥{{ totalAssets.toFixed(2) }}</div>
        </div>

        <!-- 新建按钮 -->
        <div class="create-row">
          <button @click="startCreate" class="btn-primary">＋ 新建账户</button>
        </div>
      </template>

      <!-- 新建/编辑表单 -->
      <template v-else>
        <!-- 图标选择 -->
        <div class="form-group">
          <label class="form-label">图标</label>
          <div class="icon-grid">
            <button v-for="ic in presetIcons" :key="ic" type="button" @click="accountForm.icon = ic"
              class="icon-btn" :class="{ active: accountForm.icon === ic }">
              {{ ic }}
            </button>
          </div>
        </div>

        <!-- 名称 -->
        <div class="form-group">
          <label class="form-label">账户名称</label>
          <input v-model="accountForm.name" type="text" placeholder="例：微信钱包" class="form-input" />
        </div>

        <!-- 初始余额 -->
        <div class="form-group">
          <label class="form-label">初始余额</label>
          <input v-model.number="accountForm.balance" type="number" step="0.01" placeholder="0.00" class="form-input" />
          <div class="form-hint">设置已有余额，后续收支会自动增减</div>
        </div>

        <!-- 默认账户 -->
        <div class="toggle-row">
          <span class="toggle-label">设为默认账户</span>
          <div class="toggle-switch" :class="{ on: accountForm.is_default }" @click="accountForm.is_default = !accountForm.is_default">
            <div class="toggle-thumb"></div>
          </div>
        </div>

        <!-- 按钮 -->
        <div class="modal-actions">
          <button @click="resetForm" class="cancel-btn">取消</button>
          <button @click="handleSave" class="confirm-btn"
            :style="{ opacity: accountForm.name.trim() ? 1 : 0.5, cursor: accountForm.name.trim() ? 'pointer' : 'not-allowed' }">
            {{ formMode === 'edit' ? '保存修改' : '创建账户' }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { GetAllWithBalance as GetAllAccountsWithBalance, Create as CreateAccount, Update as UpdateAccount, Delete as DeleteAccount, GetTotalAssets } from '../../../wailsjs/go/service/AccountService'
import { createLogger } from '../../utils/logger'
import type { AccountingStoreLike } from '../../types'

const log = createLogger('Account')

interface AccountInfo {
  id: number
  name: string
  icon: string
  balance: number
  real_balance: number
  is_default: boolean
}

const props = defineProps<{
  show: boolean
  isDark: boolean
  accounts: AccountInfo[]
  store: AccountingStoreLike
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const formMode = ref<'list' | 'create' | 'edit'>('list')
const editingId = ref(0)
const totalAssets = ref(0)
const presetIcons = ['💵', '💳', '🏦', '📱', '💰', '🏠', '🚗', '📈', '💎', '🎁', '🎯', '⭐']

const accountForm = ref({
  name: '',
  icon: '💵',
  balance: 0,
  is_default: false,
})

// 同步外部 accounts 到内部
const localAccounts = ref<AccountInfo[]>([])
watch(() => props.accounts, (val) => {
  localAccounts.value = val
}, { immediate: true })

async function loadTotalAssets() {
  try {
    totalAssets.value = await GetTotalAssets()
  } catch (e) {
    log.error('获取总资产失败', e)
  }
}

function resetForm() {
  formMode.value = 'list'
  editingId.value = 0
  accountForm.value = { name: '', icon: '💵', balance: 0, is_default: false }
}

function startCreate() {
  formMode.value = 'create'
  editingId.value = 0
  accountForm.value = { name: '', icon: '💵', balance: 0, is_default: localAccounts.value.length === 0 }
}

function startEdit(acc: AccountInfo) {
  formMode.value = 'edit'
  editingId.value = acc.id
  accountForm.value = { name: acc.name, icon: acc.icon, balance: acc.balance, is_default: acc.is_default }
}

async function handleSave() {
  if (!accountForm.value.name.trim()) {
    emit('toast', '请输入账户名称', 'error')
    return
  }
  try {
    if (formMode.value === 'create') {
      await CreateAccount(accountForm.value.name.trim(), accountForm.value.icon, accountForm.value.balance, accountForm.value.is_default)
    } else {
      await UpdateAccount(editingId.value, accountForm.value.name.trim(), accountForm.value.icon, accountForm.value.balance, accountForm.value.is_default)
    }
    resetForm()
    emit('refresh')
    emit('toast', formMode.value === 'create' ? '账户创建成功' : '账户保存成功', 'success')
  } catch (e) {
    emit('toast', '保存失败: ' + e, 'error')
  }
}

async function handleDelete(id: number) {
  try {
    await DeleteAccount(id)
    emit('refresh')
    emit('toast', '账户已删除', 'success')
  } catch (e) {
    emit('toast', '删除失败: ' + e, 'error')
  }
}

async function handleSetDefault(id: number) {
  try {
    const acc = localAccounts.value.find(a => a.id === id)
    if (acc) {
      await UpdateAccount(id, acc.name, acc.icon, acc.balance, true)
      emit('refresh')
      emit('toast', '已设为默认账户', 'success')
    }
  } catch (e) {
    emit('toast', '设置失败: ' + e, 'error')
  }
}

watch(() => props.show, (v) => { if (v) loadTotalAssets() })
</script>

<script lang="ts">
export default {
  components: {}
}
</script>

<style scoped>
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
.modal {
  width: 100%;
  max-width: 420px;
  border-radius: 20px;
  padding: 20px;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.close-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8e8e93;
  transition: background 0.15s;
}
.close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

/* 账户列表 */
.account-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 55vh;
  overflow-y: auto;
}
.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 12px;
  border-radius: 12px;
  background: var(--bg-input);
}
.account-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}
.account-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  background: rgba(99, 102, 241, 0.1);
}
.account-detail {
  min-width: 0;
}
.account-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.account-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}
.badge-default {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--accent-color);
  color: #fff;
}
.account-balance {
  font-size: 13px;
  margin-top: 2px;
  color: var(--text-secondary);
}
.account-init {
  font-size: 11px;
  opacity: 0.6;
}
.account-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.btn-sm {
  padding: 4px 8px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 11px;
}
.btn-ghost {
  background: var(--bg-input);
  color: var(--text-secondary);
}
.btn-danger {
  background: rgba(255, 69, 58, 0.1);
  color: #ff453a;
}

/* 总资产 */
.total-assets {
  margin-top: 12px;
  padding: 12px;
  border-radius: 10px;
  text-align: center;
  background: v-bind('isDark ? "rgba(99,102,241,0.1)" : "rgba(99,102,241,0.06)"');
}
.total-label {
  font-size: 12px;
  margin-bottom: 4px;
  color: var(--accent-color);
}
.total-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}
.create-row {
  margin-top: 12px;
}
.btn-primary {
  width: 100%;
  padding: 12px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  background: var(--accent-color);
  color: #fff;
}

/* 表单 */
.form-group {
  margin-bottom: 16px;
}
.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 6px;
  color: var(--text-secondary);
}
.form-input {
  width: 100%;
  padding: 10px 12px;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  background: var(--bg-input);
  color: var(--text-primary);
}
.form-hint {
  font-size: 11px;
  margin-top: 4px;
  color: var(--text-secondary);
}

/* Toggle switch */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
}
.toggle-label {
  font-size: 14px;
  color: var(--text-primary);
}
.toggle-switch {
  width: 48px;
  height: 28px;
  border-radius: 14px;
  padding: 2px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  background: var(--switch-track);
}
.toggle-switch.on {
  background: var(--accent-color);
}
.toggle-thumb {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #fff;
  transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}
.toggle-switch.on .toggle-thumb {
  transform: translateX(20px);
}

/* 按钮 */
.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}
.cancel-btn {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  background: var(--bg-input);
  color: var(--text-secondary);
}
.confirm-btn {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  background: var(--accent-color);
  color: #fff;
}

/* 图标选择网格 */
.icon-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.icon-btn {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  font-size: 20px;
  transition: all 0.15s;
  background: var(--bg-input);
}
.icon-btn.active {
  background: var(--accent-color);
  box-shadow: 0 0 0 2px var(--accent-color);
}
</style>
