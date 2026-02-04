<template>
  <div class="cloud-page">
    <div class="page-header">
      <h1>云服务集成</h1>
      <p class="subtitle">管理您的云服务商账号和资源</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon connected"><el-icon><Link /></el-icon></div>
        <div class="stat-info">
          <span class="stat-value">{{ connectedCount }}</span>
          <span class="stat-label">已连接</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon total"><el-icon><Cloudy /></el-icon></div>
        <div class="stat-info">
          <span class="stat-value">{{ providers.length }}</span>
          <span class="stat-label">支持的服务商</span>
        </div>
      </div>
    </div>

    <!-- 已连接的服务商 -->
    <div v-if="connectedProviders.length > 0" class="section">
      <h2>已连接的服务</h2>
      <div class="connected-providers">
        <el-card
          v-for="provider in connectedProviders"
          :key="provider.id"
          class="connected-card"
          @click="manageProvider(provider)"
        >
          <div class="card-header">
            <span class="provider-emoji">{{ provider.emoji }}</span>
            <div class="provider-title">
              <h3>{{ provider.name }}</h3>
              <el-tag type="success" size="small">已连接</el-tag>
            </div>
            <el-dropdown @command="handleProviderAction($event, provider)" trigger="click">
              <el-button text @click.stop>
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="manage">管理</el-dropdown-item>
                  <el-dropdown-item command="refresh">刷新</el-dropdown-item>
                  <el-dropdown-item command="disconnect" divided>断开连接</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="quick-actions">
            <el-button
              v-for="action in provider.quickActions"
              :key="action.name"
              size="small"
              @click.stop="executeQuickAction(provider, action)"
            >
              {{ action.name }}
            </el-button>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 可用的服务商 -->
    <div class="section">
      <h2>{{ connectedProviders.length > 0 ? '添加更多服务' : '选择云服务商' }}</h2>
      <div class="provider-grid">
        <el-card
          v-for="provider in availableProviders"
          :key="provider.id"
          class="provider-card"
          @click="connectProvider(provider)"
        >
          <div class="provider-icon">{{ provider.emoji }}</div>
          <div class="provider-info">
            <h3>{{ provider.name }}</h3>
            <p>{{ provider.description }}</p>
          </div>
          <el-button type="primary" size="small">连接</el-button>
        </el-card>
      </div>
    </div>

    <!-- 配置对话框 -->
    <el-dialog
      v-model="showConfigDialog"
      :title="`连接 ${currentProvider?.name || ''}`"
      width="500px"
    >
      <el-form v-if="currentProvider" label-width="120px">
        <el-form-item
          v-for="field in configFields[currentProvider.id] || []"
          :key="field.label"
          :label="field.label"
        >
          <el-input
            v-model="configForm[field.label]"
            :type="field.type"
            :placeholder="field.placeholder"
            :show-password="field.type === 'password'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Link, Cloudy, MoreFilled } from '@element-plus/icons-vue'

const router = useRouter()

interface QuickAction {
  name: string
  action: string
}

interface CloudProvider {
  id: string
  name: string
  description: string
  emoji: string
  connected: boolean
  config?: Record<string, string>
  quickActions: QuickAction[]
}

const showConfigDialog = ref(false)
const currentProvider = ref<CloudProvider | null>(null)
const configForm = ref<Record<string, string>>({})

