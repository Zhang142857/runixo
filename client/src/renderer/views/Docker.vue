<template>
  <div class="docker-page">
    <div class="page-header">
      <div class="header-left">
        <h1>Docker 管理</h1>
        <p class="subtitle">容器、镜像、网络和卷管理</p>
      </div>
      <div class="header-actions">
        <el-select v-if="hasMultipleServers" v-model="selectedServer" placeholder="选择服务器" size="small">
          <el-option v-for="s in connectedServers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="refresh" :loading="loading" size="small">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <div v-else-if="!dockerInstalled" class="empty-state">
      <el-empty description="Docker 未安装">
        <el-button type="primary" size="small" @click="goToEnvironment">前往安装</el-button>
      </el-empty>
    </div>

    <template v-else>
      <!-- 标签页 -->
      <el-tabs v-model="activeTab" class="docker-tabs">
        <el-tab-pane name="containers">
          <template #label>
            <span class="tab-label">容器 <el-badge :value="containers.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="images">
          <template #label>
            <span class="tab-label">镜像 <el-badge :value="images.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="networks">
          <template #label>
            <span class="tab-label">网络 <el-badge :value="networks.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="volumes">
          <template #label>
            <span class="tab-label">卷 <el-badge :value="volumes.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="compose">
          <template #label>
            <span class="tab-label">Compose <el-badge :value="composeProjects.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="hub">
          <template #label>
            <span class="tab-label"><el-icon><Shop /></el-icon> 应用商店</span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <!-- 容器标签页 -->
      <div v-show="activeTab === 'containers'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="containerSearch" placeholder="搜索容器..." size="small" clearable style="width: 200px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-radio-group v-model="containerFilter" size="small">
            <el-radio-button value="all">全部</el-radio-button>
            <el-radio-button value="running">运行中</el-radio-button>
            <el-radio-button value="stopped">已停止</el-radio-button>
          </el-radio-group>
          <el-button type="primary" size="small" @click="showCreateContainer = true">创建容器</el-button>
        </div>

        <el-table :data="filteredContainers" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="名称" min-width="150">
            <template #default="{ row }">
              <div class="cell-name">
                <span class="status-dot" :class="row.state"></span>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="image" label="镜像" min-width="180">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.image }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="140" />
          <el-table-column label="端口" width="120">
            <template #default="{ row }">
              <span v-if="row.ports">{{ formatPorts(row.ports) }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button v-if="row.state !== 'running'" type="success" @click="containerAction(row.id, 'start')">启动</el-button>
                <el-button v-if="row.state === 'running'" type="warning" @click="containerAction(row.id, 'stop')">停止</el-button>
                <el-button @click="containerAction(row.id, 'restart')">重启</el-button>
                <el-button @click="showLogs(row)">日志</el-button>
                <el-button type="danger" @click="deleteContainer(row)">删除</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 镜像标签页 -->
      <div v-show="activeTab === 'images'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="imageSearch" placeholder="搜索镜像..." size="small" clearable style="width: 200px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button size="small" @click="showPullImage = true">拉取镜像</el-button>
        </div>

        <el-table :data="filteredImages" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="repository" label="仓库" min-width="200" />
          <el-table-column prop="tag" label="标签" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.tag }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="id" label="ID" width="120">
            <template #default="{ row }">
              <code class="mono">{{ row.id?.substring(0, 12) }}</code>
            </template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="100">
            <template #default="{ row }">{{ formatSize(row.size) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button type="primary" @click="createFromImage(row)">创建容器</el-button>
                <el-button type="danger" @click="deleteImage(row)">删除</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 网络标签页 -->
      <div v-show="activeTab === 'networks'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="networkSearch" placeholder="搜索网络..." size="small" clearable style="width: 200px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" size="small" @click="showCreateNetwork = true">创建网络</el-button>
        </div>

        <el-table :data="filteredNetworks" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="driver" label="驱动" width="100">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.driver }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="scope" label="范围" width="100" />
          <el-table-column prop="subnet" label="子网" width="150" />
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="deleteNetwork(row)" 
                :disabled="['bridge', 'host', 'none'].includes(row.name)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 卷标签页 -->
      <div v-show="activeTab === 'volumes'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="volumeSearch" placeholder="搜索卷..." size="small" clearable style="width: 200px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" size="small" @click="showCreateVolume = true">创建卷</el-button>
        </div>

        <el-table :data="filteredVolumes" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="driver" label="驱动" width="100">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.driver }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="mountpoint" label="挂载点" min-width="250">
            <template #default="{ row }">
              <code class="mono">{{ row.mountpoint }}</code>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="deleteVolume(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Compose 标签页 -->
      <div v-show="activeTab === 'compose'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="composeSearch" placeholder="搜索项目..." size="small" clearable style="width: 200px">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>

        <el-table :data="filteredComposeProjects" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="项目名称" min-width="150" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="getComposeStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="config_files" label="配置文件" min-width="200" show-overflow-tooltip />
          <el-table-column label="服务数" width="80">
            <template #default="{ row }">{{ row.services?.length || 0 }}</template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button type="success" @click="composeAction(row, 'up')">启动</el-button>
                <el-button @click="composeAction(row, 'restart')">重启</el-button>
                <el-button type="danger" @click="composeAction(row, 'down')">停止</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 应用商店标签页 -->
      <div v-show="activeTab === 'hub'" class="tab-content">
        <div class="toolbar">
          <el-input v-model="hubSearch" placeholder="搜索 Docker Hub 镜像..." size="small" clearable style="width: 300px"
            @keyup.enter="searchDockerHub">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" size="small" @click="searchDockerHub" :loading="hubSearching">搜索</el-button>
        </div>

        <!-- 热门应用 -->
        <div v-if="!hubSearchResults.length && !hubSearching" class="popular-apps">
          <h3>热门应用 - 一键部署</h3>
          <div class="app-grid">
            <div v-for="app in popularApps" :key="app.name" class="app-card" @click="showDeployDialog(app)">
              <div class="app-icon">{{ app.icon }}</div>
              <div class="app-info">
                <div class="app-name">{{ app.name }}</div>
                <div class="app-desc">{{ app.description }}</div>
              </div>
              <el-button type="primary" size="small" class="deploy-btn">部署</el-button>
            </div>
          </div>
        </div>

        <!-- 搜索结果 -->
        <div v-if="hubSearchResults.length || hubSearching" class="search-results">
          <el-table :data="hubSearchResults" v-loading="hubSearching" size="small" class="data-table">
            <el-table-column prop="name" label="镜像名称" min-width="200">
              <template #default="{ row }">
                <div class="hub-name">
                  <el-icon v-if="row.is_official" color="#3b82f6"><CircleCheck /></el-icon>
                  <span>{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
            <el-table-column prop="star_count" label="Stars" width="100">
              <template #default="{ row }">
                <span>⭐ {{ formatStars(row.star_count) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button-group size="small">
                  <el-button type="primary" @click="quickDeploy(row)">部署</el-button>
                  <el-button @click="pullHubImage(row.name)">拉取</el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </template>

    <!-- 日志对话框 -->
    <el-dialog v-model="showLogDialog" :title="`容器日志 - ${currentContainer?.name}`" width="80%" top="5vh" class="dark-dialog">
      <div class="log-container">
        <pre ref="logPre">{{ logContent }}</pre>
      </div>
      <template #footer>
        <el-button size="small" @click="refreshLogs">刷新</el-button>
        <el-button size="small" @click="showLogDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 创建容器对话框 -->
    <el-dialog v-model="showCreateContainer" title="创建容器" width="500px" class="dark-dialog">
      <el-form :model="newContainer" label-width="80px" size="small">
        <el-form-item label="名称" required>
          <el-input v-model="newContainer.name" placeholder="容器名称" />
        </el-form-item>
        <el-form-item label="镜像" required>
          <el-select v-model="newContainer.image" filterable allow-create placeholder="选择或输入镜像" style="width: 100%">
            <el-option v-for="img in images" :key="img.id" :label="`${img.repository}:${img.tag}`" :value="`${img.repository}:${img.tag}`" />
          </el-select>
        </el-form-item>
        <el-form-item label="端口映射">
          <el-input v-model="newContainer.ports" placeholder="8080:80, 3000:3000" />
        </el-form-item>
        <el-form-item label="环境变量">
          <el-input v-model="newContainer.env" placeholder="KEY=value, KEY2=value2" />
        </el-form-item>
        <el-form-item label="重启策略">
          <el-select v-model="newContainer.restart" style="width: 100%">
            <el-option value="no" label="不重启" />
            <el-option value="always" label="总是重启" />
            <el-option value="on-failure" label="失败时重启" />
            <el-option value="unless-stopped" label="除非手动停止" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showCreateContainer = false">取消</el-button>
        <el-button type="primary" size="small" @click="createContainer" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 拉取镜像对话框 -->
    <el-dialog v-model="showPullImage" title="拉取镜像" width="400px" class="dark-dialog">
      <el-form label-width="80px" size="small">
        <el-form-item label="镜像名称" required>
          <el-input v-model="pullImageName" placeholder="nginx:latest" />
        </el-form-item>
      </el-form>
      <div v-if="pullOutput" class="pull-output">
        <pre>{{ pullOutput }}</pre>
      </div>
      <template #footer>
        <el-button size="small" @click="showPullImage = false">取消</el-button>
        <el-button type="primary" size="small" @click="pullImage" :loading="pulling">拉取</el-button>
      </template>
    </el-dialog>

    <!-- 创建网络对话框 -->
    <el-dialog v-model="showCreateNetwork" title="创建网络" width="400px" class="dark-dialog">
      <el-form :model="newNetwork" label-width="80px" size="small">
        <el-form-item label="名称" required>
          <el-input v-model="newNetwork.name" placeholder="网络名称" />
        </el-form-item>
        <el-form-item label="驱动">
          <el-select v-model="newNetwork.driver" style="width: 100%">
            <el-option value="bridge" label="bridge" />
            <el-option value="overlay" label="overlay" />
          </el-select>
        </el-form-item>
        <el-form-item label="子网">
          <el-input v-model="newNetwork.subnet" placeholder="172.20.0.0/16" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showCreateNetwork = false">取消</el-button>
        <el-button type="primary" size="small" @click="createNetwork">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建卷对话框 -->
    <el-dialog v-model="showCreateVolume" title="创建卷" width="400px" class="dark-dialog">
      <el-form :model="newVolume" label-width="80px" size="small">
        <el-form-item label="名称" required>
          <el-input v-model="newVolume.name" placeholder="卷名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showCreateVolume = false">取消</el-button>
        <el-button type="primary" size="small" @click="createVolume">创建</el-button>
      </template>
    </el-dialog>

    <!-- 一键部署对话框 -->
    <el-dialog v-model="showDeploy" :title="`部署 ${deployApp?.name || ''}`" width="500px" class="dark-dialog">
      <el-form :model="deployConfig" label-width="100px" size="small">
        <el-form-item label="容器名称" required>
          <el-input v-model="deployConfig.name" :placeholder="deployApp?.defaultName || 'my-app'" />
        </el-form-item>
        <el-form-item label="镜像版本">
          <el-select v-model="deployConfig.tag" style="width: 100%">
            <el-option value="latest" label="latest (最新)" />
            <el-option v-for="tag in deployApp?.tags || []" :key="tag" :value="tag" :label="tag" />
          </el-select>
        </el-form-item>
        <el-form-item v-for="port in deployApp?.ports || []" :key="port.container" :label="`端口 ${port.container}`">
          <el-input v-model="deployConfig.ports[port.container]" :placeholder="String(port.host)">
            <template #prepend>主机端口</template>
          </el-input>
        </el-form-item>
        <el-form-item v-for="env in deployApp?.envs || []" :key="env.name" :label="env.label">
          <el-input v-model="deployConfig.envs[env.name]" :placeholder="env.default" :type="env.secret ? 'password' : 'text'" />
        </el-form-item>
        <el-form-item v-for="vol in deployApp?.volumes || []" :key="vol.container" :label="vol.label">
          <el-input v-model="deployConfig.volumes[vol.container]" :placeholder="vol.host" />
        </el-form-item>
        <el-form-item label="自动重启">
          <el-switch v-model="deployConfig.restart" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="showDeploy = false">取消</el-button>
        <el-button type="primary" size="small" @click="executeDeploy" :loading="deploying">部署</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Shop, CircleCheck } from '@element-plus/icons-vue'

interface Container {
  id: string
  name: string
  image: string
  state: string
  status: string
  ports?: string
}

interface Image {
  id: string
  repository: string
  tag: string
  size: number
  created: string
}

interface Network {
  id: string
  name: string
  driver: string
  scope: string
  subnet?: string
}

interface Volume {
  name: string
  driver: string
  mountpoint: string
}

interface ComposeProject {
  name: string
  status: string
  config_files: string
  services: any[]
}

const router = useRouter()
const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const activeTab = ref('containers')
const loading = ref(false)
const dockerInstalled = ref(false)

// 数据
const containers = ref<Container[]>([])
const images = ref<Image[]>([])
const networks = ref<Network[]>([])
const volumes = ref<Volume[]>([])
const composeProjects = ref<ComposeProject[]>([])

// 搜索
const containerSearch = ref('')
const containerFilter = ref('all')
const imageSearch = ref('')
const networkSearch = ref('')
const volumeSearch = ref('')
const composeSearch = ref('')

// 对话框
const showLogDialog = ref(false)
const showCreateContainer = ref(false)
const showPullImage = ref(false)
const showCreateNetwork = ref(false)
const showCreateVolume = ref(false)
const currentContainer = ref<Container | null>(null)
const logContent = ref('')
const logPre = ref<HTMLPreElement | null>(null)

// 表单
const newContainer = ref({ name: '', image: '', ports: '', env: '', restart: 'no' })
const newNetwork = ref({ name: '', driver: 'bridge', subnet: '' })
const newVolume = ref({ name: '' })
const pullImageName = ref('')
const pullOutput = ref('')
const pulling = ref(false)
const creating = ref(false)

// Docker Hub 搜索
const hubSearch = ref('')
const hubSearching = ref(false)
const hubSearchResults = ref<any[]>([])

// 一键部署
const showDeploy = ref(false)
const deploying = ref(false)
const deployApp = ref<any>(null)
const deployConfig = ref<any>({ name: '', tag: 'latest', ports: {}, envs: {}, volumes: {}, restart: true })

// 热门应用配置
const popularApps = [
  {
    name: 'Nginx',
    icon: '🌐',
    description: '高性能 Web 服务器',
    image: 'nginx',
    defaultName: 'nginx',
    tags: ['latest', 'alpine', '1.25', '1.24'],
    ports: [{ container: 80, host: 80 }],
    envs: [],
    volumes: [{ container: '/usr/share/nginx/html', host: '/var/www/html', label: '网站目录' }]
  },
  {
    name: 'MySQL',
    icon: '🐬',
    description: '流行的关系型数据库',
    image: 'mysql',
    defaultName: 'mysql',
    tags: ['latest', '8.0', '5.7'],
    ports: [{ container: 3306, host: 3306 }],
    envs: [{ name: 'MYSQL_ROOT_PASSWORD', label: 'Root密码', default: '', secret: true }],
    volumes: [{ container: '/var/lib/mysql', host: '/data/mysql', label: '数据目录' }]
  },
  {
    name: 'Redis',
    icon: '🔴',
    description: '高性能键值存储',
    image: 'redis',
    defaultName: 'redis',
    tags: ['latest', 'alpine', '7', '6'],
    ports: [{ container: 6379, host: 6379 }],
    envs: [],
    volumes: [{ container: '/data', host: '/data/redis', label: '数据目录' }]
  },
  {
    name: 'PostgreSQL',
    icon: '🐘',
    description: '强大的开源数据库',
    image: 'postgres',
    defaultName: 'postgres',
    tags: ['latest', '16', '15', '14'],
    ports: [{ container: 5432, host: 5432 }],
    envs: [{ name: 'POSTGRES_PASSWORD', label: '密码', default: '', secret: true }],
    volumes: [{ container: '/var/lib/postgresql/data', host: '/data/postgres', label: '数据目录' }]
  },
  {
    name: 'MongoDB',
    icon: '🍃',
    description: 'NoSQL 文档数据库',
    image: 'mongo',
    defaultName: 'mongo',
    tags: ['latest', '7', '6', '5'],
    ports: [{ container: 27017, host: 27017 }],
    envs: [],
    volumes: [{ container: '/data/db', host: '/data/mongo', label: '数据目录' }]
  },
  {
    name: 'WordPress',
    icon: '📝',
    description: '流行的博客/CMS系统',
    image: 'wordpress',
    defaultName: 'wordpress',
    tags: ['latest', 'php8.2', 'php8.1'],
    ports: [{ container: 80, host: 8080 }],
    envs: [
      { name: 'WORDPRESS_DB_HOST', label: '数据库地址', default: 'mysql:3306' },
      { name: 'WORDPRESS_DB_USER', label: '数据库用户', default: 'root' },
      { name: 'WORDPRESS_DB_PASSWORD', label: '数据库密码', default: '', secret: true }
    ],
    volumes: []
  },
  {
    name: 'Portainer',
    icon: '🐳',
    description: 'Docker 可视化管理',
    image: 'portainer/portainer-ce',
    defaultName: 'portainer',
    tags: ['latest', '2.19.4'],
    ports: [{ container: 9000, host: 9000 }],
    envs: [],
    volumes: [{ container: '/var/run/docker.sock', host: '/var/run/docker.sock', label: 'Docker Socket' }]
  },
  {
    name: 'Adminer',
    icon: '📊',
    description: '轻量级数据库管理',
    image: 'adminer',
    defaultName: 'adminer',
    tags: ['latest'],
    ports: [{ container: 8080, host: 8081 }],
    envs: [],
    volumes: []
  }
]

const connectedServers = computed(() => serverStore.connectedServers)
const hasMultipleServers = computed(() => serverStore.hasMultipleServers)

const filteredContainers = computed(() => {
  let list = containers.value
  if (containerFilter.value === 'running') list = list.filter(c => c.state === 'running')
  else if (containerFilter.value === 'stopped') list = list.filter(c => c.state !== 'running')
  if (containerSearch.value) list = list.filter(c => c.name.includes(containerSearch.value) || c.image.includes(containerSearch.value))
  return list
})

const filteredImages = computed(() => {
  if (!imageSearch.value) return images.value
  return images.value.filter(i => i.repository.includes(imageSearch.value) || i.tag.includes(imageSearch.value))
})

const filteredNetworks = computed(() => {
  if (!networkSearch.value) return networks.value
  return networks.value.filter(n => n.name.includes(networkSearch.value))
})

const filteredVolumes = computed(() => {
  if (!volumeSearch.value) return volumes.value
  return volumes.value.filter(v => v.name.includes(volumeSearch.value))
})

const filteredComposeProjects = computed(() => {
  if (!composeSearch.value) return composeProjects.value
  return composeProjects.value.filter(p => p.name.includes(composeSearch.value))
})

watch(selectedServer, (val) => {
  if (val) checkDockerAndLoad()
})

onMounted(() => {
  if (connectedServers.value.length > 0) {
    selectedServer.value = serverStore.currentServerId || connectedServers.value[0].id
  }
})

async function checkDockerAndLoad() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'docker --version'])
    dockerInstalled.value = result.exit_code === 0
    if (dockerInstalled.value) {
      await loadAllData()
    }
  } catch {
    dockerInstalled.value = false
  } finally {
    loading.value = false
  }
}

