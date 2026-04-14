<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal backup-modal">
      <div class="modal-header">
        <span class="modal-title">数据备份</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 操作按钮 -->
      <div class="backup-actions">
        <button @click="handleCreateBackup" :disabled="backupLoading"
          class="btn-backup-primary" :style="{ opacity: backupLoading ? 0.6 : 1 }">
          <span>💾</span>
          {{ backupLoading ? '备份中...' : '立即备份' }}
        </button>
      </div>

      <!-- 备份列表 -->
      <div class="backup-count">共 {{ backups.length }} 个备份</div>
      <div class="backup-list">
        <div v-if="backups.length === 0" class="backup-empty">
          <div class="backup-empty-icon">☁️</div>
          <p class="backup-empty-text">暂无备份</p>
          <p class="backup-empty-hint">点击上方按钮创建第一个备份</p>
        </div>
        <div v-for="b in backups" :key="b.filename" class="backup-item">
          <div class="backup-info">
            <div class="backup-file-icon">{{ b.auto ? '🔄' : '💾' }}</div>
            <div class="backup-detail">
              <div class="backup-name-row">
                <span>{{ b.display_name || b.filename }}</span>
                <span v-if="b.auto" class="badge-auto">自动</span>
              </div>
              <div class="backup-meta">
                <span>{{ b.created_at }}</span>
                <span>{{ formatFileSize(b.file_size) }}</span>
              </div>
            </div>
          </div>
          <div class="backup-btns">
            <button v-if="backupOpLoading === b.filename" disabled class="btn-sm btn-loading">处理中…</button>
            <template v-else>
              <button @click.stop="startExportBackup(b.filename)" class="btn-sm btn-ghost">导出</button>
              <button @click.stop="handleRestoreBackup(b.filename)" class="btn-sm btn-accent">恢复</button>
              <button @click.stop="handleDeleteBackup(b.filename)" class="btn-sm btn-danger">删除</button>
            </template>
          </div>
        </div>
      </div>

      <!-- Toast -->
      <div v-if="backupToast" class="backup-toast" :class="'toast-' + backupToast.type">
        {{ backupToast.msg }}
      </div>

      <!-- 内联确认对话框 -->
      <div v-if="confirm.show" class="inline-confirm-overlay" @click.self="confirm.show = false">
        <div class="inline-confirm-box">
          <div class="confirm-icon-lg">⚠️</div>
          <div class="confirm-t">{{ confirm.title }}</div>
          <div v-if="confirm.message" class="confirm-m">{{ confirm.message }}</div>
          <div class="confirm-btns">
            <button @click="confirm.show = false" class="ic-cancel">取消</button>
            <button @click="confirm.onConfirm" :class="confirm.dangerColor ? 'ic-danger' : 'ic-ok'">{{ confirm.confirmText || '确定' }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { CreateBackup, RestoreBackup, DeleteBackup, ListBackups, ExportBackupToDir } from '../../../wailsjs/go/service/BackupService'
import { SelectDirectory } from '../../../wailsjs/go/main/App'

const props = defineProps<{
  show: boolean; isDark: boolean
}>()
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

interface BackupInfo {
  filename: string
  display_name: string
  file_size: number
  created_at: string
  auto: boolean
  ledger_name: string
}

const backups = ref<BackupInfo[]>([])
const backupLoading = ref(false)
const backupOpLoading = ref('')
const backupToast = ref<{ msg: string; type: 'success' | 'error' } | null>(null)
let backupToastTimer: ReturnType<typeof setTimeout> | null = null

const confirm = ref<{
  show: boolean; title: string; message?: string; confirmText?: string; dangerColor?: boolean
  onConfirm: () => void
}>({ show: false, title: '', onConfirm: () => {} })

function showConfirm(options: {
  title: string; message?: string; confirmText?: string; dangerColor?: boolean
}): Promise<boolean> {
  return new Promise((resolve) => {
    confirm.value = {
      show: true,
      title: options.title,
      message: options.message || '',
      confirmText: options.confirmText || '确定',
      dangerColor: options.dangerColor ?? false,
      onConfirm: () => { confirm.value.show = false; resolve(true) }
    }
  })
}

function showBackupToast(msg: string, type: 'success' | 'error') {
  if (backupToastTimer) clearTimeout(backupToastTimer)
  backupToast.value = { msg, type }
  backupToastTimer = setTimeout(() => { backupToast.value = null }, 2500)
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function loadBackups() {
  try {
    const result = await ListBackups()
    backups.value = result || []
  } catch (e) {
    showBackupToast('加载备份失败: ' + e, 'error')
    backups.value = []
  }
}

async function handleCreateBackup() {
  try {
    backupLoading.value = true
    await CreateBackup(false)
    await loadBackups()
    showBackupToast('备份成功', 'success')
  } catch (e) {
    showBackupToast('备份失败: ' + e, 'error')
  } finally {
    backupLoading.value = false
  }
}

async function handleRestoreBackup(filename: string) {
  if (backupOpLoading.value) return
  if (!await showConfirm({
    title: '确定要从此备份恢复？',
    message: `备份文件：${filename}\n\n恢复后当前数据将被覆盖（恢复前会自动创建备份）。`,
    confirmText: '恢复',
    dangerColor: true,
  })) return
  backupOpLoading.value = filename
  try {
    await RestoreBackup(filename)
    showBackupToast('恢复成功！数据已重新加载', 'success')
    emit('refresh')
    await loadBackups()
  } catch (e) {
    showBackupToast('恢复失败: ' + e, 'error')
  } finally {
    backupOpLoading.value = ''
  }
}

async function handleDeleteBackup(filename: string) {
  if (backupOpLoading.value) return
  if (!await showConfirm({
    title: `确定删除备份文件？`,
    message: filename,
    confirmText: '删除',
    dangerColor: true,
  })) return
  backupOpLoading.value = filename
  try {
    await DeleteBackup(filename)
    showBackupToast('删除成功', 'success')
    await loadBackups()
  } catch (e) {
    showBackupToast('删除失败: ' + e, 'error')
  } finally {
    backupOpLoading.value = ''
  }
}

async function startExportBackup(filename: string) {
  if (backupOpLoading.value) return
  try {
    backupOpLoading.value = filename
    const dir = await SelectDirectory('选择导出位置')
    if (!dir || dir === '') {
      showBackupToast('已取消导出', 'error')
      return
    }
    const destPath = await ExportBackupToDir(filename, dir)
    showBackupToast(`已导出到：${destPath}`, 'success')
  } catch (e) {
    showBackupToast('导出失败: ' + e, 'error')
  } finally {
    backupOpLoading.value = ''
  }
}

watch(() => props.show, (val) => {
  if (val) loadBackups()
})
</script>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px }
.backup-modal { width: 100%; max-width: 420px; max-height: 85vh; overflow-y: auto; border-radius: 20px; padding: 20px; background: var(--card-bg); position: relative; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }
.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: var(--text-secondary); }

