<template>
  <div class="environment">
    <div class="page-header">
      <div class="header-left">
        <h1>环境包管理</h1>
        <p class="subtitle">一键安装开发语言运行环境</p>
      </div>
      <div class="header-actions">
        <el-button @click="checkAllStatus" :loading="checking" size="small" round>
          <el-icon><Refresh /></el-icon>刷新状态
        </el-button>
      </div>
    </div>

    <div v-if="!serverStore.currentServer" class="no-server">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <!-- 已安装环境 -->
      <div class="section animate-in" v-if="installedPkgs.length > 0">
        <div class="section-header">
          <h2>已安装环境</h2>
          <el-tag type="success" size="small" round>{{ installedPkgs.length }} 个</el-tag>
        </div>
        <div class="pkg-grid" v-loading="checking">
          <div v-for="(pkg, idx) in installedPkgs" :key="pkg.name"
               class="pkg-card installed animate-card" :style="{ animationDelay: idx * 0.05 + 's' }">
            <div class="pkg-icon" :style="{ background: pkg.color }">
              <TechIcon :name="pkg.iconName" />
            </div>
            <div class="pkg-body">
              <div class="pkg-name">{{ pkg.name }}</div>
              <div class="pkg-version"><el-icon><CircleCheck /></el-icon> v{{ pkg.version }}</div>
              <div class="pkg-components">
                <el-tag v-for="c in pkg.components" :key="c" size="small" class="comp-tag" round>{{ c }}</el-tag>
              </div>
            </div>
            <div class="pkg-actions">
              <el-button size="small" text @click="showDetail(pkg)">详情</el-button>
              <el-button size="small" text type="danger" @click="uninstallPkg(pkg)">卸载</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 可安装环境 -->
      <div class="section animate-in" style="animation-delay: 0.1s">
        <div class="section-header">
          <h2>可安装环境</h2>
        </div>
        <div class="pkg-grid">
          <div v-for="(pkg, idx) in availablePkgs" :key="pkg.name"
               class="pkg-card animate-card" :style="{ animationDelay: (idx * 0.05 + 0.15) + 's' }">
            <div class="pkg-icon" :style="{ background: pkg.color }">
              <TechIcon :name="pkg.iconName" />
            </div>
            <div class="pkg-body">
              <div class="pkg-name">{{ pkg.name }}</div>
              <div class="pkg-desc">{{ pkg.description }}</div>
              <div class="pkg-components">
                <el-tag v-for="c in pkg.components" :key="c" size="small" class="comp-tag" round>{{ c }}</el-tag>
              </div>
            </div>
            <div class="pkg-actions">
              <el-button type="primary" size="small" round @click="openInstall(pkg)">安装</el-button>
            </div>
          </div>
        </div>
        <el-empty v-if="availablePkgs.length === 0 && !checking" description="所有环境包已安装" />
      </div>
    </template>

    <!-- 安装对话框 -->
    <el-dialog v-model="showInstallDlg" :title="`安装 ${curPkg?.name || ''}`" width="520px" class="dark-dialog" destroy-on-close>
      <div v-if="curPkg" class="install-content">
        <div class="install-preview">
          <div class="pkg-icon large" :style="{ background: curPkg.color }">
            <TechIcon :name="curPkg.iconName" />
          </div>
          <div>
            <div class="pkg-name">{{ curPkg.name }}</div>
            <div class="pkg-desc">{{ curPkg.description }}</div>
          </div>
        </div>
        <el-form label-position="top" v-if="curPkg.versions?.length">
          <el-form-item label="选择版本">
            <el-select v-model="selVersion" style="width:100%">
              <el-option v-for="v in curPkg.versions" :key="v.value" :label="v.label" :value="v.value">
                <span>{{ v.label }}</span>
                <el-tag v-if="v.recommended" type="success" size="small" style="margin-left:8px">推荐</el-tag>
              </el-option>
            </el-select>
          </el-form-item>
        </el-form>
        <div class="install-info">
          <el-icon><InfoFilled /></el-icon>
          <span>将安装：{{ curPkg.components.join('、') }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showInstallDlg = false">取消</el-button>
        <el-button type="primary" @click="doInstall" :loading="installing">
          <el-icon><Download /></el-icon>开始安装
        </el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框 -->
    <el-dialog v-model="showDetailDlg" :title="detailPkg?.name + ' 详情'" width="500px" class="dark-dialog" destroy-on-close>
      <el-descriptions :column="1" border v-if="detailPkg">
        <el-descriptions-item label="名称">{{ detailPkg.name }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ detailPkg.version || '未安装' }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ detailPkg.description }}</el-descriptions-item>
        <el-descriptions-item label="包含组件">{{ detailPkg.components.join('、') }}</el-descriptions-item>
        <el-descriptions-item label="检测命令"><code>{{ detailPkg.checkCmd }}</code></el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useServerStore } from '@/stores/server'
