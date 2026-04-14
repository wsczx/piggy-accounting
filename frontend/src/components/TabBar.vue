<template>
  <nav class="tabbar-nav">
    <!-- 首页 -->
    <button @click="$router.push('/')"
            class="tab-item"
            :class="{ 'tab-active': $route.path === '/' }">
      <svg width="22" height="22" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
        <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1"/>
      </svg>
      <span class="tab-label">首页</span>
    </button>

    <!-- 账单 -->
    <button @click="$router.push('/records')"
            class="tab-item"
            :class="{ 'tab-active': $route.path === '/records' }">
      <svg width="22" height="22" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
      </svg>
      <span class="tab-label">账单</span>
    </button>

    <!-- 记账按钮（居中突出） -->
    <button @click="$emit('newRecord')" class="tab-add-btn">
      <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="#fff" stroke-width="2.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
      </svg>
    </button>

    <!-- 统计 -->
    <button @click="$router.push('/statistics')"
            class="tab-item"
            :class="{ 'tab-active': $route.path === '/statistics' }">
      <svg width="22" height="22" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
      </svg>
      <span class="tab-label">统计</span>
    </button>

    <!-- 我的 -->
    <button @click="$router.push('/profile')"
            class="tab-item"
            :class="{ 'tab-active': $route.path === '/profile' }">
      <svg width="22" height="22" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
        <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
      </svg>
      <span class="tab-label">我的</span>
    </button>
  </nav>
</template>

<script setup lang="ts">
defineEmits<{ (e: 'newRecord'): void }>()
</script>

<style scoped>
.tabbar-nav {
  height: 60px;
  flex-shrink: 0;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  padding: 0 8px;
  z-index: 10;
  transition: background 0.3s, border-top-color 0.3s;
  border-top: 1px solid var(--border-color-light);
  background: var(--card-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 12px 8px;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  background: transparent;
  transition: color 0.15s;
  color: var(--text-muted);
  flex: 1;
}
.tab-item:hover {
  color: var(--text-secondary);
}
.tab-item.tab-active {
  color: var(--accent-color);
  animation: tabBounce 0.35s ease;
}
.tab-label {
  font-size: 10px;
  font-weight: 500;
}
.tab-add-btn {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  cursor: pointer;
  margin-top: -16px;
  transition: transform 0.15s;
  background: var(--brand-gradient);
  box-shadow: 0 4px 16px rgba(168, 85, 247, 0.35);
  padding: 0;
  flex-shrink: 0;
}
.tab-add-btn:hover {
  transform: scale(1.05);
}
.tab-add-btn:active {
  transform: scale(0.95);
}

@keyframes tabBounce {
  0% { transform: scale(1); }
  30% { transform: scale(1.15); }
  60% { transform: scale(0.95); }
  100% { transform: scale(1); }
}
</style>
