/**
 * AI 审计日志 - 记录所有 AI 工具调用
 */
import { app } from 'electron'
import { appendFile, mkdirSync, existsSync, readdirSync, unlinkSync } from 'fs'
import { join } from 'path'

const LOG_DIR = () => join(app.getPath('userData'), 'ai-audit')
const MAX_LOG_DAYS = 30

export interface AuditEntry {
  timestamp: string
  serverId: string
  toolName: string
  args: Record<string, unknown>
  result: { success: boolean; error?: string }
  userConfirmed: boolean
  dangerLevel: number
}

function ensureDir() {
  const dir = LOG_DIR()
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
}

/** 记录审计日志（异步） */
export function logToolCall(entry: AuditEntry): void {
  try {
    ensureDir()
    const date = new Date().toISOString().slice(0, 10)
    const logFile = join(LOG_DIR(), `${date}.jsonl`)
    appendFile(logFile, JSON.stringify(entry) + '\n', { mode: 0o600 }, (err) => {
      if (err) console.error('[AuditLogger] 写入失败:', err)
    })
    cleanOldLogs()
  } catch (err) {
    console.error('[AuditLogger] 错误:', err)
  }
}

/** 清理超过 MAX_LOG_DAYS 天的日志 */
function cleanOldLogs(): void {
  try {
    const dir = LOG_DIR()
    if (!existsSync(dir)) return
    const cutoff = new Date()
    cutoff.setDate(cutoff.getDate() - MAX_LOG_DAYS)
    const cutoffStr = cutoff.toISOString().slice(0, 10)

    for (const file of readdirSync(dir)) {
      if (file.endsWith('.jsonl') && file.slice(0, 10) < cutoffStr) {
        unlinkSync(join(dir, file))
      }
    }
  } catch { /* 清理失败不影响主流程 */ }
}
