# 应用、镜像和代码包组件

## 应用详情外部服务入口

源文件：`w7panel-ui/src/views/app/apps/detail.vue`

应用详情页读取 `AppGroup.spec.externalServices`，将市场、授权中心、工单系统等第三方服务作为独立菜单展示。面板不识别服务提供方，也不解析订单参数。

```yaml
spec:
  externalServices:
    - key: billing
      title: 授权与续费
      url: https://market.example.com/embed/order/xxx
      openMode: iframe
```

字段说明：

| 字段 | 必填 | 说明 |
|------|------|------|
| `key` | 是 | AppGroup 内唯一的稳定标识，同时用于页面 query 定位 |
| `title` | 是 | 面板菜单标题 |
| `url` | 是 | HTTP 或 HTTPS 服务地址 |
| `openMode` | 否 | 仅支持 `iframe`，未填写时也默认在详情区加载 |

面板会忽略 key 重复、字段不完整、协议不受支持或打开方式无效的入口。`iframe` 服务需要由目标站点允许被面板嵌入。

制品安装时，仓库信息接口可以返回同结构的 `data.external_services`。安装器会校验入口并写入根 AppGroup；内部子应用不会重复写入：

```json
{
  "data": {
    "external_services": [
      {
        "key": "billing",
        "title": "授权与续费",
        "url": "https://market.example.com/#/user-center?tab=orders&order_sn=ORDER-1",
        "openMode": "iframe"
      }
    ]
  }
}
```

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
