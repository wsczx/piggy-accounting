<template>
  <transition name="welcome-fade">
    <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
      <div class="welcome-modal">
        <!-- Logo -->
        <div class="welcome-logo">🐷</div>
        <div class="welcome-title">欢迎使用猪猪记账！</div>

        <!-- 提示说明 -->
        <div class="welcome-body">
          <p class="welcome-desc">
            当前账本包含<strong>示例测试数据</strong>，方便你快速体验各项功能。
          </p>
          <p class="welcome-desc">
            你可以随时清空这些数据，开始记录自己的真实账目。
          </p>
        </div>

        <!-- 操作按钮 -->
        <div class="welcome-actions">
          <button class="btn-secondary" @click="handleKeep">
            保留数据，直接使用
          </button>
          <button class="btn-danger" :class="{ loading: clearing }" @click="handleClear" :disabled="clearing">
            {{ clearing ? '清空中...' : '清空测试数据' }}
          </button>
        </div>

        <!-- 底部提示 -->
        <p class="welcome-hint">清空后将保留系统类别和默认账户</p>
        <p class="welcome-hint">如后续想清空，可前往「我的 → 数据管理 → 清空数据」</p>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ClearLedgerData } from '../../../wailsjs/go/service/BackupService'
import { createLogger } from '../../utils/logger'

const log = createLogger('WelcomeModal')

const props = defineProps<{
  show: boolean
  isDark: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'refresh'): void
  (e: 'data-cleared'): void
}>()

const clearing = ref(false)

function handleKeep() {
  log.info('用户选择保留测试数据')
  emit('update:show', false)
}

async function handleClear() {
  if (clearing.value) return
  clearing.value = true
  try {
    await ClearLedgerData()
    log.info('测试数据已清空')
    emit('data-cleared')
    emit('refresh')
    emit('update:show', false)
  } catch (e) {
    log.error('清空测试数据失败', e)
    clearing.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 20px;
}

.welcome-modal {
  width: 100%;
  max-width: 380px;
  border-radius: 20px;
  padding: 28px 24px 22px;
  background: var(--card-bg);
  text-align: center;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 4px 16px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.welcome-logo {
  font-size: 52px;
  margin-bottom: 10px;
}

.welcome-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.welcome-body {
  margin-bottom: 22px;
}

.welcome-desc {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-secondary);
  margin: 0;
}

.welcome-desc strong {
  color: #f472b6;
  font-weight: 600;
}

.welcome-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 14px;
}

.btn-secondary,
.btn-danger {
  padding: 12px 20px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-secondary {
  background: var(--btn-secondary-bg, rgba(168, 85, 247, 0.08));
  color: var(--accent-color, #a855f7);
}

.btn-secondary:hover {
  background: var(--btn-secondary-hover, rgba(168, 85, 247, 0.15));
}

.btn-danger {
  background: linear-gradient(135deg, #ff453a, #ff6961);
  color: #fff;
}

.btn-danger:hover:not(:disabled) {
  opacity: 0.85;
}

.btn-danger.loading {
  opacity: 0.6;
  cursor: not-allowed;
}

.welcome-hint {
  font-size: 11px;
  color: var(--text-tertiary, #98989d);
  margin: 0;
}

/* 入场/退场动画 */
.welcome-fade-enter-active {
  animation: welcomeIn 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.welcome-fade-leave-active {
  animation: welcomeOut 0.2s ease-in;
}

@keyframes welcomeIn {
  from {
    opacity: 0;
    transform: scale(0.85);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes welcomeOut {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.92);
  }
}
</style>
