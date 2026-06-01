# w7panel-ui 组件说明

本文档整理 `w7panel-ui/src/components` 下的公共组件和业务复用组件。组件按实际用途分组，字段以当前源码为准。

## 使用约定

- 全局注册组件可直接在模板中使用；其余组件需要在页面内 `import` 后注册。
- 抽屉、弹窗类组件统一用 `show` 控制显示，关闭后通常触发 `close` 或 `cancel`，父组件应在事件里把 `show` 置为 `false`。
- 涉及 K8s 资源的组件默认使用当前命名空间 store，并通过 `k8sproxy` 或 `panelApi` 请求后端。
- 组件方法只在源码中明确暴露或被父组件通过 `ref` 调用时记录。

## 当前覆盖范围

当前 `w7panel-ui/src/components` 下包含通用组件、应用安装/部署组件、域名策略组件、终端日志组件和微前端桥接组件。本文重点记录被多处复用、通过全局组件注册、或对外暴露 Props/Events/Ref 方法的组件。

尚未逐项展开的组件包括：

| 组件 | 源文件 | 说明 |
|------|--------|------|
| `ContactUs` | `src/components/contact-us.vue` | 联系我们展示组件 |
| `ContainerPlugin` | `src/components/container-plugin.vue` | 制品库/容器插件配置，已通过 Wujie `containerPlugin` 事件间接对外 |
| `CostEdit` | `src/components/cost-edit.vue` | 费用配置编辑 |
| `DcformDrawer` | `src/components/dcform-drawer.vue` | 表单抽屉 |
| `DefaultLayout` | `src/components/default-layout.vue` | 组件级默认布局 |
| `DomainStrategyFilecache` | `src/components/domain-strategy-filecache.vue` | 文件缓存微应用容器，事件详见 `wujie-events.md` |
| `DomainStrategyImagecache` | `src/components/domain-strategy-imagecache.vue` | 镜像缓存微应用容器，事件详见 `wujie-events.md` |
| `DomainStrategyPluginFilecache` | `src/components/domain-strategy-plugin-filecache.vue` | 文件缓存策略插件 |
| `DomainStrategyPluginRatelimit` | `src/components/domain-strategy-plugin-ratelimit.vue` | 限流策略插件 |
| `DomainStrategyPluginWhitelist` | `src/components/domain-strategy-plugin-whitelist.vue` | 白名单策略插件 |
| `PermissionEdit` | `src/components/permission-edit.vue` | 权限编辑 |
| `QuotaConfig` | `src/components/quota-config.vue` | 配额配置 |
| `QuotaEdit` | `src/components/quota-edit.vue` | 配额编辑 |
| `TestResource` | `src/components/test-resource.vue` | 资源测试组件 |
| `WujieModals` | `src/components/wujie-modals.vue` | 微应用通用弹窗和面板能力桥接，事件详见 `wujie-events.md` |
| `DdcNode` | `src/components/node/ddc-node.vue` | 节点相关业务组件 |
| `NbPage` | `src/components/node/nb-page.vue` | 节点绑定相关页面组件 |
| `NdSet` | `src/components/node/nd-set.vue` | 节点设置组件 |

这些组件发生 Props、Events、Ref 方法或对外行为变化时，也需要同步补充到本文。

## 全局注册组件

### `Chart`

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

### `Breadcrumb`

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

### `RouteBreadcrumb`

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

## 表单与编辑组件

### `CustomCheckbox`

源文件：`w7panel-ui/src/components/custom-checkbox.vue`

功能：支持自定义选中值和未选中值的 Checkbox，适合字段值不是简单 boolean 的表单。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `modelValue` | boolean/string/number | `false` | 当前值 |
| `checkedValue` | boolean/string/number | `true` | 选中时写入的值 |
| `uncheckedValue` | boolean/string/number | `false` | 取消选中时写入的值 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `update:modelValue` | checkedValue/uncheckedValue | 值变化时触发 |

使用示例：

```vue
<CustomCheckbox
  v-model="form.storageRwMode"
  checked-value="ReadWriteMany"
  unchecked-value="ReadWriteOnce"
>
  多写
</CustomCheckbox>
```

### `CronJob`

源文件：`w7panel-ui/src/components/cron-job.vue`

功能：用表单方式编辑 5 位 cron 表达式，支持每月、每周、每天、每小时、每隔 N 日、每隔 N 小时、每隔 N 分钟。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `value` | string | - | cron 表达式，例如 `0 2 * * *` |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `change` | string | 用户修改后输出新的 cron 表达式 |

