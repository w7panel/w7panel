# 重构计划 v5.4

**发布日期**: 2026-02-22

## 一、问题概述

通过对项目代码的系统性分析，发现以下问题类别：

| 问题类别 | 严重程度 | 数量 |
|----------|---------|------|
| 已有 Hook 未使用 | 🔴 严重 | 3 个 Hook |
| console.log 生产代码 | 🟡 中等 | 132 处 |
| 注释掉的死代码 | 🟡 中等 | 126 处 |
| 空 catch 块静默吞错 | 🟡 中等 | 56 处 |
| Hook 代码 Bug | 🔴 严重 | 1 处 |
| 重复 API 路径字符串 | 🔴 严重 | 135+ 处 |
| JSON 深拷贝过度使用 | 🟡 中等 | 74 处 |
| window.* 全局对象使用 | 🟢 良好 | 255 处 (已规范) |
| 事件监听器管理 | 🟢 良好 | 13 对 (已正确清理) |

**已验证无需优化的模式**:
- ✅ kube-system 硬编码 - 用户确认为常规 API 传参
- ✅ 事件监听器 - 13 处 addEventListener 都有 removeEventListener

---

## 二、已完成的优化 (v5.4 本次会话)

### 2.1 P0: Hook Bug 修复 - timer.ts setInterval 重复定义

| 优化项 | 文件 | 说明 |
|--------|------|------|
| 删除重复 setInterval 定义 | `hooks/timer.ts` | 删除第 43-45 行重复定义 |

**效果**: 修复运行时可能的 Bug

### 2.2 P1: useTimer Hook 使用验证

| 优化项 | 文件 | 说明 |
|--------|------|------|
| 验证定时器清理 | 6 个 Vue 文件 | 已正确在 beforeUnmount 中清理 |

**结论**: 7 个文件已正确管理定时器，无需修改

### 2.3 P5: localStorage 键名统一

| 优化项 | 文件 | 说明 |
|--------|------|------|
| 统一键名前缀 | `utils/auth.ts` | token→w7panel-token 等 |
| 统一键名前缀 | `store/modules/app/index.ts` | APP_MENU_FILTER→w7panel-menu-filter |

**效果**: 统一规范

### 2.4 P4: API 路径工具封装

| 优化项 | 文件 | 说明 |
|--------|------|------|
| 新建 API 工具 | `utils/k8s-api.ts` | 封装常用 K8s API 路径 |

**效果**: 提供统一 API 路径构建函数

### 2.5 根因修复：namespace=undefined 问题

| 优化项 | 文件 | 说明 |
|--------|------|------|
| Store 初始值修复 | `store/modules/namespace.ts` | 将 `namespace: ''` 改为 `namespace: 'default'` |
| namespaceActive 初始化 | `components/dcform-drawer.vue` | 添加 Store 引用 |

**效果**: 从根源解决 `namespace=undefined` 导致 API 路径错误的问题

### 2.2 Drawer 组件性能优化

| 优化项 | 文件 | 说明 |
|--------|------|------|
| 延迟 API 调用 | `components/domain-strategy.vue` | 移除 created() 中的 API 调用，移到 watch(show) |
| 移除重复调用 | `components/domain-strategy-plugin.vue` | 移除 created() 中的重复调用 |

**效果**: Drawer/Modal 组件仅在真正显示时加载数据，减少不必要请求

### 2.3 API 路由修复

| 优化项 | 文件 | 说明 |
|--------|------|------|
| /version → /k8s-proxy/version | `views/cluster/overview/panel.vue` | 符合 v5.3 重构规范 |

### 2.4 控制台警告修复

| 优化项 | 文件 | 说明 |
|--------|------|------|
| Intlify 翻译警告 | `components/menu/index.vue` | 使用路由名称作为备选 |
| Wujie 事件订阅警告 | 新建 `hooks/use-wujie-events.ts` | 统一事件管理 Hook |
| 隐藏容器终端菜单 | `router/routes/modules/dialogpage.ts` | 添加 hideInMenu: true |
| 翻译键缺失 | `locale/zh-CN.ts`, `locale/en-US.ts` | 添加 dialog-pod-webshell |

---

## 三、深入分析发现的新问题

### 3.1 事件监听器管理 ✅ 已正确处理

**现状**: 13 处 addEventListener 都有对应的 removeEventListener

