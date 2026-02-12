import fs from 'fs-extra'
import path from 'path'
import { app } from 'electron'
import { safeStorage } from 'electron'

/**
 * 插件存储服务
 */
export class PluginStorageService {
  private storageDir: string

  constructor() {
    this.storageDir = path.join(app.getPath('userData'), 'plugin-storage')
  }

  async initialize() {
    await fs.ensureDir(this.storageDir)
  }

  /**
   * 创建插件存储API
   */
  createStorageAPI(pluginId: string) {
    const pluginStoragePath = path.join(this.storageDir, `${pluginId}.json`)

    return {
      get: async <T = any>(key: string): Promise<T | null> => {
        const data = await this.readStorage(pluginStoragePath)
        return data[key] ?? null
      },

      set: async (key: string, value: any): Promise<void> => {
        const data = await this.readStorage(pluginStoragePath)
        data[key] = value
        await this.writeStorage(pluginStoragePath, data)
      },

      delete: async (key: string): Promise<void> => {
        const data = await this.readStorage(pluginStoragePath)
        delete data[key]
        await this.writeStorage(pluginStoragePath, data)
      },

      clear: async (): Promise<void> => {
        await this.writeStorage(pluginStoragePath, {})
      },

      keys: async (): Promise<string[]> => {
        const data = await this.readStorage(pluginStoragePath)
        return Object.keys(data)
      }
    }
  }

  /**
   * 创建加密存储API
   */
  createSecureStorageAPI(pluginId: string) {
    const secureStoragePath = path.join(this.storageDir, `${pluginId}.secure.json`)

    return {
      get: async (key: string): Promise<string | null> => {
        const data = await this.readStorage(secureStoragePath)
        const encrypted = data[key]
        if (!encrypted) return null

        if (safeStorage.isEncryptionAvailable()) {
          const buffer = Buffer.from(encrypted, 'base64')
          return safeStorage.decryptString(buffer)
        }
        return encrypted
      },

      set: async (key: string, value: string): Promise<void> => {
        const data = await this.readStorage(secureStoragePath)
        
        if (safeStorage.isEncryptionAvailable()) {
          const buffer = safeStorage.encryptString(value)
          data[key] = buffer.toString('base64')
        } else {
          data[key] = value
        }

        await this.writeStorage(secureStoragePath, data)
      },

      delete: async (key: string): Promise<void> => {
        const data = await this.readStorage(secureStoragePath)
        delete data[key]
        await this.writeStorage(secureStoragePath, data)
      }
    }
  }

  private async readStorage(filePath: string): Promise<Record<string, any>> {
    if (await fs.pathExists(filePath)) {
      return await fs.readJSON(filePath)
    }
    return {}
  }

  private async writeStorage(filePath: string, data: Record<string, any>): Promise<void> {
    await fs.writeJSON(filePath, data, { spaces: 2 })
  }
}
