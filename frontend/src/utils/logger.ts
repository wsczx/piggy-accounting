/**
 * 前端日志工具
 * 所有日志写入控制台 + 通过 Wails 桥接写入后端日志文件
 */

type LogLevel = 'INFO' | 'WARN' | 'ERROR' | 'DEBUG'

function timestamp(): string {
  return new Date().toISOString().slice(11, 23) // HH:MM:SS.mmm
}

function formatMsg(level: LogLevel, tag: string, ...args: unknown[]): string {
  const ts = timestamp()
  let msg = ''
  for (const a of args) {
    if (typeof a === 'object') {
      try { msg += JSON.stringify(a) + ' ' } catch { msg += String(a) + ' ' }
    } else {
      msg += String(a) + ' '
    }
  }
  return `[${ts}] [${level}] [${tag}] ${msg.trim()}`
}

/**
 * 前端日志器
 * - tag: 模块标签 (如 'Store', 'App', 'Panel')
 * - silent: 是否只写后端日志，不在控制台输出
 */
export function createLogger(tag: string, silent = false) {
  return {
    info(...args: unknown[]) {
      if (!silent) console.log(formatMsg('INFO', tag, ...args))
      logToBackend('INFO', formatMsg('INFO', tag, ...args))
    },
    warn(...args: unknown[]) {
      if (!silent) console.warn(formatMsg('WARN', tag, ...args))
      logToBackend('WARN', formatMsg('WARN', tag, ...args))
    },
    error(...args: unknown[]) {
      if (!silent) console.error(formatMsg('ERROR', tag, ...args))
      logToBackend('ERROR', formatMsg('ERROR', tag, ...args))
    },
    debug(...args: unknown[]) {
      if (!silent) console.debug(formatMsg('DEBUG', tag, ...args))
    },
  }
}

// 默认日志器
export const logger = createLogger('App')

/**
 * 异步写入后端日志（不 await，避免阻塞）
 */
async function logToBackend(level: string, message: string) {
  try {
    // 动态导入避免构建时依赖问题
    const { LogFrontend } = await import('../../wailsjs/go/main/App')
    LogFrontend(level, message)
  } catch {
    // Wails 不存在时静默忽略（开发模式 / 浏览器预览）
  }
}
