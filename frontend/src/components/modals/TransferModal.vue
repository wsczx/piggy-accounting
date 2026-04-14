<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ formMode === 'list' ? '转账记录' : '新建转账' }}</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 列表视图 -->
      <template v-if="formMode === 'list'">
        <div v-if="records.length === 0" class="empty-state">
          <div class="empty-icon">🔄</div>
          <p class="empty-text-main">暂无转账记录</p>
          <p class="empty-text-sub">在不同账户之间快速转账</p>
        </div>

        <div class="transfer-list">
          <div v-for="t in records" :key="t.id" class="transfer-item">
            <div class="transfer-item-body">
              <div class="item-icon-wrap">
                {{ getAccountIcon(t.from_account) }}
              </div>
              <div class="item-info">
                <div class="item-name-row">
                  <span>{{ getAccountName(t.from_account) }}</span>
                  <span class="arrow">→</span>
                  <span>{{ getAccountName(t.to_account) }}</span>
                </div>
                <div class="item-meta">
                  <span>{{ t.date }}</span>
                  <span v-if="t.note">· {{ t.note }}</span>
                </div>
              </div>
            </div>
            <div class="item-right">
              <span class="amount">¥{{ t.amount.toFixed(2) }}</span>
              <button @click.stop="handleDelete(t.id)" class="delete-btn">删除</button>
            </div>
          </div>
        </div>

        <div class="list-footer">
          <button @click="startCreate" class="create-btn">＋ 新建转账</button>
        </div>
      </template>

      <!-- 新建表单 -->
      <template v-else>
        <!-- 源账户 -->
        <div class="form-group">
          <label>转出账户</label>
          <div class="chip-grid">
            <button v-for="acc in accounts" :key="'from-' + acc.id" type="button"
              @click="transferForm.from_account = acc.id"
              class="chip-btn"
              :class="{ 'chip-sel-from': transferForm.from_account === acc.id }">
              <span>{{ acc.icon }}</span> {{ acc.name }}
            </button>
          </div>
        </div>

        <!-- 箭头指示 -->
        <div class="arrow-indicator">⬇</div>

        <!-- 目标账户 -->
        <div class="form-group">
          <label>转入账户</label>
          <div class="chip-grid">
            <button v-for="acc in accounts" :key="'to-' + acc.id" type="button"
              @click="transferForm.to_account = acc.id"
              class="chip-btn"
              :class="{ 'chip-sel-to': transferForm.to_account === acc.id }">
              <span>{{ acc.icon }}</span> {{ acc.name }}
            </button>
          </div>
        </div>

        <!-- 金额 -->
        <div class="form-group">
          <label>转账金额</label>
          <input v-model.number="transferForm.amount" type="number" step="0.01" min="0.01" placeholder="0.00"
            class="form-input form-input-lg" />
        </div>

        <!-- 日期 -->
        <div class="form-group">
          <label>日期</label>
          <input v-model="transferForm.date" type="date" class="form-input" />
        </div>

        <!-- 备注 -->
        <div class="form-group">
          <label>备注（可选）</label>
          <input v-model="transferForm.note" type="text" placeholder="例：还款" class="form-input" />
        </div>

        <!-- 预览 -->
        <div class="preview-box">
          <span>💡</span>
          <span>{{ getAccountName(transferForm.from_account) || '未选择' }} → {{ getAccountName(transferForm.to_account) || '未选择' }} ¥{{ (transferForm.amount || 0).toFixed(2) }}</span>
        </div>

        <!-- 按钮 -->
        <div class="modal-actions">
          <button @click="resetForm" class="cancel-btn">取消</button>
          <button @click="handleSave" class="confirm-btn"
            :class="{ 'btn-disabled': !(transferForm.from_account && transferForm.to_account && transferForm.amount) }">
            确认转账
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Create as CreateTransfer, GetAll as GetAllTransfers, Delete as DeleteTransfer } from '../../../wailsjs/go/service/TransferService'
import { createLogger } from '../../utils/logger'

const log = createLogger('Transfer')

interface TransferRecord {
  id: number; from_account: number; to_account: number
  amount: number; note: string; date: string; created_at: string
}

interface AccountInfo {
  id: number; name: string; icon: string; balance: number; real_balance: number; is_default: boolean
}

const props = defineProps<{ show: boolean; isDark: boolean; accounts: AccountInfo[] }>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const formMode = ref<'list' | 'create'>('list')
const records = ref<TransferRecord[]>([])

const transferForm = ref({
  from_account: 0, to_account: 0,
  amount: undefined as number | undefined, note: '', date: '',
})

async function loadTransfers() {
  try { records.value = await GetAllTransfers() }
  catch (e) { log.error('加载转账记录失败', e) }
}
function resetForm() {
  formMode.value = 'list'
  transferForm.value = { from_account: 0, to_account: 0, amount: undefined, note: '', date: '' }
}
function startCreate() {
  formMode.value = 'create'
  const now = new Date()
  transferForm.value = {
    from_account: props.accounts.find(a => a.is_default)?.id || 0,
    to_account: 0, amount: undefined, note: '',
    date: `${now.getFullYear()}-${String(now.getMonth()+1).padStart(2,'0')}-${String(now.getDate()).padStart(2,'0')}`,
  }
}
function getAccountName(id: number): string { return props.accounts.find(a=>a.id===id)?.name || '未知账户' }
function getAccountIcon(id: number): string { return props.accounts.find(a=>a.id===id)?.icon || '💳' }

