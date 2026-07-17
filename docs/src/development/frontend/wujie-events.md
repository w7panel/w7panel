# Wujie 微前端事件说明

本文档整理 `registerWujieEvent()` 注册的事件、入参、回调响应和典型调用方式。微应用容器、props 和 token 边界见 [microapps.md](./microapps.md)。源码以当前 `w7panel-ui/src` 为准。

## 整体使用方式

Wujie 事件文档用于维护面板和微应用之间的事件协议。开发时先确认事件由谁发起、是否需要回调、回调数据结构是什么，再在面板侧注册事件、在微应用侧按约定触发。

### 基本流程

1. 面板容器页面挂载需要的事件注册组件，例如 `WujieModals`。
2. 面板侧使用 `registerWujieEvent(event, handler)` 注册事件，并在卸载时清理。
3. 微应用侧使用 `window.$wujie.bus.$emit(event, payload, callback, rejectCallback)` 触发事件。
4. 事件需要响应时，明确 callback 返回字段；需要失败路径时明确 rejectCallback。
5. 新增或修改事件名、payload、callback 响应时，同步更新本文。

### 使用边界

- 事件名要避免过泛，新增通用名称前先搜索现有事件。
- payload 优先使用对象，便于后续扩展。
- callback 不是 Promise，调用方必须按回调协议处理成功和失败。
- 微应用 props 和 token 边界见 [microapps.md](./microapps.md)。

## 基础用法

源文件：`w7panel-ui/src/hooks/use-wujie-events.ts`

`registerWujieEvent(event, handler)` 是对 `wujie` 的 `bus.$on(event, handler)` 的封装，额外记录已注册事件，组件卸载时可通过 `clearAllWujieEvents()` 统一注销。

面板侧注册示例：

```ts
import { registerWujieEvent, clearAllWujieEvents } from '@/hooks/use-wujie-events';

registerWujieEvent('openPage', this.openPage);

beforeUnmount() {
  clearAllWujieEvents();
}
```

微应用侧调用示例：

```ts
window.$wujie.bus.$emit('openPage', {
  title: '外部页面',
  src: 'https://example.com',
});
```

微应用侧推荐封装：

```ts
export function getWujieBus() {
  return window.$wujie?.bus;
}

export function emitPanelEvent(eventName: string, payload?: unknown) {
  return new Promise((resolve, reject) => {
    const bus = getWujieBus();
    if (!bus) {
      reject(new Error('Wujie bus is not available'));
      return;
    }

    bus.$emit(eventName, payload, resolve, reject);
  });
}
```

使用方式：

```ts
// 无回调事件：打开面板弹窗
window.$wujie?.bus?.$emit('openPage', {
  title: '帮助',
  src: 'https://example.com/help',
});

// 有回调事件：按 Promise 处理成功和失败
await emitPanelEvent('buildImage', buildPayload);
```

带响应回调的调用示例：

```ts
window.$wujie.bus.$emit('checkSession', (token: string) => {
  // token 为空表示刷新失败或未返回
});
```

注意事项：

- `handler` 为空时会跳过注册并输出 warning。
- `clearAllWujieEvents()` 清理的是当前 hook 模块记录的全部事件；同一页面嵌套多个注册组件时，要确认清理时机不会误清其它仍在使用的事件。
- 多参数响应遵循 `window.$wujie.bus.$emit(event, data, callback, rejectCallback)` 这种约定，回调不是 Promise。
- 面板传给微应用的 props 由各容器页面 `startApp()` 或 `setupApp()` 配置决定，常见字段见 [microapps.md](./microapps.md)。
- 微应用触发事件前要判断 `window.$wujie?.bus` 是否存在；独立运行调试时可以降级为空实现或提示当前不在面板容器内。
- 微应用不要假设所有事件都有回调；是否有响应以本文每个事件的“响应”说明为准。
- payload 优先传对象，不要传位置参数数组，后续扩展字段时更稳定。

## 通用弹窗和面板能力事件

注册位置：`w7panel-ui/src/components/wujie-modals.vue`

该组件挂载在微应用容器页面中，提供文件、日志、构建镜像、域名、应用配置等面板能力。

### `toStoreInstall`

功能：跳转到应用商店安装页。

参数：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | zpk 安装路径或应用标识 |

响应：无。

示例：

```ts
window.$wujie.bus.$emit('toStoreInstall', 'https://zpk.w7.cc/zpk/respo/info/demo_app');
```

### `openPage`

功能：在面板弹窗中打开 iframe 页面。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `src` | string | 是 | iframe 地址 |
| `title` | string | 否 | 弹窗标题 |

响应：无。

示例：

```ts
window.$wujie.bus.$emit('openPage', {
  title: '帮助文档',
  src: 'https://example.com/help',
});
```

### `openApp`

