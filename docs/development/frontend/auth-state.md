# 前端鉴权和状态

本文档说明 `w7panel-ui` 前端登录态、token 注入、refresh token、本地缓存和权限状态的使用方式。微应用 props 和接入流程见 [microapps.md](./microapps.md)，通用开发约定见 [conventions.md](./conventions.md)，后端凭据边界见 [../api/credentials.md](../api/credentials.md)。

## 整体使用方式

前端授权的核心是“登录拿 token，后续请求自动带 token，过期后用 refresh token 刷新”。页面不直接管理 token 生命周期，而是通过 `src/utils/auth.ts` 读写登录态，通过 `src/api/interceptor.ts` 自动注入和刷新 token；需要覆盖 token 的特殊场景才使用 axios `customToken`。

### 授权模型

| 阶段 | 前端行为 | 后端/接口 |
|------|----------|-----------|
| 登录 | 提交用户名、密码和可选验证码 | `/panel-api/v1/login` 返回 `token`、`refreshToken`、用户类型等 |
| 保存登录态 | 保存 token、refresh token、用户信息和权限 | `src/utils/auth.ts` 写入 `w7panel-*` 本地缓存 |
| 普通请求 | axios 拦截器自动加 `Authorization` | 后端 `middleware.Auth` 校验 Bearer token |
| token 过期 | 拦截器捕获 401，尝试刷新 | `/panel-api/v1/auth/refresh-token2` 使用 refresh token 换新 token |
| 刷新失败 | 清理登录态并跳转登录页 | 前端统一退出或重新登录 |
| 特殊凭据 | 使用 `customToken` 覆盖默认 token | 文件、WebDAV、跨用户资源等明确场景 |

### token 怎么用

普通业务接口不需要页面手动写请求头。只要通过项目 axios 实例发请求，拦截器会读取 `getToken()` 并自动注入：

```http
Authorization: Bearer <user-token>
```

示例：

```ts
import { panelApi, k8sproxy } from '@/utils/api';

// 自动携带 Authorization: Bearer <getToken()>
panelApi.get('/auth/userinfo');

// K8s 代理请求同样自动携带用户 token
k8sproxy.get('/api/v1/namespaces/default/pods');
```

只有在明确需要使用另一个 token 时，才传 `customToken`：

```ts
panelApi.get('/some-api', {
  customToken: otherToken,
});
```

`customToken` 会让拦截器改用：

```http
Authorization: Bearer <customToken>
```

### token 来源

| 运行环境 | 默认 token 来源 | 说明 |
|----------|-----------------|------|
| 面板主应用 | localStorage `w7panel-token` | 登录后写入 |
| Wujie 微应用 | `window.$wujie.props.paneltoken` | 面板容器传入，优先级高于 localStorage |
| micro-app 环境 | `window.microApp.getData().token` | 兼容微应用容器 |
| 特殊请求 | axios `customToken` | 覆盖默认 token |

### refresh token 怎么用

refresh token 只用于刷新登录态，不用于普通业务 API。当前前端在收到 401 时统一处理：

1. 暂停/取消待处理请求。
2. 读取 `getRefreshToken()`。
3. 调用 `/panel-api/v1/auth/refresh-token2`。
4. 成功后写入新的 token 和 refresh token。
5. 失败后清理登录态并跳转登录页。

页面不要自行调用刷新接口，除非是在改造统一登录态逻辑。

### 场景选择

| 场景 | 使用方式 | 说明 |
|------|----------|------|
| 普通面板请求 | 默认 axios 拦截器 | 自动读取 `getToken()` |
| 刷新登录态 | refresh token | 拦截器统一处理 401 |
| 微应用内请求面板 API | Wujie props 或 micro-app data | 优先使用 `paneltoken` |
| 文件/WebDAV 等特殊 token | axios `customToken` | 覆盖默认 Authorization |
| 用户信息和权限 | `src/utils/auth.ts` | 统一读写本地缓存 |
| 全局业务状态 | Pinia store | 不把同一后端对象长期存多份 |

## token 注入

axios 拦截器位于 `src/api/interceptor.ts`。普通请求会自动注入：

```http
Authorization: Bearer <getToken()>
```

当请求配置存在 `customToken` 时，会覆盖默认 token：

```ts
panelApi.get('/some-api', {
  customToken: otherToken,
});
```

约定：

- 页面不要手动拼 `Authorization`，优先使用拦截器。
- `customToken` 只用于明确需要临时 token 的场景。
- refresh token 请求本身不要再触发普通 token 逻辑。

## token 读取优先级

`src/utils/auth.ts` 负责读取 token。常见优先级：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | Wujie props | `window.$wujie.props.paneltoken` |
| 2 | micro-app data | `window.microApp.getData().token` |
| 3 | localStorage | `w7panel-token` 或兼容 key |

refresh token 读取来源：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | Wujie props | `window.$wujie.props.refreshToken` |
| 2 | localStorage | `w7panel-refresh-token` 或兼容 key |

## localStorage key

| key | 说明 |
|-----|------|
| `w7panel-token` | 用户 token |
| `w7panel-refresh-token` | refresh token |
| `w7panel-permission` | 权限 |
| `w7panel-userinfo` | 用户信息 |
| `w7panel-k8sinfo` | K3k/K8s 信息 |
| `w7panel-fileeditor` | 文件编辑能力 |
| `w7panel-webshell` | WebShell 能力 |

约定：

- 不新增 `offline-*`、`k8soffline-*` 等旧命名 key。
- 新增本地缓存 key 必须使用 `w7panel-*` 前缀。
- 不在 localStorage 中存储明文密码、长期密钥、OIDC code 或 client secret。

## 使用边界

- 微应用 props、`paneltoken`、`w7PanelToken`、`access_token`、`Authorization` 的区别见 [microapps.md](./microapps.md)。
- token、密码、密钥、OIDC code 不要输出到 console、URL 或错误提示。
- 401、刷新 token、退出登录必须走统一逻辑，不在页面内各自处理。
- 后端修改 token 字段、刷新接口或鉴权方式时，需要同步检查本文、[../api/credentials.md](../api/credentials.md) 和前端拦截器。
