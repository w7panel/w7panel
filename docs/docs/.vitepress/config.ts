import path from 'path'
import process from 'node:process'
import versions from './versions'

const latest = versions[0]

const nav = [
  {
    text: '首页',
    link: '/'
  },
  {
    text: '文档',
    activeMatch: `^/([0-9]\.x)/`,
    items: versions.map((version) => ({
      text: version,
      link: `/${version}/`
    }))
  }
]

export const sidebar = versions.reduce(
  (sidebars, version) => ({
    ...sidebars,
    [`/${version}/`]: require(path.join(
      __dirname,
      `../src/${version}/sidebar`
    ))
  }),
  {}
)

export default {
  lang: 'zh-CN',
  title: 'W7Panel',
  description: '基于 Kubernetes 的云原生应用管理平台文档',
  srcDir: 'src',
  srcExclude: [],
  scrollOffset: 'header',
  metaChunk: true,
  ignoreDeadLinks: [],

  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }],
    // google analytics, without tracing dev
    ...(process?.argv?.[2] === 'dev' ? [] : [
      [
        'script',
        { async: '', src: 'https://www.googletagmanager.com/gtag/js?id=G-ZVHYZEP1SR' }
      ],
      [
        'script',
        {},
        `window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', 'G-ZVHYZEP1SR');`
      ],
    ]),
    // end google analytics
  ],

  markdown: {
    codeCopyButtonTitle: '复制',
    lineNumbers: true,
  },

  themeConfig: {
    nav,
    sidebar,

    logo: '/logo.png',
    siteTitle: false,

    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '搜索文档',
            buttonAriaLabel: '搜索文档'
          },
          modal: {
            noResultsText: '没有找到相关结果',
            resetButtonTitle: '清除查询',
            footer: {
              selectText: '选择',
              navigateText: '切换'
            }
          }
        }
      }
    },

    returnToTopLabel: '回到顶部',
    sidebarMenuLabel: '菜单',
    darkModeSwitchLabel: '主题模式',
    lightModeSwitchTitle: '浅色模式',
    darkModeSwitchTitle: '深色模式',

    outline: {
      level: [2, 3],
      label: '页面导航',
    },

    docFooter: {
      prev: '上一页',
      next: '下一页'
    },

    notFound: {
      title: '未找到',
      quote: '您所访问的页面未找到，或者已失效',
      linkLabel: '返回首页',
      linkText: '返回首页',
    },

    // carbonAds: {
    //   code: '',
    //   placement: ''
    // },

    socialLinks: [{ icon: 'github', link: 'https://github.com/w7panel/w7panel' }],

    editLink: {
      pattern:
        'https://github.com/w7panel/w7panel/edit/dev-v1/docs/docs/src/:path',
      text: '帮助我们改善此页面！'
    },

    license: {
      text: 'MIT License',
      link: 'https://opensource.org/licenses/MIT'
    },
    copyright: `Copyright © 2013-${new Date().getFullYear()} 微擎 <a class="ml-4" href="https://beian.miit.gov.cn/" target="_blank">皖ICP备19002904号-6</a>`
  },

  vite: {
    define: {
      __VUE_OPTIONS_API__: false
    },
    optimizeDeps: {
      include: ['gsap', 'dynamics.js'],
      exclude: []
    },
    // @ts-ignore
    ssr: {
      external: []
    },
    server: {
      host: true,
      fs: {
        // for when developing with locally linked theme
        allow: ['../..']
      }
    },
    json: {
      stringify: true
    }
  },

  vue: {
    reactivityTransform: true
  }
}
