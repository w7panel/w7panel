# w7panel 开发指南

## 环境变量

```bash
export BASE_DIR=/home/wwwroot/w7panel-dev
```

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BASE_DIR` | . | 项目根目录 |
| `W7PANEL_HTTP_SERVER_PORT` | 8080 | HTTP端口 |
| `KO_DATA_PATH` | ./kodata | 静态资源路径（相对于可执行文件目录） |
| `LOCAL_MOCK` | 自动检测 | 开发模式：true；生产模式：false |
| `CAPTCHA_ENABLED` | false | 验证码开关，开发测试时设为 false 跳过滑块验证 |
| `KUBECONFIG` | - | **开发模式专用**：kubeconfig 文件路径；生产模式不需要 |

### 启动方式（推荐使用启动脚本）

**注意**：推荐使用 `./w7panel-ctl.sh` 启动脚本，自动检测并设置正确的环境变量。

```bash
# ========== 开发模式 (需要 kubeconfig.yaml) ==========
export KUBECONFIG=/path/to/kubeconfig.yaml
./w7panel-ctl.sh start

# ========== 生产模式 (使用 ServiceAccount) ==========
export LOCAL_MOCK=false
./w7panel-ctl.sh start
```

```bash
# 错误方式 - 环境变量不会传递给子进程
nohup ./w7panel server:start &  # ❌ 不推荐
```

---

## 环境初始化规范

### 目录结构

```
/home/                          # 持久化存储目录（重启不丢失）
├── env/                        # 运行时环境
│   ├── node/                   # Node.js
│   ├── go/                     # Go
│   └── ...
├── runtime/                    # 运行时数据
│   ├── logs/                   # 日志目录
│   └── ...
└── ...
```

### 环境安装（首次部署）

项目依赖的基础运行时环境需要安装到 `/home/env/` 目录下：

```bash
# 1. 创建目录
mkdir -p /home/env

# 2. 安装 Node.js（以 v20 为例）- 使用国内源
cd /tmp
wget https://npmmirror.com/mirrors/node/v20.x.x/node-v20.x.x-linux-x64.tar.xz
tar -xf node-v20.x.x-linux-x64.tar.xz
mv node-v20.x.x-linux-x64 /home/env/node

# 或使用阿里云镜像
wget https://npmmirror.com/mirrors/node/v20.20.0/node-v20.20.0-linux-x64.tar.xz

# 3. 安装 Go（以 1.24 为例）- 使用国内源
cd /tmp
# 阿里云镜像
wget https://npmmirror.com/mirrors/golang/go1.24.linux-amd64.tar.gz
# 谷歌中国镜像
wget https://golang.google.cn/dl/go1.24.linux-amd64.tar.gz
tar -C /home/env -xzf go1.24.linux-amd64.tar.gz

# 4. 配置环境变量
export PATH="/home/env/node/bin:/home/env/go/bin:$PATH"
export NODE_HOME=/home/env/node
export GOROOT=/home/env/go

# 5. 配置 Go 模块代理（国内源）
export GOPROXY=https://goproxy.cn,direct
```

### 环境变量配置

在 `~/.bashrc` 或 `/etc/profile.d/` 中添加：

```bash
# Node.js
export PATH="/home/env/node/bin:$PATH"
export NODE_HOME=/home/env/node
# npm 使用国内源
export NPM_CONFIG_REGISTRY=https://registry.npmmirror.com

# Go
export PATH="/home/env/go/bin:$PATH"
export GOROOT=/home/env/go
export GOPATH=/home/env/gopath
# Go 模块代理（国内源）
export GOPROXY=https://goproxy.cn,direct
```

### 验证环境

```bash
# 验证 Node.js
node --version
npm --version

# 验证 Go
go version
```

---

## 交流规则

- 始终使用中文回复
- **无特别要求时，部署测试均使用测试模式 (LOCAL_MOCK=true)**

### 开发规范

- **每次开发前先拉取最新代码**，防止代码冲突：
  ```bash
  # 后端
  cd $BASE_DIR/w7panel && git fetch origin dev-v1 && git pull origin dev-v1
  
  # 前端
  cd $BASE_DIR/w7panel-ui && git fetch origin dev-v1 && git pull origin dev-v1
  ```

### Git 配置

项目使用 `gitconfig.yaml` 管理 Git 凭证和代理配置。

```yaml
[user]
    username = <用户名>
    password = <GitHub Token>

[http]
    proxy = <代理服务器地址>

[https]
    proxy = <代理服务器地址>

[url "https://github.com/"]
    insteadOf = git@github.com:
