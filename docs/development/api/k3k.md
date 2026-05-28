# w7panel-server/app/k3k API 文档

## 用户与集群

### GET `/panel-api/v1/k3k/info`

功能：获取当前 token 对应的 K3k 用户、集群和权限信息。

鉴权：需要用户 token。

请求参数：无。

响应参数：`K3kUser.ToArray()` 返回的 `map[string]string`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `w7.cc/user-mode` | string | 用户模式 |
| `k3k.io/name` | string | K3k 集群名称 |
| `k3k.io/namespace` | string | K3k 集群命名空间 |
| `k3k.io/job-name` | string | K3k 初始化 Job 名称 |
| `k3k.io/job-status` | string | K3k 初始化 Job 状态 |
| `k3k.io/cluster-mode` | string | 集群模式 |
| `k3k.io/cluster-policy` | string | 集群策略 |
| `w7.cc/username` | string | 用户名 |
| `w7.cc/debug` | string | 调试模式 |
| `w7.cc/menu` | string | 菜单权限 |
| `w7.cc/quota-limit` | string | 配额限制 |
| `w7.cc/file-editor` | string | 文件编辑权限 |
| `w7.cc/web-shell` | string | Web Shell 权限 |
| `w7.cc/domain-white-list` | string | 域名白名单 |
| `w7.cc/demo-user` | string | 是否演示用户 |
| `w7.cc/sys-storage-pvc-name` | string | 系统存储 PVC 名称 |
| `w7.cc/cost` | string | 费用配置 |
| `w7.cc/can-init-cluster` | string | 是否可初始化集群 |
| `w7.cc/need-create-order` | string | 是否需要创建初始订单 |
| `w7.cc/need-over-check` | string | 是否必须执行超卖检查 |
| `w7.cc/can-over-check` | string | 是否可执行超卖检查 |
| `w7.cc/has-over-resource` | string | 是否资源超额 |
| `w7.cc/has-password` | string | 是否设置密码 |
| `w7.cc/unit-price-total` | string | 续费单价 |
| `w7.cc/can-renew` | string | 是否可续费 |
| `w7.cc/need-renew` | string | 是否需要续费 |
| `w7.cc/can-expand` | string | 是否可扩容 |
| `w7.cc/over-mode` | string | 资源超卖模式 |
| `k3k.io/expire-time` | string | 到期时间 |
| `k3k.io/cluster-status` | string | 集群状态 |
| `w7.cc/diff-month` | string | 剩余或差异月份 |
| `w7.cc/role` | string | 用户角色 |
| `w7.cc/wh-mode` | string | 维护模式 |
| `w7.cc/wh-job` | string | 维护 Job |
| `w7.cc/wh-job-status` | string | 维护 Job 状态 |
| `w7.cc/server-pod-name` | string | K3k server Pod 名称 |
| `w7.cc/server-container-name` | string | K3k server 容器名称 |

### GET `/panel-api/v1/userinfo`

功能：兼容入口，行为同 `/panel-api/v1/k3k/info`。

鉴权：需要用户 token。

请求参数：无。

响应参数：同 `/panel-api/v1/k3k/info`。

### POST `/panel-api/v1/k3k/init`

功能：初始化当前 K3k 用户对应的集群。

鉴权：需要用户 token。

请求参数：无。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/init-cluster`

功能：创始人用户为指定 K3k 用户重新初始化集群。会清理 SDK 缓存并根据目标 ServiceAccount 初始化集群。

鉴权：需要 founder 用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `k3kUserName` | form | 是 | string | 目标 K3k 用户名，即 ServiceAccount 名称 |

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/login`

功能：创始人或 `w7panel` 用户切换登录到指定 K3k 用户，返回该用户 token。

鉴权：需要 founder 或 `w7panel` 用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `k3kUserName` | form | 是 | string | 目标 K3k 用户名 |

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | 目标用户 token |
| `expire` | int64 | token 过期时间，Unix 秒 |
| `isK3kUser` | bool | 是否 K3k 用户 |
| `refreshToken` | string | 刷新 token |

