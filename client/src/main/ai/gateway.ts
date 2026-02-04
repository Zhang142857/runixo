import axios, { AxiosInstance } from 'axios'
import { EventEmitter } from 'events'

interface AIConfig {
  provider: 'openai' | 'claude' | 'ollama' | 'custom'
  apiKey?: string
  baseUrl?: string
  model?: string
}

interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content: string
  tool_call_id?: string
  tool_calls?: ToolCall[]
}

interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

interface AIContext {
  serverId?: string
  history?: ChatMessage[]
  systemPrompt?: string
}

interface Tool {
  name: string
  description: string
  parameters: Record<string, unknown>
  execute: (params: Record<string, unknown>, context?: AIContext) => Promise<unknown>
}

interface ToolExecutor {
  executeCommand: (serverId: string, command: string, args?: string[]) => Promise<unknown>
  listContainers: (serverId: string, all?: boolean) => Promise<unknown>
  containerAction: (serverId: string, containerId: string, action: string) => Promise<unknown>
  readFile: (serverId: string, path: string) => Promise<unknown>
  writeFile: (serverId: string, path: string, content: string) => Promise<unknown>
  listDirectory: (serverId: string, path: string) => Promise<unknown>
  getSystemInfo: (serverId: string) => Promise<unknown>
}

export class AIGateway extends EventEmitter {
  private config: AIConfig = {
    provider: 'ollama',
    baseUrl: 'http://localhost:11434',
    model: 'llama3'
  }

  private httpClient: AxiosInstance
  private tools: Map<string, Tool> = new Map()
  private systemPrompt: string
  private toolExecutor: ToolExecutor | null = null
  private maxToolCalls: number = 5  // 防止无限循环

  constructor() {
    super()
    this.httpClient = axios.create({
      timeout: 120000
    })

    this.systemPrompt = `你是 ServerHub AI 助手，一个专业的服务器运维助手。你可以帮助用户：
- 执行服务器命令
- 管理 Docker 容器和镜像
- 分析系统日志和性能指标
- 诊断服务器问题
- 生成配置文件
- 提供运维建议

你有以下工具可以使用：
- execute_command: 在服务器上执行命令
- list_containers: 列出 Docker 容器
- container_action: 对容器执行操作（启动、停止、重启等）
- read_file: 读取服务器上的文件
- write_file: 写入文件到服务器
- list_directory: 列出目录内容
- get_system_info: 获取系统信息

当用户请求执行操作时，请使用相应的工具。请用简洁专业的语言回答用户问题。`

    this.registerDefaultTools()
  }

  // 设置工具执行器（由 IPC 处理器注入）
  setToolExecutor(executor: ToolExecutor): void {
    this.toolExecutor = executor
  }

  setProvider(provider: string, config: Partial<AIConfig>): boolean {
    this.config = {
      ...this.config,
      provider: provider as AIConfig['provider'],
      ...config
    }
    return true
  }

  getConfig(): AIConfig {
    return { ...this.config }
  }

  getProviders(): Array<{ id: string; name: string; description: string }> {
    return [
      { id: 'ollama', name: 'Ollama', description: '本地运行的开源大语言模型' },
      { id: 'openai', name: 'OpenAI', description: 'GPT-4 等 OpenAI 模型' },
      { id: 'claude', name: 'Claude', description: 'Anthropic Claude 模型' },
      { id: 'custom', name: '自定义', description: '自定义 API 端点' }
    ]
  }

  async chat(
    message: string,
    context?: AIContext,
    onStream?: (chunk: string) => void
  ): Promise<string> {
    const messages: ChatMessage[] = [
      { role: 'system', content: context?.systemPrompt || this.systemPrompt },
      ...(context?.history || []),
      { role: 'user', content: message }
    ]

    switch (this.config.provider) {
      case 'ollama':
        return this.chatWithOllama(messages, context, onStream)
      case 'openai':
        return this.chatWithOpenAI(messages, context, onStream)
      case 'claude':
        return this.chatWithClaude(messages, context, onStream)
      default:
        throw new Error(`Unsupported provider: ${this.config.provider}`)
    }
  }

  // 执行工具调用
  private async executeToolCall(
    toolCall: ToolCall,
    context?: AIContext
  ): Promise<string> {
    const tool = this.tools.get(toolCall.function.name)
    if (!tool) {
      return JSON.stringify({ error: `Unknown tool: ${toolCall.function.name}` })
    }

    try {
      const params = JSON.parse(toolCall.function.arguments)
      // 如果没有指定 serverId，使用 context 中的
      if (!params.serverId && context?.serverId) {
        params.serverId = context.serverId
      }
      const result = await tool.execute(params, context)
      return JSON.stringify(result)
    } catch (error) {
      return JSON.stringify({ error: `Tool execution failed: ${(error as Error).message}` })
    }
  }