```

**使用方法**:
```bash
# 创建软链接
ln -sf $BASE_DIR/gitconfig.yaml $BASE_DIR/w7panel/.gitconfig
ln -sf $BASE_DIR/gitconfig.yaml $BASE_DIR/w7panel-ui/.gitconfig
```

### 文档更新规则（重要！）

**项目文档分为两部分：**
1. **AGENTS.md** - AI助手开发指南（快速参考），包含UI设计规范
2. **/docs/** - 完整项目文档（详细说明）

```
docs/
├── README.md           # 项目概述
├── changelog/          # 更新日志（每版本独立文件）
│   ├── 1.0.0.md        # 当前版本（最新文件名即为当前版本号）
│   └── ...             # 历史版本
├── user-guide/         # 用户操作手册
│   ├── README.md       # 快速入门
│   ├── app-management.md
│   ├── file-management.md
│   ├── storage-management.md
│   ├── domain-management.md
│   └── faq.md
├── api/                # API接口文档
├── deployment/         # 部署文档
├── development/        # 开发指南
└── testing/            # 测试文档
```

**以下情况必须立即更新文档：**

| 触发条件 | 更新 AGENTS.md | 更新 /docs |
|---------|---------------|-----------|
| 新增/删除目录 | 目录结构、编译部署 | development/README.md |
| 修改构建命令 | 编译部署 | deployment/README.md |
| 新增/修改/删除 API | API 接口 | api/README.md |
| 修改环境变量 | 环境变量 | deployment/README.md |
| 新增/修改测试 | 测试流程 | testing/README.md |
| 新增功能模块 | - | user-guide/, development/ |
| 修改用户操作流程 | - | user-guide/ |
| 新增/修改UI组件 | UI设计规范（第10节） | - |
| 完成开发 | 版本管理规范 | changelog/{版本号}.md |
| 新增后端功能 | 后端 README | w7panel/README.md |
| 新增前端功能 | 前端 README | w7panel-ui/README.md |
| 新增编辑器功能 | 编辑器 README | codeblitz/README.md |

**更新检查清单：**
```
□ AGENTS.md 是否需要更新？（包括UI设计规范）
□ docs/changelog/{版本号}.md 是否需要更新？（版本日志）
□ docs/user-guide/ 是否需要更新？（用户操作）
□ docs/api/ 是否需要更新？
□ docs/deployment/ 是否需要更新？
□ docs/development/ 是否需要更新？
□ docs/testing/ 是否需要更新？
□ w7panel/README.md 是否需要更新？（后端）
□ w7panel-ui/README.md 是否需要更新？（前端）
□ codeblitz/README.md 是否需要更新？（编辑器）
□ tests/README.md 是否需要更新？（测试脚本）
```

**未遵守规则的后果：**
- 其他开发者无法正确部署
- 文档与实际代码不一致导致混乱
- AI助手无法提供准确帮助

---

## 开发流程规则

### 1. 前后端同步开发（项目特定）

**后端修改后，必须同步修改前端：**
- 修改了 API 返回字段 → 检查前端是否使用该字段
- 修改了 URL 路由格式 → 检查前端 API 调用路径
- 新增了接口 → **必须**在前端添加对应调用

**重要：每次修改后端代码后，必须执行以下检查：**
```bash
# 1. 检查是否需要修改前端
cd $BASE_DIR

# 新增路由或修改响应字段时，搜索相关关键字
grep -r "permission-agent\|permissionUrl\|compressUrl\|webdavUrl" w7panel-ui/src/ || true

