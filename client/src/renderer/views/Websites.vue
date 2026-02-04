<template>
  <div class="websites">
    <div class="page-header">
      <div class="header-left">
        <h1>网站管理</h1>
        <p class="subtitle">站点配置与项目部署</p>
      </div>
      <div class="header-actions">
        <el-select v-if="hasMultipleServers" v-model="selectedServer" placeholder="选择服务器" size="small">
          <el-option v-for="s in connectedServers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="refresh" :loading="loading" size="small">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
        <el-dropdown @command="handleAddCommand">
          <el-button type="primary" size="small">
            <el-icon><Plus /></el-icon>添加<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="static">静态站点</el-dropdown-item>
              <el-dropdown-item command="project">项目部署</el-dropdown-item>
              <el-dropdown-item command="proxy">反向代理</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <!-- 标签页 -->
      <el-tabs v-model="activeTab" class="main-tabs">
        <el-tab-pane name="sites">
          <template #label>
            <span class="tab-label">站点列表 <el-badge :value="sites.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="projects">
          <template #label>
            <span class="tab-label">项目部署 <el-badge :value="projects.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <!-- 站点列表 -->
      <div v-show="activeTab === 'sites'" class="tab-content">
        <el-table :data="sites" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="站点名称" min-width="140">
            <template #default="{ row }">
              <div class="cell-name">
                <span class="status-dot" :class="row.status"></span>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="domain" label="域名" min-width="180">
            <template #default="{ row }">
              <a :href="(row.ssl ? 'https://' : 'http://') + row.domain" target="_blank" class="domain-link">
                <el-icon v-if="row.ssl"><Lock /></el-icon>
                {{ row.domain }}
              </a>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="getTypeTag(row.type)">{{ getTypeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="目录" min-width="160">
            <template #default="{ row }">
              <code class="mono">{{ row.path }}</code>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button @click="openSite(row)">打开</el-button>
                <el-button @click="editSite(row)">设置</el-button>
                <el-button v-if="row.status === 'running'" type="warning" @click="toggleSite(row, 'stop')">停止</el-button>
                <el-button v-else type="success" @click="toggleSite(row, 'start')">启动</el-button>
                <el-button type="danger" @click="deleteSite(row)">删除</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 项目部署 -->
      <div v-show="activeTab === 'projects'" class="tab-content">
        <div v-if="projects.length === 0" class="empty-projects">
          <el-empty description="暂无部署项目">
            <el-button type="primary" size="small" @click="handleAddCommand('project')">创建项目部署</el-button>
          </el-empty>
        </div>
        <div v-else class="projects-grid">
          <div v-for="project in projects" :key="project.id" class="project-card">
            <div class="project-header">
              <div class="project-icon" :style="{ background: getProjectColor(project.type) }">
                <TechIcon :name="project.type" />
              </div>
              <div class="project-info">
                <div class="project-name">{{ project.name }}</div>
                <div class="project-domain">{{ project.domain }}</div>
              </div>
              <el-tag :type="getProjectStatusType(project.status)" size="small">{{ getProjectStatusLabel(project.status) }}</el-tag>
            </div>
            <div class="project-meta">
              <div class="meta-item"><span class="meta-label">目录:</span> <code>{{ project.path }}</code></div>
              <div class="meta-item"><span class="meta-label">端口:</span> {{ project.port }}</div>
              <div class="meta-item" v-if="project.lastDeploy"><span class="meta-label">上次部署:</span> {{ formatTime(project.lastDeploy) }}</div>
            </div>
            <div class="project-actions">
              <el-button size="small" type="primary" @click="deployProject(project)" :loading="project.deploying">
                {{ project.deploying ? '部署中' : '部署' }}
              </el-button>
              <el-button size="small" @click="viewProjectLogs(project)">日志</el-button>
              <el-button size="small" @click="editProject(project)">设置</el-button>
              <el-button size="small" v-if="project.status === 'running'" type="warning" @click="stopProject(project)">停止</el-button>
              <el-button size="small" v-else type="success" @click="startProject(project)">启动</el-button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- 添加静态站点对话框 -->
    <el-dialog v-model="showAddStatic" title="添加静态站点" width="500px" class="dark-dialog">
      <el-form :model="newSite" label-width="80px" size="small">
        <el-form-item label="站点名称" required>
          <el-input v-model="newSite.name" placeholder="my-website" />
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="newSite.domain" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="根目录" required>
          <el-input v-model="newSite.path" placeholder="/var/www/html" />
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="newSite.ssl" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showAddStatic = false">取消</el-button>
        <el-button type="primary" size="small" @click="createStaticSite" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 添加反向代理对话框 -->
    <el-dialog v-model="showAddProxy" title="添加反向代理" width="500px" class="dark-dialog">
      <el-form :model="newProxy" label-width="80px" size="small">
        <el-form-item label="站点名称" required>
          <el-input v-model="newProxy.name" placeholder="my-api" />
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="newProxy.domain" placeholder="api.example.com" />
        </el-form-item>
        <el-form-item label="代理地址" required>
          <el-input v-model="newProxy.upstream" placeholder="http://127.0.0.1:3000" />
        </el-form-item>
        <el-form-item label="WebSocket">
          <el-switch v-model="newProxy.websocket" />
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="newProxy.ssl" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showAddProxy = false">取消</el-button>
        <el-button type="primary" size="small" @click="createProxySite" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 项目部署对话框 -->
    <el-dialog v-model="showAddProject" title="项目部署" width="650px" class="dark-dialog">
      <el-form :model="newProject" label-width="100px" size="small">
        <el-form-item label="项目名称" required>
          <el-input v-model="newProject.name" placeholder="my-app" />
        </el-form-item>
        <el-form-item label="项目类型" required>
          <el-select v-model="newProject.type" style="width: 100%" @change="onProjectTypeChange">
            <el-option value="nodejs" label="Node.js" />
            <el-option value="python" label="Python" />
            <el-option value="go" label="Go" />
            <el-option value="java" label="Java" />
            <el-option value="php" label="PHP" />
            <el-option value="static-build" label="静态构建 (Vue/React)" />
          </el-select>
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="newProject.domain" placeholder="app.example.com" />
        </el-form-item>
        <el-form-item label="项目目录" required>
          <el-input v-model="newProject.path" placeholder="/var/www/my-app" />
        </el-form-item>
        <el-form-item label="运行端口" v-if="!['php', 'static-build'].includes(newProject.type)">
          <el-input-number v-model="newProject.port" :min="1024" :max="65535" />
        </el-form-item>

        <el-divider content-position="left">构建工作流</el-divider>

        <el-form-item label="构建步骤">
          <div class="workflow-steps">
            <div v-for="(step, index) in newProject.buildSteps" :key="index" class="workflow-step">
              <el-input v-model="step.command" placeholder="npm install" style="flex: 1" />
              <el-button text type="danger" @click="removeBuildStep(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" @click="addBuildStep">
              <el-icon><Plus /></el-icon> 添加步骤
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="启动命令" v-if="!['php', 'static-build'].includes(newProject.type)">
          <el-input v-model="newProject.startCommand" :placeholder="getDefaultStartCommand(newProject.type)" />
        </el-form-item>

        <el-form-item label="输出目录" v-if="newProject.type === 'static-build'">
          <el-input v-model="newProject.outputDir" placeholder="dist" />
        </el-form-item>

        <el-form-item label="环境变量">
          <div class="env-vars">
            <div v-for="(env, index) in newProject.envVars" :key="index" class="env-var-row">
              <el-input v-model="env.key" placeholder="KEY" style="width: 120px" />
              <span class="env-eq">=</span>
              <el-input v-model="env.value" placeholder="value" style="flex: 1" />
              <el-button text type="danger" @click="removeEnvVar(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" @click="addEnvVar">
              <el-icon><Plus /></el-icon> 添加变量
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="启用 SSL">
          <el-switch v-model="newProject.ssl" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showAddProject = false">取消</el-button>
        <el-button type="primary" size="small" @click="createProject" :loading="creating">创建项目</el-button>
      </template>
    </el-dialog>

    <!-- 部署日志对话框 -->
    <el-dialog v-model="showDeployLog" :title="`部署日志 - ${currentProject?.name}`" width="80%" top="5vh" class="dark-dialog">
      <div class="deploy-log">
        <pre ref="logPre">{{ deployLog }}</pre>
      </div>
      <template #footer>
        <el-button size="small" @click="showDeployLog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 站点设置对话框 -->
    <el-dialog v-model="showSiteSettings" :title="`站点设置 - ${currentSite?.name}`" width="600px" class="dark-dialog">
      <el-form :model="currentSite" label-width="100px" size="small" v-if="currentSite">
        <el-form-item label="域名">
          <el-input v-model="currentSite.domain" />
        </el-form-item>
        <el-form-item label="根目录">
          <el-input v-model="currentSite.path" />
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="currentSite.ssl" />
        </el-form-item>
        <el-form-item label="伪静态">
          <div class="rewrite-presets">
            <el-button size="small" @click="applyRewrite('vue')">Vue/React</el-button>
            <el-button size="small" @click="applyRewrite('laravel')">Laravel</el-button>
            <el-button size="small" @click="applyRewrite('wordpress')">WordPress</el-button>
          </div>
          <el-input type="textarea" v-model="currentSite.rewrite" :rows="6" class="code-input" placeholder="location / { try_files $uri $uri/ /index.html; }" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showSiteSettings = false">取消</el-button>
        <el-button type="primary" size="small" @click="saveSiteSettings" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 项目设置对话框 -->
    <el-dialog v-model="showProjectSettings" :title="`项目设置 - ${currentProject?.name}`" width="600px" class="dark-dialog">
      <el-form :model="currentProject" label-width="100px" size="small" v-if="currentProject">
        <el-form-item label="域名">
          <el-input v-model="currentProject.domain" />
        </el-form-item>
        <el-form-item label="项目目录">
          <el-input v-model="currentProject.path" />
        </el-form-item>
        <el-form-item label="运行端口" v-if="!['php', 'static-build'].includes(currentProject.type)">
          <el-input-number v-model="currentProject.port" :min="1024" :max="65535" />
        </el-form-item>
        <el-form-item label="构建步骤">
          <div class="workflow-steps">
            <div v-for="(step, index) in currentProject.buildSteps" :key="index" class="workflow-step">
              <el-input v-model="step.command" style="flex: 1" />
              <el-button text type="danger" @click="currentProject.buildSteps.splice(index, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" @click="currentProject.buildSteps.push({ command: '' })">
              <el-icon><Plus /></el-icon> 添加步骤
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="启动命令" v-if="!['php', 'static-build'].includes(currentProject.type)">
          <el-input v-model="currentProject.startCommand" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showProjectSettings = false">取消</el-button>
        <el-button type="danger" size="small" @click="deleteProject">删除项目</el-button>
        <el-button type="primary" size="small" @click="saveProjectSettings" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Lock, Delete, ArrowDown } from '@element-plus/icons-vue'
import TechIcon from '@/components/icons/TechIcons.vue'

interface Site {
  id: string
  name: string
  domain: string
  path: string
  type: string
  status: string
  ssl: boolean
  rewrite?: string
}

interface BuildStep {
  command: string
}

interface EnvVar {
  key: string
  value: string
}

interface Project {
  id: string
  name: string
  type: string
  domain: string
  path: string
  port: number
  status: string
  ssl: boolean
  buildSteps: BuildStep[]
  startCommand: string
  outputDir?: string
  envVars: EnvVar[]
  lastDeploy?: number
  deploying?: boolean
}

const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const activeTab = ref('sites')
const loading = ref(false)
const creating = ref(false)
const saving = ref(false)

// 数据
const sites = ref<Site[]>([])
const projects = ref<Project[]>([])

// 对话框
const showAddStatic = ref(false)
const showAddProxy = ref(false)
const showAddProject = ref(false)
const showDeployLog = ref(false)
const showSiteSettings = ref(false)
const showProjectSettings = ref(false)
const currentSite = ref<Site | null>(null)
const currentProject = ref<Project | null>(null)
const deployLog = ref('')
const logPre = ref<HTMLPreElement | null>(null)

// 表单
const newSite = ref({ name: '', domain: '', path: '/var/www', ssl: false })
const newProxy = ref({ name: '', domain: '', upstream: 'http://127.0.0.1:3000', websocket: false, ssl: false })
const newProject = ref<{
  name: string; type: string; domain: string; path: string; port: number; ssl: boolean;
  buildSteps: BuildStep[]; startCommand: string; outputDir: string; envVars: EnvVar[]
}>({
  name: '', type: 'nodejs', domain: '', path: '/var/www', port: 3000, ssl: false,
  buildSteps: [{ command: 'npm install' }, { command: 'npm run build' }],
  startCommand: 'npm start', outputDir: 'dist', envVars: []
})

const connectedServers = computed(() => serverStore.connectedServers)
const hasMultipleServers = computed(() => serverStore.hasMultipleServers)

watch(selectedServer, (val) => {
  if (val) loadData()
})

onMounted(() => {
  if (connectedServers.value.length > 0) {
    selectedServer.value = serverStore.currentServerId || connectedServers.value[0].id
  }
  loadProjectsFromStorage()
})

function loadProjectsFromStorage() {
  const saved = localStorage.getItem('serverhub_projects')
  if (saved) {
    projects.value = JSON.parse(saved)
  }
}

function saveProjectsToStorage() {
  localStorage.setItem('serverhub_projects', JSON.stringify(projects.value))
}

async function loadData() {
  await loadSites()
}

async function loadSites() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'ls -1 /etc/nginx/sites-enabled/ 2>/dev/null || ls -1 /etc/nginx/conf.d/*.conf 2>/dev/null']
    )
    const stdout = result.stdout || ''
    const files = stdout.trim().split('\n').filter(f => f && !f.includes('default'))
    sites.value = files.map((f, i) => ({
      id: `site_${i}`,
      name: f.replace('.conf', '').replace(/^.*\//, ''),
      domain: f.replace('.conf', '').replace(/^.*\//, ''),
      path: '/var/www/' + f.replace('.conf', '').replace(/^.*\//, ''),
      type: 'static',
      status: 'running',
      ssl: false
    }))
  } catch { sites.value = [] }
  finally { loading.value = false }
}

function refresh() { loadData() }

function handleAddCommand(cmd: string) {
  if (cmd === 'static') showAddStatic.value = true
  else if (cmd === 'proxy') showAddProxy.value = true
  else if (cmd === 'project') {
    resetNewProject()
    showAddProject.value = true
  }
}

function resetNewProject() {
  newProject.value = {
    name: '', type: 'nodejs', domain: '', path: '/var/www', port: 3000, ssl: false,
    buildSteps: [{ command: 'npm install' }, { command: 'npm run build' }],
    startCommand: 'npm start', outputDir: 'dist', envVars: []
  }
}

function onProjectTypeChange(type: string) {
  const defaults: Record<string, { buildSteps: BuildStep[]; startCommand: string; port: number }> = {
    nodejs: { buildSteps: [{ command: 'npm install' }, { command: 'npm run build' }], startCommand: 'npm start', port: 3000 },
    python: { buildSteps: [{ command: 'pip install -r requirements.txt' }], startCommand: 'python app.py', port: 5000 },
    go: { buildSteps: [{ command: 'go build -o app' }], startCommand: './app', port: 8080 },
    java: { buildSteps: [{ command: 'mvn package' }], startCommand: 'java -jar target/*.jar', port: 8080 },
    php: { buildSteps: [{ command: 'composer install' }], startCommand: '', port: 0 },
    'static-build': { buildSteps: [{ command: 'npm install' }, { command: 'npm run build' }], startCommand: '', port: 0 }
  }
  const d = defaults[type] || defaults.nodejs
  newProject.value.buildSteps = d.buildSteps
  newProject.value.startCommand = d.startCommand
  newProject.value.port = d.port
}

function getDefaultStartCommand(type: string): string {
  const cmds: Record<string, string> = {
    nodejs: 'npm start', python: 'python app.py', go: './app', java: 'java -jar target/*.jar'
  }
  return cmds[type] || ''
}

function addBuildStep() { newProject.value.buildSteps.push({ command: '' }) }
function removeBuildStep(index: number) { newProject.value.buildSteps.splice(index, 1) }
function addEnvVar() { newProject.value.envVars.push({ key: '', value: '' }) }
function removeEnvVar(index: number) { newProject.value.envVars.splice(index, 1) }

async function createStaticSite() {
  if (!selectedServer.value || !newSite.value.name || !newSite.value.domain) {
    ElMessage.warning('请填写完整信息'); return
  }
  creating.value = true
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo mkdir -p ${newSite.value.path}`])
    const config = generateStaticConfig(newSite.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newSite.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newSite.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    ElMessage.success('站点创建成功')
    showAddStatic.value = false
    newSite.value = { name: '', domain: '', path: '/var/www', ssl: false }
    loadSites()
  } catch (e) { ElMessage.error('创建失败: ' + (e as Error).message) }
  finally { creating.value = false }
}

async function createProxySite() {
  if (!selectedServer.value || !newProxy.value.name || !newProxy.value.domain) {
    ElMessage.warning('请填写完整信息'); return
  }
  creating.value = true
  try {
    const config = generateProxyConfig(newProxy.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newProxy.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newProxy.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    ElMessage.success('反向代理创建成功')
    showAddProxy.value = false
    newProxy.value = { name: '', domain: '', upstream: 'http://127.0.0.1:3000', websocket: false, ssl: false }
    loadSites()
  } catch (e) { ElMessage.error('创建失败: ' + (e as Error).message) }
  finally { creating.value = false }
}

async function createProject() {
  if (!selectedServer.value || !newProject.value.name || !newProject.value.domain) {
    ElMessage.warning('请填写完整信息'); return
  }
  creating.value = true
  try {
    // 创建项目目录
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo mkdir -p ${newProject.value.path}`])
    
    // 生成 Nginx 配置
    const config = generateProjectConfig(newProject.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newProject.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newProject.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    
    // 保存项目配置
    const project: Project = {
      id: `project_${Date.now()}`,
      name: newProject.value.name,
      type: newProject.value.type,
      domain: newProject.value.domain,
      path: newProject.value.path,
      port: newProject.value.port,
      status: 'stopped',
      ssl: newProject.value.ssl,
      buildSteps: [...newProject.value.buildSteps],
      startCommand: newProject.value.startCommand,
      outputDir: newProject.value.outputDir,
      envVars: [...newProject.value.envVars]
    }
    projects.value.push(project)
    saveProjectsToStorage()
    
    ElMessage.success('项目创建成功')
    showAddProject.value = false
    activeTab.value = 'projects'
  } catch (e) { ElMessage.error('创建失败: ' + (e as Error).message) }
  finally { creating.value = false }
}