const providers = ref<CloudProvider[]>([
  {
    id: 'cloudflare',
    name: 'Cloudflare',
    description: 'DNS、CDN、WAF、SSL证书、Tunnel',
    emoji: '☁️',
    connected: false,
    quickActions: [
      { name: 'DNS 管理', action: 'dns' },
      { name: '域名列表', action: 'domains' },
      { name: '清除缓存', action: 'purge' }
    ]
  },
  {
    id: 'aws',
    name: 'Amazon Web Services',
    description: 'EC2、S3、Route53、CloudWatch',
    emoji: '🔶',
    connected: false,
    quickActions: [
      { name: 'EC2 实例', action: 'ec2' },
      { name: 'S3 存储桶', action: 's3' }
    ]
  },
  {
    id: 'aliyun',
    name: '阿里云',
    description: 'ECS、OSS、DNS、CDN',
    emoji: '🟠',
    connected: false,
    quickActions: [
      { name: 'ECS 实例', action: 'ecs' },
      { name: 'OSS 存储', action: 'oss' }
    ]
  },
  {
    id: 'tencent',
    name: '腾讯云',
    description: 'CVM、COS、DNS',
    emoji: '🔵',
    connected: false,
    quickActions: [
      { name: 'CVM 实例', action: 'cvm' },
      { name: 'COS 存储', action: 'cos' }
    ]
  },
  {
    id: 'digitalocean',
    name: 'DigitalOcean',
    description: 'Droplet、Spaces',
    emoji: '🌊',
    connected: false,
    quickActions: [
      { name: 'Droplets', action: 'droplets' },
      { name: 'Spaces', action: 'spaces' }
    ]
  }
])

// 加载保存的配置
loadSavedConfigs()

const connectedCount = computed(() => providers.value.filter(p => p.connected).length)
const connectedProviders = computed(() => providers.value.filter(p => p.connected))
const availableProviders = computed(() => providers.value.filter(p => !p.connected))

const configFields: Record<string, { label: string; type: string; placeholder: string }[]> = {
  cloudflare: [
    { label: 'API Token', type: 'password', placeholder: '输入 Cloudflare API Token' },
    { label: 'Account ID', type: 'text', placeholder: '输入 Account ID (可选)' }
  ],
  aws: [
    { label: 'Access Key ID', type: 'text', placeholder: '输入 AWS Access Key ID' },
    { label: 'Secret Access Key', type: 'password', placeholder: '输入 AWS Secret Access Key' },
    { label: 'Region', type: 'text', placeholder: '如 us-east-1' }
  ],
  aliyun: [
    { label: 'Access Key ID', type: 'text', placeholder: '输入阿里云 AccessKey ID' },
    { label: 'Access Key Secret', type: 'password', placeholder: '输入阿里云 AccessKey Secret' }
  ],
  tencent: [
    { label: 'Secret ID', type: 'text', placeholder: '输入腾讯云 SecretId' },
    { label: 'Secret Key', type: 'password', placeholder: '输入腾讯云 SecretKey' }
  ],
  digitalocean: [
    { label: 'API Token', type: 'password', placeholder: '输入 DigitalOcean API Token' }
  ]
}

function loadSavedConfigs() {
  const saved = localStorage.getItem('serverhub_cloud_providers')
  if (saved) {
    try {
      const configs = JSON.parse(saved)
      providers.value.forEach(p => {
        if (configs[p.id]) {
          p.connected = true
          p.config = configs[p.id]
        }
      })
    } catch { /* ignore */ }
  }
}

function saveConfigs() {
  const configs: Record<string, Record<string, string>> = {}
  providers.value.forEach(p => {
    if (p.connected && p.config) {
      configs[p.id] = p.config
    }
  })
  localStorage.setItem('serverhub_cloud_providers', JSON.stringify(configs))
}

function connectProvider(provider: CloudProvider) {
  currentProvider.value = provider
  configForm.value = {}
  showConfigDialog.value = true
}

function saveConfig() {
  if (!currentProvider.value) return

  const fields = configFields[currentProvider.value.id] || []
  const firstField = fields[0]
  if (firstField && !configForm.value[firstField.label]) {
    ElMessage.warning(`请输入 ${firstField.label}`)
    return
  }

  currentProvider.value.connected = true
  currentProvider.value.config = { ...configForm.value }
  saveConfigs()
  showConfigDialog.value = false
  ElMessage.success(`${currentProvider.value.name} 已连接`)
}

function manageProvider(provider: CloudProvider) {
  if (provider.id === 'cloudflare') {
    router.push('/cloud/cloudflare')
  } else if (provider.id === 'aws') {
    router.push('/cloud/aws')
  } else if (provider.id === 'aliyun') {
    router.push('/cloud/aliyun')
  } else if (provider.id === 'tencent') {
    router.push('/cloud/tencent')
  } else if (provider.id === 'digitalocean') {
    router.push('/cloud/digitalocean')
  } else {
    ElMessage.info(`${provider.name} 管理面板即将推出`)
  }
}