# 2. 如果前端需要修改，同步修改
# 3. 前后端同时编译通过后才能提交
# 4. 验证功能是否正常工作
```

**涉及文件：**
- 后端：Controller、Service、路由
- 前端：API 接口、页面组件、类型定义
- 编辑器：`codeblitz/src/index.tsx`、`codeblitz/editor.html`
- 文档：更新 AGENTS.md

**未遵守规则的后果：**
- 功能不完整，前端无法使用新接口
- 必须返工，增加工作量
- 破坏用户体验

### 2. 项目特定规范

**w7panel 项目特定：**

#### 接口设计规范

**遵守协议标准：**
- **绝不破坏原有协议标准**（如 WebDAV 必须返回 XML）
- 在原接口基础上增强功能，而非新建冗余接口
- 保持接口统一性，避免同一功能多套实现

#### 性能优化规范

- 最大文件大小: 50MB
- 最大目录条目: 5000
- 请求超时: 10秒
- 特殊文件系统处理 (/proc, /sys, /dev)

### 项目特定检查

项目特定检查：
- [ ] 检查是否有 `offline`、`k8soffline` 等旧命名
- [ ] 检查 localStorage/sessionStorage 键名
- [ ] 检查 API 路径命名
- [ ] 检查环境变量命名

---

## 项目概述

基于 Kubernetes 的云原生应用管理平台。

| 项目 | 技术栈 | 目录 |
|------|--------|------|
| 后端 | Go 1.24 + Gin + w7-rangine-go | `$BASE_DIR/w7panel` |
| 前端 | Vue 3.5 + TypeScript + Arco Design | `$BASE_DIR/w7panel-ui` |
| Web IDE | React + TypeScript + Codeblitz | `$BASE_DIR/codeblitz` |
| 部署 | 编译输出 | `$BASE_DIR/dist` |
| Helm Charts | K8s 部署包 | `$BASE_DIR/w7panel/install/charts` |

---

## 目录结构

```
$BASE_DIR/
├── w7panel/                        # 后端源码
│   ├── app/application/http/controller/  # 控制器
│   ├── common/service/             # 业务服务
│   ├── common/middleware/          # 中间件
│   ├── install/                    # 安装相关
│   │   └── charts/                 # Helm Charts
│   │       └── w7panel/            # Chart 目录
│   ├── scripts/                    # 开发脚本
│   ├── kodata/                     # 静态资源
│   └── config.yaml
├── w7panel-ui/                     # 前端源码
│   ├── src/api/                    # API
│   ├── src/views/                  # 页面
│   ├── src/components/             # 组件
│   └── scripts/                    # 开发脚本
├── codeblitz/                      # Web IDE 源码 (基于 Codeblitz/OpenSumi)
│   ├── src/                        # 源码
│   ├── scripts/                    # 开发脚本
│   ├── package.json
│   ├── webpack.config.js           # Webpack 配置
│   └── node_modules/@codeblitzjs/ide-core/bundle/  # WASM 文件位置
├── kubeconfig.yaml               # K8S 集群配置
├── dist/                           # 编译输出目录
│   ├── w7panel                    # 可执行文件
│   ├── config.yaml                 # 配置文件
│   ├── kodata/                     # 前端+后端资源
│   ├── runtime/                    # 运行时目录
│   │   └── logs/                   # 日志目录
│   └── w7panel.db                 # SQLite 数据库
├── docs/                           # 项目文档
└── tests/                          # 测试脚本

/home/                              # 持久化存储目录（重启不丢失）
├── env/                            # 运行时环境
│   ├── node/                       # Node.js
│   └── go/                         # Go
└── runtime/                        # 运行时数据
    └── logs/                       # 日志目录
```

**注意**: 
- 编译输出目录 `$BASE_DIR/dist/` 位于持久存储分区，重启后不会丢失
- K8S 配置文件 `$BASE_DIR/kubeconfig.yaml` 用于内测和公测
- 基础运行时环境（Node.js、Go 等）存放在 `/home/env/` 目录下

---

## 编译部署

### 资源复制顺序规范（重要！）

```
复制资源时必须遵循以下顺序：
1. 先复制后端静态资源（logo.png, k3s-*.sh, ip2region 等）
2. 再复制前端资源（前端同名文件覆盖后端）

原因：前端构建产物可能与后端静态资源重名（如 index.html），
      需要后端资源优先，再让前端覆盖同名文件。
```

### 使用构建脚本（推荐）

```bash
# 完整构建（自动清理旧产物）
cd $BASE_DIR/w7panel/scripts
./build.sh
```

#### 生产环境配置

**K8s Deployment 配置示例**：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: w7panel
spec:
  template:
    spec:
      serviceAccountName: w7panel  # 使用 ServiceAccount（自动挂载 token）
      containers:
      - name: w7panel
        image: w7panel:xxx
        command: ["./w7panel-ctl.sh", "start"]
        env:
        - name: LOCAL_MOCK
          value: "false"           # 生产模式：使用 ServiceAccount
        - name: CAPTCHA_ENABLED
          value: "false"
        - name: KO_DATA_PATH
          value: "/home/wwwroot/w7panel-dev/dist/kodata"
        volumeMounts:
        - name: config
          mountPath: /home/wwwroot/w7panel-dev
      volumes:
      - name: config
        persistentVolumeClaim:
          claimName: w7panel-config
```

**注意**：生产环境使用 Pod 内置的 ServiceAccount（挂载在 `/var/run/secrets/kubernetes.io/serviceaccount/`），**不需要**设置 `KUBECONFIG` 环境变量。

**环境变量说明**：

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `LOCAL_MOCK` | ✅ 必填 | false | 生产模式设为 false，使用 ServiceAccount |
| `CAPTCHA_ENABLED` | 否 | false | 验证码开关 |
| `KO_DATA_PATH` | 否 | ./kodata | 数据目录 |
| `KUBECONFIG` | ❌ 不需要 | - | 开发模式专用，生产模式使用 ServiceAccount |

#### ⚠️ 重要：禁止直接使用 pkill/kill -9

**错误示例**（会产生僵尸进程）：
```bash
# ❌ 禁止直接使用 pkill -9
pkill -9 -f "w7panel"

# ❌ 禁止直接 kill -9
kill -9 $(pgrep w7panel)
```

