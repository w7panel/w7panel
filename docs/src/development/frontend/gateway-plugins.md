# 网关插件前端接入

## 数据来源

网关插件页直接请求制品市场公开接口 `POST https://zm.w7.com/zpk-market/formula/list`。请求体固定携带 `tag: "网关插件"`，由市场服务端先按标签筛选；前端再校验 `application_type=gateway-plugin`，避免错误标签数据进入插件列表。请求读取第一页并设置足够覆盖完整市场列表的 `limit`，超时为 10 秒。

```json
{
  "page": 1,
  "limit": 500,
  "tag": "网关插件"
}
```

市场制品通过 `identify` 与 AppGroup `spec.identifie` 关联，再通过 WasmPlugin 的 `w7.cc/group-name` 找到真实插件资源。关联到 WasmPlugin 时显示已安装状态并保留配置、启停、更新和卸载能力；未关联时显示“待安装”，点击“安装”携带 `formula_url` 进入 `/app/store-install`。市场请求失败时不能影响已安装 WasmPlugin 的管理；不在市场中的历史或手工插件归入“其他”分类继续展示。

AppGroup、MicroApp API 以及通用分组、官方和删除保护元数据集中维护在 `src/utils/w7panel-resource.ts`，网关插件工具只维护 Higress WasmPlugin 协议。插件页先读取必须展示的 WasmPlugin，再根据其中实际出现的 `w7.cc/group-name` 定向读取对应 AppGroup，并用标签选择器读取同组 MicroApp；旧版无分组 MicroApp 才按同名资源做兼容 GET。禁止全量读取 AppGroup 或 MicroApp 后在浏览器筛选。

卸载属于完整应用生命周期操作：列表必须删除插件关联的 `default` 命名空间 AppGroup，由 AppGroup Controller 清理同组 WasmPlugin、MicroApp 和其他关联资源，不能直接删除某一个 WasmPlugin。同一个 AppGroup 包含多个插件时，任一插件行触发的都是整个应用卸载；没有关联 AppGroup 的历史或手工插件不显示卸载入口。

已安装状态和插件操作仍以 Higress `WasmPlugin` 为准：

```text
/k8s-proxy/apis/extensions.higress.io/v1alpha1/namespaces/higress-system/wasmplugins
```

列表、添加、修改和删除分别使用 K8s 代理的 GET、POST、PUT 和 DELETE。MicroApp 仅提供配置操作界面，不作为市场列表或安装状态的数据源。

网关插件主列表和域名管理“更多”都使用市场 `plugin_type` 分类，且只渲染非空分类：

| `plugin_type` | 分类标题 |
|---------------|----------|
| `auth` | 认证鉴权 |
| `security` | 安全防护 |
| `traffic` | 流量管控 |
| `transform` | 请求响应转换 |
| `o11y` | 可观测性 |
| `ai` | AI |
| 空值 | 其他 |
| 未知值 | 市场原始分类名 |

网关插件主列表展示市场全部网关插件；域名管理“更多”仍只展示已安装且支持规则配置的插件，并通过关联 AppGroup 的 `spec.identifie` 取得对应市场分类。没有对应市场项的已安装插件归入“其他”。

每个非空分类默认展开，分类标题显示当前插件数量，点击标题可以独立收起或展开。所有分类置于同一个 4px 圆角的扁平列表容器中，使用中性分隔线和统一高度标题行；收起标题为白底，展开标题使用主色浅背景和左侧状态线。分类标题已经承担分组层级，因此分类内表格隐藏重复表头，展开后直接显示插件数据行；全部收起时形成紧凑、连续的分类列表。折叠状态仅属于当前组件实例，不写入浏览器存储；重新进入页面或重新打开域名策略时恢复默认展开。

列表中的插件信息与域名规则插件信息统一按“标题、标识+版本、描述”展示，标题不加粗，后两项使用辅助文字样式。“全局状态”列直接映射 Higress `spec.defaultConfigDisable`；无编辑权限时开关禁用，提交期间显示加载状态。只支持规则配置的插件不显示全局开关。

## 官方标识与卸载保护

通过制品安装的官方应用不在 WasmPlugin 上重复写保护注解。前端使用 WasmPlugin 的 `metadata.labels["w7.cc/group-name"]` 关联 `default` 命名空间的同名 AppGroup，并读取 AppGroup 的注解：

```yaml
metadata:
  annotations:
    w7.cc/official-app: "true"
    w7.cc/deny-delete: "true"
```

`w7.cc/official-app` 只表示官方身份：存在该注解时，网关插件列表和域名管理“更多”列表都在插件标题右侧显示紧凑的“官方”标识，网关插件列表同时禁止编辑 WasmPlugin 元数据。是否允许卸载独立读取 `w7.cc/deny-delete`，值为 `"true"` 时隐藏卸载入口并在执行前再次阻止删除；没有该注解时允许卸载关联 AppGroup。官方身份不能隐含卸载保护，删除保护也不能用于推断官方身份。插件启停、全局配置和域名规则配置仍然可用。AppGroup 查询失败时不能降级为普通应用，避免错误开放受保护操作。文档和代码统一使用“官方应用提供的网关插件”，不使用“官方插件”。

