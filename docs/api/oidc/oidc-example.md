# OIDC Server Example

当前项目已提供最小可用的 OIDC Provider，支持：

- `authorization_code`
- `refresh_token`
- `dynamic client registration`
- `PKCE(S256)`
- dynamic client `read/update/delete`

## 1. 启用配置

在环境变量或配置文件中开启：

```yaml
oidc:
  enabled: true
  issuer: https://panel.example.com/panel-api/v1/oidc
  cookie_secure: true
  registration_enabled: true
  registration_access_token: change-me
  access_token_ttl: 3600s
  refresh_token_ttl: 720h
  clients:
    - client_id: demo-client
      client_secret: demo-secret
      token_endpoint_auth_method: client_secret_post
      allow_any_redirect_uri: false
      require_pkce: true
      redirect_uris:
        - https://oauthdebugger.com/debug
      scopes:
        - openid
        - profile
        - offline_access
```

如果 `issuer` 留空，服务会根据请求头里的 `Host` 和 `X-Forwarded-Proto` 动态推导。

`allow_any_redirect_uri: true` 时，客户端会跳过 `redirect_uris` 白名单校验；无论 `redirect_uris` 是否为空，任意 `redirect_uri` 都会被接受。该规则现在对静态配置客户端和动态注册客户端保持一致。

## 2. Discovery

```bash
curl -s https://panel.example.com/panel-api/v1/oidc/.well-known/openid-configuration | jq
```

## 3. 动态注册 Client

```bash
curl -sS -X POST \
  https://panel.example.com/panel-api/v1/oidc/register \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer change-me' \
  -d '{
    "client_name": "local-debug-client",
    "allow_any_redirect_uri": false,
    "redirect_uris": ["http://127.0.0.1:3000/callback"],
    "token_endpoint_auth_method": "none",
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile offline_access",
    "require_pkce": true
  }' | jq
```

返回里会包含：

- `client_id`
- `client_secret`，如果是 confidential client
- `redirect_uris`
- `grant_types`
- `scope`

动态注册的 client 会存储为 Kubernetes `Secret`：

- `Secret.metadata.name = client_id`
- `Secret.metadata.labels["w7.cc/oidc-client"] = "true"`

如果动态注册时设置 `allow_any_redirect_uri=true`，则 `redirect_uris` 可以为空，且授权阶段会允许任意 `redirect_uri`。如果未开启该选项，则 `redirect_uris` 仍然必填。

## 3.1 读取 Client

```bash
curl -sS \
  https://panel.example.com/panel-api/v1/oidc/register/<client_id> \
  -H 'Authorization: Bearer change-me' | jq
```

## 3.2 更新 Client

```bash
curl -sS -X PUT \
  https://panel.example.com/panel-api/v1/oidc/register/<client_id> \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer change-me' \
  -d '{
    "client_name": "local-debug-client-v2",
    "allow_any_redirect_uri": false,
    "redirect_uris": ["http://127.0.0.1:3000/callback"],
    "scope": "openid profile offline_access",
    "require_pkce": true
  }' | jq
```

## 3.3 删除 Client

```bash
curl -i -X DELETE \
  https://panel.example.com/panel-api/v1/oidc/register/<client_id> \
  -H 'Authorization: Bearer change-me'
```

## 4. Authorization Code + PKCE

浏览器访问：

```text
https://panel.example.com/panel-api/v1/oidc/authorize?response_type=code&client_id=<client_id>&redirect_uri=http%3A%2F%2F127.0.0.1%3A3000%2Fcallback&scope=openid%20profile%20offline_access&state=test-state&code_challenge=<challenge>&code_challenge_method=S256
```

登录成功后会跳转到 `redirect_uri`，并带上：

```text
code=...&state=test-state
```

## 5. 交换 Token

公开客户端：

```bash
curl -sS -X POST \
  https://panel.example.com/panel-api/v1/oidc/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'client_id=<client_id>' \
  --data-urlencode 'code=<code>' \
  --data-urlencode 'redirect_uri=http://127.0.0.1:3000/callback' \
  --data-urlencode 'code_verifier=<verifier>' | jq
```

机密客户端：

```bash
curl -sS -X POST \
  https://panel.example.com/panel-api/v1/oidc/token \
  -u '<client_id>:<client_secret>' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'code=<code>' \
  --data-urlencode 'redirect_uri=http://127.0.0.1:3000/callback' \
  --data-urlencode 'code_verifier=<verifier>' | jq
```

如果 scope 包含 `offline_access`，响应中会额外返回 `refresh_token`。

## 6. Refresh Token

```bash
curl -sS -X POST \
  https://panel.example.com/panel-api/v1/oidc/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=refresh_token' \
  --data-urlencode 'client_id=<client_id>' \
  --data-urlencode 'refresh_token=<refresh_token>' | jq
```

## 7. UserInfo

```bash
curl -sS \
  https://panel.example.com/panel-api/v1/oidc/userinfo \
  -H "Authorization: Bearer <access_token>" | jq
```

## 8. 脚本联调

仓库已附带一个可直接跑的脚本：

```bash
dev/scripts/oidc_test.sh
```

## 9. 导入 Postman

OpenAPI 文档位于：

```text
docs/oidc-openapi.yaml
```

可直接在 Postman 中使用 `Import` 导入这份 YAML。