**正确示例**（使用启动脚本）：
```bash
# ✅ 正确：使用启动脚本停止服务
./dist/w7panel-ctl.sh stop

# ✅ 正确：先 SIGTERM 再 SIGKILL（启动脚本已实现）
kill -TERM $PID
sleep 2
kill -9 $PID
```

**原因**：
- 直接 `kill -9` 会强制终止进程，不给子进程优雅退出的机会
- 子进程变成孤儿，被 PID 1 收养，但 PID 1 不回收 → 产生僵尸
- 启动脚本的 `stop` 命令会先发 SIGTERM，等待后再 SIGKILL，减少僵尸产生

### 僵尸进程处理

#### 问题根因

在容器环境中，僵尸进程问题的根本原因是：

```
容器 PID 1 (node opencode)
    └── 不设置 PR_SET_CHILD_SUBREAPER
    └── 不收割孤儿进程
    └── 当任何进程被 kill 后，如果它有子进程，这些子进程被 PID 1 收养
    └── PID 1 不收割，这些进程变成僵尸进程（Z 状态）
```

#### 修复方案（开发/测试场景）

**1. Go 代码实现 (main.go)**

```go
import (
    "log/slog"
    "os/signal"
    "syscall"
)

func init() {
    // 1. 设置子进程收割者 (PR_SET_CHILD_SUBREAPER)
    // 当前进程会成为其所有子进程的"收养者"，负责回收它们
    const PR_SET_CHILD_SUBREAPER = 36
    _, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0, 0)
    if errno != 0 {
        slog.Warn("Failed to set child subreaper", "error", errno)
    } else {
        slog.Info("set child subreaper successfully")
    }

    // 2. 忽略 SIGCHLD，让内核自动回收僵尸子进程
    // 设置后，子进程退出时不会变成僵尸，直接被内核回收
    signal.Ignore(syscall.SIGCHLD)
    slog.Info("SIGCHLD ignored for auto child process reaping")
}
```

**2. 停止脚本优化 (start.sh)**

改进孤儿进程停止逻辑，先 SIGTERM 再 SIGKILL：

```bash
# 清理可能的孤儿进程
for orphan in $orphans; do
    kill -TERM "$orphan" 2>/dev/null || true
done
sleep 1
kill -9 $orphans 2>/dev/null || true
```

#### 效果验证

启动服务后，检查日志中是否有以下输出：

```bash
tail -f /tmp/w7panel.log | grep -E "subreaper|SIGCHLD"
```

预期输出：
```
[INFO] set child subreaper successfully
[INFO] SIGCHLD ignored for auto child process reaping
```

#### 效果说明

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| w7panel 运行期间 | 可能产生僵尸 | ✅ 自动回收 |
| 服务重启后 | 僵尸继续累积 | ✅ 新服务正常回收 |
| 手动 kill -9 | 僵尸（不可避免） | 僵尸（不可避免） |
| 容器重启 | 清理所有僵尸 | 清理所有僵尸 |

#### 局限性

当进程被 `kill -9` 强制终止时：
- 进程没有机会执行清理代码
- 所有子进程被 PID 1 收养
- 如果 PID 1 不设置 subreaper，这些进程会变成僵尸

**解决方案**：重启容器或定期清理

#### 生产环境 (K8s 部署)

如果需要根本解决僵尸进程问题，需要在 K8s 部署时使用 tini：

```yaml
# deployment.yaml
containers:
- name: w7panel
  image: w7panel:xxx
  args: ["/usr/bin/tini", "--", "./w7panel", "server:start"]
```

---

## API 接口

### 文件管理

```
# WebDAV (通过代理访问容器内文件)
# 生产环境: /panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/webdav-agent/{pid}/agent/{path}
# 开发环境(LOCAL_MOCK=true): /panel-api/v1/files/webdav-agent/{pid}/agent/{path}

# WebDAV 标准方法
GET /panel-api/v1/files/webdav-agent/{pid}/agent/{path}           # 读取文件内容（自动处理符号链接、特殊文件）
PROPFIND /panel-api/v1/files/webdav-agent/{pid}/agent/{path}      # 列出目录，返回 XML (Depth: 1)
PUT /panel-api/v1/files/webdav-agent/{pid}/agent/{path}           # 写入文件
DELETE /panel-api/v1/files/webdav-agent/{pid}/agent/{path}        # 删除文件/目录
MKCOL /panel-api/v1/files/webdav-agent/{pid}/agent/{path}         # 创建目录
MOVE /panel-api/v1/files/webdav-agent/{pid}/agent/{path}          # 移动/重命名
COPY /panel-api/v1/files/webdav-agent/{pid}/agent/{path}          # 复制

# 压缩
POST /panel-api/v1/files/compress-agent/{pid}/compress
Body: {"sources": ["/path/file"], "output": "/path/out.tar.gz"}

# 解压
POST /panel-api/v1/files/compress-agent/{pid}/extract
Body: {"source": "/path/archive.zip", "target": "/path/extract"}

# 支持的压缩格式
压缩: zip, tar, tar.gz, tar.xz
解压: zip, tar, tar.gz, tar.bz2, tar.xz, 7z

# 权限修改 (通过代理访问)
# 生产环境: /panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/permission-agent/{pid}
# 开发环境(LOCAL_MOCK=true): /panel-api/v1/files/permission-agent/{pid}
POST /panel-api/v1/files/permission-agent/{pid}/chmod
Body: {"path": "/path/file", "mode": "755"}

POST /panel-api/v1/files/permission-agent/{pid}/chown
Body: {"path": "/path/file", "owner": "root"}

# 子进程权限修改
POST /panel-api/v1/files/permission-agent/{pid}/subagent/{subpid}/chmod
POST /panel-api/v1/files/permission-agent/{pid}/subagent/{subpid}/chown
```

