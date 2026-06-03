# OAuth 与 OIDC API

本文档说明 W7Panel 内置 OIDC Provider、OIDC 动态 Client 管理、面板内获取授权码、标准授权码流程和 Console OAuth 相关接口。OIDC 标准端点的响应格式和错误对象应遵循 OAuth/OIDC 协议，不要强行套用普通业务 JSON。

## 整体使用方式

OAuth/OIDC 接口分为三类：标准 OIDC Provider 端点、面板内已登录用户直接获取授权码的辅助接口，以及 Console OAuth 登录/绑定接口。开发时先判断调用方是第三方 OIDC Client、面板内微应用，还是 Console 登录流程。

### 基本流程

1. 第三方 Client 先读取 Discovery，获取 authorize、token、userinfo、jwks 等端点。
2. 标准授权码流程走 `/authorize`、`/authorize/login`、`/token`、`/userinfo`。
3. 面板内微应用已持有用户 token 时，可调用 `/js-code` 或兼容入口直接获取一次性 code。
4. 动态 Client 管理走 `/register`，是否可用取决于 OIDC 注册配置。
5. Console OAuth 登录、绑定和信息查询走 `/panel-api/v1/auth/console/*`，详细凭据字段见 [credentials.md](./credentials.md)。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 第三方发现 Provider | `/.well-known/openid-configuration` | 返回 Discovery 元数据 |
| 标准登录授权 | `/authorize`、`/authorize/login` | OAuth/OIDC 授权码流程 |
| 换 token | `/token` | 返回 access token、id token、refresh token |
| 查询用户信息 | `/userinfo` | 使用 OIDC access token |
| 面板内获取 code | `/js-code`、`/redirect-uri` | 已登录用户直接生成 code 或 callback URL |
| 动态 Client 管理 | `/register` | 创建、查询、更新、删除 Client |
| Console OAuth | `/auth/console/*` | Console 登录、绑定、代理和授权信息 |

### 使用边界

- OIDC 标准端点必须保持协议错误格式，不要包装成面板普通 JSON。
- OIDC `access_token` 只用于 OIDC 协议接口，不等同于面板用户 token。
- `redirect_uri`、client secret、code、access token、refresh token 都不能写入日志或前端可见调试输出。
- Console OAuth 和内置 OIDC Provider 是不同能力，字段和 token 不要混用。

## 通用约定

### 基础路径

OIDC Provider 基础前缀：

```text
/panel-api/v1/oidc
```

Discovery 地址：

```text
/panel-api/v1/oidc/.well-known/openid-configuration
```

### 配置项

OIDC 配置来自 `w7panel-server/config.yaml` 的 `oidc` 节点。

| 配置 | 默认值 | 说明 |
|------|--------|------|
| `OIDC_ENABLED` | `true` | 是否启用 OIDC Provider |
| `OIDC_ISSUER` | 空 | 固定 issuer；为空时按请求 Host/Forwarded 推断 |
| `OIDC_SIGNING_KEY_PEM` | 空 | RSA 签名私钥；为空时运行时生成 |
| `OIDC_ACCESS_TOKEN_TTL` | `3600s` | access token 有效期 |
| `OIDC_REFRESH_TOKEN_TTL` | `720h` | refresh token 有效期 |
| `OIDC_CODE_TTL` | `300s` | 授权请求和 code 有效期 |
| `OIDC_REGISTRATION_ENABLED` | `false` | 是否启用动态 Client 注册接口 |
| `OIDC_REGISTRATION_ACCESS_TOKEN` | 空 | 动态注册接口访问 token；为空时开启注册后不校验固定值 |
| `OIDC_INSECURE_ALLOW_ANY_REDIRECT_URI` | `true` | Provider 级任意 redirect_uri 配置 |
| `OIDC_CLIENT_ID` | `default` | 默认静态 Client ID |
| `OIDC_CLIENT_SECRET` | 空 | 默认静态 Client Secret |
| `OIDC_TOKEN_AUTH_METHOD` | `none` | 默认静态 Client 的 token 端点认证方式 |

默认静态 Client scopes：

```text
openid profile offline_access
```

### 错误格式

OIDC 标准端点通常返回协议错误：

```json
{
  "error": "invalid_request",
  "error_description": "..."
}
```