async function handleSave() {
  const f = transferForm.value
  if (!f.from_account || !f.to_account) { emit('toast','请选择源账户和目标账户','error'); return }
  if (f.from_account === f.to_account) { emit('toast','源账户和目标账户不能相同','error'); return }
  if (!f.amount || f.amount <= 0) { emit('toast','请输入转账金额','error'); return }
  if (!f.date) { emit('toast','请选择转账日期','error'); return }
  try {
    const amt=f.amount||0
    await CreateTransfer(f.from_account,f.to_account,amt,f.note,f.date)
    resetForm(); await loadTransfers(); emit('refresh'); emit('toast','转账成功','success')
  } catch(e){ emit('toast','转账失败: '+e,'error')}
}
async function handleDelete(id:number){try{await DeleteTransfer(id);await loadTransfers();emit('refresh');emit('toast','转账记录已删除','success')}catch(e){emit('toast','删除失败: '+e,'error')}}

watch(() => props.show, (v) => { if (v) loadTransfers() })
</script>

<style scoped>
/* ====== 弹窗基础 ====== */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 100; padding: 20px;
}
.modal {
  width: 100%; max-width: 420px;
  border-radius: var(--radius-xl); padding: 20px;
  background: var(--card-bg);
}
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }

.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }

.close-btn {
  width: 32px; height: 32px; border-radius: 50%; border: none;
  background: transparent; cursor: pointer; display: flex;
  align-items: center; justify-content: center;
  color: var(--text-secondary); transition: background 0.15s;
}
.close-btn:hover { background: rgba(128,128,128,0.15); }

/* ====== 空状态 ====== */
.empty-state { text-align: center; padding: 30px 0; }
.empty-icon { font-size: 36px; margin-bottom: 8px; }
.empty-text-main { color: var(--text-secondary); font-size: 13px; margin: 0; }
.empty-text-sub { color: var(--text-secondary); font-size: 12px; margin-top: 4px; }

/* ====== 表单通用 ====== */
.form-group { margin-bottom: 16px; }
.form-group > label {
  display: block; font-size: 13px; font-weight: 500; margin-bottom: 6px;
  color: var(--text-secondary);
}
.form-input {
  width: 100%; padding: 10px 12px; border-radius: var(--radius-md);
  border: none; font-size: 14px; outline: none; box-sizing: border-box;
  background: var(--bg-input); color: var(--text-primary);
}
.form-input-lg { font-size: 18px; font-weight: 600; }

.modal-actions { display: flex; gap: 10px; margin-top: 20px; }

.cancel-btn {
  flex: 1; padding: 12px; border-radius: var(--radius-lg);
  border: none; cursor: pointer; font-size: 14px; font-weight: 500;
  background: var(--bg-input); color: var(--text-primary);
}
.confirm-btn {
  flex: 1; padding: 12px; border-radius: var(--radius-lg);
  border: none; cursor: pointer; font-size: 14px; font-weight: 600;
  background: var(--accent-color); color: #fff;
}
.btn-disabled { opacity: 0.5; cursor: not-allowed !important; }

/* ====== 列表视图 ====== */
.transfer-list { display: flex; flex-direction: column; gap: 8px; max-height: 50vh; overflow-y: auto; }

.transfer-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px; border-radius: var(--radius-lg);
  background: var(--bg-input);
}
.transfer-item-body { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }
.item-icon-wrap {
  width: 36px; height: 36px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 16px; flex-shrink: 0; background: rgba(245,158,11,0.1);
}
.item-info { min-width: 0; }
.item-name-row {
  font-size: 13px; font-weight: 500; display: flex; align-items: center; gap: 4px;
  color: var(--text-primary);
}
.arrow { color: var(--text-secondary); }
.item-meta { font-size: 11px; display: flex; gap: 8px; margin-top: 2px; color: var(--text-secondary); }

.item-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.amount { font-size: 14px; font-weight: 600; color: #f59e0b; }
.delete-btn { padding: 4px 8px; border-radius: 6px; border: none; cursor: pointer; font-size: 12px; background: rgba(255,69,58,0.1); color: #ff453a; }

.list-footer { margin-top: 12px; }
.create-btn {
  width: 100%; padding: 12px; border-radius: var(--radius-lg); border: none;
  cursor: pointer; font-size: 14px; font-weight: 600;
  background: var(--accent-color); color: #fff;
}

/* ====== Chip Grid ====== */
.chip-grid { display: flex; flex-wrap: wrap; gap: 6px; }

.chip-btn {
  display: flex; align-items: center; gap: 4px; padding: 8px 12px;
  border-radius: 10px; border: none; cursor: pointer; font-size: 13px;
  transition: all 0.15s; background: var(--bg-input);
  color: var(--text-secondary);
}
.chip-sel-from { background: rgba(255,69,58,0.1); color: #ff453a; font-weight: 600; }
.chip-sel-to { background: rgba(48,209,88,0.1); color: #30d158; font-weight: 600; }

.arrow-indicator { text-align: center; font-size: 20px; color: var(--accent-color); margin: -4px 0; }

/* Preview */
.preview-box {
  padding: 10px 12px; border-radius: 10px; margin-bottom: 16px;
  font-size: 12px; display: flex; gap: 6px; align-items: center;
  background: var(--preview-bg);
  color: #f59e0b;
}

/* 暗色模式下预览框背景稍深 */
:global(.dark) .preview-box {
  background: rgba(245, 158, 11, 0.1);
}
</style>
