<template>
  <div class="database-page">
    <div class="page-header animate-in">
      <div class="header-left">
        <h1>数据库管理</h1>
        <p class="subtitle">管理和监控服务器上的数据库服务</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="detectDatabases">
          <el-option v-for="s in connectedServers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="detectDatabases" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <el-empty v-if="!selectedServer" description="请先选择一个已连接的服务器" />

    <template v-else>
      <!-- 数据库卡片 -->
      <div class="db-grid">
        <div class="db-card" v-for="(db, i) in detectedDatabases" :key="db.type"
          :class="{ active: selectedDb?.type === db.type, running: db.running, 'animate-card': true }"
          :style="{ animationDelay: i * 0.05 + 's' }"
          @click="selectDatabase(db)">
          <div class="db-icon" :style="{ background: db.color }">
            <TechIcon :name="db.icon" :size="28" />
          </div>
          <div class="db-body">
            <div class="db-name">{{ db.name }}</div>
            <div class="db-meta">
              <el-tag :type="db.running ? 'success' : db.installed ? 'info' : 'danger'" size="small" round>
                {{ db.running ? '运行中' : db.installed ? '已停止' : '未安装' }}
              </el-tag>
              <span class="db-ver" v-if="db.version">{{ db.version }}</span>
            </div>
          </div>
          <div class="db-action" v-if="db.installed" @click.stop>
            <el-button size="small" :type="db.running ? 'danger' : 'success'" plain
              @click="dbAction(db, db.running ? 'stop' : 'start')" :loading="db.actionLoading">
              {{ db.running ? '停止' : '启动' }}
            </el-button>
          </div>
        </div>
      </div>

      <!-- 详情面板 -->
      <div class="detail-panel" v-if="selectedDb">
        <div class="panel-header">
          <div class="panel-title">
            <div class="panel-icon" :style="{ background: selectedDb.color }">
              <TechIcon :name="selectedDb.icon" :size="22" />
            </div>
            <div>
              <h2>{{ selectedDb.name }}</h2>
              <span class="panel-ver">{{ selectedDb.version || '未检测到版本' }}</span>
            </div>
          </div>
          <el-radio-group v-model="activeTab" size="small">
            <el-radio-button value="overview">概览</el-radio-button>
            <el-radio-button value="databases">数据库</el-radio-button>
            <el-radio-button value="query">查询</el-radio-button>
            <el-radio-button value="users">用户</el-radio-button>
            <el-radio-button value="config">配置</el-radio-button>
          </el-radio-group>
        </div>

        <div class="panel-body" v-if="!selectedDb.running">
          <el-empty description="数据库未运行，请先启动服务">
            <el-button type="success" @click="dbAction(selectedDb, 'start')">启动服务</el-button>
          </el-empty>
        </div>

        <!-- 概览 -->
        <div class="panel-body" v-else-if="activeTab === 'overview'">
          <div class="stats-row">
            <div class="stat-card" v-for="s in overviewStats" :key="s.label">
              <div class="stat-icon" :style="{ background: s.bg, color: s.color }">
                <el-icon><component :is="s.icon" /></el-icon>
              </div>
              <div class="stat-val">{{ s.value }}</div>
              <div class="stat-lbl">{{ s.label }}</div>
            </div>
          </div>
          <div class="info-section" v-if="Object.keys(dbStatus).length">
            <h3>服务状态</h3>
            <div class="info-grid">
              <div class="info-item" v-for="(val, key) in dbStatus" :key="key">
                <span class="info-key">{{ key }}</span>
                <span class="info-val">{{ val }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 数据库列表 -->
        <div class="panel-body" v-else-if="activeTab === 'databases'">
          <div class="tab-toolbar">
            <el-input v-model="dbSearch" placeholder="搜索数据库..." prefix-icon="Search" clearable style="width: 240px" />
            <el-button type="primary" size="small" @click="showCreateDb = true" v-if="selectedDb.type !== 'redis'">
              <el-icon><Plus /></el-icon>创建数据库
            </el-button>
          </div>
          <el-table :data="filteredDatabases" stripe v-loading="loading" empty-text="暂无数据库">
            <el-table-column label="数据库名" prop="name" min-width="180">
              <template #default="{ row }">
                <span class="db-name-cell">{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column label="大小" prop="size" width="120" />
            <el-table-column label="表/键数" prop="tables" width="100" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="showTables(row.name)">浏览</el-button>
                <el-button size="small" type="danger" plain @click="confirmDropDb(row.name)"
                  v-if="!['mysql','information_schema','performance_schema','sys','postgres','template0','template1'].includes(row.name)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 查询 -->
        <div class="panel-body" v-else-if="activeTab === 'query'">
          <div class="query-area">
            <el-input v-model="queryInput" type="textarea" :rows="5"
              :placeholder="queryPlaceholder" class="query-input" />
            <div class="query-toolbar">
              <el-button type="primary" @click="executeQuery" :loading="queryLoading">
                <el-icon><CaretRight /></el-icon>执行
              </el-button>
              <el-button @click="queryInput = ''">清空</el-button>
              <div class="query-hint" v-if="queryTime">耗时 {{ queryTime }}ms</div>
            </div>
          </div>
          <div class="query-result" v-if="queryResult">
            <div class="result-header">
              <span>查询结果</span>
              <span class="result-count" v-if="queryResultRows">{{ queryResultRows }} 行</span>
            </div>
            <el-table :data="queryResultData" stripe max-height="400" v-if="queryResultData.length" size="small">
              <el-table-column v-for="col in queryResultColumns" :key="col" :prop="col" :label="col" min-width="120" show-overflow-tooltip />
            </el-table>
            <pre class="result-text" v-else>{{ queryResult }}</pre>
          </div>
        </div>

        <!-- 用户管理 -->
        <div class="panel-body" v-else-if="activeTab === 'users'">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="showCreateUser = true">
              <el-icon><Plus /></el-icon>创建用户
            </el-button>
            <el-button size="small" @click="loadUsers">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
          </div>
          <el-table :data="userList" stripe v-loading="loading" empty-text="暂无用户数据">
            <el-table-column label="用户名" prop="user" min-width="150" />
            <el-table-column label="主机" prop="host" width="150" />
            <el-table-column label="权限" prop="privileges" min-width="200" show-overflow-tooltip />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="danger" plain @click="confirmDropUser(row)"
                  v-if="row.user !== 'root'">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 配置 -->
        <div class="panel-body" v-else-if="activeTab === 'config'">
          <div class="config-section">
            <h3>配置文件</h3>
            <div class="config-file" v-if="configContent">
              <div class="config-path">{{ configPath }}</div>
              <pre class="config-text">{{ configContent }}</pre>
            </div>
            <el-empty v-else description="无法读取配置文件" />
          </div>
          <div class="config-section" v-if="selectedDb.type !== 'redis'">
            <h3>常用操作</h3>
            <div class="quick-actions">
              <el-button @click="dbAction(selectedDb, 'restart')">重启服务</el-button>
              <el-button @click="loadConfig">重新加载配置</el-button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- 创建数据库对话框 -->
    <el-dialog v-model="showCreateDb" title="创建数据库" width="420px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="数据库名">
          <el-input v-model="newDbName" placeholder="输入数据库名称" />
        </el-form-item>
        <el-form-item label="字符集" v-if="selectedDb?.type === 'mysql' || selectedDb?.type === 'mariadb'">
          <el-select v-model="newDbCharset" style="width: 100%">
            <el-option label="utf8mb4" value="utf8mb4" />
            <el-option label="utf8" value="utf8" />
            <el-option label="latin1" value="latin1" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDb = false">取消</el-button>
        <el-button type="primary" @click="createDatabase" :loading="loading">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建用户对话框 -->
    <el-dialog v-model="showCreateUser" title="创建用户" width="420px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="用户名"><el-input v-model="newUser.name" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="newUser.password" type="password" show-password /></el-form-item>
        <el-form-item label="主机" v-if="selectedDb?.type !== 'redis'">
          <el-select v-model="newUser.host" style="width: 100%">
            <el-option label="localhost" value="localhost" />
            <el-option label="% (所有)" value="%" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateUser = false">取消</el-button>
        <el-button type="primary" @click="createUser" :loading="loading">创建</el-button>
      </template>
    </el-dialog>

    <!-- 表浏览对话框 -->
    <el-dialog v-model="showTablesDialog" :title="`${currentDbName} - 表结构`" width="650px" destroy-on-close>
      <el-table :data="tableList" stripe max-height="400" empty-text="暂无表">
        <el-table-column label="表名" prop="name" min-width="180" />
        <el-table-column label="行数" prop="rows" width="100" />
        <el-table-column label="大小" prop="size" width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, CaretRight, Connection, Coin, Timer, DataLine } from '@element-plus/icons-vue'
import TechIcon from '@/components/icons/TechIcons.vue'

interface DbType {
  type: string; name: string; icon: string; installed: boolean; running: boolean
  version: string; color: string; service: string; queryCmd: string; actionLoading?: boolean
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const loading = ref(false)
const queryLoading = ref(false)
const detectedDatabases = ref<DbType[]>([])
const selectedDb = ref<DbType | null>(null)
const activeTab = ref('overview')
const dbStatus = ref<Record<string, string>>({})
const queryInput = ref('')
const queryResult = ref('')
const queryResultData = ref<any[]>([])
const queryResultColumns = ref<string[]>([])
const queryResultRows = ref(0)
const queryTime = ref(0)
const databaseList = ref<any[]>([])
const tableList = ref<any[]>([])
const showTablesDialog = ref(false)
const currentDbName = ref('')
const dbSearch = ref('')
const showCreateDb = ref(false)
const newDbName = ref('')
const newDbCharset = ref('utf8mb4')
const showCreateUser = ref(false)
const newUser = reactive({ name: '', password: '', host: 'localhost' })
const userList = ref<any[]>([])
const configContent = ref('')
const configPath = ref('')

const connectedServers = computed(() => serverStore.connectedServers)
const filteredDatabases = computed(() => {
  if (!dbSearch.value) return databaseList.value
  const q = dbSearch.value.toLowerCase()
  return databaseList.value.filter(d => d.name.toLowerCase().includes(q))
})

const queryPlaceholder = computed(() => {
  if (!selectedDb.value) return ''
  const m: Record<string, string> = {
    mysql: 'SELECT * FROM table_name LIMIT 10;', mariadb: 'SELECT * FROM table_name LIMIT 10;',
    postgresql: 'SELECT * FROM table_name LIMIT 10;', redis: 'GET key / KEYS *',
    mongodb: 'db.collection.find().limit(10)'
  }
  return m[selectedDb.value.type] || ''
})

const overviewStats = computed(() => {
  const s = dbStatus.value
  return [
    { label: '连接数', value: s.Threads_connected || s.connected_clients || s.connections || '-', icon: Connection, bg: 'rgba(99,102,241,0.15)', color: '#6366f1' },
    { label: '数据库数', value: String(databaseList.value.length), icon: Coin, bg: 'rgba(16,185,129,0.15)', color: '#10b981' },
    { label: '运行时间', value: formatUptime(s.Uptime || s.uptime_in_seconds || ''), icon: Timer, bg: 'rgba(56,189,248,0.15)', color: '#38bdf8' },
    { label: '查询数', value: s.Questions || s.total_commands_processed || '-', icon: DataLine, bg: 'rgba(245,158,11,0.15)', color: '#f59e0b' }
  ]
})

const dbConfigs: Omit<DbType, 'installed' | 'running' | 'version' | 'actionLoading'>[] = [
  { type: 'mysql', name: 'MySQL', icon: 'mysql', color: '#4479A1', service: 'mysql', queryCmd: 'mysql' },
  { type: 'postgresql', name: 'PostgreSQL', icon: 'postgresql', color: '#336791', service: 'postgresql', queryCmd: 'psql' },
  { type: 'redis', name: 'Redis', icon: 'redis', color: '#DC382D', service: 'redis-server', queryCmd: 'redis-cli' },
  { type: 'mongodb', name: 'MongoDB', icon: 'mongodb', color: '#47A248', service: 'mongod', queryCmd: 'mongosh' },
  { type: 'mariadb', name: 'MariaDB', icon: 'mariadb', color: '#003545', service: 'mariadb', queryCmd: 'mysql' }
]

if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  detectDatabases()
}

function formatUptime(sec: string) {
  const n = parseInt(sec)
  if (!n) return '-'
  if (n < 3600) return `${Math.floor(n / 60)}分钟`
  if (n < 86400) return `${Math.floor(n / 3600)}小时`
  return `${Math.floor(n / 86400)}天`
}

async function exec(cmd: string) {
  if (!selectedServer.value) throw new Error('未选择服务器')
  return window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd], { timeout: 30 })
}

