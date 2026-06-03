# 集群资源与运维 API

本文档汇总集群资源、Helm、YAML、代理、DNS、GPU、终端和诊断类 API。这些接口分布在 `app/application`、`app/k3k`、`app/metrics` 等模块中，通常由面板页面直接调用。Longhorn 接口已拆分到 [longhorn.md](./longhorn.md)。

## 整体使用方式

集群运维接口面向 Kubernetes 原生资源和面板聚合能力。开发时先判断是要直接访问 K8s API，还是调用面板业务封装：原生资源走 `/k8s-proxy/`，Helm、YAML、终端、代理、DNS、GPU、指标等走 `/panel-api/v1/` 下的业务接口。

### 基本流程

1. 使用用户 token 调用接口，确认当前用户可访问的 namespace 和资源范围。
2. 查询原生 K8s 对象时优先走 `/k8s-proxy/api/*` 或 `/k8s-proxy/apis/*`。
3. 安装 Helm、应用 YAML、终端连接、服务代理等面板能力，调用对应 `/panel-api/v1/*` 接口。
4. 涉及长连接或代理接口时，按原始协议处理响应，不额外假设 JSON 包装。
5. 涉及 Longhorn 卷操作时跳转到 [longhorn.md](./longhorn.md)，不要在集群运维文档重复维护。

### 场景选择

| 场景 | 推荐入口 | 说明 |
|------|----------|------|
| 原生 Pod、Deployment、Service 等资源 | `/k8s-proxy/*path` | 透传 Kubernetes API Server |
| Helm Release 管理 | Helm 相关 `/panel-api/v1/*` | 面板封装 Helm list/detail/install/uninstall |
| YAML 应用和回滚 | YAML 相关接口 | 面板处理 manifest 应用、回滚和 Compose 转换 |
| Pod/Node 终端 | TTY、Exec 接口 | 通常涉及 WebSocket 或执行流 |
| Service/Pod 代理 | proxy 接口 | 透传目标服务响应 |
| DNS、GPU、指标 | 对应业务接口 | 返回面板聚合后的业务字段 |

### 使用边界

- 不要把 Kubernetes 原生 API 新增到 `/panel-api/v1/`，应使用 `/k8s-proxy/`。
- 面板业务接口返回的是业务封装字段，不应要求它完整等同 K8s 原生对象。
- 代理和终端类接口要保留原始协议语义，不要强行改成普通 JSON。
- 涉及 namespace 的接口必须明确用户上下文或显式参数，避免跨 namespace 误操作。

## 通用说明

### 鉴权

除明确标注为无鉴权的接口外，本文接口均需要用户 token：

```http
Authorization: Bearer <token>
```

`LOCAL_MOCK=true` 只改变 K8s API 调用方式，不跳过用户认证。测试这些接口时仍需携带有效 token。

### 响应格式

当前后端已将成功响应处理器设置为直接返回业务对象，因此大多数接口不会额外包裹 `code`、`data`、`message`。

示例：

```json
{
  "cpu": {
    "usage": 100,
    "total": 2000
  }
}
```

部分历史接口会直接调用 `http.JSON` 或 `http.String`，响应也以控制器实际写出的 body 为准。失败时通常返回 HTTP `500` 或校验失败响应，错误内容由框架统一处理。

### 参数位置

使用 `form` tag 的接口可从 query、表单或框架支持的绑定来源读取参数；本文按最常见调用方式标注为 `query/form`。使用 `json` tag 的接口必须传 JSON body。代理接口会透传原始方法、Header、Body 和目标响应。

## 能力概览

| 能力 | 说明 |
|------|------|
| 命名空间与 K8s 资源 | 获取命名空间、通过 `/k8s-proxy/` 访问原生 K8s API |
| Helm | Helm Release 列表、详情、安装、卸载、复用 values |
| YAML 和 Compose | 应用 YAML、回滚、Docker Compose 转换 |
| 终端与执行 | Pod TTY、Node TTY、Exec、ExecAll |
| 代理 | Service/Pod/Common proxy、kubeconfig、microapp proxy |
| DNS 和网络诊断 | DNS 解析、DNS zone/record、数据库连接测试、etcd ping |
| Longhorn | 已拆分到 [longhorn.md](./longhorn.md) |
| GPU | GPU 开关、HAMI/GPU Operator、GPU summary、设备和 GPUStack worker |
| 指标 | CPU、内存、磁盘、metrics 安装状态 |

