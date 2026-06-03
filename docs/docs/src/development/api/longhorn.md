# Longhorn API

本文档说明 W7Panel Longhorn 相关 API，包括安装、卷状态聚合、待删除副本筛选，以及 attach、detach、扩容取消、文件系统 trim、快照删除/清理等卷操作。

## 整体使用方式

Longhorn API 用于处理 Kubernetes 存储层能力，主要围绕 Longhorn CRD 状态查询和 Longhorn Backend 卷操作。开发时先确认 Longhorn 是否已安装，再根据场景选择查询 CRD、安装内置 YAML，或通过 Backend action 操作卷。

### 基本流程

1. 使用用户 token 调用 Longhorn 接口。
2. 首次部署或缺少组件时，调用 `POST /panel-api/v1/longhorn/install` 安装内置 Longhorn YAML。
3. 页面展示 PVC/卷状态时，调用状态聚合接口读取 Longhorn Volume、Snapshot、Engine、Replica 等 CRD。
4. 执行 attach、detach、取消扩容、trim、快照删除或清理时，调用对应 volume action 接口。
5. 节点或磁盘迁移前，通过待删除副本接口筛选可清理 Replica。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 安装 Longhorn | `/panel-api/v1/longhorn/install` | 从 `KO_DATA_PATH` 内置 YAML 安装 |
| 查看卷状态 | `/panel-api/v1/longhorn/volumes/status` | 面板聚合 PVC 与 Longhorn 卷状态 |
| 筛选迁移副本 | `/panel-api/v1/longhorn/need-delete-replica` | 根据磁盘选择器和节点 ID 筛选 Replica |
| 挂载/卸载卷 | `/volumes/:volumeName/attach`、`/detach` | 调用 Longhorn Backend action |
| 扩容、trim、快照 | volume action 接口 | 直接操作指定 Longhorn 卷 |

### 使用边界

- Longhorn CRD 查询固定依赖 `longhorn-system` 命名空间。
- Backend action 依赖 `longhorn-backend.longhorn-system.svc:9500`，生产环境要确认服务可达。
- 安装接口依赖 `KO_DATA_PATH` 下的内置 YAML，修改安装包时需要同步部署文档。
- 卷操作属于高风险存储操作，调用前应确认卷名、状态和目标节点。

## 通用说明

### 鉴权

所有 Longhorn 接口都需要用户 token：

```http
Authorization: Bearer <user-token>
```

路由注册在 `w7panel-server/app/application/provider.go`，路径前缀为：

```text
/panel-api/v1/longhorn
```

### 后端依赖

| 依赖 | 说明 |
|------|------|
| Longhorn CRD | 状态查询读取 `longhorn.io/v1beta2` 下的 Volume、Snapshot、Engine、VolumeAttachment、Replica |
| Longhorn Backend | 卷操作通过 `http://longhorn-backend.longhorn-system.svc:9500/v1/volumes/{volumeName}?action={action}` 调用 |
| `KO_DATA_PATH` | 安装接口读取 `${KO_DATA_PATH}/yaml/longhorn/longhornfull.yaml` |
| `longhorn-system` | Longhorn client 固定访问该命名空间 |

### 响应格式

| 场景 | 响应 |
|------|------|
| 查询接口 | 直接返回业务对象或 Kubernetes CRD List |
| `install`、`attach` | 成功返回 `"success"` |
| `detach`、`cancel-expansion`、`trim-filesystem`、`snapshot-delete`、`snapshot-purge` | 成功返回 `null` |
| Longhorn Backend 非 200 | 返回服务端错误，错误内容为 Longhorn Backend 响应体 |

### 参数位置

查询接口主要使用 query/form 参数；卷操作接口的 `volumeName` 位于 path，action 所需参数按接口说明放在 JSON body 或空 body；安装接口可通过 query/form 传入可选 namespace。所有接口都需要通过 Header 携带用户 token。

## 能力概览

