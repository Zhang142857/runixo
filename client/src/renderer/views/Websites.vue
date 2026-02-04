<template>
  <div class="websites">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>网站管理</h1>
        <p class="subtitle">站点配置与项目部署</p>
      </div>
      <div class="header-actions">
        <el-select v-if="hasMultipleServers" v-model="selectedServer" placeholder="选择服务器" size="small">
          <el-option v-for="s in connectedServers" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="refresh" :loading="loading" size="small">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
        <el-dropdown @command="handleAddCommand">
          <el-button type="primary" size="small">
            <el-icon><Plus /></el-icon>添加<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="static">静态站点</el-dropdown-item>
              <el-dropdown-item command="project">项目部署</el-dropdown-item>
              <el-dropdown-item command="proxy">反向代理</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-if="!selectedServer" class="empty-state">
      <el-empty description="请先选择一个已连接的服务器" />
    </div>

    <template v-else>
      <!-- 标签页 -->
      <el-tabs v-model="activeTab" class="main-tabs">
        <el-tab-pane name="sites">
          <template #label>
            <span class="tab-label">站点列表 <el-badge :value="sites.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
        <el-tab-pane name="projects">
          <template #label>
            <span class="tab-label">项目部署 <el-badge :value="projects.length" :max="99" type="info" /></span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <!-- 站点列表 -->
      <div v-show="activeTab === 'sites'" class="tab-content">
        <el-table :data="sites" v-loading="loading" size="small" class="data-table">
          <el-table-column prop="name" label="站点名称" min-width="140">
            <template #default="{ row }">
              <div class="cell-name">
                <span class="status-dot" :class="row.status"></span>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="domain" label="域名" min-width="180">
            <template #default="{ row }">
              <a :href="(row.ssl ? 'https://' : 'http://') + row.domain" target="_blank" class="domain-link">
                <el-icon v-if="row.ssl"><Lock /></el-icon>
                {{ row.domain }}
              </a>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="getTypeTag(row.type)">{{ getTypeLabel(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="目录" min-width="160">
            <template #default="{ row }">
              <code class="mono">{{ row.path }}</code>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button-group size="small">
                <el-button @click="openSite(row)">打开</el-button>
                <el-button @click="editSite(row)">设置</el-button>
                <el-button v-if="row.status === 'running'" type="warning" @click="toggleSite(row, 'stop')">停止</el-button>
                <el-button v-else type="success" @click="toggleSite(row, 'start')">启动</el-button>
                <el-button type="danger" @click="deleteSite(row)">删除</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 项目部署 -->
      <div v-show="activeTab === 'projects'" class="tab-content">
        <div v-if="projects.length === 0" class="empty-projects">
          <el-empty description="暂无部署项目">
            <el-button type="primary" size="small" @click="handleAddCommand('project')">创建项目部署</el-button>
          </el-empty>
        </div>
        <div v-else class="projects-grid">
          <div v-for="project in projects" :key="project.id" class="project-card">
            <div class="project-header">
              <div class="project-icon" :style="{ background: getProjectColor(project.type) }">
                <TechIcon :name="project.type" />
              </div>
              <div class="project-info">
                <div class="project-name">{{ project.name }}</div>
                <div class="project-domain">{{ project.domain }}</div>
              </div>
              <el-tag :type="getProjectStatusType(project.status)" size="small">{{ getProjectStatusLabel(project.status) }}</el-tag>
            </div>
            <div class="project-meta">
              <div class="meta-item"><span class="meta-label">目录:</span> <code>{{ project.path }}</code></div>
              <div class="meta-item"><span class="meta-label">端口:</span> {{ project.port }}</div>
              <div class="meta-item" v-if="project.lastDeploy"><span class="meta-label">上次部署:</span> {{ formatTime(project.lastDeploy) }}</div>
            </div>
            <div class="project-actions">
              <el-button size="small" type="primary" @click="deployProject(project)" :loading="project.deploying">
                {{ project.deploying ? '部署中' : '部署' }}
              </el-button>
              <el-button size="small" @click="viewProjectLogs(project)">日志</el-button>
              <el-button size="small" @click="editProject(project)">设置</el-button>
              <el-button size="small" v-if="project.status === 'running'" type="warning" @click="stopProject(project)">停止</el-button>
              <el-button size="small" v-else type="success" @click="startProject(project)">启动</el-button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- 添加静态站点对话框 -->
    <el-dialog v-model="showAddStatic" title="添加静态站点" width="520px" class="site-dialog" destroy-on-close>
      <el-form :model="newSite" label-width="80px" size="default" class="site-form">
        <el-form-item label="站点名称" required>
          <el-input v-model="newSite.name" placeholder="my-website">
            <template #prefix><el-icon><Edit /></el-icon></template>
          </el-input>
          <div class="form-tip">用于标识站点，建议使用英文</div>
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="newSite.domain" placeholder="example.com">
            <template #prefix><el-icon><Link /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="根目录" required>
          <el-input v-model="newSite.path" placeholder="/var/www/html">
            <template #prefix><el-icon><Folder /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="newSite.ssl" />
          <span class="switch-label">{{ newSite.ssl ? '使用 HTTPS' : '使用 HTTP' }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showAddStatic = false">取消</el-button>
          <el-button type="primary" @click="createStaticSite" :loading="creating">
            <el-icon><Check /></el-icon>创建站点
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 添加反向代理对话框 -->
    <el-dialog v-model="showAddProxy" title="添加反向代理" width="520px" class="site-dialog" destroy-on-close>
      <el-form :model="newProxy" label-width="80px" size="default" class="site-form">
        <el-form-item label="站点名称" required>
          <el-input v-model="newProxy.name" placeholder="my-api">
            <template #prefix><el-icon><Edit /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="域名" required>
          <el-input v-model="newProxy.domain" placeholder="api.example.com">
            <template #prefix><el-icon><Link /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="代理地址" required>
          <el-input v-model="newProxy.upstream" placeholder="http://127.0.0.1:3000">
            <template #prefix><el-icon><Position /></el-icon></template>
          </el-input>
          <div class="form-tip">后端服务地址，如 http://127.0.0.1:3000</div>
        </el-form-item>
        <el-form-item label="WebSocket">
          <el-switch v-model="newProxy.websocket" />
          <span class="switch-label">{{ newProxy.websocket ? '支持 WebSocket' : '不支持 WebSocket' }}</span>
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="newProxy.ssl" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showAddProxy = false">取消</el-button>
          <el-button type="primary" @click="createProxySite" :loading="creating">
            <el-icon><Check /></el-icon>创建代理
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 项目部署对话框 - 重新设计的多步骤向导 -->
    <el-dialog v-model="showAddProject" title="" width="960px" class="deploy-wizard-dialog" :show-close="false" destroy-on-close>
      <div class="wizard-container">
        <!-- 顶部进度条 -->
        <div class="wizard-header">
          <div class="wizard-title">
            <el-icon class="title-icon"><Promotion /></el-icon>
            <span>项目部署向导</span>
          </div>
          <el-button class="close-btn" text circle @click="showAddProject = false">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
        
        <!-- 步骤指示器 -->
        <div class="wizard-steps">
          <div 
            v-for="(step, index) in deploySteps" 
            :key="step.key"
            class="wizard-step"
            :class="{ 
              active: deployStep === step.key, 
              completed: index < deployStepIndex,
              clickable: index <= deployStepIndex
            }"
            @click="index <= deployStepIndex && (deployStep = step.key)"
          >
            <div class="step-indicator">
              <el-icon v-if="index < deployStepIndex"><Check /></el-icon>
              <span v-else>{{ index + 1 }}</span>
            </div>
            <div class="step-info">
              <div class="step-title">{{ step.title }}</div>
              <div class="step-desc">{{ step.desc }}</div>
            </div>
          </div>
        </div>

        <!-- 步骤内容区域 -->
        <div class="wizard-content">
          <!-- 步骤1: 基本信息 -->
          <div v-show="deployStep === 'basic'" class="step-panel">
            <div class="panel-header">
              <h3><el-icon><Setting /></el-icon> 基本信息</h3>
              <p>设置项目名称、类型和运行环境</p>
            </div>
            
            <el-form :model="newProject" label-position="top" size="default" class="wizard-form">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="项目名称" required>
                    <el-input v-model="newProject.name" placeholder="my-app" maxlength="32" show-word-limit>
                      <template #prefix><el-icon><Edit /></el-icon></template>
                    </el-input>
                    <div class="form-tip">用于标识项目，建议使用英文和短横线</div>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="运行端口" v-if="!['php', 'static-build'].includes(newProject.type)">
                    <el-input-number v-model="newProject.port" :min="1024" :max="65535" style="width: 100%" controls-position="right" />
                    <div class="form-tip">应用监听端口，Nginx 会转发请求到此端口</div>
                  </el-form-item>
                </el-col>
              </el-row>

              <el-form-item label="项目类型" required>
                <div class="type-selector">
                  <div 
                    v-for="pt in projectTypes" 
                    :key="pt.value"
                    class="type-card"
                    :class="{ active: newProject.type === pt.value }"
                    @click="selectProjectType(pt.value)"
                  >
                    <div class="type-icon" :style="{ background: pt.color }">
                      <TechIcon :name="pt.value" />
                    </div>
                    <div class="type-info">
                      <div class="type-name">{{ pt.label }}</div>
                      <div class="type-desc">{{ pt.desc }}</div>
                    </div>
                    <el-icon v-if="newProject.type === pt.value" class="type-check"><CircleCheck /></el-icon>
                  </div>
                </div>
              </el-form-item>

              <el-form-item label="项目目录" required>
                <div class="path-input-group">
                  <el-input v-model="newProject.path" placeholder="/var/www/my-app">
                    <template #prefix><el-icon><Folder /></el-icon></template>
                  </el-input>
                  <el-button type="primary" plain @click="showProjectPathBrowser = true">
                    <el-icon><FolderOpened /></el-icon>浏览
                  </el-button>
                </div>
                <div class="form-tip">项目代码存放的服务器目录</div>
              </el-form-item>
            </el-form>
          </div>

          <!-- 步骤2: 上传代码 -->
          <div v-show="deployStep === 'upload'" class="step-panel">
            <div class="panel-header">
              <h3><el-icon><Upload /></el-icon> 上传代码</h3>
              <p>选择本地项目文件夹上传到服务器</p>
            </div>

            <div class="upload-area">
              <!-- 选择文件夹 -->
              <div class="upload-dropzone" v-if="!selectedLocalPath" @click="selectFolder">
                <div class="dropzone-content">
                  <el-icon class="dropzone-icon"><UploadFilled /></el-icon>
                  <div class="dropzone-title">点击选择项目文件夹</div>
                  <div class="dropzone-hint">选择包含项目代码的本地文件夹</div>
                </div>
              </div>

              <!-- 已选择文件夹 -->
              <div class="upload-preview" v-else>
                <div class="preview-header">
                  <div class="preview-path">
                    <el-icon><Folder /></el-icon>
                    <span>{{ selectedLocalPath }}</span>
                  </div>
                  <el-button text type="primary" @click="selectFolder">
                    <el-icon><RefreshRight /></el-icon>重新选择
                  </el-button>
                </div>

                <!-- 检测到的项目信息 -->
                <div class="detected-info" v-if="detectedProjectInfo">
                  <div class="info-header">
                    <el-icon><InfoFilled /></el-icon>
                    <span>检测到的项目信息</span>
                  </div>
                  <div class="info-content">
                    <div class="info-item" v-if="detectedProjectInfo.name">
                      <span class="info-label">项目名称:</span>
                      <span class="info-value">{{ detectedProjectInfo.name }}</span>
                    </div>
                    <div class="info-item" v-if="detectedProjectInfo.type">
                      <span class="info-label">项目类型:</span>
                      <el-tag size="small" :color="getProjectColor(detectedProjectInfo.type)">{{ getProjectTypeLabel(detectedProjectInfo.type) }}</el-tag>
                    </div>
                    <div class="info-item" v-if="detectedProjectInfo.scripts && detectedProjectInfo.scripts.length">
                      <span class="info-label">可用脚本:</span>
                      <div class="script-tags">
                        <el-tag v-for="s in detectedProjectInfo.scripts" :key="s" size="small" type="info">{{ s }}</el-tag>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 文件列表 -->
                <div class="file-list-panel">
                  <div class="list-header">
                    <span>文件预览</span>
                    <span class="file-count">{{ uploadFiles.length }} 个文件</span>
                  </div>
                  <div class="file-list">
                    <div v-for="file in uploadFiles.slice(0, 8)" :key="file.path" class="file-row">
                      <el-icon v-if="file.isDir" class="file-icon folder"><Folder /></el-icon>
                      <el-icon v-else class="file-icon"><Document /></el-icon>
                      <span class="file-name">{{ file.name }}</span>
                      <span class="file-size" v-if="!file.isDir">{{ formatFileSize(file.size) }}</span>
                    </div>
                    <div v-if="uploadFiles.length > 8" class="file-more">
                      还有 {{ uploadFiles.length - 8 }} 个文件...
                    </div>
                  </div>
                </div>

                <!-- 上传目标 -->
                <div class="upload-target">
                  <el-icon><Right /></el-icon>
                  <span>上传到:</span>
                  <code>{{ newProject.path || '/var/www/' + newProject.name }}</code>
                </div>

                <!-- 上传进度 -->
                <div class="upload-progress" v-if="uploading">
                  <el-progress :percentage="uploadProgress" :stroke-width="10" :status="uploadProgress === 100 ? 'success' : ''" />
                  <div class="progress-text">{{ uploadLog }}</div>
                </div>
              </div>

              <!-- 跳过上传 -->
              <div class="skip-upload">
                <el-checkbox v-model="newProject.skipUpload">
                  <span>跳过上传</span>
                  <span class="skip-hint">（代码已在服务器上或稍后手动上传）</span>
                </el-checkbox>
              </div>
            </div>
          </div>

          <!-- 步骤3: 域名设置 -->
          <div v-show="deployStep === 'domain'" class="step-panel">
            <div class="panel-header">
              <h3><el-icon><Link /></el-icon> 域名设置</h3>
              <p>配置访问域名，让用户可以通过域名访问你的应用</p>
            </div>

            <!-- 服务器信息卡片 -->
            <div class="server-info-card">
              <div class="card-header">
                <el-icon><Monitor /></el-icon>
                <span>服务器信息</span>
              </div>
              <div class="card-body">
                <div class="info-row">
                  <span class="info-label">公网 IP</span>
                  <code class="info-value">{{ serverPublicIP || '获取中...' }}</code>
                  <el-button text size="small" @click="copyToClipboard(serverPublicIP)">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
                <div class="info-row" v-if="serverLocalIP">
                  <span class="info-label">内网 IP</span>
                  <code class="info-value secondary">{{ serverLocalIP }}</code>
                </div>
              </div>
            </div>

            <el-form :model="newProject" label-position="top" size="default" class="wizard-form">
              <el-form-item label="访问方式">
                <div class="access-type-cards">
                  <div 
                    class="access-card"
                    :class="{ active: newProject.domainType === 'ip' }"
                    @click="newProject.domainType = 'ip'"
                  >
                    <el-icon class="card-icon"><Monitor /></el-icon>
                    <div class="card-content">
                      <div class="card-title">IP 直接访问</div>
                      <div class="card-desc">通过服务器 IP 和端口访问，无需域名</div>
                    </div>
                    <el-icon v-if="newProject.domainType === 'ip'" class="card-check"><CircleCheck /></el-icon>
                  </div>
                  <div 
                    class="access-card"
                    :class="{ active: newProject.domainType === 'domain' }"
                    @click="newProject.domainType = 'domain'"
                  >
                    <el-icon class="card-icon"><Link /></el-icon>
                    <div class="card-content">
                      <div class="card-title">域名访问</div>
                      <div class="card-desc">需要先将域名解析到服务器 IP</div>
                    </div>
                    <el-icon v-if="newProject.domainType === 'domain'" class="card-check"><CircleCheck /></el-icon>
                  </div>
                </div>
              </el-form-item>

              <template v-if="newProject.domainType === 'ip'">
                <el-form-item label="访问地址">
                  <el-input :model-value="`http://${serverPublicIP}:${newProject.port || 80}`" disabled class="readonly-input" />
                  <div class="form-tip">部署完成后，可通过此地址访问应用</div>
                </el-form-item>
              </template>

              <template v-else>
                <el-form-item label="域名" required>
                  <el-input v-model="newProject.domain" placeholder="app.example.com">
                    <template #prepend>http(s)://</template>
                  </el-input>
                </el-form-item>

                <!-- DNS 配置指引 -->
                <div class="dns-guide">
                  <div class="guide-header">
                    <el-icon><InfoFilled /></el-icon>
                    <span>DNS 配置指引</span>
                  </div>
                  <div class="guide-content">
                    <div class="guide-step">
                      <div class="step-num">1</div>
                      <div class="step-content">
                        <div class="step-title">登录域名服务商</div>
                        <div class="step-desc">如阿里云、腾讯云、Cloudflare 等</div>
                      </div>
                    </div>
                    <div class="guide-step">
                      <div class="step-num">2</div>
                      <div class="step-content">
                        <div class="step-title">添加 DNS 解析记录</div>
                        <div class="step-desc">
                          类型: <code>A</code>，
                          主机记录: <code>{{ getDomainPrefix(newProject.domain) || 'app' }}</code>，
                          记录值: <code>{{ serverPublicIP }}</code>
                        </div>
                      </div>
                    </div>
                    <div class="guide-step">
                      <div class="step-num">3</div>
                      <div class="step-content">
                        <div class="step-title">等待生效</div>
                        <div class="step-desc">DNS 解析通常需要几分钟到几小时生效</div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </el-form>
          </div>

          <!-- 步骤4: 部署设置 -->
          <div v-show="deployStep === 'deploy'" class="step-panel">
            <div class="panel-header">
              <h3><el-icon><SetUp /></el-icon> 部署设置</h3>
              <p>配置构建命令和启动流程</p>
            </div>

            <el-form :model="newProject" label-position="top" size="default" class="wizard-form">
              <!-- 进程管理器选择 -->
              <el-form-item label="进程管理" v-if="!['php', 'static-build'].includes(newProject.type)">
                <div class="pm-selector">
                  <div 
                    v-for="pm in processManagers" 
                    :key="pm.value"
                    class="pm-card"
                    :class="{ active: newProject.processManager === pm.value }"
                    @click="newProject.processManager = pm.value"
                  >
                    <div class="pm-icon">{{ pm.icon }}</div>
                    <div class="pm-info">
                      <div class="pm-name">{{ pm.label }}</div>
                      <div class="pm-desc">{{ pm.desc }}</div>
                    </div>
                  </div>
                </div>
              </el-form-item>

              <!-- 构建步骤 -->
              <el-form-item label="构建步骤">
                <div class="build-steps">
                  <div v-for="(step, index) in newProject.buildSteps" :key="index" class="build-step">
                    <div class="step-num">{{ index + 1 }}</div>
                    <el-input v-model="step.command" placeholder="npm install" class="step-input">
                      <template #prefix><el-icon><Cpu /></el-icon></template>
                    </el-input>
                    <el-checkbox v-model="step.optional" class="step-optional">可选</el-checkbox>
                    <el-button text type="danger" @click="removeBuildStep(index)" :disabled="newProject.buildSteps.length <= 1">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button class="add-step-btn" @click="addBuildStep" text type="primary">
                    <el-icon><Plus /></el-icon>添加构建步骤
                  </el-button>
                </div>
                <div class="form-tip">
                  <el-icon><InfoFilled /></el-icon>
                  按顺序执行的构建命令。勾选"可选"的步骤如果失败不会中断部署。
                </div>
              </el-form-item>

              <!-- 启动命令 -->
              <el-form-item label="启动命令" v-if="!['php', 'static-build'].includes(newProject.type)">
                <el-input v-model="newProject.startCommand" :placeholder="getDefaultStartCommand(newProject.type)">
                  <template #prefix><el-icon><VideoPlay /></el-icon></template>
                </el-input>
                <div class="form-tip">应用启动命令，将由进程管理器管理</div>
              </el-form-item>

              <!-- 输出目录（静态构建） -->
              <el-form-item label="输出目录" v-if="newProject.type === 'static-build'">
                <el-input v-model="newProject.outputDir" placeholder="dist">
                  <template #prefix><el-icon><Folder /></el-icon></template>
                </el-input>
                <div class="form-tip">构建产物目录，Nginx 将直接托管此目录</div>
              </el-form-item>

              <!-- 环境变量 -->
              <el-form-item label="环境变量">
                <div class="env-vars">
                  <div v-for="(env, index) in newProject.envVars" :key="index" class="env-row">
                    <el-input v-model="env.key" placeholder="变量名" class="env-key" />
                    <span class="env-eq">=</span>
                    <el-input 
                      v-model="env.value" 
                      placeholder="变量值" 
                      class="env-value"
                      :type="isSecretKey(env.key) ? 'password' : 'text'" 
                      show-password 
                    />
                    <el-button text type="danger" @click="removeEnvVar(index)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button class="add-step-btn" @click="addEnvVar" text type="primary">
                    <el-icon><Plus /></el-icon>添加环境变量
                  </el-button>
                </div>
              </el-form-item>
            </el-form>
          </div>

          <!-- 步骤5: SSL 设置 -->
          <div v-show="deployStep === 'ssl'" class="step-panel">
            <div class="panel-header">
              <h3><el-icon><Lock /></el-icon> SSL 证书</h3>
              <p>启用 HTTPS 加密访问，保护数据传输安全</p>
            </div>

            <el-form :model="newProject" label-position="top" size="default" class="wizard-form">
              <el-form-item>
                <div class="ssl-cards">
                  <div 
                    class="ssl-card"
                    :class="{ active: !newProject.ssl }"
                    @click="newProject.ssl = false"
                  >
                    <el-icon class="ssl-icon"><Unlock /></el-icon>
                    <div class="ssl-info">
                      <div class="ssl-title">HTTP</div>
                      <div class="ssl-desc">不启用 SSL 加密</div>
                    </div>
                  </div>
                  <div 
                    class="ssl-card"
                    :class="{ active: newProject.ssl }"
                    @click="newProject.ssl = true"
                  >
                    <el-icon class="ssl-icon"><Lock /></el-icon>
                    <div class="ssl-info">
                      <div class="ssl-title">HTTPS</div>
                      <div class="ssl-desc">使用 Let's Encrypt 免费证书</div>
                    </div>
                  </div>
                </div>
              </el-form-item>

              <el-alert v-if="newProject.ssl" type="info" :closable="false" show-icon class="ssl-notice">
                <template #title>SSL 证书将在项目创建后自动申请</template>
                <template #default>
                  <div class="notice-content">
                    使用 Let's Encrypt 免费证书，需要确保：
                    <ul>
                      <li>域名已正确解析到服务器 IP</li>
                      <li>服务器 80 端口可被外网访问</li>
                    </ul>
                  </div>
                </template>
              </el-alert>

              <!-- 部署预览 -->
              <div class="deploy-preview">
                <div class="preview-header">
                  <el-icon><View /></el-icon>
                  <span>部署预览</span>
                </div>
                <div class="preview-content">
                  <div class="preview-item">
                    <span class="preview-label">项目名称</span>
                    <span class="preview-value">{{ newProject.name || '-' }}</span>
                  </div>
                  <div class="preview-item">
                    <span class="preview-label">项目类型</span>
                    <span class="preview-value">{{ getProjectTypeLabel(newProject.type) }}</span>
                  </div>
                  <div class="preview-item">
                    <span class="preview-label">访问地址</span>
                    <span class="preview-value">
                      {{ newProject.ssl ? 'https://' : 'http://' }}{{ newProject.domainType === 'ip' ? serverPublicIP + ':' + newProject.port : newProject.domain }}
                    </span>
                  </div>
                  <div class="preview-item">
                    <span class="preview-label">项目目录</span>
                    <code class="preview-value">{{ newProject.path }}</code>
                  </div>
                  <div class="preview-item" v-if="newProject.processManager && !['php', 'static-build'].includes(newProject.type)">
                    <span class="preview-label">进程管理</span>
                    <span class="preview-value">{{ getProcessManagerLabel(newProject.processManager) }}</span>
                  </div>
                </div>
              </div>
            </el-form>
          </div>
        </div>

        <!-- 底部操作栏 -->
        <div class="wizard-footer">
          <el-button @click="showAddProject = false">取消</el-button>
          <div class="footer-right">
            <el-button v-if="deployStepIndex > 0" @click="prevDeployStep">
              <el-icon><ArrowLeft /></el-icon>上一步
            </el-button>
            <el-button v-if="deployStepIndex < deploySteps.length - 1" type="primary" @click="nextDeployStep">
              下一步<el-icon><ArrowRight /></el-icon>
            </el-button>
            <el-button v-else type="primary" @click="createProject" :loading="creating">
              <el-icon><Check /></el-icon>创建并部署
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 部署日志对话框 -->
    <el-dialog v-model="showDeployLog" :title="`部署日志 - ${currentProject?.name}`" width="800px" top="5vh" class="log-dialog" destroy-on-close>
      <div class="deploy-log-container">
        <div class="log-toolbar">
          <el-button-group size="small">
            <el-button @click="scrollLogToTop"><el-icon><Top /></el-icon></el-button>
            <el-button @click="scrollLogToBottom"><el-icon><Bottom /></el-icon></el-button>
          </el-button-group>
          <el-button size="small" @click="copyLog"><el-icon><CopyDocument /></el-icon>复制日志</el-button>
        </div>
        <div class="log-content" ref="logContainer">
          <pre>{{ deployLog }}</pre>
        </div>
      </div>
      <template #footer>
        <el-button @click="showDeployLog = false">关闭</el-button>
        <el-button type="primary" @click="loadProjectLogs(currentProject!)" :loading="loadingLogs">刷新日志</el-button>
      </template>
    </el-dialog>

    <!-- 站点设置对话框 -->
    <el-dialog v-model="showSiteSettings" :title="`站点设置 - ${currentSite?.name}`" width="600px" class="site-dialog" destroy-on-close>
      <el-form :model="currentSite" label-width="100px" size="default" v-if="currentSite" class="site-form">
        <el-form-item label="域名">
          <el-input v-model="currentSite.domain" />
        </el-form-item>
        <el-form-item label="根目录">
          <el-input v-model="currentSite.path" />
        </el-form-item>
        <el-form-item label="启用 SSL">
          <el-switch v-model="currentSite.ssl" />
        </el-form-item>
        <el-form-item label="伪静态">
          <div class="rewrite-presets">
            <el-button size="small" @click="applyRewrite('vue')">Vue/React</el-button>
            <el-button size="small" @click="applyRewrite('laravel')">Laravel</el-button>
            <el-button size="small" @click="applyRewrite('wordpress')">WordPress</el-button>
          </div>
          <el-input type="textarea" v-model="currentSite.rewrite" :rows="6" class="code-textarea" placeholder="location / { try_files $uri $uri/ /index.html; }" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showSiteSettings = false">取消</el-button>
          <el-button type="primary" @click="saveSiteSettings" :loading="saving">保存设置</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 项目设置对话框 -->
    <el-dialog v-model="showProjectSettings" :title="`项目设置 - ${currentProject?.name}`" width="640px" class="site-dialog" destroy-on-close>
      <el-form :model="currentProject" label-width="100px" size="default" v-if="currentProject" class="site-form">
        <el-form-item label="域名">
          <el-input v-model="currentProject.domain" />
        </el-form-item>
        <el-form-item label="项目目录">
          <el-input v-model="currentProject.path" />
        </el-form-item>
        <el-form-item label="运行端口" v-if="!['php', 'static-build'].includes(currentProject.type)">
          <el-input-number v-model="currentProject.port" :min="1024" :max="65535" />
        </el-form-item>
        <el-form-item label="构建步骤">
          <div class="build-steps compact">
            <div v-for="(step, index) in currentProject.buildSteps" :key="index" class="build-step">
              <el-input v-model="step.command" style="flex: 1" />
              <el-checkbox v-model="step.optional">可选</el-checkbox>
              <el-button text type="danger" @click="currentProject.buildSteps.splice(index, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" @click="currentProject.buildSteps.push({ command: '', optional: false })">
              <el-icon><Plus /></el-icon>添加步骤
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="启动命令" v-if="!['php', 'static-build'].includes(currentProject.type)">
          <el-input v-model="currentProject.startCommand" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer split">
          <el-button type="danger" @click="deleteProject">删除项目</el-button>
          <div>
            <el-button @click="showProjectSettings = false">取消</el-button>
            <el-button type="primary" @click="saveProjectSettings" :loading="saving">保存设置</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 项目目录浏览器对话框 -->
    <el-dialog v-model="showProjectPathBrowser" title="选择项目目录" width="520px" class="browser-dialog" destroy-on-close>
      <div class="path-browser">
        <div class="browser-breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item @click="browseProjectPath('/')" class="clickable">
              <el-icon><HomeFilled /></el-icon>
            </el-breadcrumb-item>
            <el-breadcrumb-item
              v-for="(part, index) in projectBrowserPathParts"
              :key="index"
              @click="browseProjectPathIndex(index)"
              class="clickable"
            >
              {{ part }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="browser-list" v-loading="projectBrowserLoading">
          <div class="browser-item parent" @click="browseProjectPathParent" v-if="projectBrowserPath !== '/'">
            <el-icon><ArrowLeft /></el-icon>
            <span>..</span>
          </div>
          <div 
            v-for="dir in projectBrowserDirs" 
            :key="dir.path"
            class="browser-item"
            @click="browseProjectPath(dir.path)"
            @dblclick="selectProjectPath(dir.path)"
          >
            <el-icon class="folder-icon"><Folder /></el-icon>
            <span>{{ dir.name }}</span>
          </div>
          <div v-if="projectBrowserDirs.length === 0 && !projectBrowserLoading" class="browser-empty">
            此目录下没有子文件夹
          </div>
        </div>
        <div class="browser-selected">
          <span>当前选择:</span>
          <code>{{ projectBrowserPath }}</code>
        </div>
      </div>
      <template #footer>
        <el-button @click="showProjectPathBrowser = false">取消</el-button>
        <el-button type="primary" @click="selectProjectPath(projectBrowserPath)">选择此目录</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useServerStore } from '@/stores/server'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Refresh, Lock, Delete, ArrowDown, Check, Promotion, Monitor, CopyDocument, 
  InfoFilled, Unlock, ArrowLeft, ArrowRight, FolderOpened, Folder, Document, Right, 
  HomeFilled, Top, Bottom, Edit, Link, Position, Close, Setting, Upload, UploadFilled,
  RefreshRight, SetUp, VideoPlay, View, CircleCheck, Cpu
} from '@element-plus/icons-vue'
import TechIcon from '@/components/icons/TechIcons.vue'

// 类型定义
interface Site {
  id: string
  name: string
  domain: string
  path: string
  type: string
  status: string
  ssl: boolean
  rewrite?: string
}

interface BuildStep {
  command: string
  optional?: boolean
}

interface EnvVar {
  key: string
  value: string
}

interface Project {
  id: string
  name: string
  type: string
  domain: string
  path: string
  port: number
  status: string
  ssl: boolean
  buildSteps: BuildStep[]
  startCommand: string
  outputDir?: string
  envVars: EnvVar[]
  processManager?: string
  lastDeploy?: number
  deploying?: boolean
}

interface DetectedProjectInfo {
  name?: string
  type?: string
  scripts?: string[]
  hasPackageJson?: boolean
  hasRequirements?: boolean
  hasGoMod?: boolean
  hasPomXml?: boolean
  hasComposerJson?: boolean
}

// Store
const serverStore = useServerStore()
const selectedServer = ref<string | null>(null)
const activeTab = ref('sites')
const loading = ref(false)
const creating = ref(false)
const saving = ref(false)
const loadingLogs = ref(false)

// 数据
const sites = ref<Site[]>([])
const projects = ref<Project[]>([])

// 对话框状态
const showAddStatic = ref(false)
const showAddProxy = ref(false)
const showAddProject = ref(false)
const showDeployLog = ref(false)
const showSiteSettings = ref(false)
const showProjectSettings = ref(false)
const currentSite = ref<Site | null>(null)
const currentProject = ref<Project | null>(null)
const deployLog = ref('')
const logContainer = ref<HTMLElement | null>(null)

// 表单数据
const newSite = ref({ name: '', domain: '', path: '/var/www', ssl: false })
const newProxy = ref({ name: '', domain: '', upstream: 'http://127.0.0.1:3000', websocket: false, ssl: false })
const newProject = ref<{
  name: string; type: string; domain: string; domainType: string; path: string; port: number; ssl: boolean;
  buildSteps: BuildStep[]; startCommand: string; outputDir: string; envVars: EnvVar[]; skipUpload: boolean;
  processManager: string
}>({
  name: '', type: 'nodejs', domain: '', domainType: 'domain', path: '/var/www', port: 3000, ssl: false,
  buildSteps: [{ command: 'npm install', optional: false }],
  startCommand: 'npm start', outputDir: 'dist', envVars: [], skipUpload: false, processManager: 'systemd'
})

// 项目类型配置
const projectTypes = [
  { value: 'nodejs', label: 'Node.js', desc: 'Express / Koa / NestJS', color: '#68a063' },
  { value: 'static-build', label: '静态构建', desc: 'Vue / React / Next.js', color: '#42b883' },
  { value: 'python', label: 'Python', desc: 'Flask / Django / FastAPI', color: '#3776ab' },
  { value: 'go', label: 'Go', desc: 'Gin / Echo / Fiber', color: '#00add8' },
  { value: 'java', label: 'Java', desc: 'Spring Boot', color: '#f89820' },
  { value: 'php', label: 'PHP', desc: 'Laravel / ThinkPHP', color: '#777bb4' }
]

// 进程管理器配置
const processManagers = [
  { value: 'systemd', label: 'Systemd', desc: '系统服务，开机自启', icon: '⚙️' },
  { value: 'pm2', label: 'PM2', desc: 'Node.js 进程管理器', icon: '🚀' },
  { value: 'supervisor', label: 'Supervisor', desc: 'Python 进程管理器', icon: '🐍' }
]

// 部署步骤
const deployStep = ref('basic')
const deploySteps = [
  { key: 'basic', title: '基本信息', desc: '项目名称和类型' },
  { key: 'upload', title: '上传代码', desc: '上传项目文件' },
  { key: 'domain', title: '域名设置', desc: '配置访问地址' },
  { key: 'deploy', title: '部署设置', desc: '构建和启动' },
  { key: 'ssl', title: 'SSL 证书', desc: 'HTTPS 加密' }
]
const deployStepIndex = computed(() => deploySteps.findIndex(s => s.key === deployStep.value))

// 上传相关
const uploadFiles = ref<{ name: string; path: string; size: number; isDir: boolean }[]>([])
const uploadProgress = ref(0)
const uploading = ref(false)
const uploadLog = ref('')
const selectedLocalPath = ref('')
const detectedProjectInfo = ref<DetectedProjectInfo | null>(null)

// 目录浏览器
const showProjectPathBrowser = ref(false)
const projectBrowserPath = ref('/var/www')
const projectBrowserDirs = ref<{ name: string; path: string; isDir: boolean }[]>([])
const projectBrowserLoading = ref(false)
const projectBrowserPathParts = computed(() => {
  if (!projectBrowserPath.value || projectBrowserPath.value === '/') return []
  return projectBrowserPath.value.split('/').filter(Boolean)
})

// 服务器 IP
const serverPublicIP = ref('')
const serverLocalIP = ref('')

// 计算属性
const connectedServers = computed(() => serverStore.connectedServers)
const hasMultipleServers = computed(() => serverStore.hasMultipleServers)

// 监听器
watch(selectedServer, (val) => {
  if (val) loadData()
})

onMounted(() => {
  if (connectedServers.value.length > 0) {
    selectedServer.value = serverStore.currentServerId || connectedServers.value[0].id
  }
  loadProjectsFromStorage()
})

// 数据加载
function loadProjectsFromStorage() {
  const saved = localStorage.getItem('serverhub_projects')
  if (saved) {
    try {
      projects.value = JSON.parse(saved)
    } catch { projects.value = [] }
  }
}

function saveProjectsToStorage() {
  localStorage.setItem('serverhub_projects', JSON.stringify(projects.value))
}

async function loadData() {
  await loadSites()
}

async function loadSites() {
  if (!selectedServer.value) return
  loading.value = true
  try {
    const result = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'ls -1 /etc/nginx/sites-enabled/ 2>/dev/null || ls -1 /etc/nginx/conf.d/*.conf 2>/dev/null']
    )
    const stdout = result.stdout || ''
    const files = stdout.trim().split('\n').filter(f => f && !f.includes('default'))
    sites.value = files.map((f, i) => ({
      id: `site_${i}`,
      name: f.replace('.conf', '').replace(/^.*\//, ''),
      domain: f.replace('.conf', '').replace(/^.*\//, ''),
      path: '/var/www/' + f.replace('.conf', '').replace(/^.*\//, ''),
      type: 'static',
      status: 'running',
      ssl: false
    }))
  } catch { sites.value = [] }
  finally { loading.value = false }
}

function refresh() { loadData() }

// 对话框处理
function handleAddCommand(cmd: string) {
  if (cmd === 'static') showAddStatic.value = true
  else if (cmd === 'proxy') showAddProxy.value = true
  else if (cmd === 'project') {
    resetNewProject()
    showAddProject.value = true
  }
}

function resetNewProject() {
  newProject.value = {
    name: '', type: 'nodejs', domain: '', domainType: 'domain', path: '/var/www', port: 3000, ssl: false,
    buildSteps: [{ command: 'npm install', optional: false }],
    startCommand: 'npm start', outputDir: 'dist', envVars: [], skipUpload: false, processManager: 'systemd'
  }
  deployStep.value = 'basic'
  uploadFiles.value = []
  selectedLocalPath.value = ''
  uploadProgress.value = 0
  uploadLog.value = ''
  detectedProjectInfo.value = null
  fetchServerIP()
}

// 项目类型选择
function selectProjectType(type: string) {
  newProject.value.type = type
  const defaults: Record<string, { buildSteps: BuildStep[]; startCommand: string; port: number; processManager: string }> = {
    nodejs: { buildSteps: [{ command: 'npm install', optional: false }], startCommand: 'npm start', port: 3000, processManager: 'pm2' },
    python: { buildSteps: [{ command: 'pip install -r requirements.txt', optional: false }], startCommand: 'python app.py', port: 5000, processManager: 'supervisor' },
    go: { buildSteps: [{ command: 'go build -o app', optional: false }], startCommand: './app', port: 8080, processManager: 'systemd' },
    java: { buildSteps: [{ command: 'mvn package -DskipTests', optional: false }], startCommand: 'java -jar target/*.jar', port: 8080, processManager: 'systemd' },
    php: { buildSteps: [{ command: 'composer install', optional: false }], startCommand: '', port: 0, processManager: 'systemd' },
    'static-build': { buildSteps: [{ command: 'npm install', optional: false }], startCommand: '', port: 0, processManager: 'systemd' }
  }
  const d = defaults[type] || defaults.nodejs
  newProject.value.buildSteps = d.buildSteps
  newProject.value.startCommand = d.startCommand
  newProject.value.port = d.port
  newProject.value.processManager = d.processManager
}

// 服务器 IP 获取
async function fetchServerIP() {
  if (!selectedServer.value) return
  try {
    const pubResult = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', 'curl -fsSL --connect-timeout 3 ifconfig.me 2>/dev/null || curl -fsSL --connect-timeout 3 ipinfo.io/ip 2>/dev/null']
    )
    serverPublicIP.value = (pubResult.stdout || '').trim()
    
    const localResult = await window.electronAPI.server.executeCommand(
      selectedServer.value, 'bash', ['-c', "hostname -I 2>/dev/null | awk '{print $1}'"]
    )
    serverLocalIP.value = (localResult.stdout || '').trim()
  } catch {
    serverPublicIP.value = '获取失败'
  }
}

function copyToClipboard(text: string) {
  if (!text) return
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

function getDomainPrefix(domain: string): string {
  if (!domain) return ''
  const parts = domain.split('.')
  if (parts.length > 2) return parts[0]
  return '@'
}

// 步骤导航
function prevDeployStep() {
  const idx = deployStepIndex.value
  if (idx > 0) deployStep.value = deploySteps[idx - 1].key
}

function nextDeployStep() {
  // 验证当前步骤
  if (deployStep.value === 'basic') {
    if (!newProject.value.name) {
      ElMessage.warning('请输入项目名称')
      return
    }
    if (!newProject.value.path || newProject.value.path === '/var/www') {
      newProject.value.path = '/var/www/' + newProject.value.name
    }
  } else if (deployStep.value === 'upload') {
    if (!newProject.value.skipUpload && !selectedLocalPath.value) {
      ElMessage.warning('请选择要上传的项目文件夹，或勾选跳过上传')
      return
    }
  } else if (deployStep.value === 'domain') {
    if (newProject.value.domainType === 'domain' && !newProject.value.domain) {
      ElMessage.warning('请输入域名')
      return
    }
    if (newProject.value.domainType === 'ip') {
      newProject.value.domain = serverPublicIP.value
    }
  }
  
  const idx = deployStepIndex.value
  if (idx < deploySteps.length - 1) deployStep.value = deploySteps[idx + 1].key
}

// 文件夹选择和扫描
async function selectFolder() {
  try {
    const result = await window.electronAPI.dialog.showOpenDialog({
      properties: ['openDirectory'],
      title: '选择项目文件夹'
    })
    
    if (result.canceled || !result.filePaths.length) return
    
    selectedLocalPath.value = result.filePaths[0]
    await scanFolder(selectedLocalPath.value)
    await detectProjectType(selectedLocalPath.value)
  } catch (e) {
    ElMessage.error('选择文件夹失败: ' + (e as Error).message)
  }
}

async function scanFolder(folderPath: string) {
  try {
    const files = await window.electronAPI.fs.scanDirectory(folderPath, {
      ignore: ['node_modules', '.git', '__pycache__', '.venv', 'venv', 'dist', 'build', '.next', '.nuxt', 'target', 'vendor']
    })
    uploadFiles.value = files
  } catch {
    uploadFiles.value = [{ name: folderPath.split(/[/\\]/).pop() || 'project', path: folderPath, size: 0, isDir: true }]
  }
}

// 项目类型检测
async function detectProjectType(folderPath: string) {
  try {
    const info: DetectedProjectInfo = {}
    
    // 尝试读取 package.json
    try {
      const pkgContent = await window.electronAPI.fs.readFile(folderPath + '/package.json')
      if (pkgContent) {
        const pkgStr = typeof pkgContent === 'string' ? pkgContent : pkgContent.toString()
        const pkg = JSON.parse(pkgStr)
        info.hasPackageJson = true
        info.name = pkg.name
        info.scripts = Object.keys(pkg.scripts || {})
        
        // 检测是否是静态构建项目
        const deps = { ...pkg.dependencies, ...pkg.devDependencies }
        if (deps.vue || deps.react || deps['next'] || deps.nuxt || deps.vite) {
          info.type = 'static-build'
        } else {
          info.type = 'nodejs'
        }
        
        // 自动设置构建步骤
        if (info.scripts) {
          const buildSteps: BuildStep[] = [{ command: 'npm install', optional: false }]
          if (info.scripts.includes('build')) {
            buildSteps.push({ command: 'npm run build', optional: true })
          }
          newProject.value.buildSteps = buildSteps
        }
        
        // 自动设置项目名称
        if (info.name && !newProject.value.name) {
          newProject.value.name = info.name
          newProject.value.path = '/var/www/' + info.name
        }
        
        // 自动设置项目类型
        if (info.type) {
          selectProjectType(info.type)
        }
      }
    } catch { /* 没有 package.json */ }
    
    // 检测 Python 项目
    try {
      await window.electronAPI.fs.readFile(folderPath + '/requirements.txt')
      info.hasRequirements = true
      if (!info.type) {
        info.type = 'python'
        selectProjectType('python')
      }
    } catch { /* 没有 requirements.txt */ }
    
    // 检测 Go 项目
    try {
      await window.electronAPI.fs.readFile(folderPath + '/go.mod')
      info.hasGoMod = true
      if (!info.type) {
        info.type = 'go'
        selectProjectType('go')
      }
    } catch { /* 没有 go.mod */ }
    
    // 检测 Java 项目
    try {
      await window.electronAPI.fs.readFile(folderPath + '/pom.xml')
      info.hasPomXml = true
      if (!info.type) {
        info.type = 'java'
        selectProjectType('java')
      }
    } catch { /* 没有 pom.xml */ }
    
    // 检测 PHP 项目
    try {
      await window.electronAPI.fs.readFile(folderPath + '/composer.json')
      info.hasComposerJson = true
      if (!info.type) {
        info.type = 'php'
        selectProjectType('php')
      }
    } catch { /* 没有 composer.json */ }
    
    detectedProjectInfo.value = Object.keys(info).length > 0 ? info : null
  } catch {
    detectedProjectInfo.value = null
  }
}

// 目录浏览器
async function browseProjectPath(path: string) {
  if (!selectedServer.value) return
  
  projectBrowserLoading.value = true
  projectBrowserPath.value = path
  
  try {
    const result = await window.electronAPI.file.list(selectedServer.value, path)
    projectBrowserDirs.value = result.files
      .filter((f: any) => f.is_dir)
      .map((f: any) => ({ name: f.name, path: f.path, isDir: true }))
      .sort((a: any, b: any) => a.name.localeCompare(b.name))
  } catch (e) {
    ElMessage.error('加载目录失败: ' + (e as Error).message)
    projectBrowserDirs.value = []
  } finally {
    projectBrowserLoading.value = false
  }
}

function browseProjectPathParent() {
  if (projectBrowserPath.value === '/') return
  const parts = projectBrowserPath.value.split('/').filter(Boolean)
  parts.pop()
  browseProjectPath('/' + parts.join('/'))
}

function browseProjectPathIndex(index: number) {
  const parts = projectBrowserPath.value.split('/').filter(Boolean)
  browseProjectPath('/' + parts.slice(0, index + 1).join('/'))
}

function selectProjectPath(path: string) {
  newProject.value.path = path
  showProjectPathBrowser.value = false
}

watch(showProjectPathBrowser, (val) => {
  if (val) browseProjectPath(newProject.value.path || '/var/www')
})

// 构建步骤管理
function addBuildStep() { 
  newProject.value.buildSteps.push({ command: '', optional: false }) 
}

function removeBuildStep(index: number) { 
  newProject.value.buildSteps.splice(index, 1) 
}

function addEnvVar() { 
  newProject.value.envVars.push({ key: '', value: '' }) 
}

function removeEnvVar(index: number) { 
  newProject.value.envVars.splice(index, 1) 
}

function isSecretKey(key: string): boolean {
  const secretPatterns = ['secret', 'password', 'token', 'key', 'api_key', 'apikey']
  return secretPatterns.some(p => key.toLowerCase().includes(p))
}

function getDefaultStartCommand(type: string): string {
  const cmds: Record<string, string> = {
    nodejs: 'npm start', python: 'python app.py', go: './app', java: 'java -jar target/*.jar'
  }
  return cmds[type] || ''
}

function getProjectTypeLabel(type: string): string {
  const pt = projectTypes.find(p => p.value === type)
  return pt ? pt.label : type
}

function getProcessManagerLabel(pm: string): string {
  const p = processManagers.find(m => m.value === pm)
  return p ? p.label : pm
}

// 创建站点
async function createStaticSite() {
  if (!selectedServer.value || !newSite.value.name || !newSite.value.domain) {
    ElMessage.warning('请填写完整信息')
    return
  }
  creating.value = true
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo mkdir -p ${newSite.value.path}`])
    const config = generateStaticConfig(newSite.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newSite.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newSite.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    ElMessage.success('站点创建成功')
    showAddStatic.value = false
    newSite.value = { name: '', domain: '', path: '/var/www', ssl: false }
    loadSites()
  } catch (e) { 
    ElMessage.error('创建失败: ' + (e as Error).message) 
  } finally { 
    creating.value = false 
  }
}

async function createProxySite() {
  if (!selectedServer.value || !newProxy.value.name || !newProxy.value.domain) {
    ElMessage.warning('请填写完整信息')
    return
  }
  creating.value = true
  try {
    const config = generateProxyConfig(newProxy.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newProxy.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newProxy.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    ElMessage.success('反向代理创建成功')
    showAddProxy.value = false
    newProxy.value = { name: '', domain: '', upstream: 'http://127.0.0.1:3000', websocket: false, ssl: false }
    loadSites()
  } catch (e) { 
    ElMessage.error('创建失败: ' + (e as Error).message) 
  } finally { 
    creating.value = false 
  }
}

// 创建项目
async function createProject() {
  if (!selectedServer.value || !newProject.value.name || !newProject.value.domain) {
    ElMessage.warning('请填写完整信息')
    return
  }
  creating.value = true
  try {
    // 创建项目目录
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo mkdir -p ${newProject.value.path}`])
    
    // 生成 Nginx 配置
    const config = generateProjectConfig(newProject.value)
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `echo '${config.replace(/'/g, "'\\''")}' | sudo tee /etc/nginx/sites-available/${newProject.value.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo ln -sf /etc/nginx/sites-available/${newProject.value.name} /etc/nginx/sites-enabled/`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo nginx -t && sudo systemctl reload nginx'])
    
    // 保存项目配置
    const project: Project = {
      id: `project_${Date.now()}`,
      name: newProject.value.name,
      type: newProject.value.type,
      domain: newProject.value.domain,
      path: newProject.value.path,
      port: newProject.value.port,
      status: 'stopped',
      ssl: newProject.value.ssl,
      buildSteps: [...newProject.value.buildSteps],
      startCommand: newProject.value.startCommand,
      outputDir: newProject.value.outputDir,
      envVars: [...newProject.value.envVars],
      processManager: newProject.value.processManager
    }
    projects.value.push(project)
    saveProjectsToStorage()
    
    ElMessage.success('项目创建成功')
    showAddProject.value = false
    activeTab.value = 'projects'
  } catch (e) { 
    ElMessage.error('创建失败: ' + (e as Error).message) 
  } finally { 
    creating.value = false 
  }
}

