/**
 * Agent 定义和管理
 * Agent = 专业 prompt + 工具子集
 */

import { app } from 'electron'
import { join } from 'path'
import { readFileSync, writeFileSync } from 'fs'

export interface AgentDefinition {
  id: string
  name: string
  icon: string
  description: string
  systemPrompt: string
  /** 允许使用的工具名称列表，空数组 = 使用所有工具 */
  tools: string[]
  /** 是否为内置 Agent */
  builtin: boolean
}

const builtinAgents: AgentDefinition[] = [
  {
    id: 'general',
    name: '通用助手',
    icon: 'agent-general',
    description: '通用服务器运维助手，可使用所有工具',
    systemPrompt: `你是 Runixo AI 助手，一个专业的服务器运维助手。
重要规则：
1. 主动使用工具获取真实数据，不要猜测
2. 每次调用工具后，必须用自然语言解释结果
3. 简洁专业，避免冗长`,
    tools: [],
    builtin: true
  },
  {
    id: 'diagnostics',
    name: '故障诊断',
    icon: 'agent-diagnostics',
    description: '专注系统故障诊断和性能分析',
    systemPrompt: `你是一个专业的服务器故障诊断专家。你的工作流程：
1. 先收集系统信息（CPU、内存、磁盘、网络）
2. 检查系统日志中的错误
3. 分析进程和服务状态
4. 给出诊断结论和修复建议
每一步都要使用工具获取真实数据。`,
    tools: ['get_system_info', 'execute_command', 'list_processes', 'list_services', 'diagnose_system', 'analyze_logs', 'check_port'],
    builtin: true
  },
  {
    id: 'docker',
    name: 'Docker 管理',
    icon: 'agent-docker',
    description: '专注 Docker 容器和镜像管理',
    systemPrompt: `你是 Docker 容器管理专家。你可以：
- 列出和管理容器（启动、停止、重启、删除）
- 管理镜像（拉取、删除）
- 查看容器日志
- 管理网络和卷
先列出当前容器状态，再执行用户请求的操作。`,
    tools: ['list_containers', 'container_action', 'container_logs', 'list_images', 'pull_image', 'remove_image', 'list_networks', 'list_volumes', 'execute_command'],
    builtin: true
  },
  {
    id: 'security',
    name: '安全审计',
    icon: 'agent-security',
    description: '专注服务器安全检查和加固',
    systemPrompt: `你是服务器安全审计专家。你的检查清单：
1. 检查用户和权限配置
2. 检查开放端口和网络连接
3. 检查系统更新状态
4. 检查 SSH 配置安全性
5. 检查防火墙规则
给出安全评分和加固建议。`,
    tools: ['list_users', 'check_user_activity', 'execute_command', 'security_scan', 'check_port', 'list_services'],
    builtin: true
  },
  {
    id: 'deploy',
    name: '应用部署',
    icon: 'agent-deploy',
    description: '专注应用部署和环境配置',
    systemPrompt: `你是应用部署专家。你可以：
- 检查和安装运行环境
- 部署 Web 应用（Node.js, Python, PHP）
- 配置 Nginx 反向代理
- 管理系统服务
按步骤执行，每步确认成功后再继续。`,
    tools: ['check_environment', 'install_software', 'manage_service', 'execute_command', 'deploy_application', 'create_nginx_config', 'write_file', 'read_file'],
    builtin: true
  }
]

export class AgentManager {
  private configPath: string
  private customAgents: AgentDefinition[] = []

  constructor() {
    this.configPath = join(app.getPath('userData'), 'agents.json')
    try {
      this.customAgents = JSON.parse(readFileSync(this.configPath, 'utf-8'))
    } catch {}
  }

  getAll(): AgentDefinition[] {
    return [...builtinAgents, ...this.customAgents]
  }

  get(id: string): AgentDefinition | undefined {
    return this.getAll().find(a => a.id === id)
  }

  addCustom(agent: Omit<AgentDefinition, 'builtin'>): void {
    this.customAgents.push({ ...agent, builtin: false })
    this.save()
  }

  removeCustom(id: string): void {
    this.customAgents = this.customAgents.filter(a => a.id !== id)
    this.save()
  }

  private save(): void {
    try { writeFileSync(this.configPath, JSON.stringify(this.customAgents, null, 2)) } catch {}
  }
}

export const agentManager = new AgentManager()
