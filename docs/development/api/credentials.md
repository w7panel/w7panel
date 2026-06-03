# API 调用凭据

本文档说明 W7Panel API 调用中常见 token、签名和微应用凭据的来源、请求格式、响应字段和使用边界。开发新接口时，应先明确接口需要哪一种凭据，避免把用户 token、ServiceAccount token、微应用 Basic 认证、Console token 和 OIDC token 混用。

## 整体使用方式

W7Panel 的调用凭据按“谁在调用、调用什么资源”来选择。默认规则是：前端或普通业务调用使用面板用户 token；容器文件操作使用 `webdavToken`；后端访问 Kubernetes 使用 ServiceAccount 或 `LOCAL_MOCK` 注入的 kubeconfig token；微应用和第三方系统按各自协议使用独立凭据。

### 基本流程

1. 登录面板，调用 `/panel-api/v1/login` 获取 `token` 和 `refreshToken`。
2. 普通面板 API 请求统一携带 `Authorization: Bearer <token>`。
3. token 过期后，使用 `/panel-api/v1/auth/refresh-token2` 和 `refreshToken` 换取新 token。
4. 进入容器文件、WebDAV、压缩、权限修改等场景时，先从容器信息接口获取 `webdavToken`，再用 `Authorization: Bearer <webdavToken>` 调用文件类接口。
5. 微应用内调用面板 API 时优先使用 Wujie props 或 micro-app data 传入的 `paneltoken`；调用微应用自身后端时使用该微应用自己的 Basic 认证或业务 token。
6. OIDC Client 按标准 OAuth/OIDC 流程获取 `access_token`，只用于 OIDC `/userinfo` 等协议接口，不等同于面板用户 token。

### 场景选择

| 场景 | 使用凭据 | 获取方式 | 传递方式 |
|------|----------|----------|----------|
| 登录、初始化用户、刷新 token | 无用户 token 或 refresh token | 登录接口返回 `refreshToken` | form `refreshToken` |
| 普通面板业务 API | 用户 token | `/panel-api/v1/login` 或刷新接口 | `Authorization: Bearer <user-token>` |
| K8s 代理和面板聚合接口 | 用户 token | 前端登录态或微应用 `paneltoken` | `Authorization: Bearer <user-token>` |
| 容器文件、WebDAV、压缩、权限修改 | webdavToken | 容器 PID/文件入口相关接口返回 | `Authorization: Bearer <webdavToken>` |
| 后端服务访问 Kubernetes | ServiceAccount token | Pod 内自动挂载 | 后端内部使用，不暴露给前端 |
| `LOCAL_MOCK=true` 开发测试 | kubeconfig token | `KUBECONFIG` 或默认 kubeconfig 路径 | 中间件注入 `k8s_token` |
| Console 签名登录 | ConsoleSignature | Console 请求签名 | 签名请求头/参数 |
| Console OAuth 登录/绑定 | Console OAuth code | `/auth/console/oauth` 跳转后回调 | query/form `code` |
| 微应用自身后端 | 微应用 Basic 认证 | MicroApp/ZPK 配置 | `Authorization: Basic ...` |
| OIDC 第三方 Client | OIDC access token | `/panel-api/v1/oidc/token` | `Authorization: Bearer <oidc-token>` |

### 使用边界

- `user-token` 是面板业务默认凭据，适用于绝大多数 `/panel-api/v1/*` 和 `/k8s-proxy/*` 调用。
- `refreshToken` 只用于刷新登录态，不用于普通业务 API。
- `webdavToken` 只用于目标容器文件系统相关操作，不应拿来调用普通面板业务 API。
- ServiceAccount token 和 kubeconfig token 只应在后端内部或 `LOCAL_MOCK` 开发测试中使用，不应返回给前端。
- `w7PanelToken` 是 Console 第三方 CD token，不等同于 `paneltoken`。
- OIDC `access_token` 只服务于 OIDC 协议，不保证能调用面板业务 API。
- 不要在 URL、日志、响应体、localStorage 调试输出中暴露完整 token、密码、密钥或 OIDC code。

