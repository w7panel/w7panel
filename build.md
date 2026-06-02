# 构建镜像说明

## 手动构建

1. 将 `./w7panel-ui` 构建产物复制到 `./w7panel-server/kodata` 目录
2. 进入 `./w7panel-server` 目录
3. 使用 `ko` 构建并推送镜像

```bash
export KO_DOCKER_REPO=ccr.ccs.tencentyun.com/afan-public/w7panel1
export KO_DEFAULTBASEIMAGE=ccr.ccs.tencentyun.com/afan-public/ubuntu:24.04-offlineui
/root/go/bin/ko build --bare --tags=${CNB_BRANCH} --tag-only --sbom=none --platform=all
```

## GitHub Actions 自动构建

仓库已添加工作流：

`/.github/workflows/build-image-on-tag.yml`

触发方式：

```bash
git tag v1.0.0
git push origin v1.0.0
```

工作流执行顺序：

1. 检出主仓库和子模块
2. 构建 `w7panel-ui`
3. 将前端产物复制到 `w7panel-server/kodata`
4. 在 `w7panel-server` 目录执行 `ko build`
5. 使用 Git tag 名称作为镜像 tag 推送

需要提前配置的 GitHub Secrets：

- `TENCENT_REGISTRY_USERNAME`
- `TENCENT_REGISTRY_PASSWORD`

默认镜像仓库和基础镜像：

- `KO_DOCKER_REPO=ccr.ccs.tencentyun.com/afan-public/w7panel1`
- `KO_DEFAULTBASEIMAGE=ccr.ccs.tencentyun.com/afan-public/ubuntu:24.04-offlineui`
