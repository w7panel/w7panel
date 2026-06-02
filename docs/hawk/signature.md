# W7Panel Hawk 签名文档

## 推荐接入方式

客户端签名优先使用已有 Hawk 类库，不建议业务代码直接拼接标准化签名串。类库会处理 URL、Host、端口、时间戳、nonce、转义和 header 格式，能减少跨语言实现差异。

| 语言 | 推荐方式 | 说明 |
| --- | --- | --- |
| Node.js / JavaScript | `@hapi/hawk` | 推荐用于前端构建工具、Node 服务端、命令行脚本。 |
| Go | `go.mozilla.org/hawk` | W7Panel 服务端当前使用同一个库校验，请优先使用。 |
| Python | `mohawk` | Python 项目可直接生成 Hawk `Authorization` header。 |
| 其他语言 | 优先寻找实现 Hawk HTTP Authentication Scheme 的成熟库 | 找不到可靠类库时，再使用本文的“手写签名兜底算法”。 |

要求所有语言生成的 header 都满足：

```http
Authorization: Hawk id="<clientId>", mac="<mac>", nonce="<nonce>", ts="<timestamp>"
```

当前服务端没有启用 payload 校验，所以客户端示例默认不传 `hash`。如果未来服务端开启请求体校验，客户端需要按类库要求补充 body 和 content-type，使 header 中带上 `hash`。

## Node.js / JavaScript

安装：

```bash
npm install @hapi/hawk
```

生成 header：

```js
const Hawk = require("@hapi/hawk");

const credentials = {
  id: "client-1",
  key: "secret-1",
  algorithm: "sha256",
};

const url = "https://example.com/panel-api/v1/auth/hawk-test";
const method = "GET";

const { field } = Hawk.client.header(url, method, { credentials });

console.log(field);
```

调用：

```bash
AUTH="$(node sign-hawk.js)"
curl 'https://example.com/panel-api/v1/auth/hawk-test' \
  -H "Authorization: $AUTH"
```

POST JSON 示例：

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

const { field } = Hawk.client.header(url, method, {
  credentials,
});

await fetch(url, {
  method,
  headers: {
    Authorization: field,
    "Content-Type": "application/json",
  },
  body,
});
```

说明：当前 W7Panel Hawk 中间件不校验 payload hash，因此示例没有把 body 纳入 Hawk hash。请求体是否参与签名必须与服务端能力保持一致。

## Go

安装：

```bash
go get go.mozilla.org/hawk
```

生成 header：

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
	rawURL := "https://example.com/panel-api/v1/auth/hawk-test"
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
req, err := http.NewRequest("GET", "https://example.com/panel-api/v1/auth/hawk-test", nil)
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

## Python

安装：

```bash
pip install mohawk
```

生成 header：

```python
from mohawk import Sender

credentials = {
    "id": "client-1",
    "key": "secret-1",
    "algorithm": "sha256",
}

url = "https://example.com/panel-api/v1/auth/hawk-test"
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

url = "https://example.com/panel-api/v1/auth/hawk-test"
sender = Sender(credentials, url, "GET", content="", content_type="")

resp = requests.get(url, headers={"Authorization": sender.request_header})
print(resp.status_code)
print(resp.text)
```

## 手写签名兜底算法

仅当目标语言没有可靠 Hawk 类库，或需要排查类库生成结果时，才使用下面的算法手动生成 header。

## 签名算法

W7Panel 当前使用 `go.mozilla.org/hawk` 校验 Hawk header 签名。测试用例中的签名算法如下：

- HMAC 算法：`HMAC-SHA256`
- 输出编码：`Base64`
- 密钥：`ApiClient.spec.clientSecret`
- 时间戳：Unix 秒级时间戳
- 默认时间偏差：5 分钟

## 标准化签名串

当前测试和服务端 Hawk header 认证使用如下标准化串：

```text
hawk.1.header
<ts>
<nonce>
<METHOD>
<resource>
<host>
<port>




```

注意：末尾包含 4 个空字段，整段字符串最后还有一个换行符。字段含义如下：

| 行 | 字段 | 说明 |
| --- | --- | --- |
| 1 | `hawk.1.header` | Hawk header MAC 固定前缀。 |
| 2 | `<ts>` | Unix 秒级时间戳，例如 `1760000000`。 |
| 3 | `<nonce>` | 随机串，例如 `nonce-123`。 |
| 4 | `<METHOD>` | HTTP 方法大写，例如 `GET`、`POST`。 |
| 5 | `<resource>` | URL path + query，例如 `/panel-api/v1/auth/hawk-test?a=1`。无 query 时只写 path。 |
| 6 | `<host>` | 请求 Host 小写，不包含端口，例如 `example.com`。 |
| 7 | `<port>` | 请求端口。HTTPS 默认 `443`，HTTP 默认 `80`。 |
| 8 | `hash` | 当前未启用 payload 校验，留空。 |
| 9 | `ext` | 扩展字段，当前示例留空。 |
| 10 | `app` | 应用字段，当前示例留空。 |
| 11 | `dlg` | 委托字段，当前示例留空。 |

示例标准化串：

```text
hawk.1.header
1760000000
nonce-123
GET
/panel-api/v1/auth/hawk-test
example.com
443




