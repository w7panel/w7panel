# API 开发文档

本文档是 `w7panel-server/app` 业务 API 的文档入口和通用开发约定。新增或修改接口时，应先确认路由归属、鉴权方式、响应格式、前端同步范围和文档更新位置。

## 文档目录

| 文档 | 对应模块 | 说明 |
|------|----------|------|
| [application.md](./application.md) | `w7panel-server/app/application` | OpenAPI、命名空间、Helm、配置、文件、应用和面板核心能力 |
| [auth.md](./auth.md) | `w7panel-server/app/auth` | 登录、注册、刷新 token、用户信息、OIDC 等认证能力 |
| [k3k.md](./k3k.md) | `w7panel-server/app/k3k` | K3k 用户、集群初始化、集群状态和权限信息 |
| [k3s-registry.md](./k3s-registry.md) | `w7panel-server/app/k3s-registry` | Docker Registry v2 兼容入口、镜像仓库管理和 containerd 操作代理 |
| [metrics.md](./metrics.md) | `w7panel-server/app/metrics` | CPU、内存、磁盘使用量和 metrics 组件安装状态 |
| [zpk.md](./zpk.md) | `w7panel-server/app/zpk` | ZPK 配置、列表、安装、升级、构建和应用管理 |

## 模块结构

后端源码按业务模块拆分：

```text
w7panel-server/app/{module}/
├── http/          # Controller，请求绑定、参数校验、响应转换
├── provider.go   # 路由注册和模块初始化
└── ...           # 模块内部服务、类型和辅助逻辑
```

公共能力放在：

| 目录 | 说明 |
|------|------|
| `w7panel-server/common/service/` | 跨模块业务服务 |
| `w7panel-server/common/middleware/` | 中间件 |
| `w7panel-server/common/` | 公共类型、工具、客户端封装 |

约定：

- Controller 只处理 HTTP 层工作：参数读取、校验、调用服务、转换响应。
- 可复用业务逻辑放入 service，不在 Controller 中堆叠跨模块流程。
- K8s 资源对象转换成前端需要的业务字段后再返回，公开接口尤其不能直接透出完整资源对象。

## 路由分层

| 类型 | 前缀 | 用途 |
|------|------|------|
| 面板业务 API | `/panel-api/v1/` | Helm、ZPK、认证、配置、文件、权限、面板业务聚合 |
| K8s 代理 | `/k8s-proxy/` | 仅代理 Kubernetes 原生 API，例如 `/api/v1/*`、`/apis/*` |
| 未授权公开 API | `/panel-api/v1/noauth/*` | 只返回允许公开的业务字段 |

禁止事项：

- 不要把面板业务 API 放到 `/k8s-proxy/` 下。
- 不要在 `/k8s-proxy/` 下新增非 Kubernetes API 路由。
- 不要创建 `/panel-api/v1/v1/*` 这种重复版本前缀。
- 未授权接口不要返回完整 K8s 资源对象。

推荐命名：

| 场景 | 路由示例 | 说明 |
|------|----------|------|
| 模块列表 | `GET /panel-api/v1/zpk/list` | 面板聚合数据 |
| 模块详情 | `GET /panel-api/v1/zpk/detail` | 用 query 传标识 |
| 创建任务 | `POST /panel-api/v1/zpk/buildimage/job` | 明确任务对象 |
| K8s 原生资源 | `/k8s-proxy/apis/apps/v1/namespaces/{ns}/deployments` | 保持 K8s API 路径 |

## 鉴权和安全

除明确标注“无需鉴权”的接口外，请求头必须携带：

```http
Authorization: Bearer <token>
```

约定：

- `LOCAL_MOCK=true` 只改变 K8s 调用方式，不改变用户认证逻辑。
- 公开接口必须使用 `/panel-api/v1/noauth/*` 前缀，并只返回业务允许公开的字段。
- 需要调用 K8s API 的接口必须明确 token 来源：用户 token、ServiceAccount token、webdavToken 或自定义 token。
- 不在日志、URL、响应或前端可见字段中输出完整 token、密码、密钥、OIDC code、registry password。

## 请求参数

参数来源按 HTTP 语义选择：

| 场景 | 推荐方式 | 说明 |
|------|----------|------|
| 查询、过滤、分页 | query/form tag | `GET` 请求优先使用 query |
| 创建、更新复杂对象 | JSON body | `POST`、`PUT`、`PATCH` 使用结构化 JSON |
| 文件上传 | multipart/form-data | 明确文件字段名和大小限制 |
| 表单兼容接口 | form tag | 仅在历史兼容或简单提交时使用 |

说明：

