# 前端性能规范

本文档从历史前端性能报告中提炼长期开发规范，适用于 `w7panel-ui` 的页面、API 封装、组件、hooks、store 和资源加载。

## 整体要求

前端性能优化优先处理用户能感知的慢路径：

| 场景 | 风险 | 要求 |
|------|------|------|
| 页面首屏 | 多个接口串行导致等待放大 | 互不依赖的请求并行，依赖关系显式写清 |
| 列表页 | DOM 节点过多、重复渲染 | 分页、筛选、虚拟滚动或后端 limit |
| 轮询和定时器 | 页面关闭后仍请求 | 统一 hooks 管理并在生命周期中清理 |
| 全局状态 | 重复拉取 namespace、用户信息等 | 使用 Pinia 和带 TTL 的缓存 |
| 文件、日志、终端 | 大文本、连接和编辑器实例重 | 用专用组件，关闭时释放资源 |

## 页面性能预算

页面开发时先区分首屏关键数据和延迟数据。

| 数据类型 | 加载时机 | 要求 |
|----------|----------|------|
| 首屏必需数据 | 页面进入立即加载 | 互不依赖时并行，有 loading 和 error |
| 次要卡片数据 | 首屏后加载或进入视口加载 | 不阻塞页面主体 |
| 弹窗/抽屉数据 | 打开时加载 | 关闭时清理，必要时缓存 |
| 大列表详情 | 用户展开或点击时加载 | 不在列表初始化时逐项串行请求 |
| 日志/终端/文件内容 | 用户进入能力时加载 | 可取消、可关闭、可释放 |

常用预算：

| 指标 | 默认要求 |
|------|----------|
| 首屏接口 | 不依赖的请求并行 |
| 普通列表直接 DOM 渲染 | 小于 1000 条 |
| 轮询间隔 | 默认不低于 5 秒，实时场景需说明 |
| 弹窗重组件 | 默认按需创建 |
| 大文本渲染 | 使用编辑器、虚拟列表或流式组件 |

## API 请求

约定：

- 页面初始化时，互不依赖的请求使用 `Promise.all` 或统一请求 hook 并行处理。
- 不在列表循环中串行请求每一项详情；必须逐项请求时限制并发或分批请求。
- 优先使用 `src/utils/api.ts` 的 `panelApi`、`k8sproxy` 或 `src/api/` 方法。
- URL 参数使用 axios `params` 或 `encodeURIComponent`，避免因为特殊字符产生重复失败请求。
- 全局 axios 超时由 `src/api/interceptor.ts` 管理，单接口需要更长时间时在调用处明确说明。
- 同一 GET 请求在短时间内可能被多个组件触发时，应复用已有 store、hook 缓存或防重机制。

### 请求组织

| 场景 | 推荐写法 | 原因 |
|------|----------|------|
| 多个互不依赖接口 | `Promise.all` | 缩短首屏等待 |
| 同一资源多组件使用 | Pinia store 或 hook 缓存 | 避免重复请求 |
| 列表项详情 | 展开时加载、分批加载或后端聚合 | 避免 N+1 |
| 搜索输入 | debounce + 取消上一次请求 | 避免请求风暴 |
| 切换 namespace | 取消旧请求，刷新依赖状态 | 避免旧数据覆盖新数据 |

推荐模式：

```ts
const [apps, namespaces, metrics] = await Promise.all([
  fetchApps(),
  namespaceStore.fetchNamespaceList(),
  fetchMetrics(),
]);
```

需要避免：

```ts
for (const item of list) {
  item.detail = await fetchDetail(item.name);
}
```

分批并发模式：

```ts
const batchSize = 10;
for (let i = 0; i < list.length; i += batchSize) {
  const batch = list.slice(i, i + batchSize);
  await Promise.all(batch.map((item) => fetchDetail(item.name)));
}
```

搜索请求模式：

```ts
const controller = new AbortController();
await panelApi.get('/apps/search', {
  params: { keyword },
  signal: controller.signal,
});
controller.abort();
```

## Hooks、轮询和生命周期

