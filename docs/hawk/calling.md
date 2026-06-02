# W7Panel Hawk 调用文档

## 适用范围

W7Panel 的 Hawk 中间件用于给服务端接口增加基于 `clientId/clientSecret` 的请求签名认证。当前代码中已注册的验证接口为：

```text
GET /panel-api/v1/auth/hawk-test
```

业务接口接入 `middleware.Hawk{}.Process` 后，除 `OPTIONS` 预检请求外，所有请求都必须携带合法的 Hawk `Authorization` 请求头。

## 客户端凭据

Hawk 客户端凭据来自 Kubernetes `ApiClient` 资源：

```yaml
apiVersion: w7panel.w7.com/v1alpha1
kind: ApiClient
metadata:
  name: demo-client
  namespace: w7panel
spec:
  enabled: true
  clientId: client-1
  clientName: demo-client
  clientSecret: secret-1
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `spec.clientId` | Hawk `Authorization` 头中的 `id`，用于查找客户端。 |
| `spec.clientSecret` | 签名密钥，必须非空。服务端使用它重新计算 MAC。 |
| `spec.clientName` | 客户端展示名称。认证成功后会写入 Gin 上下文。 |
| `spec.enabled` | 客户端启用状态字段。当前 Hawk 中间件代码只校验客户端存在和 `clientSecret` 非空，不校验该字段。 |
| `status.lastAccessedAt` | 认证成功后由服务端异步更新最近访问时间。 |

服务端按当前命名空间加载 `ApiClient` 列表，并按 `spec.clientId` 建立缓存。缓存会被 `ApiClient` webhook 同步更新，成功请求会记录最近访问时间。

## 请求头

每个非 `OPTIONS` 请求必须携带：

```http
Authorization: Hawk id="<clientId>", mac="<base64-hmac>", nonce="<随机串>", ts="<unix秒级时间戳>"
```

必填属性：

| 属性 | 说明 |
| --- | --- |
| `id` | 客户端 ID，对应 `ApiClient.spec.clientId`。 |
| `mac` | 使用 `clientSecret` 计算得到的 Base64 HMAC-SHA256 签名。 |
| `nonce` | 单次请求随机串。建议每次请求重新生成。 |
| `ts` | Unix 秒级时间戳。服务端默认允许与服务器时间相差 5 分钟。 |

当前中间件调用 `hawk.NewAuthFromRequest(req, resolver.lookupCredentials, nil)`，没有启用 payload 校验。因此默认只签 HTTP 方法、路径、查询串、Host、端口和头部属性，不校验请求体内容。

## 成功响应示例

调用测试接口：

```bash
curl 'https://example.com/panel-api/v1/auth/hawk-test' \
  -H 'Authorization: Hawk id="client-1", mac="<mac>", nonce="nonce-123", ts="1760000000"'
```

成功返回：

```json
{
  "ok": true,
  "apiClientId": "client-1",
  "apiClientName": "demo-client",
  "hawkClientId": "client-1"
}
```

中间件认证成功后会向 Gin 上下文写入：

| Key | 值 |
| --- | --- |
| `api_client_id` | `ApiClient.spec.clientId` |
| `api_client_name` | `ApiClient.spec.clientName` |
| `hawk_client_id` | Hawk 凭据中的 `id` |

## 失败响应

认证失败时中间件会终止请求并返回：

```http
HTTP/1.1 401 Unauthorized
```

常见原因：

| 原因 | 处理方式 |
| --- | --- |
| 缺少 `Authorization` 头 | 补充 `Authorization: Hawk ...`。 |
| `id` 为空或找不到对应 `ApiClient` | 检查 `ApiClient.spec.clientId` 和所在命名空间。 |
| `clientSecret` 为空 | 给 `ApiClient.spec.clientSecret` 设置非空密钥。 |
| `mac` 不正确 | 确认签名串、Host、端口、路径、查询串、方法和密钥完全一致。 |
| `ts` 超过允许时间偏差 | 同步客户端与服务端时间，默认允许 5 分钟偏差。 |

## 调用注意事项

- 签名时使用请求中的实际 `Host` 和端口。HTTPS 默认端口为 `443`，HTTP 默认端口为 `80`。
- URL 查询串必须参与签名，且要和实际发送的请求完全一致。
- 反向代理如果修改 `Host`、端口、路径前缀或查询串，会导致签名不一致。
- `nonce` 建议每次请求生成新值，便于未来增加重放保护。
- 当前实现没有校验请求体摘要；如果业务需要防止 body 被篡改，需要扩展中间件传入 payload 校验配置并要求客户端签 `hash` 字段。
