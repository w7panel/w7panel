# W7Panel 1.x

W7Panel 是基于 Kubernetes 的云原生应用管理平台，提供应用部署、集群资源、文件管理、域名证书、存储和用户权限等能力。

## 环境要求

- 推荐使用全新的 Linux 服务器环境
- 节点配置建议不低于 2 核 4G
- 需要可访问 `6443`、`80`、`443`、`9090` 等端口
- 浏览器建议使用 Chrome、Firefox、Edge 等现代浏览器

## 快速安装

```shell:no-line-numbers
curl -sfL https://cdn.w7.cc/w7panel/install.sh | sh -
```

安装完成后，通过浏览器访问：

```text:no-line-numbers
http://{服务器 IP}:9090
```

首次进入后台后，根据页面提示设置管理员账号密码。

## 核心能力

- 可视化管理 Kubernetes 集群、节点和资源对象
- 支持应用商店、Helm、YAML、Docker Compose 等多种部署方式
- 支持在线文件管理、WebDAV、压缩解压和权限修改
- 支持域名绑定、反向代理和 Let's Encrypt HTTPS 证书
- 支持 Longhorn、NFS、Local 等多种存储方案
- 提供用户、用户组、权限策略和资源配额管理

## 开发模式

本地或测试环境默认使用测试模式：

```shell:no-line-numbers
export LOCAL_MOCK=true
export CAPTCHA_ENABLED=false
export KUBECONFIG=/path/to/kubeconfig.yaml
```

开发模式启动建议通过项目启动脚本执行，避免环境变量未传递给子进程。

## 文档入口

- [项目文档中心](https://github.com/w7panel/w7panel/blob/dev-v1/docs/README.md)
- [用户指南](https://github.com/w7panel/w7panel/tree/dev-v1/docs/user-guide)
- [部署文档](https://github.com/w7panel/w7panel/tree/dev-v1/docs/deployment)
- [开发指南](https://github.com/w7panel/w7panel/tree/dev-v1/docs/development)
- [API 文档](https://github.com/w7panel/w7panel/tree/dev-v1/docs/api)
- [测试文档](https://github.com/w7panel/w7panel/tree/dev-v1/docs/testing)