| 能力 | 说明 |
|------|------|
| 副本筛选 | 根据磁盘选择器和节点 ID 找出可清理 Longhorn Replica |
| 卷状态 | 聚合 PVC、Longhorn Volume、Snapshot、Engine、Replica 等状态 |
| 安装 | 从 `KO_DATA_PATH` 内置 YAML 安装 Longhorn |
| 卷挂载 | attach、detach 指定 Longhorn 卷 |
| 卷维护 | 取消扩容、trim 文件系统、删除和清理快照 |

## 查询接口

### GET `/panel-api/v1/longhorn/need-delete-replica`

功能：根据目标磁盘选择器和节点 ID，筛选可以删除的 Longhorn Replica。

典型用途：节点或磁盘迁移时，找出指定节点上属于某类磁盘选择器、且卷副本数大于 1 的副本。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `diskselector` | query/form | 是 | string | 磁盘选择器；当前实现按单个 selector 传入，内部包装为数组匹配 |
| `nodeid` | query/form | 是 | string | 节点 ID；多个用英文逗号分隔 |

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/longhorn/need-delete-replica?diskselector=disk-default&nodeid=node-1,node-2' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 使用当前用户 token 创建 K8s client |
| 2 | 创建 Longhorn client，读取 `longhorn-system` 命名空间下的 Volume 和 Replica |
| 3 | 按 Volume 聚合 Replica |
| 4 | 只处理 `volume.spec.diskSelector` 包含 `diskselector` 且 `volume.spec.numberOfReplicas > 1` 的卷 |
| 5 | 返回 `replica.spec.nodeID` 命中 `nodeid` 列表的 Replica |

响应参数：Kubernetes `longhorn.io/v1beta2 ReplicaList`。

常用字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | array | Replica 列表 |
| `items[].metadata.name` | string | Replica 名称 |
| `items[].metadata.namespace` | string | 通常为 `longhorn-system` |
| `items[].metadata.labels.longhornvolume` | string | 所属 Longhorn volume 名称 |
| `items[].spec.nodeID` | string | Replica 所在节点 ID |
| `items[].spec.diskID` | string | Replica 所在磁盘 ID |
| `items[].spec.volumeName` | string | 所属 Volume 名称，视 Longhorn 版本返回 |
| `items[].status` | object | Longhorn Replica 原生状态 |

响应示例：

```json
{
  "items": [
    {
      "metadata": {
        "name": "pvc-xxx-r-abc",
        "namespace": "longhorn-system",
        "labels": {
          "longhornvolume": "pvc-xxx"
        }
      },
      "spec": {
        "nodeID": "node-1",
        "diskID": "disk-default"
      },
      "status": {
        "currentState": "running"
      }
    }
  ]
}
```

### GET `/panel-api/v1/longhorn/volumes/status`

功能：获取 Longhorn Volume 状态，并按 PVC 维度聚合给前端使用。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `convertpvc` | query/form | 否 | string | 当前 controller 接收但未使用，兼容预留字段 |

请求示例：