async function loadAllData() {
  await Promise.all([loadContainers(), loadImages(), loadNetworks(), loadVolumes(), loadComposeProjects()])
}

async function loadContainers() {
  if (!selectedServer.value) return
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'docker ps -a --format "{{.ID}}|{{.Names}}|{{.Image}}|{{.State}}|{{.Status}}|{{.Ports}}"']
    )
    const stdout = result.stdout || ''
    containers.value = stdout.trim().split('\n').filter(l => l).map(line => {
      const [id, name, image, state, status, ports] = line.split('|')
      return { id, name, image, state, status, ports }
    })
  } catch { containers.value = [] }
}

async function loadImages() {
  if (!selectedServer.value) return
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'docker images --format "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}"']
    )
    const stdout = result.stdout || ''
    images.value = stdout.trim().split('\n').filter(l => l).map(line => {
      const [id, repository, tag, size, created] = line.split('|')
      return { id, repository, tag, size: parseSize(size), created }
    })
  } catch { images.value = [] }
}

async function loadNetworks() {
  if (!selectedServer.value) return
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'docker network ls --format "{{.ID}}|{{.Name}}|{{.Driver}}|{{.Scope}}"']
    )
    const stdout = result.stdout || ''
    networks.value = stdout.trim().split('\n').filter(l => l).map(line => {
      const [id, name, driver, scope] = line.split('|')
      return { id, name, driver, scope }
    })
  } catch { networks.value = [] }
}

