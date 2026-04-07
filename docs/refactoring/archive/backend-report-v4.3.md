# 路由重构 v4.3 报告

**日期**: 2026-02-21

**版本**: v4.3

---

## 1. 概述

本次重构完成了以下目标：

1. **统一 API 路由前缀**：
   - 面板业务 API: `/panel-api/v1/*`
   - K8s 代理 API: `/k8s-proxy/*`
   - 兼容旧路由: `/k8s/*` (通过 NoRoute)

2. **K8sResponseFilter 中间件**：
   - 已创建基础框架
   - 过滤逻辑验证有效（managedFields 成功移除）
   - ⚠️ 由于 HTTP 响应流式写入问题，暂未启用

---

## 2. 修改的文件

### 2.1 后端修改

| 文件 | 修改内容 |
|------|---------|
| `main.go` | 添加 `/k8s-proxy/*` 路由，修复 HTML 中间件排除规则 |
| `common/middleware/html.go` | 排除 `/k8s-proxy/*` 和 `/k8s/*` 路径 |
| `common/middleware/k8s_response_filter.go` | 新增 K8sResponseFilter 中间件 |
| `app/application/http/controller/proxy.go` | 路径归一化处理 |

### 2.2 前端修改（之前已完成）

| 文件 | 修改内容 |
|------|---------|
| `src/config/api.ts` | `API_BASE_PATH = '/panel-api/v1'` |
| `src/api/interceptor.ts` | 路径替换 |
| 多个 .vue/.ts 文件 | API 路径替换 (~50 文件) |

---

## 3. 测试结果

```
=== API 测试结果 ===

1. Panel API - Login:                ✅ Success
2. K8s Proxy - /k8s-proxy/api/v1/namespaces: ✅ 14 namespaces
3. K8s Proxy - /k8s-proxy/api/v1/pods:     ✅ 108 pods  
4. K8s Proxy - /k8s-proxy/api/v1/services: ✅ 25 services
5. Legacy route - /k8s/api/v1/namespaces:   ✅ 14 namespaces

managedFields: ⚠️ Present (not filtered - see known issues)
```

---

## 4. 已知问题

### 4.1 K8sResponseFilter 未启用

**问题**: 
- 过滤逻辑已验证有效（日志显示 `hasManagedFields: false` after filtering）
- 但由于 Gin 中间件与 K8s reverse proxy 的响应处理方式冲突，导致响应为空

**原因**:
- K8s client.Proxy 使用 httputil.ReverseProxy 直接写入 ResponseWriter
- 中间件的 ResponseWrapper 无法正确拦截和修改响应

**解决方案（待定）**:
1. 使用自定义 RoundTripper 包装 K8s client
2. 在 controller 层直接处理响应过滤
3. 使用 HTTP hijacking 完全控制连接

---

## 5. 下一步计划

1. 修复 K8sResponseFilter 中间件的响应拦截问题
2. 更新前端 API 文档
3. 完整的 UI 测试验证

---

## 6. 备份

修改前代码已备份到：
- Backend: `/backup_repo/w7panel_backend_20260221065524`
- Frontend: `/backup_repo/w7panel_frontend_20260221070130`
