# w7panel-server/app/application API 文档

## OpenAPI 文档

### GET `/docs/openapi`

功能：打开 OpenAPI 文档页面。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `url` | query | 否 | string | OpenAPI spec 地址，默认 `/docs/openapi/spec` |

响应参数：HTML 页面。

### GET `/docs/openapi/spec`

功能：返回 OpenAPI 规格内容。

请求参数：无。

响应参数：OpenAPI JSON。

### GET `/openapi.json`

功能：OpenAPI JSON 兼容入口。

请求参数：无。

响应参数：OpenAPI JSON。

## 命名空间

### GET `/panel-api/v1/namespaces`

功能：获取当前 token 可访问的 Kubernetes 命名空间列表。获取失败时，如果能读取默认命名空间，会返回只包含默认命名空间的 `NamespaceList`。

请求参数：无。

响应参数：Kubernetes `corev1.NamespaceList`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | array | 命名空间数组 |
| `items[].metadata.name` | string | 命名空间名称 |
| `items[].metadata.*` | object | Kubernetes 原生 metadata 字段 |
| `items[].status.*` | object | Kubernetes 原生 namespace status 字段 |

## Helm

### GET `/panel-api/v1/helm/releases`

功能：获取 Helm Release 列表。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/query | 否 | string | 命名空间；为空时使用 SDK 默认命名空间 |
| `labelSelector` | form/query | 否 | string | Helm release labelSelector |

响应参数：Helm release 列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | Release 列表 |
| `[].name` | string | Release 名称 |
| `[].namespace` | string | 命名空间 |
| `[].revision` | number | Release revision |
| `[].updated` | string | 更新时间 |
| `[].status` | string | Release 状态 |
| `[].chart` | string | Chart 信息 |
| `[].appVersion` | string | 应用版本 |

### GET `/panel-api/v1/helm/releases/:name`

功能：获取指定 Helm Release 详情。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | form/query | 否 | string | Release 所在命名空间 |

响应参数：Helm `release.Release` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | Release 名称 |
| `namespace` | string | 命名空间 |
| `version` | number | Release revision |
| `info` | object | Release 状态、时间等信息 |
| `chart` | object | Chart 元信息和模板 |
| `config` | object | Release values |
| `manifest` | string | 渲染后的 Kubernetes YAML |
| `hooks` | array | Helm hooks |

### POST `/panel-api/v1/helm/releases/:name`

功能：从仓库安装 Helm Release。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | 要安装的 Release 名称 |
| `namespace` | form/json | 是 | string | 安装命名空间 |
| `repository` | form/json | 否 | string | Chart 仓库地址 |
| `chartName` | form/json | 是 | string | Chart 名称 |
| `vals` | form/json | 是 | object | Helm values |
| `chartType` | form/json | 否 | string | Chart 类型，当前 Controller 接收但未直接使用 |
| `version` | form/json | 否 | string | Chart 版本 |

响应参数：Helm `release.Release` 对象，字段同 `GET /helm/releases/:name`。

### DELETE `/panel-api/v1/helm/releases/:name`

功能：卸载 Helm Release。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | form/query | 否 | string | Release 所在命名空间 |

响应参数：Helm `release.UninstallReleaseResponse`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `release` | object | 被卸载的 Release 信息 |
| `info` | string | 卸载说明 |

### PUT `/panel-api/v1/helm/releases/:name/reuse`

功能：复用当前 values 更新 Helm Release。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | Release 名称 |
| `namespace` | form/json | 是 | string | Release 所在命名空间 |
| `vals` | form/json | 是 | object | 要合并或覆盖的 values |

响应参数：Helm `release.Release` 对象，字段同 `GET /helm/releases/:name`。

### GET `/panel-api/v1/app-info`

功能：获取当前面板自身 Helm 和部署元信息。无需鉴权。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `helmVersion` | string | 面板 Helm Chart 版本 |
| `helmReleaseName` | string | 面板 Helm Release 名称 |
| `helmNamespace` | string | 面板所在命名空间 |
| `metadataName` | string | 面板元数据名称 |
| `deploymentName` | string | 面板 Deployment 名称 |

## 终端与命令执行

### GET `/panel-api/v1/tty`

功能：打开面板本地终端；K3k 集群 token 下会进入 K3k server Pod。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `shell` | form/query | 否 | string | `/bin/sh` 或 `/bin/bash`，默认 `/bin/bash` |

响应参数：WebSocket 流，不返回 JSON。

### GET `/panel-api/v1/nodetty`

