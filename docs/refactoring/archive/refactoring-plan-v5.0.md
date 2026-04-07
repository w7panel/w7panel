# W7Panel 路由重构 v5.0 - 问题分析与新方案

**版本**: 5.0  
**日期**: 2026-02-21  
**状态**: 待实施

---

## 一、问题分析

### 1.1 已完成工作

| 任务 | 状态 | 说明 |
|------|------|------|
| 后端 `/panel-api/v1/*` 路由 | ✅ 完成 | 各 Provider 路由前缀已修改 |
| 后端 `/k8s-proxy/*` 路由 | ✅ 完成 | NoRoute + explicit route |
| 前端 API 路径替换 | ⚠️ 部分完成 | 大部分已替换，需验证 |
| HTML 中间件排除 | ✅ 完成 | 排除 `/k8s-proxy/*` 路径 |
| K8sResponseFilter | ⚠️ 未启用 | 逻辑有效但响应处理有问题 |

### 1.2 未完成工作

#### 问题 1: 旧路由兼容不符合规范

**当前实现**:
- 只有 NoRoute 处理 `/k8s/*` 路径
- NoRoute 将所有未匹配路由转发到 K8s proxy

**问题**:
- 旧的面板业务路由（如 `/k8s/login`）无法正确匹配
- 缺少从旧路由到新路由的重定向/别名处理

**预期行为**:
- `/k8s/login` → 301 重定向到 `/panel-api/v1/auth/login`
- 或 `/k8s/login` → 内部转发到 `/panel-api/v1/auth/login`
- `/k8s/api/v1/namespaces` → 转发到 K8s proxy

#### 问题 2: 前端 index.html 未修改

**当前**:
```html
<div id="k8soffline"></div>
```

**应改为**:
```html
<div id="w7panel"></div>
```

**原因**: 根据 AGENTS.md 中的命名规范，应使用 `w7panel-*` 前缀

#### 问题 3: 前端 UI 空白

**可能原因**:
1. API 路径仍存在未替换的旧路径
2. Vue 应用的 mount point 不匹配
3. JavaScript 加载错误

**需要排查**:
- 检查浏览器控制台错误
- 验证 Vue 应用是否正确挂载
- 确认 API 请求是否成功

---

## 二、新方案 v5.0

### 2.1 路由兼容方案

#### 方案 A: 使用 Gin 中间件进行路径重写

在请求进入时，将旧路径转换为新路径：

```go
// common/middleware/legacy_redirect.go
func LegacyRedirect() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        
        // 面板业务 API 重定向
        redirects := map[string]string{
            "/k8s/login":                     "/panel-api/v1/auth/login",
            "/k8s/register":                  "/panel-api/v1/auth/register",
            "/k8s/refresh-token2":           "/panel-api/v1/auth/refresh-token2",
            "/k8s/userinfo":                 "/panel-api/v1/auth/userinfo",
            "/k8s/console/":                 "/panel-api/v1/console/",
            // ... 其他
        }
        
        if newPath, ok := redirects[path]; ok {
            c.Redirect(http.StatusMovedPermanently, newPath)
            c.Abort()
            return
        }
        
        // K8s 代理路径重写 (/k8s/api/* -> /k8s-proxy/api/*)
        if strings.HasPrefix(path, "/k8s/api/") {
            newPath := "/k8s-proxy" + strings.TrimPrefix(path, "/k8s")
            c.Request.URL.Path = newPath
        }
        
        c.Next()
    }
}
```

#### 方案 B: 在 NoRoute 中处理

在 NoRoute 处理器中先判断是否为旧的面板业务路由：

```go
func (self Proxy) ProxyK8s(c *gin.Context) {
    path := c.Request.URL.Path
    
    // 1. 检查是否为旧的面板业务路由
    if isLegacyPanelRoute(path) {
        // 转发到新的面板路由处理
        newPath := convertToNewPanelRoute(path)
        c.Request.URL.Path = newPath
        // 调用对应的 Handler
    }
    
    // 2. 检查是否为 K8s API 路径
    if isK8sApiPath(path) {
        // 转发到 K8s proxy
    }
}
```

### 2.2 前端修复

#### 2.2.1 修改 index.html

```bash
# 修改 app id
sed -i 's/id="k8soffline"/id="w7panel"/g' w7panel-ui/index.html
sed -i 's/id="k8soffline"/id="w7panel"/g' w7panel-ui/dist/index.html
```

#### 2.2.2 排查 UI 空白问题

1. 检查浏览器控制台错误
2. 验证 API 请求路径
3. 检查 Vue 应用挂载

---

## 三、实施清单

### 3.1 后端任务

- [ ] 创建 `w7panel/common/middleware/legacy_redirect.go`
- [ ] 在 main.go 中注册 legacy_redirect 中间件
- [ ] 测试旧路由重定向

### 3.2 前端任务

- [ ] 修改 index.html app id
- [ ] 重新构建前端
- [ ] 复制到 dist 目录
- [ ] UI 测试验证

### 3.3 排查任务

- [ ] 使用 agent-browser 打开页面
- [ ] 检查控制台错误
- [ ] 验证 API 请求
- [ ] 确认 Vue 挂载

---

## 四、测试用例

### 4.1 API 测试

```bash
# 旧路由重定向测试
curl -I http://localhost:8080/k8s/login
# 应返回 301 或正确响应

# 新路由测试
curl http://localhost:8080/panel-api/v1/auth/login

# K8s proxy 测试
curl http://localhost:8080/k8s-proxy/api/v1/namespaces
```

### 4.2 UI 测试

```bash
# 使用 agent-browser 打开页面
agent-browser open http://localhost:8080

# 检查控制台错误
agent-browser console

# 检查页面内容
agent-browser eval "document.getElementById('w7panel')"
```

---

## 五、影响分析

| 变更 | 影响范围 | 风险 |
|------|---------|------|
| 路由重定向 | 所有 API 调用 | 低 - 重定向逻辑简单 |
| index.html 修改 | 前端加载 | 中 - 需验证 Vue 挂载 |
| legacy 中间件 | 所有请求 | 低 - 仅路径匹配 |

---

**下一步**: 等待确认后开始实施 v5.0 方案