## K8s 原生代理

K8s 原生资源请求应走 `/k8s-proxy/`，不要把原生资源 API 放入面板业务路径。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `ANY` | `/k8s-proxy/*path` | 需要 token | 代理到 Kubernetes API Server |

常见请求：

```http
GET /k8s-proxy/api/v1/namespaces/default/pods
GET /k8s-proxy/apis/apps/v1/namespaces/default/deployments
```

查询参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `local` | query | 否 | string | `true` 或 `1` 时强制使用本地 K8s 客户端 |

响应：透传 Kubernetes API 响应。中间件可能对部分响应做过滤，调用方应按 K8s 原生对象解析。

## 命名空间

### GET `/panel-api/v1/namespaces`

功能：获取当前 token 可访问的命名空间。

请求参数：无。

响应：Kubernetes `corev1.NamespaceList`。

常见响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | array | 命名空间列表 |
| `items[].metadata.name` | string | 命名空间名称 |
| `items[].status.phase` | string | 状态，例如 `Active` |

请求示例：

```http
GET /panel-api/v1/namespaces
Authorization: Bearer <token>
```

## Helm

### 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/helm/releases` | Helm Release 列表 |
| `GET` | `/panel-api/v1/helm/releases/:name` | Helm Release 详情 |
| `POST` | `/panel-api/v1/helm/releases/:name` | 使用仓库安装 Release |
| `DELETE` | `/panel-api/v1/helm/releases/:name` | 卸载 Release |
| `PUT` | `/panel-api/v1/helm/releases/:name/reuse` | 复用 values 更新 Release |
| `GET` | `/panel-api/v1/app-info` | 当前面板 Helm/Deployment 信息 |

### GET `/panel-api/v1/helm/releases`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form | 否 | string | 命名空间；为空时由 Helm 服务使用默认范围 |
| `labelSelector` | query/form | 否 | string | 标签选择器 |

响应：Helm Release 列表。

常见响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | Release 名称 |
| `namespace` | string | 命名空间 |
| `revision` | number | revision |
| `updated` | string | 更新时间 |
| `status` | string | 状态 |
| `chart` | string | Chart |
| `appVersion` | string | 应用版本 |

请求示例：

```http
GET /panel-api/v1/helm/releases?namespace=default&labelSelector=app.kubernetes.io/managed-by=Helm
Authorization: Bearer <token>
```

### GET `/panel-api/v1/helm/releases/:name`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | query/form | 否 | string | 命名空间 |

响应：Helm Release 详情，字段由 Helm SDK 返回结果决定，通常包含 chart、config、manifest、hooks、info 等。

### POST `/panel-api/v1/helm/releases/:name`

功能：使用仓库中的 chart 安装 Helm Release。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | query/form/body | 是 | string | 安装命名空间 |
| `repository` | query/form/body | 否 | string | Chart 仓库地址 |
| `chartName` | query/form/body | 是 | string | Chart 名称 |
| `vals` | query/form/body | 是 | object | Helm values |
| `chartType` | query/form/body | 否 | string | Chart 类型，预留字段 |
| `version` | query/form/body | 否 | string | Chart 版本 |

请求示例：

```http
POST /panel-api/v1/helm/releases/demo
Authorization: Bearer <token>
Content-Type: application/json

{
  "namespace": "default",
  "repository": "https://charts.example.com",
  "chartName": "nginx",
  "version": "1.0.0",
  "vals": {
    "replicaCount": 1
  }
}
```

响应：Helm install 返回结果。

### DELETE `/panel-api/v1/helm/releases/:name`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | query/form | 否 | string | 命名空间 |