功能：打开指定节点终端。主集群通过 agent Pod 执行 `nsenter` 进入节点命名空间；K3k 虚拟集群按 Pod IP 查找集群 Pod。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `hostIp` | form/query | 是 | string | 节点内网 IP 或 K3k server Pod IP |
| `shell` | form/query | 否 | string | `/bin/sh` 或 `/bin/bash`，默认 `/bin/bash` |

响应参数：WebSocket 流，不返回 JSON。

### GET `/panel-api/v1/exec`

功能：在 Pod 容器中执行命令。普通 HTTP 请求返回命令输出；WebSocket 请求返回交互式流。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/query | 是 | string | Pod 命名空间 |
| `podName` | form/query | 是 | string | Pod 名称 |
| `containerName` | form/query | 是 | string | 容器名称 |
| `command` | form/query | 是 | string[] | 命令数组，例如 `command=ls&command=-la` |
| `tty` | form/query | 否 | bool | 是否开启 TTY |

响应参数：

| 场景 | 响应 |
|------|------|
| 普通 HTTP | 命令 stdout/stderr 字节流 |
| WebSocket | 交互式命令流 |

### POST `/panel-api/v1/exec2`

功能：同 `GET /exec`，用于通过 POST 提交执行参数。

请求参数：同 `GET /exec`。

响应参数：同 `GET /exec`。

### POST `/panel-api/v1/exec-all`

功能：批量在多个 Pod 中执行同一命令。不支持 WebSocket。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/json | 是 | string | Pod 命名空间 |
| `podNames` | form/json | 否 | string[] | Pod 名称列表 |
| `podName` | form/json | 否 | string | 单个 Pod 名称；`podNames` 为空时使用 |
| `containerName` | form/json | 是 | string | 容器名称 |
| `command` | form/json | 是 | string[] | 命令数组 |
| `tty` | form/json | 否 | bool | 是否开启 TTY |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | 执行结果列表 |
| `[].podName` | string | Pod 名称 |
| `[].output` | string | 命令输出 |
| `[].success` | bool | 是否成功 |
| `[].error` | string | 错误信息，失败时存在 |

### POST `/panel-api/v1/cp`

功能：调用 `kubectl cp` 在面板临时目录和 Pod 之间复制文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `from` | form | 是 | string | 源路径 |
| `to` | form | 是 | string | 目标路径 |
| `namespace` | form | 是 | string | Pod 命名空间 |
| `upload` | form | 是 | string | `1` 表示从面板上传到 Pod；其他值表示从 Pod 下载到面板 |
| `podName` | form | 是 | string | Pod 名称 |

响应参数：`"success"`。

### POST `/panel-api/v1/cppid`

功能：通过宿主机 PID rootfs 在面板临时目录和容器文件系统之间复制文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `from` | form | 是 | string | 源路径 |
| `to` | form | 是 | string | 目标路径 |
| `upload` | form | 是 | string | `1` 表示从面板临时目录复制到容器；其他值表示从容器复制到面板临时目录 |
| `pid` | form | 是 | string | 目标容器 PID |

响应参数：`"success"`。

### POST `/panel-api/v1/mvpid`

功能：当前路由绑定到 `CpPidFile`，行为同 `/cppid`。

请求参数：同 `/cppid`。

响应参数：`"success"`。

## PID 与挂载文件

### GET `/panel-api/v1/pid`

功能：根据 Pod、容器或 containerId 获取目标容器 PID，并生成文件管理相关 URL。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/query | 是 | string | 目标 Pod 命名空间 |
| `HostIp` | form/query | 是 | string | 目标 Pod 所在节点 IP，注意字段名首字母大写 |
| `containerId` | form/query | 否 | string | 容器 ID |
| `podName` | form/query | 否 | string | 目标 Pod 名称 |
| `containerName` | form/query | 否 | string | 目标容器名称 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `podName` | string | Agent Pod 名称 |
| `pid` | string | 主 PID |
| `subPid` | string | 子容器 PID；没有时为空字符串 |
| `namespace` | string | Agent Pod 命名空间 |
| `containerName` | string | Agent Pod 容器名称 |
| `podIp` | string | Agent Pod IP |
| `webdavUrl` | string | WebDAV 文件操作基础 URL |
| `webdavBasePath` | string | 前端用于过滤路径的 WebDAV base path |
| `compressUrl` | string | 压缩/解压基础 URL |
| `permissionUrl` | string | 权限修改基础 URL |
| `pwd` | string | 当前工作目录 |
| `agentUrl` | string | Agent 代理基础 URL |
| `webdavToken` | string | 当前请求 token |

### GET `/panel-api/v1/mountfiles`

