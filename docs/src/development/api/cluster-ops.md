# 集群资源与运维 API

本文档汇总集群资源、Helm、YAML、终端、集群侧代理、DNS、GPU 和诊断类 API。这些接口主要分布在 `app/application` 模块中，通常由面板集群运维页面直接调用。已经拆到独立专题的 Longhorn、metrics、微应用、容器文件、容器镜像等接口不在本文重复维护。

## 整体使用方式

集群运维接口面向 Kubernetes 原生资源和面板聚合能力。开发时先判断是要直接访问 K8s API，还是调用面板业务封装：原生资源走 `/k8s-proxy/`，Helm、YAML、终端、集群侧代理、DNS、GPU 等走 `/panel-api/v1/` 下的业务接口。

### 基本流程

1. 使用用户 token 调用接口，确认当前用户可访问的 namespace 和资源范围。
2. 查询原生 K8s 对象时优先走 `/k8s-proxy/api/*` 或 `/k8s-proxy/apis/*`。
3. 安装 Helm、应用 YAML、终端连接、服务代理等面板能力，调用对应 `/panel-api/v1/*` 接口。
4. 涉及长连接或代理接口时，按原始协议处理响应，不额外假设 JSON 包装。
5. 涉及 Longhorn、metrics、微应用、文件、镜像等专题接口时跳转对应文档，不要在集群运维文档重复维护。

### 场景选择

| 场景 | 推荐入口 | 说明 |
|------|----------|------|
| 原生 Pod、Deployment、Service 等资源 | `/k8s-proxy/*path` | 透传 Kubernetes API Server |
| Helm Release 管理 | Helm 相关 `/panel-api/v1/*` | 面板封装 Helm list/detail/install/uninstall |
| YAML 应用和回滚 | YAML 相关接口 | 面板处理 manifest 应用、回滚和 Compose 转换 |
| Pod/Node 终端 | TTY、Exec 接口 | 通常涉及 WebSocket 或执行流 |
| Service/Pod/Common 代理 | proxy 接口 | 透传目标服务响应 |
| DNS、GPU、诊断 | 对应业务接口 | 返回面板聚合后的业务字段 |

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
| 代理 | K8s API proxy、Service/Pod/Common proxy、proxy-no、proxy-url、kubeconfig |
| DNS 和网络诊断 | DNS 解析、DNS zone/record、数据库连接测试、etcd ping |
| GPU | GPU 开关、HAMI/GPU Operator、GPU summary、设备和 GPUStack worker |

## K8s 原生代理

K8s 原生资源请求应走 `/k8s-proxy/`，不要把原生资源 API 放入面板业务路径。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `ANY` | `/k8s-proxy/*path` | 需要 token | 代理到 Kubernetes API Server |

### 请求路径

`/k8s-proxy/*path` 会去掉 `/k8s-proxy` 前缀后，把剩余路径作为 Kubernetes API Server 路径。例如：

```http
GET /k8s-proxy/api/v1/namespaces/default/pods
GET /k8s-proxy/apis/apps/v1/namespaces/default/deployments
GET /k8s-proxy/api/v1/namespaces/default/services/nginx:80/proxy/health
```

查询参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `local` | query | 否 | string | `true` 或 `1` 时强制使用本地 K8s 客户端 |

### 转发行为

| 项 | 说明 |
|----|------|
| token 来源 | 从用户 token 解析 `k8s_token`，再创建对应 K8s client |
| `Authorization` | 如果请求头带 `Bearer`，后端会替换为当前 K8s client 的 bearer token |
| 方法和 body | 原始 HTTP method、query、body 会继续传给 Kubernetes API Server |
| `local=true` | 强制走本地 K8s client，主要用于开发或特殊资源读取 |
| 响应 | 透传 Kubernetes API 响应；调用方应按 K8s 原生对象、watch 流或 service proxy 响应解析 |

### 适用场景

