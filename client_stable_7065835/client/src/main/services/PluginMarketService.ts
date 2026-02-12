import type {
  PluginMarketInfo,
  PluginInstallStatus,
  PluginInstallOptions,
  PluginUpdateInfo
} from '@runixo/plugin-types'
import { EventEmitter } from 'events'
import fs from 'fs-extra'
import path from 'path'
import { app } from 'electron'

/**
 * 插件市场服务
 */
export class PluginMarketService extends EventEmitter {
  private pluginsDir: string
  private registryPath: string
  private registry: PluginRegistry

  constructor() {
    super()
    this.pluginsDir = path.join(app.getPath('userData'), 'plugins')
    this.registryPath = path.join(this.pluginsDir, 'registry.json')
    this.registry = { plugins: [], lastUpdated: new Date().toISOString() }
  }

  async initialize() {
    await fs.ensureDir(this.pluginsDir)
    await this.loadRegistry()
  }

  /**
   * 搜索插件
   */
  async search(query: string, options?: SearchOptions): Promise<PluginMarketInfo[]> {
    let results = [...this.registry.plugins]

    // 关键词搜索
    if (query) {
      const q = query.toLowerCase()
      results = results.filter(p =>
        p.name.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q) ||
        p.tags.some(t => t.toLowerCase().includes(q))
      )
    }

    // 分类过滤
    if (options?.category) {
      results = results.filter(p => p.category === options.category)
    }

    // 标签过滤
    if (options?.tags?.length) {
      results = results.filter(p =>
        options.tags!.some(t => p.tags.includes(t))
      )
    }

    // 排序
    const sortBy = options?.sort || 'downloads'
    results.sort((a, b) => {
      switch (sortBy) {
        case 'rating': return b.rating - a.rating
        case 'updated': return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        case 'name': return a.name.localeCompare(b.name)
        default: return b.downloads - a.downloads
      }
    })