功能：获取指定工作负载引用的 ConfigMap/Secret 挂载文件列表，可选择带出文件内容。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/query | 否 | string | 工作负载命名空间 |
| `apiVersion` | form/query | 是 | string | 工作负载 apiVersion |
| `kind` | form/query | 是 | string | 工作负载 Kind |
| `name` | form/query | 是 | string | 工作负载名称 |
| `includeContent` | form/query | 否 | bool | 是否返回文件内容 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 工作负载命名空间 |
| `apiVersion` | string | 工作负载 apiVersion |
| `kind` | string | 工作负载 Kind |
| `name` | string | 工作负载名称 |
| `mounts` | array | 挂载描述列表 |
| `mounts[].containerName` | string | 容器名称 |
| `mounts[].containerType` | string | 容器类型 |
| `mounts[].volumeName` | string | volume 名称 |
| `mounts[].mountPath` | string | 容器内挂载路径 |
| `mounts[].subPath` | string | subPath |
| `mounts[].readOnly` | bool | 是否只读 |
| `mounts[].sourceType` | string | 来源类型，如 ConfigMap/Secret |
| `mounts[].sourceName` | string | 来源资源名称 |
| `mounts[].files` | array | 文件列表 |
| `mounts[].files[].path` | string | 容器内文件绝对路径 |
| `mounts[].files[].relativePath` | string | 相对路径 |
| `mounts[].files[].sourceType` | string | 来源类型 |
| `mounts[].files[].sourceName` | string | 来源资源名称 |
| `mounts[].files[].key` | string | ConfigMap/Secret key |
| `mounts[].files[].optional` | bool | 是否 optional |
| `mounts[].files[].mode` | number | 文件权限十进制值 |
| `mounts[].files[].modeOctal` | string | 文件权限八进制字符串 |
| `mounts[].files[].content` | string | 文本内容，`includeContent=true` 时可能返回 |
| `mounts[].files[].contentBase64` | string | Base64 内容，二进制或非 UTF-8 时可能返回 |

### POST `/panel-api/v1/mountfiles`

功能：创建挂载文件。底层会更新对应 ConfigMap 或 Secret。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form | 否 | string | 工作负载命名空间 |
| `apiVersion` | form | 是 | string | 工作负载 apiVersion |
| `kind` | form | 是 | string | 工作负载 Kind |
| `name` | form | 是 | string | 工作负载名称 |
| `path` | form | 是 | string | 容器内文件路径 |
| `containerName` | form | 否 | string | 容器名称 |
| `content` | form | 否 | string | 文件内容 |
| `mode` | form | 否 | string | 文件权限 |

响应参数：`"success"`。

### PUT `/panel-api/v1/mountfiles`

功能：更新挂载文件内容。

请求参数：同 `POST /mountfiles`。

响应参数：`"success"`。

### DELETE `/panel-api/v1/mountfiles`

功能：删除挂载文件。

请求参数：同 `POST /mountfiles`。

响应参数：`"success"`。

### PUT `/panel-api/v1/mountfiles/chmod`

功能：修改挂载文件权限。

请求参数：同 `POST /mountfiles`，其中 `mode` 必填。

响应参数：`"success"`。

### GET `/panel-api/v1/nodepid`

功能：获取节点或 Agent 相关 PID 信息。该路由绑定 `PodExec.GetNodePid`，用于文件管理 Agent 场景。

请求参数：与具体 Agent/PID 获取场景相关，常用参数包括节点 IP、Pod、容器信息。

响应参数：PID 信息对象。

## YAML 与 Compose

### POST `/panel-api/v1/yaml`

功能：应用 Kubernetes YAML 到集群。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query | 否 | string | 默认命名空间；为空时使用 SDK 默认命名空间 |
| request body | body | 是 | string | Kubernetes YAML 内容 |

响应参数：`"success"`。

### PUT `/panel-api/v1/rollback`

功能：回滚 Kubernetes 工作负载到指定 revision。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form | 是 | string | 命名空间 |
| `name` | form | 是 | string | 资源名称 |
| `kind` | form | 是 | string | 资源 Kind |
| `apiVersion` | form | 是 | string | 资源 apiVersion |
| `toRevision` | form | 是 | int64 | 目标 revision。当前结构体字段为小写，实际绑定可能不可用，需后续修正代码 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `message` | string | 回滚结果说明 |

### POST `/panel-api/v1/kcompose`

功能：将 Docker Compose YAML 转换为 Kubernetes YAML。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| request body | body | 是 | string | Docker Compose YAML 内容 |

响应参数：`map[string]string`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `<resource-file>` | string | 转换后的 Kubernetes YAML 内容 |

## 工具接口

