# 容器镜像管理 API

本文档说明镜像仓库、containerd 镜像补丁操作、容器提交镜像、面板导出推送镜像和 ZPK 构建镜像相关 API。该能力主要服务于节点镜像管理、运行中容器生成镜像、ZPK 构建任务和本地 Registry。

## 整体使用方式

容器镜像接口覆盖 Registry v2、Agent 上的 containerd 镜像操作、运行中容器 commit、工作负载镜像导出推送和 ZPK 构建镜像。开发时先区分是标准 Registry 协议、面板业务接口，还是 Agent 代理接口。

### 基本流程

1. 查询 Registry 服务信息，确认 Agent Registry 地址和 `registry.enabled` 配置。
2. 标准 Docker Registry v2 调用走 `/v2/*path`，按 OCI/Docker Distribution 协议处理。
3. containerd 镜像列表、tag、delete、label、import 等操作走 `/panel-api/v1/registry/patch/*`。
4. 运行中容器生成镜像时，调用 `/panel-api/v1/registry/containers/:id/commit`。
5. 应用构建镜像时，优先查看 ZPK 构建接口和 [zpk.md](./zpk.md) 中的构建流程。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| Registry 标准拉推 | `/v2/*path` | 保持 Docker Registry v2 协议 |
| 获取 Registry 地址 | `/panel-api/v1/registry/server-info` | 前端获取 Agent Registry 信息 |
| containerd 镜像管理 | `/panel-api/v1/registry/patch/*` | 面板代理到 Agent 执行 |
| 容器 commit 镜像 | `/panel-api/v1/registry/containers/:id/commit` | 从运行中容器生成镜像 |
| 工作负载导出推送 | `/panel-api/v1/containers/image/export-push` | 面板 application 模块能力 |
| ZPK 构建镜像 | `/panel-api/v1/zpk/buildimage/*` | 创建 Job/CronJob 构建 |

### 使用边界

- `/v2/*path` 是标准 Registry 协议入口，不应改成面板 JSON 响应。
- Registry 相关部分接口仅在 `registry.enabled=true` 时注册。
- Agent 代理接口需要确认目标节点、containerd 和 Registry 服务可用。
- 镜像导入、删除、commit、push 都可能影响节点镜像状态，调用前确认目标镜像名和 tag。

## 通用说明

### 鉴权和注册条件

| 接口类型 | 鉴权/注册 |
|----------|-----------|
| `/v2/*path` | 仅 `registry.enabled=true` 时注册；当前路由未挂载用户鉴权中间件，访问控制由 Registry handler 或部署层处理 |
| `/panel-api/v1/registry/containers/:id/commit` | 仅 `registry.enabled=true` 时注册；需要用户 token |
| `/panel-api/v1/registry/server-info` | 始终注册；需要用户 token |
| `/panel-api/v1/registry/patch/*` | 始终注册；需要用户 token，并经过 `middleware.Proxy` 代理到 Agent |
| `/panel-api/v1/containers/image/export-push` | 需要用户 token，并经过 `middleware.Proxy` |
| `/panel-api/v1/zpk/buildimage/*` | 需要用户 token |

用户 token 请求头：

```http
Authorization: Bearer <token>
```

### 响应格式

当前后端成功响应处理器直接返回业务对象，不额外包裹 `code`、`data`、`message`。调用 `JsonSuccessResponse` 的接口返回 JSON 字符串 `"success"`。

Docker Registry v2 接口必须保持标准协议响应，不要改成面板 JSON 格式。

### 参数位置

Registry v2 兼容入口按 Docker Distribution 协议使用 path、query、Header 和 body。面板 Registry 管理接口按各自参数表使用 path、query/form 或 JSON body；构建镜像、导入、导出推送等操作类接口以接口明细中的“位置”列为准。

## 能力概览

| 能力 | 说明 |
|------|------|
| Registry v2 兼容入口 | `/v2/*path`，仅 `registry.enabled=true` 时注册 |
| Registry 服务信息 | 获取 Agent Registry 访问地址和 Host |
| containerd 镜像操作 | 列表、打标签、删除、打 label、导入 |
| 容器提交镜像 | 将运行中容器 commit 为镜像 |
| 工作负载镜像导出推送 | 面板 application 模块提供 `/containers/image/export-push` |
| ZPK 构建镜像任务 | 创建 Job 或 CronJob 构建应用镜像 |

## Registry v2 兼容入口

### ANY `/v2/*path`

