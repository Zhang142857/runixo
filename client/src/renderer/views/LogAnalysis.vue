<template>
  <div class="log-analysis-page">
    <div class="page-header">
      <div class="header-left">
        <h1>日志智能分析</h1>
        <p class="subtitle">查看、搜索和 AI 分析服务器日志</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="loadLogFiles">
          <el-option v-for="server in connectedServers" :key="server.id" :label="server.name" :value="server.id" />
        </el-select>
        <el-button @click="loadLogFiles" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <div class="main-content" v-if="selectedServer">
      <!-- 左侧日志文件列表 -->
      <div class="log-files-panel">
        <div class="panel-header">
          <h3>日志文件</h3>
          <el-input v-model="fileSearch" placeholder="搜索..." size="small" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>
        <div class="file-list">
          <div v-for="file in filteredLogFiles" :key="file.path" class="file-item" :class="{ active: selectedFile?.path === file.path }" @click="selectLogFile(file)">
            <el-icon><Document /></el-icon>
            <div class="file-info">
              <span class="file-name">{{ file.name }}</span>
              <span class="file-size">{{ formatSize(file.size) }}</span>
            </div>
          </div>
          <el-empty v-if="filteredLogFiles.length === 0" description="未找到日志文件" :image-size="60" />
        </div>
        <div class="custom-path">
          <el-input v-model="customPath" placeholder="自定义日志路径" size="small" @keyup.enter="loadCustomLog">
            <template #append>
              <el-button @click="loadCustomLog" size="small">加载</el-button>
            </template>
          </el-input>
        </div>
      </div>

      <!-- 右侧日志内容 -->
      <div class="log-content-panel">
        <div class="content-header" v-if="selectedFile">
          <div class="file-title">
            <el-icon><Document /></el-icon>
            <span>{{ selectedFile.path }}</span>
          </div>
          <div class="content-actions">
            <el-input v-model="searchQuery" placeholder="搜索日志内容..." class="search-input" clearable @input="filterLogContent">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="logLevel" placeholder="日志级别" class="level-select" clearable @change="filterLogContent">
              <el-option label="ERROR" value="error" />
              <el-option label="WARN" value="warn" />
              <el-option label="INFO" value="info" />
              <el-option label="DEBUG" value="debug" />
            </el-select>
            <el-button-group>
              <el-button @click="loadLogContent" :loading="loadingContent">刷新</el-button>
              <el-button @click="toggleTail" :type="isTailing ? 'primary' : ''">
                <el-icon><VideoPlay /></el-icon>{{ isTailing ? '停止' : '实时' }}
              </el-button>
              <el-button type="success" @click="analyzeWithAI" :loading="analyzing">
                <el-icon><MagicStick /></el-icon>AI 分析
              </el-button>
            </el-button-group>
          </div>
        </div>

        <!-- 日志统计 -->
        <div class="log-stats" v-if="selectedFile && logStats.total > 0">
          <div class="stat-item"><span class="stat-label">总行数</span><span class="stat-value">{{ logStats.total }}</span></div>
          <div class="stat-item error"><span class="stat-label">错误</span><span class="stat-value">{{ logStats.errors }}</span></div>
          <div class="stat-item warn"><span class="stat-label">警告</span><span class="stat-value">{{ logStats.warnings }}</span></div>
          <div class="stat-item info"><span class="stat-label">信息</span><span class="stat-value">{{ logStats.info }}</span></div>
        </div>

        <!-- 日志内容 -->
        <div class="log-viewer" ref="logViewer" v-if="selectedFile">
          <div v-for="(line, index) in displayedLines" :key="index" class="log-line" :class="getLineClass(line)">
            <span class="line-number">{{ line.lineNumber }}</span>
            <span class="line-content" v-html="highlightLine(line.content)"></span>
          </div>
          <div v-if="loadingContent" class="loading-indicator">
            <el-icon class="is-loading"><Loading /></el-icon>加载中...
          </div>
          <el-empty v-if="!loadingContent && displayedLines.length === 0" description="暂无日志内容" />
        </div>

        <el-empty v-else description="请选择一个日志文件" />

        <!-- AI 分析结果 -->
        <el-drawer v-model="showAnalysis" title="AI 日志分析" size="40%">
          <div class="analysis-content" v-if="analysisResult">
            <div class="analysis-section" v-if="analysisResult.summary">
              <h4>概要</h4>
              <p>{{ analysisResult.summary }}</p>
            </div>
            <div class="analysis-section" v-if="analysisResult.errors?.length">
              <h4>发现的错误 ({{ analysisResult.errors.length }})</h4>
              <div v-for="(error, i) in analysisResult.errors" :key="i" class="error-item">
                <div class="error-header">
                  <el-tag type="danger" size="small">错误</el-tag>
                  <span class="error-time">{{ error.time }}</span>
                </div>
                <pre class="error-message">{{ error.message }}</pre>
                <div class="error-suggestion" v-if="error.suggestion">
                  <strong>建议：</strong>{{ error.suggestion }}
                </div>
              </div>
            </div>
            <div class="analysis-section" v-if="analysisResult.patterns?.length">
              <h4>检测到的模式</h4>
              <div v-for="(pattern, i) in analysisResult.patterns" :key="i" class="pattern-item">
                <el-tag :type="pattern.severity === 'high' ? 'danger' : pattern.severity === 'medium' ? 'warning' : 'info'" size="small">
                  {{ pattern.severity }}
                </el-tag>
                <span>{{ pattern.description }}</span>
                <span class="pattern-count">出现 {{ pattern.count }} 次</span>
              </div>
            </div>
            <div class="analysis-section" v-if="analysisResult.recommendations?.length">
              <h4>优化建议</h4>
              <ul>
                <li v-for="(rec, i) in analysisResult.recommendations" :key="i">{{ rec }}</li>
              </ul>
            </div>
          </div>
          <div v-else class="analysis-loading">
            <el-icon class="is-loading" :size="32"><Loading /></el-icon>
            <p>正在分析日志...</p>
          </div>
        </el-drawer>
      </div>
    </div>

    <el-empty v-else description="请先选择一个已连接的服务器" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { Refresh, Search, Document, VideoPlay, MagicStick, Loading } from '@element-plus/icons-vue'

