<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">安全设置</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>
      <div class="security-content">
        <div class="security-icon">🔒</div>
        <div class="security-title">安全设置开发中</div>
        <p class="security-desc">
          即将支持：<br/>• 应用锁（密码/指纹）<br/>• 数据加密存储<br/>• 自动锁定时间设置
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  isDark: boolean
}>()
defineEmits<{ (e: 'update:show', value: boolean): void }>()
</script>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px; }
.modal { width: 100%; max-width: 400px; border-radius: 20px; padding: 20px; background: var(--card-bg); }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.modal-title { font-size: 16px; font-weight: 600; color: var(--text-primary); }
.close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: var(--text-secondary); }
.security-content { text-align: center; padding: 40px 20px; }
.security-icon { font-size: 48px; margin-bottom: 16px; }
.security-title { font-size: 15px; font-weight: 600; margin-bottom: 8px; color: var(--text-primary); }
.security-desc { font-size: 13px; line-height: 1.6; color: var(--text-secondary); }
</style>
