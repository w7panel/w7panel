# W7Panel 路由重构 v4.2 - 完整实施指南

**版本**: 4.2.0  
**创建日期**: 2026-02-21  
**状态**: 规划中

---

## 一、代码深入分析

### 1.1 接口分类总结

通过分析代码，将接口分为以下几类：

| 分类 | 特征 | 示例 | 路由前缀 |
|------|------|------|---------|
| **面板业务接口** | 面板自有的业务逻辑 | 登录、验证码、文件管理、Helm操作 | `/panel-api/v1/*` |
| **面板自采集数据** | 面板采集的监控数据 | cgroup metrics、usage | `/panel-api/v1/*` |
| **K8s 代理透传** | NoRoute 直接转发 | `/api/v1/*`、`/apis/*` | `/k8s-proxy/*` |
| **K8s 条件转发** | Proxy middleware 条件转发 | metrics/node、metrics/pod | `/k8s-proxy/*` |
| **K8s 业务封装** | Controller 封装但调用 K8s API | namespaces、helm releases | `/panel-api/v1/*` |

### 1.2 可优化/合并的接口

#### 1.2.1 重复的用户信息接口

| 接口 | Controller | 功能 |
|------|------------|------|
| `/k8s/userinfo` | `k3k.K3k{}.Info` | K3k 集群用户信息 |
| `/k8s/k3k/info` | `k3k.K3k{}.Info` | 同上，K3k 集群信息 |
| `/panel-api/v1/auth/userinfo` | `auth.Auth{}.UserInfo` | 面板用户信息 |

**建议**: 合并为 `/panel-api/v1/auth/userinfo`，在 Controller 层判断返回不同数据

#### 1.2.2 重复的 Console 接口

| 接口 | 功能 |
|------|------|
| `/k8s/console/info` | Console 信息 |
| `/k8s/k3k/info` | K3k 集群信息（也返回 Console 相关） |

**建议**: 保持分离，因为是两个不同的业务域

#### 1.2.3 指标接口整合

| 接口 | 功能 |
|------|------|
| `/k8s/metrics/usage/normal` | 面板采集 - 资源使用 |
| `/k8s/metrics/usage/disk` | 面板采集 - 磁盘使用 |
| `/k8s/metrics/node` | K8s 代理 - Node 指标 |
| `/k8s/metrics/pod` | K8s 代理 - Pod 指标 |

**建议**: 保持分离，因为数据来源不同（面板采集 vs K8s API）

---

## 二、路由对照表（最终版）

### 2.1 面板业务接口 → `/panel-api/v1/*`

