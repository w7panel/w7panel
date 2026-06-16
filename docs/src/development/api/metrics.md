# 指标 API

本文档说明 `w7panel-server/app/metrics` 模块提供的资源使用量和 metrics 组件安装状态接口。指标接口主要服务于面板资源概览、云主机资源展示和运维判断。

## 整体使用方式

指标接口用于查询当前用户上下文下的 CPU、内存、磁盘使用量，以及 metrics 组件是否已安装。开发时先确认 metrics 组件状态，再调用资源用量接口展示面板聚合指标。

### 基本流程

1. 使用用户 token 调用 metrics 接口。
2. 需要判断能力是否可用时，先调用 `/panel-api/v1/metrics/installed`。
3. 展示 CPU/内存资源时调用 `/panel-api/v1/metrics/usage/normal`。
4. 展示磁盘资源时调用 `/panel-api/v1/metrics/usage/disk`。
5. 如果 metrics 不可用，前端应展示降级状态，不要把空值误认为真实资源为 0。

### 场景选择

| 场景 | 使用接口 | 说明 |
|------|----------|------|
| CPU/内存用量 | `/panel-api/v1/metrics/usage/normal` | 返回 usage 和 total |
| 磁盘用量 | `/panel-api/v1/metrics/usage/disk` | 返回 disk usage 和 total |
| metrics 安装状态 | `/panel-api/v1/metrics/installed` | 返回组件状态和访问地址 |

### 使用边界

- 指标数据是面板聚合结果，不等同于 Kubernetes Metrics API 的原始响应。
- 获取底层指标失败时，部分接口可能返回默认或局部结果，前端需要结合状态字段判断。
- 指标展示涉及当前用户或云主机上下文，不能跨用户直接复用。

## 通用说明

### 鉴权

metrics 接口均需要用户 token：

```http
Authorization: Bearer <user-token>
```

### 响应格式

接口直接返回指标或状态业务对象，不额外包裹统一 `data` 字段。metrics 组件不可用时，安装状态接口返回状态说明；查询类接口可能返回错误。

### 参数位置

当前 metrics 查询接口主要依赖当前 token 上下文，不需要额外 body；如后续增加 namespace、cluster 等参数，应明确标注 query/form/json 来源。

## 能力概览

| 能力 | 说明 |
|------|------|
| CPU/内存用量 | 查询当前集群或云主机可展示的 CPU、内存使用量和总量 |
| 磁盘用量 | 查询磁盘使用量和总量 |
| metrics 安装状态 | 检查 VictoriaMetrics/metrics 组件是否可用及访问地址 |
| 展示状态 | 返回前端展示 metrics 能力时需要的状态字段 |

## 资源使用量

### GET `/panel-api/v1/metrics/usage/normal`

功能：获取当前 K3k 用户对应集群的 CPU 和内存使用量。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `cpu` | object | CPU 使用情况 |
| `cpu.usage` | int64 | 已使用 CPU，单位为 millicore |
| `cpu.total` | int64 | CPU 总量，单位为 millicore |
| `memory` | object | 内存使用情况 |
| `memory.usage` | int64 | 已使用内存，单位为 byte |
| `memory.total` | int64 | 内存总量，单位为 byte |

响应示例：

```json
{
  "cpu": {
    "usage": 500,
    "total": 2000
  },
  "memory": {
    "usage": 1073741824,
    "total": 4294967296
  }
}
```

### GET `/panel-api/v1/metrics/usage/disk`

功能：获取当前 K3k 用户对应集群的磁盘使用量。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `disk` | object | 磁盘使用情况 |
| `disk.usage` | int64 | 已使用磁盘容量 |
| `disk.total` | int64 | 磁盘总容量 |

说明：当底层获取磁盘使用量失败时，接口仍返回 `disk.usage` 和 `disk.total`，值来自失败时已得到的默认或局部结果。

响应示例：

```json
{
  "disk": {
    "usage": 10737418240,
    "total": 53687091200
  }
}
```

## 安装状态

### GET `/panel-api/v1/metrics/installed`

功能：检查 metrics 组件是否已安装，并返回前端访问 metrics 服务的基础地址。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `installed` | bool | metrics Helm Release 是否存在 |
| `baseUrl` | string | metrics 服务代理基础 URL |
| `namespace` | string | metrics 组件所在命名空间 |

响应逻辑：

| 场景 | `releaseName` | `namespace` | `baseUrl` |
|------|---------------|-------------|-----------|
| 主集群 | `vm-operator` | `w7-system` | `/k8s-proxy/v1/namespaces/w7-system/services/vmsingle-vm-operator-k8s-offline-metrics-single:8429/proxy/` |
| K3k virtual | `w7panel-metrics` | `default` | `/api/v1/namespaces/default/services/vmsingle-w7panel-metrics-k8s-offline-metrics-single:8429/proxy/` |

响应示例：

```json
{
  "installed": true,
  "baseUrl": "/k8s-proxy/v1/namespaces/w7-system/services/vmsingle-vm-operator-k8s-offline-metrics-single:8429/proxy/",
  "namespace": "w7-system"
}
```

## 展示状态

### GET `/panel-api/v1/metrics/state`

功能：获取仪表盘、节点和 Pod 指标是否可展示，以及是否需要安装 metrics 组件。

请求参数：无。

响应参数：

| 字段 | 类型 | 说明 |
|------|------|------|
| `canShowClusterMetrics` | bool | 是否可展示集群指标 |
| `canShowNodeMetrics` | bool | 是否可展示节点指标 |
| `canShowPodMetrics` | bool | 是否可展示 Pod 指标 |
| `needInstallMetricsInDashboard` | bool | 仪表盘是否需要提示安装 metrics |
| `needInstallMetricsInApp` | bool | 应用详情是否需要提示安装 metrics |

响应逻辑：

| 场景 | 说明 |
|------|------|
| 主集群 token | 检查 `default` 命名空间 `w7panel-metrics` Helm Release；存在时集群、节点、Pod 指标都可展示 |
| K3k token | 主集群安装后允许展示集群指标；再检查子集群 `default/w7panel-metrics`，存在时可展示 Pod 指标 |

响应示例：

```json
{
  "canShowClusterMetrics": true,
  "canShowNodeMetrics": true,
  "canShowPodMetrics": true,
  "needInstallMetricsInDashboard": false,
  "needInstallMetricsInApp": false
}
```

## 未注册 Handler 说明

`metrics.go` 中还存在以下 handler，但当前 `provider.go` 未注册对应 HTTP 路由，因此不作为现有 API 文档记录：

| Handler | 说明 |
|---------|------|
| `Promhttp` | Prometheus handler |
| `PodHandler` | Pod 历史指标，Prometheus query_range 风格响应 |
| `NamespacePodHandler` | 命名空间 Pod 历史指标 |
| `TopNodeHandler` | 节点资源 top 汇总 |
| `NodeHandler` | 节点历史指标 |
| `NamespaceResourceHandler` | 命名空间 CPU/内存汇总 |
