<template>
  <div class="qa-inline" :class="{ 'is-dark': isDark }">
    <!-- 紧凑快捷按钮行 -->
    <div class="qa-row">
      <button v-for="action in quickActions" :key="action.id"
              class="qa-chip"
              @click="handleAction(action)"
              :title="action.label"
              :style="{ '--chip-color': action.color }">
        <span class="chip-icon">{{ action.icon }}</span>
      </button>
      <button class="qa-chip chip-add" @click="startAddQuick" title="添加快捷">＋</button>
    </div>

    <!-- 编辑弹窗 -->
    <Teleport to="body">
      <div v-if="showEditModal" class="qa-overlay" @click.self="showEditModal = false">
        <div class="qa-edit-dialog" :style="isDark ? { background: '#1c1c1e' } : { background: '#fff' }">
          <div class="modal-header">
            <span :style="{ color: isDark ? '#fff' : '#1f2937', fontSize: '16px', fontWeight: 600 }">
              {{ editingIndex >= 0 ? '编辑快捷' : '添加快捷' }}
            </span>
            <button @click="showEditModal = false" class="close-btn">✕</button>
          </div>

          <!-- 已有列表 -->
          <div v-if="quickActions.length > 0" style="margin-bottom:16px;">
            <div class="section-label">当前快捷（点击可编辑）</div>
            <div class="edit-list">
              <div v-for="(a, i) in quickActions" :key="a.id" class="edit-row"
                :style="isDark ? { background: '#2c2c2e' } : { background: '#f3f4f6' }"
                @click="startEdit(i)">
                <span style="font-size:18px;">{{ a.icon }}</span>
                <span class="edit-name" :style="{ color: isDark ? '#fff' : '#1f2937' }">{{ a.label }}</span>
                <span class="edit-target" :style="{ color: isDark ? '#8e8e93' : '#9ca3af' }">
                  {{ getTargetLabel(a.action, a.route) }}
                </span>
                <button @click.stop="removeItem(i)" class="row-btn danger">删</button>
              </div>
            </div>
          </div>

          <!-- 表单：名称 -->
          <div class="form-group">
            <label>名称</label>
            <input v-model="form.label" type="text" placeholder="例如：转账" :style="{ color: isDark ? '#fff' : '#1f2937', background: isDark ? '#2c2c2e' : '#f3f4f6' }" />
          </div>

          <!-- 图标 -->
          <div class="form-group">
            <label>图标</label>
            <div class="icon-grid">
              <button v-for="ico in commonIcons" :key="ico" @click="form.icon = ico"
                      class="icon-btn" :class="{ active: form.icon === ico }"
                      :style="getIconBtnStyle(ico)">{{ ico }}</button>
            </div>
          </div>

          <!-- 颜色 -->
          <div class="form-group">
            <label>颜色</label>
            <div class="color-row">
              <button v-for="c in colors" :key="c" @click="form.color = c" class="color-dot"
                      :class="{ active: form.color === c }"
                      :style="{ background: c }"></button>
            </div>
          </div>

          <!-- 目标功能选择 -->
          <div class="form-group">
            <label>目标功能</label>
            <div class="target-grid">
              <button v-for="t in targetOptions" :key="t.action + (t.route || '')"
                      @click="selectTarget(t)"
                      class="target-btn"
                      :class="{ active: form.action === t.action && (t.route || '/') === (form.route || '/') }"
                      :style="getTargetBtnStyle(t)">
                <span class="target-ico">{{ t.icon }}</span>
                <span class="target-name">{{ t.label }}</span>
              </button>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="modal-actions">
            <button @click="resetDefaults" class="btn-secondary">恢复默认</button>
            <button @click="saveForm" class="btn-primary">{{ editingIndex >= 0 ? '保存修改' : '添加' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '../stores/theme'

export interface QuickAction {
  id: string
  label: string
  icon: string
  color: string
  /** 目标动作标识 */
  action: string
  /** 路由路径 */
  route?: string
}

/** 系统内置的可选目标功能 */
const TARGETS: Array<{ action: string; label: string; icon: string; route?: string; desc?: string }> = [
  // ===== 核心操作 =====
  { action: 'record',    label: '记一笔',       icon: '📝', desc: '打开记账面板' },

  // ===== 页面导航 =====
  { action: 'navigate',  label: '统计',         icon: '📊', route: '/statistics', desc: '收支统计页' },
  { action: 'popup',     label: '分类管理',     icon: '🏷️', route: '/profile?open=category', desc: '类别管理弹窗' },
  { action: 'navigate',  label: '全部账单',     icon: '📋', route: '/records',     desc: '所有记录列表' },
  { action: 'navigate',  label: '待办任务',     icon: '✅', route: '/tasks',        desc: '任务管理页' },
  { action: 'navigate',  label: '更多功能',     icon: '⚙️', route: '/profile',      desc: '账户/周期/导入导出等' },

  // ===== 弹窗功能（跳转"更多"页并携带参数自动打开对应弹窗）=====
  { action: 'popup',     label: '账户管理',     icon: '🏦', route: '/profile?open=account',   desc: '管理收支账户' },
  { action: 'popup',     label: '转账记录',     icon: '🔄', route: '/profile?open=transfer',  desc: '账户间转账' },
  { action: 'popup',     label: '周期记账',     icon: '⏰', route: '/profile?open=recurring', desc: '定期自动记账' },
  { action: 'popup',     label: '数据管理',     icon: '💾', route: '/profile?open=backup',    desc: '备份/导入/导出' },
  { action: 'popup',     label: '多账本',       icon: '📒', route: '/profile?open=ledger',    desc: '多账本切换' },
  { action: 'popup',     label: '提醒设置',     icon: '🔔', route: '/profile?open=reminder',  desc: '预算预警提醒' },
]

const STORAGE_KEY = 'piggy-quick-actions'

const DEFAULTS: QuickAction[] = [
  { id: 'default-record', label: '记一笔', icon: '📝', color: '#6366f1', action: 'record' },
  { id: 'default-stats', label: '统计', icon: '📊', color: '#f59e0b', action: 'navigate', route: '/statistics' },
  { id: 'default-cat', label: '分类', icon: '🏷️', color: '#ec4899', action: 'popup', route: '/profile?open=category' },
  { id: 'default-more', label: '更多', icon: '⚙️', color: '#6b7280', action: 'navigate', route: '/profile' },
]

const emit = defineEmits<{
  (e: 'action-click', action: QuickAction): void
}>()

const router = useRouter()
const theme = useThemeStore()
const isDark = computed(() => theme.isDark)

// ---- 数据持久化 ----
function load(): QuickAction[] {
  try {
    const s = localStorage.getItem(STORAGE_KEY)
    if (s) return JSON.parse(s)
  } catch { /* ignore */ }
  return [...DEFAULTS]
}
function save() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(quickActions.value))
}