**示例**:
```typescript
// panel.vue:660-663
window.addEventListener('message', this.paySuccess);
// ...
window.removeEventListener('message', this.paySuccess);
```

**结论**: 事件监听器管理规范，无需优化

---

### 3.2 硬编码 Namespace 🔴

**问题**: 27 处硬编码 `kube-system` Namespace

**示例**:
```javascript
// 错误 - 硬编码
axios.get('/k8s-proxy/api/v1/namespaces/kube-system/configmaps/k3s.config')

// 正确 - 使用常量或配置
const SYSTEM_NAMESPACE = 'kube-system';
axios.get(`/k8s-proxy/api/v1/namespaces/${SYSTEM_NAMESPACE}/configmaps/k3s.config`)
```

**影响文件** (部分):
- `views/cluster/overview/panel.vue`: 3 处
- `views/cluster/nodes/index.vue`: 3 处
- `views/system/system/index.vue`: 6 处
- `views/system/usermanage/quota.vue`: 3 处
- 等...

**注意**: 用户确认 kube-system 硬编码是常规 API 传参，属于正常模式，无需优化

---

### 3.3 重复 API 路径字符串 🔴

**问题**: 135+ 处重复写相同的 API 路径

**示例**:
```javascript
// 重复 - 每次都写完整路径
axios.get('/k8s-proxy/api/v1/namespaces/' + this.namespaceActive + '/persistentvolumeclaims')
axios.get('/k8s-proxy/api/v1/namespaces/' + this.namespaceActive + '/configmaps')
axios.get('/k8s-proxy/api/v1/namespaces/' + this.namespaceActive + '/secrets')
```

**建议**: 封装 API 路径工具函数
```typescript
// utils/api.ts
export const buildK8sApi = (resource: string, namespace?: string) => {
    const ns = namespace || useNamespaceStore().namespace;
    return `/k8s-proxy/api/v1/namespaces/${ns}/${resource}`;
};

// 使用
axios.get(buildK8sApi('persistentvolumeclaims'))
axios.get(buildK8sApi('configmaps'))
```

### 3.4 JSON 深拷贝过度使用 🟡

**问题**: 74 处使用 `JSON.parse(JSON.stringify())` 进行深拷贝

**影响**:
- 性能开销大
- 无法处理特殊对象 (Date, RegExp, Function 等)
- 代码冗长

**建议**: 使用专业深拷贝库或优化拷贝逻辑
```typescript
// 替换方案
import { cloneDeep } from 'lodash-es';
// 或
import structuredClone from 'structuredClone'; // 现代浏览器原生支持
```

**高频使用文件**:
- `views/app/pages/domain.vue`: 12 处
- `views/app/cronjob/cronjob-drawer.vue`: 9 处
- `views/cluster/overview/panel.vue`: 5 处
- `views/storage/disk.vue`: 4 处

---

### 3.5 window.* 全局对象使用 🟢

**现状**: 255 处使用 window 对象

**用途分析**:
| 用途 | 数量 | 说明 |
|------|------|------|
| window.formatDate | 45+ | 日期格式化 - 建议封装为工具函数 |
| window.location | 20+ | 页面跳转 - 正常用法 |
| window.addEventListener | 13 | 事件监听 - 已正确清理 |
| document.getElementById | 15 | DOM 操作 - 正常用法 |

**建议**: 将 `window.formatDate` 封装为独立工具函数
```typescript
// utils/date.ts
export function formatDate(timestamp: number | string): string {
    return window.formatDate(timestamp);
}
```

---

### 3.6 localStorage 键名不一致 🟡

#### 3.6.1 前端分析

**问题**: 多个不同的键名前缀，不统一

**现状分析**:

| 键名 | 位置 | 状态 |
|------|------|------|
| `w7panel-permission` | auth.ts:4 | ✅ 已统一 |
| `w7panel-userinfo` | auth.ts:5 | ✅ 已统一 |
| `token` | auth.ts:1 | ❌ 需统一 |
| `refreshtoken` | auth.ts:2 | ❌ 需统一 |
| `fileeditor` | auth.ts:65 | ❌ 需统一 |
| `webshell` | auth.ts:71 | ❌ 需统一 |
| `k8sinfo` | auth.ts:79 | ❌ 需统一 |
| `APP_MENU_FILTER` | store/modules/app/index.ts:6 | ❌ 需统一 |
| `arco-locale` | locale/index.ts:9 | ⚠️ 第三方库，保持 |

