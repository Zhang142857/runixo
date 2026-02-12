<template>
  <header class="header">
    <div class="header-left">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-if="currentRoute.meta?.title">
          {{ currentRoute.meta.title }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <div class="header-center">
      <el-input
        v-model="searchQuery"
        placeholder="搜索服务器、容器、文件..."
        prefix-icon="Search"
        clearable
        class="search-input"
        @keyup.enter="handleSearch"
      />
    </div>

    <div class="header-right">
      <!-- 当前服务器选择 -->
      <el-select
        v-if="connectedServers.length > 0"
        v-model="selectedServerId"
        placeholder="选择服务器"
        class="server-select"
        @change="handleServerChange"
      >
        <el-option
          v-for="server in connectedServers"
          :key="server.id"
          :label="server.name"
          :value="server.id"
        >
          <div class="server-option">
            <span class="status-dot connected"></span>
            <span>{{ server.name }}</span>
          </div>
        </el-option>
      </el-select>

      <!-- 版本信息 & 更新 -->
      <el-popover placement="bottom" :width="280" trigger="click">
        <template #reference>
          <el-button class="version-btn" :type="hasUpdate ? 'danger' : 'default'" plain size="small">
            <el-icon v-if="hasUpdate"><Upload /></el-icon>
            <span>v{{ clientVersion }}</span>
            <span v-if="hasUpdate" class="update-dot"></span>
          </el-button>
        </template>
        <div class="version-panel">
          <div class="version-row">
            <span class="version-label">Client</span>
            <span class="version-value">v{{ clientVersion }}</span>
            <el-tag v-if="clientUpdateVersion" size="small" type="danger">{{ clientUpdateVersion }} 可用</el-tag>
            <el-tag v-else size="small" type="success">最新</el-tag>
          </div>
          <div class="version-row">
            <span class="version-label">Agent</span>
            <span class="version-value">{{ agentVersion || '未连接' }}</span>
            <el-tag v-if="agentUpdateVersion" size="small" type="danger">{{ agentUpdateVersion }} 可用</el-tag>
            <el-tag v-else-if="agentVersion" size="small" type="success">最新</el-tag>
          </div>
          <el-button
            style="width: 100%; margin-top: 10px"
            :loading="checking"
            @click="checkAllUpdates"
          >
            检查更新
          </el-button>
        </div>
      </el-popover>

      <!-- AI 助手按钮 -->
      <el-tooltip content="AI 助手 (Ctrl+K)" placement="bottom">
        <el-button
          circle
          class="ai-button"
          @click="toggleAI"
        >
          <el-icon><ChatDotRound /></el-icon>
        </el-button>
      </el-tooltip>

      <!-- 通知 -->
      <el-badge :value="notifications" :hidden="notifications === 0" class="notification-badge">
        <el-button circle>
          <el-icon><Bell /></el-icon>
        </el-button>
      </el-badge>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useServerStore } from '@/stores/server'
import { ChatDotRound, Bell, Upload } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const serverStore = useServerStore()

const searchQuery = ref('')
const notifications = ref(0)
const clientVersion = ref('0.0.0')
const agentVersion = ref('')
const clientUpdateVersion = ref('')
const agentUpdateVersion = ref('')
const checking = ref(false)

const hasUpdate = computed(() => !!clientUpdateVersion.value || !!agentUpdateVersion.value)

const currentRoute = computed(() => route)
const connectedServers = computed(() => serverStore.connectedServers)
const selectedServerId = computed({
  get: () => serverStore.currentServerId,
  set: (val) => serverStore.setCurrentServer(val)
})

watch(connectedServers, (servers) => {
  if (servers.length > 0 && !serverStore.currentServerId) {
    serverStore.autoSelectServer()
  }
}, { immediate: true })

// 当前服务器变化时获取 agent 版本
watch(selectedServerId, async (id) => {
  if (id) {
    try {
      const result = await window.electronAPI.server.checkUpdate(id)
      agentVersion.value = result?.currentVersion || result?.version || ''
      agentUpdateVersion.value = result?.hasUpdate ? result.latestVersion : ''
    } catch {
      agentVersion.value = ''
    }
  } else {
    agentVersion.value = ''
    agentUpdateVersion.value = ''
  }
}, { immediate: true })

async function checkAllUpdates() {
  checking.value = true
  try {
    // 检查 client 更新
    const clientResult = await window.electronAPI.updater.check()
    clientUpdateVersion.value = clientResult.available ? clientResult.version || '' : ''

    // 检查 agent 更新
    if (selectedServerId.value) {
      const agentResult = await window.electronAPI.server.checkUpdate(selectedServerId.value)
      agentVersion.value = agentResult?.currentVersion || agentResult?.version || agentVersion.value
      agentUpdateVersion.value = agentResult?.hasUpdate ? agentResult.latestVersion : ''
    }

    if (!clientUpdateVersion.value && !agentUpdateVersion.value) {
      ElMessage.success('所有组件均为最新版本')
    } else {
      ElMessage.warning('发现可用更新')
    }
  } catch {
    ElMessage.error('检查更新失败')
  } finally {
    checking.value = false
  }
}

const emit = defineEmits(['toggle-ai'])

function handleSearch() {
  console.log('Search:', searchQuery.value)
}

function handleServerChange(serverId: string) {
  serverStore.setCurrentServer(serverId)
}

function toggleAI() {
  emit('toggle-ai')
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    toggleAI()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeydown)
  try {
    clientVersion.value = await window.electronAPI.app.getVersion()
  } catch {
    clientVersion.value = '0.0.0'
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style lang="scss" scoped>
.header {
  height: var(--header-height);
  padding: 0 var(--space-5);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  flex: 0 0 auto;
}

.header-center {
  flex: 1;
  max-width: 480px;
  margin: 0 var(--space-6);

  .search-input {
    :deep(.el-input__wrapper) {
      background-color: var(--bg-tertiary);
      border-radius: var(--radius-lg);
    }
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-3);

  .server-select {
    width: 160px;

    :deep(.el-input__wrapper) {
      background-color: var(--bg-tertiary);
    }
  }

  .server-option {
    display: flex;
    align-items: center;
    gap: var(--space-2);

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;

      &.connected {
        background-color: var(--success-color);
        box-shadow: 0 0 6px var(--success-color);
      }
    }
  }

  .version-btn {
    font-size: 12px;
    position: relative;

    .update-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: var(--danger-color);
      display: inline-block;
      margin-left: 4px;
      animation: pulse 2s infinite;
    }
  }

  .ai-button {
    background: var(--primary-gradient);
    border: none;
    color: white;
    transition: all var(--transition-fast);

    &:hover {
      opacity: 0.9;
      transform: scale(1.05);
    }
  }

  .notification-badge {
    :deep(.el-badge__content) {
      background-color: var(--danger-color);
    }
  }
}

.version-panel {
  .version-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 0;

    .version-label {
      font-weight: 600;
      min-width: 48px;
    }

    .version-value {
      color: var(--text-secondary);
      flex: 1;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>
