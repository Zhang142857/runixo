<template>
  <div class="terminal-page">
    <div class="terminal-header">
      <div class="header-left">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-button @click="addTab" :disabled="!selectedServer" size="small">
          <el-icon><Plus /></el-icon>
          新建标签
        </el-button>
      </div>
      <div class="header-right">
        <el-dropdown trigger="click" @command="handleQuickCommand">
          <el-button size="small">
            <el-icon><Lightning /></el-icon>
            快捷命令
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="cmd in quickCommands" :key="cmd.command" :command="cmd.command">
                <span class="quick-cmd-name">{{ cmd.name }}</span>
                <span class="quick-cmd-desc">{{ cmd.command }}</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" @command="changeTheme">
          <el-button size="small">
            <el-icon><Brush /></el-icon>
            主题
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="(theme, name) in themes"
                :key="name"
                :command="name"
                :class="{ active: currentTheme === name }"
              >
                {{ theme.label }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button-group size="small">
          <el-button @click="decreaseFontSize" :disabled="fontSize <= 10">
            <el-icon><Minus /></el-icon>
          </el-button>
          <el-button disabled style="min-width: 50px">{{ fontSize }}px</el-button>
          <el-button @click="increaseFontSize" :disabled="fontSize >= 24">
            <el-icon><Plus /></el-icon>
          </el-button>
        </el-button-group>
        <el-button @click="clearTerminal" size="small">
          <el-icon><Delete /></el-icon>
          清屏
        </el-button>
        <el-button @click="copySelection" size="small">
          <el-icon><CopyDocument /></el-icon>
          复制
        </el-button>
        <el-button @click="pasteFromClipboard" size="small">
          <el-icon><DocumentCopy /></el-icon>
          粘贴
        </el-button>
        <el-button @click="reconnect" :disabled="!selectedServer" size="small">
          <el-icon><Refresh /></el-icon>
          重连
        </el-button>
      </div>
    </div>

    <!-- 标签栏 -->
    <div class="tabs-bar" v-if="tabs.length > 0">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: activeTabId === tab.id }"
        @click="switchTab(tab.id)"
      >
        <span class="tab-title">{{ tab.title }}</span>
        <el-icon class="tab-close" @click.stop="closeTab(tab.id)" v-if="tabs.length > 1">
          <Close />
        </el-icon>
      </div>
    </div>

    <!-- 命令历史面板 -->
    <div class="history-panel" v-if="showHistory">
      <div class="history-header">
        <span>命令历史</span>
        <el-button text size="small" @click="showHistory = false">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
      <div class="history-list">
        <div
          v-for="(cmd, index) in commandHistory"
          :key="index"
          class="history-item"
          @click="executeHistoryCommand(cmd)"
        >
          <span class="history-cmd">{{ cmd }}</span>
          <span class="history-index">#{{ commandHistory.length - index }}</span>
        </div>
      </div>
    </div>

    <div class="terminal-container" ref="terminalContainer"></div>

    <!-- 状态栏 -->
    <div class="status-bar">
      <div class="status-left">
        <span class="status-item" :class="{ connected: activeTabConnected }">
          <el-icon><Connection /></el-icon>
          {{ activeTabConnected ? '已连接' : (selectedServer ? '连接中...' : '未连接') }}
        </span>
        <span class="status-item" v-if="selectedServer">
          {{ getServerName(selectedServer) }}
        </span>
      </div>
      <div class="status-right">
        <span class="status-item clickable" @click="showHistory = !showHistory">
          <el-icon><Clock /></el-icon>
          历史 ({{ commandHistory.length }})
        </span>
        <span class="status-item">
          行: {{ terminalRows }} 列: {{ terminalCols }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { WebglAddon } from 'xterm-addon-webgl'
import 'xterm/css/xterm.css'
import {
  Refresh, Plus, Close, Delete, Minus, Brush,
  Lightning, Connection, Clock, CopyDocument, DocumentCopy
} from '@element-plus/icons-vue'

interface TerminalTab {
  id: string
  title: string
  terminal: Terminal | null
  fitAddon: FitAddon | null
  webglAddon: WebglAddon | null
  sessionId: string
  connected: boolean
  cleanupFns: Array<() => void>
}

const route = useRoute()
const serverStore = useServerStore()

const terminalContainer = ref<HTMLElement | null>(null)
const selectedServer = ref<string | null>(null)
const tabs = ref<TerminalTab[]>([])
const activeTabId = ref<string>('')
const fontSize = ref(14)
const currentTheme = ref('dark')
const showHistory = ref(false)
const commandHistory = ref<string[]>([])
const terminalRows = ref(24)
const terminalCols = ref(80)

const connectedServers = computed(() => serverStore.connectedServers)

// 当前活跃标签的连接状态
const activeTabConnected = computed(() => {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  return activeTab?.connected ?? false
})

// 主题配置
const themes: Record<string, { label: string; background: string; foreground: string; cursor: string }> = {
  dark: { label: '深色', background: '#0f0f1a', foreground: '#e4e4e7', cursor: '#6366f1' },
  light: { label: '浅色', background: '#ffffff', foreground: '#1f2937', cursor: '#6366f1' },
  monokai: { label: 'Monokai', background: '#272822', foreground: '#f8f8f2', cursor: '#f8f8f0' },
  dracula: { label: 'Dracula', background: '#282a36', foreground: '#f8f8f2', cursor: '#f8f8f2' }
}

// 快捷命令
const quickCommands = [
  { name: '系统信息', command: 'uname -a' },
  { name: '磁盘使用', command: 'df -h' },
  { name: '内存使用', command: 'free -h' },
  { name: '进程列表', command: 'ps aux | head -20' },
  { name: '网络连接', command: 'netstat -tuln' },
  { name: '当前目录', command: 'pwd && ls -la' },
  { name: 'Docker 容器', command: 'docker ps' },
  { name: '系统日志', command: 'tail -50 /var/log/syslog' }
]

// 从路由参数获取服务器ID
if (route.params.serverId) {
  selectedServer.value = route.params.serverId as string
} else if (serverStore.currentServerId) {
  selectedServer.value = serverStore.currentServerId
}

watch(selectedServer, (newVal) => {
  if (newVal && tabs.value.length === 0) {
    addTab()
  }
})

function getServerName(serverId: string): string {
  const server = connectedServers.value.find(s => s.id === serverId)
  return server?.name || serverId
}

function generateTabId(): string {
  return Math.random().toString(36).substring(2, 10)
}

function addTab() {
  const id = generateTabId()
  const sessionId = generateTabId() // 用于 gRPC shell 会话
  const tab: TerminalTab = {
    id,
    title: `终端 ${tabs.value.length + 1}`,
    terminal: null,
    fitAddon: null,
    webglAddon: null,
    sessionId,
    connected: false,
    cleanupFns: []
  }
  tabs.value.push(tab)
  activeTabId.value = id
  setTimeout(() => initTerminal(tab), 50)
}

function closeTab(tabId: string) {
  const index = tabs.value.findIndex(t => t.id === tabId)
  if (index === -1) return
  const tab = tabs.value[index]

  // 清理事件监听器
  tab.cleanupFns.forEach(fn => fn())

  // 停止 shell 会话
  if (tab.connected && selectedServer.value) {
    window.electronAPI.terminal.stop(selectedServer.value, tab.sessionId)
  }

  if (tab.terminal) tab.terminal.dispose()
  tabs.value.splice(index, 1)
  if (activeTabId.value === tabId && tabs.value.length > 0) {
    activeTabId.value = tabs.value[Math.max(0, index - 1)].id
  }
}

function switchTab(tabId: string) {
  if (activeTabId.value === tabId) return
  activeTabId.value = tabId
  const tab = tabs.value.find(t => t.id === tabId)
  if (tab) showTerminal(tab)
}

function showTerminal(tab: TerminalTab) {
  tabs.value.forEach(t => {
    if (t.terminal?.element) {
      t.terminal.element.style.display = t.id === tab.id ? 'block' : 'none'
    }
  })
  if (tab.fitAddon) setTimeout(() => tab.fitAddon?.fit(), 10)
}

function initTerminal(tab: TerminalTab) {
  if (!terminalContainer.value) return

  const theme = themes[currentTheme.value]

  tab.terminal = new Terminal({
    theme: {
      background: theme.background,
      foreground: theme.foreground,
      cursor: theme.cursor,
      cursorAccent: theme.background,
      selectionBackground: theme.cursor + '80'
    },
    fontFamily: '"Fira Code", "Cascadia Code", Consolas, monospace',
    fontSize: fontSize.value,
    lineHeight: 1.2,
    cursorBlink: true,
    cursorStyle: 'bar',
    scrollback: 10000,
    allowProposedApi: true,
    rightClickSelectsWord: true
  })

  tab.fitAddon = new FitAddon()
  tab.terminal.loadAddon(tab.fitAddon)
  tab.terminal.loadAddon(new WebLinksAddon())

  tab.terminal.open(terminalContainer.value)
  
  // 尝试加载 WebGL 渲染器以提升性能
  try {
    tab.webglAddon = new WebglAddon()
    tab.terminal.loadAddon(tab.webglAddon)
  } catch (e) {
    console.warn('WebGL addon failed to load, using canvas renderer')
  }
  
  tab.fitAddon.fit()

  terminalRows.value = tab.terminal.rows
  terminalCols.value = tab.terminal.cols

  window.addEventListener('resize', handleResize)
  connectToServer(tab)
  showTerminal(tab)
}

async function connectToServer(tab: TerminalTab) {
  if (!tab.terminal || !selectedServer.value) return

  tab.terminal.writeln('\x1b[36m正在连接到服务器...\x1b[0m')

  try {
    // 启动 shell 会话
    const result = await window.electronAPI.terminal.start(
      selectedServer.value,
      tab.sessionId,
      tab.terminal.rows,
      tab.terminal.cols
    )

    if (!result.success) {
      tab.terminal.writeln('\x1b[31m连接失败\x1b[0m')
      return
    }

    tab.connected = true
    tab.terminal.writeln('\x1b[32m已连接\x1b[0m')

    // 监听服务器返回的数据
    const cleanupData = window.electronAPI.terminal.onData(tab.sessionId, (data: string) => {
      tab.terminal?.write(data)
    })
    tab.cleanupFns.push(cleanupData)

    // 监听错误
    const cleanupError = window.electronAPI.terminal.onError(tab.sessionId, (error: string) => {
      tab.terminal?.writeln(`\x1b[31m错误: ${error}\x1b[0m`)
    })
    tab.cleanupFns.push(cleanupError)

    // 监听会话结束
    const cleanupEnd = window.electronAPI.terminal.onEnd(tab.sessionId, () => {
      tab.connected = false
      tab.terminal?.writeln('\x1b[33m会话已断开\x1b[0m')
    })
    tab.cleanupFns.push(cleanupEnd)

    // 发送用户输入到服务器
    tab.terminal.onData((data) => {
      if (tab.connected && selectedServer.value) {
        window.electronAPI.terminal.write(selectedServer.value, tab.sessionId, data)

        // 记录命令历史（当按下回车时）
        if (data === '\r') {
          // 命令历史由服务端 shell 处理，这里只做本地记录
        }
      }
    })

  } catch (error) {
    tab.terminal.writeln(`\x1b[31m连接失败: ${(error as Error).message}\x1b[0m`)
  }
}

function handleResize() {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.fitAddon) {
    activeTab.fitAddon.fit()
    if (activeTab.terminal) {
      terminalRows.value = activeTab.terminal.rows
      terminalCols.value = activeTab.terminal.cols

      // 通知服务器终端大小变化
      if (activeTab.connected && selectedServer.value) {
        window.electronAPI.terminal.resize(
          selectedServer.value,
          activeTab.sessionId,
          activeTab.terminal.rows,
          activeTab.terminal.cols
        )
      }
    }
  }
}

