import fs from 'fs-extra'
import path from 'path'
import { app } from 'electron'
import type { PluginMetadata } from '@runixo/plugin-types'
import { PluginSystemManager } from './PluginSystemManager'

/**
 * 插件加载器
 */
export class PluginLoader {
  private pluginsDir: string
  private loadedPlugins = new Map<string, LoadedPlugin>()
  private manager: PluginSystemManager

  constructor(manager: PluginSystemManager) {
    this.manager = manager
    this.pluginsDir = path.join(app.getPath('userData'), 'plugins')
  }

  async initialize() {
    await fs.ensureDir(this.pluginsDir)
    await this.loadAllPlugins()
  }

  /**
   * 加载所有插件
   */
  async loadAllPlugins() {
    const dirs = await fs.readdir(this.pluginsDir)
    
    for (const dir of dirs) {
      if (dir === 'registry.json') continue
      
      try {
        await this.loadPlugin(dir)
      } catch (error) {
        console.error(`加载插件失败: ${dir}`, error)
      }
    }
  }

  /**
   * 加载单个插件
   */
  async loadPlugin(pluginId: string) {
    if (this.loadedPlugins.has(pluginId)) return

    const pluginDir = path.join(this.pluginsDir, pluginId)
    const manifestPath = path.join(pluginDir, 'plugin.json')

    if (!await fs.pathExists(manifestPath)) {
      throw new Error(`插件清单不存在: ${pluginId}`)
    }

    const metadata: PluginMetadata = await fs.readJSON(manifestPath)
    const config = await this.loadConfig(pluginId)

    // 创建插件上下文
    const context = this.manager.createPluginContext(pluginId, metadata, config)

    // 加载主文件
    const mainPath = path.join(pluginDir, metadata.main)
    const PluginClass = require(mainPath).default

    // 创建实例
    const instance = new PluginClass(context)

    // 调用生命周期
    await instance.onLoad()
    await instance.onEnable()

    this.loadedPlugins.set(pluginId, {
      metadata,
      instance,
      context,
      enabled: true
    })
  }

  /**
   * 卸载插件
   */
  async unloadPlugin(pluginId: string) {
    const loaded = this.loadedPlugins.get(pluginId)
    if (!loaded) return

    await loaded.instance.onDisable()
    await loaded.instance.onUnload()

    this.loadedPlugins.delete(pluginId)
  }

  /**
   * 重载插件
   */
  async reloadPlugin(pluginId: string) {
    await this.unloadPlugin(pluginId)
    await this.loadPlugin(pluginId)
  }

  /**
   * 获取插件实例
   */
  getPlugin(pluginId: string) {
    return this.loadedPlugins.get(pluginId)
  }

  /**
   * 加载插件配置
   */
  private async loadConfig(pluginId: string): Promise<any> {
    const configPath = path.join(this.pluginsDir, pluginId, 'config.json')
    if (await fs.pathExists(configPath)) {
      return await fs.readJSON(configPath)
    }
    return {}
  }

  /**
   * 保存插件配置
   */
  async saveConfig(pluginId: string, config: any) {
    const configPath = path.join(this.pluginsDir, pluginId, 'config.json')
    await fs.writeJSON(configPath, config, { spaces: 2 })

    const loaded = this.loadedPlugins.get(pluginId)
    if (loaded) {
      await loaded.instance.onConfigChange(config)
    }
  }
}

interface LoadedPlugin {
  metadata: PluginMetadata
  instance: any
  context: any
  enabled: boolean
}