    // 分页
    const limit = options?.limit || 20
    const offset = options?.offset || 0
    return results.slice(offset, offset + limit)
  }

  /**
   * 获取插件详情
   */
  async getPluginInfo(pluginId: string): Promise<PluginMarketInfo | null> {
    return this.registry.plugins.find(p => p.id === pluginId) || null
  }

  /**
   * 安装插件
   */
  async install(pluginId: string, options?: PluginInstallOptions): Promise<void> {
    const plugin = await this.getPluginInfo(pluginId)
    if (!plugin) throw new Error(`插件不存在: ${pluginId}`)

    const version = options?.version || plugin.version
    const pluginDir = path.join(this.pluginsDir, pluginId, version)

    this.emit('install:start', pluginId, version)

    try {
      // 检查依赖
      if (!options?.skipDependencies && plugin.dependencies) {
        await this.installDependencies(plugin.dependencies)
      }

      // 下载插件
      await this.downloadPlugin(pluginId, version, pluginDir)

      // 更新注册表
      await this.updateInstalledRegistry(pluginId, version)

      this.emit('install:complete', pluginId, version)
    } catch (error) {
      this.emit('install:error', pluginId, error)
      throw error
    }
  }

  /**
   * 卸载插件
   */
  async uninstall(pluginId: string): Promise<void> {
    const pluginDir = path.join(this.pluginsDir, pluginId)
    
    this.emit('uninstall:start', pluginId)

    try {
      await fs.remove(pluginDir)
      await this.removeFromInstalledRegistry(pluginId)
      
      this.emit('uninstall:complete', pluginId)
    } catch (error) {
      this.emit('uninstall:error', pluginId, error)
      throw error
    }
  }

  /**
   * 更新插件
   */
  async update(pluginId: string, version?: string): Promise<void> {
    const plugin = await this.getPluginInfo(pluginId)
    if (!plugin) throw new Error(`插件不存在: ${pluginId}`)

    const targetVersion = version || plugin.version

    this.emit('update:start', pluginId, targetVersion)

    try {
      await this.install(pluginId, { version: targetVersion, force: true })
      this.emit('update:complete', pluginId, targetVersion)
    } catch (error) {
      this.emit('update:error', pluginId, error)
      throw error
    }
  }

  /**
   * 检查更新
   */
  async checkUpdates(): Promise<PluginUpdateInfo[]> {
    const installed = await this.getInstalledPlugins()
    const updates: PluginUpdateInfo[] = []

    for (const plugin of installed) {
      const latest = await this.getPluginInfo(plugin.id)
      if (latest && this.compareVersions(latest.version, plugin.version) > 0) {
        updates.push({
          pluginId: plugin.id,
          currentVersion: plugin.version,
          latestVersion: latest.version,
          changelog: latest.changelog || '',
          breaking: this.isBreakingChange(plugin.version, latest.version)
        })
      }
    }

    return updates
  }

  /**
   * 获取安装状态
   */
  getInstallStatus(pluginId: string): PluginInstallStatus {
    const installedPath = path.join(this.pluginsDir, pluginId)
    if (fs.existsSync(installedPath)) {
      return PluginInstallStatus.INSTALLED
    }
    return PluginInstallStatus.NOT_INSTALLED
  }

  /**
   * 获取已安装插件
   */
  async getInstalledPlugins(): Promise<PluginMarketInfo[]> {
    const installed: PluginMarketInfo[] = []
    const dirs = await fs.readdir(this.pluginsDir)

    for (const dir of dirs) {
      if (dir === 'registry.json') continue
      const plugin = await this.getPluginInfo(dir)
      if (plugin) installed.push(plugin)
    }

    return installed
  }

  /**
   * 获取热门插件
   */
  async getFeatured(limit = 10): Promise<PluginMarketInfo[]> {
    return this.registry.plugins
      .filter(p => p.verified)
      .sort((a, b) => b.rating - a.rating)
      .slice(0, limit)
  }

  /**
   * 获取分类
   */
  async getCategories(): Promise<Array<{ id: string; name: string; count: number }>> {
    const categories = new Map<string, number>()
    
    for (const plugin of this.registry.plugins) {
      const count = categories.get(plugin.category) || 0
      categories.set(plugin.category, count + 1)
    }

    return Array.from(categories.entries()).map(([id, count]) => ({
      id,
      name: id,
      count
    }))
  }

  // 私有方法

  private async loadRegistry() {
    if (await fs.pathExists(this.registryPath)) {
      this.registry = await fs.readJSON(this.registryPath)
    } else {
      await this.saveRegistry()
    }
  }

  private async saveRegistry() {
    await fs.writeJSON(this.registryPath, this.registry, { spaces: 2 })
  }

  private async downloadPlugin(pluginId: string, version: string, targetDir: string) {
    // TODO: 实现实际下载逻辑
    // 这里可以从 GitHub Releases 或插件市场服务器下载
    await fs.ensureDir(targetDir)
  }

  private async installDependencies(dependencies: any) {
    // TODO: 实现依赖安装
    if (dependencies.plugins) {
      for (const [pluginId, version] of Object.entries(dependencies.plugins)) {
        const status = this.getInstallStatus(pluginId)
        if (status === PluginInstallStatus.NOT_INSTALLED) {
          await this.install(pluginId)
        }
      }
    }
  }

  private async updateInstalledRegistry(pluginId: string, version: string) {
    // 更新已安装插件记录
  }

  private async removeFromInstalledRegistry(pluginId: string) {
    // 从已安装记录中移除
  }

  private compareVersions(v1: string, v2: string): number {
    const parts1 = v1.split('.').map(Number)
    const parts2 = v2.split('.').map(Number)

    for (let i = 0; i < 3; i++) {
      if (parts1[i] > parts2[i]) return 1
      if (parts1[i] < parts2[i]) return -1
    }
    return 0
  }

  private isBreakingChange(oldVersion: string, newVersion: string): boolean {
    const oldMajor = parseInt(oldVersion.split('.')[0])
    const newMajor = parseInt(newVersion.split('.')[0])
    return newMajor > oldMajor
  }
}

interface PluginRegistry {
  plugins: PluginMarketInfo[]
  lastUpdated: string
}

interface SearchOptions {
  category?: string
  tags?: string[]
  sort?: 'downloads' | 'rating' | 'updated' | 'name'
  limit?: number
  offset?: number
}
