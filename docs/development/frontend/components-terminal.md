# 日志、终端和文本对比组件

## `PodLog`

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

## `JobLog`

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

## `WebShell`

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

## `WebshellTty`

源文件：`w7panel-ui/src/components/webshell-tty.vue`

功能：底层 TTY 终端连接展示组件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `token` | string | - | 连接 token |
| `command` | string | - | 执行命令 |
| `show` | boolean | - | 是否显示 |
| `type` | string | - | 连接类型 |

## `DiffTxt`

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

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
