# AI 代理

AI 代理基于 Higress，把不同厂商或自建的大模型 API 统一到一个访问域名，并支持多提供者权重分流、模型白名单和 Key Auth 认证。客户端通常只需要修改 API Base URL 和访问 Key，不需要感知后端提供者的地址和 Token。

## 适用场景

- 给 OpenAI 兼容客户端提供统一的大模型 API 地址；
- 在 DeepSeek、OpenAI、通义千问或自建 vLLM 等提供者之间切换；
- 为同一个模型配置主备或按权重分流；
- 用消费者 Key 隔离客户端，不直接暴露上游 Token；
- 限制某个域名允许请求的模型名称。

如果只是把普通网站或 HTTP 服务转发到另一个地址，请使用 [反向代理](./reverse-proxy.md)；如果要给域名增加限流、白名单等扩展能力，请使用 [网关插件](./gateway-plugins.md)。

## 使用前准备

- Higress 网关和相关 CRD 已正常安装；
- 准备一个解析到网关入口的域名；
- 准备大模型厂商 Token，或可从集群访问的 Ollama/vLLM 地址；
- 当前账号拥有 AI 代理的新增、编辑或删除权限；
- 如需自动 HTTPS，域名必须已正确解析，且集群中的 `w7-letsencrypt-prod` ClusterIssuer 可用。

## 创建域名

1. 进入 **网关管理 → AI代理**。
2. 点击 **新增**，填写访问域名并按需启用 HTTPS 证书。
3. 创建完成后，在列表中点击 **编辑** 进入该域名的配置页。

域名必须先创建，模型、认证和服务提供者均在该域名的配置页维护。

## 配置路由

域名详情页上方的 **路由配置** 只作用于当前域名：

- **模型名称**：在域名路由配置中输入该域名支持的模型名称。配置后，请求体中的 `model` 必须在列表内，否则网关返回 HTTP 403；留空表示不限制模型。
- **认证**：启用后至少要创建一个消费者。消费者区域支持添加多个消费者，并展示消费者名称和认证方式。每个消费者可以配置多个认证令牌及独立的令牌来源。

点击 **添加消费者** 后会打开配置弹窗。消费者名称由系统自动生成且不能修改；认证方式使用 Tab 展示，当前只支持 **Key Auth**。Key Auth 支持添加多个认证令牌，每个令牌都可以手动填写或随机生成。令牌来源支持：

- **Authorization: Bearer ${value}**：通过标准 `Authorization` 请求头发送，适合 OpenAI SDK 等客户端；
- **自定义 HTTP Header**：需要同时填写 Header 名称，例如 `x-api-key`；
- **查询参数**：需要同时填写参数名称，例如 `apikey`，只建议临时调试使用。

点击弹窗确认后配置立即保存，不需要再次点击页面底部的 **保存配置**。编辑消费者时可以增加、删除或轮换认证令牌，也可以修改令牌来源及对应名称。

删除消费者会立即清理对应 Secret、Key Auth 全局消费者配置和当前域名的 `allow` 引用。删除最后一个消费者时，该域名的认证会自动关闭。

## 配置服务提供者

一个域名可以创建多个 AI 服务提供者。创建和编辑表单会按照 Higress 的供应商能力显示专属配置，而不是使用统一的服务地址。服务提供者不单独限制客户端可请求的模型；模型白名单仍在域名路由配置中统一维护。

常用供应商包括 OpenAI/OpenAI 兼容服务、通义千问、DeepSeek、Azure OpenAI、Claude、智谱 AI、豆包、Gemini、Ollama、vLLM、AWS Bedrock 和 Google Vertex。不同类型会显示不同的专属字段。

- 只有启用的提供者会参与请求转发。
- 启用两个或更多提供者时，每个权重必须大于 0，且权重总和必须等于 100。
- 提供者名称只要求在当前域名内唯一；不同域名可使用相同显示名称，底层资源会自动隔离。

供应商专属配置包括：

- **OpenAI**：官方服务或自定义兼容服务；自定义 URL 支持多个同协议、同路径的静态 IP 地址；
- **通义千问**：联网搜索、OpenAI 兼容模式、文件 ID、官方/自定义域名和推理内容处理模式；
- **Azure OpenAI**：填写包含 `api-version` 的完整服务 URL；
- **Claude**：Anthropic 官方服务或自定义服务 URL，并支持 API 版本和 Claude Code 模式；
- **Ollama / vLLM**：填写主机端口或一个、多个 vLLM 服务 URL；
- **AWS Bedrock / Google Vertex**：从 Higress 区域候选中选择或搜索区域；Vertex 还支持认证 JSON、令牌刷新提前量和 Gemini 安全设置。

所有供应商都可以设置流式首包超时。启用 **Token 故障转移** 后，还需设置连续失败/成功阈值、健康检查间隔、超时和健康检查模型；模型下拉优先显示当前供应商的 Higress 预置选项，也可以输入自定义模型名。

如果 `higress-system/default` McpBridge 已配置代理服务器，提供者表单会显示 **代理服务器** 选项。选择后，该提供者访问上游模型服务的连接会经过对应代理；留空表示直接访问上游。代理服务器必须预先存在于该 McpBridge 的 `spec.proxies` 中。

选择协议时：

- **OpenAI/v1**：让客户端按 OpenAI API 格式请求，适合大部分统一接入场景；
- **原始协议**：保留对应厂商的原始请求协议，仅在客户端明确使用该厂商格式时选择。

## 完整配置示例：接入 DeepSeek

下面创建一个带消费者认证和模型限制的 OpenAI 兼容入口。

