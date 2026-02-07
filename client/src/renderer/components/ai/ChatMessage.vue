<template>
  <div class="chat-message" :class="[message.role, { streaming: message.isStreaming }]">
    <div class="message-avatar">
      <div v-if="message.role === 'user'" class="avatar user-avatar"><el-icon><User /></el-icon></div>
      <div v-else class="avatar ai-avatar"><el-icon><MagicStick /></el-icon></div>
    </div>

    <div class="message-body">
      <div class="message-header">
        <span class="role-name">{{ message.role === 'user' ? '你' : 'AI 助手' }}</span>
        <span class="message-time">{{ formatTime(message.timestamp) }}</span>
      </div>

      <!-- 思考过程（可折叠） -->
      <div v-if="message.thinking" class="thinking-block" :class="{ collapsed: !thinkingExpanded }">
        <div class="thinking-header" @click="thinkingExpanded = !thinkingExpanded">
          <el-icon class="thinking-icon"><Loading v-if="message.isStreaming && !message.content" /><Sunny v-else /></el-icon>
          <span>{{ message.isStreaming && !message.content ? '思考中...' : '思考过程' }}</span>
          <span class="thinking-toggle">{{ thinkingExpanded ? '收起' : '展开' }}</span>
        </div>
        <div v-show="thinkingExpanded" class="thinking-content">{{ message.thinking }}</div>
      </div>

      <!-- 消息正文 -->
      <div v-if="message.content" class="message-content" v-html="renderedContent"></div>

      <!-- 流式光标 -->
      <span v-if="message.isStreaming && message.content" class="streaming-cursor">▋</span>

      <!-- 工具调用 -->
      <div v-if="message.toolCalls?.length" class="tool-calls">
        <ToolExecution v-for="(tool, i) in message.toolCalls" :key="i" :tool-call="tool" :expanded="tool.expanded" @toggle="toggleToolExpand(i)" />
      </div>

      <!-- 任务计划 -->
      <TaskPlan v-if="message.plan" :plan="message.plan" class="message-plan" />

      <!-- 操作按钮 -->
      <div class="message-actions" v-if="message.role === 'assistant' && !message.isStreaming">
        <el-tooltip content="复制" placement="top"><el-button text size="small" @click="copyContent"><el-icon><CopyDocument /></el-icon></el-button></el-tooltip>
        <el-tooltip content="重新生成" placement="top"><el-button text size="small" @click="$emit('regenerate')"><el-icon><Refresh /></el-icon></el-button></el-tooltip>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { User, MagicStick, CopyDocument, Refresh, Loading, Sunny } from '@element-plus/icons-vue'
import ToolExecution from './ToolExecution.vue'
import TaskPlan from './TaskPlan.vue'
import type { Message } from '@/stores/ai'

const props = defineProps<{ message: Message }>()
defineEmits<{ regenerate: [] }>()

const thinkingExpanded = ref(false)

const renderedContent = computed(() => renderMarkdown(props.message.content))

function renderMarkdown(content: string): string {
  if (!content) return ''
  return content
    .replace(/```(\w*)\n([\s\S]*?)```/g, (_, lang, code) =>
      `<div class="code-block"><div class="code-header"><span>${lang || 'code'}</span><button class="copy-btn" onclick="navigator.clipboard.writeText(this.parentElement.nextElementSibling.textContent);this.textContent='已复制';setTimeout(()=>this.textContent='复制',1500)">复制</button></div><pre><code class="lang-${lang}">${code.replace(/</g, '&lt;')}</code></pre></div>`)
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    .replace(/^### (.+)$/gm, '<h4>$1</h4>')
    .replace(/^## (.+)$/gm, '<h3>$1</h3>')
    .replace(/^# (.+)$/gm, '<h2>$1</h2>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>[\s\S]*?<\/li>)/g, '<ul>$1</ul>')
    .replace(/\n/g, '<br>')
}

function formatTime(date: Date): string {
  if (!date) return ''
  return new Date(date).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function toggleToolExpand(index: number) {
  if (props.message.toolCalls?.[index]) props.message.toolCalls[index].expanded = !props.message.toolCalls[index].expanded
}

function copyContent() {
  navigator.clipboard.writeText(props.message.content)
  ElMessage.success('已复制')
}
</script>

<style lang="scss" scoped>
.chat-message {
  display: flex;
  gap: 14px;
  padding: 20px 24px;
  transition: background-color 0.15s;

  &:hover { background-color: rgba(255,255,255,0.02); }
  &:hover .message-actions { opacity: 1; }
}

.avatar {
  width: 32px; height: 32px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 16px; flex-shrink: 0;
}
.user-avatar { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: #fff; }
.ai-avatar { background: linear-gradient(135deg, #10b981, #06b6d4); color: #fff; }

.message-body { flex: 1; min-width: 0; }

.message-header {
  display: flex; align-items: center; gap: 8px; margin-bottom: 6px;
  .role-name { font-weight: 600; font-size: 13px; color: var(--text-color); }
  .message-time { font-size: 11px; color: var(--text-muted); }
}

// 思考过程块
.thinking-block {
  margin-bottom: 10px; border-radius: 10px; overflow: hidden;
  border: 1px solid rgba(245, 158, 11, 0.2);
  background: rgba(245, 158, 11, 0.05);

  .thinking-header {
    display: flex; align-items: center; gap: 8px;
    padding: 8px 12px; cursor: pointer; font-size: 12px;
    color: #f59e0b; user-select: none;
    &:hover { background: rgba(245, 158, 11, 0.08); }
  }
  .thinking-icon { font-size: 14px; animation: none; }
  .thinking-toggle { margin-left: auto; font-size: 11px; opacity: 0.7; }
  .thinking-content {
    padding: 0 12px 10px; font-size: 13px; line-height: 1.7;
    color: var(--text-secondary); white-space: pre-wrap; word-break: break-word;
    max-height: 300px; overflow-y: auto;
  }
  &.collapsed .thinking-content { display: none; }
}

// 消息正文
.message-content {
  font-size: 14px; line-height: 1.75; color: var(--text-color); word-break: break-word;

  :deep(.code-block) {
    margin: 10px 0; border-radius: 10px; overflow: hidden;
    border: 1px solid var(--border-color);
    .code-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 6px 14px; background: var(--bg-elevated); font-size: 12px; color: var(--text-secondary);
    }
    .copy-btn {
      background: none; border: 1px solid var(--border-color); border-radius: 4px;
      padding: 2px 8px; font-size: 11px; color: var(--text-secondary); cursor: pointer;
      &:hover { color: var(--text-color); border-color: var(--text-muted); }
    }
    pre { margin: 0; padding: 14px; background: var(--bg-tertiary); overflow-x: auto; }
    code { font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 13px; background: none; padding: 0; }
  }

  :deep(code) {
    background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;
    font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 13px;
  }
  :deep(h2), :deep(h3), :deep(h4) { margin: 14px 0 8px; font-weight: 600; }
  :deep(ul) { margin: 8px 0; padding-left: 20px; }
  :deep(li) { margin: 4px 0; }
  :deep(strong) { font-weight: 600; color: var(--text-color); }
}

.streaming-cursor { color: var(--primary-color); animation: blink 1s infinite; }
@keyframes blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }

.tool-calls { margin-top: 12px; display: flex; flex-direction: column; gap: 8px; }
.message-plan { margin-top: 12px; }

.message-actions {
  display: flex; gap: 4px; margin-top: 8px; opacity: 0; transition: opacity 0.15s;
  .el-button { color: var(--text-muted); &:hover { color: var(--primary-color); } }
}
</style>
