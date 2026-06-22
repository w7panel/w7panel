# 微前端桥接组件

| 组件 | 源文件 | Props/输入 | Events | 说明 |
|------|--------|------------|--------|------|
| `WujieModals` | `src/components/wujie-modals.vue` | Wujie bus 事件 | 事件回调 | 微应用通用弹窗和面板能力桥接，事件协议见 [wujie-events.md](./wujie-events.md) |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
