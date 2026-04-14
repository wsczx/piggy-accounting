<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal budget-modal">
      <div class="modal-header">
        <span class="modal-title">预算设置</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 月度预算 -->
      <div class="budget-section">
        <div class="section-row">
          <span class="section-label">📅 月度预算</span>
          <span v-if="monthlyBudget" class="section-current">当前: ¥{{ monthlyBudget.budget_amount.toFixed(0) }}</span>
        </div>
        <div class="input-row">
          <input v-model.number="monthlyBudgetInput" type="number" min="0" step="100" placeholder="输入月度预算金额"
            class="budget-input" />
          <button @click="saveMonthlyBudget()" class="save-btn" :disabled="!monthlyBudgetInput || monthlyBudgetInput <= 0">保存</button>
        </div>
      </div>

      <!-- 年度预算 -->
      <div class="budget-section">
        <div class="section-row">
          <span class="section-label">📆 年度预算</span>
          <span v-if="yearlyBudget" class="section-current">当前: ¥{{ yearlyBudget.budget_amount.toFixed(0) }}</span>
        </div>
        <div class="input-row">
          <input v-model.number="yearlyBudgetInput" type="number" min="0" step="1000" placeholder="输入年度预算金额"
            class="budget-input" />
          <button @click="saveYearlyBudget()" class="save-btn" :disabled="!yearlyBudgetInput || yearlyBudgetInput <= 0">保存</button>
        </div>
      </div>

      <!-- 使用情况提示 -->
      <div v-if="monthlyBudget" class="usage-card" :class="'usage-' + budgetLevel(monthlyBudget.percentage)">
        <div class="usage-header">
          <span>本月使用情况</span>
          <span class="usage-pct">{{ Math.round(monthlyBudget.percentage) }}%</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: Math.min(monthlyBudget.percentage, 100) + '%' }"></div>
        </div>
        <div class="usage-meta">
          <span>已用 ¥{{ formatAmount(monthlyBudget.spent) }}</span>
          <span :class="{ 'text-danger': monthlyBudget.remaining < 0 }">{{
            monthlyBudget.remaining >= 0 ? ('剩余 ¥' + formatAmount(monthlyBudget.remaining)) : ('超支 ¥' + formatAmount(Math.abs(monthlyBudget.remaining))) }}</span>
        </div>
      </div>

      <!-- 删除按钮 -->
      <div v-if="monthlyBudget || yearlyBudget" class="delete-row">
        <button v-if="monthlyBudget" @click="handleDeleteBudget('monthly')" class="del-btn">删除月度预算</button>
        <button v-if="yearlyBudget" @click="handleDeleteBudget('yearly')" class="del-btn">删除年度预算</button>
      </div>

      <!-- 内联确认对话框 -->
      <Teleport to="body">
        <div v-if="confirm.show" class="ic-overlay" @click.self="confirm.show = false">
          <div class="ic-dialog">
            <div class="ic-icon">⚠️</div>
            <div class="ic-title">{{ confirm.title }}</div>
            <div v-if="confirm.message" class="ic-msg">{{ confirm.message }}</div>
            <div class="ic-actions">
              <button @click="confirm.show = false" class="ic-cancel">取消</button>
              <button @click="confirm.onConfirm" :class="confirm.dangerColor ? 'ic-danger' : 'ic-ok'">{{ confirm.confirmText || '确定' }}</button>
            </div>
          </div>
        </div>
      </Teleport>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { formatAmount } from '../../utils/formatters'
import type { BudgetInfo } from '../../stores/accounting'

const props = defineProps<{
  show: boolean; isDark: boolean
  monthlyBudget: BudgetInfo | null; yearlyBudget: BudgetInfo | null
  store: any
}>()
const emit = defineEmits<{ (e:'update:show',v:boolean):void; (e:'toast',msg:string,type:'success'|'error'):void; (e:'refresh'):void }>()

const monthlyBudgetInput = ref<number | null>(null)
const yearlyBudgetInput = ref<number | null>(null)

const confirm = ref<{ show:boolean; title:string; message?:string; confirmText?:string; dangerColor?:boolean; onConfirm:()=>void }>({ show:false,title:'',onConfirm:()=>{} })

function showConfirm(opts:{ title:string; message?:string; confirmText?:string; dangerColor?:boolean }): Promise<boolean> {
  return new Promise(resolve => { confirm.value={ show:true,title:opts.title,message:opts.message||'',confirmText:opts.confirmText||'确定',dangerColor:opts.dangerColor??false,onConfirm:()=>{confirm.value.show=false;resolve(true)} } })
}

function budgetLevel(pct: number): string { if (pct >= 100) return 'over'; if (pct >= 80) return 'warn'; return 'ok'; }

watch(() => props.show, (val) => {
  if (val) { monthlyBudgetInput.value=props.monthlyBudget?props.monthlyBudget.budget_amount:null; yearlyBudgetInput.value=props.yearlyBudget?props.yearlyBudget.budget_amount:null }
})