interface LogFile {
  name: string
  path: string
  size: number
}

interface LogLine {
  lineNumber: number
  content: string
  level?: string
}

interface AnalysisResult {
  summary: string
  errors: Array<{ time: string; message: string; suggestion?: string }>
  patterns: Array<{ description: string; count: number; severity: string }>
  recommendations: string[]
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const loading = ref(false)
const loadingContent = ref(false)
const analyzing = ref(false)
const logFiles = ref<LogFile[]>([])
const selectedFile = ref<LogFile | null>(null)
const logContent = ref<string[]>([])
const filteredContent = ref<LogLine[]>([])
const fileSearch = ref('')
const searchQuery = ref('')
const logLevel = ref('')
const customPath = ref('')
const isTailing = ref(false)
const showAnalysis = ref(false)
const analysisResult = ref<AnalysisResult | null>(null)
const logViewer = ref<HTMLElement | null>(null)

let tailCleanup: (() => void) | null = null

const connectedServers = computed(() => serverStore.connectedServers)

const filteredLogFiles = computed(() => {
  if (!fileSearch.value) return logFiles.value
  const query = fileSearch.value.toLowerCase()
  return logFiles.value.filter(f => f.name.toLowerCase().includes(query) || f.path.toLowerCase().includes(query))
})

const displayedLines = computed(() => filteredContent.value.slice(-1000))

const logStats = computed(() => {
  const stats = { total: logContent.value.length, errors: 0, warnings: 0, info: 0 }
  logContent.value.forEach(line => {
    const lower = line.toLowerCase()
    if (lower.includes('error') || lower.includes('exception') || lower.includes('fatal')) stats.errors++
    else if (lower.includes('warn')) stats.warnings++
    else if (lower.includes('info')) stats.info++
  })
  return stats
})

// 常见日志路径
const commonLogPaths = [
  '/var/log/syslog', '/var/log/messages', '/var/log/auth.log', '/var/log/secure',
  '/var/log/nginx/access.log', '/var/log/nginx/error.log',
  '/var/log/apache2/access.log', '/var/log/apache2/error.log',
  '/var/log/mysql/error.log', '/var/log/postgresql/postgresql-*.log',
  '/var/log/docker.log', '/var/log/cron', '/var/log/boot.log'
]

if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  loadLogFiles()
}