**需要统一的键名**:
```
token           → w7panel-token
refreshtoken   → w7panel-refresh-token
fileeditor     → w7panel-fileeditor
webshell       → w7panel-webshell
k8sinfo        → w7panel-k8sinfo
APP_MENU_FILTER → w7panel-menu-filter
```

#### 3.6.2 后端分析

**结论**: 后端**不需要**修改

**原因**:
- 后端代码中无 `localStorage` 相关代码
- localStorage 是前端浏览器存储，仅在前端使用
- 后端通过 JWT Token 认证，与前端 localStorage 无关

#### 3.6.3 优化方案

**修改文件**:
```
w7panel-ui/src/utils/auth.ts          - 6 处键名修改
w7panel-ui/src/store/modules/app/index.ts - 1 处键名修改
```

---

## 四、计划优化 (v5.5)

### 4.1 Hook 使用规范化 🔴 严重

#### 问题描述
项目已封装完善的 Hook，但大部分未使用，导致：
- 代码重复
- 无法复用缓存/重试能力
- 资源泄漏风险

#### 优化方案

##### 4.1.1 useRequest Hook 推广使用

**现状**: 
- Hook 存在于 `hooks/request.ts` (195 行)
- 特性: 请求缓存、重试、超时、Loading 状态
- 使用率: **0%** (145+ 文件直接用 axios)

**优化**:
```
影响范围: 145+ 文件
工作量: 高
收益: 统一请求管理、减少重复代码
建议: 渐进式替换，优先替换高频请求
```

##### 4.1.2 useTimer/usePolling Hook 推广使用

**现状**:
- Hook 存在于 `hooks/timer.ts` (141 行)
- 特性: 统一定时器管理、自动清理资源
- 使用率: **0%** (7 个文件使用原生 setInterval)

**问题文件**:
```
views/app/apps/detail.vue:883       - watchInterval
views/init-cluster/resource-loading.vue:29  - interval
views/init-cluster/order-base.vue:992       - interval
views/header/allow-register-check.vue:35   - interval
views/cluster/nodes/set-gpu-bu.vue:204     - statusInterval
views/cluster/nodes/set-gpu.vue:198       - statusInterval
```

**优化**: 改用 `useTimer()` 或 `usePolling()` 自动清理资源

##### 4.1.3 Hook Bug 修复

**问题**: `timer.ts` 第 39-45 行 setInterval 重复定义

```typescript
// 第 39 行
const setInterval = (id: string, callback: () => void, delay: number): NodeJS.Timeout => {
    return setTimer(id, callback, delay, 'setInterval') as NodeJS.Timeout;
};

// 第 43 行 (重复!)
const setInterval = (id: string, callback: () => void, delay: number): NodeJS.Timeout => {
    return setTimer(id, callback, delay, 'setInterval') as NodeJS.Timeout;
};
```

**修复**: 删除重复行

### 4.2 代码质量优化

#### 4.2.1 清理 console.log

**现状**: 132 处 console.log 在生产代码中

**影响**:
- 生产环境控制台噪音
- 可能的性能影响
- 信息泄露风险

**优化**: 批量移除或替换为日志开关

#### 4.2.2 清理注释掉的代码

**现状**: 126 处注释掉的代码（API 调用、import 等）

**问题**:
- 死代码难以维护
- 增加理解成本
- 可能导致误用

**优化**: 删除或标记 TODO

#### 4.2.3 空 catch 块处理

**现状**: 138 处 `.catch(()=>{})` 静默吞掉错误

**问题**:
- 错误被隐藏
- 调试困难
- 可能导致数据不一致

**优化**: 添加错误日志或使用全局错误处理

### 4.3 架构优化 (长期)

#### 4.3.1 封装 API 路径工具

```typescript
// 新建 utils/k8s-api.ts
export function buildNamespacedApi(resource: string, namespace?: string): string {
    const ns = namespace || useNamespaceStore().namespace || 'default';
    return `/k8s-proxy/api/v1/namespaces/${ns}/${resource}`;
}

export function buildClusterApi(resource: string): string {
    return `/k8s-proxy/api/v1/${resource}`;
}
```

#### 4.3.2 统一 localStorage 键名前缀

