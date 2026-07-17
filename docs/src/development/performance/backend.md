# 后端性能规范

本文档从历史后端性能报告中提炼长期开发规范，适用于 `w7panel-server/app`、`common/service`、`common/middleware` 和 WebDAV、K8s、认证相关能力。

## 整体要求

后端接口开发时先判断接口是否属于高频路径。以下场景默认按高频路径处理：

| 场景 | 风险 | 要求 |
|------|------|------|
| 认证中间件 | 每个 API 请求都会经过 | 避免每次请求都产生外部 K8s 调用 |
| K8s 客户端和 token 解析 | 多数业务接口会依赖 | 复用解析结果和客户端实例，设置过期策略 |
| WebDAV 文件读写 | 文件大小和目录规模不稳定 | 使用流式处理、大小限制和特殊文件处理 |
| 列表接口 | K8s 对象可能很大 | 分页、limit、字段裁剪或明确数据规模 |
| 压缩、复制、删除 | 可能触发长任务和高 I/O | 校验路径、限制并发、设置超时 |

## 接口设计预算

新增接口或改动高频接口时，必须在实现里体现以下边界：

| 项目 | 要求 |
|------|------|
| 超时 | 外部 HTTP、K8s API、exec、proxy 默认参考 10 秒 |
| 列表 | 必须支持分页、limit、字段裁剪，或说明不会超过固定规模 |
| 文件 | 单文件编辑大小上限 50MB，目录条目上限 5000 |
| 并发 | 批量操作必须显式限制并发数 |
| 缓存 | 跨请求缓存必须有 TTL、容量上限和失效策略 |
| 日志 | 高频路径不输出完整请求体、响应体、token 和大对象 |

接口设计时优先返回前端直接需要的业务字段。只有 YAML 编辑、K8s proxy、协议兼容接口等明确场景才返回原始对象。

## WebDAV 和文件处理

文件接口必须遵守以下规则：

- 不要在检查大小前用 `io.ReadAll` 全量读取文件。
- 读取文件内容时优先使用流式处理；需要缓存时必须明确最大大小。
- 大文件检查使用 `io.LimitedReader` 或等价机制，达到上限后立即停止读取。
- `device`、`fifo`、`socket`、`/proc`、`/sys`、`/dev` 等特殊文件系统必须单独处理。
- 目录列表必须遵守最大条目限制，当前标准为 5000 条。
- WebDAV 标准方法保持协议语义，`PROPFIND` 返回 XML，不包装成普通 JSON。
- `LockSystem`、可复用 Handler 状态和昂贵对象不要在每个请求里重复创建，除非有明确隔离需求。

### 读取策略

| 操作 | 推荐方式 | 说明 |
|------|----------|------|
| 下载文件 | 流式响应 | 不把完整文件放入内存 |
| 编辑小文件 | 限制大小后读取 | 超过 50MB 直接拒绝或提示 |
| 目录列表 | 迭代读取并限制条目 | 超过 5000 条返回截断或错误 |
| 压缩/解压 | 后台任务或受控同步 | 校验路径和格式，限制并发 |
| 特殊文件 | 短路处理 | device、fifo、socket 返回空内容或明确错误 |

示例：

```go
limited := &io.LimitedReader{
    R: file,
    N: maxSize + 1,
}
content, err := io.ReadAll(limited)
if err != nil {
    return nil, err
}
if int64(len(content)) > maxSize {
    return nil, fmt.Errorf("file too large")
}
```

反例：

```go
content, err := io.ReadAll(file)
if len(content) > MaxFileSize {
    return nil, fmt.Errorf("file too large")
}
```

上面的写法会先把超限文件读入内存，再判断大小。对于文件管理和编辑器路径，必须先限制再读取。

## 认证和 Token

认证路径不能把外部校验放在无缓存的每次请求中。

约定：

- TokenReview、JWT 解析、K3k audience 解析等结果应按 token 或 token 指纹缓存。
- 缓存必须有 TTL 或基于 token 过期时间的失效策略。
- 缓存 key 不保存完整敏感 token；需要记录时使用哈希或短指纹。
- token、密码、密钥、OIDC code 不写入日志、URL、响应体或前端可见字段。

