# 云主机 API

面板中 K3k/CKM/CVM 统一按“云主机”理解。本文档说明当前 `w7panel-server/app/k3k` 实际注册的云主机用户、CKM/CVM、同步、菜单和套餐接口。登录、token、用户信息兼容入口的凭据说明见 [credentials.md](./credentials.md)；云主机订单、授权订单和资源超卖接口见 [orders.md](./orders.md)。

## 整体使用方式

云主机接口用于处理 K3k 用户、子集群、CKM/CVM 资源、资源套餐和子集群同步。开发时先通过用户信息接口确定当前用户身份和权限，再根据角色决定是查看当前云主机、切换登录其它云主机、还是操作 CKM/CVM 列表。

### 基本流程

1. 使用用户 token 调用 `/panel-api/v1/k3k/info` 或 `/panel-api/v1/userinfo` 获取当前用户、角色、集群和权限信息。
2. 需要初始化当前用户云主机时，调用 `/panel-api/v1/k3k/init`。
3. Founder 或 `w7panel` 用户需要切换到指定云主机时，调用 `/panel-api/v1/k3k/login` 获取目标用户 token。
4. 查询云主机实例列表和详情时，使用 CKM/CVM 接口。
5. 子集群资源同步由 `/panel-api/v1/k3k/sync-*` 接口处理，订单和资源超卖进入 [orders.md](./orders.md)。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 当前用户与权限 | `/panel-api/v1/k3k/info`、`/panel-api/v1/userinfo` | 返回 `K3kUser.ToArray()` |
| 初始化当前云主机 | `/panel-api/v1/k3k/init` | 当前实现直接返回成功，调用前确认预期 |
| 切换云主机用户 | `/panel-api/v1/k3k/login` | Founder 或 `w7panel` 用户使用 |
| CKM/CVM 列表与详情 | `/panel-api/v1/k3k/ckm`、`/cvm` | 查询云主机资源对象 |
| 子集群同步 | `/panel-api/v1/k3k/sync-*` | 同步 Ingress、ConfigMap、Secret、MicroApp 等 |
| 套餐列表 | `/panel-api/v1/idc-list` | 商店可展示的资源套餐 |

### 使用边界

- 云主机是面板侧业务概念，底层仍对应 K3k、CKM/CVM 和 K8s 资源。
- 普通用户只能访问自身上下文；Founder 查询 CKM/CVM 时可指定 namespace。
- 同步接口当前未挂载用户鉴权中间件，部署侧应限制来源。
- 历史未注册接口不作为可调用 API 维护，具体以 `app/k3k/provider.go` 为准。

## 通用说明

### 鉴权与凭据

云主机接口默认使用面板用户 token，传递方式和其他面板业务 API 一致：

```http
Authorization: Bearer <user-token>
```

凭据来源、刷新方式和兼容取 token 位置见 [credentials.md](./credentials.md)。本文不重复说明登录、refresh token 和前端注入逻辑。

| 接口类型 | 鉴权 | 说明 |
|----------|------|------|
| 用户信息、初始化、CKM/CVM、菜单 | `Authorization: Bearer <user-token>` | 经过 `middleware.Auth`，后端从 token 解析 K3k 用户和命名空间 |
| 切换云主机用户 | `Authorization: Bearer <founder-or-w7panel-token>` | 路由经过 `middleware.Auth`，Controller 内限制 Founder 或 `w7panel` 用户 |
| CVM/CKM 登录 | `Authorization: Bearer <user-token>` | 返回目标 K3k 用户 token 和 refreshToken |
| 同步接口 | 无用户 token | 当前路由未挂 `middleware.Auth`，应由部署侧限制来源 |
| 套餐列表 | 无用户 token | 公开展示商店可用资源套餐 |

使用边界：

- `k3k/login` 和 `cvm/:namespace/action/:name/login` 返回的是目标云主机用户 token，可作为后续面板 API 的用户 token 使用。
- `refreshToken` 只用于刷新登录态，不用于普通业务 API。
- 同步接口不应暴露到公网或不可信网络。
- `idc-list` 无需用户 token，但只用于展示套餐，不代表用户可购买或可操作资源。

