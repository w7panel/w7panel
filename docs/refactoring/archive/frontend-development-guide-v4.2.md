# W7Panel 路由重构前端开发详细指南 v4.2

**版本**: 4.2.0  
**创建日期**: 2026-02-21  
**状态**: 待开发

---

## 一、概述

本文档是前端开发的详细实施步骤，基于 v4.2 方案：
- 面板业务接口：`/panel-api/v1/*`
- K8s 代理接口：`/k8s-proxy/*`

### 前端修改统计

| 类型 | 数量 | 说明 |
|------|------|------|
| 面板业务 API | ~149 处 | 需替换 `/k8s/*` → `/panel-api/v1/*` |
| K8s Core API | ~300 处 | 需替换 `/api/v1/` → `/k8s-proxy/api/v1/` |
| K8s CRD API | ~360 处 | 需替换 `/apis/` → `/k8s-proxy/apis/` |
| 涉及文件 | 50+ | .vue 和 .ts 文件 |

---

## 二、前置条件

### 2.1 备份代码

```bash
export BACKUP_DIR=/home/wwwroot/w7panel-dev/backup_repo
mkdir -p $BACKUP_DIR
cp -r /home/wwwroot/w7panel-dev/w7panel-ui $BACKUP_DIR/w7panel_frontend_$(date +%Y%m%d%H%M%S)
echo "备份完成: $BACKUP_DIR/w7panel_frontend_$(date +%Y%m%d%H%M%S)"
```

### 2.2 准备工作

```bash
# 进入前端目录
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 确保依赖已安装
npm install

# 记录需要修改的文件数量
find src -name "*.ts" -o -name "*.vue" | wc -l
```

---

## 三、开发步骤

### Step 1: 修改 API 配置

**文件**: `src/config/api.ts`

#### 1.1 修改基础路径配置

```typescript
// 修改前
export const API_BASE_PATH = '/k8s';

// 修改后 - 支持双 base URL
export const API_BASE_PATH = '/panel-api/v1';
export const K8S_PROXY_BASE_PATH = '/k8s-proxy';
```

#### 1.2 新增 K8s 代理 API 配置

在 `API_PATHS` 中添加 K8s 代理相关配置：

```typescript
// 新增 K8s 代理 API 配置
export const K8S_API_PATHS = {
  // Core API
  NAMESPACES: '/api/v1/namespaces',
  PODS: '/api/v1/namespaces/{namespace}/pods',
  SERVICES: '/api/v1/namespaces/{namespace}/services',
  NODES: '/api/v1/nodes',
  CONFIGMAPS: '/api/v1/namespaces/{namespace}/configmaps',
  SECRETS: '/api/v1/namespaces/{namespace}/secrets',
  EVENTS: '/api/v1/namespaces/{namespace}/events',
  PVCS: '/api/v1/namespaces/{namespace}/persistentvolumeclaims',
  
  // Apps API
  DEPLOYMENTS: '/apis/apps/v1/namespaces/{namespace}/deployments',
  STATEFULSETS: '/apis/apps/v1/namespaces/{namespace}/statefulsets',
  DAEMONSETS: '/apis/apps/v1/namespaces/{namespace}/daemonsets',
  JOBS: '/apis/batch/v1/namespaces/{namespace}/jobs',
  CRONJOBS: '/apis/batch/v1/namespaces/{namespace}/cronjobs',
  INGRESSES: '/apis/networking.k8s.io/v1/namespaces/{namespace}/ingresses',
  
  // Custom Resources
  APPGROUPS: '/apis/appgroup.w7.cc/v1alpha1/namespaces/{namespace}/appgroups',
  MICROAPPS: '/apis/microapp.w7.cc/v1alpha1/namespaces/{namespace}/microapps',
  ZPK_LIST: '/panel-api/v1/zpk/list',
  
  // Metrics
  METRICS_NODES: '/apis/metrics.k8s.io/v1beta1/nodes',
  METRICS_PODS: '/apis/metrics.k8s.io/v1beta1/namespaces/{namespace}/pods',
  
  // Storage
  STORAGECLASSES: '/apis/storage.k8s.io/v1/storageclasses',
  
  // Autoscaling
  HPA: '/apis/autoscaling/v2/namespaces/{namespace}/horizontalpodautoscalers',
} as const;

// 构建完整 K8s 代理路径
export function buildK8sApiPath(pathKey: keyof typeof K8S_API_PATHS, params?: Record<string, string>): string {
  let path = K8S_API_PATHS[pathKey];
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      path = path.replace(`{${key}}`, value);
    });
  }
  return `${K8S_PROXY_BASE_PATH}${path}`;
}

// 修改 buildApiPath 支持面板业务 API
export function buildApiPath(pathKey: keyof typeof API_PATHS): string {
  return `${API_BASE_PATH}${API_PATHS[pathKey]}`;
}
```

---

### Step 2: 修改 Interceptor

**文件**: `src/api/interceptor.ts`

#### 2.1 修改 Token 刷新路径

