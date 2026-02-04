<template>
  <div class="nginx-page">
    <div class="page-header">
      <div class="header-left">
        <h1>Nginx 管理</h1>
        <p class="subtitle">管理 Nginx 配置和虚拟主机</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="loadNginxStatus">
          <el-option v-for="server in connectedServers" :key="server.id" :label="server.name" :value="server.id" />
        </el-select>
        <el-button type="primary" @click="openAddSiteDialog" :disabled="!selectedServer || !nginxInstalled">
          <el-icon><Plus /></el-icon>添加站点
        </el-button>
        <el-button @click="loadNginxStatus" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <!-- Nginx 状态 -->
    <div class="status-section" v-if="selectedServer">
      <div class="status-card" :class="{ active: nginxRunning }">
        <div class="status-icon"><el-icon :size="32"><Connection /></el-icon></div>
        <div class="status-info">
          <div class="status-title">Nginx 状态</div>
          <el-tag :type="nginxRunning ? 'success' : 'danger'" size="large">
            {{ nginxInstalled ? (nginxRunning ? '运行中' : '已停止') : '未安装' }}
          </el-tag>
          <div class="status-version" v-if="nginxVersion">{{ nginxVersion }}</div>
        </div>
        <div class="status-actions">
          <el-button-group>
            <el-button @click="nginxAction('start')" :disabled="!nginxInstalled || nginxRunning">启动</el-button>
            <el-button @click="nginxAction('stop')" :disabled="!nginxInstalled || !nginxRunning">停止</el-button>
            <el-button @click="nginxAction('reload')" :disabled="!nginxInstalled || !nginxRunning">重载</el-button>
            <el-button @click="testConfig" :disabled="!nginxInstalled">测试配置</el-button>
          </el-button-group>
        </div>
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon sites"><el-icon :size="24"><Document /></el-icon></div>
          <div class="stat-info"><div class="stat-value">{{ sites.length }}</div><div class="stat-label">站点数</div></div>
        </div>
        <div class="stat-card">
          <div class="stat-icon enabled"><el-icon :size="24"><Check /></el-icon></div>
          <div class="stat-info"><div class="stat-value">{{ enabledSites }}</div><div class="stat-label">已启用</div></div>
        </div>
        <div class="stat-card">
          <div class="stat-icon disabled"><el-icon :size="24"><Close /></el-icon></div>
          <div class="stat-info"><div class="stat-value">{{ disabledSites }}</div><div class="stat-label">已禁用</div></div>
        </div>
      </div>
    </div>

    <!-- 站点列表 -->
    <div class="sites-section" v-if="selectedServer && nginxInstalled">
      <el-table :data="sites" v-loading="loading" stripe>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="站点名称" prop="name" width="200" />
        <el-table-column label="域名" min-width="200">
          <template #default="{ row }"><code>{{ row.serverName || '-' }}</code></template>
        </el-table-column>
        <el-table-column label="监听端口" width="120">
          <template #default="{ row }"><code>{{ row.listen || '80' }}</code></template>
        </el-table-column>
        <el-table-column label="根目录" min-width="200">
          <template #default="{ row }"><code class="root-path">{{ row.root || '-' }}</code></template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button-group size="small">
              <el-button @click="toggleSite(row)" :type="row.enabled ? 'warning' : 'success'">
                {{ row.enabled ? '禁用' : '启用' }}
              </el-button>
              <el-button @click="editSite(row)">编辑</el-button>
              <el-button type="danger" @click="deleteSite(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && sites.length === 0" description="暂无站点配置" />
    </div>

    <el-empty v-else-if="!selectedServer" description="请先选择一个已连接的服务器" />
    <el-empty v-else-if="!nginxInstalled" description="该服务器未安装 Nginx" />

    <!-- 添加/编辑站点对话框 -->
    <el-dialog v-model="showSiteDialog" :title="editingSite ? '编辑站点' : '添加站点'" width="600px">
      <el-form :model="siteForm" label-width="100px">
        <el-form-item label="站点名称" required>
          <el-input v-model="siteForm.name" placeholder="如: example.com" :disabled="!!editingSite" />
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="siteForm.serverName" placeholder="如: example.com www.example.com" />
        </el-form-item>
        <el-form-item label="监听端口">
          <el-input v-model="siteForm.listen" placeholder="80" />
        </el-form-item>
        <el-form-item label="根目录">
          <el-input v-model="siteForm.root" placeholder="/var/www/html" />
        </el-form-item>
        <el-form-item label="配置类型">
          <el-radio-group v-model="siteForm.type">
            <el-radio value="static">静态站点</el-radio>
            <el-radio value="proxy">反向代理</el-radio>
            <el-radio value="php">PHP 站点</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="代理地址" v-if="siteForm.type === 'proxy'">
          <el-input v-model="siteForm.proxyPass" placeholder="http://127.0.0.1:3000" />
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="siteForm.ssl" />
        </el-form-item>
        <el-form-item label="SSL 证书" v-if="siteForm.ssl">
          <el-input v-model="siteForm.sslCert" placeholder="/etc/ssl/certs/example.crt" />
        </el-form-item>
        <el-form-item label="SSL 密钥" v-if="siteForm.ssl">
          <el-input v-model="siteForm.sslKey" placeholder="/etc/ssl/private/example.key" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSiteDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSite" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 编辑配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="编辑配置文件" width="800px">
      <el-input v-model="configContent" type="textarea" :rows="20" class="config-editor" />
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Connection, Document, Check, Close } from '@element-plus/icons-vue'