## 通用约定

### 认证请求头

多数面板业务 API 默认使用 Bearer token：

```http
Authorization: Bearer <user-token>
```

后端 `helper.GetToken()` 当前按以下优先级取 token：

| 优先级 | 位置 | 示例 | 说明 |
|--------|------|------|------|
| 1 | query `api-token` | `?api-token=xxx` | 兼容入口，不建议新接口使用 |
| 2 | header `x-w7panel-token` | `x-w7panel-token: xxx` | 兼容入口 |
| 3 | header `AuthorizationX` | `AuthorizationX: Bearer xxx` | 兼容入口 |
| 4 | header `Authorization` | `Authorization: Bearer xxx` | 标准入口，优先用于新增接口 |

### 认证失败响应

缺少 token 或 TokenReview 校验失败时，`middleware.Auth` 返回：

```json
{
  "code": 401,
  "msg": "请登录"
}
```

TokenReview 失败时 `msg` 后会拼接具体错误信息。

### 成功响应形态

当前项目接口存在两类成功响应：

| 形态 | 示例 | 说明 |
|------|------|------|
| 直接业务对象 | `{"token":"...","expire":1710000000}` | `JsonResponseWithoutError` 返回，登录、用户信息、Console 信息等接口常用 |
| 字符串成功 | `"success"` | `JsonSuccessResponse` 返回，注册、绑定、导入证书等操作类接口常用 |

新增文档时应按实际 controller 返回形态描述，不要默认包一层 `data`。

## 凭据类型

| 凭据 | 来源 | 传递方式 | 适用范围 | 说明 |
|------|------|----------|----------|------|
| 用户 token | `/panel-api/v1/login`、`/panel-api/v1/auth/refresh-token2`、浏览器 `w7panel-token` | `Authorization: Bearer <token>` | 面板业务 API、K8s 代理、文件管理、应用管理 | 多数接口的默认凭据 |
| refresh token | 登录接口返回，浏览器保存为 `w7panel-refresh-token` | form `refreshToken` | 刷新用户 token | 只用于刷新登录态，不用于普通业务 API |
| ServiceAccount token | Pod 内 `/var/run/secrets/kubernetes.io/serviceaccount/token` | 后端内部使用 | 生产环境服务访问 Kubernetes API | 不应暴露给前端 |
| kubeconfig token | `KUBECONFIG` 指向的 kubeconfig | `LOCAL_MOCK` 中间件注入 `k8s_token` | `LOCAL_MOCK=true` 开发测试 | 用于本地模式模拟 K8s token |
| webdavToken | `/panel-api/v1/pid` 返回 | `Authorization: Bearer <webdavToken>` | WebDAV、压缩、权限修改、文件编辑器 | 用于目标容器文件系统访问，字段来源见 [container-files.md](container-files.md) |
| ConsoleSignature | Console 签名中间件 | 签名请求头/参数 | `/panel-api/v1/auth/login` | 用于控制台签名登录 |
| Console 第三方 CD token | `/panel-api/v1/auth/console/info` | 微应用 props `w7PanelToken` | 微应用、制品库、第三方持续交付 | 不等同于面板用户 token |
| 微应用 Basic 认证 | MicroApp/ZPK 配置中的 `username/password` | `Authorization: Basic ...` | 微应用自身后端 | 不用于面板 API |
| OIDC access token | `/panel-api/v1/oidc/token` | `Authorization: Bearer <oidc-token>` | OIDC `/userinfo` | 只能按 OIDC 协议获取用户信息 |

## 前端 token 注入

前端 axios 拦截器会自动从 `src/utils/auth.ts` 读取 token 并注入：

```http
Authorization: Bearer <getToken()>
```

`getToken()` 读取优先级：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | Wujie props | `window.$wujie.props.paneltoken` |
| 2 | micro-app data | `window.microApp.getData().token` |
| 3 | localStorage | `w7panel-token` 或 `iframe-w7panel-token` |