| Provider | 旧路径 | 新路径 |
|----------|--------|--------|
| **auth** | | |
| | `/k8s/login` | `/panel-api/v1/auth/login` |
| | `/k8s/register` | `/panel-api/v1/auth/register` |
| | `/k8s/refresh-token2` | `/panel-api/v1/auth/refresh-token2` |
| | `/k8s/init-user` | `/panel-api/v1/auth/init-user` |
| | `/k8s/reset-password` | `/panel-api/v1/auth/reset-password` |
| | `/k8s/reset-password-current` | `/panel-api/v1/auth/reset-password-current` |
| | `/k8s/userinfo` | `/panel-api/v1/auth/userinfo` |
| **Console** | | |
| | `/k8s/console/oauth` | `/panel-api/v1/console/oauth` |
| | `/k8s/console/login` | `/panel-api/v1/console/login` |
| | `/k8s/console/bind` | `/panel-api/v1/console/bind` |
| | `/k8s/console/info` | `/panel-api/v1/console/info` |
| | `/k8s/console/register-to-console` | `/panel-api/v1/console/register-to-console` |
| | `/k8s/console/import-cert` | `/panel-api/v1/console/import-cert` |
| | `/k8s/console/verify-cert` | `/panel-api/v1/console/verify-cert` |
| | `/k8s/console/import-cert-console` | `/panel-api/v1/console/import-cert-console` |
| **应用操作** | | |
| | `/k8s/pid` | `/panel-api/v1/pid` |
| | `/k8s/nodepid` | `/panel-api/v1/nodepid` |
| | `/k8s/tty` | `/panel-api/v1/tty` |
| | `/k8s/nodetty` | `/panel-api/v1/nodetty` |
| | `/k8s/exec` | `/panel-api/v1/exec` |
| | `/k8s/exec2` | `/panel-api/v1/exec2` |
| | `/k8s/yaml` | `/panel-api/v1/yaml` |
| | `/k8s/rollback` | `/panel-api/v1/rollback` |
| | `/k8s/kcompose` | `/panel-api/v1/kcompose` |
| **工具** | | |
| | `/k8s/captcha` | `/panel-api/v1/captcha` |
| | `/k8s/verify-captcha` | `/panel-api/v1/verify-captcha` |
| | `/k8s/pinyin` | `/panel-api/v1/pinyin` |
| | `/k8s/dnsip` | `/panel-api/v1/dnsip` |
| | `/k8s/dns-cname` | `/panel-api/v1/dns-cname` |
| | `/k8s/myip` | `/panel-api/v1/myip` |
| | `/k8s/db-conn-test` | `/panel-api/v1/db-conn-test` |
| | `/k8s/ping-etcd` | `/panel-api/v1/ping-etcd` |
| **文件操作** | | |
| | `/k8s/webdav-agent/:pid/agent/*path` | `/panel-api/v1/files/webdav-agent/:pid/agent/*path` |
| | `/k8s/webdav-agent/:pid/subagent/:subpid/agent/*path` | `/panel-api/v1/files/webdav-agent/:pid/subagent/:subpid/agent/*path` |
| | `/k8s/compress-agent/:pid/compress` | `/panel-api/v1/files/compress-agent/:pid/compress` |
| | `/k8s/compress-agent/:pid/extract` | `/panel-api/v1/files/compress-agent/:pid/extract` |
| | `/k8s/compress-agent/:pid/subagent/:subpid/compress` | `/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/compress` |
| | `/k8s/compress-agent/:pid/subagent/:subpid/extract` | `/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/extract` |
| | `/k8s/permission-agent/:pid/chmod` | `/panel-api/v1/files/permission-agent/:pid/chmod` |
| | `/k8s/permission-agent/:pid/chown` | `/panel-api/v1/files/permission-agent/:pid/chown` |
| | `/k8s/permission-agent/:pid/subagent/:subpid/chmod` | `/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chmod` |
| | `/k8s/permission-agent/:pid/subagent/:subpid/chown` | `/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chown` |
| | `/k8s/download/*path` | `/panel-api/v1/download/*path` |
| | `/k8s/cp` | `/panel-api/v1/cp` |
| | `/k8s/mvpid` | `/panel-api/v1/mvpid` |
| | `/k8s/cppid` | `/panel-api/v1/cppid` |
| **Helm 应用** | | |
| | `/api/v1/helm/releases` | `/panel-api/v1/helm/releases` |
| | `/api/v1/helm/releases/:name` | `/panel-api/v1/helm/releases/:name` |
| | `/api/v1/app-info` | `/panel-api/v1/app-info` |
| **Zpk 应用** | | |
| | `/api/v1/zpk/*` | `/panel-api/v1/zpk/*` |
| **集群管理** | | |
| | `/k8s/longhorn/*` | `/panel-api/v1/longhorn/*` |
| | `/k8s/k3s/env/gogc` | `/panel-api/v1/k3s/env/gogc` |
| | `/k8s/kubeblocks/*` | `/panel-api/v1/kubeblocks/*` |
| **GPU** | | |
| | `/k8s/gpu/*` | `/panel-api/v1/gpu/*` |
| **K3k** | | |
| | `/k8s/k3k/*` | `/panel-api/v1/k3k/*` |
| | `/k8s/userinfo` | `/panel-api/v1/userinfo` |
| | `/k8s/idc-list` | `/panel-api/v1/idc-list` |
| **Microapp** | | |
| | `/k8s/microapp/top` | `/panel-api/v1/microapp/top` |
| | `/k8s/microapp/:name/proxy/*path` | `/panel-api/v1/microapp/:name/proxy/*path` |
| **面板自采集 Metrics** | | |
| | `/k8s/metrics/usage/normal` | `/panel-api/v1/metrics/usage/normal` |
| | `/k8s/metrics/usage/disk` | `/panel-api/v1/metrics/usage/disk` |
| | `/k8s/metrics/installed` | `/panel-api/v1/metrics/installed` |
| | `/k8s/metrics/state` | `/panel-api/v1/metrics/state` |
| **特殊** | | |
| | `/k8s/kubeconfig` | `/panel-api/v1/kubeconfig` |
| | `/k8s/noauth/api/v1/namespaces/:namespace/configmaps/:name` | `/panel-api/v1/noauth/namespaces/:namespace/configmaps/:name` |