### 响应格式

云主机接口成功时直接返回业务对象，不额外包裹 `code`、`data`、`message`。用户信息和菜单返回 `K3kUser.ToArray()` 生成的字段 map；登录类接口返回目标用户 `token`、`expire`、`isK3kUser`、`refreshToken` 等字段；同步和初始化类接口多返回 JSON 字符串 `"success"`；套餐列表返回 `types.Params` 结构。

### 参数位置

用户信息、菜单和套餐列表接口通常无请求参数。`/panel-api/v1/k3k/login` 使用 form 参数；CKM/CVM 列表按 query 指定 namespace；详情和登录操作使用 path 中的 namespace、name；同步接口使用 form 参数，具体字段见对应接口说明。

## 能力概览

| 能力 | 说明 |
|------|------|
| 用户与集群 | 查询当前云主机用户、角色、权限和 K3k 集群信息 |
| 用户切换 | Founder 或 `w7panel` 用户切换到指定云主机用户 |
| 同步 | 主集群和 K3k 子集群之间同步 Ingress、ConfigMap、Secret、MicroApp 等资源 |
| CKM/CVM | 查询云主机实例列表、详情，并登录指定 CVM/CKM |
| 菜单与套餐 | 查询当前用户菜单权限和可展示资源套餐 |

## 用户与集群

### GET `/panel-api/v1/k3k/info`

功能：获取当前 token 对应的 K3k 用户、集群、角色和功能权限信息。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：`K3kUser.ToArray()` 返回的 `map[string]string`。常用字段如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| `w7.cc/username` | string | 用户名 |
| `w7.cc/role` | string | 用户角色 |
| `w7.cc/user-mode` | string | 用户模式 |
| `w7.cc/menu` | string | 菜单权限 |
| `w7.cc/quota-limit` | string | 配额限制 |
| `w7.cc/file-editor` | string | 文件编辑权限 |
| `w7.cc/web-shell` | string | Web Shell 权限 |
| `w7.cc/domain-white-list` | string | 域名白名单 |
| `w7.cc/demo-user` | string | 是否演示用户 |
| `w7.cc/has-password` | string | 是否设置密码 |
| `w7.cc/can-init-cluster` | string | 是否可初始化集群 |
| `w7.cc/can-renew` | string | 是否可续费 |
| `w7.cc/need-renew` | string | 是否需要续费 |
| `w7.cc/can-expand` | string | 是否可扩容 |
| `k3k.io/name` | string | K3k 集群名称 |
| `k3k.io/namespace` | string | K3k 集群命名空间 |
| `k3k.io/cluster-mode` | string | 集群模式 |
| `k3k.io/cluster-policy` | string | 集群策略 |
| `k3k.io/cluster-status` | string | 集群状态 |
| `k3k.io/expire-time` | string | 到期时间 |
| `w7.cc/server-pod-name` | string | K3k server Pod 名称 |
| `w7.cc/server-container-name` | string | K3k server 容器名称 |

### GET `/panel-api/v1/userinfo`

功能：兼容入口，行为同 `/panel-api/v1/k3k/info`。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：同 `/panel-api/v1/k3k/info`。

### POST `/panel-api/v1/k3k/init`

功能：初始化当前 K3k 用户对应的集群。当前 Controller 返回成功，实际初始化逻辑保留在代码注释中，调用前需要结合实现确认预期效果。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：

```json
"success"
```

### POST `/panel-api/v1/k3k/login`

功能：创始人或 `w7panel` 用户切换登录到指定 K3k 用户，返回该用户 token。

鉴权：`Authorization: Bearer &lt;founder-or-w7panel-token&gt;`。路由经过 `middleware.Auth`，Controller 内要求当前用户为 Founder 或用户名为 `w7panel`。

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `k3kUserName` | form | 是 | string | 目标 K3k 用户名，即 ServiceAccount 名称 |
| `cvmName` | form | 是 | string | 关联的 CVM/CKM 名称 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | 目标用户 token |
| `expire` | int64 | token 过期时间，Unix 秒 |
| `isK3kUser` | bool | 是否 K3k 用户 |
| `refreshToken` | string | 刷新 token |