`getRefreshToken()` 读取优先级：

| 优先级 | 来源 | 字段 |
|--------|------|------|
| 1 | Wujie props | `window.$wujie.props.refreshToken` |
| 2 | localStorage | `w7panel-refresh-token` 或 `iframe-w7panel-refresh-token` |

请求配置中的 `customToken` 会覆盖默认 `Authorization`，适用于文件编辑器、容器文件等需要使用 `webdavToken` 的场景。

## 登录与刷新

### POST `/panel-api/v1/login`

功能：使用用户名密码登录面板，返回用户 token 和 refresh token。

认证：无需 Bearer token。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名，对应 K8s ServiceAccount 名称 |
| `password` | form | 是 | string | 密码 |
| `point` | form | 条件必填 | string | 验证码坐标，`captcha.enabled=true` 时必填 |
| `key` | form | 条件必填 | string | 验证码 key，`captcha.enabled=true` 时必填 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/login' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'username=admin' \
  --data-urlencode 'password=123456'
```

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | 用户 token，后续请求放入 `Authorization: Bearer` |
| `expire` | int64 | token 过期时间，Unix 秒 |
| `refreshToken` | string | refresh token，用于刷新登录态 |
| `isClusterUser` | bool | 是否集群用户 |
| `isK3kUser` | bool | 兼容字段，是否 K3k 用户 |

响应示例：

```json
{
  "token": "eyJhbGciOi...",
  "expire": 1710000000,
  "isK3kUser": false,
  "isClusterUser": false,
  "refreshToken": "refresh-token-value"
}
```

### POST `/panel-api/v1/auth/login`

功能：Console 签名登录入口，登录参数与 `/panel-api/v1/login` 相同。

认证：需要通过 `ConsoleSignature` 中间件。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |
| `point` | form | 否 | string | 当前入口不校验验证码 |
| `key` | form | 否 | string | 当前入口不校验验证码 |

响应参数：同 `/panel-api/v1/login`。

### POST `/panel-api/v1/auth/refresh-token2`

功能：使用 refresh token 换取新的用户 token。

认证：无需旧用户 token，但请求必须携带有效 refresh token。

请求类型：后端当前绑定 `application/x-www-form-urlencoded`。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `refreshToken` | form | 是 | string | 登录接口返回的 refresh token |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/auth/refresh-token2' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'refreshToken=refresh-token-value'
```

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | 新用户 token |
| `expire` | int64 | 新 token 过期时间，Unix 秒 |
| `refreshToken` | string | 新 refresh token |
| `isK3kUser` | bool | 兼容字段 |
| `isClusterUser` | bool | 是否集群用户，视登录路径返回 |

实现注意：当前前端拦截器 `src/api/interceptor.ts` 在 401 后向该接口发送 JSON `{ "token": getRefreshToken() }`，但后端 `RefreshToken2` 实际绑定的是 form 字段 `refreshToken`。调整刷新逻辑时必须同步前后端，避免刷新登录态失效。

## 当前用户信息

### GET `/panel-api/v1/auth/userinfo`

功能：返回当前 token 对应的 K3k 用户、角色、集群和功能权限信息。

认证：`Authorization: Bearer <user-token>`

请求参数：无。

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/auth/userinfo' \
  -H 'Authorization: Bearer <user-token>'
