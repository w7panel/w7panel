# W7Panel 文档中心

这里汇总 W7Panel 的用户手册、开发文档、部署说明与版本记录。项目总览、能力介绍和安装入口请先查看仓库根目录的 [README.md](../README.md)。

## 文档站

`docs/` 是 VitePress 文档站源码目录，主入口位于 [src/index.md](./src/index.md)，侧边栏配置位于：

- [用户文档侧边栏](./src/user-guide/sidebar.js)
- [开发文档侧边栏](./src/development/sidebar.js)

本地预览：

```bash
cd docs
pnpm install
pnpm run dev
```

构建静态站点：

```bash
cd docs
pnpm run build
```

发布流程：

- `.github/workflows/release.yml` 会在发布时构建 `docs`。
- 构建产物 `docs/.vitepress/dist` 会打包进 nginx 静态资源镜像，默认放在 `/usr/share/nginx/html`。
- 文档镜像推送为 `${IMAGE_PUSH_REGISTRY}/${owner/repo}-docs:${tag}`，默认是 `ghcr.io/w7panel/w7panel-docs:${tag}`。
- 文档服务固定按根路径 `/` 构建和部署。

## 目录结构

当前文档内容主要集中在 `docs/src`：

```
docs/
├── README.md              # 本说明
├── src/
│   ├── index.md           # 文档站首页
│   ├── user-guide/        # 用户操作手册、部署运维、概述与版本日志
│   ├── development/       # 开发文档、API、前端规范
│   └── public/            # 文档站静态资源
│       └── user-guide/    # 仓库 README 展示图片
├── .vitepress/            # VitePress 配置、主题和构建缓存
├── package.json
└── pnpm-lock.yaml
```

## 快速入口

### 用户文档

- [文档首页](./src/index.md)
- [用户文档首页](./src/user-guide/index.md)
- [快速入门](./src/user-guide/quick-start.md)
- [集群管理](./src/user-guide/cluster-management.md)
- [应用管理](./src/user-guide/app-management.md)
- [站点管理](./src/user-guide/site-management.md)
- [文件管理](./src/user-guide/file-management.md)
- [存储管理](./src/user-guide/storage-management.md)
- [域名管理](./src/user-guide/domain-management.md)
- [镜像管理](./src/user-guide/image-management.md)
- [计划任务](./src/user-guide/scheduled-tasks.md)
- [快速开始](./src/user-guide/quick-start.md)
- [常见问题](./src/user-guide/overview/faq.md)

### 开发文档

- [开发文档首页](./src/development/index.md)
- [API 文档首页](./src/development/api/index.md)
- [API 设计约定](./src/development/api/conventions.md)
- [认证与凭据 API](./src/development/api/credentials.md)
- [应用管理 API](./src/development/api/zpk.md)
- [容器文件 API](./src/development/api/container-files.md)
- [容器镜像 API](./src/development/api/container-images.md)
- [集群运维 API](./src/development/api/cluster-ops.md)
- [Longhorn API](./src/development/api/longhorn.md)
- [指标 API](./src/development/api/metrics.md)
- [前端开发文档](./src/development/frontend/index.md)
- [前端组件文档](./src/development/frontend/components.md)
- [微前端接入说明](./src/development/frontend/microapps.md)
- [Wujie 事件说明](./src/development/frontend/wujie-events.md)
- [ZPK 开发示例](./src/development/examples/zpk-product-workflow.md)

### 版本

- [版本日志](./src/user-guide/overview/changelog/1.0.0.md)

## 维护规则

- 新增用户操作流程时，优先更新 `docs/src/user-guide/`。
- 新增或调整 API 时，优先更新 `docs/src/development/api/`。
- 新增前端组件、接口调用或微前端事件时，优先更新 `docs/src/development/frontend/`。
- 修改部署方式、环境变量或故障处理步骤时，同步更新 `docs/src/user-guide/quick-start.md`、`docs/src/user-guide/overview/faq.md` 或新增相应部署专题文档。
- 完成面向用户或开发者的变更后，同步更新 `docs/src/user-guide/overview/changelog/1.0.0.md`。

### 子项目说明

- `../w7panel-server/`：后端服务源码
- `../w7panel-ui/`：Vue 管理端源码
- Web IDE：当前以 `../w7panel-server/kodata/plugin/codeblitz.zip` 静态包随服务分发；如恢复独立源码目录，再补充对应入口。
- `../tests/`：测试脚本与测试资料

## 说明

- 如果你想先了解产品能力、技术架构、适用场景和安装方式，请阅读仓库根目录的 [README.md](../README.md)。
- 如果要查看线上文档站源码和导航结构，请从 [src/index.md](./src/index.md) 开始。