import { useTaskStore } from '@/stores/task'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, CircleCheck, InfoFilled, Download } from '@element-plus/icons-vue'
import TechIcon from '@/components/icons/TechIcons.vue'

interface PkgVersion { value: string; label: string; recommended?: boolean }
interface PkgDef {
  name: string; iconName: string; color: string; description: string
  components: string[]; checkCmd: string; versions?: PkgVersion[]
  getSteps: (v: string) => { cmd: string; desc: string }[]
  uninstallSteps?: { cmd: string; desc: string }[]
  installed?: boolean; version?: string
}

const serverStore = useServerStore()
const taskStore = useTaskStore()
const checking = ref(false)
const installing = ref(false)
const showInstallDlg = ref(false)
const showDetailDlg = ref(false)
const curPkg = ref<PkgDef | null>(null)
const detailPkg = ref<PkgDef | null>(null)
const selVersion = ref('')

const pkgDefs: PkgDef[] = [
  {
    name: 'Node.js 环境包', iconName: 'nodejs', color: '#68a063',
    description: 'JavaScript 运行时环境',
    components: ['Node.js', 'npm', 'PM2', 'yarn'],
    checkCmd: 'node --version 2>/dev/null | tr -d "v"',
    versions: [
      { value: '20', label: 'Node.js 20 LTS', recommended: true },
      { value: '22', label: 'Node.js 22 (最新)' },
      { value: '18', label: 'Node.js 18 LTS' }
    ],
    getSteps: (v) => [
      { cmd: `curl -fsSL https://deb.nodesource.com/setup_${v}.x -o /tmp/nodesource_setup.sh`, desc: '下载安装脚本' },
      { cmd: 'sudo bash /tmp/nodesource_setup.sh', desc: '配置软件源' },
      { cmd: 'sudo apt-get install -y nodejs', desc: '安装 Node.js' },
      { cmd: 'sudo npm install -g pm2 yarn', desc: '安装 PM2 和 Yarn' }
    ],
    uninstallSteps: [
      { cmd: 'sudo npm uninstall -g pm2 yarn 2>/dev/null; sudo apt-get remove -y nodejs', desc: '卸载 Node.js' },
      { cmd: 'sudo apt-get autoremove -y', desc: '清理依赖' }
    ]
  },
  {
    name: 'Python 环境包', iconName: 'python', color: '#3776ab',
    description: 'Python 解释器与包管理',
    components: ['Python 3', 'pip', 'venv', 'virtualenv'],
    checkCmd: 'python3 --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    getSteps: () => [
      { cmd: 'sudo apt-get update', desc: '更新软件源' },
      { cmd: 'sudo apt-get install -y python3 python3-pip python3-venv', desc: '安装 Python 3' },
      { cmd: 'pip3 install virtualenv', desc: '安装 virtualenv' }
    ],
    uninstallSteps: [
      { cmd: 'sudo apt-get remove -y python3 python3-pip python3-venv', desc: '卸载 Python' },
      { cmd: 'sudo apt-get autoremove -y', desc: '清理依赖' }
    ]
  },
  {
    name: 'PHP 环境包', iconName: 'php', color: '#777bb4',
    description: 'Web 服务端脚本语言',
    components: ['PHP', 'Composer', 'php-fpm', '常用扩展'],
    checkCmd: 'php -v 2>/dev/null | head -1 | grep -oP "\\d+\\.\\d+\\.\\d+"',
    versions: [
      { value: '8.3', label: 'PHP 8.3', recommended: true },
      { value: '8.2', label: 'PHP 8.2' },
      { value: '8.1', label: 'PHP 8.1' }
    ],
    getSteps: (v) => [
      { cmd: 'sudo add-apt-repository ppa:ondrej/php -y && sudo apt-get update', desc: '添加 PHP 仓库' },
      { cmd: `sudo apt-get install -y php${v} php${v}-fpm php${v}-cli php${v}-common php${v}-mysql php${v}-curl php${v}-gd php${v}-mbstring php${v}-xml php${v}-zip`, desc: '安装 PHP 和扩展' },
      { cmd: 'curl -sS https://getcomposer.org/installer | php && sudo mv composer.phar /usr/local/bin/composer', desc: '安装 Composer' }
    ],
    uninstallSteps: [
      { cmd: 'sudo apt-get remove -y php* 2>/dev/null', desc: '卸载 PHP' },
      { cmd: 'sudo apt-get autoremove -y', desc: '清理依赖' }
    ]
  },
  {
    name: 'Java 环境包', iconName: 'java', color: '#f89820',
    description: '企业级应用开发平台',
    components: ['OpenJDK', 'Maven'],
    checkCmd: 'java --version 2>&1 | head -1 | grep -oP "\\d+\\.\\d+\\.\\d+" || java -version 2>&1 | head -1 | grep -oP "\\d+\\.\\d+\\.\\d+"',
    versions: [
      { value: '21', label: 'OpenJDK 21', recommended: true },
      { value: '17', label: 'OpenJDK 17 LTS' },
      { value: '11', label: 'OpenJDK 11 LTS' }
    ],
    getSteps: (v) => [
      { cmd: `sudo apt-get update && sudo apt-get install -y openjdk-${v}-jdk`, desc: '安装 OpenJDK' },
      { cmd: 'sudo apt-get install -y maven', desc: '安装 Maven' }
    ],
    uninstallSteps: [
      { cmd: 'sudo apt-get remove -y openjdk-*-jdk maven', desc: '卸载 Java' },
      { cmd: 'sudo apt-get autoremove -y', desc: '清理依赖' }
    ]
  },
  {
    name: 'Go 环境包', iconName: 'go', color: '#00add8',
    description: 'Go 语言环境',
    components: ['Go', 'Go Modules'],
    checkCmd: 'go version 2>/dev/null | grep -oP "go\\K[0-9.]+"',
    getSteps: () => [
      { cmd: 'sudo apt-get update && sudo apt-get install -y golang-go', desc: '安装 Go' }
    ],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y golang-go && sudo apt-get autoremove -y', desc: '卸载 Go' }]
  },
  {
    name: '.NET 环境包', iconName: 'dotnet', color: '#512bd4',
    description: '微软跨平台开发框架',
    components: ['.NET SDK', 'ASP.NET Core', 'dotnet CLI'],
    checkCmd: 'dotnet --version 2>/dev/null',
    versions: [
      { value: '8', label: '.NET 8 LTS', recommended: true },
      { value: '7', label: '.NET 7' }
    ],
    getSteps: (v) => [
      { cmd: 'wget https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb -O /tmp/ms.deb && sudo dpkg -i /tmp/ms.deb', desc: '添加 Microsoft 仓库' },
      { cmd: `sudo apt-get update && sudo apt-get install -y dotnet-sdk-${v}.0`, desc: '安装 .NET SDK' }
    ],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y dotnet-sdk-* && sudo apt-get autoremove -y', desc: '卸载 .NET' }]
  },
  {
    name: 'Rust 环境包', iconName: 'rust', color: '#ce422b',
    description: '系统级编程语言',
    components: ['Rust', 'Cargo', 'rustup'],
    checkCmd: 'rustc --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    getSteps: () => [
      { cmd: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y", desc: '安装 Rust 工具链' },
      { cmd: 'source $HOME/.cargo/env && rustc --version', desc: '验证安装' }
    ],
    uninstallSteps: [{ cmd: 'rustup self uninstall -y 2>/dev/null || true', desc: '卸载 Rust' }]
  },
  {
    name: 'Ruby 环境包', iconName: 'ruby', color: '#cc342d',
    description: '动态面向对象编程语言',
    components: ['Ruby', 'Bundler', 'RubyGems'],
    checkCmd: 'ruby --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    getSteps: () => [
      { cmd: 'sudo apt-get update && sudo apt-get install -y ruby-full', desc: '安装 Ruby' },
      { cmd: 'sudo gem install bundler', desc: '安装 Bundler' }
    ],
    uninstallSteps: [{ cmd: 'sudo apt-get remove -y ruby-full && sudo apt-get autoremove -y', desc: '卸载 Ruby' }]
  },
  {
    name: 'Docker 环境包', iconName: 'docker', color: '#2496ed',
    description: '容器化应用平台',
    components: ['Docker Engine', 'Docker Compose', 'Docker CLI'],
    checkCmd: 'docker --version 2>/dev/null | grep -oP "\\d+\\.\\d+\\.\\d+"',
    getSteps: () => [
      { cmd: 'curl -fsSL https://get.docker.com -o /tmp/get-docker.sh', desc: '下载安装脚本' },
      { cmd: 'sudo sh /tmp/get-docker.sh', desc: '安装 Docker' },
      { cmd: 'sudo usermod -aG docker $USER', desc: '添加用户到 docker 组' },
      { cmd: 'sudo systemctl enable docker && sudo systemctl start docker', desc: '启动 Docker' }
    ],
    uninstallSteps: [
      { cmd: 'sudo systemctl stop docker', desc: '停止 Docker' },
      { cmd: 'sudo apt-get remove -y docker-ce docker-ce-cli containerd.io', desc: '卸载 Docker' }
    ]
  }
]

