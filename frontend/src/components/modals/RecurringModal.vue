<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ formMode === 'list' ? '周期记账' : (formMode === 'edit' ? '编辑周期记账' : '新建周期记账') }}</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <!-- ===== 列表视图 ===== -->
      <template v-if="formMode === 'list'">
        <!-- 空状态 -->
        <div v-if="records.length === 0" class="empty-state">
          <div class="empty-icon">⏰</div>
          <p class="empty-text-main">暂无周期记账规则</p>
          <p class="empty-text-sub">设置定期自动记账，比如每月房租、每周交通费</p>
        </div>

        <!-- 规则列表 -->
        <div class="recurring-list">
          <div v-for="record in records" :key="record.id" class="recurring-item" :class="{ disabled: !record.enabled }">
            <div class="recurring-item-body">
              <div class="item-icon-wrap" :class="'type-' + record.type">{{ getCategoryIcon(record.category) }}</div>
              <div class="item-info">
                <div class="item-name-row">
                  <span class="item-name">{{ record.category }}</span>
                  <span v-if="record.note" class="item-note">{{ record.note }}</span>
                </div>
                <div class="item-meta">{{ getFrequencyDetail(record) }} · 下次: {{ record.next_date }}</div>
              </div>
            </div>
            <div class="item-right">
              <span class="item-amount" :class="'amount-' + record.type">¥{{ record.amount.toFixed(2) }}</span>
              <button @click.stop="handleToggle(record.id)" class="toggle-switch toggle-sm" :class="{ on: record.enabled }"><div class="toggle-thumb"></div></button>
              <div class="recurring-actions">
                <button @click.stop="startEdit(record)" class="action-btn action-edit">编辑</button>
                <button @click.stop="handleDelete(record.id)" class="action-btn action-delete">删除</button>
              </div>
            </div>
          </div>
        </div>

        <!-- 底部按钮 -->
        <div v-if="formMode === 'list'" class="list-footer">
          <button @click="startCreate" class="create-btn">＋ 新建周期记账</button>
        </div>
      </template>

      <!-- ===== 编辑/新建表单 ===== -->
      <template v-else>
        <!-- 类型切换 -->
        <div class="type-switcher">
          <button v-for="t in [{ value: 'expense', label: '支出' }, { value: 'income', label: '收入' }]" :key="t.value"
            @click="form.type = t.value as 'expense' | 'income'"
            class="type-btn" :class="[form.type === t.value ? ('active type-active-' + t.value) : '']">{{ t.label }}</button>
        </div>

        <!-- 金额输入 -->
        <div class="form-group"><label>金额</label><input v-model.number="form.amount" type="number" step="0.01" min="0.01" placeholder="0.00" class="form-input form-input-lg" /></div>

        <!-- 类别选择 -->
        <div class="form-group"><label>类别</label><div class="chip-grid">
          <button v-for="cat in categories" :key="cat.name" type="button" @click="form.category = cat.name"
            class="chip-btn" :class="[form.category === cat.name ? 'chip-sel' : '', 'chip-' + (form.type)]"><span>{{ cat.icon }}</span> {{ cat.name }}</button>
        </div></div>

        <!-- 周期选择 -->
        <div class="form-group"><label>周期</label><div class="chip-grid chip-grid-sm">
          <button v-for="opt in frequencyOptions" :key="opt.value" @click="form.frequency = opt.value"
            class="chip-btn" :class="{ 'chip-sel': form.frequency === opt.value }">{{ opt.label }}</button>
        </div></div>

        <!-- 星期几 / 每月几号 / 年度月份日期 -->
        <div v-if="form.frequency === 'weekly'" class="form-group"><label>星期几</label><div class="chip-grid chip-grid-xs">
          <button v-for="(day, idx) in weekDayOptions" :key="day" @click="form.week_day = idx + 1"
            class="chip-btn" :class="{ 'chip-sel': form.week_day === idx + 1 }">{{ day }}</button>
        </div></div>
        <div v-if="form.frequency === 'monthly'" class="form-group"><label>每月几号</label><div class="day-grid">
          <button v-for="d in 31" :key="d" type="button" @click="form.month_day = d"
            class="day-btn" :class="{ 'day-sel': form.month_day === d }">{{ d }}</button>
        </div></div>
        <div v-if="form.frequency === 'yearly'" class="form-group"><label>月份</label><div class="chip-grid chip-grid-xs">
          <button v-for="m in 12" :key="m" type="button" @click="form.year_month = m"
            class="chip-btn" :class="{ 'chip-sel': form.year_month === m }">{{ m }}月</button>
        </div></div>
        <div v-if="form.frequency === 'yearly'" class="form-group"><label>日期</label><div class="day-grid">
          <button v-for="d in 31" :key="d" type="button" @click="form.month_day = d"
            class="day-btn" :class="{ 'day-sel': form.month_day === d }">{{ d }}</button>
        </div></div>

        <!-- 账户选择 -->
        <div v-if="accounts.length > 0" class="form-group"><label>账户（可选）</label><div class="chip-grid">
          <button v-for="acc in accounts" :key="'rec-acc-' + acc.id" type="button"
            @click="form.account_id = form.account_id === acc.id ? 0 : acc.id"
            class="chip-btn" :class="{ 'chip-sel-account': form.account_id === acc.id }"><span>{{ acc.icon }}</span> {{ acc.name }}</button>
        </div></div>

        <!-- 备注 -->
        <div class="form-group"><label>备注（可选）</label><input v-model="form.note" type="text" placeholder="例：房租" class="form-input" /></div>

        <!-- 预览 -->
        <div class="preview-box"><span>💡</span><span>将每{{ getFrequencyLabel(form.frequency) }}自动记账 ¥{{ (form.amount || 0).toFixed(2) }} → {{ form.category || '未选择' }}{{ form.note ? ' (' + form.note + ')' : '' }}</span></div>

        <!-- 按钮 -->
        <div class="modal-actions">
          <button @click="cancelForm" class="cancel-btn">取消</button>
          <button @click="handleSave" :disabled="!canSubmit" class="confirm-btn" :class="{ 'btn-disabled': !canSubmit }">{{ formMode === 'edit' ? '保存修改' : '创建规则' }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { GetAll as GetAllRecurringRecords, Create as CreateRecurringRecord, Update as UpdateRecurringRecord, Delete as DeleteRecurringRecord, ToggleEnabled as ToggleRecurringRecord } from '../../../wailsjs/go/service/RecurringService'
import { getCategoryIcon as _getCatIcon } from '../../utils/category'
import { createLogger } from '../../utils/logger'

const log = createLogger('Recurring')

interface RecurringRecord { id: number; type: string; amount: number; category: string; note: string; account_id: number; frequency: string; week_day: number; month_day: number; year_month: number; next_date: string; enabled: boolean }
interface AccountInfo { id: number; name: string; icon: string; balance: number; real_balance: number; is_default: boolean }

const frequencyOptions = [ { value: 'daily', label: '每天' }, { value: 'weekly', label: '每周' }, { value: 'monthly', label: '每月' }, { value: 'yearly', label: '每年' } ]
const weekDayOptions = ['周一','周二','周三','周四','周五','周六','周日']

const props = defineProps<{ show: boolean; isDark: boolean; accounts: AccountInfo[]; store: { expenseCategories?: Array<{id?:number;name:string;icon:string;type:string;is_system?:boolean}>; incomeCategories?: any; categories?: any; loadRecords?():Promise<void>} }>()

const emit = defineEmits<{ (e:'update:show', value:boolean):void; (e:'toast', msg:string, type:'success'|'error'):void; (e:'refresh'):void }>()

const formMode = ref<'list'|'create'|'edit'>('list')
const editingId = ref(0)
const records = ref<RecurringRecord[]>([])
const form = ref({ type: 'expense' as 'income'|'expense', amount: undefined as number|undefined, category:'', note:'', frequency:'monthly', week_day:1, month_day:1, year_month:1, account_id:0, enabled:true })

const categories = computed(() => form.value.type === 'income' ? props.store.incomeCategories : props.store.expenseCategories)
const canSubmit = computed(() => form.value.amount && form.value.amount > 0 && !!form.value.category)

async function loadRecords() { try { records.value = await GetAllRecurringRecords() } catch(e){log.error('加载周期记账失败',e)} }
function resetForm() { formMode.value='list';editingId.value=0;form.value={type:'expense',amount:undefined,category:'',note:'',frequency:'monthly',week_day:1,month_day:1,year_month:1,account_id:0,enabled:true} }
function startCreate() {
  formMode.value='create';editingId.value=0
  form.value={type:'expense',amount:undefined,category:'',note:'',frequency:'monthly',week_day:1,month_day:1,year_month:1,account_id:props.accounts.find(a=>a.is_default)?.id||0,enabled:true}
  const cats=props.store.expenseCategories||[]
  if(cats.length>0) form.value.category=cats[0].name
}
function startEdit(r:RecurringRecord) {formMode.value='edit';editingId.value=r.id;form.value={type:r.type as any,amount:r.amount,category:r.category,note:r.note,frequency:r.frequency,week_day:r.week_day||1,month_day:r.month_day||1,year_month:r.year_month||1,account_id:r.account_id||0,enabled:r.enabled}}
function cancelForm(){resetForm()}
function getFrequencyLabel(f:string){if(f==='daily')return'每天';if(f==='weekly')return'每周';if(f==='monthly')return'每月';return f==='yearly'?'每年':f}

function getFrequencyDetail(r:RecurringRecord){
  switch(r.frequency){
    case'daily':return'每天执行'
    case'weekly':{const d=['','周一','周二','周三','周四','周五','周六','周日'];return`每${r.week_day>=1&&r.week_day<=7?d[r.week_day]:'周'}`}
    case'monthly':return r.month_day>0?`每月${r.month_day}日`:'每月执行'
    case'yearly':{if(r.year_month>0&&r.month_day>0)return`每年${r.year_month}月${r.month_day}日`;if(r.year_month>0)return`每年${r.year_month}月`;return'每年执行'}
    default:return r.frequency
  }
}
function getCategoryIcon(name:string):string{return _getCatIcon(name,(props.store.categories||[])as any)}

async function handleSave(){
  if(!canSubmit.value)return
  try{
    const f=form.value;const amt=f.amount||0
    if(formMode.value==='create')await CreateRecurringRecord(f.type,f.category,f.note,f.frequency,amt,f.week_day,f.month_day,f.year_month,f.account_id||0)
    else await UpdateRecurringRecord(editingId.value,f.type,f.category,f.note,f.frequency,amt,f.week_day,f.month_day,f.year_month,f.account_id||0,f.enabled)
    const wasCreate=formMode.value==='create';formMode.value='list';editingId.value=0;f.amount=undefined;f.category='';f.note=''
    await loadRecords();emit('refresh');emit('toast',wasCreate?'周期记账创建成功':'周期记账已保存','success')
  }catch(e){emit('toast','保存失败: '+e,'error')}
}
async function handleDelete(id:number){try{await DeleteRecurringRecord(id);await loadRecords();emit('toast','周期记账已删除','success')}catch(e){emit('toast','删除失败: '+e,'error')}}
async function handleToggle(id:number){try{await ToggleRecurringRecord(id);await loadRecords()}catch(e){emit('toast','操作失败: '+e,'error')}}
watch(() => props.show, (v) => { if (v) loadRecords() })
</script>

<style scoped>
/* ====== 弹窗基础 ====== */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
.modal { width: 100%; max-width: 420px; border-radius: var(--radius-xl); padding: 20px; background: var(--card-bg); max-height: 80vh; overflow-y: auto; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }

.close-btn { width: 32px; height: 32px; border-radius: 50%; border: none; background: transparent; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-secondary); transition: background .15s; }
.close-btn:hover { background: var(--bg-input); }

