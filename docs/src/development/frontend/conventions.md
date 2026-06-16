# 前端开发约定

本文档说明 `w7panel-ui` 的目录结构、启动入口、API 调用、页面开发、UI 交互、状态缓存、性能和提交流程规范。鉴权和登录态见 [auth-state.md](./auth-state.md)，微应用接入见 [microapps.md](./microapps.md)，组件明细见 [components.md](./components.md)，Wujie 微前端事件见 [wujie-events.md](./wujie-events.md)。

## 整体使用方式

前端开发按“页面入口 -> API 封装 -> 状态缓存 -> 组件复用 -> 文档同步”的顺序处理。页面只负责组织交互和展示，复杂请求路径、公共状态和复用组件应放到已有封装里。

### 基本流程

1. 确认页面所在模块和路由，优先复用已有 views、components、hooks 和 API 方法。
2. 调用后端时优先使用 `src/utils/api.ts` 的 `panelApi`、`k8sproxy` 或 `src/api/` 方法。
3. 登录态、用户信息、权限、K8s 信息按 [auth-state.md](./auth-state.md) 统一处理。
4. 公共 UI 放入 `src/components/`，新增 Props、Events、Ref 方法后更新 [components.md](./components.md)。
5. 微应用交互按 [microapps.md](./microapps.md) 接入，新增事件后更新 [wujie-events.md](./wujie-events.md)。
6. 修改后端 API 路径、字段或鉴权方式时，同步检查前端调用、类型、页面展示和 API 文档。

### 场景选择

| 场景 | 推荐位置 | 说明 |
|------|----------|------|
| 面板业务请求 | `src/utils/api.ts`、`src/api/` | 使用 `/panel-api/v1` 前缀 |
| Kubernetes 原生资源 | `k8sproxy` | 使用 `/k8s-proxy` 前缀 |
| 登录态和权限 | [auth-state.md](./auth-state.md) | 统一处理 token、userinfo、权限等 |
| 跨页面复用逻辑 | `src/hooks/` | Composition API 复用逻辑 |
| 公共 UI | `src/components/` | 需要记录组件外部契约 |
| 微应用桥接 | [microapps.md](./microapps.md) | 微应用容器、props、token 边界和事件入口 |
| 页面私有能力 | `src/views/{module}/` | 不跨页面复用时放页面目录 |

### 使用边界

- 页面内不要散落复杂 API 路径，尤其不要手写生产 Agent 代理路径。
- 不新增 `offline-*`、`k8soffline-*` 等旧 localStorage/sessionStorage key，具体 key 约定见 [auth-state.md](./auth-state.md)。
- token、密码、密钥、OIDC code 不要输出到 console、URL 或长期缓存。
- 终端、日志、轮询、Wujie 事件必须在组件卸载、弹窗关闭或 deactivated 时清理。
- 前端文档只记录前端契约；后端接口请求响应字段以 [../api/](../api/) 下专题文档为准。

## 目录结构

```text
w7panel-ui/src/
├── api/          # axios 拦截器和少量业务 API 方法
├── components/   # 公共组件和业务复用组件
├── config/       # 全局配置，例如 API 路径配置
├── hooks/        # Vue Composition API 复用逻辑
├── layout/       # 页面布局
├── locale/       # i18n 文案
├── router/       # 路由配置和应用菜单
├── store/        # Pinia 状态管理
├── types/        # 全局类型
├── utils/        # 工具方法、认证、本地缓存、事件工具
└── views/        # 页面模块
```

启动入口：

| 文件 | 说明 |
|------|------|
| `src/main.ts` | 创建 Vue 应用，注册 Arco、Pinia、Vue Router、i18n、全局组件、VMdPreview、GoCaptcha，并加载 axios 拦截器 |
| `src/App.vue` | 应用根组件 |
| `src/api/interceptor.ts` | axios 全局超时、token 注入、GET 防重、401 刷新 token 和错误提示 |
| `src/utils/api.ts` | `panelApi`、`k8sproxy` 路径前缀封装 |
| `src/utils/auth.ts` | token、refresh token、权限、用户信息、K8s 信息和文件/终端权限缓存，详见 [auth-state.md](./auth-state.md) |

## API 调用规范

路由前缀：

| 类型 | 前缀 | 用途 |
|------|------|------|
| 面板业务 API | `/panel-api/v1/` | 面板聚合能力、认证、ZPK、文件、Helm 等 |
| K8s 代理 API | `/k8s-proxy/` | Kubernetes 原生 API 代理 |

约定：

- `panelApi` 和 `k8sproxy` 位于 `src/utils/api.ts`，只负责拼接前缀，仍使用全局 axios 拦截器。
- 面板业务接口不要写到 `/k8s-proxy/`。
- K8s 原生资源请求使用 `/k8s-proxy/api/v1/*` 或 `/k8s-proxy/apis/*`。
- 新增后端接口时，同步新增或更新前端 API 方法，避免页面直接拼接复杂路径。
- URL 中的 namespace、name、path、labelSelector 必须使用 `encodeURIComponent` 或 axios `params`。
- 文件上传、WebDAV、压缩下载等接口要明确 token 来源和 Content-Type。
- GET 请求会按 `url + params` 做防重并共享同一个 Promise；流式请求、手动取消请求不走防重逻辑。
- 401 时会取消待处理请求，尝试调用 `/panel-api/v1/auth/refresh-token2` 刷新 token；刷新失败后清理登录态并跳转登录页。

示例：

```ts
import { panelApi, k8sproxy } from '@/utils/api';

export function getConsoleInfo() {
  return panelApi.get('/auth/console/info');
}

export function getDeployments(namespace: string) {
  return k8sproxy.get(`/apis/apps/v1/namespaces/${namespace}/deployments`);
}
```

