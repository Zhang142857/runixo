// 对话相关类型定义
export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  tokens?: { prompt: number; completion: number; total: number }
  attachments?: Array<{ name: string; path: string; type: string }>
}

export interface Conversation {
  id: string
  title: string
  agentId?: string
  serverId?: string
  messages: Message[]
  createdAt: number
  updatedAt: number
  tokenUsage: { prompt: number; completion: number; total: number }
}

export interface ConversationIndex {
  id: string
  title: string
  agentId?: string
  serverId?: string
  messageCount: number
  lastMessage?: string
  createdAt: number
  updatedAt: number
}

export interface SearchResult {
  conversationId: string
  messageId: string
  content: string
  score: number
  context: string
}

export interface ExportData {
  version: string
  exportedAt: number
  conversations: Conversation[]
}