async function deployProject(project: Project) {
  if (!selectedServer.value) return
  project.deploying = true
  deployLog.value = `开始部署 ${project.name}...\n\n`
  showDeployLog.value = true
  currentProject.value = project

  try {
    // 执行构建步骤
    for (const step of project.buildSteps) {
      if (!step.command.trim()) continue
      deployLog.value += `> ${step.command}\n`
      await nextTick()
      if (logPre.value) logPre.value.scrollTop = logPre.value.scrollHeight

      const envStr = project.envVars.map(e => `${e.key}=${e.value}`).join(' ')
      const cmd = envStr ? `cd ${project.path} && ${envStr} ${step.command}` : `cd ${project.path} && ${step.command}`
      const result = await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', cmd])
      
      deployLog.value += (result.stdout || '') + '\n'
      if (result.stderr) deployLog.value += result.stderr + '\n'
      
      if (result.exit_code !== 0) {
        deployLog.value += `\n✗ 步骤失败 (退出码: ${result.exit_code})\n`
        ElMessage.error('部署失败')
        project.deploying = false
        return
      }
    }

    // 启动服务（非静态项目）
    if (!['php', 'static-build'].includes(project.type) && project.startCommand) {
      deployLog.value += `\n> 启动服务: ${project.startCommand}\n`
      
      // 先停止旧进程
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
        `pkill -f "node.*${project.path}" || true`])
      
      // 使用 systemd 或 pm2 启动
      const serviceName = `serverhub-${project.name}`
      const envStr = project.envVars.map(e => `Environment="${e.key}=${e.value}"`).join('\n')
      const serviceContent = `[Unit]
Description=${project.name}
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${project.path}
ExecStart=/bin/bash -c '${project.startCommand}'
Restart=on-failure
${envStr}

[Install]
WantedBy=multi-user.target`
      
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
        `echo '${serviceContent.replace(/'/g, "'\\''")}' | sudo tee /etc/systemd/system/${serviceName}.service`])
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
        `sudo systemctl daemon-reload && sudo systemctl enable ${serviceName} && sudo systemctl restart ${serviceName}`])
      
      deployLog.value += `服务已启动: ${serviceName}\n`
      project.status = 'running'
    } else if (project.type === 'static-build') {
      // 静态构建项目，复制到 Nginx 目录
      const outputPath = `${project.path}/${project.outputDir || 'dist'}`
      deployLog.value += `\n> 部署静态文件: ${outputPath}\n`
      project.status = 'running'
    }

    project.lastDeploy = Date.now()
    saveProjectsToStorage()
    deployLog.value += '\n✓ 部署成功！\n'
    ElMessage.success('部署成功')
  } catch (e) {
    deployLog.value += `\n错误: ${(e as Error).message}\n`
    ElMessage.error('部署失败')
  } finally {
    project.deploying = false
  }
}