功能：在面板弹窗中打开应用组微应用页面。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `appgroup` | string | 是 | 应用组名称 |
| `path` | string | 否 | 微应用内部路径，会写入 `path` query |
| `title` | string | 否 | 弹窗标题 |

响应：无。

示例：

```ts
window.$wujie.bus.$emit('openApp', {
  appgroup: 'demo',
  path: '/settings',
  title: '应用设置',
});
```

### `toFile`

功能：跳转到当前应用组下的文件管理页面。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `kind` | string | 是 | Workload 复数类型，例如 `deployments`、`statefulsets` |
| `appname` | string | 是 | 应用名称 |
| `path` | string | 否 | 初始文件路径 |

响应：无。

示例：

```ts
window.$wujie.bus.$emit('toFile', {
  kind: 'deployments',
  appname: 'nginx',
  path: '/etc/nginx/',
});
```

### `openFile`

功能：在面板弹窗中打开文件管理组件。

参数同 `toFile`。

响应：无。

示例：

```ts
window.$wujie.bus.$emit('openFile', {
  kind: 'statefulsets',
  appname: 'mysql',
  path: '/var/lib/mysql/',
});
```

### `buildImage`

功能：提交 zpk 镜像构建任务。

参数：直接透传给 `POST /panel-api/v1/zpk/buildimage/job`。

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `callback` | 无 | 后端任务创建成功后调用 |

示例：

```ts
window.$wujie.bus.$emit('buildImage', buildPayload, () => {
  // 构建任务已提交
});
```

### `buildImageLog`

功能：打开构建 Job 日志弹窗。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `buildJobName` | string | 是 | Job 名称 |

响应：无。

示例：

```ts
window.$wujie.bus.$emit('buildImageLog', { buildJobName: 'build-demo-xxx' });
```

### `closeBuildImageLog`

功能：关闭构建 Job 日志弹窗。

参数：无。

响应：无。

```ts
window.$wujie.bus.$emit('closeBuildImageLog');
```

### `zip`

功能：压缩容器内文件并生成下载链接。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `pid` | object | 是 | 目标容器定位信息 |
| `pid.namespace` | string | 是 | 命名空间 |
| `pid.HostIp` | string | 否 | 节点 IP |
| `pid.containerId` | string | 否 | 容器 ID |
| `pid.containerName` | string | 是 | 容器名 |
| `pid.podName` | string | 是 | Pod 名称 |
| `input` | array | 是 | 待压缩源路径列表，传给 `compressFiles()` |
| `output` | string | 是 | 压缩包输出路径 |
| `callback` | function | 否 | data 内部回调 |

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `data.callback` | `{ link }` | 兼容 data 内回调 |
| 第二参数 `callback` | `{ link }` | 标准事件回调 |

示例：

```ts
window.$wujie.bus.$emit('zip', {
  pid: {
    namespace: 'default',
    podName: 'nginx-xxx',
    containerName: 'nginx',
  },
  input: ['/etc/nginx'],
  output: '/tmp/nginx.tar.gz',
}, ({ link }) => {
  window.open(link);
});
```

### `uploadFile`

功能：上传文件到容器内指定目录。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `pid` | object | 是 | 目标容器定位信息，字段同 `zip.pid` |
| `path` | string | 是 | 上传目录，代码中会拼接 `path + file.name` |
| `file` | File | 是 | 浏览器 File 对象 |

响应：

| 回调参数 | 说明 |
|----------|------|
| 无参数 | 上传成功 |
| `Error/AxiosError` | 上传失败 |

示例：

```ts
window.$wujie.bus.$emit('uploadFile', {
  pid: {
    namespace: 'default',
    podName: 'nginx-xxx',
    containerName: 'nginx',
  },
  path: '/tmp/',
  file,
}, (err?: unknown) => {
  if (err) return;
  // 上传成功
});
```

### `domainCert`

功能：打开或更新域名证书相关数据，由 `views/topapp/domain-cert.vue` 处理展示。

参数：证书组件所需的 domain cert 数据对象。

响应：无。

```ts
window.$wujie.bus.$emit('domainCert', certData);
```

### `podLog`

功能：打开 Pod 日志弹窗。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | Pod 名称 |
| `namespace` | string | 否 | 命名空间 |
| `container` | string | 否 | 默认容器名 |
| `containerList` | array | 否 | 容器列表 |

响应：无。

```ts
window.$wujie.bus.$emit('podLog', {
  name: 'nginx-xxx',
  namespace: 'default',
  container: 'nginx',
});
```

### `openAppForm`

功能：打开微应用配置表单。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `yaml` | string | 否 | YAML 配置 |
| `json` | object | 否 | JSON 配置 |

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `callback` | 由 `MicroAppForm` 决定 | 表单提交后的结果 |

示例：

```ts
window.$wujie.bus.$emit('openAppForm', {
  yaml: yamlText,
}, (result) => {
  // 处理表单提交结果
});
```

