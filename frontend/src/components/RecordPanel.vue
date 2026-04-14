<template>
  <Teleport to="body">
    <!-- 遮罩 -->
    <transition name="overlay">
      <div v-if="visible"
           class="record-modal-overlay"
           :style="{background: isDark ? 'rgba(0,0,0,0.6)' : 'rgba(0,0,0,0.25)'}"
           @click.stop="handleClose" />
    </transition>

    <!-- 弹窗卡片 -->
    <transition name="modal">
      <div v-if="visible" class="record-modal-card" :style="{ background: modalBg }">

        <!-- 头部 -->
        <div class="modal-header-row">
          <h2 class="panel-title">{{ editingRecord ? '编辑记录' : '记一笔' }}</h2>
          <button @click="handleClose" class="panel-close-btn">
            <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- 类别为空提示 -->
        <div v-if="currentCategories.length === 0" class="panel-empty-hint">
          ⚠️ 类别列表为空，请先加载类别数据
        </div>

        <!-- 表单内容 -->
        <div class="panel-body">
          <!-- 智能识别入口 -->
          <div v-if="!editingRecord" class="smart-section">
            <div v-if="!showSmartResult" class="smart-input-row">
              <input v-model="smartText" type="text"
                     style="flex:1; padding:10px 12px; border-radius:12px; font-size:13px; border:none;"
                     :style="{background: isDark ? 'rgba(255,255,255,0.05)' : '#f9fafb', color: isDark ? '#f5f5f7' : '#1f2937'}"
                     placeholder="💡 粘贴文本智能识别，如「美团外卖 25元」"
                     @keydown.enter="handleSmartRecognize" />
              <button @click="handleSmartRecognize" class="smart-btn">识别</button>
            </div>
            <!-- 智能识别结果 -->
            <div v-if="showSmartResult && smartParsed" class="smart-result-card">
              <div class="smart-result-header">
                <span>识别结果</span>
                <button @click="cancelSmartResult" class="smart-cancel-btn">取消</button>
              </div>
              <div class="smart-result-amount" :class="'type-' + smartParsed.type">
                {{ smartParsed.type === 'income' ? '+' : '-' }}¥{{ smartParsed.amount.toFixed(2) }}
              </div>
              <span class="smart-result-tag" :class="'tag-' + smartParsed.type">{{ smartParsed.category }}</span>
              <div v-if="smartParsed.note" class="smart-result-note">备注：{{ smartParsed.note }}</div>
              <div v-if="smartParsed.tags && smartParsed.tags.length > 0" class="smart-result-tags">
                <span v-for="tag in smartParsed.tags" :key="tag" class="result-tag-chip">#{{ tag }}</span>
              </div>
              <button @click="applySmartResult" class="smart-apply-btn">应用到表单</button>
            </div>
          </div>

          <!-- 类型切换 -->
          <div class="type-switcher" :style="{ background: typeSwitchBg }">
            <button type="button" @click="switchType('expense')"
                    class="type-btn"
                    :class="{ active: form.type === 'expense', 'btn-expense': form.type === 'expense' }">
              <span class="type-arrow">↓</span> 支出
            </button>
            <button type="button" @click="switchType('income')"
                    class="type-btn"
                    :class="{ active: form.type === 'income', 'btn-income': form.type === 'income' }">
              <span class="type-arrow">↑</span> 收入
            </button>
          </div>

          <!-- 金额输入 -->
          <div class="form-section">
            <label class="form-label">金额</label>
            <div style="display:flex; align-items:baseline; gap:4px; padding:8px 4px; border-radius:12px;"
                 :style="{background: isDark ? 'rgba(255,255,255,0.05)' : 'rgba(243,244,246,0.8)'}">
              <span class="amount-symbol" :class="'sym-' + form.type">¥</span>
              <input v-model.number="form.amount" type="number" step="0.01" min="0.01" required
                     class="amount-input" :class="'input-' + form.type"
                     placeholder="0.00" inputmode="decimal" />
            </div>
          </div>

          <!-- 类别网格 -->
          <div class="form-section">
            <label class="form-label">类别 {{ form.category ? '· ' + form.category : '' }}</label>
            <div class="category-grid">
              <button v-for="cat in currentCategories" :key="cat.name" type="button"
                      @click="selectCategory(cat)"
                      class="cat-btn"
                      :class="{ active: form.category === cat.name, 'cat-sel-expense': form.category === cat.name && form.type === 'expense', 'cat-sel-income': form.category === cat.name && form.type === 'income' }">
                <span class="cat-icon" :class="{ icon_scale: form.category === cat.name }">{{ cat.icon }}</span>
                <span class="cat-name">{{ cat.name }}</span>
              </button>
            </div>
          </div>

          <!-- 日期和备注 -->
          <div class="date-note-row">
            <div class="date-note-field flex-1">
              <label class="form-label">日期</label>
              <input v-model="form.date" type="date" required
                     style="width:100%; padding:10px 12px; border-radius:12px; font-size:14px; border:none;"
                     :style="{background: isDark ? 'rgba(255,255,255,0.05)' : '#f9fafb', color: isDark ? '#f5f5f7' : '#1f2937'}" />
            </div>
            <div class="date-note-field flex-2">
              <label class="form-label">备注</label>
              <input v-model="form.note" type="text" placeholder="添加备注（可选）"
                     style="width:100%; padding:10px 12px; border-radius:12px; font-size:14px; border:none;"
                     :style="{background: isDark ? 'rgba(255,255,255,0.05)' : '#f9fafb', color: isDark ? '#f5f5f7' : '#1f2937'}" />
            </div>
          </div>

          <!-- 账户选择 -->
          <div v-if="accountsList.length > 1" class="form-section">
            <label class="form-label">
              账户 {{ accountsList.find(a => a.id === form.account_id)?.name ? '· ' + accountsList.find(a => a.id === form.account_id)!.name : '' }}
            </label>
            <div class="account-chips">
              <button v-for="acc in accountsList" :key="acc.id" type="button"
                      @click="form.account_id = acc.id"
                      class="acc-chip"
                      :class="{ active: form.account_id === acc.id }">
                <span class="acc-icon">{{ acc.icon }}</span>{{ acc.name }}
              </button>
            </div>
          </div>

          <!-- 标签选择 -->
          <div v-if="allTags.length > 0" class="form-section">
            <label class="form-label">标签</label>
            <div class="tag-select-chips">
              <button v-for="tag in allTags" :key="tag.id" type="button"
                      @click="toggleTagSelection(tag.id)"
                      class="tag-select-chip"
                      :class="{ active: selectedTags.includes(tag.id) }"
                      :style="selectedTags.includes(tag.id) ? { background: tag.color + '20', color: tag.color, borderColor: tag.color + '50' } : {}">
                <span class="tag-dot" :style="{ background: tag.color }"></span>{{ tag.name }}
              </button>
            </div>
          </div>

          <!-- 提交按钮 -->
          <button type="button" @click="handleSubmit" :disabled="!canSubmit"
                  class="submit-btn"
                  :class="{ 'btn-expense': form.type === 'expense', 'btn-income': form.type === 'income' }"
                  :style="{ opacity: canSubmit ? 1 : 0.3, cursor: canSubmit ? 'pointer' : 'not-allowed' }">
            {{ editingRecord ? '保存修改' : '确认记账' }}
          </button>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, reactive, nextTick } from 'vue'