响应：Helm uninstall 返回结果。

### PUT `/panel-api/v1/helm/releases/:name/reuse`

功能：复用已有 values 更新 Release。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | query/form/body | 是 | string | 命名空间 |
| `vals` | query/form/body | 是 | object | 要覆盖的 values |

响应：Helm upgrade 返回结果。

### GET `/panel-api/v1/app-info`

鉴权：当前路由未注册 `middleware.Auth`。

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `helmVersion` | string | 配置中的面板 Helm 版本 |
| `helmReleaseName` | string | 面板 Release 名称 |
| `helmNamespace` | string | 面板所在命名空间 |
| `metadataName` | string | 面板 metadata 名称 |
| `deploymentName` | string | 面板 Deployment 名称 |

## YAML、Compose 和回滚

### 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/panel-api/v1/yaml` | 直接提交 Kubernetes YAML |
| `PUT` | `/panel-api/v1/rollback` | 回滚资源 |
| `POST` | `/panel-api/v1/kcompose` | Docker Compose 转换为 Kubernetes YAML |

### POST `/panel-api/v1/yaml`

功能：读取请求原始 body 作为 YAML 并应用到集群。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query | 否 | string | 默认 namespace；为空时使用当前 K8s 客户端 namespace |
| body | raw body | 是 | string/bytes | Kubernetes YAML 内容，可包含多文档 YAML |

请求示例：

```http
POST /panel-api/v1/yaml?namespace=default
Authorization: Bearer <token>
Content-Type: application/yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
data:
  key: value
```

成功响应：成功对象或空成功响应，取决于框架 `JsonSuccessResponse` 输出。

### POST `/panel-api/v1/kcompose`

功能：读取请求原始 body 作为 Docker Compose 内容，转换为 Kubernetes YAML。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| body | raw body | 是 | string/bytes | Docker Compose YAML |

响应：`kompose.ConvertToK8sYaml` 的转换结果，通常为对象/map，value 为生成的 Kubernetes YAML 文本。

请求示例：

```http
POST /panel-api/v1/kcompose
Authorization: Bearer <token>
Content-Type: application/yaml

services:
  web:
    image: nginx:latest
```

### PUT `/panel-api/v1/rollback`

功能：回滚 Kubernetes 资源到指定 revision。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 是 | string | 命名空间 |
| `name` | query/form/body | 是 | string | 资源名称 |
| `kind` | query/form/body | 是 | string | 资源 Kind，例如 `Deployment` |
| `apiVersion` | query/form/body | 是 | string | 资源 apiVersion，例如 `apps/v1` |
| `toRevision` | query/form/body | 是 | int64 | 目标 revision |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `message` | string | 回滚结果信息 |

## 终端与执行

### 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/tty` | 面板本地/K3K server TTY，WebSocket |
| `GET` | `/panel-api/v1/nodetty` | 节点 TTY，WebSocket |
| `GET` | `/panel-api/v1/exec` | Pod 执行命令，支持 WebSocket 或普通 HTTP |
| `POST` | `/panel-api/v1/exec2` | Pod 执行命令，支持复杂 payload |
| `POST` | `/panel-api/v1/exec-all` | 批量执行命令，不支持 WebSocket |
| `GET` | `/panel-api/v1/nodepid` | 获取节点 PID 信息 |
| `POST` | `/panel-api/v1/cp` | Pod 与临时目录之间复制文件 |

### GET/POST `/panel-api/v1/exec`

`GET /exec` 和 `POST /exec2` 使用同一个控制器方法。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 是 | string | Pod 命名空间 |
| `podName` | query/form/body | 是 | string | Pod 名称 |
| `containerName` | query/form/body | 是 | string | 容器名称 |
| `command` | query/form/body | 是 | array<string> | 命令数组，例如 `["/bin/sh","-c","ls -la"]` |
| `tty` | query/form/body | 否 | bool | 是否以 TTY 模式执行 |

普通 HTTP 响应：命令输出文本。