### POST `/panel-api/v1/k3k/wh`

功能：切换当前集群用户的维护模式。非集群用户直接返回成功。

鉴权：需要用户 token。

请求参数：无。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/whjob`

功能：重新创建当前 K3k 用户的维护/救援模式 Job。

鉴权：需要用户 token。

请求参数：无。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/storage/resize`

功能：扩容当前 K3k 用户系统存储。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `size` | form | 否 | int | 目标容量，单位 Gi |

响应参数：`"success"`。

### GET `/panel-api/v1/idc-list`

功能：获取可展示在商店中的 IDC/K3k 集群资源套餐列表。只返回 `w7.cc/showInShop=true` 或未设置该 label 的策略。

鉴权：无需用户 token。

请求参数：无。

响应参数：`types.Params`，即 `[]map[string]string`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `[]` | array | 套餐参数列表 |
| `[][key]` | string | 套餐参数键值，由 `K3kClusterPolicy.ToPackageItemsParams(true)` 生成 |

## 同步接口

同步接口用于在主集群和 K3k 子集群之间同步资源。以下接口均无需用户 token。

通用请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `virtualName` | form | 否 | string | 子集群内资源名称 |
| `virtualNamespace` | form | 否 | string | 子集群内资源命名空间 |
| `k3kName` | form | 否 | string | K3k 集群名称 |
| `k3kNamespace` | form | 否 | string | K3k 集群命名空间 |
| `k3kMode` | form | 否 | string | K3k 模式，如 `virtual` |
| `version` | form | 否 | string | 版本号 |

### POST `/panel-api/v1/k3k/sync-ingress`

功能：同步 Ingress。

请求参数：同步通用参数。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/sync-configmap`

功能：同步 ConfigMap。`k3kMode=virtual` 且 `virtualName=registries` 时，代码中保留了重启集群 Pod 的逻辑，但当前被注释。

请求参数：同步通用参数。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/sync-mcpbridge`

功能：同步 MCP Bridge。

请求参数：同步通用参数。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/sync-secret`

功能：同步 Secret。

请求参数：同步通用参数。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/sync-down-static`

功能：触发子集群静态资源下载。