### POST `/panel-api/v1/pinyin`

功能：将中文转换为拼音。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `words` | form | 是 | string | 待转换文本 |

响应参数：字符串数组或拼音结果数组，由 `go-pinyin` 返回。

### GET `/panel-api/v1/dnsip`

功能：查询域名解析 IP。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `domain` | form/query | 是 | string | 域名 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | string[] | IP 地址列表；解析失败返回空数组 |

### GET `/panel-api/v1/dns-cname`

功能：查询域名 CNAME。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `domain` | form/query | 是 | string | 域名 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | string[] | CNAME 列表；解析失败返回空数组 |

### GET `/panel-api/v1/myip`

功能：获取服务器出口公网 IP。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `ip` | string | 出口公网 IP；获取失败时可能返回空数组 |

### POST `/panel-api/v1/db-conn-test`

功能：测试数据库 DSN 是否可连接。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `dsn` | form | 是 | string | 数据库连接 DSN |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canConnect` | bool | 是否可连接 |
| `msg` | string | 错误信息，连接失败时返回 |

### POST `/panel-api/v1/ping-etcd`

功能：检查 etcd `/health` 是否可访问。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `url` | form | 是 | string | etcd 地址，接口会自动补 `/health` |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canConnect` | bool | HTTP 状态码是否为 200 |
| `msg` | string | 错误信息，请求失败时返回 |

### GET `/panel-api/v1/captcha`

功能：生成滑块验证码。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | number | 固定为 0 |
| `captcha_key` | string | 加密后的验证码 key |
| `image_base64` | string | 背景图 Base64 |
| `tile_base64` | string | 滑块图 Base64 |
| `tile_width` | number | 滑块宽度 |
| `tile_height` | number | 滑块高度 |
| `tile_x` | number | 滑块目标 X 坐标 |
| `tile_y` | number | 滑块目标 Y 坐标 |

### POST `/panel-api/v1/verify-captcha`

功能：验证滑块验证码。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `point` | form | 是 | string | 前端提交的滑块坐标 |
| `key` | form | 是 | string | `captcha` 返回的 `captcha_key` |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `ok` | bool | 是否验证通过 |
| `msg` | string | 错误信息，验证失败时返回 |

## 代理接口

### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy-root/*path`

功能：代理访问指定 Service，使用集群内域名，不经过 `Proxy` 中间件转发。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | Service 命名空间 |
| `name` | path | 是 | string | Service 名称，可携带端口或协议，格式见下方 |
| `path` | path | 是 | string | 代理路径 |

`name` 支持格式：

| 格式 | 示例 | 说明 |
|------|------|------|
| `host` | `nginx` | 默认 `http://host:80` |
| `host:port` | `nginx:8080` | 默认 http |
| `schema:host:port` | `https:nginx:443` | 指定协议和端口 |

响应参数：被代理服务的原始响应。

### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy/*path`

功能：代理访问指定 Service，经过 `Proxy` 中间件。

请求参数：同 `proxy-root`。

响应参数：被代理服务的原始响应。

### ANY `/panel-api/v1/namespaces/:namespace/pods/:name/proxy/*path`

功能：代理访问指定 Pod IP 和端口。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | Pod 命名空间 |
| `name` | path | 是 | string | Pod 名称，可用 `podName:port` 指定端口 |
| `path` | path | 是 | string | 代理路径 |

响应参数：被代理 Pod 的原始响应。

### ANY `/panel-api/v1/:name/proxy/*path`

功能：通用代理，直接按 `name` 组装目标地址。

请求参数：同 Service 代理的 `name` 和 `path`。

响应参数：被代理地址的原始响应。

### ANY `/panel-api/v1/proxy-url/`

功能：请求指定 URL 并返回文本内容。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `proxyUrl` | form/query | 是 | string | 要请求的完整 URL |

响应参数：目标 URL 响应文本；请求失败返回空字符串。

### ANY `/panel-api/v1/namespaces/:namespace/services/:name/proxy-no/*path`

功能：无需用户鉴权的 Service 代理入口。由 `ProxyNoAuth` 中间件处理访问控制。

请求参数：同 Service 代理。

响应参数：被代理服务的原始响应。

### GET `/panel-api/v1/kubeconfig`

功能：生成当前 token 对应的 kubeconfig。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `apiServerUrl` | query | 否 | string | kubeconfig 中使用的 API Server 地址 |

响应参数：kubeconfig 内容对象或字符串，取决于 `Sdk.ToKubeconfig` 返回值。

## Longhorn

### GET `/panel-api/v1/longhorn/need-delete-replica`

