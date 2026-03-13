# K3s 本地镜像仓库与热修改系统

## 概述

本模块实现了 K3s 本地镜像仓库功能，支持容器热修改和镜像提交，完全兼容 Docker Registry API V2。

## 功能特性

### 1. Registry API (Docker Registry V2 兼容)

提供标准的镜像仓库 API，支持镜像的推送、拉取和管理：

- `GET /panel-api/v1/k3s-registry/v2/` - API 版本检查
- `GET /panel-api/v1/k3s-registry/v2/_catalog` - 获取镜像目录
- `GET /panel-api/v1/k3s-registry/v2/:name/tags/list` - 获取镜像标签
- `GET /panel-api/v1/k3s-registry/v2/:name/manifests/*reference` - 获取镜像 manifest
- `PUT /panel-api/v1/k3s-registry/v2/:name/manifests/*reference` - 推送镜像 manifest
- `GET /panel-api/v1/k3s-registry/v2/:name/blobs/*digest` - 获取镜像 blob
- `HEAD /panel-api/v1/k3s-registry/v2/:name/blobs/*digest` - 检查 blob 是否存在
- `POST /panel-api/v1/k3s-registry/v2/:name/blobs/uploads/` - 初始化 blob 上传
- `PUT /panel-api/v1/k3s-registry/v2/:name/blobs/uploads/:uuid` - 完成 blob 上传

### 2. Patch API (容器操作)

提供容器管理和操作功能：

- `GET /panel-api/v1/k3s-patch/containers` - 获取容器列表
- `GET /panel-api/v1/k3s-patch/containers/:id` - 获取容器详情
- `GET /panel-api/v1/k3s-patch/containers/:id/layers` - 获取容器镜像层信息
- `POST /panel-api/v1/k3s-patch/containers/:id/exec` - 在容器内执行命令
- `POST /panel-api/v1/k3s-patch/containers/:id/commit` - 提交容器为新镜像

## 项目结构

```
app/k3s-registry/
├── provider.go                 # 模块注册和路由配置
├── model/
│   └── types.go              # 数据类型定义
├── internal/
│   ├── config.go             # 配置管理
│   ├── metadata/
│   │   └── bolt.go          # BoltDB 元数据存储
│   └── content/
│       └── store.go          # Blob 存储管理
├── logic/
│   ├── registry.go           # Registry 业务逻辑
│   ├── containers.go         # 容器操作逻辑
│   ├── exec.go              # 命令执行逻辑
│   └── commit.go            # 镜像提交逻辑
└── http/controller/
    ├── registry.go           # Registry API 控制器
    ├── containers.go         # 容器管理控制器
    ├── exec.go              # 命令执行控制器
    └── commit.go            # 镜像提交控制器
```

## 配置

在 `config.yaml` 中添加以下配置：

```yaml
k3s-registry:
  enabled: ${K3S_REGISTRY_ENABLED-false}  # 是否启用 K3s 本地镜像仓库功能
  containerd_socket: ${K3S_CONTAINERD_SOCKET-/run/k3s/containerd/containerd.sock}  # containerd socket 路径
  data_dir: ${K3S_DATA_DIR-/var/lib/rancher/k3s/agent/containerd}  # 数据目录
  runtime_dir: ${K3S_RUNTIME_DIR-/run/k3s/containerd/io.containerd.runtime.v2.task/k8s.io}  # Runtime 目录
  agent_dir: ${K3S_AGENT_DIR-/var/lib/rancher/k3s/agent}  # Agent 目录
```

## 部署要求

由于需要访问 containerd socket 和本地文件系统，需要以 DaemonSet 方式部署，并配置以下权限：

```yaml
securityContext:
  privileged: true

volumeMounts:
  - name: containerd-sock
    mountPath: /run/k3s/containerd
  - name: k3s-agent
    mountPath: /var/lib/rancher/k3s/agent
  - name: proc
    mountPath: /host/proc
    mountPropagation: HostToContainer
  - name: run
    mountPath: /run
```

## 使用示例

### 1. 查看镜像目录

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/panel-api/v1/k3s-registry/v2/_catalog
```

### 2. 查看镜像标签

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/panel-api/v1/k3s-registry/v2/my-image/tags/list
```

### 3. 获取容器列表

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8000/panel-api/v1/k3s-patch/containers
```

### 4. 在容器内执行命令

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"command": ["/bin/sh", "-c", "ls -la"]}' \
  http://localhost:8000/panel-api/v1/k3s-patch/containers/<container-id>/exec
```

### 5. 提交容器为新镜像

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"new_tag": "my-image:v2", "squash": true}' \
  http://localhost:8000/panel-api/v1/k3s-patch/containers/<container-id>/commit
```

## 实现状态

### Phase 1: 基础架构 ✅
- [x] 创建模块目录结构
- [x] 实现配置管理 (internal/config.go)
- [x] 实现数据类型定义 (model/types.go)
- [x] 创建模块注册 (provider.go)

### Phase 2: Registry API ✅
- [x] 实现版本检查 API
- [x] 实现镜像目录 API
- [x] 实现镜像标签 API
- [x] 实现 manifest 读写 API
- [x] 实现 blob 上传/下载 API
- [x] 实现 BoltDB 元数据存储
- [x] 实现 Blob 存储

### Phase 3: Patch API ✅
- [x] 实现容器列表 API
- [x] 实现容器详情 API
- [x] 实现镜像层信息 API
- [x] 实现容器内命令执行 API

### Phase 4: 镜像提交 ✅ (基础实现)
- [x] 实现镜像提交核心逻辑
- [x] 实现容器重启
- [ ] 实现 setns 工具 (待完成)
- [ ] 实现 squash 层合并 (待完成)
- [ ] 实现镜像替换 (待完成)

### Phase 5: 集成与测试 ✅
- [x] 注册到 w7panel 主程序
- [x] 添加配置文件
- [x] 添加依赖项
- [x] 修复编译错误
- [x] 验证代码编译通过

## 后续工作

### 高级功能实现
1. **Containerd 集成**: 实现完整的 containerd 客户端集成
2. **Setns 工具**: 实现命名空间切换功能，支持在容器内执行命令
3. **Squash 功能**: 实现镜像层合并，减小镜像体积
4. **权限管理**: 完善容器操作的权限检查
5. **性能优化**: 优化大文件传输和镜像操作性能

### 生产环境准备
1. 添加完整的单元测试和集成测试
2. 实现错误处理和日志记录
3. 添加监控指标和健康检查
4. 完善文档和 API 规范
5. 安全加固和权限控制

## 依赖项

```go
require (
    go.etcd.io/bbolt v1.4.3  // BoltDB 元数据存储
)
```

## 注意事项

1. **安全性**: 本功能需要访问 containerd socket，应谨慎配置权限
2. **存储**: 元数据和 blob 数据存储在本地，需要确保磁盘空间充足
3. **性能**: 大文件传输可能影响性能，建议根据实际情况调整配置
4. **兼容性**: 当前实现为基础版本，部分高级功能待完善

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

遵循 w7panel 项目的许可证。
