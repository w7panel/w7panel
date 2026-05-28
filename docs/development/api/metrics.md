# w7panel-server/app/metrics API 文档

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
