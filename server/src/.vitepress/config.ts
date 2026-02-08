import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Runixo',
  description: 'AI-Native 服务器管理平台',
  lang: 'zh-CN',
  head: [
    ['link', { rel: 'icon', href: '/favicon.svg' }],
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: '首页', link: '/' },
      { text: '文档', link: '/guide/' },
      { text: 'GitHub', link: 'https://github.com/Zhang142857/runixo' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: '入门',
          items: [
            { text: '简介', link: '/guide/' },
            { text: '快速开始', link: '/guide/quickstart' },
            { text: '配置', link: '/guide/configuration' },
          ]
        },
        {
          text: '进阶',
          items: [
            { text: '插件开发', link: '/guide/plugins' },
            { text: 'TLS 证书', link: '/guide/tls' },
          ]
        }
      ]
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/Zhang142857/runixo' }
    ],
    footer: {
      message: 'MIT License',
      copyright: '© 2026 Runixo'
    },
    search: { provider: 'local' },
  }
})