async function detectDatabases() {
  if (!selectedServer.value) return
  loading.value = true
  detectedDatabases.value = []
  try {
    for (const cfg of dbConfigs) {
      const db: DbType = { ...cfg, installed: false, running: false, version: '', actionLoading: false }
      const w = await exec(`which ${cfg.queryCmd} 2>/dev/null`)
      db.installed = w.exit_code === 0
      if (db.installed) {
        const vCmd = cfg.type === 'redis' ? 'redis-server --version' : cfg.type === 'mongodb' ? 'mongod --version' : cfg.type === 'postgresql' ? 'psql --version' : `${cfg.queryCmd} --version`
        const vr = await exec(`${vCmd} 2>&1`)
        db.version = (vr.stdout + vr.stderr).match(/(\d+\.\d+(\.\d+)?)/)?.[1] ? `v${(vr.stdout + vr.stderr).match(/(\d+\.\d+(\.\d+)?)/)?.[1]}` : ''
        const pr = await exec(`pgrep -x ${cfg.service} 2>/dev/null`)
        db.running = pr.exit_code === 0
      }
      detectedDatabases.value.push(db)
    }
  } catch (e) { ElMessage.error('检测失败: ' + (e as Error).message) }
  finally { loading.value = false }
}

async function selectDatabase(db: DbType) {
  selectedDb.value = db
  activeTab.value = 'overview'
  if (db.running) { await Promise.all([loadDbStatus(), loadDatabaseList()]) }
}