```typescript
// 修改前 (第 108 行)
if(!error?.config?.customToken && error?.config?.url!='/k8s/refresh-token2'){

// 修改后
if(!error?.config?.customToken && error?.config?.url!='/panel-api/v1/auth/refresh-token2'){
```

#### 2.2 修改 Token 刷新请求

```typescript
// 修改前 (第 109 行)
let t = await axios.post('/k8s/refresh-token2',{token: getRefreshToken()},{

// 修改后
let t = await axios.post('/panel-api/v1/auth/refresh-token2',{token: getRefreshToken()},{
```

#### 2.3 修改登录错误处理

```typescript
// 修改前 (第 153 行)
}else if(error?.response?.status == 500 && error?.config?.url=='/k8s/login'){

// 修改后
}else if(error?.response?.status == 500 && error?.config?.url=='/panel-api/v1/auth/login'){
```

---

### Step 3: 批量替换面板业务 API

**文件**: 多个 .vue 和 .ts 文件

#### 3.1 认证 API

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 登录相关
sed -i "s|'/k8s/login|'/panel-api/v1/auth/login|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/register|'/panel-api/v1/auth/register|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/refresh-token2|'/panel-api/v1/auth/refresh-token2|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/init-user|'/panel-api/v1/auth/init-user|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/userinfo|'/panel-api/v1/auth/userinfo|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/reset-password|'/panel-api/v1/auth/reset-password|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.2 Console API

```bash
# Console
sed -i "s|'/k8s/console/|'/panel-api/v1/console/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.3 文件操作 API

```bash
# WebDAV
sed -i "s|'/k8s/webdav-agent|'/panel-api/v1/files/webdav-agent|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/webdav/|'/panel-api/v1/files/webdav/|g" $(find src -name "*.ts" -o -name "*.vue")

# Compress
sed -i "s|'/k8s/compress-agent|'/panel-api/v1/files/compress-agent|g" $(find src -name "*.ts" -o -name "*.vue")

# Permission
sed -i "s|'/k8s/permission-agent|'/panel-api/v1/files/permission-agent|g" $(find src -name "*.ts" -o -name "*.vue")

# Download
sed -i "s|'/k8s/download/|'/panel-api/v1/download/|g" $(find src -name "*.ts" -o -name "*.vue")

# File operations
sed -i "s|'/k8s/cp'|'/panel-api/v1/cp'|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/mvpid|'/panel-api/v1/mvpid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/cppid|'/panel-api/v1/cppid|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.4 工具 API

```bash
# Captcha
sed -i "s|'/k8s/captcha|'/panel-api/v1/captcha|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/verify-captcha|'/panel-api/v1/verify-captcha|g" $(find src -name "*.ts" -o -name "*.vue")

# Utils
sed -i "s|'/k8s/pinyin|'/panel-api/v1/pinyin|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/dnsip|'/panel-api/v1/dnsip|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/dns-cname|'/panel-api/v1/dns-cname|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/myip|'/panel-api/v1/myip|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/db-conn-test|'/panel-api/v1/db-conn-test|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/ping-etcd|'/panel-api/v1/ping-etcd|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.5 应用操作 API

```bash
# Pod/Exec
sed -i "s|'/k8s/pid|'/panel-api/v1/pid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/nodepid|'/panel-api/v1/nodepid|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/tty|'/panel-api/v1/tty|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/nodetty|'/panel-api/v1/nodetty|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/exec|'/panel-api/v1/exec|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/exec2|'/panel-api/v1/exec2|g" $(find src -name "*.ts" -o -name "*.vue")

# YAML
sed -i "s|'/k8s/yaml|'/panel-api/v1/yaml|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/rollback|'/panel-api/v1/rollback|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/kcompose|'/panel-api/v1/kcompose|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.6 Helm/Zpk API

```bash
# Helm
sed -i "s|'/api/v1/helm/|'/panel-api/v1/helm/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/api/v1/app-info|'/panel-api/v1/app-info|g" $(find src -name "*.ts" -o -name "*.vue")

# Zpk
sed -i "s|'/api/v1/zpk/|'/panel-api/v1/zpk/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.7 K3k API

```bash
# K3k
sed -i "s|'/k8s/k3k/|'/panel-api/v1/k3k/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/idc-list|'/panel-api/v1/idc-list|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.8 集群管理 API

```bash
# Longhorn
sed -i "s|'/k8s/longhorn/|'/panel-api/v1/longhorn/|g" $(find src -name "*.ts" -o -name "*.vue")

# K3s
sed -i "s|'/k8s/k3s/|'/panel-api/v1/k3s/|g" $(find src -name "*.ts" -o -name "*.vue")

# KubeBlocks
sed -i "s|'/k8s/kubeblocks/|'/panel-api/v1/kubeblocks/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.9 GPU API

```bash
# GPU
sed -i "s|'/k8s/gpu/|'/panel-api/v1/gpu/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.10 Microapp API