```bash
curl 'http://localhost:8080/panel-api/v1/longhorn/volumes/status' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 使用 root SDK 创建 Longhorn client |
| 2 | 读取 Volume、Snapshot、Engine、VolumeAttachment 列表 |
| 3 | 跳过未绑定 PVC 的 Volume |
| 4 | 使用 key `&lt;pvcName&gt;:&lt;namespace&gt;` 聚合状态 |
| 5 | 计算快照大小、是否扩容中、是否存在 `longhorn-ui` lock、当前挂载节点 |

响应参数：对象 map，key 为 `&lt;pvcName&gt;:&lt;namespace&gt;`，value 为卷状态对象。

| 字段 | 类型 | 说明 |
|------|------|------|
| `numberOfReplicas` | int | `volume.spec.numberOfReplicas` |
| `robustness` | string | `volume.status.robustness`，如 `healthy`、`degraded` |
| `size` | int64/string | `volume.spec.size`；Go tag 为 `json:"size,string"`，JSON 中可能表现为字符串 |
| `actualSize` | int64 | `volume.status.actualSize`，实际占用容量 |
| `creationTimestamp` | string | Volume 创建时间，格式 `2006-01-02 15:04:05` |
| `accessMode` | string | `volume.spec.accessMode` |
| `snapShotSize` | int64 | 当前 Volume 快照总大小 |
| `isExpanding` | bool | 是否处于扩容中 |
| `expandErr` | string | 扩容错误信息 |
| `state` | string | `volume.status.state` |
| `volumeName` | string | Longhorn Volume 名称 |
| `isLock` | string | 是否存在 lock，字符串 `true`/`false` |
| `lockNodeId` | string | lock 指向的节点 ID |
| `attachedNodeId` | string | 当前挂载节点 ID |

响应示例：

```json
{
  "data-default:default": {
    "numberOfReplicas": 3,
    "robustness": "healthy",
    "size": "10737418240",
    "actualSize": 2147483648,
    "creationTimestamp": "2026-06-03 10:20:30",
    "accessMode": "rwo",
    "snapShotSize": 1048576,
    "isExpanding": false,
    "expandErr": "",
    "state": "attached",
    "volumeName": "pvc-xxx",
    "isLock": "false",
    "lockNodeId": "",
    "attachedNodeId": "node-1"
  }
}
```

实现注意：

| 项 | 说明 |
|----|------|
| root SDK | 当前接口使用 `k8s.NewK8sClient().Sdk`，不是用户 channel SDK |
| mock lock | `MOCK_LONGHORN_LOCK=true` 时会强制 `isLock=true`，仅用于测试 |
| 空结果 | 没有绑定 PVC 的 Longhorn Volume 不会出现在结果中 |

## 安装接口

### POST `/panel-api/v1/longhorn/install`

功能：使用内置 YAML 安装 Longhorn。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `namespace` | query | 否 | string | 安装命名空间；为空时使用当前用户 K8s client 的默认 namespace |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/install?namespace=longhorn-system' \
  -H 'Authorization: Bearer <user-token>'
```

处理逻辑：

| 步骤 | 说明 |
|------|------|
| 1 | 使用当前用户 token 创建 K8s channel client |
| 2 | 从 `${KO_DATA_PATH}/yaml/longhorn/longhornfull.yaml` 读取 YAML |
| 3 | 调用 `ApplyBytes` 应用到目标 namespace |

响应示例：

```json
"success"
```

错误场景：

| 场景 | 说明 |
|------|------|
| `KO_DATA_PATH` 未设置 | 返回错误 `找不到longhorn yaml` |
| YAML 文件不存在或不可读 | 返回文件读取错误 |
| Apply 失败 | 返回 K8s apply 错误 |

## 卷操作接口

卷操作通过 Longhorn Backend action API 实现，目标地址形态：

```text
POST http://longhorn-backend.longhorn-system.svc:9500/v1/volumes/{volumeName}?action={action}
```

后端期望 Longhorn Backend 返回 HTTP 200；非 200 会把响应体作为错误返回。

### POST `/panel-api/v1/longhorn/volumes/:volumeName/attach`

功能：挂载 Longhorn Volume 到指定节点。

请求类型：`application/json`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `volumeName` | path | 是 | string | - | Longhorn Volume 名称 |
| `hostId` | JSON | 是 | string | - | 目标节点 ID |
| `disableFrontend` | JSON | 否 | bool | `false` | 是否禁用 frontend |
| `AttachedBy` | JSON | 否 | string | `""` | 挂载发起方；注意字段首字母大写 |
| `AttachmentID` | JSON | 否 | string | `longhorn-ui` | Attachment ticket ID；注意字段首字母大写 |
| `attacherType` | JSON | 否 | string | `""` | attacher 类型 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/attach' \
  -H 'Authorization: Bearer <user-token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "hostId": "node-1",
    "disableFrontend": false,
    "AttachmentID": "longhorn-ui"
  }'
