# Runixo

<p align="center">
  <img src="server/src/public/logo.svg" width="80" height="80" alt="Runixo">
</p>

<p align="center">
  <strong>AI-Native 服务器管理平台</strong><br>
  将 AI 能力深度融合到服务器运维的每个环节
</p>

<p align="center">
  <a href="https://runixo.top">官网</a> ·
  <a href="https://runixo.top/guide/">文档</a> ·
  <a href="https://github.com/Zhang142857/runixo/releases">下载</a>
</p>

## 特性

- 🔒 **安全架构** — Agent 不暴露 Web 端口，gRPC + TLS 加密通信
- 🤖 **AI 深度融合** — 自然语言运维、智能故障诊断、自动化工作流
- 🖥️ **多服务器管理** — 多节点、批量操作、跨节点编排
- 🐳 **容器管理** — Docker 容器/镜像/Compose 全生命周期管理
- 🧩 **插件生态** — 插件市场，功能即装即用
- ☁️ **云服务集成** — Cloudflare、AWS 等一键接入
- 📊 **实时监控** — CPU、内存、磁盘、网络实时指标

## 架构

```
┌──────────────────────────────────────────────┐
│           Runixo Client (Electron)           │
│   Vue 3 + Element Plus  │  AI 模块  │  管理  │
└──────────────┬───────────────────────────────┘
               │ gRPC (TLS)
┌──────────────▼───────────────────────────────┐
│           Runixo Agent (Go 单二进制)          │
│   gRPC 服务  │  命令执行器  │  数据采集器     │
└──────────────────────────────────────────────┘
```

## 快速开始

### SSH 自动安装（推荐）

客户端 →「服务器」→「SSH 安装」→ 填写连接信息 → 自动完成 Agent 安装、证书生成和配置。

### 手动安装

```bash
# 一键安装 Agent
curl -fsSL https://raw.githubusercontent.com/Zhang142857/runixo/main/scripts/install.sh | sudo bash

# 查看连接信息
sudo runixo info
```

在客户端添加服务器：填写 IP、端口、Token，如连接失败则导入证书。

### 下载客户端

从 [Releases](https://github.com/Zhang142857/runixo/releases) 下载：

| 平台 | 文件 |
|------|------|
| Windows | `Runixo-Setup-x.x.x.exe` |
| macOS | `Runixo-x.x.x.dmg` |
| Linux | `Runixo-x.x.x.AppImage` |

## 项目结构

```
runixo/
├── client/          # Electron 客户端 (Vue 3 + TypeScript)
├── agent/           # Go Agent (gRPC 服务)
├── server/          # 官网 + API (Cloudflare Pages + Functions)
├── proto/           # Protocol Buffers 定义
├── plugins/         # 官方插件
├── packages/        # 插件 SDK / CLI
├── sdk/             # 插件开发 SDK
├── scripts/         # 安装/构建脚本
└── examples/        # 示例插件
```

## 开发

```bash
# 客户端
cd client && pnpm install && pnpm electron:dev

# Agent
cd agent && go run cmd/agent/main.go

# 官网
cd server && npm install && npm run dev
```

## 构建

```bash
# 客户端
cd client && pnpm build

# Agent
cd agent && go build -o runixo-agent cmd/agent/main.go

# 官网（部署到 Cloudflare Pages）
cd server && npm run deploy
```

## 许可证

MIT License