async function reconnect() {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab) {
    // 清理旧的事件监听器
    activeTab.cleanupFns.forEach(fn => fn())
    activeTab.cleanupFns = []

    // 停止旧的 shell 会话
    if (activeTab.connected && selectedServer.value) {
      await window.electronAPI.terminal.stop(selectedServer.value, activeTab.sessionId)
    }

    // 生成新的 sessionId
    activeTab.sessionId = generateTabId()
    activeTab.connected = false

    if (activeTab.terminal) {
      activeTab.terminal.clear()
    }

    // 重新连接
    await connectToServer(activeTab)
  }
}

function clearTerminal() {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.terminal) {
    activeTab.terminal.clear()
    activeTab.terminal.write('\x1b[33m$ \x1b[0m')
  }
}

function increaseFontSize() {
  if (fontSize.value < 24) {
    fontSize.value += 2
    updateAllTerminalsFontSize()
  }
}

function decreaseFontSize() {
  if (fontSize.value > 10) {
    fontSize.value -= 2
    updateAllTerminalsFontSize()
  }
}

function updateAllTerminalsFontSize() {
  tabs.value.forEach(tab => {
    if (tab.terminal) {
      tab.terminal.options.fontSize = fontSize.value
      tab.fitAddon?.fit()
    }
  })
}