import { useThemeStore } from '../stores/theme'
import { useAccountingStore, type Record } from '../stores/accounting'
import { createLogger } from '../utils/logger'
import { RecognizeText as SmartRecognize } from '../../wailsjs/go/service/SmartRecognizeService'
import { GetAll as GetAllAccounts } from '../../wailsjs/go/service/AccountService'
import { GetAllTags as GetAllTagsAPI, SetRecordTags as SetRecordTagsAPI, GetRecordTags as GetRecordTagsAPI } from '../../wailsjs/go/service/TagService'

const log = createLogger('Panel')

const props = defineProps<{
  visible: boolean
  editingRecord: Record | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'done'): void
}>()

const theme = useThemeStore()
const store = useAccountingStore()
const isDark = computed(() => theme.isDark)

// 计算属性：暗色模式下的样式值
const modalBg = computed(() => isDark.value ? '#1c1c1e' : '#fff')
const typeSwitchBg = computed(() => isDark.value ? 'rgba(255,255,255,0.05)' : '#f3f4f6')

const form = reactive({
  type: 'expense' as 'income' | 'expense',
  amount: undefined as number | undefined,
  category: '',
  note: '',
  date: new Date().toISOString().split('T')[0],
  account_id: 0,
})

interface Account {
  id: number
  name: string
  icon: string
  is_default: boolean
}

const accountsList = ref<Account[]>([])
const allTags = ref<{ id: number; name: string; color: string }[]>([])
const selectedTags = ref<number[]>([])