async function startProject(project: Project) {
  if (!selectedServer.value) return
  try {
    const serviceName = `serverhub-${project.name}`
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo systemctl start ${serviceName}`])
    project.status = 'running'
    saveProjectsToStorage()
    ElMessage.success('项目已启动')
  } catch (e) { ElMessage.error('启动失败') }
}

async function stopProject(project: Project) {
  if (!selectedServer.value) return
  try {
    const serviceName = `serverhub-${project.name}`
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo systemctl stop ${serviceName}`])
    project.status = 'stopped'
    saveProjectsToStorage()
    ElMessage.success('项目已停止')
  } catch (e) { ElMessage.error('停止失败') }
}

function viewProjectLogs(project: Project) {
  currentProject.value = project
  deployLog.value = '加载日志中...'
  showDeployLog.value = true
  loadProjectLogs(project)
}

async function loadProjectLogs(project: Project) {
  if (!selectedServer.value) return
  try {
    const serviceName = `serverhub-${project.name}`
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo journalctl -u ${serviceName} -n 100 --no-pager`])
    deployLog.value = result.stdout || '无日志'
  } catch { deployLog.value = '获取日志失败' }
}

function editProject(project: Project) {
  currentProject.value = { ...project, buildSteps: [...project.buildSteps], envVars: [...project.envVars] }
  showProjectSettings.value = true
}

async function saveProjectSettings() {
  if (!currentProject.value) return
  saving.value = true
  try {
    const index = projects.value.findIndex(p => p.id === currentProject.value!.id)
    if (index !== -1) {
      projects.value[index] = { ...currentProject.value }
      saveProjectsToStorage()
    }
    ElMessage.success('设置已保存')
    showProjectSettings.value = false
  } finally { saving.value = false }
}

async function deleteProject() {
  if (!currentProject.value || !selectedServer.value) return
  try {
    await ElMessageBox.confirm(`确定删除项目 ${currentProject.value.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  
  try {
    const serviceName = `serverhub-${currentProject.value.name}`
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo systemctl stop ${serviceName} || true; sudo systemctl disable ${serviceName} || true; sudo rm -f /etc/systemd/system/${serviceName}.service`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo rm -f /etc/nginx/sites-enabled/${currentProject.value.name} /etc/nginx/sites-available/${currentProject.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    
    projects.value = projects.value.filter(p => p.id !== currentProject.value!.id)
    saveProjectsToStorage()
    showProjectSettings.value = false
    ElMessage.success('项目已删除')
  } catch (e) { ElMessage.error('删除失败') }
}

function editSite(site: Site) {
  currentSite.value = { ...site }
  showSiteSettings.value = true
}

async function saveSiteSettings() {
  if (!currentSite.value || !selectedServer.value) return
  saving.value = true
  try {
    ElMessage.success('设置已保存')
    showSiteSettings.value = false
    loadSites()
  } finally { saving.value = false }
}

function openSite(site: Site) {
  const url = site.ssl ? `https://${site.domain}` : `http://${site.domain}`
  window.electronAPI.shell.openExternal(url)
}

async function toggleSite(site: Site, action: 'start' | 'stop') {
  if (!selectedServer.value) return
  try {
    if (action === 'stop') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo rm -f /etc/nginx/sites-enabled/${site.name}`])
    } else {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo ln -sf /etc/nginx/sites-available/${site.name} /etc/nginx/sites-enabled/`])
    }
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    ElMessage.success(action === 'stop' ? '站点已停止' : '站点已启动')
    loadSites()
  } catch { ElMessage.error('操作失败') }
}

