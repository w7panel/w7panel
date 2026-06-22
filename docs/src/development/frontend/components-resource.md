# 资源、账号和权限组件

| 组件 | 源文件 | Props | Events | 说明 |
|------|--------|-------|--------|------|
| `ContainerPlugin` | `src/components/container-plugin.vue` | `propsData` | - | 容器插件配置 |
| `QuotaConfig` | `src/components/quota-config.vue` | `data` | `setQuotaPage` | 配额入口配置 |
| `QuotaEdit` | `src/components/quota-edit.vue` | `show`, `data`, `name`, `clustermode` | `submit`, `close` | 配额编辑 |
| `PermissionEdit` | `src/components/permission-edit.vue` | `show`, `list`, `name`, `type`, `debug`, `fileeditor`, `whitelist`, `webshell`, `permissionPackage`, `disabledBase`, `disabledMenu`, `noCustom` | `submit`, `close` | 权限编辑 |

## 维护要求

- 修改组件 Props、Events、Ref 方法或对外行为时，同步更新本文。
- 涉及 Wujie 事件的组件，只在本文记录组件职责，事件协议维护到 [wujie-events.md](./wujie-events.md)。
- 返回 [组件说明入口](./components.md)。