async function loadAccounts() {
  try {
    const accs = await GetAllAccounts()
    accountsList.value = accs.map(a => ({ id: a.id, name: a.name, icon: a.icon, is_default: a.is_default }))
    if (form.account_id === 0) {
      const def = accs.find(a => a.is_default)
      if (def) form.account_id = def.id
    }
  } catch (e) {
    log.error('加载账户失败', e)
  }
}

async function loadTags() {
  try {
    const tags = await GetAllTagsAPI()
    allTags.value = (tags || []).map((t: any) => ({ id: t.id, name: t.name, color: t.color }))
  } catch (e) {
    log.error('加载标签失败', e)
  }
}

function toggleTagSelection(tagId: number) {
  const idx = selectedTags.value.indexOf(tagId)
  if (idx > -1) {
    selectedTags.value.splice(idx, 1)
  } else {
    selectedTags.value.push(tagId)
  }
}

// 智能识别
const smartText = ref('')
const showSmartResult = ref(false)
const smartParsed = ref<{
  type: string
  amount: number
  category: string
  note: string
  date: string
  tags: string[]
} | null>(null)

async function handleSmartRecognize() {
  if (!smartText.value.trim()) return
  try {
    const result = await SmartRecognize(smartText.value.trim())
    if (result) {
      smartParsed.value = result
      showSmartResult.value = true
    }
  } catch (e) {
    log.error('智能识别失败', e)
  }
}

function applySmartResult() {
  if (!smartParsed.value) return
  const p = smartParsed.value
  form.type = p.type as 'income' | 'expense'
  form.amount = p.amount
  form.note = p.note || ''
  form.date = p.date || new Date().toISOString().split('T')[0]
  form.category = ''
  smartText.value = ''
  showSmartResult.value = false
  smartParsed.value = null
  nextTick(() => { form.category = p.category; ensureCategory() })
}

function cancelSmartResult() {
  smartText.value = ''
  showSmartResult.value = false
  smartParsed.value = null
}

const currentCategories = computed(() =>
  form.type === 'income' ? store.incomeCategories : store.expenseCategories
)

const canSubmit = computed(() =>
  form.amount && form.amount > 0 && !!form.category && form.date
)

function ensureCategory() {
  if (form.category) return
  const cats = currentCategories.value
  if (cats.length > 0) { form.category = cats[0].name; log.info('自动选中类别:', form.category) }
  else { log.warn('当前类型无可用类别', form.type) }
}

watch(() => form.type, () => { form.category = ''; nextTick(() => ensureCategory()); log.info('切换类型:', form.type) })

watch(() => props.visible, (v) => {
  if (v) {
    loadAccounts()
    loadTags()
    if (props.editingRecord) {
      form.type = props.editingRecord.type as 'income' | 'expense'
      form.amount = props.editingRecord.amount
      form.category = props.editingRecord.category
      form.note = props.editingRecord.note
      form.date = props.editingRecord.date
      form.account_id = (props.editingRecord as Record).account_id || 0
      selectedTags.value = (props.editingRecord.tags || []).map(t => t.id)
      log.info('编辑模式, category:', form.category, 'tags:', selectedTags.value)
    } else {
      resetForm()
    }
    log.info('面板打开, 可用类别:', currentCategories.value.length)
  }
})

watch(() => store.categories, (cats) => {
  log.info('类别数据更新, 总数:', cats.length,
    '支出:', cats.filter(c => c.type === 'expense').length,
    '收入:', cats.filter(c => c.type === 'income').length)
  if (props.visible && !props.editingRecord) { ensureCategory() }
}, { immediate: true, deep: true })

watch(() => props.editingRecord, (record) => {
  if (record && props.visible) {
    form.type = record.type as 'income' | 'expense'
    form.amount = record.amount
    form.category = record.category
    form.note = record.note
    form.date = record.date
  }
}, { immediate: true })

function switchType(type: 'income' | 'expense') { form.type = type }

function selectCategory(cat: { name: string; icon: string }) {
  form.category = cat.name
  log.info('选择类别:', cat.name)
}

function handleClose() {
  log.info('关闭面板')
  emit('close')
}

function resetForm() {
  form.amount = undefined; form.note = ''; form.date = new Date().toISOString().split('T')[0]
  form.type = 'expense'; form.category = ''; form.account_id = 0
  selectedTags.value = []
  ensureCategory()
}