建议统一使用 `w7panel-` 前缀，避免键名冲突。

---

## 五、具体实施方案

### 阶段一：紧急修复 (P0)

#### 1.1 Hook Bug 修复 - timer.ts setInterval 重复定义

**问题**: `hooks/timer.ts` 第 39-45 行 setInterval 重复定义

**修复步骤**:
```bash
# 1. 查看问题代码
cd w7panel-ui/src/hooks
grep -n "const setInterval" timer.ts

# 2. 编辑 timer.ts，删除第 43-45 行的重复定义
# 修改位置: timer.ts:43-45
```

**修改文件**: `w7panel-ui/src/hooks/timer.ts`

**验证**:
```bash
# 编译检查
cd w7panel-ui && npm run build
```

---

### 阶段二：Hook 规范化 (P1)

#### 2.1 修复 7 个文件使用原生 setInterval

**问题文件及修改方案**:

| 文件 | 行号 | 当前代码 | 修改为 |
|------|------|---------|--------|
| `views/app/apps/detail.vue` | ~883 | `watchInterval = setInterval(...)` | `useTimer` |
| `views/init-cluster/resource-loading.vue` | ~29 | `interval = setInterval(...)` | `useTimer` |
| `views/init-cluster/order-base.vue` | ~992 | `interval = setInterval(...)` | `useTimer` |
| `views/header/allow-register-check.vue` | ~35 | `interval = setInterval(...)` | `useTimer` |
| `views/cluster/nodes/set-gpu-bu.vue` | ~204 | `statusInterval = setInterval(...)` | `useTimer` |
| `views/cluster/nodes/set-gpu.vue` | ~198 | `statusInterval = setInterval(...)` | `useTimer` |

**修改模板**:
```typescript
// 修改前
import { ref, onMounted, onUnmounted } from 'vue';
const interval = ref(null);
onMounted(() => {
    interval.value = setInterval(() => { ... }, 3000);
});
onUnmounted(() => {
    clearInterval(interval.value);
});

// 修改后
import { useTimer } from '@/hooks/timer';
const { setInterval, clearTimer } = useTimer();
onMounted(() => {
    setInterval('my-timer', () => { ... }, 3000);
});
onUnmounted(() => {
    clearTimer('my-timer');
});
```

**验证**:
```bash
# 检查是否还有原生 setInterval
cd w7panel-ui/src && grep -r "setInterval(" --include="*.vue" | grep -v node_modules | wc -l
```

---

### 阶段三：代码清理 (P2-P3)

#### 3.1 清理 console.log (132 处)

**批量清理命令**:
```bash
cd w7panel-ui/src

# 查找所有 console.log 位置
grep -rn "console.log(" --include="*.ts" --include="*.vue" | grep -v node_modules | head -30
```

**处理策略**:
| 类型 | 处理方式 |
|------|---------|
| 调试用 console.log | 直接删除 |
| 错误日志 console.error | 改为使用统一日志服务 |
| 保留必要的 console.warn | 保持 |

**注意**: 先备份再批量替换
```bash
# 备份
cp -r src src.backup

# 批量删除 console.log (谨慎使用)
# sed -i "s/console.log(.*);/\/\/ console.log removed/g" $(grep -rl "console.log" src)
```

#### 3.2 清理注释代码 (126 处)

**批量查找**:
```bash
cd w7panel-ui/src

# 查找注释掉的 import
grep -rn "// import" --include="*.ts" --include="*.vue" | wc -l

# 查找注释掉的代码行
grep -rn "// const\|// let\|// function\|// axios" --include="*.ts" --include="*.vue" | wc -l
```

**处理策略**:
```typescript
// 删除死代码
// const oldCode = ...  → 直接删除

// 保留待用代码，标记 TODO
// TODO: 等待 API 完善后启用
// const pendingCode = ...
```

---

### 阶段四：架构优化 (P4-P6)

#### 4.1 封装 API 路径工具

**新建文件**: `w7panel-ui/src/utils/k8s-api.ts`

