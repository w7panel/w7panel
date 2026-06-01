# 开发指南

本文档汇总 W7Panel 开发相关资料，面向后端、前端、Web IDE、部署联调和接口文档维护。

## 快速入口

### API 文档

`docs/development/api/` 按 `w7panel-server/app` 下的模块目录拆分接口文档。每个模块文档包含接口功能、请求参数和响应参数说明。

| 模块 | 文档 | 说明 |
|------|------|------|
| 通用约定与开发规范 | [api/README.md](./api/README.md) | API 路由分层、鉴权、响应格式、安全、前后端同步和测试规范 |
| application | [api/application.md](./api/application.md) | `w7panel-server/app/application` 接口 |
| auth | [api/auth.md](./api/auth.md) | `w7panel-server/app/auth` 接口 |
| k3k | [api/k3k.md](./api/k3k.md) | `w7panel-server/app/k3k` 接口 |
| k3s-registry | [api/k3s-registry.md](./api/k3s-registry.md) | `w7panel-server/app/k3s-registry` 接口 |
| metrics | [api/metrics.md](./api/metrics.md) | `w7panel-server/app/metrics` 接口 |
| zpk | [api/zpk.md](./api/zpk.md) | `w7panel-server/app/zpk` 接口 |

### 前端文档

| 文档 | 说明 |
|------|------|
| [frontend/README.md](./frontend/README.md) | `w7panel-ui` 前端文档入口，包含页面、组件、API、状态、Wujie 和提交流程规范 |
| [frontend/components.md](./frontend/components.md) | 前端公共组件和业务复用组件说明 |
| [frontend/api-methods.md](./frontend/api-methods.md) | API 路径配置、axios 拦截器和业务 API 方法说明 |
| [frontend/wujie-events.md](./frontend/wujie-events.md) | Wujie 微前端事件、参数、回调响应和调用示例 |

## 项目结构

```text
$BASE_DIR/
├── w7panel-server/                  # 后端源码，Go + Gin + w7-rangine-go
│   ├── app/                         # 后端应用模块
│   │   ├── application/             # 面板核心业务 API
│   │   ├── auth/                    # 认证 API
│   │   ├── k3k/                     # K3k API
│   │   ├── k3s-registry/            # K3s 镜像仓库 API
│   │   ├── metrics/                 # 指标 API
│   │   └── zpk/                     # ZPK 应用 API
│   ├── common/                      # 公共服务、中间件和工具
│   ├── install/                     # 安装资源和 Helm Charts
│   ├── scripts/                     # 构建脚本
│   └── kodata/                      # 静态资源
├── w7panel-ui/                      # 前端源码，Vue 3 + TypeScript + Arco Design
│   ├── src/api/                     # 前端 API 调用封装
│   ├── src/views/                   # 页面
│   ├── src/components/              # 公共组件
│   ├── src/hooks/                   # Hooks
│   └── src/router/                  # 路由配置
├── codeblitz/                       # Web IDE 源码
├── dist/                            # 编译输出目录
├── docs/                            # 项目文档
└── tests/                           # 测试脚本
```

说明：历史文档中可能仍出现 `w7panel/` 作为后端目录名；当前开发文档按 `w7panel-server/` 表述。

## 开发规范

### 后端

- 控制器目录：`w7panel-server/app/{module}/http/`
- 路由注册：`w7panel-server/app/{module}/provider.go`
- 公共服务：`w7panel-server/common/service/`
- 中间件：`w7panel-server/common/middleware/`
- 日志使用 `log/slog` 键值对格式：

```go
slog.Info("操作成功", "user", userID, "action", "create")
```

### 前端

- API 封装放在 `w7panel-ui/src/api/`
- 页面放在 `w7panel-ui/src/views/`
- 通用组件放在 `w7panel-ui/src/components/`
- 业务逻辑复用优先放在 `w7panel-ui/src/hooks/`
- 可复用组件、API 方法和 Wujie 微前端事件说明见 [frontend/README.md](./frontend/README.md)

前端 API 示例：

```typescript
// src/api/example.ts
export function getExampleList() {
  return axios.get('/panel-api/v1/example');
}
```

## 开发流程

### 1. 开发前同步代码

```bash
cd $BASE_DIR
git fetch origin dev-v1
git pull origin dev-v1
```

如果本地分支和远端分支已分叉，需要先明确使用 merge、rebase 或 fast-forward 策略，再执行 `git pull`。

### 2. 后端 API 修改

修改、新增或删除后端接口时，需要同步检查：

| 检查项 | 说明 |
|--------|------|
| 路由前缀 | 面板业务 API 使用 `/panel-api/v1/`，K8s 代理使用 `/k8s-proxy/` |
| 鉴权 | 除公开接口外，业务接口需要用户 token |
| 前端调用 | 检查 `w7panel-ui/src/api/` 和相关页面是否需要同步 |
| 文档 | 更新 `docs/development/api/{module}.md` |
| 版本日志 | 更新 `docs/changelog/{version}.md` |

常用搜索：

```bash
rg "接口路径或字段名" w7panel-ui/src
rg "接口路径或字段名" w7panel-server
```

### 3. 前端修改

后端字段、URL、鉴权方式变化后，前端需要同步检查：

| 检查项 | 说明 |
|--------|------|
| API 封装 | `src/api/` 中的路径和参数是否匹配后端 |
| 类型定义 | TypeScript 类型是否覆盖新增或变更字段 |
| 页面使用 | 页面是否仍使用旧字段、旧路径或旧状态含义 |
| 构建验证 | 修改后执行前端构建或相关检查 |

### 4. 文档维护

新增接口文档时按模块目录命名：

```text
docs/development/api/{app子目录名}.md
```

文档内容至少包含：

- 接口功能
- 请求方法和路径
- 请求参数说明
- 响应参数说明
- 必要的响应示例或特殊逻辑说明

## 常用命令

### 后端

```bash
cd $BASE_DIR/w7panel-server
go build -o $BASE_DIR/dist/w7panel .
```

### 前端

```bash
cd $BASE_DIR/w7panel-ui
npm run dev
npm run build
```

### 启动服务

默认开发和部署测试使用测试模式：

```bash
cd $BASE_DIR/dist
export LOCAL_MOCK=true
export CAPTCHA_ENABLED=false
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
./w7panel-ctl.sh start
```

### 测试

```bash
cd $BASE_DIR/tests
bash compress-ui-test.sh all
```

## 调试

### 查看日志

```bash
tail -f /tmp/w7panel.log
```

### API 调试

请求面板业务接口时携带用户 token：

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/panel-api/v1/example"
```

### 常见问题排查

| 问题 | 检查方向 |
|------|----------|
| 接口 401 | token 是否存在、是否走了公开接口、LOCAL_MOCK 是否错误绕过认证假设 |
| 接口 404 | 路由是否注册、前缀是否为 `/panel-api/v1/` 或 `/k8s-proxy/` |
| 前端字段为空 | 后端响应字段、前端类型定义、页面取值路径是否一致 |
| K8s 权限错误 | `LOCAL_MOCK`、`KUBECONFIG`、ServiceAccount 和 RBAC 是否正确 |

## 提交前检查

| 检查项 | 命令或说明 |
|--------|------------|
| 文档格式 | `git diff --check` |
| 后端编译 | `cd $BASE_DIR/w7panel-server && go build` |
| 前端构建 | `cd $BASE_DIR/w7panel-ui && npm run build` |
| 接口文档 | 确认新增或变更 API 已更新到 `docs/development/api/` |
| 变更日志 | 确认 `docs/changelog/{version}.md` 已记录 |