const packages = ref<PkgDef[]>([])
const installedPkgs = computed(() => packages.value.filter(p => p.installed))
const availablePkgs = computed(() => packages.value.filter(p => !p.installed))

watch(() => serverStore.currentServer, () => { if (serverStore.currentServer) checkAllStatus() })

async function checkAllStatus() {
  const server = serverStore.currentServer
  if (!server) return
  checking.value = true
  for (const pkg of packages.value) {
    try {
      const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', pkg.checkCmd])
      if (result.exit_code === 0 && (result.stdout || '').trim()) {
        pkg.installed = true
        pkg.version = (result.stdout || '').trim().split('\n')[0]
      } else {
        pkg.installed = false
        pkg.version = undefined
      }
    } catch { pkg.installed = false; pkg.version = undefined }
  }
  checking.value = false
}

function openInstall(pkg: PkgDef) {
  curPkg.value = pkg
  if (pkg.versions?.length) {
    selVersion.value = (pkg.versions.find(v => v.recommended) || pkg.versions[0]).value
  } else { selVersion.value = '' }
  showInstallDlg.value = true
}

function showDetail(pkg: PkgDef) { detailPkg.value = pkg; showDetailDlg.value = true }

async function doInstall() {
  const server = serverStore.currentServer
  if (!server || !curPkg.value) return
  const pkg = curPkg.value
  const steps = pkg.getSteps(selVersion.value)
  showInstallDlg.value = false
  installing.value = true

  const task = taskStore.createTask(`安装 ${pkg.name}`, 'env-install', server.id, steps)

  try {
    for (let i = 0; i < steps.length; i++) {
      const step = steps[i]
      taskStore.updateStep(task.id, i, 'running')
      taskStore.appendLog(task.id, `\n[${i + 1}/${steps.length}] ${step.desc}\n$ ${step.cmd}\n`)

      // 显示耗时计时
      const startTime = Date.now()
      const timer = setInterval(() => {
        const elapsed = Math.round((Date.now() - startTime) / 1000)
        taskStore.appendLog(task.id, `\r⏳ 执行中... ${elapsed}s`)
      }, 5000)

      try {
        const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', step.cmd], { timeout: 600 })
        clearInterval(timer)
        const elapsed = ((Date.now() - startTime) / 1000).toFixed(1)
        if (result.stdout) taskStore.appendLog(task.id, result.stdout + '\n')
        if (result.stderr) taskStore.appendLog(task.id, result.stderr + '\n')
        const ok = result.exit_code === 0
        taskStore.updateStep(task.id, i, ok ? 'success' : 'failed')
        taskStore.appendLog(task.id, ok ? `✓ 完成 (${elapsed}s)\n` : `⚠️ 退出码: ${result.exit_code} (${elapsed}s)\n`)
      } catch (e) {
        clearInterval(timer)
        taskStore.updateStep(task.id, i, 'failed')
        taskStore.appendLog(task.id, `❌ 错误: ${(e as Error).message}\n`)
      }
    }

    taskStore.appendLog(task.id, '\n🔍 验证安装...\n')
    const check = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', pkg.checkCmd])
    if (check.exit_code === 0 && (check.stdout || '').trim()) {
      taskStore.appendLog(task.id, check.stdout + '\n')
      taskStore.completeTask(task.id, true)
      ElMessage.success(`${pkg.name} 安装成功`)
    } else {
      taskStore.completeTask(task.id, false)
      ElMessage.error(`${pkg.name} 安装可能未成功`)
    }
    await checkAllStatus()
  } catch (e) {
    taskStore.appendLog(task.id, `\n❌ ${(e as Error).message}\n`)
    taskStore.completeTask(task.id, false)
    ElMessage.error('安装失败')
  } finally { installing.value = false }
}

