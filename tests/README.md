# W7Panel 测试脚本

项目测试脚本集，包含自动化测试用例。

## 目录结构

```
tests/
├── run.sh                    # 统一运行器
├── test-cases/              # 测试用例 (可执行脚本)
│   ├── api/                # API测试
│   │   ├── webdav-list.sh
│   │   └── compress.sh
│   ├── ui/                 # UI测试
│   │   └── login.sh
│   └── performance/        # 性能测试
│       └── webdav-perf.sh
├── reports/                 # 测试报告
└── *.sh                    # 原有测试脚本
```

## 快速开始

### BootstrapProfile 单元测试

后端测试会严格解析并校验内置
`w7panel-server/kodata/yaml/bootstrap-profile.yaml`，检查应用清单、职责边界和 Bootstrap 内部所有权注解：

```bash
cd "$BASE_DIR/w7panel-server"
go test ./common/service/k8s/bootstrap
```

Bootstrap 使用 ServiceAccount Token 安装 ZPK 时，`PackageApp.K8sToken` 允许为空；以下回归测试确认 Helm Job 会安全回退到 `RealToken`，且用户 Token 存在时仍优先使用用户 Token：

```bash
cd "$BASE_DIR/w7panel-server"
LOCAL_MOCK=true go test ./app/zpk/logic -run TestPackageTokens -count=1
LOCAL_MOCK=true go test ./common/service/k8s/bootstrap -run TestClusterTokenFromRESTConfig -count=1
```

### 运行测试

```bash
cd /home/wwwroot/w7panel-dev/tests

# 运行所有测试
bash run.sh

# 只运行API测试
bash run.sh api

# 只运行UI测试
bash run.sh ui

# 调试模式（显示输出）
DEBUG=1 bash run.sh

# 单个测试
bash test-cases/api/webdav-list.sh
```

### 环境变量

```bash
# 配置服务地址和Token
BASE_URL=http://localhost:8000 TOKEN=xxx bash run.sh

# 或直接运行单个测试
BASE_URL=http://localhost:8000 TOKEN=xxx bash test-cases/api/webdav-list.sh
```

## 测试用例

| 脚本 | 类型 | 说明 |
|------|------|------|
| `test-cases/api/webdav-list.sh` | API | WebDAV目录列表 |
| `test-cases/api/compress.sh` | API | 压缩功能 |
| `test-cases/ui/login.sh` | UI | 登录测试 |
| `test-cases/performance/webdav-perf.sh` | Performance | 性能测试 |

## 创建新测试用例

```bash
#!/bin/bash
#========================================
# 测试用例: 功能名称
#========================================
#
# ## 测试信息
# | 项目 | 内容 |
# |------|------|
# | 类型 | API/UI/Performance |
# | 优先级 | P0/P1/P2 |
#
# ## 环境变量
# BASE_URL - 服务地址
# TOKEN    - 认证Token
#
#========================================

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
TOKEN="${TOKEN:-xxx}"

# 测试代码
echo "执行测试..."

# 退出码: 0成功, 1失败
```

## 测试工具

### LOCAL_MOCK 鉴权单元测试

验证测试模式只改变 Kubernetes 访问方式，不会绕过面板用户 Token 校验：

```bash
cd $BASE_DIR/w7panel-server
go test ./common/middleware -run TestAuthRequiresTokenInLocalMockMode -count=1
```

### Go 权限单元测试

网关插件菜单权限由后端权限服务维护，可运行定向测试确认创始人默认权限完整：

```bash
cd $BASE_DIR/w7panel-server
go test ./common/service/k8s/permission -run TestFounderFallbackIncludesGatewayPluginPermissions -count=1
```

### Higress 插件迁移兼容测试

验证域名白名单读取时优先选择制品资源，并继续兼容旧固定资源名：

```bash
cd $BASE_DIR/w7panel-server
go test ./common/service/k8s/higress -run TestPreferredWasmPlugin -count=1
```

### BootstrapProfile 协调单元测试

验证通用 Lease 的持有者隔离、过期接管、安全释放、原槽冲突重试和分布式并发限制，以及 Bootstrap Profile revision 状态汇总、失败 revision 清理重装、依赖失败策略、AppGroup 真实 Ready 判定、安装超时与 Lease 生命周期：

```bash
cd $BASE_DIR/w7panel-server
LOCAL_MOCK=true go test ./k8s/pkg/apis/bootstrapinstallation/v1alpha1 ./common/service/k8s/coordination ./common/service/k8s/bootstrap -count=1
```

### 制品域名传递测试

验证面板会移除仓库 URL 中不受信任的 `reinstall`，仅在用户确认强制覆盖时重新添加受控标记；同时验证域名、应用标识和 HTTP 409 订单绑定冲突解析：

```bash
cd $BASE_DIR/w7panel-server
LOCAL_MOCK=true go test ./app/zpk/logic -run TestLoadPackageByHTTPPassesDomainWithoutReinstall -count=1
LOCAL_MOCK=true go test ./app/zpk/logic -run TestLoadPackageByHTTPReturnsArtifactInstallConflict -count=1
LOCAL_MOCK=true go test ./app/zpk/logic -run TestLoadPackageByHTTPReturnsConflictWhenProxyChangesStatus -count=1
LOCAL_MOCK=true go test ./app/zpk/logic -run TestLoadPackageByHTTPPassesControlledReinstall -count=1

cd $BASE_DIR/../w7panel-zpk
LOCAL_MOCK=true go test ./app/respo/logic -run TestTicketPreservesDomain -count=1

cd $BASE_DIR/../cd-artifact-market
LOCAL_MOCK=true go test ./app/respo/logic -run '^(TestValidateOrderDomain|TestOrderAppIdentifyMatches|TestOrderBindingConflictReason|TestDiscardUsedOrderClearsInstallationBinding|TestUseOrderReinstallOverwritesInstallationBinding)$' -count=1
```

### 插件微应用入口过滤测试

验证带有 `w7.cc/manifest-type=gateway-plugin` 注解的 MicroApp 不会进入顶部菜单和“应用直达”共用列表：

```bash
cd $BASE_DIR/w7panel-server
LOCAL_MOCK=true go test ./common/service/k8s/microapp -run TestIsPluginMicroApp -count=1
```

### agent-browser

用于UI测试的浏览器自动化工具：

```bash
agent-browser open "http://localhost:8000"
agent-browser snapshot -i
agent-browser click @e1
agent-browser close
```

详见 [.opencode/skills/agent-browser](../.opencode/skills/agent-browser/SKILL.md)

## 相关文档

- [test-case-creation技能](../.opencode/skills/test-case-creation/SKILL.md)
- [UI菜单地图](../docs/testing/ui/ui-menu-map.md)
