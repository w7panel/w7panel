# AI 代理前端实现

AI 代理是“网关管理”中的大模型转发能力。它不经过面板业务 API，前端通过 `k8sproxy` 直接维护 Kubernetes 和 Higress 资源。

## 代码入口

| 位置 | 职责 |
|------|------|
| `src/router/routes/modules/gateway.ts` | 注册 `/gateway/aiproxy` 和隐藏的域名详情路由 |
| `src/views/app/aiproxy/aiproxy.vue` | 域名列表、新增/删除域名、初始化 `ai-proxy` 插件 |
| `src/views/app/aiproxy/domain.vue` | 模型、认证、消费者和服务提供者配置 |
| `src/utils/ai-proxy.ts` | 标签、注解、作用域资源名、提供者权重校验 |

页面权限键为 `gateway/aiproxy/add`、`gateway/aiproxy/edit` 和 `gateway/aiproxy/delete`。域名详情页虽然不显示在菜单中，但沿用 AI 代理的权限键。

## 资源关系

| 配置 | Kubernetes/Higress 资源 |
|------|-------------------------|
| 访问域名 | 业务命名空间中的 Ingress |
| 域名路由元数据 | Ingress 的 `w7.cc/gateway-ai-*` 注解和 `higress.io/destination` |
| AI 提供者配置 | `higress-system/ai-proxy.internal` 内置 WasmPlugin |
| 上游服务发现 | `higress-system/default` McpBridge |
| 消费者管理 | 业务命名空间中的 Secret |
| 请求认证 | `higress-system/key-auth.internal` Higress 内置 WasmPlugin |
| 模型白名单 | `higress-system/request-validation.internal` Higress 内置 WasmPlugin |

请求链路如下：

```text
客户端 -> AI 代理域名 Ingress
       -> key-auth / request-validation（按域名启用）
       -> higress.io/destination 选择 McpBridge 提供者服务
       -> ai-proxy.internal 按 activeProviderId 转换请求 -> 实际大模型服务
```

## 域名创建流程

创建 AI 代理域名时，页面先在当前业务命名空间创建带 `w7.cc/gateway-ai-proxy=true` 的 Ingress，根路径后端指向 `McpBridge/default`，然后确保 `higress-system/ai-proxy.internal` 存在且为 2.0.0。插件初始化失败时会回滚刚创建的 Ingress。

启用 HTTPS 时，Ingress 会使用 `w7-letsencrypt-prod` ClusterIssuer 并创建域名作用域的 TLS Secret。前端显示的域名白名单来自用户信息中的 `w7.cc/domain-white-list`。

## 作用域与敏感数据

提供者 ID、消费者 ID 和 McpBridge registry 名称必须使用 `ai-proxy.ts` 的作用域函数生成，不得直接使用用户填写的显示名称。作用域包含 Ingress 名称，保证不同域名下的同名提供者和消费者不会覆盖彼此。

`ai-proxy.internal` 必须使用 2.0.0 镜像，并包含 `higress.io/resource-definer=higress`、`higress.io/wasm-plugin-name=ai-proxy`、`higress.io/wasm-plugin-version=2.0.0` 和 `higress.io/wasm-plugin-built-in=true` 标签。Higress 的服务提供者列表只从该内置全局实例读取配置，不能改用自定义名称的 WasmPlugin。

AI 代理域名以业务命名空间的 Ingress 为数据源，不创建或同步 Higress Console 的 AI Route ConfigMap。当前部署默认关闭 `higress-console`，两者不是同一个路由管理数据源。

Ingress 注解只保存非敏感路由数据：模型列表、认证开关、提供者 ID、显示名称、权重和状态。上游 Token 不得写入 Ingress 注解。

消费者 Key 保存在业务命名空间的 Secret。上游 Token 按 Higress `ai-proxy` 协议保存在 `ai-proxy.internal.spec.defaultConfig.providers[].apiTokens`，因此必须限制 WasmPlugin 的读取权限，日志和错误信息中也不得输出凭据。