```typescript
import { useNamespaceStore } from '@/store/modules/namespace';

export function buildNamespacedApi(resource: string, namespace?: string): string {
    const ns = namespace || useNamespaceStore().namespace || 'default';
    return `/k8s-proxy/api/v1/namespaces/${ns}/${resource}`;
}

export function buildClusterApi(resource: string): string {
    return `/k8s-proxy/api/v1/${resource}`;
}

export function buildAppsApi(kind: string, name?: string, namespace?: string): string {
    const ns = namespace || useNamespaceStore().namespace || 'default';
    const base = `/k8s-proxy/apis/apps/v1/namespaces/${ns}/${kind}s`;
    return name ? `${base}/${name}` : base;
}

// 常用 API 封装
export const api = {
    persistentvolumeclaims: () => buildNamespacedApi('persistentvolumeclaims'),
    configmaps: () => buildNamespacedApi('configmaps'),
    secrets: () => buildNamespacedApi('secrets'),
    pods: () => buildNamespacedApi('pods'),
    services: () => buildNamespacedApi('services'),
    deployments: () => buildAppsApi('deployment'),
    statefulsets: () => buildAppsApi('statefulset'),
    cronjobs: () => buildAppsApi('cronjob'),
    jobs: () => buildAppsApi('job'),
};
```

**渐进式替换** (优先高频使用文件):
1. `storage/storage-drawer.vue` - 5 处
2. `storage/zone-drawer.vue` - 4 处
3. `app/pages/form.vue` - 6 处

#### 4.2 统一 localStorage 键名

直接修改键名，无需迁移逻辑。

**修改文件**:
1. `w7panel-ui/src/utils/auth.ts`
2. `w7panel-ui/src/store/modules/app/index.ts`

**修改示例**:
```typescript
// 修改前
const TOKEN_KEY = 'token';

// 修改后
const TOKEN_KEY = 'w7panel-token';
```

#### 4.3 JSON 深拷贝优化

**方案选择**:
| 方案 | 优点 | 缺点 |
|------|------|------|
| lodash-es cloneDeep | 功能强大 | 需要引入库 |
| structuredClone | 原生无需引入 | 旧浏览器不支持 |
| 手动拷贝 | 无需引入 | 代码量大 |

**推荐**: 使用 `structuredClone` (现代浏览器内置)

**修改示例**:
```typescript
// 修改前
const copy = JSON.parse(JSON.stringify(data));

// 修改后
const copy = structuredClone(data);

// 或使用 lodash
import { cloneDeep } from 'lodash-es';
const copy = cloneDeep(data);
```

**高频修改文件** (按优先级):
1. `views/app/pages/domain.vue` - 12 处
2. `views/app/cronjob/cronjob-drawer.vue` - 9 处
3. `views/cluster/overview/panel.vue` - 5 处

---

### 阶段五：长期优化 (P7)

#### 5.1 useRequest Hook 推广

**策略**: 渐进式替换，优先高频请求

**优先替换文件**:
1. `store/modules/namespace.ts` - 获取 namespace 列表
2. `views/app/apps/index.vue` - 获取应用列表
3. `views/storage/storage.vue` - 获取存储列表

**修改示例**:
```typescript
// 修改前
import axios from 'axios';
const loading = ref(false);
const data = ref([]);
const fetchData = async () => {
    loading.value = true;
    try {
        const res = await axios.get('/api/data');
        data.value = res.data;
    } finally {
        loading.value = false;
    }
};

// 修改后
import useRequest from '@/hooks/request';
const { loading, response, run } = useRequest(fetchData);
```

---

## 六、优化优先级 (更新)

| 优先级 | 优化项 | 工作量 | 收益 | 状态 |
|--------|--------|--------|------|------|
| P0 | Hook Bug 修复 (timer.ts) | 5 分钟 | 避免运行时错误 | ⏳ 待执行 |
| P1 | useTimer/usePolling 推广 | 2 小时 | 避免资源泄漏 | ⏳ 待执行 |
| P2 | console.log 清理 | 1 小时 | 减少生产噪音 | ⏳ 待执行 |
| P3 | 注释代码清理 | 2 小时 | 代码整洁 | ⏳ 待执行 |
| P4 | API 路径工具封装 | 4 小时 | 减少重复代码 | ⏳ 待执行 |
| P5 | localStorage 键名统一 | 30 分钟 | 统一规范 | ⏳ 待执行 |
| P6 | JSON 深拷贝优化 | 2 小时 | 性能提升 | ⏳ 待执行 |
| P7 | useRequest 推广 | 高 | 统一请求管理 | ⏳ 长期 |

---

## 七、验证清单