```

发送给 Longhorn Backend 的 action：

```json
{
  "hostId": "node-1",
  "disableFrontend": false,
  "AttachedBy": "",
  "attacherType": "",
  "AttachmentID": "longhorn-ui"
}
```

响应示例：

```json
"success"
```

### POST `/panel-api/v1/longhorn/volumes/:volumeName/detach`

功能：卸载 Longhorn Volume。

请求类型：`application/json`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `volumeName` | path | 是 | string | - | Longhorn Volume 名称 |
| `forceDetach` | JSON | 否 | bool | `false` | 是否强制卸载 |
| `attachmentID` | JSON | 否 | string | `longhorn-ui` | Attachment ticket ID |
| `hostId` | JSON | 否 | string | `""` | 当前 controller 接收但不传给底层方法 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/detach' \
  -H 'Authorization: Bearer <user-token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "forceDetach": true,
    "attachmentID": "longhorn-ui"
  }'
```

发送给 Longhorn Backend 的 action：

```json
{
  "forceDetach": true,
  "attachmentID": "longhorn-ui",
  "hostId": ""
}
```

响应示例：

```json
null
```

### POST `/panel-api/v1/longhorn/volumes/:volumeName/cancel-expansion`

功能：取消 Longhorn Volume 扩容。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn Volume 名称 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/cancel-expansion' \
  -H 'Authorization: Bearer <user-token>'
```

发送给 Longhorn Backend 的 action：

```json
{
  "name": "pvc-xxx"
}
```

响应示例：

```json
null
```

### POST `/panel-api/v1/longhorn/volumes/:volumeName/trim-filesystem`

功能：对 Longhorn Volume 执行文件系统 trim。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn Volume 名称 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/trim-filesystem' \
  -H 'Authorization: Bearer <user-token>'
```

发送给 Longhorn Backend 的 action：

```json
{
  "name": "pvc-xxx"
}
```

响应示例：

```json
null
```

### POST `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-delete`

功能：删除指定 Longhorn Snapshot。

请求类型：`application/json`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn Volume 名称 |
| `name` | JSON | 否 | string | Snapshot 名称；当前后端未设置 required，但 Longhorn Backend 通常需要有效名称 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/snapshot-delete' \
  -H 'Authorization: Bearer <user-token>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"snapshot-001"}'
```

发送给 Longhorn Backend 的 action：

```json
{
  "name": "snapshot-001"
}
```

响应示例：

```json
null
```

### POST `/panel-api/v1/longhorn/volumes/:volumeName/snapshot-purge`

功能：清理 Longhorn Volume 上可清理的快照数据。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `volumeName` | path | 是 | string | Longhorn Volume 名称 |

请求示例：

```bash
curl -X POST 'http://localhost:8080/panel-api/v1/longhorn/volumes/pvc-xxx/snapshot-purge' \
  -H 'Authorization: Bearer <user-token>'
```

发送给 Longhorn Backend 的 action：

```json
{
  "name": "pvc-xxx"
}
```

响应示例：

```json
null
```

## 前端和调用注意

| 注意项 | 说明 |
|--------|------|
| `AttachmentID` 大小写 | attach 使用 `AttachmentID`，detach 使用 `attachmentID`，字段大小写不同 |
| `size` 类型 | 卷状态 `size` 使用 `json:",string"`，前端应兼容字符串 |
| Longhorn Backend | 卷操作依赖集群内 `longhorn-backend.longhorn-system.svc:9500` 可访问 |
| 权限 | 查询和操作都需要面板用户 token，部分接口通过 `Proxy` 中间件转发时还依赖目标集群访问权限 |
| 安装 YAML | 安装前确认 `KO_DATA_PATH/yaml/longhorn/longhornfull.yaml` 存在 |

## 开发检查

- 修改 Longhorn 路由时同步更新 [cluster-ops.md](./cluster-ops.md) 中的 Longhorn 跳转说明和本文能力概览。
- 修改卷状态字段时同步检查前端 PVC/存储页面对 `size`、`isLock`、`attachedNodeId` 等字段的使用。
- 新增 Longhorn action 时，文档中要写清楚面板请求参数和实际发送给 Longhorn Backend 的 action body。
- 不要在日志或响应中输出不必要的集群内部错误堆栈，Longhorn Backend 响应体可用于定位但需要避免泄露敏感信息。