WebSocket 响应：升级为 WebSocket 后进行终端流式交互。超时时间来自 `k8s.exec_timeout_seconds`，小于等于 0 时默认 1800 秒。

请求示例：

```http
POST /panel-api/v1/exec2
Authorization: Bearer <token>
Content-Type: application/json

{
  "namespace": "default",
  "podName": "nginx-xxx",
  "containerName": "nginx",
  "command": ["/bin/sh", "-c", "ls -la /"],
  "tty": false
}
```

### POST `/panel-api/v1/exec-all`

功能：对一个或多个 Pod 批量执行命令。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 是 | string | 命名空间 |
| `podNames` | query/form/body | 否 | array<string> | Pod 名称数组 |
| `podName` | query/form/body | 否 | string | 单个 Pod 名称；当 `podNames` 为空时会转为单元素数组 |
| `containerName` | query/form/body | 是 | string | 容器名称 |
| `command` | query/form/body | 是 | array<string> | 命令数组 |
| `tty` | query/form/body | 否 | bool | 是否 TTY |

约束：不支持 WebSocket；`podNames` 和 `podName` 至少传一个。

响应元素：

| 字段 | 类型 | 说明 |
|------|------|------|
| `podName` | string | Pod 名称 |
| `output` | string | 命令输出 |
| `success` | bool | 是否执行成功 |
| `error` | string | 错误信息；失败时返回 |

响应示例：

```json
[
  {
    "podName": "nginx-xxx",
    "output": "ok\n",
    "success": true
  }
]
```

### GET `/panel-api/v1/tty`

功能：WebSocket 终端。普通集群执行面板本地 shell；K3K token 下进入 K3K server 容器。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `shell` | query/form | 否 | string | `/bin/bash` | 仅允许 `/bin/sh` 或 `/bin/bash`；K3K 模式会强制使用 `/bin/sh` |

响应：WebSocket 终端流。

### GET `/panel-api/v1/nodetty`

功能：WebSocket 节点终端。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `hostIp` | query/form | 是 | string | - | 节点内网 IP |
| `shell` | query/form | 否 | string | `/bin/bash` | 仅允许 `/bin/sh` 或 `/bin/bash` |

普通集群下通过 agent pod 执行 `nsenter` 进入节点 namespace；K3K 集群下查找对应 server pod 执行 shell。

### GET `/panel-api/v1/nodepid`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form | K3K 模式必填 | string | 目标 Pod namespace |
| `podName` | query/form | K3K 模式必填 | string | 目标 Pod 名称 |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `pid` | int/string | 普通集群固定返回 `1`；K3K 模式返回容器在宿主/agent 侧的 PID |

### POST `/panel-api/v1/cp`

功能：Pod 与 `s3.base_dir` 临时目录之间复制文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `from` | query/form/body | 是 | string | 源路径 |
| `to` | query/form/body | 是 | string | 目标路径 |
| `namespace` | query/form/body | 是 | string | Pod namespace |
| `upload` | query/form/body | 是 | string | `1` 表示从临时目录上传到 Pod；其它值表示从 Pod 下载到临时目录 |
| `podName` | query/form/body | 是 | string | Pod 名称 |

成功响应：空成功响应。

## 代理接口

### 接口清单

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| 多方法 | `/panel-api/v1/namespaces/:namespace/services/:name/proxy-root/*path` | 需要 token | Service proxy，路径按根路径转发 |
| 多方法 | `/panel-api/v1/namespaces/:namespace/services/:name/proxy/*path` | 需要 token | Service proxy |
| 多方法 | `/panel-api/v1/namespaces/:namespace/pods/:name/proxy/*path` | 需要 token | Pod proxy |
| 多方法 | `/panel-api/v1/:name/proxy/*path` | 需要 token | Common proxy |
| `ANY` | `/panel-api/v1/proxy-url/` | 当前路由未注册 Auth | 读取远程 URL 内容 |
| `GET` | `/panel-api/v1/kubeconfig` | 需要 token + Proxy middleware | 获取 kubeconfig |
| `ANY` | `/panel-api/v1/namespaces/:namespace/services/:name/proxy-no/*path` | 无 Auth | 无鉴权 Service proxy，历史/TODO 风险接口 |
| 多方法 | `/panel-api/v1/microapp/:name/proxy/*path` | 需要 token | 微应用 proxy |