```bash
# Microapp
sed -i "s|'/k8s/microapp/|'/panel-api/v1/microapp/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.11 Metrics API (面板自采集)

```bash
# Metrics (面板采集)
sed -i "s|'/k8s/metrics/usage/|'/panel-api/v1/metrics/usage/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/installed|'/panel-api/v1/metrics/installed|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|'/k8s/metrics/state|'/panel-api/v1/metrics/state|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 3.12 特殊路由

```bash
# Kubeconfig
sed -i "s|'/k8s/kubeconfig|'/panel-api/v1/kubeconfig|g" $(find src -name "*.ts" -o -name "*.vue")

# Proxy URL
sed -i "s|'/k8s/proxy-url|'/panel-api/v1/proxy-url|g" $(find src -name "*.ts" -o -name "*.vue")
```

---

### Step 4: 批量替换 K8s 代理 API

#### 4.1 K8s Core API

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 替换 /api/v1/ 为 /k8s-proxy/api/v1/
sed -i "s|'/api/v1/|'/k8s-proxy/api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")

# 注意：上面会误伤 /panel-api/v1/，需要修正
sed -i "s|'/k8s-proxy/panel-api/v1/|'/panel-api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 4.2 K8s CRD API

```bash
# 替换 /apis/ 为 /k8s-proxy/apis/
sed -i "s|'/apis/|'/k8s-proxy/apis/|g" $(find src -name "*.ts" -o -name "*.vue")
```

#### 4.3 特殊 K8s 代理路径

```bash
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

---

### Step 5: 检查并修正

#### 5.1 检查替换遗漏

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 检查是否还有遗留的 /k8s/ 路径（排除 noauth）
grep -r "'/k8s/" src --include="*.ts" --include="*.vue" | grep -v "k8s-proxy" | head -20

# 检查是否还有遗留的 /api/v1/ 路径
grep -r "'/api/v1/" src --include="*.ts" --include="*.vue" | grep -v "k8s-proxy" | head -20

# 检查是否还有遗留的 /apis/ 路径
grep -r "'/apis/" src --include="*.ts" --include="*.vue" | grep -v "k8s-proxy" | head -20
```

#### 5.2 修正替换错误

```bash
# 修正误替换
sed -i "s|'/k8s-proxy/panel-api/v1/|'/panel-api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")

# 修正路径格式问题（如 /k8s-proxy//api/ → /k8s-proxy/api/）
sed -i "s|/k8s-proxy/+|/k8s-proxy/|g" $(find src -name "*.ts" -o -name "*.vue")
```

---

### Step 6: 编译验证

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 编译前端
npm run build

if [ $? -eq 0 ]; then
    echo "✅ 前端编译成功"
else
    echo "❌ 前端编译失败"
    exit 1
fi
```

---

## 四、常见问题处理

### 4.1 路径格式问题

```bash
# 检查双斜杠
grep -r "//" src --include="*.ts" --include="*.vue" | grep "k8s-proxy"

# 修正
sed -i "s|/k8s-proxy/+|/k8s-proxy/|g" $(find src -name "*.ts" -o -name "*.vue")
```

### 4.2 遗漏的面板业务路径

```bash
# 检查未被替换的 /k8s/ 路径
grep -r "'/k8s/" src --include="*.ts" --include="*.vue" | grep -v "k8s-proxy"
```

### 4.3 特殊字符问题

```bash
# 检查反引号替换
sed -i "s|\`/k8s/|\`/panel-api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|\`/api/v1/|\`/k8s-proxy/api/v1/|g" $(find src -name "*.ts" -o -name "*.vue")
sed -i "s|\`/apis/|\`/k8s-proxy/apis/|g" $(find src -name "*.ts" -o -name "*.vue")
```

---

## 五、文件修改清单

### 5.1 核心配置文件

| 文件 | 修改内容 |
|------|---------|
| `src/config/api.ts` | API_BASE_PATH, 新增 K8S_API_PATHS |
| `src/api/interceptor.ts` | 3 处 URL 替换 |

### 5.2 批量替换统计

| 类型 | 数量 |
|------|------|
| .vue 文件 | ~40 |
| .ts 文件 | ~15 |
| 总替换次数 | ~800+ |

---

## 六、验证测试

### 6.1 本地开发测试

```bash
# 启动开发服务器
npm run dev

# 验证登录页面
# 访问 http://localhost:5173

# 验证 API 调用
# 打开浏览器控制台，检查 network 中的请求路径
```

### 6.2 构建测试

```bash
# 构建生产版本
npm run build

# 检查构建产物
ls -la dist/
```

### 6.3 UI 自动化测试

```bash
cd /home/wwwroot/w7panel-dev/tests

# 运行 UI 测试
bash panel-ui-test.sh all
```

---

## 七、注意事项

1. **sed 替换风险**: 批量替换可能误伤，替换后必须检查
2. **双斜杠问题**: 注意 `/k8s-proxy//api/` 格式错误
3. **特殊字符**: 反引号、模板字符串中的路径也需要替换
4. **顺序重要**: 先替换面板业务 API，再替换 K8s 代理 API
5. **编译验证**: 替换后必须编译验证

---

*本文档是前端开发详细步骤，请按顺序执行。完成后请进行完整的 UI 测试验证。*