async function loadVolumes() {
  if (!selectedServer.value) return
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'docker volume ls --format "{{.Name}}|{{.Driver}}|{{.Mountpoint}}"']
    )
    const stdout = result.stdout || ''
    volumes.value = stdout.trim().split('\n').filter(l => l).map(line => {
      const [name, driver, mountpoint] = line.split('|')
      return { name, driver, mountpoint }
    })
  } catch { volumes.value = [] }
}

async function loadComposeProjects() {
  if (!selectedServer.value) return
  try {
    const result = await window.electronAPI.compose.list(selectedServer.value)
    composeProjects.value = result.projects || []
  } catch { composeProjects.value = [] }
}

function refresh() {
  checkDockerAndLoad()
}

function goToEnvironment() {
  router.push('/environment')
}

async function containerAction(id: string, action: string) {
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker ${action} ${id}`])
    ElMessage.success(`容器${action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启'}成功`)
    loadContainers()
  } catch (e) {
    ElMessage.error('操作失败: ' + (e as Error).message)
  }
}

async function deleteContainer(container: Container) {
  try {
    await ElMessageBox.confirm(`确定删除容器 ${container.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker rm -f ${container.id}`])
    ElMessage.success('容器已删除')
    loadContainers()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

async function showLogs(container: Container) {
  currentContainer.value = container
  logContent.value = '加载中...'
  showLogDialog.value = true
  await refreshLogs()
}

async function refreshLogs() {
  if (!selectedServer.value || !currentContainer.value) return
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', `docker logs --tail 200 ${currentContainer.value.id}`]
    )
    logContent.value = result.stdout || result.stderr || '无日志'
  } catch (e) {
    logContent.value = '获取日志失败: ' + (e as Error).message
  }
}