### Service proxy

路径参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | Service namespace |
| `name` | string | Service 名称和端口，支持 `svc`、`svc:8080`、`https:svc:443` |
| `path` | string | 目标路径 |

转发目标：`<schema>://<service>.<namespace>.svc:<port>/<path>`，默认 `schema=http`、`port=80`。

响应：透传目标服务状态、Header 和 Body；代理会删除目标响应中的 `Access-Control-Allow-Origin`。

### Pod proxy

路径参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | Pod namespace |
| `name` | string | Pod 名称和端口，支持 `pod`、`pod:8080` |
| `path` | string | 目标路径 |

转发目标：Pod IP + 端口。Pod IP 为空时返回错误。

### GET `/panel-api/v1/proxy-url/`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `proxyUrl` | query/form | 是 | string | 要请求的完整 URL |

响应：远程 URL 的响应 body 字符串；请求失败时返回空字符串。

### GET `/panel-api/v1/kubeconfig`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `apiServerUrl` | query | 否 | string | kubeconfig 中使用的 API Server 地址 |

响应：kubeconfig 内容或结构，取决于 `client.ToKubeconfig` 返回值。

## DNS 和诊断

### 查询工具

| 方法 | 路径 | 参数 | 响应 |
|------|------|------|------|
| `POST` | `/panel-api/v1/pinyin` | `words` query/form，必填 | 拼音转换结果数组 |
| `GET` | `/panel-api/v1/dnsip` | `domain` query/form，必填 | IP 地址数组；解析失败返回 `[]` |
| `GET` | `/panel-api/v1/dns-cname` | `domain` query/form，必填 | CNAME 数组；解析失败返回 `[]` |
| `GET` | `/panel-api/v1/myip` | 无 | `{ "ip": "<出口IP>" }`，失败时可能为空 |
| `POST` | `/panel-api/v1/db-conn-test` | `dsn` query/form，必填 | `{ "canConnect": true/false, "msg": "错误信息" }` |
| `POST` | `/panel-api/v1/ping-etcd` | `url` query/form，必填 | `{ "canConnect": true/false, "msg": "错误信息" }` |

`ping-etcd` 会自动补齐尾部 `/`，并请求 `<url>/health`，HTTP 状态码 `200` 视为可连接。

### DNS Zone

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/dns/zones` | DNS zone 列表 |
| `POST` | `/panel-api/v1/dns/zones` | 创建 zone |
| `DELETE` | `/panel-api/v1/dns/zones/:domain` | 删除 zone |

创建 zone 请求体：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `domain` | 是 | string | 域名，会转小写、去掉末尾 `.`，至少两段 label |

Zone 响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `domain` | string | 域名 |
| `recordNum` | int | 记录数量 |
| `updateTime` | string | 更新时间，可能为空 |

请求示例：

```http
POST /panel-api/v1/dns/zones
Authorization: Bearer <token>
Content-Type: application/json

{
  "domain": "example.com"
}
```

### DNS Record

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/dns/zones/:domain/records` | 记录列表 |
| `POST` | `/panel-api/v1/dns/zones/:domain/records` | 创建记录 |
| `PUT` | `/panel-api/v1/dns/zones/:domain/records/:id` | 更新记录 |
| `DELETE` | `/panel-api/v1/dns/zones/:domain/records/:id` | 删除记录 |

Record 请求/响应字段：

| 字段 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `id` | 更新时可选 | string | 自动生成 | 记录 ID，默认由 domain、name、type、value、ttl、mxPriority 计算 |
| `name` | 否 | string | `@` | 主机记录；非 `@` 时校验 label |
| `type` | 是 | string | - | 记录类型，代码会转大写 |
| `value` | 是 | string | - | 记录值 |
| `ttl` | 否 | int | `60` | TTL，小于等于 0 时使用默认值 |
| `mxPriority` | 否 | int | `10` | MX 优先级，小于等于 0 时使用默认值 |

