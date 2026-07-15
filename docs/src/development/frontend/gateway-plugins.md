# 网关插件前端接入

## 数据来源

网关插件以 Higress `WasmPlugin` 为唯一数据源：

```text
/k8s-proxy/apis/extensions.higress.io/v1alpha1/namespaces/higress-system/wasmplugins
```

列表、添加、修改和删除分别使用 K8s 代理的 GET、POST、PUT 和 DELETE。MicroApp 仅提供配置操作界面，不作为插件列表数据源。

## Metadata 约定

扩展能力写入 `metadata.annotations`，避免修改 Higress CRD：

| Annotation | 默认值 | 说明 |
|------------|--------|------|
| `w7.cc/plugin-enabled` | `true` | 插件级启停状态 |
| `w7.cc/plugin-support-global` | `true` | 是否支持全局配置 |
| `w7.cc/plugin-support-rule` | `false` | 是否支持域名规则配置 |
| `w7.cc/plugin-microapp` | 空 | 旧资源或手工添加时显式关联的 MicroApp 名称 |
| `w7.cc/plugin-disabled-state` | 空 | 停用前的全局与规则开关快照 |

旧插件没有 `plugin-support-rule` 时，仅在已经存在 `matchRules` 的情况下兼容识别为支持规则配置。

## 配置映射

| 页面作用域 | WasmPlugin 字段 |
|------------|-----------------|
| 全局配置 | `spec.defaultConfig`、`spec.defaultConfigDisable` |
| 规则配置 | `spec.matchRules[].config`、`spec.matchRules[].configDisable` |
| 规则目标 | `spec.matchRules[].ingress = ["{namespace}/{ingressName}"]` |

域名“更多”必须同时满足 `plugin-enabled != false` 和支持规则配置才展示。

## MicroApp 加载

制品安装的插件通过 `metadata.labels["w7.cc/group-name"]` 自动关联同组 MicroApp。前端从 `default` 命名空间读取 MicroApp 列表，只有同组唯一匹配时才建立关联；找不到或同组存在多个 MicroApp 时回退 YAML。没有 `w7.cc/group-name` 的旧资源继续兼容 `w7.cc/plugin-microapp` 显式关联。

找到关联资源后，通过 `/panel-api/v1/microapp/{name}/info` 获取按当前用户角色过滤的资源信息，并沿用 MicroApp 静态资源状态、下载、`frontprops`、Wujie/iframe 加载流程。

是否配置了插件前端包，以 MicroApp 当前作用域可用的 `spec.bindings[].menu` 是否包含菜单项为准，不能只根据 `frontendUrl` 或 `url` 判断。全局配置检查允许展示的创始人端/普通用户端菜单；规则配置只检查 `normal` binding。没有对应菜单时直接回退 YAML。

插件前端会收到以下 props：

```js
{
  configScope: 'global' | 'rule',
  pluginId,
  namespace,
  ingressName,
  domain,
  path,
  pluginConfig,
  pluginConfigEnabled,
  savePluginConfig(config, enabled)
}
```

- 全局配置：创始人可见创始人端和普通用户端菜单，普通用户只见自身菜单，菜单竖排。
- 规则配置：只读取 `normal` binding，菜单横排。
- MicroApp 启动前会在浏览器控制台输出注入的配置 props，覆盖全局和规则作用域；认证 Token 和 Authorization 不写入日志。
- Wujie 实例名必须包含插件、作用域和 Ingress，避免多个配置抽屉相互覆盖。
- MicroApp 不可用时必须回退 YAML，不阻断插件配置。