功能：Docker Registry v2 兼容入口，底层由 containerd registry handler 处理。

启用条件：`registry.enabled=true`。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `path` | path | 是 | string | Docker Registry v2 路径 |
| body | body | 否 | bytes/json | Docker Registry 标准请求体 |
| headers | header | 否 | string | Docker Registry 标准请求头 |

常见标准路径：

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/v2/` | Registry API 版本检查 |
| `GET` | `/v2/_catalog` | 镜像仓库列表 |
| `GET` | `/v2/{name}/tags/list` | tag 列表 |
| `GET` / `HEAD` | `/v2/{name}/manifests/{reference}` | 获取/检查 manifest |
| `PUT` | `/v2/{name}/manifests/{reference}` | 推送 manifest |
| `GET` / `HEAD` | `/v2/{name}/blobs/{digest}` | 获取/检查 blob |
| `POST` | `/v2/{name}/blobs/uploads/` | 初始化 blob 上传 |
| `PATCH` | `/v2/{name}/blobs/uploads/{uuid}` | 上传 blob 分片 |
| `PUT` | `/v2/{name}/blobs/uploads/{uuid}?digest=...` | 完成 blob 上传 |

响应：Docker Registry v2 标准响应。示例：`GET /v2/` 通常返回 200，并带 Registry 版本相关 Header；manifest/blob 接口按 OCI/Docker Distribution 协议返回。

## Registry 管理接口

### GET `/panel-api/v1/registry/server-info`

功能：获取前端访问 Agent Registry 的地址。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `hostIp` | query | 否 | string | 目标节点 IP；传入时查找该节点上的 Agent Registry Pod |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `requestUrl` | string | 前端请求 Registry/Agent 的基础 URL；非 K3K token 下为 `/panel-api/v1/{podIp}:8000/proxy`，K3K token 下可能为 `/` |
| `requestHost` | string | Registry 服务 Host，格式 `{podIp}:8000` |

响应示例：

```json
{
  "requestUrl": "/panel-api/v1/10.0.0.10:8000/proxy",
  "requestHost": "10.0.0.10:8000"
}
```

错误：找不到可用 Agent Registry Pod 时返回服务端错误，错误信息类似 `not found agent registry ip`。

### POST `/panel-api/v1/registry/containers/:id/commit`

功能：将运行中的 containerd 容器提交为镜像。

启用条件：`registry.enabled=true`。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `id` | path | 是 | string | containerd 容器 ID |
| `ref` | context | 是 | string | 目标镜像引用，例如 `registry.example.com/ns/app:v1`。Controller 从 `ctx.GetString("ref")` 读取，当前 HTTP 参数不会直接绑定该字段 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `digest` | string | 新镜像 digest |

响应示例：

```json
{
  "digest": "sha256:..."
}
```

注意：如果调用链没有把 `ref` 注入到 Gin context，提交会因镜像引用为空或非法而失败。

## containerd 镜像补丁接口

这些接口路由前缀为 `/panel-api/v1/registry/patch`，均经过 `Auth` 和 `Proxy` 中间件，实际在目标 Agent 环境中操作 containerd。底层 containerd namespace 为服务配置中的 registry containerd namespace。

### GET `/panel-api/v1/registry/patch/images/list`

功能：获取 containerd 镜像列表。

请求参数：无。当前过滤条件固定为 `dangling=false`。

响应：`[]images.Image`，来自 containerd `github.com/containerd/containerd/v2/core/images.Image`。

常见响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 镜像引用 |
| `target` | object | OCI descriptor |
| `target.mediaType` | string | manifest/config media type |
| `target.digest` | string | digest |
| `target.size` | int64 | descriptor size |
| `labels` | object&lt;string,string&gt; | 镜像 labels |
| `createdAt` | string | 创建时间 |
| `updatedAt` | string | 更新时间 |

### PUT `/panel-api/v1/registry/patch/images/tag`

功能：给镜像添加 tag。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `source` | query/form/body | 是 | string | 源镜像引用 |
| `target` | query/form/body | 是 | string | 目标镜像引用 |

请求示例：

```http
PUT /panel-api/v1/registry/patch/images/tag
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

