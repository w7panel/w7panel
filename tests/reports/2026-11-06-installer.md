# 本地 HTTPS 安装回归（2026-11-06）

## 环境

- 独立 Docker 测试容器：`w7panel-installer-k3s-test`，CentOS Stream 9 / systemd。
- mkcert 签发 `w7panel.com`，测试 CA 仅导入测试容器；容器内该域名解析到 `127.0.0.1`。
- Python HTTPS 服务发布本地 `/installer/`；curl 无 `-k` 下载成功，SHA256 与源脚本一致。
- K3s：`v1.36.5+k3s1`；第二轮采用 `native`、最终重装采用 `fuse-overlayfs` snapshotter，并准备 shared 挂载传播。
- 安装命令携带 `LOCAL_MOCK=true`；此变量不会自动注入 Helm 部署中的面板 Pod。

## 已发现并修复

- 镜像 tar、systemd env 文件和 Chart URL 引用缺失资源。
- 面板 Chart 包缺失，manifest 版本与本地 Chart 不一致。
- 启动阶段 Helm Pod 直接下载本地域名 Chart 会依赖 DNS 和私有 CA，改为 K3s 静态目录。
- 下载跳过 TLS 验证、空文件被视为缓存、失败管道未传播、下载重试时间过长。
- 非交互终端进度动画导致密集轮询和短进程创建，改为两秒间隔并设置十分钟超时。
- 旧 registry 域名不可解析、阿里云普通仓库被当作通用 Docker Hub 镜像源。
- `K3S_SNAPSHOTTER` 需要显式转换为 K3s 的 `--snapshotter` 参数。

## 静态验证

- `bash -n installer/install.sh`：通过。
- `python3 tests/check-installer-assets.py`：四个 Chart 的路径及版本全部通过。
- `helm lint charts/w7panel`：通过。
- 测试容器通过受信任 HTTPS 获取四个 Chart，并核对包内版本：通过。

## 实机回归

- 第一轮：K3s 下载和安装成功，启动受嵌套 overlayfs 限制；执行官方卸载后确认二进制不存在。
- 第二轮：K3s API 启动，Cilium Helm 安装完成、Operator 运行；修正测试容器挂载传播后继续验证。

- 第二轮最终节点 Ready，Cilium Agent/Operator 正常，Higress 和面板 Helm 安装完成；业务镜像拉取无进展，因此没有判定端到端通过。
- 第二次执行官方卸载后，确认 `/usr/local/bin/k3s`、`/var/lib/rancher/k3s` 均不存在；最终重装未使用导出的镜像缓存。


## 最终结果：通过

第四轮使用最终安装脚本和本地打包的面板 Chart 1.1.102，先执行官方卸载并确认
K3s 二进制、数据目录不存在，再从受信任 HTTPS 下载脚本重新安装。

- 安装进程退出码 **0**，日志打印安装成功。
- 下载脚本与发布目录 SHA256 均为 `8b57f4710ce80134d05cfe2dd57acbd3fa2b4ccf19baf6d90db15a8545eb2ace`。
- `server1`：**Ready**，K3s `v1.36.5+k3s1`。
- `http://172.20.0.3:9090`：**HTTP 200**。
- `w7panel-offline-1-upgrade-pzgme`：**Complete 1/1**，耗时 **2m41s**。
- Cilium、Operator、CoreDNS、metrics-server、local-path-provisioner、cert-manager、
  cert-manager CSI、Higress、面板与 Agent 均已 Running/Ready。
- 初始化早期曾因 cert-manager Webhook 尚未启动而重试，最终 Job 成功，失败尝试已由初始化脚本清理。
- 全程安装均在 Docker 内完成；宿主机仅只读检查，没有安装 K3s。

最终还修复了默认路由 IP 选择（残留 cilium_host 会干扰首网卡法），并在面板
初始化 Job 中保留外部 Helm 组件管理的 CRD。此检查兼容没有 release 注解、
只有 `managedFields.manager=helm` 的 CRD。Dockerfile 补齐 jq 依赖；使用
Alpine 3.20 实际安装 jq 并执行过滤表达式，结果通过。

测试容器 `w7panel-installer-k3s-test` 保留供复查。关键证据位于容器：

- `/tmp/install-verified.log`：最终安装日志。
- `/tmp/uninstall-final.log`：最终一轮重装前的官方卸载日志。
- `/tmp/upgrade-fixed.log`：独立验证 CRD 冲突修复的初始化日志。

证书位于开发工作区宿主可见的 `/tmp/w7panel-installer-test/certs/`，未提交至仓库。
复用流程见项目 Skill：`skills/w7panel-installer-test/SKILL.md`。
