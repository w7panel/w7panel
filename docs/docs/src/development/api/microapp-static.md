# 微应用与静态资源 API

本文档说明面板内微应用信息、微应用后端代理、前端静态资源下载和静态资源回源代理相关 API。该能力服务于 Wujie 微前端容器、ZPK 应用前端包、应用组弹窗和微应用联调。

## 整体使用方式

微应用与静态资源接口用于让面板发现微应用、代理微应用后端、管理前端静态包下载状态，并在本地缺少资源时回源到制品库。开发时先查询微应用信息，再根据 frontendUrl/backendUrl 决定加载前端、代理后端或触发静态资源下载。

### 基本流程

1. 使用用户 token 查询微应用列表或详情，获取 frontendUrl、backendUrl、角色配置和绑定信息。
2. 面板通过 Wujie 加载微应用前端，并通过 props 注入 paneltoken、用户信息、Console token 等上下文。
3. 微应用调用自身后端时，走 `/panel-api/v1/microapp/:name/proxy/*path` 代理。
4. 对 ZPK 前端静态包，先查 `/panel-api/v1/static/:identifie/status`，本地未就绪时触发下载或使用回源代理。
5. 静态下载后用状态接口轮询，不要假设下载接口同步返回完整结果。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 顶部微应用列表 | `/panel-api/v1/microapp/top` | 按当前角色过滤 |
| 微应用详情 | `/panel-api/v1/microapp/:name/info` | 获取前后端 URL、角色配置和绑定 |
| 后端代理 | `/panel-api/v1/microapp/:name/proxy/*path` | 透传微应用后端响应 |
| 静态资源状态 | `/panel-api/v1/static/:identifie/status` | 查询本地资源和回源地址 |
| 触发静态下载 | `/panel-api/v1/static/:namespace/download/:name` | 异步下载前端包 |
| 静态回源代理 | `/panel-api/v1/static/proxy/*` | 本地未下载时代理远端静态资源 |

### 使用边界

- 微应用 props 中不同 token 含义不同，`paneltoken`、`w7PanelToken`、`access_token` 不要混用。
- 微应用后端代理和静态回源代理都可能透传非 JSON 响应。
- 静态下载是异步行为，前端应轮询状态接口。
- 修改微应用字段时需要同步检查 Wujie 事件和前端 props 使用。

## 通用说明

### 鉴权

本文档中的接口都需要用户 token：

```http
Authorization: Bearer <user-token>
```

后端通过 `middleware.Auth` 校验 token，并在请求上下文注入：

| 上下文字段 | 类型 | 说明 |
|------------|------|------|
| `k8s_token` | string | 当前用户的 K8s token，微应用查询、代理和静态下载都依赖该字段 |

### 响应格式

| 类型 | 响应形态 | 说明 |
|------|----------|------|
| 微应用查询 | 直接返回 Kubernetes CRD 对象或列表 | `MicroApp`、`MicroAppList` |
| 静态状态 | 直接返回对象 | 包含 `status`、`proxyUrl`、`zpkUrl`、`ticket` |
| 代理接口 | 透传 | 透传微应用后端或远程静态资源响应 |
| 静态下载 | 当前无显式响应体 | 调用后用状态接口轮询 |

缺少 token 或 token 无效时返回 401：

```json
{
  "code": 401,
  "msg": "请登录"
}
```

业务错误通常通过 `JsonResponseWithServerError` 返回，具体结构取决于框架封装。静态回源代理失败时会直接返回 502：

```json
{
  "code": 502,
  "error": "proxy error message"
}
```

### 参数位置

微应用名称、静态资源标识、namespace、name、zpkUrl、version 和静态文件路径主要位于 path。微应用后端代理和静态回源代理会透传 query、Header 和 body；列表、详情、状态接口通常无 body；触发下载接口通过 path 指定 AppGroup。

## 能力概览