// 部署项目 - 优化版本，支持可选步骤
async function deployProject(project: Project) {
  if (!selectedServer.value) return
  project.deploying = true
  deployLog.value = `🚀 开始部署 ${project.name}...\n\n`
  showDeployLog.value = true
  currentProject.value = project

  try {
    // 执行构建步骤
    for (const step of project.buildSteps) {
      const cmd = step.command?.trim()
      if (!cmd) continue
      
      deployLog.value += `📦 执行: ${cmd}\n`
      await nextTick()
      scrollLogToBottom()

      const envStr = project.envVars.map(e => `${e.key}=${e.value}`).join(' ')
      const fullCmd = envStr ? `cd ${project.path} && ${envStr} ${cmd}` : `cd ${project.path} && ${cmd}`
      
      try {
        const result = await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', fullCmd])
        
        if (result.stdout) deployLog.value += result.stdout + '\n'
        if (result.stderr) deployLog.value += result.stderr + '\n'
        
        if (result.exit_code !== 0) {
          if (step.optional) {
            deployLog.value += `⚠️ 可选步骤失败，继续执行...\n\n`
          } else {
            deployLog.value += `\n❌ 步骤失败 (退出码: ${result.exit_code})\n`
            ElMessage.error('部署失败')
            project.deploying = false
            return
          }
        } else {
          deployLog.value += `✅ 完成\n\n`
        }
      } catch (e) {
        if (step.optional) {
          deployLog.value += `⚠️ 可选步骤出错: ${(e as Error).message}，继续执行...\n\n`
        } else {
          throw e
        }
      }
    }

    // 启动服务（非静态项目）
    if (!['php', 'static-build'].includes(project.type) && project.startCommand) {
      deployLog.value += `\n🔧 配置服务...\n`
      
      const pm = project.processManager || 'systemd'
      
      if (pm === 'pm2') {
        await startWithPM2(project)
      } else if (pm === 'supervisor') {
        await startWithSupervisor(project)
      } else {
        await startWithSystemd(project)
      }
      
      project.status = 'running'
    } else if (project.type === 'static-build') {
      deployLog.value += `\n📁 静态文件已部署到: ${project.path}/${project.outputDir || 'dist'}\n`
      project.status = 'running'
    }

    project.lastDeploy = Date.now()
    saveProjectsToStorage()
    deployLog.value += '\n✅ 部署成功！\n'
    ElMessage.success('部署成功')
  } catch (e) {
    deployLog.value += `\n❌ 错误: ${(e as Error).message}\n`
    ElMessage.error('部署失败')
  } finally {
    project.deploying = false
  }
}