async function createContainer() {
  if (!selectedServer.value || !newContainer.value.name || !newContainer.value.image) {
    ElMessage.warning('请填写容器名称和镜像')
    return
  }
  creating.value = true
  try {
    let cmd = `docker run -d --name ${newContainer.value.name}`
    if (newContainer.value.ports) {
      newContainer.value.ports.split(',').forEach(p => { cmd += ` -p ${p.trim()}` })
    }
    if (newContainer.value.env) {
      newContainer.value.env.split(',').forEach(e => { cmd += ` -e ${e.trim()}` })
    }
    if (newContainer.value.restart !== 'no') {
      cmd += ` --restart ${newContainer.value.restart}`
    }
    cmd += ` ${newContainer.value.image}`
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
    ElMessage.success('容器创建成功')
    showCreateContainer.value = false
    newContainer.value = { name: '', image: '', ports: '', env: '', restart: 'no' }
    loadContainers()
  } catch (e) {
    ElMessage.error('创建失败: ' + (e as Error).message)
  } finally {
    creating.value = false
  }
}

function createFromImage(image: Image) {
  newContainer.value.image = `${image.repository}:${image.tag}`
  showCreateContainer.value = true
}

async function pullImage() {
  if (!selectedServer.value || !pullImageName.value) {
    ElMessage.warning('请输入镜像名称')
    return
  }
  pulling.value = true
  pullOutput.value = '正在拉取...\n'
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', `docker pull ${pullImageName.value}`]
    )
    pullOutput.value += result.stdout || ''
    if (result.exit_code === 0) {
      ElMessage.success('镜像拉取成功')
      loadImages()
    } else {
      pullOutput.value += '\n拉取失败'
    }
  } catch (e) {
    pullOutput.value += '\n错误: ' + (e as Error).message
  } finally {
    pulling.value = false
  }
}