记录值校验：

| 类型 | 校验 |
|------|------|
| `A` | 必须是 IPv4 |
| `AAAA` | 必须是 IPv6 |
| `CNAME` | 必须是合法域名 |
| `TXT` | 必须是单行文本，不能包含换行 |
| `MX` | 使用 `mxPriority` 和域名值 |

请求示例：

```http
POST /panel-api/v1/dns/zones/example.com/records
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "www",
  "type": "A",
  "value": "1.2.3.4",
  "ttl": 60
}
```

### DNS Server

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/dns/server` | DNS 服务配置 |
| `PUT` | `/panel-api/v1/dns/server` | 开启/关闭 DNS 服务 |

更新请求体：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `enabled` | 是 | bool | 是否启用 |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `enabled` | bool | 是否启用 |
| `serviceName` | string | Service 名称 |
| `serviceType` | string | Service 类型，可能为空 |
| `externalIPs` | array<string> | 外部 IP 列表 |

## Longhorn

Longhorn 接口已拆分到独立文档，详细请求参数、响应字段、Longhorn Backend action body 和调用注意见 [longhorn.md](./longhorn.md)。

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/longhorn/need-delete-replica` | 获取需要删除的副本 |
| `GET` | `/panel-api/v1/longhorn/volumes/status` | Longhorn 卷状态 |
| `POST` | `/panel-api/v1/longhorn/install` | 安装 Longhorn |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/attach` | attach 卷 |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/detach` | detach 卷 |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/cancel-expansion` | 取消扩容 |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/trim-filesystem` | trim 文件系统 |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-delete` | 删除快照 |
| `POST` | `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-purge` | 清理快照 |

## GPU

### 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/panel-api/v1/gpu/enabled-gpu` | 开启/关闭 GPU |
| `POST` | `/panel-api/v1/gpu/install-hami` | 安装 HAMI |
| `POST` | `/panel-api/v1/gpu/install-gpu-operator` | 安装 GPU Operator |
| `GET` | `/panel-api/v1/gpu/config` | GPU 配置 |
| `GET` | `/panel-api/v1/gpu/hami/metrics/real` | HAMI 实时指标 |
| `GET` | `/panel-api/v1/gpu/summary` | GPU 汇总 |
| `GET` | `/panel-api/v1/gpu/node/devices` | 节点 GPU 设备 |
| `POST` | `/panel-api/v1/gpu/gpustack/worker` | 创建 GPUStack worker |

这些接口所在 group 注册了 `middleware.Auth`，均需要 token。多数接口可选传 `namespace` query/form，用于初始化 GPU manager。

### POST `/panel-api/v1/gpu/enabled-gpu`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `enabled` | query/form/body | 否 | bool | 是否启用 GPU |
| `namespace` | query/form/body | 否 | string | 命名空间 |

成功响应：空成功响应。

### POST `/panel-api/v1/gpu/install-hami`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `runtimeClassName` | query/form/body | 否 | string | RuntimeClass 名称 |
| `namespace` | query/form/body | 否 | string | 命名空间 |

成功响应：空成功响应。

### POST `/panel-api/v1/gpu/install-gpu-operator`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `driverVersion` | query/form/body | 否 | string | NVIDIA driver 版本 |
| `driverEnabled` | query/form/body | 否 | bool | 是否启用 driver |
| `namespace` | query/form/body | 否 | string | 命名空间 |

成功响应：空成功响应。

### GET `/panel-api/v1/gpu/config`

响应：`GpuManager.ToJsonStruct()` 返回的 GPU 配置对象。

### GET `/panel-api/v1/gpu/hami/metrics/real`

响应元素：

| 字段 | 类型 | 说明 |
|------|------|------|
| `NodeName` | string | 节点名称 |
| `HostCoreUtilization` | string | HAMI host core utilization，字符串化浮点数 |

### GET `/panel-api/v1/gpu/summary`