  private async chatWithOllama(
    messages: ChatMessage[],
    _context?: AIContext,
    onStream?: (chunk: string) => void
  ): Promise<string> {
    const url = `${this.config.baseUrl}/api/chat`

    // Ollama 也支持 tools，但格式略有不同
    const tools = this.getToolDefinitions()

    if (onStream) {
      // 流式响应
      const response = await this.httpClient.post(url, {
        model: this.config.model,
        messages: this.convertMessagesForOllama(messages),
        tools: tools.length > 0 ? tools : undefined,
        stream: true
      }, {
        responseType: 'stream'
      })

      let fullResponse = ''

      return new Promise((resolve, reject) => {
        response.data.on('data', (chunk: Buffer) => {
          const lines = chunk.toString().split('\n').filter((line: string) => line.trim())
          for (const line of lines) {
            try {
              const json = JSON.parse(line)
              if (json.message?.content) {
                fullResponse += json.message.content
                onStream(json.message.content)
              }
              if (json.done) {
                resolve(fullResponse)
              }
            } catch {
              // 忽略解析错误
            }
          }
        })
        response.data.on('error', reject)
      })
    } else {
      // 非流式响应
      const response = await this.httpClient.post(url, {
        model: this.config.model,
        messages: this.convertMessagesForOllama(messages),
        tools: tools.length > 0 ? tools : undefined,
        stream: false
      })
      return response.data.message?.content || ''
    }
  }

  private convertMessagesForOllama(messages: ChatMessage[]): Array<{role: string; content: string}> {
    return messages.map(m => ({
      role: m.role === 'tool' ? 'assistant' : m.role,
      content: m.content
    }))
  }

  private async chatWithOpenAI(
    messages: ChatMessage[],
    context?: AIContext,
    onStream?: (chunk: string) => void
  ): Promise<string> {
    const url = `${this.config.baseUrl || 'https://api.openai.com'}/v1/chat/completions`
    const tools = this.getToolDefinitions()

    let currentMessages = [...messages]
    let toolCallCount = 0

    // 循环处理工具调用
    while (toolCallCount < this.maxToolCalls) {
      const response = await this.httpClient.post(url, {
        model: this.config.model || 'gpt-4',
        messages: currentMessages,
        stream: false,  // 工具调用时不使用流式
        tools: tools.length > 0 ? tools : undefined,
        tool_choice: tools.length > 0 ? 'auto' : undefined
      }, {
        headers: {
          'Authorization': `Bearer ${this.config.apiKey}`,
          'Content-Type': 'application/json'
        }
      })

      const assistantMessage = response.data.choices?.[0]?.message
      if (!assistantMessage) {
        throw new Error('No response from OpenAI')
      }

      // 检查是否有工具调用
      if (assistantMessage.tool_calls && assistantMessage.tool_calls.length > 0) {
        // 添加助手消息（包含工具调用）
        currentMessages.push({
          role: 'assistant',
          content: assistantMessage.content || '',
          tool_calls: assistantMessage.tool_calls
        })

        // 执行每个工具调用并添加结果
        for (const toolCall of assistantMessage.tool_calls) {
          this.emit('tool_call', {
            name: toolCall.function.name,
            arguments: toolCall.function.arguments
          })

          const result = await this.executeToolCall(toolCall, context)

          this.emit('tool_result', {
            name: toolCall.function.name,
            result
          })

          currentMessages.push({
            role: 'tool',
            content: result,
            tool_call_id: toolCall.id
          })
        }

        toolCallCount++
      } else {
        // 没有工具调用，返回最终响应
        const finalContent = assistantMessage.content || ''

        // 如果需要流式输出，一次性发送
        if (onStream && finalContent) {
          onStream(finalContent)
        }

        return finalContent
      }
    }

    // 达到最大工具调用次数
    return '抱歉，操作过于复杂，请尝试简化您的请求。'
  }