async function deleteImage(image: Image) {
  try {
    await ElMessageBox.confirm(`确定删除镜像 ${image.repository}:${image.tag}？`, '确认删除', { type: 'warning' })
  } catch { return }
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker rmi ${image.id}`])
    ElMessage.success('镜像已删除')
    loadImages()
  } catch (e) {
    ElMessage.error('删除失败，可能有容器正在使用此镜像')
  }
}

async function createNetwork() {
  if (!selectedServer.value || !newNetwork.value.name) {
    ElMessage.warning('请输入网络名称')
    return
  }
  try {
    let cmd = `docker network create -d ${newNetwork.value.driver}`
    if (newNetwork.value.subnet) cmd += ` --subnet ${newNetwork.value.subnet}`
    cmd += ` ${newNetwork.value.name}`
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
    ElMessage.success('网络创建成功')
    showCreateNetwork.value = false
    newNetwork.value = { name: '', driver: 'bridge', subnet: '' }
    loadNetworks()
  } catch (e) {
    ElMessage.error('创建失败')
  }
}

async function deleteNetwork(network: Network) {
  try {
    await ElMessageBox.confirm(`确定删除网络 ${network.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker network rm ${network.id}`])
    ElMessage.success('网络已删除')
    loadNetworks()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

async function createVolume() {
  if (!selectedServer.value || !newVolume.value.name) {
    ElMessage.warning('请输入卷名称')
    return
  }
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker volume create ${newVolume.value.name}`])
    ElMessage.success('卷创建成功')
    showCreateVolume.value = false
    newVolume.value = { name: '' }
    loadVolumes()
  } catch (e) {
    ElMessage.error('创建失败')
  }
}