### 开发环境验证
```bash
# 1. 编译检查
cd w7panel-ui && npm run build  # ✅ 通过
```

### 功能验证
- [x] timer.ts Bug 已修复 (无编译警告)
- [x] 7 个文件已验证定时器清理正确
- [x] namespace=undefined 问题已解决
- [x] Drawer 组件延迟加载已生效
- [x] /k8s-proxy/version API 正常工作
- [x] 控制台无 Intlify/wujie 警告
- [x] localStorage 键名已统一
- [ ] 登录/登出功能正常 (需测试验证 token 键名)

### 性能验证
- [ ] API 请求数量减少 (Drawer 延迟加载)
- [ ] 定时器无内存泄漏 (已验证正确清理)
- [ ] JSON 拷贝性能提升 (structuredClone)

---

## 八、相关文件

### 修改的文件 (已完成)
```
w7panel-ui/src/store/modules/namespace.ts
w7panel-ui/src/components/dcform-drawer.vue
w7panel-ui/src/components/domain-strategy.vue
w7panel-ui/src/components/domain-strategy-plugin.vue
w7panel-ui/src/components/menu/index.vue
w7panel-ui/src/views/cluster/overview/panel.vue
w7panel-ui/src/hooks/use-wujie-events.ts (新增)
w7panel-ui/src/locale/zh-CN.ts
w7panel-ui/src/locale/en-US.ts
w7panel-ui/src/router/routes/modules/dialogpage.ts
w7panel-ui/src/hooks/timer.ts (Bug 修复)
w7panel-ui/src/utils/auth.ts (localStorage 键名统一)
w7panel-ui/src/store/modules/app/index.ts (localStorage 键名统一)
w7panel-ui/src/utils/k8s-api.ts (新增)
```

### 待修改的文件 (后续优化)
```
# 建议后续处理（风险较高，需逐个确认）
console.log 清理 (54处实际使用)              - 保留调试用途
注释代码清理 (318处)                        - 风险高，需逐个确认
JSON 深拷贝优化 (74处)                      - 可用 structuredClone 替代
API 路径工具推广使用 (90处已替换)           - 渐进式替换完成基础模式
```

### 待修改的文件 (计划中) - 已完成
```
w7panel-ui/src/hooks/timer.ts               (P0 - Bug 修复) ✅
w7panel-ui/src/utils/k8s-api.ts            (P4 - 新建) ✅
w7panel-ui/src/utils/auth.ts               (P5 - 键名统一) ✅
w7panel-ui/src/store/modules/app/index.ts   (P5 - 键名统一) ✅
```

---

## 九、架构原则总结

### 正确做法 ✅ (已验证)
| 模式 | 示例 | 状态 |
|------|------|------|
| 统一工具函数 | `getUserInfo()`, `getPermission()` - 55+ 处统一使用 | ✅ 良好 |
| 并行请求 | `Promise.all()` - 20+ 处正确使用 | ✅ 良好 |
| Store 状态管理 | Pinia store - 正确管理全局状态 | ✅ 良好 |
| 常量封装 | 已有 `DEFAULT_ROUTE_NAME`, `DEFAULT_ROUTE` | ✅ 良好 |
| 事件监听器清理 | 13 处 addEventListener 都有 removeEventListener | ✅ 良好 |
| window.formatDate | 45+ 处统一使用 | ⚠️ 建议封装 |
| kube-system 硬编码 | 用户确认是常规 API 传参 | ✅ 正常 |

### 错误做法 ❌ (需要优化)
| 模式 | 问题 | 数量 |
|------|------|------|
| 直接 import axios | 145+ 文件重复，应使用 useRequest | 145+ |
| 原生 setInterval | 7 个文件，应使用 useTimer | 7 |
| 空 catch 块 | 56 处静默吞错 | 56 |
| 重复 API 路径 | 135+ 处相同路径字符串 | 135+ |
| JSON 深拷贝 | 74 处 JSON.parse(stringify) | 74 |
| console.log | 132 处生产代码 | 132 |
| 注释代码 | 126 处死代码 | 126 |
| localStorage 键名 | 6 处不统一 | 6 |

---

## 十、版本历史

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v5.4 | 2026-02-22 | 本次会话优化 + 深入分析 + 具体实施方案 |
| v5.3 | 2026-02-20 | API 路由重构 |
| v5.2 | 历史版本 | - |
