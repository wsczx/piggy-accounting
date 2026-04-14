<template>
  <Teleport to="body">
    <div v-if="modelValue" class="cm-overlay" @click.self="$emit('update:modelValue', false)">
      <div class="cm-dialog" :style="isDark ? {background:'#1c1c1e'} : {background:'#fff'}">
        <div class="cm-icon">{{ icon }}</div>
        <div class="cm-title" :style="{color: isDark ? '#fff' : '#1f2937'}">{{ title }}</div>
        <div class="cm-message" :style="{color: isDark ? '#8e8e93' : '#6b7280'}">
          <slot>{{ message }}</slot>
        </div>
        <div class="cm-actions">
          <button @click="$emit('update:modelValue', false)" class="cm-cancel">取消</button>
          <button @click="$emit('confirm')" :class="confirmClass">{{ confirmText || '确定' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeStore } from '../stores/theme'

const props = withDefaults(defineProps<{
  modelValue: boolean
  icon?: string
  title?: string
  message?: string
  confirmText?: string
  dangerous?: boolean
}>(), {
  icon: '🗑️',
  title: '确认删除',
  message: '',
  confirmText: '确定',
  dangerous: true,
})

defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
}>()

const theme = useThemeStore()
const isDark = computed(() => theme.isDark)

const confirmClass = computed(() => props.dangerous ? 'cm-danger' : 'cm-ok')
</script>

<style scoped>
.cm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.cm-dialog {
  width: 100%;
  max-width: 320px;
  border-radius: 20px;
  padding: 28px 24px;
  text-align: center;
}
.cm-icon { font-size: 48px; margin-bottom: 16px; }
.cm-title { font-size: 18px; font-weight: 600; margin-bottom: 8px; }
.cm-message { font-size: 14px; line-height: 1.5; margin-bottom: 24px; }
.cm-actions {
  display: flex;
  gap: 12px;
}
.cm-actions button {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.15s;
}
.cm-actions button:hover { opacity: 0.9; }
.cm-cancel {
  background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"');
  color: v-bind('isDark ? "#8e8e93" : "#6b7280"');
}
.cm-ok {
  background: var(--accent-color);
  color: #fff;
}
.cm-danger {
  background: #ff453a;
  color: #fff;
}
</style>