async function dbAction(db: DbType, action: string) {
  db.actionLoading = true
  try {
    const r = await exec(`systemctl ${action} ${db.service} 2>&1`)
    if (r.exit_code === 0) {
      ElMessage.success(`${db.name} ${action === 'start' ? '已启动' : action === 'stop' ? '已停止' : '已重启'}`)
      await detectDatabases()
      if (selectedDb.value?.type === db.type) {
        selectedDb.value = detectedDatabases.value.find(d => d.type === db.type) || null
        if (selectedDb.value?.running) await Promise.all([loadDbStatus(), loadDatabaseList()])
      }
    } else { ElMessage.error(r.stderr || r.stdout || '操作失败') }
  } finally { db.actionLoading = false }
}

async function loadDbStatus() {
  if (!selectedDb.value) return
  dbStatus.value = {}
  const cmds: Record<string, string> = {
    mysql: "mysql -e \"SHOW GLOBAL STATUS WHERE Variable_name IN ('Uptime','Threads_connected','Questions','Slow_queries','Bytes_received','Bytes_sent')\" -N 2>/dev/null",
    mariadb: "mysql -e \"SHOW GLOBAL STATUS WHERE Variable_name IN ('Uptime','Threads_connected','Questions','Slow_queries')\" -N 2>/dev/null",
    postgresql: "sudo -u postgres psql -t -c \"SELECT 'connections', count(*) FROM pg_stat_activity UNION ALL SELECT 'databases', count(*) FROM pg_database\" 2>/dev/null",
    redis: "redis-cli INFO 2>/dev/null | grep -E '^(redis_version|uptime_in_seconds|connected_clients|used_memory_human|total_commands_processed):'",
    mongodb: "mongosh --quiet --eval 'const s=db.serverStatus(); print(\"connections:\" + s.connections.current + \"\\nuptime:\" + s.uptime)' 2>/dev/null"
  }
  const cmd = cmds[selectedDb.value.type]
  if (!cmd) return
  try {
    const r = await exec(cmd)
    if (r.stdout) r.stdout.trim().split('\n').forEach(line => {
      const [k, v] = line.split(/[\t:|]/).map(s => s.trim()).filter(Boolean)
      if (k && v) dbStatus.value[k] = v
    })
  } catch {}
}

