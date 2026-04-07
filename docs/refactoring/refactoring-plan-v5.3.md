# API 路由重构与性能优化方案

**版本**: v5.3  
**日期**: 2026-02-22  
**主题**: API 路由重构与性能优化方案

---

## 一、架构原则

### 1.1 核心原则

| 原则 | 说明 |
|------|------|
| **直接访问 K8s API 是允许的** | `/k8s-proxy/` 是访问 K8s 资源的标准方式 |
| **需要未授权访问时才封装** | 只有需要公开访问（无需登录）的接口才封装 |
| **减少冗余** | 不要为了"封装"而封装，直接调用 K8s API 更高效 |
| **性能优先** | 减少中间层转发，直接调用 K8s API 性能更好 |
| **安全优先** | 未授权接口不能泄露敏感信息 |
| **代码简洁** | 及时清理死代码、废弃方法和注释，保持代码高效简洁 |
| **必要注释** | 需要说明的地方添加注释，方便其他开发人员理解 |
| **国际化完整** | 确保 i18n 翻译文件完整，避免运行时警告 |

### 1.2 代码整洁规范

```
✅ 鼓励：及时清理死代码
   - 删除不再使用的控制器、方法、变量
   - 移除废弃的代码注释
   - 保持代码库健康

❌ 禁止：保留死代码
   - 废弃的控制器文件
   - 不再调用的方法
   - 被注释掉的路由定义
   - 未使用的导入包

✅ 鼓励：完整的国际化
   - 新增菜单项需要添加翻译 key
   - 检查翻译文件是否包含所有文本
   - 避免运行时 i18n 警告

❌ 禁止：遗漏翻译
   - 菜单文本未添加到翻译文件
   - 按钮文本未翻译
   - 提示信息未翻译
```

### 1.3 死代码清理检查清单

| 检查项 | 说明 |
|--------|------|
| 控制器文件 | 是否有路由被注释但控制器方法仍存在？ |
| 方法调用 | 是否有方法定义但无任何调用？ |
| 路由定义 | 是否有废弃的路由注释？ |
| 导入包 | 是否有未使用的 import？ |

### 1.4 国际化检查清单

| 检查项 | 说明 |
|--------|------|
| 菜单文本 | 新增菜单是否添加到翻译文件？ |
| 按钮文本 | 按钮文字是否都有翻译 key？ |
| 提示信息 | 提示、警告、错误信息是否翻译？ |
| 占位符 | 输入框占位符是否翻译？ |

### 1.2 安全的 API 使用模式

```
✅ 直接使用 /k8s-proxy/ (推荐)
   - 获取 Pod 列表: GET /k8s-proxy/api/v1/namespaces/{ns}/pods
   - 获取 ConfigMap: GET /k8s-proxy/api/v1/namespaces/{ns}/configmaps/{name}
   - 获取 Secrets: GET /k8s-proxy/api/v1/namespaces/{ns}/secrets

✅ 需要未授权访问时使用 /panel-api/v1/noauth/ (仅用于公开接口)
   - 但必须：只返回必要的业务字段，过滤敏感信息
   - 不能：返回完整的 K8s 资源对象

❌ 错误示例：返回完整 ConfigMap
   {
     "kind": "ConfigMap",
     "metadata": { "name": "beian", "namespace": "default", ... },
     "data": { "icpnumber": "xxx", "internal_config": "xxx", ... }
   }

✅ 正确示例：只返回必要的公开字段
   {
     "icpnumber": "xxx",
     "number": "xxx", 
     "location": "xxx"
   }
```

---

## 二、安全问题（已修复）

### 2.1 问题描述

原有未授权接口返回**完整的 K8s ConfigMap 对象**，包含敏感信息：

```json
{
  "kind": "ConfigMap",
  "apiVersion": "v1",
  "metadata": {
    "name": "beian",
    "namespace": "default",
    "uid": "xxx",
    "creationTimestamp": "xxx",
    "annotations": { ... }
  },
  "data": {
    "icpnumber": "xxx",           // 需要
    "number": "xxx",               // 需要
    "location": "xxx",             // 需要
    "internal_config": "xxx",       // 不该暴露
    "debug_mode": "xxx",           // 不该暴露
    "custom_settings": "xxx"       // 不该暴露
  }
}
```

### 2.2 风险分析

| 风险类型 | 说明 |
|---------|------|
| **信息泄露** | 暴露内部配置、业务逻辑、调试开关等 |
| **安全绕过** | 攻击者可能通过配置信息发现系统漏洞 |
| **隐私问题** | 不必要的数据传输到客户端 |

### 2.3 受影响的接口（已修复）

| 原接口 | 问题 | 修复方案 |
|--------|------|---------|
| `/panel-api/v1/noauth/namespaces/:ns/configmaps/:name` | 返回完整 ConfigMap | 已废弃，使用新接口 |
| `/panel-api/v1/init-user` | 返回敏感配置 | 已删除，使用 `/noauth/site/init-user` |
| `/panel-api/v1/auth/init-user` (GET) | 返回敏感配置 | 已删除，使用 `/noauth/site/init-user` |

