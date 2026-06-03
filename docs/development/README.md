# 开发指南

本文档是 W7Panel 开发资料入口，面向后端、前端、Web IDE 静态资源、应用示例、联调测试和文档维护。更完整的用户、部署、测试和版本资料见 [docs/README.md](../README.md)。

## 快速入口

### 后端 API

`docs/development/api/` 以业务专题为主、模块文档为辅组织接口说明。新增或修改接口时，优先更新对应业务专题；无法归入专题或需要从 Controller 反查实现时，再维护模块文档。

业务专题：

| 专题 | 文档 | 说明 |
|------|------|------|
| 调用凭据 | [api/credentials.md](./api/credentials.md) | 用户 token、webdavToken、微应用 props、OIDC token、LOCAL_MOCK |
| 集群运维 | [api/cluster-ops.md](./api/cluster-ops.md) | Helm、YAML、终端、代理、DNS、GPU、指标 |
| 指标 | [api/metrics.md](./api/metrics.md) | CPU、内存、磁盘使用量和 metrics 组件安装状态 |
| Longhorn | [api/longhorn.md](./api/longhorn.md) | Longhorn 安装、卷状态、副本筛选、attach/detach、快照和文件系统操作 |
| 容器文件管理 | [api/container-files.md](./api/container-files.md) | WebDAV、压缩解压、权限修改、分片上传和文件编辑器 |
| 容器镜像管理 | [api/container-images.md](./api/container-images.md) | Registry、containerd 镜像操作、容器 commit、ZPK 构建镜像 |
| 应用管理 | [api/zpk.md](./api/zpk.md) | ZPK 配置、列表、安装、升级、构建和应用管理 |
| 微应用静态资源 | [api/microapp-static.md](./api/microapp-static.md) | 微应用信息、后端代理、静态资源状态、下载和回源 |
| OAuth/OIDC | [api/oauth-oidc.md](./api/oauth-oidc.md) | OIDC Provider、Client 管理、授权码、Console OAuth |
| 云主机 | [api/k3k.md](./api/k3k.md) | 云主机用户、CKM/CVM、同步、菜单、套餐和订单入口 |

模块文档：

| 模块 | 文档 | 说明 |
|------|------|------|
| 通用约定 | [api/conventions.md](./api/conventions.md) | API 路由分层、鉴权、响应格式、安全、前后端同步和测试规范 |
| application | [api/application.md](./api/application.md) | 面板核心业务接口 |
| audit | [api/audit.md](./api/audit.md) | 登录日志、操作日志和审计状态 |

### 前端

| 文档 | 说明 |
|------|------|
| [frontend/README.md](./frontend/README.md) | `w7panel-ui` 前端开发文档入口 |
| [frontend/conventions.md](./frontend/conventions.md) | 前端目录、API、页面、UI、状态、性能和提交流程规范 |
| [frontend/auth-state.md](./frontend/auth-state.md) | 前端 token 注入、刷新、本地缓存和权限状态 |
| [frontend/microapps.md](./frontend/microapps.md) | 微应用容器、Wujie props、token 边界、后端代理和接入流程 |
| [frontend/components.md](./frontend/components.md) | 前端组件文档入口，组件明细按类型拆分 |
| [frontend/wujie-events.md](./frontend/wujie-events.md) | Wujie 微前端事件、参数、回调响应和调用示例 |

### 示例和专题

| 文档 | 说明 |
|------|------|
| [examples/README.md](./examples/README.md) | 面板内应用开发示例，覆盖独立应用仓库、后端服务、前端微应用、镜像发布和可选 Helm Chart |
| [CHANNELLOCAL_PERFORMANCE_ANALYSIS.md](./CHANNELLOCAL_PERFORMANCE_ANALYSIS.md) | ChannelLocal 性能分析专题 |