async function loadDatabaseList() {
  if (!selectedDb.value) return
  databaseList.value = []
  const cmds: Record<string, string> = {
    mysql: "mysql -N -e \"SELECT table_schema, ROUND(SUM(data_length+index_length)/1024/1024,2), COUNT(*) FROM information_schema.tables GROUP BY table_schema\" 2>/dev/null",
    mariadb: "mysql -N -e \"SELECT table_schema, ROUND(SUM(data_length+index_length)/1024/1024,2), COUNT(*) FROM information_schema.tables GROUP BY table_schema\" 2>/dev/null",
    postgresql: "sudo -u postgres psql -t -c \"SELECT datname, pg_size_pretty(pg_database_size(datname)), 0 FROM pg_database WHERE datistemplate=false\" 2>/dev/null",
    redis: "redis-cli INFO keyspace 2>/dev/null",
    mongodb: "mongosh --quiet --eval 'db.adminCommand(\"listDatabases\").databases.forEach(d=>print(d.name+\"\\t\"+Math.round(d.sizeOnDisk/1024)+\" KB\\t-\"))' 2>/dev/null"
  }
  const cmd = cmds[selectedDb.value.type]
  if (!cmd) return
  try {
    const r = await exec(cmd)
    if (r.stdout) {
      r.stdout.trim().split('\n').filter(l => l && !l.startsWith('#')).forEach(line => {
        const p = line.split(/[\t|]/).map(s => s.trim()).filter(Boolean)
        if (p[0]) databaseList.value.push({ name: p[0], size: p[1] ? (p[1].includes('KB') || p[1].includes('MB') || p[1].includes('GB') ? p[1] : p[1] + ' MB') : '-', tables: p[2] || '-' })
      })
    }
  } catch {}
}