Provider 未启用时返回：

```json
{
  "error": "oidc disabled"
}
```

动态注册未启用时返回：

```json
{
  "error": "oidc registration disabled"
}
```

## 能力概览

| 能力 | 接口 | 说明 |
|------|------|------|
| Discovery | `ANY /.well-known/openid-configuration` | 暴露 OpenID Provider 元数据 |
| JWKS | `ANY /jwks` | 暴露 RS256 签名公钥 |
| 授权端点 | `ANY /authorize` | 标准 OAuth/OIDC 授权端点 |
| 内置登录页 | `GET/POST /authorize/login` | 授权码流程中的登录授权页 |
| Token | `ANY /token` | 标准 token 端点 |
| UserInfo | `ANY /userinfo` | 标准 userinfo 端点 |
| 动态 Client 管理 | `/register` | 注册、查询、更新、删除动态 Client |
| 面板内获取 code | `POST /js-code` | 已登录面板用户直接获取一次性 code |
| 面板内 callback URL | `POST /redirect-uri` | 完成授权请求并生成回调 URL |
| 旧版入口 | `/panel-api/v1/code`、`/panel-api/v1/callback-url` | 兼容入口 |
| Console OAuth | `/panel-api/v1/auth/console/*` | 面板绑定/登录 Console |

## OIDC 标准端点

### ANY `/panel-api/v1/oidc/.well-known/openid-configuration`

功能：返回 OIDC Discovery 元数据。

认证：无需用户 token。

请求参数：无。

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/oidc/.well-known/openid-configuration'
```

响应参数：OIDC Discovery JSON。常用字段如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| `issuer` | string | issuer；配置 `OIDC_ISSUER` 时使用固定值，否则从请求推断 |
| `authorization_endpoint` | string | `/panel-api/v1/oidc/authorize` |
| `token_endpoint` | string | `/panel-api/v1/oidc/token` |
| `userinfo_endpoint` | string | `/panel-api/v1/oidc/userinfo` |
| `jwks_uri` | string | `/panel-api/v1/oidc/jwks` |
| `response_types_supported` | array | 支持的 response type，当前 Client 支持 code、id_token 等 |
| `grant_types_supported` | array | 支持授权码；Client scopes 含 `offline_access` 时支持 refresh token |
| `scopes_supported` | array | `openid`、`profile`、`offline_access` |
| `id_token_signing_alg_values_supported` | array | 当前为 `RS256` |

响应示例：

```json
{
  "issuer": "http://localhost:8080/panel-api/v1/oidc",
  "authorization_endpoint": "http://localhost:8080/panel-api/v1/oidc/authorize",
  "token_endpoint": "http://localhost:8080/panel-api/v1/oidc/token",
  "userinfo_endpoint": "http://localhost:8080/panel-api/v1/oidc/userinfo",
  "jwks_uri": "http://localhost:8080/panel-api/v1/oidc/jwks",
  "scopes_supported": ["openid", "profile", "offline_access"]
}
```

### ANY `/panel-api/v1/oidc/jwks`

功能：返回 JWT 验签用公钥集合。

认证：无需用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `keys` | array | JWK 公钥列表 |
| `keys[].kid` | string | Key ID |
| `keys[].kty` | string | Key 类型，RSA |
| `keys[].alg` | string | 签名算法，`RS256` |
| `keys[].use` | string | 用途，`sig` |
| `keys[].n` | string | RSA modulus |
| `keys[].e` | string | RSA exponent |

响应示例：

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "key-id",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

### ANY `/panel-api/v1/oidc/authorize`

功能：标准授权端点，创建授权请求并引导用户登录授权。

认证：标准 OIDC 流程中无需面板 Bearer token；如果后续访问 `/authorize/login` 时已携带有效 Bearer token，可自动完成授权。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `client_id` | query/form | 是 | string | OIDC Client ID |
| `redirect_uri` | query/form | 是 | string | 回调地址；必须匹配 Client 配置，除非允许任意地址 |
| `response_type` | query/form | 是 | string | 常用 `code` |
| `scope` | query/form | 是 | string | 空格分隔，至少包含 `openid` |
| `state` | query/form | 否 | string | 客户端状态，回调时原样返回 |
| `nonce` | query/form | 否 | string | OIDC nonce |
| `response_mode` | query/form | 否 | string | 默认 query |
| `code_challenge` | query/form | 否 | string | PKCE challenge |
| `code_challenge_method` | query/form | 否 | string | PKCE 方法，推荐 `S256` |
| `prompt` | query/form | 否 | string | prompt 参数 |

请求示例：

```bash
curl -i 'http://localhost:8080/panel-api/v1/oidc/authorize?client_id=default&redirect_uri=http%3A%2F%2F127.0.0.1%3A3000%2Fcallback&response_type=code&scope=openid%20profile&state=abc'
```

响应：

| 情况 | 响应 |
|------|------|
| 需要登录 | 302 到 `/login?authRequestID=...`，由 Provider 映射到内置登录流程 |
| 请求非法 | OIDC 标准错误 |
| 登录完成 | 302 到 `redirect_uri?code=...&state=...` |

实现注意：Client 的 `LoginURL()` 返回 `/login?authRequestID=...`，项目实际注册的登录页是 `/panel-api/v1/oidc/authorize/login`。部署侧如没有额外 rewrite，需要重点验证完整授权码流程；面板内直接获取 code 可使用 `/js-code`。

### GET `/panel-api/v1/oidc/authorize/login`

功能：显示内置登录授权页，或在已携带 Bearer token 时自动完成授权并跳转回 callback。

认证：可选。携带有效 token 时自动授权；未携带时展示 HTML 登录页。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | query | 是 | string | 授权请求 ID |

响应：

| 情况 | 响应 |
|------|------|
| 已携带有效 Bearer token | 302 到授权回调 URL |
| 未携带 token | 返回 `text/html; charset=utf-8` 登录页 |
| 缺少 `authRequestID` | `400 {"error":"missing authRequestID"}` |
| `authRequestID` 无效 | `400 {"error":"invalid authRequestID","error_description":"..."}` |

### POST `/panel-api/v1/oidc/authorize/login`

功能：提交用户名密码完成授权请求。

认证：无需 Bearer token。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | query | 是 | string | 授权请求 ID |
| `username` | form | 是 | string | 用户名 |
| `password` | form | 是 | string | 密码 |

请求示例：

```bash
curl -i -X POST 'http://localhost:8080/panel-api/v1/oidc/authorize/login?authRequestID=<id>' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'username=admin' \
  --data-urlencode 'password=123456'
