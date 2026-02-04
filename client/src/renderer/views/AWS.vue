<template>
  <div class="aws-page">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="$router.push('/cloud')"><el-icon><ArrowLeft /></el-icon></el-button>
        <span class="provider-icon">🔶</span>
        <div>
          <h1>Amazon Web Services</h1>
          <p class="subtitle">EC2、S3、Route53 管理</p>
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
      <!-- EC2 实例 -->
      <el-tab-pane label="EC2 实例" name="ec2">
        <div class="tab-header">
          <el-input v-model="ec2Search" placeholder="搜索实例..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>
        <el-table :data="filteredInstances" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="instanceId" label="实例 ID" width="180" />
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="state" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStateType(row.state)" size="small">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="publicIp" label="公网 IP" width="140" />
          <el-table-column prop="privateIp" label="私网 IP" width="140" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.state === 'stopped'" text size="small" type="success" @click="startInstance(row)">启动</el-button>
              <el-button v-if="row.state === 'running'" text size="small" type="warning" @click="stopInstance(row)">停止</el-button>
              <el-button text size="small" @click="showInstanceDetail(row)">详情</el-button>
              <el-button text size="small" type="danger" @click="terminateInstance(row)">终止</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- S3 存储桶 -->
      <el-tab-pane label="S3 存储桶" name="s3">
        <div class="tab-header">
          <el-input v-model="s3Search" placeholder="搜索存储桶..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateBucketDialog"><el-icon><Plus /></el-icon>创建存储桶</el-button>
        </div>
        <el-table :data="filteredBuckets" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="region" label="区域" width="150" />
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="access" label="访问权限" width="120">
            <template #default="{ row }">
              <el-tag :type="row.access === 'private' ? 'info' : 'warning'" size="small">{{ row.access === 'private' ? '私有' : '公开' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="browseBucket(row)">浏览</el-button>
              <el-button text size="small" type="danger" @click="deleteBucket(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Route53 -->
      <el-tab-pane label="Route53" name="route53">
        <div class="tab-header">
          <el-select v-model="selectedZone" placeholder="选择托管区域" @change="loadDnsRecords">
            <el-option v-for="z in hostedZones" :key="z.id" :label="z.name" :value="z.id" />
          </el-select>
          <el-button type="primary" @click="showAddRecordDialog" :disabled="!selectedZone"><el-icon><Plus /></el-icon>添加记录</el-button>
        </div>
        <el-table :data="dnsRecords" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="value" label="值" min-width="200" show-overflow-tooltip />
          <el-table-column prop="ttl" label="TTL" width="100" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="editDnsRecord(row)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deleteDnsRecord(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- CloudWatch -->
      <el-tab-pane label="CloudWatch" name="cloudwatch">
        <div class="cloudwatch-section">
          <el-card v-for="metric in cloudwatchMetrics" :key="metric.name" class="metric-card">
            <template #header><span>{{ metric.label }}</span></template>
            <div class="metric-value">{{ metric.value }}{{ metric.unit }}</div>
            <div class="metric-trend" :class="metric.trend">
              <el-icon v-if="metric.trend === 'up'"><Top /></el-icon>
              <el-icon v-else-if="metric.trend === 'down'"><Bottom /></el-icon>
              <span>{{ metric.change }}</span>
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 实例详情对话框 -->
    <el-dialog v-model="instanceDetailVisible" title="实例详情" width="600px">
      <el-descriptions v-if="currentInstance" :column="2" border>
        <el-descriptions-item label="实例 ID">{{ currentInstance.instanceId }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ currentInstance.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ currentInstance.type }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="getStateType(currentInstance.state)" size="small">{{ currentInstance.state }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="公网 IP">{{ currentInstance.publicIp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="私网 IP">{{ currentInstance.privateIp }}</el-descriptions-item>
        <el-descriptions-item label="可用区">{{ currentInstance.az }}</el-descriptions-item>
        <el-descriptions-item label="AMI">{{ currentInstance.ami }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 创建存储桶对话框 -->
    <el-dialog v-model="createBucketVisible" title="创建 S3 存储桶" width="500px">
      <el-form :model="bucketForm" label-width="100px">
        <el-form-item label="存储桶名称"><el-input v-model="bucketForm.name" placeholder="全局唯一名称" /></el-form-item>
        <el-form-item label="区域"><el-select v-model="bucketForm.region" style="width: 100%">
          <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createBucketVisible = false">取消</el-button>
        <el-button type="primary" @click="createBucket" :loading="saving">创建</el-button>
      </template>
    </el-dialog>

    <!-- DNS 记录对话框 -->
    <el-dialog v-model="dnsDialogVisible" :title="editingRecord ? '编辑记录' : '添加记录'" width="500px">
      <el-form :model="dnsForm" label-width="80px">
        <el-form-item label="类型"><el-select v-model="dnsForm.type" :disabled="!!editingRecord">
          <el-option v-for="t in ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS']" :key="t" :label="t" :value="t" />
        </el-select></el-form-item>
        <el-form-item label="名称"><el-input v-model="dnsForm.name" placeholder="记录名称" /></el-form-item>
        <el-form-item label="值"><el-input v-model="dnsForm.value" placeholder="记录值" /></el-form-item>
        <el-form-item label="TTL"><el-input-number v-model="dnsForm.ttl" :min="60" :max="86400" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dnsDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDnsRecord" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Refresh, Search, Plus, Top, Bottom } from '@element-plus/icons-vue'

interface EC2Instance { instanceId: string; name: string; type: string; state: string; publicIp: string; privateIp: string; az: string; ami: string }
interface S3Bucket { name: string; region: string; createdAt: string; access: string }
interface DnsRecord { id: string; type: string; name: string; value: string; ttl: number }
interface CloudWatchMetric { name: string; label: string; value: string; unit: string; trend: string; change: string }

const route = useRoute()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('ec2')
const selectedRegion = ref('us-east-1')
const ec2Search = ref('')
const s3Search = ref('')
const selectedZone = ref('')

const regions = [
  { value: 'us-east-1', label: '美国东部 (弗吉尼亚)' },
  { value: 'us-west-2', label: '美国西部 (俄勒冈)' },
  { value: 'eu-west-1', label: '欧洲 (爱尔兰)' },
  { value: 'ap-northeast-1', label: '亚太 (东京)' },
  { value: 'ap-southeast-1', label: '亚太 (新加坡)' }
]

const instances = ref<EC2Instance[]>([
  { instanceId: 'i-0abc123def456', name: 'web-server-01', type: 't3.medium', state: 'running', publicIp: '54.123.45.67', privateIp: '10.0.1.10', az: 'us-east-1a', ami: 'ami-0123456789' },
  { instanceId: 'i-0def456abc789', name: 'db-server-01', type: 't3.large', state: 'running', publicIp: '', privateIp: '10.0.2.20', az: 'us-east-1b', ami: 'ami-9876543210' },
  { instanceId: 'i-0ghi789jkl012', name: 'dev-server', type: 't3.small', state: 'stopped', publicIp: '', privateIp: '10.0.3.30', az: 'us-east-1a', ami: 'ami-1122334455' }
])

const buckets = ref<S3Bucket[]>([
  { name: 'my-app-assets', region: 'us-east-1', createdAt: '2024-01-15T10:30:00Z', access: 'private' },
  { name: 'backup-data-2024', region: 'us-east-1', createdAt: '2024-02-20T14:00:00Z', access: 'private' },
  { name: 'static-website', region: 'us-west-2', createdAt: '2024-03-10T09:15:00Z', access: 'public' }
])

const hostedZones = ref([{ id: 'Z123456', name: 'example.com' }, { id: 'Z789012', name: 'myapp.io' }])
const dnsRecords = ref<DnsRecord[]>([])

const cloudwatchMetrics = ref<CloudWatchMetric[]>([
  { name: 'cpu', label: 'CPU 使用率', value: '45.2', unit: '%', trend: 'up', change: '+5.3%' },
  { name: 'network_in', label: '网络流入', value: '1.2', unit: ' GB/h', trend: 'down', change: '-12%' },
  { name: 'network_out', label: '网络流出', value: '0.8', unit: ' GB/h', trend: 'up', change: '+8%' },
  { name: 'requests', label: '请求数', value: '12.5', unit: 'K/min', trend: 'up', change: '+15%' }
])

const instanceDetailVisible = ref(false)
const createBucketVisible = ref(false)
const dnsDialogVisible = ref(false)
const currentInstance = ref<EC2Instance | null>(null)
const editingRecord = ref<DnsRecord | null>(null)
const bucketForm = ref({ name: '', region: 'us-east-1' })
const dnsForm = ref({ type: 'A', name: '', value: '', ttl: 300 })

const filteredInstances = computed(() => {
  if (!ec2Search.value) return instances.value
  const q = ec2Search.value.toLowerCase()
  return instances.value.filter(i => i.name.toLowerCase().includes(q) || i.instanceId.toLowerCase().includes(q))
})

const filteredBuckets = computed(() => {
  if (!s3Search.value) return buckets.value
  return buckets.value.filter(b => b.name.toLowerCase().includes(s3Search.value.toLowerCase()))
})

onMounted(() => {
  const tab = route.query.tab as string
  if (tab && ['ec2', 's3', 'route53', 'cloudwatch'].includes(tab)) activeTab.value = tab
})

function loadRegionData() { ElMessage.info(`已切换到 ${selectedRegion.value} 区域`) }
function refreshData() { ElMessage.success('数据已刷新') }

function getStateType(state: string) {
  if (state === 'running') return 'success'
  if (state === 'stopped') return 'info'
  if (state === 'pending' || state === 'stopping') return 'warning'
  return 'danger'
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString('zh-CN') }

function showInstanceDetail(instance: EC2Instance) { currentInstance.value = instance; instanceDetailVisible.value = true }

async function startInstance(instance: EC2Instance) {
  await ElMessageBox.confirm(`确定启动实例 ${instance.name}？`, '确认')
  instance.state = 'pending'
  setTimeout(() => { instance.state = 'running'; ElMessage.success('实例已启动') }, 1000)
}

async function stopInstance(instance: EC2Instance) {
  await ElMessageBox.confirm(`确定停止实例 ${instance.name}？`, '确认')
  instance.state = 'stopping'
  setTimeout(() => { instance.state = 'stopped'; instance.publicIp = ''; ElMessage.success('实例已停止') }, 1000)
}

async function terminateInstance(instance: EC2Instance) {
  await ElMessageBox.confirm(`确定终止实例 ${instance.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  instances.value = instances.value.filter(i => i.instanceId !== instance.instanceId)
  ElMessage.success('实例已终止')
}

function showCreateBucketDialog() { bucketForm.value = { name: '', region: selectedRegion.value }; createBucketVisible.value = true }

async function createBucket() {
  if (!bucketForm.value.name) { ElMessage.warning('请输入存储桶名称'); return }
  saving.value = true
  setTimeout(() => {
    buckets.value.push({ name: bucketForm.value.name, region: bucketForm.value.region, createdAt: new Date().toISOString(), access: 'private' })
    saving.value = false; createBucketVisible.value = false
    ElMessage.success('存储桶已创建')
  }, 500)
}

function browseBucket(bucket: S3Bucket) { ElMessage.info(`浏览 ${bucket.name} 功能即将推出`) }

async function deleteBucket(bucket: S3Bucket) {
  await ElMessageBox.confirm(`确定删除存储桶 ${bucket.name}？`, '确认')
  buckets.value = buckets.value.filter(b => b.name !== bucket.name)
  ElMessage.success('存储桶已删除')
}

function loadDnsRecords() {
  dnsRecords.value = [
    { id: '1', type: 'A', name: 'example.com', value: '192.168.1.1', ttl: 300 },
    { id: '2', type: 'CNAME', name: 'www.example.com', value: 'example.com', ttl: 300 },
    { id: '3', type: 'MX', name: 'example.com', value: '10 mail.example.com', ttl: 3600 }
  ]
}

function showAddRecordDialog() { editingRecord.value = null; dnsForm.value = { type: 'A', name: '', value: '', ttl: 300 }; dnsDialogVisible.value = true }

function editDnsRecord(record: DnsRecord) { editingRecord.value = record; dnsForm.value = { ...record }; dnsDialogVisible.value = true }

async function saveDnsRecord() {
  if (!dnsForm.value.name || !dnsForm.value.value) { ElMessage.warning('请填写完整信息'); return }
  saving.value = true
  setTimeout(() => {
    if (editingRecord.value) { Object.assign(editingRecord.value, dnsForm.value); ElMessage.success('记录已更新') }
    else { dnsRecords.value.push({ id: Date.now().toString(), ...dnsForm.value }); ElMessage.success('记录已添加') }
    saving.value = false; dnsDialogVisible.value = false
  }, 500)
}

async function deleteDnsRecord(record: DnsRecord) {
  await ElMessageBox.confirm(`确定删除 ${record.name} 的 ${record.type} 记录？`, '确认')
  dnsRecords.value = dnsRecords.value.filter(r => r.id !== record.id)
  ElMessage.success('记录已删除')
}
</script>

<style lang="scss" scoped>
.aws-page { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;
  .header-left { display: flex; align-items: center; gap: 12px; .provider-icon { font-size: 32px; } h1 { font-size: 24px; font-weight: 600; margin: 0; } .subtitle { color: var(--text-secondary); font-size: 14px; margin: 0; } }
  .header-right { display: flex; gap: 12px; align-items: center; }
}

.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; .search-input { width: 300px; } }

.cloudwatch-section { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; }

.metric-card {
  .metric-value { font-size: 28px; font-weight: 600; margin-bottom: 8px; }
  .metric-trend { display: flex; align-items: center; gap: 4px; font-size: 14px; &.up { color: var(--el-color-success); } &.down { color: var(--el-color-danger); } }
}
</style>