| 能力 | 接口 | 说明 |
|------|------|------|
| 微应用列表 | `GET /panel-api/v1/microapp/top` | 获取当前角色可见的顶部微应用列表 |
| 微应用详情 | `GET /panel-api/v1/microapp/:name/info` | 获取指定微应用的 frontendUrl、backendUrl、roleConfig、bindings 等 |
| 微应用后端代理 | `ANY /panel-api/v1/microapp/:name/proxy/*path` | 面板代理请求微应用后端服务 |
| 静态资源状态 | `GET /panel-api/v1/static/:identifie/status` | 查询前端包下载状态，并尝试生成远程回源地址 |
| 静态资源下载 | `POST /panel-api/v1/static/:namespace/download/:name` | 触发指定 AppGroup 的前端包下载 |
| 静态资源回源代理 | `GET /panel-api/v1/static/proxy/:zpkUrl/:identifie/:version/frontend/*path` | 本地未下载时从制品库回源静态文件 |

## 微应用接口

### GET `/panel-api/v1/microapp/top`

功能：获取当前用户角色可见的顶部微应用列表。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/microapp/top' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 从 token 解析当前角色 `role` |
| 2 | 从 root 集群读取所有 `MicroApp` CRD |
| 3 | 只保留 `spec.config-v2.props.roleConfig` 中包含当前角色的多角色微应用 |
| 4 | 非 `founder` 角色会过滤 `bindings` 和 `roleConfig`，并标记 `metadata.labels.microapp.w7.cc/from=root` |
| 5 | 若读取失败，controller 返回空 `MicroAppList`，不会直接报错 |

