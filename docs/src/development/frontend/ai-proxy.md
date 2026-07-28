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

提供者使用可增删的 Token 输入列表。Bedrock 和 Vertex 使用供应商专属凭据，不显示通用 Token；Ollama、vLLM 以及 OpenAI/Claude 自定义服务允许 Token 为空。

## Higress 插件协议

- `ai-proxy` 只写入官方支持的 `providers`、`activeProviderId` 和各供应商字段。不要向域名规则写入自定义 `auth`、`models` 或带权重的 provider 对象。
- `key-auth` 的实例配置维护全局 consumers，域名 matchRule 使用 `allow` 关联当前域名的消费者。
- `request-validation` 的域名 matchRule 使用 JSON Schema `enum` 校验请求体 `model`；模型列表为空时删除该域名规则。
- 认证和模型校验必须复用 Higress 内置的 `key-auth.internal`、`request-validation.internal` 实例；两个实例是网关通用插件，标题和描述固定为“Key Auth 认证”和“请求校验”，不能添加 AI 专属前缀。创建或修复实例时使用 2.0.0 镜像、默认 `AUTHN` 阶段和优先级以及 `FAIL_OPEN`，避免 Wasm 下载或初始化异常导致未匹配域名返回 500；不得删除其他路由共享的配置。
- 多提供者权重通过 Ingress `higress.io/destination` 配置，两个或更多启用项的权重总和必须为 100。
- Provider 不保存支持模型；模型名称只在域名路由配置中维护，并同步到 `request-validation` 规则。
- ProviderForm 按 Higress Console 显示供应商专属字段。健康检查模型的预置候选来自 Higress `aiModelProviders[].targetModelList`，只用于帮助填写 `failover.healthCheckModel`，不等同于域名模型白名单。
- OpenAI/vLLM 多 URL 保存为首个 `openaiCustomUrl`/`vllmCustomUrl` 加备用 URL 数组，并同步为同一个 static registry 的逗号分隔地址；多个 URL 只允许 IP，且协议和路径必须一致。
- Token 故障转移同时写入 `failover` 和 `retryOnFailure.enabled`；Vertex 安全设置从表格数组规范化为插件协议使用的 `geminiSafetySetting` 对象。读取时兼容旧的 `geminiSafetySettings` 复数字段，并在保存时迁移为单数字段。
- Provider 的代理服务器选项来自 `higress-system/default` McpBridge 的 `spec.proxies`。选择结果写入该 Provider 对应 `spec.registries[]` 项的 `proxyName`，不写入 `ai-proxy.internal.spec.defaultConfig.providers[]`。

保存旧版域名配置时，页面会把旧的域名级 `ai-proxy` 自定义规则迁移到 Ingress 注解和域名作用域资源，并清理不再被其他域名引用的旧 provider。

## 提供者保存流程

保存一个提供者时，前端按以下顺序同步资源：

1. 校验名称、Token、供应商专属字段、多 URL 约束、故障转移、代理服务器和启用项权重。
2. 在 `ai-proxy.internal.spec.defaultConfig.providers` 中新增或更新域名作用域 Provider。
3. 为 Provider 写入只匹配其内部服务的 `activeProviderId` 规则。
4. 在 `higress-system/default` McpBridge 中新增或更新 DNS/static registry。
5. 把当前域名的 Provider 引用写入 Ingress 注解，并生成 `higress.io/destination` 分流配置。

前端当前支持 OpenAI/OpenAI 兼容服务、通义千问、DeepSeek、Azure OpenAI、Claude、智谱、豆包、Gemini、Ollama、vLLM、Bedrock、Vertex 等类型。供应商专属字段必须通过 `normalizeProviderRawConfigs`、`providerToPluginConfig` 和 `applyEndpointConfig` 转换为 Higress 官方字段，不能把表单辅助字段直接写入插件。

例如两个启用的提供者权重为 70 和 30 时，Ingress 目标值结构类似（实际名称由作用域函数生成）：

```text
70% llm-ai-scope-example-primary.internal.dns:443
30% llm-ai-scope-example-backup.internal.dns:443
```

只有一个启用项时不写百分比。没有启用项时删除 `higress.io/destination`。

## 认证和模型限制

- 开启认证时至少要有一个消费者；消费者名和认证令牌在当前域名内不能重复，认证令牌也不能与其他域名复用。
- 消费者通过列表展示，新增和编辑使用弹窗。认证方式按照 Higress Console 使用 Tab 展示，当前只展示并支持 Key Auth；OAuth2 和 JWT 在真正支持前不显示。
- 显示名称使用 `consumer-` 加 6 字节随机数的十六进制字符串自动生成，并在新增、编辑时保持只读。自动生成的认证令牌为 24 字节随机数的十六进制字符串，一个消费者可以保存多个令牌。
- Key Auth 令牌来源支持 `BEARER`、`HEADER` 和 `QUERY`。Bearer 映射为 `keys: [Authorization]` 和带 `Bearer ` 前缀的 `credentials`；Header、Query 分别映射用户填写的 Header 名称或参数名称，并写入每个 consumer 自身的 `keys`、`in_header`、`in_query`。
- Secret 使用 `values`、`source`、`tokenKey` 保存消费者配置，并保留首个令牌的 `key` 字段兼容旧数据。读取旧 Secret 或旧插件 consumer 时默认迁移为 Header 来源 `x-api-key`；同步插件时要先把仍依赖全局 `keys/in_header/in_query` 的其他消费者展开为消费者级配置，再把全局 key 收敛为 Higress Console 使用的占位值，避免破坏其他域名。
- 模型列表为空表示不限制；非空时请求体必须包含 `model` 且值在枚举内，否则返回 HTTP 403。
- 新增、编辑消费者时立即同步消费者 Secret 和 Key Auth；删除时先清理 Key Auth consumer/matchRule，再删除 Secret。删除最后一个消费者时同步关闭 Ingress 的认证注解。
- 保存路由配置时仍会依次收敛消费者 Secret、Key Auth、模型校验、Provider 和 Ingress 状态。修改这些流程时要检查部分成功后的再次操作是否可收敛。

## 删除规则

删除域名时必须同步清理 AI Provider、service matchRule、McpBridge registry、Key Auth consumer/matchRule、模型校验 matchRule 和消费者 Secret，不能只删除 Ingress。

列表页删除使用三阶段任务：

1. 消费者阶段删除 Key Auth consumer/matchRule 和消费者 Secret；
2. 服务提供者阶段删除 `ai-proxy` Provider、service matchRule、McpBridge registry 和旧版提供者 Secret；
3. 域名阶段删除模型校验 matchRule 和 Ingress。

每个阶段结束后使用轮询反查对应资源，确认无残留后才进入下一阶段。单资源 GET 只有 HTTP 404 可以解释为已删除；列表请求、网络错误或其他状态码不能通过检测。任务上下文要保留首次发现的 Provider/consumer ID，使插件已删除但 registry 等后续资源删除失败时仍可重试。已成功阶段在重试时跳过，未完成阶段的删除方法必须保持幂等。

## 开发检查

修改 AI 代理后至少检查：

```bash
cd w7panel-ui
npm run build
```

联调时使用测试模式并重点验证：单提供者、多提供者权重、启停、Token 留空编辑、模型允许/拒绝、请求头和查询参数两种认证、不同域名同名提供者，以及删除后是否有孤儿 Provider、matchRule、registry 或 Secret。
