# W7Panel 文档中心

这里汇总项目的用户文档、开发文档、部署文档与测试资料。项目总览、能力介绍、安装入口与快速导航已统一整理到仓库根目录的 [README.md](../README.md)。

## 文档目录

### 用户文档

面向面板使用者的操作手册：

```
docs/user-guide/
├── README.md             # 快速入门
├── app-management.md     # 应用管理
├── file-management.md    # 文件管理
├── storage-management.md # 存储管理
├── domain-management.md  # 域名管理
├── cluster-management.md # 集群管理
└── faq.md                # 常见问题
```

### 开发与运维文档

面向开发、部署与维护人员的技术资料：

```
docs/
├── api/          # API 接口文档
├── deployment/   # 部署文档
├── development/  # 开发指南
├── refactoring/  # 重构方案与历史资料
├── testing/      # 测试文档和报告
└── changelog/    # 版本更新日志
```

## 快速入口

### 用户文档

- [快速入门](./user-guide/README.md)
- [集群管理](./user-guide/cluster-management.md)
- [应用管理](./user-guide/app-management.md)
- [文件管理](./user-guide/file-management.md)
- [存储管理](./user-guide/storage-management.md)
- [域名管理](./user-guide/domain-management.md)
- [常见问题](./user-guide/faq.md)

### 开发与部署文档

- [API 文档](./api/README.md)
- [部署文档](./deployment/README.md)
- [开发指南](./development/README.md)
- [测试文档](./testing/README.md)
- [版本日志](./changelog/1.0.0.md)

### 子项目说明

- `../w7panel/`：后端源码
- `../w7panel-ui/`：前端源码
- `../codeblitz/`：Web IDE 源码
- `../tests/`：测试脚本与测试资料

## 说明

- 如果你想先了解产品能力、技术架构、适用场景和安装方式，请先阅读 [../README.md](../README.md)。
- 如果你已经明确要查找某类文档，可以直接从本页进入对应子目录。
