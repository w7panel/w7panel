# W7Panel 文档中心

这里汇总项目的用户文档、开发文档、部署文档与测试资料。项目总览、能力介绍、安装入口与快速导航已统一整理到仓库根目录的 [README.md](https://github.com/w7panel/w7panel/blob/dev-v1/README.md)。

## 文档前端

`docs/docs/` 目录已接入 VitePress 文档前端项目，侧边栏索引位于 [docs/src/1.x/sidebar.js](./docs/src/1.x/sidebar.js)。

```bash
cd docs/docs
pnpm install
pnpm run dev
```

构建静态文档：

```bash
cd docs/docs
pnpm run build
```

发布流程：

- `.github/workflows/release.yml` 会在发布时构建 `docs/docs`。
- 构建产物 `docs/docs/.vitepress/dist` 会打包进 nginx 静态资源镜像，默认放在 `/usr/share/nginx/html`。
- 文档镜像推送为 `${IMAGE_PUSH_REGISTRY}/${owner/repo}-docs:${tag}`，默认是 `ghcr.io/w7panel/w7panel-docs:${tag}`。
- 文档服务固定按根路径 `/` 构建和部署。

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
├── docs/         # VitePress 文档前端
│   └── src/
│       └── development/ # 开发指南、API、前端组件和 Wujie 微前端事件说明
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

- [部署文档](./deployment/README.md)
- [部署排障](./deployment/troubleshooting.md)
- [开发指南](./docs/src/development/)
- [API 文档](./api/README.md)
- [API 调用凭据与认证接口说明](./docs/src/development/api/credentials.md)
- [集群运维 API 接口说明](./docs/src/development/api/cluster-ops.md)
- [指标 API 接口说明](./docs/src/development/api/metrics.md)
- [Longhorn API 接口说明](./docs/src/development/api/longhorn.md)
- [容器文件管理 API 接口说明](./docs/src/development/api/container-files.md)
- [容器镜像管理 API 接口说明](./docs/src/development/api/container-images.md)
- [应用管理 API 接口说明](./docs/src/development/api/zpk.md)
- [前端开发文档](./docs/src/development/frontend/)
- [前端微应用接入说明](./docs/src/development/frontend/microapps.md)
- [w7panel-server/app/application API 接口说明](./docs/src/development/api/application.md)
- [云主机 API 接口说明](./docs/src/development/api/k3k.md)
- [测试文档](./testing/README.md)
- [版本日志](./changelog/1.0.0.md)

### 子项目说明

- `../w7panel/`：后端源码
- `../w7panel-ui/`：前端源码
- `../codeblitz/`：Web IDE 源码
- `../tests/`：测试脚本与测试资料

## 说明

- 如果你想先了解产品能力、技术架构、适用场景和安装方式，请先阅读仓库根目录的 [README.md](https://github.com/w7panel/w7panel/blob/dev-v1/README.md)。
- 如果你已经明确要查找某类文档，可以直接从本页进入对应子目录。
