import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(false)

  function init() {
    const saved = localStorage.getItem('piggy-theme')
    if (saved === 'dark') {
      isDark.value = true
    } else if (saved === 'light') {
      isDark.value = false
    } else {
      // 跟随系统
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
      // 监听系统主题变化（仅在跟随系统模式下响应）
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (!localStorage.getItem('piggy-theme')) {
          isDark.value = e.matches
        }
      })
    }
  }

  function toggle() {
    isDark.value = !isDark.value
    localStorage.setItem('piggy-theme', isDark.value ? 'dark' : 'light')
  }

  return { isDark, init, toggle }
})
