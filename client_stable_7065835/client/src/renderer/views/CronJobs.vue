<template>
  <div class="cron-page">
    <div class="page-header">
      <div class="header-left">
        <h1>定时任务管理</h1>
        <p class="subtitle">管理服务器上的 Cron 定时任务</p>
      </div>
      <div class="header-right">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select" @change="loadCronJobs">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-button type="primary" @click="openAddDialog" :disabled="!selectedServer">
          <el-icon><Plus /></el-icon>
          添加任务
        </el-button>
        <el-button @click="loadCronJobs" :disabled="!selectedServer" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid" v-if="selectedServer">
      <div class="stat-card">
        <div class="stat-icon total"><el-icon :size="24"><Timer /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ cronJobs.length }}</div>
          <div class="stat-label">总任务数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon active"><el-icon :size="24"><VideoPlay /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ activeJobs }}</div>
          <div class="stat-label">已启用</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon disabled"><el-icon :size="24"><VideoPause /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ disabledJobs }}</div>
          <div class="stat-label">已禁用</div>
        </div>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="cron-list" v-if="selectedServer">
      <el-table :data="cronJobs" v-loading="loading" stripe>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间表达式" width="180">
          <template #default="{ row }">
            <code class="cron-expression">{{ row.schedule }}</code>
            <div class="cron-readable">{{ parseCronExpression(row.schedule) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="命令" min-width="300">
          <template #default="{ row }">
            <code class="cron-command">{{ row.command }}</code>
          </template>
        </el-table-column>
        <el-table-column label="描述" width="200">
          <template #default="{ row }">
            <span class="cron-comment">{{ row.comment || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row, $index }">
            <el-button-group size="small">
              <el-button @click="toggleJob($index)" :type="row.enabled ? 'warning' : 'success'">
                {{ row.enabled ? '禁用' : '启用' }}
              </el-button>
              <el-button @click="openEditDialog(row, $index)">编辑</el-button>
              <el-button type="danger" @click="deleteJob($index)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && cronJobs.length === 0" description="暂无定时任务">
        <el-button type="primary" @click="openAddDialog">添加任务</el-button>
      </el-empty>
    </div>

    <el-empty v-else description="请先选择一个已连接的服务器" />

    <!-- 添加/编辑任务对话框 -->
    <el-dialog v-model="showDialog" :title="editIndex >= 0 ? '编辑定时任务' : '添加定时任务'" width="600px">
      <el-form :model="formData" label-width="100px">
        <el-form-item label="描述">
          <el-input v-model="formData.comment" placeholder="任务描述（可选）" />
        </el-form-item>

        <el-form-item label="时间表达式" required>
          <div class="cron-input-group">
            <el-input v-model="formData.schedule" placeholder="* * * * *" class="cron-input" />
            <el-dropdown trigger="click" @command="applyPreset">
              <el-button>
                常用模板 <el-icon><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-for="preset in cronPresets" :key="preset.value" :command="preset.value">
                    <span class="preset-name">{{ preset.label }}</span>
                    <code class="preset-value">{{ preset.value }}</code>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="cron-help">
            <span>格式: 分 时 日 月 周</span>
            <span class="cron-preview" v-if="formData.schedule">
              {{ parseCronExpression(formData.schedule) }}
            </span>
          </div>
        </el-form-item>

        <el-form-item label="自定义时间">
          <div class="cron-fields">
            <div class="cron-field">
              <label>分钟</label>
              <el-input v-model="cronFields.minute" placeholder="*" size="small" @input="updateScheduleFromFields" />
            </div>
            <div class="cron-field">
              <label>小时</label>
              <el-input v-model="cronFields.hour" placeholder="*" size="small" @input="updateScheduleFromFields" />
            </div>
            <div class="cron-field">
              <label>日期</label>
              <el-input v-model="cronFields.day" placeholder="*" size="small" @input="updateScheduleFromFields" />
            </div>
            <div class="cron-field">
              <label>月份</label>
              <el-input v-model="cronFields.month" placeholder="*" size="small" @input="updateScheduleFromFields" />
            </div>
            <div class="cron-field">
              <label>星期</label>
              <el-input v-model="cronFields.weekday" placeholder="*" size="small" @input="updateScheduleFromFields" />
            </div>
          </div>
        </el-form-item>

        <el-form-item label="执行命令" required>
          <el-input
            v-model="formData.command"
            type="textarea"
            :rows="3"
            placeholder="要执行的命令"
          />
        </el-form-item>

        <el-form-item label="启用状态">
          <el-switch v-model="formData.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveJob" :loading="saving">
          {{ editIndex >= 0 ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Timer, VideoPlay, VideoPause, ArrowDown } from '@element-plus/icons-vue'

interface CronJob {
  schedule: string
  command: string
  comment: string
  enabled: boolean
  originalLine: string
}

const serverStore = useServerStore()

const selectedServer = ref<string | null>(null)
const loading = ref(false)
const saving = ref(false)
const showDialog = ref(false)
const editIndex = ref(-1)
const cronJobs = ref<CronJob[]>([])

const formData = ref({
  schedule: '',
  command: '',
  comment: '',
  enabled: true
})

const cronFields = ref({
  minute: '*',
  hour: '*',
  day: '*',
  month: '*',
  weekday: '*'
})

const connectedServers = computed(() => serverStore.connectedServers)
const activeJobs = computed(() => cronJobs.value.filter(j => j.enabled).length)
const disabledJobs = computed(() => cronJobs.value.filter(j => !j.enabled).length)

// 常用 cron 模板
const cronPresets = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天凌晨', value: '0 0 * * *' },
  { label: '每天早上6点', value: '0 6 * * *' },
  { label: '每天中午12点', value: '0 12 * * *' },
  { label: '每周一凌晨', value: '0 0 * * 1' },
  { label: '每月1号凌晨', value: '0 0 1 * *' },
  { label: '每5分钟', value: '*/5 * * * *' },
  { label: '每30分钟', value: '*/30 * * * *' },
  { label: '工作日早上9点', value: '0 9 * * 1-5' }
]

// 初始化选择第一个已连接的服务器
if (connectedServers.value.length > 0) {
  selectedServer.value = connectedServers.value[0].id
  loadCronJobs()
}

// 监听表达式变化，更新字段
watch(() => formData.value.schedule, (newVal) => {
  if (newVal) {
    const parts = newVal.trim().split(/\s+/)
    if (parts.length >= 5) {
      cronFields.value = {
        minute: parts[0],
        hour: parts[1],
        day: parts[2],
        month: parts[3],
        weekday: parts[4]
      }
    }
  }
})

function updateScheduleFromFields() {
  formData.value.schedule = `${cronFields.value.minute} ${cronFields.value.hour} ${cronFields.value.day} ${cronFields.value.month} ${cronFields.value.weekday}`
}

function applyPreset(value: string) {
  formData.value.schedule = value
}

// 解析 cron 表达式为可读文本
function parseCronExpression(expr: string): string {
  if (!expr) return ''
  const parts = expr.trim().split(/\s+/)
  if (parts.length < 5) return '无效表达式'

  const [minute, hour, day, month, weekday] = parts

  // 简单的解析逻辑
  if (expr === '* * * * *') return '每分钟执行'
  if (expr === '0 * * * *') return '每小时整点执行'
  if (expr === '0 0 * * *') return '每天凌晨执行'
  if (expr.match(/^\*\/(\d+) \* \* \* \*$/)) {
    const interval = expr.match(/^\*\/(\d+)/)?.[1]
    return `每 ${interval} 分钟执行`
  }
  if (expr.match(/^0 \*\/(\d+) \* \* \*$/)) {
    const interval = expr.match(/\*\/(\d+)/)?.[1]
    return `每 ${interval} 小时执行`
  }
  if (minute !== '*' && hour !== '*' && day === '*' && month === '*' && weekday === '*') {
    return `每天 ${hour}:${minute.padStart(2, '0')} 执行`
  }
  if (minute !== '*' && hour !== '*' && day === '*' && month === '*' && weekday !== '*') {
    const weekdays = ['日', '一', '二', '三', '四', '五', '六']
    if (weekday.includes('-')) {
      return `每周${weekday} ${hour}:${minute.padStart(2, '0')} 执行`
    }
    return `每周${weekdays[parseInt(weekday)] || weekday} ${hour}:${minute.padStart(2, '0')} 执行`
  }
  if (minute !== '*' && hour !== '*' && day !== '*' && month === '*') {
    return `每月 ${day} 日 ${hour}:${minute.padStart(2, '0')} 执行`
  }

  return `${minute} ${hour} ${day} ${month} ${weekday}`
}

// 加载 cron 任务
async function loadCronJobs() {
  if (!selectedServer.value) return

  loading.value = true
  cronJobs.value = []

  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value,
      'crontab',
      ['-l']
    )

    if (result.exit_code === 0 && result.stdout) {
      parseCrontab(result.stdout)
    } else if (result.stderr?.includes('no crontab')) {
      // 没有 crontab，这是正常的
      cronJobs.value = []
    } else if (result.exit_code !== 0) {
      ElMessage.warning('无法读取 crontab: ' + (result.stderr || '未知错误'))
    }
  } catch (error) {
    ElMessage.error('加载失败: ' + (error as Error).message)
  } finally {
    loading.value = false
  }
}

// 解析 crontab 内容
function parseCrontab(content: string) {
  const lines = content.split('\n')
  const jobs: CronJob[] = []
  let currentComment = ''

  for (const line of lines) {
    const trimmed = line.trim()

    // 跳过空行
    if (!trimmed) {
      currentComment = ''
      continue
    }

    // 注释行（可能是任务描述）
    if (trimmed.startsWith('#')) {
      // 检查是否是被禁用的任务
      const disabledMatch = trimmed.match(/^#\s*(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$/)
      if (disabledMatch) {
        jobs.push({
          schedule: disabledMatch[1],
          command: disabledMatch[2],
          comment: currentComment,
          enabled: false,
          originalLine: line
        })
        currentComment = ''
      } else {
        // 普通注释，可能是下一个任务的描述
        currentComment = trimmed.substring(1).trim()
      }
      continue
    }

    // 解析 cron 任务行
    const match = trimmed.match(/^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$/)
    if (match) {
      jobs.push({
        schedule: match[1],
        command: match[2],
        comment: currentComment,
        enabled: true,
        originalLine: line
      })
      currentComment = ''
    }
  }

  cronJobs.value = jobs
}

// 生成 crontab 内容
function generateCrontab(): string {
  const lines: string[] = []

  for (const job of cronJobs.value) {
    if (job.comment) {
      lines.push(`# ${job.comment}`)
    }
    if (job.enabled) {
      lines.push(`${job.schedule} ${job.command}`)
    } else {
      lines.push(`# ${job.schedule} ${job.command}`)
    }
  }

  return lines.join('\n') + '\n'
}

// 保存 crontab
async function saveCrontab() {
  if (!selectedServer.value) return false

  const content = generateCrontab()

  try {
    // 使用 echo 和管道来设置 crontab
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value,
      'bash',
      ['-c', `echo '${content.replace(/'/g, "'\\''")}' | crontab -`]
    )

    if (result.exit_code !== 0) {
      ElMessage.error('保存失败: ' + (result.stderr || '未知错误'))
      return false
    }

    return true
  } catch (error) {
    ElMessage.error('保存失败: ' + (error as Error).message)
    return false
  }
}

function openAddDialog() {
  editIndex.value = -1
  formData.value = {
    schedule: '0 0 * * *',
    command: '',
    comment: '',
    enabled: true
  }
  cronFields.value = {
    minute: '0',
    hour: '0',
    day: '*',
    month: '*',
    weekday: '*'
  }
  showDialog.value = true
}

function openEditDialog(job: CronJob, index: number) {
  editIndex.value = index
  formData.value = {
    schedule: job.schedule,
    command: job.command,
    comment: job.comment,
    enabled: job.enabled
  }
  showDialog.value = true
}

async function saveJob() {
  if (!formData.value.schedule.trim() || !formData.value.command.trim()) {
    ElMessage.warning('请填写时间表达式和命令')
    return
  }

  // 验证 cron 表达式
  const parts = formData.value.schedule.trim().split(/\s+/)
  if (parts.length < 5) {
    ElMessage.warning('无效的 cron 表达式')
    return
  }

  saving.value = true

  const job: CronJob = {
    schedule: formData.value.schedule.trim(),
    command: formData.value.command.trim(),
    comment: formData.value.comment.trim(),
    enabled: formData.value.enabled,
    originalLine: ''
  }

  if (editIndex.value >= 0) {
    cronJobs.value[editIndex.value] = job
  } else {
    cronJobs.value.push(job)
  }

  const success = await saveCrontab()

  if (success) {
    ElMessage.success(editIndex.value >= 0 ? '任务已更新' : '任务已添加')
    showDialog.value = false
    await loadCronJobs() // 重新加载以确保同步
  } else {
    // 回滚更改
    await loadCronJobs()
  }

  saving.value = false
}

async function toggleJob(index: number) {
  const job = cronJobs.value[index]
  job.enabled = !job.enabled

  saving.value = true
  const success = await saveCrontab()

  if (success) {
    ElMessage.success(job.enabled ? '任务已启用' : '任务已禁用')
  } else {
    job.enabled = !job.enabled // 回滚
  }

  saving.value = false
}

async function deleteJob(index: number) {
  const job = cronJobs.value[index]

  try {
    await ElMessageBox.confirm(
      `确定要删除这个定时任务吗？\n命令: ${job.command}`,
      '确认删除',
      { type: 'warning' }
    )
  } catch {
    return
  }

  cronJobs.value.splice(index, 1)

  saving.value = true
  const success = await saveCrontab()

  if (success) {
    ElMessage.success('任务已删除')
  } else {
    await loadCronJobs() // 回滚
  }

  saving.value = false
}
</script>

<style lang="scss" scoped>
.cron-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;

  .header-left {
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

  .header-right {
    display: flex;
    gap: 12px;

    .server-select {
      width: 200px;
    }
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;

    &.total {
      background-color: rgba(99, 102, 241, 0.1);
      color: #6366f1;
    }

    &.active {
      background-color: rgba(34, 197, 94, 0.1);
      color: #22c55e;
    }

    &.disabled {
      background-color: rgba(156, 163, 175, 0.1);
      color: #9ca3af;
    }
  }

  .stat-info {
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

.cron-list {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 16px;
}

.cron-expression {
  font-family: 'Fira Code', monospace;
  font-size: 13px;
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}

.cron-readable {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.cron-command {
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  word-break: break-all;
}

.cron-comment {
  color: var(--text-secondary);
  font-size: 13px;
}

.cron-input-group {
  display: flex;
  gap: 12px;

  .cron-input {
    flex: 1;
  }
}

.cron-help {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);

  .cron-preview {
    color: var(--primary-color);
  }
}

.cron-fields {
  display: flex;
  gap: 12px;

  .cron-field {
    flex: 1;
    text-align: center;

    label {
      display: block;
      font-size: 12px;
      color: var(--text-secondary);
      margin-bottom: 4px;
    }
  }
}

.preset-name {
  display: block;
  font-weight: 500;
}

.preset-value {
  display: block;
  font-size: 11px;
  color: var(--text-secondary);
  font-family: monospace;
}

:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
}
</style>