// 进程管理器启动方法
async function startWithSystemd(project: Project) {
  const serviceName = `serverhub-${project.name}`
  const envStr = project.envVars.map(e => `Environment="${e.key}=${e.value}"`).join('\n')
  const serviceContent = `[Unit]
Description=${project.name}
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${project.path}
ExecStart=/bin/bash -c '${project.startCommand}'
Restart=on-failure
${envStr}

[Install]
WantedBy=multi-user.target`
  
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 
    `echo '${serviceContent.replace(/'/g, "'\\''")}' | sudo tee /etc/systemd/system/${serviceName}.service`])
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 
    `sudo systemctl daemon-reload && sudo systemctl enable ${serviceName} && sudo systemctl restart ${serviceName}`])
  
  deployLog.value += `✅ Systemd 服务已启动: ${serviceName}\n`
}

async function startWithPM2(project: Project) {
  // 先停止旧进程
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 
    `pm2 delete ${project.name} 2>/dev/null || true`])
  
  const envStr = project.envVars.map(e => `${e.key}="${e.value}"`).join(' ')
  const cmd = envStr 
    ? `cd ${project.path} && ${envStr} pm2 start --name ${project.name} -- ${project.startCommand}`
    : `cd ${project.path} && pm2 start --name ${project.name} -- ${project.startCommand}`
  
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', cmd])
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 'pm2 save'])
  
  deployLog.value += `✅ PM2 进程已启动: ${project.name}\n`
}

