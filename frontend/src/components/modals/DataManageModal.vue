<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">数据管理</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 分段标签 -->
      <div class="dm-seg" :class="'dm-seg-' + activeTab">
        <button v-for="t in tabs" :key="t.key" class="dm-seg-btn" :class="{ active: activeTab === t.key }"
          @click="activeTab = t.key">{{ t.label }}</button>
      </div>

      <!-- Tab 1: 备份恢复 -->
      <template v-if="activeTab === 'backup'">
        <div class="dm-actions">
          <button @click="handleCreateBackup" :disabled="backupLoading"
            class="dm-btn-primary" :style="{ opacity: backupLoading ? 0.6 : 1 }">
            <span>💾</span> {{ backupLoading ? '备份中...' : '立即备份' }}
          </button>
        </div>
        <div class="dm-count">共 {{ backups.length }} 个备份</div>
        <div class="dm-list">
          <div v-if="backups.length === 0" class="dm-empty">
            <div class="dm-empty-icon">☁️</div>
            <p>暂无备份</p>
            <p class="dm-empty-hint">点击上方按钮创建第一个备份</p>
          </div>
          <div v-for="b in backups" :key="b.filename" class="dm-item">
            <div class="dm-item-info">
              <div class="dm-item-icon">{{ b.auto ? '🔄' : '💾' }}</div>
              <div class="dm-item-detail">
                <div class="dm-item-name">
                  <span>{{ b.display_name || b.filename }}</span>
                  <span v-if="b.auto" class="dm-badge-auto">自动</span>
                </div>
                <div class="dm-item-meta">
                  <span>{{ b.created_at }}</span>
                  <span>{{ formatFileSize(b.file_size) }}</span>
                </div>
              </div>
            </div>
            <div class="dm-item-btns">
              <button v-if="backupOpLoading === b.filename" disabled class="dm-btn-sm dm-btn-loading">处理中…</button>
              <template v-else>
                <button @click.stop="startExportBackup(b.filename)" class="dm-btn-sm dm-btn-ghost">导出</button>
                <button @click.stop="handleRestoreBackup(b.filename)" class="dm-btn-sm dm-btn-accent">恢复</button>
                <button @click.stop="handleDeleteBackup(b.filename)" class="dm-btn-sm dm-btn-danger">删除</button>
              </template>
            </div>
          </div>
        </div>
      </template>

      <!-- Tab 2: 数据导入 -->
      <template v-if="activeTab === 'import'">
        <div class="dm-form-group">
          <label>导入来源</label>
          <div class="dm-type-selector">
            <button v-for="t in importTypes" :key="t.key" @click="importType = t.key as 'csv' | 'wechat' | 'alipay'"
              class="dm-type-btn" :class="{ active: importType === t.key }">{{ t.label }}</button>
          </div>
        </div>
        <div class="dm-hint dm-hint-import">
          <p v-if="importType === 'csv'">支持标准 CSV 格式：日期,类型,类别,金额,备注</p>
          <p v-else-if="importType === 'wechat'">支持微信账单 CSV 导出文件</p>
          <p v-else>支持支付宝账单 CSV 导出文件</p>
        </div>
        <div v-if="importResult" class="dm-result">
          <div class="dm-result-item success"><span class="dm-result-icon">✓</span><span>成功导入 {{ importResult.successCount }} 条</span></div>
          <div v-if="importResult.skipCount > 0" class="dm-result-item skip"><span class="dm-result-icon">⊘</span><span>跳过重复 {{ importResult.skipCount }} 条</span></div>
          <div v-if="importResult.errorCount > 0" class="dm-result-item error"><span class="dm-result-icon">✗</span><span>失败 {{ importResult.errorCount }} 条</span></div>
        </div>
        <div class="dm-actions">
          <label class="dm-btn-primary" :class="{ disabled: importLoading }" style="cursor:pointer">
            <input type="file" accept=".csv" class="dm-file-input" @change="handleFileSelect" :disabled="importLoading" />
            {{ importLoading ? '导入中...' : '选择文件导入' }}
          </label>
        </div>
      </template>

      <!-- Tab 3: 数据导出 -->
      <template v-if="activeTab === 'export'">
        <div class="dm-form-group">
          <label>开始日期</label>
          <input v-model="exportStart" type="date" class="dm-input" />
        </div>
        <div class="dm-form-group">
          <label>结束日期</label>
          <input v-model="exportEnd" type="date" class="dm-input" />
        </div>
        <div class="dm-actions">
          <button @click="handleExportCSV" class="dm-btn-primary">导出 CSV</button>
        </div>
      </template>

      <!-- Tab 4: 清空数据 -->
      <template v-if="activeTab === 'clear'">
        <div class="dm-clear-section">
          <div class="dm-clear-icon">🗑️</div>
          <div class="dm-clear-title">清空当前账本数据</div>
          <div class="dm-clear-desc">删除当前账本的所有记账记录、预算、转账、周期记账、标签和任务。</div>
          <div class="dm-clear-keep">
            <span class="dm-clear-keep-icon">✅</span>
            <span>保留系统类别和默认账户</span>
          </div>
        </div>
        <div v-if="clearResult" class="dm-clear-result" :class="clearResult.success ? 'dm-result-success' : 'dm-result-error'">
          <span class="dm-clear-result-icon">{{ clearResult.success ? '✅' : '❌' }}</span>
          <span>{{ clearResult.message }}</span>
        </div>
        <div class="dm-actions">
          <button @click="handleClearData" :disabled="clearLoading"
            class="dm-btn-danger-full dm-btn-clear" :style="{ opacity: clearLoading ? 0.6 : 1 }">
            {{ clearLoading ? '清空中...' : '清空数据' }}
          </button>
        </div>
      </template>

      <!-- Toast -->
      <div v-if="toast" class="dm-toast" :class="'toast-' + toast.type">{{ toast.msg }}</div>

      <!-- 内联确认对话框 -->
      <div v-if="confirm.show" class="dm-confirm-overlay" @click.self="confirm.show = false">
        <div class="dm-confirm-box">
          <div class="dm-confirm-icon">⚠️</div>
          <div class="dm-confirm-t">{{ confirm.title }}</div>
          <div v-if="confirm.message" class="dm-confirm-m">{{ confirm.message }}</div>
          <div class="dm-confirm-btns">
            <button @click="confirm.show = false" class="dm-btn-cancel">取消</button>
            <button @click="confirm.onConfirm" :class="confirm.dangerColor ? 'dm-btn-danger-full' : 'dm-btn-ok-full'">{{ confirm.confirmText || '确定' }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  CreateBackup, RestoreBackup, DeleteBackup, ListBackups, ExportBackupToDir, ClearLedgerData
} from '../../../wailsjs/go/service/BackupService'
import { SelectDirectory, SaveExportToDir } from '../../../wailsjs/go/main/App'
import {
  ExportToCSV as ExportRecordsToCSV,
  ImportFromCSV as ImportRecordsFromCSV, ImportFromWeChat, ImportFromAlipay
} from '../../../wailsjs/go/service/ExportImportService'