响应：GPU 集群汇总对象；失败时返回空 summary。

### GET `/panel-api/v1/gpu/node/devices`

响应：节点 GPU 设备列表；失败时返回 `[]`。

### POST `/panel-api/v1/gpu/gpustack/worker`

功能：创建 GPUStack worker StatefulSet。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `image` | query/form/body | 是 | string | - | worker 镜像 |
| `pvcName` | query/form/body | 是 | string | - | 挂载的 PVC 名称 |
| `serverUrl` | query/form/body | 是 | string | - | GPUStack server URL |
| `group` | query/form/body | 是 | string | - | GPUStack backend/group 名称 |
| `workerToken` | query/form/body | 是 | string | - | worker token |
| `namespace` | query/form/body | 是 | string | - | 创建 namespace |
| `gpucores` | query/form/body | 是 | string | - | `nvidia.com/gpucores` 资源值 |
| `gpu` | query/form/body | 是 | string | - | `nvidia.com/gpu` 资源值 |
| `gpumem` | query/form/body | 是 | string | - | `nvidia.com/gpumem` 资源值 |
| `cpu` | query/form/body | 否 | string | - | CPU 资源值；注意当前实现中非空时会被改为 `0` |
| `memory` | query/form/body | 否 | string | - | Memory 资源值；注意当前实现中非空时会被改为 `0` |
| `runtimeClassName` | query/form/body | 否 | string | `nvidia` | RuntimeClass 名称 |
| `password` | query/form/body | 否 | string | - | 预留字段，当前创建容器时未写入环境变量 |

成功响应：当前实现返回空数组 `[]`。

## 指标

指标接口详见 [metrics.md](./metrics.md)。

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/panel-api/v1/metrics/usage/normal` | 当前 token 资源 CPU/内存用量 |
| `GET` | `/panel-api/v1/metrics/usage/disk` | 当前 token 磁盘用量 |
| `GET` | `/panel-api/v1/metrics/usage/cvm/:namespace/name/:name/normal` | 指定 CVM CPU/内存用量 |
| `GET` | `/panel-api/v1/metrics/usage/cvm/:namespace/name/:name/disk` | 指定 CVM 磁盘用量 |
| `GET` | `/panel-api/v1/metrics/installed` | metrics 组件安装状态 |
| `GET` | `/panel-api/v1/metrics/state` | metrics 展示/安装状态 |

### GET `/panel-api/v1/metrics/usage/normal`

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu.usage` | int64 | CPU 已用，milliValue |
| `cpu.total` | int64 | CPU 总量，milliValue |
| `memory.usage` | int64 | 内存已用，bytes |
| `memory.total` | int64 | 内存总量，bytes |

### GET `/panel-api/v1/metrics/usage/disk`

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `disk.usage` | number | 磁盘已用 |
| `disk.total` | number | 磁盘总量 |

### GET `/panel-api/v1/metrics/installed`

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `installed` | bool | metrics Helm Release 是否存在 |
| `baseUrl` | string | metrics 服务代理地址 |
| `namespace` | string | metrics 所在 namespace |

### GET `/panel-api/v1/metrics/state`

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canShowClusterMetrics` | bool | 是否可展示集群指标 |
| `canShowNodeMetrics` | bool | 是否可展示节点指标 |
| `canShowPodMetrics` | bool | 是否可展示 Pod 指标 |
| `needInstallMetricsInDashboard` | bool | Dashboard 是否需要安装 metrics |
| `needInstallMetricsInApp` | bool | 应用内是否需要安装 metrics |

## 开发检查

- 面板业务接口使用 `/panel-api/v1/`，原生 K8s API 使用 `/k8s-proxy/`。
- 代理接口要明确是否需要 `middleware.Proxy`，避免误将业务接口放到 proxy 路径。
- 高风险操作需要前端确认、后端权限和审计日志。
- 不要通过公开接口返回完整 K8s 资源对象。
- 后端新增或修改这些接口时，需要同步前端 API 调用、类型定义和本文档。