使用示例：

```vue
<CronJob :value="form.cron" @change="value => form.cron = value" />
```

### `YamlInput`

源文件：`w7panel-ui/src/components/yaml-input.vue`

功能：轻量 YAML 输入组件，适合在表单里编辑一段 YAML 文本。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `domid` | string | - | 编辑器 DOM id |
| `value` | string | - | YAML 初始内容 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `submit` | string | 提交后的 YAML 文本 |

使用示例：

```vue
<YamlInput domid="config-yaml" :value="yamlText" @submit="saveYaml" />
```

### `YamlEditor`

源文件：`w7panel-ui/src/components/yaml-editor.vue`

功能：基于 CodeMirror 的 YAML 编辑器，内置 YAML 语法、Tab/Shift+Tab 缩进和 K8s 元数据清理。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `yaml` | string | - | YAML 文本 |
| `disabled` | boolean | `false` | 是否只读 |
| `nofooter` | boolean | `false` | 是否隐藏底部确定/取消按钮 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `submit` | string | 提交后的 YAML 文本 |
| `cancel` | - | 取消编辑 |

Ref 方法：

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `getValue()` | string | 获取当前 YAML，并清理 `resourceVersion`、`uid`、`creationTimestamp`、`managedFields` 等字段 |
| `write(value)` | void | 覆盖编辑器内容 |

使用示例：

```vue
<YamlEditor
  ref="yamlEditor"
  :yaml="yamlText"
  @submit="handleSubmit"
  @cancel="visible = false"
/>
```

### `YamlDrawer`

源文件：`w7panel-ui/src/components/yaml-drawer.vue`

功能：抽屉式 YAML 编辑器。传入对象时会转为 YAML，提交时默认转回对象。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `title` | string | - | 抽屉标题 |
| `data` | object/string | - | YAML 对象或 YAML 字符串 |
| `show` | boolean | `false` | 是否显示抽屉 |
| `returnYaml` | boolean | `false` | 为 `true` 时提交原始 YAML 字符串 |
| `disabled` | boolean | `false` | 是否只读 |
| `nofooter` | boolean | `false` | 是否隐藏编辑器底部按钮 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `submit` | object/string | `returnYaml=false` 时返回对象，否则返回 YAML 字符串 |
| `cancel` | - | 关闭抽屉 |

使用示例：

```vue
<YamlDrawer
  title="编辑资源"
  :show="yamlVisible"
  :data="resource"
  @submit="updateResource"
  @cancel="yamlVisible = false"
/>
```

### `YamlView`

源文件：`w7panel-ui/src/components/yaml-view.vue`

功能：YAML 只读展示组件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object/string | - | 待展示的 YAML 对象或字符串 |

### `K8sYamlDrawer`

源文件：`w7panel-ui/src/components/k8syaml-drawer.vue`

功能：K8s 资源 YAML 抽屉，常用于资源查看、编辑和刷新列表。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 是否需要刷新列表 |

## 日志、终端和差异组件

### `PodLog`

源文件：`w7panel-ui/src/components/pod-log.vue`

功能：查看 Pod 容器日志，使用 xterm 渲染日志内容，支持弹窗或内联模式。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | - | 是否显示，必填 |
| `podName` | string | `''` | Pod 名称 |
| `data` | object | `null` | 兼容旧格式的 Pod 数据 |
| `namespace` | string | `''` | 命名空间；为空时使用当前命名空间 |
| `container` | string | `''` | 默认容器名 |
| `containers` | array | `null` | 容器列表 |
| `tailLines` | number | `100` | 默认日志行数 |
| `token` | string | `''` | 自定义 token |
| `local` | boolean | `false` | 是否 local 模式 |
| `mode` | string | `modal` | 显示模式：`modal` 或 `inline` |
| `title` | string | `查看日志` | 弹窗标题 |
| `width` | number | `1000` | 弹窗宽度 |
| `height` | number | `400` | 终端高度 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | - | 关闭日志视图 |

使用示例：

```vue
<PodLog
  :show="logVisible"
  pod-name="nginx-xxx"
  namespace="default"
  container="nginx"
  @close="logVisible = false"
/>
```

注意事项：

- `mode="inline"` 时由父容器控制布局高度。
- 组件关闭或卸载时会清理终端和请求，新增日志类组件也要保持同样行为。

### `JobLog`

源文件：`w7panel-ui/src/components/job-log.vue`