async function executeQuery() {
  if (!selectedDb.value || !queryInput.value.trim()) return
  queryLoading.value = true; queryResult.value = ''; queryResultData.value = []; queryResultColumns.value = []; queryTime.value = 0
  const q = queryInput.value.trim().replace(/'/g, "'\\''")
  const cmds: Record<string, string> = {
    mysql: `mysql -e '${q}' 2>&1`, mariadb: `mysql -e '${q}' 2>&1`,
    postgresql: `sudo -u postgres psql -c '${q}' 2>&1`, redis: `redis-cli ${q} 2>&1`,
    mongodb: `mongosh --quiet --eval '${q}' 2>&1`
  }
  try {
    const start = Date.now()
    const r = await exec(cmds[selectedDb.value.type] || '')
    queryTime.value = Date.now() - start
    queryResult.value = r.stdout || r.stderr || '查询完成'
    if (r.stdout && ['mysql', 'mariadb'].includes(selectedDb.value.type)) {
      const lines = r.stdout.trim().split('\n')
      if (lines.length > 1) {
        queryResultColumns.value = lines[0].split('\t')
        queryResultData.value = lines.slice(1).map(l => {
          const vals = l.split('\t'); const row: Record<string, string> = {}
          queryResultColumns.value.forEach((c, i) => row[c] = vals[i] || ''); return row
        })
        queryResultRows.value = queryResultData.value.length
      }
    }
  } catch (e) { queryResult.value = '查询失败: ' + (e as Error).message }
  finally { queryLoading.value = false }
}

async function loadUsers() {
  if (!selectedDb.value) return
  userList.value = []; loading.value = true
  const cmds: Record<string, string> = {
    mysql: "mysql -N -e \"SELECT user, host, '' FROM mysql.user\" 2>/dev/null",
    mariadb: "mysql -N -e \"SELECT user, host, '' FROM mysql.user\" 2>/dev/null",
    postgresql: "sudo -u postgres psql -t -c \"SELECT usename, 'local', CASE WHEN usesuper THEN 'SUPERUSER' ELSE 'USER' END FROM pg_user\" 2>/dev/null",
    redis: "redis-cli ACL LIST 2>/dev/null", mongodb: "mongosh --quiet --eval 'db.getUsers().users.forEach(u=>print(u.user+\"\\t\"+u.db+\"\\t\"+u.roles.map(r=>r.role).join(\",\")))' 2>/dev/null"
  }
  try {
    const r = await exec(cmds[selectedDb.value.type] || '')
    if (r.stdout) r.stdout.trim().split('\n').filter(Boolean).forEach(line => {
      const p = line.split(/\t/).map(s => s.trim())
      userList.value.push({ user: p[0] || '', host: p[1] || '-', privileges: p[2] || '-' })
    })
  } catch {} finally { loading.value = false }
}

async function loadConfig() {
  if (!selectedDb.value) return
  const paths: Record<string, string> = {
    mysql: '/etc/mysql/my.cnf', mariadb: '/etc/mysql/mariadb.conf.d/50-server.cnf',
    postgresql: '/etc/postgresql/*/main/postgresql.conf', redis: '/etc/redis/redis.conf',
    mongodb: '/etc/mongod.conf'
  }
  configPath.value = paths[selectedDb.value.type] || ''
  try {
    const r = await exec(`cat ${configPath.value} 2>/dev/null | head -100`)
    configContent.value = r.stdout || '无法读取配置文件'
  } catch { configContent.value = '' }
}

async function createDatabase() {
  if (!selectedDb.value || !newDbName.value.trim()) return
  loading.value = true
  const cmds: Record<string, string> = {
    mysql: `mysql -e "CREATE DATABASE \\\`${newDbName.value}\\\` CHARACTER SET ${newDbCharset.value}" 2>&1`,
    mariadb: `mysql -e "CREATE DATABASE \\\`${newDbName.value}\\\` CHARACTER SET ${newDbCharset.value}" 2>&1`,
    postgresql: `sudo -u postgres createdb "${newDbName.value}" 2>&1`,
    mongodb: `mongosh --quiet --eval 'db.getSiblingDB("${newDbName.value}").createCollection("init")' 2>&1`
  }
  try {
    const r = await exec(cmds[selectedDb.value.type] || '')
    if (r.exit_code === 0) { ElMessage.success('创建成功'); showCreateDb.value = false; newDbName.value = ''; await loadDatabaseList() }
    else ElMessage.error(r.stdout || r.stderr || '创建失败')
  } finally { loading.value = false }
}

async function confirmDropDb(name: string) {
  await ElMessageBox.confirm(`确定删除数据库 "${name}"？此操作不可恢复！`, '危险操作', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  const cmds: Record<string, string> = {
    mysql: `mysql -e "DROP DATABASE \\\`${name}\\\`" 2>&1`, mariadb: `mysql -e "DROP DATABASE \\\`${name}\\\`" 2>&1`,
    postgresql: `sudo -u postgres dropdb "${name}" 2>&1`, mongodb: `mongosh --quiet --eval 'db.getSiblingDB("${name}").dropDatabase()' 2>&1`
  }
  const r = await exec(cmds[selectedDb.value!.type] || '')
  if (r.exit_code === 0) { ElMessage.success('已删除'); await loadDatabaseList() } else ElMessage.error(r.stderr || '删除失败')
}

async function createUser() {
  if (!selectedDb.value || !newUser.name) return
  loading.value = true
  const cmds: Record<string, string> = {
    mysql: `mysql -e "CREATE USER '${newUser.name}'@'${newUser.host}' IDENTIFIED BY '${newUser.password}'" 2>&1`,
    mariadb: `mysql -e "CREATE USER '${newUser.name}'@'${newUser.host}' IDENTIFIED BY '${newUser.password}'" 2>&1`,
    postgresql: `sudo -u postgres psql -c "CREATE USER ${newUser.name} WITH PASSWORD '${newUser.password}'" 2>&1`,
    mongodb: `mongosh --quiet --eval 'db.createUser({user:"${newUser.name}",pwd:"${newUser.password}",roles:["readWrite"]})' 2>&1`
  }
  try {
    const r = await exec(cmds[selectedDb.value.type] || '')
    if (r.exit_code === 0) { ElMessage.success('用户创建成功'); showCreateUser.value = false; newUser.name = ''; newUser.password = ''; await loadUsers() }
    else ElMessage.error(r.stdout || r.stderr || '创建失败')
  } finally { loading.value = false }
}

async function confirmDropUser(row: any) {
  await ElMessageBox.confirm(`确定删除用户 "${row.user}"？`, '确认', { type: 'warning' })
  const cmds: Record<string, string> = {
    mysql: `mysql -e "DROP USER '${row.user}'@'${row.host}'" 2>&1`, mariadb: `mysql -e "DROP USER '${row.user}'@'${row.host}'" 2>&1`,
    postgresql: `sudo -u postgres psql -c "DROP USER ${row.user}" 2>&1`, mongodb: `mongosh --quiet --eval 'db.dropUser("${row.user}")' 2>&1`
  }
  const r = await exec(cmds[selectedDb.value!.type] || '')
  if (r.exit_code === 0) { ElMessage.success('已删除'); await loadUsers() } else ElMessage.error(r.stderr || '删除失败')
}

async function showTables(dbName: string) {
  currentDbName.value = dbName; tableList.value = []; showTablesDialog.value = true
  const cmds: Record<string, string> = {
    mysql: `mysql -N -e "SELECT table_name, table_rows, ROUND((data_length+index_length)/1024,2) FROM information_schema.tables WHERE table_schema='${dbName}'" 2>/dev/null`,
    mariadb: `mysql -N -e "SELECT table_name, table_rows, ROUND((data_length+index_length)/1024,2) FROM information_schema.tables WHERE table_schema='${dbName}'" 2>/dev/null`,
    postgresql: `sudo -u postgres psql -d ${dbName} -t -c "SELECT tablename, n_live_tup, pg_size_pretty(pg_total_relation_size(tablename::text)) FROM pg_stat_user_tables" 2>/dev/null`,
    mongodb: `mongosh --quiet ${dbName} --eval 'db.getCollectionNames().forEach(c=>{const s=db[c].stats();print(c+"\\t"+s.count+"\\t"+Math.round(s.size/1024)+" KB")})' 2>/dev/null`
  }
  try {
    const r = await exec(cmds[selectedDb.value!.type] || '')
    if (r.stdout) r.stdout.trim().split('\n').filter(Boolean).forEach(line => {
      const p = line.split(/\t/).map(s => s.trim()).filter(Boolean)
      tableList.value.push({ name: p[0] || '', rows: p[1] || '-', size: p[2] ? (p[2].includes('KB') || p[2].includes('MB') ? p[2] : p[2] + ' KB') : '-' })
    })
  } catch { ElMessage.error('加载失败') }
}

// 切换 tab 时加载数据
import { watch } from 'vue'
watch(activeTab, (tab) => {
  if (tab === 'users') loadUsers()
  if (tab === 'config') loadConfig()
})
</script>

<style lang="scss" scoped>
@keyframes fadeSlideUp { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: translateY(0); } }
@keyframes cardIn { from { opacity: 0; transform: scale(0.96) translateY(10px); } to { opacity: 1; transform: scale(1) translateY(0); } }
.animate-in { animation: fadeSlideUp 0.4s ease both; }
.animate-card { animation: cardIn 0.35s ease both; }

.database-page { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px;
  .header-left { h1 { font-size: 22px; font-weight: 700; margin: 0 0 4px; } .subtitle { color: var(--text-secondary); font-size: 13px; margin: 0; } }
  .header-right { display: flex; gap: 10px; .server-select { width: 200px; } }
}

.db-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; margin-bottom: 20px; }

