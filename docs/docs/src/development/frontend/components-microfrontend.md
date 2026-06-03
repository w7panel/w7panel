# 微前端桥接组件

| 组件 | 源文件 | Props/输入 | Events | 说明 |
|------|--------|------------|--------|------|
| `WujieModals` | `src/components/wujie-modals.vue` | Wujie bus 事件 | 事件回调 | 微应用通用弹窗和面板能力桥接，事件协议见 [wujie-events.md](./wujie-events.md) |
| `SubaccountPanel` | `src/components/subaccount-panel.vue` | `token`, `refreshToken`, `path`, `name` | `close` | 子账号面板 Wujie 容器，向子面板注入 `paneltoken` 和 refresh token |
| `DomainStrategyFilecache` | `src/components/domain-strategy-filecache.vue` | `data`, `activeName` | `submit`, `cancel` | 文件缓存策略微应用容器，业务归在域名策略类型 |
| `DomainStrategyImagecache` | `src/components/domain-strategy-imagecache.vue` | `data`, `activeName` | `submit`, `cancel` | 图片缓存策略微应用容器，业务归在域名策略类型 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
