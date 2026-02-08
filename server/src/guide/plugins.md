# 插件开发

Runixo 提供强大的插件系统，支持 UI 扩展、AI 能力注册、工具注册等。

## 插件能力

- **UI 扩展** — 添加页面、菜单和组件
- **AI 能力** — 注册 Agent、工作流、提示词模板
- **工具注册** — 为 AI 助手提供工具能力
- **配置 UI** — 基于 JSON Schema 自动生成配置界面

## 快速开始

```bash
npx @runixo/plugin-cli create my-plugin
cd my-plugin
npm run dev
```

## 示例：注册 AI Agent

```typescript
import { Plugin } from '@runixo/plugin-sdk'

export default class MyPlugin extends Plugin {
  async onLoad() {
    this.registerAgent({
      id: 'my-assistant',
      name: '我的助手',
      description: '专业的AI助手',
      systemPrompt: '你是一个专业的助手...',
      tools: ['tool1', 'tool2']
    })
  }
}
```

## 插件市场

在客户端中打开插件市场，可浏览、搜索、一键安装官方和社区插件。