.db-card {
  display: flex; align-items: center; gap: 12px; padding: 14px 16px;
  background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px;
  cursor: pointer; transition: all 0.25s cubic-bezier(.4,0,.2,1);
  &:hover { border-color: var(--text-muted); transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.15); }
  &.active { border-color: var(--primary-color); background: linear-gradient(135deg, var(--bg-secondary), rgba(99,102,241,0.06)); }
  &.running .db-icon { box-shadow: 0 0 0 2px rgba(34,197,94,0.3); }
}

.db-icon {
  width: 44px; height: 44px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  box-shadow: 0 3px 10px rgba(0,0,0,0.15);
}

.db-body { flex: 1; min-width: 0; }
.db-name { font-size: 14px; font-weight: 600; margin-bottom: 4px; }
.db-meta { display: flex; align-items: center; gap: 8px; }
.db-ver { font-size: 11px; color: var(--text-muted); }
.db-action { flex-shrink: 0; }

.detail-panel {
  background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px; overflow: hidden;
}

.panel-header {
  display: flex; justify-content: space-between; align-items: center; padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.panel-title {
  display: flex; align-items: center; gap: 12px;
  h2 { font-size: 16px; font-weight: 600; margin: 0; }
  .panel-ver { font-size: 12px; color: var(--text-muted); }
}

.panel-icon { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.panel-body { padding: 20px; min-height: 300px; }

.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 20px; }
.stat-card {
  padding: 16px; background: var(--bg-tertiary); border-radius: 10px; text-align: center;
  .stat-icon { width: 40px; height: 40px; border-radius: 10px; display: inline-flex; align-items: center; justify-content: center; font-size: 18px; margin-bottom: 8px; }
  .stat-val { font-size: 22px; font-weight: 700; margin-bottom: 2px; }
  .stat-lbl { font-size: 12px; color: var(--text-secondary); }
}

.info-section {
  h3 { font-size: 14px; font-weight: 600; margin-bottom: 12px; }
}

.info-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 8px;
  .info-item { display: flex; justify-content: space-between; padding: 8px 12px; background: var(--bg-tertiary); border-radius: 6px; font-size: 13px;
    .info-key { color: var(--text-secondary); }
    .info-val { font-family: 'Consolas', monospace; }
  }
}

