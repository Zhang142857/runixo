<template>
  <div class="containers-page">
    <div class="page-header">
      <h1>容器管理</h1>
      <div class="header-actions">
        <el-select v-model="selectedServer" placeholder="选择服务器" class="server-select">
          <el-option
            v-for="server in connectedServers"
            :key="server.id"
            :label="server.name"
            :value="server.id"
          />
        </el-select>
        <el-input
          v-model="searchQuery"
          placeholder="搜索容器..."
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button @click="refreshContainers" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建容器
        </el-button>
      </div>
    </div>

    <!-- 标签页切换 -->
    <el-tabs v-model="activeTab" class="main-tabs">
      <el-tab-pane label="容器" name="containers" />
      <el-tab-pane label="镜像" name="images" />
      <el-tab-pane label="网络" name="networks" />
      <el-tab-pane label="卷" name="volumes" />
    </el-tabs>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <!-- 容器标签页 -->
      <div v-show="activeTab === 'containers'">
        <!-- 快速过滤和批量操作 -->
        <div class="filter-bar">
          <div class="filter-buttons">
            <el-radio-group v-model="statusFilter" size="small">
              <el-radio-button value="all">全部</el-radio-button>
              <el-radio-button value="running">
                <span class="filter-dot running"></span> 运行中 ({{ runningCount }})
              </el-radio-button>
              <el-radio-button value="stopped">
                <span class="filter-dot stopped"></span> 已停止 ({{ stoppedCount }})
              </el-radio-button>
              <el-radio-button value="paused">
                <span class="filter-dot paused"></span> 已暂停 ({{ pausedCount }})
              </el-radio-button>
            </el-radio-group>
          </div>
          <div class="batch-actions" v-if="selectedContainers.length > 0">
            <span class="selected-count">已选择 {{ selectedContainers.length }} 个容器</span>
            <el-button-group size="small">
              <el-button type="success" @click="batchAction('start')">批量启动</el-button>
              <el-button type="warning" @click="batchAction('stop')">批量停止</el-button>
              <el-button @click="batchAction('restart')">批量重启</el-button>
              <el-button type="danger" @click="batchAction('delete')">批量删除</el-button>
            </el-button-group>
            <el-button size="small" @click="selectedContainers = []">取消选择</el-button>
          </div>
        </div>

        <!-- 统计信息 -->
        <div class="stats-row">
          <div class="stat-item running">
            <span class="stat-value">{{ runningCount }}</span>
            <span class="stat-label">运行中</span>
          </div>
          <div class="stat-item stopped">
            <span class="stat-value">{{ stoppedCount }}</span>
            <span class="stat-label">已停止</span>
          </div>
          <div class="stat-item paused">
            <span class="stat-value">{{ pausedCount }}</span>
            <span class="stat-label">已暂停</span>
          </div>
          <div class="stat-item total">
            <span class="stat-value">{{ containers.length }}</span>
            <span class="stat-label">总计</span>
          </div>
          <div class="stat-item resources">
            <div class="resource-summary">
              <span class="resource-label">CPU 总使用:</span>
              <el-progress :percentage="totalCpuUsage" :stroke-width="8" :show-text="false" />
              <span class="resource-value">{{ totalCpuUsage.toFixed(1) }}%</span>
            </div>
            <div class="resource-summary">
              <span class="resource-label">内存总使用:</span>
              <el-progress :percentage="totalMemoryPercent" :stroke-width="8" :show-text="false" />
              <span class="resource-value">{{ formatMemory(totalMemoryUsage) }}</span>
            </div>
          </div>
        </div>

        <!-- 容器列表 -->
        <el-table
          :data="filteredContainers"
          v-loading="loading"
          class="container-table"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column prop="name" label="名称" min-width="180">
            <template #default="{ row }">
              <div class="container-name" @click="showContainerDetail(row)">
                <span class="status-dot" :class="row.state"></span>
                <span class="name-text">{{ row.name }}</span>
                <el-tag v-if="row.health" :type="getHealthType(row.health)" size="small" class="health-tag">
                  {{ row.health }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="image" label="镜像" min-width="200">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.image }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="160" />
          <el-table-column label="资源使用" width="200">
            <template #default="{ row }">
              <div v-if="row.state === 'running'" class="resource-bars">
                <div class="resource-bar">
                  <span class="bar-label">CPU</span>
                  <el-progress
                    :percentage="row.cpu || 0"
                    :stroke-width="6"
                    :show-text="false"
                    :color="getResourceColor(row.cpu)"
                  />
                  <span class="bar-value">{{ row.cpu?.toFixed(1) }}%</span>
                </div>
                <div class="resource-bar">
                  <span class="bar-label">MEM</span>
                  <el-progress
                    :percentage="getMemoryPercent(row.memory)"
                    :stroke-width="6"
                    :show-text="false"
                    :color="getResourceColor(getMemoryPercent(row.memory))"
                  />
                  <span class="bar-value">{{ formatMemory(row.memory) }}</span>
                </div>
              </div>
              <span v-else class="text-secondary">-</span>
            </template>
          </el-table-column>
          <el-table-column label="端口" width="150">
            <template #default="{ row }">
              <span v-if="row.ports && row.ports.length > 0">
                {{ formatPorts(row.ports) }}
              </span>
              <span v-else class="text-secondary">-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <div class="action-buttons">
                <el-button-group size="small">
                  <el-button
                    v-if="row.state !== 'running'"
                    type="success"
                    @click="containerAction(row.id, 'start')"
                  >
                    启动
                  </el-button>
                  <el-button
                    v-if="row.state === 'running'"
                    type="warning"
                    @click="containerAction(row.id, 'stop')"
                  >
                    停止
                  </el-button>
                  <el-button
                    v-if="row.state === 'running'"
                    @click="containerAction(row.id, 'pause')"
                  >
                    暂停
                  </el-button>
                  <el-button
                    v-if="row.state === 'paused'"
                    type="success"
                    @click="containerAction(row.id, 'unpause')"
                  >
                    恢复
                  </el-button>
                  <el-button @click="containerAction(row.id, 'restart')">重启</el-button>
                </el-button-group>
                <el-dropdown trigger="click" @command="handleMoreAction($event, row)">
                  <el-button size="small">
                    更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="logs">
                        <el-icon><Document /></el-icon> 查看日志
                      </el-dropdown-item>
                      <el-dropdown-item command="terminal">
                        <el-icon><Monitor /></el-icon> 进入终端
                      </el-dropdown-item>
                      <el-dropdown-item command="inspect">
                        <el-icon><View /></el-icon> 查看详情
                      </el-dropdown-item>
                      <el-dropdown-item command="stats">
                        <el-icon><DataLine /></el-icon> 实时监控
                      </el-dropdown-item>
                      <el-dropdown-item command="env" divided>
                        <el-icon><Edit /></el-icon> 编辑环境变量
                      </el-dropdown-item>
                      <el-dropdown-item command="export">
                        <el-icon><Download /></el-icon> 导出配置
                      </el-dropdown-item>
                      <el-dropdown-item command="clone">
                        <el-icon><CopyDocument /></el-icon> 克隆容器
                      </el-dropdown-item>
                      <el-dropdown-item command="delete" divided>
                        <span class="danger-text"><el-icon><Delete /></el-icon> 删除容器</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 镜像标签页 -->
      <div v-show="activeTab === 'images'">
        <div class="images-header">
          <el-input
            v-model="imageSearchQuery"
            placeholder="搜索镜像..."
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <div class="images-actions">
            <el-button @click="showPullImageDialog = true">
              <el-icon><Download /></el-icon>
              拉取镜像
            </el-button>
            <el-button type="primary" @click="showBuildImageDialog = true">
              <el-icon><Box /></el-icon>
              构建镜像
            </el-button>
          </div>
        </div>

        <el-table :data="filteredImages" v-loading="imagesLoading" class="images-table">
          <el-table-column prop="repository" label="仓库" min-width="200" />
          <el-table-column prop="tag" label="标签" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ row.tag }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="id" label="镜像ID" width="140">
            <template #default="{ row }">
              <span class="image-id">{{ row.id.substring(0, 12) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="100">
            <template #default="{ row }">
              {{ formatImageSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="created" label="创建时间" width="180" />
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button type="primary" @click="createContainerFromImage(row)">
                  创建容器
                </el-button>
                <el-button @click="openRetagDialog(row)">
                  标记
                </el-button>
                <el-button @click="showImageHistory(row)">
                  历史
                </el-button>
                <el-button type="danger" @click="deleteImage(row)">
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 网络标签页 -->
      <div v-show="activeTab === 'networks'">
        <div class="networks-header">
          <el-input
            v-model="networkSearchQuery"
            placeholder="搜索网络..."
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" @click="showCreateNetworkDialog = true">
            <el-icon><Plus /></el-icon>
            创建网络
          </el-button>
        </div>

        <el-table :data="filteredNetworks" v-loading="networksLoading" class="networks-table">
          <el-table-column prop="name" label="网络名称" min-width="180">
            <template #default="{ row }">
              <div class="network-name">
                <el-icon><Connection /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="driver" label="驱动" width="120">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.driver }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="scope" label="范围" width="100" />
          <el-table-column prop="subnet" label="子网" width="160" />
          <el-table-column prop="gateway" label="网关" width="140" />
          <el-table-column label="容器数" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.containers || 0 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button @click="showNetworkDetail(row)">详情</el-button>
                <el-button
                  type="danger"
                  @click="deleteNetwork(row)"
                  :disabled="row.name === 'bridge' || row.name === 'host' || row.name === 'none'"
                >
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 卷标签页 -->
      <div v-show="activeTab === 'volumes'">
        <div class="volumes-header">
          <el-input
            v-model="volumeSearchQuery"
            placeholder="搜索卷..."
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" @click="showCreateVolumeDialog = true">
            <el-icon><Plus /></el-icon>
            创建卷
          </el-button>
        </div>

        <el-table :data="filteredVolumes" v-loading="volumesLoading" class="volumes-table">
          <el-table-column prop="name" label="卷名称" min-width="200">
            <template #default="{ row }">
              <div class="volume-name">
                <el-icon><Box /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="driver" label="驱动" width="100">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.driver }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="mountpoint" label="挂载点" min-width="250">
            <template #default="{ row }">
              <span class="mountpoint">{{ row.mountpoint }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="created" label="创建时间" width="180" />
          <el-table-column label="使用状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.inUse ? 'success' : 'info'" size="small">
                {{ row.inUse ? '使用中' : '未使用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button @click="showVolumeDetail(row)">详情</el-button>
                <el-button type="danger" @click="deleteVolume(row)" :disabled="row.inUse">
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <!-- 日志对话框 -->
    <el-dialog
      v-model="showLogDialog"
      :title="`容器日志 - ${currentContainer?.name}`"
      width="85%"
      top="3vh"
      class="log-dialog"
    >
      <div class="log-toolbar">
        <div class="log-toolbar-left">
          <el-select v-model="logLines" size="small" style="width: 120px">
            <el-option :value="100" label="最近100行" />
            <el-option :value="500" label="最近500行" />
            <el-option :value="1000" label="最近1000行" />
          </el-select>
          <el-select v-model="logLevelFilter" size="small" style="width: 100px" placeholder="日志级别">
            <el-option value="all" label="全部" />
            <el-option value="error" label="ERROR" />
            <el-option value="warn" label="WARN" />
            <el-option value="info" label="INFO" />
            <el-option value="debug" label="DEBUG" />
          </el-select>
          <el-input
            v-model="logSearchQuery"
            placeholder="搜索日志..."
            size="small"
            clearable
            style="width: 200px"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <div class="log-toolbar-right">
          <el-button size="small" @click="refreshLogs">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
          <el-button size="small" @click="scrollToTop">
            <el-icon><ArrowUp /></el-icon>
            顶部
          </el-button>
          <el-button size="small" @click="scrollToBottom">
            <el-icon><ArrowDown /></el-icon>
            底部
          </el-button>
          <el-button size="small" @click="downloadLogs">
            <el-icon><Download /></el-icon>
            下载
          </el-button>
        </div>
      </div>
      <div class="log-stats">
        <el-tag size="small" type="info">总行数: {{ logContent.split('\n').length }}</el-tag>
        <el-tag size="small" type="success" v-if="logSearchQuery || logLevelFilter !== 'all'">
          匹配: {{ filteredLogContent.split('\n').filter(l => l).length }} 行
        </el-tag>
      </div>
      <div class="log-container" ref="logContainer">
        <pre class="log-content" v-html="highlightLogs(filteredLogContent)"></pre>
      </div>
      <template #footer>
        <el-checkbox v-model="followLogs">实时跟踪</el-checkbox>
        <el-button @click="showLogDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 容器详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      :title="`容器详情 - ${currentContainer?.name}`"
      width="700px"
    >
      <el-descriptions :column="2" border v-if="currentContainer">
        <el-descriptions-item label="容器ID">
          <span class="mono-text">{{ currentContainer.id?.substring(0, 12) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStateType(currentContainer.state)">{{ currentContainer.state }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="镜像" :span="2">{{ currentContainer.image }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentContainer.created }}</el-descriptions-item>
        <el-descriptions-item label="启动时间">{{ currentContainer.startedAt || '-' }}</el-descriptions-item>
        <el-descriptions-item label="端口映射" :span="2">
          {{ currentContainer.ports?.length > 0 ? formatPorts(currentContainer.ports) : '无' }}
        </el-descriptions-item>
        <el-descriptions-item label="网络模式">{{ currentContainer.networkMode || 'bridge' }}</el-descriptions-item>
        <el-descriptions-item label="重启策略">{{ currentContainer.restartPolicy || 'no' }}</el-descriptions-item>
        <el-descriptions-item label="CPU 使用" v-if="currentContainer.state === 'running'">
          {{ currentContainer.cpu?.toFixed(2) }}%
        </el-descriptions-item>
        <el-descriptions-item label="内存使用" v-if="currentContainer.state === 'running'">
          {{ formatMemory(currentContainer.memory) }}
        </el-descriptions-item>
        <el-descriptions-item label="环境变量" :span="2">
          <div class="env-list" v-if="currentContainer.env && currentContainer.env.length > 0">
            <el-tag v-for="env in currentContainer.env.slice(0, 5)" :key="env" size="small" class="env-tag">
              {{ env }}
            </el-tag>
            <span v-if="currentContainer.env && currentContainer.env.length > 5" class="more-env">
              +{{ currentContainer.env.length - 5 }} 更多
            </span>
          </div>
          <span v-else>无</span>
        </el-descriptions-item>
        <el-descriptions-item label="挂载卷" :span="2">
          <div v-if="currentContainer.mounts && currentContainer.mounts.length > 0">
            <div v-for="mount in currentContainer.mounts" :key="mount.source" class="mount-item">
              {{ mount.source }} → {{ mount.destination }}
            </div>
          </div>
          <span v-else>无</span>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 创建容器对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建容器"
      width="600px"
    >
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="容器名称" required>
          <el-input v-model="createForm.name" placeholder="输入容器名称" />
        </el-form-item>
        <el-form-item label="镜像" required>
          <el-select v-model="createForm.image" filterable allow-create placeholder="选择或输入镜像">
            <el-option
              v-for="img in images"
              :key="img.id"
              :label="`${img.repository}:${img.tag}`"
              :value="`${img.repository}:${img.tag}`"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="端口映射">
          <div v-for="(port, index) in createForm.ports" :key="index" class="port-row">
            <el-input v-model="port.host" placeholder="主机端口" style="width: 100px" />
            <span class="port-separator">:</span>
            <el-input v-model="port.container" placeholder="容器端口" style="width: 100px" />
            <el-button text type="danger" @click="removePort(index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <el-button text type="primary" @click="addPort">
            <el-icon><Plus /></el-icon> 添加端口
          </el-button>
        </el-form-item>
        <el-form-item label="环境变量">
          <div v-for="(env, index) in createForm.envs" :key="index" class="env-row">
            <el-input v-model="env.key" placeholder="变量名" style="width: 120px" />
            <span class="env-separator">=</span>
            <el-input v-model="env.value" placeholder="值" style="width: 180px" />
            <el-button text type="danger" @click="removeEnv(index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <el-button text type="primary" @click="addEnv">
            <el-icon><Plus /></el-icon> 添加变量
          </el-button>
        </el-form-item>
        <el-form-item label="重启策略">
          <el-select v-model="createForm.restartPolicy">
            <el-option value="no" label="不重启" />
            <el-option value="always" label="总是重启" />
            <el-option value="on-failure" label="失败时重启" />
            <el-option value="unless-stopped" label="除非手动停止" />
          </el-select>
        </el-form-item>
        <el-form-item label="启动命令">
          <el-input v-model="createForm.command" placeholder="可选，覆盖默认命令" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createContainer">创建</el-button>
      </template>
    </el-dialog>

    <!-- 拉取镜像对话框 -->
    <el-dialog
      v-model="showPullImageDialog"
      title="拉取镜像"
      width="500px"
    >
      <el-form label-width="80px">
        <el-form-item label="镜像名称" required>
          <el-input v-model="pullImageName" placeholder="例如: nginx:latest" />
        </el-form-item>
      </el-form>
      <div v-if="pullProgress" class="pull-progress">
        <el-progress :percentage="pullProgress" :status="pullStatus" />
        <p class="pull-message">{{ pullMessage }}</p>
      </div>
      <template #footer>
        <el-button @click="showPullImageDialog = false">取消</el-button>
        <el-button type="primary" @click="pullImage" :loading="isPulling">拉取</el-button>
      </template>
    </el-dialog>

    <!-- 构建镜像对话框 -->
    <el-dialog
      v-model="showBuildImageDialog"
      title="构建镜像"
      width="500px"
    >
      <el-form :model="buildImageForm" label-width="100px">
        <el-form-item label="镜像名称" required>
          <el-input v-model="buildImageForm.name" placeholder="例如: my-app" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="buildImageForm.tag" placeholder="latest" />
        </el-form-item>
        <el-form-item label="Dockerfile">
          <el-input
            v-model="buildImageForm.dockerfile"
            type="textarea"
            :rows="6"
            placeholder="FROM node:18-alpine&#10;WORKDIR /app&#10;COPY . .&#10;RUN npm install&#10;CMD [&quot;npm&quot;, &quot;start&quot;]"
          />
        </el-form-item>
        <el-form-item label="构建上下文">
          <el-input v-model="buildImageForm.context" placeholder="." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBuildImageDialog = false">取消</el-button>
        <el-button type="primary" @click="buildImage">构建</el-button>
      </template>
    </el-dialog>

    <!-- 创建网络对话框 -->
    <el-dialog
      v-model="showCreateNetworkDialog"
      title="创建网络"
      width="500px"
    >
      <el-form :model="createNetworkForm" label-width="80px">
        <el-form-item label="网络名称" required>
          <el-input v-model="createNetworkForm.name" placeholder="输入网络名称" />
        </el-form-item>
        <el-form-item label="驱动">
          <el-select v-model="createNetworkForm.driver" style="width: 100%">
            <el-option value="bridge" label="bridge" />
            <el-option value="host" label="host" />
            <el-option value="overlay" label="overlay" />
            <el-option value="macvlan" label="macvlan" />
          </el-select>
        </el-form-item>
        <el-form-item label="子网">
          <el-input v-model="createNetworkForm.subnet" placeholder="例如: 172.20.0.0/16" />
        </el-form-item>
        <el-form-item label="网关">
          <el-input v-model="createNetworkForm.gateway" placeholder="例如: 172.20.0.1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateNetworkDialog = false">取消</el-button>
        <el-button type="primary" @click="createNetwork">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建卷对话框 -->
    <el-dialog
      v-model="showCreateVolumeDialog"
      title="创建卷"
      width="450px"
    >
      <el-form :model="createVolumeForm" label-width="80px">
        <el-form-item label="卷名称" required>
          <el-input v-model="createVolumeForm.name" placeholder="输入卷名称" />
        </el-form-item>
        <el-form-item label="驱动">
          <el-select v-model="createVolumeForm.driver" style="width: 100%">
            <el-option value="local" label="local" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateVolumeDialog = false">取消</el-button>
        <el-button type="primary" @click="createVolume">创建</el-button>
      </template>
    </el-dialog>

    <!-- 容器实时监控对话框 -->
    <el-dialog
      v-model="showStatsDialog"
      :title="`实时监控 - ${currentContainer?.name}`"
      width="700px"
    >
      <div class="stats-container" v-if="currentContainer">
        <div class="stats-grid">
          <div class="stats-card">
            <div class="stats-card-header">
              <el-icon><Cpu /></el-icon>
              <span>CPU 使用率</span>
            </div>
            <div class="stats-card-value">{{ currentContainer.cpu?.toFixed(2) }}%</div>
            <div class="stats-chart">
              <div class="mini-chart">
                <div
                  v-for="(val, idx) in containerStats.cpu"
                  :key="idx"
                  class="chart-bar"
                  :style="{ height: `${Math.min(100, val)}%` }"
                ></div>
              </div>
            </div>
          </div>
          <div class="stats-card">
            <div class="stats-card-header">
              <el-icon><Box /></el-icon>
              <span>内存使用</span>
            </div>
            <div class="stats-card-value">{{ formatMemory(currentContainer.memory) }}</div>
            <div class="stats-chart">
              <div class="mini-chart">
                <div
                  v-for="(val, idx) in containerStats.memory"
                  :key="idx"
                  class="chart-bar memory"
                  :style="{ height: `${Math.min(100, val / 10000000)}%` }"
                ></div>
              </div>
            </div>
          </div>
          <div class="stats-card">
            <div class="stats-card-header">
              <el-icon><Download /></el-icon>
              <span>网络接收</span>
            </div>
            <div class="stats-card-value">{{ (containerStats.networkRx[containerStats.networkRx.length - 1] / 1024).toFixed(1) }} KB/s</div>
          </div>
          <div class="stats-card">
            <div class="stats-card-header">
              <el-icon><ArrowDown /></el-icon>
              <span>网络发送</span>
            </div>
            <div class="stats-card-value">{{ (containerStats.networkTx[containerStats.networkTx.length - 1] / 1024).toFixed(1) }} KB/s</div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showStatsDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 镜像历史对话框 -->
    <el-dialog
      v-model="showImageHistoryDialog"
      :title="`镜像历史 - ${currentImage?.repository}:${currentImage?.tag}`"
      width="800px"
    >
      <el-table :data="imageHistory" class="history-table">
        <el-table-column prop="id" label="层ID" width="140">
          <template #default="{ row }">
            <span class="layer-id">{{ row.id.substring(0, 12) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created" label="创建时间" width="180" />
        <el-table-column prop="createdBy" label="创建命令" min-width="300">
          <template #default="{ row }">
            <span class="created-by">{{ row.createdBy }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="100">
          <template #default="{ row }">
            {{ row.size > 0 ? formatImageSize(row.size) : '-' }}
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="showImageHistoryDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 环境变量编辑对话框 -->
    <el-dialog
      v-model="showEnvEditDialog"
      :title="`编辑环境变量 - ${currentContainer?.name}`"
      width="600px"
    >
      <div class="env-edit-toolbar">
        <el-button type="primary" size="small" @click="addEditEnv">
          <el-icon><Plus /></el-icon>
          添加变量
        </el-button>
        <span class="env-count">共 {{ editingEnvs.length }} 个变量</span>
      </div>
      <div class="env-edit-list">
        <div v-for="(env, index) in editingEnvs" :key="index" class="env-edit-row" :class="{ 'new-env': env.isNew }">
          <el-input
            v-model="env.key"
            placeholder="变量名"
            size="small"
            style="width: 180px"
            :disabled="!env.isNew"
          />
          <span class="env-equals">=</span>
          <el-input
            v-model="env.value"
            placeholder="值"
            size="small"
            style="flex: 1"
            :type="env.key.toLowerCase().includes('password') || env.key.toLowerCase().includes('secret') ? 'password' : 'text'"
            show-password
          />
          <el-button
            type="danger"
            size="small"
            text
            @click="removeEditEnv(index)"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <div v-if="editingEnvs.length === 0" class="no-envs">
          暂无环境变量
        </div>
      </div>
      <template #footer>
        <el-button @click="showEnvEditDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEnvChanges">保存更改</el-button>
      </template>
    </el-dialog>

    <!-- 容器终端对话框 -->
    <el-dialog
      v-model="showTerminalDialog"
      :title="`容器终端 - ${currentContainer?.name}`"
      width="80%"
      top="5vh"
      class="terminal-dialog"
      @opened="focusTerminalInput"
    >
      <div class="terminal-container" ref="terminalContainer">
        <div class="terminal-output">
          <div v-for="(line, index) in terminalOutput" :key="index" class="terminal-line" v-html="line"></div>
        </div>
        <div class="terminal-input-line">
          <span class="terminal-prompt">{{ currentContainer?.name }}:~# </span>
          <input
            ref="terminalInputRef"
            v-model="terminalInput"
            class="terminal-input"
            @keyup.enter="executeTerminalCommand"
            @keyup.up="historyUp"
            @keyup.down="historyDown"
            placeholder="输入命令..."
          />
        </div>
      </div>
      <template #footer>
        <el-button size="small" @click="clearTerminal">清屏</el-button>
        <el-button @click="showTerminalDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 镜像重新标记对话框 -->
    <el-dialog
      v-model="showRetagDialog"
      :title="`重新标记镜像 - ${currentImage?.repository}:${currentImage?.tag}`"
      width="450px"
    >
      <el-form :model="retagForm" label-width="100px">
        <el-form-item label="当前标签">
          <el-tag size="large">{{ currentImage?.repository }}:{{ currentImage?.tag }}</el-tag>
        </el-form-item>
        <el-form-item label="新仓库名" required>
          <el-input v-model="retagForm.newRepository" placeholder="例如: my-app" />
        </el-form-item>
        <el-form-item label="新标签" required>
          <el-input v-model="retagForm.newTag" placeholder="例如: v1.0.0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRetagDialog = false">取消</el-button>
        <el-button type="primary" @click="retagImage">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, onMounted, nextTick } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh, Search, Plus, ArrowDown, Delete, Download,
  Document, Monitor, View, DataLine, CopyDocument, Connection, Box, Cpu,
  Edit, ArrowUp
} from '@element-plus/icons-vue'

interface Port {
  hostPort: string
  containerPort: string
}

interface Container {
  id: string
  name: string
  image: string
  state: string
  status: string
  ports: Port[]
  cpu?: number
  memory?: number
  created?: string
  startedAt?: string
  networkMode?: string
  restartPolicy?: string
  env?: string[]
  mounts?: Array<{ source: string; destination: string }>
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
  subnet: string
  gateway: string
  containers: number
  created: string
}

interface Volume {
  name: string
  driver: string
  mountpoint: string
  created: string
  inUse: boolean
  size?: number
}

const serverStore = useServerStore()

const selectedServer = ref<string | null>(serverStore.currentServerId)
const containers = ref<Container[]>([])
const images = ref<Image[]>([])
const networks = ref<Network[]>([])
const volumes = ref<Volume[]>([])
const loading = ref(false)
const imagesLoading = ref(false)
const networksLoading = ref(false)
const volumesLoading = ref(false)
const activeTab = ref('containers')
const searchQuery = ref('')
const imageSearchQuery = ref('')
const networkSearchQuery = ref('')
const volumeSearchQuery = ref('')

// 过滤和批量选择
const statusFilter = ref('all')
const selectedContainers = ref<string[]>([])

// 对话框状态
const showLogDialog = ref(false)
const showDetailDialog = ref(false)
const showCreateDialog = ref(false)
const showPullImageDialog = ref(false)
const showBuildImageDialog = ref(false)
const showCreateNetworkDialog = ref(false)
const showCreateVolumeDialog = ref(false)
const showStatsDialog = ref(false)
const showImageHistoryDialog = ref(false)

const currentContainer = ref<Container | null>(null)
const logContent = ref('')
const followLogs = ref(true)
const logLines = ref(100)
const logContainer = ref<HTMLElement | null>(null)

// 日志搜索和过滤
const logSearchQuery = ref('')
const logLevelFilter = ref('all')
const filteredLogContent = computed(() => {
  let logs = logContent.value

  // 按日志级别过滤
  if (logLevelFilter.value !== 'all') {
    const lines = logs.split('\n')
    logs = lines.filter(line => {
      const level = logLevelFilter.value.toUpperCase()
      return line.includes(`[${level}]`) || line.includes(level)
    }).join('\n')
  }

  // 按搜索词过滤
  if (logSearchQuery.value) {
    const query = logSearchQuery.value.toLowerCase()
    const lines = logs.split('\n')
    logs = lines.filter(line => line.toLowerCase().includes(query)).join('\n')
  }

  return logs
})

// 环境变量编辑
const showEnvEditDialog = ref(false)
const editingEnvs = ref<Array<{ key: string; value: string; isNew?: boolean }>>([])

// 容器终端
const showTerminalDialog = ref(false)
const terminalOutput = ref<string[]>([])
const terminalInput = ref('')
const terminalContainer = ref<HTMLElement | null>(null)

// 镜像标签管理
const showRetagDialog = ref(false)
const retagForm = ref({
  newRepository: '',
  newTag: ''
})

// 拉取镜像
const pullImageName = ref('')
const isPulling = ref(false)
const pullProgress = ref(0)
const pullStatus = ref<'' | 'success' | 'exception'>('')
const pullMessage = ref('')

// 创建容器表单
const createForm = ref({
  name: '',
  image: '',
  ports: [] as Array<{ host: string; container: string }>,
  envs: [] as Array<{ key: string; value: string }>,
  restartPolicy: 'no',
  command: ''
})

// 构建镜像表单
const buildImageForm = ref({
  name: '',
  tag: 'latest',
  dockerfile: '',
  context: '.'
})

// 创建网络表单
const createNetworkForm = ref({
  name: '',
  driver: 'bridge',
  subnet: '',
  gateway: ''
})

// 创建卷表单
const createVolumeForm = ref({
  name: '',
  driver: 'local'
})

// 容器实时监控数据
const containerStats = ref({
  cpu: [] as number[],
  memory: [] as number[],
  networkRx: [] as number[],
  networkTx: [] as number[],
  timestamps: [] as string[]
})

// 镜像历史
const imageHistory = ref<Array<{
  id: string
  created: string
  createdBy: string
  size: number
}>>([])
const currentImage = ref<Image | null>(null)

// 模拟数据更新定时器
let metricsInterval: ReturnType<typeof setInterval> | null = null

const connectedServers = computed(() => serverStore.connectedServers)
const runningCount = computed(() => containers.value.filter(c => c.state === 'running').length)
const stoppedCount = computed(() => containers.value.filter(c => c.state === 'exited').length)
const pausedCount = computed(() => containers.value.filter(c => c.state === 'paused').length)

const filteredContainers = computed(() => {
  let result = containers.value

  // 状态过滤
  if (statusFilter.value !== 'all') {
    result = result.filter(c => {
      if (statusFilter.value === 'running') return c.state === 'running'
      if (statusFilter.value === 'stopped') return c.state === 'exited'
      if (statusFilter.value === 'paused') return c.state === 'paused'
      return true
    })
  }

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(c =>
      c.name.toLowerCase().includes(query) ||
      c.image.toLowerCase().includes(query)
    )
  }

  return result
})

const filteredImages = computed(() => {
  if (!imageSearchQuery.value) return images.value
  const query = imageSearchQuery.value.toLowerCase()
  return images.value.filter(img =>
    img.repository.toLowerCase().includes(query) ||
    img.tag.toLowerCase().includes(query)
  )
})

const filteredNetworks = computed(() => {
  if (!networkSearchQuery.value) return networks.value
  const query = networkSearchQuery.value.toLowerCase()
  return networks.value.filter(n =>
    n.name.toLowerCase().includes(query) ||
    n.driver.toLowerCase().includes(query)
  )
})

const filteredVolumes = computed(() => {
  if (!volumeSearchQuery.value) return volumes.value
  const query = volumeSearchQuery.value.toLowerCase()
  return volumes.value.filter(v =>
    v.name.toLowerCase().includes(query)
  )
})

// 资源使用统计
const totalCpuUsage = computed(() => {
  const runningContainers = containers.value.filter(c => c.state === 'running')
  if (runningContainers.length === 0) return 0
  return Math.min(100, runningContainers.reduce((sum, c) => sum + (c.cpu || 0), 0))
})

const totalMemoryUsage = computed(() => {
  const runningContainers = containers.value.filter(c => c.state === 'running')
  return runningContainers.reduce((sum, c) => sum + (c.memory || 0), 0)
})

const totalMemoryPercent = computed(() => {
  // 假设总内存为 16GB
  const totalMemory = 16 * 1024 * 1024 * 1024
  return Math.min(100, (totalMemoryUsage.value / totalMemory) * 100)
})

function formatPorts(ports: Port[]): string {
  return ports.map(p => `${p.hostPort}:${p.containerPort}`).join(', ')
}

function formatMemory(bytes?: number): string {
  if (!bytes) return '-'
  const mb = bytes / (1024 * 1024)
  if (mb < 1024) return `${mb.toFixed(1)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

function formatImageSize(bytes: number): string {
  const mb = bytes / (1024 * 1024)
  if (mb < 1024) return `${mb.toFixed(0)} MB`
  return `${(mb / 1024).toFixed(2)} GB`
}

function getStateType(state: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (state) {
    case 'running': return 'success'
    case 'paused': return 'warning'
    case 'exited': return 'danger'
    default: return 'info'
  }
}

function getHealthType(health: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (health) {
    case 'healthy': return 'success'
    case 'unhealthy': return 'danger'
    case 'starting': return 'warning'
    default: return 'info'
  }
}

function getResourceColor(percent: number): string {
  if (percent > 80) return '#f56c6c'
  if (percent > 60) return '#e6a23c'
  return '#67c23a'
}

function getMemoryPercent(bytes?: number): number {
  if (!bytes) return 0
  // 假设单个容器最大内存为 4GB
  const maxMemory = 4 * 1024 * 1024 * 1024
  return Math.min(100, (bytes / maxMemory) * 100)
}

// 批量选择处理
function handleSelectionChange(selection: Container[]) {
  selectedContainers.value = selection.map(c => c.id)
}

// 批量操作
async function batchAction(action: string) {
  if (selectedContainers.value.length === 0) return

  const actionNames: Record<string, string> = {
    start: '启动',
    stop: '停止',
    restart: '重启',
    delete: '删除'
  }

  if (action === 'delete') {
    try {
      await ElMessageBox.confirm(
        `确定要删除选中的 ${selectedContainers.value.length} 个容器吗？此操作不可恢复。`,
        '确认批量删除',
        { type: 'warning' }
      )
    } catch {
      return
    }
  }

  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 800))

    for (const containerId of selectedContainers.value) {
      const container = containers.value.find(c => c.id === containerId)
      if (container) {
        switch (action) {
          case 'start':
            container.state = 'running'
            container.status = 'Up 1 second'
            container.cpu = 5
            container.memory = 100000000
            break
          case 'stop':
            container.state = 'exited'
            container.status = 'Exited (0) just now'
            container.cpu = undefined
            container.memory = undefined
            break
          case 'restart':
            container.status = 'Up 1 second'
            break
          case 'delete':
            containers.value = containers.value.filter(c => c.id !== containerId)
            break
        }
      }
    }

    ElMessage.success(`已${actionNames[action]} ${selectedContainers.value.length} 个容器`)
    selectedContainers.value = []
  } catch (error) {
    ElMessage.error(`批量操作失败: ${(error as Error).message}`)
  } finally {
    loading.value = false
  }
}

// 初始化模拟容器数据
function initSimulatedData() {
  containers.value = [
    {
      id: 'a1b2c3d4e5f6',
      name: 'nginx-proxy',
      image: 'nginx:latest',
      state: 'running',
      status: 'Up 5 days',
      ports: [{ hostPort: '80', containerPort: '80' }, { hostPort: '443', containerPort: '443' }],
      cpu: 2.5,
      memory: 134217728,
      created: '2024-01-15 10:30:00',
      startedAt: '2024-01-20 08:00:00',
      networkMode: 'bridge',
      restartPolicy: 'always',
      env: ['NGINX_HOST=localhost', 'NGINX_PORT=80'],
      mounts: [{ source: '/etc/nginx/conf.d', destination: '/etc/nginx/conf.d' }]
    },
    {
      id: 'b2c3d4e5f6a7',
      name: 'mysql-db',
      image: 'mysql:8.0',
      state: 'running',
      status: 'Up 5 days',
      ports: [{ hostPort: '3306', containerPort: '3306' }],
      cpu: 8.3,
      memory: 536870912,
      created: '2024-01-15 10:35:00',
      startedAt: '2024-01-20 08:00:00',
      networkMode: 'bridge',
      restartPolicy: 'always',
      env: ['MYSQL_ROOT_PASSWORD=***', 'MYSQL_DATABASE=app'],
      mounts: [{ source: '/var/lib/mysql', destination: '/var/lib/mysql' }]
    },
    {
      id: 'c3d4e5f6a7b8',
      name: 'redis-cache',
      image: 'redis:7-alpine',
      state: 'running',
      status: 'Up 5 days',
      ports: [{ hostPort: '6379', containerPort: '6379' }],
      cpu: 0.8,
      memory: 67108864,
      created: '2024-01-15 10:40:00',
      startedAt: '2024-01-20 08:00:00',
      networkMode: 'bridge',
      restartPolicy: 'always',
      env: [],
      mounts: []
    },
    {
      id: 'd4e5f6a7b8c9',
      name: 'app-backend',
      image: 'node:18-alpine',
      state: 'running',
      status: 'Up 3 days',
      ports: [{ hostPort: '3000', containerPort: '3000' }],
      cpu: 15.2,
      memory: 268435456,
      created: '2024-01-17 14:20:00',
      startedAt: '2024-01-22 09:30:00',
      networkMode: 'bridge',
      restartPolicy: 'on-failure',
      env: ['NODE_ENV=production', 'PORT=3000', 'DB_HOST=mysql-db'],
      mounts: [{ source: '/app', destination: '/app' }]
    },
    {
      id: 'e5f6a7b8c9d0',
      name: 'postgres-test',
      image: 'postgres:15',
      state: 'exited',
      status: 'Exited (0) 2 days ago',
      ports: [{ hostPort: '5432', containerPort: '5432' }],
      created: '2024-01-10 16:00:00',
      networkMode: 'bridge',
      restartPolicy: 'no',
      env: ['POSTGRES_PASSWORD=***'],
      mounts: []
    },
    {
      id: 'f6a7b8c9d0e1',
      name: 'mongodb-dev',
      image: 'mongo:6',
      state: 'paused',
      status: 'Up 1 day (Paused)',
      ports: [{ hostPort: '27017', containerPort: '27017' }],
      cpu: 0,
      memory: 157286400,
      created: '2024-01-18 11:00:00',
      startedAt: '2024-01-24 10:00:00',
      networkMode: 'bridge',
      restartPolicy: 'unless-stopped',
      env: ['MONGO_INITDB_ROOT_USERNAME=admin'],
      mounts: [{ source: '/data/db', destination: '/data/db' }]
    }
  ]

  images.value = [
    { id: 'sha256:abc123', repository: 'nginx', tag: 'latest', size: 142000000, created: '2024-01-10' },
    { id: 'sha256:def456', repository: 'mysql', tag: '8.0', size: 565000000, created: '2024-01-08' },
    { id: 'sha256:ghi789', repository: 'redis', tag: '7-alpine', size: 32000000, created: '2024-01-12' },
    { id: 'sha256:jkl012', repository: 'node', tag: '18-alpine', size: 178000000, created: '2024-01-15' },
    { id: 'sha256:mno345', repository: 'postgres', tag: '15', size: 412000000, created: '2024-01-05' },
    { id: 'sha256:pqr678', repository: 'mongo', tag: '6', size: 698000000, created: '2024-01-11' },
    { id: 'sha256:stu901', repository: 'python', tag: '3.11-slim', size: 125000000, created: '2024-01-14' },
    { id: 'sha256:vwx234', repository: 'golang', tag: '1.21-alpine', size: 258000000, created: '2024-01-09' }
  ]

  // 初始化网络数据
  networks.value = [
    { id: 'net1', name: 'bridge', driver: 'bridge', scope: 'local', subnet: '172.17.0.0/16', gateway: '172.17.0.1', containers: 4, created: '2024-01-01' },
    { id: 'net2', name: 'host', driver: 'host', scope: 'local', subnet: '', gateway: '', containers: 0, created: '2024-01-01' },
    { id: 'net3', name: 'none', driver: 'null', scope: 'local', subnet: '', gateway: '', containers: 0, created: '2024-01-01' },
    { id: 'net4', name: 'app-network', driver: 'bridge', scope: 'local', subnet: '172.18.0.0/16', gateway: '172.18.0.1', containers: 3, created: '2024-01-15' },
    { id: 'net5', name: 'db-network', driver: 'bridge', scope: 'local', subnet: '172.19.0.0/16', gateway: '172.19.0.1', containers: 2, created: '2024-01-15' }
  ]

  // 初始化卷数据
  volumes.value = [
    { name: 'mysql-data', driver: 'local', mountpoint: '/var/lib/docker/volumes/mysql-data/_data', created: '2024-01-15', inUse: true },
    { name: 'redis-data', driver: 'local', mountpoint: '/var/lib/docker/volumes/redis-data/_data', created: '2024-01-15', inUse: true },
    { name: 'nginx-config', driver: 'local', mountpoint: '/var/lib/docker/volumes/nginx-config/_data', created: '2024-01-15', inUse: true },
    { name: 'app-logs', driver: 'local', mountpoint: '/var/lib/docker/volumes/app-logs/_data', created: '2024-01-17', inUse: true },
    { name: 'backup-data', driver: 'local', mountpoint: '/var/lib/docker/volumes/backup-data/_data', created: '2024-01-10', inUse: false },
    { name: 'temp-storage', driver: 'local', mountpoint: '/var/lib/docker/volumes/temp-storage/_data', created: '2024-01-20', inUse: false }
  ]
}

// 更新容器资源使用
function updateContainerMetrics() {
  containers.value.forEach(container => {
    if (container.state === 'running') {
      container.cpu = Math.max(0.1, Math.min(100, (container.cpu || 5) + (Math.random() - 0.5) * 3))
      container.memory = Math.max(50000000, (container.memory || 100000000) * (1 + (Math.random() - 0.5) * 0.1))
    }
  })
}

watch(selectedServer, (newVal) => {
  if (newVal) {
    refreshContainers()
  }
}, { immediate: true })

watch(activeTab, (newVal) => {
  if (newVal === 'images' && images.value.length === 0) {
    refreshImages()
  }
  if (newVal === 'networks' && networks.value.length === 0) {
    loadNetworks()
  }
  if (newVal === 'volumes' && volumes.value.length === 0) {
    loadVolumes()
  }
})

onMounted(() => {
  metricsInterval = setInterval(updateContainerMetrics, 3000)
})

onUnmounted(() => {
  if (metricsInterval) clearInterval(metricsInterval)
})

async function refreshContainers() {
  if (!selectedServer.value) return

  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    initSimulatedData()
    ElMessage.success('容器列表已刷新')
  } catch (error) {
    ElMessage.error(`获取容器列表失败: ${(error as Error).message}`)
  } finally {
    loading.value = false
  }
}

async function refreshImages() {
  imagesLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 300))
    // 数据已在 initSimulatedData 中初始化
  } finally {
    imagesLoading.value = false
  }
}

async function containerAction(containerId: string, action: string) {
  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    const container = containers.value.find(c => c.id === containerId)
    if (container) {
      switch (action) {
        case 'start':
          container.state = 'running'
          container.status = 'Up 1 second'
          container.cpu = 5
          container.memory = 100000000
          break
        case 'stop':
          container.state = 'exited'
          container.status = 'Exited (0) just now'
          container.cpu = undefined
          container.memory = undefined
          break
        case 'pause':
          container.state = 'paused'
          container.status = container.status.replace('Up', 'Up') + ' (Paused)'
          break
        case 'unpause':
          container.state = 'running'
          container.status = container.status.replace(' (Paused)', '')
          break
        case 'restart':
          container.status = 'Up 1 second'
          break
      }
    }

    ElMessage.success(`${action} 操作成功`)
  } catch (error) {
    ElMessage.error(`操作失败: ${(error as Error).message}`)
  }
}

function handleMoreAction(command: string, container: Container) {
  currentContainer.value = container
  switch (command) {
    case 'logs':
      showLogs(container)
      break
    case 'terminal':
      openTerminalDialog(container)
      break
    case 'inspect':
      showContainerDetail(container)
      break
    case 'stats':
      showContainerStats(container)
      break
    case 'export':
      exportContainerConfig(container)
      break
    case 'clone':
      cloneContainer(container)
      break
    case 'env':
      openEnvEditDialog(container)
      break
    case 'delete':
      deleteContainer(container)
      break
  }
}

function showContainerDetail(container: Container) {
  currentContainer.value = container
  showDetailDialog.value = true
}

function showLogs(container: Container) {
  currentContainer.value = container
  logContent.value = generateSimulatedLogs(container)
  showLogDialog.value = true
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

function generateSimulatedLogs(container: Container): string {
  const logs: string[] = []
  const now = new Date()

  for (let i = 0; i < logLines.value; i++) {
    const time = new Date(now.getTime() - (logLines.value - i) * 60000)
    const timestamp = time.toISOString()

    if (container.name.includes('nginx')) {
      const ips = ['192.168.1.100', '10.0.0.50', '172.16.0.25']
      const paths = ['/', '/api/users', '/api/data', '/static/js/app.js', '/health']
      const codes = ['200', '200', '200', '304', '404', '500']
      logs.push(`${timestamp} ${ips[i % 3]} - - "GET ${paths[i % 5]} HTTP/1.1" ${codes[i % 6]} ${Math.floor(Math.random() * 5000)}`)
    } else if (container.name.includes('mysql')) {
      const queries = ['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'COMMIT']
      logs.push(`${timestamp} [Note] ${queries[i % 5]} query executed in ${Math.floor(Math.random() * 100)}ms`)
    } else if (container.name.includes('redis')) {
      const commands = ['GET', 'SET', 'HGET', 'LPUSH', 'EXPIRE']
      logs.push(`${timestamp} ${commands[i % 5]} key:${Math.floor(Math.random() * 1000)}`)
    } else {
      const levels = ['INFO', 'DEBUG', 'WARN', 'ERROR']
      const messages = ['Request processed', 'Connection established', 'Cache hit', 'Task completed', 'Health check passed']
      logs.push(`${timestamp} [${levels[i % 4]}] ${messages[i % 5]}`)
    }
  }

  return logs.join('\n')
}

function refreshLogs() {
  if (currentContainer.value) {
    logContent.value = generateSimulatedLogs(currentContainer.value)
    ElMessage.success('日志已刷新')
  }
}

function downloadLogs() {
  const blob = new Blob([logContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${currentContainer.value?.name}-logs.txt`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('日志已下载')
}

// 日志高亮显示
function highlightLogs(content: string): string {
  return content.split('\n').map(line => {
    // 高亮 ERROR
    if (line.includes('[ERROR]') || line.includes('ERROR') || line.includes('error')) {
      return `<span class="log-error">${escapeHtml(line)}</span>`
    }
    // 高亮 WARN
    if (line.includes('[WARN]') || line.includes('WARN') || line.includes('warning')) {
      return `<span class="log-warn">${escapeHtml(line)}</span>`
    }
    // 高亮 INFO
    if (line.includes('[INFO]') || line.includes('INFO')) {
      return `<span class="log-info">${escapeHtml(line)}</span>`
    }
    // 高亮 DEBUG
    if (line.includes('[DEBUG]') || line.includes('DEBUG')) {
      return `<span class="log-debug">${escapeHtml(line)}</span>`
    }
    // 高亮搜索词
    if (logSearchQuery.value) {
      const regex = new RegExp(`(${escapeRegExp(logSearchQuery.value)})`, 'gi')
      return escapeHtml(line).replace(regex, '<mark class="log-highlight">$1</mark>')
    }
    return escapeHtml(line)
  }).join('\n')
}

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

function escapeRegExp(string: string): string {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function scrollToTop() {
  if (logContainer.value) {
    logContainer.value.scrollTop = 0
  }
}

function scrollToBottom() {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

// 环境变量编辑功能
function openEnvEditDialog(container: Container) {
  currentContainer.value = container
  editingEnvs.value = (container.env || []).map(e => {
    const [key, ...valueParts] = e.split('=')
    return { key, value: valueParts.join('='), isNew: false }
  })
  showEnvEditDialog.value = true
}

function addEditEnv() {
  editingEnvs.value.push({ key: '', value: '', isNew: true })
}

function removeEditEnv(index: number) {
  editingEnvs.value.splice(index, 1)
}

async function saveEnvChanges() {
  if (!currentContainer.value) return

  // 验证
  const invalidEnvs = editingEnvs.value.filter(e => e.key && !e.key.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/))
  if (invalidEnvs.length > 0) {
    ElMessage.warning('环境变量名只能包含字母、数字和下划线，且不能以数字开头')
    return
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    // 更新容器环境变量
    currentContainer.value.env = editingEnvs.value
      .filter(e => e.key)
      .map(e => `${e.key}=${e.value}`)

    showEnvEditDialog.value = false
    ElMessage.success('环境变量已更新（需要重启容器生效）')
  } catch (error) {
    ElMessage.error(`保存失败: ${(error as Error).message}`)
  }
}

// 容器终端功能
const terminalInputRef = ref<HTMLInputElement | null>(null)
const terminalHistory = ref<string[]>([])
const terminalHistoryIndex = ref(-1)

function openTerminalDialog(container: Container) {
  currentContainer.value = container
  terminalOutput.value = [
    `<span class="terminal-info">连接到容器 ${container.name}...</span>`,
    `<span class="terminal-success">已连接</span>`,
    ''
  ]
  terminalInput.value = ''
  terminalHistory.value = []
  terminalHistoryIndex.value = -1
  showTerminalDialog.value = true
}

function focusTerminalInput() {
  nextTick(() => {
    terminalInputRef.value?.focus()
  })
}

async function executeTerminalCommand() {
  const command = terminalInput.value.trim()
  if (!command) return

  // 添加到历史
  terminalHistory.value.push(command)
  terminalHistoryIndex.value = terminalHistory.value.length

  // 显示命令
  terminalOutput.value.push(`<span class="terminal-prompt">${currentContainer.value?.name}:~# </span>${escapeHtml(command)}`)

  // 模拟命令执行
  const output = simulateTerminalCommand(command)
  terminalOutput.value.push(...output)
  terminalOutput.value.push('')

  terminalInput.value = ''

  // 滚动到底部
  nextTick(() => {
    if (terminalContainer.value) {
      terminalContainer.value.scrollTop = terminalContainer.value.scrollHeight
    }
  })
}

function simulateTerminalCommand(command: string): string[] {
  const cmd = command.toLowerCase().trim()

  if (cmd === 'ls' || cmd === 'ls -la') {
    return [
      'total 48',
      'drwxr-xr-x 5 root root 4096 Jan 25 10:30 .',
      'drwxr-xr-x 1 root root 4096 Jan 15 08:00 ..',
      '-rw-r--r-- 1 root root 3526 Jan 15 10:00 app.js',
      'drwxr-xr-x 2 root root 4096 Jan 20 14:30 node_modules',
      '-rw-r--r-- 1 root root 1234 Jan 25 09:00 package.json'
    ]
  }

  if (cmd === 'pwd') {
    return ['/app']
  }

  if (cmd === 'whoami') {
    return ['root']
  }

  if (cmd === 'hostname') {
    return [currentContainer.value?.id?.substring(0, 12) || 'container']
  }

  if (cmd === 'env' || cmd === 'printenv') {
    return currentContainer.value?.env || ['No environment variables']
  }

  if (cmd === 'ps aux' || cmd === 'ps') {
    return [
      'USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND',
      'root         1  0.0  0.1  18236  3456 ?        Ss   10:30   0:00 /bin/sh',
      'root        15  2.5  3.1 245678 52428 ?        S    10:30   1:25 node app.js'
    ]
  }

  if (cmd === 'df -h') {
    return [
      'Filesystem      Size  Used Avail Use% Mounted on',
      'overlay          50G   15G   35G  30% /',
      'tmpfs            64M     0   64M   0% /dev',
      'shm              64M     0   64M   0% /dev/shm'
    ]
  }

  if (cmd === 'free -h' || cmd === 'free') {
    return [
      '              total        used        free      shared  buff/cache   available',
      'Mem:          7.8Gi       2.1Gi       3.2Gi       256Mi       2.5Gi       5.2Gi',
      'Swap:            0B          0B          0B'
    ]
  }

  if (cmd === 'cat /etc/os-release') {
    return [
      'NAME="Alpine Linux"',
      'ID=alpine',
      'VERSION_ID=3.18.4',
      'PRETTY_NAME="Alpine Linux v3.18"'
    ]
  }

  if (cmd === 'exit') {
    showTerminalDialog.value = false
    return []
  }

  if (cmd === 'help') {
    return [
      '可用命令: ls, pwd, whoami, hostname, env, ps, df, free, cat, exit, help',
      '这是一个模拟终端，实际命令将通过 Docker exec 执行'
    ]
  }

  if (cmd === 'clear') {
    terminalOutput.value = []
    return []
  }

  if (cmd.startsWith('echo ')) {
    return [command.substring(5)]
  }

  return [`<span class="terminal-error">命令未找到: ${escapeHtml(command)}</span>`]
}

function historyUp() {
  if (terminalHistoryIndex.value > 0) {
    terminalHistoryIndex.value--
    terminalInput.value = terminalHistory.value[terminalHistoryIndex.value]
  }
}

function historyDown() {
  if (terminalHistoryIndex.value < terminalHistory.value.length - 1) {
    terminalHistoryIndex.value++
    terminalInput.value = terminalHistory.value[terminalHistoryIndex.value]
  } else {
    terminalHistoryIndex.value = terminalHistory.value.length
    terminalInput.value = ''
  }
}

function clearTerminal() {
  terminalOutput.value = []
}

// 镜像重新标记功能
function openRetagDialog(image: Image) {
  currentImage.value = image
  retagForm.value = {
    newRepository: image.repository,
    newTag: ''
  }
  showRetagDialog.value = true
}

async function retagImage() {
  if (!retagForm.value.newRepository || !retagForm.value.newTag) {
    ElMessage.warning('请填写新的仓库名和标签')
    return
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    // 添加新标签的镜像
    const newImage: Image = {
      id: currentImage.value!.id,
      repository: retagForm.value.newRepository,
      tag: retagForm.value.newTag,
      size: currentImage.value!.size,
      created: new Date().toISOString().split('T')[0]
    }

    images.value.unshift(newImage)
    showRetagDialog.value = false
    ElMessage.success(`镜像已标记为 ${retagForm.value.newRepository}:${retagForm.value.newTag}`)
  } catch (error) {
    ElMessage.error(`标记失败: ${(error as Error).message}`)
  }
}

async function deleteContainer(container: Container) {
  try {
    await ElMessageBox.confirm(
      `确定要删除容器 "${container.name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'warning' }
    )

    await new Promise(resolve => setTimeout(resolve, 500))
    containers.value = containers.value.filter(c => c.id !== container.id)
    ElMessage.success('容器已删除')
  } catch {
    // 用户取消
  }
}

// 容器实时监控
function showContainerStats(container: Container) {
  currentContainer.value = container
  // 生成模拟监控数据
  const now = new Date()
  containerStats.value = {
    cpu: Array.from({ length: 20 }, () => Math.random() * 30 + (container.cpu || 5)),
    memory: Array.from({ length: 20 }, () => Math.random() * 100000000 + (container.memory || 100000000)),
    networkRx: Array.from({ length: 20 }, () => Math.random() * 1000000),
    networkTx: Array.from({ length: 20 }, () => Math.random() * 500000),
    timestamps: Array.from({ length: 20 }, (_, i) => {
      const t = new Date(now.getTime() - (19 - i) * 3000)
      return t.toLocaleTimeString()
    })
  }
  showStatsDialog.value = true
}

// 导出容器配置
function exportContainerConfig(container: Container) {
  const config = {
    name: container.name,
    image: container.image,
    ports: container.ports,
    env: container.env,
    mounts: container.mounts,
    restartPolicy: container.restartPolicy,
    networkMode: container.networkMode
  }

  const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${container.name}-config.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('配置已导出')
}

// 克隆容器
function cloneContainer(container: Container) {
  createForm.value = {
    name: `${container.name}-clone`,
    image: container.image,
    ports: container.ports?.map(p => ({ host: p.hostPort, container: p.containerPort })) || [],
    envs: container.env?.map(e => {
      const [key, ...valueParts] = e.split('=')
      return { key, value: valueParts.join('=') }
    }) || [],
    restartPolicy: container.restartPolicy || 'no',
    command: ''
  }
  showCreateDialog.value = true
  ElMessage.info('已加载容器配置，请修改名称后创建')
}

// 网络操作
async function loadNetworks() {
  networksLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 300))
    // 数据已在 initSimulatedData 中初始化
  } finally {
    networksLoading.value = false
  }
}

async function createNetwork() {
  if (!createNetworkForm.value.name) {
    ElMessage.warning('请输入网络名称')
    return
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    const newNetwork: Network = {
      id: 'net' + Math.random().toString(36).substring(2, 8),
      name: createNetworkForm.value.name,
      driver: createNetworkForm.value.driver,
      scope: 'local',
      subnet: createNetworkForm.value.subnet || '172.20.0.0/16',
      gateway: createNetworkForm.value.gateway || '172.20.0.1',
      containers: 0,
      created: new Date().toISOString().split('T')[0]
    }

    networks.value.push(newNetwork)
    showCreateNetworkDialog.value = false
    createNetworkForm.value = { name: '', driver: 'bridge', subnet: '', gateway: '' }
    ElMessage.success('网络创建成功')
  } catch (error) {
    ElMessage.error(`创建失败: ${(error as Error).message}`)
  }
}

async function deleteNetwork(network: Network) {
  try {
    await ElMessageBox.confirm(
      `确定要删除网络 "${network.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await new Promise(resolve => setTimeout(resolve, 300))
    networks.value = networks.value.filter(n => n.id !== network.id)
    ElMessage.success('网络已删除')
  } catch {
    // 用户取消
  }
}

function showNetworkDetail(network: Network) {
  ElMessageBox.alert(
    `<div style="line-height: 1.8">
      <p><strong>网络ID:</strong> ${network.id}</p>
      <p><strong>名称:</strong> ${network.name}</p>
      <p><strong>驱动:</strong> ${network.driver}</p>
      <p><strong>范围:</strong> ${network.scope}</p>
      <p><strong>子网:</strong> ${network.subnet || '无'}</p>
      <p><strong>网关:</strong> ${network.gateway || '无'}</p>
      <p><strong>连接容器数:</strong> ${network.containers}</p>
      <p><strong>创建时间:</strong> ${network.created}</p>
    </div>`,
    '网络详情',
    { dangerouslyUseHTMLString: true }
  )
}

// 卷操作
async function loadVolumes() {
  volumesLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 300))
    // 数据已在 initSimulatedData 中初始化
  } finally {
    volumesLoading.value = false
  }
}

async function createVolume() {
  if (!createVolumeForm.value.name) {
    ElMessage.warning('请输入卷名称')
    return
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    const newVolume: Volume = {
      name: createVolumeForm.value.name,
      driver: createVolumeForm.value.driver,
      mountpoint: `/var/lib/docker/volumes/${createVolumeForm.value.name}/_data`,
      created: new Date().toISOString().split('T')[0],
      inUse: false
    }

    volumes.value.push(newVolume)
    showCreateVolumeDialog.value = false
    createVolumeForm.value = { name: '', driver: 'local' }
    ElMessage.success('卷创建成功')
  } catch (error) {
    ElMessage.error(`创建失败: ${(error as Error).message}`)
  }
}

async function deleteVolume(volume: Volume) {
  try {
    await ElMessageBox.confirm(
      `确定要删除卷 "${volume.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await new Promise(resolve => setTimeout(resolve, 300))
    volumes.value = volumes.value.filter(v => v.name !== volume.name)
    ElMessage.success('卷已删除')
  } catch {
    // 用户取消
  }
}

function showVolumeDetail(volume: Volume) {
  ElMessageBox.alert(
    `<div style="line-height: 1.8">
      <p><strong>名称:</strong> ${volume.name}</p>
      <p><strong>驱动:</strong> ${volume.driver}</p>
      <p><strong>挂载点:</strong> ${volume.mountpoint}</p>
      <p><strong>创建时间:</strong> ${volume.created}</p>
      <p><strong>使用状态:</strong> ${volume.inUse ? '使用中' : '未使用'}</p>
    </div>`,
    '卷详情',
    { dangerouslyUseHTMLString: true }
  )
}

// 镜像历史
function showImageHistory(image: Image) {
  currentImage.value = image
  // 生成模拟历史数据
  imageHistory.value = [
    { id: image.id, created: '2024-01-15 10:00:00', createdBy: '/bin/sh -c #(nop) CMD ["nginx" "-g" "daemon off;"]', size: 0 },
    { id: 'sha256:layer1', created: '2024-01-15 09:58:00', createdBy: '/bin/sh -c #(nop) EXPOSE 80', size: 0 },
    { id: 'sha256:layer2', created: '2024-01-15 09:55:00', createdBy: '/bin/sh -c apt-get update && apt-get install -y nginx', size: 85000000 },
    { id: 'sha256:layer3', created: '2024-01-15 09:50:00', createdBy: '/bin/sh -c #(nop) ENV NGINX_VERSION=1.25.3', size: 0 },
    { id: 'sha256:base', created: '2024-01-10 08:00:00', createdBy: '/bin/sh -c #(nop) ADD file:xxx in /', size: 57000000 }
  ]
  showImageHistoryDialog.value = true
}

// 构建镜像
async function buildImage() {
  if (!buildImageForm.value.name) {
    ElMessage.warning('请输入镜像名称')
    return
  }

  try {
    ElMessage.info('开始构建镜像...')
    await new Promise(resolve => setTimeout(resolve, 2000))

    const newImage: Image = {
      id: 'sha256:' + Math.random().toString(36).substring(2, 14),
      repository: buildImageForm.value.name,
      tag: buildImageForm.value.tag || 'latest',
      size: Math.floor(Math.random() * 300000000) + 100000000,
      created: new Date().toISOString().split('T')[0]
    }

    images.value.unshift(newImage)
    showBuildImageDialog.value = false
    buildImageForm.value = { name: '', tag: 'latest', dockerfile: '', context: '.' }
    ElMessage.success('镜像构建成功')
  } catch (error) {
    ElMessage.error(`构建失败: ${(error as Error).message}`)
  }
}

// 创建容器相关
function addPort() {
  createForm.value.ports.push({ host: '', container: '' })
}

function removePort(index: number) {
  createForm.value.ports.splice(index, 1)
}

function addEnv() {
  createForm.value.envs.push({ key: '', value: '' })
}

function removeEnv(index: number) {
  createForm.value.envs.splice(index, 1)
}

async function createContainer() {
  if (!createForm.value.name || !createForm.value.image) {
    ElMessage.warning('请填写容器名称和镜像')
    return
  }

  try {
    await new Promise(resolve => setTimeout(resolve, 800))

    const newContainer: Container = {
      id: Math.random().toString(36).substring(2, 14),
      name: createForm.value.name,
      image: createForm.value.image,
      state: 'running',
      status: 'Up 1 second',
      ports: createForm.value.ports
        .filter(p => p.host && p.container)
        .map(p => ({ hostPort: p.host, containerPort: p.container })),
      cpu: 1,
      memory: 50000000,
      created: new Date().toISOString().split('T')[0],
      startedAt: new Date().toISOString(),
      networkMode: 'bridge',
      restartPolicy: createForm.value.restartPolicy,
      env: createForm.value.envs
        .filter(e => e.key)
        .map(e => `${e.key}=${e.value}`),
      mounts: []
    }

    containers.value.unshift(newContainer)
    showCreateDialog.value = false

    // 重置表单
    createForm.value = {
      name: '',
      image: '',
      ports: [],
      envs: [],
      restartPolicy: 'no',
      command: ''
    }

    ElMessage.success('容器创建成功')
  } catch (error) {
    ElMessage.error(`创建失败: ${(error as Error).message}`)
  }
}

function createContainerFromImage(image: Image) {
  createForm.value.image = `${image.repository}:${image.tag}`
  showCreateDialog.value = true
}

async function pullImage() {
  if (!pullImageName.value) {
    ElMessage.warning('请输入镜像名称')
    return
  }

  isPulling.value = true
  pullProgress.value = 0
  pullStatus.value = ''
  pullMessage.value = '正在连接...'

  try {
    // 模拟拉取进度
    for (let i = 0; i <= 100; i += 10) {
      await new Promise(resolve => setTimeout(resolve, 300))
      pullProgress.value = i
      if (i < 30) pullMessage.value = '正在下载层...'
      else if (i < 70) pullMessage.value = '正在解压...'
      else if (i < 100) pullMessage.value = '正在验证...'
      else pullMessage.value = '完成'
    }

    pullStatus.value = 'success'

    // 添加新镜像
    const [repo, tag] = pullImageName.value.split(':')
    images.value.unshift({
      id: 'sha256:' + Math.random().toString(36).substring(2, 14),
      repository: repo,
      tag: tag || 'latest',
      size: Math.floor(Math.random() * 500000000) + 50000000,
      created: new Date().toISOString().split('T')[0]
    })

    ElMessage.success('镜像拉取成功')

    setTimeout(() => {
      showPullImageDialog.value = false
      pullImageName.value = ''
      pullProgress.value = 0
    }, 1000)
  } catch (error) {
    pullStatus.value = 'exception'
    pullMessage.value = '拉取失败'
    ElMessage.error(`拉取失败: ${(error as Error).message}`)
  } finally {
    isPulling.value = false
  }
}

async function deleteImage(image: Image) {
  try {
    await ElMessageBox.confirm(
      `确定要删除镜像 "${image.repository}:${image.tag}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )

    await new Promise(resolve => setTimeout(resolve, 300))
    images.value = images.value.filter(img => img.id !== image.id)
    ElMessage.success('镜像已删除')
  } catch {
    // 用户取消
  }
}
</script>

<style lang="scss" scoped>
.containers-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h1 {
    font-size: 24px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 12px;

    .server-select {
      width: 200px;
    }

    .search-input {
      width: 200px;
    }
  }
}

.main-tabs {
  margin-bottom: 20px;
}

.stats-row {
  display: flex;
  gap: 24px;
  margin-bottom: 24px;

  .stat-item {
    display: flex;
    align-items: baseline;
    gap: 8px;
    padding: 12px 20px;
    background: var(--bg-secondary);
    border-radius: 8px;

    .stat-value {
      font-size: 28px;
      font-weight: 600;
    }

    .stat-label {
      color: var(--text-secondary);
    }

    &.running .stat-value {
      color: var(--success-color);
    }

    &.stopped .stat-value {
      color: var(--danger-color);
    }

    &.paused .stat-value {
      color: var(--warning-color);
    }
  }
}

.container-table {
  .container-name {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;

    &:hover .name-text {
      color: var(--primary-color);
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background-color: var(--text-secondary);
      flex-shrink: 0;

      &.running {
        background-color: var(--success-color);
      }

      &.exited {
        background-color: var(--danger-color);
      }

      &.paused {
        background-color: var(--warning-color);
      }
    }

    .name-text {
      font-weight: 500;
      transition: color 0.2s;
    }
  }

  .action-buttons {
    display: flex;
    gap: 8px;
  }

  .cpu-high {
    color: var(--danger-color);
    font-weight: 600;
  }

  .cpu-medium {
    color: var(--warning-color);
  }

  .cpu-low {
    color: var(--success-color);
  }
}

.danger-text {
  color: var(--danger-color);
}

.images-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .search-input {
    width: 250px;
  }
}

.images-table {
  .image-id {
    font-family: monospace;
    color: var(--text-secondary);
  }
}

.log-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.log-container {
  height: 60vh;
  overflow: auto;
  background-color: var(--bg-color);
  border-radius: 8px;
  padding: 16px;

  .log-content {
    font-family: 'Fira Code', monospace;
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }
}

.mono-text {
  font-family: monospace;
}

.env-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  .env-tag {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .more-env {
    color: var(--text-secondary);
    font-size: 12px;
  }
}

.mount-item {
  font-family: monospace;
  font-size: 12px;
  padding: 4px 0;
  color: var(--text-secondary);
}

.port-row,
.env-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.port-separator,
.env-separator {
  color: var(--text-secondary);
}

.pull-progress {
  margin-top: 16px;

  .pull-message {
    margin-top: 8px;
    font-size: 13px;
    color: var(--text-secondary);
  }
}

.empty-state {
  padding: 60px 0;
}

.text-secondary {
  color: var(--text-secondary);
}

// 过滤栏样式
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-radius: 8px;

  .filter-buttons {
    .filter-dot {
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 4px;

      &.running {
        background-color: var(--success-color);
      }

      &.stopped {
        background-color: var(--danger-color);
      }

      &.paused {
        background-color: var(--warning-color);
      }
    }
  }

  .batch-actions {
    display: flex;
    align-items: center;
    gap: 12px;

    .selected-count {
      font-size: 13px;
      color: var(--text-secondary);
    }
  }
}

// 资源使用统计样式
.stats-row {
  .stat-item.resources {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .resource-summary {
      display: flex;
      align-items: center;
      gap: 12px;

      .resource-label {
        font-size: 12px;
        color: var(--text-secondary);
        min-width: 80px;
      }

      .el-progress {
        flex: 1;
        max-width: 200px;
      }

      .resource-value {
        font-size: 12px;
        font-weight: 500;
        min-width: 60px;
      }
    }
  }
}

// 资源进度条样式
.resource-bars {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .resource-bar {
    display: flex;
    align-items: center;
    gap: 6px;

    .bar-label {
      font-size: 10px;
      color: var(--text-secondary);
      width: 28px;
    }

    .el-progress {
      flex: 1;
    }

    .bar-value {
      font-size: 10px;
      color: var(--text-secondary);
      min-width: 50px;
      text-align: right;
    }
  }
}

// 健康状态标签
.health-tag {
  margin-left: 8px;
}

// 网络和卷页面样式
.networks-header,
.volumes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .search-input {
    width: 250px;
  }
}

.networks-table,
.volumes-table {
  .network-name,
  .volume-name {
    display: flex;
    align-items: center;
    gap: 8px;

    .el-icon {
      color: var(--primary-color);
    }
  }

  .mountpoint {
    font-family: monospace;
    font-size: 12px;
    color: var(--text-secondary);
  }
}

// 镜像操作按钮
.images-actions {
  display: flex;
  gap: 12px;
}

// 容器监控对话框样式
.stats-container {
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .stats-card {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 16px;

    .stats-card-header {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      color: var(--text-secondary);
      margin-bottom: 8px;
    }

    .stats-card-value {
      font-size: 24px;
      font-weight: 600;
      margin-bottom: 12px;
    }

    .stats-chart {
      height: 60px;
    }

    .mini-chart {
      display: flex;
      align-items: flex-end;
      gap: 2px;
      height: 100%;

      .chart-bar {
        flex: 1;
        background: var(--primary-color);
        border-radius: 2px 2px 0 0;
        min-height: 2px;
        transition: height 0.3s;

        &.memory {
          background: var(--success-color);
        }
      }
    }
  }
}

// 镜像历史对话框样式
.history-table {
  .layer-id {
    font-family: monospace;
    color: var(--text-secondary);
  }

  .created-by {
    font-family: monospace;
    font-size: 12px;
    color: var(--text-secondary);
    display: block;
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

// 增强的日志对话框样式
.log-dialog {
  .log-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    flex-wrap: wrap;
    gap: 12px;

    .log-toolbar-left,
    .log-toolbar-right {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .log-stats {
    display: flex;
    gap: 8px;
    margin-bottom: 8px;
  }

  .log-container {
    height: 55vh;
    overflow: auto;
    background-color: #1e1e1e;
    border-radius: 8px;
    padding: 16px;

    .log-content {
      font-family: 'Fira Code', 'Consolas', monospace;
      font-size: 12px;
      line-height: 1.6;
      white-space: pre-wrap;
      word-break: break-all;
      color: #d4d4d4;
      margin: 0;

      .log-error {
        color: #f85149;
      }

      .log-warn {
        color: #d29922;
      }

      .log-info {
        color: #58a6ff;
      }

      .log-debug {
        color: #8b949e;
      }

      .log-highlight {
        background-color: #634d00;
        color: #fff;
        padding: 0 2px;
        border-radius: 2px;
      }
    }
  }
}

// 环境变量编辑对话框样式
.env-edit-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .env-count {
    font-size: 13px;
    color: var(--text-secondary);
  }
}

.env-edit-list {
  max-height: 400px;
  overflow-y: auto;

  .env-edit-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px;
    border-radius: 6px;
    margin-bottom: 8px;
    background: var(--bg-tertiary);

    &.new-env {
      border: 1px dashed var(--primary-color);
    }

    .env-equals {
      color: var(--text-secondary);
      font-weight: 600;
    }
  }

  .no-envs {
    text-align: center;
    color: var(--text-secondary);
    padding: 40px;
  }
}

// 容器终端对话框样式
.terminal-dialog {
  .terminal-container {
    background: #0d1117;
    border-radius: 8px;
    padding: 16px;
    height: 50vh;
    overflow-y: auto;
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 13px;
    line-height: 1.5;

    .terminal-output {
      .terminal-line {
        color: #c9d1d9;
        white-space: pre-wrap;
        word-break: break-all;

        .terminal-prompt {
          color: #58a6ff;
        }

        .terminal-info {
          color: #8b949e;
        }

        .terminal-success {
          color: #3fb950;
        }

        .terminal-error {
          color: #f85149;
        }
      }
    }

    .terminal-input-line {
      display: flex;
      align-items: center;
      margin-top: 8px;

      .terminal-prompt {
        color: #58a6ff;
        white-space: nowrap;
      }

      .terminal-input {
        flex: 1;
        background: transparent;
        border: none;
        outline: none;
        color: #c9d1d9;
        font-family: inherit;
        font-size: inherit;
        padding: 0;
        margin-left: 4px;

        &::placeholder {
          color: #484f58;
        }
      }
    }
  }
}

// 镜像重新标记对话框样式
.retag-dialog {
  .current-tag {
    font-family: monospace;
    background: var(--bg-tertiary);
    padding: 8px 12px;
    border-radius: 4px;
  }
}
</style>