const props = defineProps<{
  show: boolean
  isDark: boolean
  store?: { loadRecords(): Promise<void> }
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
  (e: 'fullRefresh'): void
}>()

// ====== 分段标签 ======
type TabKey = 'backup' | 'import' | 'export' | 'clear'
const activeTab = ref<TabKey>('backup')
const tabs = [
  { key: 'backup' as TabKey, label: '备份恢复' },
  { key: 'import' as TabKey, label: '数据导入' },
  { key: 'export' as TabKey, label: '数据导出' },
  { key: 'clear' as TabKey, label: '清空数据' },
]

// ====== Toast & Confirm ======
const toast = ref<{ msg: string; type: 'success' | 'error' } | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string, type: 'success' | 'error') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { msg, type }
  toastTimer = setTimeout(() => { toast.value = null }, 2500)
}

const confirm = ref<{
  show: boolean; title: string; message?: string; confirmText?: string; dangerColor?: boolean
  onConfirm: () => void
}>({ show: false, title: '', onConfirm: () => {} })

function showConfirm(options: {
  title: string; message?: string; confirmText?: string; dangerColor?: boolean
}): Promise<boolean> {
  return new Promise((resolve) => {
    confirm.value = {
      show: true, title: options.title, message: options.message || '',
      confirmText: options.confirmText || '确定', dangerColor: options.dangerColor ?? false,
      onConfirm: () => { confirm.value.show = false; resolve(true) }
    }
  })
}

// ====== 备份 ======
interface BackupInfo { filename: string; display_name: string; file_size: number; created_at: string; auto: boolean; ledger_name: string }
const backups = ref<BackupInfo[]>([])
const backupLoading = ref(false)
const backupOpLoading = ref('')

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function loadBackups() {
  try { backups.value = (await ListBackups()) || [] }
  catch (e) { showToast('加载备份失败: ' + e, 'error'); backups.value = [] }
}

