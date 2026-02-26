# W7Panel 后端

基于 Go + Gin 的 Kubernetes 云原生应用管理平台后端服务。

## 技术栈

- **Go 1.24** - 编程语言
- **Gin** - Web 框架
- **w7-rangine-go** - 应用框架
- **Kubernetes** - 容器编排
- **SQLite** - 嵌入式数据库

## 项目结构

```
w7panel/
├── app/                      # 业务应用
│   └── application/http/controller/  # HTTP 控制器
├── common/                   # 公共模块
│   ├── service/             # 业务服务
│   ├── middleware/          # 中间件
│   └── helper/              # 工具函数
├── install/                  # 安装相关
│   └── charts/             # Helm Charts
├── scripts/                  # 构建脚本
├── kodata/                  # 静态资源
└── config.yaml              # 配置文件
```

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+ (用于前端构建)
- Kubernetes 集群

### 开发模式

```bash
# 设置环境变量
export BASE_DIR=/home/wwwroot/w7panel-dev

# 编译
cd $BASE_DIR/w7panel
go build -o ../dist/w7panel .

# 启动服务
cd $BASE_DIR/dist
CAPTCHA_ENABLED=false \
LOCAL_MOCK=true \
KO_DATA_PATH=$BASE_DIR/dist/kodata \
KUBECONFIG=$BASE_DIR/kubeconfig.yaml \
./w7panel server:start
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LOCAL_MOCK` | true | 开发模式 |
| `CAPTCHA_ENABLED` | true | 验证码开关 |
| `KO_DATA_PATH` | ./kodata | 静态资源路径 |
| `KUBECONFIG` | ./kubeconfig.yaml | K8S 配置 |
| `W7PANEL_OFFLINE_HTTP_SERVER_PORT` | 8080 | HTTP 端口 |

## 主要功能

- **WebDAV 文件管理** - 容器内文件在线管理
- **压缩/解压** - 支持 zip, tar, tar.gz, tar.xz
- **权限管理** - chmod, chown 操作
- **应用部署** - Helm, Docker Compose, YAML
- **集群管理** - 节点、资源对象管理

## API 接口

详见 [API 文档](../docs/api/README.md)

## 测试

```bash
# 运行 API 测试
cd $BASE_DIR/tests
bash webdav.sh

# 运行压缩功能测试
bash compress.sh
```

## 相关文档

- [部署文档](../docs/deployment/README.md)
- [开发指南](../docs/development/README.md)
- [测试文档](../docs/testing/README.md)