请求参数：同步通用参数，其中主要使用：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `virtualNamespace` | form | 否 | string | AppGroup 命名空间 |
| `virtualName` | form | 否 | string | AppGroup 名称 |

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/sync-microapp`

功能：将 MicroApp 同步到子集群。

请求参数：同步通用参数，其中主要使用：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `k3kName` | form | 否 | string | K3k 集群名称 |
| `k3kNamespace` | form | 否 | string | K3k 集群命名空间 |

响应参数：`"success"`。

## 订单接口

### GET `/panel-api/v1/k3k/order/config`

功能：获取订单配置。当前 Controller 方法为空，没有显式响应。

鉴权：需要用户 token。

请求参数：无。

响应参数：当前无显式响应。

### GET `/panel-api/v1/k3k/order/price`

功能：计算基础购买、续费、扩容三种购买模式价格。

鉴权：需要集群用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `cpu` | form/query | 否 | int64 | CPU 资源 |
| `storage` | form/query | 否 | int64 | 存储资源，单位 Gi |
| `memory` | form/query | 否 | int64 | 内存资源，单位 Gi |
| `bandwidth` | form/query | 否 | int64 | 带宽资源，单位 M |
| `quantity` | form/query | 否 | int64 | 购买数量 |
| `unit` | form/query | 否 | string | 时间单位，如 `hour`、`month` |

响应参数：map，key 为购买模式。

| 字段 | 类型 | 说明 |
|------|------|------|
| `base.price` | decimal | 基础购买折后价 |
| `base.originPrice` | decimal | 基础购买原价 |
| `base.discount` | int64 | 折扣 |
| `base.buyMode` | string | 购买模式 |
| `renew.price` | decimal | 续费折后价 |
| `renew.originPrice` | decimal | 续费原价 |
| `renew.discount` | int64 | 折扣 |
| `renew.buyMode` | string | 购买模式 |
| `expand.price` | decimal | 扩容价格 |
| `expand.originPrice` | decimal | 扩容原价 |
| `expand.discount` | int64 | 折扣 |
| `expand.buyMode` | string | 购买模式 |

### POST `/panel-api/v1/k3k/order/base`

功能：创建基础资源购买订单。

鉴权：需要集群用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `cpu` | form | 否 | int64 | CPU 资源 |
| `storage` | form | 否 | int64 | 存储资源，单位 Gi |
| `memory` | form | 否 | int64 | 内存资源，单位 Gi |
| `bandwidth` | form | 否 | int64 | 带宽资源，单位 M |
| `quantity` | form | 否 | int64 | 购买数量 |
| `unit` | form | 否 | string | 时间单位 |
| `couponCode` | form | 否 | string | 优惠码 |

响应参数：支付结果对象，由 `order.CreateBaseResourceOrder` 返回。

### POST `/panel-api/v1/k3k/order/renew`

功能：创建续费订单。

鉴权：需要集群用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `quantity` | form | 否 | int64 | 续费数量 |
| `unit` | form | 否 | string | 时间单位 |
| `couponCode` | form | 否 | string | 优惠码 |

响应参数：支付结果对象，由 `order.CreateRenewOrder` 返回。

### POST `/panel-api/v1/k3k/order/expand`

功能：创建扩容订单。至少需要提交一项资源。

鉴权：需要集群用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `cpu` | form | 否 | int64 | 增加的 CPU |
| `storage` | form | 否 | int64 | 增加的存储，单位 Gi |
| `memory` | form | 否 | int64 | 增加的内存，单位 Gi |
| `bandwidth` | form | 否 | int64 | 增加的带宽，单位 M |

响应参数：支付结果对象，由 `order.CreateExpandOrder` 返回。

### POST `/panel-api/v1/k3k/order/license`

功能：创建面板授权购买订单。

鉴权：需要用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `productId` | form | 是 | string | Console 产品 ID |

响应参数：Console 默认产品订单结果。

### POST `/panel-api/v1/k3k/order/notify`

功能：订单支付回调，根据订单号通知并更新 K3k 用户订单状态。

鉴权：无需用户 token。

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `k3kName` | form | 是 | string | K3k 用户名 |
| `orderSn` | form | 是 | string | 订单号 |

响应参数：`"success"`。

### POST `/k8s/k3k/order/notify`

功能：旧版订单支付回调，行为同 `/panel-api/v1/k3k/order/notify`。

鉴权：无需用户 token。

请求参数：同 `/panel-api/v1/k3k/order/notify`。

响应参数：`"success"`。

### POST `/panel-api/v1/k3k/order/refresh`

功能：当前集群用户主动刷新订单支付状态。

鉴权：需要集群用户 token。

请求参数：无。

响应参数：`"success"`。

## 超卖资源接口

### GET `/panel-api/v1/k3k/overselling/config`

功能：获取资源超卖配置。读取失败时返回默认配置。

鉴权：需要用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu` | int32 | CPU 超卖百分比，默认 100 |
| `memory` | int32 | 内存超卖百分比，默认 100 |
| `storage` | int32 | 存储超卖百分比，默认 100 |
| `bandwidth` | int32 | 带宽超卖配置，默认 1000 |
| `bandwidthNum` | int32 | 带宽数量配置，默认 100 |

### GET `/panel-api/v1/k3k/overselling/current-resource`

功能：获取按超卖配置计算后的当前可用资源。

鉴权：需要用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu` | int64 | CPU |
| `memory` | int64 | 内存，单位 Gi |
| `storage` | int64 | 存储，单位 Gi |
| `bandwidth` | int64 | 带宽，单位 M |

### POST `/panel-api/v1/k3k/overselling/check`

功能：检查当前集群用户所需资源是否超过集群超卖后可用资源。

鉴权：需要集群用户 token。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `pass` | bool | 是否通过资源检查 |
