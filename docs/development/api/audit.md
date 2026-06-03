# w7panel-server/app/audit API 文档

本文档记录审计日志相关 API。审计接口用于查看登录日志、操作日志和日志采集状态，当前路由统一挂载在 `/panel-api/v1/audit` 下，并需要用户 token。

## 整体使用方式

审计接口用于排查登录行为、写操作记录和日志系统状态。开发时先确认日志能力是否启用以及 VictoriaLogs 是否可访问，再按登录日志或操作日志场景查询。

### 基本流程

1. 使用用户 token 调用 `/panel-api/v1/audit/logs/status`，确认审计日志能力和 VictoriaLogs 状态。
2. 查询登录记录时调用 `/panel-api/v1/audit/login-logs`。
3. 查询写操作记录时调用 `/panel-api/v1/audit/operation-logs`。
4. 管理员可按用户、租户、时间、方法、路径等条件过滤；普通用户会被限制到自身上下文。
5. 查询失败时优先检查 `logs.base_url`、VictoriaLogs 健康状态和 LogSQL 参数。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 日志能力状态 | `/panel-api/v1/audit/logs/status` | 查看启用状态和 VictoriaLogs 可达性 |
| 登录记录 | `/panel-api/v1/audit/login-logs` | 查询登录成功/失败 |
| 操作记录 | `/panel-api/v1/audit/operation-logs` | 查询写操作记录 |

### 使用边界

- 审计查询依赖 VictoriaLogs，服务不可达时接口会返回错误。
- 非管理员查询会强制使用当前用户和租户上下文。
- 时间参数当前不做格式转换，建议使用 RFC3339。
- 审计日志可能包含路径、方法、错误信息等排障字段，不应记录完整 token、密码或密钥。

## 概览

| 能力 | 接口 | 说明 |
|------|------|------|
| 日志状态 | `GET /panel-api/v1/audit/logs/status` | 查看审计日志能力是否启用，以及 VictoriaLogs 是否可访问 |
| 登录日志 | `GET /panel-api/v1/audit/login-logs` | 查询用户登录成功/失败记录 |
| 操作日志 | `GET /panel-api/v1/audit/operation-logs` | 查询用户写操作记录 |

## 鉴权

所有接口都注册了 `middleware.Auth`，请求必须携带有效用户 token。

```http
Authorization: Bearer <token>
```

`LOCAL_MOCK=true` 只改变 K8s API 调用方式，不跳过用户认证。审计查询接口在测试模式下同样需要 token。

## 响应格式

审计接口当前直接返回业务对象，不额外包裹 `code`、`data`、`message`。例如状态接口直接返回：

```json
{
  "enabled": true,
  "installed": true,
  "baseUrl": "http://vlsingle-w7panel-logs.default.svc:9428"
}
```

查询 VictoriaLogs 失败时接口返回 HTTP `500`，错误内容由框架统一处理；常见原因包括 `logs.base_url` 为空、VictoriaLogs 不可达、VictoriaLogs 返回非 2xx 状态。

## 配置项

审计日志依赖 `w7panel-server/config.yaml` 中的 `logs` 配置：

| 配置 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| `logs.enabled` | `LOGS_ENABLED` | `true` | 是否启用日志写入和状态能力 |
| `logs.base_url` | `LOGS_BASE_URL` | `http://vlsingle-w7panel-logs.default.svc:9428` | VictoriaLogs 基础地址，代码会自动去掉末尾 `/` |
| `logs.timeout` | `LOGS_TIMEOUT` | `5s` | 查询和健康检查超时时间；代码默认兜底为 5 秒 |
| `logs.batch_size` | `LOGS_BATCH_SIZE` | `1` | 预留批量配置，当前审计写入按单条 JSON Line 写入 |

状态接口会访问 `{logs.base_url}/health`。登录和操作日志写入使用 `{logs.base_url}/insert/jsonline`，查询使用 `{logs.base_url}/select/logsql/query`。

## 通用查询参数

`login-logs` 和 `operation-logs` 共用以下 query 参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 约束/处理 | 说明 |
|------|------|------|------|--------|-----------|------|
| `page` | query | 否 | int | `1` | 小于等于 0 时按 `1` 处理；非数字会被解析为 0 后按 `1` 处理 | 页码 |
| `pageSize` | query | 否 | int | `20` | 小于等于 0 时按 `20` 处理；最大 `200`；非数字会被解析为 0 后按 `20` 处理 | 每页条数 |
| `username` | query | 否 | string | - | 仅管理员生效；非管理员会强制使用当前 token 中的用户名 | 用户名精确过滤 |
| `tenant` | query | 否 | string | - | 仅管理员生效；非管理员会强制使用当前 token 中的租户 | 租户精确过滤 |
| `success` | query | 否 | string | - | `true`/`false` 按布尔条件查询，其它值按字符串精确匹配 | 成功状态 |
| `method` | query | 否 | string | - | 按字符串精确匹配，建议传大写 HTTP 方法 | HTTP 方法过滤，主要用于操作日志 |
| `path` | query | 否 | string | - | 按字符串精确匹配 | 请求路径过滤，主要用于操作日志 |
| `startTime` | query | 否 | string | - | 拼接为 `time:>="<startTime>"`，不做格式转换 | 开始时间 |
| `endTime` | query | 否 | string | - | 拼接为 `time:<="<endTime>"`，不做格式转换 | 结束时间 |

