# W7Panel 路由重构 v5.1 - 问题汇总与新方案

**版本**: 5.1  
**日期**: 2026-02-21  
**状态**: 待实施

---

## 一、问题汇总

### 1.1 已修复

| 问题 | 状态 |
|------|------|
| index.html app id | ✅ 已修复 (`k8soffline` → `w7panel`) |
| 前端构建 | ✅ 已完成 |
| 后端路由 | ✅ 基本完成 |

### 1.2 未完成 - 前端 API 路径替换

**问题**: 前端仍有大量旧 API 路径未替换

**统计**:
- `/k8s/` 路径: 65 处未替换
- 主要集中在:
  - `/k8s/console/info` - 大量使用
  - `/k8s/exec*` - WebSocket 调用
  - `/k8s/v1/namespaces/...` - K8s API 代理

**影响**:
- 登录失败 (404 错误)
- 页面空白/功能异常

---

## 二、详细问题分析

### 2.1 前端需要替换的路径

#### 2.1.1 Console API (约 15 处)
```bash
/k8s/console/info    → /panel-api/v1/console/info
/k8s/console/login  → /panel-api/v1/console/login
/k8s/console/bind   → /panel-api/v1/console/bind
/k8s/console/...    → /panel-api/v1/console/...
```

#### 2.1.2 Exec API (约 10 处)
```bash
/k8s/exec    → /panel-api/v1/exec
/k8s/exec2   → /panel-api/v1/exec2
/k8s/nodetty → /panel-api/v1/nodetty
```

#### 2.1.3 K8s 代理 API (约 40 处)
```bash
/k8s/v1/namespaces/...  → /k8s-proxy/v1/namespaces/...
/k8s/api/v1/...       → /k8s-proxy/api/v1/...
```

---

## 三、实施方案

### 3.1 后端 - 旧路由兼容

**问题**: 前端仍在使用 `/k8s/console/*` 等旧路径

**解决方案**: 在 NoRoute 中添加旧路由兼容处理

```go
// main.go NoRoute 处理
engine.NoRoute(
    middleware2.Html{}.Process,
    middleware2.Auth{}.Process,
    middleware2.K8sFilter{}.Process,
    controller.Proxy{}.ProxyK8sWithLegacy,
)
```

在 `ProxyK8sWithLegacy` 中:
1. 如果路径是 `/k8s/console/*` → 转发到 `/panel-api/v1/console/*`
2. 如果路径是 `/k8s/exec*` → 转发到 `/panel-api/v1/exec*`
3. 其他 → 原有 K8s proxy 处理

### 3.2 前端 - 批量替换 (推荐方案)

使用 sed 批量替换:

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# Console API
sed -i 's|"/k8s/console/|"/panel-api/v1/console/|g' $(find src -name "*.ts" -o -name "*.vue")

# Exec API
sed -i 's|"/k8s/exec|"/panel-api/v1/exec|g' $(find src -name "*.ts" -o -name "*.vue")

# K8s proxy API
sed -i 's|"/k8s/v1/|"/k8s-proxy/v1/|g' $(find src -name "*.ts" -o -name "*.vue")
```

---

## 四、实施检查清单

### 4.1 后端 (快速修复 - 兼容模式)

- [ ] 修改 `app/application/http/controller/proxy.go` - 添加旧路由兼容处理
- [ ] 后端重新编译
- [ ] 测试登录流程

### 4.2 前端 (完整替换)

- [ ] 批量替换 Console API 路径
- [ ] 批量替换 Exec API 路径
- [ ] 批量替换 K8s 代理 API 路径
- [ ] 前端重新构建
- [ ] 复制到 dist 目录
- [ ] UI 自动化测试

---

## 五、测试用例

### 5.1 API 测试
```bash
# 测试旧路由兼容
curl http://localhost:8080/k8s/console/info
# 应返回正确响应 (通过兼容处理)

# 测试新路由
curl http://localhost:8080/panel-api/v1/console/info
# 应返回正确响应
```

### 5.2 UI 测试
```bash
# 打开登录页面
agent-browser open http://localhost:8080
sleep 5

# 填写登录信息
agent-browser snapshot -i
agent-browser fill @e1 "admin"
agent-browser fill @e2 "123456"
agent-browser click @e4  # 登录按钮
sleep 3

# 验证登录成功
agent-browser get url
# 应跳转到应用页面
```

---

**下一步**: 等待确认后开始实施
