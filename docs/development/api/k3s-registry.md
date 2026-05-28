# w7panel-server/app/k3s-registry API 文档

## 通用约定

### 启用条件

`/v2/*path` 和 `/panel-api/v1/registry/containers/:id/commit` 只有在配置 `registry.enabled=true` 时注册。

`/panel-api/v1/registry/patch/images/*` 和 `/panel-api/v1/registry/server-info` 始终注册，其中 patch 接口会经过 `Auth` 和 `Proxy` 中间件，通常用于代理到 Agent 节点执行 containerd 操作。

## Docker Registry v2 兼容入口

### ANY `/v2/*path`

功能：本地 Docker Registry v2 兼容接口，底层由 containerd registry handler 处理。

启用条件：`registry.enabled=true`。

鉴权：当前路由未挂载用户鉴权中间件，访问控制由 registry handler 或部署层处理。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `path` | path | 是 | string | Docker Registry v2 路径 |
| request body | body | 否 | bytes/json | Docker Registry 标准请求体 |
| headers | header | 否 | string | Docker Registry 标准请求头 |

常见路径：

| 方法 | 路径 | 功能 |
|------|------|------|
| `GET` | `/v2/` | 检查 Registry API 版本 |
| `GET` | `/v2/_catalog` | 获取镜像仓库列表 |
| `GET` | `/v2/{name}/tags/list` | 获取镜像 tag 列表 |
| `GET` | `/v2/{name}/manifests/{reference}` | 获取 manifest |
| `PUT` | `/v2/{name}/manifests/{reference}` | 推送 manifest |
| `GET` | `/v2/{name}/blobs/{digest}` | 获取 blob |
| `HEAD` | `/v2/{name}/blobs/{digest}` | 检查 blob 是否存在 |
| `POST` | `/v2/{name}/blobs/uploads/` | 初始化 blob 上传 |
| `PUT` | `/v2/{name}/blobs/uploads/{uuid}` | 完成 blob 上传 |

响应参数：Docker Registry v2 标准响应。

## Registry 管理接口

### GET `/panel-api/v1/registry/server-info`

功能：获取面板 Registry 服务访问信息。主集群 token 下会返回通过面板代理访问 Agent Registry 的地址；K3k token 下可能直接使用 `/`。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `hostIp` | query | 否 | string | 目标节点 IP，用于定位 Agent Registry 服务 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `requestUrl` | string | 前端请求 Registry 时使用的基础 URL |
| `requestHost` | string | Registry 服务 Host，格式一般为 `ip:8000` |

### POST `/panel-api/v1/registry/containers/:id/commit`

功能：将指定容器提交为 containerd 镜像。

启用条件：`registry.enabled=true`。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `id` | path | 是 | string | 容器 ID |
| `ref` | context | 否 | string | 目标镜像引用，例如 `registry.example.com/ns/app:v1`。Controller 从 `ctx.GetString("ref")` 读取，通常由上游中间件或调用链注入 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `digest` | string | 提交后的镜像 digest |

## 镜像 Patch 接口

以下接口路由前缀为 `/panel-api/v1/registry/patch`，均经过 `Auth` 和 `Proxy` 中间件，实际在目标 Agent 环境中操作 containerd。

### GET `/panel-api/v1/registry/patch/images/list`

功能：获取 containerd 镜像列表。接口固定过滤 `dangling=false`。

鉴权：需要用户 token。

请求参数：无。

响应参数：`[]images.Image`，来源于 `github.com/containerd/nerdctl/v2/pkg/cmd/image.List`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | 镜像列表 |
| `[].Name` / `[].name` | string | 镜像名称，实际字段取决于 nerdctl `images.Image` JSON 定义 |
| `[].Tag` / `[].tag` | string | 镜像标签 |
| `[].ID` / `[].id` | string | 镜像 ID |
| `[].Size` / `[].size` | string/number | 镜像大小 |
| `[].Created` / `[].created` | string | 创建时间 |

### PUT `/panel-api/v1/registry/patch/images/tag`

功能：给本地 containerd 镜像打新标签。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `source` | form | 是 | string | 源镜像引用 |
| `target` | form | 是 | string | 目标镜像引用 |

响应参数：`"success"`。

### POST `/panel-api/v1/registry/patch/images/delete`

功能：删除本地 containerd 镜像。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `target` | form | 是 | string | 要删除的镜像引用 |
| `force` | form | 否 | bool | 是否强制删除 |
| `async` | form | 否 | bool | 是否异步删除 |

响应参数：`"success"`。

### POST `/panel-api/v1/registry/patch/images/label`

功能：修改本地 containerd 镜像 label。

鉴权：需要用户 token。

请求类型：`application/json`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | json | 是 | string | 镜像引用 |
| `labels` | json | 是 | object | 要设置的 label 键值对 |
| `replace` | json | 是 | bool | 是否替换全部已有 labels |

响应参数：`"success"`。

### POST `/panel-api/v1/registry/patch/images/import`

功能：从文件导入 containerd 镜像。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | form | 是 | string | 导入后的镜像引用 |
| `path` | form | 是 | string | 镜像文件路径；Agent 或 K3k virtual 环境下会拼接到 `s3.base_dir` |
| `pinned` | form | 否 | bool | Controller 接收但当前未传给底层导入逻辑 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 实际导入后的镜像名称 |

## 未注册 Controller 说明

`w7panel-server/app/k3s-registry/http/controller/containers.go` 和 `exec.go` 当前存在 Controller 代码，但未在 `provider.go` 中注册路由，因此不作为现有 HTTP API 记录。