source=registry.local.w7.cc/default/app:v1&target=registry.local.w7.cc/default/app:v2
```

成功响应：

```json
"success"
```

### POST `/panel-api/v1/registry/patch/images/delete`

功能：删除镜像。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `target` | query/form/body | 是 | string | 要删除的镜像引用 |
| `force` | query/form/body | 否 | bool | 是否强制删除 |
| `async` | query/form/body | 否 | bool | 是否异步删除 |

成功响应：

```json
"success"
```

### POST `/panel-api/v1/registry/patch/images/label`

功能：修改镜像 labels。

请求类型：`application/json`

请求参数：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `name` | 是 | string | 镜像引用 |
| `labels` | 是 | object&lt;string,string&gt; | label 键值 |
| `replace` | 是 | bool | 是否替换全部 labels；`true` 时 field path 为 `labels`，否则只更新传入 key |

请求示例：

```json
{
  "name": "registry.local.w7.cc/default/app:v1",
  "labels": {
    "w7.cc/source": "panel"
  },
  "replace": false
}
```

成功响应：

```json
"success"
```

### POST `/panel-api/v1/registry/patch/images/import`

功能：从 tar 文件导入镜像。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | query/form/body | 是 | string | 导入后的镜像引用 |
| `path` | query/form/body | 是 | string | 镜像 tar 文件路径 |
| `pinned` | query/form/body | 否 | bool | 当前控制器绑定但未传给底层导入逻辑 |

路径处理：

| 场景 | `path` 解释 |
|------|-------------|
| 普通面板环境 | 直接使用传入路径 |
| Agent 或 K3K virtual 环境 | 拼接为 `{s3.base_dir}/{path}` |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 实际导入后的镜像名 |

响应示例：

```json
{
  "name": "registry.local.w7.cc/default/app:v1"
}
```

## 面板容器镜像导出推送

### POST `/panel-api/v1/containers/image/export-push`

功能：把运行中容器 rootfs 导出为单层 OCI 镜像，并推送到指定 Registry。

鉴权：需要用户 token，并经过 `middleware.Proxy`。实际在目标 Agent 节点执行。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `containerId` | query/form/body | 是 | string | containerd 容器 ID，支持不带 `containerd://` 前缀的 ID |
| `registryDomain` | query/form/body | 是 | string | 目标 Registry 域名，例如 `registry.local.w7.cc` |
| `imageName` | query/form/body | 是 | string | 镜像名和 tag，例如 `default/app:v1` |

响应类型：`text/event-stream`。

响应内容：逐行输出构建/推送进度文本，每行末尾 `\n`。当前不是标准 SSE `data:` 格式。

进度示例：

```text
exporting container image, containerId: 123456
push image to container registry, containerId: 123456, image: default/app:v1, registry: registry.local.w7.cc
push image status: complete: 1024, total: 2048, err: <nil>
```

导出规则：

| 项 | 说明 |
|----|------|
| containerd socket | `/var/run/k3s/containerd/containerd.sock` |
| containerd namespace | `k8s.io` |
| rootfs 来源 | container task spec 的 rootfs path |
| 默认排除 | `.dockerenv`、`dev`、`proc`、`sys`、`run/secrets`、`var/run/secrets` |
| 挂载点排除 | 会排除容器 spec 中的 mount destination |
| 镜像格式 | 单层 OCI image，gzip 压缩 layer |
| 运行配置 | 尝试继承原镜像的 Cmd、Entrypoint、Env、Labels、User、Volumes、WorkingDir、ExposedPorts 等 |

注意：控制器的 `PushRequest` 支持 Registry 账号和 plain HTTP，但当前 HTTP 参数只绑定 `containerId`、`registryDomain`、`imageName`，没有绑定用户名、密码或 plain HTTP。

## ZPK 构建镜像

### POST `/panel-api/v1/zpk/buildimage/job`

功能：创建 Kubernetes `batch/v1 Job`，使用 Kaniko 构建并推送镜像。

请求类型：JSON。

请求字段：

