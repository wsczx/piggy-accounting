<template>
  <Teleport to="body">
    <div v-if="modelValue" class="tf-overlay" @click.self="$emit('update:modelValue', false)">
      <div class="tf-dialog" :style="isDark ? {background:'#1c1c1e'} : {background:'#fff'}">
        <div class="tf-header">
          <span :style="{color: isDark ? '#fff' : '#1f2937'}">{{ editing ? '编辑任务' : '添加任务' }}</span>
          <button @click="$emit('update:modelValue', false)" class="tf-close-btn">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <div v-if="validationError" class="tf-error">{{ validationError }}</div>

        <div class="tf-body">
          <div class="form-group">
            <label :style="{color: isDark ? '#8e8e93' : '#6b7280'}">任务名称</label>
            <input v-model="title" type="text" placeholder="例如：还信用卡"
                   :style="{color: isDark ? '#fff' : '#1f2937', background: isDark ? '#2c2c2e' : '#f3f4f6'}" />
          </div>

          <div class="form-group">
            <label :style="{color: isDark ? '#8e8e93' : '#6b7280'}">到期日期</label>
            <input v-model="dueDate" type="date"
                   :style="{color: isDark ? '#fff' : '#1f2937', background: isDark ? '#2c2c2e' : '#f3f4f6'}" />
          </div>

          <div class="form-group">
            <label :style="{color: isDark ? '#8e8e93' : '#6b7280'}">金额（可选）</label>
            <input v-model.number="amount" type="number" placeholder="0.00"
                   :style="{color: isDark ? '#fff' : '#1f2937', background: isDark ? '#2c2c2e' : '#f3f4f6'}" />
          </div>
        </div>

        <div class="tf-actions">
          <button @click="$emit('update:modelValue', false)" class="tf-cancel">取消</button>
          <button @click="handleSave" class="tf-ok">{{ editing ? '保存' : '确定' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useThemeStore } from '../stores/theme'

const props = defineProps<{
  modelValue: boolean
  /** 编辑模式时传入已有数据 */
  initialTitle?: string
  initialDueDate?: string
  initialAmount?: number | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'save', data: { title: string; dueDate: string; amount: number }): void
  (e: 'toast', msg: string, type: 'error'): void
}>()

const theme = useThemeStore()
const isDark = computed(() => theme.isDark)

/** 表单验证错误提示 */
const validationError = ref('')

const title = ref('')
const dueDate = ref('')
const amount = ref<number | undefined>(undefined)

const editing = computed(() => !!props.initialTitle)

watch(() => props.modelValue, (v) => {
  if (v) {
    title.value = props.initialTitle || ''
    dueDate.value = props.initialDueDate || ''
    amount.value = props.initialAmount ?? undefined
  }
})

function handleSave() {
  validationError.value = ''
  if (!title.value.trim() || !dueDate.value) {
    validationError.value = '请填写任务名称和到期日期'
    emit('toast', validationError.value, 'error')
    return
  }
  emit('save', {
    title: title.value,
    dueDate: dueDate.value,
    amount: amount.value || 0,
  })
}
</script>

<style scoped>
.tf-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.tf-dialog {
  width: 100%;
  max-width: 360px;
  border-radius: 20px;
  padding: 20px;
}
.tf-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.tf-header span {
  font-size: 16px;
  font-weight: 600;
}
.tf-close-btn {
  padding: 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 8px;
  color: #9ca3af;
}
.tf-close-btn:hover { background: rgba(0,0,0,0.05); }

.tf-error {
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  color: #ff453a;
  background: rgba(255,69,58,0.1);
}

.form-group { margin-bottom: 14px; }
.form-group label {
  display: block;
  font-size: 13px;
  margin-bottom: 6px;
}
.form-group input {
  width: 100%;
  padding: 12px 14px;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.tf-actions {
  display: flex;
  gap: 10px;
  margin-top: 4px;
}
.tf-actions button {
  flex: 1;
  padding: 12px;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.15s;
}
.tf-actions button:hover { opacity: 0.9; }
.tf-cancel {
  background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"');
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}
.tf-ok {
  background: var(--accent-color);
  color: #fff;
}
</style>