```

响应参数：后端返回 `K3kUser.ToArray()` 的 `map[string]string`。常用字段如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| `w7.cc/username` | string | 当前用户名称 |
| `w7.cc/role` | string | 用户角色，如 `founder`、`super`、`normal`、`technician` |
| `w7.cc/file-editor` | string | 是否允许文件编辑器，字符串布尔值 |
| `w7.cc/web-shell` | string | 是否允许 Web Shell，字符串布尔值 |
| `w7.cc/domain-white-list` | string | 域名白名单配置 |
| `w7.cc/quota-limit` | string | 配额限制 |
| `w7.cc/k3k-name` | string | K3k 名称 |
| `w7.cc/k3k-namespace` | string | K3k 命名空间 |
| `w7.cc/cluster-mode` | string | 集群模式 |
| `w7.cc/cluster-policy` | string | 集群策略 |
| `w7.cc/cluster-status` | string | 集群状态 |
| `w7.cc/expire-time` | string | 到期时间，格式 `2006-01-02 15:04:05` |
| `w7.cc/has-password` | string | 是否已设置密码，字符串布尔值 |
| `w7.cc/cvm-name` | string | CVM 名称 |
| `w7.cc/cvm-namespace` | string | CVM 命名空间 |

响应示例：

```json
{
  "w7.cc/username": "admin",
  "w7.cc/role": "founder",
  "w7.cc/file-editor": "true",
  "w7.cc/web-shell": "true",
  "w7.cc/k3k-name": "admin",
  "w7.cc/k3k-namespace": "default"
}
```

## 用户初始化与注册

### POST `/panel-api/v1/auth/register`

功能：注册用户。

认证：按路由配置使用，调用前确认当前环境是否开放该入口。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |
| `policyName` | form | 是 | string | 绑定策略名称 |

响应示例：

```json
"success"
```

### POST `/panel-api/v1/auth/init-user`

功能：初始化首个用户。

认证：按路由配置使用，通常配合初始化状态接口判断是否允许调用。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 初始用户名 |
| `password` | form | 是 | string | 初始密码 |

响应示例：

```json
"success"
```

## 密码管理

### POST `/panel-api/v1/auth/reset-password`

功能：校验指定用户旧密码后重置该用户密码。

认证：`Authorization: Bearer <user-token>`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 需要重置密码的用户名 |
| `password` | form | 是 | string | 原密码 |
| `newPassword` | form | 是 | string | 新密码 |

响应示例：

```json
{}
```

### POST `/panel-api/v1/auth/reset-password-current`

功能：修改当前登录用户密码。当前用户已有密码注解时会校验原密码。

认证：`Authorization: Bearer <user-token>`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `password` | form | 否 | string | 原密码；当前账号已有密码注解时需要 |
| `newPassword` | form | 是 | string | 新密码 |

响应示例：

```json
{}
```

## Console 凭据接口

### GET `/panel-api/v1/auth/console/oauth`

功能：生成 Console OAuth 跳转地址。

认证：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `redirect_uri` | query/form | 是 | string | OAuth 回调地址 |

响应规则：

| 请求头 | 响应 |
|--------|------|
| `Accept: application/json` | 返回 `{"url":"<redirect-url>"}` |
| 其他 | HTTP 302 跳转到 Console OAuth 地址 |

JSON 响应示例：

```json
{
  "url": "https://console.example.com/oauth/authorize?..."
}
```

实现注意：当前 controller 在返回 JSON 后没有显式 `return`，代码仍会继续执行 302 redirect。前端调用该接口时应优先按实际响应验证。

### GET `/panel-api/v1/auth/console/login`

功能：使用 Console OAuth `code` 登录面板；若 Console 用户尚未绑定面板用户，会尝试注册并绑定。

认证：无需 Bearer token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | query/form | 是 | string | Console OAuth 授权码 |
| `policyName` | query/form | 否 | string | 自动注册时使用的策略名称 |

响应参数：同 `/panel-api/v1/login`。

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/auth/console/login?code=<oauth-code>&policyName=default'
```

### GET `/panel-api/v1/auth/console/bind`

功能：把当前面板用户与 Console OAuth 授权结果绑定。

认证：`Authorization: Bearer <user-token>`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | query/form | 是 | string | Console OAuth 授权码 |

响应：

| 位置 | 字段 | 类型 | 说明 |
|------|------|------|------|
| header | `uid` | string | Console 用户 ID |
| body | - | string | `"success"` |

### GET `/panel-api/v1/auth/console/info`

功能：返回面板与 Console 的注册、授权、第三方 CD token 和证书信息。

