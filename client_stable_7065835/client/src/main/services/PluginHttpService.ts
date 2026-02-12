import axios, { AxiosInstance } from 'axios'
import type { HttpRequestConfig, HttpResponse } from '@runixo/plugin-types'

/**
 * 插件HTTP服务
 */
export class PluginHttpService {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      timeout: 30000,
      headers: {
        'User-Agent': 'Runixo-Plugin/2.0'
      }
    })
  }

  /**
   * 创建HTTP API
   */
  createHttpAPI() {
    return {
      get: async <T = any>(url: string, config?: HttpRequestConfig): Promise<HttpResponse<T>> => {
        const response = await this.client.get(url, config)
        return this.formatResponse(response)
      },

      post: async <T = any>(url: string, data?: any, config?: HttpRequestConfig): Promise<HttpResponse<T>> => {
        const response = await this.client.post(url, data, config)
        return this.formatResponse(response)
      },

      put: async <T = any>(url: string, data?: any, config?: HttpRequestConfig): Promise<HttpResponse<T>> => {
        const response = await this.client.put(url, data, config)
        return this.formatResponse(response)
      },

      delete: async <T = any>(url: string, config?: HttpRequestConfig): Promise<HttpResponse<T>> => {
        const response = await this.client.delete(url, config)
        return this.formatResponse(response)
      },

      request: async <T = any>(config: HttpRequestConfig): Promise<HttpResponse<T>> => {
        const response = await this.client.request(config)
        return this.formatResponse(response)
      }
    }
  }

  private formatResponse(response: any): HttpResponse {
    return {
      data: response.data,
      status: response.status,
      statusText: response.statusText,
      headers: response.headers
    }
  }
}