时间建议使用 RFC3339 格式，例如 `2026-06-03T00:00:00Z`。接口当前不会校验时间格式，格式是否可用由 VictoriaLogs LogSQL 决定。

分页查询固定按 `_time desc` 排序，内部查询形式为：

```text
<filters> | sort by (_time desc) | offset <(page - 1) * pageSize> | limit <pageSize>
```

## 查询权限范围

查询接口会根据当前 token 解析用户上下文：

| 用户类型 | 判断方式 | 查询范围 |
|----------|----------|----------|
| 管理员 | `user_mode` 为 `founder` 或 `cluster` | 可通过 `tenant`、`username` 查询指定范围；未传时可看全部日志 |
| 非管理员 | 其它 `user_mode` | 忽略请求中的 `tenant`、`username`，强制限定为当前 token 对应租户和用户名 |

租户默认来自 `k8s.default_namespace`，如果 token 包含 K3K 配置，则使用 K3K namespace。

## 分页响应字段

`login-logs` 和 `operation-logs` 均返回以下结构：

| 字段 | 类型 | 说明 |
|------|------|------|
| `list` | array<object> | 日志列表。每个元素为 VictoriaLogs 查询结果解码后的字段 map |
| `total` | int | 满足过滤条件的总数，由 `stats count() as total` 查询得到 |
| `page` | int | 归一化后的当前页码 |
| `pageSize` | int | 归一化后的每页条数 |

VictoriaLogs 查询结果可能包含 `_time` 或写入字段 `time`。前端展示应优先兼容 `_time`，同时兼容 `time`。

## 接口清单

### GET `/panel-api/v1/audit/logs/status`

功能：获取审计日志状态。

鉴权：需要用户 token。

请求参数：无。

处理逻辑：

| 场景 | 返回说明 |
|------|----------|
| `logs.enabled=false` | `enabled=false`，`installed=false`，`message="logs disabled"` |
| `logs.base_url` 为空 | `enabled=true`，`installed=false`，`message="logs base_url empty"` |
| `{baseUrl}/health` 返回 2xx | `installed=true` |
| `{baseUrl}/health` 返回非 2xx | `installed=false`，`message="status <HTTP状态码>"` |
| 健康检查请求异常 | `installed=false`，`message` 为错误信息 |

响应参数：

| 字段 | 类型 | 必返 | 说明 |
|------|------|------|------|
| `enabled` | bool | 是 | 是否启用日志能力，对应 `logs.enabled` |
| `installed` | bool | 是 | VictoriaLogs 健康检查是否通过 |
| `baseUrl` | string | 是 | VictoriaLogs 基础地址，对应 `logs.base_url` |
| `message` | string | 否 | 状态说明或错误信息；为空时因 `omitempty` 不返回 |

请求示例：

```http
GET /panel-api/v1/audit/logs/status
Authorization: Bearer <token>
```

响应示例：

```json
{
  "enabled": true,
  "installed": true,
  "baseUrl": "http://vlsingle-w7panel-logs.default.svc:9428"
}
```

未启用示例：

```json
{
  "enabled": false,
  "installed": false,
  "baseUrl": "http://vlsingle-w7panel-logs.default.svc:9428",
  "message": "logs disabled"
}
```

### GET `/panel-api/v1/audit/login-logs`

功能：获取登录日志列表。

鉴权：需要用户 token。

请求参数：见“通用查询参数”。登录日志常用参数为 `page`、`pageSize`、`username`、`tenant`、`success`、`startTime`、`endTime`。

查询固定追加 `audit_type:"login"`。

登录日志字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `_time` | string | VictoriaLogs 查询时间字段 |
| `time` | string | 写入日志时的时间字段 |
| `audit_type` | string | 固定为 `login` |
| `tenant` | string | 租户/namespace |
| `username` | string | 用户名；失败日志可能是请求中提交的用户名，OAuth 失败场景可能为空或用户 ID |
| `user_mode` | string | 用户模式，例如 `normal`、`founder`、`cluster` |
| `login_method` | string | 登录方式，例如 `k8s`、`oauth`，由认证流程传入 |
| `success` | bool | 是否登录成功 |
| `reason` | string | 登录失败原因；最长保留 512 字符；成功日志不返回 |
| `ip` | string | 客户端 IP |
| `user_agent` | string | 请求 User-Agent |
| `message` | string | `login success` 或 `login failed` |

请求示例：

```http
GET /panel-api/v1/audit/login-logs?page=1&pageSize=20&success=true&username=admin
Authorization: Bearer <token>
```

