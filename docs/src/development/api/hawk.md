# Hawk 签名认证

本文档说明 `w7panel-server/common/middleware/hawk.go` 中 Hawk 中间件的认证流程、凭据来源和客户端签名方式。Hawk 适用于服务端到服务端调用，使用 `clientId/clientSecret` 计算请求签名，不替代面板用户 Bearer token。

## 当前实现状态

后端已实现 `middleware.Hawk{}.Process`，但默认路由中没有启用 Hawk 保护的业务接口。`w7panel-server/app/auth/provider.go` 中的 `registerHawkTestRoute(localApiGroup, middleware.Hawk{}.Process)` 当前为注释状态。

接入新接口时，在路由注册处显式挂载中间件：

```go
group.POST("/example", middleware.Hawk{}.Process, controller.Example{}.Handle)
```

`OPTIONS` 预检请求会直接放行，其他方法必须携带合法的 Hawk `Authorization` 请求头。

## 服务端认证流程

1. 从 `Authorization` 头解析 Hawk header。
2. 使用 header 中的 `id` 作为 `clientId` 查找 Kubernetes `ApiClient` 资源。
3. 读取 `ApiClient.spec.clientSecret` 作为签名密钥。
4. 调用 `go.mozilla.org/hawk` 校验 MAC、时间戳和请求信息。
5. 认证成功后记录访问时间，并把客户端信息写入 Gin 上下文。

当前默认时间偏差为 5 分钟，可通过 `middleware.Hawk{MaxSkew: ...}` 调整。

## ApiClient 凭据

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
|------|------|
| `spec.clientId` | Hawk header 中的 `id`，用于查找客户端。 |
| `spec.clientSecret` | 签名密钥，必须非空。 |
| `spec.clientName` | 客户端展示名称，认证成功后写入 Gin 上下文。 |
| `spec.enabled` | 启用状态字段。当前 Hawk 中间件未校验该字段，启停逻辑需要接入方自行补充或扩展中间件。 |
| `status.lastAccessedAt` | 最近访问时间，认证成功后异步写回。 |

中间件按当前运行命名空间加载 `ApiClient` 列表，并通过缓存按 `clientId` 查询。缓存由 `common/service/k8s/apiclient` 维护，默认 1 分钟周期刷新最近访问时间。

## 请求头格式

请求必须携带：

```http
Authorization: Hawk id="<clientId>", mac="<base64-hmac>", nonce="<nonce>", ts="<unix-seconds>"
```

必填属性：

| 属性 | 说明 |
|------|------|
| `id` | 客户端 ID，对应 `ApiClient.spec.clientId`。 |
| `mac` | 使用 `clientSecret` 计算得到的 Base64 HMAC-SHA256 签名。 |
| `nonce` | 单次请求随机串，建议每次请求重新生成。 |
| `ts` | Unix 秒级时间戳，默认允许与服务端时间相差 5 分钟。 |

当前中间件调用 `hawk.NewAuthFromRequest(req, resolver.lookupCredentials, nil)`，没有启用 payload 校验。因此默认只校验 HTTP 方法、路径、查询串、Host、端口和 Hawk header 属性，不校验请求体内容。

## 认证成功后的上下文

认证成功后，中间件写入以下 Gin 上下文 key：

| Key | 值 |
|-----|-----|
| `api_client_id` | `ApiClient.spec.clientId` |
| `api_client_name` | `ApiClient.spec.clientName` |
| `hawk_client_id` | Hawk 凭据中的 `id` |

业务 controller 可从上下文读取调用方身份，用于审计、限流或授权判断。

## 失败响应

认证失败时中间件终止请求并返回：

```http
HTTP/1.1 401 Unauthorized
```

常见原因：

| 原因 | 处理方式 |
|------|----------|
| 缺少 `Authorization` 头 | 补充 `Authorization: Hawk ...`。 |
| `id` 为空或找不到对应 `ApiClient` | 检查 `ApiClient.spec.clientId` 和资源所在命名空间。 |
| `clientSecret` 为空 | 给 `ApiClient.spec.clientSecret` 设置非空密钥。 |
| `mac` 不正确 | 确认方法、Host、端口、路径、查询串、时间戳、nonce 和密钥完全一致。 |
| `ts` 超过允许时间偏差 | 同步客户端和服务端时间。 |