async function saveMonthlyBudget() {
  const a=monthlyBudgetInput.value; if(!a||a<=0)return;
  try{const[y,m]=(props.store.currentMonth||'').split('-').map(Number); await props.store.setBudget('monthly',y,m,a); emit('toast','月度预算已保存','success'); emit('refresh')}
  catch(e){emit('toast','设置月度预算失败: '+e,'error')}
}
async function saveYearlyBudget() {
  const a=yearlyBudgetInput.value; if(!a||a<=0)return;
  try{const[y]=(props.store.currentMonth||'').split('-').map(Number); await props.store.setBudget('yearly',y,0,a); emit('toast','年度预算已保存','success'); emit('refresh')}
  catch(e){emit('toast','设置年度预算失败: '+e,'error')}
}
async function handleDeleteBudget(type:'monthly'|'yearly'){
  if(!await showConfirm({title:'确定要删除'+(type==='monthly'?'月度':'年度')+'预算吗？',dangerColor:true}))return;
  try{const[y,m]=(props.store.currentMonth||'').split('-').map(Number);const mn=type==='monthly'?m:0; await props.store.deleteBudget(type,y,mn); emit('toast',(type==='monthly'?'月度':'年度')+'预算已删除','success'); emit('refresh')}
  catch(e){emit('toast','删除预算失败: '+e,'error')}
}
</script>

<style scoped>
/* ====== 弹窗基础 ====== */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.5); display:flex; align-items:center; justify-content:center; z-index:100; padding:20px; }
.modal { width:340px; max-height:85vh; overflow-y:auto; border-radius:20px; padding:20px; background:var(--card-bg); box-shadow:var(--card-shadow); position:relative;}
.modal-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:20px; }
.modal-title { font-size:16px; font-weight:600; color:v-bind('isDark ? "#fff" : "#1f2937"'); }
.close-btn { padding:4px; border:none; background:transparent; cursor:pointer; border-radius:8px; color:#9ca3af; }

/* ====== 预算区块 ====== */
.budget-section { padding:14px 12px; border-radius:12px; margin-bottom:10px; background:var(--bg-input); }
.section-row { display:flex; align-items:center; justify-content:space-between; margin-bottom:10px; }
.section-label { font-size:14px; font-weight:600; color:v-bind('isDark ? "#f5f5f7" : "#1f2937"'); }
.section-current { font-size:12px; color:v-bind('isDark ? "#636366" : "#9ca3af"'); }

.input-row { display:flex; gap:8px; }
.budget-input { flex:1; padding:10px 14px; border-radius:12px; font-size:15px; font-weight:500; border:none; outline:none; background:var(--card-bg); color:v-bind('isDark ? "#fff" : "#1f2937"'); }
.save-btn { padding:10px 16px; border-radius:12px; border:none; cursor:pointer; font-size:13px; font-weight:600; background:var(--accent-color); color:#fff; white-space:nowrap; }
.save-btn:disabled { opacity:0.5; cursor:not-allowed; }

/* ====== 使用情况 ====== */
.usage-card { margin-top:12px; padding:12px; border-radius:12px; border:1px solid transparent; transition:border-color 0.2s; }
.usage-ok { background:v-bind('isDark ? "rgba(48,209,88,0.06)" : "rgba(48,209,88,0.04)"'); border-color: rgba(48,209,88,0.1); }
.usage-warn { background:v-bind('isDark ? "rgba(255,159,10,0.08)" : "rgba(255,159,10,0.04)"'); border-color: rgba(255,159,10,0.2); }
.usage-over { background:v-bind('isDark ? "rgba(255,69,58,0.08)" : "rgba(255,69,58,0.04)"'); border-color: rgba(255,69,58,0.2); }

.usage-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:6px; }
.usage-header span:first-child { font-size:12px; font-weight:600; color:v-bind('isDark ? "#f5f5f7" : "#374151"'); }
.usage-pct { font-size:13px; font-weight:700; }
.usage-ok .usage-pct { color:#30d158; } .usage-warn .usage-pct { color:#ff9f0a; } .usage-over .usage-pct { color:#ff453a; }

.progress-track { height:6px; background:rgba(128,128,128,0.1); border-radius:3px; overflow:hidden; }
.progress-fill { height:100%; border-radius:3px; transition:width 0.4s; }
.usage-ok .progress-fill { background:#30d158; } .usage-warn .progress-fill { background:#ff9f0a; } .usage-over .progress-fill { background:#ff453a; }

.usage-meta { display:flex; justify-content:space-between; margin-top:6px; font-size:11px; color:v-bind('isDark ? "#636366" : "#9ca3af"'); }
.text-danger { color:#ff453a; }

/* ====== 删除 ====== */
.delete-row { margin-top:8px; display:flex; gap:8px; }
.del-btn { flex:1; padding:8px; border-radius:10px; border:1px solid rgba(255,69,58,0.25); cursor:pointer; font-size:12px; background:transparent; color:#ff453a; }

/* ====== 内联确认弹窗（Teleport到body）===== */
.ic-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.5); display:flex; align-items:center; justify-content:center; z-index:1100; padding:20px; }
.ic-dialog { width:280px; padding:24px; border-radius:16px; text-align:center; background:var(--card-bg); }
.ic-icon { font-size:28px; margin-bottom:8px; }
.ic-title { font-size:15px; font-weight:600; margin-bottom:6px; color:v-bind('isDark ? "#fff" : "#1f2937"'); }
.ic-msg { font-size:13px; margin-bottom:16px; line-height:1.5; color:v-bind('isDark ? "#8e8e93" : "#6b7280"'); }
.ic-actions { display:flex; gap:8px; }
.ic-cancel,.ic-ok,.ic-danger { flex:1; padding:8px; border-radius:8px; border:none; cursor:pointer; font-size:13px; }
.ic-cancel { background:var(--bg-input); color:v-bind('isDark ? "#8e8e93" : "#6b7280"'); }
.ic-ok { background:var(--accent-color); color:#fff; font-weight:600; }
.ic-danger { background:#ff453a; color:#fff; font-weight:600; }
</style>
