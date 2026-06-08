# w7panel 部署指南

## 启动方式（推荐使用 w7panel-ctl.sh）

启动脚本会自动检测并设置正确的环境变量：

```bash
# ========== 开发模式 (需要 kubeconfig.yaml) ==========
export KUBECONFIG=/path/to/kubeconfig.yaml
./w7panel-ctl.sh start

# ========== 生产模式 (使用 ServiceAccount) ==========
export LOCAL_MOCK=false
./w7panel-ctl.sh start
```

## 环境变量说明

| 变量名 | 开发模式 | 生产模式 | 说明 |
|--------|---------|---------|------|
| `CAPTCHA_ENABLED` | false | false | 验证码开关 |
| `LOCAL_MOCK` | true | false | K8s 访问方式 |
| `KO_DATA_PATH` | ./kodata | ./kodata | 静态资源目录 |
| `KUBECONFIG` | 必填 | 不需要 | kubeconfig 文件路径 |

## 常见错误

### 权限错误

**错误信息**：
```
serviceaccounts "admin" is forbidden: User "system:serviceaccount:default:default" 
cannot get resource "serviceaccounts" in API group "" in the namespace "default"
```

**原因**：未正确设置模式，系统使用了 ServiceAccount 而非 kubeconfig。

**解决**：
- 开发模式：设置 `KUBECONFIG` 环境变量
- 生产模式：确认 ServiceAccount 权限配置正确

## 相关文档

- [部署与运维排障](./troubleshooting.md)

## GitHub Actions 发布流程

仓库支持通过 Git tag 或手动触发 Release workflow，自动构建文档镜像、前端、主服务镜像和 Helm Chart。

工作流文件：

- `/.github/workflows/release.yml`

触发方式：

```bash
git tag v1.0.0
git push origin v1.0.0
```

也可以在 GitHub Actions 页面手动运行 `Release` workflow，并填写 `tag`。

功能分类：

- 文档服务：按根路径 `/` 构建 `docs/docs` VitePress 文档项目，并将 `docs/docs/.vitepress/dist` 打包到 nginx 默认静态目录 `/usr/share/nginx/html/`
- 前端资源：构建 `w7panel-ui`，并同步产物到 `w7panel-server/kodata/`
- 主服务发布：构建并推送 w7panel 主服务镜像，打包 Helm Chart
- Release 资产：将 Helm Chart 上传到 GitHub Release assets
- ZPK 分发：将 Helm Chart 作为 ZPK 附件推送到 `ZPK_ARTIFACT`

执行流程：

1. 解析发布 tag：tag push 使用 `GITHUB_REF_NAME`，手动触发使用输入的 `tag`
2. 构建文档：执行 `docs/docs` 的 pnpm 构建，固定 `DOCS_BASE_PATH=/`
3. 推送文档镜像：登录 `IMAGE_PUSH_REGISTRY`，将静态文件复制到 nginx 默认目录，并推送 `${IMAGE_PUSH_REGISTRY}/${owner/repo}-docs:${tag}`
4. 构建前端：执行 `w7panel-ui` 的 npm 构建
5. 同步前端资源：复制 `w7panel-ui/dist` 到 `w7panel-server/kodata/`
6. 推送主服务镜像并打包 Chart：登录 `IMAGE_PUSH_REGISTRY`，执行 `make publish`
7. 发布 Helm 包：上传 `charts/*.tgz` 到 GitHub Release assets
8. 推送 ZPK：登录 `ZPK_HOST`，将 Helm 包附加到 `ZPK_ARTIFACT` 后推送

前置配置：

- 配置 GitHub Secrets `ZPK_DOCKER_USERNAME`
- 配置 GitHub Secrets `ZPK_DOCKER_PASSWORD`
- 如使用非默认镜像仓库，可配置仓库变量 `IMAGE_PUSH_REGISTRY`
- 如使用非默认 Helm 包镜像仓库，可配置仓库变量 `HELM_PACKAGE_IMAGE_REGISTRY`

默认构建参数：

- `IMAGE_PUSH_REGISTRY=ghcr.io`
- `HELM_PACKAGE_IMAGE_REGISTRY=ghcr.registry.cdn.w7.cc`
- `DOCS_BASE_PATH=/`
- 文档镜像：`${IMAGE_PUSH_REGISTRY}/${owner/repo}-docs:${tag}`，默认是 `ghcr.io/w7panel/w7panel-docs:${tag}`
