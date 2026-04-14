<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ formMode === 'list' ? '多账本' : '新建账本' }}</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 列表视图 -->
      <template v-if="formMode === 'list'">
        <div class="ledger-list">
          <div v-for="ledger in ledgers" :key="ledger.name"
            class="ledger-item" :class="{ 'is-active': ledger.is_active }">
            <div class="ledger-main" @click="handleSwitch(ledger.name)">
              <div class="ledger-icon" :class="{ 'icon-active': ledger.is_active }">
                {{ ledger.is_active ? '📒' : '📔' }}
              </div>
              <div class="ledger-detail">
                <div class="ledger-name-row" :class="{ 'name-active': ledger.is_active }">
                  <span>{{ ledger.name }}</span>
                  <span v-if="ledger.is_active" class="badge-active">使用中</span>
                </div>
                <div class="ledger-meta" :class="{ 'meta-active': ledger.is_active }">
                  {{ ledger.record_count }} 条记录
                  <span class="meta-date">创建于 {{ ledger.created_at }}</span>
                </div>
              </div>
            </div>
            <div class="ledger-actions">
              <template v-if="renaming && renamingName === ledger.name">
                <input v-model="editingName" type="text" class="rename-input"
                  @keydown.enter="handleRename(ledger.name, editingName)" />
                <button @click.stop="handleRename(ledger.name, editingName)" class="btn-sm btn-accent">✓</button>
                <button @click.stop="cancelRename" class="btn-sm btn-ghost">✗</button>
              </template>
              <template v-else>
                <button @click.stop="startRename(ledger.name)" class="btn-sm btn-ghost">重命名</button>
                <button v-if="!ledger.is_active" @click.stop="handleDelete(ledger.name)" class="btn-sm btn-danger">删除</button>
              </template>
            </div>
          </div>
        </div>

        <!-- 提示 -->
        <div class="ledger-tip">
          <span>💡</span>
          <span>点击账本即可切换，切换后页面数据会自动刷新</span>
        </div>

        <!-- 新建按钮 -->
        <div class="create-row">
          <button @click="startCreate" class="btn-primary">＋ 新建账本</button>
        </div>
      </template>

      <!-- 新建表单 -->
      <template v-else>
        <div class="form-group">
          <label>账本名称</label>
          <input v-model="newLedgerName" type="text" placeholder="例：旅行账本、装修专用" @keydown.enter="handleCreate"
            class="form-input" />
          <div class="form-hint">新账本会自动创建独立的收支数据和账户</div>
        </div>

        <div class="modal-actions">
          <button @click="formMode = 'list'" class="cancel-btn">取消</button>
          <button @click="handleCreate" class="confirm-btn" :class="{ disabled: !newLedgerName.trim() }">
            创建账本
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { GetAllLedgers, CreateLedger, SwitchLedger, DeleteLedger, RenameLedger } from '../../../wailsjs/go/service/LedgerService'
import { createLogger } from '../../utils/logger'
import type { AccountingStoreLike } from '../../types'

const log = createLogger('Ledger')

interface LedgerInfo {
  name: string
  is_active: boolean
  record_count: number
  created_at: string
}

const props = defineProps<{
  show: boolean
  isDark: boolean
  store: AccountingStoreLike
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const formMode = ref<'list' | 'create'>('list')
const ledgers = ref<LedgerInfo[]>([])
const newLedgerName = ref('')
const editingName = ref('')
const renaming = ref(false)
const renamingName = ref('')

async function loadLedgers() {
  try {
    ledgers.value = await GetAllLedgers()
  } catch (e) {
    log.error('加载账本列表失败', e)
  }
}

function startCreate() {
  formMode.value = 'create'
  newLedgerName.value = ''
}

async function handleCreate() {
  if (!newLedgerName.value.trim()) {
    emit('toast', '请输入账本名称', 'error')
    return
  }
  try {
    await CreateLedger(newLedgerName.value.trim())
    newLedgerName.value = ''
    formMode.value = 'list'
    await loadLedgers()
    emit('refresh')
    emit('toast', '账本创建成功', 'success')
  } catch (e) {
    emit('toast', '创建失败: ' + e, 'error')
  }
}

async function handleSwitch(name: string) {
  if (name === ledgers.value.find(l => l.is_active)?.name) return
  try {
    await SwitchLedger(name)
    await props.store.init()
    await loadLedgers()
    emit('refresh')
    emit(`toast`, `已切换到「${name}」`, 'success')
  } catch (e) {
    emit('toast', '切换失败: ' + e, 'error')
  }
}

async function handleDelete(name: string) {
  try {
    await DeleteLedger(name)
    await loadLedgers()
    emit('refresh')
    emit('toast', '账本已删除', 'success')
  } catch (e) {
    emit('toast', '删除失败: ' + e, 'error')
  }
}

function startRename(name: string) {
  editingName.value = name
  renaming.value = true
  renamingName.value = name
}

function cancelRename() {
  renaming.value = false
  renamingName.value = ''
  editingName.value = ''
}

async function handleRename(oldName: string, newName: string) {
  if (!newName.trim()) {
    emit('toast', '请输入新名称', 'error')
    return
  }
  if (newName.trim() === oldName) {
    renaming.value = false; renamingName.value = ''; editingName.value = ''; return
  }
  try {
    await RenameLedger(oldName, newName.trim())
    renaming.value = false; renamingName.value = ''; editingName.value = ''
    await loadLedgers()
    emit('refresh')
    emit('toast', '重命名成功', 'success')
  } catch (e) {
    emit('toast', '重命名失败: ' + e, 'error')
  }
}

watch(() => props.show, (v) => { if (v) loadLedgers() })
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
  color: var(--text-secondary);
  transition: background 0.15s;
}
.close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

/* 账本列表 */
.ledger-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 55vh;
  overflow-y: auto;
}
.ledger-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.15s;
  background: var(--bg-input);
}
.ledger-item.is-active {
  background: var(--accent-color);
  opacity: 1;
}
.ledger-item:not(.is-active) {
  opacity: 0.85;
}
.ledger-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}
.ledger-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  background: rgba(132, 204, 22, 0.1);
}
.ledger-icon.icon-active {
  background: rgba(255, 255, 255, 0.2);
}
.ledger-detail {
  min-width: 0;
}
.ledger-name-row {
  font-size: 14px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary);
}
.ledger-name-row.name-active {
  color: #fff;
}
.badge-active {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(255,255,255,0.25);
  color: #fff;
  flex-shrink: 0;
}
.ledger-meta {
  font-size: 13px;
  margin-top: 2px;
  color: var(--text-secondary);
}
.ledger-meta.meta-active {
  color: rgba(255, 255, 255, 0.7);
}
.meta-date {
  font-size: 11px;
  opacity: 0.6;
  margin-left: 8px;
}
.ledger-actions {
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
  white-space: nowrap;
}
.btn-ghost {
  background: var(--bg-input);
  color: var(--text-secondary);
}
.btn-accent {
  background: var(--accent-color);
  color: #fff;
}
.btn-danger {
  background: rgba(255,69,58,0.1);
  color: #ff453a;
}
.rename-input {
  width: 100px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  border: none;
  outline: none;
  background: var(--card-bg);
  color: var(--text-primary);
}

/* 提示 */
.ledger-tip {
  margin-top: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  font-size: 12px;
  display: flex;
  gap: 6px;
  align-items: center;
  background: v-bind('isDark ? "rgba(132,204,22,0.1)" : "rgba(132,204,22,0.06)"');
  color: v-bind('isDark ? "#84cc16" : "#65a30d"');
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
.form-group label {
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
.confirm-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