- `form` 表示 Controller 使用 `form` tag 绑定，通常支持 query、form-urlencoded 或 multipart form，具体取决于 HTTP 方法和 Content-Type。
- `json` 表示 JSON body。
- namespace、name、kind、podName、containerName 等 K8s 标识必须校验为空情况。
- 用户输入路径必须清理和限制访问范围，文件相关接口要处理符号链接、特殊文件和目录规模。
- 布尔值不要混用字符串和 boolean，确需兼容时在 Controller 层统一转换。
- 大对象列表必须考虑分页、limit 或字段裁剪，避免直接返回超大 K8s 对象。

## 响应格式

`w7panel-server/main.go` 覆盖了框架默认成功响应处理器，成功响应直接返回业务数据。

普通成功：

```json
"success"
```

对象成功：

```json
{
  "field": "value"
}
```

错误响应仍使用框架默认格式：

```json
{
  "error": "错误信息",
  "code": 500
}
```

响应规范：

- 字段名保持稳定，修改字段含义时必须同步前端、文档和 changelog。
- 公开接口返回业务字段，不返回完整 K8s `metadata`、`status`、`managedFields`。
- WebDAV、Docker Registry v2、OIDC 等标准协议接口必须遵守各自协议响应，不要强行包成普通 JSON。
- 列表接口应说明排序、分页、过滤规则；没有分页时说明数据规模边界。

## 错误处理和日志

日志统一使用 `log/slog` 键值对格式：

```go
slog.Info("create app success", "namespace", namespace, "name", name)
slog.Warn("get pod failed", "namespace", namespace, "pod", podName, "error", err)
```

约定：

- 用户可理解的错误信息返回给前端，内部错误细节写日志。
- 不吞掉关键错误；可选依赖失败时需要明确降级行为。
- 批量操作应记录失败项，避免只返回笼统失败。
- 外部请求和 K8s 请求必须设置合理超时，默认参考 10 秒。

## K8s 资源处理

约定：

- 对 K8s 原生对象做业务封装时，优先抽取前端所需字段。
- 响应中需要保留 YAML 编辑能力时，应清理 `managedFields`、`resourceVersion`、`uid` 等系统字段。
- 修改 K8s 资源时明确使用 `PUT`、strategic merge patch 或 JSON Patch，不混用语义。
- 涉及 namespace 的接口必须使用当前用户上下文或显式参数，不允许默认误写其它 namespace。

## 文件和 WebDAV

限制：

| 限制项 | 值 |
|--------|----|
| 最大文件大小 | 50MB |
| 最大目录条目 | 5000 |
| 默认请求超时 | 10 秒 |

约定：

- `/proc`、`/sys`、`/dev`、device、fifo、socket 等特殊文件要单独处理。
- `LOCAL_MOCK=true` 时使用本地 `/host/proc` 路径，但仍必须校验用户 token。
- WebDAV 标准方法保持协议兼容，不用 JSON 替代 XML。
- 压缩和解压接口必须校验源路径、目标路径和格式。

## 前后端同步

后端有以下变化时，必须同步检查 `w7panel-ui`：

| 后端变化 | 前端检查 |
|----------|----------|
| 新增接口 | 是否需要新增 `src/api/` 或 `src/utils/api.ts` 方法 |
| 修改路由 | 搜索旧路径调用并替换 |
| 修改响应字段 | 搜索字段使用处、类型定义和 UI 展示 |
| 修改字段含义 | 检查按钮状态、权限判断、读写逻辑是否一致 |
| 修改鉴权方式 | 检查 token 获取、刷新、错误拦截 |
| 修改文件/WebDAV 字段 | 检查文件管理、Codeblitz、Wujie 微应用调用 |

常用检查：

```bash
rg "permission-agent|permissionUrl|compressUrl|webdavUrl" w7panel-ui/src
rg "/panel-api/v1|/k8s-proxy" w7panel-ui/src
rg "offline|k8soffline" .
```

## 文档维护

新增或修改接口时，同步维护对应模块文档：

```text
docs/development/api/{module}.md
```

接口文档至少包含：

- 接口功能
- 请求方法和路径
- 鉴权要求
- 请求参数说明
- 响应参数说明
- 必要的响应示例、边界限制或特殊逻辑说明

新增模块时，需要同时更新本文档的“文档目录”表格。

## 测试和验证

无特别要求时，部署测试使用测试模式：

```bash
LOCAL_MOCK=true
```

建议流程：

1. Controller 参数校验覆盖必填、非法值、边界值。
2. Service 覆盖主要成功路径和关键失败路径。
3. 涉及 K8s 的接口至少用当前命名空间做一次真实或 mock 请求验证。
4. 涉及前端调用时，确认浏览器网络请求路径、token、响应字段和错误提示。
5. 涉及 WebDAV、压缩、权限修改时，用实际文件路径验证协议和边界行为。

提交前检查：

```bash
go test ./...
rg "offline|k8soffline" .
rg "/api/v1/" w7panel-ui/src
```

`/api/v1/` 只允许作为 K8s 原生路径出现在 `/k8s-proxy/api/v1/` 场景中。
