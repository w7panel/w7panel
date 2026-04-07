# W7Panel 路由重构 v5.2 - 问题分析与解决方案

**版本**: 5.2  
**日期**: 2026-02-21  
**状态**: 待确认

---

## 一、当前问题

### 1.1 用户反馈

用户反馈：
1. **兼容旧路由属于未按重构规范执行** - 后端添加了旧路由兼容，但没有按规范执行
2. **前端UI测试空白** - 前端修改后没有进行UI测试验证

### 1.2 问题分析

#### 问题1：后端旧路由兼容不规范

**当前状态**：
- 后端在 `main.go` 中显式注册了 `/k8s/console/*` 等旧路由
- 但这种做法不符合重构规范（应该是前端全部替换，而不是后端兼容）

**问题**：
- 65+ 处前端代码仍使用旧路径
- 后端需要维护两套路由，增加维护成本
- 重构不完整

#### 问题2：前端UI测试缺失

**当前状态**：
- 前端构建后没有进行UI测试
- 无法验证登录流程是否正常
- 无法发现潜在的前端问题

---

## 二、解决方案选择

### 方案A：后端兼容 + 前端替换（推荐）

| 步骤 | 工作内容 | 预估工作量 |
|------|---------|-----------|
| 1 | 保留后端旧路由兼容（确保现有前端能工作） | 已完成 |
| 2 | 前端批量替换旧API路径 | 30分钟 |
| 3 | 前端重新构建 | 5分钟 |
| 4 | 复制到dist目录 | 2分钟 |
| 5 | UI自动化测试 | 20分钟 |

**优点**：
- 重构完整，前后端统一使用新路由
- 减少后端维护成本
- 符合重构规范

**缺点**：
- 需要修改前端代码
- 需要重新测试

### 方案B：仅后端兼容（快速修复）

| 步骤 | 工作内容 | 预估工作量 |
|------|---------|-----------|
| 1 | 完善后端旧路由兼容 | 10分钟 |
| 2 | 后端重新编译 | 2分钟 |
| 3 | 简单API测试 | 5分钟 |

**优点**：
- 快速，不需要修改前端
- 现有代码不需要大改

**缺点**：
- 重构不完整
- 65+旧路径仍在前端代码 处中
- 后续需要继续完成前端替换

---

## 三、实施方案（推荐方案A）

### 3.1 后端（保持现状）

后端已有旧路由兼容代码，保持不变：

```go
// main.go 中的旧路由兼容
legacyGroup := engine.Group("/k8s")
legacyGroup.Use(middleware2.Html{}.Process, middleware2.Auth{}.Process)
{
    legacyGroup.GET("/console/info", appauth.Console{}.Info)
    legacyGroup.GET("/console/login", appauth.Auth{}.ConsoleLogin)
    // ... 其他 console 路由
    legacyGroup.GET("/exec", controller.PodExec{}.Exec)
    legacyGroup.POST("/exec2", controller.PodExec{}.Exec)
    legacyGroup.GET("/nodetty", controller.PodExec{}.NodeTty)
}
```

### 3.2 前端（批量替换）

使用 sed 批量替换前端源码中的旧路径：

```bash
cd /home/wwwroot/w7panel-dev/w7panel-ui

# 1. Console API 路径替换
sed -i 's|"/k8s/console/|"/panel-api/v1/auth/console/|g' $(find src -name "*.ts" -o -name "*.vue")

# 2. Exec API 路径替换  
sed -i 's|"/k8s/exec|"/panel-api/v1/exec|g' $(find src -name "*.ts" -o -name "*.vue")
sed -i 's|"/k8s/exec2|"/panel-api/v1/exec2|g' $(find src -name "*.ts" -o -name "*.vue")
sed -i 's|"/k8s/nodetty|"/panel-api/v1/nodetty|g' $(find src -name "*.ts" -o -name "*.vue")
sed -i 's|"/k8s/tty|"/panel-api/v1/tty|g' $(find src -name "*.ts" -o -name "*.vue")

# 3. K8s 代理 API 路径替换
sed -i 's|"/k8s/v1/|"/k8s-proxy/v1/|g' $(find src -name "*.ts" -o -name "*.vue")
sed -i 's|"/k8s/api/v1/|"/k8s-proxy/api/v1/|g' $(find src -name "*.ts" -o -name "*.vue")
```

### 3.3 构建部署

```bash
# 1. 前端构建
cd /home/wwwroot/w7panel-dev/w7panel-ui
npm run build

# 2. 复制到 dist
rm -rf /home/wwwroot/w7panel-dev/dist/kodata/assets
cp -r /home/wwwroot/w7panel-dev/w7panel-ui/dist/* /home/wwwroot/w7panel-dev/dist/kodata/

# 3. 启动服务
cd /home/wwwroot/w7panel-dev/dist
CAPTCHA_ENABLED=false LOCAL_MOCK=true KO_DATA_PATH=/home/wwwroot/w7panel-dev/dist/kodata KUBECONFIG=/home/wwwroot/w7panel-dev/kubeconfig.yaml ./w7panel server:start
```

---

## 四、测试计划

### 4.1 API 测试

```bash
# 获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/panel-api/v1/auth/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=123456" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")

# 测试新路由
curl http://localhost:8080/panel-api/v1/auth/console/info -H "Authorization: Bearer $TOKEN"

# 测试旧路由（兼容）
curl http://localhost:8080/k8s/console/info -H "Authorization: Bearer $TOKEN"

# 测试 K8s 代理（新路由）
curl http://localhost:8080/k8s-proxy/api/v1/namespaces -H "Authorization: Bearer $TOKEN"
```

### 4.2 UI 自动化测试

```bash
# 打开登录页
agent-browser open "http://localhost:8080"
sleep 5

# 获取元素
agent-browser snapshot -i

# 登录
agent-browser fill @e1 "admin"
agent-browser fill @e2 "123456"
agent-browser click @e3

sleep 3

# 验证跳转
agent-browser get url
# 应跳转到 /app/apps
```

### 4.3 UI 审美检查

根据 AGENTS.md 第10节 UI设计规范检查：
- [ ] 主题一致性（深色/浅色主题）
- [ ] 布局合理性（无溢出）
- [ ] 组件规范（标签页、弹窗、侧边栏）
- [ ] 交互体验（空状态、加载反馈）
- [ ] 文字显示（中文、特殊字符）

---

## 五、检查清单

### 5.1 代码修改

- [ ] 后端旧路由兼容（已实现）
- [ ] 前端 Console API 路径替换
- [ ] 前端 Exec API 路径替换
- [ ] 前端 K8s 代理 API 路径替换

### 5.2 构建部署

- [ ] 前端构建
- [ ] 复制到 dist 目录
- [ ] 服务启动

### 5.3 测试验证

- [ ] API 测试通过
- [ ] UI 登录流程正常
- [ ] UI 审美检查通过

---

## 六、文档更新

完成重构后需要更新以下文档：

- [ ] AGENTS.md - 如有变更
- [ ] docs/changelog/{版本号}.md - 更新变更日志
- [ ] docs/refactoring/ - 添加本次实施报告

---

**请确认要采用的方案：**

- [ ] **方案A（推荐）**：后端兼容 + 前端完整替换
- [ ] **方案B**：仅后端兼容（快速修复）

确认后我再开始实施。
