<template>
  <div class="software-store">
    <div class="page-header">
      <div class="header-left">
        <h1>软件商城</h1>
        <p class="subtitle">一键安装常用服务器软件</p>
      </div>
      <div class="header-actions">
        <el-button @click="checkAllStatus" :loading="checking" size="small" round>
          <el-icon><Refresh /></el-icon>刷新状态
        </el-button>
      </div>
    </div>

    <div v-if="!serverStore.currentServer" class="no-server">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <div class="category-tabs animate-in">
        <el-tag v-for="cat in categories" :key="cat.key" :effect="activeCat === cat.key ? 'dark' : 'plain'"
                class="cat-tab" round @click="activeCat = cat.key">{{ cat.label }}</el-tag>
      </div>

      <!-- 已安装 -->
      <div class="section animate-in" v-if="installedItems.length > 0" style="animation-delay:0.05s">
        <div class="section-header">
          <h2>已安装</h2>
          <el-tag type="success" size="small" round>{{ installedItems.length }}</el-tag>
        </div>
        <div class="sw-grid">
          <div v-for="(item, idx) in installedItems" :key="item.name"
               class="sw-card installed animate-card" :style="{ animationDelay: idx * 0.04 + 's' }">
            <div class="sw-icon" :style="{ background: item.color }">
              <TechIcon :name="item.iconName" />
            </div>
            <div class="sw-body">
              <div class="sw-name">{{ item.name }}</div>
              <div class="sw-version"><el-icon><CircleCheck /></el-icon> v{{ item.version }}</div>
            </div>
            <div class="sw-actions">
              <el-button size="small" text type="danger" @click="uninstallItem(item)">卸载</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 可安装 -->
      <div class="section animate-in" style="animation-delay:0.1s">
        <div class="section-header">
          <h2>可安装软件</h2>
        </div>
        <div class="sw-grid">
          <div v-for="(item, idx) in availableItems" :key="item.name"
               class="sw-card animate-card" :style="{ animationDelay: (idx * 0.04 + 0.12) + 's' }">
            <div class="sw-icon" :style="{ background: item.color }">
              <TechIcon :name="item.iconName" />
            </div>
            <div class="sw-body">
              <div class="sw-name">{{ item.name }}</div>
              <div class="sw-desc">{{ item.description }}</div>
            </div>
            <div class="sw-actions">
              <el-button type="primary" size="small" round @click="installItem(item)">安装</el-button>
            </div>
          </div>
        </div>
        <el-empty v-if="availableItems.length === 0 && !checking" description="当前分类下所有软件已安装" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useServerStore } from '@/stores/server'
import { useTaskStore } from '@/stores/task'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, CircleCheck } from '@element-plus/icons-vue'
import TechIcon from '@/components/icons/TechIcons.vue'

interface SwItem {
  name: string; iconName: string; color: string; description: string; category: string
  checkCmd: string
  installSteps: { cmd: string; desc: string }[]
  uninstallSteps: { cmd: string; desc: string }[]
  installed?: boolean; version?: string
}

const serverStore = useServerStore()
const taskStore = useTaskStore()
const checking = ref(false)
const activeCat = ref('all')

const categories = [
  { key: 'all', label: '全部' },
  { key: 'database', label: '数据库' },
  { key: 'webserver', label: 'Web 服务器' },
  { key: 'middleware', label: '中间件' },
  { key: 'monitoring', label: '监控' },
  { key: 'tools', label: '工具' }
]

