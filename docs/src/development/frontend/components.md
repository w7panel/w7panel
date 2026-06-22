# w7panel-ui 组件说明

本文档是 `w7panel-ui/src/components` 组件文档入口。组件明细已按类型拆分到独立文档；新增或修改组件时，先确认组件类型，再更新对应专题文档。

## 整体使用方式

组件文档用于记录可复用组件的外部契约。开发时先判断组件是否跨页面复用：跨页面复用或被父组件通过 `ref` 调用的组件需要在对应专题文档记录 Props、Events、Ref 方法和使用注意；页面私有组件只在对应页面内维护即可。

### 基本流程

1. 新增公共组件时放到 `src/components/`，并在对应专题文档补充用途、Props、Events 和示例。
2. 修改组件外部参数、事件名、事件参数或 `ref` 方法时，同步更新对应专题文档。
3. 涉及 Wujie 微应用事件的组件，只在组件专题中记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
4. 组件内依赖 API、token、localStorage 或全局状态时，应遵守 [conventions.md](./conventions.md)。

### 场景选择

| 场景 | 维护方式 |
|------|----------|
| 全局注册组件 | 记录组件名、源文件、Props、使用示例 |
| 业务复用组件 | 记录业务场景、外部参数、事件和边界 |
| 弹窗/抽屉组件 | 记录显示控制字段、关闭事件和提交事件 |
| 通过 `ref` 暴露方法 | 记录方法名、参数、返回值和调用时机 |
| 微前端桥接组件 | 组件职责在本文，事件协议在 Wujie 文档 |

## 使用约定

- 全局注册组件可直接在模板中使用；其余组件需要在页面内 `import` 后注册。
- 抽屉、弹窗类组件统一用 `show` 控制显示，关闭后通常触发 `close` 或 `cancel`，父组件应在事件里把 `show` 置为 `false`。
- 涉及 K8s 资源的组件默认使用当前命名空间 store，并通过 `k8sproxy` 或 `panelApi` 请求后端。
- 组件方法只在源码中明确暴露或被父组件通过 `ref` 调用时记录。

## 组件分类文档

| 类型 | 文档 | 组件 |
|------|------|------|
| 基础展示与导航 | [components-basic.md](./components-basic.md) | `Chart`、`Breadcrumb`、`RouteBreadcrumb`、`DefaultLayout`、`Menu`、`Navbar`、`ContactUs` |
| 表单控件 | [components-forms.md](./components-forms.md) | `CustomCheckbox`、`CronJob`、`HealthProbe`、`DcformDrawer` |
| YAML 编辑 | [components-yaml.md](./components-yaml.md) | `YamlInput`、`YamlEditor`、`YamlDrawer`、`YamlView`、`K8sYamlDrawer` |
| 日志、终端和文本对比 | [components-terminal.md](./components-terminal.md) | `PodLog`、`JobLog`、`WebShell`、`WebshellTty`、`DiffTxt` |
| 应用、镜像和代码包 | [components-apps.md](./components-apps.md) | `StoreInstall`、`StoreInstallDrawer`、`AddappDrawer`、`AppForm`、`AppFormContainer`、`AppFormVolumes`、`MicroAppForm`、`BuildImageDrawer`、`CodepackDrawer` |
| 域名、灰度和缓存策略 | [components-domain.md](./components-domain.md) | `DomainStrategy`、`DomainMicroEdit`、`DomainParseAlert`、`DomainGrayRelease`、`DomainStrategyPlugin`、缓存策略插件 |
| 集群、节点和容器选择 | [components-cluster.md](./components-cluster.md) | `NodeSelect`、`NbPage`、`NdSet`、`SelectContainer` |
| 资源、账号和权限 | [components-resource.md](./components-resource.md) | `ContainerPlugin`、`QuotaConfig`、`QuotaEdit`、`PermissionEdit` |
| 业务卡片和入口 | [components-business.md](./components-business.md) | `PodsCharts`、`HelmItem`、`RespoItem`、`TopappMenu`、`SlideCapt` |
| 微前端桥接 | [components-microfrontend.md](./components-microfrontend.md) | `WujieModals` |

## 维护建议

- 新增组件时补充：用途、Props、Events、Ref 方法、使用示例和依赖的 store/API。
- 跨模块复用组件放在 `src/components/`；只被单个页面使用的组件优先保留在对应 `views/` 目录。
- 抽屉、弹窗类组件要延迟请求数据，避免页面加载时触发无用 API。
- 涉及终端、日志、WebShell 的组件要在 `onUnmounted` 或关闭动作中释放连接、终止请求和清理 xterm 实例。
- 修改后端 API 返回字段时，同步检查本文件涉及的组件是否需要调整。
