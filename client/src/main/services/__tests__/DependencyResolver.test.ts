import { describe, it, expect } from 'vitest'
import { DependencyResolver } from '../DependencyResolver'

describe('DependencyResolver', () => {
  it('应该解析简单依赖', () => {
    const resolver = new DependencyResolver()

    resolver.registerPlugin({
      id: 'plugin-a',
      name: 'Plugin A',
      version: '1.0.0',
      description: '',
      author: '',
      main: '',
      permissions: []
    })

    resolver.registerPlugin({
      id: 'plugin-b',
      name: 'Plugin B',
      version: '1.0.0',
      description: '',
      author: '',
      main: '',
      permissions: [],
      dependencies: {
        plugins: { 'plugin-a': '^1.0.0' }
      }
    })

    const result = resolver.resolve('plugin-b')
    expect(result.errors).toHaveLength(0)
    expect(result.resolved).toEqual(['plugin-a', 'plugin-b'])
  })

  it('应该检测循环依赖', () => {
    const resolver = new DependencyResolver()

    resolver.registerPlugin({
      id: 'plugin-a',
      name: 'Plugin A',
      version: '1.0.0',
      description: '',
      author: '',
      main: '',
      permissions: [],
      dependencies: {
        plugins: { 'plugin-b': '^1.0.0' }
      }
    })

    resolver.registerPlugin({
      id: 'plugin-b',
      name: 'Plugin B',
      version: '1.0.0',
      description: '',
      author: '',
      main: '',
      permissions: [],
      dependencies: {
        plugins: { 'plugin-a': '^1.0.0' }
      }
    })

    const cycles = resolver.detectCycles('plugin-a')
    expect(cycles.length).toBeGreaterThan(0)
  })

  it('应该检查版本兼容性', () => {
    const resolver = new DependencyResolver()

    resolver.registerPlugin({
      id: 'plugin-a',
      name: 'Plugin A',
      version: '1.5.0',
      description: '',
      author: '',
      main: '',
      permissions: []
    })

    expect(resolver.checkCompatibility('plugin-a', '^1.0.0')).toBe(true)
    expect(resolver.checkCompatibility('plugin-a', '^2.0.0')).toBe(false)
  })
})
