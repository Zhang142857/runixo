<template>
  <div class="services-page">
    <div class="page-header">
      <h1>服务管理</h1>
      <div class="header-actions">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-input
          v-model="searchText"
          placeholder="搜索服务..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="filterStatus" placeholder="状态" clearable class="filter-select">
          <el-option v-for="status in statusOptions" :key="status" :label="status" :value="status" />
        </el-select>
        <el-switch v-model="autoRefresh" active-text="自动刷新" />
        <el-button @click="loadServices" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <div class="stats-bar">
        <div class="stats-left">
          <el-tag type="success">运行中: {{ runningCount }}</el-tag>
          <el-tag type="info">已停止: {{ stoppedCount }}</el-tag>
          <el-tag type="danger">失败: {{ failedCount }}</el-tag>
          <el-tag>总计: {{ services.length }}</el-tag>
        </div>
        <div class="stats-right" v-if="selectedServices.length > 0">
          <span class="selected-count">已选择 {{ selectedServices.length }} 个服务</span>
          <el-button-group size="small">
            <el-button type="success" @click="batchAction('start')">批量启动</el-button>
            <el-button type="warning" @click="batchAction('stop')">批量停止</el-button>
            <el-button @click="batchAction('restart')">批量重启</el-button>
          </el-button-group>
          <el-button size="small" @click="selectedServices = []">取消选择</el-button>
        </div>
      </div>

      <el-table :data="filteredServices" v-loading="loading" class="services-table" row-key="name">
        <el-table-column width="50">
          <template #header>
            <el-checkbox
              :model-value="selectedServices.length === filteredServices.length && filteredServices.length > 0"
              :indeterminate="selectedServices.length > 0 && selectedServices.length < filteredServices.length"
              @change="toggleSelectAll"
            />
          </template>
          <template #default="{ row }">
            <el-checkbox
              :model-value="selectedServices.includes(row.name)"
              @change="toggleSelection(row.name)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="服务名称" min-width="180" sortable />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="PID" width="80">
          <template #default="{ row }">
            {{ row.pid > 0 ? row.pid : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="资源" width="120">
          <template #default="{ row }">
            <div v-if="row.status === 'running'" class="resource-info">
              <span>{{ row.cpu.toFixed(1) }}% / {{ formatMemory(row.memory) }}</span>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="运行时间" width="100">
          <template #default="{ row }">
            {{ row.uptime > 0 ? formatUptime(row.uptime) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="开机启动" width="90">
          <template #default="{ row }">
            <el-switch
              :model-value="row.enabled"
              @change="toggleEnabled(row)"
              :loading="row.loading"
              size="small"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" text @click="viewServiceDetail(row)">
              <el-icon><View /></el-icon>
            </el-button>
            <el-button type="info" size="small" text @click="viewServiceLogs(row)">
              <el-icon><Document /></el-icon>
            </el-button>
            <el-button-group size="small">
              <el-button
                v-if="row.status !== 'running'"
                type="success"
                @click="serviceAction(row.name, 'start')"
                :loading="row.loading"
              >启动</el-button>
              <el-button
                v-if="row.status === 'running'"
                type="warning"
                @click="serviceAction(row.name, 'stop')"
                :loading="row.loading"
              >停止</el-button>
              <el-button
                @click="serviceAction(row.name, 'restart')"
                :loading="row.loading"
              >重启</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- 服务详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="服务详情" width="600px">
      <template v-if="currentService">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="服务名称">{{ currentService.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentService.status)" size="small">
              {{ currentService.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentService.type }}</el-descriptions-item>
          <el-descriptions-item label="PID">{{ currentService.pid > 0 ? currentService.pid : '-' }}</el-descriptions-item>
          <el-descriptions-item label="开机启动">
            <el-tag :type="currentService.enabled ? 'success' : 'info'" size="small">
              {{ currentService.enabled ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="运行时间">
            {{ currentService.uptime > 0 ? formatUptime(currentService.uptime) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="CPU 使用">
            {{ currentService.status === 'running' ? currentService.cpu.toFixed(2) + '%' : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="内存使用">
            {{ currentService.status === 'running' ? formatMemory(currentService.memory) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ currentService.description }}
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button type="info" @click="viewServiceLogs(currentService!); showDetailDialog = false">
          查看日志
        </el-button>
        <el-button
          v-if="currentService?.status !== 'running'"
          type="success"
          @click="serviceAction(currentService!.name, 'start'); showDetailDialog = false"
        >启动</el-button>
        <el-button
          v-if="currentService?.status === 'running'"
          type="warning"
          @click="serviceAction(currentService!.name, 'stop'); showDetailDialog = false"
        >停止</el-button>
      </template>
    </el-dialog>

    <!-- 服务日志对话框 -->
    <el-dialog v-model="showLogDialog" :title="`服务日志 - ${currentService?.name}`" width="800px">
      <div class="log-container">
        <div v-if="serviceLogs.length === 0" class="log-loading">
          <el-icon class="is-loading"><Refresh /></el-icon>
          加载日志中...
        </div>
        <pre v-else class="log-content">{{ serviceLogs.join('\n') }}</pre>
      </div>
      <template #footer>
        <el-button @click="showLogDialog = false">关闭</el-button>
        <el-button type="primary" @click="viewServiceLogs(currentService!)">刷新日志</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { Refresh, Search, View, Document } from '@element-plus/icons-vue'

interface ServiceInfo {
  name: string
  status: string
  description: string
  enabled: boolean
  pid: number
  uptime: number
  type: string
  memory: number
  cpu: number
  loading?: boolean
  selected?: boolean
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(serverStore.currentServerId)
const services = ref<ServiceInfo[]>([])
const loading = ref(false)
const searchText = ref('')
const filterStatus = ref('')
const autoRefresh = ref(false)
const showDetailDialog = ref(false)
const showLogDialog = ref(false)
const currentService = ref<ServiceInfo | null>(null)
const serviceLogs = ref<string[]>([])
const selectedServices = ref<string[]>([])
let refreshInterval: ReturnType<typeof setInterval> | null = null

const connectedServers = computed(() => serverStore.connectedServers)

const statusOptions = ['running', 'stopped', 'failed']

const filteredServices = computed(() => {
  let result = services.value

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(search) ||
      s.description?.toLowerCase().includes(search)
    )
  }

  // 状态过滤
  if (filterStatus.value) {
    result = result.filter(s => {
      if (filterStatus.value === 'running') {
        return s.status === 'running' || s.status === 'active'
      } else if (filterStatus.value === 'stopped') {
        return s.status === 'stopped' || s.status === 'inactive'
      }
      return s.status === filterStatus.value
    })
  }

  return result
})

const runningCount = computed(() => services.value.filter(s => s.status === 'running').length)
const stoppedCount = computed(() => services.value.filter(s => s.status === 'stopped' || s.status === 'inactive').length)
const failedCount = computed(() => services.value.filter(s => s.status === 'failed').length)

watch(selectedServer, (newVal) => {
  if (newVal) loadServices()
}, { immediate: true })

watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshInterval = setInterval(loadServices, 5000)
  } else if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
})

onMounted(() => {
  if (selectedServer.value) loadServices()
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})

// 模拟服务数据
function generateSimulatedServices(): ServiceInfo[] {
  const serviceTemplates = [
    { name: 'nginx', description: 'A high performance web server and reverse proxy', type: 'web', enabled: true, status: 'running' },
    { name: 'mysql', description: 'MySQL Community Server', type: 'database', enabled: true, status: 'running' },
    { name: 'redis-server', description: 'Advanced key-value store', type: 'database', enabled: true, status: 'running' },
    { name: 'docker', description: 'Docker Application Container Engine', type: 'container', enabled: true, status: 'running' },
    { name: 'sshd', description: 'OpenSSH server daemon', type: 'system', enabled: true, status: 'running' },
    { name: 'cron', description: 'Regular background program processing daemon', type: 'system', enabled: true, status: 'running' },
    { name: 'postgresql', description: 'PostgreSQL RDBMS', type: 'database', enabled: false, status: 'stopped' },
    { name: 'mongodb', description: 'MongoDB Database Server', type: 'database', enabled: false, status: 'stopped' },
    { name: 'apache2', description: 'Apache HTTP Server', type: 'web', enabled: false, status: 'stopped' },
    { name: 'php-fpm', description: 'PHP FastCGI Process Manager', type: 'web', enabled: true, status: 'running' },
    { name: 'fail2ban', description: 'Daemon to ban hosts that cause multiple authentication errors', type: 'security', enabled: true, status: 'running' },
    { name: 'ufw', description: 'Uncomplicated Firewall', type: 'security', enabled: true, status: 'running' },
    { name: 'rsyslog', description: 'System Logging Service', type: 'system', enabled: true, status: 'running' },
    { name: 'systemd-resolved', description: 'Network Name Resolution', type: 'system', enabled: true, status: 'running' },
    { name: 'NetworkManager', description: 'Network Manager', type: 'system', enabled: true, status: 'running' },
    { name: 'elasticsearch', description: 'Elasticsearch distributed search engine', type: 'database', enabled: false, status: 'failed' },
    { name: 'rabbitmq-server', description: 'RabbitMQ Messaging Server', type: 'messaging', enabled: false, status: 'stopped' },
    { name: 'memcached', description: 'High-performance memory object caching system', type: 'database', enabled: true, status: 'running' },
  ]

  return serviceTemplates.map((template, idx) => ({
    name: template.name,
    description: template.description,
    type: template.type,
    enabled: template.enabled,
    status: template.status,
    pid: template.status === 'running' ? 1000 + idx * 100 + Math.floor(Math.random() * 50) : 0,
    uptime: template.status === 'running' ? Math.floor(Math.random() * 86400 * 30) : 0,
    memory: template.status === 'running' ? Math.floor(Math.random() * 500) + 50 : 0,
    cpu: template.status === 'running' ? Math.random() * 5 : 0,
    loading: false,
    selected: false
  }))
}

async function loadServices() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    // 使用模拟数据
    await new Promise(resolve => setTimeout(resolve, 300))
    services.value = generateSimulatedServices()
  } catch (error) {
    ElMessage.error(`加载服务列表失败: ${(error as Error).message}`)
  } finally {
    loading.value = false
  }
}

async function serviceAction(name: string, action: string) {
  if (!selectedServer.value) return

  const service = services.value.find(s => s.name === name)
  if (service) service.loading = true

  const actionNames: Record<string, string> = {
    start: '启动',
    stop: '停止',
    restart: '重启',
    enable: '启用',
    disable: '禁用'
  }

  try {
    // 模拟操作
    await new Promise(resolve => setTimeout(resolve, 800))
    ElMessage.success(`${actionNames[action] || action} ${name} 成功`)
    await loadServices()
  } catch (error) {
    ElMessage.error(`操作失败: ${(error as Error).message}`)
  } finally {
    if (service) service.loading = false
  }
}

// 批量操作
async function batchAction(action: string) {
  if (selectedServices.value.length === 0) return

  loading.value = true
  const actionNames: Record<string, string> = {
    start: '启动',
    stop: '停止',
    restart: '重启'
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success(`已${actionNames[action]} ${selectedServices.value.length} 个服务`)
    selectedServices.value = []
    await loadServices()
  } catch (error) {
    ElMessage.error(`批量操作失败: ${(error as Error).message}`)
  } finally {
    loading.value = false
  }
}

// 查看服务详情
function viewServiceDetail(service: ServiceInfo) {
  currentService.value = service
  showDetailDialog.value = true
}

// 查看服务日志
async function viewServiceLogs(service: ServiceInfo) {
  currentService.value = service
  serviceLogs.value = []
  showLogDialog.value = true

  // 模拟日志数据
  await new Promise(resolve => setTimeout(resolve, 500))
  const now = new Date()
  const logTemplates = [
    `Starting ${service.name} service...`,
    `Loading configuration from /etc/${service.name}/${service.name}.conf`,
    `Binding to port ${1000 + Math.floor(Math.random() * 9000)}`,
    `Service ${service.name} started successfully`,
    `Accepting connections...`,
    `Connection from 192.168.1.${Math.floor(Math.random() * 255)}`,
    `Processing request...`,
    `Request completed in ${Math.floor(Math.random() * 100)}ms`,
    `Health check passed`,
    `Memory usage: ${Math.floor(Math.random() * 500)}MB`
  ]

  serviceLogs.value = logTemplates.map((log, idx) => {
    const time = new Date(now.getTime() - (logTemplates.length - idx) * 60000)
    return `[${time.toISOString()}] ${log}`
  })
}

// 选择/取消选择服务
function toggleSelection(name: string) {
  const index = selectedServices.value.indexOf(name)
  if (index === -1) {
    selectedServices.value.push(name)
  } else {
    selectedServices.value.splice(index, 1)
  }
}

// 全选/取消全选
function toggleSelectAll() {
  if (selectedServices.value.length === filteredServices.value.length) {
    selectedServices.value = []
  } else {
    selectedServices.value = filteredServices.value.map(s => s.name)
  }
}

// 格式化内存
function formatMemory(mb: number): string {
  if (mb < 1024) return `${mb} MB`
  return `${(mb / 1024).toFixed(1)} GB`
}

async function toggleEnabled(service: ServiceInfo) {
  const action = service.enabled ? 'disable' : 'enable'
  await serviceAction(service.name, action)
}

function getStatusType(status: string): 'success' | 'info' | 'danger' | 'warning' {
  switch (status) {
    case 'running':
    case 'active':
      return 'success'
    case 'stopped':
    case 'inactive':
      return 'info'
    case 'failed':
      return 'danger'
    case 'starting':
    case 'stopping':
      return 'warning'
    default:
      return 'info'
  }
}

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}小时`
  return `${Math.floor(seconds / 86400)}天`
}
</script>

<style lang="scss" scoped>
.services-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;

  h1 {
    font-size: 24px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;

    .server-select {
      width: 180px;
    }

    .search-input {
      width: 180px;
    }

    .filter-select {
      width: 120px;
    }
  }
}

.stats-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-radius: 8px;

  .stats-left {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .stats-right {
    display: flex;
    gap: 12px;
    align-items: center;

    .selected-count {
      font-size: 13px;
      color: var(--text-secondary);
    }
  }
}

.services-table {
  flex: 1;

  .resource-info {
    font-size: 12px;
    color: var(--text-secondary);
  }
}

.empty-state {
  padding: 60px 0;
}

.log-container {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 16px;
  max-height: 400px;
  overflow: auto;

  .log-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #888;
    padding: 40px;
  }

  .log-content {
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 12px;
    line-height: 1.6;
    color: #d4d4d4;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
  }
}

:deep(.el-dialog__body) {
  padding-top: 10px;
}

:deep(.el-descriptions__label) {
  width: 100px;
}
</style>