  private async chatWithClaude(
    messages: ChatMessage[],
    _context?: AIContext,
    onStream?: (chunk: string) => void
  ): Promise<string> {
    const url = `${this.config.baseUrl || 'https://api.anthropic.com'}/v1/messages`

    // 转换消息格式（Claude API 格式略有不同）
    const systemMessage = messages.find(m => m.role === 'system')
    const chatMessages = messages.filter(m => m.role !== 'system')

    const response = await this.httpClient.post(url, {
      model: this.config.model || 'claude-3-opus-20240229',
      max_tokens: 4096,
      system: systemMessage?.content,
      messages: chatMessages.map(m => ({
        role: m.role === 'tool' ? 'user' : m.role,
        content: m.content
      })),
      stream: !!onStream
    }, {
      headers: {
        'x-api-key': this.config.apiKey,
        'anthropic-version': '2023-06-01',
        'Content-Type': 'application/json'
      },
      responseType: onStream ? 'stream' : 'json'
    })

    if (onStream) {
      let fullResponse = ''

      return new Promise((resolve, reject) => {
        response.data.on('data', (chunk: Buffer) => {
          const lines = chunk.toString().split('\n').filter((line: string) => line.startsWith('data: '))
          for (const line of lines) {
            try {
              const json = JSON.parse(line.slice(6))
              if (json.type === 'content_block_delta') {
                const text = json.delta?.text
                if (text) {
                  fullResponse += text
                  onStream(text)
                }
              }
              if (json.type === 'message_stop') {
                resolve(fullResponse)
              }
            } catch {
              // 忽略解析错误
            }
          }
        })
        response.data.on('error', reject)
      })
    } else {
      return response.data.content?.[0]?.text || ''
    }
  }

  // 工具注册
  registerTool(tool: Tool): void {
    this.tools.set(tool.name, tool)
  }

  private registerDefaultTools(): void {
    // 执行命令工具
    this.registerTool({
      name: 'execute_command',
      description: '在服务器上执行 shell 命令。返回命令的输出结果。',
      parameters: {
        type: 'object',
        properties: {
          command: { type: 'string', description: '要执行的命令' },
          args: {
            type: 'array',
            items: { type: 'string' },
            description: '命令参数（可选）'
          },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: ['command']
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.executeCommand(
            serverId,
            params.command as string,
            params.args as string[] | undefined
          )
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 列出容器工具
    this.registerTool({
      name: 'list_containers',
      description: '列出服务器上的 Docker 容器',
      parameters: {
        type: 'object',
        properties: {
          all: { type: 'boolean', description: '是否包含已停止的容器，默认 false' },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: []
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.listContainers(serverId, params.all as boolean)
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 容器操作工具
    this.registerTool({
      name: 'container_action',
      description: '对 Docker 容器执行操作（启动、停止、重启、暂停、恢复、删除）',
      parameters: {
        type: 'object',
        properties: {
          containerId: { type: 'string', description: '容器ID或名称' },
          action: {
            type: 'string',
            enum: ['start', 'stop', 'restart', 'pause', 'unpause', 'remove'],
            description: '要执行的操作'
          },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: ['containerId', 'action']
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.containerAction(
            serverId,
            params.containerId as string,
            params.action as string
          )
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 文件读取工具
    this.registerTool({
      name: 'read_file',
      description: '读取服务器上的文件内容',
      parameters: {
        type: 'object',
        properties: {
          path: { type: 'string', description: '文件的绝对路径' },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: ['path']
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.readFile(serverId, params.path as string)
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 文件写入工具
    this.registerTool({
      name: 'write_file',
      description: '将内容写入服务器上的文件',
      parameters: {
        type: 'object',
        properties: {
          path: { type: 'string', description: '文件的绝对路径' },
          content: { type: 'string', description: '要写入的内容' },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: ['path', 'content']
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.writeFile(
            serverId,
            params.path as string,
            params.content as string
          )
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 列出目录工具
    this.registerTool({
      name: 'list_directory',
      description: '列出服务器上目录的内容',
      parameters: {
        type: 'object',
        properties: {
          path: { type: 'string', description: '目录的绝对路径' },
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: ['path']
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.listDirectory(serverId, params.path as string)
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })

    // 获取系统信息工具
    this.registerTool({
      name: 'get_system_info',
      description: '获取服务器的系统信息，包括 CPU、内存、磁盘、网络等',
      parameters: {
        type: 'object',
        properties: {
          serverId: { type: 'string', description: '服务器ID（如果上下文中已有则可省略）' }
        },
        required: []
      },
      execute: async (params, context) => {
        if (!this.toolExecutor) {
          return { error: '工具执行器未初始化' }
        }
        const serverId = params.serverId as string || context?.serverId
        if (!serverId) {
          return { error: '未指定服务器ID' }
        }
        try {
          const result = await this.toolExecutor.getSystemInfo(serverId)
          return result
        } catch (error) {
          return { error: (error as Error).message }
        }
      }
    })
  }

  private getToolDefinitions(): Array<{type: string; function: {name: string; description: string; parameters: Record<string, unknown>}}> {
    return Array.from(this.tools.values()).map(tool => ({
      type: 'function',
      function: {
        name: tool.name,
        description: tool.description,
        parameters: tool.parameters
      }
    }))
  }
}