编辑提供者时，Token 输入框默认不回显；留空表示保留原 Token。新增提供者时输入 Token 后即使没有按回车，保存逻辑也会把当前输入合并到 Token 列表。

## Higress 插件协议

- `ai-proxy` 只写入官方支持的 `providers`、`activeProviderId` 和各供应商字段。不要向域名规则写入自定义 `auth`、`models` 或带权重的 provider 对象。
- `key-auth` 的实例配置维护全局 consumers，域名 matchRule 使用 `allow` 关联当前域名的消费者。
- `request-validation` 的域名 matchRule 使用 JSON Schema `enum` 校验请求体 `model`；模型列表为空时删除该域名规则。
- 认证和模型校验必须复用 Higress 内置的 `key-auth.internal`、`request-validation.internal` 实例；创建或修复实例时使用 2.0.0 镜像、默认 `AUTHN` 阶段和优先级以及 `FAIL_OPEN`，避免 Wasm 下载或初始化异常导致未匹配域名返回 500；不得删除其他路由共享的配置。
- 多提供者权重通过 Ingress `higress.io/destination` 配置，两个或更多启用项的权重总和必须为 100。
- Provider 不保存支持模型；模型名称只在域名路由配置中维护，并同步到 `request-validation` 规则。

保存旧版域名配置时，页面会把旧的域名级 `ai-proxy` 自定义规则迁移到 Ingress 注解和域名作用域资源，并清理不再被其他域名引用的旧 provider。

## 提供者保存流程

保存一个提供者时，前端按以下顺序同步资源：

1. 校验名称、供应商专属字段和启用项权重。
2. 在 `ai-proxy.internal.spec.defaultConfig.providers` 中新增或更新域名作用域 Provider。
3. 为 Provider 写入只匹配其内部服务的 `activeProviderId` 规则。
4. 在 `higress-system/default` McpBridge 中新增或更新 DNS/static registry。
5. 把当前域名的 Provider 引用写入 Ingress 注解，并生成 `higress.io/destination` 分流配置。

前端当前支持 OpenAI/OpenAI 兼容服务、通义千问、DeepSeek、Azure OpenAI、Claude、智谱、豆包、Gemini、Ollama、vLLM、Bedrock、Vertex 等类型。供应商专属字段必须通过 `providerToPluginConfig` 和 `applyEndpointConfig` 转换为 Higress 官方字段，不能把表单对象直接写入插件。

例如两个启用的提供者权重为 70 和 30 时，Ingress 目标值结构类似（实际名称由作用域函数生成）：

```text
70% llm-ai-scope-example-primary.internal.dns:443
30% llm-ai-scope-example-backup.internal.dns:443
```

只有一个启用项时不写百分比。没有启用项时删除 `higress.io/destination`。

## 认证和模型限制

- 开启认证时至少要有一个消费者；消费者名和 Key 在当前域名内不能重复，Key 也不能与其他域名复用。
- 自动生成的消费者 Key 为 24 字节随机数的十六进制字符串，只在保存后的消息中显示一次。
- Key Auth 接受请求头 `x-api-key`，也接受查询参数 `apikey`。
- 模型列表为空表示不限制；非空时请求体必须包含 `model` 且值在枚举内，否则返回 HTTP 403。
- 保存路由配置时会依次同步消费者 Secret、Key Auth、模型校验、Provider 和 Ingress 状态。修改这一流程时要检查部分成功后的再次保存是否可收敛。

## 删除规则

删除域名时必须同步清理 AI Provider、service matchRule、McpBridge registry、Key Auth consumer/matchRule、模型校验 matchRule 和消费者 Secret，不能只删除 Ingress。

## 开发检查

修改 AI 代理后至少检查：

```bash
cd w7panel-ui
npm run build
```

联调时使用测试模式并重点验证：单提供者、多提供者权重、启停、Token 留空编辑、模型允许/拒绝、请求头和查询参数两种认证、不同域名同名提供者，以及删除后是否有孤儿 Provider、matchRule、registry 或 Secret。
