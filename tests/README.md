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
BASE_URL=http://localhost:8080 TOKEN=xxx bash run.sh

# 或直接运行单个测试
BASE_URL=http://localhost:8080 TOKEN=xxx bash test-cases/api/webdav-list.sh
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
BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-xxx}"

# 测试代码
echo "执行测试..."

# 退出码: 0成功, 1失败
```

## 测试工具

### Go 权限单元测试

网关插件菜单权限由后端权限服务维护，可运行定向测试确认创始人默认权限完整：

```bash
cd $BASE_DIR/w7panel-server
go test ./common/service/k8s/permission -run TestFounderFallbackIncludesGatewayPluginPermissions -count=1
```

### agent-browser

用于UI测试的浏览器自动化工具：

```bash
agent-browser open "http://localhost:8080"
agent-browser snapshot -i
agent-browser click @e1
agent-browser close
```

详见 [.opencode/skills/agent-browser](../.opencode/skills/agent-browser/SKILL.md)

## 相关文档

- [test-case-creation技能](../.opencode/skills/test-case-creation/SKILL.md)
- [UI菜单地图](../docs/testing/ui/ui-menu-map.md)
