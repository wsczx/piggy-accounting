<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">帮助与反馈</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="hf-body">
        <!-- 常见问题 -->
        <div class="hf-section">
          <h3 class="section-label">常见问题</h3>
          <div class="faq-list">
            <div v-for="(item, idx) in faqItems" :key="idx"
              class="faq-item" :class="{ expanded: expandedFaq === idx }"
              @click="toggleFaq(idx)">
              <div class="faq-q">
                {{ item.q }}
                <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
                  class="faq-arrow" :class="{ rotated: expandedFaq === idx }">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
              <div v-if="expandedFaq === idx" class="faq-a">{{ item.a }}</div>
            </div>
          </div>
        </div>
        <!-- 反馈表单 -->
        <div class="hf-section">
          <h3 class="section-label">提交反馈</h3>
          <textarea v-model="feedbackContent" placeholder="请描述您遇到的问题或建议..."
            class="feedback-textarea"></textarea>
          <button @click="submitFeedback" class="btn-submit">提交反馈</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  isDark: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const faqItems = [
  { q: '如何导出我的记账数据？', a: '进入"我的"页面，点击"数据管理"，选择"数据导出"标签，选择日期范围后即可导出CSV格式的数据文件。' },
  { q: '如何设置预算提醒？', a: '进入"我的"页面，点击"提醒设置"，开启预算预警并设置提醒阈值（如80%），当支出达到该比例时会自动提醒。' },
  { q: '数据会丢失吗？', a: '应用支持自动备份和手动备份，您可以在"数据管理"中创建备份或导出数据。建议定期将备份文件保存到安全位置。' },
  { q: '如何删除错误的记录？', a: '在首页或全部账单页面找到该记录，点击即可进入编辑模式，选择删除即可。' },
]

const expandedFaq = ref<number | null>(null)
const feedbackContent = ref('')

watch(() => props.show, (val) => {
  if (!val) expandedFaq.value = null
})

function toggleFaq(idx: number) {
  expandedFaq.value = expandedFaq.value === idx ? null : idx
}

function submitFeedback() {
  if (!feedbackContent.value.trim()) {
    emit('toast', '请输入反馈内容', 'error')
    return
  }
  // 用 mailto 打开默认邮件客户端发送到 wsc@wsczx.com
  const subject = encodeURIComponent('猪猪记账 - 用户反馈')
  const body = encodeURIComponent(feedbackContent.value.trim() + '\n\n---\n来自猪猪记账桌面客户端')
  window.location.href = `mailto:wsc@wsczx.com?subject=${subject}&body=${body}`
  emit('toast', '正在打开邮件客户端...', 'success')
  setTimeout(() => { emit('update:show', false) }, 800)
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
  z-index: 100;
  padding: 20px;
}
.modal {
  width: 100%;
  max-width: 400px;
  border-radius: 20px;
  padding: 20px;
  background: var(--card-bg);
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.close-btn {
  padding: 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 8px;
  color: var(--text-secondary);
}
.hf-body { padding: 8px 0; }
.hf-section { margin-bottom: 20px; }
.section-label { font-size: 13px; font-weight: 600; margin-bottom: 12px; color: var(--text-secondary); }

.faq-list { display: flex; flex-direction: column; gap: 10px; }
.faq-item {
  padding: 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s;
  background: var(--bg-input);
}
.faq-q {
  font-size: 14px;
  font-weight: 500;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text-primary);
}
.faq-arrow { transition: transform 0.2s; flex-shrink: 0; margin-left: 8px; }
.faq-arrow.rotated { transform: rotate(180deg); }
.faq-a {
  font-size: 13px;
  margin-top: 8px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.feedback-textarea {
  width: 100%;
  min-height: 100px;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  resize: vertical;
  outline: none;
  box-sizing: border-box;
  background: var(--bg-input);
  color: var(--text-primary);
}
.btn-submit {
  width: 100%;
  margin-top: 12px;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  background: var(--accent-color);
  color: #fff;
}
</style>
