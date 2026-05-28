# w7panel-server/app/auth API 文档

## 通用约定

OIDC 部分接口直接返回 OIDC 标准错误对象，例如：

```json
{
  "error": "invalid_request",
  "error_description": "..."
}
```

## 登录与用户

### POST `/panel-api/v1/login`

功能：使用用户名密码登录。验证码开启时会校验滑块验证码。

鉴权：无需鉴权。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |
| `point` | form | 否 | string | 验证码坐标；`captcha.enabled=true` 时必填 |
| `key` | form | 否 | string | 验证码 key；`captcha.enabled=true` 时必填 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | Kubernetes 访问 token |
| `expire` | int64 | token 过期时间，Unix 秒 |
| `isK3kUser` | bool | 是否 K3k 用户，废弃兼容字段 |
| `isClusterUser` | bool | 是否集群用户 |
| `refreshToken` | string | 刷新 token |

### POST `/panel-api/v1/auth/login`

功能：Console 签名登录，内部复用用户名密码登录逻辑，但不校验验证码。

鉴权：需要 `ConsoleSignature` 中间件校验。

请求参数：同 `/panel-api/v1/login`，其中 `point`、`key` 不参与验证码校验。

响应参数：同 `/panel-api/v1/login`。

### POST `/panel-api/v1/auth/register`

功能：注册 ServiceAccount 用户并绑定策略。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |
| `policyName` | form | 是 | string | 权限策略名称 |

响应参数：`"success"`。

### POST `/panel-api/v1/auth/refresh-token2`

功能：使用 `refreshToken` 刷新登录 token。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `refreshToken` | form | 是 | string | 登录接口返回的刷新 token |

响应参数：同 `/panel-api/v1/login`。

### POST `/panel-api/v1/auth/init-user`

功能：首次部署时初始化管理员用户。只有初始化 ConfigMap 存在时可调用，成功后会删除初始化 ConfigMap。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 管理员用户名 |
| `password` | form | 是 | string | 管理员密码 |

响应参数：`"success"`。

### POST `/panel-api/v1/auth/reset-password`

功能：校验指定用户旧密码后重置该用户密码。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 原密码 |
| `newPassword` | form | 是 | string | 新密码 |

响应参数：空对象 `{}`。

### POST `/panel-api/v1/auth/reset-password-current`

功能：修改当前登录用户密码。当前用户已有密码注解时会校验原密码。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `password` | form | 否 | string | 原密码 |
| `newPassword` | form | 是 | string | 新密码 |

响应参数：空对象 `{}`。

### GET `/panel-api/v1/auth/userinfo`

功能：获取当前登录用户信息。该路由复用 `app/k3k` 的 `K3k.Info` Controller。

鉴权：需要用户 token。

请求参数：无。

响应参数：用户和集群信息对象，字段以 `w7panel-server/app/k3k/http/controller/k3k.go` 的 `Info` 返回为准。

## Console 集成

### GET `/panel-api/v1/auth/console/oauth`

功能：获取 Console OAuth 登录地址；如果请求头 `Accept: application/json`，返回 JSON，否则 302 跳转。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `redirect_uri` | form/query | 是 | string | OAuth 回调地址 |

JSON 响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `url` | string | Console OAuth 登录地址 |

非 JSON 响应：`302` 跳转到 Console OAuth 登录地址。

### GET `/panel-api/v1/auth/console/login`

功能：Console OAuth 登录回调。通过 Console code 获取用户信息，已绑定用户直接登录，未绑定用户自动注册并绑定。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | form/query | 是 | string | Console OAuth code |
| `policyName` | form/query | 否 | string | 自动注册时使用的策略名称 |

响应参数：同 `/panel-api/v1/login`。

### GET `/panel-api/v1/auth/console/bind`

功能：将当前登录用户绑定到 Console 账号。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | form/query | 是 | string | Console OAuth code |

响应参数：`"success"`。

响应 Header：

| Header | 说明 |
|--------|------|
| `uid` | Console 用户 ID |