```

响应：

| 情况 | 响应 |
|------|------|
| 登录成功 | 302 到授权回调 URL |
| 登录失败 | 返回 HTML 登录页并显示错误 |
| 参数缺失 | 返回 HTML 登录页并显示“用户名和密码不能为空” |

### ANY `/panel-api/v1/oidc/token`

功能：标准 token 端点，用授权码或 refresh token 换取 token。

认证方式由 Client 的 `token_endpoint_auth_method` 决定：

| 值 | 请求方式 | 说明 |
|----|----------|------|
| `client_secret_basic` | HTTP Basic | `Authorization: Basic base64(client_id:client_secret)` |
| `client_secret_post` | form | `client_id`、`client_secret` 放在表单 |
| `none` | 无 secret | 公共/Native Client；PKCE 场景常用 |

授权码换 token 请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `grant_type` | form | 是 | string | `authorization_code` |
| `code` | form | 是 | string | 授权码 |
| `redirect_uri` | form | 视客户端要求 | string | 回调地址；当前适配器没有强制比对，但客户端仍应传入 |
| `client_id` | form | 条件必填 | string | `client_secret_post` 或 `none` 时传入 |
| `client_secret` | form | 条件必填 | string | `client_secret_post` 时传入 |
| `code_verifier` | form | PKCE 时必填 | string | PKCE verifier |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/oidc/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'client_id=default' \
  --data-urlencode 'code=<code>' \
  --data-urlencode 'redirect_uri=http://127.0.0.1:3000/callback'
```

