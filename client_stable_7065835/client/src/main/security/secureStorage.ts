/**
 * 安全凭据存储模块
 * 使用 Electron 的 safeStorage API 进行加密存储
 */

import { safeStorage } from 'electron'
import * as fs from 'fs'
import * as path from 'path'
import { app } from 'electron'

// 存储文件路径
const getStoragePath = () => path.join(app.getPath('userData'), 'secure-credentials.enc')

interface CredentialStore {
  [key: string]: string
}

/**
 * 检查加密是否可用
 */
export function isEncryptionAvailable(): boolean {
  return safeStorage.isEncryptionAvailable()
}

/**
 * 加密字符串
 */
export function encryptString(plainText: string): Buffer {
  if (!isEncryptionAvailable()) {
    throw new Error('系统加密服务不可用')
  }
  return safeStorage.encryptString(plainText)
}

/**
 * 解密字符串
 */
export function decryptString(encrypted: Buffer): string {
  if (!isEncryptionAvailable()) {
    throw new Error('系统加密服务不可用')
  }
  return safeStorage.decryptString(encrypted)
}

/**
 * 读取加密存储
 */
function readStore(): CredentialStore {
  const storagePath = getStoragePath()

  if (!fs.existsSync(storagePath)) {
    return {}
  }

  try {
    const encryptedData = fs.readFileSync(storagePath)
    const decryptedJson = decryptString(encryptedData)
    return JSON.parse(decryptedJson)
  } catch (error) {
    console.error('读取加密存储失败:', error)
    return {}
  }
}

/**
 * 写入加密存储
 */
function writeStore(store: CredentialStore): void {
  const storagePath = getStoragePath()
  const jsonData = JSON.stringify(store)
  const encryptedData = encryptString(jsonData)

  // 确保目录存在
  const dir = path.dirname(storagePath)
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }

  // 写入文件，设置安全权限
  fs.writeFileSync(storagePath, encryptedData, { mode: 0o600 })
}

/**
 * 存储凭据
 */
export function setCredential(key: string, value: string): void {
  if (!isEncryptionAvailable()) {
    throw new Error('系统加密服务不可用，无法安全存储凭据')
  }

  // 验证 key
  if (!key || typeof key !== 'string') {
    throw new Error('无效的凭据键名')
  }

  // 验证 value
  if (typeof value !== 'string') {
    throw new Error('凭据值必须是字符串')
  }

  const store = readStore()
  store[key] = value
  writeStore(store)
}

/**
 * 获取凭据
 */
export function getCredential(key: string): string | null {
  if (!isEncryptionAvailable()) {
    console.warn('系统加密服务不可用')
    return null
  }

  const store = readStore()
  return store[key] || null
}

/**
 * 删除凭据
 */
export function deleteCredential(key: string): void {
  const store = readStore()
  delete store[key]
  writeStore(store)
}

/**
 * 检查凭据是否存在
 */
export function hasCredential(key: string): boolean {
  const store = readStore()
  return key in store
}

/**
 * 列出所有凭据键名（不返回值）
 */
export function listCredentialKeys(): string[] {
  const store = readStore()
  return Object.keys(store)
}

/**
 * 清除所有凭据
 */
export function clearAllCredentials(): void {
  const storagePath = getStoragePath()
  if (fs.existsSync(storagePath)) {
    fs.unlinkSync(storagePath)
  }
}

/**
 * 凭据键名常量
 */
export const CredentialKeys = {
  // 服务器连接令牌
  SERVER_TOKEN: (serverId: string) => `server_token_${serverId}`,
  // Cloudflare API
  CLOUDFLARE_API_TOKEN: 'cloudflare_api_token',
  CLOUDFLARE_ACCOUNT_ID: 'cloudflare_account_id',
  // AWS
  AWS_ACCESS_KEY: 'aws_access_key',
  AWS_SECRET_KEY: 'aws_secret_key',
  // 阿里云
  ALIYUN_ACCESS_KEY: 'aliyun_access_key',
  ALIYUN_SECRET_KEY: 'aliyun_secret_key',
  // 腾讯云
  TENCENT_SECRET_ID: 'tencent_secret_id',
  TENCENT_SECRET_KEY: 'tencent_secret_key',
  // 对象存储
  OBJECT_STORAGE_ACCESS_KEY: 'object_storage_access_key',
  OBJECT_STORAGE_SECRET_KEY: 'object_storage_secret_key',
} as const