async function startWithSupervisor(project: Project) {
  const confName = `serverhub-${project.name}`
  const envStr = project.envVars.map(e => `${e.key}="${e.value}"`).join(',')
  const confContent = `[program:${confName}]
command=${project.startCommand}
directory=${project.path}
autostart=true
autorestart=true
${envStr ? `environment=${envStr}` : ''}
stdout_logfile=/var/log/supervisor/${confName}.log
stderr_logfile=/var/log/supervisor/${confName}.err.log`
  
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 
    `echo '${confContent.replace(/'/g, "'\\''")}' | sudo tee /etc/supervisor/conf.d/${confName}.conf`])
  await window.electronAPI.server.executeCommand(selectedServer.value!, 'bash', ['-c', 
    `sudo supervisorctl reread && sudo supervisorctl update && sudo supervisorctl restart ${confName}`])
  
  deployLog.value += `✅ Supervisor 进程已启动: ${confName}\n`
}

// 项目控制
async function startProject(project: Project) {
  if (!selectedServer.value) return
  try {
    const pm = project.processManager || 'systemd'
    if (pm === 'pm2') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `pm2 start ${project.name}`])
    } else if (pm === 'supervisor') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo supervisorctl start serverhub-${project.name}`])
    } else {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo systemctl start serverhub-${project.name}`])
    }
    project.status = 'running'
    saveProjectsToStorage()
    ElMessage.success('项目已启动')
  } catch (e) { 
    ElMessage.error('启动失败') 
  }
}

