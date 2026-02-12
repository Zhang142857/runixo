// AI 助手快捷键
import { onMounted, onUnmounted } from 'vue'

export interface AIHotkeys {
  send?: () => void
  newConversation?: () => void
  clearInput?: () => void
  focusInput?: () => void
  deleteConversation?: () => void
  exportConversation?: () => void
}

export function useAIHotkeys(handlers: AIHotkeys) {
  const handleKeydown = (e: KeyboardEvent) => {
    // Ctrl/Cmd + Enter: 发送消息
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault()
      handlers.send?.()
      return
    }

    // Ctrl/Cmd + N: 新建对话
    if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
      e.preventDefault()
      handlers.newConversation?.()
      return
    }

    // Ctrl/Cmd + K: 清空输入
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
      e.preventDefault()
      handlers.clearInput?.()
      return
    }

    // Ctrl/Cmd + L: 聚焦输入框
    if ((e.ctrlKey || e.metaKey) && e.key === 'l') {
      e.preventDefault()
      handlers.focusInput?.()
      return
    }

    // Ctrl/Cmd + D: 删除当前对话
    if ((e.ctrlKey || e.metaKey) && e.key === 'd') {
      e.preventDefault()
      handlers.deleteConversation?.()
      return
    }

    // Ctrl/Cmd + E: 导出对话
    if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
      e.preventDefault()
      handlers.exportConversation?.()
      return
    }

    // Esc: 停止生成
    if (e.key === 'Escape') {
      // 由组件自己处理
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })
}
