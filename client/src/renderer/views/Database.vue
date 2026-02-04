<template>
  <div class="database-page">
    <div class="page-header">
      <div class="header-left">
        <h1>数据库管理</h1>
        <p class="subtitle">管理 MySQL、PostgreSQL、Redis、MongoDB 等数据库</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="detectDatabases">
          <el-option v-for="server in connectedServers" :key="server.id" :label="server.name" :value="server.id" />
        </el-select>
        <el-button @click="detectDatabases" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <!-- 数据库类型卡片 -->
    <div class="db-types-grid" v-if="selectedServer">
      <div class="db-card" v-for="db in detectedDatabases" :key="db.type" :class="{ active: db.running }" @click="selectDatabase(db)">
        <div class="db-icon" :style="{ background: db.color }">
          <span class="db-letter">{{ db.type[0] }}</span>
        </div>
        <div class="db-info">
          <div class="db-name">{{ db.name }}</div>
          <el-tag :type="db.running ? 'success' : 'info'" size="small">{{ db.running ? '运行中' : (db.installed ? '已停止' : '未安装') }}</el-tag>
          <div class="db-version" v-if="db.version">{{ db.version }}</div>
        </div>
        <div class="db-actions" v-if="db.installed">
          <el-button size="small" @click.stop="dbAction(db, db.running ? 'stop' : 'start')">
            {{ db.running ? '停止' : '启动' }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- 数据库详情面板 -->
    <div class="db-detail-panel" v-if="selectedDb && selectedDb.running">
      <div class="panel-header">
        <h2>{{ selectedDb.name }} 管理</h2>
        <el-button-group size="small">
          <el-button @click="activeTab = 'status'" :type="activeTab === 'status' ? 'primary' : ''">状态</el-button>
          <el-button @click="activeTab = 'query'" :type="activeTab === 'query' ? 'primary' : ''">查询</el-button>
          <el-button @click="activeTab = 'databases'" :type="activeTab === 'databases' ? 'primary' : ''">数据库</el-button>
        </el-button-group>
      </div>

      <!-- 状态面板 -->
      <div class="tab-content" v-if="activeTab === 'status'">
        <div class="status-grid">
          <div class="status-item" v-for="(value, key) in dbStatus" :key="key">
            <div class="status-label">{{ key }}</div>
            <div class="status-value">{{ value }}</div>
          </div>
        </div>
      </div>

      <!-- 查询面板 -->
      <div class="tab-content" v-if="activeTab === 'query'">
        <div class="query-section">
          <el-input v-model="queryInput" type="textarea" :rows="4" placeholder="输入 SQL 查询语句..." class="query-input" />
          <div class="query-actions">
            <el-button type="primary" @click="executeQuery" :loading="queryLoading">执行查询</el-button>
            <el-button @click="queryInput = ''">清空</el-button>
          </div>
        </div>
        <div class="query-result" v-if="queryResult">
          <div class="result-header">查询结果 ({{ queryResultRows }} 行)</div>
          <el-table :data="queryResultData" stripe max-height="400" v-if="queryResultData.length">
            <el-table-column v-for="col in queryResultColumns" :key="col" :prop="col" :label="col" min-width="120" />
          </el-table>
          <pre class="result-text" v-else>{{ queryResult }}</pre>
        </div>
      </div>

      <!-- 数据库列表面板 -->
      <div class="tab-content" v-if="activeTab === 'databases'">
        <el-table :data="databaseList" stripe v-loading="loading">
          <el-table-column label="数据库名" prop="name" />
          <el-table-column label="大小" prop="size" width="120" />
          <el-table-column label="表数量" prop="tables" width="100" />
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" @click="showTables(row.name)">查看表</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-empty v-else-if="!selectedServer" description="请先选择一个已连接的服务器" />
    <el-empty v-else-if="selectedDb && !selectedDb.running" description="请先启动数据库服务" />

    <!-- 表列表对话框 -->
    <el-dialog v-model="showTablesDialog" :title="`${currentDbName} 的表`" width="600px">
      <el-table :data="tableList" stripe max-height="400">
        <el-table-column label="表名" prop="name" />
        <el-table-column label="行数" prop="rows" width="100" />
        <el-table-column label="大小" prop="size" width="100" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

interface DatabaseType {
  type: string
  name: string
  installed: boolean
  running: boolean
  version: string
  color: string
  service: string
  queryCmd: string
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const loading = ref(false)
const queryLoading = ref(false)
const detectedDatabases = ref<DatabaseType[]>([])
const selectedDb = ref<DatabaseType | null>(null)
const activeTab = ref('status')
const dbStatus = ref<Record<string, string>>({})
const queryInput = ref('')
const queryResult = ref('')
const queryResultData = ref<any[]>([])
const queryResultColumns = ref<string[]>([])
const queryResultRows = ref(0)
const databaseList = ref<any[]>([])
const tableList = ref<any[]>([])
const showTablesDialog = ref(false)
const currentDbName = ref('')

const connectedServers = computed(() => serverStore.connectedServers)

const dbConfigs: Omit<DatabaseType, 'installed' | 'running' | 'version'>[] = [
  { type: 'mysql', name: 'MySQL', color: '#4479A1', service: 'mysql', queryCmd: 'mysql' },
  { type: 'postgresql', name: 'PostgreSQL', color: '#336791', service: 'postgresql', queryCmd: 'psql' },
  { type: 'redis', name: 'Redis', color: '#DC382D', service: 'redis-server', queryCmd: 'redis-cli' },
  { type: 'mongodb', name: 'MongoDB', color: '#47A248', service: 'mongod', queryCmd: 'mongosh' },
  { type: 'mariadb', name: 'MariaDB', color: '#003545', service: 'mariadb', queryCmd: 'mysql' }
]

if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  detectDatabases()
}

async function detectDatabases() {
  if (!selectedServer.value) return
  loading.value = true
  detectedDatabases.value = []

  try {
    for (const config of dbConfigs) {
      const db: DatabaseType = { ...config, installed: false, running: false, version: '' }

      // 检测是否安装
      const whichResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'which', [config.queryCmd])
      db.installed = whichResult.exit_code === 0

      if (db.installed) {
        // 检测版本
        let versionCmd = config.type === 'redis' ? 'redis-server --version' :
                         config.type === 'mongodb' ? 'mongod --version' :
                         config.type === 'postgresql' ? 'psql --version' :
                         `${config.queryCmd} --version`
        const versionResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', versionCmd])
        const output = versionResult.stdout || versionResult.stderr || ''
        const versionMatch = output.match(/(\d+\.\d+(\.\d+)?)/)?.[1]
        if (versionMatch) db.version = `v${versionMatch}`

        // 检测是否运行
        const pgrepResult = await window.electronAPI.server.executeCommand(selectedServer.value, 'pgrep', ['-x', config.service])
        db.running = pgrepResult.exit_code === 0
      }

      detectedDatabases.value.push(db)
    }
  } catch (error) {
    ElMessage.error('检测失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

async function selectDatabase(db: DatabaseType) {
  selectedDb.value = db
  if (db.running) {
    await loadDbStatus()
    await loadDatabaseList()
  }
}

async function dbAction(db: DatabaseType, action: string) {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'systemctl', [action, db.service])
    if (result.exit_code === 0) {
      ElMessage.success(`${db.name} ${action === 'start' ? '已启动' : '已停止'}`)
      await detectDatabases()
    } else {
      ElMessage.error(result.stderr || '操作失败')
    }
  } finally {
    loading.value = false
  }
}

async function loadDbStatus() {
  if (!selectedServer.value || !selectedDb.value) return
  dbStatus.value = {}

  try {
    let cmd = ''
    switch (selectedDb.value.type) {
      case 'mysql':
      case 'mariadb':
        cmd = "mysql -e \"SHOW STATUS WHERE Variable_name IN ('Uptime','Threads_connected','Questions','Slow_queries')\" 2>/dev/null | tail -n +2"
        break
      case 'postgresql':
        cmd = "sudo -u postgres psql -c \"SELECT 'connections' as name, count(*) as value FROM pg_stat_activity UNION SELECT 'databases', count(*) FROM pg_database\" -t 2>/dev/null"
        break
      case 'redis':
        cmd = "redis-cli INFO server 2>/dev/null | grep -E '^(redis_version|uptime_in_seconds|connected_clients):'"
        break
      case 'mongodb':
        cmd = "mongosh --quiet --eval 'db.serverStatus().connections' 2>/dev/null"
        break
    }

    if (cmd) {
      const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
      if (result.stdout) {
        const lines = result.stdout.trim().split('\n')
        lines.forEach(line => {
          const [key, value] = line.split(/[\t:|]/).map(s => s.trim()).filter(Boolean)
          if (key && value) dbStatus.value[key] = value
        })
      }
    }
  } catch (error) {
    console.error('Load status error:', error)
  }
}

async function loadDatabaseList() {
  if (!selectedServer.value || !selectedDb.value) return
  databaseList.value = []

  try {
    let cmd = ''
    switch (selectedDb.value.type) {
      case 'mysql':
      case 'mariadb':
        cmd = "mysql -e \"SELECT table_schema as name, ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size, COUNT(*) as tables FROM information_schema.tables GROUP BY table_schema\" -N 2>/dev/null"
        break
      case 'postgresql':
        cmd = "sudo -u postgres psql -c \"SELECT datname as name, pg_size_pretty(pg_database_size(datname)) as size FROM pg_database WHERE datistemplate = false\" -t 2>/dev/null"
        break
      case 'redis':
        cmd = "redis-cli INFO keyspace 2>/dev/null"
        break
      case 'mongodb':
        cmd = "mongosh --quiet --eval 'db.adminCommand(\"listDatabases\").databases.forEach(d => print(d.name + \"|\" + d.sizeOnDisk))' 2>/dev/null"
        break
    }

    if (cmd) {
      const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
      if (result.stdout) {
        const lines = result.stdout.trim().split('\n').filter(Boolean)
        databaseList.value = lines.map(line => {
          const parts = line.split(/[\t|]/).map(s => s.trim()).filter(Boolean)
          return { name: parts[0] || '', size: parts[1] || '-', tables: parts[2] || '-' }
        })
      }
    }
  } catch (error) {
    console.error('Load databases error:', error)
  }
}

async function executeQuery() {
  if (!selectedServer.value || !selectedDb.value || !queryInput.value.trim()) return
  queryLoading.value = true
  queryResult.value = ''
  queryResultData.value = []
  queryResultColumns.value = []

  try {
    let cmd = ''
    const query = queryInput.value.trim().replace(/'/g, "'\\''")

    switch (selectedDb.value.type) {
      case 'mysql':
      case 'mariadb':
        cmd = `mysql -e '${query}' 2>&1`
        break
      case 'postgresql':
        cmd = `sudo -u postgres psql -c '${query}' 2>&1`
        break
      case 'redis':
        cmd = `redis-cli ${query} 2>&1`
        break
      case 'mongodb':
        cmd = `mongosh --quiet --eval '${query}' 2>&1`
        break
    }

    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
    queryResult.value = result.stdout || result.stderr || '查询完成'

    // 尝试解析表格数据
    if (result.stdout && (selectedDb.value.type === 'mysql' || selectedDb.value.type === 'mariadb')) {
      const lines = result.stdout.trim().split('\n')
      if (lines.length > 1) {
        queryResultColumns.value = lines[0].split('\t')
        queryResultData.value = lines.slice(1).map(line => {
          const values = line.split('\t')
          const row: Record<string, string> = {}
          queryResultColumns.value.forEach((col, i) => row[col] = values[i] || '')
          return row
        })
        queryResultRows.value = queryResultData.value.length
      }
    }
  } catch (error) {
    queryResult.value = '查询失败: ' + (error as Error).message
  } finally {
    queryLoading.value = false
  }
}

async function showTables(dbName: string) {
  if (!selectedServer.value || !selectedDb.value) return
  currentDbName.value = dbName
  tableList.value = []
  showTablesDialog.value = true

  try {
    let cmd = ''
    switch (selectedDb.value.type) {
      case 'mysql':
      case 'mariadb':
        cmd = `mysql -e "SELECT table_name as name, table_rows as rows, ROUND((data_length + index_length) / 1024, 2) as size FROM information_schema.tables WHERE table_schema='${dbName}'" -N 2>/dev/null`
        break
      case 'postgresql':
        cmd = `sudo -u postgres psql -d ${dbName} -c "SELECT tablename as name, n_live_tup as rows FROM pg_stat_user_tables" -t 2>/dev/null`
        break
    }

    if (cmd) {
      const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
      if (result.stdout) {
        tableList.value = result.stdout.trim().split('\n').filter(Boolean).map(line => {
          const parts = line.split(/[\t|]/).map(s => s.trim()).filter(Boolean)
          return { name: parts[0] || '', rows: parts[1] || '-', size: parts[2] ? `${parts[2]} KB` : '-' }
        })
      }
    }
  } catch (error) {
    ElMessage.error('加载表失败')
  }
}
</script>

<style lang="scss" scoped>
.database-page { max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;
  .header-left { h1 { font-size: 24px; font-weight: 600; margin-bottom: 4px; } .subtitle { color: var(--text-secondary); font-size: 14px; } }
  .header-right { display: flex; gap: 12px; .server-select { width: 200px; } }
}
.db-types-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; margin-bottom: 24px; }
.db-card { display: flex; align-items: center; gap: 12px; padding: 16px; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px; cursor: pointer; transition: all 0.2s;
  &:hover { border-color: var(--primary-color); }
  &.active { border-color: #22c55e; background: linear-gradient(135deg, rgba(34, 197, 94, 0.05), transparent); }
  .db-icon { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center;
    .db-letter { color: white; font-size: 20px; font-weight: 700; }
  }
  .db-info { flex: 1; .db-name { font-weight: 600; margin-bottom: 4px; } .db-version { font-size: 11px; color: var(--text-secondary); margin-top: 2px; } }
}
.db-detail-panel { background: var(--bg-secondary); border-radius: 12px; padding: 20px;
  .panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;
    h2 { font-size: 18px; font-weight: 600; }
  }
}
.tab-content { min-height: 300px; }
.status-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 16px;
  .status-item { padding: 16px; background: var(--bg-tertiary); border-radius: 8px;
    .status-label { font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }
    .status-value { font-size: 18px; font-weight: 600; }
  }
}
.query-section { margin-bottom: 20px;
  .query-input { margin-bottom: 12px; font-family: 'Fira Code', monospace; }
  .query-actions { display: flex; gap: 8px; }
}
.query-result { .result-header { font-size: 14px; color: var(--text-secondary); margin-bottom: 12px; }
  .result-text { background: var(--bg-tertiary); padding: 12px; border-radius: 8px; font-family: monospace; font-size: 13px; white-space: pre-wrap; max-height: 400px; overflow: auto; }
}
:deep(.el-table) { --el-table-bg-color: transparent; --el-table-tr-bg-color: transparent; }
</style>