## 推荐客户端类库

优先使用成熟 Hawk 类库，不建议业务代码直接拼接标准化签名串。类库会处理 URL、Host、端口、时间戳、nonce、转义和 header 格式，能减少跨语言差异。

| 语言 | 推荐方式 | 说明 |
|------|----------|------|
| Node.js / JavaScript | `@hapi/hawk` | 适合 Node 服务端、构建工具和命令行脚本。 |
| Go | `go.mozilla.org/hawk` | 与服务端校验库一致，优先使用。 |
| Python | `mohawk` | Python 项目可直接生成 Hawk `Authorization` header。 |
| 其他语言 | Hawk HTTP Authentication Scheme 的成熟实现 | 找不到可靠类库时，再参考手写签名算法。 |

## Node.js 示例

安装：

```bash
npm install @hapi/hawk
```

GET 请求：

```js
const Hawk = require("@hapi/hawk");

const credentials = {
  id: "client-1",
  key: "secret-1",
  algorithm: "sha256",
};

const url = "https://example.com/panel-api/v1/example";
const method = "GET";

const { field } = Hawk.client.header(url, method, { credentials });
console.log(field);
```

调用：

```bash
AUTH="$(node sign-hawk.js)"
curl 'https://example.com/panel-api/v1/example' \
  -H "Authorization: $AUTH"
```

POST JSON：

```js
const Hawk = require("@hapi/hawk");

const credentials = {
  id: "client-1",
  key: "secret-1",
  algorithm: "sha256",
};

const url = "https://example.com/panel-api/v1/example";
const method = "POST";
const body = JSON.stringify({ name: "demo" });

const { field } = Hawk.client.header(url, method, { credentials });

await fetch(url, {
  method,
  headers: {
    Authorization: field,
    "Content-Type": "application/json",
  },
  body,
});
```

当前服务端未校验 payload hash，因此示例没有把 body 纳入 Hawk hash。若未来开启 payload 校验，客户端和服务端必须同时启用请求体 hash。

## Go 示例

安装：

```bash
go get go.mozilla.org/hawk
```

生成请求头：

```go
package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"

	"go.mozilla.org/hawk"
)

func main() {
	rawURL := "https://example.com/panel-api/v1/example"
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Host:   u.Host,
	}

	auth := hawk.NewRequestAuth(req, &hawk.Credentials{
		ID:   "client-1",
		Key:  "secret-1",
		Hash: sha256.New,
	}, 0)

	fmt.Println(auth.RequestHeader())
}
```

发起请求：

```go
req, err := http.NewRequest("GET", "https://example.com/panel-api/v1/example", nil)
if err != nil {
	return err
}

auth := hawk.NewRequestAuth(req, &hawk.Credentials{
	ID:   "client-1",
	Key:  "secret-1",
	Hash: sha256.New,
}, 0)
req.Header.Set("Authorization", auth.RequestHeader())

resp, err := http.DefaultClient.Do(req)
```

## Python 示例

安装：

```bash
pip install mohawk
```

生成请求头：

```python
from mohawk import Sender

credentials = {
    "id": "client-1",
    "key": "secret-1",
    "algorithm": "sha256",
}

url = "https://example.com/panel-api/v1/example"
sender = Sender(credentials, url, "GET", content="", content_type="")

print(sender.request_header)
```

使用 `requests` 调用：

```python
import requests
from mohawk import Sender

credentials = {
    "id": "client-1",
    "key": "secret-1",
    "algorithm": "sha256",
}

url = "https://example.com/panel-api/v1/example"
sender = Sender(credentials, url, "GET", content="", content_type="")

resp = requests.get(url, headers={"Authorization": sender.request_header})
print(resp.status_code)
print(resp.text)
```

## 手写签名

