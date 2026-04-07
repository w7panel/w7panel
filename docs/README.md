# W7Panel 离线版文档

## 项目简介

W7Panel（微擎面板）是一款基于 Kubernetes 的云原生应用管理平台，由微擎团队开发维护。它提供了简洁直观的 Web 界面，帮助用户轻松管理 Kubernetes 集群和应用，无需掌握复杂的 kubectl 命令。

### 核心特性

- **可视化集群管理** - 实时监控集群资源使用情况，管理节点和资源对象
- **一键应用部署** - 支持应用商店、Docker Compose、Helm、YAML 等多种部署方式
- **在线文件管理** - WebDAV 协议支持，提供文件上传、编辑、压缩解压等完整操作
- **内置 Web IDE** - 基于 Codeblitz 的在线代码编辑器，支持语法高亮和实时保存
- **灵活域名管理** - 支持自定义域名绑定和 Let's Encrypt 自动 HTTPS 证书
- **持久化存储** - 支持多种存储类型，提供存储分区和数据备份功能

### 技术架构

| 组件 | 技术栈 | 说明 |
|------|--------|------|
| 后端 | Go 1.24 + Gin + w7-rangine-go | RESTful API、WebDAV代理、K8S交互 |
| 前端 | Vue 3.5 + TypeScript + Arco Design | 响应式管理界面 |
| Web IDE | React + Codeblitz (OpenSumi) | 在线代码编辑器 |
| 容器运行时 | containerd / Docker | 兼容主流容器运行时 |
| 存储 | Longhorn / NFS / Local | 多种存储后端支持 |

### 适用场景

- **个人开发者** - 快速部署和管理个人项目
- **中小企业** - 简化 K8S 集群运维，降低学习成本
- **开发团队** - 提供统一的开发、测试、部署平台
- **私有化部署** - 支持离线环境，数据完全自主可控

## 文档目录

### 用户文档

面向面板使用者的操作手册：

```
docs/user-guide/
├── README.md            # 快速入门
├── app-management.md    # 应用管理
├── file-management.md   # 文件管理
├── storage-management.md # 存储管理
├── domain-management.md # 域名管理
├── cluster-management.md # 集群管理
└── faq.md             # 常见问题
```

### 开发者文档

面向开发和运维人员的技术文档：

```
docs/
├── api/           # API接口文档
├── deployment/   # 部署文档
├── development/   # 开发指南
├── refactoring/  # 重构方案 (v5.3)
├── testing/       # 测试文档和报告
│   ├── ui/       # UI测试报告
│   ├── performance/ # 性能测试报告
│   └── backend/  # 后端测试报告
└── changelog/    # 版本更新日志
```

### 项目 README

各项目独立 README：

```
w7panel/           # 后端 (Go) - 技术栈、项目结构、快速开始
w7panel-ui/        # 前端 (Vue 3) - 技术栈、组件、页面说明
codeblitz/         # Web IDE - 编辑器配置、功能特性
tests/              # 测试脚本 - 测试脚本说明、运行方式
```

## 快速开始

```bash
# 设置环境变量
export BASE_DIR=/home/wwwroot/w7panel-dev

# 开发模式启动（需要 kubeconfig.yaml）
cd $BASE_DIR/dist
export KUBECONFIG=$BASE_DIR/kubeconfig.yaml
./w7panel-ctl.sh start

# 访问面板
# http://localhost:8080/
# 用户名: admin
# 密码: 123456
```

## 核心功能

### 集群管理
- **概览仪表盘** - CPU/内存/硬盘使用率实时监控，节点/应用/域名统计
- **节点管理** - 节点注册、镜像源配置、内存优化、节点封锁/驱逐
- **资源对象浏览器** - 浏览 K8S 所有资源对象（Pod、Service、ConfigMap、Secret 等）
- **集群终端** - Web 终端直接访问集群
- **配置字典** - ConfigMap 创建和管理
- **密钥管理** - Secret 创建和管理
- **证书管理** - TLS 证书管理

### 应用管理
- **应用列表** - 查看和管理已部署应用
- **应用商店** - 在线应用模板一键部署
- **Helm 应用** - Helm Chart 部署和管理
- **Docker Compose** - Compose 文件部署
- **YAML 创建** - 直接提交 K8S YAML
- **代码包创建** - 上传代码包部署
- **计划任务** - CronJob 定时任务管理
- **反向代理** - Ingress 代理规则配置
- **集群数据库** - 数据库服务管理（MySQL/PostgreSQL/Redis 等）
- **AI 应用管理** - GPUStack GPU 应用部署

### 应用详情
- **应用信息** - 状态、配置、资源使用
- **容器列表** - Pod 和容器状态
- **文件管理** - WebDAV 文件操作、在线编辑、压缩解压
- **域名管理** - 域名绑定、HTTPS 证书
- **运行状态** - 实时监控图表
- **事件日志** - K8S 事件查看
- **历史版本** - Revision 历史和回滚
- **执行脚本** - 一次性任务执行

### 存储管理
- **存储设备** - 节点磁盘监控、I/O 性能统计
- **存储分区** - PVC 创建、扩容、快照备份
- **Longhorn 管理** - 分布式存储管理

### 制品管理（ZPK）
- **ZPK 制品** - 安装、升级、卸载
- **传统应用** - 传统 PHP 应用安装
- **镜像构建** - Job/CronJob 构建镜像
- **Helm 生成** - 从代码生成 Helm Chart

### 系统管理
- **云配置** - 云账号绑定
- **API 密钥** - API 访问密钥管理
- **许可证** - 授权管理
- **权限策略** - RBAC 权限配置
- **订单中心** - 订单管理
- **费用中心** - 资源费用统计

### 用户管理
- **用户列表** - 用户创建和管理
- **用户组** - 用户组管理
- **权限策略** - 角色和权限分配
- **用户资源** - 用户资源配额
- **白名单** - 域名白名单

### 其他功能
- **K3S 优化** - GoGC 内存优化
- **KubeBlocks** - 数据库运维
- **GPU 管理** - HAMi/GPU Operator
- **DNS 工具** - DNS 解析测试
- **连接测试** - 数据库/ETCD 连接测试

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BASE_DIR` | . | 项目根目录 |
| `LOCAL_MOCK` | true | 开发模式 |
| `CAPTCHA_ENABLED` | true | 验证码开关 |
| `KO_DATA_PATH` | ./kodata | 静态资源路径 |

## 相关链接

### 用户文档
- [快速入门](./user-guide/README.md)
- [集群管理](./user-guide/cluster-management.md)
- [应用管理](./user-guide/app-management.md)
- [文件管理](./user-guide/file-management.md)
- [存储管理](./user-guide/storage-management.md)
- [域名管理](./user-guide/domain-management.md)
- [常见问题](./user-guide/faq.md)

### 开发者文档
- [API文档](./api/README.md)
- [部署文档](./deployment/README.md)
- [开发指南](./development/README.md)
- [测试文档](./testing/README.md)