async function handleSubmit() {
  if (!canSubmit.value) { log.warn('提交被阻止', 'canSubmit:', canSubmit.value, 'amount:', form.amount, 'category:', form.category, 'date:', form.date); return }
  try {
    const isEdit = !!props.editingRecord
    const recordId = props.editingRecord?.id
    let newRecordId = recordId
    const formData = { ...form }
    if (isEdit && recordId) {
      await store.updateRecord(recordId, formData.type, formData.category, formData.note, formData.date, formData.amount!)
    } else {
      newRecordId = await store.addRecord(formData.type, formData.category, formData.note, formData.date, formData.amount!, formData.account_id)
    }
    // 保存标签关联
    if (newRecordId) {
      await SetRecordTagsAPI(newRecordId, selectedTags.value)
      log.info('标签已保存, recordId:', newRecordId, 'tags:', selectedTags.value)
    }
    emit('done'); resetForm()
  } catch (e) { log.error('提交失败', e) }
}
</script>

<style scoped>
/* ===== 遮罩 & 弹窗 ===== */
.record-modal-overlay {
  position: fixed; inset: 0; z-index: 999;
}
.record-modal-card {
  position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 380px; max-height: 85vh; border-radius: 20px;
  box-shadow: 0 24px 80px rgba(0,0,0,0.2), 0 8px 24px rgba(0,0,0,0.12);
  z-index: 1000; overflow-y: auto; display: flex; flex-direction: column;
}
.modal-header-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18px 20px 14px; flex-shrink: 0;
}
.panel-title { font-size: 17px; font-weight: 700; color: v-bind('isDark ? "#f5f5f7" : "#111827"'); }

.panel-close-btn {
  width: 32px; height: 32px; border-radius: 50%; border: none; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  background: transparent; transition: all 0.15s; color: v-bind('isDark ? "#98989d" : "#9ca3af"');
}
.panel-close-btn:hover { background: v-bind('isDark ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.05)"'); }

/* 动画 */
.overlay-enter-active { transition: opacity 0.25s ease; }
.overlay-leave-active { transition: opacity 0.15s ease; }
.overlay-enter-from, .overlay-leave-to { opacity: 0; }
.modal-enter-active { transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1); }
.modal-leave-active { transition: all 0.18s ease-in; }
.modal-enter-from { opacity: 0; transform: translate(-50%, -45%) scale(0.95); }
.modal-leave-to { opacity: 0; transform: translate(-50%, -50%) scale(0.97); }

/* 空提示 */
.panel-empty-hint {
  margin: 0 20px 12px; padding: 10px; border-radius: 10px;
  background: rgba(255,69,58,0.1); color: #ff453a; font-size: 13px; text-align: center;
}

/* 表单容器 */
.panel-body { padding: 0 20px 28px; }

/* 智能识别 */
.smart-section { margin-bottom: 16px; }
.smart-input-row { display: flex; gap: 8px; align-items: center; }
.smart-btn {
  padding: 10px 14px; border-radius: var(--input-radius); border: none; cursor: pointer;
  font-size: 12px; font-weight: 600; white-space: nowrap;
  background: var(--accent-color); color: #fff;
}
.smart-result-card {
  border-radius: 12px; padding: 12px;
  background: v-bind('isDark ? "var(--bg-input)" : "#f9fafb"');
}
.smart-result-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;
}
.smart-result-header span { font-size: 13px; font-weight: 600; color: v-bind('isDark ? "#f5f5f7" : "#1f2937"'); }
.smart-cancel-btn {
  border: none; background: transparent; cursor: pointer; font-size: 12px;
  padding: 2px 6px; border-radius: 6px; color: v-bind('isDark ? "#636366" : "#9ca3af"');
}
.smart-result-amount { font-size: 20px; font-weight: 700; margin-bottom: 8px; }
.smart-result-amount.type-income { color: var(--income-color); }
.smart-result-amount.type-expense { color: var(--expense-color); }
.smart-result-tag {
  font-size: 12px; padding: 2px 8px; border-radius: 6px; display: inline-block; margin-bottom: 10px;
}
.smart-result-tag.tag-income { background: var(--income-bg); color: var(--income-color); }
.smart-result-tag.tag-expense { background: var(--expense-bg); color: var(--expense-color); }
.smart-result-note {
  font-size: 12px; margin-bottom: 6px; color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}