refresh token 请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `grant_type` | form | 是 | string | `refresh_token` |
| `refresh_token` | form | 是 | string | 上一次 token 响应返回的 refresh token |
| `client_id` | form | 条件必填 | string | `client_secret_post` 或 `none` 时传入 |
| `client_secret` | form | 条件必填 | string | `client_secret_post` 时传入 |
| `scope` | form | 否 | string | 请求缩小 scope |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `access_token` | string | JWT access token |
| `token_type` | string | 通常为 `Bearer` |
| `expires_in` | int | access token 有效秒数 |
| `refresh_token` | string | scope 包含 `offline_access` 且 Client 支持时返回 |
| `id_token` | string | OIDC ID Token |
| `scope` | string | 实际授权 scope |

响应示例：

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh-token-value",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "scope": "openid profile offline_access"
}
```

实现注意：refresh token 采用轮换机制，使用旧 refresh token 换新 token 成功后，旧 refresh token 和旧 access token 会被删除。

### ANY `/panel-api/v1/oidc/userinfo`

功能：标准 userinfo 端点。

认证：

```http
Authorization: Bearer <access_token>
```

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `sub` | string | 用户主体，当前为 ServiceAccount/用户名 |
| `preferred_username` | string | 用户名 |
| `name` | string | scope 包含 `profile` 时返回用户名 |
| `role` | string | W7Panel 角色 |
| `is_founder` | bool | 是否创始人 |

响应示例：

```json
{
  "sub": "admin",
  "preferred_username": "admin",
  "name": "admin",
  "role": "founder",
  "is_founder": true
}
```

## 动态 OIDC Client 管理

这些接口用于动态注册 Client，和前端系统设置页直接通过 K8s 代理管理 `OIDCClient` CRD 是两套入口。动态注册接口使用 snake_case JSON；CRD 使用 camelCase 字段。

### 访问控制

动态注册接口要求：

| 条件 | 说明 |
|------|------|
| `OIDC_ENABLED=true` | OIDC Provider 启用 |
| `OIDC_REGISTRATION_ENABLED=true` | 动态注册启用 |
| `OIDC_REGISTRATION_ACCESS_TOKEN` | 非空时，请求 token 必须等于该值 |

注册管理 token 可通过通用 token 解析逻辑传入，推荐使用：

```http
Authorization: Bearer <OIDC_REGISTRATION_ACCESS_TOKEN>
```

如果 `OIDC_REGISTRATION_ACCESS_TOKEN` 为空，只要启用动态注册，当前实现会允许访问。

### POST `/panel-api/v1/oidc/register`

功能：注册动态 OIDC Client。

认证：registration access token。

请求类型：`application/json`

请求参数：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `redirect_uris` | 条件必填 | array<string> | 允许的回调地址；`allow_any_redirect_uri=false` 时必填 |
| `allow_any_redirect_uri` | 否 | bool | 是否允许任意回调地址 |
| `token_endpoint_auth_method` | 否 | string | `client_secret_basic`、`client_secret_post`、`none`；为空时按 secret 规则默认 |
| `grant_types` | 否 | array<string> | 仅支持 `["authorization_code"]` 或 `["authorization_code","refresh_token"]` |
| `scope` | 否 | string | 空格分隔；只保留 `openid`、`profile`、`offline_access` |
| `client_name` | 否 | string | Client 名称 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/oidc/register' \
  -H 'Authorization: Bearer <registration-token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "redirect_uris": ["https://app.example.com/callback"],
    "token_endpoint_auth_method": "client_secret_post",
    "grant_types": ["authorization_code", "refresh_token"],
    "scope": "openid profile offline_access",
    "client_name": "demo"
  }'
```

响应状态：`201 Created`

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `client_id` | string | 自动生成，格式类似 `oidc_xxx`，会规范化为小写和短横线 |
| `client_secret` | string | `token_endpoint_auth_method != none` 时返回 |
| `client_id_issued_at` | int64 | Client 创建时间，Unix 秒 |
| `client_secret_expires_at` | int64 | 当前固定为 `0`，表示不过期 |
| `redirect_uris` | array<string> | 回调地址 |
| `allow_any_redirect_uri` | bool | 是否允许任意回调 |
| `token_endpoint_auth_method` | string | token 端点认证方式 |
| `grant_types` | array<string> | 当前响应固定为 `["authorization_code","refresh_token"]` |
| `scope` | string | 实际保存的 scope |
| `client_name` | string | Client 名称 |

响应示例：