认证：`Authorization: Bearer <user-token>`

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `thirdparty_cd_token` | string | Console 第三方 CD token |
| `cluster_id` | string | Console 侧集群 ID |
| `offline_url` | string | 离线/私有化 Console 地址 |
| `access_token` | string | Console OAuth access token |
| `expire_time` | int | `access_token` 过期时间，Unix 秒 |
| `api_server_url` | string | Console API 地址 |
| `require_oauth` | bool | 是否需要发起 OAuth 授权 |
| `is_register` | bool | 是否已注册到 Console |
| `userinfo` | object/null | Console 用户信息 |
| `license_type` | string | 证书类型，过期时回落为 `free` |
| `license_id` | string | 证书序列号 |
| `license_end_time` | string | 证书到期时间，格式 `2006-01-02 15:04:05` |
| `license_is_expired` | bool | 证书是否过期 |
| `debug_value` | string | 调试值 |

响应示例：

```json
{
  "thirdparty_cd_token": "cd-token-value",
  "cluster_id": "cluster-001",
  "offline_url": "",
  "access_token": "console-access-token",
  "expire_time": 1710000000,
  "api_server_url": "https://console.example.com",
  "require_oauth": false,
  "is_register": true,
  "userinfo": {
    "user_id": 1,
    "nickname": "admin"
  },
  "license_type": "free",
  "license_id": "0",
  "license_end_time": "2026-06-03 12:00:00",
  "license_is_expired": false,
  "debug_value": ""
}
```

### POST `/panel-api/v1/auth/console/register-to-console`

功能：把当前面板注册到 Console。

认证：`Authorization: Bearer <user-token>`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `offline_url` | form | 是 | string | Console 地址 |
| `api_server_url` | form | 否 | string | Console API 地址 |

响应示例：

```json
"success"
```

### POST `/panel-api/v1/auth/console/import-cert`

功能：导入证书内容。

认证：`Authorization: Bearer <user-token>`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `cert` | form | 是 | string | 证书内容 |

响应示例：

```json
"success"
```

### POST `/panel-api/v1/auth/console/import-cert-console`

功能：从 Console 按 licenseId 导入证书。

认证：`Authorization: Bearer <user-token>`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `licenseId` | form | 是 | string | Console license ID |

响应示例：

```json
"success"
```

### POST `/panel-api/v1/auth/console/verify-cert`

功能：校验证书状态。

认证：`Authorization: Bearer <user-token>`

请求参数：无。

响应示例：

```json
"success"
```

### POST `/panel-api/v1/auth/console/thirdparty-cd-token`

功能：预留的第三方 CD token 接口。

当前实现为空，controller 未写入响应体。前端或微应用不要依赖该接口获取 token，应优先使用 `/panel-api/v1/auth/console/info` 返回的 `thirdparty_cd_token`。

### ANY `/panel-api/v1/auth/console/proxy/*path`

功能：代理到 Console SDK。

认证：`middleware.NewAuth("founder")`，仅创始人/高权限场景使用。

请求参数：透传原始 method、path、query、body 和 header。

响应参数：透传 Console 返回。

### GET `/panel-api/v1/auth/console/code/:code`

功能：把优惠码请求代理到 Console 第三方 CD SDK。

认证：`Authorization: Bearer <user-token>`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | path | 是 | string | 优惠码 |

响应参数：透传 Console 返回。

### POST `/panel-api/v1/auth/console/register-zpk-site`

功能：注册 ZPK 站点，并将返回的 app secret 写入对应应用容器配置。

认证：无需用户 token。Controller 内部通过 `releaseName` 与 `installId` 做临时校验。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `host` | form | 是 | string | 站点域名 |
| `siteIdentifie` | form | 是 | string | 站点标识 |
| `installId` | form | 是 | string | 安装 ID |
| `appName` | form | 是 | string | 应用名称 |
| `containerName` | form | 是 | string | 容器名称 |
| `namespace` | form | 是 | string | 应用命名空间 |
| `releaseName` | form | 是 | string | Release 名称 |

