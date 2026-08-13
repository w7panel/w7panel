# w7panel 开发指南

#====================================================
# 全局规则 (必须遵守)
#====================================================

## 全局规则

### 一、交流规则

- 始终使用中文回复
- **无特别要求时，部署测试均使用测试模式 (LOCAL_MOCK=true)**

### 二、开发规范

- **每次开发前先拉取最新代码**，防止代码冲突：
  ```bash
  # 后端
  cd $BASE_DIR/w7panel-server && git fetch origin dev-v1 && git pull origin dev-v1
  
  # 前端
  cd $BASE_DIR/w7panel-ui && git fetch origin dev-v1 && git pull origin dev-v1
  ```

### 三、Git 配置

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
ln -sf $BASE_DIR/gitconfig.yaml $BASE_DIR/w7panel-server/.gitconfig
ln -sf $BASE_DIR/gitconfig.yaml $BASE_DIR/w7panel-ui/.gitconfig
```

### 四、文档更新规则

**项目文档分为两部分：**
1. **AGENTS.md** - AI助手开发指南（快速参考），包含UI设计规范
2. **/docs/** - 完整项目文档（详细说明）

```
docs/
├── README.md           # 项目概述
├── src/
│   ├── user-guide/     # 用户操作手册、部署运维
│   │   └── overview/
│   │       └── changelog/ # 更新日志（每版本独立文件）
│   └── development/    # 开发指南、API、前端规范
├── .vitepress/         # VitePress 配置和主题
├── package.json
└── pnpm-lock.yaml
```

**以下情况必须立即更新文档：**

| 触发条件 | 更新 AGENTS.md | 更新 /docs |
|---------|---------------|-----------|
| 新增/删除目录 | 目录结构、编译部署 | docs/src/development/index.md |
| 修改构建命令 | 编译部署 | docs/src/user-guide/quick-start.md 或新增部署专题文档 |
| 新增/修改/删除 API | API 接口 | docs/src/development/api/index.md 及对应专题 |
| 修改环境变量 | 环境变量 | docs/src/user-guide/quick-start.md 或新增部署专题文档 |
| 新增/修改测试 | 测试流程 | tests/README.md |
| 新增功能模块 | - | docs/src/user-guide/, docs/src/development/ |
| 修改用户操作流程 | - | docs/src/user-guide/ |
| 新增/修改UI组件 | UI设计规范（第10节） | - |
| 完成开发 | 版本管理规范 | docs/src/user-guide/overview/changelog/{版本号}.md |
| 新增后端功能 | 后端 README | w7panel-server/README.md |
| 新增前端功能 | 前端 README | w7panel-ui/README.md |

**更新检查清单：**
```
□ AGENTS.md 是否需要更新？（包括UI设计规范）
□ docs/src/user-guide/overview/changelog/{版本号}.md 是否需要更新？（版本日志）
□ docs/src/user-guide/ 是否需要更新？（用户操作）
□ docs/src/development/api/ 是否需要更新？
□ docs/src/user-guide/quick-start.md 或部署专题文档是否需要更新？
□ docs/src/development/ 是否需要更新？
□ w7panel-server/README.md 是否需要更新？（后端）
□ w7panel-ui/README.md 是否需要更新？（前端）
□ tests/README.md 是否需要更新？（测试脚本）
```

**未遵守规则的后果：**
- 其他开发者无法正确部署
- 文档与实际代码不一致导致混乱
- AI助手无法提供准确帮助

### 五、开发流程规则

#### 1. 前后端同步开发

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
- 文档：更新 AGENTS.md

**未遵守规则的后果：**
- 功能不完整，前端无法使用新接口
- 必须返工，增加工作量
- 破坏用户体验

#### 2. 接口设计规范

**遵守协议标准：**
- **绝不破坏原有协议标准**（如 WebDAV 必须返回 XML）
- 在原接口基础上增强功能，而非新建冗余接口
- 保持接口统一性，避免同一功能多套实现

#### 3. 性能优化规范

- 最大文件大小: 50MB
- 最大目录条目: 5000
- 请求超时: 10秒
- 特殊文件系统处理 (/proc, /sys, /dev)

### 六、项目特定检查

项目特定检查：
- [ ] 检查是否有 `offline`、`k8soffline` 等旧命名
- [ ] 检查 localStorage/sessionStorage 键名
- [ ] 检查 API 路径命名
- [ ] 检查环境变量命名

### 七、代码审查规范

#### 审查原则

1. **代码存在 ≠ 代码被使用**
   - 搜索到方法存在，还要确认是否被调用
   - 检查外部库如何关联（接口、实现、调用链）
   - 区分：类型定义 vs 类型使用

2. **关联分析**
   - 对于接口：确认谁实现了它
   - 对于库方法：确认调用路径
   - 对于自定义方法：确认调用方

3. **逻辑一致性审查**（重要！）
   - 修改了某处字段含义，其他相关位置也需要同步修改
   - 例如：后端修改了 editable 含义，前端按钮控制、Read/Seek 等都要同步
   - 审查时问自己：这段代码和前面讨论的逻辑一致吗？

4. **深度审查清单**
   ```
   □ 新增的方法是否被调用？（搜索调用处）
   □ 新增的接口是否被实现？（搜索 implements/implementation）
   □ 自定义属性是否被前端解析？（搜索解析代码）
   □ 删除的属性前端是否还在使用？（搜索使用处）
   □ 安全逻辑是否完整？（边界条件、日志记录）
   □ 字段含义变更是否同步到所有相关位置？（逻辑一致性）
   ```

#### 常见遗漏点

| 遗漏类型 | 说明 | 审查方法 |
|---------|------|---------|
| 未使用的方法 | 功能预留但未启用 | 搜索调用处 |
| 未实现的接口 | 定义但未实现 | 搜索 implements |
| 未解析的属性 | 后端返回但前端未解析 | 搜索解析代码 |
| 残留代码 | 旧功能删除但代码遗留 | 对比前后变更 |
| 导入未使用 | import 后未引用 | 编译检查 |

#### WebDAV 审查示例

```
后端新增属性 → 检查前端是否解析 → 检查 UI 是否使用
后端删除属性 → 检查前端是否还在解析 → 检查 UI 是否还在使用
新增安全逻辑 → 检查边界条件 → 检查日志记录
```

### 八、API 路由规范

#### 路由分层原则

```
/panel-api/v1/     面板业务 API (核心)
/k8s-proxy/        纯粹的 K8s API 代理
```

#### 详细规范

| 类型 | 前缀 | 说明 |
|------|------|------|
| 面板业务 | `/panel-api/v1/` | Helm、配置、密钥、事件、代理等 |
| K8s 代理 | `/k8s-proxy/` | 仅转发 K8s API (api/v1/*, apis/*) |
| 未授权公开 | `/panel-api/v1/noauth/site/*` | 公开接口（必须只返回业务字段） |

#### 未授权接口规范 (安全优先)

```
✅ 正确：只返回业务数据
GET /panelauth/site/beian
Response: { "icpnumber": "xxx", "number": "xxx", "location": "xxx" }

❌ 错误：返回完整 K8s 资源
GET /panel-api/v1/noauth/namespaces/default/configmaps/beian
Response: { "kind": "ConfigMap", "metadata": {...}, "data": {...} }
```

#### 禁止事项

- ❌ 禁止将面板业务 API 放到 `/k8s-proxy/` 下
- ❌ 禁止在 `/k8s-proxy/` 下添加非 K8s API 路由
- ❌ 禁止创建 `/panel-api/v1/v1/*` 这样的重复前缀
- ❌ 禁止未授权接口返回完整 K8s 资源对象

#====================================================
# 环境配置
#====================================================
## 环境变量

```bash
export BASE_DIR=/home/wwwroot/w7panel-dev
```

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BASE_DIR` | . | 项目根目录 |
| `W7PANEL_HTTP_SERVER_PORT` | 8000 | HTTP端口 |
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

# 3. 安装 Go（以 1.26 为例）- 使用国内源
cd /tmp
# 阿里云镜像
wget https://npmmirror.com/mirrors/golang/go1.26.linux-amd64.tar.gz
# 谷歌中国镜像
wget https://golang.google.cn/dl/go1.26.linux-amd64.tar.gz
tar -C /home/env -xzf go1.26.linux-amd64.tar.gz

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

## 项目概述

基于 Kubernetes 的云原生应用管理平台。

| 项目 | 技术栈 | 目录 |
|------|--------|------|
| 后端 | Go 1.26 + Gin + w7-rangine-go | `$BASE_DIR/w7panel-server` |
| 前端 | Vue 3 + TypeScript + Arco Design | `$BASE_DIR/w7panel-ui` |
| 部署 | 编译输出 | `$BASE_DIR/dist` |
| Helm Charts | K8s 部署包 | `$BASE_DIR/charts/w7panel` |

---

## 目录结构

```
$BASE_DIR/
├── w7panel-server/                 # 后端源码
│   ├── app/                        # 后端应用模块
│   │   ├── application/            # 面板核心业务 API
│   │   ├── auth/                   # 认证 API
│   │   ├── k3k/                    # K3k API
│   │   ├── k3s-registry/           # K3s 镜像仓库 API
│   │   ├── metrics/                # 指标 API
│   │   └── zpk/                    # ZPK 应用 API
│   ├── common/service/             # 业务服务
│   │   └── k8s/coordination/       # 通用 Kubernetes Lease 与分布式并发协调
│   ├── common/middleware/          # 中间件
│   ├── k8s/pkg/apis/bootstrapinstallation/ # BootstrapInstallation API（w7panel.w7.com）
│   ├── dev-tools/scripts/          # 开发脚本
│   ├── kodata/                     # 静态资源
│   └── config.yaml
├── w7panel-ui/                     # 前端源码
│   ├── src/api/                    # API
│   ├── src/views/                  # 页面
│   ├── src/components/             # 组件
│   └── scripts/                    # 开发脚本
├── charts/                         # Helm Charts
│   └── w7panel/                    # 当前维护的面板 Chart
├── kubeconfig.yaml               # K8S 集群配置
├── dist/                           # 编译输出目录
│   ├── w7panel                    # 可执行文件
│   ├── config.yaml                 # 配置文件
│   ├── kodata/                     # 前端+后端资源
│   ├── runtime/                    # 运行时目录
│   │   └── logs/                   # 日志目录
│   └── w7panel.db                 # SQLite 数据库
├── docs/                           # 项目文档
│   └── src/development/examples/    # 应用开发示例
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
cd $BASE_DIR/w7panel-server/dev-tools/scripts
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
```

### WebDAV 性能限制

| 限制项 | 值 | 说明 |
|--------|-----|------|
| 最大目录条目 | 5000 | 单次请求返回的最大目录项数 |
| 不可读文件 | device/fifo/socket | 返回空内容 |

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
POST /panel-api/v1/login
POST /panel-api/v1/auth/refresh-token2
GET  /panel-api/v1/auth/userinfo
```

### TOKEN 获取方式

**测试时需要 TOKEN 进行认证，获取方式如下：**

| 方式 | 适用场景 | 命令/操作 |
|------|---------|----------|
| K8S ServiceAccount | 在 K8S 容器内运行 | `cat /var/run/secrets/kubernetes.io/serviceaccount/token` |
| 浏览器 localStorage | 前端登录后 | 打开浏览器控制台: `localStorage.getItem('webdavToken')` |
| /panel-api/v1/pid 接口 | API 测试 | 从接口返回的 `webdavToken` 字段获取 |
| Kubeconfig | 本地开发 | 从 kubeconfig.yaml 的 token 字段提取 |

**推荐方式（在 K8S 容器内）：**
```bash
# 方式1: 直接使用 K8S token（最常用）
export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# 方式2: 从 /panel-api/v1/pid 接口获取（包含 webdavToken）
RESPONSE=$(curl -s -G "http://localhost:8000/panel-api/v1/pid" \
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

## 技术规范

### 后端规范

- **日志**: 使用 `log/slog` 键值对格式
```go
slog.Info("操作成功", "user", userID, "action", "create")
```

- **目录**: 控制器 `w7panel-server/app/{module}/http/controller/`，服务 `w7panel-server/common/service/`

#### ZPK 更新规范

- 跨应用更新允许新制品 `identifie` 与已有 AppGroup 的 `spec.identifie` 不同；`/panel-api/v1/zpk/config` 返回的根应用 `releaseName`、`deployName` 必须沿用已有 AppGroup 名称，以便安装界面读取原实例参数。安装接口的资源命名逻辑独立维护，不能仅因配置回填调整而改变。
- 内置预装应用直接声明在 `w7panel-server/kodata/yaml/bootstrap-installations.yaml`；不再使用 BootstrapProfile，也不维护 revision。内置资源必须保留 `w7.cc/bootstrap-builtin=true` 标签；从清单移除条目后，升级脚本会在该标签和 BootstrapInstallation 类型范围内安全清理，并由 finalizer 自动卸载其拥有的 AppGroup。

### 前端规范

- **目录**: API `w7panel-ui/src/api/`，页面 `w7panel-ui/src/views/`，组件 `w7panel-ui/src/components/`，Hooks `w7panel-ui/src/hooks/`
- **Hooks 规范**: 见下方「前端性能规范」第5节 Hooks 使用规范

#### 网关插件 UI 规范

- AppGroup、MicroApp 的通用 API 地址以及 `w7.cc/group-name`、`w7.cc/official-app`、`w7.cc/deny-delete` 等跨模块资源元数据统一维护在 `src/utils/w7panel-resource.ts`，不得放入网关插件专用工具。已知名称或标识的资源必须使用单资源 GET 或 Kubernetes `labelSelector` 定向查询，禁止先全量读取 AppGroup/MicroApp 再在前端筛选。
- 网关插件列表以 `POST https://zm.w7.com/zpk-market/formula/list` 为展示主数据源，请求固定携带 `tag=网关插件`，并在前端二次筛选 `application_type=gateway-plugin`；随后与 `higress-system` 命名空间的 Higress `WasmPlugin`、同组 AppGroup 合并安装状态。未安装制品显示“待安装”，并可在列表直接进入统一制品安装页；市场中不存在的历史或手工 WasmPlugin 继续在“其他”分类展示。
- 网关插件列表和域名管理“更多”的插件列表统一按制品市场 `plugin_type` 分类：`auth`、`security`、`traffic`、`transform`、`o11y`、`ai` 分别显示“认证鉴权、安全防护、流量管控、请求响应转换、可观测性、AI”；无分类显示“其他”，未知分类保留市场原始名称，没有插件的分类不渲染。域名列表仍只展示已安装且支持规则配置的插件。
- 插件分类默认展开，分类标题显示插件数量并支持独立展开或收起；所有分类使用一个 4px 圆角的扁平列表容器，分类之间以中性分隔线区分。收起标题使用面板白底列表行，展开标题使用主色浅背景和左侧状态线；分类内表格隐藏重复表头，直接展示插件数据行。折叠状态只保留在当前页面生命周期内，不写入 localStorage/sessionStorage。
- 网关插件列表不单独展示前端包；关联的 MicroApp 仅用于决定配置时加载插件页面还是回退 YAML。
- 网关插件列表的“插件”列固定为 360px，避免在宽屏下过度拉伸。
- 网关插件列表与域名规则插件列表统一按“标题、标识+版本、描述”三层展示，标题使用常规字重，标识版本和描述使用 12px 灰色辅助文字。
- `w7.cc/manifest-type=gateway-plugin` 的插件应用只在“网关管理 → 网关插件”中管理，不在顶部菜单、应用直达和普通应用列表中展示。
- 域名管理“更多”的插件列表只展示插件、规则状态和操作列，不展示“配置方式”列；配置时仍根据关联 MicroApp 自动打开操作界面或回退 YAML。
- 网关插件列表使用“全局状态”列直接控制 Higress `spec.defaultConfigDisable`，不能连带修改 `matchRules[].configDisable`；只支持规则配置的插件不显示全局开关。无编辑权限时禁用开关，切换期间显示加载状态。
- 网关插件通过 `metadata.labels["w7.cc/group-name"]` 关联 `default` 命名空间同名 AppGroup；AppGroup 带 `metadata.annotations["w7.cc/official-app"] == "true"` 时，网关插件列表和域名管理“更多”列表都在插件标题右侧显示紧凑的“官方”标识，并禁止编辑插件元数据。是否允许卸载独立读取关联 AppGroup 的 `metadata.annotations["w7.cc/deny-delete"]`：值为 `"true"` 时禁止卸载，否则允许卸载。官方身份不能隐含卸载保护，删除保护也不能用于推断官方身份；插件启停及全局/规则配置不受这两个注解影响。此类资源统一称为“官方应用提供的网关插件”，不要称为“官方插件”。
- 网关插件列表的“卸载”必须删除插件关联的 AppGroup，由 AppGroup Controller 按标准应用卸载流程清理同组 WasmPlugin、MicroApp 及其他关联资源，禁止直接删除单个 WasmPlugin。没有关联 AppGroup 的历史或手工 WasmPlugin 不显示应用卸载入口。
- 网关插件更新通过 `w7.cc/group-name` 关联的 AppGroup 调用 `/panel-api/v1/zpk/upgrade-info` 检查，同一 AppGroup 只请求一次并向组内插件共享结果；“立即更新”必须携带 AppGroup 名称作为 `releaseName` 进入统一制品更新页，更新整个应用制品而不是单独修改 WasmPlugin。
- 网关插件列表顶部的添加按钮与搜索框至少保留 12px 间距。
- 网关插件列表应在表格上方提示“全局状态仅控制插件的全局配置，不影响域名规则”；域名管理“更多”的插件列表应提示“规则状态仅控制当前域名，不影响插件的全局配置或其他域名”。提示框与后续表格至少保留 12px 间距。
- 网关插件列表右侧操作的确认浮层向左展开并限制内容宽度，避免浮层打开时触发页面横向滚动。
- 插件支持范围等扩展能力写入 `metadata.annotations`，实际全局和规则开关分别使用 Higress 原生 `spec.defaultConfigDisable` 与 `spec.matchRules[].configDisable`；制品安装的 WasmPlugin 与 MicroApp 通过共同的 `metadata.labels["w7.cc/group-name"]` 关联，同组必须唯一匹配。兼容尚未写入分组标签的旧 MicroApp 时，只允许用与分组完全同名的 `metadata.name` 精确匹配，不再读取 `w7.cc/plugin-microapp`。
- 编辑插件时取消“支持规则配置”必须明确提示：插件将从域名“更多”中隐藏，所有已启用的 `matchRules` 会停用，重新开启支持范围不会自动恢复；用户确认后才能保存。
- 全局配置的前端包只读取创始人端（兼容旧名称 `found`）菜单并竖排展示，不显示普通用户端菜单；规则配置只读取普通用户端菜单并横排展示。当前作用域过滤后仅剩一个菜单项时，直接加载该页面，不显示左侧或顶部菜单栏。
- MicroApp 当前作用域的 `bindings.menu` 有菜单项时才视为配置了前端包，并按照统一的静态资源与 Wujie/iframe 流程加载；无菜单或加载失败时回退 YAML。
- 关联到可用 MicroApp 配置页面时必须提供统一的“YAML 详情”入口；未关联到可用页面或加载失败时直接进入同一个 YAML 界面。两种入口都默认只读预览，通过“编辑”按钮切换为可修改状态。
- 全局配置和域名规则按 Higress 原生逻辑独立启停：关闭全局状态不能关闭或隐藏路由规则，关闭某条路由规则也不能影响全局配置。
- 网关插件 MicroApp 注入只读的 `globalPluginConfig`、`globalPluginEnabled` 上下文，供全局或规则配置页面判断全局状态；仅全局配置页面额外注入 `ruleConfigs`（`spec.matchRules` 原始配置的深拷贝），规则配置页面不注入该字段。这些字段不得作为保存目标。`savePluginConfig` 必须严格按 `configScope` 写入：全局入口只能保存 `spec.defaultConfig`，规则入口只能保存当前 Ingress 的 `matchRules[].config`。
- Higress 的 `defaultConfigDisable` 或已有 `matchRules[].configDisable` 未设置时按 `false`（启用）处理；没有匹配规则时才显示为规则未启用。
- 域名规则兼容 Higress 原生 `ingress`、`domain`、`service` 三种匹配目标；修改同时覆盖多个同类目标的共享规则前，必须拆出当前匹配目标并保留其他目标的原配置和状态。
- Ingress 的 Host/Path 重写只写 Ingress annotation，不得向所有 WasmPlugin 的 `matchRules[].config` 注入 `rewrite_host` 等插件私有字段。
- 删除域名或路径对应的 Ingress 前，必须从所有 WasmPlugin 规则中移除该 `namespace/ingressName`；共享规则保留其他匹配目标，避免同名 Ingress 重建后旧插件配置意外恢复。
- 域名管理“更多”显示所有已安装且支持规则配置的插件，不受全局状态影响；不能在域名页面安装或卸载插件。
- 域名管理“更多”的提示框与插件表格之间保留 12px 间距。

#### AI 代理消费者 UI 规范

- AI 代理复用的 `key-auth.internal` 和 `request-validation.internal` 是网关通用插件，展示名称固定为“Key Auth 认证”和“请求校验”，不得添加 AI 专属前缀。
- AI 代理、Key Auth、请求校验的固定制品 identify 分别为 `w7panel-pluginaiproxy`、`w7panel-pluginkeyauth`、`w7panel-pluginrequestvalidation`，其规范化 identify 直接作为 `w7.cc/group-name` 定向查询 WasmPlugin；无需先查询 AppGroup，也不得按 `*.internal` 资源名兜底。AI 代理列表初始化只检测 AI 代理插件，Key Auth 与请求校验仅在域名配置页检测；列表删除任务因清理关联规则可按需解析后两者。
- AI 代理未安装时禁用新增并引导安装 `https://zpk.w7.cc/zpk/respo/info/w7panel-pluginaiproxy`；Key Auth 未安装时禁用认证开关并在同一栏引导安装 `https://zpk.w7.cc/zpk/respo/info/w7panel-pluginkeyauth`；请求校验未安装时禁用模型名称输入并引导安装 `https://zpk.w7.cc/zpk/respo/info/w7panel-pluginrequestvalidation`。这些入口不得运行时请求制品市场列表。
- AI 代理、Key Auth 与请求校验都按域名或服务 `matchRules` 生效，不能用 `spec.defaultConfigDisable` 判断 AI 代理域名、认证或模型限制是否启用；该字段只控制插件无规则命中时的全局配置。
- AI 代理列表页检测到 AI Proxy WasmPlugin 已安装后，应复用域名管理“更多”的规则方法，按 `ingress: ["namespace/ingressName"]` 检测并自动开启当前 AI 域名规则；规则不存在时创建，存在共享目标时拆出当前目标。不得使用 `domain`、裸 Ingress 名或 Provider `service` rule 判断域名插件状态。域名编辑页不提供 AI 代理总开关，Provider service rule 仍负责 `activeProviderId`。不得修改 `spec.defaultConfigDisable`，也不得另建 Ingress 注解保存重复启停状态。
- AI 代理页面禁止自动创建 `ai-proxy.internal`、`key-auth.internal` 或 `request-validation.internal`，也不得用硬编码镜像和版本标签覆盖制品安装结果；缺少依赖时必须引导安装或阻止相关配置保存。
- 域名编辑页的消费者使用列表展示，新增和编辑使用弹窗；消费者名称由系统自动生成并始终只读。
- 认证方式使用 Tab 组织。当前只展示 Key Auth，OAuth2、JWT 等未实现方式不得提前显示。
- Key Auth 表单对齐 Higress Console：支持多个认证令牌，并支持 Bearer Token、自定义 HTTP Header、查询参数三种令牌来源；Header 和 Query 来源必须显示并校验对应名称字段。
- 删除消费者必须同步清理消费者 Secret、`key-auth.internal` 的 consumer 配置和当前域名 `allow` 引用；删除最后一个消费者时同步关闭该域名认证。

#### AI 代理服务提供者 UI 规范

- 新增、编辑服务提供者的字段、联动和候选值必须对齐 Higress Console ProviderForm，不能用一个通用“服务地址”代替供应商专属配置。
- OpenAI、Qwen、Claude 必须区分官方服务与自定义服务；Azure 使用包含 `api-version` 的完整服务 URL；OpenAI 和 vLLM 支持多个同协议、同路径的静态 IP URL。
- Qwen 支持搜索、兼容模式、文件 ID、自定义域名和推理内容处理模式；Bedrock、Vertex 的区域使用 Higress 候选列表并允许搜索；Vertex 支持 Gemini 安全类别和阈值设置。
- 通用配置支持流式首包超时和 Token 故障转移。健康检查模型优先展示当前供应商的 Higress 预置模型，同时允许输入自定义模型名称。
- 代理服务器从 `higress-system/default` McpBridge 的 `spec.proxies` 读取，并写入当前 Provider 对应 registry 的 `proxyName`；不得写入 `ai-proxy` Provider 配置。
- 表单字段必须转换为 `ai-proxy.internal.spec.defaultConfig.providers[]` 的官方字段；多 URL、`failover`、`retryOnFailure`、`geminiSafetySetting` 等配置不能只保存在页面状态中。

#### AI 代理删除任务 UI 规范

- 删除 AI 代理必须使用任务弹窗依次展示“消费者数据、服务提供者数据、域名配置数据”的等待、执行、成功或失败状态，任务执行期间禁止关闭弹窗。
- 任务弹窗复用 w7panel 现有任务状态风格：顶部居中展示 80px 成功/失败图标或 60px 旋转加载图，下方使用中性边框任务列表，单行右侧展示图标和“已删除、删除中、未执行、失败”，结束操作按钮居中排列。
- 每一步不能只以 DELETE/PUT 请求成功作为完成条件，必须重新读取 Secret、WasmPlugin、McpBridge registry 或 Ingress，确认该步骤的关联资源已经不存在后才显示“已删除”并进入下一步。
- 删除任务必须幂等。部分步骤成功后失败时保留错误信息和重试入口；重试应跳过已确认完成的步骤，并继续清理未完成资源。
- 资源查询失败不得当作资源不存在。只有明确的 HTTP 404 可以作为单资源已删除的依据，列表查询或其他网络错误必须令任务失败。

#### 资源列表 UI 规范

- 镜像管理的镜像 ID 作为名称的辅助信息放在名称下方，使用 12px 灰色小字，不单独占用表格列。
- 私有 DNS 属于网关管理能力，菜单名固定为“私有DNS”并放在“网关管理”下。

### 性能规范

**后端性能规范:**

1. **文件操作限制**
   - 最大目录条目: 5000
   - 使用 `io.LimitedReader` 限制读取大小

2. **内存管理**
   - 避免一次性加载大文件到内存
   - 使用流式处理 (stream) 而非缓冲处理
   - 大文件使用 `io.Copy` 直接传输
   - 及时关闭文件句柄和释放资源

3. **文件类型处理**
   - device/fifo/socket 不可读写，返回空内容
   - 符号链接安全验证：防止路径穿越

4. **符号链接处理**
   - 使用 `os.Lstat()` 获取链接信息
   - 使用 `os.Readlink()` 读取链接目标
   - 安全验证：防止符号链接逃逸容器根目录

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
   - 封装请求逻辑时必须提供取消机制

5. **Hooks 使用规范**

- 复用现有 hooks 前先确认文件仍存在，避免引用历史遗留路径。
- 定时器、轮询、WebSocket、xterm、CodeMirror、ECharts 实例必须在组件卸载或抽屉关闭时清理。
- keep-alive 页面需要在 `onDeactivated` 中停止轮询，在 `onActivated` 中按需恢复。

**API 请求示例**:
```typescript
// src/api/cluster.ts
// compressUrl 从 /panel-api/v1/pid 接口返回，格式: /panel-api/v1/files/compress-agent/{pid}
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
| WebDAV PROPFIND 400 错误 | Content-Type 必须是 `text/xml; charset=utf-8` |
| 文件管理列表空白 | 检查 WebDAV API 是否正常，查看浏览器控制台日志 |
| WebDAV 401 认证失败 | 需要有效的 K8S Token，从 kubeconfig 或 ServiceAccount 获取 |
| agent-browser 找不到元素 | 使用 `snapshot -i` 查看交互元素，用 ref (@e1) 定位 |

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

**Helm Charts 项目**: `$BASE_DIR/charts/`

---

## Helm Charts 维护

### 项目结构

```
$BASE_DIR/charts/w7panel/
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
helm upgrade --install w7panel ./charts/w7panel -n default \
  --set image.tag=1.0.20
```

---

## 故障排查

---