---

## 三、API 分层架构

```
/panel-api/v1/                  面板业务 API
├── /noauth/site/*               未授权公开接口 (安全)
│   ├── /beian                   备案信息 (只返回公开字段)
│   ├── /k3k-config              K3K 配置 (只返回公开字段)
│   └── /init-user               初始化配置 (只返回必要字段)
├── /auth/*                      认证接口 (需登录)
│   ├── /login                   登录
│   ├── /init-user (POST)        创建初始用户
│   ├── /userinfo                用户信息
│   └── /console/*               控制台接口
├── /helm/*                      Helm 管理
├── /files/*                     文件操作 (WebDAV)
├── /metrics/*                   监控数据
├── /gpu/*                       GPU 管理
├── /k3k/*                       K3K 集群管理
├── /zpk/*                       ZPK 应用管理
└── ...

/k8s-proxy/                     K8s 资源 API (需认证)
├── /api/v1/*                   K8s 核心 API
│   ├── /namespaces/{ns}/pods
│   ├── /namespaces/{ns}/services
│   ├── /namespaces/{ns}/configmaps
│   ├── /namespaces/{ns}/secrets
│   └── ...
└── /apis/*                     K8s CRD API
```

---

## 四、安全的未授权接口

### 4.1 设计原则

1. **最小化原则** - 只返回前端需要的字段
2. **过滤敏感信息** - 不返回 metadata、内部配置等
3. **返回业务数据** - 直接返回数据对象，而非 K8s 资源对象

### 4.2 已创建的安全接口

| 接口 | 认证 | 返回字段 | 用途 |
|------|------|---------|------|
| `GET /panel-api/v1/noauth/site/beian` | ❌ 无 | icpnumber, number, location | 备案信息 |
| `GET /panel-api/v1/noauth/site/k3k-config` | ❌ 无 | indexpage | K3K 公开配置 |
| `GET /panel-api/v1/noauth/site/init-user` | ❌ 无 | canInitUser, allowConsoleRegister, captchaEnabled | 初始化配置 |

### 4.3 init-user 接口说明

| 接口 | 方法 | 认证 | 用途 |
|------|------|------|------|
| `/panel-api/v1/noauth/site/init-user` | GET | ❌ 无 | 获取初始化配置（公开） |
| `/panel-api/v1/auth/init-user` | POST | ✅ 有 | 创建初始用户（需认证） |

---

## 五、K8s 资源分类

### 5.1 直接访问 (使用 `/k8s-proxy/`)

以下 K8s 资源**直接访问即可**：

| 资源类型 | API 路径示例 |
|---------|-------------|
| Pods | `/k8s-proxy/api/v1/namespaces/{ns}/pods` |
| Services | `/k8s-proxy/api/v1/namespaces/{ns}/services` |
| ConfigMaps | `/k8s-proxy/api/v1/namespaces/{ns}/configmaps/{name}` |
| Secrets | `/k8s-proxy/api/v1/namespaces/{ns}/secrets/{name}` |
| Events | `/k8s-proxy/api/v1/namespaces/{ns}/events` |
| Deployments | `/k8s-proxy/apis/apps/v1/namespaces/{ns}/deployments` |

### 5.2 需要封装的场景

| 场景 | 封装原因 |
|------|---------|
| **未授权公开接口** | 需要过滤敏感字段，只返回必要数据 |
| **业务逻辑聚合** | 需要组合多个 K8s 接口 |
| **自定义返回格式** | 需要转换 K8s 格式为业务格式 |

---

## 六、实施计划

### 6.1 修复安全问题

| 步骤 | 内容 | 状态 |
|------|------|------|
| 1 | 创建 `/panel-api/v1/noauth/site/beian` 接口 | ✅ 完成 |
| 2 | 创建 `/panel-api/v1/noauth/site/k3k-config` 接口 | ✅ 完成 |
| 3 | 创建 `/panel-api/v1/noauth/site/init-user` 接口 | ✅ 完成 |
| 4 | 前端登录页改用新接口 | ✅ 完成 |
| 5 | 前端初始化页面改用新接口 | ✅ 完成 |
| 6 | 移除旧的安全问题代码 | ✅ 完成 |

### 6.2 代码清理

| 清理项 | 说明 | 状态 |
|--------|------|------|
| 删除 `noauth.go` | 废弃的未授权接口控制器 | ✅ 完成 |
| 删除 `handleNoAuthRequest()` | 废弃的代理方法 | ✅ 完成 |
| 清理 `provider.go` 废弃注释 | 删除所有废弃代码注释 | ✅ 完成 |
| 修复 `init-user` 路由 | 删除旧的 auth 路由引用 | ✅ 完成 |

---

## 七、前端改动示例

```javascript
// ❌ 旧代码 - 暴露敏感信息
axios.get('/k8s-proxy/noauth/api/v1/namespaces/default/configmaps/beian')
  .then(res => {
    site.value = res.data?.data || {};
  })

// ✅ 新代码 - 只获取必要字段
axios.get('/panel-api/v1/noauth/site/beian')
  .then(res => {
    site.value = res.data || {};
  })
```