| 字段 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `dockerRegistry` | 否 | object | Registry 账号信息 |
| `dockerRegistry.host` | 否 | string | Registry 域名；为 `registry.local.w7.cc` 时自动设置 `hostNetwork=true`，并使用内置 `admin/w7-secret` |
| `dockerRegistry.username` | 否 | string | Registry 用户名 |
| `dockerRegistry.password` | 否 | string | Registry 密码 |
| `dockerRegistry.namespace` | 否 | string | Registry namespace |
| `dockerfilePath` | 否 | string | Dockerfile 路径 |
| `buildContext` | 否 | string | 构建上下文 |
| `pushImage` | 是 | string | 目标镜像引用 |
| `zipUrl` | 是 | string | 构建上下文 zip 下载 URL |
| `identifie` | 否 | string | 应用标识，用于环境变量 `MODULE_NAME` |
| `notifyCompletionUrl` | 否 | string | 构建成功通知 URL |
| `notifyFailedUrl` | 否 | string | 构建失败通知 URL |
| `hostNetwork` | 否 | bool | 是否使用 hostNetwork；也影响 `--insecure --insecure-pull` |
| `hostAliases` | 否 | array | Kubernetes `corev1.HostAlias` 列表 |
| `title` | 否 | string | Job annotation 标题 |
| `labels` | 否 | object&lt;string,string&gt; | 写入 Job labels |
| `buildJobName` | 否 | string | Job 名称；为空时由底层生成 |
| `schedule` | 否 | string | 仅 CronJob 使用 |
| `dockerRegistrySecretName` | 否 | string | Docker config Secret 名称；会从 `default` namespace 读取 `.dockerconfigjson` 填充账号 |
| `panelRegistryHost` | 否 | string | 面板 Registry host；K3K token 下后端可能自动填充 |

请求示例：

```json
{
  "dockerRegistry": {
    "host": "registry.local.w7.cc",
    "username": "admin",
    "password": "w7-secret",
    "namespace": "default"
  },
  "dockerfilePath": "./Dockerfile",
  "buildContext": ".",
  "pushImage": "registry.local.w7.cc/default/demo:v1",
  "zipUrl": "https://example.com/source.zip",
  "identifie": "demo",
  "notifyCompletionUrl": "https://example.com/callback",
  "title": "demo build",
  "labels": {
    "app": "demo"
  },
  "buildJobName": "demo-build"
}
```

响应：创建后的 Kubernetes Job 对象。

常见响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `apiVersion` | string | `batch/v1` |
| `kind` | string | `Job` |
| `metadata.name` | string | Job 名称 |
| `metadata.namespace` | string | 当前实现创建在 `default` namespace |
| `metadata.labels` | object | 请求中的 labels |
| `metadata.annotations.title` | string | 请求中的 title |
| `spec.template.spec.containers[0].name` | string | `build-image` |
| `spec.template.spec.containers[0].image` | string | Kaniko 构建镜像 |

### POST `/panel-api/v1/zpk/buildimage/cronjob`

功能：创建 Kubernetes `batch/v1 CronJob`，定时执行镜像构建。

请求字段：同 `/buildimage/job`，其中 `schedule` 用于 `spec.schedule`。

响应：创建后的 Kubernetes CronJob 对象。

### ANY `/panel-api/v1/zpk/build-image-success`

功能：构建成功回调/兼容入口。

请求/响应：由 ZPK 构建流程调用，详细字段见 [zpk.md](./zpk.md)。该接口需要用户 token。

### GET `/panel-api/v1/zpk/local-url`

功能：获取本地 ZPK 服务 Ingress 地址和 OAuth token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `instance` | query | 否 | string | `w7-zpkv2` | Helm instance label |

响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `host` | string | Ingress host；未找到时为空 |
| `hasIngress` | bool/string | 是否存在 Ingress；未找到时当前实现返回字符串 `"false"` |
| `isHttps` | bool | 是否配置 TLS；找到 Ingress 时返回 |
| `oauthToken` | string | 从 Deployment 第一个容器环境变量 `OAUTH_TOKEN` 读取 |

## 前端关联

- 节点镜像：`w7panel-ui/src/views/cluster/nodes/image-list.vue`
- 构建镜像状态：`w7panel-ui/src/views/cluster/nodes/build-image-status.vue`
- 构建镜像抽屉：`w7panel-ui/src/components/build-image-drawer.vue`
- 微应用事件：`buildContainerImage`、`buildImage`，见 [../frontend/wujie-events.md](../frontend/wujie-events.md)

## 开发检查

- 镜像操作通常涉及节点 Agent，确认是否必须使用 `middleware.Proxy`。
- `registry.enabled=false` 时 `/v2/*path` 和 `/registry/containers/:id/commit` 不注册，前端需要处理不可用状态。
- Docker Registry v2 标准响应不能包装成普通 JSON。
- 镜像名、tag、digest、namespace、registry host 等参数要清理和校验。
- 不要在日志、URL、响应或前端可见字段中输出 registry password。
- 修改后端镜像接口时，同步检查前端节点镜像页、构建镜像抽屉和微应用事件。
