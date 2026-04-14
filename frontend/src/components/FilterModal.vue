<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fm-overlay" @click.self="$emit('update:modelValue', false)">
      <div class="fm-dialog" :style="isDark ? {background:'#1c1c1e'} : {background:'#fff'}">
        <div class="fm-header">
          <span :style="{color: isDark ? '#fff' : '#1f2937', fontSize:'16px', fontWeight:600}">筛选账单</span>
          <button @click="$emit('update:modelValue', false)" class="fm-close-btn">
            <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- 类型筛选 -->
        <div class="fm-section">
          <div class="fm-section-label">类型</div>
          <div class="fm-options">
            <button v-for="t in typeOptions" :key="t.value"
                    @click="selectedType = t.value"
                    class="fm-option"
                    :class="{ active: selectedType === t.value }"
                    :style="selectedType === t.value
                      ? {background:'#6366f1', color:'#fff'}
                      : (isDark ? {background:'#2c2c2e', color:'#8e8e93'} : {background:'#f3f4f6', color:'#6b7280'})">
              {{ t.label }}
            </button>
          </div>
        </div>

        <!-- 分类筛选 -->
        <div v-if="categories.length > 0" class="fm-section">
          <div class="fm-section-label">分类</div>
          <div class="fm-options">
            <button v-for="cat in categories" :key="cat.id"
                    @click="toggleCategory(cat.id)"
                    class="fm-option"
                    :class="{ active: selectedCategories.includes(cat.id) }"
                    :style="selectedCategories.includes(cat.id)
                      ? {background:'#6366f1', color:'#fff'}
                      : (isDark ? {background:'#2c2c2e', color:'#8e8e93'} : {background:'#f3f4f6', color:'#6b7280'})">
              {{ cat.icon }} {{ cat.name }}
            </button>
          </div>
        </div>

        <div class="fm-actions">
          <button @click="handleReset" class="fm-reset">重置</button>
          <button @click="handleApply" class="fm-apply">确定</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useThemeStore } from '../stores/theme'

export interface FilterResult {
  type: 'all' | 'income' | 'expense'
  categories: number[]
}

const props = withDefaults(defineProps<{
  modelValue: boolean
  categories?: Array<{ id: number; name: string; icon: string }>
  initialType?: 'all' | 'income' | 'expense'
  initialCategories?: number[]
}>(), {
  categories: () => [],
  initialType: 'all',
  initialCategories: () => [],
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'apply', result: FilterResult): void
  (e: 'reset'): void
}>()

const theme = useThemeStore()
const isDark = computed(() => theme.isDark)

const typeOptions = [
  { label: '全部', value: 'all' as const },
  { label: '收入', value: 'income' as const },
  { label: '支出', value: 'expense' as const },
]

const selectedType = ref<'all' | 'income' | 'expense'>(props.initialType)
const selectedCategories = ref<number[]>([...props.initialCategories])

watch(() => props.modelValue, (v) => {
  if (v) {
    selectedType.value = props.initialType
    selectedCategories.value = [...props.initialCategories]
  }
})

function toggleCategory(id: number) {
  const idx = selectedCategories.value.indexOf(id)
  if (idx > -1) selectedCategories.value.splice(idx, 1)
  else selectedCategories.value.push(id)
}

function handleReset() { emit('reset') }
function handleApply() {
  emit('apply', { type: selectedType.value, categories: selectedCategories.value })
  emit('update:modelValue', false)
}
</script>

<style scoped>
.fm-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}
.fm-dialog {
  width: 100%; max-width: 400px; border-radius: 20px; padding: 20px;
  max-height: 80vh; overflow-y: auto;
}
.fm-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;
}
.fm-close-btn { padding: 4px; border: none; background: transparent; cursor: pointer; border-radius: 8px; color: #9ca3af; }
.fm-close-btn:hover { background: rgba(0,0,0,0.05); }

.fm-section { margin-bottom: 20px; }
.fm-section-label { font-size: 13px; font-weight: 500; margin-bottom: 10px; color: #8e8e93; }
.fm-options { display: flex; flex-wrap: wrap; gap: 8px; }
.fm-option {
  padding: 8px 14px; border-radius: 10px; border: none; font-size: 13px;
  cursor: pointer; transition: all 0.15s;
}

.fm-actions {
  display: flex; gap: 12px; padding-top: 16px;
  border-top: 1px solid v-bind('isDark ? "rgba(255,255,255,0.1)" : "#f3f4f6"');
}
.fm-reset, .fm-apply {
  flex: 1; padding: 12px; border-radius: 12px; border: none;
  font-size: 14px; font-weight: 500; cursor: pointer; transition: opacity 0.15s;
}
.fm-reset { background: v-bind('isDark ? "#2c2c2e" : "#f3f4f6"'); color: v-bind('isDark ? "#8e8e93" : "#6b7280"'); }
.fm-apply { background: var(--accent-color); color: #fff; }
.fm-reset:hover, .fm-apply:hover { opacity: 0.9; }
</style>
