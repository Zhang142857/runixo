// Docker Hub 搜索代理
// GET /api/v1/docker/search?q=nginx&page_size=25&page=1

export const onRequestGet: PagesFunction = async (context) => {
  const url = new URL(context.request.url)
  const query = url.searchParams.get('q') || ''
  const pageSize = url.searchParams.get('page_size') || '25'
  const page = url.searchParams.get('page') || '1'

  if (!query) {
    return Response.json({ error: 'Missing query parameter: q' }, { status: 400 })
  }

  const hubUrl = `https://hub.docker.com/v2/search/repositories/?query=${encodeURIComponent(query)}&page_size=${pageSize}&page=${page}`

  const resp = await fetch(hubUrl, {
    headers: { 'User-Agent': 'Runixo-Docker-Proxy' },
  })

  if (!resp.ok) {
    return Response.json({ error: `Docker Hub returned ${resp.status}` }, { status: resp.status })
  }

  const data: any = await resp.json()

  return Response.json({
    results: (data.results || []).map((r: any) => ({
      name: r.repo_name || r.name,
      description: r.short_description || r.description || '',
      star_count: r.star_count || 0,
      pull_count: r.pull_count || 0,
      is_official: r.is_official || false,
      is_automated: r.is_automated || false,
    })),
    total_count: data.count || 0,
    page: Number(page),
    page_size: Number(pageSize),
  })
}