async function handleCreateBackup() {
  try {
    backupLoading.value = true; await CreateBackup(false); await loadBackups(); showToast('备份成功', 'success')
  } catch (e) { showToast('备份失败: ' + e, 'error') } finally { backupLoading.value = false }
}

async function handleRestoreBackup(filename: string) {
  if (backupOpLoading.value) return
  if (!await showConfirm({ title: '确定要从此备份恢复？', message: `备份文件：${filename}\n\n恢复后当前数据将被覆盖（恢复前会自动创建备份）。`, confirmText: '恢复', dangerColor: true })) return
  backupOpLoading.value = filename
  try { await RestoreBackup(filename); showToast('恢复成功！', 'success'); emit('refresh'); await loadBackups() }
  catch (e) { showToast('恢复失败: ' + e, 'error') } finally { backupOpLoading.value = '' }
}

async function handleDeleteBackup(filename: string) {
  if (backupOpLoading.value) return
  if (!await showConfirm({ title: '确定删除备份文件？', message: filename, confirmText: '删除', dangerColor: true })) return
  backupOpLoading.value = filename
  try { await DeleteBackup(filename); showToast('删除成功', 'success'); await loadBackups() }
  catch (e) { showToast('删除失败: ' + e, 'error') } finally { backupOpLoading.value = '' }
}

async function startExportBackup(filename: string) {
  if (backupOpLoading.value) return
  try {
    backupOpLoading.value = filename
    const dir = await SelectDirectory('选择导出位置')
    if (!dir || dir === '') { showToast('已取消导出', 'error'); return }
    const destPath = await ExportBackupToDir(filename, dir)
    showToast(`已导出到：${destPath}`, 'success')
  } catch (e) { showToast('导出失败: ' + e, 'error') } finally { backupOpLoading.value = '' }
}

// ====== 导入 ======
const importType = ref<'csv' | 'wechat' | 'alipay'>('csv')
const importLoading = ref(false)
const importResult = ref<{ successCount: number; skipCount: number; errorCount: number } | null>(null)

interface ImportTypeOption { key: string; label: string }
const importTypes: ImportTypeOption[] = [
  { key: 'csv', label: '标准CSV' },
  { key: 'wechat', label: '微信账单' },
  { key: 'alipay', label: '支付宝' },
]

async function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    importLoading.value = true; importResult.value = null
    const data = Array.from(new Uint8Array(await file.arrayBuffer()))
    let result
    switch (importType.value) {
      case 'wechat': result = await ImportFromWeChat(data, true); break
      case 'alipay': result = await ImportFromAlipay(data, true); break
      default: result = await ImportRecordsFromCSV(data, true);
    }
    importResult.value = result
    if (props.store) await props.store.loadRecords()
    emit('refresh')
    if (result.errorCount === 0) { setTimeout(() => { emit('update:show', false); importResult.value = null }, 2000) }
  } catch (error) { showToast(`导入失败: ${error}`, 'error') } finally { importLoading.value = false; input.value = '' }
}

// ====== 导出 ======
const exportStart = ref('')
const exportEnd = ref('')

function initExportDates() {
  const now = new Date()
  const y = now.getFullYear(), m = now.getMonth() + 1
  const lastDay = new Date(y, m, 0).getDate()
  exportStart.value = `${y}-${String(m).padStart(2, '0')}-01`
  exportEnd.value = `${y}-${String(m).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`
}

async function handleExportCSV() {
  try {
    const dir = await SelectDirectory('选择导出位置')
    if (!dir || dir === '') { showToast('已取消导出', 'error'); return }
    const data = await ExportRecordsToCSV(exportStart.value, exportEnd.value)
    const defaultFilename = `记账数据_${exportStart.value}_${exportEnd.value}.csv`
    const destPath = await SaveExportToDir(defaultFilename, dir, data)
    showToast(`已导出到：${destPath}`, 'success')
  } catch (error) { showToast('导出失败: ' + error, 'error') }
}

// ====== 清空数据 ======
const clearLoading = ref(false)
const clearResult = ref<{ success: boolean; message: string } | null>(null)