async function stopProject(project: Project) {
  if (!selectedServer.value) return
  try {
    const pm = project.processManager || 'systemd'
    if (pm === 'pm2') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `pm2 stop ${project.name}`])
    } else if (pm === 'supervisor') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo supervisorctl stop serverhub-${project.name}`])
    } else {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo systemctl stop serverhub-${project.name}`])
    }
    project.status = 'stopped'
    saveProjectsToStorage()
    ElMessage.success('项目已停止')
  } catch (e) { 
    ElMessage.error('停止失败') 
  }
}

function viewProjectLogs(project: Project) {
  currentProject.value = project
  deployLog.value = '加载日志中...'
  showDeployLog.value = true
  loadProjectLogs(project)
}

async function loadProjectLogs(project: Project) {
  if (!selectedServer.value) return
  loadingLogs.value = true
  try {
    const pm = project.processManager || 'systemd'
    let cmd = ''
    if (pm === 'pm2') {
      cmd = `pm2 logs ${project.name} --lines 100 --nostream 2>/dev/null || echo "无日志"`
    } else if (pm === 'supervisor') {
      cmd = `sudo tail -n 100 /var/log/supervisor/serverhub-${project.name}.log 2>/dev/null || echo "无日志"`
    } else {
      cmd = `sudo journalctl -u serverhub-${project.name} -n 100 --no-pager 2>/dev/null || echo "无日志"`
    }
    const result = await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', cmd])
    deployLog.value = result.stdout || '无日志'
  } catch { 
    deployLog.value = '获取日志失败' 
  } finally {
    loadingLogs.value = false
  }
}