功能：查看 Job 执行日志，支持 Job 列表、容器切换、日志行数和弹窗/内联模式。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | - | 是否显示，必填 |
| `jobName` | string | `''` | Job 名称 |
| `jobList` | array | `[]` | Job 列表 |
| `name` | string | `''` | 兼容旧字段的 Job 名称 |
| `namespace` | string | `''` | 命名空间 |
| `labelSelector` | string | `''` | Pod label selector |
| `containers` | array | `null` | 容器列表 |
| `tailLines` | number | `100` | 默认日志行数 |
| `local` | boolean | `false` | 是否 local 模式 |
| `mode` | string | `modal` | 显示模式：`modal` 或 `inline` |
| `title` | string | `执行记录` | 弹窗标题 |
| `width` | number | `1100` | 弹窗宽度 |
| `height` | number | `400` | 终端高度 |
| `showTabs` | boolean | `true` | 是否显示左侧 Tab |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | - | 关闭日志视图 |

使用示例：

```vue
<JobLog
  :show="jobLogVisible"
  name="install-job"
  :tail-lines="200"
  @close="jobLogVisible = false"
/>
```

### `WebShell`

源文件：`w7panel-ui/src/components/web-shell.vue`

功能：容器 WebShell 页面组件，负责组装连接参数并展示终端。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `api_token` | string | - | API token |
| `type` | string | - | 连接类型 |
| `pod` | string/object | - | Pod 信息 |
| `defaultCommand` | string | - | 默认执行命令 |
| `namespace` | string | - | 命名空间 |
| `containerName` | string | - | 容器名 |
| `show` | boolean | - | 是否显示 |
| `origin` | string | - | WebShell 源地址 |
| `ip` | string | - | Pod/Agent IP |

使用示例：

```vue
<WebShell
  :show="shellVisible"
  namespace="default"
  pod="nginx-xxx"
  container-name="nginx"
/>
```

### `WebshellTty`

源文件：`w7panel-ui/src/components/webshell-tty.vue`

功能：底层 TTY 终端连接展示组件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `token` | string | - | 连接 token |
| `command` | string | - | 执行命令 |
| `show` | boolean | - | 是否显示 |
| `type` | string | - | 连接类型 |

### `DiffTxt`

源文件：`w7panel-ui/src/components/diff-txt.vue`

功能：文本差异展示弹窗。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `title` | string | - | 标题 |
| `new` | string | - | 新文本 |
| `old` | string | - | 旧文本 |
| `show` | boolean | `false` | 是否显示 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `cancel` | - | 关闭弹窗 |

使用示例：

```vue
<DiffTxt
  title="YAML 差异"
  :old="oldYaml"
  :new="newYaml"
  :show="diffVisible"
  @cancel="diffVisible = false"
/>
```

## 应用安装和部署组件

### `StoreInstall`

源文件：`w7panel-ui/src/components/store-install.vue`

功能：应用商店安装流程组件，包含配置项填写、依赖检测、安装执行、状态轮询和日志入口。可作为页面使用，也可嵌入抽屉。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `is_component` | boolean | `false` | 是否作为组件嵌入；为 `false` 时从路由 query 读取安装参数 |
| `path_identifie` | string | - | 组件模式下的 zpk 路径或标识 |
| `version` | string | - | 指定安装版本 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `needInstall` | identifie, callback | 依赖应用未安装时触发 |
| `installed` | identifie | 安装任务已提交或安装完成节点触发 |
| `installedStatusSuccess` | identifie | 安装状态检测成功 |
| `close` | - | 组件模式下关闭 |

使用示例：

```vue
<StoreInstall
  :is_component="true"
  :path_identifie="installPath"
  @installed="refreshList"
  @close="installVisible = false"
/>
```

注意事项：

- 非组件模式依赖路由参数：`path`、`releasename`、`completeName`、`domain`、`thirdpartyCDToken` 等。
- 组件内部会请求存储、镜像仓库、IngressClass、白名单、zpk 配置等资源。

### `StoreInstallDrawer`

源文件：`w7panel-ui/src/components/store-install-drawer.vue`

功能：抽屉式应用安装入口，内部承载 `StoreInstall`。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示抽屉 |
| `path` | string | - | 安装路径或标识 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | - | 关闭抽屉 |
| `installed` | - | 安装已触发 |
| `installedStatusSuccess` | - | 安装状态成功 |

使用示例：

```vue
<StoreInstallDrawer
  :show="installVisible"
  :path="zpkPath"
  @close="installVisible = false"
  @installed="refreshList"
/>
```