| 场景 | 示例 |
|------|------|
| 读取原生资源 | `GET /k8s-proxy/api/v1/namespaces/default/pods` |
| 修改原生资源 | `PATCH /k8s-proxy/apis/apps/v1/namespaces/default/deployments/nginx` |
| K8s Service proxy | `GET /k8s-proxy/api/v1/namespaces/default/services/nginx:80/proxy/` |
| CRD 管理 | `GET /k8s-proxy/apis/{group}/{version}/...` |

使用边界：

- 只用于 Kubernetes 原生 API 路径，不要在 `/k8s-proxy/` 下新增面板业务接口。
- Service proxy 返回的是目标服务响应，不一定是 JSON。
- watch、exec、log 等流式接口需要前端按原协议处理。

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

### GET/POST `/panel-api/v1/exec`

`GET /exec` 和 `POST /exec2` 使用同一个控制器方法。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query/form/body | 是 | string | Pod 命名空间 |
| `podName` | query/form/body | 是 | string | Pod 名称 |
| `containerName` | query/form/body | 是 | string | 容器名称 |
| `command` | query/form/body | 是 | array[string] | 命令数组，例如 `["/bin/sh","-c","ls -la"]` |
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
| `podNames` | query/form/body | 否 | array[string] | Pod 名称数组 |
| `podName` | query/form/body | 否 | string | 单个 Pod 名称；当 `podNames` 为空时会转为单元素数组 |
| `containerName` | query/form/body | 是 | string | 容器名称 |
| `command` | query/form/body | 是 | array[string] | 命令数组 |
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

集群侧代理接口用于从面板访问集群内 Service、Pod 或指定主机。除 `proxy-no` 和 `proxy-url` 的特殊说明外，代理响应会透传目标状态码、Header 和 Body；目标不是 JSON 时，前端不能按普通业务 JSON 解析。

### 代理入口对比

| 入口 | 鉴权 | 是否经过 K3K Agent 转发 | 目标 | 典型用途 |
|------|------|--------------------------|------|----------|
| `/k8s-proxy/*path` | 需要用户 token | 否，直接通过 K8s client 代理到 API Server | Kubernetes API Server | 原生 K8s 资源和 K8s service proxy |
| `/panel-api/v1/namespaces/:namespace/services/:name/proxy/*path` | 需要用户 token | K3K token 下会先转发到子集群 Agent | 集群内 Service DNS | 访问当前用户集群内服务 |
| `/panel-api/v1/namespaces/:namespace/services/:name/proxy-root/*path` | 需要用户 token | 不经过 `middleware.Proxy` | root 集群 Service DNS | 强制访问 root 集群服务 |
| `/panel-api/v1/namespaces/:namespace/services/:name/proxy-no/*path` | 可无用户 token | 有 token 且为 K3K token 时会转发到子集群 Agent | 集群内 Service DNS | 微应用、公开回调或服务自身鉴权场景 |
| `/panel-api/v1/namespaces/:namespace/pods/:name/proxy/*path` | 需要用户 token | K3K token 下会先转发到子集群 Agent | Pod IP | 直接访问某个 Pod 暴露端口 |
| `/panel-api/v1/:name/proxy/*path` | 需要用户 token | 不经过 `middleware.Proxy` | `:name` 中指定的 host:port | Agent、文件管理等直接 host 代理 |
| `/panel-api/v1/proxy-url/` | 当前未挂用户鉴权 | 否 | query 中的完整 URL | Helm chart 等远程文本拉取 |
| `/panel-api/v1/kubeconfig` | 需要用户 token | K3K token 下会先转发到子集群 Agent | 当前用户上下文 Kubeconfig | 前端下载 kubeconfig |

### 通用转发规则

| 项 | 说明 |
|----|------|
| 方法 | Service、Pod、Common、proxy-no 路由注册了 WebDAV 方法集合，包含常见 `GET`、`POST`、`PUT`、`DELETE`、`PATCH`、`HEAD`、`OPTIONS` 等 |
| query/body | 原始 query 和 body 透传给目标服务 |
| Header | 原始 Header 透传；`Destination` Header 会针对 WebDAV 路径做改写 |
| Host | 反向代理会把 `Host` 设置为目标地址 host |
| CORS | 代理响应会删除目标响应中的 `Access-Control-Allow-Origin` |
| 错误 | 代理连接失败返回 HTTP `502`，body 形如 `{"code":502,"error":"..."}` |