响应参数：`MicroAppList`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata` | object | Kubernetes List 元信息 |
| `items` | array | MicroApp 列表 |
| `items[].apiVersion` | string | CRD API 版本，通常为 `microapp.w7.cc/v1alpha1` |
| `items[].kind` | string | 资源类型，通常为 `MicroApp` |
| `items[].metadata.name` | string | MicroApp 名称；非 founder 从 root 集群过滤时可能带 `-root` |
| `items[].metadata.namespace` | string | 命名空间，当前读取固定为 `default` |
| `items[].metadata.labels` | object | 标签，常见 `w7.cc/identifie`、`w7.cc/version`、`microapp.w7.cc/from` |
| `items[].spec.framework` | string | 微应用框架类型 |
| `items[].spec.frontendUrl` | string | 前端入口地址 |
| `items[].spec.backendUrl` | string | 后端入口或兼容字段 |
| `items[].spec.proxyUrl` | string | 代理地址，视 CRD 配置返回 |
| `items[].spec.title` | string | 微应用标题 |
| `items[].spec.logo` | string | 图标地址 |
| `items[].spec.version` | string | 微应用版本 |
| `items[].spec.config.props` | object | 旧版 props，`map[string]string` |
| `items[].spec.config-v2.props.roleConfig` | object | 按角色划分的前端和代理配置 |
| `items[].spec.bindings` | array | 角色入口、菜单和展示配置 |

响应示例：

```json
{
  "metadata": {
    "resourceVersion": "12345"
  },
  "items": [
    {
      "apiVersion": "microapp.w7.cc/v1alpha1",
      "kind": "MicroApp",
      "metadata": {
        "name": "demo-root",
        "namespace": "default",
        "labels": {
          "w7.cc/identifie": "demo",
          "w7.cc/version": "1.0.0",
          "microapp.w7.cc/from": "root"
        }
      },
      "spec": {
        "framework": "wujie",
        "frontendUrl": "/ui/microapp/demo/index.html",
        "backendUrl": "/panel-api/v1/microapp/demo-root/proxy",
        "title": "Demo",
        "version": "1.0.0",
        "config-v2": {
          "props": {
            "roleConfig": {
              "normal": {
                "load_mode": "iframe",
                "serverUrl": "http://demo.default.svc:8080",
                "frontend_props": {
                  "group": "demo"
                },
                "proxy_request": {
                  "headers": {
                    "X-User": "${system.userid}"
                  },
                  "query": {
                    "role": "${system.role}"
                  }
                }
              }
            }
          }
        },
        "bindings": [
          {
            "name": "normal",
            "title": "Demo",
            "support": "thirdparty_cd",
            "status": 1,
            "menu": [
              {
                "title": "首页",
                "do": "index",
                "location": "top"
              }
            ]
          }
        ]
      }
    }
  ]
}
```

### GET `/panel-api/v1/microapp/:name/info`

功能：获取单个微应用详情。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | MicroApp 名称；带 `-root` 后缀时强制从 root 集群读取原始名称 |

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/microapp/demo-root/info' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 场景 | 行为 |
|------|------|
| `name` 以 `-root` 结尾 | 去掉 `-root` 后从 root 集群读取，并按当前角色过滤 |
| 当前用户集群可读 | 从当前用户集群读取 MicroApp |
| 当前用户集群 NotFound/Forbidden | 回退到 root 集群读取，并按当前角色过滤 |
| 当前角色为空 | 返回错误 `role is empty` |
| 找不到 MicroApp | 返回服务端错误 |

响应参数：`MicroApp` 对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `metadata.name` | string | MicroApp 名称，可能被过滤逻辑追加 `-root` |
| `metadata.namespace` | string | 命名空间 |
| `metadata.labels.microapp.w7.cc/from` | string | 值为 `root` 时表示来自 root 集群过滤结果 |
| `spec.framework` | string | 微应用框架 |
| `spec.frontendUrl` | string | 前端入口 |
| `spec.backendUrl` | string | 后端入口或兼容字段 |
| `spec.proxyUrl` | string | 代理 URL 配置 |
| `spec.title` | string | 标题 |
| `spec.logo` | string | 图标 |
| `spec.description` | string | 描述 |
| `spec.version` | string | 版本 |
| `spec.config.props` | object | 旧版 props |
| `spec.config-v2.props.roleConfig.{role}.load_mode` | string | 当前角色加载模式 |
| `spec.config-v2.props.roleConfig.{role}.serverUrl` | string | 微应用后端代理目标地址 |
| `spec.config-v2.props.roleConfig.{role}.frontend_props` | object | 注入前端的角色 props |
| `spec.config-v2.props.roleConfig.{role}.proxy_request.headers` | object | 代理时附加或覆盖的请求头 |
| `spec.config-v2.props.roleConfig.{role}.proxy_request.query` | object | 代理时附加的 query |
| `spec.bindings[]` | array | 角色入口和菜单配置 |

响应示例：

```json
{
  "apiVersion": "microapp.w7.cc/v1alpha1",
  "kind": "MicroApp",
  "metadata": {
    "name": "demo",
    "namespace": "default",
    "labels": {
      "w7.cc/identifie": "demo",
      "w7.cc/version": "1.0.0"
    }
  },
  "spec": {
    "framework": "wujie",
    "frontendUrl": "/ui/microapp/demo/index.html",
    "backendUrl": "/panel-api/v1/microapp/demo/proxy",
    "title": "Demo",
    "logo": "/assets/demo.png",
    "version": "1.0.0",
    "config": {
      "props": {
        "group": "demo"
      }
    },
    "config-v2": {
      "props": {
        "roleConfig": {
          "founder": {
            "load_mode": "wujie",
            "serverUrl": "http://demo.default.svc:8080",
            "frontend_props": {
              "group": "demo"
            },
            "proxy_request": {
              "headers": {
                "X-Panel-User": "${system.userid}"
              },
              "query": {
                "role": "${system.role}"
              }
            }
          }
        }
      }
    },
    "bindings": [
      {
        "name": "founder",
        "title": "Demo",
        "support": "thirdparty_cd",
        "status": 1,
        "menu": [
          {
            "displayorder": 0,
            "do": "index",
            "icon": "apps",
            "is_default": 1,
            "location": "top",
            "title": "首页"
          }
        ]
      }
    ]
  }
}
```

### ANY `/panel-api/v1/microapp/:name/proxy/*path`

功能：代理请求微应用后端服务。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `name` | path | 是 | string | MicroApp 名称 |
| `path` | path | 是 | string | 代理到微应用后端的路径，包含前导 `/` |
| query | query | 否 | any | 原始 query 会参与透传，并叠加 roleConfig 中的 `proxy_request.query` |
| body | body | 否 | any | 原始 body 透传 |

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/microapp/demo/proxy/api/health?debug=1' \
  -H 'Authorization: Bearer <user-token>'
```

代理目标选择：

| 场景 | 行为 |
|------|------|
| MicroApp 来自 root 或当前 token 不是 K3k 集群 token | 使用 `spec.config-v2.props.roleConfig[role].serverUrl` 作为代理目标 |
| K3k 虚拟模式 | role 强制设置为 `founder` 后再读取 roleConfig |
| K3k 集群用户访问非 root 微应用 | 先代理到子集群 Agent，再由子集群处理同一路径 |
| 当前 role 找不到 roleConfig | 集群用户可能回退 `founder`，仍找不到则返回 `not found proxy role config` |

`proxy_request` 变量替换：

| 变量 | 替换值 |
|------|--------|
| `${system.group}` | MicroApp 名称 |
| `${system.userid}` | 当前 K3k 用户名 |
| `${system.openid}` | 当前 K3k 用户名 |
| `${system.nickname}` | 当前用户昵称 |
| `${system.role}` | 当前角色 |
| `${system.access_token}` | OIDC 默认 access token；只有配置中包含该变量时才生成 |

响应参数：透传微应用后端响应，包含状态码、响应头和响应体。面板可能会附加/覆盖请求头和 query，但不改变微应用业务响应结构。

## 静态资源接口

静态资源下载依赖以下运行时配置：

| 配置 | 默认值 | 说明 |
|------|--------|------|
| `STATIC_DOWN_ENABLED` | `false` | 只有值为 `true` 时，`POST /static/:namespace/download/:name` 才会执行下载 |
| `MICROAPP_PATH` | `./kodata/microapp` | 静态资源解压目录 |

状态值：

| 值 | 说明 |
|----|------|
| `no_download` | 未下载、下载失败或缓存不存在 |
| `downloading` | 正在下载或解压 |
| `download_success` | 下载并解压成功 |

状态只存放在当前进程内存缓存中，有效期 24 小时；服务重启后会回到 `no_download`。

### GET `/panel-api/v1/static/:identifie/status`

功能：查询指定应用前端静态资源下载状态，并在传入 `releaseName` 时尝试生成远程回源信息。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `identifie` | path | 是 | string | 应用标识；内部会把 `_` 替换为 `-` |
| `version` | query | 否 | string | 应用版本；非空时状态缓存 key 为 `identifie + version` |
| `releaseName` | query | 否 | string | AppGroup 名称；包含 `-root` 时会去掉该后缀 |

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/static/demo/status?version=1.0.0&releaseName=demo' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 条件 | 行为 |
|------|------|
| `version` 非空 | 调用 `DownStaticStatus(identifie, version, releaseName)` 查询 `static-download-{identifie}{version}` |
| `version` 为空 | 查询 `static-download-{releaseName}` |
| `releaseName` 非空且 AppGroup 存在 | 从 root 集群 `default` 命名空间读取 AppGroup |
| AppGroup 有 `spec.zpkUrl` | 对 `zpkUrl` 做 base64url 编码并生成 `proxyUrl` |
| AppGroup annotation 有 `w7.cc/ticket` | 返回 `ticket`，并缓存到 `frontend-ticket-{identifie}` 供回源代理使用，缓存 2 小时 |
| AppGroup 不存在或读取失败 | 仍返回状态，但 `zpkUrl`、`ticket`、`proxyUrl` 为空 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | `no_download`、`downloading`、`download_success` |
| `proxyUrl` | string | 静态资源回源代理基础路径，只有 `releaseName` 对应 AppGroup 且 `zpkUrl` 非空时返回 |
| `zpkUrl` | string | AppGroup `spec.zpkUrl` |
| `ticket` | string | AppGroup annotation `w7.cc/ticket` |

响应示例：

```json
{
  "status": "no_download",
  "proxyUrl": "/panel-api/v1/static/aHR0cHM6Ly96cGsuZXhhbXBsZS5jb20/demo/1.0.0/frontend/",
  "zpkUrl": "https://zpk.example.com",
  "ticket": "ticket-value"
}
```

实现注意：当前路由实际注册为 `/panel-api/v1/static/proxy/:zpkUrl/:identifie/:version/frontend/*path`，但 `StaticInfo` 生成的 `proxyUrl` 当前少了 `/proxy` 前缀。调用方使用 `proxyUrl` 前应按实际环境验证；正确回源路径应为：

```text
/panel-api/v1/static/proxy/{base64url(zpkUrl)}/{identifie}/{version}/frontend/{path}
```

### POST `/panel-api/v1/static/:namespace/download/:name`

功能：触发指定 AppGroup 的前端静态资源下载与解压。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | path | 是 | string | AppGroup 所在命名空间 |
| `name` | path | 是 | string | AppGroup 名称；包含 `-root` 时会去掉该后缀并使用 root SDK |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/static/default/download/demo' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 使用当前用户 token 创建 K8s client |
| 2 | 如果 `name` 包含 `-root`，去掉后缀并使用 root SDK |
| 3 | 优先从指定 namespace 和当前用户集群读取 AppGroup |
| 4 | NotFound 时回退 root 集群读取 |
| 5 | 调用 `appgroup.DownStatic(appgroupObj)` |
| 6 | `STATIC_DOWN_ENABLED != true` 时只记录日志，不下载 |
| 7 | AppGroup annotation `w7.cc/front-type` 包含 `thirdparty_cd` 时才下载 |

下载数据来源：

| 字段 | 说明 |
|------|------|
| `appgroup.spec.zpkUrl` | ZPK 信息接口地址 |
| `appgroup.spec.version` | 作为 `cur_version` query 参数传给 ZPK 信息接口 |
| `ZPK info.data.webzip_url` | 前端 zip 包下载地址映射 |
| `MICROAPP_PATH/{releaseName}` | 解压兼容目录 |
| `MICROAPP_PATH/{identifie}/{version}` | 解压版本目录 |

响应参数：当前 controller 调用 `DownStatic` 后没有显式写入响应体。HTTP 状态通常为 200，调用方应通过 `GET /panel-api/v1/static/:identifie/status` 轮询确认状态。

下载失败时，状态会回到 `no_download`；下载成功时状态为 `download_success`。

### GET `/panel-api/v1/static/proxy/:zpkUrl/:identifie/:version/frontend/*path`

功能：当前端静态资源本地不存在或尚未下载完成时，从远程 ZPK 制品库回源代理静态文件。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `zpkUrl` | path | 是 | string | base64url 编码后的 ZPK 制品库基础地址，使用 `base64.RawURLEncoding` |
| `identifie` | path | 是 | string | 应用标识 |
| `version` | path | 是 | string | 应用版本 |
| `path` | path | 否 | string | 静态资源路径；为空时按 `/` 处理 |

请求示例：

```bash
# zpkUrl=https://zpk.example.com 的 base64url 值为 aHR0cHM6Ly96cGsuZXhhbXBsZS5jb20
curl 'http://localhost:8080/panel-api/v1/static/proxy/aHR0cHM6Ly96cGsuZXhhbXBsZS5jb20/demo/1.0.0/frontend/index.html' \
  -H 'Authorization: Bearer <user-token>'
```

远程请求：

```text
{zpkUrl}/zpk/respo/attach/frontend/{identifie}/{version}/{path}?ticket={ticket}
```

`ticket` 来源于 `GET /static/:identifie/status` 缓存的 `frontend-ticket-{identifie}`。如果未先调用状态接口或缓存已过期，则不附加 `ticket`。

响应参数：透传远程静态文件响应。

| 响应 | 说明 |
|------|------|
| 200 + 文件内容 | 远程制品库返回成功 |
| 500/框架错误 | `zpkUrl` 解码失败、必填参数缺失、远程 URL 解析失败 |
| 400 `{"error":"zpkUrl is empty"}` | 解码后 `zpkUrl` 为空 |
| 502 `{"code":502,"error":"..."}` | 远程代理失败 |

实现细节：

| 项 | 说明 |
|----|------|
| CORS | 代理响应会删除远程 `Access-Control-Allow-Origin` |
| Host | 请求 Host 会设置为远程制品库 Host |
| query | 只保留 `ticket` query，不透传浏览器原始 query |
| path | 保证以 `/` 开头后拼接到远程路径 |

## 典型流程

1. 前端使用用户 token 调用 `/panel-api/v1/microapp/top` 获取可见微应用列表。
2. 用户进入某个微应用时，调用 `/panel-api/v1/microapp/:name/info` 获取前端入口和 roleConfig。
3. 若微应用前端包需要下载，使用 `identifie`、`version`、`releaseName` 调用 `/panel-api/v1/static/:identifie/status`。
4. 状态为 `no_download` 时，调用 `/panel-api/v1/static/:namespace/download/:name` 触发下载。
5. 状态为 `downloading` 时保持 loading 并轮询。
6. 状态为 `download_success` 时使用 `frontendUrl` 启动 Wujie 应用。
7. 本地资源缺失时，可使用 `/panel-api/v1/static/proxy/:zpkUrl/:identifie/:version/frontend/*path` 回源。
8. 微应用后端请求统一走 `/panel-api/v1/microapp/:name/proxy/*path`，由面板注入 header/query 并代理到 `serverUrl`。

## Wujie props

常见 props 包括：

| 字段 | 类型 | 说明 |
|------|------|------|
| `url` | string | 面板代理请求微应用后端服务的地址，通常指向 `/panel-api/v1/microapp/:name/proxy` |
| `group` | string | 应用标识分组；多子应用时为主应用标识 |
| `paneltoken` | string | 面板用户 token |
| `w7PanelToken` | string | Console 第三方 CD token |
| `Authorization` | string | 微应用自身 Basic 认证 |
| `userid` | string/number | 用户 ID |
| `openid` | string | openid |
| `nickname` | string | 昵称 |
| `role` | string | `founder`、`super`、`normal`、`technician` |
| `access_token` | string | 面板登录用户自身维护的 access token，只能获取用户信息 |

详细事件见 [../frontend/wujie-events.md](../frontend/wujie-events.md)。凭据边界见 [credentials.md](credentials.md)。

## 前端关联

| 位置 | 说明 |
|------|------|
| `w7panel-ui/src/views/topapp/micro-container.vue` | 通用应用组微应用容器 |
| `w7panel-ui/src/views/app/apps/detail.vue` | 应用详情微应用容器 |
| `w7panel-ui/src/views/app/apps/micro.vue` | 应用详情内嵌微应用 |
| `w7panel-ui/src/views/app/store/store-zpk.vue` | 应用商店 ZPK 微应用 |
| `w7panel-ui/src/components/wujie-modals.vue` | 微应用调用面板能力的事件桥 |

## 开发检查

- 需要访问微应用后端时优先使用面板代理 `url`，避免浏览器跨域和内网地址暴露。
- 静态资源下载是异步过程，前端必须处理 `no_download`、`downloading`、`download_success` 三种状态。
- `STATIC_DOWN_ENABLED` 默认关闭，测试下载流程前要确认环境变量为 `true`。
- `MICROAPP_PATH` 需要可写，否则下载解压会失败并回到 `no_download`。
- 修改 roleConfig 合并或过滤逻辑时，同步检查 `ListTop`、`ListInfo`、微应用容器和本文档。
- props 中 token 含义不同，不要混用 `paneltoken`、`w7PanelToken`、`access_token`。
- 修改静态回源路径时，同步检查 `StaticInfo` 生成的 `proxyUrl` 和实际注册路由 `/static/proxy/:zpkUrl/...` 是否一致。