```

## 计算 MAC

伪代码：

```text
mac = base64(hmac_sha256(clientSecret, normalizedString))
```

生成请求头：

```text
Authorization: Hawk id="client-1", mac="<mac>", nonce="nonce-123", ts="1760000000"
```

## Node.js 手写示例

```js
const crypto = require("crypto");

function defaultPort(url) {
  if (url.port) return url.port;
  return url.protocol === "https:" ? "443" : "80";
}

function signHawk({ method, requestUrl, clientId, clientSecret, nonce, timestamp }) {
  const url = new URL(requestUrl);
  const resource = `${url.pathname}${url.search}`;
  const host = url.hostname.toLowerCase();
  const port = defaultPort(url);

  const normalized = [
    "hawk.1.header",
    String(timestamp),
    nonce,
    method.toUpperCase(),
    resource,
    host,
    port,
    "",
    "",
    "",
    "",
  ].join("\n") + "\n";

  const mac = crypto
    .createHmac("sha256", clientSecret)
    .update(normalized)
    .digest("base64");

  return `Hawk id="${clientId}", mac="${mac}", nonce="${nonce}", ts="${timestamp}"`;
}

const authorization = signHawk({
  method: "GET",
  requestUrl: "https://example.com/panel-api/v1/auth/hawk-test",
  clientId: "client-1",
  clientSecret: "secret-1",
  nonce: "nonce-123",
  timestamp: Math.floor(Date.now() / 1000),
});

console.log(authorization);
```

调用：

```bash
AUTH="$(node sign-hawk.js)"
curl 'https://example.com/panel-api/v1/auth/hawk-test' \
  -H "Authorization: $AUTH"
```

## Go 手写示例

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func defaultPort(u *url.URL) string {
	if u.Port() != "" {
		return u.Port()
	}
	if strings.EqualFold(u.Scheme, "https") {
		return "443"
	}
	return "80"
}

func signHawk(method, rawURL, clientID, clientSecret, nonce string, requestTime time.Time) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	resource := u.EscapedPath()
	if u.RawQuery != "" {
		resource += "?" + u.RawQuery
	}

	ts := strconv.FormatInt(requestTime.UTC().Unix(), 10)
	normalized := strings.Join([]string{
		"hawk.1.header",
		ts,
		nonce,
		strings.ToUpper(method),
		resource,
		strings.ToLower(u.Hostname()),
		defaultPort(u),
		"",
		"",
		"",
		"",
	}, "\n") + "\n"

	mac := hmac.New(sha256.New, []byte(clientSecret))
	_, _ = mac.Write([]byte(normalized))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf(`Hawk id="%s", mac="%s", nonce="%s", ts="%s"`, clientID, signature, nonce, ts), nil
}
```

## curl 调试示例

下面的命令使用 Node.js 生成签名并调用测试接口：

```bash
export URL='https://example.com/panel-api/v1/auth/hawk-test'
export CLIENT_ID='client-1'
export CLIENT_SECRET='secret-1'
export NONCE="$(date +%s)-$RANDOM"
export TS="$(date +%s)"

AUTH="$(node -e '
const crypto = require("crypto");
const url = new URL(process.env.URL);
const port = url.port || (url.protocol === "https:" ? "443" : "80");
const normalized = [
  "hawk.1.header",
  process.env.TS,
  process.env.NONCE,
  "GET",
  `${url.pathname}${url.search}`,
  url.hostname.toLowerCase(),
  port,
  "",
  "",
  "",
  "",
].join("\n") + "\n";
const mac = crypto.createHmac("sha256", process.env.CLIENT_SECRET).update(normalized).digest("base64");
process.stdout.write(`Hawk id="${process.env.CLIENT_ID}", mac="${mac}", nonce="${process.env.NONCE}", ts="${process.env.TS}"`);
')"

curl "$URL" -H "Authorization: $AUTH"
```

## 排查清单

签名失败时优先核对：

- `Authorization` 必须以 `Hawk ` 开头。
- `id` 必须等于 `ApiClient.spec.clientId`。
- `clientSecret` 必须与服务端 `ApiClient.spec.clientSecret` 完全一致。
- `METHOD` 必须大写。
- `resource` 必须包含 query，且 query 顺序、编码与实际请求一致。
- `host` 必须小写，且与服务端收到的 Host 一致。
- `port` 必须是服务端收到的端口；默认 HTTPS 为 `443`，HTTP 为 `80`。
- `ts` 使用秒，不是毫秒。
- 标准化串最后必须保留换行符。