.smart-result-tags { display: flex; gap: 4px; margin-bottom: 10px; }
.result-tag-chip {
  font-size: 11px; padding: 2px 6px; border-radius: 4px;
  background: rgba(99,102,241,0.1); color: #6366f1;
}
.smart-apply-btn {
  width: 100%; padding: 8px 0; border-radius: 10px; border: none; cursor: pointer;
  font-size: 13px; font-weight: 600; background: var(--accent-color); color: #fff;
}

/* 类型切换 */
.type-switcher {
  display: flex; gap: 4px; padding: 4px; border-radius: 16px; margin-bottom: 20px;
}
.type-btn {
  flex: 1; padding: 10px 0; border-radius: 12px; border: none; cursor: pointer;
  font-size: 15px; font-weight: 600; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  background: transparent; color: v-bind('isDark ? "#636366" : "#6b7280"');
}
.type-btn.active.btn-expense { background: var(--expense-color); color: #fff; box-shadow: 0 2px 8px rgba(255,69,58,0.25); }
.type-btn.active.btn-income { background: var(--income-color); color: #fff; box-shadow: 0 2px 8px rgba(48,209,88,0.25); }
.type-arrow { font-size: 14px; }

/* 表单区块 */
.form-section { margin-bottom: 20px; }
.form-label {
  font-size: 13px; margin-bottom: 8px; display: block;
  color: v-bind('isDark ? "#636366" : "#9ca3af"');
}

/* 金额输入 */
.amount-symbol { font-size: 28px; font-weight: 700; }
.amount-symbol.sym-income { color: var(--income-color); }
.amount-symbol.sym-expense { color: var(--expense-color); }
.amount-input {
  font-size: 32px; font-weight: 700; background: transparent; border: none;
  width: 100%; padding: 0; outline: none;
}
.amount-input.input-income { color: var(--income-color); }
.amount-input.input-expense { color: var(--expense-color); }

/* 类别网格 */
.category-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
.cat-btn {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 12px 0; border-radius: 16px; border: none; cursor: pointer; transition: all 0.2s;
  background: transparent;
}
.cat-icon { font-size: 22px; transition: transform 0.15s; }
.cat-icon.icon_scale { transform: scale(1.1); }
.cat-name { font-size: 11px; font-weight: 500; line-height: 1; color: v-bind('isDark ? "#636366" : "#6b7280"'); }
.cat-btn.active .cat-name { color: v-bind('isDark ? "#f5f5f7" : "#1f2937"'); }
.cat-btn.cat-sel-expense { background: var(--expense-bg); transform: scale(1.05); }
.cat-btn.cat-sel-expense .cat-name { color: var(--expense-color); font-weight: 600; }
.cat-btn.cat-sel-income { background: var(--income-bg); transform: scale(1.05); }
.cat-btn.cat-sel-income .cat-name { color: var(--income-color); font-weight: 600; }

/* 日期 + 备注 */
.date-note-row { display: flex; gap: 12px; margin-bottom: 16px; }
.date-note-field.flex-1 { flex: 1; }
.date-note-field.flex-2 { flex: 2; }

/* 账户选择 */
.account-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.acc-chip {
  display: flex; align-items: center; gap: 4px; padding: 6px 12px; border-radius: 10px;
  border: none; cursor: pointer; font-size: 13px; transition: all 0.15s;
  background: v-bind('isDark ? "var(--bg-input)" : "#f9fafb"');
  color: v-bind('isDark ? "#636366" : "#6b7280"');
}
.acc-chip.active { background: rgba(99,102,241,0.1); color: #6366f1; font-weight: 600; }
.acc-icon { font-size: 14px; }

/* 标签选择 */
.tag-select-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.tag-select-chip {
  display: flex; align-items: center; gap: 5px; padding: 6px 14px; border-radius: 10px;
  border: 1.5px solid transparent; cursor: pointer; font-size: 13px; transition: all 0.15s;
  background: v-bind('isDark ? "var(--bg-input)" : "#f9fafb"');
  color: v-bind('isDark ? "#636366" : "#6b7280"');
  font-weight: 500;
}
.tag-select-chip:hover {
  background: v-bind('isDark ? "rgba(255,255,255,0.08)" : "#f3f4f6"');
}
.tag-select-chip.active { font-weight: 600; }
.tag-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

/* 提交按钮 */
.submit-btn {
  width: 100%; padding: 14px 0; border-radius: 16px; border: none; cursor: pointer;
  font-size: 16px; font-weight: 700; color: #fff; transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(0,0,0,0.15);
}
.submit-btn.btn-expense { background: var(--expense-color); }
.submit-btn.btn-income { background: var(--income-color); }
</style>
