import { describe, it, expect, beforeEach } from 'vitest'
import { WorkflowEngine } from '../WorkflowEngine'

describe('WorkflowEngine', () => {
  let engine: WorkflowEngine

  beforeEach(() => {
    engine = new WorkflowEngine()
  })

  it('应该注册工作流', () => {
    const workflow = {
      id: 'test-workflow',
      name: '测试工作流',
      description: '测试',
      steps: []
    }

    engine.registerWorkflow(workflow)
    expect(engine['workflows'].has('test-workflow')).toBe(true)
  })

  it('应该执行简单工作流', async () => {
    // 注册工具
    engine.registerTool({
      name: 'test_tool',
      displayName: '测试工具',
      description: '测试',
      category: '测试',
      parameters: {},
      handler: async () => ({ success: true })
    })

    // 注册工作流
    engine.registerWorkflow({
      id: 'simple',
      name: '简单工作流',
      description: '测试',
      steps: [
        {
          id: 'step1',
          type: 'tool',
          name: '步骤1',
          config: { tool: 'test_tool' }
        }
      ]
    })

    const result = await engine.execute('simple', {})
    expect(result.success).toBe(true)
  })

  it('应该处理条件分支', async () => {
    engine.registerTool({
      name: 'check',
      displayName: '检查',
      description: '检查',
      category: '测试',
      parameters: {},
      handler: async () => ({ success: true })
    })

    engine.registerWorkflow({
      id: 'conditional',
      name: '条件工作流',
      description: '测试',
      steps: [
        {
          id: 'check',
          type: 'tool',
          name: '检查',
          config: { tool: 'check' },
          next: 'condition'
        },
        {
          id: 'condition',
          type: 'condition',
          name: '条件',
          config: { condition: 'result.success === true' },
          next: ['success', 'failure']
        },
        {
          id: 'success',
          type: 'tool',
          name: '成功',
          config: { tool: 'check' }
        },
        {
          id: 'failure',
          type: 'tool',
          name: '失败',
          config: { tool: 'check' }
        }
      ]
    })

    const result = await engine.execute('conditional', {})
    expect(result.success).toBe(true)
  })
})