/* 空状态 */
.empty-state { text-align: center; padding: 30px 0; }
.empty-icon { font-size: 36px; margin-bottom: 8px; }
.empty-text-main { color: var(--text-secondary); font-size: 13px; margin: 0; }
.empty-text-sub { color: var(--text-muted); font-size: 12px; margin-top: 4px; }

/* 表单通用 */
.form-group { margin-bottom: 16px; }
.form-group > label { display: block; font-size: 13px; font-weight: 500; margin-bottom: 6px; color: var(--text-secondary); }
.form-input { width: 100%; padding: 10px 12px; border-radius: var(--radius-md); border: none; font-size: 14px; outline: none; box-sizing: border-box; background: var(--bg-input); color: var(--text-primary); }
.form-input-lg { font-size: 18px; font-weight: 600; }

.modal-actions { display: flex; gap: 10px; margin-top: 20px; }
.cancel-btn { flex: 1; padding: 12px; border-radius: var(--radius-lg); border: none; cursor: pointer; font-size: 14px; font-weight: 500; background: var(--bg-input); color: var(--text-primary); }
.confirm-btn { flex: 1; padding: 12px; border-radius: var(--radius-lg); border: none; cursor: pointer; font-size: 14px; font-weight: 600; background: var(--accent-color); color: #fff; }
.btn-disabled { opacity: 0.5; cursor: not-allowed !important; }

/* 列表视图 */
.recurring-list { display: flex; flex-direction: column; gap: 8px; max-height: 50vh; overflow-y: auto; }
.recurring-item { display: flex; align-items: center; gap: 10px; padding: 12px; border-radius: var(--radius-lg); background: var(--bg-input); transition: opacity 0.15s; }
.recurring-item.disabled { opacity: 0.5; }
.recurring-item:hover .recurring-actions { display: flex; }

.recurring-item-body { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }
.item-icon-wrap { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 16px; flex-shrink: 0; }
.item-icon-wrap.type-income { background: var(--income-bg); }
.item-icon-wrap.type-expense { background: var(--expense-bg); }

.item-info { min-width: 0; flex: 1; }
.item-name-row { font-size: 14px; font-weight: 500; display: flex; align-items: center; gap: 6px; }
.item-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-primary); }
.item-note { font-size: 11px; opacity: 0.6; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-secondary); }
.item-meta { font-size: 11px; display: flex; gap: 8px; margin-top: 2px; color: var(--text-muted); }