## 插件更新

插件更新复用应用列表的 AppGroup/ZPK 更新链路。前端按 WasmPlugin 的 `w7.cc/group-name` 找到 AppGroup，以 AppGroup 的命名空间和名称调用 `/panel-api/v1/zpk/upgrade-info`。同一 AppGroup 可能包含多个 WasmPlugin，更新检测必须按 AppGroup 名称去重，并将结果共享给组内插件。

存在新版本时在插件标题右侧显示“新版本”，可查看后端返回的 `description`，并携带 AppGroup 名称作为 `releaseName`、更新结果或 AppGroup `spec.zpkUrl` 作为 `path` 进入 `/app/store-install`。更新操作升级整个应用制品，包括同组 WasmPlugin、MicroApp 和其他资源，不能只替换单个 WasmPlugin。

## Metadata 约定

扩展能力写入 `metadata.annotations`，避免修改 Higress CRD：

| Annotation | 默认值 | 说明 |
|------------|--------|------|
| `w7.cc/plugin-support-global` | `true` | 是否支持全局配置 |
| `w7.cc/plugin-support-rule` | `false` | 是否支持域名规则配置 |

旧插件没有 `plugin-support-rule` 时，仅在已经存在 `matchRules` 的情况下兼容识别为支持规则配置。

“支持规则配置”同时约束域名入口可见性。编辑插件并取消该能力时，如果存在 `configDisable != true` 的规则，保存前必须展示危险确认，说明插件会从域名“更多”中隐藏、现有启用规则将全部停用，且重新开启能力不会自动恢复规则状态。确认后将所有 `matchRules[].configDisable` 写为 `true`。

## 配置映射

| 页面作用域 | WasmPlugin 字段 |
|------------|-----------------|
| 全局配置 | `spec.defaultConfig`、`spec.defaultConfigDisable` |
| 规则配置 | `spec.matchRules[].config`、`spec.matchRules[].configDisable` |
| 规则目标 | 统一使用 `ingress = ["{namespace}/{ingressName}"]` |

全局配置和规则配置遵循 Higress 原生的独立开关语义：列表“全局状态”只修改 `defaultConfigDisable`，不能连带修改任何 `matchRules[].configDisable`。域名“更多”只要 WasmPlugin 存在且支持规则配置就展示；卸载 WasmPlugin 后自然消失。网关插件表格上方必须提示全局状态不影响域名规则；域名“更多”列表上方必须提示规则状态只影响当前域名，不影响全局配置或其他域名。

域名“更多”的插件表格只保留插件、规则状态和操作列，不单独展示“配置方式”；用户点击配置后，仍按关联 MicroApp 是否存在可用菜单决定加载操作界面或回退 YAML。

Higress 的 Disable 字段采用缺省启用语义：`defaultConfigDisable` 或已匹配规则的 `configDisable` 不存在时按 `false` 处理；只有没有匹配规则时，规则状态才显示为未启用。所有域名级插件配置只识别精确的 `namespace/ingressName` 目标，不读取或迁移旧 `domain`/裸 Ingress 目标；AI Provider 自行维护的 `service` 规则不由通用域名配置接管。修改一条包含多个 namespaced Ingress 目标的共享规则时，前端先复制其配置和状态，为当前 Ingress 建立独立规则，再保留其他目标。

YAML 模式只编辑当前作用域的配置对象，不提供启用或停用操作；保存时必须保留原 `defaultConfigDisable` 或 `matchRules[].configDisable` 状态。全局启停统一在插件列表操作，规则启停统一在域名“更多”列表操作。

删除 Ingress 前，前端遍历 `higress-system` 下的 WasmPlugin，从每条规则的 `ingress` 数组移除对应的 `namespace/name`；共享规则仍保留其他 Ingress、域名或服务目标。Ingress 重写由 Ingress annotation 独立维护，不能向任意插件配置写入通用 `rewrite_host` 字段。

## MicroApp 加载

面向用户的标准安装入口是 **应用管理 → 制品市场**。插件制品应把 WasmPlugin、可选的配置前端 MicroApp 和依赖资源一起交付；安装或升级后，用户再到 **网关管理 → 网关插件** 管理启停和配置。网关插件页的“添加插件”只用于开发调试、私有镜像验证和历史资源兼容，不负责安装 MicroApp 或维护完整制品生命周期。

