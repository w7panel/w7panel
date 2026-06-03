# 业务卡片和入口组件

## `PodsCharts`

源文件：`w7panel-ui/src/components/pods-charts.vue`

功能：Pod 指标图表展示。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `list` | array | `[]` | 指标列表 |
| `type` | string | - | 图表类型 |
| `noTitle` | boolean | `false` | 是否隐藏标题 |
| `pickerValue` | any | - | 时间选择值 |
| `step` | string/number | - | 查询步长 |

## `HelmItem`

源文件：`w7panel-ui/src/components/helm-item.vue`

功能：Helm 应用项展示。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | Helm 项数据 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `refresh` | - | 请求刷新 |

## `RespoItem`

源文件：`w7panel-ui/src/components/respo-item.vue`

功能：仓库项展示，支持标签点击和支付/授权流程。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | 仓库项数据 |
| `webUrl` | string | - | Web 地址 |
| `thirdparty_cd_token` | string | - | 第三方 token |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `toPay` | ticket | 触发支付 |
| `tagClick` | item | 标签点击 |

## `TopappMenu`

源文件：`w7panel-ui/src/components/topapp-menu.vue`

功能：顶部应用菜单。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `roles` | array | - | 角色列表 |
| `info` | object | - | 当前应用信息 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `routeChange` | route | 路由变化 |

## `SlideCapt`

源文件：`w7panel-ui/src/components/slide-capt.vue`

功能：滑块验证码。

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `confirm` | object | 验证通过数据 |
| `close` | - | 关闭 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
