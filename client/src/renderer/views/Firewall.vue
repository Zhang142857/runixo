<template>
  <div class="firewall-page">
    <div class="page-header">
      <div class="header-left">
        <h1>防火墙管理</h1>
        <p class="subtitle">管理服务器防火墙规则 (UFW/iptables)</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="loadFirewallStatus">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-button type="primary" @click="openAddRuleDialog" :disabled="!selectedServer">
          <el-icon><Plus /></el-icon>
          添加规则
        </el-button>
        <el-button @click="loadFirewallStatus" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 防火墙状态卡片 -->
    <div class="status-section" v-if="selectedServer">
      <div class="status-card" :class="{ active: firewallStatus.enabled }">
        <div class="status-icon">
          <el-icon :size="32"><Lock /></el-icon>
        </div>
        <div class="status-info">
          <div class="status-title">防火墙状态</div>
          <div class="status-value">
            <el-tag :type="firewallStatus.enabled ? 'success' : 'danger'" size="large">
              {{ firewallStatus.enabled ? '已启用' : '已禁用' }}
            </el-tag>
          </div>
          <div class="status-type">{{ firewallStatus.type || '检测中...' }}</div>
        </div>
        <div class="status-actions">
          <el-button
            :type="firewallStatus.enabled ? 'warning' : 'success'"
            @click="toggleFirewall"
            :loading="toggling"
          >
            {{ firewallStatus.enabled ? '禁用防火墙' : '启用防火墙' }}
          </el-button>
        </div>
      </div>

      <!-- 统计卡片 -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon allow"><el-icon :size="24"><Check /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ allowRulesCount }}</div>
            <div class="stat-label">允许规则</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon deny"><el-icon :size="24"><Close /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ denyRulesCount }}</div>
            <div class="stat-label">拒绝规则</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon total"><el-icon :size="24"><List /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ rules.length }}</div>
            <div class="stat-label">总规则数</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 规则列表 -->
    <div class="rules-section" v-if="selectedServer">
      <div class="section-header">
        <h2>防火墙规则</h2>
        <el-input
          v-model="searchQuery"
          placeholder="搜索规则..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <el-table :data="filteredRules" v-loading="loading" stripe>
        <el-table-column label="序号" width="80" type="index" />
        <el-table-column label="动作" width="100">
          <template #default="{ row }">
            <el-tag :type="row.action === 'ALLOW' ? 'success' : 'danger'" size="small">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="方向" width="100">
          <template #default="{ row }">
            <el-tag :type="row.direction === 'IN' ? 'primary' : 'warning'" size="small">
              {{ row.direction === 'IN' ? '入站' : '出站' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="端口/服务" width="150">
          <template #default="{ row }">
            <code class="port-code">{{ row.port || 'Any' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="协议" width="100">
          <template #default="{ row }">
            <span class="protocol">{{ row.protocol || 'Any' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="150">
          <template #default="{ row }">
            <code class="source-code">{{ row.from || 'Anywhere' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="目标" min-width="150">
          <template #default="{ row }">
            <code class="source-code">{{ row.to || 'Anywhere' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="deleteRule(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && rules.length === 0" description="暂无防火墙规则">
        <el-button type="primary" @click="openAddRuleDialog">添加规则</el-button>
      </el-empty>
    </div>

    <el-empty v-else description="请先选择一个已连接的服务器" />

    <!-- 添加规则对话框 -->
    <el-dialog v-model="showAddDialog" title="添加防火墙规则" width="500px">
      <el-form :model="newRule" label-width="100px">
        <el-form-item label="规则类型">
          <el-radio-group v-model="newRule.type">
            <el-radio value="simple">简单规则</el-radio>
            <el-radio value="advanced">高级规则</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="动作" required>
          <el-radio-group v-model="newRule.action">
            <el-radio value="allow">允许 (ALLOW)</el-radio>
            <el-radio value="deny">拒绝 (DENY)</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="方向" v-if="newRule.type === 'advanced'">
          <el-radio-group v-model="newRule.direction">
            <el-radio value="in">入站 (IN)</el-radio>
            <el-radio value="out">出站 (OUT)</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="端口" required>
          <el-input v-model="newRule.port" placeholder="如: 80, 443, 8080:8090">
            <template #append>
              <el-select v-model="newRule.protocol" style="width: 100px">
                <el-option label="TCP/UDP" value="" />
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
              </el-select>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="来源IP" v-if="newRule.type === 'advanced'">
          <el-input v-model="newRule.from" placeholder="如: 192.168.1.0/24 (留空表示任意)" />
        </el-form-item>

        <el-form-item label="常用端口">
          <div class="preset-ports">
            <el-tag
              v-for="preset in portPresets"
              :key="preset.port"
              class="preset-tag"
              @click="applyPortPreset(preset)"
            >
              {{ preset.name }} ({{ preset.port }})
            </el-tag>
          </div>
        </el-form-item>

        <el-form-item label="备注">
          <el-input v-model="newRule.comment" placeholder="规则备注（可选）" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addRule" :loading="saving">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Refresh, Lock, Check, Close, List, Search
} from '@element-plus/icons-vue'

interface FirewallRule {
  index: number
  action: string
  direction: string
  port: string
  protocol: string
  from: string
  to: string
  raw: string
}

interface FirewallStatus {
  enabled: boolean
  type: string
}

const serverStore = useServerStore()

const selectedServer = ref<string | null>(null)
const loading = ref(false)
const saving = ref(false)
const toggling = ref(false)
const showAddDialog = ref(false)
const searchQuery = ref('')
const rules = ref<FirewallRule[]>([])
const firewallStatus = ref<FirewallStatus>({ enabled: false, type: '' })

const newRule = ref({
  type: 'simple',
  action: 'allow',
  direction: 'in',
  port: '',
  protocol: '',
  from: '',
  comment: ''
})

const connectedServers = computed(() => serverStore.connectedServers)
const allowRulesCount = computed(() => rules.value.filter(r => r.action === 'ALLOW').length)
const denyRulesCount = computed(() => rules.value.filter(r => r.action === 'DENY' || r.action === 'REJECT').length)

const filteredRules = computed(() => {
  if (!searchQuery.value) return rules.value
  const query = searchQuery.value.toLowerCase()
  return rules.value.filter(rule =>
    rule.port?.toLowerCase().includes(query) ||
    rule.from?.toLowerCase().includes(query) ||
    rule.to?.toLowerCase().includes(query) ||
    rule.action?.toLowerCase().includes(query)
  )
})

// 常用端口预设
const portPresets = [
  { name: 'SSH', port: '22', protocol: 'tcp' },
  { name: 'HTTP', port: '80', protocol: 'tcp' },
  { name: 'HTTPS', port: '443', protocol: 'tcp' },
  { name: 'MySQL', port: '3306', protocol: 'tcp' },
  { name: 'PostgreSQL', port: '5432', protocol: 'tcp' },
  { name: 'Redis', port: '6379', protocol: 'tcp' },
  { name: 'MongoDB', port: '27017', protocol: 'tcp' },
  { name: 'FTP', port: '21', protocol: 'tcp' }
]

// 初始化选择第一个已连接的服务器
if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  loadFirewallStatus()
}

function applyPortPreset(preset: { port: string; protocol: string }) {
  newRule.value.port = preset.port
  newRule.value.protocol = preset.protocol
}

async function loadFirewallStatus() {
  if (!selectedServer.value) return

  loading.value = true
  rules.value = []

  try {
    // 首先检测防火墙类型和状态
    const statusResult = await window.electronAPI.server.executeCommand(
      selectedServer.value,
      'bash',
      ['-c', 'which ufw && ufw status || (which iptables && iptables -L -n)']
    )

    if (statusResult.stdout?.includes('ufw')) {
      firewallStatus.value.type = 'UFW'
      firewallStatus.value.enabled = statusResult.stdout.includes('Status: active')
      await loadUfwRules()
    } else if (statusResult.stdout?.includes('iptables') || statusResult.stderr?.includes('iptables')) {
      firewallStatus.value.type = 'iptables'
      firewallStatus.value.enabled = true // iptables 总是"启用"的
      await loadIptablesRules()
    } else {
      firewallStatus.value.type = '未检测到'
      firewallStatus.value.enabled = false
      ElMessage.warning('未检测到 UFW 或 iptables')
    }
  } catch (error) {
    ElMessage.error('加载防火墙状态失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadUfwRules() {
  if (!selectedServer.value) return

  const result = await window.electronAPI.server.executeCommand(
    selectedServer.value,
    'ufw',
    ['status', 'numbered']
  )

  if (result.exit_code === 0 && result.stdout) {
    parseUfwRules(result.stdout)
  }
}

function parseUfwRules(output: string) {
  const lines = output.split('\n')
  const parsedRules: FirewallRule[] = []

  for (const line of lines) {
    // 匹配 UFW 规则格式: [ 1] 22/tcp                     ALLOW IN    Anywhere
    const match = line.match(/\[\s*(\d+)\]\s+(.+?)\s+(ALLOW|DENY|REJECT)\s+(IN|OUT)?\s*(.*)/)
    if (match) {
      const [, index, portProto, action, direction, fromTo] = match
      const [port, protocol] = portProto.split('/')

      parsedRules.push({
        index: parseInt(index),
        action,
        direction: direction || 'IN',
        port: port.trim(),
        protocol: protocol?.toUpperCase() || '',
        from: fromTo?.trim() || 'Anywhere',
        to: '',
        raw: line
      })
    }
  }

  rules.value = parsedRules
}

async function loadIptablesRules() {
  if (!selectedServer.value) return

  const result = await window.electronAPI.server.executeCommand(
    selectedServer.value,
    'iptables',
    ['-L', '-n', '--line-numbers']
  )

  if (result.exit_code === 0 && result.stdout) {
    parseIptablesRules(result.stdout)
  }
}

function parseIptablesRules(output: string) {
  const lines = output.split('\n')
  const parsedRules: FirewallRule[] = []
  let currentChain = ''

  for (const line of lines) {
    // 检测链名
    if (line.startsWith('Chain ')) {
      currentChain = line.split(' ')[1]
      continue
    }

    // 跳过表头
    if (line.startsWith('num') || !line.trim()) continue

    // 解析规则行
    const parts = line.trim().split(/\s+/)
    if (parts.length >= 4 && /^\d+$/.test(parts[0])) {
      const [index, target, protocol, , source, destination, ...rest] = parts

      let port = ''
      const dptMatch = rest.join(' ').match(/dpt:(\d+)/)
      if (dptMatch) port = dptMatch[1]

      parsedRules.push({
        index: parseInt(index),
        action: target === 'ACCEPT' ? 'ALLOW' : target,
        direction: currentChain === 'INPUT' ? 'IN' : 'OUT',
        port,
        protocol: protocol.toUpperCase(),
        from: source === '0.0.0.0/0' ? 'Anywhere' : source,
        to: destination === '0.0.0.0/0' ? 'Anywhere' : destination,
        raw: line
      })
    }
  }

  rules.value = parsedRules
}

async function toggleFirewall() {
  if (!selectedServer.value) return

  const action = firewallStatus.value.enabled ? 'disable' : 'enable'

  try {
    await ElMessageBox.confirm(
      `确定要${firewallStatus.value.enabled ? '禁用' : '启用'}防火墙吗？`,
      '确认操作',
      { type: 'warning' }
    )
  } catch {
    return
  }

  toggling.value = true

  try {
    if (firewallStatus.value.type === 'UFW') {
      const result = await window.electronAPI.server.executeCommand(
        selectedServer.value,
        'bash',
        ['-c', `echo "y" | ufw ${action}`]
      )

      if (result.exit_code === 0) {
        firewallStatus.value.enabled = !firewallStatus.value.enabled
        ElMessage.success(`防火墙已${firewallStatus.value.enabled ? '启用' : '禁用'}`)
      } else {
        ElMessage.error('操作失败: ' + (result.stderr || '未知错误'))
      }
    } else {
      ElMessage.warning('iptables 不支持直接启用/禁用，请使用规则管理')
    }
  } catch (error) {
    ElMessage.error('操作失败: ' + (error as Error).message)
  } finally {
    toggling.value = false
  }
}

function openAddRuleDialog() {
  newRule.value = {
    type: 'simple',
    action: 'allow',
    direction: 'in',
    port: '',
    protocol: '',
    from: '',
    comment: ''
  }
  showAddDialog.value = true
}

async function addRule() {
  if (!selectedServer.value || !newRule.value.port) {
    ElMessage.warning('请填写端口')
    return
  }

  saving.value = true

  try {
    let command: string
    const { action, direction, port, protocol, from, comment } = newRule.value

    if (firewallStatus.value.type === 'UFW') {
      // 构建 UFW 命令
      let ufwCmd = `ufw ${action}`

      if (newRule.value.type === 'advanced' && direction === 'out') {
        ufwCmd += ' out'
      }

      if (from && newRule.value.type === 'advanced') {
        ufwCmd += ` from ${from}`
      }

      ufwCmd += ` ${port}`

      if (protocol) {
        ufwCmd += `/${protocol}`
      }

      if (comment) {
        ufwCmd += ` comment '${comment}'`
      }

      command = ufwCmd
    } else {
      // 构建 iptables 命令
      const chain = direction === 'in' ? 'INPUT' : 'OUTPUT'
      const target = action === 'allow' ? 'ACCEPT' : 'DROP'

      command = `iptables -A ${chain}`

      if (protocol) {
        command += ` -p ${protocol}`
      } else {
        command += ' -p tcp'
      }

      if (from) {
        command += ` -s ${from}`
      }

      command += ` --dport ${port} -j ${target}`
    }

    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value,
      'bash',
      ['-c', command]
    )

    if (result.exit_code === 0) {
      ElMessage.success('规则添加成功')
      showAddDialog.value = false
      await loadFirewallStatus()
    } else {
      ElMessage.error('添加失败: ' + (result.stderr || '未知错误'))
    }
  } catch (error) {
    ElMessage.error('添加失败: ' + (error as Error).message)
  } finally {
    saving.value = false
  }
}

async function deleteRule(rule: FirewallRule) {
  if (!selectedServer.value) return

  try {
    await ElMessageBox.confirm(
      `确定要删除这条规则吗？\n${rule.raw || `${rule.action} ${rule.port}`}`,
      '确认删除',
      { type: 'warning' }
    )
  } catch {
    return
  }

  loading.value = true

  try {
    let command: string

    if (firewallStatus.value.type === 'UFW') {
      command = `echo "y" | ufw delete ${rule.index}`
    } else {
      const chain = rule.direction === 'IN' ? 'INPUT' : 'OUTPUT'
      command = `iptables -D ${chain} ${rule.index}`
    }

    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value,
      'bash',
      ['-c', command]
    )

    if (result.exit_code === 0) {
      ElMessage.success('规则已删除')
      await loadFirewallStatus()
    } else {
      ElMessage.error('删除失败: ' + (result.stderr || '未知错误'))
    }
  } catch (error) {
    ElMessage.error('删除失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.firewall-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;

  .header-left {
    h1 {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 4px;
    }

    .subtitle {
      color: var(--text-secondary);
      font-size: 14px;
    }
  }

  .header-right {
    display: flex;
    gap: 12px;

    .server-select {
      width: 200px;
    }
  }
}

.status-section {
  margin-bottom: 24px;
}

.status-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  margin-bottom: 16px;

  &.active {
    border-color: #22c55e;
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.05), transparent);
  }

  .status-icon {
    width: 64px;
    height: 64px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(99, 102, 241, 0.1);
    color: #6366f1;
  }

  .status-info {
    flex: 1;

    .status-title {
      font-size: 14px;
      color: var(--text-secondary);
      margin-bottom: 4px;
    }

    .status-value {
      margin-bottom: 4px;
    }

    .status-type {
      font-size: 12px;
      color: var(--text-secondary);
    }
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.stat-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;

    &.allow {
      background-color: rgba(34, 197, 94, 0.1);
      color: #22c55e;
    }

    &.deny {
      background-color: rgba(239, 68, 68, 0.1);
      color: #ef4444;
    }

    &.total {
      background-color: rgba(99, 102, 241, 0.1);
      color: #6366f1;
    }
  }

  .stat-info {
    .stat-value {
      font-size: 28px;
      font-weight: 600;
    }

    .stat-label {
      font-size: 13px;
      color: var(--text-secondary);
    }
  }
}

.rules-section {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 20px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h2 {
      font-size: 18px;
      font-weight: 600;
    }

    .search-input {
      width: 250px;
    }
  }
}

.port-code,
.source-code {
  font-family: 'Fira Code', monospace;
  font-size: 13px;
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}

.protocol {
  text-transform: uppercase;
  font-size: 12px;
  color: var(--text-secondary);
}

.preset-ports {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  .preset-tag {
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      background-color: var(--primary-color);
      color: white;
    }
  }
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
}
</style>