async function deleteVolume(volume: Volume) {
  try {
    await ElMessageBox.confirm(`确定删除卷 ${volume.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker volume rm ${volume.name}`])
    ElMessage.success('卷已删除')
    loadVolumes()
  } catch (e) {
    ElMessage.error('删除失败，可能有容器正在使用此卷')
  }
}

async function composeAction(project: ComposeProject, action: 'up' | 'down' | 'restart') {
  if (!selectedServer.value) return
  const actionNames: Record<string, string> = { up: '启动', down: '停止', restart: '重启' }
  try {
    const options = { project_path: project.config_files }
    switch (action) {
      case 'up':
        await window.electronAPI.compose.up(selectedServer.value, { ...options, detach: true })
        break
      case 'down':
        await window.electronAPI.compose.down(selectedServer.value, options)
        break
      case 'restart':
        await window.electronAPI.compose.restart(selectedServer.value, options)
        break
    }
    ElMessage.success(`${actionNames[action]}成功`)
    loadComposeProjects()
  } catch (e) {
    ElMessage.error(`${actionNames[action]}失败: ${(e as Error).message}`)
  }
}

function formatPorts(ports: string): string {
  if (!ports) return '-'
  const parts = ports.split(',').slice(0, 2)
  return parts.map(p => p.split('->')[0]).join(', ') + (ports.split(',').length > 2 ? '...' : '')
}

