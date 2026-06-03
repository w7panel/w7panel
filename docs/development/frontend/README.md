# 前端开发文档

`docs/development/frontend/` 是 `w7panel-ui` 前端开发资料入口。文档按“通用约定 + 专题明细”组织；新增或修改前端公开能力时，先更新对应专题文档，再按需补充通用约定。

源码目录：`w7panel-ui/src`

## 文档目录

| 文档 | 说明 |
|------|------|
| [conventions.md](./conventions.md) | 前端目录、启动入口、API 调用、页面开发、UI、性能和提交流程规范 |
| [auth-state.md](./auth-state.md) | 前端 token 注入、刷新、本地缓存和权限状态 |
| [components.md](./components.md) | 公共组件、全局注册组件、业务复用组件的 Props、Events、Ref 方法和使用说明 |
| [wujie-events.md](./wujie-events.md) | Wujie 微前端事件参数、回调响应和调用示例 |
| [microapps.md](./microapps.md) | 微应用容器、Wujie props、token 边界、后端代理和接入流程 |

## 使用顺序

1. 新开发页面或改页面流程，先看 [conventions.md](./conventions.md)。
2. 修改 token、权限、本地缓存或刷新逻辑，更新 [auth-state.md](./auth-state.md)。
3. 新增或修改微应用容器、props 或接入流程，更新 [microapps.md](./microapps.md)。
4. 新增或修改公共组件，更新 [components.md](./components.md)。
5. 新增或修改微应用事件参数或回调，更新 [wujie-events.md](./wujie-events.md)。
6. 涉及后端 API 字段、路径或鉴权方式变化时，同步检查 [../api/README.md](../api/README.md) 和前端调用封装。

## 分类原则

| 类型 | 维护位置 | 说明 |
|------|----------|------|
| 通用开发规范 | [conventions.md](./conventions.md) | 适用于所有页面和组件的约定 |
| 鉴权和状态 | [auth-state.md](./auth-state.md) | 记录 token、refresh token、本地缓存和权限状态 |
| 微应用接入 | [microapps.md](./microapps.md) | 记录微应用容器、props、token 边界和接入流程 |
| 组件明细 | [components.md](./components.md) | 记录可复用组件的外部契约 |
| 微前端事件 | [wujie-events.md](./wujie-events.md) | 记录面板和微应用之间的事件协议 |

## 维护规则

- README 只维护入口、分类和跳转，不放具体组件或事件参数。
- 组件、Wujie 事件、鉴权状态、localStorage key、API 路径前缀发生变化时必须同步更新对应文档。
- 后端接口变更后，前端 API 方法、类型、页面取值和文档要一起检查。
