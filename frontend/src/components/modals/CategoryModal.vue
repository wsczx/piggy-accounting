<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">分类管理</span>
        <button @click="$emit('update:show', false)" class="close-btn"><svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg></button>
      </div>
      <div class="cat-section">
        <h3 class="section-title">添加自定义类别</h3>
        <div class="type-segment">
          <button type="button" @click="newCat.type = 'expense'" class="seg-btn" :class="{ active: newCat.type === 'expense' }">支出</button>
          <button type="button" @click="newCat.type = 'income'" class="seg-btn" :class="{ active: newCat.type === 'income' }">收入</button>
        </div>
        <div class="cat-input-row">
          <input v-model="newCat.icon" type="text" maxlength="2" class="input-icon" placeholder="图标"/>
          <input v-model="newCat.name" type="text" class="input-name" placeholder="类别名称"/>
          <button @click="handleAddCategory" :disabled="!newCat.name||!newCat.icon" class="btn-add-cat" :style="{opacity:(!newCat.name||!newCat.icon)?0.3:1,cursor:(!newCat.name||!newCat.icon)?'not-allowed':'pointer'}">添加</button>
        </div>
      </div>
      <!-- 支出类别 -->
      <div class="cat-section">
        <h3 class="section-title">支出类别</h3>
        <div class="cat-grid">
          <div v-for="cat in expenseCategories" :key="cat.id" class="cat-item">
            <span class="cat-item-icon">{{ cat.icon }}</span>
            <span class="cat-item-name">{{ cat.name }}</span>
            <span v-if="cat.is_system" class="badge-system">系统</span>
            <button v-else @click="handleDeleteCategory(cat.id!,cat.name)" class="btn-delete-cat"><svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path d="M6 18L18 6M6 6l12 12"/></svg></button>
          </div>
        </div>
      </div>
      <!-- 收入类别 -->
      <div class="cat-section">
        <h3 class="section-title">收入类别</h3>
        <div class="cat-grid">
          <div v-for="cat in incomeCategories" :key="cat.id" class="cat-item">
            <span class="cat-item-icon">{{ cat.icon }}</span>
            <span class="cat-item-name">{{ cat.name }}</span>
            <span v-if="cat.is_system" class="badge-system">系统</span>
            <button v-else @click="handleDeleteCategory(cat.id!,cat.name)" class="btn-delete-cat"><svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path d="M6 18L18 6M6 6l12 12"/></svg></button>
          </div>
        </div>
      </div>
      <ConfirmModal v-model="showDeleteConfirm" icon="⚠️" title="删除类别" :message="`确定要删除类别「${deleteCatName}」吗？`" dangerous @confirm="confirmDeleteCategory"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import ConfirmModal from '../ConfirmModal.vue'

export interface CategoryItem {
  id?: number
  name: string
  icon: string
  type: string
  is_system?: boolean
}

const props = defineProps<{
  show: boolean; isDark: boolean
  expenseCategories: CategoryItem[]; incomeCategories: CategoryItem[]
  store: {
    addCategory?(name: string, icon: string, type: string): Promise<void>
    deleteCategory?(id: number): Promise<void>
    loadRecords?(): Promise<void>
    init?(): Promise<void>
  }
}>()
const emit = defineEmits<{ (e: 'update:show', value: boolean): void; (e: 'toast', msg: string, type: 'success'|'error'): void; (e: 'refresh'): void }>()

const newCat = reactive({ name:'', icon:'', type:'expense' as 'income'|'expense' })
const showDeleteConfirm = ref(false)
const deleteCatId = ref<number|null>(null)
const deleteCatName = ref('')

async function handleAddCategory() {
  if (!newCat.name || !newCat.icon) return
  try {
    await props.store.addCategory!(newCat.name, newCat.icon, newCat.type)
    newCat.name=''; newCat.icon=''
    emit('refresh')
  } catch(e:any){ emit('toast','添加失败: '+e,'error') }
}
function handleDeleteCategory(id:number, name:string){ deleteCatId.value=id; deleteCatName.value=name; showDeleteConfirm.value=true }
async function confirmDeleteCategory(){
  if(deleteCatId.value===null)return
  try{ await props.store.deleteCategory!(deleteCatId.value); emit('refresh') }
  catch(e:any){ emit('toast','删除失败: '+e,'error') }
}
</script>

<style scoped>
.modal-overlay{position:fixed;inset:0;background:rgba(0,0,0,0.5);display:flex;align-items:center;justify-content:center;z-index:100;padding:20px}
.modal{width:100%;max-width:400px;border-radius:20px;padding:20px;background:var(--card-bg);max-height:80vh;overflow-y:auto;}
.modal-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}
.modal-title{font-size:16px;font-weight:600;color:var(--text-primary)}
.close-btn{padding:4px;border:none;background:transparent;cursor:pointer;border-radius:8px;color:var(--text-secondary);transition:background .15s}
.close-btn:hover{background:rgba(128,128,128,.15)}

.cat-section{border-radius:14px;padding:16px;margin-bottom:12px;background:var(--bg-input)}
.section-title{font-size:13px;font-weight:600;margin-bottom:10px;color:var(--text-secondary)}
.type-segment{display:flex;gap:4px;padding:2px;border-radius:10px;margin-bottom:10px;background:var(--chip-inactive-bg)}
.seg-btn{flex:1;padding:6px 0;border-radius:8px;border:none;cursor:pointer;font-size:12px;font-weight:600;transition:all .2s;background:transparent;color:var(--text-secondary)}
.seg-btn.active{color:#fff}
.seg-btn.active[data-type-expense],:scope > .seg-btn:nth-child(1).active{background:var(--expense-color)}
.cat-section .type-segment .seg-btn:nth-child(1).active{background:var(--expense-color)}
.cat-section .type-segment .seg-btn:nth-child(2).active{background:var(--income-color)}
.cat-input-row{display:flex;gap:8px}
.input-icon{width:56px;padding:8px 10px;border-radius:10px;font-size:13px;text-align:center;border:none;background:var(--card-bg);color:var(--text-primary);box-sizing:border-box}
.input-name{flex:1;padding:8px 10px;border-radius:10px;font-size:13px;border:none;background:var(--card-bg);color:var(--text-primary);box-sizing:border-box}
.btn-add-cat{padding:8px 16px;border-radius:10px;border:none;cursor:pointer;font-size:12px;font-weight:600;color:#fff;background:var(--accent-color)}

.cat-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:8px}
.cat-item{display:flex;flex-direction:column;align-items:center;gap:5px;padding:10px 0;border-radius:14px;position:relative;background:var(--card-bg);transition:background .15s}
.cat-item:hover{background:rgba(99,102,241,.04)!important}
.cat-item-icon{font-size:20px}
.cat-item-name{font-size:12px;font-weight:500;color:var(--text-secondary)}
.badge-system{font-size:9px;padding:2px 6px;border-radius:8px;background:var(--chip-inactive-bg);color:var(--text-muted)}
.btn-delete-cat{position:absolute;top:-3px;right:-3px;width:18px;height:18px;border-radius:50%;display:flex;align-items:center;justify-content:center;border:none;cursor:pointer;background:var(--danger-bg);color:var(--danger-color);opacity:0;transition:opacity .15s}
.cat-item:hover .btn-delete-cat{opacity:1}
</style>
