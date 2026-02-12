<template>
  <div class="alerts-page">
    <div class="page-header">
      <div class="header-left">
        <h1>告警系统</h1>
        <p class="subtitle">配置告警规则和通知渠道</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" @change="loadData">
          <el-option v-for="s in servers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="loadData" :loading="loading">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 告警规则 -->
      <el-tab-pane label="告警规则" name="rules">
        <div class="tab-header">
          <span class="tab-desc">配置资源监控告警阈值</span>
          <el-button type="primary" @click="showAddRuleDialog">
            <el-icon><Plus /></el-icon>添加规则
          </el-button>
        </div>
        <el-table :data="rules" v-loading="loading" stripe>
          <el-table-column prop="name" label="规则名称" min-width="150" />
          <el-table-column prop="metric" label="监控指标" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ getMetricLabel(row.metric) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="condition" label="触发条件" width="150">
            <template #default="{ row }">
              {{ row.operator }} {{ row.threshold }}{{ row.metric.includes('percent') ? '%' : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="持续时间" width="100">
            <template #default="{ row }">{{ row.duration }}s</template>
          </el-table-column>
          <el-table-column prop="severity" label="严重程度" width="100">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="80">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" @change="toggleRule(row)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="editRule(row)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deleteRule(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 通知渠道 -->
      <el-tab-pane label="通知渠道" name="channels">
        <div class="tab-header">
          <span class="tab-desc">配置告警通知方式</span>
          <el-button type="primary" @click="showAddChannelDialog">
            <el-icon><Plus /></el-icon>添加渠道
          </el-button>
        </div>
        <el-table :data="channels" v-loading="loading" stripe>
          <el-table-column prop="name" label="渠道名称" min-width="150" />
          <el-table-column prop="type" label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ getChannelTypeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="target" label="目标" min-width="200" show-overflow-tooltip />
          <el-table-column prop="enabled" label="状态" width="80">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button text size="small" @click="testChannel(row)">测试</el-button>
              <el-button text size="small" @click="editChannel(row)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deleteChannel(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 告警历史 -->
      <el-tab-pane label="告警历史" name="history">
        <div class="tab-header">
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" />
          <el-select v-model="historyFilter" placeholder="筛选状态" clearable>
            <el-option label="触发中" value="firing" />
            <el-option label="已确认" value="acknowledged" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </div>
        <el-table :data="filteredHistory" v-loading="loading" stripe>
          <el-table-column prop="ruleName" label="规则" min-width="150" />
          <el-table-column prop="severity" label="严重程度" width="100">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="value" label="触发值" width="100" />
          <el-table-column prop="triggeredAt" label="触发时间" width="180">
            <template #default="{ row }">{{ formatDate(row.triggeredAt) }}</template>
          </el-table-column>
          <el-table-column prop="resolvedAt" label="解决时间" width="180">
            <template #default="{ row }">{{ row.resolvedAt ? formatDate(row.resolvedAt) : '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'firing'" text size="small" @click="acknowledgeAlert(row)">确认</el-button>
              <el-button v-if="row.status !== 'resolved'" text size="small" type="success" @click="resolveAlert(row)">解决</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 添加/编辑规则对话框 -->
    <el-dialog v-model="ruleDialogVisible" :title="editingRule ? '编辑规则' : '添加规则'" width="500px">
      <el-form :model="ruleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="如: CPU 使用率过高" />
        </el-form-item>
        <el-form-item label="监控指标">
          <el-select v-model="ruleForm.metric">
            <el-option label="CPU 使用率" value="cpu_percent" />
            <el-option label="内存使用率" value="memory_percent" />
            <el-option label="磁盘使用率" value="disk_percent" />
            <el-option label="网络入流量" value="network_in" />
            <el-option label="网络出流量" value="network_out" />
            <el-option label="系统负载" value="load_avg" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发条件">
          <div style="display: flex; gap: 8px;">
            <el-select v-model="ruleForm.operator" style="width: 100px">
              <el-option label=">" value=">" />
              <el-option label=">=" value=">=" />
              <el-option label="<" value="<" />
              <el-option label="<=" value="<=" />
            </el-select>
            <el-input-number v-model="ruleForm.threshold" :min="0" :max="100" style="width: 150px" />
          </div>
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input-number v-model="ruleForm.duration" :min="10" :max="3600" />
          <span style="margin-left: 8px">秒</span>
        </el-form-item>
        <el-form-item label="严重程度">
          <el-select v-model="ruleForm.severity">
            <el-option label="信息" value="info" />
            <el-option label="警告" value="warning" />
            <el-option label="严重" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="通知渠道">
          <el-select v-model="ruleForm.channels" multiple placeholder="选择通知渠道">
            <el-option v-for="c in channels" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加/编辑渠道对话框 -->
    <el-dialog v-model="channelDialogVisible" :title="editingChannel ? '编辑渠道' : '添加渠道'" width="500px">
      <el-form :model="channelForm" label-width="100px">
        <el-form-item label="渠道名称">
          <el-input v-model="channelForm.name" placeholder="如: 运维邮箱" />
        </el-form-item>
        <el-form-item label="渠道类型">
          <el-select v-model="channelForm.type" @change="onChannelTypeChange">
            <el-option label="邮件" value="email" />
            <el-option label="Webhook" value="webhook" />
            <el-option label="Telegram" value="telegram" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="企业微信" value="wecom" />
          </el-select>
        </el-form-item>
        <el-form-item :label="getTargetLabel()">
          <el-input v-model="channelForm.target" :placeholder="getTargetPlaceholder()" />
        </el-form-item>
        <el-form-item label="Secret" v-if="channelForm.type === 'dingtalk'">
          <el-input v-model="channelForm.secret" type="password" placeholder="钉钉机器人加签密钥" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="channelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveChannel" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { useServerStore } from '../stores/server'

interface AlertRule {
  id: string; name: string; metric: string; operator: string; threshold: number
  duration: number; severity: string; enabled: boolean; channels: string[]
}
interface NotifyChannel {
  id: string; name: string; type: string; target: string; secret?: string; enabled: boolean
}
interface AlertHistory {
  id: string; ruleId: string; ruleName: string; severity: string; status: string
  value: string; triggeredAt: string; resolvedAt?: string
}

const serverStore = useServerStore()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('rules')
const selectedServer = ref('')
const servers = ref<Array<{ id: string; name: string }>>([])

const rules = ref<AlertRule[]>([])
const channels = ref<NotifyChannel[]>([])
const history = ref<AlertHistory[]>([])
const dateRange = ref<[Date, Date] | null>(null)
const historyFilter = ref('')

const ruleDialogVisible = ref(false)
const channelDialogVisible = ref(false)
const editingRule = ref<AlertRule | null>(null)
const editingChannel = ref<NotifyChannel | null>(null)

const ruleForm = ref({ name: '', metric: 'cpu_percent', operator: '>', threshold: 80, duration: 60, severity: 'warning', channels: [] as string[] })
const channelForm = ref({ name: '', type: 'email', target: '', secret: '' })

const filteredHistory = computed(() => {
  let result = history.value
  if (historyFilter.value) result = result.filter(h => h.status === historyFilter.value)
  if (dateRange.value) {
    const [start, end] = dateRange.value
    result = result.filter(h => {
      const t = new Date(h.triggeredAt)
      return t >= start && t <= end
    })
  }
  return result
})

onMounted(() => {
  servers.value = serverStore.servers.map(s => ({ id: s.id, name: s.name }))
  if (servers.value.length > 0) { selectedServer.value = servers.value[0].id; loadData() }
})

async function loadData() {
  loading.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    rules.value = [
      { id: '1', name: 'CPU 使用率过高', metric: 'cpu_percent', operator: '>', threshold: 80, duration: 60, severity: 'warning', enabled: true, channels: ['1'] },
      { id: '2', name: '内存使用率过高', metric: 'memory_percent', operator: '>', threshold: 90, duration: 60, severity: 'critical', enabled: true, channels: ['1', '2'] },
      { id: '3', name: '磁盘空间不足', metric: 'disk_percent', operator: '>', threshold: 85, duration: 300, severity: 'warning', enabled: true, channels: ['1'] }
    ]
    channels.value = [
      { id: '1', name: '运维邮箱', type: 'email', target: 'ops@example.com', enabled: true },
      { id: '2', name: '钉钉群', type: 'dingtalk', target: 'https://oapi.dingtalk.com/robot/send?access_token=xxx', enabled: true },
      { id: '3', name: 'Webhook', type: 'webhook', target: 'https://api.example.com/alerts', enabled: false }
    ]
    history.value = [
      { id: '1', ruleId: '1', ruleName: 'CPU 使用率过高', severity: 'warning', status: 'resolved', value: '85%', triggeredAt: '2024-06-15T10:30:00Z', resolvedAt: '2024-06-15T10:45:00Z' },
      { id: '2', ruleId: '2', ruleName: '内存使用率过高', severity: 'critical', status: 'firing', value: '92%', triggeredAt: '2024-06-15T14:00:00Z' },
      { id: '3', ruleId: '3', ruleName: '磁盘空间不足', severity: 'warning', status: 'acknowledged', value: '87%', triggeredAt: '2024-06-14T08:00:00Z' }
    ]
  } finally { loading.value = false }
}

function getMetricLabel(m: string) {
  const labels: Record<string, string> = { cpu_percent: 'CPU', memory_percent: '内存', disk_percent: '磁盘', network_in: '网络入', network_out: '网络出', load_avg: '负载' }
  return labels[m] || m
}
function getSeverityType(s: string) { return s === 'critical' ? 'danger' : s === 'warning' ? 'warning' : 'info' }
function getStatusType(s: string) { return s === 'firing' ? 'danger' : s === 'acknowledged' ? 'warning' : 'success' }
function getStatusLabel(s: string) { return s === 'firing' ? '触发中' : s === 'acknowledged' ? '已确认' : '已解决' }
function getChannelTypeLabel(t: string) {
  const labels: Record<string, string> = { email: '邮件', webhook: 'Webhook', telegram: 'Telegram', dingtalk: '钉钉', wecom: '企业微信' }
  return labels[t] || t
}
function getTargetLabel() {
  const labels: Record<string, string> = { email: '邮箱地址', webhook: 'Webhook URL', telegram: 'Bot Token', dingtalk: 'Webhook URL', wecom: 'Webhook URL' }
  return labels[channelForm.value.type] || '目标'
}
function getTargetPlaceholder() {
  const ph: Record<string, string> = { email: 'user@example.com', webhook: 'https://...', telegram: '123456:ABC-DEF...', dingtalk: 'https://oapi.dingtalk.com/...', wecom: 'https://qyapi.weixin.qq.com/...' }
  return ph[channelForm.value.type] || ''
}

function showAddRuleDialog() { editingRule.value = null; ruleForm.value = { name: '', metric: 'cpu_percent', operator: '>', threshold: 80, duration: 60, severity: 'warning', channels: [] }; ruleDialogVisible.value = true }
function editRule(r: AlertRule) { editingRule.value = r; ruleForm.value = { ...r }; ruleDialogVisible.value = true }
async function saveRule() {
  if (!ruleForm.value.name) { ElMessage.warning('请输入规则名称'); return }
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    if (editingRule.value) { Object.assign(editingRule.value, ruleForm.value); ElMessage.success('规则已更新') }
    else { rules.value.push({ id: Date.now().toString(), ...ruleForm.value, enabled: true }); ElMessage.success('规则已添加') }
    ruleDialogVisible.value = false
  } finally { saving.value = false }
}
async function deleteRule(r: AlertRule) { await ElMessageBox.confirm(`确定删除规则 "${r.name}" 吗？`, '确认'); rules.value = rules.value.filter(x => x.id !== r.id); ElMessage.success('规则已删除') }
function toggleRule(r: AlertRule) { ElMessage.success(`规则 "${r.name}" 已${r.enabled ? '启用' : '禁用'}`) }