### Service proxy

#### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy/*path`

功能：代理访问当前用户上下文中的 Kubernetes Service。K3K 集群 token 下，`middleware.Proxy` 会先把请求转发到对应子集群 Agent，再由子集群侧访问 Service。

路径参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | Service namespace |
| `name` | path | 是 | string | Service 名称、协议和端口，支持 `svc`、`svc:8080`、`https:svc:443` |
| `path` | path | 是 | string | 目标服务路径，空路径按 `/` 处理 |

`name` 解析规则：

| 形式 | 解析结果 |
|------|----------|
| `svc` | `http://svc.<namespace>.svc:80` |
| `svc:8080` | `http://svc.<namespace>.svc:8080` |
| `https:svc:443` | `https://svc.<namespace>.svc:443` |

请求示例：

```http
GET /panel-api/v1/namespaces/default/services/nginx:8080/proxy/api/health?debug=1
Authorization: Bearer <token>
```

响应：透传目标服务状态、Header 和 Body。

#### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy-root/*path`

功能：代理访问 root 集群 Service。该入口需要用户 token，但不挂 `middleware.Proxy`，因此不会因 K3K token 自动转到子集群 Agent。

请求参数同 `proxy/*path`。

典型用途：K3K 或子集群用户仍需要访问主集群里的共享服务时使用。调用前应确认当前用户有访问该 root 集群服务的业务权限。

### Pod proxy

#### ANY `/panel-api/v1/namespaces/:namespace/pods/:name/proxy/*path`

功能：通过 Pod IP 访问指定 Pod。K3K 集群 token 下会先转发到子集群 Agent。

路径参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | Pod namespace |
| `name` | path | 是 | string | Pod 名称和端口，支持 `pod`、`pod:8080` |
| `path` | path | 是 | string | 目标路径，空路径按 `/` 处理 |

`name` 解析规则：

| 形式 | 解析结果 |
|------|----------|
| `pod-name` | `http://{podIP}:80` |
| `pod-name:8080` | `http://{podIP}:8080` |

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 使用当前 `k8s_token` 创建 K8s client |
| 2 | 按 namespace/name 读取 Pod |
| 3 | 使用 `pod.status.podIP` 和端口拼接目标地址 |
| 4 | Pod IP 为空时返回服务端错误 |

请求示例：

```http
GET /panel-api/v1/namespaces/default/pods/nginx-xxx:8080/proxy/metrics
Authorization: Bearer <token>
```

### Common proxy

#### ANY `/panel-api/v1/:name/proxy/*path`

功能：按 `:name` 中的 host、协议和端口直接代理，不自动拼接 Kubernetes Service DNS。该入口需要用户 token，但不经过 `middleware.Proxy`，常用于文件管理 Agent 等需要直接访问目标 host 的场景。

路径参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | 目标 host、协议和端口，支持 `host`、`host:8000`、`http:host:8000` |
| `path` | path | 是 | string | 目标路径，空路径按 `/` 处理 |

`name` 解析规则：

| 形式 | 解析结果 |
|------|----------|
| `10.0.0.10` | `http://10.0.0.10:80` |
| `10.0.0.10:8000` | `http://10.0.0.10:8000` |
| `https:example.com:443` | `https://example.com:443` |

请求示例：

```http
GET /panel-api/v1/10.0.0.10:8000/proxy/panel-api/v1/files/webdav-agent/123/agent/etc/hosts
Authorization: Bearer <token>
```

### No-auth Service proxy

#### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy-no/*path`

功能：免面板登录代理访问 Service。该入口不挂 `middleware.Auth`，用于服务自身已经有鉴权、微应用公开入口、回调入口等场景。

