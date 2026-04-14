<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="about-logo">🐷</div>
      <div class="about-name">猪猪记账</div>
      <div class="about-version">版本 v1.0.0</div>
      <p class="about-desc">一款简洁好用的个人记账工具<br/>轻松记录每一笔收支，掌控财务状况</p>
      <button @click="$emit('update:show', false)" class="btn-gotit">知道了</button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  isDark: boolean
}>()

defineEmits<{
  (e: 'update:show', value: boolean): void
}>()
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}
.modal {
  width: 100%;
  max-width: 400px;
  border-radius: 20px;
  padding: 32px 24px;
  background: var(--card-bg);
  text-align: center;
}
.about-logo {
  font-size: 56px;
  margin-bottom: 12px;
}
.about-name {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 4px;
  color: var(--text-primary);
}
.about-version {
  font-size: 13px;
  margin-bottom: 20px;
  color: var(--text-secondary);
}
.about-desc {
  font-size: 13px;
  line-height: 1.7;
  margin-bottom: 24px;
  color: var(--text-secondary);
}
.btn-gotit {
  padding: 12px 32px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  background: var(--accent-color);
  color: #fff;
}
</style>