interface NginxSite {
  name: string
  enabled: boolean
  serverName: string
  listen: string
  root: string
  configPath: string
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const loading = ref(false)
const saving = ref(false)
const nginxInstalled = ref(false)
const nginxRunning = ref(false)
const nginxVersion = ref('')
const sites = ref<NginxSite[]>([])
const showSiteDialog = ref(false)
const showConfigDialog = ref(false)
const editingSite = ref<NginxSite | null>(null)
const configContent = ref('')

const siteForm = ref({
  name: '', serverName: '', listen: '80', root: '/var/www/html',
  type: 'static', proxyPass: '', ssl: false, sslCert: '', sslKey: ''
})

const connectedServers = computed(() => serverStore.connectedServers)
const enabledSites = computed(() => sites.value.filter(s => s.enabled).length)
const disabledSites = computed(() => sites.value.filter(s => !s.enabled).length)

if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  loadNginxStatus()
}

async function loadNginxStatus() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const versionResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'nginx', ['-v'])
    nginxInstalled.value = versionResult.exit_code === 0 || versionResult.stderr?.includes('nginx version')
    if (nginxInstalled.value) {
      nginxVersion.value = versionResult.stderr?.match(/nginx\/[\d.]+/)?.[0] || ''
      const statusResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'pgrep -x nginx > /dev/null && echo running || echo stopped'])
      nginxRunning.value = statusResult.stdout?.trim() === 'running'
      await loadSites()
    }
  } catch (error) {
    ElMessage.error('加载失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadSites() {
  if (!selectedServer.value) return
  const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `
    for f in /etc/nginx/sites-available/*; do
      [ -f "$f" ] || continue
      name=$(basename "$f")
      enabled="false"
      [ -L "/etc/nginx/sites-enabled/$name" ] && enabled="true"
      server_name=$(grep -m1 'server_name' "$f" 2>/dev/null | sed 's/.*server_name\\s*//;s/;.*//' || echo "")
      listen=$(grep -m1 'listen' "$f" 2>/dev/null | sed 's/.*listen\\s*//;s/;.*//' || echo "80")
      root=$(grep -m1 'root' "$f" 2>/dev/null | sed 's/.*root\\s*//;s/;.*//' || echo "")
      echo "$name|$enabled|$server_name|$listen|$root|$f"
    done
  `])
  if (result.exit_code === 0 && result.stdout) {
    sites.value = result.stdout.trim().split('\n').filter(Boolean).map(line => {
      const [name, enabled, serverName, listen, root, configPath] = line.split('|')
      return { name, enabled: enabled === 'true', serverName, listen, root, configPath }
    })
  }
}

async function nginxAction(action: string) {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'systemctl', [action, 'nginx'])
    if (result.exit_code === 0) {
      ElMessage.success(`Nginx ${action} 成功`)
      await loadNginxStatus()
    } else {
      ElMessage.error(result.stderr || '操作失败')
    }
  } finally {
    loading.value = false
  }
}

async function testConfig() {
  if (!selectedServer.value) return
  const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'nginx', ['-t'])
  if (result.exit_code === 0 || result.stderr?.includes('successful')) {
    ElMessage.success('配置测试通过')
  } else {
    ElMessage.error('配置错误: ' + (result.stderr || ''))
  }
}

function openAddSiteDialog() {
  editingSite.value = null
  siteForm.value = { name: '', serverName: '', listen: '80', root: '/var/www/html', type: 'static', proxyPass: '', ssl: false, sslCert: '', sslKey: '' }
  showSiteDialog.value = true
}

async function editSite(site: NginxSite) {
  if (!selectedServer.value) return
  editingSite.value = site
  const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'cat', [site.configPath])
  if (result.exit_code === 0) {
    configContent.value = result.stdout || ''
    showConfigDialog.value = true
  }
}