1. 在 **网关管理 → AI代理** 新增域名 `ai.example.com`，建议启用 HTTPS。
2. 点击该域名的 **编辑**，在“服务提供者”区域点击 **新增**。
3. 按下表填写提供者并保存：

| 字段 | 示例值 |
|------|--------|
| 供应商 | DeepSeek |
| 名称 | `deepseek-primary` |
| 协议 | `OpenAI/v1` |
| API Token | DeepSeek 提供的 Token |
| 服务地址 | 使用页面默认值 `https://api.deepseek.com` |
| 权重 | `100` |
| 启用 | 是 |

4. 在页面上方“路由配置”的模型名称中输入 `deepseek-chat` 并回车。
5. 开启认证并点击 **添加消费者**，保留默认的 Bearer Token 来源，记录自动生成的认证令牌后保存弹窗。
6. 点击页面底部的 **保存配置**，保存模型名称等路由设置。

注意：客户端使用的是消费者认证令牌，不是 DeepSeek 的上游 Token。

## 调用示例

### curl 对话请求

```bash
export AI_BASE_URL='https://ai.example.com/v1'
export AI_API_KEY='替换为消费者Key'

curl "$AI_BASE_URL/chat/completions" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $AI_API_KEY" \
  -d '{
    "model": "deepseek-chat",
    "messages": [
      {"role": "system", "content": "你是一个简洁的中文助手。"},
      {"role": "user", "content": "用一句话解释 Kubernetes。"}
    ]
  }'
```

如果消费者的令牌来源选择了“查询参数”，并把参数名称设为 `apikey`，可以这样调用：

```bash
curl "$AI_BASE_URL/models?apikey=$AI_API_KEY"
```

### 流式响应

```bash
curl -N "$AI_BASE_URL/chat/completions" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $AI_API_KEY" \
  -d '{
    "model": "deepseek-chat",
    "stream": true,
    "messages": [{"role": "user", "content": "写一首四行短诗"}]
  }'
```

### OpenAI Python SDK

OpenAI 兼容 SDK 默认发送 `Authorization: Bearer`，因此消费者使用默认的 Bearer Token 来源时不需要额外设置请求头：

```python
from openai import OpenAI

consumer_key = "替换为消费者Key"
client = OpenAI(
    api_key=consumer_key,
    base_url="https://ai.example.com/v1",
)

response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[{"role": "user", "content": "你好，请介绍一下自己"}],
)
print(response.choices[0].message.content)
```

### OpenAI Node.js SDK

```javascript
import OpenAI from 'openai';

const consumerKey = process.env.AI_API_KEY;
const client = new OpenAI({
  apiKey: consumerKey,
  baseURL: 'https://ai.example.com/v1',
});

const response = await client.chat.completions.create({
  model: 'deepseek-chat',
  messages: [{ role: 'user', content: '你好，请介绍一下自己' }],
});
console.log(response.choices[0].message.content);
```

## 多提供者权重示例

例如配置 `primary` 权重 70、`backup` 权重 30，并同时启用，网关会按 70/30 分流。需要注意：模型白名单只校验请求中的模型名，不会把模型绑定到某个提供者；所有参与分流的提供者都应支持客户端请求的模型名，否则部分请求可能被上游拒绝。

如果只需要主备手动切换，可把备用提供者设为停用；发生故障时先停用主提供者，再启用备用提供者，并确保启用项权重满足规则。

## 常见问题

| 现象 | 检查项 |
|------|--------|
| HTTP 401/403，提示认证失败 | 是否启用了认证；请求携带方式、Header/参数名称和认证令牌是否与消费者配置一致 |
| HTTP 403，提示模型不在允许列表 | 请求体是否包含 `model`，名称是否与路由配置完全一致 |
| 保存多个提供者失败 | 所有启用项权重是否大于 0，权重总和是否为 100 |
| 网关返回上游连接错误 | 服务地址能否从 Higress 访问；DNS、端口、协议和 TLS 是否正确 |
| SDK 请求失败但 curl 成功 | SDK 是否使用正确的 `base_url`；消费者是否选择了 SDK 默认支持的 Bearer Token 来源 |
| 只有部分请求失败 | 多个启用提供者是否都支持相同模型和请求协议 |
| HTTPS 证书未签发 | 域名是否解析到网关，ClusterIssuer 和证书控制器是否正常 |

## 安全建议

- 不要把上游 Token 当作消费者 Key 分发；
- 不要把 Key 写入代码仓库、截图、URL 或日志；
- 为不同客户端创建不同消费者，便于单独轮换和撤销；
- 定期轮换消费者 Key 和上游 Token；
- 生产环境建议启用认证和 HTTPS；
- 删除消费者或域名前，先确认使用该 Key 的客户端已经迁移。

## 删除域名

删除 AI 代理域名会同时清理该域名的：

- Ingress 路由；
- AI 服务提供者和 Higress 服务匹配规则；
- McpBridge 服务注册；
- 消费者 Secret 与 Key Auth 规则；
- 模型白名单规则。

确认删除后会打开任务检测界面，并依次执行：

1. 删除并检查消费者 Secret、Key Auth 消费者和域名授权规则；
2. 删除并检查 AI Provider、服务匹配规则、McpBridge registry 和旧版提供者 Secret；
3. 删除并检查模型校验规则和 Ingress 域名配置。

只有反查确认该阶段没有残留资源后，状态才会变为“已删除”。任务执行期间不能关闭弹窗；如果某一步失败，界面会保留错误信息，可以点击 **重试未完成任务** 继续清理，已经确认完成的步骤不会重复执行。

删除操作不可恢复，执行前应确认域名已不再承载请求。