没有可用 Hawk SDK，或需要排查 SDK 生成结果时，可以按本节手写签名。当前 W7Panel 服务端只校验 header MAC，不校验 payload hash，因此所有手写示例都不计算请求体摘要。

### 公共规则

所有语言都按同一套规则生成 `Authorization`：

```http
Authorization: Hawk id="<clientId>", mac="<mac>", nonce="<nonce>", ts="<unix-seconds>"
```

签名输入是 `go.mozilla.org/hawk` 的 Hawk header MAC 标准化串。W7Panel 当前不使用 `hash`、`ext`、`app`、`dlg`，所以默认只保留 `hash` 和 `ext` 两个空字段：

```text
hawk.1.header
<ts>
<nonce>
<METHOD>
<resource>
<host>
<port>


```

字段说明：

| 行 | 字段 | 说明 |
|----|------|------|
| 1 | `hawk.1.header` | Hawk header MAC 固定前缀。 |
| 2 | `<ts>` | Unix 秒级时间戳。 |
| 3 | `<nonce>` | 随机串。 |
| 4 | `<METHOD>` | HTTP 方法大写。 |
| 5 | `<resource>` | URL path + query，无 query 时只写 path。 |
| 6 | `<host>` | 服务端收到的 Host，不包含端口。建议使用小写域名并保持请求头一致。 |
| 7 | `<port>` | 请求端口，HTTPS 默认 `443`，HTTP 默认 `80`。 |
| 8 | `hash` | 当前未启用 payload 校验，留空。 |
| 9 | `ext` | 扩展字段，当前留空。 |

说明：`go.mozilla.org/hawk` 只有在 header 携带非空 `app` 时才会把 `app` 和 `dlg` 追加到标准化串。W7Panel `ApiClient` 当前没有使用 `app/dlg`，手写签名不要添加这两个字段，也不要在 MAC 输入里追加两个空行。

计算步骤：

1. 使用 `clientSecret` 对标准化串做 `HMAC-SHA256`。
2. 对 HMAC 结果做 Base64 编码。
3. 将结果写入 Hawk header 的 `mac` 字段。

实现时要特别注意：

| 项 | 要求 |
|----|------|
| `METHOD` | 必须大写，例如 `GET`、`POST`。 |
| `resource` | 必须是 path + query，例如 `/panel-api/v1/example?a=1`。没有 query 时只写 path。 |
| `host` | 必须和服务端收到的 Host 一致，不包含端口。建议请求 URL 和 Host header 都使用小写域名。 |
| `port` | URL 有显式端口时使用显式端口；否则 HTTPS 为 `443`，HTTP 为 `80`。 |
| 换行 | 每一行用 `\n`，标准化串末尾也必须有一个 `\n`。默认只有 `hash`、`ext` 两个空字段。 |
| 字符集 | 用 UTF-8 字节计算 HMAC。 |
| nonce | 建议使用字母、数字、短横线、下划线，避免 header 转义问题。 |

下面示例使用相同输入：

```text
method       = GET
url          = https://example.com/panel-api/v1/example?a=1
clientId     = client-1
clientSecret = secret-1
```

生成的 header 可直接用于：

```bash
curl 'https://example.com/panel-api/v1/example?a=1' \
  -H "$AUTHORIZATION"
```

其中 `$AUTHORIZATION` 形如：

```http
Authorization: Hawk id="client-1", mac="<base64-hmac>", nonce="<nonce>", ts="<unix-seconds>"
```

### Java 标准库

适用于没有 Hawk SDK 的 Java 服务。依赖 JDK 自带的 `javax.crypto` 和 `java.net.URI`。

