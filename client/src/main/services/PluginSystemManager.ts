import { PluginMarketService } from './PluginMarketService'
import { WorkflowEngine } from './WorkflowEngine'
import { AgentService } from './AgentService'
import { DependencyResolver } from './DependencyResolver'
import { PluginLoader } from './PluginLoader'
import { PluginStorageService } from './PluginStorageService'
import { PluginHttpService } from './PluginHttpService'
import type { PluginContext } from '@runixo/plugin-sdk'

/**
 * 插件系统管理器
 * 整合所有插件相关服务
 */
export class PluginSystemManager {
  public market: PluginMarketService
  public workflow: WorkflowEngine
  public agent: AgentService
  public dependency: DependencyResolver
  public loader: PluginLoader
  public storage: PluginStorageService
  public http: PluginHttpService

  private plugins = new Map<string, any>()

  constructor() {
    this.market = new PluginMarketService()
    this.workflow = new WorkflowEngine()
    this.agent = new AgentService()
    this.dependency = new DependencyResolver()
    this.loader = new PluginLoader(this)
    this.storage = new PluginStorageService()
    this.http = new PluginHttpService()
  }

  async initialize() {
    await this.market.initialize()
    await this.storage.initialize()
    await this.loader.initialize()
    this.setupEventHandlers()
  }

  /**
   * 加载插件
   */
  async loadPlugin(pluginId: string): Promise<void> {
    await this.loader.loadPlugin(pluginId)
  }

  /**
   * 卸载插件
   */
  async unloadPlugin(pluginId: string): Promise<void> {
    await this.loader.unloadPlugin(pluginId)
  }

  /**
   * 创建插件上下文
   */
  createPluginContext(pluginId: string, metadata: any, config: any): PluginContext {
    return {
      pluginId,
      metadata,
      config,
      
      storage: this.storage.createStorageAPI(pluginId),
      secureStorage: this.storage.createSecureStorageAPI(pluginId),
      http: this.http.createHttpAPI(),
      ui: this.createUIAPI(),
      server: this.createServerAPI(),
      file: this.createFileAPI(),
      events: this.createEventAPI(),
      
      agent: {
        registerTool: (tool) => this.workflow.registerTool(tool),
        registerAgent: (agent) => this.agent.registerAgent(agent),
        registerWorkflow: (workflow) => this.workflow.registerWorkflow(workflow),
        registerPromptTemplate: (template) => this.agent.registerPromptTemplate(template),
        chat: (prompt, options) => this.agent.chat(pluginId, prompt, options),
        executeWorkflow: (id, inputs) => this.workflow.execute(id, inputs),
        renderPrompt: (id, vars) => this.agent.renderPrompt(id, vars),
        listAgents: () => this.agent.listAgents(),
        listWorkflows: () => [],
        listPromptTemplates: () => this.agent.listPromptTemplates()
      },
      
      tools: this.createToolsAPI(),
      menus: this.createMenusAPI(),
      routes: this.createRoutesAPI(),
      commands: this.createCommandsAPI(),
      logger: this.createLoggerAPI(pluginId)
    }
  }

  private async loadPluginInstance(pluginId: string) {
    // 由 PluginLoader 处理
  }

  private setupEventHandlers() {
    this.market.on('install:complete', (pluginId) => {
      this.loadPlugin(pluginId)
    })

    this.market.on('uninstall:complete', (pluginId) => {
      this.unloadPlugin(pluginId)
    })
  }

  // API 工厂方法已移至各服务类

  private createUIAPI() {
    return {
      showNotification: (title: string, message: string, type?: any) => {},
      showDialog: async (options: any) => ({ response: 0 }),
      showMessage: (message: string, type?: any) => {}
    }
  }

  private createServerAPI() {
    return {
      execute: async (serverId: string, command: string, options?: any) => ({ stdout: '', stderr: '', exitCode: 0 }),
      getSystemInfo: async (serverId: string) => ({} as any),
      listServers: async () => []
    }
  }

  private createFileAPI() {
    return {
      read: async (serverId: string, path: string) => '',
      write: async (serverId: string, path: string, content: string) => {},
      exists: async (serverId: string, path: string) => false,
      delete: async (serverId: string, path: string) => {},
      list: async (serverId: string, path: string) => []
    }
  }

  private createEventAPI() {
    return {
      on: (event: string, handler: any) => {},
      off: (event: string, handler: any) => {},
      emit: (event: string, data?: any) => {},
      once: (event: string, handler: any) => {}
    }
  }

  private createToolsAPI() {
    return {
      register: (tool: any) => this.workflow.registerTool(tool),
      unregister: (name: string) => {},
      list: () => []
    }
  }

  private createMenusAPI() {
    return {
      register: (menu: any) => {},
      unregister: (id: string) => {},
      list: () => []
    }
  }

  private createRoutesAPI() {
    return {
      register: (route: any) => {},
      unregister: (path: string) => {},
      list: () => []
    }
  }

  private createCommandsAPI() {
    return {
      register: (command: any) => {},
      unregister: (id: string) => {},
      execute: async (id: string, ...args: any[]) => {},
      list: () => []
    }
  }

  private createLoggerAPI(pluginId: string) {
    return {
      debug: (message: string, ...args: any[]) => console.debug(`[${pluginId}]`, message, ...args),
      info: (message: string, ...args: any[]) => console.info(`[${pluginId}]`, message, ...args),
      warn: (message: string, ...args: any[]) => console.warn(`[${pluginId}]`, message, ...args),
      error: (message: string, ...args: any[]) => console.error(`[${pluginId}]`, message, ...args)
    }
  }
}

// 导出单例
export const pluginSystem = new PluginSystemManager()
