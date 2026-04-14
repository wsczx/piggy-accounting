<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal" :class="{ 'modal-dark': isDark }">
      <div class="modal-header">
        <span class="modal-title">导入数据</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>
      <div class="form-group">
        <label class="im-label">导入来源</label>
        <div class="import-type-selector">
          <button v-for="type in types" :key="type.key" @click="importType = type.key as 'csv' | 'wechat' | 'alipay'"
            class="type-btn" :class="{ active: importType === type.key }">{{ type.label }}</button>
        </div>
      </div>
      <div class="import-hint">
        <p v-if="importType === 'csv'">支持标准 CSV 格式：日期,类型,类别,金额,备注</p>
        <p v-else-if="importType === 'wechat'">支持微信账单 CSV 导出文件</p>
        <p v-else>支持支付宝账单 CSV 导出文件</p>
      </div>
      <div v-if="importResult" class="import-result">
        <div class="result-item success"><span class="result-icon">✓</span><span>成功导入 {{ importResult.successCount }} 条</span></div>
        <div v-if="importResult.skipCount > 0" class="result-item skip"><span class="result-icon">⊘</span><span>跳过重复 {{ importResult.skipCount }} 条</span></div>
        <div v-if="importResult.errorCount > 0" class="result-item error"><span class="result-icon">✗</span><span>失败 {{ importResult.errorCount }} 条</span></div>
      </div>
      <div class="modal-actions">
        <button @click="$emit('update:show', false)" class="cancel-btn">取消</button>
        <label class="confirm-btn" :class="{ disabled: loading }">
          <input type="file" accept=".csv" class="im-file-input" @change="handleFileSelect" :disabled="loading" />
          {{ loading ? '导入中...' : '选择文件' }}
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ImportFromCSV as ImportRecordsFromCSV, ImportFromWeChat, ImportFromAlipay } from '../../../wailsjs/go/service/ExportImportService'

const props = defineProps<{
  show: boolean
  isDark: boolean
  store: { loadRecords(): Promise<void> }
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const importType = ref<'csv' | 'wechat' | 'alipay'>('csv')
const loading = ref(false)
const importResult = ref<{ successCount: number; skipCount: number; errorCount: number } | null>(null)

interface ImportTypeOption {
  key: string
  label: string
}

const types: ImportTypeOption[] = [
  { key: 'csv', label: '标准CSV' },
  { key: 'wechat', label: '微信账单' },
  { key: 'alipay', label: '支付宝' },
]

async function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    loading.value = true
    importResult.value = null
    const arrayBuffer = await file.arrayBuffer()
    const data = Array.from(new Uint8Array(arrayBuffer))
    let result
    switch (importType.value) {
      case 'wechat': result = await ImportFromWeChat(data, true); break
      case 'alipay': result = await ImportFromAlipay(data, true); break
      default: result = await ImportRecordsFromCSV(data, true);
    }
    importResult.value = result
    await props.store.loadRecords()
    if (result.errorCount === 0) {
      setTimeout(() => { emit('update:show', false); importResult.value = null }, 2000)
    }
  } catch (error) {
    emit('toast', `导入失败: ${error}`, 'error')
  } finally {
    loading.value = false
    input.value = ''
  }
}
</script>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
.modal { width: 100%; max-width: 400px; border-radius: 20px; padding: 20px; background: #fff; }
.modal-dark.modal { background: #1c1c1e; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.modal-title { font-size: 16px; font-weight: 600; color: #1f2937; }
.modal-dark .modal-title { color: #fff; }
.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: #9ca3af; }
.form-group { margin-bottom: 16px; }
.im-label { display: block; font-size: 13px; margin-bottom: 6px; color: #6b7280; }
.modal-dark .im-label { color: #8e8e93; }
.import-type-selector { display: flex; gap: 8px; }
.type-btn {
  flex: 1; padding: 10px; border-radius: 10px; border: none;
  font-size: 13px; cursor: pointer; transition: all 0.15s;
  background: #f3f4f6; color: #6b7280;
}
.type-btn.active { background: var(--accent-color); color: #fff; }
.modal-dark .type-btn:not(.active) { background: #2c2c2e; color: #8e8e93; }
.import-hint { font-size: 12px; margin-bottom: 16px; padding: 10px; border-radius: 8px; background: rgba(0,0,0,0.03); color: #9ca3af; }
.modal-dark .import-hint { color: #636366; background: rgba(255,255,255,0.05); }
.import-result { margin-bottom: 16px; padding: 12px; border-radius: 12px; background: rgba(0,0,0,0.03); }
.result-item { display: flex; align-items: center; gap: 8px; font-size: 13px; margin-bottom: 6px; }
.result-item:last-child { margin-bottom: 0; }
.result-item.success { color: #30d158; }
.result-item.skip { color: #ff9f0a; }
.result-item.error { color: #ff453a; }
.result-icon { width: 18px; height: 18px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: bold; }
.result-item.success .result-icon { background: rgba(48,209,88,0.15); }
.result-item.skip .result-icon { background: rgba(255,159,10,0.15); }
.result-item.error .result-icon { background: rgba(255,69,58,0.15); }
.modal-actions { display: flex; gap: 12px; padding-top: 16px; }
.cancel-btn, .confirm-btn { flex: 1; padding: 14px; border-radius: 12px; border: none; font-size: 14px; font-weight: 500; cursor: pointer; text-align: center; }
.cancel-btn { background: #f3f4f6; color: #6b7280; }
.confirm-btn { background: var(--accent-color); color: #fff; cursor: pointer; }
.confirm-btn.disabled { opacity: 0.6; cursor: not-allowed; }
.im-file-input { display: none; }
</style>
