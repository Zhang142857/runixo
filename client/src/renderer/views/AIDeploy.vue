<template>
  <div class="ai-deploy-page">
    <!-- 顶部标题 -->
    <div class="page-header">
      <div class="header-left">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2L4 8l8 6 8-6-8-6z"/><path d="M4 12l8 6 8-6"/><path d="M4 16l8 6 8-6"/>
          </svg>
        </div>
        <div>
          <h2>AI 部署</h2>
          <p class="subtitle">告诉 AI 你想部署什么，它会自动完成一切</p>
        </div>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" size="default" clearable style="width: 180px">
          <el-option v-for="s in connectedServers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="deploy-content">
      <!-- 无服务器提示 -->
      <div v-if="!selectedServer" class="empty-state">
        <el-icon :size="48" color="var(--text-muted)"><Monitor /></el-icon>
        <p>请先选择一个已连接的服务器</p>
      </div>

      <template v-else>
        <!-- 部署输入区 -->
        <div class="deploy-input-section">
          <div class="input-card">
            <el-input
              v-model="deployRequest"
              type="textarea"
              :rows="3"
              :placeholder="placeholders[placeholderIdx]"
              resize="none"
              @keydown.enter.ctrl="startDeploy"
            />
            <div class="input-footer">
              <span class="hint">Ctrl + Enter 发送</span>
              <el-button type="primary" :loading="deploying" :disabled="!deployRequest.trim()" @click="startDeploy">
                <el-icon v-if="!deploying"><Promotion /></el-icon>
                {{ deploying ? '部署中...' : '开始部署' }}
              </el-button>
            </div>
          </div>
        </div>

        <!-- 快捷部署卡片 -->
        <div v-if="!deploying && deployLogs.length === 0" class="quick-deploy">
          <h3>快速部署</h3>
          <div class="quick-grid">
            <div v-for="item in quickItems" :key="item.name" class="quick-card" @click="deployRequest = item.prompt">
              <span class="quick-icon">{{ item.icon }}</span>
              <div class="quick-info">
                <div class="quick-name">{{ item.name }}</div>
                <div class="quick-desc">{{ item.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 部署日志 -->
        <div v-if="deployLogs.length > 0" class="deploy-logs">
          <div class="logs-header">
            <h3>部署进度</h3>
            <el-tag :type="deployStatus === 'success' ? 'success' : deployStatus === 'error' ? 'danger' : 'warning'" size="small">
              {{ deployStatus === 'success' ? '完成' : deployStatus === 'error' ? '失败' : '进行中' }}
            </el-tag>
          </div>
          <div class="logs-container" ref="logsRef">
            <div v-for="(log, i) in deployLogs" :key="i" class="log-line" :class="log.type">
              <span class="log-time">{{ log.time }}</span>
              <span class="log-icon">{{ log.type === 'success' ? '✓' : log.type === 'error' ? '✗' : log.type === 'info' ? 'ℹ' : '›' }}</span>
              <span class="log-text">{{ log.text }}</span>
            </div>
            <div v-if="deploying" class="log-line step">
              <span class="log-time">{{ now() }}</span>
              <span class="log-icon spinning">⟳</span>
              <span class="log-text">{{ currentStep }}</span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useServerStore } from '@/stores/server'
import { useAIStore } from '@/stores/ai'
import { Monitor, Promotion } from '@element-plus/icons-vue'

const serverStore = useServerStore()
const aiStore = useAIStore()

const selectedServer = ref<string | null>(serverStore.currentServerId)
const deployRequest = ref('')
const deploying = ref(false)
const deployStatus = ref<'idle' | 'running' | 'success' | 'error'>('idle')
const currentStep = ref('')
const deployLogs = ref<Array<{ time: string; type: string; text: string }>>([])
const logsRef = ref<HTMLElement | null>(null)
const placeholderIdx = ref(0)

const connectedServers = computed(() => serverStore.connectedServers)

const placeholders = [
  '例如：帮我部署一个 WordPress 博客',
  '例如：部署 Nextcloud 私有云盘，端口 8080',
  '例如：安装 Redis 缓存服务',
  '例如：部署 Jellyfin 媒体服务器'
]

const quickItems = [
  { icon: '📝', name: 'WordPress', desc: '博客/CMS 系统', prompt: '帮我部署一个 WordPress 博客' },
  { icon: '☁️', name: 'Nextcloud', desc: '私有云盘', prompt: '部署 Nextcloud 私有云盘' },
  { icon: '📊', name: 'Grafana', desc: '监控仪表盘', prompt: '部署 Grafana 监控面板' },
  { icon: '🐳', name: 'Portainer', desc: '容器管理', prompt: '部署 Portainer 容器管理面板' },
  { icon: '📸', name: 'Immich', desc: '照片管理', prompt: '部署 Immich 照片管理服务' },
  { icon: '🎬', name: 'Jellyfin', desc: '媒体服务器', prompt: '部署 Jellyfin 媒体服务器' },
]

setInterval(() => { placeholderIdx.value = (placeholderIdx.value + 1) % placeholders.length }, 4000)

function now() { return new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) }

function addLog(type: string, text: string) {
  deployLogs.value.push({ time: now(), type, text })
  nextTick(() => { logsRef.value?.scrollTo({ top: logsRef.value.scrollHeight, behavior: 'smooth' }) })
}