### GET `/panel-api/v1/auth/console/info`

功能：获取当前用户 Console 绑定配置和 license 信息。

鉴权：需要用户 token。

请求参数：无。

响应参数：`W7Config.ToArray()` 返回对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `license` | string | License 内容，存在时返回 |
| 其他字段 | any | Console 配置字段，以 `common/service/config` 定义为准 |

### GET `/panel-api/v1/auth/console/code/:code`

功能：代理 Console 优惠码接口，并附带当前用户策略组参数。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `code` | path | 是 | string | 优惠码 |

响应参数：Console 代理接口原始响应。

### ANY `/panel-api/v1/auth/console/proxy/*path`

功能：代理 Console SDK API。

鉴权：需要 founder 权限。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `path` | path | 是 | string | Console API 路径 |

响应参数：Console 代理接口原始响应。

### POST `/panel-api/v1/auth/console/register-to-console`

功能：将当前集群注册到 Console 交付系统。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `offline_url` | form | 是 | string | 面板访问地址；LOCAL_MOCK 下会被替换为固定测试地址 |
| `api_server_url` | form | 否 | string | Kubernetes API Server 地址 |

响应参数：`"success"`。

### POST `/panel-api/v1/auth/console/thirdparty-cd-token`

功能：预留接口。当前 Controller 内容被注释，无显式响应。

鉴权：需要用户 token。

请求参数：无。

响应参数：当前无显式响应。

### POST `/panel-api/v1/auth/console/import-cert`

功能：导入 license 证书内容。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `cert` | form | 是 | string | 证书内容 |

响应参数：`"success"`。

### POST `/panel-api/v1/auth/console/verify-cert`

功能：重新验证默认 license。

鉴权：需要用户 token。

请求参数：无。

响应参数：`"success"`。

### POST `/panel-api/v1/auth/console/import-cert-console`

功能：通过 Console license ID 验证并导入 license。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `licenseId` | form | 是 | string | Console license ID |

响应参数：`"success"`。

### POST `/panel-api/v1/auth/console/register-zpk-site`

功能：注册 ZPK 站点，并将返回的 app secret 写入对应应用容器配置。

鉴权：无需用户 token。Controller 内部通过 `releaseName` 与 `installId` 做临时校验。

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

## OIDC 标准接口

### ANY `/panel-api/v1/oidc/.well-known/openid-configuration`

功能：OIDC Discovery。OIDC 未启用时返回 `404`。

鉴权：无需用户 token。

请求参数：无。

响应参数：OIDC Discovery JSON，字段由 OIDC Provider 生成。

### ANY `/panel-api/v1/oidc/jwks`

功能：OIDC JWKS 公钥接口。

鉴权：无需用户 token。

请求参数：无。

响应参数：JWKS JSON。

### POST `/panel-api/v1/oidc/register`

功能：动态注册 OIDC Client。需要 registration access token。

鉴权：使用 `Authorization` token，由 OIDC server 校验 registration access token。

请求参数 JSON：

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `redirect_uris` | 否 | string[] | 回调地址列表 |
| `allow_any_redirect_uri` | 否 | bool | 是否允许任意回调地址 |
| `token_endpoint_auth_method` | 否 | string | token endpoint 认证方式 |
| `grant_types` | 否 | string[] | grant type 列表 |
| `scope` | 否 | string | scope 字符串 |
| `client_name` | 否 | string | Client 名称 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `client_id` | string | Client ID |
| `client_secret` | string | Client Secret，可能为空 |
| `client_id_issued_at` | int64 | Client ID 签发时间 |
| `client_secret_expires_at` | int64 | Secret 过期时间 |
| `redirect_uris` | string[] | 回调地址列表 |
| `allow_any_redirect_uri` | bool | 是否允许任意回调地址 |
| `token_endpoint_auth_method` | string | token endpoint 认证方式 |
| `grant_types` | string[] | grant type 列表 |
| `scope` | string | scope |
| `client_name` | string | Client 名称 |

### GET `/panel-api/v1/oidc/register/:clientId`

