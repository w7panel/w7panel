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

面向用户的标准安装入口是 **应用管理 → 制品市场**。插件制品应把 WasmPlugin、可选的配置前端 MicroApp 和依赖资源一起交付；安装或升级后，用户再到 **网关管理 → 网关插件** 管理启停和配置。网关插件页的“添加插件”只用于开发调试、私有镜像验证和历史资源兼容，不负责安装 MicroApp 或维护完整制品生命周期。

制品市场安装的插件通过 `metadata.labels["w7.cc/group-name"]` 自动关联同组 MicroApp。前端从 `default` 命名空间读取 MicroApp 列表，只有同组唯一匹配时才建立关联；找不到或同组存在多个 MicroApp 时回退 YAML。没有 `w7.cc/group-name` 的旧资源继续兼容 `w7.cc/plugin-microapp` 显式关联。

制品开发时应保证：

- WasmPlugin 与配置前端 MicroApp 由同一制品版本交付；
- 两类资源的 `w7.cc/group-name` 完全一致；
- 同组只有一个作为插件配置前端的 MicroApp；
- MicroApp 在需要展示的角色 binding 中配置 `menu`；
- 安装、升级和完整卸载走制品流程，避免只更新某一个资源造成版本不一致。

找到关联资源后，通过 `/panel-api/v1/microapp/{name}/info` 获取按当前用户角色过滤的资源信息，并沿用 MicroApp 静态资源状态、下载、`frontprops`、Wujie/iframe 加载流程。

是否配置了插件前端包，以 MicroApp 当前作用域可用的 `spec.bindings[].menu` 是否包含菜单项为准，不能只根据 `frontendUrl` 或 `url` 判断。全局配置检查允许展示的创始人端/普通用户端菜单；规则配置只检查 `normal` binding。没有对应菜单时直接回退 YAML。

## 配置前端 props

插件配置前端通过 `window.$wujie.props` 接收面板注入的数据和保存函数。除通用 MicroApp props 外，网关插件专用字段如下：

| 字段 | 类型 | 全局配置 | 规则配置 | 说明 |
|------|------|----------|----------|------|
| `configScope` | `'global' \| 'rule'` | `global` | `rule` | 当前配置作用域 |
| `pluginId` | `string` | 有值 | 有值 | WasmPlugin 的 `metadata.name` |
| `namespace` | `string` | 空字符串 | 当前命名空间 | 目标 Ingress 所在命名空间 |
| `ingressName` | `string` | 空字符串 | 有值 | 目标 Ingress 名称 |
| `domain` | `string` | 空字符串 | 有值 | Ingress 第一条规则的域名 |
| `path` | `string` | 空字符串 | 有值 | Ingress 第一条规则的第一条路径 |
| `pluginConfig` | `Record<string, unknown>` | `spec.defaultConfig` | `matchRules[].config` | 当前配置的深拷贝快照 |
| `pluginEnabled` | `boolean` | 全局开关 | 当前规则开关 | 推荐使用的启用状态字段 |
| `pluginConfigEnabled` | `boolean` | 同上 | 同上 | 兼容旧插件前端的字段名 |
| `microappRole` | `string` | 按当前角色加载 | `normal` | 当前插件配置前端使用的角色 |
| `savePluginConfig` | `(config, enabled?) => Promise<WasmPlugin>` | 可用 | 可用 | 保存当前作用域配置并返回更新后的 WasmPlugin |

`pluginConfig` 不包含整个 WasmPlugin，只包含当前作用域的配置对象。需要插件名称、域名或 Ingress 信息时使用对应的独立字段，不要从配置内容中反推。

### 读取 props

建议在插件前端封装统一入口，并为独立运行开发提供安全的默认值：

```ts
export type GatewayPluginProps = {
  configScope?: 'global' | 'rule';
  pluginId?: string;
  namespace?: string;
  ingressName?: string;
  domain?: string;
  path?: string;
  pluginConfig?: Record<string, unknown>;
  pluginEnabled?: boolean;
  pluginConfigEnabled?: boolean;
  microappRole?: string;
  savePluginConfig?: (
    config: Record<string, unknown>,
    enabled?: boolean,
  ) => Promise<Record<string, unknown>>;
};

export function getGatewayPluginProps(): GatewayPluginProps {
  return window.$wujie?.props || {};
}

export function clonePluginConfig() {
  const config = getGatewayPluginProps().pluginConfig || {};
  return JSON.parse(JSON.stringify(config));
}
```

