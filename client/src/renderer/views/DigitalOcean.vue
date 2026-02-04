<template>
  <div class="digitalocean-page">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="$router.push('/cloud')"><el-icon><ArrowLeft /></el-icon></el-button>
        <span class="provider-icon">🌊</span>
        <div>
          <h1>DigitalOcean</h1>
          <p class="subtitle">Droplets、Spaces、Kubernetes 管理</p>
        </div>
      </div>
      <div class="header-right">
        <el-select v-model="selectedRegion" placeholder="选择区域" @change="loadRegionData">
          <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select>
        <el-button @click="refreshData" :loading="loading"><el-icon><Refresh /></el-icon>刷新</el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <!-- Droplets -->
      <el-tab-pane label="Droplets" name="droplets">
        <div class="tab-header">
          <el-input v-model="dropletSearch" placeholder="搜索 Droplet..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateDropletDialog"><el-icon><Plus /></el-icon>创建 Droplet</el-button>
        </div>
        <el-table :data="filteredDroplets" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="id" label="ID" width="120" />
          <el-table-column prop="size" label="规格" width="140" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="publicIp" label="公网 IP" width="140" />
          <el-table-column prop="privateIp" label="私网 IP" width="140" />
          <el-table-column prop="region" label="区域" width="100" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'off'" text size="small" type="success" @click="powerOnDroplet(row)">开机</el-button>
              <el-button v-if="row.status === 'active'" text size="small" type="warning" @click="powerOffDroplet(row)">关机</el-button>
              <el-button text size="small" @click="showDropletDetail(row)">详情</el-button>
              <el-button text size="small" @click="openConsole(row)">控制台</el-button>
              <el-button text size="small" type="danger" @click="destroyDroplet(row)">销毁</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Spaces -->
      <el-tab-pane label="Spaces" name="spaces">
        <div class="tab-header">
          <el-input v-model="spaceSearch" placeholder="搜索 Space..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateSpaceDialog"><el-icon><Plus /></el-icon>创建 Space</el-button>
        </div>
        <el-table :data="filteredSpaces" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="region" label="区域" width="120" />
          <el-table-column prop="endpoint" label="Endpoint" min-width="250" show-overflow-tooltip />
          <el-table-column prop="filesCount" label="文件数" width="100" />
          <el-table-column prop="size" label="大小" width="100" />
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="browseSpace(row)">浏览</el-button>
              <el-button text size="small" type="danger" @click="deleteSpace(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Kubernetes -->
      <el-tab-pane label="Kubernetes" name="kubernetes">
        <div class="tab-header">
          <el-input v-model="k8sSearch" placeholder="搜索集群..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateClusterDialog"><el-icon><Plus /></el-icon>创建集群</el-button>
        </div>
        <el-table :data="filteredClusters" v-loading="loading" stripe>
          <el-table-column prop="name" label="集群名称" min-width="180" />
          <el-table-column prop="version" label="版本" width="100" />
          <el-table-column prop="region" label="区域" width="100" />
          <el-table-column prop="nodeCount" label="节点数" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'running' ? 'success' : row.status === 'provisioning' ? 'warning' : 'info'" size="small">
                {{ row.status === 'running' ? '运行中' : row.status === 'provisioning' ? '创建中' : '已停止' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="endpoint" label="API Endpoint" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="downloadKubeconfig(row)">Kubeconfig</el-button>
              <el-button text size="small" @click="showClusterDetail(row)">详情</el-button>
              <el-button text size="small" type="danger" @click="deleteCluster(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Databases -->
      <el-tab-pane label="Databases" name="databases">
        <div class="tab-header">
          <el-input v-model="dbSearch" placeholder="搜索数据库..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateDbDialog"><el-icon><Plus /></el-icon>创建数据库</el-button>
        </div>
        <el-table :data="filteredDatabases" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="engine" label="引擎" width="120">
            <template #default="{ row }"><el-tag size="small">{{ row.engine }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="version" label="版本" width="80" />
          <el-table-column prop="size" label="规格" width="120" />
          <el-table-column prop="region" label="区域" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : 'warning'" size="small">{{ row.status === 'online' ? '在线' : '创建中' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="showDbConnection(row)">连接信息</el-button>
              <el-button text size="small" type="danger" @click="deleteDatabase(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Droplet 详情对话框 -->
    <el-dialog v-model="dropletDetailVisible" title="Droplet 详情" width="600px">
      <el-descriptions v-if="currentDroplet" :column="2" border>
        <el-descriptions-item label="ID">{{ currentDroplet.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ currentDroplet.name }}</el-descriptions-item>
        <el-descriptions-item label="规格">{{ currentDroplet.size }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="getStatusType(currentDroplet.status)" size="small">{{ currentDroplet.status }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="公网 IP">{{ currentDroplet.publicIp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="私网 IP">{{ currentDroplet.privateIp }}</el-descriptions-item>
        <el-descriptions-item label="区域">{{ currentDroplet.region }}</el-descriptions-item>
        <el-descriptions-item label="镜像">{{ currentDroplet.image }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 创建 Droplet 对话框 -->
    <el-dialog v-model="createDropletVisible" title="创建 Droplet" width="500px">
      <el-form :model="dropletForm" label-width="100px">
        <el-form-item label="名称"><el-input v-model="dropletForm.name" placeholder="Droplet 名称" /></el-form-item>
        <el-form-item label="区域"><el-select v-model="dropletForm.region" style="width: 100%">
          <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select></el-form-item>
        <el-form-item label="规格"><el-select v-model="dropletForm.size" style="width: 100%">
          <el-option label="s-1vcpu-1gb ($6/mo)" value="s-1vcpu-1gb" />
          <el-option label="s-1vcpu-2gb ($12/mo)" value="s-1vcpu-2gb" />
          <el-option label="s-2vcpu-4gb ($24/mo)" value="s-2vcpu-4gb" />
          <el-option label="s-4vcpu-8gb ($48/mo)" value="s-4vcpu-8gb" />
        </el-select></el-form-item>
        <el-form-item label="镜像"><el-select v-model="dropletForm.image" style="width: 100%">
          <el-option label="Ubuntu 22.04 LTS" value="ubuntu-22-04-x64" />
          <el-option label="Ubuntu 20.04 LTS" value="ubuntu-20-04-x64" />
          <el-option label="Debian 11" value="debian-11-x64" />
          <el-option label="CentOS Stream 9" value="centos-stream-9-x64" />
        </el-select></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDropletVisible = false">取消</el-button>
        <el-button type="primary" @click="createDroplet" :loading="saving">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建 Space 对话框 -->
    <el-dialog v-model="createSpaceVisible" title="创建 Space" width="500px">
      <el-form :model="spaceForm" label-width="100px">
        <el-form-item label="名称"><el-input v-model="spaceForm.name" placeholder="Space 名称（全局唯一）" /></el-form-item>
        <el-form-item label="区域"><el-select v-model="spaceForm.region" style="width: 100%">
          <el-option v-for="r in spaceRegions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createSpaceVisible = false">取消</el-button>
        <el-button type="primary" @click="createSpace" :loading="saving">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Refresh, Search, Plus } from '@element-plus/icons-vue'

interface Droplet { id: string; name: string; size: string; status: string; publicIp: string; privateIp: string; region: string; image: string }
interface Space { name: string; region: string; endpoint: string; filesCount: number; size: string; createdAt: string }
interface K8sCluster { name: string; version: string; region: string; nodeCount: number; status: string; endpoint: string }
interface Database { name: string; engine: string; version: string; size: string; region: string; status: string }

const route = useRoute()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('droplets')
const selectedRegion = ref('nyc1')
const dropletSearch = ref('')
const spaceSearch = ref('')
const k8sSearch = ref('')
const dbSearch = ref('')

const regions = [
  { value: 'nyc1', label: 'New York 1' },
  { value: 'nyc3', label: 'New York 3' },
  { value: 'sfo3', label: 'San Francisco 3' },
  { value: 'ams3', label: 'Amsterdam 3' },
  { value: 'sgp1', label: 'Singapore 1' },
  { value: 'lon1', label: 'London 1' },
  { value: 'fra1', label: 'Frankfurt 1' }
]

const spaceRegions = [
  { value: 'nyc3', label: 'New York 3' },
  { value: 'sfo3', label: 'San Francisco 3' },
  { value: 'ams3', label: 'Amsterdam 3' },
  { value: 'sgp1', label: 'Singapore 1' },
  { value: 'fra1', label: 'Frankfurt 1' }
]

const droplets = ref<Droplet[]>([
  { id: '123456789', name: 'web-server-01', size: 's-2vcpu-4gb', status: 'active', publicIp: '167.99.123.45', privateIp: '10.132.0.2', region: 'nyc1', image: 'Ubuntu 22.04' },
  { id: '234567890', name: 'db-server-01', size: 's-4vcpu-8gb', status: 'active', publicIp: '167.99.123.46', privateIp: '10.132.0.3', region: 'nyc1', image: 'Ubuntu 20.04' },
  { id: '345678901', name: 'dev-server', size: 's-1vcpu-2gb', status: 'off', publicIp: '', privateIp: '10.132.0.4', region: 'nyc3', image: 'Debian 11' }
])

const spaces = ref<Space[]>([
  { name: 'my-app-assets', region: 'nyc3', endpoint: 'https://my-app-assets.nyc3.digitaloceanspaces.com', filesCount: 1250, size: '2.5 GB', createdAt: '2024-01-15T10:30:00Z' },
  { name: 'backup-storage', region: 'nyc3', endpoint: 'https://backup-storage.nyc3.digitaloceanspaces.com', filesCount: 450, size: '15.8 GB', createdAt: '2024-02-20T14:00:00Z' }
])

const clusters = ref<K8sCluster[]>([
  { name: 'production-cluster', version: '1.28', region: 'nyc1', nodeCount: 3, status: 'running', endpoint: 'https://abc123.k8s.ondigitalocean.com' },
  { name: 'staging-cluster', version: '1.28', region: 'nyc3', nodeCount: 2, status: 'running', endpoint: 'https://def456.k8s.ondigitalocean.com' }
])

const databases = ref<Database[]>([
  { name: 'main-postgres', engine: 'PostgreSQL', version: '15', size: 'db-s-2vcpu-4gb', region: 'nyc1', status: 'online' },
  { name: 'cache-redis', engine: 'Redis', version: '7', size: 'db-s-1vcpu-1gb', region: 'nyc1', status: 'online' }
])

const dropletDetailVisible = ref(false)
const createDropletVisible = ref(false)
const createSpaceVisible = ref(false)
const currentDroplet = ref<Droplet | null>(null)
const dropletForm = ref({ name: '', region: 'nyc1', size: 's-1vcpu-1gb', image: 'ubuntu-22-04-x64' })
const spaceForm = ref({ name: '', region: 'nyc3' })

const filteredDroplets = computed(() => {
  if (!dropletSearch.value) return droplets.value
  const q = dropletSearch.value.toLowerCase()
  return droplets.value.filter(d => d.name.toLowerCase().includes(q) || d.id.includes(q))
})

const filteredSpaces = computed(() => {
  if (!spaceSearch.value) return spaces.value
  return spaces.value.filter(s => s.name.toLowerCase().includes(spaceSearch.value.toLowerCase()))
})

const filteredClusters = computed(() => {
  if (!k8sSearch.value) return clusters.value
  return clusters.value.filter(c => c.name.toLowerCase().includes(k8sSearch.value.toLowerCase()))
})

const filteredDatabases = computed(() => {
  if (!dbSearch.value) return databases.value
  return databases.value.filter(d => d.name.toLowerCase().includes(dbSearch.value.toLowerCase()))
})

onMounted(() => {
  const tab = route.query.tab as string
  if (tab && ['droplets', 'spaces', 'kubernetes', 'databases'].includes(tab)) activeTab.value = tab
})

function loadRegionData() { ElMessage.info(`已切换到 ${selectedRegion.value} 区域`) }
function refreshData() { ElMessage.success('数据已刷新') }

function getStatusType(status: string) {
  if (status === 'active') return 'success'
  if (status === 'off') return 'info'
  if (status === 'new') return 'warning'
  return 'danger'
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString('zh-CN') }

function showDropletDetail(droplet: Droplet) { currentDroplet.value = droplet; dropletDetailVisible.value = true }

function showCreateDropletDialog() { dropletForm.value = { name: '', region: selectedRegion.value, size: 's-1vcpu-1gb', image: 'ubuntu-22-04-x64' }; createDropletVisible.value = true }

async function createDroplet() {
  if (!dropletForm.value.name) { ElMessage.warning('请输入 Droplet 名称'); return }
  saving.value = true
  setTimeout(() => {
    droplets.value.push({
      id: Date.now().toString(),
      name: dropletForm.value.name,
      size: dropletForm.value.size,
      status: 'new',
      publicIp: '',
      privateIp: `10.132.0.${droplets.value.length + 5}`,
      region: dropletForm.value.region,
      image: dropletForm.value.image.replace(/-/g, ' ')
    })
    saving.value = false; createDropletVisible.value = false
    ElMessage.success('Droplet 创建中')
    setTimeout(() => {
      const newDroplet = droplets.value.find(d => d.name === dropletForm.value.name)
      if (newDroplet) { newDroplet.status = 'active'; newDroplet.publicIp = `167.99.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}` }
    }, 2000)
  }, 500)
}

async function powerOnDroplet(droplet: Droplet) {
  await ElMessageBox.confirm(`确定开机 ${droplet.name}？`, '确认')
  droplet.status = 'new'
  setTimeout(() => { droplet.status = 'active'; ElMessage.success('Droplet 已开机') }, 1000)
}

async function powerOffDroplet(droplet: Droplet) {
  await ElMessageBox.confirm(`确定关机 ${droplet.name}？`, '确认')
  setTimeout(() => { droplet.status = 'off'; droplet.publicIp = ''; ElMessage.success('Droplet 已关机') }, 1000)
}

function openConsole(droplet: Droplet) { ElMessage.info(`打开 ${droplet.name} 控制台功能即将推出`) }

async function destroyDroplet(droplet: Droplet) {
  await ElMessageBox.confirm(`确定销毁 ${droplet.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  droplets.value = droplets.value.filter(d => d.id !== droplet.id)
  ElMessage.success('Droplet 已销毁')
}

function showCreateSpaceDialog() { spaceForm.value = { name: '', region: 'nyc3' }; createSpaceVisible.value = true }

async function createSpace() {
  if (!spaceForm.value.name) { ElMessage.warning('请输入 Space 名称'); return }
  saving.value = true
  setTimeout(() => {
    spaces.value.push({
      name: spaceForm.value.name,
      region: spaceForm.value.region,
      endpoint: `https://${spaceForm.value.name}.${spaceForm.value.region}.digitaloceanspaces.com`,
      filesCount: 0,
      size: '0 B',
      createdAt: new Date().toISOString()
    })
    saving.value = false; createSpaceVisible.value = false
    ElMessage.success('Space 已创建')
  }, 500)
}

function browseSpace(space: Space) { ElMessage.info(`浏览 ${space.name} 功能即将推出`) }

async function deleteSpace(space: Space) {
  await ElMessageBox.confirm(`确定删除 Space ${space.name}？`, '确认')
  spaces.value = spaces.value.filter(s => s.name !== space.name)
  ElMessage.success('Space 已删除')
}

function showCreateClusterDialog() { ElMessage.info('创建 Kubernetes 集群功能即将推出') }

function downloadKubeconfig(cluster: K8sCluster) { ElMessage.success(`已下载 ${cluster.name} 的 Kubeconfig`) }

function showClusterDetail(cluster: K8sCluster) { ElMessage.info(`${cluster.name} 详情功能即将推出`) }

async function deleteCluster(cluster: K8sCluster) {
  await ElMessageBox.confirm(`确定删除集群 ${cluster.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  clusters.value = clusters.value.filter(c => c.name !== cluster.name)
  ElMessage.success('集群已删除')
}

function showCreateDbDialog() { ElMessage.info('创建数据库功能即将推出') }

function showDbConnection(db: Database) { ElMessage.info(`${db.name} 连接信息功能即将推出`) }

async function deleteDatabase(db: Database) {
  await ElMessageBox.confirm(`确定删除数据库 ${db.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  databases.value = databases.value.filter(d => d.name !== db.name)
  ElMessage.success('数据库已删除')
}
</script>

<style lang="scss" scoped>
.digitalocean-page { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;
  .header-left { display: flex; align-items: center; gap: 12px; .provider-icon { font-size: 32px; } h1 { font-size: 24px; font-weight: 600; margin: 0; } .subtitle { color: var(--text-secondary); font-size: 14px; margin: 0; } }
  .header-right { display: flex; gap: 12px; align-items: center; }
}

.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; .search-input { width: 300px; } }
</style>
