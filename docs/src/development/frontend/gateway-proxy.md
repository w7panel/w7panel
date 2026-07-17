# 网关代理（反向代理）前端实现

前端菜单中的正式名称是“反向代理”，路由为 `/gateway/rvproxy`。它用于把 Higress 的外部访问入口转发到已有域名或 IP 服务；不要与面向大模型协议的 [AI 代理](./ai-proxy.md) 或扩展网关行为的 [网关插件](./gateway-plugins.md) 混用。

## 代码入口

| 位置 | 职责 |
|------|------|
| `src/router/routes/modules/gateway.ts` | 注册反向代理列表和域名详情路由 |
| `src/views/app/rvproxy/rvproxy.vue` | McpBridge 列表、上游编辑、关联域名和删除 |
| `src/views/app/rvproxy/domain.vue` | 反向代理域名页容器 |
| `src/views/app/pages/domain.vue` | 复用应用域名管理页面 |
| `src/components/domain-strategy.vue` | 重写、跨域、重试等域名策略 |
| `src/components/domain-gray-release.vue` | 反向代理场景的多目标和灰度配置 |

页面权限键为 `gateway/rvproxy/add`、`gateway/rvproxy/edit` 和 `gateway/rvproxy/delete`。

## 资源模型

每个反向代理对象对应当前业务命名空间中的一个 `networking.higress.io/v1` `McpBridge`：

- 一个 registry 代表一个上游分组；
- `type=dns` 使用目标域名、端口和 HTTP/HTTPS 协议；
- `type=static` 使用一个或多个 `IP:端口`；
- 多个目标地址以逗号写入 registry 的 `domain`，由网关进行负载均衡；
- 名为 `default` 的 McpBridge 不在反向代理列表展示，它由其他网关能力共享。

创建反向代理时，页面还会为每个填写了“访问域名”的 registry 创建 Ingress。Ingress 后端引用当前 McpBridge，并通过 `higress.io/destination=<registry-name>.<type>` 选择上游。启用证书时使用 `w7-letsencrypt-prod` ClusterIssuer。

## 编辑与删除一致性

registry 名称由“标识 + 代理对象后缀”组成。编辑标识或类型时，必须同步修改关联 Ingress 的 `destination` 标签和注解；删除 registry 时必须清理其关联 Ingress。

删除整个反向代理对象时，应先删除引用其中 registry 的 Ingress，再删除 McpBridge。新增、修改删除逻辑时必须同时检查 `higress.io/destination` 和兼容字段 `destination`，避免留下仍指向旧上游的路由。

## 与 AI 代理的区别

| 能力 | 反向代理 | AI 代理 |
|------|----------|---------|
| 主要用途 | 通用 HTTP/HTTPS 服务转发 | 大模型 API 统一入口 |
| 上游模型 | McpBridge registry | AI Provider + McpBridge registry |
| 域名入口 | Ingress | Ingress |
| 内置认证/模型白名单 | 无 | Key Auth、请求体模型枚举 |
| 多上游 | registry 内地址负载均衡、域名灰度 | Provider 权重分流 |

## 开发检查

联调时至少覆盖 DNS 与 static 两种上游、HTTP 与 HTTPS、多个上游地址、自动证书、关联域名编辑、registry 删除及整个代理删除。修改页面后执行：

```bash
cd w7panel-ui
npm run build
```
