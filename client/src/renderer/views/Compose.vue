<template>
  <div class="compose-page">
    <div class="page-header">
      <h1>Compose 项目</h1>
      <div class="header-actions">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-button @click="refresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <el-table :data="projects" v-loading="loading" class="compose-table">
        <el-table-column prop="name" label="项目名称" min-width="150" />
        <el-table-column label="状态" width="150">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config_files" label="配置文件" min-width="250" show-overflow-tooltip />
        <el-table-column label="服务数" width="100">
          <template #default="{ row }">
            {{ row.services?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button-group size="small">
              <el-button type="primary" @click="viewProject(row)">
                详情
              </el-button>
              <el-button type="success" @click="composeAction(row, 'up')">
                启动
              </el-button>
              <el-button type="warning" @click="composeAction(row, 'restart')">
                重启
              </el-button>
              <el-button type="danger" @click="composeAction(row, 'down')">
                停止
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 项目详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      :title="`项目: ${currentProject?.name}`"
      width="70%"
      top="5vh"
    >
      <div v-if="currentProject" class="project-detail">
        <div class="detail-header">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="项目名称">{{ currentProject.name }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getStatusType(currentProject.status)">{{ currentProject.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="配置文件" :span="2">{{ currentProject.config_files }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <h4>服务列表</h4>
        <el-table :data="currentProject.services" size="small">
          <el-table-column prop="name" label="服务名" width="150" />
          <el-table-column prop="image" label="镜像" min-width="200" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="getServiceStatusType(row.state)" size="small">
                {{ row.state || row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="health" label="健康状态" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.health" :type="getHealthType(row.health)" size="small">
                {{ row.health }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="ports" label="端口" width="150" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 操作输出对话框 -->
    <el-dialog
      v-model="showOutputDialog"
      :title="outputTitle"
      width="60%"
      :close-on-click-modal="false"
    >
      <div class="output-container">
        <pre class="output-content">{{ outputContent }}</pre>
      </div>
      <template #footer>
        <el-button @click="showOutputDialog = false" :disabled="isRunning">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

interface ComposeProject {
  name: string
  status: string
  config_files: string
  services: ComposeService[]
}

interface ComposeService {
  name: string
  image: string
  status: string
  state: string
  health: string
  ports: string
  replicas: string
  exit_code: number
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(serverStore.currentServerId)
const projects = ref<ComposeProject[]>([])
const loading = ref(false)
const showDetailDialog = ref(false)
const showOutputDialog = ref(false)
const currentProject = ref<ComposeProject | null>(null)
const outputContent = ref('')
const outputTitle = ref('')
const isRunning = ref(false)

const connectedServers = computed(() => serverStore.connectedServers)

watch(selectedServer, (newVal) => {
  if (newVal) loadProjects()
}, { immediate: true })

async function loadProjects() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.compose.list(selectedServer.value)
    projects.value = result.projects || []
  } catch (error) {
    ElMessage.error(`加载 Compose 项目失败: ${(error as Error).message}`)
  } finally {
    loading.value = false
  }
}

function refresh() {
  loadProjects()
}

async function viewProject(project: ComposeProject) {
  if (!selectedServer.value) return
  try {
    const detail = await window.electronAPI.compose.get(selectedServer.value, project.name)
    currentProject.value = detail
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error(`获取项目详情失败: ${(error as Error).message}`)
  }
}

async function composeAction(project: ComposeProject, action: 'up' | 'down' | 'restart') {
  if (!selectedServer.value) return

  const actionNames: Record<string, string> = {
    up: '启动',
    down: '停止',
    restart: '重启'
  }

  outputTitle.value = `${actionNames[action]} ${project.name}`
  outputContent.value = ''
  showOutputDialog.value = true
  isRunning.value = true

  try {
    const options = { project_path: project.config_files }
    const onOutput = (line: string) => {
      outputContent.value += line + '\n'
    }

    switch (action) {
      case 'up':
        await window.electronAPI.compose.up(selectedServer.value, { ...options, detach: true }, onOutput)
        break
      case 'down':
        await window.electronAPI.compose.down(selectedServer.value, options, onOutput)
        break
      case 'restart':
        await window.electronAPI.compose.restart(selectedServer.value, options, onOutput)
        break
    }

    ElMessage.success(`${actionNames[action]}成功`)
    await loadProjects()
  } catch (error) {
    ElMessage.error(`${actionNames[action]}失败: ${(error as Error).message}`)
    outputContent.value += `\n错误: ${(error as Error).message}`
  } finally {
    isRunning.value = false
  }
}

function getStatusType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  if (status?.includes('running')) return 'success'
  if (status?.includes('exited') || status?.includes('stopped')) return 'danger'
  if (status?.includes('paused')) return 'warning'
  return 'info'
}

function getServiceStatusType(state: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (state?.toLowerCase()) {
    case 'running': return 'success'
    case 'exited':
    case 'dead': return 'danger'
    case 'paused': return 'warning'
    case 'created':
    case 'restarting': return 'info'
    default: return 'info'
  }
}

function getHealthType(health: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (health?.toLowerCase()) {
    case 'healthy': return 'success'
    case 'unhealthy': return 'danger'
    case 'starting': return 'warning'
    default: return 'info'
  }
}
</script>

<style lang="scss" scoped>
.compose-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  h1 {
    font-size: 24px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;

    .server-select {
      width: 200px;
    }
  }
}

.compose-table {
  flex: 1;
}

.empty-state {
  padding: 60px 0;
}

.project-detail {
  .detail-header {
    margin-bottom: 20px;
  }

  h4 {
    margin: 16px 0 12px;
    font-size: 14px;
    font-weight: 600;
  }
}

.output-container {
  background-color: #1e1e1e;
  border-radius: 4px;
  padding: 12px;
  max-height: 400px;
  overflow-y: auto;

  .output-content {
    margin: 0;
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 12px;
    line-height: 1.5;
    color: #d4d4d4;
    white-space: pre-wrap;
    word-break: break-all;
  }
}
</style>