```json
{
  "client_id": "oidc-abcd1234efgh5678",
  "client_secret": "secret-value",
  "client_id_issued_at": 1710000000,
  "client_secret_expires_at": 0,
  "redirect_uris": ["https://app.example.com/callback"],
  "allow_any_redirect_uri": false,
  "token_endpoint_auth_method": "client_secret_post",
  "grant_types": ["authorization_code", "refresh_token"],
  "scope": "openid profile offline_access",
  "client_name": "demo"
}
```

### GET `/panel-api/v1/oidc/register/:clientId`

功能：查询动态 Client。

认证：registration access token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `clientId` | path | 是 | string | 动态 Client ID |

响应状态：`200 OK`

响应参数：同注册响应。

错误：

| 状态 | 响应 | 说明 |
|------|------|------|
| 404 | `{"error":"not_found","error_description":"client not found"}` | Client 不存在或不是动态 Client |

### PUT `/panel-api/v1/oidc/register/:clientId`

功能：更新动态 Client。

认证：registration access token。

请求类型：`application/json`

请求参数：同注册请求。更新逻辑如下：

| 字段 | 行为 |
|------|------|
| `redirect_uris` | 非空时覆盖 |
| `scope` | 非空时覆盖 |
| `client_name` | 非空时覆盖 |
| `allow_any_redirect_uri` | 总是覆盖 |
| `token_endpoint_auth_method` | 非空时覆盖 |
| `grant_types` | 当前校验请求合法性，但不单独持久化 |

响应状态：`200 OK`

响应参数：同注册响应。

### DELETE `/panel-api/v1/oidc/register/:clientId`

功能：删除动态 Client。

认证：registration access token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `clientId` | path | 是 | string | 动态 Client ID |

响应状态：`204 No Content`

响应体：无。

## OIDCClient CRD 管理

前端系统设置页 `OIDC密钥管理` 当前直接通过 K8s 代理管理 CRD：

```text
/apis/w7panel.w7.com/v1alpha1/namespaces/{namespace}/oidcclients
```

CRD 字段使用 camelCase：

| 字段 | 类型 | 说明 |
|------|------|------|
| `spec.enabled` | bool | 是否启用，前端可通过 merge-patch 切换 |
| `spec.clientId` | string | Client ID |
| `spec.clientName` | string | Client 名称 |
| `spec.clientSecret` | string | Client Secret |
| `spec.redirectUris` | array<string> | 回调地址 |
| `spec.allowAnyRedirectUri` | bool | 是否允许任意回调地址 |
| `spec.scopes` | array<string> | scopes，常见 `openid`、`profile`、`offline_access` |
| `spec.tokenEndpointAuthMethod` | string | 前端当前创建时固定为 `client_secret_post` |

示例：

```json
{
  "apiVersion": "w7panel.w7.com/v1alpha1",
  "kind": "OIDCClient",
  "metadata": {
    "name": "demo",
    "namespace": "default"
  },
  "spec": {
    "clientId": "demo",
    "clientName": "Demo",
    "clientSecret": "secret-value",
    "allowAnyRedirectUri": false,
    "redirectUris": ["https://app.example.com/callback"],
    "scopes": ["openid", "profile", "offline_access"],
    "tokenEndpointAuthMethod": "client_secret_post"
  }
}
```

注意：动态注册 API 写入的后端 CRD 也是 `OIDCClient`，但 API 请求/响应字段为 snake_case；前端 K8s 代理操作 CRD 时必须使用 camelCase。

## 面板内获取授权信息

### POST `/panel-api/v1/oidc/js-code`

功能：已登录面板用户直接获取 OIDC 一次性 code，常用于 Wujie 微应用事件 `getOidcCode`。

认证：`Authorization: Bearer <user-token>`

请求类型：支持 JSON 或 form。

请求参数：

| 字段 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `client_id` | JSON/form | 是 | string | OIDC Client ID，前端默认 `default` |
| `redirect_uri` | JSON/form | 是 | string | 回调地址 |
| `scope` | JSON/form | 否 | string | 空格分隔，前端默认 `openid` |
| `response_type` | JSON/form | 否 | string | 默认 `code` |
| `state` | JSON/form | 否 | string | 回调状态 |
| `nonce` | JSON/form | 否 | string | OIDC nonce |
| `response_mode` | JSON/form | 否 | string | 默认 query |
| `code_challenge` | JSON/form | 否 | string | PKCE challenge |
| `code_challenge_method` | JSON/form | 否 | string | PKCE 方法 |
| `prompt` | JSON/form | 否 | string | prompt |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/oidc/js-code' \
  -H 'Authorization: Bearer <user-token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "client_id": "default",
    "redirect_uri": "http://127.0.0.1:3000/callback",
    "scope": "openid profile",
    "state": "abc"
  }'
