<template>
  <div class="tencent-page">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="$router.push('/cloud')"><el-icon><ArrowLeft /></el-icon></el-button>
        <span class="provider-icon">🔵</span>
        <div>
          <h1>腾讯云</h1>
          <p class="subtitle">CVM、COS、DNS、CDN 管理</p>
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
      <!-- CVM 实例 -->
      <el-tab-pane label="CVM 实例" name="cvm">
        <div class="tab-header">
          <el-input v-model="cvmSearch" placeholder="搜索实例..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>
        <el-table :data="filteredInstances" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="instanceId" label="实例 ID" width="180" />
          <el-table-column prop="type" label="机型" width="140" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="publicIp" label="公网 IP" width="140" />
          <el-table-column prop="privateIp" label="内网 IP" width="140" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === '已关机'" text size="small" type="success" @click="startInstance(row)">开机</el-button>
              <el-button v-if="row.status === '运行中'" text size="small" type="warning" @click="stopInstance(row)">关机</el-button>
              <el-button text size="small" @click="showInstanceDetail(row)">详情</el-button>
              <el-button text size="small" type="danger" @click="terminateInstance(row)">销毁</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- COS 存储 -->
      <el-tab-pane label="COS 存储" name="cos">
        <div class="tab-header">
          <el-input v-model="cosSearch" placeholder="搜索存储桶..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="showCreateBucketDialog"><el-icon><Plus /></el-icon>创建存储桶</el-button>
        </div>
        <el-table :data="filteredBuckets" v-loading="loading" stripe>
          <el-table-column prop="name" label="存储桶名称" min-width="200" />
          <el-table-column prop="region" label="所属地域" width="150" />
          <el-table-column prop="acl" label="访问权限" width="120">
            <template #default="{ row }">
              <el-tag :type="row.acl === 'private' ? 'info' : 'warning'" size="small">{{ row.acl === 'private' ? '私有读写' : '公有读' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="browseBucket(row)">文件管理</el-button>
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
          <el-table-column prop="type" label="记录类型" width="100">
            <template #default="{ row }"><el-tag size="small">{{ row.type }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="name" label="主机记录" min-width="150" />
          <el-table-column prop="value" label="记录值" min-width="200" show-overflow-tooltip />
          <el-table-column prop="line" label="线路类型" width="100" />
          <el-table-column prop="ttl" label="TTL" width="80" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ENABLE' ? 'success' : 'info'" size="small">{{ row.status === 'ENABLE' ? '启用' : '暂停' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="editDnsRecord(row)">修改</el-button>
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
          <el-table-column prop="domain" label="域名" min-width="200" />
          <el-table-column prop="cname" label="CNAME" min-width="250" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : row.status === 'processing' ? 'warning' : 'info'" size="small">
                {{ row.status === 'online' ? '已启动' : row.status === 'processing' ? '部署中' : '已关闭' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="serviceType" label="业务类型" width="120" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="purgeCache(row)">刷新缓存</el-button>
              <el-button text size="small" @click="prefetchUrl(row)">预热URL</el-button>
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
        <el-descriptions-item label="机型">{{ currentInstance.type }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="getStatusType(currentInstance.status)" size="small">{{ currentInstance.status }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="公网 IP">{{ currentInstance.publicIp || '-' }}</el-descriptions-item>
        <el-descriptions-item label="内网 IP">{{ currentInstance.privateIp }}</el-descriptions-item>
        <el-descriptions-item label="可用区">{{ currentInstance.zone }}</el-descriptions-item>
        <el-descriptions-item label="镜像">{{ currentInstance.image }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 创建存储桶对话框 -->
    <el-dialog v-model="createBucketVisible" title="创建 COS 存储桶" width="500px">
      <el-form :model="bucketForm" label-width="100px">
        <el-form-item label="存储桶名称"><el-input v-model="bucketForm.name" placeholder="全局唯一名称" /></el-form-item>
        <el-form-item label="所属地域"><el-select v-model="bucketForm.region" style="width: 100%">
          <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
        </el-select></el-form-item>
        <el-form-item label="访问权限"><el-select v-model="bucketForm.acl" style="width: 100%">
          <el-option label="私有读写" value="private" />
          <el-option label="公有读私有写" value="public-read" />
        </el-select></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createBucketVisible = false">取消</el-button>
        <el-button type="primary" @click="createBucket" :loading="saving">创建</el-button>
      </template>
    </el-dialog>

    <!-- DNS 记录对话框 -->
    <el-dialog v-model="dnsDialogVisible" :title="editingRecord ? '修改记录' : '添加记录'" width="500px">
      <el-form :model="dnsForm" label-width="80px">
        <el-form-item label="记录类型"><el-select v-model="dnsForm.type" :disabled="!!editingRecord">
          <el-option v-for="t in ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV']" :key="t" :label="t" :value="t" />
        </el-select></el-form-item>
        <el-form-item label="主机记录"><el-input v-model="dnsForm.name" placeholder="如 www、@、*" /></el-form-item>
        <el-form-item label="记录值"><el-input v-model="dnsForm.value" placeholder="记录值" /></el-form-item>
        <el-form-item label="线路类型"><el-select v-model="dnsForm.line">
          <el-option label="默认" value="默认" />
          <el-option label="电信" value="电信" />
          <el-option label="联通" value="联通" />
          <el-option label="移动" value="移动" />
        </el-select></el-form-item>
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

interface CVMInstance { instanceId: string; name: string; type: string; status: string; publicIp: string; privateIp: string; zone: string; image: string }
interface COSBucket { name: string; region: string; acl: string; createdAt: string }
interface DnsRecord { id: string; type: string; name: string; value: string; line: string; ttl: number; status: string }
interface CdnDomain { domain: string; cname: string; status: string; serviceType: string }

const route = useRoute()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('cvm')
const selectedRegion = ref('ap-guangzhou')
const cvmSearch = ref('')
const cosSearch = ref('')
const cdnSearch = ref('')
const selectedDomain = ref('')

const regions = [
  { value: 'ap-guangzhou', label: '华南地区（广州）' },
  { value: 'ap-shanghai', label: '华东地区（上海）' },
  { value: 'ap-beijing', label: '华北地区（北京）' },
  { value: 'ap-chengdu', label: '西南地区（成都）' },
  { value: 'ap-hongkong', label: '港澳台地区（香港）' },
  { value: 'ap-singapore', label: '亚太东南（新加坡）' }
]

const instances = ref<CVMInstance[]>([
  { instanceId: 'ins-abc123def', name: 'web-server-01', type: 'S5.MEDIUM4', status: '运行中', publicIp: '119.29.123.45', privateIp: '10.0.0.10', zone: 'ap-guangzhou-3', image: 'CentOS 7.9' },
  { instanceId: 'ins-def456ghi', name: 'db-server-01', type: 'S5.LARGE8', status: '运行中', publicIp: '', privateIp: '10.0.0.20', zone: 'ap-guangzhou-3', image: 'Ubuntu 20.04' },
  { instanceId: 'ins-ghi789jkl', name: 'test-server', type: 'S5.SMALL2', status: '已关机', publicIp: '', privateIp: '10.0.0.30', zone: 'ap-guangzhou-4', image: 'TencentOS Server 3.1' }
])

const buckets = ref<COSBucket[]>([
  { name: 'my-app-1250000000', region: 'ap-guangzhou', acl: 'private', createdAt: '2024-01-15T10:30:00Z' },
  { name: 'backup-data-1250000000', region: 'ap-guangzhou', acl: 'private', createdAt: '2024-02-20T14:00:00Z' },
  { name: 'static-files-1250000000', region: 'ap-shanghai', acl: 'public-read', createdAt: '2024-03-10T09:15:00Z' }
])

const domains = ref(['example.com', 'myapp.cn'])
const dnsRecords = ref<DnsRecord[]>([])

const cdnDomains = ref<CdnDomain[]>([
  { domain: 'cdn.example.com', cname: 'cdn.example.com.cdn.dnsv1.com', status: 'online', serviceType: '静态加速' },
  { domain: 'download.myapp.cn', cname: 'download.myapp.cn.cdn.dnsv1.com', status: 'online', serviceType: '下载加速' }
])

const instanceDetailVisible = ref(false)
const createBucketVisible = ref(false)
const dnsDialogVisible = ref(false)
const currentInstance = ref<CVMInstance | null>(null)
const editingRecord = ref<DnsRecord | null>(null)
const bucketForm = ref({ name: '', region: 'ap-guangzhou', acl: 'private' })
const dnsForm = ref({ type: 'A', name: '', value: '', line: '默认', ttl: 600 })

const filteredInstances = computed(() => {
  if (!cvmSearch.value) return instances.value
  const q = cvmSearch.value.toLowerCase()
  return instances.value.filter(i => i.name.toLowerCase().includes(q) || i.instanceId.toLowerCase().includes(q))
})

const filteredBuckets = computed(() => {
  if (!cosSearch.value) return buckets.value
  return buckets.value.filter(b => b.name.toLowerCase().includes(cosSearch.value.toLowerCase()))
})

const filteredCdnDomains = computed(() => {
  if (!cdnSearch.value) return cdnDomains.value
  return cdnDomains.value.filter(d => d.domain.toLowerCase().includes(cdnSearch.value.toLowerCase()))
})

onMounted(() => {
  const tab = route.query.tab as string
  if (tab && ['cvm', 'cos', 'dns', 'cdn'].includes(tab)) activeTab.value = tab
})

function loadRegionData() { ElMessage.info(`已切换到 ${selectedRegion.value} 地域`) }
function refreshData() { ElMessage.success('数据已刷新') }

function getStatusType(status: string) {
  if (status === '运行中') return 'success'
  if (status === '已关机') return 'info'
  if (status === '开机中' || status === '关机中') return 'warning'
  return 'danger'
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString('zh-CN') }

function showInstanceDetail(instance: CVMInstance) { currentInstance.value = instance; instanceDetailVisible.value = true }

async function startInstance(instance: CVMInstance) {
  await ElMessageBox.confirm(`确定开机实例 ${instance.name}？`, '确认')
  instance.status = '开机中'
  setTimeout(() => { instance.status = '运行中'; ElMessage.success('实例已开机') }, 1000)
}

async function stopInstance(instance: CVMInstance) {
  await ElMessageBox.confirm(`确定关机实例 ${instance.name}？`, '确认')
  instance.status = '关机中'
  setTimeout(() => { instance.status = '已关机'; instance.publicIp = ''; ElMessage.success('实例已关机') }, 1000)
}

async function terminateInstance(instance: CVMInstance) {
  await ElMessageBox.confirm(`确定销毁实例 ${instance.name}？此操作不可恢复！`, '警告', { type: 'warning' })
  instances.value = instances.value.filter(i => i.instanceId !== instance.instanceId)
  ElMessage.success('实例已销毁')
}

function showCreateBucketDialog() { bucketForm.value = { name: '', region: selectedRegion.value, acl: 'private' }; createBucketVisible.value = true }

async function createBucket() {
  if (!bucketForm.value.name) { ElMessage.warning('请输入存储桶名称'); return }
  saving.value = true
  setTimeout(() => {
    buckets.value.push({ name: `${bucketForm.value.name}-1250000000`, region: bucketForm.value.region, acl: bucketForm.value.acl, createdAt: new Date().toISOString() })
    saving.value = false; createBucketVisible.value = false
    ElMessage.success('存储桶已创建')
  }, 500)
}

function browseBucket(bucket: COSBucket) { ElMessage.info(`浏览 ${bucket.name} 功能即将推出`) }

async function deleteBucket(bucket: COSBucket) {
  await ElMessageBox.confirm(`确定删除存储桶 ${bucket.name}？`, '确认')
  buckets.value = buckets.value.filter(b => b.name !== bucket.name)
  ElMessage.success('存储桶已删除')
}

function loadDnsRecords() {
  dnsRecords.value = [
    { id: '1', type: 'A', name: '@', value: '119.29.123.45', line: '默认', ttl: 600, status: 'ENABLE' },
    { id: '2', type: 'CNAME', name: 'www', value: 'example.com', line: '默认', ttl: 600, status: 'ENABLE' },
    { id: '3', type: 'MX', name: '@', value: '10 mx.example.com', line: '默认', ttl: 3600, status: 'ENABLE' }
  ]
}

function showAddRecordDialog() { editingRecord.value = null; dnsForm.value = { type: 'A', name: '', value: '', line: '默认', ttl: 600 }; dnsDialogVisible.value = true }

function editDnsRecord(record: DnsRecord) { editingRecord.value = record; dnsForm.value = { type: record.type, name: record.name, value: record.value, line: record.line, ttl: record.ttl }; dnsDialogVisible.value = true }

async function saveDnsRecord() {
  if (!dnsForm.value.name || !dnsForm.value.value) { ElMessage.warning('请填写完整信息'); return }
  saving.value = true
  setTimeout(() => {
    if (editingRecord.value) { Object.assign(editingRecord.value, dnsForm.value); ElMessage.success('记录已更新') }
    else { dnsRecords.value.push({ id: Date.now().toString(), ...dnsForm.value, status: 'ENABLE' }); ElMessage.success('记录已添加') }
    saving.value = false; dnsDialogVisible.value = false
  }, 500)
}

async function deleteDnsRecord(record: DnsRecord) {
  await ElMessageBox.confirm(`确定删除 ${record.name} 的 ${record.type} 记录？`, '确认')
  dnsRecords.value = dnsRecords.value.filter(r => r.id !== record.id)
  ElMessage.success('记录已删除')
}

function showAddCdnDialog() { ElMessage.info('添加 CDN 域名功能即将推出') }

function purgeCache(domain: CdnDomain) { ElMessage.success(`已提交 ${domain.domain} 的缓存刷新请求`) }

function prefetchUrl(domain: CdnDomain) { ElMessage.success(`已提交 ${domain.domain} 的 URL 预热请求`) }

async function deleteCdnDomain(domain: CdnDomain) {
  await ElMessageBox.confirm(`确定删除加速域名 ${domain.domain}？`, '确认')
  cdnDomains.value = cdnDomains.value.filter(d => d.domain !== domain.domain)
  ElMessage.success('域名已删除')
}
</script>

<style lang="scss" scoped>
.tencent-page { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;
  .header-left { display: flex; align-items: center; gap: 12px; .provider-icon { font-size: 32px; } h1 { font-size: 24px; font-weight: 600; margin: 0; } .subtitle { color: var(--text-secondary); font-size: 14px; margin: 0; } }
  .header-right { display: flex; gap: 12px; align-items: center; }
}

.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; .search-input { width: 300px; } }
</style>
