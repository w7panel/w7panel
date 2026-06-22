# 集群、节点和容器选择组件

## `NodeSelect`

源文件：`w7panel-ui/src/components/node/node-select.vue`

功能：节点调度选择组件，读取节点标签并生成 `affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution`。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `ns` | object | - | 已有 affinity 数据，用于回显 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `select` | object/string | 选择后的 affinity；未选择时返回空字符串 |

使用示例：

```vue
<NodeSelect :ns="form.affinity" @select="affinity => form.affinity = affinity" />
```

## 节点展示片段

| 组件 | 源文件 | Props | 说明 |
|------|--------|-------|------|
| `NbPage` | `src/components/node/nb-page.vue` | `list` | 节点页面片段 |
| `NdSet` | `src/components/node/nd-set.vue` | - | 节点设置 |

## `SelectContainer`

源文件：`w7panel-ui/src/components/select-container.vue`

功能：级联选择应用组、应用和容器。选择容器后返回容器对象、Pod 标签和 volumes。

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `change` | object | 应用组或应用变化时触发 |
| `complete` | object | 容器选择完成时触发 |

返回对象字段：

| 字段 | 说明 |
|------|------|
| `group` | 应用组名称 |
| `app` | 应用名称 |
| `kind` | Workload 复数类型 |
| `container` | 容器名 |
| `containerObj` | 容器完整配置 |
| `podMatchLabels` | selector labels |
| `podLabels` | template labels |
| `volumes` | workload volumes |

使用示例：

```vue
<SelectContainer @complete="container => selectedContainer = container" />
```

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