```

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | string | 授权码 |
| `state` | string | 原样返回 state；为空时省略 |
| `session_state` | string | 会话状态；为空时省略 |

响应示例：

```json
{
  "code": "authorization-code",
  "state": "abc",
  "session_state": "session-state"
}
```

前端事件示例见 [../frontend/wujie-events.md](../frontend/wujie-events.md)。

### POST `/panel-api/v1/oidc/redirect-uri`

功能：根据 `authRequestID` 完成授权请求并构造回调 URL。

认证：`Authorization: Bearer <user-token>`

请求类型：支持 JSON 或 form。

请求参数：

| 字段 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `authRequestID` | JSON/form | 是 | string | 授权请求 ID |
| `callbackUrl` | JSON/form | 否 | string | 兼容字段；当前实现未使用该值 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/oidc/redirect-uri' \
  -H 'Authorization: Bearer <user-token>' \
  -H 'Content-Type: application/json' \
  -d '{"authRequestID":"<id>"}'
```

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `callbackUrl` | string | 完成授权后的回调地址，包含 `code`、`state` 等 query |

响应示例：

```json
{
  "callbackUrl": "https://app.example.com/callback?code=authorization-code&state=abc"
}
```

### 旧版兼容入口

| 方法 | 路径 | 对应新接口 | 认证 |
|------|------|------------|------|
| `POST` | `/panel-api/v1/code` | `/panel-api/v1/oidc/js-code` | 用户 token |
| `POST` | `/panel-api/v1/callback-url` | `/panel-api/v1/oidc/redirect-uri` | 用户 token |

请求/响应参数与新接口相同。

## Console OAuth 与凭据

Console OAuth 详细请求/响应字段见 [credentials.md](credentials.md)。本节只列出与 OAuth/OIDC 相关的入口。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `GET` | `/panel-api/v1/auth/console/oauth` | 无用户 token | 生成或跳转 Console OAuth URL |
| `GET` | `/panel-api/v1/auth/console/login` | 无用户 token | 使用 Console OAuth code 登录面板 |
| `GET` | `/panel-api/v1/auth/console/bind` | 用户 token | 当前面板用户绑定 Console OAuth code |
| `GET` | `/panel-api/v1/auth/console/info` | 用户 token | 获取 Console 注册状态、第三方 CD token、license 等 |
| `GET` | `/panel-api/v1/auth/console/code/:code` | 用户 token | 代理优惠码请求 |
| `ANY` | `/panel-api/v1/auth/console/proxy/*path` | founder token | Founder 代理 Console API |

## access_token 使用边界

微应用 props 中的 `access_token` 是面板登录用户自身维护的 access token，只能用于获取用户信息，不能准确定位 appid。需要定位具体应用时，应使用 `group`、应用配置、MicroApp 信息或后端返回的应用上下文。

`/userinfo` 返回的 `role`、`is_founder` 是 W7Panel 附加 claims，不属于最小 OIDC 标准字段；第三方 Client 使用时应做好兼容。

## 开发检查

- OIDC 标准端点保持协议兼容，错误格式遵守 OAuth/OIDC 约定。
- `redirect_uri` 必须校验，避免开放重定向；只有明确需要时才允许 `allow_any_redirect_uri`。
- 不在日志、URL、响应体、前端 localStorage 中输出完整 code、access token、refresh token、client secret。
- 动态注册 API 使用 snake_case，`OIDCClient` CRD 使用 camelCase，前后端不要混用字段名。
- Founder 级 Console 代理接口必须保留更高权限检查。
- 修改 OIDC Client CRD 后，同步检查系统设置页 `w7panel-ui/src/views/system/access-key/oidc-key.vue`。
- 修改 `js-code` 后，同步检查 Wujie 事件 `getOidcCode` 和 [../frontend/wujie-events.md](../frontend/wujie-events.md)。