- 定时器、轮询、WebSocket、xterm、CodeMirror、ECharts 实例必须在 `onUnmounted` 中清理。
- keep-alive 页面需要在 `onDeactivated` 中停止轮询，在 `onActivated` 中按需恢复。
- 弹窗、抽屉内的重组件关闭时必须释放连接和实例；必要时使用 `unmountOnClose`。
- 轮询间隔按业务实时性设置，避免多个组件同时轮询同一接口。
- 使用 `src/hooks/timer.ts`、`src/hooks/request.ts` 等已有 hooks 时，优先复用其取消、缓存和清理能力。

检查项：

| 能力 | 必须清理 |
|------|----------|
| `setInterval`、`setTimeout` | timer id |
| 轮询请求 | polling handle、AbortController |
| WebSocket、终端 | socket、xterm 实例 |
| 编辑器 | CodeMirror、Monaco 或 Codeblitz 实例 |
| 图表 | ECharts instance |
| Wujie 事件 | bus listener |

### 生命周期规则

| 组件类型 | 初始化 | 清理 |
|----------|--------|------|
| 普通页面 | `onMounted` 或路由参数准备后 | `onUnmounted` |
| keep-alive 页面 | `onActivated` 恢复可见数据 | `onDeactivated` 停止轮询 |
| 弹窗/抽屉 | `visible` 变为 true 后加载 | `visible` 变为 false 后取消请求和释放重实例 |
| Wujie 微应用容器 | 子应用挂载后注册事件 | 容器销毁时注销事件 |
| 文件/日志/终端 | 用户进入能力后连接 | 关闭 tab、弹窗或路由离开时断开 |

watch 使用要求：

- `watch` 里发请求时必须处理连续变更，避免旧响应覆盖新状态。
- 监听 namespace、pod、container、path 等关键参数时，参数为空应直接短路。
- 使用 `immediate: true` 时确认不会和 `onMounted` 重复请求。

## 状态和缓存

- namespace、用户信息、权限、K8s 基础信息等跨页面状态放 Pinia。
- localStorage/sessionStorage key 必须使用模块前缀，遵守 [auth-state.md](../frontend/auth-state.md)。
- 缓存必须定义 TTL、强制刷新入口和清理入口。
- 不把同一后端响应长期复制到多个 store 或组件变量中，避免状态不同步。
- token、密码、密钥、OIDC code 不缓存到非必要位置，不输出到 console。

示例：

```ts
await namespaceStore.fetchNamespaceList();
await namespaceStore.fetchNamespaceList(true); // 强制刷新
```

### 缓存设计

| 字段 | 要求 |
|------|------|
| key | 包含模块、namespace、资源名、用户上下文 |
| ttl | 按数据变化频率设置，不能永久有效 |
| force refresh | 页面刷新、用户手动刷新或 namespace 切换可绕过缓存 |
| clear | 登出、token 失效、权限变化时能清理 |
| stale behavior | 过期期间是显示旧数据并刷新，还是清空后加载，需要明确 |

缓存适用性：

| 数据 | 适合缓存 | 注意 |
|------|----------|------|
| namespace 列表 | 适合 | namespace 变更后强制刷新 |
| 用户信息和权限 | 适合 | token 刷新、登出时清理 |
| 监控实时数据 | 短 TTL | 不要长期缓存 |
| 日志流 | 不适合长期缓存 | 关闭连接优先 |
| 文件内容 | 谨慎 | 保存、外部变化和编码问题要处理 |

## 列表和渲染

- 1000 条以上列表必须使用分页、虚拟滚动或后端 limit，不直接渲染完整 DOM。
- 表格 `row-key` 和 `v-for :key` 使用稳定业务字段，不使用数组 index。
- 大型 YAML、日志、文本文件使用专用编辑器或终端组件，不放到普通 textarea 长时间渲染。
- ECharts 只注册需要的图表类型和组件，新增类型同步全局组件注册文档。
- 图片、应用图标、预览图等可延迟资源使用懒加载或按需加载。
- 计算量大的过滤、排序和格式化结果应使用 `computed` 或后端过滤，不在模板里重复计算。

示例：

```vue
<tr v-for="item in list" :key="item.metadata?.uid || item.name">
  ...
</tr>

<a-image :src="item.icon" loading="lazy" />
```

### 列表策略

| 规模 | 策略 |
|------|------|
| 小于 200 条 | 普通表格即可 |
| 200 到 1000 条 | 分页、筛选、减少列渲染和复杂 formatter |
| 1000 条以上 | 虚拟滚动或后端分页 |
| 不可预估规模 | 默认后端分页或 limit |

