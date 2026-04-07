# 重构方案 v5.5 - API 请求工具统一

**日期**: 2026-02-22

**目标**: 统一 K8s API 和面板 API 的请求方式，封装公共前缀，便于统一维护

---

## 背景

### 问题

当前代码中存在大量硬编码的 API 路径：

| 类型 | 示例 | 数量 |
|------|------|------|
| K8s API | `axios.get('/k8s-proxy/api/v1/namespaces/...')` | 596处 |
| 面板 API | `axios.get('/panel-api/v1/gpu/config')` | 171处 |

**问题**：
- 前缀硬编码，变更时需要改大量文件
- 鉴权等公共逻辑分散在各处
- 代码冗余，不够简洁

### 解决方案

创建统一的 API 请求工具，封装公共前缀：

```typescript
// Before
axios.get('/k8s-proxy/api/v1/namespaces/'+ ns +'/pods')
axios.get('/panel-api/v1/gpu/config')

// After
k8sproxy.get(`/api/v1/namespaces/${ns}/pods`)
panelApi.get('/gpu/config')
```

---

## 设计方案

### 1. 统一请求工具

**文件**: `src/utils/api.ts`

```typescript
import axios from 'axios';

// K8s 集群 API 前缀
const K8S_PROXY_PREFIX = '/k8s-proxy';

// 面板业务 API 前缀（v1 版本号放工具里，便于未来升级 v2）
const PANEL_API_PREFIX = '/panel-api/v1';

// K8s 集群代理请求工具
export const k8sproxy = {
    get: (path: string, config?: any) => axios.get(`${K8S_PROXY_PREFIX}${path}`, config),
    post: (path: string, data?: any, config?: any) => axios.post(`${K8S_PROXY_PREFIX}${path}`, data, config),
    patch: (path: string, data?: any, config?: any) => axios.patch(`${K8S_PROXY_PREFIX}${path}`, data, config),
    put: (path: string, data?: any, config?: any) => axios.put(`${K8S_PROXY_PREFIX}${path}`, data, config),
    delete: (path: string, config?: any) => axios.delete(`${K8S_PROXY_PREFIX}${path}`, config),
};

// 面板业务 API 请求工具
export const panelApi = {
    get: (path: string, config?: any) => axios.get(`${PANEL_API_PREFIX}${path}`, config),
    post: (path: string, data?: any, config?: any) => axios.post(`${PANEL_API_PREFIX}${path}`, data, config),
    patch: (path: string, data?: any, config?: any) => axios.patch(`${PANEL_API_PREFIX}${path}`, data, config),
    put: (path: string, data?: any, config?: any) => axios.put(`${PANEL_API_PREFIX}${path}`, data, config),
    delete: (path: string, config?: any) => axios.delete(`${PANEL_API_PREFIX}${path}`, config),
};
```

### 2. 保留现有的 k8sApi 工具函数

对于高频使用的特定 API（如 podsLog、serviceProxy），保留现有的便捷函数：

```typescript
// src/utils/k8s-api.ts
// 仅保留高频工具函数，不做全量替换
export const k8sApi = {
    podsLog: (name: string, namespace?: string) => buildNamespacedApi(`pods/${name}/log`, namespace),
    serviceProxy: (name: string, port?: string, namespace?: string) => { ... },
    // ...
};
```

### 3. 版本号管理策略

| 场景 | 方案 | 示例 |
|------|------|------|
| 面板 API 升级 v2 | 新增 `panelApiV2` 工具 | `panelApiV2.get('/users')` |
| K8s API 版本 | 由 K8s 自身管理 | `/apis/apps/v1/...` |

---

## 替换规则

### K8s API 替换