async function uninstallPkg(pkg: PkgDef) {
  const server = serverStore.currentServer
  if (!server || !pkg.uninstallSteps) return
  try { await ElMessageBox.confirm(`确定卸载 ${pkg.name}？`, '确认卸载', { type: 'warning' }) } catch { return }

  const task = taskStore.createTask(`卸载 ${pkg.name}`, 'env-uninstall', server.id, pkg.uninstallSteps)
  try {
    for (let i = 0; i < pkg.uninstallSteps.length; i++) {
      const step = pkg.uninstallSteps[i]
      taskStore.updateStep(task.id, i, 'running')
      taskStore.appendLog(task.id, `\n[${i + 1}/${pkg.uninstallSteps.length}] ${step.desc}\n$ ${step.cmd}\n`)

      const startTime = Date.now()
      const timer = setInterval(() => {
        const elapsed = Math.round((Date.now() - startTime) / 1000)
        taskStore.appendLog(task.id, `\r⏳ 执行中... ${elapsed}s`)
      }, 5000)

      try {
        const result = await window.electronAPI.server.executeCommand(server.id, 'bash', ['-c', step.cmd], { timeout: 300 })
        clearInterval(timer)
        const elapsed = ((Date.now() - startTime) / 1000).toFixed(1)
        if (result.stdout) taskStore.appendLog(task.id, result.stdout + '\n')
        if (result.stderr) taskStore.appendLog(task.id, result.stderr + '\n')
        const ok = result.exit_code === 0
        taskStore.updateStep(task.id, i, ok ? 'success' : 'failed')
        taskStore.appendLog(task.id, ok ? `✓ 完成 (${elapsed}s)\n` : `⚠️ 退出码: ${result.exit_code} (${elapsed}s)\n`)
      } catch (e) {
        clearInterval(timer)
        taskStore.updateStep(task.id, i, 'failed')
        taskStore.appendLog(task.id, `❌ ${(e as Error).message}\n`)
      }
    }
    taskStore.completeTask(task.id, true)
    ElMessage.success(`${pkg.name} 已卸载`)
    await checkAllStatus()
  } catch { taskStore.completeTask(task.id, false); ElMessage.error('卸载失败') }
}

