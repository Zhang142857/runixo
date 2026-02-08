// 插件市场 API
// GET /api/v1/plugins/list?search=xxx

interface Env {
  GITHUB_REPO: string
}

const PLUGINS = [
  {
    id: 'devops-assistant',
    name: 'DevOps 助手',
    version: '1.0.0',
    description: '智能DevOps助手，提供自动化部署、监控和故障诊断',
    author: 'Runixo Team',
    category: 'ai',
    icon: '🤖',
    keywords: ['devops', 'deployment', 'monitoring', 'ai'],
    download_url: '',  // 运行时填充
  },
  {
    id: 'cloudflare',
    name: 'Cloudflare 管理',
    version: '2.0.0',
    description: 'Cloudflare DNS、CDN、WAF、SSL证书、Tunnel 管理',
    author: 'Runixo Team',
    category: 'cloud-service',
    icon: '☁️',
    keywords: ['cloudflare', 'dns', 'cdn', 'waf'],
    download_url: '',
  },
]

export const onRequestGet: PagesFunction<Env> = async (context) => {
  const url = new URL(context.request.url)
  const search = url.searchParams.get('search')?.toLowerCase() || ''
  const repo = context.env.GITHUB_REPO
  const base = `https://raw.githubusercontent.com/${repo}/main/plugins`

  let results = PLUGINS.map((p) => ({
    ...p,
    download_url: `${base}/${p.id === 'cloudflare' ? 'cloudflare-v2' : p.id}/`,
  }))

  if (search) {
    results = results.filter(
      (p) =>
        p.name.toLowerCase().includes(search) ||
        p.description.toLowerCase().includes(search) ||
        p.keywords.some((k) => k.includes(search))
    )
  }

  return Response.json({ plugins: results, total: results.length })
}