### `AddappDrawer`

源文件：`w7panel-ui/src/components/addapp-drawer.vue`

功能：创建或编辑应用的抽屉，支持主应用和子应用 tab，内部使用 `AppForm`。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示抽屉 |
| `tabs` | array | - | 编辑模式下的应用 tab 列表；为空时进入创建模式 |
| `activeName` | string | - | 默认激活 tab key |
| `groupname` | string | - | 应用组名，创建时写入 label |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 关闭抽屉；参数为 `true` 时通常表示需要刷新列表 |

使用示例：

```vue
<AddappDrawer
  :show="drawerVisible"
  :tabs="editTabs"
  :active-name="activeTab"
  :groupname="groupName"
  @close="refresh => { drawerVisible = false; refresh && getList(); }"
/>
```

### `AppForm`

源文件：`w7panel-ui/src/components/app-form.vue`

功能：应用表单主体，负责基本信息、应用类型、数据卷、容器、节点调度、容忍度配置，并支持创建或更新 Deployment/StatefulSet/DaemonSet。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `id` | string | - | 已有应用名称；传入时进入编辑模式 |
| `kind` | string | - | K8s workload 复数类型：`deployments`、`statefulsets`、`daemonsets` |
| `defaultData` | object | - | 外部传入的 workload 数据；优先用于初始化表单 |
| `parent` | string/boolean | - | 子应用父级名称；`false` 表示主应用 |
| `afterName` | string | - | 创建时附加到应用标识后的后缀 |
| `groupname` | string | - | 应用组名 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `getInfo` | `{ name, title, isSubmit, submitStatus }` | 表单名称变化或提交后通知父组件 |
| `showTestResource` | object | 资源检测信息，当前由父级抽屉展示 |

Ref 方法：

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `validate()` | Promise | 校验表单 |
| `exportFormData()` | Promise<object> | 导出 K8s workload 对象，不提交 |
| `submit(hideMessage)` | Promise | 创建或更新 K8s workload |

使用示例：

```vue
<AppForm
  ref="appFormRef"
  :id="appName"
  kind="deployments"
  :groupname="groupName"
  @getInfo="info => appInfo = info"
/>
```

```ts
await appFormRef.value.submit(true);
```

### `AppFormContainer`

源文件：`w7panel-ui/src/components/app-form-container.vue`

功能：容器配置表单，支持业务容器、初始化容器、端口、环境变量、镜像仓库、资源配置、挂载卷、健康检查等。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | Workload 原始数据 |
| `volumes` | array | `[]` | 可挂载的 volumes |
| `volumeClaimTemplates` | array | `[]` | StatefulSet 动态 PVC 模板 |
| `mirror` | array | `[]` | 镜像仓库列表 |
| `isPlugin` | boolean | `false` | 插件模式 |
| `pluginData` | object | - | 插件数据 |
| `layout` | string | - | 布局模式 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `getMirror` | - | 请求父组件刷新镜像仓库 |
| `editMirror` | string | 编辑或新建镜像仓库；空字符串表示新建 |
| `delMirror` | string | 删除镜像仓库 |
| `showExtra` | boolean | 是否显示节点调度等额外配置 |
| `addVolumes` | array | 请求父组件追加 volumes |

Ref 方法：

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `formToData()` | object | 输出 `initContainers`、`containers`、`hostPorts`、`imagePullSecrets` |

使用示例：

```vue
<AppFormContainer
  ref="containerRef"
  :data="workload"
  :volumes="volumes"
  :volume-claim-templates="volumeClaimTemplates"
  :mirror="mirror"
  @getMirror="getMirror"
/>
```

### `AppFormVolumes`

源文件：`w7panel-ui/src/components/app-form-volumes.vue`

功能：应用数据卷配置组件，支持 NFS、emptyDir、hostPath、已有 PVC、StatefulSet 动态 PVC、ConfigMap、Secret。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `id` | string | - | 已有应用名称；编辑已有 StatefulSet 时禁止修改动态 PVC |
| `data` | object | - | Workload 原始数据 |
| `kind` | string | - | Workload 类型，决定是否可创建动态 PVC |
| `readonly` | boolean | `false` | 是否只读展示 |
| `isPlugin` | boolean | `false` | 插件模式下隐藏 ConfigMap/Secret 类型 |
| `isTemplate` | boolean | `false` | 模板模式，不请求集群资源 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `submit` | `{ volumes, volumeClaimTemplates }` | 每次数据卷变化后输出 K8s 字段 |