async function startDeploy() {
  const content = deployRequest.value.trim()
  if (!content || deploying.value || !selectedServer.value) return

  deploying.value = true
  deployStatus.value = 'running'
  deployLogs.value = []
  currentStep.value = '分析部署需求...'
  addLog('info', `开始部署: ${content}`)

  let cleanupListener: (() => void) | null = null

  try {
    // 使用 AI stream 进行部署对话
    aiStore.createConversation(true, selectedServer.value)
    aiStore.addUserMessage(content)
    aiStore.startProcessing('部署中...')
    aiStore.createStreamingMessage()

    let fullResponse = ''
    cleanupListener = window.electronAPI.ai.onStreamDelta((delta: any) => {
      if (delta.content) fullResponse += delta.content
      if (delta.type === 'tool-call') {
        currentStep.value = `执行: ${delta.toolName}`
        addLog('step', `调用工具: ${delta.toolName}`)
      }
      if (delta.type === 'tool-confirm') {
        // 部署页面自动批准工具执行
        addLog('step', `授权执行: ${delta.toolName}`)
        window.electronAPI.ai.confirmTool(delta.confirmId, true)
      }
      if (delta.type === 'tool-result') {
        const r = delta.result
        addLog(r?.success ? 'success' : 'error', r?.message || (r?.success ? '执行成功' : '执行失败'))
      }
      aiStore.appendToLastMessage(delta)
    })

    await window.electronAPI.ai.streamChat(content, {
      serverId: selectedServer.value,
      agentId: 'deploy',
      history: []
    })

    deployStatus.value = 'success'
    addLog('success', '部署流程完成')
    if (fullResponse) addLog('info', 'AI 总结: ' + fullResponse.slice(0, 200))
  } catch (e) {
    deployStatus.value = 'error'
    addLog('error', `部署失败: ${(e as Error).message}`)
  } finally {
    cleanupListener?.()
    deploying.value = false
    currentStep.value = ''
    aiStore.finalizeStreamingMessage()
    aiStore.endProcessing()
  }
}
</script>

<style lang="scss" scoped>
.ai-deploy-page { height: 100%; display: flex; flex-direction: column; background: var(--bg-color); }

.page-header {
  padding: 20px 28px; display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid var(--border-color); background: var(--bg-secondary);
  .header-left { display: flex; align-items: center; gap: 14px; }
  .header-icon {
    width: 40px; height: 40px; border-radius: 10px;
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    color: #fff; display: flex; align-items: center; justify-content: center;
  }
  h2 { margin: 0; font-size: 18px; font-weight: 700; }
  .subtitle { margin: 2px 0 0; font-size: 13px; color: var(--text-secondary); }
}

.deploy-content { flex: 1; overflow-y: auto; padding: 24px 28px; }

.empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 300px; gap: 12px; color: var(--text-muted); font-size: 14px;
}

.deploy-input-section { max-width: 700px; margin: 0 auto 28px; }
.input-card {
  border: 1px solid var(--border-color); border-radius: 14px;
  background: var(--bg-secondary); overflow: hidden;
  transition: border-color 0.3s, box-shadow 0.3s;
  &:focus-within { border-color: var(--primary-color); box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.1); }
  :deep(.el-textarea__inner) {
    background: transparent !important; border: none !important;
    box-shadow: none !important; padding: 16px 18px 8px;
    font-size: 14px; line-height: 1.6; color: var(--text-color); resize: none;
  }
  .input-footer {
    display: flex; justify-content: space-between; align-items: center; padding: 8px 14px 12px;
    .hint { font-size: 11px; color: var(--text-muted); }
  }
}

.quick-deploy {
  max-width: 700px; margin: 0 auto;
  h3 { font-size: 15px; font-weight: 600; margin: 0 0 14px; }
}
.quick-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.quick-card {
  display: flex; align-items: center; gap: 10px;
  padding: 14px; border-radius: 10px; cursor: pointer;
  border: 1px solid var(--border-color); background: var(--bg-secondary);
  transition: all 0.2s;
  &:hover { border-color: var(--primary-color); transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
  .quick-icon { font-size: 24px; }
  .quick-name { font-size: 13px; font-weight: 600; }
  .quick-desc { font-size: 11px; color: var(--text-secondary); margin-top: 2px; }
}

.deploy-logs {
  max-width: 700px; margin: 0 auto;
  .logs-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
  h3 { font-size: 15px; font-weight: 600; margin: 0; }
}
.logs-container {
  background: var(--bg-secondary); border: 1px solid var(--border-color);
  border-radius: 10px; padding: 14px; max-height: 400px; overflow-y: auto;
  font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 12px;
}
.log-line {
  display: flex; align-items: flex-start; gap: 8px; padding: 3px 0; line-height: 1.6;
  .log-time { color: var(--text-muted); min-width: 70px; }
  .log-icon { min-width: 14px; text-align: center; }
  .log-text { flex: 1; word-break: break-all; }
  &.success .log-icon { color: #10b981; }
  &.error .log-icon { color: #ef4444; }
  &.info .log-icon { color: #6366f1; }
  &.step .log-icon { color: #f59e0b; }
}
.spinning { animation: spin 1s linear infinite; display: inline-block; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