功能：获取动态注册的 OIDC Client。

鉴权：使用 registration access token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `clientId` | path | 是 | string | Client ID |

响应参数：同动态注册响应。

### PUT `/panel-api/v1/oidc/register/:clientId`

功能：更新动态注册的 OIDC Client。

鉴权：使用 registration access token。

请求参数：`clientId` path 参数 + 动态注册 JSON body。

响应参数：同动态注册响应。

### DELETE `/panel-api/v1/oidc/register/:clientId`

功能：删除动态注册的 OIDC Client。

鉴权：使用 registration access token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `clientId` | path | 是 | string | Client ID |

响应参数：成功时 HTTP `204 No Content`。

### GET `/panel-api/v1/oidc/authorize/login`

功能：OIDC 登录授权页面。若请求携带 token 且可识别用户名，会自动完成授权并跳转 callback。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | query | 是 | string | OIDC 授权请求 ID |

响应参数：HTML 页面或 `302` 跳转。

### POST `/panel-api/v1/oidc/authorize/login`

功能：提交 OIDC 登录授权页面用户名密码。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | query | 是 | string | OIDC 授权请求 ID |
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |

响应参数：成功时 `302` 跳转 callback；失败时返回 HTML 登录页并显示错误。

### ANY `/panel-api/v1/oidc/authorize`

功能：OIDC authorize 标准入口。

鉴权：无需用户 token。

请求参数：OIDC 标准 authorize 参数，如 `client_id`、`redirect_uri`、`scope`、`response_type`、`state`、`nonce`、PKCE 参数等。

响应参数：OIDC Provider 标准响应。

### ANY `/panel-api/v1/oidc/token`

功能：OIDC token 标准入口。

鉴权：按 OIDC Client 配置执行。

请求参数：OIDC 标准 token 参数。

响应参数：OIDC token 响应。

### ANY `/panel-api/v1/oidc/userinfo`

功能：OIDC userinfo 标准入口。

鉴权：OIDC access token。

请求参数：OIDC 标准 userinfo 请求。

响应参数：OIDC userinfo JSON。

## OIDC 面板辅助接口

### POST `/panel-api/v1/oidc/js-code`

功能：当前登录用户直接创建 OIDC 授权 code。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `response_type` | form/json | 否 | string | response type，默认 `code` |
| `client_id` | form/json | 是 | string | OIDC Client ID |
| `redirect_uri` | form/json | 是 | string | 回调地址 |
| `scope` | form/json | 否 | string | scope 字符串 |
| `state` | form/json | 否 | string | state |
| `nonce` | form/json | 否 | string | nonce |
| `response_mode` | form/json | 否 | string | response mode，默认 query |
| `code_challenge` | form/json | 否 | string | PKCE challenge |
| `code_challenge_method` | form/json | 否 | string | PKCE challenge 方法 |
| `prompt` | form/json | 否 | string | prompt |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | string | 授权 code |
| `state` | string | state，存在时返回 |
| `session_state` | string | session state，存在时返回 |

### POST `/panel-api/v1/oidc/redirect-uri`

功能：完成授权请求并生成 OIDC 回调 URL。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | form/json | 是 | string | OIDC 授权请求 ID |
| `callbackUrl` | form/json | 否 | string | Controller 接收但当前未使用 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `callbackUrl` | string | 生成后的 OIDC callback URL |

### POST `/panel-api/v1/code`

功能：旧版直接获取 OIDC 授权 code 接口，行为同 `/panel-api/v1/oidc/js-code`。

鉴权：需要用户 token。

请求参数：同 `/panel-api/v1/oidc/js-code`。

响应参数：同 `/panel-api/v1/oidc/js-code`。

### POST `/panel-api/v1/callback-url`

功能：旧版获取 OIDC callback URL 接口，行为同 `/panel-api/v1/oidc/redirect-uri`。

鉴权：需要用户 token。

请求参数：同 `/panel-api/v1/oidc/redirect-uri`。

响应参数：同 `/panel-api/v1/oidc/redirect-uri`。