## 组件开发规范

组件放置：

| 场景 | 位置 |
|------|------|
| 跨页面复用 | `src/components/` |
| 单页面私有 | `src/views/{module}/` |
| 布局级组件 | `src/layout/` 或 `src/components/{layout}/` |
| 微前端弹窗集合 | `src/components/wujie-modals.vue` |

约定：

- 组件使用明确的 props 和 events，不依赖隐式全局变量。
- 抽屉、弹窗统一用 `show` 或 `visible` 控制显示，关闭后触发 `close` 或 `cancel`。
- 表单类组件提交时触发 `submit`，参数使用对象。
- 需要父组件通过 `ref` 调用的方法，必须在 [components.md](./components.md) 记录方法名、参数和返回值。
- 终端、日志、WebShell 类组件必须在卸载或关闭时清理连接、请求和 xterm 实例。

## 页面开发规范

- 页面初始化请求优先并行化，互不依赖的接口不要串行等待。
- 列表页保留 loading、empty、error 或 noAlert 降级策略。
- 删除、重启、覆盖等危险操作必须有确认流程。
- 表单提交前做前端必填和格式校验，后端错误要展示可理解信息。
- namespace 默认使用 store 中当前 namespace，不允许出现 `namespace=undefined`。
- 路由 query 中的外部 URL、文件路径、应用路径必须编码和解码成对处理。

## UI 和交互规范

项目使用 Vue 3、TypeScript、Arco Design、Pinia、Vue Router、Wujie、ECharts、xterm、CodeMirror 和 VMdPreview。

约定：

- 优先使用 Arco 现有组件，不重复造基础控件。
- 图标按钮使用 Arco 或项目已有图标，复杂操作提供 tooltip 或明确按钮文案。
- 表格、表单、抽屉和弹窗布局保持密度一致，避免营销式大卡片布局。
- 按钮文案使用动词，危险动作使用 `status="danger"` 或确认弹窗。
- 文本必须在移动端和桌面端都不溢出、不重叠；长路径、镜像名、域名使用换行或省略。
- 组件内不要用大段说明文字解释功能，必要说明放到文档或 tooltip。

## 微应用和 Wujie 规范

微应用容器、props、token 边界和接入流程见 [microapps.md](./microapps.md)。事件调用统一使用：

```ts
window.$wujie.bus.$emit('eventName', payload, callback);
window.$wujie.bus.$on('eventName', handler);
```

约定：

- 面板侧注册事件使用 `registerWujieEvent()`，组件卸载时调用 `clearAllWujieEvents()`。
- 微应用侧发送事件时使用对象参数，除非只有一个简单字符串。
- 需要响应时使用 callback；有失败路径时明确是否使用 rejectCallback。
- 新增事件必须更新 [wujie-events.md](./wujie-events.md)。
- 避免通用事件名冲突；`submit`、`close` 只能在明确生命周期和嵌套场景中使用。

主要注册位置：

| 位置 | 说明 |
|------|------|
| `src/components/wujie-modals.vue` | 微应用通用弹窗、文件、日志、构建、域名、OIDC、登录切换等事件 |
| `src/components/domain-strategy-filecache.vue` | 文件缓存微应用 `submit`、`close` |
| `src/components/domain-strategy-imagecache.vue` | 镜像缓存微应用 `submit`、`close` |
| `src/views/app/gpustack/index.vue` | GPUStack 专用事件 |

## 状态和缓存

- 全局状态放 Pinia store，页面局部状态放组件自身。
- 可跨页面复用的异步状态、轮询、响应式窗口状态放 `src/hooks/`。
- 轮询和定时器必须在组件卸载、keep-alive deactivated 或弹窗关闭时清理。
- 缓存 key 必须有模块前缀，缓存失效要能按模块清理。
- 不把后端响应对象直接长期保存在多个状态源中，避免更新不同步。

## 文件、日志和终端能力

- 文件管理依赖后端返回的 `webdavUrl`、`compressUrl`、`permissionUrl` 和 `webdavToken`。
- WebDAV 请求保持协议方法语义：`GET`、`PROPFIND`、`PUT`、`DELETE`、`MOVE`、`COPY`。
- 上传文件必须设置正确 Content-Type，并处理失败回调。
- 日志和终端组件关闭时必须释放连接。
- 前端不要自行拼接生产模式 agent 代理路径，应使用后端返回字段。

## 性能规范

- 页面首屏请求能并行则并行。
- 大列表使用分页、筛选或后端 limit，避免一次渲染过多节点。
- ECharts 组件使用按需注册的图表类型，新增类型同步 `src/components/index.ts`。
- 大 YAML、日志、文件内容使用专用编辑器或终端组件，不放入普通 textarea 长时间渲染。
- 弹窗内部重组件使用 `unmountOnClose` 或显示时再创建。

## 提交前检查

建议命令：

```bash
npm run build
rg "offline|k8soffline" w7panel-ui/src
rg "localStorage\\.|sessionStorage\\." w7panel-ui/src
rg "/api/v1/" w7panel-ui/src
```

检查清单：

| 检查项 | 说明 |
|--------|------|
| 路径前缀 | 面板业务 API 是否使用 `/panel-api/v1`，K8s 代理是否使用 `/k8s-proxy` |
| token | 是否统一从 auth 工具或微应用 props 获取 |
| 组件文档 | 公共组件是否更新 Props、Events、Ref 方法 |
| Wujie 文档 | 新事件是否记录参数和回调 |
| 错误处理 | API 失败时是否有合理提示或降级 |
| 清理逻辑 | 定时器、终端、日志、Wujie 事件是否清理 |