.tab-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.db-name-cell { font-family: 'Consolas', monospace; font-size: 13px; }

.query-area { margin-bottom: 16px; }
.query-input { font-family: 'Fira Code', 'Consolas', monospace; margin-bottom: 10px; }
.query-toolbar { display: flex; align-items: center; gap: 8px; .query-hint { margin-left: auto; font-size: 12px; color: var(--text-muted); } }
.query-result {
  .result-header { display: flex; justify-content: space-between; font-size: 13px; color: var(--text-secondary); margin-bottom: 10px;
    .result-count { color: var(--primary-color); font-weight: 500; }
  }
  .result-text { background: var(--bg-tertiary); padding: 12px; border-radius: 8px; font-family: monospace; font-size: 13px; white-space: pre-wrap; max-height: 400px; overflow: auto; }
}

.config-section {
  margin-bottom: 20px;
  h3 { font-size: 14px; font-weight: 600; margin-bottom: 12px; }
}
.config-file { .config-path { font-size: 12px; color: var(--text-muted); margin-bottom: 8px; font-family: monospace; } }
.config-text { background: var(--bg-tertiary); padding: 14px; border-radius: 8px; font-family: 'Fira Code', monospace; font-size: 12px; max-height: 400px; overflow: auto; white-space: pre-wrap; line-height: 1.6; }
.quick-actions { display: flex; gap: 8px; }

:deep(.el-table) { --el-table-bg-color: transparent; --el-table-tr-bg-color: transparent; }
</style>