功能：根据磁盘选择器和节点 ID 获取需要删除的 Longhorn 副本。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `diskselector` | form/query | 是 | string | 磁盘选择器 |
| `nodeid` | form/query | 是 | string | 节点 ID，多个用逗号分隔 |

响应参数：Longhorn replica 列表。

### GET `/panel-api/v1/longhorn/volumes/status`

功能：获取 Longhorn PVC 对应卷状态。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `convertpvc` | form/query | 否 | string | Controller 接收但当前未使用 |

响应参数：map，key 为 `pvcName:namespace`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `<pvc>:<namespace>.numberOfReplicas` | number | 副本数 |
| `<pvc>:<namespace>.robustness` | string | 健康状态 |
| `<pvc>:<namespace>.size` | int64 | 规格容量 |
| `<pvc>:<namespace>.actualSize` | int64 | 实际使用容量 |
| `<pvc>:<namespace>.creationTimestamp` | string | 创建时间 |
| `<pvc>:<namespace>.accessMode` | string | 访问模式 |
| `<pvc>:<namespace>.snapShotSize` | int64 | 快照大小 |
| `<pvc>:<namespace>.isExpanding` | bool | 是否扩容中 |
| `<pvc>:<namespace>.expandErr` | string | 扩容错误 |
| `<pvc>:<namespace>.state` | string | 卷状态 |
| `<pvc>:<namespace>.volumeName` | string | Longhorn 卷名称 |
| `<pvc>:<namespace>.isLock` | string | 是否锁定，字符串布尔值 |
| `<pvc>:<namespace>.lockNodeId` | string | 锁定节点 |
| `<pvc>:<namespace>.attachedNodeId` | string | 当前挂载节点 |

### POST `/panel-api/v1/longhorn/install`

功能：从 `KO_DATA_PATH/yaml/longhorn/longhornfull.yaml` 安装 Longhorn。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query | 否 | string | 安装命名空间；为空时使用当前 SDK 默认命名空间 |

响应参数：`"success"`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/attach`

功能：挂载 Longhorn 卷。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |
| `hostId` | json | 是 | string | 目标节点 ID |
| `disableFrontend` | json | 否 | bool | 是否禁用 frontend |
| `AttachedBy` | json | 否 | string | 挂载发起方 |
| `AttachmentID` | json | 否 | string | attachment ID，默认 `longhorn-ui` |
| `attacherType` | json | 否 | string | attacher 类型 |

响应参数：`"success"`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/detach`

功能：卸载 Longhorn 卷。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |
| `forceDetach` | json | 否 | bool | 是否强制卸载 |
| `attachmentID` | json | 否 | string | attachment ID，默认 `longhorn-ui` |
| `hostId` | json | 否 | string | 节点 ID，Controller 接收但未使用 |

响应参数：`null`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/cancel-expansion`

功能：取消 Longhorn 卷扩容。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |

响应参数：`null`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/trim-filesystem`

功能：执行 Longhorn 卷文件系统 trim。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |

响应参数：`null`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-delete`

功能：删除 Longhorn 卷快照。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |
| `name` | json | 否 | string | 快照名称 |

响应参数：`null`。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-purge`

功能：清理 Longhorn 卷快照。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn 卷名称 |

响应参数：`null`。

## KubeBlocks

### GET `/panel-api/v1/kubeblocks/installjobyaml`

功能：生成安装 KubeBlocks 的 Kubernetes Job YAML/对象。

请求参数：无。

