// Agent 更新检查 API
// GET /api/v1/agent/check?version=0.1.0&os=linux&arch=amd64

interface Env {
  GITHUB_REPO: string
}

interface GitHubRelease {
  tag_name: string
  body: string
  published_at: string
  assets: { name: string; size: number; browser_download_url: string }[]
}

export const onRequestGet: PagesFunction<Env> = async (context) => {
  const url = new URL(context.request.url)
  const version = url.searchParams.get('version') || '0.0.0'
  const os = url.searchParams.get('os') || 'linux'
  const arch = url.searchParams.get('arch') || 'amd64'
  const repo = context.env.GITHUB_REPO

  // 查询 GitHub 最新 Release（使用 Cloudflare Cache）
  const cacheKey = `https://api.github.com/repos/${repo}/releases/latest`
  const cache = caches.default
  let response = await cache.match(cacheKey)

  let release: GitHubRelease
  if (response) {
    release = await response.json()
  } else {
    const ghResp = await fetch(cacheKey, {
      headers: { 'User-Agent': 'Runixo-Update-Server', Accept: 'application/vnd.github+json' },
    })
    if (!ghResp.ok) {
      return Response.json({ available: false, current_version: version, latest_version: version }, { status: 200 })
    }
    release = await ghResp.json()
    // 缓存 1 小时
    const cached = new Response(JSON.stringify(release), { headers: { 'Cache-Control': 's-maxage=3600' } })
    context.waitUntil(cache.put(cacheKey, cached))
  }

  const latest = release.tag_name.replace(/^v/, '')
  if (!isNewer(latest, version)) {
    return Response.json({ available: false, current_version: version, latest_version: latest })
  }

  // 匹配平台对应的 asset
  const suffix = `${os}_${arch}`
  const asset = release.assets.find((a) => a.name.includes(suffix) && a.name.endsWith('.tar.gz'))
  const checksumAsset = release.assets.find((a) => a.name === 'checksums.txt')

  // 获取 checksum
  let checksum = ''
  if (checksumAsset) {
    const csResp = await fetch(checksumAsset.browser_download_url, { headers: { 'User-Agent': 'Runixo' } })
    if (csResp.ok) {
      const text = await csResp.text()
      const line = text.split('\n').find((l) => l.includes(suffix))
      if (line) checksum = line.split(/\s+/)[0]
    }
  }

  return Response.json({
    available: true,
    current_version: version,
    latest_version: latest,
    download_url: asset?.browser_download_url || '',
    size: asset?.size || 0,
    checksum,
    release_notes: release.body || '',
    release_date: release.published_at || '',
    is_critical: false,
  })
}

function isNewer(latest: string, current: string): boolean {
  const a = latest.split('.').map(Number)
  const b = current.split('.').map(Number)
  for (let i = 0; i < 3; i++) {
    if ((a[i] || 0) > (b[i] || 0)) return true
    if ((a[i] || 0) < (b[i] || 0)) return false
  }
  return false
}