function handleProviderAction(action: string, provider: CloudProvider) {
  switch (action) {
    case 'manage':
      manageProvider(provider)
      break
    case 'refresh':
      ElMessage.success('已刷新')
      break
    case 'disconnect':
      ElMessageBox.confirm(`确定要断开 ${provider.name} 的连接吗？`, '确认').then(() => {
        provider.connected = false
        provider.config = undefined
        saveConfigs()
        ElMessage.info('已断开连接')
      }).catch(() => {})
      break
  }
}

function executeQuickAction(provider: CloudProvider, action: QuickAction) {
  if (provider.id === 'cloudflare') {
    // Navigate to Cloudflare page with specific tab
    const tabMap: Record<string, string> = {
      dns: 'dns',
      domains: 'dns',
      purge: 'cache'
    }
    const tab = tabMap[action.action] || 'dns'
    router.push(`/cloud/cloudflare?tab=${tab}`)
  } else if (provider.id === 'aws') {
    // Navigate to AWS page with specific tab
    const tabMap: Record<string, string> = {
      ec2: 'ec2',
      s3: 's3'
    }
    const tab = tabMap[action.action] || 'ec2'
    router.push(`/cloud/aws?tab=${tab}`)
  } else if (provider.id === 'aliyun') {
    // Navigate to Aliyun page with specific tab
    const tabMap: Record<string, string> = {
      ecs: 'ecs',
      oss: 'oss'
    }
    const tab = tabMap[action.action] || 'ecs'
    router.push(`/cloud/aliyun?tab=${tab}`)
  } else if (provider.id === 'tencent') {
    // Navigate to Tencent Cloud page with specific tab
    const tabMap: Record<string, string> = {
      cvm: 'cvm',
      cos: 'cos'
    }
    const tab = tabMap[action.action] || 'cvm'
    router.push(`/cloud/tencent?tab=${tab}`)
  } else if (provider.id === 'digitalocean') {
    // Navigate to DigitalOcean page with specific tab
    const tabMap: Record<string, string> = {
      droplets: 'droplets',
      spaces: 'spaces'
    }
    const tab = tabMap[action.action] || 'droplets'
    router.push(`/cloud/digitalocean?tab=${tab}`)
  } else {
    ElMessage.info(`${provider.name} - ${action.name} 功能即将推出`)
  }
}
</script>

<style lang="scss" scoped>
.cloud-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;

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

.stats-row {
  display: flex;
  gap: 16px;
  margin-bottom: 32px;

  .stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px 24px;
    background: var(--bg-secondary);
    border-radius: 12px;
    min-width: 180px;

    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 24px;

      &.connected {
        background: rgba(var(--el-color-success-rgb), 0.1);
        color: var(--el-color-success);
      }

      &.total {
        background: rgba(var(--el-color-primary-rgb), 0.1);
        color: var(--el-color-primary);
      }
    }

    .stat-info {
      display: flex;
      flex-direction: column;

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
}

.section {
  margin-bottom: 32px;

  h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-secondary);
  }
}

.connected-providers {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;

  .connected-card {
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: var(--el-color-primary);
    }

    .card-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 16px;

      .provider-emoji {
        font-size: 32px;
      }

      .provider-title {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 8px;

        h3 {
          font-size: 16px;
          font-weight: 600;
          margin: 0;
        }
      }
    }

    .quick-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
  }
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.provider-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: var(--el-color-primary);
  }

  .provider-icon {
    font-size: 40px;
    flex-shrink: 0;
  }

  .provider-info {
    flex: 1;

    h3 {
      font-size: 15px;
      font-weight: 600;
      margin-bottom: 4px;
    }

    p {
      font-size: 12px;
      color: var(--text-secondary);
      margin: 0;
    }
  }
}
</style>
