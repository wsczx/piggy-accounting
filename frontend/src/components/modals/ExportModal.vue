<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">导出数据</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>
      <div class="form-group">
        <label>开始日期</label>
        <input v-model="startDate" type="date" class="form-input" />
      </div>
      <div class="form-group">
        <label>结束日期</label>
        <input v-model="endDate" type="date" class="form-input" />
      </div>
      <div class="modal-actions">
        <button @click="$emit('update:show', false)" class="cancel-btn">取消</button>
        <button @click="handleExport" class="confirm-btn">导出 CSV</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ExportToCSV as ExportRecordsToCSV } from '../../../wailsjs/go/service/ExportImportService'
import { SaveExportToDir, SelectDirectory } from '../../../wailsjs/go/main/App'

const props = defineProps<{
  show: boolean
  isDark: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
}>()

const startDate = ref('')
const endDate = ref('')

watch(() => props.show, (val) => {
  if (val) initDates()
})

function initDates() {
  const now = new Date()
  const year = now.getFullYear()
  const month = now.getMonth() + 1
  const lastDay = new Date(year, month, 0).getDate()
  startDate.value = `${year}-${String(month).padStart(2,'0')}-01`
  endDate.value = `${year}-${String(month).padStart(2,'0')}-${String(lastDay).padStart(2,'0')}`
}

async function handleExport() {
  try {
    const dir = await SelectDirectory('选择导出位置')
    if (!dir || dir === '') {
      emit('toast', '已取消导出', 'error')
      return
    }
    const data = await ExportRecordsToCSV(startDate.value, endDate.value)
    const defaultFilename = `记账数据_${startDate.value}_${endDate.value}.csv`
    const destPath = await SaveExportToDir(defaultFilename, dir, data)
    emit('update:show', false)
    emit('toast', `已导出到：${destPath}`, 'success')
  } catch (error) {
    emit('toast', '导出失败: ' + error, 'error')
  }
}
</script>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
.modal { width: 100%; max-width: 400px; border-radius: 20px; padding: 20px; background: var(--card-bg); }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }
.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: var(--text-secondary); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; margin-bottom: 6px; color: var(--text-secondary); }
.form-input { width: 100%; padding: 12px 14px; border-radius: 12px; border: none; font-size: 14px; outline: none; box-sizing: border-box; background: var(--bg-input); color: var(--text-primary); }
.modal-actions { display: flex; gap: 12px; padding-top: 16px; }
.cancel-btn, .confirm-btn { flex: 1; padding: 14px; border-radius: 12px; border: none; font-size: 14px; font-weight: 500; cursor: pointer; }
.cancel-btn { background: var(--bg-input); color: var(--text-secondary); }
.confirm-btn { background: var(--accent-color); color: #fff; }
</style>