### WebDAV 性能限制

| 限制项 | 值 | 说明 |
|--------|-----|------|
| 最大文件大小 | 50MB | 单次请求可读取的最大文件 |
| 最大目录条目 | 5000 | 单次请求返回的最大目录项数 |
| 特殊目录 | /proc, /sys, /dev, /run | PROPFIND 使用高效实现，返回标准 XML |
| 符号链接 | 自动解析 | `/etc/mtab` 等符号链接文件自动读取目标内容 |

### 开发模式 (LOCAL_MOCK)

当设置 `LOCAL_MOCK=true` 时，系统进入开发模式：

| 接口 | 生产环境 | 开发环境 |
|------|---------|---------|
| webdavUrl | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/webdav-agent/{pid}/agent` | `/panel-api/v1/files/webdav-agent/{pid}/agent` |
| compressUrl | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/compress-agent/{pid}` | `/panel-api/v1/files/compress-agent/{pid}` |
| permissionUrl | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/permission-agent/{pid}` | `/panel-api/v1/files/permission-agent/{pid}` |

### LOCAL_MOCK 架构设计

#### 两种模式的核心区别

| 模式 | Agent 位置 | 文件访问方式 |
|------|-----------|-------------|
| **LOCAL_MOCK** | 与面板同服务 | 直接读取本地文件系统 |
| **生产模式** | 独立 Agent Pod | 面板代理到 Agent，Agent 访问目标 Pod |

#### LOCAL_MOCK 模式原理

**核心思想**：Agent 和面板是同一个服务，不需要代理请求，直接通过 procpath 读取本地文件系统。

```
┌─────────────────────────────────────────────────────────────────┐
│                    LOCAL_MOCK 模式架构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   开发/测试 Pod                                                 │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  w7panel (面板 + Agent)                                  │   │
│   │                                                          │   │
│   │  请求处理流程:                                           │   │
│   │  1. 接收请求: GET /panel-api/v1/files/webdav-agent/1/agent/etc/    │   │
│   │  2. procpath.GetRootPath(1) → /host/proc/1/root    │   │
│   │  3. 读取本地文件系统: /host/proc/1/root/etc/         │   │
│   │  4. 返回文件内容                                        │   │
│   └─────────────────────────────────────────────────────────┘   │
│                           ↓                                     │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  挂载卷: /host/proc → 宿主机 /proc                   │   │
│   │                                                          │   │
│   │  宿主机 (Node)                                          │   │
│   │  └── /proc/{pid}/root/ - 各容器的根目录               │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**procpath 关键代码** (`procpath.go`):
```go
func GetBasePath() string {
    localMock := facade.Config.GetBool("app.local_mock")
    if localMock {
        return HostProcPath  // /host/proc
    }
    return ProcPath  // /proc
}

func GetRootPath(pid string) string {
    return filepath.Join(GetBasePath(), pid, "root")
}
```

#### 生产模式原理

**核心思想**：面板 API 代理到独立的 Agent Pod，由 Agent 实现对各个 Pod 的文件操作。