.item-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.item-amount { font-size: 14px; font-weight: 600; }
.item-amount.amount-income { color: var(--income-color); }
.item-amount.amount-expense { color: var(--expense-color); }

.recurring-actions { display: none; gap: 4px; }
.action-btn { padding: 4px 8px; border-radius: 6px; border: none; cursor: pointer; font-size: 12px; }
.action-edit { background: var(--bg-input); color: var(--text-secondary); }
.action-delete { background: rgba(255,69,58,0.1); color: #ff453a; }

.list-footer { margin-top: 16px; }
.create-btn { width: 100%; padding: 12px; border-radius: var(--radius-lg); border: none; cursor: pointer; font-size: 14px; font-weight: 600; background: var(--accent-color); color: #fff; transition: opacity 0.15s; }
.create-btn:hover { opacity: 0.9; }

/* Type Switcher */
.type-switcher { display: flex; gap: 4px; padding: 4px; border-radius: 14px; margin-bottom: 16px; background: var(--bg-input); }
.type-btn { flex: 1; padding: 8px 0; border-radius: 10px; border: none; cursor: pointer; font-size: 14px; font-weight: 600; transition: all 0.2s; background: transparent; color: var(--text-muted); }
.type-btn.active { box-shadow: 0 2px 8px rgba(0,0,0,0.25); }
.type-active-expense { background: var(--expense-color); color: #fff; }
.type-active-income { background: var(--income-color); color: #fff; }

/* Chip Grid */
.chip-grid { display: flex; flex-wrap: wrap; gap: 6px; }
.chip-grid-sm .chip-btn { font-size: 13px; font-weight: 500; }
.chip-grid-xs .chip-btn { font-size: 12px; padding: 8px 0; }
.chip-btn { display: flex; align-items: center; gap: 4px; padding: 6px 12px; border-radius: 10px; border: none; cursor: pointer; font-size: 13px; transition: all 0.15s; background: var(--bg-input); color: var(--text-muted); }
.chip-sel { font-weight: 600; }
.chip-expense.chip-sel { background: var(--expense-bg); color: var(--expense-color); }
.chip-income.chip-sel { background: var(--income-bg); color: var(--income-color); }
.chip-sel-account { background: rgba(99,102,241,0.15); color: #6366f1; font-weight: 600; }

.day-grid { display: flex; flex-wrap: wrap; gap: 6px; max-height: 160px; overflow-y: auto; padding: 2px; }
.day-btn { width: 40px; height: 36px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 500; transition: all 0.15s; background: var(--bg-input); color: var(--text-muted); }
.day-sel { background: var(--accent-color); color: #fff; }

/* Preview */
.preview-box { padding: 10px 12px; border-radius: 10px; margin-bottom: 16px; font-size: 12px; display: flex; gap: 6px; align-items: center; background: var(--bg-input); color: var(--text-secondary); }

/* Toggle Switch */
.toggle-switch { width: 40px; height: 24px; border-radius: 12px; padding: 2px; border: none; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; flex-shrink: 0; background: var(--switch-track); }
.toggle-switch.on { background: var(--accent-color); }
.toggle-switch.toggle-sm .toggle-thumb { width: 20px; height: 20px; }
.toggle-thumb { width: 20px; height: 20px; border-radius: 50%; background: #fff; transition: all 0.2s; box-shadow: 0 1px 2px rgba(0,0,0,0.2); flex-shrink: 0; }
.toggle-switch.on .toggle-thumb { transform: translateX(16px); }
</style>