function changeTheme(themeName: string) {
  currentTheme.value = themeName
  const theme = themes[themeName]
  tabs.value.forEach(tab => {
    if (tab.terminal) {
      tab.terminal.options.theme = {
        background: theme.background,
        foreground: theme.foreground,
        cursor: theme.cursor,
        cursorAccent: theme.background,
        selectionBackground: theme.cursor + '80'
      }
    }
  })
  // 更新容器背景色
  if (terminalContainer.value) {
    terminalContainer.value.style.backgroundColor = theme.background
  }
}

function handleQuickCommand(command: string) {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.connected && selectedServer.value) {
    // 直接发送命令到 shell（包含回车）
    window.electronAPI.terminal.write(selectedServer.value, activeTab.sessionId, command + '\r')
  }
}

function executeHistoryCommand(command: string) {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.connected && selectedServer.value) {
    // 直接发送命令到 shell（包含回车）
    window.electronAPI.terminal.write(selectedServer.value, activeTab.sessionId, command + '\r')
    showHistory.value = false
  }
}

// 复制选中的文本
async function copySelection() {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.terminal) {
    const selection = activeTab.terminal.getSelection()
    if (selection) {
      await navigator.clipboard.writeText(selection)
      ElMessage.success('已复制到剪贴板')
    } else {
      ElMessage.info('请先选择要复制的文本')
    }
  }
}

