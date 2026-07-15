exports = module.exports = [
  {
    text: '目录',
    collapsible: true,
    items: [
      { text: '使用文档', link: '/user-guide/' },
      {
        text: '概述',
        items: [
          { text: '微擎面板是什么', link: '/user-guide/' },
          { text: '版本日志', link: '/user-guide/overview/changelog/1.0.0' },
          { text: 'FAQ', link: '/user-guide/overview/faq' },
        ]
      },
      {
        text: '用户指南',
        items: [
          { text: '快速开始', link: '/user-guide/quick-start' },
          {
            text: '应用与交付',
            items: [
              { text: '应用管理', link: '/user-guide/app-management' },
              { text: '计划任务', link: '/user-guide/scheduled-tasks' },
            ]
          },
          {
            text: '集群与节点',
            items: [
              { text: '集群管理', link: '/user-guide/cluster-management' },
              { text: '镜像管理', link: '/user-guide/image-management' },
            ]
          },
          {
            text: '文件与存储',
            items: [
              { text: '文件管理', link: '/user-guide/file-management' },
              { text: '存储管理', link: '/user-guide/storage-management' },
            ]
          },
          {
            text: '访问与网络',
            items: [
              { text: '域名管理', link: '/user-guide/domain-management' },
              { text: '反向代理', link: '/user-guide/reverse-proxy' },
              { text: '网关插件', link: '/user-guide/gateway-plugins' },
              { text: '私有 DNS 解析', link: '/user-guide/private-dns' },
            ]
          },
        ]
      },
      {
        text: '官方应用',
        items: [
          { text: '站点管理', link: '/user-guide/site-management' },
          { text: '制品开发', link: '/user-guide/zpk-development' },
          { text: 'CDN 文件缓存', link: '/user-guide/cdn-cache' },
          { text: '镜像仓库缓存', link: '/user-guide/registry-cache' },
        ]
      },
    ]
  }
]
