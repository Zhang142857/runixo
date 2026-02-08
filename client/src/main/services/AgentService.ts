import type {
  AgentDefinition,
  PromptTemplateDefinition,
  AgentCallOptions,
  AgentResponse
} from '@runixo/plugin-types'

/**
 * Agent 管理服务
 */
export class AgentService {
  private agents = new Map<string, AgentDefinition>()
  private prompts = new Map<string, PromptTemplateDefinition>()

  registerAgent(agent: AgentDefinition) {
    this.agents.set(agent.id, agent)
  }

  registerPromptTemplate(template: PromptTemplateDefinition) {
    this.prompts.set(template.id, template)
  }

  getAgent(id: string): AgentDefinition | undefined {
    return this.agents.get(id)
  }

  listAgents(): AgentDefinition[] {
    return Array.from(this.agents.values())
  }

  listPromptTemplates(): PromptTemplateDefinition[] {
    return Array.from(this.prompts.values())
  }

  /**
   * 渲染提示词模板
   */
  renderPrompt(templateId: string, variables: Record<string, any>): string {
    const template = this.prompts.get(templateId)
    if (!template) throw new Error(`提示词模板不存在: ${templateId}`)

    // 验证必需变量
    for (const variable of template.variables) {
      if (variable.required && !(variable.name in variables)) {
        throw new Error(`缺少必需变量: ${variable.name}`)
      }
    }

    // 替换变量
    let result = template.template
    for (const [key, value] of Object.entries(variables)) {
      result = result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), String(value))
    }

    return result
  }

  /**
   * 调用 Agent
   */
  async chat(agentId: string, prompt: string, options?: AgentCallOptions): Promise<AgentResponse> {
    const agent = this.agents.get(agentId)
    if (!agent) throw new Error(`Agent不存在: ${agentId}`)

    // TODO: 实际调用 AI 服务
    // 这里需要集成实际的 AI API
    return {
      message: '这是模拟响应',
      toolCalls: []
    }
  }
}