响应参数：Kubernetes `batchv1.Job` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata` | object | Job 元信息 |
| `spec` | object | Job 规格 |
| `spec.template` | object | Pod 模板 |

### POST `/panel-api/v1/kubeblocks/install`

功能：在 `default` 命名空间创建 KubeBlocks 安装 Job。

请求参数：无。

响应参数：Kubernetes `batchv1.Job` 对象。

## 静态资源

### GET `/panel-api/v1/static/:identifie/status`

功能：查询应用静态资源下载状态。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `identifie` | path | 是 | string | 静态资源标识 |
| `version` | query | 否 | string | 版本 |
| `releaseName` | query | 否 | string | Release 名称，包含 `-root` 时会被移除 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | 下载状态 |

### POST `/panel-api/v1/static/:namespace/download/:name`

功能：触发应用静态资源下载。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | AppGroup 命名空间 |
| `name` | path | 是 | string | AppGroup 名称；包含 `-root` 时使用 root SDK 查询 |

响应参数：该 Controller 未显式写入 JSON，下载逻辑由 `appgroup.DownStatic` 执行。

## GPU

### GET `/panel-api/v1/gpu/config`

功能：获取 GPU 管理状态。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | form/query | 否 | string | Controller 接收但当前创建 GpuManager 时未使用 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `gpuOperatorIsSuccess` | bool | GPU Operator 是否安装成功 |
| `hamiIsSuccess` | bool | HAMi 是否安装成功 |
| `gpuOperatorIsDeployed` | bool | GPU Operator 是否已部署 |
| `hamiIsDeployed` | bool | HAMi 是否已部署 |
| `canInstallGpuOperator` | bool | 是否可安装 GPU Operator |
| `canInstallHami` | bool | 是否可安装 HAMi |
| `gpuEnabled` | bool | GPU 功能是否启用 |
| `canEnabledGpu` | bool | 是否可启用 GPU |
| `gpuOperatorMode` | string | GPU Operator 模式 |
| `hamiMode` | string | HAMi 模式 |
| `vgpuMode` | string | vGPU 模式 |

### POST `/panel-api/v1/gpu/enabled-gpu`

功能：启用或关闭 GPU 能力。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `enabled` | form | 否 | bool | 是否启用 GPU |

响应参数：`"success"`。

### POST `/panel-api/v1/gpu/install-hami`

功能：安装 HAMi。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `runtimeClassName` | form | 否 | string | RuntimeClass 名称 |

响应参数：`"success"`。

### POST `/panel-api/v1/gpu/install-gpu-operator`

功能：安装 GPU Operator。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `driverVersion` | form | 否 | string | 驱动版本 |
| `driverEnabled` | form | 否 | bool | 是否启用驱动安装 |

响应参数：`"success"`。

### GET `/panel-api/v1/gpu/hami/metrics/real`

功能：获取节点 HAMi 实时 GPU 利用率。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | 节点 GPU 指标列表 |
| `[].NodeName` | string | 节点名称 |
| `[].HostCoreUtilization` | string | 节点 GPU 核心利用率字符串 |

### GET `/panel-api/v1/gpu/summary`

功能：获取集群 GPU 汇总。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `GPUDeviceSharedNum` | int32 | 已分配 vGPU 数 |
| `GPUDeviceSharedTotal` | int32 | vGPU 总数 |
| `GPUDeviceCoreAllocated` | int32 | 已分配算力 |
| `GPUDeviceCoreLimit` | int32 | 算力总量 |
| `GPUDeviceMemoryAllocated` | int64 | 已分配显存 |
| `GPUDeviceMemoryAllocatedTotal` | int32 | 显存总量 |

### GET `/panel-api/v1/gpu/node/devices`

功能：获取各节点 GPU 设备列表。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | GPU 设备列表 |
| `[].id` | string | 设备 ID |
| `[].aliasId` | string | 设备别名 |
| `[].index` | number | 设备序号 |
| `[].count` | int32 | 数量 |
| `[].devmem` | int32 | 显存 |
| `[].devcore` | int32 | 算力 |
| `[].type` | string | 设备类型 |
| `[].numa` | number | NUMA 编号 |
| `[].mode` | string | 模式 |
| `[].health` | bool | 是否健康 |
| `[].driver` | string | 驱动版本 |
| `[].nodeName` | string | 节点名称 |

### POST `/panel-api/v1/gpu/gpustack/worker`

功能：创建 GPUStack worker。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `image` | form | 是 | string | worker 镜像 |
| `pvcName` | form | 是 | string | PVC 名称 |
| `serverUrl` | form | 是 | string | GPUStack Server 地址 |
| `group` | form | 是 | string | worker 分组 |
| `password` | form | 否 | string | 密码 |
| `workerToken` | form | 是 | string | worker token |
| `namespace` | form | 是 | string | 部署命名空间 |
| `gpucores` | form | 是 | string | GPU 算力 |
| `gpu` | form | 是 | string | GPU 数量 |
| `gpumem` | form | 是 | string | GPU 显存 |
| `cpu` | form | 否 | string | CPU 请求/限制 |
| `memory` | form | 否 | string | 内存请求/限制 |
| `runtimeClassName` | form | 否 | string | RuntimeClass 名称 |

响应参数：空数组 `[]`。

## WebDAV、压缩、权限

### ANY `/panel-api/v1/files/webdav-agent/:pid/agent/*path`

功能：通过 WebDAV 操作目标容器文件系统。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 目标容器 PID |
| `path` | path | 是 | string | 文件路径 |
| `Depth` | header | 否 | string | WebDAV PROPFIND 深度 |
| `Destination` | header | MOVE/COPY 时 | string | 目标路径 |
| request body | body | 否 | bytes/xml | PUT 写入内容或 WebDAV XML |

响应参数：WebDAV 标准响应。`PROPFIND` 返回 XML；`GET` 返回文件内容；其他方法按 WebDAV 状态码返回。

### ANY `/panel-api/v1/files/webdav-agent/:pid/subagent/:subpid/agent/*path`

功能：通过 WebDAV 操作子进程容器文件系统。

请求参数：比普通 WebDAV 多 `subpid` path 参数。

响应参数：同普通 WebDAV。

### ANY `/panel-api/v1/files/webdav-test/*path`

功能：WebDAV 测试入口，无鉴权。

请求参数：WebDAV 标准参数。

响应参数：WebDAV 标准响应。

### POST `/panel-api/v1/files/compress-agent/:pid/compress`

功能：压缩目标容器内文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 目标容器 PID |
| `sources` | json/form | 是 | string[] | 要压缩的文件或目录路径 |
| `output` | json/form | 是 | string | 输出压缩包路径 |

响应参数：`"success"`。

### POST `/panel-api/v1/files/compress-agent/:pid/extract`

功能：解压目标容器内压缩包。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 目标容器 PID |
| `source` | json/form | 是 | string | 压缩包路径 |
| `target` | json/form | 是 | string | 解压目标目录 |

响应参数：`"success"`。

### POST `/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/compress`

功能：压缩子进程容器内文件。

请求参数：比普通压缩多 `subpid` path 参数，其余同 `/compress`。

响应参数：`"success"`。

### POST `/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/extract`

功能：解压子进程容器内压缩包。

请求参数：比普通解压多 `subpid` path 参数，其余同 `/extract`。

响应参数：`"success"`。

### POST `/panel-api/v1/files/permission-agent/:pid/chmod`

功能：修改目标容器内文件权限。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 目标容器 PID |
| `path` | json/form | 是 | string | 文件路径 |
| `mode` | json/form | 是 | string | 八进制权限字符串，如 `755` |
| `recursive` | json/form | 否 | bool | 是否递归处理 |

响应参数：`"success"`。

### POST `/panel-api/v1/files/permission-agent/:pid/chown`

功能：修改目标容器内文件所有者。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | path | 是 | string | 目标容器 PID |
| `path` | json/form | 是 | string | 文件路径 |
| `owner` | json/form | 是 | string | 所有者，支持 `uid`、`user`、`user:group`、`user.group` |
| `recursive` | json/form | 否 | bool | 是否递归处理 |

响应参数：`"success"`。

### POST `/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chmod`

功能：修改子进程容器内文件权限。

请求参数：比普通 chmod 多 `subpid` path 参数，其余同 `/chmod`。

响应参数：`"success"`。

### POST `/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chown`

功能：修改子进程容器内文件所有者。

请求参数：比普通 chown 多 `subpid` path 参数，其余同 `/chown`。

响应参数：`"success"`。

## 分片上传与下载

### GET `/panel-api/v1/download/*path`

功能：从 `s3.base_dir` 下载文件。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `path` | path | 是 | string | 相对 `s3.base_dir` 的文件路径 |

响应参数：文件流，Header 包含：

| Header | 说明 |
|--------|------|
| `Content-Type` | `application/octet-stream` |
| `Content-Disposition` | 附件下载文件名 |

### POST `/panel-api/v1/files/upload-chunk`

功能：上传文件分片到 `s3.base_dir/.chunks/{identifier}`。

请求类型：`multipart/form-data`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `file` | form file | 是 | file | 当前分片文件 |
| `chunkIndex` | form | 是 | string/int | 分片序号，从 0 开始 |
| `chunkTotal` | form | 是 | string/int | 总分片数 |
| `identifier` | form | 是 | string | 文件唯一标识 |
| `fileName` | form | 否 | string | 原始文件名 |
| `relativePath` | form | 否 | string | 相对路径，当前预留 |
| `fileSize` | form | 否 | string/int64 | 文件总大小，当前仅解析预留 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `chunkExists` | bool | 分片是否已存在 |
| `chunkIndex` | number | 分片序号 |
| `chunkTotal` | number | 总分片数；分片已存在时可能不返回 |
| `written` | number | 本次写入字节数；分片已存在时不返回 |

### GET `/panel-api/v1/files/check-chunk`

功能：检查指定分片是否已经上传。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `identifier` | query | 是 | string | 文件唯一标识 |
| `chunkIndex` | query | 是 | string/int | 分片序号 |
| `chunkTotal` | query | 是 | string/int | 总分片数 |
| `fileName` | query | 否 | string | 文件名 |
| `relativePath` | query | 否 | string | 相对路径 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `chunkExists` | bool | 分片是否存在 |

### POST `/panel-api/v1/files/merge-chunks`

功能：合并已上传分片。传入 `pid` 时直接合并到目标容器 rootfs，否则合并到 `s3.base_dir`。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `identifier` | json | 是 | string | 文件唯一标识 |
| `fileName` | json | 是 | string | 合并后的文件名或相对路径 |
| `totalChunks` | json | 是 | int | 总分片数 |
| `fileSize` | json | 否 | int64 | 文件总大小，当前预留 |
| `pid` | json | 否 | string | 目标容器 PID |
| `subPid` | json | 否 | string | 子容器 PID |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `fileUrl` | string | 下载 URL |
| `fileName` | string | 文件名 |
| `fileSize` | int64 | 合并后写入字节数 |

### POST `/panel-api/v1/files/mvtopod`

功能：将临时目录中的文件移动到目标容器文件系统。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `pid` | form | 是 | string | 目标容器 PID |
| `subpid` | form | 否 | string | 子容器 PID，`0` 会被当作空值 |
| `fromPath` | form | 是 | string | 相对系统临时目录的源路径 |
| `toPath` | form | 是 | string | 容器内目标路径 |

响应参数：`"success"`。

### ANY `/panel-api/v1/s3bucket`

功能：S3 兼容上传入口，数据落到 `s3.base_dir`。

请求参数：S3 协议参数。

响应参数：S3 兼容响应。

### ANY `/s3bucket`

功能：S3 兼容上传入口，兼容不支持多路径的 s3 fake server。

请求参数：S3 协议参数。

响应参数：S3 兼容响应。

## 公开站点接口

### GET `/panel-api/v1/noauth/site/beian`

功能：从 `default/beian` ConfigMap 获取备案信息。无需鉴权。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `icpnumber` | string | ICP 备案号 |
| `number` | string | 公安备案号 |
| `location` | string | 备案地区 |

如果 ConfigMap 不存在，返回 `"success"`。

### GET `/panel-api/v1/noauth/site/beian2`

功能：使用内部 Kubernetes client 获取备案信息。无需鉴权。

请求参数：无。

响应参数：同 `/noauth/site/beian`。

### GET `/panel-api/v1/noauth/site/k3k-config`

功能：获取 K3k 公开配置。无需鉴权。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `indexpage` | string | 首页类型或登录页配置 |

如果 ConfigMap 不存在，返回 `"success"`。

### GET `/panel-api/v1/noauth/site/init-user`

功能：获取初始化用户、Console 注册和验证码开关状态。无需鉴权。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canInitUser` | string | 是否允许初始化用户，字符串布尔值 |
| `allowConsoleRegister` | string | 是否允许 Console 注册，字符串布尔值 |
| `captchaEnabled` | string | 是否开启验证码，字符串布尔值 |

### GET `/panel-api/v1/noauth/site/lianxi`

功能：获取联系信息 ConfigMap 列表。无需鉴权。

请求参数：无。

响应参数：Kubernetes `corev1.ConfigMapList`。查询失败返回空 `ConfigMapList`。

### GET `/panel-api/v1/noauth/site/{name}/configmap`

功能：获取允许公开访问的 ConfigMap。仅当 ConfigMap label `w7.cc/noauth=true` 时返回真实对象。无需鉴权。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | ConfigMap 名称 |

响应参数：Kubernetes `corev1.ConfigMap`；不存在或未授权公开时返回空 ConfigMap。

## MicroApp 与容器镜像

### GET `/panel-api/v1/microapp/top`

功能：获取顶层 MicroApp 列表。

请求参数：无。

响应参数：MicroApp 列表。具体字段由 `common/service/k8s/microapp.ListTop` 返回。

### GET `/panel-api/v1/microapp/:name/info`

功能：获取指定 MicroApp 详情。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | MicroApp 名称 |

响应参数：MicroApp 详情对象。具体字段由 `common/service/k8s/microapp.ListInfo` 返回。

### ANY `/panel-api/v1/microapp/:name/proxy/*path`

功能：代理访问 MicroApp。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | MicroApp 名称 |
| `path` | path | 是 | string | MicroApp 内部路径 |

响应参数：被代理 MicroApp 的原始响应。

### POST `/panel-api/v1/containers/image/export-push`

功能：从容器 rootfs 导出镜像并推送到指定镜像仓库，响应为 SSE 文本流。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `containerId` | form | 是 | string | 容器 ID |
| `registryDomain` | form | 是 | string | 镜像仓库域名 |
| `imageName` | form | 是 | string | 推送的镜像名 |

响应参数：`text/event-stream`，每行是一条进度文本。