async function handleClearData() {
  if (clearLoading.value) return
  if (!await showConfirm({
    title: '确定清空当前账本的所有数据？',
    message: '此操作将删除当前账本的所有记账记录、预算、转账、周期记账、标签和任务。\n\n建议先创建备份，清空后无法恢复！',
    confirmText: '清空',
    dangerColor: true,
  })) return

  // 二次确认
  if (!await showConfirm({
    title: '再次确认：真的要清空吗？',
    message: '请确保已经备份重要数据，清空后不可撤销！',
    confirmText: '确认清空',
    dangerColor: true,
  })) return

  clearLoading.value = true
  clearResult.value = null
  try {
    const count = await ClearLedgerData() as number
    clearResult.value = { success: true, message: `已清空 ${count} 条记录` }
    showToast('数据已清空', 'success')
    emit('fullRefresh')
  } catch (e) {
    clearResult.value = { success: false, message: '清空失败: ' + e }
  } finally {
    clearLoading.value = false
  }
}

// ====== 生命周期 ======
watch(() => props.show, (val) => {
  if (val) {
    loadBackups()
    initExportDates()
    importResult.value = null
    clearResult.value = null
    activeTab.value = 'backup'
  }
})
</script>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px }
.modal { width: 100%; max-width: 420px; max-height: 85vh; overflow-y: auto; border-radius: 20px; padding: 20px; background: var(--card-bg); position: relative; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }
.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: var(--text-secondary); }