---

### 2.2 K8s 代理接口 → `/k8s-proxy/*`

| 类型 | 旧路径 | 新路径 |
|------|--------|--------|
| **K8s Core API** | | |
| | `/api/v1/*` (未匹配) | `/k8s-proxy/api/v1/*` |
| **K8s CRD API** | | |
| | `/apis/*` (未匹配) | `/k8s-proxy/apis/*` |
| **Service 代理** | | |
| | `/k8s/v1/namespaces/:ns/services/:name/proxy/*path` | `/k8s-proxy/v1/namespaces/:ns/services/:name/proxy/*path` |
| | `/k8s/v1/namespaces/:ns/services/:name/proxy-root/*path` | `/k8s-proxy/v1/namespaces/:ns/services/:name/proxy-root/*path` |
| **Pod 代理** | | |
| | `/k8s/v1/namespaces/:ns/pods/:name/proxy/*path` | `/k8s-proxy/v1/namespaces/:ns/pods/:name/proxy/*path` |
| **通用代理** | | |
| | `/k8s/v1/:name/proxy/*path` | `/k8s-proxy/v1/:name/proxy/*path` |
| | `noauth/v1/:name/proxy/*path` | `/k8s-proxy/noauth/v1/:name/proxy/*path` |
| | `/k8s/v1/namespaces/:ns/services/:name/proxy-no/*path` | `/k8s-proxy/v1/namespaces/:ns/services/:name/proxy-no/*path` |
| **Metrics 条件转发** | | |
| | `/k8s/metrics/node` | `/k8s-proxy/metrics/node` |
| | `/k8s/metrics/pod` | `/k8s-proxy/metrics/pod` |
| | `/k8s/metrics/top/node` | `/k8s-proxy/metrics/top/node` |
| | `/k8s/metrics/namespace/:ns/pod` | `/k8s-proxy/metrics/namespace/:ns/pod` |
| | `/k8s/metrics/namespace/:ns/resource` | `/k8s-proxy/metrics/namespace/:ns/resource` |

---

## 三、前端修改清单

### 3.1 API 配置修改

**文件**: `w7panel-ui/src/config/api.ts`

```typescript
// 修改前
export const API_BASE_PATH = '/k8s';

// 修改后 - 支持双 base URL
export const API_BASE_PATH = '/panel-api/v1';
export const K8S_PROXY_BASE_PATH = '/k8s-proxy';

// 新增 K8s 代理 API 配置
export const K8S_API_PATHS = {
  // Core API
  NAMESPACES: '/api/v1/namespaces',
  PODS: '/api/v1/namespaces/{namespace}/pods',
  SERVICES: '/api/v1/namespaces/{namespace}/services',
  NODES: '/api/v1/nodes',
  CONFIGMAPS: '/api/v1/namespaces/{namespace}/configmaps',
  SECRETS: '/api/v1/namespaces/{namespace}/secrets',
  
  // Apps API
  DEPLOYMENTS: '/apis/apps/v1/namespaces/{namespace}/deployments',
  STATEFULSETS: '/apis/apps/v1/namespaces/{namespace}/statefulsets',
  DAEMONSETS: '/apis/apps/v1/namespaces/{namespace}/daemonsets',
  
  // Custom Resources
  APPGROUPS: '/apis/appgroup.w7.cc/v1alpha1/namespaces/{namespace}/appgroups',
  MICROAPPS: '/apis/microapp.w7.cc/v1alpha1/namespaces/{namespace}/microapps',
  
  // Metrics
  METRICS_NODES: '/apis/metrics.k8s.io/v1beta1/nodes',
  METRICS_PODS: '/apis/metrics.k8s.io/v1beta1/namespaces/{namespace}/pods',
} as const;

// 构建 K8s 代理 API 路径
export function buildK8sApiPath(pathKey: keyof typeof K8S_API_PATHS, params?: Record<string, string>): string {
  let path = K8S_API_PATHS[pathKey];
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      path = path.replace(`{${key}}`, value);
    });
  }
  return `${K8S_PROXY_BASE_PATH}${path}`;
}
```

