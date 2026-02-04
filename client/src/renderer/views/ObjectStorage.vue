<template>
  <div class="object-storage-page">
    <div class="page-header">
      <div class="header-left">
        <h1>对象存储管理</h1>
        <p class="subtitle">管理云存储服务中的文件和存储桶</p>
      </div>
      <div class="header-right">
        <el-button @click="showProviderDialog" :type="currentProvider ? '' : 'warning'">
          <el-icon><Setting /></el-icon>{{ currentProvider ? '切换服务商' : '配置服务商' }}
        </el-button>
        <el-button @click="refreshData" :loading="loading" :disabled="!currentProvider">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <!-- 未配置提示 -->
    <el-alert v-if="!currentProvider" title="请先配置存储服务商" type="warning" show-icon :closable="false" class="config-alert">
      <template #default>
        点击右上角"配置服务商"按钮，选择并配置您的云存储服务。
      </template>
    </el-alert>

    <!-- 主内容区 -->
    <div v-else class="main-content">
      <!-- 存储桶选择 -->
      <div class="bucket-selector">
        <el-select v-model="selectedBucket" placeholder="选择存储桶" @change="loadObjects" filterable>
          <el-option v-for="bucket in buckets" :key="bucket.name" :label="bucket.name" :value="bucket.name">
            <span>{{ bucket.name }}</span>
            <span class="bucket-region">{{ bucket.region }}</span>
          </el-option>
        </el-select>
        <el-button type="primary" @click="showCreateBucketDialog">
          <el-icon><Plus /></el-icon>创建存储桶
        </el-button>
        <el-button type="danger" @click="deleteBucket" :disabled="!selectedBucket">
          <el-icon><Delete /></el-icon>删除存储桶
        </el-button>
      </div>

      <!-- 文件浏览器 -->
      <div v-if="selectedBucket" class="file-browser">
        <!-- 路径导航 -->
        <div class="path-nav">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item @click="navigateTo('')">
              <el-icon><FolderOpened /></el-icon>{{ selectedBucket }}
            </el-breadcrumb-item>
            <el-breadcrumb-item
              v-for="(part, index) in pathParts"
              :key="index"
              @click="navigateTo(pathParts.slice(0, index + 1).join('/'))"
            >
              {{ part }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <!-- 工具栏 -->
        <div class="toolbar">
          <el-input v-model="searchQuery" placeholder="搜索文件..." class="search-input" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <div class="toolbar-actions">
            <el-button @click="showUploadDialog">
              <el-icon><Upload /></el-icon>上传文件
            </el-button>
            <el-button @click="createFolder">
              <el-icon><FolderAdd /></el-icon>新建文件夹
            </el-button>
            <el-button type="danger" :disabled="selectedObjects.length === 0" @click="deleteSelected">
              <el-icon><Delete /></el-icon>删除选中
            </el-button>
          </div>
        </div>

        <!-- 文件列表 -->
        <el-table
          :data="filteredObjects"
          v-loading="loading"
          stripe
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="50" />
          <el-table-column label="名称" min-width="300">
            <template #default="{ row }">
              <div class="file-name" @click="handleObjectClick(row)">
                <el-icon v-if="row.isFolder" class="folder-icon"><Folder /></el-icon>
                <el-icon v-else class="file-icon"><Document /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="120">
            <template #default="{ row }">
              {{ row.isFolder ? '-' : formatSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="lastModified" label="修改时间" width="180">
            <template #default="{ row }">
              {{ row.isFolder ? '-' : formatDate(row.lastModified) }}
            </template>
          </el-table-column>
          <el-table-column prop="storageClass" label="存储类型" width="120">
            <template #default="{ row }">
              <el-tag v-if="!row.isFolder" size="small">{{ row.storageClass || 'STANDARD' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <template v-if="!row.isFolder">
                <el-button text size="small" @click="downloadObject(row)">下载</el-button>
                <el-button text size="small" @click="copyUrl(row)">复制链接</el-button>
                <el-button text size="small" type="danger" @click="deleteObject(row)">删除</el-button>
              </template>
              <template v-else>
                <el-button text size="small" @click="handleObjectClick(row)">打开</el-button>
                <el-button text size="small" type="danger" @click="deleteObject(row)">删除</el-button>
              </template>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="filteredObjects.length === 0 && !loading" description="暂无文件" />
      </div>
    </div>

    <!-- 服务商配置对话框 -->
    <el-dialog v-model="providerDialogVisible" title="配置存储服务商" width="550px">
      <el-form :model="providerForm" label-width="100px">
        <el-form-item label="服务商">
          <el-select v-model="providerForm.provider" placeholder="选择服务商" @change="onProviderChange">
            <el-option label="AWS S3" value="s3">
              <span class="provider-option">☁️ AWS S3</span>
            </el-option>
            <el-option label="阿里云 OSS" value="oss">
              <span class="provider-option">🌐 阿里云 OSS</span>
            </el-option>
            <el-option label="腾讯云 COS" value="cos">
              <span class="provider-option">🌐 腾讯云 COS</span>
            </el-option>
            <el-option label="MinIO" value="minio">
              <span class="provider-option">🗄️ MinIO (自托管)</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="Access Key">
          <el-input v-model="providerForm.accessKey" placeholder="输入 Access Key ID" />
        </el-form-item>
        <el-form-item label="Secret Key">
          <el-input v-model="providerForm.secretKey" type="password" placeholder="输入 Secret Access Key" show-password />
        </el-form-item>
        <el-form-item label="区域/端点">
          <el-input v-model="providerForm.endpoint" :placeholder="getEndpointPlaceholder()" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="clearProvider" v-if="currentProvider">清除配置</el-button>
        <el-button type="primary" @click="saveProvider" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 创建存储桶对话框 -->
    <el-dialog v-model="createBucketDialogVisible" title="创建存储桶" width="450px">
      <el-form :model="bucketForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="bucketForm.name" placeholder="输入存储桶名称（全局唯一）" />
        </el-form-item>
        <el-form-item label="区域">
          <el-select v-model="bucketForm.region" placeholder="选择区域">
            <el-option v-for="r in regions" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="访问权限">
          <el-select v-model="bucketForm.acl">
            <el-option label="私有" value="private" />
            <el-option label="公共读" value="public-read" />
            <el-option label="公共读写" value="public-read-write" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createBucketDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createBucket" :loading="saving">创建</el-button>
      </template>
    </el-dialog>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="500px">
      <el-upload
        ref="uploadRef"
        drag
        multiple
        :auto-upload="false"
        :file-list="uploadFileList"
        @change="handleUploadChange"
      >
        <el-icon class="el-icon--upload"><Upload /></el-icon>
        <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
      </el-upload>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="uploadFiles" :loading="uploading" :disabled="uploadFileList.length === 0">
          上传 ({{ uploadFileList.length }} 个文件)
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import {
  Setting, Refresh, Plus, Delete, Search, Upload,
  Folder, FolderOpened, FolderAdd, Document
} from '@element-plus/icons-vue'

interface Bucket {
  name: string
  region: string
  creationDate: string
}

interface StorageObject {
  key: string
  name: string
  size: number
  lastModified: string
  storageClass: string
  isFolder: boolean
}

interface ProviderConfig {
  provider: string
  accessKey: string
  secretKey: string
  endpoint: string
}

const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)

// 服务商配置
const currentProvider = ref<ProviderConfig | null>(null)
const providerDialogVisible = ref(false)
const providerForm = ref<ProviderConfig>({
  provider: 's3',
  accessKey: '',
  secretKey: '',
  endpoint: ''
})

// 存储桶
const buckets = ref<Bucket[]>([])
const selectedBucket = ref('')
const createBucketDialogVisible = ref(false)
const bucketForm = ref({
  name: '',
  region: 'us-east-1',
  acl: 'private'
})

// 对象浏览
const currentPath = ref('')
const objects = ref<StorageObject[]>([])
const selectedObjects = ref<StorageObject[]>([])
const searchQuery = ref('')

// 上传
const uploadDialogVisible = ref(false)
const uploadFileList = ref<UploadFile[]>([])
const uploadRef = ref()

// 区域列表
const regions = computed(() => {
  if (providerForm.value.provider === 's3') {
    return [
      { label: '美国东部 (弗吉尼亚)', value: 'us-east-1' },
      { label: '美国西部 (俄勒冈)', value: 'us-west-2' },
      { label: '欧洲 (爱尔兰)', value: 'eu-west-1' },
      { label: '亚太 (新加坡)', value: 'ap-southeast-1' },
      { label: '亚太 (东京)', value: 'ap-northeast-1' }
    ]
  } else if (providerForm.value.provider === 'oss') {
    return [
      { label: '华东1 (杭州)', value: 'oss-cn-hangzhou' },
      { label: '华东2 (上海)', value: 'oss-cn-shanghai' },
      { label: '华北2 (北京)', value: 'oss-cn-beijing' },
      { label: '华南1 (深圳)', value: 'oss-cn-shenzhen' }
    ]
  } else if (providerForm.value.provider === 'cos') {
    return [
      { label: '北京', value: 'ap-beijing' },
      { label: '上海', value: 'ap-shanghai' },
      { label: '广州', value: 'ap-guangzhou' },
      { label: '成都', value: 'ap-chengdu' }
    ]
  }
  return [{ label: '默认', value: 'default' }]
})

const pathParts = computed(() => {
  return currentPath.value ? currentPath.value.split('/').filter(p => p) : []
})

const filteredObjects = computed(() => {
  if (!searchQuery.value) return objects.value
  const q = searchQuery.value.toLowerCase()
  return objects.value.filter(o => o.name.toLowerCase().includes(q))
})

onMounted(() => {
  // 从 localStorage 加载配置
  const saved = localStorage.getItem('object_storage_config')
  if (saved) {
    try {
      currentProvider.value = JSON.parse(saved)
      loadBuckets()
    } catch (e) {
      console.error('Failed to parse saved config:', e)
    }
  }
})

function getEndpointPlaceholder() {
  switch (providerForm.value.provider) {
    case 's3': return '区域，如 us-east-1'
    case 'oss': return '区域，如 oss-cn-hangzhou'
    case 'cos': return '区域，如 ap-beijing'
    case 'minio': return '端点 URL，如 http://localhost:9000'
    default: return '端点或区域'
  }
}

function showProviderDialog() {
  if (currentProvider.value) {
    providerForm.value = { ...currentProvider.value }
  } else {
    providerForm.value = { provider: 's3', accessKey: '', secretKey: '', endpoint: '' }
  }
  providerDialogVisible.value = true
}

function onProviderChange() {
  providerForm.value.endpoint = ''
}

async function saveProvider() {
  if (!providerForm.value.accessKey || !providerForm.value.secretKey) {
    ElMessage.warning('请填写 Access Key 和 Secret Key')
    return
  }
  saving.value = true
  try {
    localStorage.setItem('object_storage_config', JSON.stringify(providerForm.value))
    currentProvider.value = { ...providerForm.value }
    providerDialogVisible.value = false
    ElMessage.success('配置已保存')
    await loadBuckets()
  } finally {
    saving.value = false
  }
}

function clearProvider() {
  ElMessageBox.confirm('确定清除存储配置吗？', '确认').then(() => {
    localStorage.removeItem('object_storage_config')
    currentProvider.value = null
    buckets.value = []
    objects.value = []
    selectedBucket.value = ''
    providerDialogVisible.value = false
    ElMessage.success('配置已清除')
  }).catch(() => {})
}

async function loadBuckets() {
  if (!currentProvider.value) return
  loading.value = true
  try {
    // 模拟 API 调用
    await new Promise(r => setTimeout(r, 500))
    buckets.value = [
      { name: 'my-app-assets', region: 'us-east-1', creationDate: '2024-01-15' },
      { name: 'backup-data', region: 'us-west-2', creationDate: '2024-02-20' },
      { name: 'logs-archive', region: 'us-east-1', creationDate: '2024-03-10' }
    ]
    if (buckets.value.length > 0 && !selectedBucket.value) {
      selectedBucket.value = buckets.value[0].name
      await loadObjects()
    }
  } finally {
    loading.value = false
  }
}

async function loadObjects() {
  if (!selectedBucket.value) return
  loading.value = true
  try {
    // 模拟 API 调用
    await new Promise(r => setTimeout(r, 300))
    objects.value = [
      { key: 'images/', name: 'images', size: 0, lastModified: '', storageClass: '', isFolder: true },
      { key: 'documents/', name: 'documents', size: 0, lastModified: '', storageClass: '', isFolder: true },
      { key: 'index.html', name: 'index.html', size: 15234, lastModified: '2024-06-15T10:30:00Z', storageClass: 'STANDARD', isFolder: false },
      { key: 'style.css', name: 'style.css', size: 8456, lastModified: '2024-06-15T10:30:00Z', storageClass: 'STANDARD', isFolder: false },
      { key: 'app.js', name: 'app.js', size: 45678, lastModified: '2024-06-14T15:20:00Z', storageClass: 'STANDARD', isFolder: false },
      { key: 'logo.png', name: 'logo.png', size: 125890, lastModified: '2024-06-10T08:00:00Z', storageClass: 'STANDARD', isFolder: false }
    ]
  } finally {
    loading.value = false
  }
}

function refreshData() {
  if (selectedBucket.value) {
    loadObjects()
  } else {
    loadBuckets()
  }
}

function navigateTo(path: string) {
  currentPath.value = path
  loadObjects()
}

function handleObjectClick(obj: StorageObject) {
  if (obj.isFolder) {
    currentPath.value = currentPath.value ? `${currentPath.value}/${obj.name}` : obj.name
    loadObjects()
  }
}

function handleSelectionChange(selection: StorageObject[]) {
  selectedObjects.value = selection
}

function showCreateBucketDialog() {
  bucketForm.value = { name: '', region: regions.value[0]?.value || '', acl: 'private' }
  createBucketDialogVisible.value = true
}

async function createBucket() {
  if (!bucketForm.value.name) {
    ElMessage.warning('请输入存储桶名称')
    return
  }
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    buckets.value.push({
      name: bucketForm.value.name,
      region: bucketForm.value.region,
      creationDate: new Date().toISOString()
    })
    createBucketDialogVisible.value = false
    ElMessage.success('存储桶已创建')
    selectedBucket.value = bucketForm.value.name
    await loadObjects()
  } finally {
    saving.value = false
  }
}

async function deleteBucket() {
  if (!selectedBucket.value) return
  await ElMessageBox.confirm(`确定删除存储桶 "${selectedBucket.value}" 吗？此操作不可恢复。`, '确认删除', { type: 'warning' })
  loading.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    buckets.value = buckets.value.filter(b => b.name !== selectedBucket.value)
    selectedBucket.value = buckets.value[0]?.name || ''
    objects.value = []
    ElMessage.success('存储桶已删除')
  } finally {
    loading.value = false
  }
}

function showUploadDialog() {
  uploadFileList.value = []
  uploadDialogVisible.value = true
}

function handleUploadChange(_file: UploadFile, fileList: UploadFile[]) {
  uploadFileList.value = fileList
}

async function uploadFiles() {
  if (uploadFileList.value.length === 0) return
  uploading.value = true
  try {
    // 模拟上传
    await new Promise(r => setTimeout(r, 1000))
    for (const file of uploadFileList.value) {
      objects.value.push({
        key: `${currentPath.value ? currentPath.value + '/' : ''}${file.name}`,
        name: file.name,
        size: file.size || 0,
        lastModified: new Date().toISOString(),
        storageClass: 'STANDARD',
        isFolder: false
      })
    }
    uploadDialogVisible.value = false
    ElMessage.success(`已上传 ${uploadFileList.value.length} 个文件`)
    uploadFileList.value = []
  } finally {
    uploading.value = false
  }
}

async function createFolder() {
  try {
    const result = await ElMessageBox.prompt('请输入文件夹名称', '新建文件夹', {
      inputPattern: /^[a-zA-Z0-9_-]+$/,
      inputErrorMessage: '文件夹名称只能包含字母、数字、下划线和连字符'
    })
    const name = typeof result === 'object' ? result.value : ''
    if (name) {
      objects.value.unshift({
        key: `${currentPath.value ? currentPath.value + '/' : ''}${name}/`,
        name: name,
        size: 0,
        lastModified: '',
        storageClass: '',
        isFolder: true
      })
      ElMessage.success('文件夹已创建')
    }
  } catch {
    // 用户取消
  }
}

async function deleteObject(obj: StorageObject) {
  await ElMessageBox.confirm(`确定删除 "${obj.name}" 吗？`, '确认删除')
  objects.value = objects.value.filter(o => o.key !== obj.key)
  ElMessage.success('已删除')
}

async function deleteSelected() {
  if (selectedObjects.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedObjects.value.length} 个项目吗？`, '确认删除')
  const keys = selectedObjects.value.map(o => o.key)
  objects.value = objects.value.filter(o => !keys.includes(o.key))
  selectedObjects.value = []
  ElMessage.success('已删除选中项目')
}

function downloadObject(obj: StorageObject) {
  // 模拟下载
  ElMessage.success(`开始下载 ${obj.name}`)
}

function copyUrl(obj: StorageObject) {
  const url = `https://${selectedBucket.value}.s3.amazonaws.com/${obj.key}`
  navigator.clipboard.writeText(url)
  ElMessage.success('链接已复制到剪贴板')
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.object-storage-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  .header-left {
    h1 { font-size: 24px; font-weight: 600; margin: 0; }
    .subtitle { color: var(--text-secondary); font-size: 14px; margin: 4px 0 0; }
  }

  .header-right {
    display: flex;
    gap: 12px;
  }
}

.config-alert {
  margin-bottom: 20px;
}

.main-content {
  .bucket-selector {
    display: flex;
    gap: 12px;
    margin-bottom: 20px;

    .el-select {
      width: 300px;
    }

    .bucket-region {
      float: right;
      color: var(--text-secondary);
      font-size: 12px;
    }
  }
}

.file-browser {
  background: var(--bg-color);
  border-radius: 8px;
  padding: 16px;

  .path-nav {
    margin-bottom: 16px;
    padding: 8px 12px;
    background: var(--bg-secondary);
    border-radius: 4px;

    .el-breadcrumb__item {
      cursor: pointer;
      &:hover { color: var(--el-color-primary); }
    }
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .search-input {
      width: 300px;
    }

    .toolbar-actions {
      display: flex;
      gap: 8px;
    }
  }

  .file-name {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;

    &:hover {
      color: var(--el-color-primary);
    }

    .folder-icon {
      color: #f0c040;
      font-size: 18px;
    }

    .file-icon {
      color: #909399;
      font-size: 18px;
    }
  }
}

.provider-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
