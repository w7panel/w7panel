# W7Panel 构建发布说明

本文档说明当前仓库的镜像、Helm Chart 和 GitHub Release 构建流程。

## 构建产物

| 产物 | 来源 | 输出位置 |
|------|------|----------|
| 前端静态资源 | `w7panel-ui` | `w7panel-ui/dist/` |
| 后端二进制 | `w7panel-server` | Docker 多阶段构建内生成 |
| 容器镜像 | `Dockerfile` + `docker buildx` | 推送到镜像仓库 |
| Helm Chart 包 | `charts/` | `charts/*.tgz` |

## 前置条件

本地构建需要提前准备：

- Docker，并启用 Buildx
- Node.js 20 或兼容版本
- npm
- Helm
- 可推送目标镜像仓库的登录凭证

检查命令：

```bash
docker buildx version
node --version
npm --version
helm version
```

## 构建流程

### 1. 构建前端

```bash
cd w7panel-ui
HUSKY=0 npm install --legacy-peer-deps
npm run build
cd ..
```

### 2. 复制前端资源

前端构建产物需要复制到后端静态资源目录，最终由镜像打包到 `/var/run/ko/`。

```bash
cp -R w7panel-ui/dist/. w7panel-server/kodata/
```

### 3. 构建并推送镜像

推荐使用 `Makefile`：

```bash
make build-image \
  IMAGE_REPOSITORY=ghcr.io/w7panel/w7panel \
  IMAGE_TAG=v1.0.0
```

默认构建平台：

```bash
linux/amd64,linux/arm64
```

如需调整平台：

```bash
make build-image \
  IMAGE_REPOSITORY=ghcr.io/w7panel/w7panel \
  IMAGE_TAG=v1.0.0 \
  PLATFORMS=linux/amd64
```

### 4. 打包 Helm Chart

```bash
make package-chart \
  HELM_PACKAGE_IMAGE_REPOSITORY=ghcr.registry.cdn.w7.cc/w7panel/w7panel \
  HELM_PACKAGE_IMAGE_TAG=v1.0.0 \
  HELM_CHART_VERSION=1.0.0 \
  HELM_APP_VERSION=v1.0.0
```

### 5. 一次性发布镜像和 Chart

```bash
make publish \
  IMAGE_REPOSITORY=ghcr.io/w7panel/w7panel \
  IMAGE_TAG=v1.0.0 \
  HELM_PACKAGE_IMAGE_REPOSITORY=ghcr.registry.cdn.w7.cc/w7panel/w7panel \
  HELM_PACKAGE_IMAGE_TAG=v1.0.0 \
  HELM_CHART_VERSION=1.0.0 \
  HELM_APP_VERSION=v1.0.0
```

## 本地测试镜像

只构建当前测试平台并加载到本地 Docker，不推送镜像：

```bash
make build-image-test \
  IMAGE_REPOSITORY=ghcr.io/w7panel/w7panel \
  IMAGE_TAG=v1.0.0 \
  TEST_PLATFORM=linux/amd64
```

生成的本地镜像 tag 为：

```bash
ghcr.io/w7panel/w7panel:v1.0.0-test
```

## Makefile 变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CHART_DIR` | `charts` | Helm Chart 目录 |
| `CHART_PACKAGE_DIR` | `charts` | Chart 包输出目录 |
| `IMAGE_REPOSITORY` | 读取 `charts/values.yaml` | 构建并推送的镜像仓库 |
| `IMAGE_TAG` | 读取 `charts/values.yaml` | 构建并推送的镜像 tag |
| `TAG` | 空 | `IMAGE_TAG` 的简写别名 |
| `PLATFORMS` | `linux/amd64,linux/arm64` | 推送镜像的目标平台 |
| `TEST_PLATFORM` | `linux/amd64` | 本地测试镜像平台 |
| `DOCKERFILE` | `Dockerfile` | Dockerfile 路径 |
| `HELM_PACKAGE_IMAGE_REPOSITORY` | 等于 `IMAGE_REPOSITORY` | 写入 Chart 包的镜像仓库 |
| `HELM_PACKAGE_IMAGE_TAG` | 等于 `IMAGE_TAG` | 写入 Chart 包的镜像 tag |
| `HELM_CHART_VERSION` | 去掉 `v` 前缀后的 `IMAGE_TAG` | Chart version |
| `HELM_APP_VERSION` | `IMAGE_TAG` | Chart appVersion |

查看完整帮助：

```bash
make help
```

## GitHub Actions 自动发布

自动发布工作流：

```bash
.github/workflows/release.yml
```

触发方式：

```bash
git tag v1.0.0
git push origin v1.0.0
```

也可以在 GitHub Actions 页面手动触发 `Release` workflow，并填写 `tag`。

### 自动发布步骤

1. 检出主仓库、子模块和 Git LFS 文件
2. 使用 Node.js 20 构建 `w7panel-ui`
3. 将 `w7panel-ui/dist/` 复制到 `w7panel-server/kodata/`
4. 登录镜像仓库
5. 执行 `make publish`
6. 将 `charts/*.tgz` 上传到 GitHub Release
7. 使用 `zpk` 将 Helm 包推送到 ZPK 平台

### GitHub Variables

| 名称 | 默认值 | 说明 |
|------|--------|------|
| `IMAGE_PUSH_REGISTRY` | `ghcr.io` | 镜像实际推送仓库 |
| `HELM_PACKAGE_IMAGE_REGISTRY` | `ghcr.registry.cdn.w7.cc` | 写入 Helm Chart 的镜像仓库域名 |

### GitHub Secrets

| 名称 | 用途 |
|------|------|
| `GITHUB_TOKEN` | 登录 GitHub Container Registry、创建 Release |
| `ZPK_DOCKER_USERNAME` | 登录 ZPK 平台 |
| `ZPK_DOCKER_PASSWORD` | 登录 ZPK 平台 |

## Dockerfile 说明

当前镜像使用多阶段构建：

1. `golang:1.26-alpine` 阶段编译 `w7panel-server`
2. `alpine:3.20` 阶段安装运行依赖
3. 将后端二进制复制到 `/app/w7panel`
4. 将 `w7panel-server/kodata/` 复制到 `/var/run/ko/`

运行时关键配置：

| 配置 | 值 |
|------|----|
| 工作目录 | `/app` |
| 入口命令 | `/app/w7panel` |
| 暴露端口 | `8000` |
| `KO_DATA_PATH` | `/var/run/ko` |
| 时区 | `Asia/Shanghai` |

## 发布前检查

发布前至少确认：

- `w7panel-ui` 能正常执行 `npm run build`
- `w7panel-server/kodata/` 已包含最新前端构建产物
- 镜像 tag 与 Helm Chart version/appVersion 一致
- `charts/*.tgz` 是本次构建生成的包
- 镜像仓库和 ZPK 凭证可用
- Git tag 使用 `v` 前缀，例如 `v1.0.0`

## 常见问题

### 前端资源没有更新

检查是否执行了资源复制：

```bash
cp -R w7panel-ui/dist/. w7panel-server/kodata/
```

### 本地测试镜像没有推送

`make build-image-test` 使用 `--load`，只加载到本地 Docker。需要推送时使用：

```bash
make build-image IMAGE_REPOSITORY=... IMAGE_TAG=...
```

### Chart 中镜像地址不符合预期

检查发布命令是否显式传入：

```bash
HELM_PACKAGE_IMAGE_REPOSITORY=...
HELM_PACKAGE_IMAGE_TAG=...
```

### GitHub Actions 找不到 Helm 包

确认 `make publish` 已执行成功，并且 `CHART_PACKAGE_DIR` 指向 `charts`。