function formatSize(size: number | string): string {
  if (typeof size === 'string') return size
  if (size < 1024) return size + 'B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + 'KB'
  if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(1) + 'MB'
  return (size / 1024 / 1024 / 1024).toFixed(2) + 'GB'
}

function parseSize(sizeStr: string): number {
  const match = sizeStr.match(/^([\d.]+)\s*(B|KB|MB|GB)$/i)
  if (!match) return 0
  const num = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const multipliers: Record<string, number> = { B: 1, KB: 1024, MB: 1024 * 1024, GB: 1024 * 1024 * 1024 }
  return num * (multipliers[unit] || 1)
}

function getComposeStatusType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  if (status?.includes('running')) return 'success'
  if (status?.includes('exited') || status?.includes('stopped')) return 'danger'
  return 'info'
}

// Docker Hub 搜索
async function searchDockerHub() {
  if (!hubSearch.value.trim()) return
  hubSearching.value = true
  hubSearchResults.value = []
  try {
    const response = await fetch(`https://hub.docker.com/v2/search/repositories/?query=${encodeURIComponent(hubSearch.value)}&page_size=20`)
    const data = await response.json()
    hubSearchResults.value = data.results || []
  } catch (e) {
    ElMessage.error('搜索失败，请检查网络连接')
  } finally {
    hubSearching.value = false
  }
}

function formatStars(count: number): string {
  if (count >= 1000000) return (count / 1000000).toFixed(1) + 'M'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'K'
  return String(count)
}