const quickActions = ref<QuickAction[]>(load())

// ---- 弹窗状态 ----
const showEditModal = ref(false)
const editingIndex = ref(-1)
const form = ref<QuickAction>({
  id: '', label: '', icon: '📌', color: '#6366f1',
  action: 'navigate', route: '/statistics',
})

const commonIcons = ['📝','📊','💰','🏷️','🔄','💳','📅','⚙️','🔍','📈','🎯','⭐','💡','❤️','🔥','🌟','✨','🛒','🚗','🍜','☕','🎮','📱','💻','📋','🏦','🎁','🏠']
const colors = ['#6366f1','#8b5cf6','#ec4899','#ef4444','#f59e0b','#10b981','#06b6d4','#3b82f6']

/** 可选目标列表（用于表单展示） */
const targetOptions = computed(() => TARGETS)

// ---- 工具函数 ----
function adjustColor(hex: string, amount: number): string {
  const h = hex.replace('#', '')
  const r = Math.min(255, parseInt(h.substring(0, 2), 16) + amount)
  const g = Math.min(255, parseInt(h.substring(2, 4), 16) + amount)
  const b = Math.min(255, parseInt(h.substring(4, 6), 16) + amount)
  return `#${r.toString(16).padStart(2,'0')}${g.toString(16).padStart(2,'0')}${b.toString(16).padStart(2,'0')}`
}

function getIconBtnStyle(ico: string) {
  if (form.value.icon !== ico) return isDark.value
    ? { background: '#2c2c2e', borderColor: 'transparent' }
    : { background: '#f3f4f6', borderColor: 'transparent' }
  return isDark.value
    ? { background: form.value.color, borderColor: form.value.color }
    : { background: form.value.color + '20', borderColor: form.value.color }
}

function getTargetBtnStyle(t: { action: string; route?: string }) {
  const active = form.value.action === t.action && (t.route || '/') === (form.value.route || '/')
  if (!active) return isDark.value
    ? { background: '#2c2c2e', borderColor: 'transparent' }
    : { background: '#f3f4f6', borderColor: 'transparent' }
  return isDark.value
    ? { background: 'rgba(99,102,241,0.15)', border: '1px solid #6366f1', color: '#a5b4fc' }
    : { background: 'rgba(99,102,241,0.08)', border: '1px solid #6366f1', color: '#4f46e5' }
}

/** 根据 action+route 获取可读标签 */
function getTargetLabel(action: string, route?: string): string {
  const t = TARGETS.find(x => x.action === action && (x.route || '') === (route || ''))
  if (t) return t.label
  if (action === 'record') return '记一笔'
  if (action === 'navigate') return route || '页面'
  if (action === 'popup') return route?.replace('/profile?open=','') || '弹窗'
  return '自定义'
}

// ---- 操作 ----
function handleAction(action: QuickAction) {
  emit('action-click', action)
}

function startAddQuick() {
  editingIndex.value = -1
  form.value = {
    id: Date.now().toString(), label: '', icon: '📌', color: '#6366f1',
    action: 'navigate', route: '/statistics',
  }
  showEditModal.value = true
}
function startEdit(i: number) {
  editingIndex.value = i
  form.value = { ...quickActions.value[i] }
  showEditModal.value = true
}