.backup-actions { display: flex; gap: 8px; margin-bottom: 16px; }
.btn-backup-primary { flex: 1; padding: 12px; border-radius: 12px; border: none; cursor: pointer; font-size: 14px; font-weight: 600; background: var(--accent-color); color: #fff; display: flex; align-items: center; justify-content: center; gap: 6px; }
.backup-count { font-size: 12px; margin-bottom: 8px; color: var(--text-secondary); }
.backup-list { display: flex; flex-direction: column; gap: 8px; max-height: 50vh; overflow-y: auto; }

.backup-empty { text-align: center; padding: 32px 0; }
.backup-empty-icon { font-size: 36px; margin-bottom: 8px; }
.backup-empty-text { font-size: 13px; color: var(--text-secondary); }
.backup-empty-hint { font-size: 12px; color: var(--text-secondary); margin-top: 4px; }

.backup-item { display: flex; align-items: center; justify-content: space-between; padding: 12px; border-radius: 12px; background: var(--bg-input); }
.backup-info { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }
.backup-file-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 16px; flex-shrink: 0; background: rgba(6,182,212,0.1); }
.backup-detail { min-width: 0; }
.backup-name-row { font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 6px; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.badge-auto { font-size: 10px; padding: 1px 6px; border-radius: 4px; background: rgba(132,204,22,0.15); color: #84cc16; flex-shrink: 0; }
.backup-meta { font-size: 11px; display: flex; gap: 8px; margin-top: 2px; color: var(--text-secondary); }
.backup-btns { display: flex; align-items: center; gap: 6px; flex-shrink: 0; padding-left: 8px; }

.btn-sm { padding: 6px 10px; border-radius: 8px; border: none; cursor: pointer; font-size: 12px; font-weight: 500; }
.btn-ghost { background: var(--bg-input); color: var(--text-secondary); }
.btn-accent { background: rgba(99,102,241,0.1); color: #6366f1; }
.btn-danger { background: rgba(255,69,58,0.1); color: #ff453a; }
.btn-loading { color: #6366f1; cursor: wait; }

/* Toast */
.backup-toast { position: fixed; bottom: 40px; left: 50%; transform: translateX(-50%); padding: 10px 20px; border-radius: 10px; font-size: 13px; font-weight: 500; white-space: nowrap; z-index: 200; box-shadow: 0 4px 16px rgba(0,0,0,0.15); color: #fff; }
.toast-success { background: rgba(52,199,89,0.95); }
.toast-error { background: rgba(255,69,58,0.95); }

/* Inline Confirm */
.inline-confirm-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 200; }
.inline-confirm-box { width: 300px; padding: 24px; border-radius: 16px; text-align: center; background: var(--card-bg); }
.confirm-icon-lg { font-size: 28px; margin-bottom: 8px; }
.confirm-t { font-size: 15px; font-weight: 600; margin-bottom: 6px; color: var(--text-primary); }
.confirm-m { font-size: 13px; margin-bottom: 16px; line-height: 1.5; word-break: break-all; color: var(--text-secondary); }
.confirm-btns { display: flex; gap: 8px; }
.ic-cancel { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; background: var(--bg-input); color: var(--text-secondary); }
.ic-ok { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 600; color: #fff; background: var(--accent-color); }
.ic-danger { flex: 1; padding: 8px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 600; color: #fff; background: #ff453a; }
</style>