Ref 方法：

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `addItemFromOut(addVolumes)` | void | 从外部追加 volumes 并重新输出 |

使用示例：

```vue
<AppFormVolumes
  :data="workload"
  kind="statefulsets"
  @submit="v => { volumes = v.volumes; volumeClaimTemplates = v.volumeClaimTemplates; }"
/>
```

### `MicroAppForm`

源文件：`w7panel-ui/src/components/micro-app-form.vue`

功能：微应用配置表单抽屉，用于 YAML/JSON 配置编辑。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `yaml` | string | - | YAML 内容 |
| `json` | object | - | JSON 内容 |
| `callback` | function | - | 提交后的回调 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 关闭抽屉 |

### `BuildImageDrawer`

源文件：`w7panel-ui/src/components/build-image-drawer.vue`

功能：构建镜像抽屉，适合从节点、应用或代码包入口触发构建。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `data` | object | - | 构建上下文数据 |
| `nodeName` | string | - | 节点名称 |
| `nodeIp` | string | - | 节点 IP |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 关闭抽屉；参数为 `true` 时通常表示需要刷新 |

### `CodepackDrawer`

源文件：`w7panel-ui/src/components/codepack-drawer.vue`

功能：代码包抽屉，支持 tab 化展示和关闭刷新。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `tabs` | array | - | tab 列表 |
| `activeName` | string | - | 默认激活 tab |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 关闭抽屉；参数为 `true` 时刷新 |

## 域名和策略组件

### `DomainStrategy`

源文件：`w7panel-ui/src/components/domain-strategy.vue`

功能：域名策略主组件，负责域名、端口、路径、重写、插件、灰度、缓存等策略配置。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `title` | string | - | 标题 |
| `show` | boolean | `false` | 是否显示 |
| `data` | object | - | 域名策略数据 |
| `hideRewrite` | boolean | `false` | 是否隐藏重写配置 |
| `multiple` | boolean | `false` | 是否多选/批量模式 |
| `isMicroComponents` | boolean | `false` | 是否微应用组件模式 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `cancel` | - | 取消或关闭 |
| `submit` | data, callback | 提交策略数据 |
| `refresh` | - | 子组件要求刷新 |

使用示例：

```vue
<DomainStrategy
  :show="strategyVisible"
  :data="strategy"
  title="域名策略"
  @submit="saveStrategy"
  @cancel="strategyVisible = false"
/>
```

### `DomainMicroEdit`

源文件：`w7panel-ui/src/components/domain-micro-edit.vue`

功能：微应用域名编辑组件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `data` | object | - | 当前域名数据 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `submit` | object | 提交后的域名配置 |
| `close` | boolean | 关闭 |

### `DomainParseAlert`

源文件：`w7panel-ui/src/components/domain-parse-alert.vue`

功能：域名解析提示组件，用于展示 DNS 解析类型和值。

### `DomainGrayRelease`

源文件：`w7panel-ui/src/components/domain-gray-release.vue`

功能：灰度发布配置组件，支持选择应用、端口、路径和发布规则。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `appList` | array | `[]` | 应用列表 |
| `appPorts` | object/array | - | 应用端口数据 |
| `parentName` | string | - | 父应用名称 |
| `parentPath` | string | - | 父路径 |
| `checkList` | array | `[]` | 已选列表 |
| `multiple` | boolean | `false` | 是否多选 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `cancel` | boolean | 关闭；参数表示是否刷新或确认 |

### `DomainStrategyPlugin`

源文件：`w7panel-ui/src/components/domain-strategy-plugin.vue`

功能：域名策略插件入口，统一管理限流、白名单、缓存等插件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | 插件配置 |
| `show` | boolean | `false` | 是否显示 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | - | 关闭 |
| `pluginbadge` | boolean | 插件启用状态变化 |

### 插件配置组件

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `DomainStrategyPluginRatelimit` | `src/components/domain-strategy-plugin-ratelimit.vue` | `show`, `config` | `submit`, `close` | 限流插件配置 |
| `DomainStrategyPluginWhitelist` | `src/components/domain-strategy-plugin-whitelist.vue` | `show`, `data` | `submit`, `close` | 白名单插件配置 |
| `DomainStrategyPluginFilecache` | `src/components/domain-strategy-plugin-filecache.vue` | `show`, `rules`, `keyrules`, `data` | `submit`, `close` | 文件缓存插件配置 |

使用示例：