const swDefs: Omit<SwItem, 'installed' | 'version'>[] = [
  { name: 'MySQL', iconName: 'mysql', color: '#4479a1', description: '关系型数据库', category: 'database',
    checkCmd: 'mysql --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    installSteps: [
      { cmd: 'sudo apt-get update', desc: '更新软件源' },
      { cmd: 'sudo apt-get install -y mysql-server', desc: '安装 MySQL' },
      { cmd: 'sudo systemctl enable mysql && sudo systemctl start mysql', desc: '启动 MySQL' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop mysql', desc: '停止 MySQL' },
      { cmd: 'sudo apt-get remove -y mysql-server mysql-client && sudo apt-get autoremove -y', desc: '卸载 MySQL' }
    ]
  },
  { name: 'PostgreSQL', iconName: 'postgresql', color: '#336791', description: '高级关系型数据库', category: 'database',
    checkCmd: 'psql --version 2>/dev/null | grep -oP "\\d+\\.\\d+"',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y postgresql postgresql-contrib', desc: '安装 PostgreSQL' },
      { cmd: 'sudo systemctl enable postgresql', desc: '启动 PostgreSQL' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop postgresql', desc: '停止' },
      { cmd: 'sudo apt-get remove -y postgresql postgresql-contrib && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Redis', iconName: 'redis', color: '#dc382d', description: '内存数据库 / 缓存', category: 'database',
    checkCmd: 'redis-server --version 2>/dev/null | grep -oP "v=\\K[0-9.]+"',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y redis-server', desc: '安装 Redis' },
      { cmd: 'sudo systemctl enable redis-server && sudo systemctl start redis-server', desc: '启动 Redis' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop redis-server', desc: '停止' },
      { cmd: 'sudo apt-get remove -y redis-server && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'MongoDB', iconName: 'mongodb', color: '#47a248', description: '文档型数据库', category: 'database',
    checkCmd: 'mongod --version 2>/dev/null | grep -oP "v\\K[0-9.]+"',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y gnupg curl', desc: '安装依赖' },
      { cmd: 'curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | sudo gpg --dearmor -o /usr/share/keyrings/mongodb-server-7.0.gpg', desc: '添加 GPG 密钥' },
      { cmd: 'echo "deb [ signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list', desc: '添加仓库' },
      { cmd: 'sudo apt-get update && sudo apt-get install -y mongodb-org', desc: '安装 MongoDB' },
      { cmd: 'sudo systemctl enable mongod && sudo systemctl start mongod', desc: '启动' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop mongod', desc: '停止' },
      { cmd: 'sudo apt-get remove -y mongodb-org && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Nginx', iconName: 'nginx', color: '#009639', description: 'Web 服务器 / 反向代理', category: 'webserver',
    checkCmd: 'nginx -v 2>&1 | grep -oP "nginx/\\K[0-9.]+"',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y nginx', desc: '安装 Nginx' },
      { cmd: 'sudo systemctl enable nginx && sudo systemctl start nginx', desc: '启动 Nginx' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop nginx && sudo systemctl disable nginx', desc: '停止' },
      { cmd: 'sudo apt-get remove -y nginx nginx-common && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Caddy', iconName: 'caddy', color: '#1f88e5', description: '自动 HTTPS 的 Web 服务器', category: 'webserver',
    checkCmd: 'caddy version 2>/dev/null | grep -oP "v[0-9.]+"',
    installSteps: [
      { cmd: 'sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl', desc: '安装依赖' },
      { cmd: 'curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/gpg.key" | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg', desc: '添加密钥' },
      { cmd: 'curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt" | sudo tee /etc/apt/sources.list.d/caddy-stable.list', desc: '添加仓库' },
      { cmd: 'sudo apt-get update && sudo apt-get install -y caddy', desc: '安装 Caddy' }
    ],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y caddy && sudo apt-get autoremove -y', desc: '卸载' }]
  },
  { name: 'RabbitMQ', iconName: 'rabbitmq', color: '#ff6600', description: '消息队列中间件', category: 'middleware',
    checkCmd: 'rabbitmqctl version 2>/dev/null',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y rabbitmq-server', desc: '安装 RabbitMQ' },
      { cmd: 'sudo systemctl enable rabbitmq-server && sudo systemctl start rabbitmq-server', desc: '启动' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop rabbitmq-server', desc: '停止' },
      { cmd: 'sudo apt-get remove -y rabbitmq-server && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Elasticsearch', iconName: 'elasticsearch', color: '#005571', description: '分布式搜索引擎', category: 'middleware',
    checkCmd: 'curl -s localhost:9200 2>/dev/null | grep -oP "number.*?:\\s*\"\\K[0-9.]+"',
    installSteps: [
      { cmd: 'wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo gpg --dearmor -o /usr/share/keyrings/elasticsearch-keyring.gpg', desc: '添加密钥' },
      { cmd: 'echo "deb [signed-by=/usr/share/keyrings/elasticsearch-keyring.gpg] https://artifacts.elastic.co/packages/8.x/apt stable main" | sudo tee /etc/apt/sources.list.d/elastic-8.x.list', desc: '添加仓库' },
      { cmd: 'sudo apt-get update && sudo apt-get install -y elasticsearch', desc: '安装' },
      { cmd: 'sudo systemctl enable elasticsearch && sudo systemctl start elasticsearch', desc: '启动' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop elasticsearch', desc: '停止' },
      { cmd: 'sudo apt-get remove -y elasticsearch && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Grafana', iconName: 'grafana', color: '#f46800', description: '可视化监控面板', category: 'monitoring',
    checkCmd: 'grafana-server -v 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    installSteps: [
      { cmd: 'sudo apt-get install -y apt-transport-https software-properties-common', desc: '安装依赖' },
      { cmd: 'wget -q -O - https://apt.grafana.com/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/grafana.gpg', desc: '添加密钥' },
      { cmd: 'echo "deb [signed-by=/usr/share/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list', desc: '添加仓库' },
      { cmd: 'sudo apt-get update && sudo apt-get install -y grafana', desc: '安装 Grafana' },
      { cmd: 'sudo systemctl enable grafana-server && sudo systemctl start grafana-server', desc: '启动' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop grafana-server', desc: '停止' },
      { cmd: 'sudo apt-get remove -y grafana && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Prometheus', iconName: 'prometheus', color: '#e6522c', description: '监控与告警系统', category: 'monitoring',
    checkCmd: 'prometheus --version 2>&1 | head -1 | grep -oP "\\d+\\.\\d+\\.\\d+"',
    installSteps: [
      { cmd: 'sudo apt-get update && sudo apt-get install -y prometheus', desc: '安装 Prometheus' },
      { cmd: 'sudo systemctl enable prometheus && sudo systemctl start prometheus', desc: '启动' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop prometheus', desc: '停止' },
      { cmd: 'sudo apt-get remove -y prometheus && sudo apt-get autoremove -y', desc: '卸载' }
    ]
  },
  { name: 'Git', iconName: 'git', color: '#f05032', description: '版本控制工具', category: 'tools',
    checkCmd: 'git --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    installSteps: [{ cmd: 'sudo apt-get update && sudo apt-get install -y git', desc: '安装 Git' }],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y git && sudo apt-get autoremove -y', desc: '卸载' }]
  },
  { name: 'Certbot', iconName: 'certbot', color: '#003a70', description: 'SSL 证书自动化工具', category: 'tools',
    checkCmd: 'certbot --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    installSteps: [{ cmd: 'sudo apt-get update && sudo apt-get install -y certbot python3-certbot-nginx', desc: '安装 Certbot' }],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y certbot python3-certbot-nginx && sudo apt-get autoremove -y', desc: '卸载' }]
  },
  { name: 'MinIO', iconName: 'minio', color: '#c72c48', description: '高性能对象存储', category: 'tools',
    checkCmd: 'minio --version 2>/dev/null | grep -oP "\\d{4}-\\d{2}-\\d{2}"',
    installSteps: [
      { cmd: 'wget https://dl.min.io/server/minio/release/linux-amd64/minio -O /tmp/minio', desc: '下载 MinIO' },
      { cmd: 'sudo mv /tmp/minio /usr/local/bin/ && sudo chmod +x /usr/local/bin/minio', desc: '安装' }
    ],
    uninstallSteps: [{ cmd: 'sudo rm -f /usr/local/bin/minio', desc: '卸载' }]
  }
]

const items = ref<SwItem[]>([])
const filteredItems = computed(() => activeCat.value === 'all' ? items.value : items.value.filter(i => i.category === activeCat.value))
const installedItems = computed(() => filteredItems.value.filter(i => i.installed))
const availableItems = computed(() => filteredItems.value.filter(i => !i.installed))

watch(() => serverStore.currentServer, () => { if (serverStore.currentServer) checkAllStatus() })

async function checkAllStatus() {
  const server = serverStore.currentServer
  if (!server) return
  checking.value = true
  for (const item of items.value) {
    try {
      const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', item.checkCmd])
      if (result.exit_code === 0 && (result.stdout || '').trim()) {
        item.installed = true
        item.version = (result.stdout || '').trim().split('\n')[0]
      } else { item.installed = false; item.version = undefined }
    } catch { item.installed = false; item.version = undefined }
  }
  checking.value = false
}

async function installItem(item: SwItem) {
  const server = serverStore.currentServer
  if (!server) return

  const task = taskStore.createTask(`安装 ${item.name}`, 'sw-install', server.id, item.installSteps)
  try {
    for (let i = 0; i < item.installSteps.length; i++) {
      const step = item.installSteps[i]
      taskStore.updateStep(task.id, i, 'running')
      taskStore.appendLog(task.id, `\n[${i + 1}/${item.installSteps.length}] ${step.desc}\n$ ${step.cmd}\n`)
      try {
        const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', step.cmd], { timeout: 300 })
        if (result.stdout) taskStore.appendLog(task.id, result.stdout + '\n')
        if (result.stderr) taskStore.appendLog(task.id, result.stderr + '\n')
        taskStore.updateStep(task.id, i, result.exit_code === 0 ? 'success' : 'failed')
        taskStore.appendLog(task.id, result.exit_code === 0 ? '✓ 完成\n' : `⚠️ 退出码: ${result.exit_code}\n`)
      } catch (e) {
        taskStore.updateStep(task.id, i, 'failed')
        taskStore.appendLog(task.id, `❌ ${(e as Error).message}\n`)
      }
    }
    taskStore.appendLog(task.id, '🔍 验证安装...\n')
    const check = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', item.checkCmd])
    if (check.exit_code === 0 && (check.stdout || '').trim()) {
      taskStore.completeTask(task.id, true)
      ElMessage.success(`${item.name} 安装成功`)
    } else {
      taskStore.completeTask(task.id, false)
      ElMessage.error(`${item.name} 安装可能未成功`)
    }
    await checkAllStatus()
  } catch (e) {
    taskStore.appendLog(task.id, `\n❌ ${(e as Error).message}\n`)
    taskStore.completeTask(task.id, false)
    ElMessage.error('安装失败')
  }
}

async function uninstallItem(item: SwItem) {
  const server = serverStore.currentServer
  if (!server) return
  try { await ElMessageBox.confirm(`确定卸载 ${item.name}？`, '确认', { type: 'warning' }) } catch { return }

  const task = taskStore.createTask(`卸载 ${item.name}`, 'sw-install', server.id, item.uninstallSteps)
  try {
    for (let i = 0; i < item.uninstallSteps.length; i++) {
      const step = item.uninstallSteps[i]
      taskStore.updateStep(task.id, i, 'running')
      taskStore.appendLog(task.id, `\n$ ${step.cmd}\n`)
      try {
        const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', step.cmd], { timeout: 120 })
        if (result.stdout) taskStore.appendLog(task.id, result.stdout + '\n')
        if (result.stderr) taskStore.appendLog(task.id, result.stderr + '\n')
        taskStore.updateStep(task.id, i, result.exit_code === 0 ? 'success' : 'failed')
      } catch (e) {
        taskStore.updateStep(task.id, i, 'failed')
        taskStore.appendLog(task.id, `❌ ${(e as Error).message}\n`)
      }
    }
    taskStore.completeTask(task.id, true)
    ElMessage.success(`${item.name} 已卸载`)
    await checkAllStatus()
  } catch { taskStore.completeTask(task.id, false); ElMessage.error('卸载失败') }
}

onMounted(() => {
  items.value = swDefs.map(d => ({ ...d, installed: false, version: undefined }))
  if (serverStore.currentServer) checkAllStatus()
})
</script>

<style lang="scss" scoped>
@keyframes fadeSlideUp { from { opacity:0; transform:translateY(16px) } to { opacity:1; transform:translateY(0) } }
@keyframes cardIn { from { opacity:0; transform:scale(0.96) translateY(10px) } to { opacity:1; transform:scale(1) translateY(0) } }
.animate-in { animation: fadeSlideUp 0.4s ease both; }
.animate-card { animation: cardIn 0.35s ease both; }

.software-store { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px;
  h1 { font-size: 22px; font-weight: 700; margin: 0 0 4px; }
  .subtitle { color: var(--text-secondary); font-size: 13px; margin: 0; }
}

.no-server { padding: 60px 0; }

.category-tabs { display: flex; gap: 8px; margin-bottom: 20px; flex-wrap: wrap; }
.cat-tab { cursor: pointer; transition: all 0.2s; &:hover { transform: translateY(-1px); } }

.section { margin-bottom: 24px; }
.section-header {
  display: flex; align-items: center; gap: 10px; margin-bottom: 12px;
  h2 { font-size: 15px; font-weight: 600; margin: 0; }
}

.sw-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 12px;
}

.sw-card {
  display: flex; align-items: center; gap: 12px;
  padding: 16px; background: var(--bg-secondary);
  border: 1px solid var(--border-color); border-radius: 12px;
  transition: all 0.25s cubic-bezier(.4,0,.2,1);
  &:hover { border-color: var(--text-muted); transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
  &.installed { border-color: rgba(34,197,94,0.35); background: linear-gradient(135deg, var(--bg-secondary), rgba(34,197,94,0.05)); }
}

.sw-icon {
  width: 42px; height: 42px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.sw-body { flex: 1; min-width: 0; }
.sw-name { font-size: 14px; font-weight: 600; margin-bottom: 2px; }
.sw-desc { font-size: 12px; color: var(--text-secondary); }
.sw-version {
  font-size: 12px; color: var(--success-color); display: flex; align-items: center; gap: 4px;
  .el-icon { font-size: 14px; }
}
.sw-actions { flex-shrink: 0; }
</style>
