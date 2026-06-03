# 前端鉴权和状态

本文档说明 `w7panel-ui` 前端登录态、token 注入、refresh token、本地缓存和权限状态的使用方式。通用开发约定见 [conventions.md](./conventions.md)，后端凭据边界见 [../api/credentials.md](../api/credentials.md)。

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

### 面板主体 token 使用方式

面板主体里常见的 token 分为五类：用户 token、refresh token、customToken、WebDAV token、第三方/集群专用 token。它们用途不同，不要互相替代。

| 类型 | 来源 | 用途 | 前端使用方式 |
|------|------|------|--------------|
| 用户 token | 登录接口返回并写入 `w7panel-token` | 请求面板业务 API 和 K8s 代理 API | 默认 axios 拦截器自动注入 |
| refresh token | 登录接口返回并写入 `w7panel-refresh-token` | 用户 token 过期后换新 token | 只由 401 刷新逻辑使用 |
| `customToken` | 页面或接口拿到的临时/代用户 token | 用另一个身份访问指定接口 | axios 请求配置里传 `customToken` |
| `webdavToken` | 文件接口返回 | WebDAV、外部编辑器、文件代理请求 | 手动放到 `Authorization: Bearer` |
| 第三方持续交付 token | Console/第三方接口返回 | 调用 Console 或第三方持续交付接口 | 通常作为 `customToken` 或专用 Header |

用户 token 是默认凭据。页面请求面板接口时只调用 `panelApi` 或 `k8sproxy`，不要手动拼 Header：

```ts
import { panelApi, k8sproxy } from '@/utils/api';

// 使用 w7panel-token
await panelApi.get('/auth/userinfo');

// 使用同一个用户 token 请求 K8s 代理
await k8sproxy.get('/api/v1/namespaces/default/pods');
```

需要读取当前 token 时，从 `src/utils/auth.ts` 取：

```ts
import { getToken } from '@/utils/auth';

const token = getToken();
```

只在需要传给子窗口、外部编辑器或构造非 axios 请求时读取 token。普通页面请求仍然走拦截器。

refresh token 只用于刷新登录态。页面不要把 refresh token 当作业务 token 使用，也不要主动把它传给普通接口：

```ts
import { getRefreshToken } from '@/utils/auth';

const refreshToken = getRefreshToken();
// 仅用于 /panel-api/v1/auth/refresh-token2 或统一刷新逻辑
```

`customToken` 用于明确要覆盖当前登录用户 token 的场景，例如代某个用户查看资源、访问第三方接口、访问临时授权资源：

```ts
await panelApi.get('/some-api', {
  customToken: otherUserToken,
});

await k8sproxy.put(resourceUrl, resourceBody, {
  customToken: clusterOrUserToken,
});
```

`customToken` 会覆盖默认 `getToken()`，所以不要在普通页面请求里随手传空字符串或旧 token。

WebDAV 和文件编辑器使用后端返回的 `webdavToken`，它不是面板登录 token。WebDAV 请求一般不走普通 `panelApi`，需要按协议手动带 Header：

```ts
await fetch(webdavUrl, {
  method: 'PROPFIND',
  headers: {
    Authorization: `Bearer ${webdavToken}`,
    Depth: '1',
  },
});
```

第三方持续交付、Console、License、成本中心等接口会使用第三方 token。这类 token 只用于对应外部接口，不写入 `w7panel-token`：

```ts
await axios.get('https://console.w7.cc/api/thirdparty-cd/k8s-offline/license', {
  customToken: thirdpartyCDToken,
});
```

### token 来源

| 运行环境 | 默认 token 来源 | 说明 |
|----------|-----------------|------|
| 面板主应用 | localStorage `w7panel-token` | 登录后写入 |
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
| 代用户或临时授权请求 | axios `customToken` | 覆盖默认 Authorization |
| 文件/WebDAV 请求 | `webdavToken` | 手动设置 `Authorization: Bearer &lt;webdavToken&gt;` |
| Console/第三方接口 | 第三方持续交付 token | 作为 `customToken` 或专用 Header |
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
- `webdavToken`、第三方持续交付 token、代用户 token 不写入 `w7panel-token`。
- refresh token 请求本身不要再触发普通 token 逻辑。

## token 读取优先级

`src/utils/auth.ts` 负责读取 token。常见优先级：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | localStorage | `w7panel-token` 或兼容 key |

refresh token 读取来源：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | localStorage | `w7panel-refresh-token` 或兼容 key |

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

## 用户信息和权限状态

用户信息、权限、K8s 信息和功能开关不是 token，但和登录态一起维护。页面不要直接解析 localStorage，统一通过 `src/utils/auth.ts`。

常用方法：

| 方法 | 读取/写入 | 说明 |
|------|-----------|------|
| `getUserInfo()` / `setUserInfo()` | 用户信息 | 读取用户模式、debug 标记、openid、昵称等 |
| `getPermission()` / `setPermission()` | 权限列表 | 控制菜单、按钮、资源访问范围 |
| `getK8sinfo()` / `setK8sinfo()` | K8s 信息 | 当前集群或初始化状态 |
| `getFileEditor()` / `setFileEditor()` | 文件编辑能力 | 控制文件编辑入口 |
| `getWebshell()` / `setWebshell()` | WebShell 能力 | 控制终端入口 |

示例：

```ts
import { getPermission, getUserInfo, getFileEditor, getWebshell } from '@/utils/auth';

const permission = getPermission() || [];
const userInfo = getUserInfo();
const fileEditorEnabled = getFileEditor() === 'true';
const webshellEnabled = getWebshell() === 'true';

const debug = userInfo?.['w7.cc/debug'] === 'true';
```

约定：

- 权限和能力状态只用于前端展示控制，后端仍必须做权限校验。
- 登录、刷新或切换用户后，要同步更新用户信息、权限和能力状态。
- 退出登录时使用 `clearToken()` 统一清理 token、refresh token、用户信息、权限和能力状态。

## 使用边界

- token、密码、密钥、OIDC code 不要输出到 console、URL 或错误提示。
- 401、刷新 token、退出登录必须走统一逻辑，不在页面内各自处理。
- 权限、用户信息和能力状态不能当作鉴权凭据使用。
- 后端修改 token 字段、刷新接口或鉴权方式时，需要同步检查本文、[../api/credentials.md](../api/credentials.md) 和前端拦截器。
