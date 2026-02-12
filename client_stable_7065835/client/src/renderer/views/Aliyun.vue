<template>
  <div class="aliyun-page">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="$router.push('/cloud')"><el-icon><ArrowLeft /></el-icon></el-button>
        <span class="provider-icon">🟠</span>
        <div>
          <h1>阿里云</h1>
          <p class="subtitle">ECS、OSS、DNS、CDN 管理</p>
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
      <!-- ECS 实例 -->
      <el-tab-pane label="ECS 实例" name="ecs">
        <div class="tab-header">
          <el-input v-model="ecsSearch" placeholder="搜索实例..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>
        <el-table :data="filteredInstances" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="instanceId" label="实例 ID" width="180" />
          <el-table-column prop="type" label="规格" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="publicIp" label="公网 IP" width="140" />
          <el-table-column prop="privateIp" label="私网 IP" width="140" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === '已停止'" text size="small" type="success" @click="startInstance(row)">启动</el-button>
              <el-button v-if="row.status === '运行中'" text size="small" type="warning" @click="stopInstance(row)">停止</el-button>
              <el-button text size="small" @click="showInstanceDetail(row)">详情</el-button>
              <el-button text size="small" type="danger" @click="releaseInstance(row)">释放</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- OSS 存储 -->
      <el-tab-pane label="OSS 存储" name="oss">
        <div class="tab-header">
          <el-input v-model="ossSearch" placeholder="搜索存储桶..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateBucketDialog"><el-icon><Plus /></el-icon>创建 Bucket</el-button>
        </div>
        <el-table :data="filteredBuckets" v-loading="loading" stripe>
          <el-table-column prop="name" label="Bucket 名称" min-width="200" />
          <el-table-column prop="region" label="区域" width="150" />
          <el-table-column prop="storageClass" label="存储类型" width="120" />
          <el-table-column prop="acl" label="读写权限" width="120">
            <template #default="{ row }">
              <el-tag :type="row.acl === 'private' ? 'info' : 'warning'" size="small">{{ row.acl === 'private' ? '私有' : '公共读' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="browseBucket(row)">浏览</el-button>
              <el-button text size="small" type="danger" @click="deleteBucket(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- DNS 解析 -->
      <el-tab-pane label="DNS 解析" name="dns">
        <div class="tab-header">
          <el-select v-model="selectedDomain" placeholder="选择域名" @change="loadDnsRecords">
            <el-option v-for="d in domains" :key="d" :label="d" :value="d" />
          </el-select>
          <el-button type="primary" @click="showAddRecordDialog" :disabled="!selectedDomain"><el-icon><Plus /></el-icon>添加记录</el-button>
        </div>
        <el-table :data="dnsRecords" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="80">
            <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="rr" label="主机记录" min-width="150" />
          <el-table-column prop="value" label="记录值" min-width="200" show-overflow-tooltip />
          <el-table-column prop="ttl" label="TTL" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ENABLE' ? 'success' : 'info'" size="small">{{ row.status === 'ENABLE' ? '正常' : '暂停' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="editDnsRecord(row)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deleteDnsRecord(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- CDN -->
      <el-tab-pane label="CDN" name="cdn">
        <div class="tab-header">
          <el-input v-model="cdnSearch" placeholder="搜索域名..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showAddCdnDialog"><el-icon><Plus /></el-icon>添加域名</el-button>
        </div>
        <el-table :data="filteredCdnDomains" v-loading="loading" stripe>
          <el-table-column prop="domain" label="加速域名" min-width="200" />
          <el-table-column prop="cname" label="CNAME" min-width="250" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small">{{ row.status === 'online' ? '正常' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="业务类型" width="120" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="refreshCdn(row)">刷新缓存</el-button>
              <el-button text size="small" type="danger" @click="deleteCdnDomain(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 实例详情对话框 -->
    <el-dialog v-model="instanceDetailVisible" title="实例详情" width="600px">
      <el-descriptions v-if="currentInstance" :column="2" border>
        <el-descriptions-item label="实例 ID">{{ currentInstance.instanceId }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ currentInstance.name }}</el-descriptions-item>
        <el-descriptions-item label="规格">{{ currentInstance.type }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="getStatusType(currentInstance.status)" size="small">{{ currentInstance.status }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="公网 IP">{{ currentInstance.publicIp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="私网 IP">{{ currentInstance.privateIp }}</el-descriptions-item>
        <el-descriptions-item label="可用区">{{ currentInstance.zone }}</el-descriptions-item>
        <el-descriptions-item label="镜像">{{ currentInstance.image }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 创建 Bucket 对话框 -->
    <el-dialog v-model="createBucketVisible" title="创建 OSS Bucket" width="500px">
      <el-form :model="bucketForm" label-width="100px">
        <el-form-item label="Bucket 名称"><el-input v-model="bucketForm.name" placeholder="全局唯一名称" /></el-form-item>
        <el-form-item label="区域"><el-select v-model="bucketForm.region" style="width: 100%">
          <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select></el-form-item>
        <el-form-item label="存储类型"><el-select v-model="bucketForm.storageClass" style="width: 100%">
          <el-option label="标准存储" value="Standard" />
          <el-option label="低频访问" value="IA" />
          <el-option label="归档存储" value="Archive" />
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
          <el-option v-for="t in ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV']" :key="t" :label="t" :value="t" />
        </el-select></el-form-item>
        <el-form-item label="主机记录"><el-input v-model="dnsForm.rr" placeholder="如 www、@、*" /></el-form-item>
        <el-form-item label="记录值"><el-input v-model="dnsForm.value" placeholder="记录值" /></el-form-item>
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
import { ArrowLeft, Refresh, Search, Plus } from '@element-plus/icons-vue'

interface ECSInstance { instanceId: string; name: string; type: string; status: string; publicIp: string; privateIp: string; zone: string; image: string }
interface OSSBucket { name: string; region: string; storageClass: string; acl: string; createdAt: string }
interface DnsRecord { id: string; type: string; rr: string; value: string; ttl: number; status: string }
interface CdnDomain { domain: string; cname: string; status: string; type: string }

const route = useRoute()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('ecs')
const selectedRegion = ref('cn-hangzhou')
const ecsSearch = ref('')
const ossSearch = ref('')
const cdnSearch = ref('')
const selectedDomain = ref('')

const regions = [
  { value: 'cn-hangzhou', label: '华东1（杭州）' },
  { value: 'cn-shanghai', label: '华东2（上海）' },
  { value: 'cn-beijing', label: '华北2（北京）' },
  { value: 'cn-shenzhen', label: '华南1（深圳）' },
  { value: 'cn-hongkong', label: '中国香港' }
]

const instances = ref<ECSInstance[]>([
  { instanceId: 'i-bp1abc123def', name: 'web-server-01', type: 'ecs.c6.large', status: '运行中', publicIp: '47.98.123.45', privateIp: '172.16.0.10', zone: 'cn-hangzhou-h', image: 'CentOS 7.9' },
  { instanceId: 'i-bp2def456ghi', name: 'db-server-01', type: 'ecs.r6.xlarge', status: '运行中', publicIp: '', privateIp: '172.16.0.20', zone: 'cn-hangzhou-h', image: 'Ubuntu 20.04' },
  { instanceId: 'i-bp3ghi789jkl', name: 'dev-server', type: 'ecs.t6-c1m1.large', status: '已停止', publicIp: '', privateIp: '172.16.0.30', zone: 'cn-hangzhou-i', image: 'Alibaba Cloud Linux 3' }
])

const buckets = ref<OSSBucket[]>([
  { name: 'my-app-static', region: 'cn-hangzhou', storageClass: 'Standard', acl: 'private', createdAt: '2024-01-15T10:30:00Z' },
  { name: 'backup-data-2024', region: 'cn-hangzhou', storageClass: 'IA', acl: 'private', createdAt: '2024-02-20T14:00:00Z' },
  { name: 'cdn-origin', region: 'cn-shanghai', storageClass: 'Standard', acl: 'public-read', createdAt: '2024-03-10T09:15:00Z' }
])

const domains = ref(['example.com', 'myapp.cn'])
const dnsRecords = ref<DnsRecord[]>([])

const cdnDomains = ref<CdnDomain[]>([
  { domain: 'cdn.example.com', cname: 'cdn.example.com.w.kunlunsl.com', status: 'online', type: '图片小文件' },
  { domain: 'static.myapp.cn', cname: 'static.myapp.cn.w.kunlunsl.com', status: 'online', type: '大文件下载' }
])

const instanceDetailVisible = ref(false)
const createBucketVisible = ref(false)
const dnsDialogVisible = ref(false)
const currentInstance = ref<ECSInstance | null>(null)
const editingRecord = ref<DnsRecord | null>(null)
const bucketForm = ref({ name: '', region: 'cn-hangzhou', storageClass: 'Standard' })
const dnsForm = ref({ type: 'A', rr: '', value: '', ttl: 600 })

const filteredInstances = computed(() => {
  if (!ecsSearch.value) return instances.value
  const q = ecsSearch.value.toLowerCase()
  return instances.value.filter(i => i.name.toLowerCase().includes(q) || i.instanceId.toLowerCase().includes(q))
})

const filteredBuckets = computed(() => {
  if (!ossSearch.value) return buckets.value
  return buckets.value.filter(b => b.name.toLowerCase().includes(ossSearch.value.toLowerCase()))
})

const filteredCdnDomains = computed(() => {
  if (!cdnSearch.value) return cdnDomains.value
  return cdnDomains.value.filter(d => d.domain.toLowerCase().includes(cdnSearch.value.toLowerCase()))
})

onMounted(() => {
  const tab = route.query.tab as string
  if (tab && ['ecs', 'oss', 'dns', 'cdn'].includes(tab)) activeTab.value = tab
})

function loadRegionData() { ElMessage.info(`已切换到 ${selectedRegion.value} 区域`) }
function refreshData() { ElMessage.success('数据已刷新') }

function getStatusType(status: string) {
  if (status === '运行中') return 'success'
  if (status === '已停止') return 'info'
  if (status === '启动中' || status === '停止中') return 'warning'
  return 'danger'
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString('zh-CN') }

function showInstanceDetail(instance: ECSInstance) { currentInstance.value = instance; instanceDetailVisible.value = true }

async function startInstance(instance: ECSInstance) {
  await ElMessageBox.confirm(`确定启动实例 ${instance.name}？`, '确认')
  instance.status = '启动中'
  setTimeout(() => { instance.status = '运行中'; ElMessage.success('实例已启动') }, 1000)
}

async function stopInstance(instance: ECSInstance) {
  await ElMessageBox.confirm(`确定停止实例 ${instance.name}？`, '确认')
  instance.status = '停止中'
  setTimeout(() => { instance.status = '已停止'; instance.publicIp = ''; ElMessage.success('实例已停止') }, 1000)
}

async function releaseInstance(instance: ECSInstance) {
  await ElMessageBox.confirm(`确定释放实例 ${instance.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  instances.value = instances.value.filter(i => i.instanceId !== instance.instanceId)
  ElMessage.success('实例已释放')
}

function showCreateBucketDialog() { bucketForm.value = { name: '', region: selectedRegion.value, storageClass: 'Standard' }; createBucketVisible.value = true }

async function createBucket() {
  if (!bucketForm.value.name) { ElMessage.warning('请输入 Bucket 名称'); return }
  saving.value = true
  setTimeout(() => {
    buckets.value.push({ name: bucketForm.value.name, region: bucketForm.value.region, storageClass: bucketForm.value.storageClass, acl: 'private', createdAt: new Date().toISOString() })
    saving.value = false; createBucketVisible.value = false
    ElMessage.success('Bucket 已创建')
  }, 500)
}

function browseBucket(bucket: OSSBucket) { ElMessage.info(`浏览 ${bucket.name} 功能即将推出`) }

async function deleteBucket(bucket: OSSBucket) {
  await ElMessageBox.confirm(`确定删除 Bucket ${bucket.name}？`, '确认')
  buckets.value = buckets.value.filter(b => b.name !== bucket.name)
  ElMessage.success('Bucket 已删除')
}

function loadDnsRecords() {
  dnsRecords.value = [
    { id: '1', type: 'A', rr: '@', value: '47.98.123.45', ttl: 600, status: 'ENABLE' },
    { id: '2', type: 'CNAME', rr: 'www', value: 'example.com', ttl: 600, status: 'ENABLE' },
    { id: '3', type: 'MX', rr: '@', value: '10 mx.example.com', ttl: 3600, status: 'ENABLE' }
  ]
}

function showAddRecordDialog() { editingRecord.value = null; dnsForm.value = { type: 'A', rr: '', value: '', ttl: 600 }; dnsDialogVisible.value = true }

function editDnsRecord(record: DnsRecord) { editingRecord.value = record; dnsForm.value = { type: record.type, rr: record.rr, value: record.value, ttl: record.ttl }; dnsDialogVisible.value = true }

async function saveDnsRecord() {
  if (!dnsForm.value.rr || !dnsForm.value.value) { ElMessage.warning('请填写完整信息'); return }
  saving.value = true
  setTimeout(() => {
    if (editingRecord.value) { Object.assign(editingRecord.value, dnsForm.value); ElMessage.success('记录已更新') }
    else { dnsRecords.value.push({ id: Date.now().toString(), ...dnsForm.value, status: 'ENABLE' }); ElMessage.success('记录已添加') }
    saving.value = false; dnsDialogVisible.value = false
  }, 500)
}

async function deleteDnsRecord(record: DnsRecord) {
  await ElMessageBox.confirm(`确定删除 ${record.rr} 的 ${record.type} 记录？`, '确认')
  dnsRecords.value = dnsRecords.value.filter(r => r.id !== record.id)
  ElMessage.success('记录已删除')
}

function showAddCdnDialog() { ElMessage.info('添加 CDN 域名功能即将推出') }

function refreshCdn(domain: CdnDomain) { ElMessage.success(`已提交 ${domain.domain} 的缓存刷新请求`) }

async function deleteCdnDomain(domain: CdnDomain) {
  await ElMessageBox.confirm(`确定删除加速域名 ${domain.domain}？`, '确认')
  cdnDomains.value = cdnDomains.value.filter(d => d.domain !== domain.domain)
  ElMessage.success('域名已删除')
}
</script>

<style lang="scss" scoped>
.aliyun-page { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;
  .header-left { display: flex; align-items: center; gap: 12px; .provider-icon { font-size: 32px; } h1 { font-size: 24px; font-weight: 600; margin: 0; } .subtitle { color: var(--text-secondary); font-size: 14px; margin: 0; } }
  .header-right { display: flex; gap: 12px; align-items: center; }
}

.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; .search-input { width: 300px; } }
</style>