响应示例：

```json
{
  "list": [
    {
      "_time": "2026-06-03T10:00:00Z",
      "time": "2026-06-03T10:00:00Z",
      "audit_type": "login",
      "tenant": "default",
      "username": "admin",
      "user_mode": "founder",
      "login_method": "k8s",
      "success": true,
      "ip": "10.0.0.10",
      "user_agent": "Mozilla/5.0",
      "message": "login success"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

失败日志示例：

```json
{
  "list": [
    {
      "_time": "2026-06-03T10:05:00Z",
      "audit_type": "login",
      "tenant": "default",
      "username": "admin",
      "login_method": "oauth",
      "success": false,
      "reason": "invalid credentials",
      "ip": "10.0.0.10",
      "user_agent": "Mozilla/5.0",
      "message": "login failed"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

### GET `/panel-api/v1/audit/operation-logs`

功能：获取操作日志列表。

鉴权：需要用户 token。

请求参数：见“通用查询参数”。操作日志常用参数为 `page`、`pageSize`、`username`、`tenant`、`success`、`method`、`path`、`startTime`、`endTime`。

查询固定追加 `audit_type:"operation"`。

操作日志记录范围：

| 条件 | 说明 |
|------|------|
| 请求方法 | 仅记录 `POST`、`PUT`、`PATCH`、`DELETE`、`PROPPATCH`、`MKCOL`、`COPY`、`MOVE`、`LOCK`、`UNLOCK`、`LINK`、`UNLINK` |
| 路径排除 | `/panel-api/v1/audit/` 开头的审计查询接口不会再次写入操作日志 |
| 用户要求 | 请求上下文必须同时存在 `k8s_token` 和 `username` |
| 成功判断 | HTTP 状态码 `< 400` 记为 `success=true` |

操作日志字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `_time` | string | VictoriaLogs 查询时间字段 |
| `time` | string | 写入日志时的时间字段 |
| `audit_type` | string | 固定为 `operation` |
| `tenant` | string | 租户/namespace |
| `username` | string | 当前用户，对应 token 中的 ServiceAccount 名称优先 |
| `user_mode` | string | 用户模式，例如 `normal`、`founder`、`cluster` |
| `method` | string | HTTP 方法 |
| `path` | string | 实际请求路径，不含 query |
| `route` | string | Gin 匹配到的路由模板，例如 `/panel-api/v1/helm/releases/:name` |
| `route_description` | string | 路由中文描述；未配置时使用方法语义和路由拼接 |
| `params` | string | 路由参数 JSON 字符串；参数名包含 `token`、`password`、`secret` 时值会被替换为 `***` |
| `status_code` | int | HTTP 响应状态码 |
| `success` | bool | 是否成功，`status_code < 400` 为 true |
| `duration_ms` | int | 请求耗时，毫秒 |
| `ip` | string | 客户端 IP |
| `user_agent` | string | 请求 User-Agent |
| `message` | string | 操作说明，当前与 `route_description` 一致 |

请求示例：

```http
GET /panel-api/v1/audit/operation-logs?page=1&pageSize=20&method=POST&path=/panel-api/v1/zpk/install
Authorization: Bearer <token>
```

响应示例：

```json
{
  "list": [
    {
      "_time": "2026-06-03T10:10:00Z",
      "time": "2026-06-03T10:10:00Z",
      "audit_type": "operation",
      "tenant": "default",
      "username": "admin",
      "user_mode": "founder",
      "method": "POST",
      "path": "/panel-api/v1/zpk/install",
      "route": "/panel-api/v1/zpk/install",
      "route_description": "安装或更新 ZPK 应用",
      "params": "",
      "status_code": 200,
      "success": true,
      "duration_ms": 135,
      "ip": "10.0.0.10",
      "user_agent": "Mozilla/5.0",
      "message": "安装或更新 ZPK 应用"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

## 调试示例

```bash
TOKEN="<token>"

curl -s "http://localhost:8080/panel-api/v1/audit/logs/status" \
  -H "Authorization: Bearer ${TOKEN}"

curl -s "http://localhost:8080/panel-api/v1/audit/login-logs?page=1&pageSize=20&success=false" \
  -H "Authorization: Bearer ${TOKEN}"

curl -s "http://localhost:8080/panel-api/v1/audit/operation-logs?page=1&pageSize=20&method=POST" \
  -H "Authorization: Bearer ${TOKEN}"
```

## 开发检查

- 日志接口不得返回 token、密码、密钥、OIDC code 等敏感信息。
- 新增可写接口后，如需更友好的操作说明，应同步补充 `common/service/audit/route_descriptions.go`。
- 新增日志字段后，需要同步更新本文档、前端展示字段和相关测试。
- 查询接口应控制分页和时间范围，避免一次返回过大日志。
- 如果新增日志类型，需要同步更新本文档和 [README.md](./README.md) 的目录。
