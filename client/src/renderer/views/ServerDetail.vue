<template>
  <div class="server-detail">
    <div class="page-header">
      <div class="server-title">
        <el-button text @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <h1>{{ server?.name || '服务器详情' }}</h1>
        <el-tag :type="server?.status === 'connected' ? 'success' : 'info'">
          {{ server?.status === 'connected' ? '已连接' : '未连接' }}
        </el-tag>
        <span class="uptime-badge" v-if="metrics.uptime">运行 {{ formatUptime(metrics.uptime) }}</span>
      </div>
      <div class="header-actions">
        <el-button @click="refreshData" :loading="refreshing"><el-icon><Refresh /></el-icon>刷新</el-button>
        <el-button @click="openTerminal">终端</el-button>
        <el-button @click="openFiles">文件管理</el-button>
      </div>
    </div>

    <div class="info-grid">
      <!-- 系统信息卡片 -->
      <el-card class="info-card">
        <template #header>系统信息</template>
        <div class="info-list">
          <div class="info-item">
            <span class="label">主机名</span>
            <span class="value">{{ simulatedSystemInfo.hostname }}</span>
          </div>
          <div class="info-item">
            <span class="label">操作系统</span>
            <span class="value">{{ simulatedSystemInfo.os }} {{ simulatedSystemInfo.arch }}</span>
          </div>
          <div class="info-item">
            <span class="label">内核版本</span>
            <span class="value">{{ simulatedSystemInfo.kernelVersion }}</span>
          </div>
          <div class="info-item">
            <span class="label">运行时间</span>
            <span class="value">{{ formatUptime(metrics.uptime) }}</span>
          </div>
        </div>
      </el-card>

      <!-- CPU 卡片 -->
      <el-card class="info-card">
        <template #header>CPU</template>
        <div class="metric-display">
          <div class="metric-chart">
            <el-progress
              type="dashboard"
              :percentage="Math.round(metrics.cpu)"
              :color="getProgressColor(metrics.cpu)"
              :stroke-width="10"
            />
          </div>
          <div class="metric-info">
            <div class="info-item">
              <span class="label">型号</span>
              <span class="value">{{ simulatedSystemInfo.cpu?.model }}</span>
            </div>
            <div class="info-item">
              <span class="label">核心数</span>
              <span class="value">{{ simulatedSystemInfo.cpu?.cores }} 核</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 内存卡片 -->
      <el-card class="info-card">
        <template #header>内存</template>
        <div class="metric-display">
          <div class="metric-chart">
            <el-progress
              type="dashboard"
              :percentage="Math.round(metrics.memory)"
              :color="getProgressColor(metrics.memory)"
              :stroke-width="10"
            />
          </div>
          <div class="metric-info">
            <div class="info-item">
              <span class="label">总量</span>
              <span class="value">{{ formatBytes(simulatedSystemInfo.memory?.total) }}</span>
            </div>
            <div class="info-item">
              <span class="label">已用</span>
              <span class="value">{{ formatBytes(simulatedSystemInfo.memory?.used) }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 网络卡片 -->
      <el-card class="info-card">
        <template #header>网络流量</template>
        <div class="network-stats">
          <div class="network-item">
            <div class="network-label">
              <span class="arrow up">↑</span>
              <span>上传</span>
            </div>
            <div class="network-value">{{ metrics.networkOut.toFixed(1) }} MB/s</div>
          </div>
          <div class="network-item">
            <div class="network-label">
              <span class="arrow down">↓</span>
              <span>下载</span>
            </div>
            <div class="network-value">{{ metrics.networkIn.toFixed(1) }} MB/s</div>
          </div>
        </div>
      </el-card>

      <!-- 磁盘卡片 -->
      <el-card class="info-card disk-card">
        <template #header>磁盘</template>
        <div class="disk-list">
          <div v-for="disk in simulatedSystemInfo.disks" :key="disk.mountpoint" class="disk-item">
            <div class="disk-header">
              <span class="disk-mount">{{ disk.mountpoint }}</span>
              <span class="disk-usage">{{ disk.usedPercent.toFixed(1) }}%</span>
            </div>
            <el-progress
              :percentage="disk.usedPercent"
              :stroke-width="8"
              :show-text="false"
              :color="getProgressColor(disk.usedPercent)"
            />
            <div class="disk-size">
              {{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }}
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 快捷操作 -->
    <el-card class="quick-actions">
      <template #header>快捷操作</template>
      <div class="action-grid">
        <el-button @click="executeQuickCommand('uptime')">查看运行时间</el-button>
        <el-button @click="executeQuickCommand('df -h')">磁盘使用</el-button>
        <el-button @click="executeQuickCommand('free -h')">内存使用</el-button>
        <el-button @click="executeQuickCommand('top -bn1 | head -20')">进程列表</el-button>
        <el-button @click="executeQuickCommand('docker ps')">Docker 容器</el-button>
        <el-button @click="executeQuickCommand('systemctl list-units --failed')">失败服务</el-button>
      </div>
    </el-card>

    <!-- 命令输出 -->
    <el-card v-if="commandOutput" class="command-output">
      <template #header>
        <div class="output-header">
          <span>命令输出</span>
          <el-button text size="small" @click="commandOutput = ''">清除</el-button>
        </div>
      </template>
      <pre class="output-content">{{ commandOutput }}</pre>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'

interface ServerMetric {
  cpu: number
  memory: number
  disk: number
  networkIn: number
  networkOut: number
  uptime: number
}

const route = useRoute()
const router = useRouter()
const serverStore = useServerStore()

const serverId = route.params.id as string
const server = computed(() => serverStore.servers.find(s => s.id === serverId))
const metrics = ref<ServerMetric>({ cpu: 0, memory: 0, disk: 0, networkIn: 0, networkOut: 0, uptime: 0 })
const commandOutput = ref('')
const refreshing = ref(false)
let metricsInterval: ReturnType<typeof setInterval> | null = null

// 模拟系统信息
const simulatedSystemInfo = {
  hostname: server.value?.name || 'server-01',
  os: 'Ubuntu 22.04.3 LTS',
  arch: 'x86_64',
  kernelVersion: '5.15.0-91-generic',
  uptime: 1234567,
  cpu: { model: 'Intel Xeon E5-2680 v4 @ 2.40GHz', cores: 8 },
  memory: { total: 17179869184, used: 10737418240 },
  disks: [
    { mountpoint: '/', total: 536870912000, used: 161061273600, usedPercent: 30 },
    { mountpoint: '/home', total: 1099511627776, used: 439804651110, usedPercent: 40 },
    { mountpoint: '/var', total: 214748364800, used: 96636764160, usedPercent: 45 }
  ]
}

// 生成随机波动值
function generateMetricValue(base: number, variance: number, min = 0, max = 100): number {
  const change = (Math.random() - 0.5) * variance
  return Math.max(min, Math.min(max, base + change))
}

// 初始化和更新指标
function initMetrics() {
  metrics.value = {
    cpu: 35 + Math.random() * 20,
    memory: 55 + Math.random() * 15,
    disk: 35 + Math.random() * 10,
    networkIn: 20 + Math.random() * 30,
    networkOut: 10 + Math.random() * 20,
    uptime: simulatedSystemInfo.uptime
  }
}

function updateMetrics() {
  metrics.value = {
    cpu: generateMetricValue(metrics.value.cpu, 8, 10, 90),
    memory: generateMetricValue(metrics.value.memory, 3, 30, 85),
    disk: generateMetricValue(metrics.value.disk, 0.5, 20, 80),
    networkIn: generateMetricValue(metrics.value.networkIn, 15, 5, 150),
    networkOut: generateMetricValue(metrics.value.networkOut, 10, 2, 100),
    uptime: metrics.value.uptime + 2
  }
}

onMounted(() => {
  initMetrics()
  metricsInterval = setInterval(updateMetrics, 2000)
})

onUnmounted(() => {
  if (metricsInterval) clearInterval(metricsInterval)
})

function refreshData() {
  refreshing.value = true
  setTimeout(() => {
    initMetrics()
    refreshing.value = false
    ElMessage.success('数据已刷新')
  }, 500)
}

function openTerminal() {
  router.push(`/terminal/${serverId}`)
}

function openFiles() {
  router.push(`/files/${serverId}`)
}

async function executeQuickCommand(command: string) {
  // 模拟命令执行
  const outputs: Record<string, string> = {
    'uptime': ' 20:45:32 up 14 days,  6:23,  2 users,  load average: 0.52, 0.48, 0.45',
    'df -h': `Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       500G  150G  350G  30% /
/dev/sdb1       1.0T  400G  600G  40% /home
/dev/sdc1       200G   90G  110G  45% /var
tmpfs           7.8G  1.2M  7.8G   1% /run`,
    'free -h': `              total        used        free      shared  buff/cache   available
Mem:           16Gi       10Gi       2.1Gi       256Mi       3.8Gi       5.4Gi
Swap:         4.0Gi       512Mi      3.5Gi`,
    'top -bn1 | head -20': `top - 20:45:32 up 14 days,  6:23,  2 users,  load average: 0.52, 0.48, 0.45
Tasks: 156 total,   1 running, 155 sleeping,   0 stopped,   0 zombie
%Cpu(s): 35.2 us,  8.1 sy,  0.0 ni, 54.3 id,  1.8 wa,  0.0 hi,  0.6 si
MiB Mem :  16384.0 total,   2150.4 free,  10240.0 used,   3993.6 buff/cache
MiB Swap:   4096.0 total,   3584.0 free,    512.0 used.   5530.4 avail Mem

  PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
 1234 root      20   0 2456780 524288  45678 S  12.5   3.1   125:32 nginx
 2345 mysql     20   0 4567890 1048576 65432 S   8.3   6.4   234:56 mysqld
 3456 root      20   0 1234567 262144  32456 S   5.2   1.6    45:23 node
 4567 www-data  20   0  987654 131072  21345 S   3.1   0.8    12:45 php-fpm`,
    'docker ps': `CONTAINER ID   IMAGE          COMMAND                  STATUS          PORTS                    NAMES
a1b2c3d4e5f6   nginx:latest   "/docker-entrypoint.…"   Up 5 days       0.0.0.0:80->80/tcp       nginx-proxy
b2c3d4e5f6a7   mysql:8.0      "docker-entrypoint.s…"   Up 5 days       0.0.0.0:3306->3306/tcp   mysql-db
c3d4e5f6a7b8   redis:7        "docker-entrypoint.s…"   Up 5 days       0.0.0.0:6379->6379/tcp   redis-cache
d4e5f6a7b8c9   node:18        "docker-entrypoint.s…"   Up 3 days       0.0.0.0:3000->3000/tcp   app-backend`,
    'systemctl list-units --failed': `  UNIT                         LOAD   ACTIVE SUB    DESCRIPTION
0 loaded units listed.`
  }
  commandOutput.value = outputs[command] || `$ ${command}\n命令执行完成`
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  if (days > 0) return `${days}天${hours}小时`
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours}小时${minutes}分钟`
}

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function getProgressColor(percentage: number): string {
  if (percentage < 60) return '#22c55e'
  if (percentage < 80) return '#f59e0b'
  return '#ef4444'
}
</script>

<style lang="scss" scoped>
.server-detail {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  .server-title {
    display: flex;
    align-items: center;
    gap: 12px;

    h1 {
      font-size: 24px;
      font-weight: 600;
    }
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.info-card {
  .info-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .info-item {
    display: flex;
    justify-content: space-between;

    .label {
      color: var(--text-secondary);
    }

    .value {
      font-family: monospace;
    }
  }
}

.metric-display {
  display: flex;
  gap: 24px;

  .metric-chart {
    flex-shrink: 0;
  }

  .metric-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 12px;
  }
}

.disk-card {
  grid-column: span 2;

  .disk-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 16px;
  }

  .disk-item {
    .disk-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;

      .disk-mount {
        font-family: monospace;
      }

      .disk-usage {
        font-weight: 500;
      }
    }

    .disk-size {
      margin-top: 4px;
      font-size: 12px;
      color: var(--text-secondary);
    }
  }
}

.network-stats {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .network-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    background-color: var(--bg-tertiary);
    border-radius: 8px;

    .network-label {
      display: flex;
      align-items: center;
      gap: 8px;

      .arrow {
        font-size: 16px;
        font-weight: bold;

        &.up {
          color: #22c55e;
        }

        &.down {
          color: #3b82f6;
        }
      }
    }

    .network-value {
      font-family: monospace;
      font-size: 18px;
      font-weight: 600;
    }
  }
}

.uptime-badge {
  padding: 4px 12px;
  background-color: var(--bg-tertiary);
  border-radius: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.quick-actions {
  margin-bottom: 24px;

  .action-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
}

.command-output {
  .output-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .output-content {
    background-color: var(--bg-color);
    padding: 16px;
    border-radius: 8px;
    font-family: 'Fira Code', monospace;
    font-size: 13px;
    line-height: 1.5;
    overflow-x: auto;
    white-space: pre-wrap;
    max-height: 400px;
    overflow-y: auto;
  }
}
</style>