---

## 八、已完成的工作

| 任务 | 状态 |
|------|------|
| 前端 `/api/v1/` → `/k8s-proxy/api/v1/` | ✅ 330+ 处 |
| 前端 `/apis/` → `/k8s-proxy/apis/` | ✅ 24+ 处 |
| 创建安全未授权接口 | ✅ beian, k3k-config, init-user |
| 前端改用新安全接口 | ✅ |
| 删除 `noauth.go` 废弃文件 | ✅ |
| 删除 `handleNoAuthRequest()` 方法 | ✅ |
| 清理废弃注释代码 | ✅ |
| 修复 init-user 路由 | ✅ |
| 后端编译验证 | ✅ 通过 |
| 前端构建验证 | ✅ 通过 |

### 无需修改的接口

以下接口规划正确，**无需修改**：

| 接口类型 | 说明 |
|---------|------|
| `/k8s-proxy/*` | K8s 资源 API，需要认证，直接访问即可 |
| `/panel-api/v1/metrics/*` | 面板业务聚合接口，需要认证，返回聚合数据 |
| `/panel-api/v1/helm/*` | Helm 管理接口，需要认证 |
| `/panel-api/v1/files/*` | 文件操作接口，需要认证 |
| `/panel-api/v1/gpu/*` | GPU 管理接口，需要认证 |
| `/panel-api/v1/k3k/*` | K3K 集群管理接口，需要认证 |

---

## 九、最终验证结果

### 接口验证

| 接口 | 方法 | 认证 | 返回类型 | 状态 |
|------|------|------|---------|------|
| `/panel-api/v1/noauth/site/beian` | GET | ❌ 无 | 业务数据 | ✅ |
| `/panel-api/v1/noauth/site/k3k-config` | GET | ❌ 无 | 业务数据 | ✅ |
| `/panel-api/v1/noauth/site/init-user` | GET | ❌ 无 | 业务数据 | ✅ |
| `/panel-api/v1/login` | POST | ❌ 无 | 登录接口 | ✅ |
| `/panel-api/v1/auth/init-user` | POST | ✅ 有 | 创建用户 | ✅ |
| `/k8s-proxy/*` | * | ✅ 有 | K8s 资源 | ✅ |
| `/panel-api/v1/helm/*` | * | ✅ 有 | Helm 管理 | ✅ |
| `/panel-api/v1/metrics/*` | * | ✅ 有 | 监控数据 | ✅ |

### 构建验证

| 项目 | 状态 |
|------|------|
| 后端 Go 编译 | ✅ 通过 |
| 前端 Vite 构建 | ✅ 通过 |

### 测试发现的问题

| 问题 | 严重程度 | 状态 |
|------|---------|------|
| i18n 翻译缺失（大量 `[intlify] Not found 'xxx' key` 警告） | 低 | ⚠️ 已有问题 |
| `/panel-api/v1/helm/releases/w7panel-metrics` 500 错误 | 中 | ⚠️ 已有问题 |

**说明**：上述问题为已有问题，与本次 API 重构无关。

---

## 十、架构决策总结

```
┌─────────────────────────────────────────────────────────────────┐
│  架构决策 (安全优先 + 代码简洁 + 必要注释 + 国际化完整)            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ✅ 直接使用 K8s API (推荐)                                    │
│     /k8s-proxy/api/v1/namespaces/{ns}/configmaps/{name}       │
│                                                                 │
│  ✅ 未授权接口必须安全                                           │
│     - 只返回必要的业务字段                                       │
│     - 过滤 metadata、annotations 等                              │
│     - 不返回完整的 K8s 资源对象                                 │
│                                                                 │
│  ✅ 代码简洁无冗余                                               │
│     - 删除废弃的控制器和方法                                     │
│     - 清理废弃注释代码                                           │
│     - 死代码清理                                                │
│     - 未使用的导入包                                            │
│                                                                 │
│  ✅ 必要注释                                                    │
│     - 复杂业务逻辑添加说明                                       │
│     - 安全相关代码添加注释                                       │
│     - 特殊处理逻辑添加说明                                       │
│                                                                 │
│  ✅ 国际化完整                                                  │
│     - 菜单文本添加翻译 key                                      │
│     - 按钮文本添加翻译 key                                      │
│     - 提示信息添加翻译 key                                      │
│                                                                 │
│  ❌ 不要返回完整的 K8s 资源                                     │
│     - 暴露内部配置信息                                          │
│     - 泄露安全相关设置                                          │
│     - 不必要的敏感数据传输                                      │
│                                                                 │
│  ❌ 不要保留死代码                                              │
│     - 废弃的控制器文件                                          │
│     - 不再调用的方法                                            │
│     - 被注释掉的路由定义                                         │
│                                                                 │
│  ❌ 不要遗漏翻译                                                │
│     - 菜单文本                                                 │
│     - 按钮文本                                                 │
│     - 提示信息                                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```