/* 分段标签 — 品牌色系三色 */
.dm-seg { display: flex; background: var(--bg-input); border-radius: 10px; padding: 3px; margin-bottom: 16px; }
.dm-seg-btn { flex: 1; padding: 8px 0; border: none; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; background: transparent; color: var(--text-secondary); transition: all 0.2s; }
/* 备份 tab 激活 — 品牌粉 */
.dm-seg-backup .dm-seg-btn.active { background: #f472b6; color: #fff; }
/* 导入 tab 激活 — 品牌紫 */
.dm-seg-import .dm-seg-btn.active { background: #a855f7; color: #fff; }
/* 导出 tab 激活 — 蓝紫 */
.dm-seg-export .dm-seg-btn.active { background: #818cf8; color: #fff; }
/* 清空 tab 激活 — 红色 */
.dm-seg-clear .dm-seg-btn.active { background: #ff453a; color: #fff; }

/* 操作按钮区 */
.dm-actions { display: flex; gap: 8px; margin-bottom: 16px; }

/* 通用主按钮 */
.dm-btn-primary { flex: 1; padding: 12px; border-radius: 12px; border: none; cursor: pointer; font-size: 14px; font-weight: 600; color: #fff; display: flex; align-items: center; justify-content: center; gap: 6px; background: var(--accent-color); }
.dm-btn-primary:hover { opacity: 0.85; }
.dm-btn-primary.disabled { opacity: 0.6; cursor: not-allowed; }
.dm-count { font-size: 12px; margin-bottom: 8px; color: var(--text-secondary); }
.dm-list { display: flex; flex-direction: column; gap: 8px; max-height: 50vh; overflow-y: auto; }
.dm-empty { text-align: center; padding: 32px 0; }
.dm-empty-icon { font-size: 36px; margin-bottom: 8px; }
.dm-empty p { font-size: 13px; color: var(--text-secondary); }
.dm-empty-hint { font-size: 12px !important; margin-top: 4px; }
.dm-item { display: flex; align-items: center; justify-content: space-between; padding: 12px; border-radius: 12px; background: var(--bg-input); }
.dm-item-info { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }
.dm-item-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 16px; flex-shrink: 0; background: rgba(6,182,212,0.1); }
.dm-item-detail { min-width: 0; }
.dm-item-name { font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dm-badge-auto { font-size: 10px; padding: 1px 6px; border-radius: 4px; background: rgba(132,204,22,0.15); color: #84cc16; flex-shrink: 0; }
.dm-item-meta { font-size: 11px; display: flex; gap: 8px; margin-top: 2px; color: var(--text-secondary); }
.dm-item-btns { display: flex; align-items: center; gap: 6px; flex-shrink: 0; padding-left: 8px; }
.dm-btn-sm { padding: 6px 10px; border-radius: 8px; border: none; cursor: pointer; font-size: 12px; font-weight: 500; }
.dm-btn-ghost { background: var(--bg-input); color: var(--text-secondary); }
.dm-btn-accent { background: rgba(168,85,247,0.1); color: #a855f7; }
.dm-btn-danger { background: rgba(255,69,58,0.1); color: #ff453a; }
.dm-btn-loading { color: #a855f7; cursor: wait; }

/* 表单 */
.dm-form-group { margin-bottom: 14px; }
.dm-form-group label { display: block; font-size: 13px; margin-bottom: 6px; color: var(--text-secondary); }
.dm-input { width: 100%; padding: 12px 14px; border-radius: 12px; border: 1.5px solid transparent; font-size: 14px; outline: none; box-sizing: border-box; background: var(--bg-input); color: var(--text-primary); transition: border-color 0.2s; }
.dm-input:focus { border-color: var(--accent-color); }
.dm-type-selector { display: flex; gap: 8px; }
.dm-type-btn { flex: 1; padding: 10px; border-radius: 10px; border: none; font-size: 13px; cursor: pointer; transition: all 0.15s; background: var(--bg-input); color: var(--text-secondary); }
.dm-type-btn.active { background: var(--accent-color); color: #fff; }
.dm-hint { font-size: 12px; margin-bottom: 14px; padding: 10px; border-radius: 8px; background: rgba(168,85,247,0.06); color: var(--text-secondary); }
.dm-hint-import { background: rgba(168,85,247,0.06); }
.dm-result { margin-bottom: 14px; padding: 12px; border-radius: 12px; background: rgba(0,0,0,0.03); }
.dm-result-item { display: flex; align-items: center; gap: 8px; font-size: 13px; margin-bottom: 6px; }
.dm-result-item:last-child { margin-bottom: 0; }
.dm-result-item.success { color: #30d158; }
.dm-result-item.skip { color: #ff9f0a; }
.dm-result-item.error { color: #ff453a; }
.dm-result-icon { width: 18px; height: 18px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: bold; }
.dm-result-item.success .dm-result-icon { background: rgba(48,209,88,0.15); }
.dm-result-item.skip .dm-result-icon { background: rgba(255,159,10,0.15); }
.dm-result-item.error .dm-result-icon { background: rgba(255,69,58,0.15); }
.dm-file-input { display: none; }

/* Toast */
.dm-toast { position: fixed; bottom: 40px; left: 50%; transform: translateX(-50%); padding: 10px 20px; border-radius: 10px; font-size: 13px; font-weight: 500; white-space: nowrap; z-index: 200; box-shadow: 0 4px 16px rgba(0,0,0,0.15); color: #fff; }
.toast-success { background: rgba(52,199,89,0.95); }
.toast-error { background: rgba(255,69,58,0.95); }

/* Inline Confirm */
.dm-confirm-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 200; }
.dm-confirm-box { width: 300px; padding: 24px; border-radius: 16px; text-align: center; background: var(--card-bg); }
.dm-confirm-icon { font-size: 28px; margin-bottom: 8px; }
.dm-confirm-t { font-size: 15px; font-weight: 600; margin-bottom: 6px; color: var(--text-primary); }
.dm-confirm-m { font-size: 13px; margin-bottom: 16px; line-height: 1.5; word-break: break-all; color: var(--text-secondary); }
.dm-confirm-btns { display: flex; gap: 8px; }
.dm-btn-cancel { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; background: var(--bg-input); color: var(--text-secondary); }
.dm-btn-ok-full { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 600; color: #fff; background: var(--accent-color); }
.dm-btn-danger-full { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 600; color: #fff; background: #ff453a; }

/* 清空数据 Tab */
.dm-clear-section { text-align: center; padding: 20px 0 16px; }
.dm-clear-icon { font-size: 36px; margin-bottom: 10px; }
.dm-clear-title { font-size: 15px; font-weight: 600; margin-bottom: 8px; color: var(--text-primary); }
.dm-clear-desc { font-size: 13px; line-height: 1.6; margin-bottom: 12px; color: var(--text-secondary); }
.dm-clear-keep { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; padding: 6px 12px; border-radius: 8px; background: rgba(52,199,89,0.08); color: #30d158; }
.dm-clear-keep-icon { font-size: 14px; }
.dm-clear-result { display: flex; align-items: center; gap: 8px; padding: 12px; border-radius: 12px; margin-bottom: 14px; font-size: 13px; }
.dm-clear-result.dm-result-success { background: rgba(52,199,89,0.1); color: #30d158; }
.dm-clear-result.dm-result-error { background: rgba(255,69,58,0.1); color: #ff453a; }
.dm-clear-result-icon { font-size: 16px; }
.dm-btn-clear { flex: 1; padding: 12px; border-radius: 12px; font-size: 14px; border: none; cursor: pointer; color: #fff; display: flex; align-items: center; justify-content: center; gap: 6px; background: #ff453a; }
.dm-btn-clear:hover { opacity: 0.85; }
</style>
