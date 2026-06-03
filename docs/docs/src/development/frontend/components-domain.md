# 域名、灰度和缓存策略组件

## `DomainStrategy`

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

## `DomainMicroEdit`

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

## `DomainParseAlert`

源文件：`w7panel-ui/src/components/domain-parse-alert.vue`

功能：域名解析提示组件，用于展示 DNS 解析类型和值。

## `DomainGrayRelease`

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

## `DomainStrategyPlugin`

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

## 插件配置组件

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

## 缓存策略组件

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `DomainStrategyFilecache` | `src/components/domain-strategy-filecache.vue` | `data`, `activeName` | `submit`, `cancel` | 文件缓存策略配置，提交 operations |
| `DomainStrategyImagecache` | `src/components/domain-strategy-imagecache.vue` | `data`, `activeName` | `submit`, `cancel` | 图片缓存策略配置 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
