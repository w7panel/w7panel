# 微应用接入

本文档说明 `w7panel-ui` 中微应用容器、Wujie props、token 边界、后端代理和接入流程。具体事件参数和回调见 [wujie-events.md](./wujie-events.md)。

## 整体使用方式

微应用接入的核心是：面板负责加载微应用前端、注入运行上下文、代理微应用后端，并通过 Wujie 事件暴露文件、日志、构建、域名、OIDC 等面板能力。微应用侧只按 props 和事件协议调用，不直接假设面板内部页面结构。

### 基本流程

1. 面板查询微应用信息，获取 frontendUrl、backendUrl、roleConfig、bindings 等字段。
2. 面板容器通过 Wujie 加载微应用前端。
3. 面板通过 props/data 注入 `paneltoken`、`url`、`requestUrl`、用户信息和 Console 相关 token。
4. 微应用请求面板 API 时使用 `paneltoken`。
5. 微应用请求自身后端时使用 `url` 或 `requestUrl`，由面板代理到后端服务。
6. 微应用需要面板能力时，通过 `window.$wujie.bus.$emit()` 触发事件，事件明细见 [wujie-events.md](./wujie-events.md)。

### 场景选择

| 场景 | 使用方式 | 说明 |
|------|----------|------|
| 加载微应用前端 | Wujie 容器 | 按 MicroApp frontendUrl 或静态资源状态加载 |
| 请求面板 API | `paneltoken` | 作为 `Authorization: Bearer` 使用 |
| 请求微应用后端 | `url` / `requestUrl` | 面板代理到微应用 backendUrl |
| 使用微应用自身认证 | `Authorization` | 通常为 Basic 认证 |
| 使用 Console 能力 | `w7PanelToken` / `access_token` | 只能用于对应 Console/OIDC 场景 |
| 调用面板能力 | Wujie 事件 | 文件、日志、构建、域名、OIDC code 等 |

## 常见微应用容器

| 容器 | 源码位置 | 说明 |
|------|----------|------|
| 通用应用组微应用 | `w7panel-ui/src/views/topapp/micro-container.vue` | 启动 `appmicro`，同时挂载 `WujieModals` |
| 应用详情微应用 | `w7panel-ui/src/views/app/apps/detail.vue`、`w7panel-ui/src/views/app/apps/micro.vue` | 应用详情内嵌微应用 |
| 应用商店 ZPK 页面 | `w7panel-ui/src/views/app/store/store-zpk.vue` | 应用商店页面内嵌微应用 |
| GPUStack | `w7panel-ui/src/views/app/gpustack/index.vue` | GPUStack 专用微应用和 worker 管理事件 |
| 文件/镜像缓存策略 | `w7panel-ui/src/components/domain-strategy-filecache.vue`、`domain-strategy-imagecache.vue` | 域名策略内嵌缓存微应用 |

## 常见 props

| 字段 | 说明 |
|------|------|
| `url` | 面板代理请求微应用后端服务的地址，可能会把相对 `backendUrl` 拼成当前 origin 下的绝对地址 |
| `group` | 应用标识分组；如果应用下有多个子应用，该值为主应用标识 |
| `userid` | 面板登录用户 ID |
| `openid` | 面板登录用户 openid |
| `nickname` | 面板登录用户昵称 |
| `role` | 面板用户角色，取值包括 `founder`、`super`、`normal`、`technician` |
| `access_token` | 面板登录用户自身维护的 access token，只能用于获取用户信息，不能准确定位 appid |
| `Authorization` | 应用自身 Basic 认证，通常由应用配置中的 `username/password` 生成 |
| `paneltoken` | 面板用户 token，来自 `getToken()` |
| `w7PanelToken` | Console 第三方持续交付 token，部分容器通过 `/panel-api/v1/auth/console/info` 获取 |
| `isRegister` | Console 注册状态 |
| `requestUrl` | 微应用环境下 axios baseURL 或资源访问基准地址，部分容器传入 |
| `appImage` | 应用镜像 |
| `domain` | 微应用绑定域名，GPUStack 等场景使用 |

## token 边界

| 字段 | 能否请求面板 API | 说明 |
|------|------------------|------|
| `paneltoken` | 可以 | 面板用户 token |
| `w7PanelToken` | 不可以 | Console 第三方持续交付 token |
| `access_token` | 不建议 | Console/OIDC access token，只能按对应协议使用 |
| `Authorization` | 不可以 | 微应用自身 Basic 认证 |

约定：

- 微应用请求面板 API 时使用 `paneltoken`。
- 微应用请求自身后端时使用 `url`、`requestUrl` 和应用自身认证。
- `paneltoken`、`w7PanelToken`、`access_token`、`Authorization` 不要混用。
- 不在微应用 URL、console 或错误提示中输出完整 token、密码和密钥。

## 事件接入

微应用事件调用统一使用：

```ts
window.$wujie.bus.$emit('eventName', payload, callback, rejectCallback);
```

面板侧注册和清理事件见 [wujie-events.md](./wujie-events.md)。新增事件时需要同时确认：

| 检查项 | 说明 |
|--------|------|
| 事件名 | 是否和现有事件冲突 |
| payload | 是否使用对象参数，字段是否稳定 |
| callback | 是否需要成功响应 |
| rejectCallback | 是否需要失败响应 |
| 清理逻辑 | 注册组件卸载时是否清理事件 |

## 后端接口文档

微应用后端代理、静态资源状态、静态资源下载和回源代理接口见 [../api/microapp-static.md](../api/microapp-static.md)。