async function deleteSite(site: Site) {
  try { await ElMessageBox.confirm(`确定删除站点 ${site.name}？`, '确认删除', { type: 'warning' }) }
  catch { return }
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo rm -f /etc/nginx/sites-enabled/${site.name} /etc/nginx/sites-available/${site.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    ElMessage.success('站点已删除')
    loadSites()
  } catch { ElMessage.error('删除失败') }
}

function applyRewrite(preset: string) {
  if (!currentSite.value) return
  const presets: Record<string, string> = {
    vue: 'location / { try_files $uri $uri/ /index.html; }',
    laravel: 'location / { try_files $uri $uri/ /index.php?$query_string; }',
    wordpress: 'location / { try_files $uri $uri/ /index.php?$args; }'
  }
  currentSite.value.rewrite = presets[preset] || ''
}

// 配置生成
function generateStaticConfig(site: { name: string; domain: string; path: string; ssl: boolean }): string {
  return `server {
    listen 80;
    server_name ${site.domain};
    root ${site.path};
    index index.html index.htm;
    
    location / {
        try_files $uri $uri/ =404;
    }
    
    location ~ /\\. { deny all; }
}`
}

function generateProxyConfig(proxy: { name: string; domain: string; upstream: string; websocket: boolean; ssl: boolean }): string {
  return `server {
    listen 80;
    server_name ${proxy.domain};
    
    location / {
        proxy_pass ${proxy.upstream};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        ${proxy.websocket ? `proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";` : ''}
    }
}`
}

function generateProjectConfig(project: { name: string; domain: string; path: string; port: number; type: string; outputDir?: string }): string {
  if (project.type === 'static-build') {
    return `server {
    listen 80;
    server_name ${project.domain};
    root ${project.path}/${project.outputDir || 'dist'};
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location ~ /\\. { deny all; }
}`
  }
  if (project.type === 'php') {
    return `server {
    listen 80;
    server_name ${project.domain};
    root ${project.path}/public;
    index index.php index.html;
    
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    
    location ~ \\.php$ {
        fastcgi_pass unix:/var/run/php/php8.2-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    
    location ~ /\\. { deny all; }
}`
  }
  return `server {
    listen 80;
    server_name ${project.domain};
    
    location / {
        proxy_pass http://127.0.0.1:${project.port};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}`
}

// 辅助函数
function getTypeTag(type: string): 'success' | 'warning' | 'info' | 'danger' | undefined {
  const map: Record<string, 'success' | 'warning' | 'info' | 'danger' | undefined> = {
    static: undefined, php: 'warning', node: 'success', python: 'info', java: 'danger', proxy: undefined
  }
  return map[type]
}

function getTypeLabel(type: string): string {
  const labels: Record<string, string> = { static: '静态', php: 'PHP', node: 'Node', python: 'Python', java: 'Java', proxy: '代理' }
  return labels[type] || type
}

function getProjectColor(type: string): string {
  const colors: Record<string, string> = {
    nodejs: '#68a063', python: '#3776ab', go: '#00add8', java: '#f89820', php: '#777bb4', 'static-build': '#42b883'
  }
  return colors[type] || '#6366f1'
}

function getProjectStatusType(status: string): 'success' | 'danger' | 'info' {
  return status === 'running' ? 'success' : status === 'error' ? 'danger' : 'info'
}

function getProjectStatusLabel(status: string): string {
  const labels: Record<string, string> = { running: '运行中', stopped: '已停止', error: '错误' }
  return labels[status] || status
}

function formatTime(ts: number): string {
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
.websites {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;

  .header-left {
    h1 { font-size: 20px; font-weight: 600; margin-bottom: 4px; }
    .subtitle { color: var(--text-secondary); font-size: 13px; }
  }

  .header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }
}

.empty-state { padding: 60px 0; }

.main-tabs { margin-bottom: 16px; }
.tab-label { display: flex; align-items: center; gap: 6px; }

.tab-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}

.data-table {
  .cell-name {
    display: flex;
    align-items: center;
    gap: 8px;

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      &.running { background: #22c55e; }
      &.stopped { background: #ef4444; }
    }
  }

  .domain-link {
    color: var(--primary-color);
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 4px;
    &:hover { text-decoration: underline; }
  }

  .mono {
    font-family: 'Consolas', monospace;
    font-size: 12px;
    background: var(--bg-tertiary);
    padding: 2px 6px;
    border-radius: 4px;
  }
}

.empty-projects { padding: 40px 0; }

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}

.project-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;

  .project-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;

    .project-icon {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      :deep(svg) { width: 24px; height: 24px; }
    }

    .project-info {
      flex: 1;
      .project-name { font-weight: 600; font-size: 14px; }
      .project-domain { font-size: 12px; color: var(--text-secondary); }
    }
  }

  .project-meta {
    margin-bottom: 12px;
    .meta-item {
      font-size: 12px;
      color: var(--text-secondary);
      margin-bottom: 4px;
      .meta-label { color: var(--text-color); }
      code { background: var(--bg-secondary); padding: 1px 4px; border-radius: 3px; font-size: 11px; }
    }
  }

  .project-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
}

.workflow-steps {
  .workflow-step {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
  }
}

.env-vars {
  .env-var-row {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 8px;
    .env-eq { color: var(--text-secondary); }
  }
}

.rewrite-presets {
  margin-bottom: 8px;
  display: flex;
  gap: 8px;
}

.code-input {
  :deep(.el-textarea__inner) {
    font-family: 'Consolas', monospace;
    font-size: 12px;
    background: var(--bg-tertiary);
  }
}

.deploy-log {
  background: #1a1a1a;
  border-radius: 6px;
  padding: 12px;
  max-height: 500px;
  overflow: auto;

  pre {
    margin: 0;
    font-size: 12px;
    color: #d4d4d4;
    white-space: pre-wrap;
    word-break: break-all;
    font-family: 'Consolas', monospace;
  }
}

:deep(.dark-dialog) {
  .el-dialog { background: var(--bg-secondary) !important; }
  .el-dialog__header { background: var(--bg-secondary); border-bottom: 1px solid var(--border-color); }
  .el-dialog__title { color: var(--text-color); }
  .el-dialog__body { background: var(--bg-secondary); }
  .el-dialog__footer { background: var(--bg-secondary); border-top: 1px solid var(--border-color); }
}
</style>
