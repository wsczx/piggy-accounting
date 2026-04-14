<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('update:show', false)">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">标签管理</span>
        <button @click="$emit('update:show', false)" class="close-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 新建标签 -->
      <div class="tag-create-row">
        <input v-model="newTagName" type="text" placeholder="标签名称"
          class="form-input" @keydown.enter="handleCreateTag" />
        <button @click="handleCreateTag" class="btn-add">添加</button>
      </div>

      <!-- 颜色选择 -->
      <div class="color-row">
        <span class="color-label">颜色</span>
        <div class="color-options">
          <button v-for="c in presetColors" :key="c" @click="newTagColor = c"
            class="color-dot" :class="{ active: newTagColor === c }"
            :style="{ background: c }" />
        </div>
      </div>

      <!-- 标签列表 -->
      <div class="tag-list">
        <div v-if="tags.length === 0" class="tag-empty">暂无标签</div>
        <div v-for="tag in tags" :key="tag.id" class="tag-item">
          <div v-if="editingTag?.id === tag.id" class="tag-edit">
            <input v-model="editingTag.name" type="text" class="edit-input" />
            <div class="edit-colors">
              <button v-for="c in presetColors" :key="c" @click="editingTag.color = c"
                class="edit-dot" :class="{ active: editingTag.color === c }"
                :style="{ background: c, borderColor: editingTag.color === c ? (isDark ? '#fff' : '#1f2937') : 'transparent' }" />
            </div>
          </div>
          <div v-else class="tag-info">
            <span class="tag-color-dot" :style="{ background: tag.color }"></span>
            <span class="tag-name">{{ tag.name }}</span>
          </div>
          <div class="tag-actions">
            <template v-if="editingTag?.id === tag.id">
              <button @click="handleSaveTag" class="btn-sm btn-accent">保存</button>
              <button @click="cancelEditTag" class="btn-sm btn-ghost">取消</button>
            </template>
            <template v-else>
              <button @click="startEditTag(tag)" class="btn-sm btn-ghost">编辑</button>
              <button @click="handleDeleteTag(tag.id)" class="btn-sm btn-danger">删除</button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { GetAllTags, CreateTag, DeleteTag, UpdateTag } from '../../../wailsjs/go/service/TagService'

const props = defineProps<{
  show: boolean
  isDark: boolean
}>()
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'toast', msg: string, type: 'success' | 'error'): void
  (e: 'refresh'): void
}>()

const tags = ref<Array<{ id: number, name: string, color: string }>>([])
const newTagName = ref('')
const newTagColor = ref('#6366f1')
const editingTag = ref<{ id: number, name: string, color: string } | null>(null)

const presetColors = ['#6366f1', '#ec4899', '#f97316', '#10b981', '#06b6d4', '#f59e0b', '#ef4444', '#8b5cf6']

async function loadTags() {
  try {
    tags.value = await GetAllTags()
  } catch (e) {
    emit('toast', '加载标签失败: ' + e, 'error')
  }
}

async function handleCreateTag() {
  if (!newTagName.value.trim()) return
  try {
    await CreateTag(newTagName.value.trim(), newTagColor.value)
    newTagName.value = ''
    newTagColor.value = '#6366f1'
    await loadTags()
    emit('refresh')
  } catch (e) {
    emit('toast', '创建标签失败: ' + e, 'error')
  }
}

async function handleDeleteTag(id: number) {
  try {
    await DeleteTag(id)
    await loadTags()
    emit('refresh')
  } catch (e) {
    emit('toast', '删除失败: ' + e, 'error')
  }
}

function startEditTag(tag: { id: number, name: string, color: string }) {
  editingTag.value = { ...tag }
}

function cancelEditTag() {
  editingTag.value = null
}

async function handleSaveTag() {
  if (!editingTag.value) return
  try {
    await UpdateTag(editingTag.value.id, editingTag.value.name, editingTag.value.color)
    editingTag.value = null
    await loadTags()
    emit('refresh')
  } catch (e) {
    emit('toast', '更新失败: ' + e, 'error')
  }
}

// 弹窗打开时加载标签
import { watch } from 'vue'
watch(() => props.show, (val) => {
  if (val) loadTags()
})
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
  max-height: 80vh;
  overflow-y: auto;
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
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: background 0.15s;
}
.close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

/* 创建行 */
.tag-create-row {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.form-input {
  flex: 1;
  padding: 10px 12px;
  border-radius: 12px;
  font-size: 14px;
  border: none;
  outline: none;
  box-sizing: border-box;
  background: var(--bg-input);
  color: var(--text-primary);
}
.btn-add {
  padding: 10px 14px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  background: var(--accent-color);
  color: #fff;
  white-space: nowrap;
}

/* 颜色选择 */
.color-row {
  display: flex;
  gap: 6px;
  margin-bottom: 16px;
  align-items: center;
}
.color-label {
  font-size: 12px;
  margin-right: 4px;
  color: var(--text-secondary);
  white-space: nowrap;
}
.color-options {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.color-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.15s;
}
.color-dot.active {
  border-color: v-bind('isDark ? "#fff" : "#1f2937"');
}

/* 标签列表 */
.tag-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.tag-empty {
  text-align: center;
  padding: 20px;
  font-size: 13px;
  color: var(--text-secondary);
}
.tag-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--bg-input);
}
.tag-edit {
  display: flex;
  gap: 8px;
  flex: 1;
  align-items: center;
}
.edit-input {
  flex: 1;
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 13px;
  border: none;
  outline: none;
  background: var(--bg-input);
  color: var(--text-primary);
}
.edit-colors {
  display: flex;
  gap: 4px;
}
.edit-dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
}
.tag-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.tag-color-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.tag-name {
  font-size: 14px;
  color: var(--text-primary);
}
.tag-actions {
  display: flex;
  gap: 4px;
}
.btn-sm {
  padding: 4px 8px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 12px;
}
.btn-ghost {
  background: var(--bg-input);
  color: var(--text-secondary);
}
.btn-accent {
  background: var(--accent-color);
  color: #fff;
}
.btn-danger {
  background: rgba(255, 69, 58, 0.1);
  color: #ff453a;
}
</style>