async function loadLogFiles() {
  if (!selectedServer.value) return
  loading.value = true
  logFiles.value = []

  try {
    // 检查常见日志文件
    for (const path of commonLogPaths) {
      const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `ls -la ${path} 2>/dev/null | awk '{print $5, $9}'`])
      if (result.exit_code === 0 && result.stdout?.trim()) {
        const lines = result.stdout.trim().split('\n')
        for (const line of lines) {
          const [size, filePath] = line.split(' ')
          if (filePath && size) {
            logFiles.value.push({ name: filePath.split('/').pop() || filePath, path: filePath, size: parseInt(size) || 0 })
          }
        }
      }
    }

    // 列出 /var/log 目录
    const listResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', "find /var/log -maxdepth 2 -type f -name '*.log' 2>/dev/null | head -50"])
    if (listResult.exit_code === 0 && listResult.stdout) {
      const paths = listResult.stdout.trim().split('\n').filter(Boolean)
      for (const path of paths) {
        if (!logFiles.value.find(f => f.path === path)) {
          const sizeResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'stat', ['-c', '%s', path])
          logFiles.value.push({ name: path.split('/').pop() || path, path, size: parseInt(sizeResult.stdout?.trim() || '0') })
        }
      }
    }

    // 去重
    logFiles.value = [...new Map(logFiles.value.map(f => [f.path, f])).values()]
  } catch (error) {
    ElMessage.error('加载日志文件失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

async function selectLogFile(file: LogFile) {
  if (isTailing.value) stopTail()
  selectedFile.value = file
  await loadLogContent()
}

async function loadLogContent() {
  if (!selectedServer.value || !selectedFile.value) return
  loadingContent.value = true

  try {
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'tail', ['-n', '500', selectedFile.value.path])
    if (result.exit_code === 0) {
      logContent.value = (result.stdout || '').split('\n')
      filterLogContent()
    } else {
      ElMessage.error(result.stderr || '读取日志失败')
    }
  } catch (error) {
    ElMessage.error('读取日志失败: ' + (error as Error).message)
  } finally {
    loadingContent.value = false
  }
}

function filterLogContent() {
  let lines = logContent.value.map((content, i) => ({ lineNumber: i + 1, content, level: detectLevel(content) }))

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    lines = lines.filter(l => l.content.toLowerCase().includes(query))
  }

  if (logLevel.value) {
    lines = lines.filter(l => l.level === logLevel.value)
  }

  filteredContent.value = lines
}

function detectLevel(line: string): string {
  const lower = line.toLowerCase()
  if (lower.includes('error') || lower.includes('exception') || lower.includes('fatal')) return 'error'
  if (lower.includes('warn')) return 'warn'
  if (lower.includes('info')) return 'info'
  if (lower.includes('debug')) return 'debug'
  return ''
}

function getLineClass(line: LogLine): string {
  return line.level ? `level-${line.level}` : ''
}

