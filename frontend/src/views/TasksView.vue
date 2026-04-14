<template>
  <div class="page-container">
    <div class="page-content">
      <!-- 页面标题 -->
      <div class="page-header">
        <h1>待办任务</h1>
        <button @click="openAddTask" class="add-btn">
          <svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
          </svg>
        </button>
      </div>

      <!-- 任务统计 -->
      <div class="task-stats stats-animate">
        <div class="stat-item stat-pending"><div class="stat-number">{{ pendingTasks.length }}</div><div class="stat-label">待处理</div></div>
        <div class="stat-item stat-today"><div class="stat-number">{{ todayTasks.length }}</div><div class="stat-label">今日到期</div></div>
        <div class="stat-item stat-overdue"><div class="stat-number">{{ overdueTasks.length }}</div><div class="stat-label">已逾期</div></div>
      </div>

      <!-- 任务列表 -->
      <div class="task-list list-animate">
        <div v-if="tasks.length === 0" class="empty-state">
          <div class="empty-icon">✅</div>
          <p class="empty-text">暂无待办任务</p>
          <button @click="openAddTask" class="add-task-btn">添加任务</button>
        </div>

        <div v-else>
          <!-- 逾期任务 -->
          <div v-if="overdueTasks.length > 0" class="task-group">
            <div class="group-title group-title-overdue">已逾期</div>
            <div v-for="task in overdueTasks" :key="task.id" class="task-item">
              <div class="task-checkbox" @click="toggleTask(task.id)">
                <div v-if="task.completed" class="checked">✓</div>
              </div>
              <div class="task-content" @click="startEditTask(task)">
                <div class="task-title" :class="{ completed: task.completed }">{{ task.title }}</div>
                <div class="task-meta">
                  <span class="task-date task-date-overdue">逾期 {{ Math.abs(task.daysLeft) }} 天</span>
                  <span v-if="task.amount" class="task-amount">¥{{ formatAmount(task.amount) }}</span>
                </div>
              </div>
              <button @click="confirmDelete(task.id)" class="delete-btn">🗑️</button>
            </div>
          </div>

          <!-- 今日任务 -->
          <div v-if="todayTasks.length > 0" class="task-group">
            <div class="group-title group-title-today">今日到期</div>
            <div v-for="task in todayTasks" :key="task.id" class="task-item">
              <div class="task-checkbox" @click="toggleTask(task.id)">
                <div v-if="task.completed" class="checked">✓</div>
              </div>
              <div class="task-content" @click="startEditTask(task)">
                <div class="task-title" :class="{ completed: task.completed }">{{ task.title }}</div>
                <div class="task-meta">
                  <span class="task-date task-date-today">今天</span>
                  <span v-if="task.amount" class="task-amount">¥{{ formatAmount(task.amount) }}</span>
                </div>
              </div>
              <button @click="confirmDelete(task.id)" class="delete-btn">🗑️</button>
            </div>
          </div>

          <!-- 即将到期 -->
          <div v-if="upcomingTasks.length > 0" class="task-group">
            <div class="group-title group-title-upcoming">即将到期</div>
            <div v-for="task in upcomingTasks" :key="task.id" class="task-item">
              <div class="task-checkbox" @click="toggleTask(task.id)">
                <div v-if="task.completed" class="checked">✓</div>
              </div>
              <div class="task-content" @click="startEditTask(task)">
                <div class="task-title" :class="{ completed: task.completed }">{{ task.title }}</div>
                <div class="task-meta">
                  <span class="task-date">还有 {{ task.daysLeft }} 天</span>
                  <span v-if="task.amount" class="task-amount">¥{{ formatAmount(task.amount) }}</span>
                </div>
              </div>
              <button @click="confirmDelete(task.id)" class="delete-btn">🗑️</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加/编辑任务 - 复用 TaskForm 组件 -->
    <TaskForm
      v-model="showTaskModal"
      :initial-title="taskInitialData?.title"
      :initial-due-date="taskInitialData?.dueDate"
      :initial-amount="taskInitialData?.amount ? Number(taskInitialData.amount) : undefined"
      @save="handleSave"
      @toast="(msg: string) => log.error(msg)"
    />

    <!-- 删除确认 - 复用 ConfirmModal 组件 -->
    <ConfirmModal
      v-model="showDeleteConfirm"
      icon="🗑️"
      title="删除任务"
      message="删除后将无法恢复，确定要删除这个任务吗？"
      dangerous
      @confirm="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useThemeStore } from '../stores/theme'
import { formatAmount } from '../utils/formatters'
import { calcDaysLeft } from '../utils/date'
import type { TaskItem } from '../types'
import TaskForm from '../components/TaskForm.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import { GetAll as GetAllTasks, Create as CreateTask, ToggleComplete as ToggleTask, Delete as DeleteTask, Update as UpdateTask } from '../../wailsjs/go/service/TaskService'
import { createLogger } from '../utils/logger'

const log = createLogger('Tasks')

const theme = useThemeStore()
const isDark = computed(() => theme.isDark)

// 使用共享 TaskItem 接口，扩展本地字段
interface Task extends TaskItem {
  created_at?: string
  updated_at?: string
}

const tasks = ref<Task[]>([])

// TaskForm 组件状态
const showTaskModal = ref(false)
const taskMode = ref<'add' | 'edit'>('add')
const taskInitialData = ref<{ title: string; dueDate: string; amount: string } | null>(null)
const editingTaskId = ref<number | null>(null)

// 删除确认状态
const showDeleteConfirm = ref(false)
const deleteTaskId = ref<number | null>(null)

// 计算天数（复用共享工具）
function calculateDaysLeft(dueDate: string): number {
  return calcDaysLeft(dueDate)
}