## 仓库结构

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
│   ├── kodata/                      # 后端静态资源、CRD、内置 Chart、Web IDE 插件包
│   └── install/                     # 安装相关资源
├── w7panel-ui/                      # 前端源码，Vue 3 + TypeScript + Arco Design
│   ├── config/                      # Vite 和构建配置
│   ├── scripts/                     # 前端构建辅助脚本
│   ├── src/api/                     # API 调用封装
│   ├── src/components/              # 公共组件
│   ├── src/hooks/                   # 复用逻辑
│   ├── src/router/                  # 路由配置
│   └── src/views/                   # 页面模块
├── codeblitz/                       # Web IDE 源码占位目录；当前静态包在 w7panel-server/kodata/plugin/codeblitz.zip
├── charts/                          # w7panel Helm Chart
├── installer/                       # 安装脚本、离线清单和系统配置
├── docs/                            # 项目文档
└── tests/                           # 测试脚本和测试资料
```

说明：外层目录名通常是 `w7panel`，后端源码目录是 `w7panel-server/`。历史文档或部署路径里出现的 `w7panel/` 可能指仓库根目录，需要结合上下文判断。

## 开发前准备

### 同步代码

每次开发前先同步 `dev-v1`。本文中的 `$BASE_DIR` 指仓库根目录：

```bash
cd $BASE_DIR
git fetch origin dev-v1
git pull origin dev-v1
```

如果本地分支和远端分支已分叉，先确认使用 merge、rebase 还是 fast-forward，再执行 `git pull`。

### 默认测试模式

无特别要求时，部署测试使用测试模式：

```bash
export LOCAL_MOCK=true
export CAPTCHA_ENABLED=false
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
```

当前实现中，`w7panel-server/common/middleware/auth.go` 在 `LOCAL_MOCK=true` 时会从 `KUBECONFIG` 或默认路径读取 token 并填充 `k8s_token`，用于本地开发测试。新增接口不要依赖该行为作为生产鉴权逻辑；生产环境必须使用真实用户 token 或 ServiceAccount。

## 开发规范

### 后端

| 类型 | 位置 |
|------|------|
| Controller | `w7panel-server/app/{module}/http/` |
| 路由注册 | `w7panel-server/app/{module}/provider.go` |
| 公共服务 | `w7panel-server/common/service/` |
| 中间件 | `w7panel-server/common/middleware/` |
| 静态资源 | `w7panel-server/kodata/` |
| 开发脚本 | `w7panel-server/dev-tools/scripts/` |

约定：

- 面板业务 API 使用 `/panel-api/v1/`。
- K8s 原生代理使用 `/k8s-proxy/`，不要放入面板业务接口。
- 公开接口使用 `/panel-api/v1/noauth/*`，只返回允许公开的业务字段。
- WebDAV 等标准协议接口必须保持协议响应格式，例如 `PROPFIND` 返回 XML。
- 日志使用 `log/slog` 键值对格式，不输出完整 token、密码、密钥和 OIDC code。

```go
slog.Info("create app success", "namespace", namespace, "name", name)
```

### 前端

| 类型 | 位置 |
|------|------|
| API 封装 | `w7panel-ui/src/api/` |
| 页面 | `w7panel-ui/src/views/` |
| 公共组件 | `w7panel-ui/src/components/` |
| 复用逻辑 | `w7panel-ui/src/hooks/` |
| 鉴权和本地状态 | `w7panel-ui/src/utils/auth.ts` |

约定：

- 页面优先使用已有 API 封装，不在页面内散落复杂路径。
- URL 参数使用 axios `params` 或 `encodeURIComponent`。
- localStorage/sessionStorage key 使用 `w7panel-*` 命名，不新增 `offline-*`、`k8soffline-*` 等旧命名。
- 新增公共组件、Wujie 事件或可复用前端能力时，同步更新 [frontend/README.md](./frontend/README.md) 及相关子文档。

## 前后端同步

后端改动后必须检查前端影响：

| 后端变化 | 前端检查 |
|----------|----------|
| 新增接口 | 是否需要新增 `src/api/` 方法和页面调用 |
| 修改路由 | 搜索旧路径并同步替换 |
| 修改响应字段 | 检查类型定义、页面取值和展示逻辑 |
| 修改字段含义 | 检查按钮状态、权限判断、读写逻辑是否一致 |
| 修改鉴权 | 检查 token 获取、刷新、拦截器和错误处理 |
| 修改 WebDAV、压缩、权限接口 | 检查 `webdavUrl`、`compressUrl`、`permissionUrl`、`webdavToken` 使用处 |

常用搜索：

```bash
rg "接口路径或字段名" w7panel-server w7panel-ui/src
rg "permission-agent|permissionUrl|compressUrl|webdavUrl" w7panel-ui/src
rg "offline|k8soffline" .
rg "localStorage\\.|sessionStorage\\." w7panel-ui/src
```

## 文档维护

开发过程中按变更类型同步文档：

| 变更类型 | 需要更新 |
|----------|----------|
| 新增、修改、删除 API | `docs/development/api/{module}.md`，必要时同步 `docs/api/README.md` |
| 新增前端公共组件 | `docs/development/frontend/components.md` |
| 新增 Wujie 事件 | `docs/development/frontend/wujie-events.md` |
| 修改构建、部署、环境变量 | `docs/deployment/README.md` |
| 新增或修改测试 | `docs/testing/README.md` 或 `tests/README.md` |
| 新增功能模块或用户流程变化 | `docs/user-guide/` |
| 完成开发 | `docs/changelog/{version}.md` |

新增接口文档至少包含：

- 接口功能
- 请求方法和路径
- 请求参数说明
- 响应参数说明
- 鉴权要求
- 必要的响应示例、边界限制或特殊逻辑说明

## 常用命令

### 后端

```bash
cd $BASE_DIR/w7panel-server
mkdir -p $BASE_DIR/dist
go build -o $BASE_DIR/dist/w7panel .
go test ./...
```

### 前端

```bash
cd $BASE_DIR/w7panel-ui
npm run dev
npm run build
```

### 启动服务

推荐通过启动脚本启动：

```bash
cd $BASE_DIR/dist
export LOCAL_MOCK=true
export CAPTCHA_ENABLED=false
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
./w7panel-ctl.sh start
```

如果当前 `dist/` 只有手动编译出的二进制、还没有复制启动脚本，可以直接启动：

```bash
cd $BASE_DIR/dist
export LOCAL_MOCK=true
export CAPTCHA_ENABLED=false
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
export KO_DATA_PATH=$BASE_DIR/w7panel-server/kodata
./w7panel server:start
```

停止服务优先使用：

```bash
./w7panel-ctl.sh stop
```

不要直接使用 `pkill -9` 或 `kill -9` 停止服务，避免子进程无法优雅退出。

### 测试

```bash
cd $BASE_DIR/tests
bash run.sh
```

## 调试排查

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

### 常见问题

| 问题 | 检查方向 |
|------|----------|
| 接口 401 | token 是否存在、是否走了公开接口、`LOCAL_MOCK` 下是否正确读取到 kubeconfig token |
| 接口 404 | 路由是否注册，前缀是否为 `/panel-api/v1/` 或 `/k8s-proxy/` |
| 前端字段为空 | 后端响应字段、前端类型定义、页面取值路径是否一致 |
| K8s 权限错误 | `LOCAL_MOCK`、`KUBECONFIG`、ServiceAccount、RBAC 是否正确 |
| WebDAV 行为异常 | 方法语义、XML 响应、token、agent URL、特殊文件处理是否正确 |

## 提交前检查

| 检查项 | 命令或说明 |
|--------|------------|
| 文档格式 | `git diff --check` |
| 后端编译和测试 | `cd $BASE_DIR/w7panel-server && go test ./...` |
| 前端构建 | `cd $BASE_DIR/w7panel-ui && npm run build` |
| 旧命名检查 | `rg "offline|k8soffline" .` |
| API 路径检查 | 确认面板业务 API 使用 `/panel-api/v1/`，K8s 代理使用 `/k8s-proxy/` |
| 接口文档 | 确认新增或变更 API 已更新到 `docs/development/api/` |
| 版本日志 | 确认 `docs/changelog/{version}.md` 已记录 |