### 3.2 需要批量替换的路径

#### 3.2.1 面板业务 API（149 处 → 替换）

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 认证 API
sed -i "s|'/k8s/login|'/panel-api/v1/auth/login|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/register|'/panel-api/v1/auth/register|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/refresh-token2|'/panel-api/v1/auth/refresh-token2|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/init-user|'/panel-api/v1/auth/init-user|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/userinfo|'/panel-api/v1/auth/userinfo|g" $(find src -name "*.ts" -o -name "*.vue")

# Console API
sed -i "s|'/k8s/console/|'/panel-api/v1/console/|g" $(find src -name "*.ts" -o -name "*.vue")

# 文件操作 API
sed -i "s|'/k8s/webdav-agent|'/panel-api/v1/files/webdav-agent|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/compress-agent|'/panel-api/v1/files/compress-agent|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/permission-agent|'/panel-api/v1/files/permission-agent|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/download/|'/panel-api/v1/download/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/cp'|'/panel-api/v1/cp'|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/mvpid|'/panel-api/v1/mvpid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/cppid|'/panel-api/v1/cppid|g" $(find src -name "*.ts" -o -name "*.vue")

# 工具 API
sed -i "s|'/k8s/captcha|'/panel-api/v1/captcha|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/verify-captcha|'/panel-api/v1/verify-captcha|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/pinyin|'/panel-api/v1/pinyin|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/dnsip|'/panel-api/v1/dnsip|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/dns-cname|'/panel-api/v1/dns-cname|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/myip|'/panel-api/v1/myip|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/db-conn-test|'/panel-api/v1/db-conn-test|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/ping-etcd|'/panel-api/v1/ping-etcd|g" $(find src -name "*.ts" -o -name "*.vue")

# 应用操作 API
sed -i "s|'/k8s/pid|'/panel-api/v1/pid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/nodepid|'/panel-api/v1/nodepid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/tty|'/panel-api/v1/tty|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/nodetty|'/panel-api/v1/nodetty|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/exec|'/panel-api/v1/exec|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/yaml|'/panel-api/v1/yaml|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/rollback|'/panel-api/v1/rollback|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/kcompose|'/panel-api/v1/kcompose|g" $(find src -name "*.ts" -o -name "*.vue")

# Helm API
sed -i "s|'/api/v1/helm/|'/panel-api/v1/helm/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/api/v1/app-info|'/panel-api/v1/app-info|g" $(find src -name "*.ts" -o -name "*.vue")

# Zpk API
sed -i "s|'/api/v1/zpk/|'/panel-api/v1/zpk/|g" $(find src -name "*.ts" -o -name "*.vue")

# K3k API
sed -i "s|'/k8s/k3k/|'/panel-api/v1/k3k/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/idc-list|'/panel-api/v1/idc-list|g" $(find src -name "*.ts" -o -name "*.vue")

# 集群管理 API
sed -i "s|'/k8s/longhorn/|'/panel-api/v1/longhorn/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/k3s/|'/panel-api/v1/k3s/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/kubeblocks/|'/panel-api/v1/kubeblocks/|g" $(find src -name "*.ts" -o -name "*.vue")

