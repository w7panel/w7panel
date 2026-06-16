# YAML 编辑组件

## `YamlInput`

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

## `YamlEditor`

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

## `YamlDrawer`

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

## `YamlView`

源文件：`w7panel-ui/src/components/yaml-view.vue`

功能：YAML 只读展示组件。

Props：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | object/string | - | 待展示的 YAML 对象或字符串 |

## `K8sYamlDrawer`

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

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
