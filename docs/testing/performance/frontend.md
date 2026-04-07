# W7Panel 前端性能分析报告

**分析日期**: 2026-02-20  
**项目**: w7panel-ui (Vue 3 + Vite + Arco Design)  
**范围**: API 请求、组件渲染、状态管理、资源加载

---

## 📊 总体评估

| 类别 | 状态 | 说明 |
|------|------|------|
| 构建配置 | 🟢 良好 | 已做代码分割，有资源哈希 |
| API 请求 | 🟢 已优化 | 批量请求、全局超时、缓存机制 |
| 组件渲染 | 🟡 中等 | 大列表缺少虚拟滚动 |
| 状态管理 | 🟢 已优化 | 添加了 localStorage 缓存 |
| 内存安全 | 🟢 已优化 | 定时器统一管理、组件卸载清理 |

---

## ✅ 已修复问题

### 1. 串行 API 请求 → 批量请求

**位置**: `w7panel-ui/src/views/app/apps/index.vue` (行480-510)

**修复前**:
```javascript
for(let i in this.data){
    // 每个应用都单独请求一次 - 性能极差
    let { data } = await axios.get('/api/v1/zpk/upgrade-info?...');
    this.data[i].upgrade = { ...data };
}
```

**修复后**:
```javascript
// 使用批量请求替代串行请求，每批10个
const batchSize = 10;
for (const batch of batches) {
    const promises = batch.map(item => axios.get(...).then(...));
    const results = await Promise.all(promises);
}
```

**状态**: ✅ 已修复

---

### 2. 定时器泄漏 → 统一管理

**新增文件**: `src/hooks/timer.ts`

```typescript
import { useTimer, usePolling } from '@/hooks/timer';

// 使用示例
const { setInterval, clearTimer } = useTimer();

// 轮询场景
const { startPolling, stopPolling } = usePolling(callback, 5000);
```

**修复的组件**:
- `yaml-input.vue` - 添加了 timeout 和 editor 清理

**状态**: ✅ 已修复

---

### 3. useRequest Hook 增强

**位置**: `src/hooks/request.ts`

**新增功能**:
- ✅ 请求缓存（基于 API 函数字符串）
- ✅ 请求取消（AbortController）
- ✅ 自动重试机制
- ✅ 请求超时配置
- ✅ 组件卸载自动取消

```typescript
const { loading, response, run, cancel, refresh } = useRequest(api, {
    cache: true,           // 启用缓存
    cacheTime: 5 * 60 * 1000, // 5分钟
    retry: 3,               // 重试3次
    retryDelay: 1000,      // 1秒后重试
    timeout: 30000,        // 30秒超时
    onSuccess: (data) => {},
    onError: (err) => {},
});
```

**状态**: ✅ 已修复

---

### 4. namespace store 缓存

**位置**: `src/store/modules/namespace.ts`

**新增功能**:
- ✅ localStorage 缓存（默认5分钟）
- ✅ `fetchNamespaceList(forceRefresh)` 强制刷新
- ✅ `refreshNamespaceList()` 清缓存并刷新
- ✅ 加载状态和错误处理

```typescript
const namespaceStore = useNamespaceStore();

// 使用缓存
await namespaceStore.fetchNamespaceList();

// 强制刷新
await namespaceStore.fetchNamespaceList(true);
```

**状态**: ✅ 已修复

---

### 5. 用户登录并行请求

**位置**: `src/store/modules/user/index.ts`

**修复前**:
```javascript
await axios.get('/k8s/userinfo').then(res => { ... })
await axios.get("/k8s/console/info?code=test").then(res => { ... })
```

**修复后**:
```javascript
const [userInfoRes, consoleInfoRes] = await Promise.all([
    axios.get('/k8s/userinfo'),
    axios.get("/k8s/console/info?code=test")
]);
```

**状态**: ✅ 已修复

---

### 6. 全局请求超时

**位置**: `src/api/interceptor.ts`

```typescript
axios.defaults.timeout = 30000;
axios.defaults.timeoutErrorMessage = '请求超时，请稍后重试';
```

**状态**: ✅ 已修复

---

## 🚧 待修复问题

### 1. 大列表无虚拟滚动

**位置**:
- `apps/index.vue` - 应用列表
- `files.vue` - 文件列表

**影响**: 1000+ 条数据时 DOM 节点过多，页面卡顿

**建议**: 使用 `vue-virtual-scroller` 或 `@vueuse/core` 的 `useVirtualList`

---

### 2. v-for key 不稳定

约 300 处 v-for，部分使用 index 作为 key：

```vue
<!-- 不推荐 -->
<tr v-for="(item,index) in list" :key="index">

<!-- 推荐 -->
<tr v-for="item in list" :key="item.id">
```

---

### 3. 图片懒加载

应用图标等静态资源建议使用懒加载：

```vue
<a-image :src="item.icon" loading="lazy" />
```

---

## 📈 性能优化汇总

| 问题 | 状态 | 修复方式 |
|------|------|----------|
| 串行 API 请求 | ✅ 已修复 | 批量 Promise.all |
| 定时器泄漏 | ✅ 已修复 | useTimer composable |
| useRequest 缺陷 | ✅ 已修复 | 重写 Hook |
| namespace 重复请求 | ✅ 已修复 | localStorage 缓存 |
| 登录串行请求 | ✅ 已修复 | Promise.all 并行 |
| 缺少请求超时 | ✅ 已修复 | 全局默认超时 |
| 大列表无虚拟滚动 | ⏳ 待修复 | - |
| v-for key 不稳定 | ⏳ 待修复 | - |
| 图片懒加载 | ⏳ 待修复 | - |

---

## 📝 修改文件清单

| 文件 | 修改类型 | 说明 |
|------|----------|------|
| `src/hooks/request.ts` | 新增/重写 | useRequest Hook 增强 |
| `src/hooks/timer.ts` | 新增 | 定时器管理 composable |
| `src/store/modules/namespace.ts` | 修改 | 添加缓存机制 |
| `src/store/modules/user/index.ts` | 修改 | 登录并行请求 |
| `src/api/interceptor.ts` | 修改 | 全局超时配置 |
| `src/views/app/apps/index.vue` | 修改 | 批量请求 |
| `src/components/yaml-input.vue` | 修改 | 清理定时器 |

---

## 🎯 后续优化建议

### 高优先级
1. 添加虚拟滚动（列表页面）
2. 优化 v-for key

### 中优先级
1. 图片懒加载
2. 大型 JSON 解析优化

### 低优先级
1. 提取静态资源为常量
2. 计算属性缓存优化
