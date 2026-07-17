exports = module.exports = [
  {
    text: '开发文档',
    collapsible: true,
    items: [
      { text: '开发指南', link: '/development/' },
    ]
  },
  {
    text: '开发规范',
    collapsible: true,
    items: [
      { text: '性能规范总览', link: '/development/performance/' },
      { text: '后端性能规范', link: '/development/performance/backend' },
      { text: '前端性能规范', link: '/development/performance/frontend' },
    ]
  },
  {
    text: '后端',
    collapsible: true,
    items: [
      { text: '说明', link: '/development/api/' },
      { text: '开发约定', link: '/development/api/conventions' },
      { text: '调用凭据', link: '/development/api/credentials' },
      { text: 'OAuth/OIDC', link: '/development/api/oauth-oidc' },
      { text: 'Hawk 签名认证', link: '/development/api/hawk' },
      {
        text: '集群',
        collapsible: true,
        items: [
          { text: '集群资源', link: '/development/api/cluster-ops' },
          { text: '存储', link: '/development/api/longhorn' },
          { text: '文件管理', link: '/development/api/container-files' },
          { text: '镜像管理', link: '/development/api/container-images' },
          {
            text: '云主机',
            collapsible: true,
            items: [
              { text: '云主机', link: '/development/api/k3k' },
              { text: '订单与超卖', link: '/development/api/orders' },
            ]
          },
          { text: '集群指标', link: '/development/api/metrics' },
          {
            text: '网关',
            collapsible: true,
            items: [
              { text: '反向代理', link: '/development/frontend/gateway-proxy' },
              { text: 'AI 代理', link: '/development/frontend/ai-proxy' },
              { text: '网关插件', link: '/development/frontend/gateway-plugins' },
            ]
          },
        ]
      },
      { text: '微应用', link: '/development/api/microapp-static' },
      { text: 'Audit', link: '/development/api/audit' },
      { text: '其他', link: '/development/api/application' },
      { text: '制品库应用管理', link: '/development/api/zpk' },
    ]
  },
  {
    text: '前端',
    collapsible: true,
    items: [
      { text: '说明', link: '/development/frontend/' },
      { text: '开发约定', link: '/development/frontend/conventions' },
      { text: '调用凭据', link: '/development/frontend/auth-state' },
      {
        text: '组件',
        collapsible: true,
        items: [
          { text: '说明', link: '/development/frontend/components' },
          { text: '基础展示与导航', link: '/development/frontend/components-basic' },
          { text: '表单控件', link: '/development/frontend/components-forms' },
          { text: 'YAML 编辑', link: '/development/frontend/components-yaml' },
          { text: '日志、终端和文本对比', link: '/development/frontend/components-terminal' },
          { text: '应用、镜像和代码包', link: '/development/frontend/components-apps' },
          { text: '域名、灰度和缓存策略', link: '/development/frontend/components-domain' },
          { text: '集群、节点和容器选择', link: '/development/frontend/components-cluster' },
          { text: '资源、账号和权限', link: '/development/frontend/components-resource' },
          { text: '业务卡片和入口', link: '/development/frontend/components-business' },
          { text: '微前端桥接', link: '/development/frontend/components-microfrontend' },
        ]
      },
      { text: 'Wujie 事件', link: '/development/frontend/wujie-events' },
      { text: '微应用接入', link: '/development/frontend/microapps' },
    ]
  },
  {
    text: '应用开发示例',
    collapsible: true,
    items: [
      { text: '应用开发示例', link: '/development/examples/' },
      { text: '制品库操作示例', link: '/development/examples/zpk-product-workflow' },
    ]
  }
]