# GPU API
sed -i "s|'/k8s/gpu/|'/panel-api/v1/gpu/|g" $(find src -name "*.ts" -o -name "*.vue")

# Microapp API
sed -i "s|'/k8s/microapp/|'/panel-api/v1/microapp/|g" $(find src -name "*.ts" -o -name "*.vue")

# Metrics API (面板自采集)
sed -i "s|'/k8s/metrics/usage/|'/panel-api/v1/metrics/usage/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/installed|'/panel-api/v1/metrics/installed|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/state|'/panel-api/v1/metrics/state|g" $(find src -name "*.ts" -o -name "*.vue")

# 特殊
sed -i "s|'/k8s/kubeconfig|'/panel-api/v1/kubeconfig|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.2.2 K8s 代理 API（660 处 → 替换）

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# K8s Core API - 替换 /api/v1/ 为 /k8s-proxy/api/v1/
sed -i "s|'/api/v1/|'/k8s-proxy/api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")

# K8s CRD API - 替换 /apis/ 为 /k8s-proxy/apis/
sed -i "s|'/apis/|'/k8s-proxy/apis/|g" $(find src -name "*.ts" -o -name "*.vue")

# K8s 代理路径
sed -i "s|'/k8s/v1/namespaces|'/k8s-proxy/v1/namespaces|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/v1/pods|'/k8s-proxy/v1/pods|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/v1/nodes|'/k8s-proxy/v1/nodes|g" $(find src -name "*.ts" -o -name "*.vue")

# Metrics 条件转发
sed -i "s|'/k8s/metrics/node|'/k8s-proxy/metrics/node|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/pod|'/k8s-proxy/metrics/pod|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/top/node|'/k8s-proxy/metrics/top/node|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/namespace|'/k8s-proxy/metrics/namespace|g" $(find src -name "*.ts" -o -name "*.vue")

# noauth 路径
sed -i "s|'/k8s/noauth/|'/k8s-proxy/noauth/|g" $(find src -name "*.ts" -o -name "*.vue")
```

### 3.3 Interceptor 修改

**文件**: `w7panel-ui/src/api/interceptor.ts`

```typescript
// 修改 Token 刷新路径
// 修改前
if (error?.config?.url != '/k8s/refresh-token2')

// 修改后
if (error?.config?.url != '/panel-api/v1/auth/refresh-token2')

// 修改登录错误处理
// 修改前
error?.config?.url=='/k8s/login'

// 修改后
error?.config?.url=='/panel-api/v1/auth/login'
```

---

## 四、实施检查清单

### 4.1 后端修改

- [ ] 创建 `w7panel/common/middleware/k8s_response_filter.go`
- [ ] 修改 `w7panel/main.go` - 添加 `/k8s-proxy` 路由组
- [ ] 修改 `w7panel/app/auth/provider.go` - 路由前缀
- [ ] 修改 `w7panel/app/application/provider.go` - 路由前缀
- [ ] 修改 `w7panel/app/k3k/provider.go` - 路由前缀
- [ ] 修改 `w7panel/app/zpk/provider.go` - 路由前缀
- [ ] 修改 `w7panel/app/metrics/provider.go` - 路由前缀
- [ ] 后端编译验证

### 4.2 前端修改

- [ ] 修改 `w7panel-ui/src/config/api.ts` - API 配置
- [ ] 批量替换面板业务 API 路径
- [ ] 批量替换 K8s 代理 API 路径
- [ ] 修改 `w7panel-ui/src/api/interceptor.ts`
- [ ] 前端编译验证

### 4.3 测试验证

- [ ] API 登录测试
- [ ] API 用户信息测试
- [ ] API K8s 代理测试
- [ ] managedFields 过滤测试
- [ ] UI 登录流程测试
- [ ] UI 应用列表测试

---

## 五、文件修改统计

| 项目 | 修改文件数 | 说明 |
|------|-----------|------|
| 后端 | 7 | 5 个 Provider + 1 个中间件 + main.go |
| 前端 | 50+ | 批量替换后预估 |

---

*本方案已完成深入分析，包含前端详细的修改清单和批量替换命令。*
