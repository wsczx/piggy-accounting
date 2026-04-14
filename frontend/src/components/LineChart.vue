<template>
  <div class="line-chart-container">
    <svg :width="width" :height="height" :viewBox="`0 0 ${width} ${height}`">
      <!-- 网格线 -->
      <g class="grid-lines">
        <line
          v-for="i in 5"
          :key="i"
          :x1="paddingLeft"
          :y1="paddingTop + (chartHeight / 4) * (i - 1)"
          :x2="width - paddingRight"
          :y2="paddingTop + (chartHeight / 4) * (i - 1)"
          class="grid-line"
          stroke-width="1"
        />
      </g>

      <!-- 收入区域填充 -->
      <path
        v-if="incomePoints.length > 1"
        :d="incomeAreaPath"
        fill="url(#incomeGradient)"
        opacity="0.3"
        class="area-path"
      />

      <!-- 支出区域填充 -->
      <path
        v-if="expensePoints.length > 1"
        :d="expenseAreaPath"
        fill="url(#expenseGradient)"
        opacity="0.3"
        class="area-path"
      />

      <!-- 收入折线 -->
      <path
        v-if="incomePoints.length > 1"
        :d="incomeLinePath"
        fill="none"
        stroke="var(--income-color)"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="line-path"
      />

      <!-- 支出折线 -->
      <path
        v-if="expensePoints.length > 1"
        :d="expenseLinePath"
        fill="none"
        stroke="var(--expense-color)"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="line-path"
      />

      <!-- 收入数据点 -->
      <g v-if="showPoints">
        <circle
          v-for="(point, index) in incomePoints"
          :key="`income-${index}`"
          :cx="point.x"
          :cy="point.y"
          r="4"
          fill="var(--income-color)"
          class="data-point"
          :style="{ animationDelay: index * 30 + 'ms' }"
        />
      </g>

      <!-- 支出数据点 -->
      <g v-if="showPoints">
        <circle
          v-for="(point, index) in expensePoints"
          :key="`expense-${index}`"
          :cx="point.x"
          :cy="point.y"
          r="4"
          fill="var(--expense-color)"
          class="data-point"
          :style="{ animationDelay: index * 30 + 'ms' }"
        />
      </g>

      <!-- X轴标签 -->
      <g class="x-labels">
        <text
          v-for="(label, index) in xLabels"
          :key="index"
          :x="label.x"
          :y="height - 5"
          text-anchor="middle"
          class="x-label-text"
          font-size="10"
        >
          {{ label.text }}
        </text>
      </g>

      <!-- 渐变定义 -->
      <defs>
        <linearGradient id="incomeGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--income-color)" stop-opacity="0.6"/>
          <stop offset="100%" stop-color="var(--income-color)" stop-opacity="0"/>
        </linearGradient>
        <linearGradient id="expenseGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--expense-color)" stop-opacity="0.6"/>
          <stop offset="100%" stop-color="var(--expense-color)" stop-opacity="0"/>
        </linearGradient>
      </defs>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface DataPoint {
  label: string
  income: number
  expense: number
}

const props = defineProps<{
  data: DataPoint[]
  width?: number
  height?: number
  isDark?: boolean
  showPoints?: boolean
}>()

const width = computed(() => props.width || 400)
const height = computed(() => props.height || 160)
const paddingLeft = 10
const paddingRight = 10
const paddingTop = 20
const paddingBottom = 25

const chartWidth = computed(() => width.value - paddingLeft - paddingRight)
const chartHeight = computed(() => height.value - paddingTop - paddingBottom)

const maxValue = computed(() => {
  const allValues = props.data.flatMap(d => [d.income, d.expense])
  const max = Math.max(...allValues, 1)
  return max * 1.1 // 留 10% 边距
})

const getX = (index: number) => {
  if (props.data.length <= 1) return paddingLeft + chartWidth.value / 2
  return paddingLeft + (index / (props.data.length - 1)) * chartWidth.value
}

const getY = (value: number) => {
  return paddingTop + chartHeight.value - (value / maxValue.value) * chartHeight.value
}

const incomePoints = computed(() => {
  return props.data.map((d, i) => ({
    x: getX(i),
    y: getY(d.income),
    value: d.income
  }))
})

const expensePoints = computed(() => {
  return props.data.map((d, i) => ({
    x: getX(i),
    y: getY(d.expense),
    value: d.expense
  }))
})

const generateLinePath = (points: { x: number; y: number }[]) => {
  if (points.length < 2) return ''

  // 使用简单的直线连接
  return points.reduce((path, point, i) => {
    if (i === 0) return `M ${point.x} ${point.y}`
    return `${path} L ${point.x} ${point.y}`
  }, '')
}

const generateAreaPath = (points: { x: number; y: number }[]) => {
  if (points.length < 2) return ''

  const linePath = generateLinePath(points)
  const bottomY = paddingTop + chartHeight.value
  const firstX = points[0].x
  const lastX = points[points.length - 1].x

  return `${linePath} L ${lastX} ${bottomY} L ${firstX} ${bottomY} Z`
}

const incomeLinePath = computed(() => generateLinePath(incomePoints.value))
const expenseLinePath = computed(() => generateLinePath(expensePoints.value))
const incomeAreaPath = computed(() => generateAreaPath(incomePoints.value))
const expenseAreaPath = computed(() => generateAreaPath(expensePoints.value))

const xLabels = computed(() => {
  // 根据数据量决定显示哪些标签
  const total = props.data.length
  if (total <= 6) {
    return props.data.map((d, i) => ({
      x: getX(i),
      text: d.label
    }))
  }
  // 数据多时只显示部分标签
  const step = Math.ceil(total / 6)
  return props.data
    .map((d, i) => ({ x: getX(i), text: d.label, index: i }))
    .filter((_, i) => i % step === 0 || i === total - 1)
})
</script>

<style scoped>
.line-chart-container {
  width: 100%;
  height: 100%;
}

.line-chart-container svg {
  width: 100%;
  height: 100%;
}

.grid-line {
  stroke: var(--border-color-light);
}

.x-label-text {
  fill: var(--text-secondary);
}

.line-path {
  animation: drawLine 0.8s ease-out forwards;
  stroke-dasharray: 1000;
  stroke-dashoffset: 1000;
}

@keyframes drawLine {
  to {
    stroke-dashoffset: 0;
  }
}

.area-path {
  animation: fadeInArea 0.6s ease-out 0.3s forwards;
  opacity: 0;
}

@keyframes fadeInArea {
  to {
    opacity: 0.3;
  }
}

.data-point {
  animation: popIn 0.3s ease-out forwards;
  opacity: 0;
  transform-origin: center;
}

@keyframes popIn {
  0% {
    opacity: 0;
    transform: scale(0);
  }
  70% {
    transform: scale(1.2);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