```java
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Base64;
import java.util.UUID;

public class HawkSigner {
    public static String sign(String method, String rawUrl, String clientId, String clientSecret) throws Exception {
        URI uri = URI.create(rawUrl);
        String ts = String.valueOf(Instant.now().getEpochSecond());
        String nonce = UUID.randomUUID().toString().replace("-", "");
        String resource = uri.getRawPath();
        if (resource == null || resource.isEmpty()) {
            resource = "/";
        }
        if (uri.getRawQuery() != null && !uri.getRawQuery().isEmpty()) {
            resource += "?" + uri.getRawQuery();
        }
        String host = uri.getHost().toLowerCase();
        int port = uri.getPort() > 0 ? uri.getPort() : ("https".equalsIgnoreCase(uri.getScheme()) ? 443 : 80);

        String normalized = String.join("\n",
                "hawk.1.header",
                ts,
                nonce,
                method.toUpperCase(),
                resource,
                host,
                String.valueOf(port),
                "",
                ""
        ) + "\n";

        Mac mac = Mac.getInstance("HmacSHA256");
        mac.init(new SecretKeySpec(clientSecret.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
        String signature = Base64.getEncoder().encodeToString(mac.doFinal(normalized.getBytes(StandardCharsets.UTF_8)));

        return "Authorization: Hawk id=\"" + clientId + "\", mac=\"" + signature + "\", nonce=\"" + nonce + "\", ts=\"" + ts + "\"";
    }

    public static void main(String[] args) throws Exception {
        System.out.println(sign("GET", "https://example.com/panel-api/v1/example?a=1", "client-1", "secret-1"));
    }
}
```

### PHP 标准库

适用于 Laravel、Symfony 或普通 PHP 项目，无需安装 Hawk 包。

```php
<?php

function hawk_sign($method, $rawUrl, $clientId, $clientSecret) {
    $parts = parse_url($rawUrl);
    $scheme = strtolower($parts['scheme'] ?? 'http');
    $host = strtolower($parts['host']);
    $port = $parts['port'] ?? ($scheme === 'https' ? 443 : 80);
    $path = $parts['path'] ?? '/';
    $resource = $path . (isset($parts['query']) ? '?' . $parts['query'] : '');
    $ts = (string) time();
    $nonce = bin2hex(random_bytes(12));

    $normalized = implode("\n", [
        'hawk.1.header',
        $ts,
        $nonce,
        strtoupper($method),
        $resource,
        $host,
        (string) $port,
        '',
        '',
    ]) . "\n";

    $mac = base64_encode(hash_hmac('sha256', $normalized, $clientSecret, true));

    return 'Authorization: Hawk id="' . $clientId . '", mac="' . $mac . '", nonce="' . $nonce . '", ts="' . $ts . '"';
}

echo hawk_sign('GET', 'https://example.com/panel-api/v1/example?a=1', 'client-1', 'secret-1') . PHP_EOL;
```

## 手写签名排查清单

签名失败时优先检查：

- `Authorization` header 必须以 `Hawk ` 开头。
- `id` 必须等于 `ApiClient.spec.clientId`。
- `clientSecret` 必须与服务端 `ApiClient.spec.clientSecret` 完全一致，不要多空格或换行。
- `ts` 是 Unix 秒，不是毫秒。
- 客户端和服务端时间差默认不能超过 5 分钟。
- `METHOD` 必须是实际请求方法的大写形式。
- `resource` 必须包含 query，且 query 顺序和编码要与实际请求一致。
- `host` 必须是服务端收到的 Host；经过网关或反向代理时尤其要检查，大小写也要保持一致。
- `port` 必须是服务端收到的端口；默认 HTTPS 为 `443`，HTTP 为 `80`。
- 标准化串最后必须保留换行符，默认末尾只有 `hash`、`ext` 两个空字段。
- 不要追加空的 `app/dlg` 两行；`go.mozilla.org/hawk` 只有在 `app` 非空时才追加这两行。
- 当前 W7Panel 没有校验 body hash，手写 header 不要添加 `hash`，除非服务端同步开启 payload 校验。

## 接入注意事项

- 签名时必须使用实际发送请求中的 Host、端口、路径和查询串。
- 反向代理如果改写 Host、端口、路径前缀或查询串，会导致签名失败。
- `nonce` 当前未做重放校验，但仍应每次请求生成新值，便于后续增强。
- `spec.enabled` 当前未参与中间件校验，新增正式业务接口前应明确是否需要启停检查。
- 当前不校验请求体摘要；需要防止 body 篡改时，应扩展中间件 payload 校验并要求客户端签 `hash` 字段。
