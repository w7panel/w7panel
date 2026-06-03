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

仓库支持通过 Git tag 或手动触发 Release workflow，自动构建前端、文档、镜像和 Helm Chart，并发布文档到 GitHub Pages。

工作流文件：

- `/.github/workflows/release.yml`

触发方式：

```bash
git tag v1.0.0
git push origin v1.0.0
```

也可以在 GitHub Actions 页面手动运行 `Release` workflow，并填写 `tag`。

主要流程：

- 构建 `w7panel-ui`，并同步产物到 `w7panel-server/kodata/`
- 构建 `docs/docs` VitePress 文档项目
- 打包并推送镜像、Helm Chart 和 ZPK 附件
- 将 `docs/docs/.vitepress/dist` 发布到 GitHub Pages

前置配置：

- GitHub Pages Source 需要设置为 GitHub Actions
- 配置 GitHub Secrets `ZPK_DOCKER_USERNAME`
- 配置 GitHub Secrets `ZPK_DOCKER_PASSWORD`
- 如使用非默认镜像仓库，可配置仓库变量 `IMAGE_PUSH_REGISTRY`
- 如使用非默认 Helm 包镜像仓库，可配置仓库变量 `HELM_PACKAGE_IMAGE_REGISTRY`
- 如 GitHub Pages 发布到域名根路径或自定义路径，可配置仓库变量 `DOCS_BASE_PATH`

默认构建参数：

- `IMAGE_PUSH_REGISTRY=ghcr.io`
- `HELM_PACKAGE_IMAGE_REGISTRY=ghcr.registry.cdn.w7.cc`
- `DOCS_BASE_PATH=/${仓库名}/`