function editProject(project: Project) {
  currentProject.value = { ...project, buildSteps: [...project.buildSteps], envVars: [...project.envVars] }
  showProjectSettings.value = true
}

async function saveProjectSettings() {
  if (!currentProject.value) return
  saving.value = true
  try {
    const index = projects.value.findIndex(p => p.id === currentProject.value!.id)
    if (index !== -1) {
      projects.value[index] = { ...currentProject.value }
      saveProjectsToStorage()
    }
    ElMessage.success('设置已保存')
    showProjectSettings.value = false
  } finally { 
    saving.value = false 
  }
}

async function deleteProject() {
  if (!currentProject.value || !selectedServer.value) return
  try {
    await ElMessageBox.confirm(`确定删除项目 ${currentProject.value.name}？`, '确认删除', { type: 'warning' })
  } catch { return }
  
  try {
    const pm = currentProject.value.processManager || 'systemd'
    const name = currentProject.value.name
    
    if (pm === 'pm2') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `pm2 delete ${name} 2>/dev/null || true; pm2 save`])
    } else if (pm === 'supervisor') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
        `sudo supervisorctl stop serverhub-${name} || true; sudo rm -f /etc/supervisor/conf.d/serverhub-${name}.conf; sudo supervisorctl reread; sudo supervisorctl update`])
    } else {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
        `sudo systemctl stop serverhub-${name} || true; sudo systemctl disable serverhub-${name} || true; sudo rm -f /etc/systemd/system/serverhub-${name}.service`])
    }
    
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo rm -f /etc/nginx/sites-enabled/${name} /etc/nginx/sites-available/${name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    
    projects.value = projects.value.filter(p => p.id !== currentProject.value!.id)
    saveProjectsToStorage()
    showProjectSettings.value = false
    ElMessage.success('项目已删除')
  } catch (e) { 
    ElMessage.error('删除失败') 
  }
}

// 站点操作
function editSite(site: Site) {
  currentSite.value = { ...site }
  showSiteSettings.value = true
}

async function saveSiteSettings() {
  if (!currentSite.value || !selectedServer.value) return
  saving.value = true
  try {
    ElMessage.success('设置已保存')
    showSiteSettings.value = false
    loadSites()
  } finally { 
    saving.value = false 
  }
}

function openSite(site: Site) {
  const url = site.ssl ? `https://${site.domain}` : `http://${site.domain}`
  window.electronAPI.shell.openExternal(url)
}