缓存设计至少说明：

| 字段 | 要求 |
|------|------|
| key | token 指纹、用户标识或上下文标识 |
| ttl | 固定 TTL 或 token 过期时间 |
| max size | 防止无限增长 |
| invalidation | 过期、登出、刷新 token、权限变化时的处理 |
| concurrency | 并发读写必须安全 |

### Token 缓存规则

| 内容 | 可以缓存 | 失效条件 |
|------|----------|----------|
| TokenReview 成功结果 | 可以 | token 过期、TTL 到期、用户登出 |
| JWT audience/claims 解析 | 可以 | token 变化 |
| K3k config 解析 | 可以 | token 变化、配置版本变化 |
| 用户权限判断结果 | 谨慎 | 权限变更、用户组变更、TTL 到期 |

实现建议：

- 使用 token hash 作为缓存 key，不使用完整 token。
- 使用 `singleflight` 或等价机制合并同一 token 的并发校验。
- 失败结果只短时间缓存，避免临时网络故障导致长时间不可用。
- 鉴权缓存只减少重复校验，不改变授权语义。

## K8s 客户端和 Exec

K8s API 和 kubelet exec 是高成本路径。新增或修改相关功能时必须检查调用频率。

要求：

- 不在循环内重复创建 K8s client、REST client 或 transport。
- 对同一请求内多次使用的 token、namespace、cluster config 和 SDK 实例做局部复用。
- 对跨请求复用的 SDK/token 缓存设置 TTL 和最大容量。
- 高频接口避免每次都通过 kubelet 10250 exec 获取信息；能使用本地缓存、Pod annotation 或已有状态时优先复用。
- 对 exec、proxy、外部 HTTP 请求设置超时，默认参考 10 秒。
- 批量操作需要控制并发，不要无上限启动 goroutine。

PID、容器 ID、Agent Pod 等缓存需要同时考虑：

| 场景 | 处理 |
|------|------|
| Pod 重建 | containerID 变化后缓存失效 |
| 节点变化 | hostIP 或 nodeName 变化后缓存失效 |
| 权限变化 | 使用当前请求 token 重新确认权限 |
| 缓存未命中 | 只在必要时执行高成本查询 |

### Client 复用

| 对象 | 建议生命周期 | 注意事项 |
|------|--------------|----------|
| HTTP transport | 进程级复用 | 配置连接池、超时和 TLS |
| K8s REST config | token 或集群级复用 | 不泄漏用户 token |
| K8s SDK/client | token 指纹 + 集群维度缓存 | TTL 和最大容量必需 |
| namespace、cluster info | 请求级或短 TTL 缓存 | namespace 切换时刷新 |
| Pod/Container PID | containerID + hostIP 维度缓存 | Pod 重建后失效 |

### Exec 使用边界

只有在没有更低成本来源时才使用 kubelet exec。优先级从高到低：

1. 请求上下文、前端传入或业务状态中已有的数据。
2. 本地内存缓存，确认 key 和失效条件可靠。
3. Pod annotation、K8s 对象字段或 informer/缓存。
4. K8s API 查询。
5. kubelet exec 或 `crictl inspect`。

进入第 5 步时，需要记录为什么不能使用前四种方式，并限制并发。

## 响应和列表

- 公开接口和面板业务接口只返回业务需要字段，不直接透出完整 K8s 对象。
- 列表接口需要说明排序、分页、过滤和默认 limit。
- 大对象响应应裁剪 `managedFields`、无用 metadata 和重复 status 字段。
- 返回 YAML 编辑原始对象时，清理会导致误提交的系统字段。
- 批量接口返回失败项，不要只返回笼统错误。

列表接口建议响应结构：

```json
{
  "items": [],
  "total": 0,
  "limit": 50,
  "continue": "",
  "truncated": false
}
```

