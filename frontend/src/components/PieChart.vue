<template>
  <div class="pie-chart-container">
    <svg :width="size" :height="size" viewBox="0 0 100 100">
      <!-- 背景圆环 -->
      <circle
        cx="50"
        cy="50"
        :r="radius"
        fill="none"
        class="bg-ring"
        :stroke-width="strokeWidth"
      />
      <!-- 数据段 -->
      <circle
        v-for="(segment, index) in segments"
        :key="index"
        cx="50"
        cy="50"
        :r="radius"
        fill="none"
        :stroke="segment.color"
        :stroke-width="strokeWidth"
        :stroke-dasharray="segment.dashArray"
        :stroke-dashoffset="segment.offset"
        stroke-linecap="round"
        class="pie-segment"
        :style="{ animationDelay: index * 100 + 'ms' }"
      />
      <!-- 中心文字 -->
      <text
        x="50"
        y="46"
        text-anchor="middle"
        class="center-label"
        font-size="14"
        font-weight="700"
      >
        {{ centerLabel }}
      </text>
      <text
        x="50"
        y="60"
        text-anchor="middle"
        class="center-value"
        font-size="11"
        font-weight="600"
      >
        {{ centerValue }}
      </text>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface DataItem {
  label?: string
  value: number
  color: string
}

const props = defineProps<{
  data: DataItem[]
  size?: number
  strokeWidth?: number
  centerLabel?: string
  centerValue?: string
  isDark?: boolean
}>()

const size = computed(() => props.size || 140)
const strokeWidth = computed(() => props.strokeWidth || 10)
const radius = computed(() => 40)

const circumference = computed(() => 2 * Math.PI * radius.value)

const total = computed(() => props.data.reduce((sum, item) => sum + item.value, 0))

const segments = computed(() => {
  let accumulated = 0
  return props.data.map((item) => {
    const percentage = total.value > 0 ? item.value / total.value : 0
    const dashArray = `${percentage * circumference.value} ${circumference.value}`
    const offset = -accumulated * circumference.value
    accumulated += percentage
    return {
      color: item.color,
      dashArray,
      offset,
      percentage: percentage * 100,
      label: item.label,
      value: item.value
    }
  })
})
</script>

<style scoped>
.pie-chart-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-ring {
  stroke: var(--bg-input);
}

.center-label {
  fill: var(--text-primary);
}

.center-value {
  fill: var(--text-secondary);
}

.pie-segment {
  transform-origin: center;
  transform: rotate(-90deg);
  animation: segmentGrow 0.6s ease-out forwards;
  opacity: 0;
}

@keyframes segmentGrow {
  from {
    opacity: 0;
    stroke-dasharray: 0 251.2;
  }
  to {
    opacity: 1;
  }
}
</style>