/** 选择目标功能 */
function selectTarget(t: typeof TARGETS[number]) {
  form.value.action = t.action
  form.value.route = t.route
  // 自动填充名称（仅当用户还没手动输入时）
  if (!form.value.label.trim()) {
    form.value.label = t.label
  }
  // 自动填充图标（仅当还是默认图标时）
  if (form.value.icon === '📌') {
    form.value.icon = t.icon
  }
}

function saveForm() {
  if (!form.value.label.trim()) return
  const item: QuickAction = { ...form.value }
  if (editingIndex.value >= 0) {
    quickActions.value[editingIndex.value] = item
  } else {
    quickActions.value.push(item)
  }
  save()
  showEditModal.value = false
}
function removeItem(i: number) {
  quickActions.value.splice(i, 1)
  save()
}
function resetDefaults() {
  quickActions.value = [...DEFAULTS]
  save()
  showEditModal.value = false
}

defineExpose({ quickActions })
</script>

<style scoped>
.qa-inline { display: inline-flex; }

/* 紧凑按钮行 */
.qa-row {
  display: flex;
  align-items: center;
  gap: 5px;
}

.qa-chip {
  width: 32px; height: 32px; border-radius: 10px;
  border: none; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  font-size: 15px;
  background: var(--surface-secondary, #f3f4f6);
  transition: transform 0.15s, box-shadow 0.15s;
  position: relative;
}
.qa-chip:hover {
  transform: scale(1.12);
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.qa-chip:active { transform: scale(0.95); }

.chip-icon {
  filter: grayscale(0.15);
}

.chip-add {
  font-size: 18px;
  font-weight: 300;
  color: #9ca3af;
  min-width: 32px;
}
.is-dark .chip-add { color: #636366; }
.is-dark .qa-chip { background: rgba(255,255,255,0.08); }

/* ====== 编辑弹窗（保持不变）====== */
.qa-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}
.qa-edit-dialog {
  width: 100%; max-width: 400px; border-radius: 20px;
  padding: 20px; max-height: 80vh; overflow-y: auto;
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 16px;
}
.close-btn {
  padding: 4px 8px; border: none; background: transparent;
  cursor: pointer; border-radius: 6px; font-size: 14px; color: #9ca3af;
}

.section-label {
  font-size: 13px; font-weight: 500; margin-bottom: 8px; color: #8e8e93;
}

.edit-list {
  display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px;
}
.edit-row {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 12px; border-radius: 10px;
  cursor: pointer; transition: background 0.15s;
}
.edit-row:hover { opacity: 0.85; }

.edit-name { flex: 1; font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.edit-target { font-size: 11px; flex-shrink: 0; }

.row-btn {
  padding: 4px 10px; border-radius: 8px; border: none;
  font-size: 11px; cursor: pointer;
  background: v-bind('isDark ? "#374151" : "#e5e7eb"');
  color: v-bind('isDark ? "#d1d5db" : "#4b5563"');
  transition: opacity 0.15s;
}
.row-btn:hover { opacity: 0.7; }
.row-btn.danger { background: rgba(239,68,68,0.1); color: #ef4444; }

.form-group { margin-bottom: 14px; }
.form-group > label {
  display: block; font-size: 13px; margin-bottom: 6px;
  color: var(--text-secondary, #9ca3af);
}
.form-group input {
  width: 100%; padding: 12px 14px; border-radius: 12px;
  border: none; font-size: 14px; outline: none;
}

/* 图标网格 */
.icon-grid { display: grid; grid-template-columns: repeat(8, 1fr); gap: 6px; }
.icon-btn {
  width: 36px; height: 36px; border-radius: 8px;
  border: 2px solid transparent; font-size: 18px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.icon-btn.active { transform: scale(1.1); }

/* 颜色 */
.color-row { display: flex; gap: 8px; flex-wrap: wrap; }
.color-dot {
  width: 28px; height: 28px; border-radius: 50%;
  border: 2px solid transparent; cursor: pointer; transition: transform 0.15s;
}
.color-dot.active { transform: scale(1.2); box-shadow: 0 0 0 2px #fff, 0 0 0 4px currentColor; }

/* 目标功能选择器 */
.target-grid {
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px;
}
.target-btn {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 12px; border-radius: 12px;
  border: 1px solid transparent; cursor: pointer;
  transition: all 0.15s; text-align: left;
}
.target-btn .target-ico { font-size: 18px; flex-shrink: 0; }
.target-btn .target-name { font-size: 13px; font-weight: 500; }

.modal-actions {
  display: flex; gap: 12px; padding-top: 16px;
  border-top: 1px solid v-bind('isDark ? "rgba(255,255,255,0.1)" : "#f3f4f6"');
}
.btn-primary, .btn-secondary {
  flex: 1; padding: 12px; border-radius: 12px;
  border: none; font-size: 14px; font-weight: 500;
  cursor: pointer; transition: opacity 0.15s;
}
.btn-primary { background: var(--accent-color); color: #fff; }
.btn-secondary { background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"'); color: v-bind('isDark ? "#8e8e93" : "#6b7280"'); }
.btn-primary:hover, .btn-secondary:hover { opacity: 0.9; }
</style>