function highlightLine(content: string): string {
  let html = content.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  if (searchQuery.value) {
    const regex = new RegExp(`(${searchQuery.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    html = html.replace(regex, '<mark>$1</mark>')
  }
  // 高亮时间戳
  html = html.replace(/(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2})/g, '<span class="timestamp">$1</span>')
  // 高亮 IP 地址
  html = html.replace(/(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/g, '<span class="ip">$1</span>')
  return html
}

async function loadCustomLog() {
  if (!customPath.value.trim()) return
  const file: LogFile = { name: customPath.value.split('/').pop() || customPath.value, path: customPath.value.trim(), size: 0 }
  await selectLogFile(file)
}

function toggleTail() {
  if (isTailing.value) {
    stopTail()
  } else {
    startTail()
  }
}

async function startTail() {
  if (!selectedServer.value || !selectedFile.value) return
  isTailing.value = true

  try {
    await window.electronAPI.log.tail(selectedServer.value, selectedFile.value.path, 100, true)
    tailCleanup = window.electronAPI.log.onData(selectedFile.value.path, (data: { line: string }) => {
      if (data.line) {
        logContent.value.push(data.line)
        if (logContent.value.length > 2000) logContent.value.shift()
        filterLogContent()
        scrollToBottom()
      }
    })
  } catch (error) {
    ElMessage.error('启动实时日志失败')
    isTailing.value = false
  }
}

function stopTail() {
  if (tailCleanup) {
    tailCleanup()
    tailCleanup = null
  }
  if (selectedServer.value && selectedFile.value) {
    window.electronAPI.log.stop(selectedServer.value, selectedFile.value.path)
  }
  isTailing.value = false
}

function scrollToBottom() {
  if (logViewer.value) {
    logViewer.value.scrollTop = logViewer.value.scrollHeight
  }
}

async function analyzeWithAI() {
  if (!selectedServer.value || logContent.value.length === 0) return
  analyzing.value = true
  showAnalysis.value = true
  analysisResult.value = null

  try {
    const logsToAnalyze = logContent.value.slice(-200).join('\n')
    const prompt = `请分析以下服务器日志，找出错误、警告和异常模式，并给出优化建议。以 JSON 格式返回结果，包含 summary、errors、patterns、recommendations 字段。

日志内容：
${logsToAnalyze}`

    const response = await window.electronAPI.ai.chat(prompt, { serverId: selectedServer.value })

    // 尝试解析 JSON 响应
    try {
      const jsonMatch = response.match(/\{[\s\S]*\}/)
      if (jsonMatch) {
        analysisResult.value = JSON.parse(jsonMatch[0])
      } else {
        analysisResult.value = { summary: response, errors: [], patterns: [], recommendations: [] }
      }
    } catch {
      analysisResult.value = { summary: response, errors: [], patterns: [], recommendations: [] }
    }
  } catch (error) {
    ElMessage.error('AI 分析失败: ' + (error as Error).message)
    showAnalysis.value = false
  } finally {
    analyzing.value = false
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

onUnmounted(() => {
  stopTail()
})
</script>

<style lang="scss" scoped>
.log-analysis-page { height: 100%; display: flex; flex-direction: column; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px;
  .header-left { h1 { font-size: 24px; font-weight: 600; margin-bottom: 4px; } .subtitle { color: var(--text-secondary); font-size: 14px; } }
  .header-right { display: flex; gap: 12px; .server-select { width: 200px; } }
}
.main-content { flex: 1; display: flex; gap: 16px; min-height: 0; }
.log-files-panel { width: 280px; background: var(--bg-secondary); border-radius: 12px; padding: 16px; display: flex; flex-direction: column;
  .panel-header { margin-bottom: 12px; h3 { font-size: 16px; font-weight: 600; margin-bottom: 8px; } }
  .file-list { flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 4px;
    .file-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 6px; cursor: pointer; transition: all 0.2s;
      &:hover { background: var(--bg-tertiary); }
      &.active { background: var(--primary-color); color: white; }
      .file-info { flex: 1; display: flex; flex-direction: column; overflow: hidden;
        .file-name { font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .file-size { font-size: 11px; opacity: 0.7; }
      }
    }
  }
  .custom-path { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--border-color); }
}
.log-content-panel { flex: 1; background: var(--bg-secondary); border-radius: 12px; display: flex; flex-direction: column; min-width: 0;
  .content-header { padding: 12px 16px; border-bottom: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px;
    .file-title { display: flex; align-items: center; gap: 8px; font-weight: 500; }
    .content-actions { display: flex; gap: 8px; align-items: center;
      .search-input { width: 200px; }
      .level-select { width: 120px; }
    }
  }
  .log-stats { display: flex; gap: 16px; padding: 8px 16px; background: var(--bg-tertiary); border-bottom: 1px solid var(--border-color);
    .stat-item { display: flex; align-items: center; gap: 6px; font-size: 12px;
      .stat-label { color: var(--text-secondary); }
      .stat-value { font-weight: 600; }
      &.error .stat-value { color: #ef4444; }
      &.warn .stat-value { color: #f59e0b; }
      &.info .stat-value { color: #3b82f6; }
    }
  }
  .log-viewer { flex: 1; overflow: auto; padding: 12px; font-family: 'Fira Code', monospace; font-size: 12px; background: #1e1e1e; color: #d4d4d4;
    .log-line { display: flex; padding: 2px 0; line-height: 1.5;
      &.level-error { background: rgba(239, 68, 68, 0.1); color: #f87171; }
      &.level-warn { background: rgba(245, 158, 11, 0.1); color: #fbbf24; }
      &.level-info { color: #60a5fa; }
      &.level-debug { color: #9ca3af; }
      .line-number { width: 50px; text-align: right; padding-right: 12px; color: #6b7280; user-select: none; flex-shrink: 0; }
      .line-content { flex: 1; white-space: pre-wrap; word-break: break-all;
        :deep(mark) { background: #fbbf24; color: #1e1e1e; padding: 0 2px; border-radius: 2px; }
        :deep(.timestamp) { color: #a78bfa; }
        :deep(.ip) { color: #34d399; }
      }
    }
    .loading-indicator { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 20px; color: var(--text-secondary); }
  }
}
.analysis-content { padding: 0 4px;
  .analysis-section { margin-bottom: 24px;
    h4 { font-size: 16px; font-weight: 600; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--border-color); }
    p { line-height: 1.6; color: var(--text-secondary); }
    ul { padding-left: 20px; li { margin-bottom: 8px; line-height: 1.5; } }
  }
  .error-item { background: var(--bg-tertiary); border-radius: 8px; padding: 12px; margin-bottom: 12px;
    .error-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; .error-time { font-size: 12px; color: var(--text-secondary); } }
    .error-message { background: #1e1e1e; color: #f87171; padding: 8px; border-radius: 4px; font-size: 12px; margin: 8px 0; overflow-x: auto; }
    .error-suggestion { font-size: 13px; color: var(--text-secondary); padding: 8px; background: rgba(34, 197, 94, 0.1); border-radius: 4px; border-left: 3px solid #22c55e; }
  }
  .pattern-item { display: flex; align-items: center; gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--border-color);
    &:last-child { border-bottom: none; }
    .pattern-count { margin-left: auto; font-size: 12px; color: var(--text-secondary); }
  }
}
.analysis-loading { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 200px; color: var(--text-secondary); }
</style>