路径参数同 Service proxy。

鉴权和转发行为：

| 场景 | 行为 |
|------|------|
| 请求不带 token | 直接访问 root 集群 Service DNS |
| 请求带 K3K token 且当前不是子 Agent | 先转发到子集群 Agent，并附加 `PANEL_ROLE`、`PANEL_USERNAME` Header |
| 请求带非 K3K token | 直接访问 root 集群 Service DNS |

安全边界：

- `proxy-no` 不做面板用户认证，目标服务必须自己校验权限或只暴露安全内容。
- 不应把管理类、写操作类、内部敏感接口直接暴露为 `proxy-no`。
- 如果目标服务依赖面板用户身份，应通过 token 或服务端注入 Header 明确处理，不要假设一定有用户上下文。

### URL 内容拉取

#### GET `/panel-api/v1/proxy-url/`

功能：通过后端发起 GET 请求并返回远程响应 body 字符串。当前路由未挂 `middleware.Auth`，主要用于读取远程 chart、配置或文本资源。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `proxyUrl` | query/form | 是 | string | 要请求的完整 URL |

响应：HTTP `200` 文本。远程请求失败时返回空字符串。

请求示例：

```http
GET /panel-api/v1/proxy-url/?proxyUrl=https%3A%2F%2Fcharts.example.com%2Findex.yaml
```

使用边界：

- 该接口只返回远程 body 字符串，不透传远程状态码和 Header。
- 新增调用时要避免把内网敏感地址、带 token 的 URL 或用户可控任意地址直接传入。

### Kubeconfig

#### GET `/panel-api/v1/kubeconfig`

功能：根据当前用户 token 生成 kubeconfig。K3K token 下会经过 `middleware.Proxy` 转到子集群 Agent，由子集群上下文生成 kubeconfig。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `apiServerUrl` | query | 否 | string | kubeconfig 中使用的 API Server 地址 |

响应：`client.ToKubeconfig(apiServerUrl)` 返回的 kubeconfig 结构。

请求示例：

```http
GET /panel-api/v1/kubeconfig?apiServerUrl=https://api.example.com
Authorization: Bearer <token>
```

使用边界：

- kubeconfig 可用于访问集群，前端下载和日志记录时不能泄露其中 token。
- `apiServerUrl` 只影响生成配置中的 server 地址，不改变当前用户实际权限。

## DNS 和诊断

### 查询工具

| 方法 | 路径 | 参数 | 响应 |
|------|------|------|------|
| `POST` | `/panel-api/v1/pinyin` | `words` query/form，必填 | 拼音转换结果数组 |
| `GET` | `/panel-api/v1/dnsip` | `domain` query/form，必填 | IP 地址数组；解析失败返回 `[]` |
| `GET` | `/panel-api/v1/dns-cname` | `domain` query/form，必填 | CNAME 数组；解析失败返回 `[]` |
| `GET` | `/panel-api/v1/myip` | 无 | `{ "ip": "{出口IP}" }`，失败时可能为空 |
| `POST` | `/panel-api/v1/db-conn-test` | `dsn` query/form，必填 | `{ "canConnect": true/false, "msg": "错误信息" }` |
| `POST` | `/panel-api/v1/ping-etcd` | `url` query/form，必填 | `{ "canConnect": true/false, "msg": "错误信息" }` |

`ping-etcd` 会自动补齐尾部 `/`，并请求 `{url}/health`，HTTP 状态码 `200` 视为可连接。

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
| `externalIPs` | array[string] | 外部 IP 列表 |

## GPU

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

## 开发检查

- 面板业务接口使用 `/panel-api/v1/`，原生 K8s API 使用 `/k8s-proxy/`。
- 代理接口要明确是否需要 `middleware.Proxy`，避免误将业务接口放到 proxy 路径。
- 高风险操作需要前端确认、后端权限和审计日志。
- 不要通过公开接口返回完整 K8s 资源对象。
- 后端新增或修改这些接口时，需要同步前端 API 调用、类型定义和本文档。