## 同步接口

同步接口用于在主集群和 K3k 子集群之间同步资源。以下接口当前未挂载用户鉴权中间件，调用方应限制来源并避免暴露到不可信网络。

鉴权：无需用户 token。

通用请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `virtualName` | form | 否 | string | 子集群内资源名称 |
| `virtualNamespace` | form | 否 | string | 子集群内资源命名空间 |
| `k3kName` | form | 否 | string | K3k 集群名称 |
| `k3kNamespace` | form | 否 | string | K3k 集群命名空间 |
| `k3kMode` | form | 否 | string | K3k 模式，如 `virtual` |
| `version` | form | 否 | string | 版本号 |
| `ckmName` | form | 否 | string | CKM/CVM 名称，`sync-microapp` 使用 |

| 方法 | 路径 | 功能 | 响应 |
|------|------|------|------|
| `POST` | `/panel-api/v1/k3k/sync-ingress` | 同步 Ingress | `"success"` |
| `POST` | `/panel-api/v1/k3k/sync-configmap` | 同步 ConfigMap | `"success"` |
| `POST` | `/panel-api/v1/k3k/sync-mcpbridge` | 同步 MCP Bridge | `"success"` |
| `POST` | `/panel-api/v1/k3k/sync-secret` | 同步 Secret | `"success"` |
| `POST` | `/panel-api/v1/k3k/sync-down-static` | 触发子集群静态资源下载 | `"success"` |
| `POST` | `/panel-api/v1/k3k/sync-microapp` | 将 MicroApp 同步到子集群 | `"success"` |

## CKM/CVM

### GET `/panel-api/v1/k3k/ckm`

功能：获取 CKM 列表。Founder 可通过 `namespace` 查询指定命名空间；非 Founder 强制使用当前 token 所属命名空间。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query | 否 | string | 命名空间，仅 Founder 查询时生效 |

响应参数：`CkmList` Kubernetes 自定义资源列表；每个条目会执行 `ComputeStatus()` 后返回。

### GET `/panel-api/v1/k3k/cvm`

功能：旧版 CVM 列表入口，行为同 `/panel-api/v1/k3k/ckm`。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：同 `/panel-api/v1/k3k/ckm`。

响应参数：同 `/panel-api/v1/k3k/ckm`。

### GET `/panel-api/v1/k3k/ckm/v1/:namespace/info/:name`

功能：获取指定 CKM 详情。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | CKM 命名空间 |
| `name` | path | 是 | string | CKM 名称 |

响应参数：`Ckm` Kubernetes 自定义资源对象；返回前会执行 `ComputeStatus()`。

### GET `/panel-api/v1/k3k/cvm/v1/:namespace/info/:name`

功能：旧版 CVM 详情入口，行为同 `/panel-api/v1/k3k/ckm/v1/:namespace/info/:name`。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：同 CKM 详情接口。

响应参数：同 CKM 详情接口。

### POST `/panel-api/v1/k3k/cvm/:namespace/action/:name/login`

功能：登录指定 CVM/CKM，返回对应 K3k 用户 token。K3k 子集群用户不能再次登录 CVM。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | CVM/CKM 命名空间 |
| `name` | path | 是 | string | CVM/CKM 名称 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | 目标用户 token |
| `expire` | int64 | token 过期时间，Unix 秒 |
| `isK3kUser` | bool | 是否 K3k 用户 |
| `refreshToken` | string | 刷新 token |

## 菜单与套餐

### GET `/panel-api/v1/menu`

功能：返回当前用户权限信息，数据来源同 `K3kUser.ToArray()`。

鉴权：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：`map[string]string`，字段同 `/panel-api/v1/k3k/info`。

### GET `/panel-api/v1/idc-list`

功能：获取可展示在商店中的 IDC/K3k 集群资源套餐列表。只返回未设置 `w7.cc/showInShop` 或该 label 为 `true` 的策略。

鉴权：无需用户 token。

请求参数：无。

响应参数：`types.Params`，即 `[]map[string]string`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | 套餐参数列表 |
| `[][key]` | string | 套餐参数键值，由 Cost 资源注解 `w7.cc/package-items` 解析得到 |
