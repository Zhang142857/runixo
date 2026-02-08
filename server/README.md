# Runixo Server

Runixo 官网 + API，部署在 Cloudflare Pages + Functions 上。

## 项目结构

```
server/
├── src/                        # VitePress 官网源码
│   ├── .vitepress/config.ts    # 站点配置
│   ├── index.md                # 首页
│   ├── guide/                  # 文档
│   └── public/                 # 静态资源（logo.svg 替换此文件即可换 Logo）
├── functions/                  # Cloudflare Pages Functions (API)
│   └── api/v1/
│       ├── agent/check.ts      # Agent 更新检查
│       ├── plugins/list.ts     # 插件市场
│       └── docker/search.ts    # Docker Hub 搜索代理
├── wrangler.toml               # Cloudflare 配置
└── package.json
```

## 部署步骤

### 1. 安装依赖

```bash
cd server
npm install
```

### 2. 本地预览

```bash
npm run dev          # 预览官网
```

### 3. 部署到 Cloudflare Pages

**方式 A：连接 GitHub（推荐，自动部署）**

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com) → Pages → Create a project
2. 连接 GitHub 仓库 `Zhang142857/runixo`
3. 构建设置：
   - Build command: `cd server && npm install && npm run build`
   - Build output directory: `server/src/.vitepress/dist`
   - Root directory: `/`
4. 环境变量：`GITHUB_REPO` = `Zhang142857/runixo`
5. 保存，每次 push 自动部署

**方式 B：手动部署**

```bash
npm run build
npx wrangler pages deploy src/.vitepress/dist --project-name=runixo
```

### 4. 绑定域名

1. Cloudflare Dashboard → Pages → runixo → Custom domains
2. 添加 `runixo.top`
3. DNS 会自动配置

## API 端点

| 端点 | 说明 | 参数 |
|------|------|------|
| `GET /api/v1/agent/check` | Agent 更新检查 | `version`, `os`, `arch` |
| `GET /api/v1/plugins/list` | 插件列表 | `search`（可选） |
| `GET /api/v1/docker/search` | Docker 搜索代理 | `q`, `page_size`, `page` |

## 更换 Logo

替换 `src/public/logo.svg` 和 `src/public/favicon.svg` 即可，无需改代码。