// 从剪贴板粘贴
async function pasteFromClipboard() {
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.connected && selectedServer.value) {
    try {
      const text = await navigator.clipboard.readText()
      if (text) {
        window.electronAPI.terminal.write(selectedServer.value, activeTab.sessionId, text)
      }
    } catch (e) {
      ElMessage.error('无法访问剪贴板')
    }
  }
}

onMounted(() => {
  if (selectedServer.value && tabs.value.length === 0) {
    addTab()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  tabs.value.forEach(tab => {
    // 清理事件监听器
    tab.cleanupFns.forEach(fn => fn())

    // 停止 shell 会话
    if (tab.connected && selectedServer.value) {
      window.electronAPI.terminal.stop(selectedServer.value, tab.sessionId)
    }

    if (tab.terminal) tab.terminal.dispose()
  })
})
</script>

<style lang="scss" scoped>
.terminal-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;

    .server-select {
      width: 200px;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.tabs-bar {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  padding: 4px;
  background: var(--bg-secondary);
  border-radius: 6px;

  .tab-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: transparent;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background: var(--bg-tertiary);
    }

    &.active {
      background: var(--primary-color);
      color: white;
    }

    .tab-title {
      font-size: 13px;
    }

    .tab-close {
      font-size: 12px;
      opacity: 0.6;
      cursor: pointer;

      &:hover {
        opacity: 1;
      }
    }
  }
}

.history-panel {
  position: absolute;
  right: 20px;
  top: 120px;
  width: 300px;
  max-height: 400px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 100;
  overflow: hidden;

  .history-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
    font-weight: 600;
  }

  .history-list {
    max-height: 340px;
    overflow-y: auto;
  }

  .history-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: var(--bg-tertiary);
    }

    .history-cmd {
      font-family: monospace;
      font-size: 13px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      max-width: 220px;
    }

    .history-index {
      font-size: 11px;
      color: var(--text-secondary);
    }
  }
}

.terminal-container {
  flex: 1;
  background-color: #0f0f1a;
  border-radius: 8px;
  padding: 12px;
  overflow: hidden;
  position: relative;

  :deep(.xterm) {
    height: 100%;
  }

  :deep(.xterm-viewport) {
    &::-webkit-scrollbar {
      width: 8px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: #2e2e4a;
      border-radius: 4px;
    }
  }
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: 0 0 8px 8px;
  font-size: 12px;
  color: var(--text-secondary);

  .status-left,
  .status-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .status-item {
    display: flex;
    align-items: center;
    gap: 4px;

    &.connected {
      color: #22c55e;
    }

    &.clickable {
      cursor: pointer;
      &:hover {
        color: var(--primary-color);
      }
    }
  }
}

.quick-cmd-name {
  display: block;
  font-weight: 500;
}

.quick-cmd-desc {
  display: block;
  font-size: 11px;
  color: var(--text-secondary);
  font-family: monospace;
}

:deep(.el-dropdown-menu__item.active) {
  background-color: var(--primary-color);
  color: white;
}
</style>