```
┌─────────────────────────────────────────────────────────────────┐
│                      生产模式架构                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   用户请求                                                       │
│   │                                                             │
│   ▼                                                             │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  面板服务 (w7panel)                                      │   │
│   │  1. 接收请求: GET /panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/webdav-agent/1/agent/etc/    │   │
│   │  2. 代理到: http://{podIp}:8000/panel-api/v1/files/...         │   │
│   └─────────────────────────────────────────────────────────┘   │
│                           ↓                                     │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  K8S API 代理                                          │   │
│   └─────────────────────────────────────────────────────────┘   │
│                           ↓                                     │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  Agent Pod (特权容器)                                  │   │
│   │  └── 访问目标 Pod 文件系统: /proc/{pid}/root/        │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 生产模式 vs LOCAL_MOCK 模式对比

| 特性 | 生产模式 | LOCAL_MOCK 模式 |
|------|---------|-----------------|
| Agent 位置 | 独立 Agent Pod | 与面板同服务 |
| 文件访问 | Agent 代理访问 | 直接读取 /host/proc |
| 部署要求 | 每节点部署 Agent Pod | 挂载宿主机 /proc |
| 网络 | 需要 K8S 网络 | 本地文件系统 |
| 性能 | 有网络延迟 | 无网络延迟 |
| **用户认证** | 必须验证 JWT token | **必须验证用户 token**（重要！） |

**重要安全原则**：`LOCAL_MOCK=true` 只改变 K8s API 调用方式（使用本地 kubeconfig），**不改变用户认证逻辑**。所有 API 请求都必须携带有效的用户 token。

#### 路由对比

| 接口 | 生产环境 | LOCAL_MOCK 环境 |
|------|---------|-----------------|
| WebDAV | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/webdav-agent/{pid}/agent/*` | `/panel-api/v1/files/webdav-agent/{pid}/agent/*` |
| Compress | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/compress-agent/{pid}/*` | `/panel-api/v1/files/compress-agent/{pid}/*` |
| Permission | `/panel-api/v1/{podIp}:8000/proxy/panel-api/v1/files/permission-agent/{pid}/*` | `/panel-api/v1/files/permission-agent/{pid}/*` |

#### 依赖条件

LOCAL_MOCK 模式正常工作需要以下条件：

| 条件 | 说明 | 必需 |
|------|------|------|
| 挂载 /proc | 将宿主机 /proc 挂载到容器的 /host/proc | ✅ 必须 |
| 节点访问权限 | 开发 Pod 和目标 Pod 在同一节点 | ✅ 必须 |
| Agent Pod | 不需要（面板和 Agent 是同一个服务） | ❌ 不需要 |

#### 当前问题

在**分离式开发环境**中存在以下问题：

| 问题 | 说明 |
|------|------|
| 无 /host/proc 挂载 | 开发环境是独立 Pod，无法挂载测试集群的宿主机 /proc |
| 跨集群访问 | 开发环境和测试集群不在同一内网，无法访问 |

**解决方案**: 使用公测模式，将服务部署到测试集群进行完整功能测试。

### 认证

```
POST /api/auth/login
POST /api/auth/refresh
GET  /api/auth/user
```

### TOKEN 获取方式

**测试时需要 TOKEN 进行认证，获取方式如下：**

| 方式 | 适用场景 | 命令/操作 |
|------|---------|----------|
| K8S ServiceAccount | 在 K8S 容器内运行 | `cat /var/run/secrets/kubernetes.io/serviceaccount/token` |
| 浏览器 localStorage | 前端登录后 | 打开浏览器控制台: `localStorage.getItem('webdavToken')` |
| /k8s/pid 接口 | API 测试 | 从接口返回的 `webdavToken` 字段获取 |
| Kubeconfig | 本地开发 | 从 kubeconfig.yaml 的 token 字段提取 |

**推荐方式（在 K8S 容器内）：**
```bash
# 方式1: 直接使用 K8S token（最常用）
export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# 方式2: 从 /k8s/pid 接口获取（包含 webdavToken）
RESPONSE=$(curl -s -G "http://localhost:8080/k8s/pid" \
  --data-urlencode "namespace=default" \
  --data-urlencode "HostIp=10.0.0.206" \
  --data-urlencode "containerName=w7-python" \
  --data-urlencode "podName=w7-python-xxx" \
  -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)")

TOKEN=$(echo $RESPONSE | python3 -c "import sys,json; print(json.load(sys.stdin).get('webdavToken',''))")
```

**在浏览器中获取（调试用）：**
```javascript
// 打开浏览器控制台 (F12)，执行：
localStorage.getItem('webdavToken')
// 或查看完整配置
console.log({
  token: localStorage.getItem('webdavToken'),
  apiUrl: localStorage.getItem('apiUrl'),
  wsBaseUrl: localStorage.getItem('wsBaseUrl')
})
```

---

## API 路由规范 (v5.3)

### 路由分层原则

```
/panel-api/v1/     面板业务 API (核心)
/k8s-proxy/        纯粹的 K8s API 代理
```

### 详细规范

| 类型 | 前缀 | 说明 |
|------|------|------|
| 面板业务 | `/panel-api/v1/` | Helm、配置、密钥、事件、代理等 |
| K8s 代理 | `/k8s-proxy/` | 仅转发 K8s API (api/v1/*, apis/*) |
| 未授权公开 | `/panel-api/v1/noauth/site/*` | 公开接口（必须只返回业务字段） |

### 未授权接口规范 (安全优先)

```
✅ 正确：只返回业务数据-api/v1/no
GET /panelauth/site/beian
Response: { "icpnumber": "xxx", "number": "xxx", "location": "xxx" }

❌ 错误：返回完整 K8s 资源
GET /panel-api/v1/noauth/namespaces/default/configmaps/beian
Response: { "kind": "ConfigMap", "metadata": {...}, "data": {...} }
```

### 禁止事项

- ❌ 禁止将面板业务 API 放到 `/k8s-proxy/` 下
- ❌ 禁止在 `/k8s-proxy/` 下添加非 K8s API 路由
- ❌ 禁止创建 `/panel-api/v1/v1/*` 这样的重复前缀
- ❌ 禁止未授权接口返回完整 K8s 资源对象

### 参考文档

- v5.4: `docs/refactoring/refactoring-plan-v5.4.md` (Hook 规范化、代码质量优化、深入分析)
- v5.3: `docs/refactoring/refactoring-plan-v5.3.md` (API 路由重构)

---

## 技术规范

### 后端规范

- **日志**: 使用 `log/slog` 键值对格式
```go
slog.Info("操作成功", "user", userID, "action", "create")
```

- **目录**: 控制器 `app/{module}/http/controller/`，服务 `common/service/`

### 前端规范

- **目录**: API `src/api/`，页面 `src/views/`，组件 `src/components/`，Hooks `src/hooks/`
- **Hooks 规范**: 见下方「前端性能规范」第5节 Hooks 使用规范

### 性能规范

**后端性能规范:**

1. **文件操作限制**
   - 最大文件大小: 50MB (`webdav.MaxFileSize`)
   - 最大目录条目: 5000 (`webdav.MaxDirEntries`)
   - 使用 `io.LimitedReader` 限制读取大小
   ```go
   limitedReader := &io.LimitedReader{R: file, N: MaxFileSize}
   ```

2. **内存管理**
   - 避免一次性加载大文件到内存
   - 使用流式处理 (stream) 而非缓冲处理
   - 大文件使用 `io.Copy` 直接传输
   - 及时关闭文件句柄和释放资源

3. **特殊目录处理**
   - `/proc`、`/sys`、`/dev`、`/run` 使用优化的 PROPFIND 实现
   - 虚拟文件（如 `/proc` 下的文件）不支持 seek，需要完整读取到 buffer

4. **符号链接处理**
   - 使用 `os.Lstat()` 获取链接信息
   - 使用 `os.Readlink()` 读取链接目标
   - GET 请求自动解析符号链接并返回目标内容

5. **认证优化 (已实现)**
   - JWT 解析结果缓存: 避免重复解析 (`token.go`)
   - Mock Token 缓存: 避免每次请求读文件 (`auth.go`)
   - Token 缓存: 避免重复调用 K8s TokenReview API

**前端性能规范:**

1. **大列表处理**
   - 使用虚拟滚动 (Virtual Scroll)
   - 分页加载，避免一次性请求大量数据
   - 设置请求超时（建议 10 秒）

2. **大文件处理**
   - 文件内容超过 10MB 时提示用户
   - 使用流式下载，避免完整加载到内存
   - 显示加载进度

3. **内存管理**
   - 组件销毁时清理定时器和事件监听
   - 避免在循环中创建大量对象
   - 使用 `v-if` 替代 `v-show` 减少不必要的渲染

4. **API 请求优化**
   - 避免串行请求，使用 `Promise.all` 批量请求
   - 避免重复请求，添加请求缓存
   - useRequest Hook 添加取消机制

5. **Hooks 使用规范**

**useRequest Hook** (`src/hooks/request.ts`):
```typescript
import useRequest from '@/hooks/request';

// 基础用法
const { loading, response, run, cancel, refresh } = useRequest(api);

// 完整配置
const { loading, response, run, cancel, refresh, clearCache } = useRequest(api, {
    defaultValue: [],           // 默认值
    isLoading: true,            // 初始加载状态
    cache: true,               // 启用缓存
    cacheTime: 5 * 60 * 1000,  // 缓存时间 (5分钟)
    retry: 3,                   // 重试次数
    retryDelay: 1000,          // 重试延迟 (1秒)
    timeout: 30000,            // 请求超时 (30秒)
    onSuccess: (data) => {},   // 成功回调
    onError: (err) => {},      // 错误回调
});

// 手动运行请求
run();

// 取消请求
cancel();

// 刷新 (清除缓存后重新请求)
refresh();
```

**定时器管理** (`src/hooks/timer.ts`):
```typescript
import { useTimer, usePolling } from '@/hooks/timer';

// 定时器管理
const { setTimeout, setInterval, clearTimer, clearAllTimers } = useTimer();

// 设置定时器
const timerId = setTimeout('my-timer', () => {
    console.log('执行一次');
}, 3000);

const intervalId = setInterval('my-interval', () => {
    console.log('每3秒执行');
}, 3000);

// 清理定时器
clearTimer('my-timer');

// 轮询 (自动处理激活/停用状态)
const { startPolling, stopPolling, restartPolling } = usePolling(async () => {
    await fetchData();
}, 5000);
```

**API 请求示例**:
```typescript
// src/api/cluster.ts
// compressUrl 从 /k8s/pid 接口返回，格式: /panel-api/v1/files/compress-agent/{pid}
export function compressFiles(compressUrl: string, sources: string[], output: string) {
    return axios.post(`${compressUrl}/compress`, { sources, output });
}
```

---

## 常见问题

| 问题 | 解决方案 |
|------|----------|
| slog 格式错误 | 必须使用键值对: `slog.Info("msg", "key", value)` |
| kodata 丢失 | 运行时依赖 `kodata/` 目录，需正确复制 |
| K8S 连接失败 | 检查 kubeconfig.yaml 路径和内容 |
| 编辑器 WASM 401 错误 | 复制 WASM 文件到 `kodata/plugin/codeblitz/` |
| 编辑器 PROPFIND 400 错误 | Content-Type 必须是 `text/xml; charset=utf-8` |
| 编辑器资源管理器空白 | 检查 WebDAV API 是否正常，查看浏览器控制台日志 |
| editor.html JS 加载失败 | 检查 main.*.js 文件名是否与 editor.html 中引用一致 |
| 编辑器写入失败 ENOTSUP | 使用 OverlayFS + 事件回调同步到 WebDAV |
| WebDAV 401 认证失败 | 需要有效的 K8S Token，从 kubeconfig 或 ServiceAccount 获取 |
| agent-browser 找不到元素 | 使用 `snapshot -i` 查看交互元素，用 ref (@e1) 定位 |

---

## Web IDE 编辑器

### 架构说明

编辑器基于 Codeblitz (OpenSumi) 构建，使用以下文件系统方案：

| 组件 | 说明 |
|------|------|
| OverlayFS | 叠加文件系统，提供写入能力 |
| InMemory | 可写层（内存），临时存储修改 |
| DynamicRequest | 只读层，通过 WebDAV 读取远程文件 |

### 写入同步机制

编辑器通过事件回调将本地修改同步到 WebDAV：

| 事件 | WebDAV 操作 | 说明 |
|------|-------------|------|
| `onDidSaveTextDocument` | PUT | 保存文件时同步到服务器 |
| `onDidCreateFiles` | MKCOL | 创建文件/目录时同步 |
| `onDidDeleteFiles` | DELETE | 删除文件时同步 |

### LOCAL_MOCK 模式

在 `LOCAL_MOCK=true` 环境下：
- WebDAV 直接访问本地文件系统（通过 `/host/proc` 映射）
- 测试时使用 `pid=1` 模拟特权 pod
- **注意**: 认证仍需要有效的 K8S Token

---

## 默认账号

- 用户名: `admin`
- 密码: `123456`

---

## 测试模式

### 内测模式

本地运行后端服务，远程连接 K8s 集群进行测试。

```bash
# 启动内测模式（开发模式，需要 kubeconfig.yaml）
cd $BASE_DIR/dist
CAPTCHA_ENABLED=false LOCAL_MOCK=true KO_DATA_PATH=$BASE_DIR/dist/kodata KUBECONFIG=$BASE_DIR/kubeconfig.yaml ./w7panel server:start
```

**限制**：
- 无法访问测试集群的宿主机文件系统（无 /host/proc 挂载）
- 文件管理功能需要通过 Agent Pod 代理访问
- 适用于：API 测试、前端 UI 测试（非文件管理）

### 公测模式

正式部署测试，构建镜像并部署到 K8s 集群。

**流程**：
1. 构建镜像
2. 推送到镜像仓库
3. 更新 Helm Charts 镜像地址
4. 部署测试

**Helm Charts 项目**: `$BASE_DIR/w7panel/install/charts/`

---

## Helm Charts 维护

### 项目结构

```
$BASE_DIR/w7panel/install/charts/w7panel/
├── Chart.yaml          # Chart 元数据
├── values.yaml         # 默认配置值
└── templates/          # K8s 资源模板
    ├── deployment.yaml
    ├── daemonset.yaml
    ├── service.yaml
    └── ...
```

### 镜像配置

```yaml
# values.yaml
image:
  repository: ccr.ccs.tencentyun.com/afan/w7panel
  pullPolicy: IfNotPresent
  tag: "1.0.19"
```

### 公测部署命令

```bash
# 1. 构建镜像
docker build -t ccr.ccs.tencentyun.com/afan/w7panel:1.0.20 .

# 2. 推送镜像
docker push ccr.ccs.tencentyun.com/afan/w7panel:1.0.20

# 3. 部署（通过 helm --set 指定镜像版本）
helm upgrade --install w7panel ./w7panel/install/charts/w7panel -n default \
  --set image.tag=1.0.20
```

---

## 故障排查

---
