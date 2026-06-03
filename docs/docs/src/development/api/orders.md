# 云主机订单与资源超卖 API

本文档是 [云主机 API](./k3k.md) 的子文档，说明当前实际注册的云主机订单通知、面板授权订单和资源超卖接口。云主机用户、CKM/CVM、同步和套餐接口见 [k3k.md](./k3k.md)。

## 整体使用方式

云主机订单接口主要服务于订单回调、面板授权购买和资源超卖信息展示。开发时先从 [k3k.md](./k3k.md) 确认当前用户和云主机上下文，再进入订单或资源超卖接口。

### 基本流程

1. 查询云主机用户信息，确认当前用户是否属于云主机用户以及是否具备相关权限。
2. 订单系统回调时使用订单通知接口；当前实现直接返回成功。
3. 购买面板授权时调用授权订单接口，并透传 Console 返回结果。
4. 展示或校验资源余量时调用资源超卖配置和当前可用资源接口。
5. 历史未注册的购买、续费、扩容接口不作为当前可调用 API。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| 订单支付回调 | `/panel-api/v1/k3k/order/notify` | 当前实现直接返回成功 |
| 旧版订单回调 | `/k8s/k3k/order/notify` | 兼容旧路径 |
| 面板授权购买 | `/panel-api/v1/k3k/order/license` | 创建 Console 默认产品订单 |
| 查询超卖配置 | `/panel-api/v1/k3k/overselling/config` | 读取或返回默认配置 |
| 查询当前资源 | `/panel-api/v1/k3k/overselling/current-resource` | 返回按超卖配置计算后的可用资源 |

### 使用边界

- 订单通知接口当前没有实际通知处理逻辑，调用方不要依赖其更新订单状态。
- 授权订单接口需要用户 token，并依赖 Console SDK。
- 资源超卖接口返回的是面板聚合值，不等同于 Kubernetes 原始 allocatable。
- 订单类接口属于云主机业务入口，不在 API 首页作为独立专题暴露。

## 通用说明

### 鉴权

| 接口类型 | 鉴权 |
|----------|------|
| 订单通知 | 无需用户 token |
| 授权订单 | `Authorization: Bearer <user-token>` |
| 资源超卖 | `Authorization: Bearer <user-token>` |

### 响应格式

订单通知成功返回 JSON 字符串 `"success"`。授权订单接口透传 Console 订单结果。资源超卖接口直接返回配置或资源聚合对象。

### 参数位置

订单通知当前实现未绑定参数。授权订单使用 form 参数；资源超卖查询接口不需要 body。

## 能力概览

| 能力 | 说明 |
|------|------|
| 订单通知 | 接收云主机订单支付回调，当前实现直接返回成功 |
| 授权订单 | 创建面板授权购买订单并透传 Console 返回结果 |
| 资源超卖配置 | 查询当前资源超卖配置 |
| 当前资源 | 按超卖配置计算并返回当前可用资源 |

## 订单接口

### POST `/panel-api/v1/k3k/order/notify`

功能：订单支付回调。当前 Controller 中订单通知处理逻辑已注释，接口直接返回成功。

认证：无需用户 token。

请求参数：当前实现未绑定参数。

响应参数：

```json
"success"
```

### POST `/k8s/k3k/order/notify`

功能：旧版订单支付回调，行为同 `/panel-api/v1/k3k/order/notify`。

认证：无需用户 token。

请求参数：当前实现未绑定参数。

响应参数：

```json
"success"
```

### POST `/panel-api/v1/k3k/order/license`

功能：创建面板授权购买订单。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求类型：`application/x-www-form-urlencoded`

请求参数：

| 参数 | 位置 | 必填 | 类型 | 说明 |
|------|------|------|------|------|
| `productId` | form | 是 | string | Console 产品 ID |

响应参数：Console 默认产品订单结果，原样返回。

## 资源超卖接口

### GET `/panel-api/v1/k3k/overselling/config`

功能：获取资源超卖配置。读取失败时返回默认配置。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu` | int32 | CPU 超卖百分比，默认 100 |
| `memory` | int32 | 内存超卖百分比，默认 100 |
| `storage` | int32 | 存储超卖百分比，默认 100 |
| `bandwidth` | int32 | 带宽配置，默认 1000 |
| `bandwidthNum` | int32 | 带宽数量配置，默认 100 |

### GET `/panel-api/v1/k3k/overselling/current-resource`

功能：获取按超卖配置计算后的当前可用资源。读取失败时返回全 0。

认证：`Authorization: Bearer &lt;user-token&gt;`

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu` | int64 | CPU |
| `memory` | int64 | 内存，单位 Gi |
| `storage` | int64 | 存储，单位 Gi |
| `bandwidth` | int64 | 带宽，单位 Mi |