微应用初始化时要允许 `window.$wujie` 暂时不存在。不要在模块顶层永久缓存 props；应在组件初始化或页面进入时读取，以便菜单路由重新挂载后拿到当前作用域数据。

### Vue 3 完整示例

下面示例读取插件配置、修改字段并调用面板保存：

```vue
<script setup lang="ts">
import { reactive, ref } from 'vue';
import { getGatewayPluginProps, clonePluginConfig } from './gateway-plugin-props';

type RateLimitConfig = {
  requests_per_second?: number;
  rejected_code?: number;
};

const panelProps = getGatewayPluginProps();
const form = reactive<RateLimitConfig>(clonePluginConfig());
const enabled = ref(
  panelProps.pluginEnabled
    ?? panelProps.pluginConfigEnabled
    ?? true,
);
const saving = ref(false);
const errorMessage = ref('');

async function save() {
  if (!panelProps.savePluginConfig) {
    errorMessage.value = '当前页面未运行在 W7Panel 网关插件容器中';
    return;
  }

  saving.value = true;
  errorMessage.value = '';
  try {
    const config = JSON.parse(JSON.stringify(form));
    await panelProps.savePluginConfig(config, enabled.value);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '保存失败';
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <form @submit.prevent="save">
    <div>作用域：{{ panelProps.configScope }}</div>
    <div v-if="panelProps.configScope === 'rule'">
      域名：{{ panelProps.domain }}{{ panelProps.path }}
    </div>
    <label>
      每秒请求数
      <input v-model.number="form.requests_per_second" type="number" min="1">
    </label>
    <label>
      <input v-model="enabled" type="checkbox">
      启用当前配置
    </label>
    <button type="submit" :disabled="saving">
      {{ saving ? '保存中…' : '保存' }}
    </button>
    <p v-if="errorMessage">{{ errorMessage }}</p>
  </form>
</template>
```

同一个配置前端可通过 `configScope` 适配两种页面：

```ts
const props = getGatewayPluginProps();

if (props.configScope === 'global') {
  // 编辑 WasmPlugin.spec.defaultConfig
} else {
  // 编辑当前 namespace/ingressName 对应的 matchRules[].config
  console.info('当前规则', props.namespace, props.ingressName, props.domain, props.path);
}
```

### 保存语义

```ts
await props.savePluginConfig(nextConfig, true);  // 保存并启用
await props.savePluginConfig(nextConfig, false); // 保存并停用
await props.savePluginConfig(nextConfig);        // enabled 省略时默认启用
```

- `config` 必须是普通对象；不要传 Vue Proxy、数组、函数、DOM 节点或循环引用。
- 保存是**整体替换当前作用域配置**，不是字段级合并。表单只编辑部分字段时，应先复制 `pluginConfig`，再覆盖目标字段，以免删除插件不认识但需要保留的配置。
- 全局作用域保存到 `spec.defaultConfig`，开关保存到 `spec.defaultConfigDisable`。
- 规则作用域保存到当前 Ingress 对应的 `spec.matchRules[].config`，开关保存到 `configDisable`；规则不存在时面板会自动创建。
- `savePluginConfig` 返回 Promise，只有 Promise resolve 后才能提示成功或离开页面；失败时应保留表单并允许重试。
- 注入的 `pluginConfig` 是启动时快照。保存成功后如需继续编辑，以本地表单为准；不要假设旧的 props 对象会自动更新。
- 插件前端不得直接调用 K8s API 修改 WasmPlugin，统一通过 `savePluginConfig`，以保留作用域和开关映射逻辑。
- 配置中如包含密钥，不得输出到控制台、埋点或错误上报。

- 全局配置：创始人可见创始人端和普通用户端菜单，普通用户只见自身菜单，菜单竖排。
- 规则配置：只读取 `normal` binding，菜单横排。
- MicroApp 启动前只在浏览器控制台输出作用域、插件、域名和启停状态等非敏感上下文；`pluginConfig`、认证 Token、Authorization 和保存函数不写入日志。
- Wujie 实例名必须包含插件、作用域和 Ingress，避免多个配置抽屉相互覆盖。
- MicroApp 不可用时必须回退 YAML，不阻断插件配置。