表格优化：

- `row-key` 使用 `metadata.uid`、资源名、路径等稳定字段。
- 避免在模板中调用复杂函数，提前用 `computed` 或数据预处理。
- 操作列中的弹窗、菜单、tooltip 不要为每行提前创建重组件。
- 大量标签、图标、进度条会放大渲染成本，需要按需展示或折叠。

图表优化：

- ECharts 初始化前确认容器可见且有尺寸。
- 页面切换、弹窗关闭时调用 `dispose`。
- 高频数据更新使用节流，不要每个点都触发完整 `setOption`。
- 图表数据超过当前视图需要时做采样或聚合。

## 组件和资源

- 公共组件要避免在 props 变化时重复发起完整初始化。
- 抽屉和弹窗中的大组件按需创建，关闭后释放。
- 路由级页面和低频能力优先异步加载。
- 大依赖新增前必须确认是否已有替代组件或按需导入能力。
- 图标优先使用项目已有图标库，避免为单个图标引入整包。

### 组件初始化

组件内部不要在 `created`、`setup` 或模块顶层直接发起和显示状态无关的大请求。推荐按场景触发：

| 组件 | 触发时机 |
|------|----------|
| 页面主体 | 路由参数和 store 初始化完成后 |
| Drawer/Modal | `visible=true` 后 |
| Tab 面板 | tab 首次激活时 |
| 折叠详情 | 用户展开时 |
| 微应用能力 | 微应用 ready 后 |

### 资源加载

- 图片使用懒加载、错误兜底和固定尺寸，避免布局抖动。
- 大型 JSON、YAML、日志文本按需加载，避免打进首屏 bundle。
- 新增第三方库前检查是否支持 tree-shaking 和按需导入。
- 静态常量和大枚举不要在多个组件重复定义。

## 反模式清单

| 反模式 | 风险 | 替代方案 |
|--------|------|----------|
| 页面 `onMounted` 串行 await 多个接口 | 首屏等待累加 | `Promise.all` 或拆分首屏/延迟数据 |
| `v-for :key="index"` | 更新错乱、重复渲染 | 稳定业务 key |
| 每行创建弹窗或编辑器 | DOM 和实例过多 | 单例弹窗 + 当前行状态 |
| 关闭弹窗不取消请求 | 旧响应污染状态 | AbortController 或 hook cancel |
| watch 里无条件请求 | 参数变化导致请求风暴 | 参数校验 + debounce + 去重 |
| 长日志放普通 textarea | 页面卡顿 | 终端、虚拟列表或专用编辑器 |
| 多组件各自读取 localStorage | 状态分散 | auth 工具或 Pinia store |

## 验证方式

提交前端性能相关改动时至少执行：

```bash
npm run build
rg "setInterval|setTimeout|addEventListener|\\$on|new WebSocket|echarts\\.init" w7panel-ui/src
rg ":key=\"index\"|:key='index'|console\\.log" w7panel-ui/src
```

涉及列表、文件、日志、终端或图表时补充验证：

- 首屏请求是否并行，Network 中是否出现不必要串行瀑布。
- 关闭弹窗、切换路由、退出页面后是否停止轮询和 WebSocket。
- 1000 条以上数据是否仍能滚动和筛选。
- 大文本、大 YAML、日志流是否不会阻塞页面。
- 刷新 token、namespace 切换、缓存强制刷新后页面数据是否一致。

## 评审清单

提交前端性能相关 PR 前自查：

| 检查项 | 通过标准 |
|--------|----------|
| 请求 | 首屏请求无不必要串行；列表无 N+1 串行请求 |
| 取消 | 路由切换、弹窗关闭、namespace 切换时请求可取消或旧响应不会覆盖新状态 |
| 生命周期 | timer、WebSocket、事件、图表、编辑器实例均清理 |
| 列表 | 大列表有分页、虚拟滚动或后端 limit |
| key | `v-for` 和表格 row-key 使用稳定业务字段 |
| 缓存 | key、TTL、强刷、清理逻辑明确 |
| 资源 | 新依赖、新图表、新图片不会进入不必要首屏路径 |
| 文档 | 新增公共性能模式同步更新本规范或前端约定 |
