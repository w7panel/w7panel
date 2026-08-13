# 应用、镜像和代码包组件

## 应用详情制品市场菜单

源文件：`w7panel-ui/src/views/app/apps/detail.vue`

应用详情页仅使用 MicroApp Binding 渲染菜单。制品库在 `info` 请求完成订单校验后，动态覆盖本次 Helm 包中 MicroApp 的 `name: zpk-market` Binding；前端按普通 MicroApp 菜单统一渲染和切换。

```yaml
bindings:
  - name: zpk-market
    title: 服务中心
    support: thirdparty_cd
    menu:
      - title: 授权与续费
        do: '#/user-orders?tab=orders&order_sn=ORDER-1'
        location: back
```

字段说明：

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 固定为 `zpk-market`，用于覆盖制品内原有市场分组 |
| `title` | 是 | 分组标题 |
| `support` | 是 | 使用现有值 `thirdparty_cd`，进入普通微应用菜单流程 |
| `menu[].location` | 否 | 菜单位置，服务中心使用 `back` 展示在应用菜单底部 |
| `menu[].do` | 是 | 微应用内路由，不包含域名；可包含安装后获得的订单参数 |

面板不识别或硬编码 `zpk-market`：`founder` 按现有规则展示全部 Binding 菜单，其他角色只展示自身角色菜单。菜单分组使用现有默认图标，不新增 Binding `icon` 或菜单 `key` 字段。

`location: back` 的菜单统一在侧边栏底部平铺展示，不显示 Binding 分组标题；Binding 归属仍用于查找对应的运行配置。

制品库同步修改 Helm `values.yaml` 中的 `bindings` 和 `backend_config`：`backend_config[role=zpk-market]` 独立定义为 `type=external`、`load_mode=iframe`，后端地址保存市场域名，菜单 `do` 只保存站内路由。应用详情页和顶部微应用页切换菜单时都按 Binding 名称选择对应的 `roleConfig`，再依据 `load_mode` 与 `serverUrl` 加载，因此市场入口不会继承应用 `founder` 的后端、代理或前端属性。没有市场 Binding 时直接返回原 Helm 包地址，不执行动态替换；有市场 Binding 时其他配置不变。

动态包从 `PackFormulaToHelmAndPack(..., false)` 取得当前缓存包，并使用基础包状态与 Bindings 内容生成缓存键；相同内容直接复用，变化时才解包重打。同一进程的并发生成会合并。动态文件与原包位于同一目录，按 `{原文件名去扩展名}-{hash}.tgz` 命名；公共缓存包不会写入订单 URL。

## `StoreInstall`

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

## `StoreInstallDrawer`

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

## `AddappDrawer`

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

## `AppForm`

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
| `exportFormData()` | Promise[object] | 导出 K8s workload 对象，不提交 |
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

## `AppFormContainer`

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
| `isTemplate` | boolean | `false` | 模板模式，隐藏复用已有应用、代码包构建等依赖集群资源的功能 |
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

## `AppFormVolumes`

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

## `MicroAppForm`

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

## `BuildImageDrawer`

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

## `CodepackDrawer`

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

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
