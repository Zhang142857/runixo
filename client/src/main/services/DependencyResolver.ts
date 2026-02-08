import type { PluginDependencies, PluginMetadata } from '@runixo/plugin-types'

/**
 * 依赖解析器
 */
export class DependencyResolver {
  private plugins = new Map<string, PluginMetadata>()

  registerPlugin(metadata: PluginMetadata) {
    this.plugins.set(metadata.id, metadata)
  }

  /**
   * 解析依赖树
   */
  resolve(pluginId: string): ResolveResult {
    const visited = new Set<string>()
    const resolved: string[] = []
    const errors: string[] = []

    try {
      this.resolveRecursive(pluginId, visited, resolved)
    } catch (error: any) {
      errors.push(error.message)
    }

    return { resolved, errors }
  }

  /**
   * 检测循环依赖
   */
  detectCycles(pluginId: string): string[] {
    const cycles: string[] = []
    const visiting = new Set<string>()
    const visited = new Set<string>()

    const visit = (id: string, path: string[]) => {
      if (visiting.has(id)) {
        cycles.push([...path, id].join(' -> '))
        return
      }
      if (visited.has(id)) return

      visiting.add(id)
      path.push(id)

      const plugin = this.plugins.get(id)
      if (plugin?.dependencies?.plugins) {
        for (const depId of Object.keys(plugin.dependencies.plugins)) {
          visit(depId, [...path])
        }
      }

      visiting.delete(id)
      visited.add(id)
    }

    visit(pluginId, [])
    return cycles
  }

  /**
   * 检查版本兼容性
   */
  checkCompatibility(pluginId: string, requiredVersion: string): boolean {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) return false

    return this.satisfiesVersion(plugin.version, requiredVersion)
  }

  private resolveRecursive(pluginId: string, visited: Set<string>, resolved: string[]) {
    if (visited.has(pluginId)) return
    visited.add(pluginId)

    const plugin = this.plugins.get(pluginId)
    if (!plugin) throw new Error(`插件不存在: ${pluginId}`)

    // 解析依赖
    if (plugin.dependencies?.plugins) {
      for (const [depId, version] of Object.entries(plugin.dependencies.plugins)) {
        if (!this.checkCompatibility(depId, version)) {
          throw new Error(`版本不兼容: ${depId}@${version}`)
        }
        this.resolveRecursive(depId, visited, resolved)
      }
    }

    resolved.push(pluginId)
  }

  private satisfiesVersion(actual: string, required: string): boolean {
    // 简化的版本匹配
    if (required.startsWith('^')) {
      const reqMajor = parseInt(required.slice(1).split('.')[0])
      const actMajor = parseInt(actual.split('.')[0])
      return actMajor === reqMajor
    }
    if (required.startsWith('>=')) {
      return this.compareVersions(actual, required.slice(2)) >= 0
    }
    return actual === required
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
}

interface ResolveResult {
  resolved: string[]
  errors: string[]
}