function showAddChannelDialog() { editingChannel.value = null; channelForm.value = { name: '', type: 'email', target: '', secret: '' }; channelDialogVisible.value = true }
function editChannel(c: NotifyChannel) { editingChannel.value = c; channelForm.value = { ...c, secret: c.secret || '' }; channelDialogVisible.value = true }
function onChannelTypeChange() { channelForm.value.target = ''; channelForm.value.secret = '' }
async function saveChannel() {
  if (!channelForm.value.name || !channelForm.value.target) { ElMessage.warning('请填写完整信息'); return }
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    if (editingChannel.value) { Object.assign(editingChannel.value, channelForm.value); ElMessage.success('渠道已更新') }
    else { channels.value.push({ id: Date.now().toString(), ...channelForm.value, enabled: true }); ElMessage.success('渠道已添加') }
    channelDialogVisible.value = false
  } finally { saving.value = false }
}
async function deleteChannel(c: NotifyChannel) { await ElMessageBox.confirm(`确定删除渠道 "${c.name}" 吗？`, '确认'); channels.value = channels.value.filter(x => x.id !== c.id); ElMessage.success('渠道已删除') }
async function testChannel(c: NotifyChannel) { ElMessage.success(`测试通知已发送到 ${c.name}`) }

async function acknowledgeAlert(a: AlertHistory) { a.status = 'acknowledged'; ElMessage.success('告警已确认') }
async function resolveAlert(a: AlertHistory) { a.status = 'resolved'; a.resolvedAt = new Date().toISOString(); ElMessage.success('告警已解决') }
function formatDate(d: string) { return new Date(d).toLocaleString('zh-CN') }
</script>

<style lang="scss" scoped>
.alerts-page { max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;
  .header-left { h1 { font-size: 24px; font-weight: 600; margin: 0; } .subtitle { color: var(--text-secondary); font-size: 14px; margin: 4px 0 0; } }
  .header-right { display: flex; gap: 12px; }
}
.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; gap: 12px;
  .tab-desc { color: var(--text-secondary); font-size: 14px; }
}
</style>