async function toggleSite(site: Site, action: 'start' | 'stop') {
  if (!selectedServer.value) return
  try {
    if (action === 'stop') {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo rm -f /etc/nginx/sites-enabled/${site.name}`])
    } else {
      await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', `sudo ln -sf /etc/nginx/sites-available/${site.name} /etc/nginx/sites-enabled/`])
    }
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    ElMessage.success(action === 'stop' ? '站点已停止' : '站点已启动')
    loadSites()
  } catch { 
    ElMessage.error('操作失败') 
  }
}

async function deleteSite(site: Site) {
  try { 
    await ElMessageBox.confirm(`确定删除站点 ${site.name}？`, '确认删除', { type: 'warning' }) 
  } catch { return }
  
  if (!selectedServer.value) return
  try {
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 
      `sudo rm -f /etc/nginx/sites-enabled/${site.name} /etc/nginx/sites-available/${site.name}`])
    await window.electronAPI.server.executeCommand(selectedServer.value, 'bash', ['-c', 'sudo systemctl reload nginx'])
    ElMessage.success('站点已删除')
    loadSites()
  } catch { 
    ElMessage.error('删除失败') 
  }
}

function applyRewrite(preset: string) {
  if (!currentSite.value) return
  const presets: Record<string, string> = {
    vue: 'location / { try_files $uri $uri/ /index.html; }',
    laravel: 'location / { try_files $uri $uri/ /index.php?$query_string; }',
    wordpress: 'location / { try_files $uri $uri/ /index.php?$args; }'
  }
  currentSite.value.rewrite = presets[preset] || ''
}

// 日志操作
function scrollLogToTop() {
  if (logContainer.value) logContainer.value.scrollTop = 0
}

function scrollLogToBottom() {
  if (logContainer.value) logContainer.value.scrollTop = logContainer.value.scrollHeight
}

function copyLog() {
  navigator.clipboard.writeText(deployLog.value)
  ElMessage.success('日志已复制')
}

// Nginx 配置生成
function generateStaticConfig(site: { name: string; domain: string; path: string; ssl: boolean }): string {
  return `server {
    listen 80;
    server_name ${site.domain};
    root ${site.path};
    index index.html index.htm;
    
    location / {
        try_files $uri $uri/ =404;
    }
    
    location ~ /\\. { deny all; }
}`
}

function generateProxyConfig(proxy: { name: string; domain: string; upstream: string; websocket: boolean; ssl: boolean }): string {
  return `server {
    listen 80;
    server_name ${proxy.domain};
    
    location / {
        proxy_pass ${proxy.upstream};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        ${proxy.websocket ? `proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";` : ''}
    }
}`
}

function generateProjectConfig(project: { name: string; domain: string; path: string; port: number; type: string; outputDir?: string }): string {
  if (project.type === 'static-build') {
    return `server {
    listen 80;
    server_name ${project.domain};
    root ${project.path}/${project.outputDir || 'dist'};
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location ~ /\\. { deny all; }
}`
  }
  if (project.type === 'php') {
    return `server {
    listen 80;
    server_name ${project.domain};
    root ${project.path}/public;
    index index.php index.html;
    
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    
    location ~ \\.php$ {
        fastcgi_pass unix:/var/run/php/php8.2-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    
    location ~ /\\. { deny all; }
}`
  }
  return `server {
    listen 80;
    server_name ${project.domain};
    
    location / {
        proxy_pass http://127.0.0.1:${project.port};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}`
}

// 辅助函数
function getTypeTag(type: string): 'success' | 'warning' | 'info' | 'danger' | undefined {
  const map: Record<string, 'success' | 'warning' | 'info' | 'danger' | undefined> = {
    static: undefined, php: 'warning', node: 'success', python: 'info', java: 'danger', proxy: undefined
  }
  return map[type]
}

function getTypeLabel(type: string): string {
  const labels: Record<string, string> = { static: '静态', php: 'PHP', node: 'Node', python: 'Python', java: 'Java', proxy: '代理' }
  return labels[type] || type
}

function getProjectColor(type: string): string {
  const colors: Record<string, string> = {
    nodejs: '#68a063', python: '#3776ab', go: '#00add8', java: '#f89820', php: '#777bb4', 'static-build': '#42b883'
  }
  return colors[type] || '#6366f1'
}

function getProjectStatusType(status: string): 'success' | 'danger' | 'info' {
  return status === 'running' ? 'success' : status === 'error' ? 'danger' : 'info'
}

function getProjectStatusLabel(status: string): string {
  const labels: Record<string, string> = { running: '运行中', stopped: '已停止', error: '错误' }
  return labels[status] || status
}

function formatTime(ts: number): string {
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style lang="scss" scoped>
.websites {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;

  .header-left {
    h1 { font-size: 22px; font-weight: 600; margin-bottom: 4px; }
    .subtitle { color: var(--text-secondary); font-size: 13px; }
  }

  .header-actions {
    display: flex;
    gap: 10px;
    align-items: center;
  }
}

.empty-state { padding: 80px 0; }

.main-tabs { margin-bottom: 16px; }
.tab-label { display: flex; align-items: center; gap: 8px; }

.tab-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 20px;
}

// 数据表格
.data-table {
  .cell-name {
    display: flex;
    align-items: center;
    gap: 10px;

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      &.running { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.5); }
      &.stopped { background: #ef4444; }
    }
  }

  .domain-link {
    color: var(--primary-color);
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 4px;
    &:hover { text-decoration: underline; }
  }

  .mono {
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    font-size: 12px;
    background: var(--bg-tertiary);
    padding: 3px 8px;
    border-radius: 4px;
  }
}

// 项目卡片
.empty-projects { padding: 60px 0; }

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 20px;
}

.project-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  transition: all 0.2s;

  &:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .project-header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 16px;

    .project-icon {
      width: 44px;
      height: 44px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      :deep(svg) { width: 26px; height: 26px; }
    }

    .project-info {
      flex: 1;
      .project-name { font-weight: 600; font-size: 15px; }
      .project-domain { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
    }
  }

  .project-meta {
    margin-bottom: 16px;
    padding: 12px;
    background: var(--bg-secondary);
    border-radius: 8px;
    
    .meta-item {
      font-size: 12px;
      color: var(--text-secondary);
      margin-bottom: 6px;
      display: flex;
      align-items: center;
      gap: 8px;
      &:last-child { margin-bottom: 0; }
      .meta-label { color: var(--text-color); min-width: 60px; }
      code { 
        background: var(--bg-tertiary); 
        padding: 2px 6px; 
        border-radius: 4px; 
        font-size: 11px;
        font-family: 'JetBrains Mono', monospace;
      }
    }
  }

  .project-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
}

// 对话框通用样式
:deep(.site-dialog) {
  .el-dialog {
    background: var(--bg-secondary) !important;
    border-radius: 12px;
    overflow: hidden;
  }
  .el-dialog__header {
    background: var(--bg-tertiary);
    padding: 16px 20px;
    margin: 0;
    border-bottom: 1px solid var(--border-color);
  }
  .el-dialog__title { font-weight: 600; }
  .el-dialog__body { padding: 24px; }
  .el-dialog__footer { 
    padding: 16px 24px; 
    border-top: 1px solid var(--border-color);
    background: var(--bg-tertiary);
  }
}

.site-form {
  .form-tip {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 6px;
  }
  
  .switch-label {
    margin-left: 10px;
    font-size: 13px;
    color: var(--text-secondary);
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  
  &.split {
    justify-content: space-between;
  }
}

// 部署向导对话框
:deep(.deploy-wizard-dialog) {
  .el-dialog {
    background: var(--bg-secondary) !important;
    border-radius: 16px;
    overflow: hidden;
  }
  .el-dialog__header { display: none; }
  .el-dialog__body { padding: 0; }
}

.wizard-container {
  display: flex;
  flex-direction: column;
  height: 680px;
}

.wizard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, var(--primary-color) 0%, #818cf8 100%);
  
  .wizard-title {
    display: flex;
    align-items: center;
    gap: 12px;
    color: #fff;
    font-size: 18px;
    font-weight: 600;
    
    .title-icon { font-size: 24px; }
  }
  
  .close-btn {
    color: rgba(255, 255, 255, 0.8);
    &:hover { color: #fff; background: rgba(255, 255, 255, 0.1); }
  }
}

.wizard-steps {
  display: flex;
  padding: 20px 24px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
}

.wizard-step {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 10px;
  cursor: default;
  transition: all 0.2s;
  
  &.clickable { cursor: pointer; }
  &.clickable:hover { background: var(--bg-secondary); }
  
  &.active {
    background: var(--primary-color);
    .step-title { color: #fff; }
    .step-desc { color: rgba(255, 255, 255, 0.7); }
    .step-indicator { background: rgba(255, 255, 255, 0.2); color: #fff; }
  }
  
  &.completed .step-indicator {
    background: #22c55e;
    color: #fff;
  }
  
  .step-indicator {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: var(--bg-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    font-weight: 600;
    flex-shrink: 0;
  }
  
  .step-info {
    flex: 1;
    min-width: 0;
  }
  
  .step-title {
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .step-desc {
    font-size: 11px;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.wizard-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.step-panel {
  .panel-header {
    margin-bottom: 24px;
    
    h3 {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 6px;
      
      .el-icon { color: var(--primary-color); }
    }
    
    p {
      color: var(--text-secondary);
      font-size: 13px;
    }
  }
}

.wizard-form {
  .form-tip {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 6px;
    display: flex;
    align-items: center;
    gap: 4px;
  }
}

// 项目类型选择器
.type-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.type-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 2px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  
  &:hover { border-color: var(--primary-color); }
  
  &.active {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.08);
  }
  
  .type-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    :deep(svg) { width: 24px; height: 24px; }
  }
  
  .type-info {
    flex: 1;
    min-width: 0;
  }
  
  .type-name {
    font-weight: 500;
    font-size: 14px;
  }
  
  .type-desc {
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 2px;
  }
  
  .type-check {
    position: absolute;
    top: 8px;
    right: 8px;
    color: var(--primary-color);
    font-size: 18px;
  }
}

.path-input-group {
  display: flex;
  gap: 10px;
  
  .el-input { flex: 1; }
}

// 上传区域
.upload-area {
  .upload-dropzone {
    border: 2px dashed var(--border-color);
    border-radius: 12px;
    padding: 60px 40px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s;
    
    &:hover {
      border-color: var(--primary-color);
      background: rgba(99, 102, 241, 0.05);
    }
    
    .dropzone-icon {
      font-size: 56px;
      color: var(--text-secondary);
      margin-bottom: 16px;
    }
    
    .dropzone-title {
      font-size: 16px;
      font-weight: 500;
      margin-bottom: 8px;
    }
    
    .dropzone-hint {
      font-size: 13px;
      color: var(--text-secondary);
    }
  }
  
  .upload-preview {
    .preview-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;
      background: var(--bg-tertiary);
      border-radius: 10px;
      margin-bottom: 16px;
      
      .preview-path {
        display: flex;
        align-items: center;
        gap: 10px;
        font-family: 'JetBrains Mono', monospace;
        font-size: 13px;
      }
    }
    
    .detected-info {
      background: rgba(99, 102, 241, 0.08);
      border: 1px solid rgba(99, 102, 241, 0.2);
      border-radius: 10px;
      margin-bottom: 16px;
      
      .info-header {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 12px 16px;
        border-bottom: 1px solid rgba(99, 102, 241, 0.2);
        font-weight: 500;
        color: var(--primary-color);
      }
      
      .info-content {
        padding: 12px 16px;
      }
      
      .info-item {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 8px;
        font-size: 13px;
        &:last-child { margin-bottom: 0; }
        
        .info-label {
          color: var(--text-secondary);
          min-width: 70px;
        }
      }
      
      .script-tags {
        display: flex;
        gap: 6px;
        flex-wrap: wrap;
      }
    }
    
    .file-list-panel {
      background: var(--bg-tertiary);
      border-radius: 10px;
      margin-bottom: 16px;
      
      .list-header {
        display: flex;
        justify-content: space-between;
        padding: 12px 16px;
        border-bottom: 1px solid var(--border-color);
        font-size: 13px;
        
        .file-count { color: var(--text-secondary); }
      }
      
      .file-list {
        padding: 8px 16px;
        max-height: 180px;
        overflow-y: auto;
      }
      
      .file-row {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 8px 0;
        font-size: 13px;
        
        .file-icon { 
          color: var(--text-secondary);
          &.folder { color: #f0b429; }
        }
        
        .file-name {
          flex: 1;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        
        .file-size {
          color: var(--text-secondary);
          font-size: 12px;
        }
      }
      
      .file-more {
        padding: 8px 0;
        color: var(--text-secondary);
        font-size: 12px;
      }
    }
    
    .upload-target {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 14px 16px;
      background: rgba(34, 197, 94, 0.1);
      border: 1px solid rgba(34, 197, 94, 0.2);
      border-radius: 10px;
      font-size: 13px;
      margin-bottom: 16px;
      
      code {
        font-family: 'JetBrains Mono', monospace;
        color: #22c55e;
      }
    }
    
    .upload-progress {
      .progress-text {
        margin-top: 10px;
        font-size: 12px;
        color: var(--text-secondary);
        text-align: center;
      }
    }
  }
  
  .skip-upload {
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px solid var(--border-color);
    
    .skip-hint {
      color: var(--text-secondary);
      font-size: 12px;
    }
  }
}

// 服务器信息卡片
.server-info-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin-bottom: 24px;
  
  .card-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color);
    font-weight: 500;
  }
  
  .card-body {
    padding: 14px 16px;
  }
  
  .info-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 10px;
    &:last-child { margin-bottom: 0; }
    
    .info-label {
      font-size: 13px;
      color: var(--text-secondary);
      min-width: 60px;
    }
    
    .info-value {
      font-family: 'JetBrains Mono', monospace;
      font-size: 14px;
      background: var(--bg-secondary);
      padding: 6px 12px;
      border-radius: 6px;
      color: var(--primary-color);
      
      &.secondary { color: var(--text-secondary); }
    }
  }
}

// 访问方式卡片
.access-type-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.access-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  
  &:hover { border-color: var(--primary-color); }
  
  &.active {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.08);
  }
  
  .card-icon {
    font-size: 28px;
    color: var(--text-secondary);
  }
  
  &.active .card-icon { color: var(--primary-color); }
  
  .card-content {
    flex: 1;
  }
  
  .card-title {
    font-weight: 500;
    margin-bottom: 4px;
  }
  
  .card-desc {
    font-size: 12px;
    color: var(--text-secondary);
  }
  
  .card-check {
    position: absolute;
    top: 10px;
    right: 10px;
    color: var(--primary-color);
    font-size: 20px;
  }
}

.readonly-input {
  :deep(.el-input__inner) {
    background: var(--bg-tertiary);
  }
}

// DNS 指引
.dns-guide {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin-top: 20px;
  
  .guide-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color);
    font-weight: 500;
    color: var(--primary-color);
  }
  
  .guide-content {
    padding: 16px;
  }
  
  .guide-step {
    display: flex;
    gap: 14px;
    margin-bottom: 18px;
    &:last-child { margin-bottom: 0; }
    
    .step-num {
      width: 26px;
      height: 26px;
      border-radius: 50%;
      background: var(--primary-color);
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      font-weight: 600;
      flex-shrink: 0;
    }
    
    .step-content {
      flex: 1;
    }
    
    .step-title {
      font-weight: 500;
      font-size: 13px;
      margin-bottom: 4px;
    }
    
    .step-desc {
      font-size: 12px;
      color: var(--text-secondary);
      
      code {
        background: var(--bg-secondary);
        padding: 2px 8px;
        border-radius: 4px;
        font-family: 'JetBrains Mono', monospace;
        color: var(--primary-color);
      }
    }
  }
}