响应参数：当前 Controller 成功后未显式写入 JSON。

## LOCAL_MOCK 行为

`LOCAL_MOCK=true` 或 `LOCAL_MOCK=1` 时，`middleware.Auth` 会绕过 TokenReview，并把 kubeconfig token 注入到请求上下文：

```go
ctx.Set("k8s_token", token)
```

token 读取顺序：

| 优先级 | 路径 |
|--------|------|
| 1 | `$KUBECONFIG` 环境变量指定文件 |
| 2 | `./kubeconfig.yaml` |
| 3 | `./config/kubeconfig.yaml` |
| 4 | `./w7panel/kubeconfig.yaml` |
| 5 | `filepath.Join(filepath.Dir(KO_DATA_PATH), "kubeconfig.yaml")` |
| 6 | 固定回退值 `local-mock-token` |

注意事项：

| 项 | 说明 |
|----|------|
| 用户身份 | `LOCAL_MOCK` 分支只设置 `k8s_token`，不设置 `username` |
| 使用范围 | 仅用于本地开发测试，默认部署测试使用 `LOCAL_MOCK=true` |
| 安全边界 | 生产接口不要依赖 `LOCAL_MOCK` 绕过真实用户鉴权 |
| 文件接口 | `LOCAL_MOCK` 只改变 Agent/文件访问路径，不改变前端应携带 token 的约定 |

## 公开接口

公开接口使用 `/panel-api/v1/noauth/*` 前缀，必须只返回业务允许公开的字段。

| 接口 | 说明 |
|------|------|
| `GET /panel-api/v1/noauth/site/beian` | 备案信息 |
| `GET /panel-api/v1/noauth/site/beian2` | 备案信息兼容入口 |
| `GET /panel-api/v1/noauth/site/k3k-config` | K3k 公开配置 |
| `GET /panel-api/v1/noauth/site/init-user` | 初始化用户相关公开配置 |
| `GET /panel-api/v1/noauth/site/lianxi` | 联系信息 |
| `GET /panel-api/v1/noauth/site/{name}/configmap` | 允许公开的 ConfigMap 业务字段 |

禁止公开接口返回完整 K8s Resource 对象、metadata、managedFields、Secret、token 或密码。

## 微应用常见 props

面板通过 Wujie 向微应用传递的 props 不是统一 token，字段含义不同：

| 字段 | 类型 | 说明 |
|------|------|------|
| `url` | string | 面板代理请求微应用后端服务的地址 |
| `group` | string | 应用标识分组；有多个子应用时为主应用标识 |
| `userid` | string/number | Console 用户 ID |
| `openid` | string | Console openid |
| `nickname` | string | Console 昵称 |
| `role` | string | 面板角色，如 `founder`、`super`、`normal`、`technician` |
| `paneltoken` | string | 面板用户 token，可用于请求面板 API |
| `w7PanelToken` | string | Console 第三方 CD token |
| `Authorization` | string | 微应用自身 Basic 认证 |
| `access_token` | string | 面板登录用户自身维护的 Console access token，只能获取用户信息 |
| `isRegister` | bool/string | 是否已注册 Console |
| `requestUrl` | string | 微应用请求基地址 |

更详细的微应用事件和 props 说明见 [../frontend/wujie-events.md](../frontend/wujie-events.md)。

## 开发检查

- 新增接口时明确是否需要 `middleware.Auth`、`middleware.NewAuth("founder")`、`ConsoleSignature`、OIDC 标准处理或公开访问。
- 不在日志、URL、响应体、前端 localStorage 中输出完整 token、密码、密钥、OIDC code。
- token 字段不要复用同名但不同含义的变量，尤其是 `access_token`、`paneltoken`、`w7PanelToken`、`webdavToken`。
- 前端新增调用时同步检查 `src/api/interceptor.ts`、`src/utils/auth.ts` 和微应用 props 传递逻辑。
- 修改登录、刷新、鉴权中间件后，同步检查本文档、[container-files.md](container-files.md) 和前端登录态处理。
