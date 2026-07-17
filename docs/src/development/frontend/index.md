# 前端开发文档

`docs/src/development/frontend/` 是 `w7panel-ui` 前端开发资料入口。文档按“通用约定 + 专题明细”组织；新增或修改前端公开能力时，先更新对应专题文档，再按需补充通用约定。

源码目录：`w7panel-ui/src`

## 文档目录

| 文档 | 说明 |
|------|------|
| [conventions.md](./conventions.md) | 前端目录、启动入口、API 调用、页面开发、UI、性能和提交流程规范 |
| [auth-state.md](./auth-state.md) | 前端 token 注入、刷新、本地缓存和权限状态 |
| [components.md](./components.md) | 组件文档入口，按类型跳转到各组件专题 |
| [wujie-events.md](./wujie-events.md) | Wujie 微前端事件参数、回调响应和调用示例 |
| [microapps.md](./microapps.md) | 微应用容器、Wujie props、token 边界、后端代理和接入流程 |
| [gateway-plugins.md](./gateway-plugins.md) | Higress WasmPlugin 管理、作用域、MicroApp 配置界面和 YAML 回退 |
| [gateway-proxy.md](./gateway-proxy.md) | 网关反向代理的 McpBridge、Ingress、上游和删除一致性 |
| [ai-proxy.md](./ai-proxy.md) | AI 代理域名、提供者隔离、权重、Key Auth 和模型白名单资源关系 |

## 使用顺序

1. 新开发页面或改页面流程，先看 [conventions.md](./conventions.md)。
2. 修改 token、权限、本地缓存或刷新逻辑，更新 [auth-state.md](./auth-state.md)。
3. 新增或修改微应用容器、props 或接入流程，更新 [microapps.md](./microapps.md)。
4. 修改网关插件、配置作用域或插件 MicroApp 接入时，更新 [gateway-plugins.md](./gateway-plugins.md)。
5. 修改网关反向代理的 McpBridge、Ingress 或域名管理流程时，更新 [gateway-proxy.md](./gateway-proxy.md)。
6. 修改 AI 代理资源、域名配置或 Higress 插件协议时，更新 [ai-proxy.md](./ai-proxy.md)。
7. 新增或修改公共组件，先从 [components.md](./components.md) 确认类型，再更新对应组件专题。
8. 新增或修改微应用事件参数或回调，更新 [wujie-events.md](./wujie-events.md)。
9. 涉及后端 API 字段、路径或鉴权方式变化时，同步检查 [../api/](../api/) 和前端调用封装。

## 维护规则

- README 只维护入口、分类和跳转，不放具体组件或事件参数。
- 组件、Wujie 事件、鉴权状态、localStorage key、API 路径前缀发生变化时必须同步更新对应文档。
- 后端接口变更后，前端 API 方法、类型、页面取值和文档要一起检查。
