# API 开发文档

`docs/src/development/api/` 是 `w7panel-server/app` 业务 API 的文档入口。接口说明以业务专题为主、模块文档为辅；新增或修改接口时，先更新对应专题文档，再按需维护模块文档。

开发约定见 [conventions.md](./conventions.md)，包括路由分层、鉴权、安全、请求参数、响应格式、前后端同步和测试要求。

## 业务专题

| 文档 | 说明 |
|------|------|
| [credentials.md](./credentials.md) | API 调用凭据、用户 token、webdavToken、微应用 props、OIDC token 和 LOCAL_MOCK 行为 |
| [cluster-ops.md](./cluster-ops.md) | 集群资源、Helm、YAML、终端、集群侧代理、DNS、GPU 和诊断类接口 |
| [longhorn.md](./longhorn.md) | Longhorn 安装、卷状态、副本筛选、attach/detach、快照和文件系统操作 |
| [container-files.md](./container-files.md) | 容器文件管理、WebDAV、压缩解压、权限修改、分片上传和文件编辑器调用流程 |
| [container-images.md](./container-images.md) | Registry v2、containerd 镜像操作、容器提交镜像 |
| [microapp-static.md](./microapp-static.md) | 微应用信息、后端代理、前端静态资源状态、下载和回源代理 |
| [oauth-oidc.md](./oauth-oidc.md) | OAuth、OIDC Provider、OIDCClient CRD、Console OAuth 和微应用获取 code |
| [metrics.md](./metrics.md) | CPU、内存、磁盘使用量和 metrics 组件安装状态 |
| [k3k.md](./k3k.md) | 云主机用户、CKM/CVM、同步、菜单、套餐和订单入口 |
| [zpk.md](./zpk.md) | ZPK 配置、列表、安装、升级、构建和应用管理 |

## 模块文档

模块文档按 `w7panel-server/app` 目录拆分，适合从 Controller 和 provider 代码反查接口实现。若某个模块的接口已完整归入业务专题，则不再保留单独模块文档，避免重复维护。

| 文档 | 对应模块 | 说明 |
|------|----------|------|
| [application.md](./application.md) | `w7panel-server/app/application` | OpenAPI、验证码和公开站点接口 |
| [audit.md](./audit.md) | `w7panel-server/app/audit` | 登录日志、操作日志和审计状态 |

## 维护规则

- 新增接口优先补充业务专题文档。
- 无法归入现有专题，或需要保留 Controller 反查入口时，再补充模块文档。
- 新增专题或模块文档时，同步更新本目录。