// 一键部署
function showDeployDialog(app: any) {
  deployApp.value = app
  deployConfig.value = {
    name: app.defaultName,
    tag: 'latest',
    ports: {},
    envs: {},
    volumes: {},
    restart: true
  }
  // 初始化默认值
  app.ports?.forEach((p: any) => { deployConfig.value.ports[p.container] = String(p.host) })
  app.envs?.forEach((e: any) => { deployConfig.value.envs[e.name] = e.default || '' })
  app.volumes?.forEach((v: any) => { deployConfig.value.volumes[v.container] = v.host })
  showDeploy.value = true
}

function quickDeploy(hubImage: any) {
  // 从 Hub 搜索结果快速部署
  const app = {
    name: hubImage.name,
    image: hubImage.name,
    defaultName: hubImage.name.split('/').pop()?.replace(/[^a-z0-9]/gi, '-') || 'app',
    tags: ['latest'],
    ports: [],
    envs: [],
    volumes: []
  }
  showDeployDialog(app)
}

async function executeDeploy() {
  if (!selectedServer.value || !deployApp.value) return
  if (!deployConfig.value.name) {
    ElMessage.warning('请输入容器名称')
    return
  }
  
  deploying.value = true
  try {
    const app = deployApp.value
    const cfg = deployConfig.value
    
    // 构建 docker run 命令
    let cmd = `docker run -d --name ${cfg.name}`
    
    // 端口映射
    Object.entries(cfg.ports).forEach(([container, host]) => {
      if (host) cmd += ` -p ${host}:${container}`
    })
    
    // 环境变量
    Object.entries(cfg.envs).forEach(([name, value]) => {
      if (value) cmd += ` -e ${name}="${value}"`
    })
    
    // 卷挂载
    Object.entries(cfg.volumes).forEach(([container, host]) => {
      if (host) cmd += ` -v ${host}:${container}`
    })
    
    // 重启策略
    if (cfg.restart) cmd += ' --restart unless-stopped'
    
    // 镜像
    cmd += ` ${app.image}:${cfg.tag}`
    
    // 先拉取镜像
    ElMessage.info('正在拉取镜像...')
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `docker pull ${app.image}:${cfg.tag}`])
    
    // 创建容器
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
    
    if (result.exit_code === 0) {
      ElMessage.success(`${app.name} 部署成功！`)
      showDeploy.value = false
      activeTab.value = 'containers'
      loadContainers()
    } else {
      ElMessage.error('部署失败: ' + (result.stderr || result.stdout))
    }
  } catch (e) {
    ElMessage.error('部署失败: ' + (e as Error).message)
  } finally {
    deploying.value = false
  }
}

async function pullHubImage(imageName: string) {
  pullImageName.value = imageName + ':latest'
  showPullImage.value = true
  await pullImage()
}
</script>

<style lang="scss" scoped>
.docker-page {
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

.empty-state {
  padding: 60px 0;
}

.docker-tabs {
  margin-bottom: 16px;

  .tab-label {
    display: flex;
    align-items: center;
    gap: 6px;
  }
}

.tab-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}

.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
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
      &.exited, &.stopped { background: #ef4444; }
      &.paused { background: #f59e0b; }
    }
  }

  .mono {
    font-family: 'Consolas', monospace;
    font-size: 12px;
    background: var(--bg-tertiary);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .text-muted {
    color: var(--text-secondary);
  }
}

.log-container {
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

.pull-output {
  background: var(--bg-tertiary);
  border-radius: 6px;
  padding: 12px;
  margin-top: 12px;
  max-height: 200px;
  overflow: auto;

  pre {
    margin: 0;
    font-size: 12px;
    color: var(--text-color);
    white-space: pre-wrap;
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

// 应用商店样式
.popular-apps {
  h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-color);
  }
}

.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.app-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: var(--primary-color);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .app-icon {
    font-size: 32px;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-secondary);
    border-radius: 8px;
  }

  .app-info {
    flex: 1;
    min-width: 0;

    .app-name {
      font-weight: 600;
      font-size: 14px;
      margin-bottom: 4px;
    }

    .app-desc {
      font-size: 12px;
      color: var(--text-secondary);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .deploy-btn {
    flex-shrink: 0;
  }
}

.hub-name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.search-results {
  margin-top: 16px;
}
</style>
