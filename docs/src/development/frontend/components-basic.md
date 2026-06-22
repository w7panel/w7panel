# 基础展示与全局组件

## `Chart`

源文件：`w7panel-ui/src/components/chart/index.vue`

功能：基于 `vue-echarts` 的图表封装，统一处理 ECharts 渲染和自适应尺寸。由 `src/components/index.ts` 全局注册。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `options` | object | `{}` | ECharts option |
| `autoResize` | boolean | `true` | 容器尺寸变化时是否自动 resize |
| `width` | string | `100%` | 图表宽度 |
| `height` | string | `100%` | 图表高度 |

使用示例：

```vue
<template>
  <Chart :options="chartOption" height="320px" />
</template>
```

注意事项：

- `options` 必须是 ECharts 标准配置。
- 当前全局只注册 CanvasRenderer、Bar、Line、Pie、Radar、Grid、Tooltip、Legend、DataZoom、Graphic；新增图表类型时需要同步更新 `src/components/index.ts`。

## `Breadcrumb`

源文件：`w7panel-ui/src/components/breadcrumb/index.vue`

功能：Arco Breadcrumb 封装，支持按路由列表渲染并点击跳转。由 `src/components/index.ts` 全局注册。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `items` | array | `[]` | 兼容字段，当前模板主要使用 `routes` |
| `routes` | array | - | 面包屑路由列表，元素通常包含 `name`、`params`、`label` |

使用示例：

```vue
<Breadcrumb :routes="[{ label: '应用管理', name: 'app' }, { label: '详情' }]" />
```

## `RouteBreadcrumb`

源文件：`w7panel-ui/src/components/route-breadcrumb.vue`

功能：根据当前路由渲染页面面包屑。由 `src/components/index.ts` 全局注册。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | any | - | 自定义面包屑数据，未传时按当前路由信息渲染 |

使用示例：

```vue
<RouteBreadcrumb />
```

## 布局、导航和联系入口

| 组件 | 源文件 | 说明 |
|------|--------|------|
| `DefaultLayout` | `src/components/default-layout.vue` | 默认布局 |
| `Menu` | `src/components/menu/index.vue` | 侧边菜单 |
| `Navbar` | `src/components/navbar/index.vue` | 顶部导航 |
| `ContactUs` | `src/components/contact-us.vue` | 联系入口，读取公开站点联系配置 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