// 进程管理器选择
.pm-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.pm-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 2px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover { border-color: var(--primary-color); }
  
  &.active {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.08);
  }
  
  .pm-icon {
    font-size: 24px;
  }
  
  .pm-info {
    flex: 1;
  }
  
  .pm-name {
    font-weight: 500;
    font-size: 14px;
  }
  
  .pm-desc {
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 2px;
  }
}

// 构建步骤
.build-steps {
  &.compact .build-step {
    margin-bottom: 8px;
  }
  
  .build-step {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
    
    .step-num {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      background: var(--bg-tertiary);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      color: var(--text-secondary);
      flex-shrink: 0;
    }
    
    .step-input {
      flex: 1;
    }
    
    .step-optional {
      flex-shrink: 0;
      font-size: 12px;
    }
  }
  
  .add-step-btn {
    margin-top: 8px;
  }
}

// 环境变量
.env-vars {
  .env-row {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 10px;
    
    .env-key { width: 140px; flex-shrink: 0; }
    .env-eq { color: var(--text-secondary); }
    .env-value { flex: 1; }
  }
}

// SSL 卡片
.ssl-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.ssl-card {
  padding: 30px;
  border: 2px solid var(--border-color);
  border-radius: 14px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover { border-color: var(--primary-color); }
  
  &.active {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.08);
    
    .ssl-icon { color: var(--primary-color); }
  }
  
  .ssl-icon {
    font-size: 40px;
    color: var(--text-secondary);
    margin-bottom: 14px;
  }
  
  .ssl-info {
    .ssl-title {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 6px;
    }
    
    .ssl-desc {
      font-size: 13px;
      color: var(--text-secondary);
    }
  }
}

.ssl-notice {
  margin-top: 24px;
  
  .notice-content {
    margin-top: 10px;
    font-size: 13px;
    color: var(--text-secondary);
    
    ul {
      margin: 8px 0 0 20px;
      padding: 0;
      
      li { margin-bottom: 4px; }
    }
  }
}

// 部署预览
.deploy-preview {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin-top: 24px;
  
  .preview-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color);
    font-weight: 500;
  }
  
  .preview-content {
    padding: 16px;
  }
  
  .preview-item {
    display: flex;
    align-items: center;
    margin-bottom: 12px;
    font-size: 13px;
    &:last-child { margin-bottom: 0; }
    
    .preview-label {
      color: var(--text-secondary);
      min-width: 80px;
    }
    
    .preview-value {
      font-weight: 500;
    }
    
    code.preview-value {
      font-family: 'JetBrains Mono', monospace;
      background: var(--bg-secondary);
      padding: 4px 10px;
      border-radius: 4px;
    }
  }
}

// 向导底部
.wizard-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-tertiary);
  
  .footer-right {
    display: flex;
    gap: 10px;
  }
}

// 日志对话框
:deep(.log-dialog) {
  .el-dialog {
    background: var(--bg-secondary) !important;
    border-radius: 12px;
  }
  .el-dialog__header {
    background: var(--bg-tertiary);
    padding: 16px 20px;
    margin: 0;
    border-bottom: 1px solid var(--border-color);
  }
  .el-dialog__body { padding: 0; }
  .el-dialog__footer { 
    padding: 16px 20px; 
    border-top: 1px solid var(--border-color);
    background: var(--bg-tertiary);
  }
}

.deploy-log-container {
  .log-toolbar {
    display: flex;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-bottom: 1px solid var(--border-color);
  }
  
  .log-content {
    background: #0d1117;
    padding: 16px;
    max-height: 500px;
    overflow: auto;
    
    pre {
      margin: 0;
      font-size: 13px;
      color: #c9d1d9;
      white-space: pre-wrap;
      word-break: break-all;
      font-family: 'JetBrains Mono', 'Consolas', monospace;
      line-height: 1.6;
    }
  }
}

// 目录浏览器
:deep(.browser-dialog) {
  .el-dialog {
    background: var(--bg-secondary) !important;
    border-radius: 12px;
  }
  .el-dialog__header {
    background: var(--bg-tertiary);
    padding: 16px 20px;
    margin: 0;
    border-bottom: 1px solid var(--border-color);
  }
  .el-dialog__body { padding: 20px; }
  .el-dialog__footer { 
    padding: 16px 20px; 
    border-top: 1px solid var(--border-color);
    background: var(--bg-tertiary);
  }
}

.path-browser {
  .browser-breadcrumb {
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: 8px;
    margin-bottom: 12px;
    
    .clickable {
      cursor: pointer;
      &:hover { color: var(--primary-color); }
    }
  }
  
  .browser-list {
    border: 1px solid var(--border-color);
    border-radius: 8px;
    max-height: 300px;
    overflow-y: auto;
    min-height: 200px;
    
    .browser-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 16px;
      cursor: pointer;
      transition: background 0.15s;
      
      &:hover { background: var(--bg-tertiary); }
      
      &.parent {
        color: var(--text-secondary);
        border-bottom: 1px solid var(--border-color);
      }
      
      .folder-icon { color: #f0b429; }
    }
    
    .browser-empty {
      padding: 50px;
      text-align: center;
      color: var(--text-secondary);
    }
  }
  
  .browser-selected {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: 8px;
    margin-top: 12px;
    font-size: 13px;
    
    code {
      font-family: 'JetBrains Mono', monospace;
      color: var(--primary-color);
    }
  }
}

// 伪静态预设
.rewrite-presets {
  margin-bottom: 10px;
  display: flex;
  gap: 8px;
}

.code-textarea {
  :deep(.el-textarea__inner) {
    font-family: 'JetBrains Mono', 'Consolas', monospace;
    font-size: 12px;
    background: var(--bg-tertiary);
  }
}
</style>