### `ingressEdit`

功能：打开微应用域名编辑弹窗。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ingress` | object | 是 | Ingress 数据 |
| `appList` | array | 否 | 应用列表 |
| `appPorts` | object/array | 否 | 应用端口数据 |
| `callback` | function | 否 | 提交后的回调 |

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `data.callback` | object | 域名编辑提交值 |

```ts
window.$wujie.bus.$emit('ingressEdit', {
  ingress,
  appList,
  appPorts,
  callback: (value) => {},
});
```

### `ingressStrategy`

功能：打开域名策略配置弹窗。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ingress` | object | 是 | Ingress 数据，读取 `ingress.backend.strategy` 回显 |
| `callback` | function | 否 | 策略提交后的回调 |

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `data.callback` | object | 策略数据，来自 `DomainStrategy` 的 `v[0].value` |

```ts
window.$wujie.bus.$emit('ingressStrategy', {
  ingress,
  callback: (strategy) => {},
});
```

### `checkSession`

功能：使用 refresh token 请求新 token。

参数：第一个参数就是回调函数。

响应：

| 回调参数 | 说明 |
|----------|------|
| `token` | 刷新成功后的访问 token；失败时当前实现不调用回调 |

```ts
window.$wujie.bus.$emit('checkSession', (token: string) => {
  // 使用新 token
});
```

### `containerPlugin`

功能：打开容器插件配置抽屉。

参数：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `data` | object | 是 | 传给 `ContainerPlugin` 的 `propsData` |
| `callback` | function | 否 | 抽屉确认时传给 `ContainerPlugin.submit(callback)` |

响应：由 `ContainerPlugin.submit()` 决定。

```ts
window.$wujie.bus.$emit('containerPlugin', pluginData, (result) => {
  // 保存插件配置
});
```

### `buildContainerImage`

功能：打开“从运行中容器构建镜像”的状态弹窗。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `podName` | string | 是 | Pod 名称 |
| `containerName` | string | 是 | 容器名称 |
| `imageName` | string | 是 | 原镜像名，成功回调会原样带回 |
| `cmd` | string/array | 否 | 构建命令 |
| `pinned` | boolean | 否 | 是否固定镜像 |
| `updateImage` | boolean | 否 | 是否更新 workload 镜像 |

响应：

| 回调 | 参数 | 说明 |
|------|------|------|
| `callback` | `{ ...result, imageName }` | 构建完成回调，额外补充原始 `imageName` |
| `rejectCallback` | 由构建组件决定 | 构建失败或取消回调 |

```ts
window.$wujie.bus.$emit('buildContainerImage', {
  podName: 'nginx-xxx',
  containerName: 'nginx',
  imageName: 'registry.local.w7.cc/default/nginx:latest',
  updateImage: true,
}, (result) => {
  // 构建完成
}, (err) => {
  // 构建失败
});
```

### `getOidcCode`

功能：为微应用获取 OIDC 一次性 code。

参数：

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `client_id` | string | 否 | `default` | OIDC client id |
| `redirect_uri` | string | 否 | `http://127.0.0.1:3000/callback` | 回调地址 |
| `scope` | string | 否 | `openid` | scope |

响应：

| 回调参数 | 说明 |
|----------|------|
| `code` | 成功时返回 code，失败时返回空字符串 |

```ts
window.$wujie.bus.$emit('getOidcCode', {
  client_id: 'demo',
  redirect_uri: 'https://demo.example.com/callback',
  scope: 'openid profile',
}, (code: string) => {});
```

### `changeLogin`

功能：切换面板登录态并跳转首页。

参数：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `token` | string | 是 | 新访问 token |
| `refreshToken` | string | 是 | 新 refresh token |

响应：无。

```ts
window.$wujie.bus.$emit('changeLogin', {
  token,
  refreshToken,
});
```

## 面板下发给微应用的事件

这些事件不是通过 `registerWujieEvent()` 注册，而是面板通过 Wujie bus 主动下发给微应用。微应用侧使用 `window.$wujie.bus.$on()` 监听。

| 事件 | 触发位置 | 参数 | 说明 |
|------|----------|------|------|
| `routeChange` | `topapp/micro-container.vue`、`app/apps/detail.vue` | string | 面板菜单或路由变化 |

监听示例：

```ts
window.$wujie.bus.$on('routeChange', (path: string) => {
  // 微应用内部切换路由
});

```

## 新增事件规范

- 事件名使用动词或动词短语，例如 `openFile`、`buildContainerImage`。
- 入参使用对象，除非事件只有一个简单字符串参数。
- 需要响应时优先使用第二参数 callback；失败路径需要明确是否调用 callback 或 rejectCallback。
- 新增注册事件后同步更新本文档和前端开发文档入口。
- 组件卸载时必须清理注册事件，避免微应用切换后重复响应。
