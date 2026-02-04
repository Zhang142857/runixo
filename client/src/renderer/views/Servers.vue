<template>
  <div class="servers-page">
    <div class="page-header">
      <h1>服务器管理</h1>
      <div class="header-actions">
        <el-select
          v-model="filterGroup"
          placeholder="筛选分组"
          clearable
          class="group-filter"
        >
          <el-option label="全部分组" value="" />
          <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
        </el-select>
        <el-button @click="showGroupDialog = true">
          <el-icon><FolderAdd /></el-icon>
          管理分组
        </el-button>
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon>
          添加服务器
        </el-button>
      </div>
    </div>

    <!-- 分组统计 -->
    <div class="group-stats">
      <div class="stat-item">
        <span class="stat-value">{{ totalServers }}</span>
        <span class="stat-label">总服务器</span>
      </div>
      <div class="stat-item connected">
        <span class="stat-value">{{ connectedCount }}</span>
        <span class="stat-label">已连接</span>
      </div>
      <div class="stat-item disconnected">
        <span class="stat-value">{{ disconnectedCount }}</span>
        <span class="stat-label">未连接</span>
      </div>
      <div class="stat-item groups">
        <span class="stat-value">{{ groups.length }}</span>
        <span class="stat-label">分组数</span>
      </div>
    </div>

    <!-- 服务器分组 -->
    <div v-for="(groupServers, groupName) in filteredServersByGroup" :key="groupName" class="server-group">
      <div class="group-header" v-if="groupServers.length > 0">
        <el-icon><Folder /></el-icon>
        <span class="group-name">{{ groupName }}</span>
        <span class="group-count">{{ groupServers.length }} 台服务器</span>
        <span class="group-connected">{{ getGroupConnectedCount(groupServers) }} 已连接</span>
      </div>

      <div class="server-list">
        <div
          v-for="server in groupServers"
          :key="server.id"
          class="server-item"
          :class="{ selected: selectedServers.includes(server.id) }"
        >
          <el-checkbox
            :model-value="selectedServers.includes(server.id)"
            @change="toggleSelect(server.id)"
          />

          <div class="server-status" :class="server.status"></div>

          <div class="server-info" @click="goToServer(server)">
            <div class="server-name">{{ server.name }}</div>
            <div class="server-host">{{ server.host }}:{{ server.port }}</div>
          </div>

          <div class="server-system" v-if="server.systemInfo">
            <span>{{ server.systemInfo.os }}</span>
            <span>{{ server.systemInfo.arch }}</span>
          </div>

          <div class="server-actions">
            <el-button
              v-if="server.status !== 'connected'"
              type="primary"
              size="small"
              :loading="server.status === 'connecting'"
              @click="connectServer(server)"
            >
              连接
            </el-button>
            <el-button
              v-else
              size="small"
              @click="disconnectServer(server)"
            >
              断开
            </el-button>

            <el-dropdown trigger="click" @command="handleAction($event, server)">
              <el-button size="small" text>
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="terminal">终端</el-dropdown-item>
                  <el-dropdown-item command="files">文件管理</el-dropdown-item>
                  <el-dropdown-item command="edit" divided>编辑</el-dropdown-item>
                  <el-dropdown-item command="delete">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <div v-if="selectedServers.length > 0" class="batch-actions">
      <span>已选择 {{ selectedServers.length }} 台服务器</span>
      <el-button size="small" @click="batchConnect">批量连接</el-button>
      <el-button size="small" @click="batchDisconnect">批量断开</el-button>
      <el-button size="small" @click="showBatchCommand = true">批量执行命令</el-button>
      <el-button size="small" text @click="selectedServers = []">取消选择</el-button>
    </div>

    <!-- 添加服务器对话框 -->
    <el-dialog v-model="showAddDialog" title="添加服务器" width="500px">
      <el-form :model="newServer" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="newServer.name" placeholder="服务器名称" />
        </el-form-item>
        <el-form-item label="主机" required>
          <el-input v-model="newServer.host" placeholder="IP 或域名" />
        </el-form-item>
        <el-form-item label="端口" required>
          <el-input-number v-model="newServer.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="Token" required>
          <el-input v-model="newServer.token" type="password" show-password />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="newServer.group">
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
        </el-form-item>
        <el-form-item label="TLS">
          <el-switch v-model="newServer.useTls" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addServer">添加</el-button>
      </template>
    </el-dialog>

    <!-- 批量命令对话框 -->
    <el-dialog v-model="showBatchCommand" title="批量执行命令" width="800px" :close-on-click-modal="false">
      <div class="batch-command-content">
        <el-input
          v-model="batchCommand"
          type="textarea"
          :rows="3"
          placeholder="输入要执行的命令"
          :disabled="batchExecuting"
        />

        <div class="batch-targets">
          <div class="targets-header">
            <span>目标服务器 ({{ selectedServers.length }} 台)：</span>
            <div class="quick-select">
              <el-dropdown @command="selectByGroup" size="small">
                <el-button size="small">
                  按分组选择 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="__all__">全部已连接</el-dropdown-item>
                    <el-dropdown-item v-for="g in groups" :key="g" :command="g">
                      {{ g }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
          <div class="targets-list">
            <el-tag
              v-for="id in selectedServers"
              :key="id"
              size="small"
              closable
              @close="removeFromSelection(id)"
            >
              {{ getServerName(id) }}
            </el-tag>
          </div>
        </div>

        <!-- 执行结果 -->
        <div v-if="batchResults.length > 0" class="batch-results">
          <div class="results-header">
            <span>执行结果</span>
            <el-tag type="success" size="small">成功: {{ batchSuccessCount }}</el-tag>
            <el-tag type="danger" size="small">失败: {{ batchFailCount }}</el-tag>
          </div>
          <div class="results-list">
            <div
              v-for="result in batchResults"
              :key="result.serverId"
              class="result-item"
              :class="{ success: result.success, error: !result.success }"
            >
              <div class="result-header" @click="result.expanded = !result.expanded">
                <el-icon v-if="result.success"><SuccessFilled /></el-icon>
                <el-icon v-else><CircleCloseFilled /></el-icon>
                <span class="server-name">{{ result.serverName }}</span>
                <span class="exit-code" v-if="result.exitCode !== undefined">
                  退出码: {{ result.exitCode }}
                </span>
                <el-icon class="expand-icon">
                  <ArrowDown v-if="!result.expanded" />
                  <ArrowUp v-else />
                </el-icon>
              </div>
              <div v-if="result.expanded" class="result-output">
                <pre v-if="result.stdout">{{ result.stdout }}</pre>
                <pre v-if="result.stderr" class="stderr">{{ result.stderr }}</pre>
                <pre v-if="result.error" class="error">{{ result.error }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="closeBatchCommand">关闭</el-button>
        <el-button
          type="primary"
          @click="executeBatchCommand"
          :loading="batchExecuting"
          :disabled="!batchCommand.trim() || selectedServers.length === 0"
        >
          {{ batchExecuting ? '执行中...' : '执行' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 分组管理对话框 -->
    <el-dialog v-model="showGroupDialog" title="管理分组" width="500px">
      <div class="group-manager">
        <div class="add-group">
          <el-input
            v-model="newGroupName"
            placeholder="输入新分组名称"
            @keyup.enter="addGroup"
          >
            <template #append>
              <el-button @click="addGroup" :disabled="!newGroupName.trim()">
                <el-icon><Plus /></el-icon>
                添加
              </el-button>
            </template>
          </el-input>
        </div>
        <div class="group-list">
          <div
            v-for="group in groups"
            :key="group"
            class="group-item"
          >
            <el-icon><Folder /></el-icon>
            <span class="group-name">{{ group }}</span>
            <span class="group-server-count">
              {{ getGroupServerCount(group) }} 台服务器
            </span>
            <el-button
              v-if="group !== '默认'"
              type="danger"
              size="small"
              text
              @click="deleteGroup(group)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showGroupDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 编辑服务器对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑服务器" width="500px">
      <el-form :model="editServer" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="editServer.name" placeholder="服务器名称" />
        </el-form-item>
        <el-form-item label="主机" required>
          <el-input v-model="editServer.host" placeholder="IP 或域名" />
        </el-form-item>
        <el-form-item label="端口" required>
          <el-input-number v-model="editServer.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="Token">
          <el-input v-model="editServer.token" type="password" show-password placeholder="留空则不修改" />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="editServer.group" style="width: 100%">
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
        </el-form-item>
        <el-form-item label="TLS">
          <el-switch v-model="editServer.useTls" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEditServer">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useServerStore, type Server } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Folder, FolderAdd, MoreFilled, Delete, ArrowDown, ArrowUp, SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'

const router = useRouter()
const serverStore = useServerStore()

const showAddDialog = ref(false)
const showGroupDialog = ref(false)
const showBatchCommand = ref(false)
const showEditDialog = ref(false)
const selectedServers = ref<string[]>([])
const batchCommand = ref('')
const batchExecuting = ref(false)
const batchResults = ref<BatchResult[]>([])
const filterGroup = ref('')
const newGroupName = ref('')

interface BatchResult {
  serverId: string
  serverName: string
  success: boolean
  exitCode?: number
  stdout?: string
  stderr?: string
  error?: string
  expanded: boolean
}

const newServer = ref({
  name: '',
  host: '',
  port: 9527,
  token: '',
  group: '默认',
  useTls: false
})

const editServer = ref({
  id: '',
  name: '',
  host: '',
  port: 9527,
  token: '',
  group: '默认',
  useTls: false
})

const serversByGroup = computed(() => serverStore.serversByGroup)
const groups = computed(() => serverStore.groups)

// 统计数据
const totalServers = computed(() => serverStore.servers.length)
const connectedCount = computed(() => serverStore.servers.filter(s => s.status === 'connected').length)
const disconnectedCount = computed(() => serverStore.servers.filter(s => s.status !== 'connected').length)

// 批量执行结果统计
const batchSuccessCount = computed(() => batchResults.value.filter(r => r.success).length)
const batchFailCount = computed(() => batchResults.value.filter(r => !r.success).length)

// 按分组筛选
const filteredServersByGroup = computed(() => {
  if (!filterGroup.value) {
    return serversByGroup.value
  }
  return {
    [filterGroup.value]: serversByGroup.value[filterGroup.value] || []
  }
})

function getGroupConnectedCount(servers: Server[]): number {
  return servers.filter(s => s.status === 'connected').length
}

function getGroupServerCount(groupName: string): number {
  return serverStore.servers.filter(s => (s.group || '默认') === groupName).length
}

function addGroup() {
  const name = newGroupName.value.trim()
  if (!name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  if (groups.value.includes(name)) {
    ElMessage.warning('分组已存在')
    return
  }
  serverStore.addGroup(name)
  newGroupName.value = ''
  ElMessage.success('分组已添加')
}

function deleteGroup(name: string) {
  ElMessageBox.confirm(
    `确定要删除分组"${name}"吗？该分组下的服务器将移至"默认"分组。`,
    '确认删除'
  ).then(() => {
    serverStore.removeGroup(name)
    ElMessage.success('分组已删除')
  }).catch(() => {})
}

function openAddDialog() {
  newServer.value = {
    name: '',
    host: '',
    port: 9527,
    token: '',
    group: filterGroup.value || '默认',
    useTls: false
  }
  showAddDialog.value = true
}

function openEditDialog(server: Server) {
  editServer.value = {
    id: server.id,
    name: server.name,
    host: server.host,
    port: server.port,
    token: '',
    group: server.group || '默认',
    useTls: server.useTls
  }
  showEditDialog.value = true
}

function saveEditServer() {
  if (!editServer.value.name || !editServer.value.host) {
    ElMessage.warning('请填写必要信息')
    return
  }

  const updates: Partial<Server> = {
    name: editServer.value.name,
    host: editServer.value.host,
    port: editServer.value.port,
    group: editServer.value.group,
    useTls: editServer.value.useTls
  }

  // 只有填写了新 token 才更新
  if (editServer.value.token) {
    updates.token = editServer.value.token
  }

  serverStore.updateServer(editServer.value.id, updates)
  showEditDialog.value = false
  ElMessage.success('服务器已更新')
}

function toggleSelect(id: string) {
  const index = selectedServers.value.indexOf(id)
  if (index === -1) {
    selectedServers.value.push(id)
  } else {
    selectedServers.value.splice(index, 1)
  }
}

function goToServer(server: any) {
  if (server.status === 'connected') {
    serverStore.setCurrentServer(server.id)
    router.push(`/server/${server.id}`)
  }
}

async function connectServer(server: any) {
  try {
    await serverStore.connectServer(server.id)
    ElMessage.success('连接成功')
  } catch (error) {
    ElMessage.error(`连接失败: ${(error as Error).message}`)
  }
}

async function disconnectServer(server: any) {
  await serverStore.disconnectServer(server.id)
  ElMessage.info('已断开连接')
}

async function addServer() {
  if (!newServer.value.name || !newServer.value.host || !newServer.value.token) {
    ElMessage.warning('请填写必要信息')
    return
  }

  serverStore.addServer({
    name: newServer.value.name,
    host: newServer.value.host,
    port: newServer.value.port,
    token: newServer.value.token,
    group: newServer.value.group,
    useTls: newServer.value.useTls
  })

  showAddDialog.value = false
  ElMessage.success('服务器已添加')
}

function handleAction(action: string, server: Server) {
  switch (action) {
    case 'terminal':
      router.push(`/terminal/${server.id}`)
      break
    case 'files':
      router.push(`/files/${server.id}`)
      break
    case 'edit':
      openEditDialog(server)
      break
    case 'delete':
      ElMessageBox.confirm('确定要删除这个服务器吗？', '确认').then(() => {
        serverStore.removeServer(server.id)
        ElMessage.success('已删除')
      }).catch(() => {})
      break
  }
}

async function batchConnect() {
  for (const id of selectedServers.value) {
    try {
      await serverStore.connectServer(id)
    } catch (e) {
      // 继续连接其他服务器
    }
  }
  ElMessage.success('批量连接完成')
}

async function batchDisconnect() {
  for (const id of selectedServers.value) {
    await serverStore.disconnectServer(id)
  }
  ElMessage.info('已批量断开')
}

function selectByGroup(groupName: string) {
  if (groupName === '__all__') {
    // 选择所有已连接的服务器
    selectedServers.value = serverStore.servers
      .filter(s => s.status === 'connected')
      .map(s => s.id)
  } else {
    // 选择指定分组中已连接的服务器
    selectedServers.value = serverStore.servers
      .filter(s => (s.group || '默认') === groupName && s.status === 'connected')
      .map(s => s.id)
  }
}

function removeFromSelection(id: string) {
  const index = selectedServers.value.indexOf(id)
  if (index !== -1) {
    selectedServers.value.splice(index, 1)
  }
}

function closeBatchCommand() {
  showBatchCommand.value = false
  // 不清空结果，让用户可以查看
}

async function executeBatchCommand() {
  if (!batchCommand.value.trim()) {
    ElMessage.warning('请输入命令')
    return
  }

  if (selectedServers.value.length === 0) {
    ElMessage.warning('请选择目标服务器')
    return
  }

  batchExecuting.value = true
  batchResults.value = []

  // 并行执行命令
  const promises = selectedServers.value.map(async (serverId) => {
    const server = serverStore.servers.find(s => s.id === serverId)
    const result: BatchResult = {
      serverId,
      serverName: server?.name || serverId,
      success: false,
      expanded: false
    }

    try {
      const response = await window.electronAPI.server.executeCommand(serverId, batchCommand.value)
      result.success = response.exit_code === 0
      result.exitCode = response.exit_code
      result.stdout = response.stdout
      result.stderr = response.stderr
    } catch (error) {
      result.success = false
      result.error = (error as Error).message
    }

    return result
  })

  const results = await Promise.all(promises)
  batchResults.value = results

  // 自动展开第一个失败的结果
  const firstFailed = results.find(r => !r.success)
  if (firstFailed) {
    firstFailed.expanded = true
  } else if (results.length > 0) {
    results[0].expanded = true
  }

  batchExecuting.value = false
  ElMessage.success(`命令执行完成: ${batchSuccessCount.value} 成功, ${batchFailCount.value} 失败`)
}

function getServerName(id: string): string {
  const server = serverStore.servers.find(s => s.id === id)
  return server?.name || id
}
</script>

<style lang="scss" scoped>
.servers-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h1 {
    font-size: 24px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 12px;

    .group-filter {
      width: 150px;
    }
  }
}

.group-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;

  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 24px;
    background: var(--bg-secondary);
    border-radius: 8px;
    min-width: 100px;

    .stat-value {
      font-size: 24px;
      font-weight: 600;
    }

    .stat-label {
      font-size: 12px;
      color: var(--text-secondary);
      margin-top: 4px;
    }

    &.connected .stat-value {
      color: var(--success-color);
    }

    &.disconnected .stat-value {
      color: var(--text-secondary);
    }

    &.groups .stat-value {
      color: var(--el-color-primary);
    }
  }
}

.server-group {
  margin-bottom: 24px;

  .group-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    color: var(--text-secondary);

    .group-name {
      font-weight: 500;
      color: var(--text-color);
    }

    .group-count {
      font-size: 12px;
    }

    .group-connected {
      font-size: 12px;
      color: var(--success-color);
      margin-left: 8px;
    }
  }
}

.server-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.server-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  transition: all 0.2s;

  &:hover {
    border-color: var(--primary-color);
  }

  &.selected {
    background-color: var(--bg-tertiary);
    border-color: var(--primary-color);
  }

  .server-status {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: var(--text-secondary);

    &.connected {
      background-color: var(--success-color);
    }

    &.connecting {
      background-color: var(--warning-color);
    }

    &.error {
      background-color: var(--danger-color);
    }
  }

  .server-info {
    flex: 1;
    cursor: pointer;

    .server-name {
      font-weight: 500;
      margin-bottom: 2px;
    }

    .server-host {
      font-size: 12px;
      color: var(--text-secondary);
      font-family: monospace;
    }
  }

  .server-system {
    display: flex;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .server-actions {
    display: flex;
    gap: 8px;
  }
}

.batch-actions {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 24px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.batch-targets {
  margin-top: 16px;

  .targets-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    span {
      color: var(--text-secondary);
    }
  }

  .targets-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
}

.batch-command-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.batch-results {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;

  .results-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--bg-color-overlay);
    border-bottom: 1px solid var(--border-color);

    span:first-child {
      font-weight: 500;
    }
  }

  .results-list {
    max-height: 300px;
    overflow-y: auto;
  }

  .result-item {
    border-bottom: 1px solid var(--border-color);

    &:last-child {
      border-bottom: none;
    }

    .result-header {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 16px;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: var(--bg-color-overlay);
      }

      .server-name {
        flex: 1;
        font-weight: 500;
      }

      .exit-code {
        font-size: 12px;
        color: var(--text-secondary);
      }

      .expand-icon {
        color: var(--text-secondary);
      }
    }

    &.success .result-header .el-icon:first-child {
      color: var(--success-color);
    }

    &.error .result-header .el-icon:first-child {
      color: var(--el-color-danger);
    }

    .result-output {
      padding: 12px 16px;
      background: var(--bg-color-overlay);
      border-top: 1px solid var(--border-color);

      pre {
        margin: 0;
        padding: 8px;
        background: rgba(0, 0, 0, 0.2);
        border-radius: 4px;
        font-size: 12px;
        font-family: 'Fira Code', monospace;
        white-space: pre-wrap;
        word-break: break-all;
        max-height: 200px;
        overflow-y: auto;

        &.stderr {
          color: var(--el-color-warning);
        }

        &.error {
          color: var(--el-color-danger);
        }
      }

      pre + pre {
        margin-top: 8px;
      }
    }
  }
}

.group-manager {
  .add-group {
    margin-bottom: 16px;
  }

  .group-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 300px;
    overflow-y: auto;

    .group-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 16px;
      background: var(--bg-color-overlay);
      border-radius: 8px;

      .el-icon {
        color: var(--el-color-primary);
      }

      .group-name {
        flex: 1;
        font-weight: 500;
      }

      .group-server-count {
        font-size: 12px;
        color: var(--text-secondary);
      }
    }
  }
}
</style>