const tasksWithDays = computed(() => {
  return tasks.value.map(task => ({
    ...task,
    daysLeft: calculateDaysLeft(task.due_date)
  }))
})

const pendingTasks = computed(() => tasksWithDays.value.filter(t => !t.completed))
const todayTasks = computed(() => tasksWithDays.value.filter(t => !t.completed && t.daysLeft === 0))
const overdueTasks = computed(() => tasksWithDays.value.filter(t => !t.completed && t.daysLeft < 0))
const upcomingTasks = computed(() => tasksWithDays.value.filter(t => !t.completed && t.daysLeft > 0))

async function loadTasks() {
  try {
    tasks.value = await GetAllTasks()
  } catch (e) {
    log.error('加载任务失败', e)
  }
}

onMounted(() => {
  loadTasks()
})

// 打开添加任务
function openAddTask() {
  taskMode.value = 'add'
  taskInitialData.value = null
  editingTaskId.value = null
  showTaskModal.value = true
}

// 开始编辑
function startEditTask(task: Task) {
  editingTaskId.value = task.id
  taskMode.value = 'edit'
  taskInitialData.value = {
    title: task.title,
    dueDate: task.due_date,
    amount: task.amount ? String(task.amount) : ''
  }
  showTaskModal.value = true
}

// TaskForm save 回调
async function handleSave(data: { title: string; dueDate: string; amount: number }) {
  try {
    if (taskMode.value === 'edit' && editingTaskId.value !== null) {
      await UpdateTask(editingTaskId.value, data.title, data.dueDate, data.amount)
    } else {
      await CreateTask(data.title, data.dueDate, data.amount)
    }
    await loadTasks()
  } catch (e) {
    log.error(`${taskMode.value === 'edit' ? '更新' : '添加'}任务失败`, e)
  }
}

// 删除确认流程
function confirmDelete(id: number) {
  deleteTaskId.value = id
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (deleteTaskId.value === null) return
  try {
    await DeleteTask(deleteTaskId.value)
    await loadTasks()
  } catch (e) {
    log.error('删除任务失败', e)
  } finally {
    deleteTaskId.value = null
  }
}

// 切换完成状态
async function toggleTask(id: number) {
  try {
    await ToggleTask(id)
    await loadTasks()
  } catch (e) {
    log.error('切换任务状态失败', e)
  }
}
</script>

<style scoped>
/* ====== 页面布局（使用 CSS 变量和 v-bind 替代内联 style） ====== */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h1 {
  font-size: 20px;
  font-weight: 600;
  color: v-bind('isDark ? "#fff" : "#1f2937"');
}

.add-btn {
  width: 40px; height: 40px; border-radius: 12px; border: none;
  background: var(--accent-color); color: #fff; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: transform 0.15s;
}
.add-btn:hover { transform: scale(1.05); }

.task-stats {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px;
  margin-bottom: 16px;
}
.stats-animate { animation-delay: 0.1s; }
.list-animate { animation-delay: 0.2s; }

.stat-item {
  border-radius: 12px; padding: 16px; text-align: center;
  background: v-bind('isDark ? "#1c1c1e" : "#fff"');
}
.stat-number { font-size: 24px; font-weight: 700; margin-bottom: 4px; }
.stat-label { font-size: 12px; color: v-bind('isDark ? "#8e8e93" : "#6b7280"'); }
.stat-pending .stat-number { color: #a855f7; }
.stat-today .stat-number { color: #30d158; }
.stat-overdue .stat-number { color: #ff9f0a; }

.task-list {
  border-radius: 16px; padding: 16px; flex: 1;
  background: v-bind('isDark ? "#1c1c1e" : "#fff"');
}

.empty-state { text-align: center; padding: 40px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-text { color: v-bind('isDark ? "#636366" : "#9ca3af"'); }

.add-task-btn {
  margin-top: 16px; padding: 10px 24px; border-radius: 12px; border: none;
  background: var(--accent-color); color: #fff; font-size: 14px; cursor: pointer;
}

.task-group { margin-bottom: 16px; }
.task-group:last-child { margin-bottom: 0; }

.group-title { font-size: 12px; font-weight: 500; margin-bottom: 8px; padding-left: 4px; }
.group-title-overdue { color: #ff453a; }
.group-title-today { color: #ff9f0a; }
.group-title-upcoming { color: #30d158; }

.task-item {
  display: flex; align-items: center; gap: 12px;
  padding: 12px; border-radius: 12px; margin-bottom: 8px;
  transition: background 0.15s;
}
.task-item:hover {
  background: v-bind('isDark ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.02)"');
}

.task-checkbox {
  width: 22px; height: 22px; border-radius: 50%;
  border: 2px solid #a855f7; /* 品牌色替代靛蓝 */
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; flex-shrink: 0;
}
.task-checkbox .checked {
  width: 14px; height: 14px; border-radius: 50%;
  background: #a855f7; color: #fff; font-size: 10px;
  display: flex; align-items: center; justify-content: center;
}

.task-content { flex: 1; min-width: 0; cursor: pointer; }

.task-title {
  font-size: 14px; font-weight: 500; margin-bottom: 4px;
  color: v-bind('isDark ? "#fff" : "#1f2937"');
}
.task-title.completed { text-decoration: line-through; opacity: 0.5; }

.task-meta { display: flex; gap: 8px; font-size: 12px; }
.task-date { color: #8e8e93; }
.task-date-overdue { color: #ff453a; }
.task-date-today { color: #ff9f0a; }
.task-amount { color: #a855f7; font-weight: 500; } /* 品牌色 */

.delete-btn {
  padding: 4px; border: none; background: transparent;
  cursor: pointer; opacity: 0.6; transition: opacity 0.15s;
}
.delete-btn:hover { opacity: 1; }
</style>