async function saveSite() {
  if (!selectedServer.value || !siteForm.value.name) return
  saving.value = true
  try {
    const config = generateNginxConfig()
    const path = `/etc/nginx/sites-available/${siteForm.value.name}`
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `cat > ${path} << 'NGINX_EOF'\n${config}\nNGINX_EOF`])
    if (result.exit_code === 0) {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'ln', ['-sf', path, `/etc/nginx/sites-enabled/${siteForm.value.name}`])
      ElMessage.success('站点保存成功')
      showSiteDialog.value = false
      await loadSites()
    } else {
      ElMessage.error(result.stderr || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

function generateNginxConfig(): string {
  const { serverName, listen, root, type, proxyPass, ssl, sslCert, sslKey } = siteForm.value
  let config = `server {\n    listen ${listen}${ssl ? ' ssl' : ''};\n    server_name ${serverName};\n`
  if (ssl) config += `    ssl_certificate ${sslCert};\n    ssl_certificate_key ${sslKey};\n`
  if (type === 'proxy') {
    config += `    location / {\n        proxy_pass ${proxyPass};\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n    }\n`
  } else if (type === 'php') {
    config += `    root ${root};\n    index index.php index.html;\n    location ~ \\.php$ {\n        fastcgi_pass unix:/var/run/php/php-fpm.sock;\n        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n        include fastcgi_params;\n    }\n`
  } else {
    config += `    root ${root};\n    index index.html;\n    location / {\n        try_files $uri $uri/ =404;\n    }\n`
  }
  return config + '}'
}

async function saveConfig() {
  if (!selectedServer.value || !editingSite.value) return
  saving.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `cat > ${editingSite.value.configPath} << 'NGINX_EOF'\n${configContent.value}\nNGINX_EOF`])
    if (result.exit_code === 0) {
      ElMessage.success('配置保存成功')
      showConfigDialog.value = false
    } else {
      ElMessage.error(result.stderr || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

async function toggleSite(site: NginxSite) {
  if (!selectedServer.value) return
  const cmd = site.enabled
    ? `rm -f /etc/nginx/sites-enabled/${site.name}`
    : `ln -sf /etc/nginx/sites-available/${site.name} /etc/nginx/sites-enabled/${site.name}`
  const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
  if (result.exit_code === 0) {
    ElMessage.success(site.enabled ? '站点已禁用' : '站点已启用')
    await loadSites()
  }
}

async function deleteSite(site: NginxSite) {
  try {
    await ElMessageBox.confirm(`确定删除站点 ${site.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  if (!selectedServer.value) return
  const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `rm -f /etc/nginx/sites-enabled/${site.name} /etc/nginx/sites-available/${site.name}`])
  if (result.exit_code === 0) {
    ElMessage.success('站点已删除')
    await loadSites()
  }
}
</script>

<style lang="scss" scoped>
.nginx-page { max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;
  .header-left { h1 { font-size: 24px; font-weight: 600; margin-bottom: 4px; } .subtitle { color: var(--text-secondary); font-size: 14px; } }
  .header-right { display: flex; gap: 12px; .server-select { width: 200px; } }
}
.status-section { margin-bottom: 24px; }
.status-card { display: flex; align-items: center; gap: 20px; padding: 24px; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px; margin-bottom: 16px;
  &.active { border-color: #22c55e; }
  .status-icon { width: 64px; height: 64px; border-radius: 16px; display: flex; align-items: center; justify-content: center; background: rgba(99, 102, 241, 0.1); color: #6366f1; }
  .status-info { flex: 1; .status-title { font-size: 14px; color: var(--text-secondary); margin-bottom: 4px; } .status-version { font-size: 12px; color: var(--text-secondary); margin-top: 4px; } }
}
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.stat-card { background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px; padding: 20px; display: flex; align-items: center; gap: 16px;
  .stat-icon { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center;
    &.sites { background: rgba(99, 102, 241, 0.1); color: #6366f1; }
    &.enabled { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
    &.disabled { background: rgba(156, 163, 175, 0.1); color: #9ca3af; }
  }
  .stat-info { .stat-value { font-size: 28px; font-weight: 600; } .stat-label { font-size: 13px; color: var(--text-secondary); } }
}
.sites-section { background: var(--bg-secondary); border-radius: 12px; padding: 20px; }
.root-path { font-family: monospace; font-size: 12px; }
.config-editor { font-family: 'Fira Code', monospace; font-size: 13px; }
:deep(.el-table) { --el-table-bg-color: transparent; --el-table-tr-bg-color: transparent; }
</style>