如果底层 K8s API 暂不支持总数，也需要返回 `limit`、`continue` 或 `truncated` 中至少一个边界字段，避免前端误以为拿到了全量数据。

字段裁剪建议：

| 类型 | 默认返回 | 避免默认返回 |
|------|----------|--------------|
| K8s 对象列表 | name、namespace、status、age、labels 摘要 | managedFields、完整 events、完整 spec |
| Pod 列表 | phase、restartCount、node、containers 摘要 | 完整 container status 原始对象 |
| 文件列表 | name、type、size、mtime、mode | 文件内容、完整 stat 扩展字段 |
| 监控数据 | 当前值、单位、时间戳 | 未使用的历史序列 |

## 日志和并发

- 统一使用 `log/slog` 键值对日志。
- 不在高频路径使用 `fmt.Printf` 输出请求详情。
- 调试日志需要级别控制，避免生产环境输出大量请求 URL、响应体或敏感参数。
- 全局 map、缓存、license 状态、用户组缓存等共享状态必须有锁、`sync.Map` 或单线程拥有者。
- goroutine、文件句柄、HTTP body、WebSocket 和子进程必须明确释放。

### 并发控制模式

批量处理使用固定并发，不使用无上限 goroutine：

```go
sem := make(chan struct{}, 8)
for _, item := range items {
    item := item
    sem <- struct{}{}
    go func() {
        defer func() { <-sem }()
        handle(item)
    }()
}
```

涉及请求生命周期时优先传递 `context.Context`：

```go
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
```

### 日志规则

| 日志内容 | 规则 |
|----------|------|
| 请求路径 | 可以记录方法、路径模板、namespace、资源名 |
| token、密码、密钥 | 禁止记录完整值 |
| 大响应体 | 禁止默认记录 |
| 高频成功日志 | 使用 debug 或采样 |
| 批量失败 | 记录失败数量和关键标识，不输出全部对象 |

## 反模式清单

| 反模式 | 风险 | 替代方案 |
|--------|------|----------|
| Controller 中循环调用 K8s API | 请求时间线性增长 | 批量查询、缓存或并发限制 |
| 每个请求创建 client/transport | 连接池失效、CPU 高 | 复用 client factory |
| 缓存无 TTL、无容量 | 内存无限增长 | TTL + max size + 清理任务 |
| 全量返回 K8s 对象 | 响应大、前端慢、泄漏字段 | 字段裁剪和专题响应 |
| 忽略 `resp.Body.Close()` | 连接泄漏 | `defer resp.Body.Close()` |
| `go func()` 不受控 | goroutine 泄漏 | context + semaphore |
| 生产路径 `fmt.Printf` | 日志噪声、阻塞 I/O | `slog` + 级别控制 |

## 验证方式

提交后端性能相关改动时至少执行：

```bash
go test ./...
rg "fmt\\.Print|io\\.ReadAll|TokenReview|NewMemLS|NewK8sClient" w7panel-server common app
```

涉及 WebDAV、认证、K8s exec 或缓存时补充验证：

- 大文件边界：小于、等于、大于 50MB。
- 目录边界：空目录、普通目录、超过 5000 条目录。
- 特殊文件：`/proc`、`/sys`、`/dev`、socket、fifo。
- 缓存命中和失效：token 变化、Pod 重建、containerID 变化。
- 并发请求：确认没有 data race、无上限 goroutine、重复外部调用。

## 评审清单

提交后端性能相关 PR 前自查：

| 检查项 | 通过标准 |
|--------|----------|
| 调用次数 | 同一请求内没有不必要重复 K8s/API/exec 调用 |
| 数据边界 | 文件、目录、列表和响应体有大小或数量限制 |
| 缓存 | key、TTL、容量、失效、并发保护都明确 |
| 超时 | 外部调用、exec、proxy、长任务都有 timeout/context |
| 资源释放 | 文件、HTTP body、连接、goroutine、子进程都可释放 |
| 安全 | 性能优化没有绕过用户鉴权和 namespace 权限 |
| 文档 | 新增接口或边界变化同步更新 API/性能文档 |