制品市场安装的插件通过 `metadata.labels["w7.cc/group-name"]` 自动关联同组 MicroApp。前端从 `default` 命名空间读取 MicroApp 列表，只有同组唯一匹配时才建立关联；同组存在多个 MicroApp 时回退 YAML。旧版 MicroApp 没有分组标签时，兼容用与 WasmPlugin 分组完全同名的 `metadata.name` 精确匹配，不再读取 `w7.cc/plugin-microapp` 注解。新生成的 WasmPlugin 与 MicroApp 都必须写入相同的分组标签。

制品开发时应保证：

- WasmPlugin 与配置前端 MicroApp 由同一制品版本交付；
- 两类资源的 `w7.cc/group-name` 完全一致；
- 同组只有一个作为插件配置前端的 MicroApp；
- MicroApp 在需要展示的角色 binding 中配置 `menu`；
- 安装、升级和完整卸载走制品流程，避免只更新某一个资源造成版本不一致。

找到关联资源后，通过 `/panel-api/v1/microapp/{name}/info` 获取按当前用户角色过滤的资源信息，并沿用 MicroApp 静态资源状态、下载、`frontprops`、Wujie/iframe 加载流程。

是否配置了插件前端包，以 MicroApp 当前作用域可用的 `spec.bindings[].menu` 是否包含菜单项为准，不能只根据 `frontendUrl` 或 `url` 判断。全局配置检查允许展示的创始人端/普通用户端菜单；规则配置只检查 `normal` binding。没有对应菜单时直接回退 YAML。

无论是否关联到可用 MicroApp 配置页面，配置抽屉都使用同一份当前作用域 YAML。存在可用页面时，操作界面提供 **YAML 详情** 按钮；没有可用页面或页面加载失败时直接进入同一个 YAML 界面。两种入口都先显示只读预览，通过 **编辑** 切换为可修改状态；取消预览时，有可用页面会返回操作界面，没有可用页面则关闭配置抽屉。YAML 保存仍通过当前作用域映射更新 `defaultConfig` 或对应 `matchRules[].config`，不能用规则配置入口修改整个 WasmPlugin。

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
| `globalPluginConfig` | `Record<string, unknown>` | 当前全局配置 | 当前全局配置 | 全局配置只读快照；规则配置前端可据此判断全局值 |
| `globalPluginEnabled` | `boolean` | 全局开关 | 全局开关 | 全局启用状态，只读，不代表当前规则状态 |
| `ruleConfigs` | `Array<Record<string, unknown>>` | 原始 `spec.matchRules` 深拷贝 | 不注入 | 仅全局配置页面提供的规则只读上下文，不展开或改写规则结构 |
| `pluginConfigEnabled` | `boolean` | 同上 | 同上 | 兼容旧插件前端的字段名 |
| `microappRole` | `string` | 按当前角色加载 | `normal` | 当前插件配置前端使用的角色 |
| `savePluginConfig` | `(config, enabled?) => Promise<WasmPlugin>` | 可用 | 可用 | 保存当前作用域配置并返回更新后的 WasmPlugin |

`pluginConfig` 不包含整个 WasmPlugin，只包含当前作用域的配置对象。`globalPluginConfig` 和 `globalPluginEnabled` 是全局/规则页面都提供的只读判断上下文；`ruleConfigs` 仅在全局页面提供，同样只读，不代表保存目标。插件前端不得通过修改这些字段来保存配置。需要插件名称、域名或 Ingress 信息时使用对应的独立字段，不要从配置内容中反推。

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
  globalPluginConfig?: Record<string, unknown>;
  globalPluginEnabled?: boolean;
  ruleConfigs?: Array<Record<string, unknown>>;
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
- `savePluginConfig` 只保存打开配置抽屉时的当前作用域：全局入口不能修改规则，规则入口不能修改全局；`globalPluginConfig`、`globalPluginEnabled`、`ruleConfigs` 永远不会成为保存目标。
- `savePluginConfig` 返回 Promise，只有 Promise resolve 后才能提示成功或离开页面；失败时应保留表单并允许重试。
- 注入的 `pluginConfig` 是启动时快照。保存成功后如需继续编辑，以本地表单为准；不要假设旧的 props 对象会自动更新。
- 插件前端不得直接调用 K8s API 修改 WasmPlugin，统一通过 `savePluginConfig`，以保留作用域和开关映射逻辑。
- 配置中如包含密钥，不得输出到控制台、埋点或错误上报。

- 全局配置：创始人可见创始人端和普通用户端菜单，普通用户只见自身菜单，菜单竖排。
- 规则配置：只读取 `normal` binding，菜单横排。
- MicroApp 启动前只在浏览器控制台输出作用域、插件、域名和启停状态等非敏感上下文；`pluginConfig`、认证 Token、Authorization 和保存函数不写入日志。
- Wujie 实例名必须包含插件、作用域和 Ingress，避免多个配置抽屉相互覆盖。
- MicroApp 不可用时必须回退 YAML，不阻断插件配置。