```typescript
// Before
axios.get('/k8s-proxy/api/v1/namespaces/'+ ns +'/pods')
axios.patch('/k8s-proxy/api/v1/nodes/'+ name, data)
axios.post('/k8s-proxy/apis/appgroup.w7.cc/...', data)

// After
k8sproxy.get(`/api/v1/namespaces/${ns}/pods`)
k8sproxy.patch(`/api/v1/nodes/${name}`, data)
k8sproxy.post('/apis/appgroup.w7.cc/...', data)
```

### 面板 API 替换

```typescript
// Before
axios.get('/panel-api/v1/gpu/config')
axios.post('/panel-api/v1/gpu/enabled-gpu?enabled='+v)

// After
panelApi.get('/gpu/config')
panelApi.post(`/gpu/enabled-gpu?enabled=${v}`)
```

---

## 实施步骤

### Step 1: 创建统一请求工具
- [x] 新建 `src/utils/api.ts`
- [ ] 保留 `src/utils/k8s-api.ts` 高频函数

### Step 2: 回滚之前的修改
- [x] 删除 k8s-api.ts 中冗余的 buildAppGroupApi、buildMicroAppApi 等函数
- [x] 回滚 set-gpu.vue 中已添加的导入

### Step 3: 替换 K8s API（约596处）
- [ ] 批量替换 `axios.get('/k8s-proxy` → `k8sproxy.get('`
- [ ] 批量替换 `axios.post('/k8s-proxy` → `k8sproxy.post('
- [ ] 批量替换 `axios.patch('/k8s-proxy` → `k8sproxy.patch('
- [ ] 批量替换 `axios.put('/k8s-proxy` → `k8sproxy.put('
- [ ] 批量替换 `axios.delete('/k8s-proxy` → `k8sproxy.delete('

### Step 4: 替换面板 API（约171处）
- [ ] 批量替换 `axios.get('/panel-api/v1` → `panelApi.get('
- [ ] 批量替换 `axios.post('/panel-api/v1` → `panelApi.post('
- [ ] 批量替换 `axios.patch('/panel-api/v1` → `panelApi.patch('
- [ ] 批量替换 `axios.put('/panel-api/v1` → `panelApi.put('
- [ ] 批量替换 `axios.delete('/panel-api/v1` → `panelApi.delete('

### Step 5: 验证
- [ ] 前端编译通过
- [ ] UI 功能测试通过

---

## 受影响文件统计

### 高频文件（需重点测试）

| 文件 | K8s API | 面板 API |
|------|---------|----------|
| views/app/pages/files.vue | ~50 | ~20 |
| views/storage/disk.vue | ~40 | ~5 |
| views/system/users/users.vue | ~30 | ~2 |
| views/system/usergroup/list.vue | ~25 | ~1 |
| views/cluster/overview/panel.vue | ~25 | ~10 |
| views/system/usermanage/quota.vue | ~15 | ~3 |
| components/store-install.vue | ~15 | ~10 |

### 替换后效果

| 场景 | 修改前 | 修改后 |
|------|--------|--------|
| K8s 路由变更 | 改 596 处 | 改 1 处 (api.ts) |
| 面板路由变更 | 改 171 处 | 改 1 处 (api.ts) |
| 鉴权逻辑变更 | 改所有文件 | 改 1 处 (api.ts) |

---

## 回滚方案

如需回滚，使用以下命令：

```bash
# 撤销 k8sproxy 替换
sed -i "s/k8sproxy\.get('/axios.get('\/k8s-proxy/g" $(grep -rl "k8sproxy.get" src/)
sed -i "s/k8sproxy\.post('/axios.post('\/k8s-proxy/g" $(grep -rl "k8sproxy.post" src/)
# ... 其他方法同理

# 撤销 panelApi 替换
sed -i "s/panelApi\.get('/axios.get('\/panel-api\/v1/g" $(grep -rl "panelApi.get" src/)
# ... 其他方法同理
```

---

## 相关文档

- v5.4: `docs/refactoring/refactoring-plan-v5.4.md`
- v5.3: `docs/refactoring/refactoring-plan-v5.3.md`