```vue
<DomainStrategyPluginRatelimit
  :show="rateLimitVisible"
  :config="rateLimitConfig"
  @submit="config => rateLimitConfig = config"
  @close="rateLimitVisible = false"
/>
```

### 缓存策略组件

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `DomainStrategyFilecache` | `src/components/domain-strategy-filecache.vue` | `data`, `activeName` | `submit`, `cancel` | 文件缓存策略配置，提交 operations |
| `DomainStrategyImagecache` | `src/components/domain-strategy-imagecache.vue` | `data`, `activeName` | `submit`, `cancel` | 图片缓存策略配置 |

## 集群、节点和资源组件

### `NodeSelect`

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

### `NodeBind`

源文件：`w7panel-ui/src/components/node/node-bind.vue`

功能：节点绑定抽屉。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `show` | boolean | `false` | 是否显示 |
| `list` | array | `[]` | 节点或绑定列表 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `close` | boolean | 关闭 |

### 节点展示片段

| 组件 | 源文件 | Props | 说明 |
|------|--------|-------|------|
| `DdcNode` | `src/components/node/ddc-node.vue` | `list` | 节点展示 |
| `NbPage` | `src/components/node/nb-page.vue` | `list` | 节点页面片段 |
| `NdSet` | `src/components/node/nd-set.vue` | - | 节点设置 |

### `SelectContainer`

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

### 资源和权限配置组件

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `ContainerPlugin` | `src/components/container-plugin.vue` | `propsData` | - | 容器插件配置 |
| `QuotaConfig` | `src/components/quota-config.vue` | `data` | `setQuotaPage` | 配额入口配置 |
| `QuotaEdit` | `src/components/quota-edit.vue` | `show`, `data`, `name`, `clustermode` | `submit`, `close` | 配额编辑 |
| `CostEdit` | `src/components/cost-edit.vue` | `show`, `data`, `name`, `list`, `onlypackage` | `submit`, `close` | 成本编辑 |
| `PermissionEdit` | `src/components/permission-edit.vue` | `show`, `list`, `name`, `type`, `debug`, `fileeditor`, `whitelist`, `webshell`, `permissionPackage`, `disabledBase`, `disabledMenu`, `noCustom` | `submit`, `close` | 权限编辑 |
| `TestResource` | `src/components/test-resource.vue` | `cpu`, `memory`, `replica`, `novisible` | `changeStatus`, `onlyshow` | 资源检测 |

## 展示和业务卡片组件

### `PodsCharts`

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

### `HealthProbe`

源文件：`w7panel-ui/src/components/health-probe.vue`

功能：健康检查配置组件，支持 liveness/readiness/startup 等 probe 数据编辑。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object | - | probe 数据 |

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `returnData` | object | 输出健康检查配置 |

### `HelmItem`

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

### `RespoItem`

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

### `TopappMenu`

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

### `SlideCapt`

源文件：`w7panel-ui/src/components/slide-capt.vue`

功能：滑块验证码。

Events：

| 事件 | 参数 | 说明 |
|------|------|------|
| `confirm` | object | 验证通过数据 |
| `close` | - | 关闭 |

### 其它布局和入口组件

| 组件 | 源文件 | 说明 |
|------|--------|------|
| `DefaultLayout` | `src/components/default-layout.vue` | 默认布局 |
| `Footer` | `src/components/footer/index.vue` | 页脚 |
| `Menu` | `src/components/menu/index.vue` | 侧边菜单 |
| `Navbar` | `src/components/navbar/index.vue` | 顶部导航 |
| `TabBar` | `src/components/tab-bar/index.vue` | 标签栏 |
| `TabItem` | `src/components/tab-bar/tab-item.vue` | 标签栏条目 |
| `WujieModals` | `src/components/wujie-modals.vue` | 微前端弹窗集合 |
| `ContactUs` | `src/components/contact-us.vue` | 联系入口 |
| `DcformDrawer` | `src/components/dcform-drawer.vue` | 动态表单抽屉，关闭时触发 `close(refreshList)` |

## 维护建议

- 新增组件时补充：用途、Props、Events、Ref 方法、使用示例和依赖的 store/API。
- 跨模块复用组件放在 `src/components/`；只被单个页面使用的组件优先保留在对应 `views/` 目录。
- 抽屉、弹窗类组件要延迟请求数据，避免页面加载时触发无用 API。
- 涉及终端、日志、WebShell 的组件要在 `onUnmounted` 或关闭动作中释放连接、终止请求和清理 xterm 实例。
- 修改后端 API 返回字段时，同步检查本文件涉及的组件是否需要调整。