onMounted(() => {
  packages.value = pkgDefs.map(d => ({ ...d, installed: false, version: undefined }))
  if (serverStore.currentServer) checkAllStatus()
})
</script>

<style lang="scss" scoped>
@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes cardIn {
  from { opacity: 0; transform: scale(0.96) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.animate-in { animation: fadeSlideUp 0.4s ease both; }
.animate-card { animation: cardIn 0.35s ease both; }

.environment { max-width: 1200px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;
  h1 { font-size: 22px; font-weight: 700; margin: 0 0 4px; }
  .subtitle { color: var(--text-secondary); font-size: 13px; margin: 0; }
}

.no-server { padding: 60px 0; }

.section { margin-bottom: 28px; }
.section-header {
  display: flex; align-items: center; gap: 10px; margin-bottom: 14px;
  h2 { font-size: 15px; font-weight: 600; margin: 0; }
}

.pkg-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 14px;
}

.pkg-card {
  display: flex; align-items: flex-start; gap: 14px;
  padding: 18px; background: var(--bg-secondary);
  border: 1px solid var(--border-color); border-radius: 12px;
  transition: all 0.25s cubic-bezier(.4,0,.2,1);

  &:hover { border-color: var(--text-muted); transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
  &.installed { border-color: rgba(34,197,94,0.35); background: linear-gradient(135deg, var(--bg-secondary), rgba(34,197,94,0.05)); }
}

.pkg-icon {
  width: 46px; height: 46px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  &.large { width: 56px; height: 56px; }
}

.pkg-body { flex: 1; min-width: 0; }
.pkg-name { font-size: 15px; font-weight: 600; margin-bottom: 3px; }
.pkg-desc { font-size: 12px; color: var(--text-secondary); margin-bottom: 8px; }
.pkg-version {
  font-size: 12px; color: var(--success-color); display: flex; align-items: center; gap: 4px; margin-bottom: 8px;
  .el-icon { font-size: 14px; }
}

.pkg-components { display: flex; flex-wrap: wrap; gap: 4px; }
.comp-tag {
  background: rgba(255,255,255,0.06) !important;
  border-color: rgba(255,255,255,0.1) !important;
  color: var(--text-secondary) !important;
  font-size: 11px !important;
}

.pkg-actions {
  display: flex; flex-direction: column; gap: 4px; flex-shrink: 0;
}

.install-content {
  .install-preview { display: flex; align-items: center; gap: 14px; margin-bottom: 20px; }
  .install-info {
    display: flex; align-items: center; gap: 8px; padding: 12px;
    background: var(--bg-tertiary); border-radius: 8px; font-size: 13px; color: var(--text-secondary);
  }
}
</style>
