# K3s 1.37 本地镜像安装回归

日期：2026-11-27。隔离环境：Docker 容器 `w7panel-k3s137-build-20261127`，基础镜像 `ghcr.registry.cdn.w7.cc/w7panel/sysbox-deploy-k3s:v0.7.1-32`，K3s `v1.37.2+k3s1`。未在宿主安装 K3s。后端构建基线为 `c203f7a4998a06a94115ee8f8190ba679572230f` 加本地升级脚本改动。

## 镜像与测试资产

- `../w7panel-server/Makefile` 的 `make image` 先执行前端 `build.sh`，再以 `ko build --local` 加载镜像。前端构建成功；原项目的两个未跟踪 Go 文件引用未定义的 `InnerFilePath`，所以在不含它们的临时 Git 工作树 `/tmp/w7panel-server-build-20261127` 中执行 `make ko-build`，并复制已生成的 `kodata` 与升级脚本。构建使用 Go 1.26.6 与 `GOSUMDB=sum.golang.org` 覆盖 Makefile 默认值。
- 第二版镜像 `w7panel-k3s137-local:crd-fix2-20261127` 的 Docker ID 为 `sha256:ac49de221195935b9825464d4c7b69cabc03b44fcadf4b59825e610e9c626281`，已导入测试容器 K3s containerd。两个版本均为本地测试镜像，未推送。
- 仅在 `/tmp/w7panel-k3s137-assets-20261127` 的测试副本中给 `w7panel-offline` HelmChart 添加 `image.repository=w7panel-k3s137-local`、`image.tag=crd-fix2-20261127`、`image.pullPolicy=IfNotPresent`。仓库原安装器 Chart 仍引用 `1.1.106`。测试资产通过容器内本地 HTTPS 提供，mkcert 根 CA 只导入测试容器；安装脚本下载 SHA256 与源文件均为 `7c15a7df260a3a1ecaf2cf3a778aa843d07add03313c55b8851b9180c3bfa466`。
- 安装命令设置 `LOCAL_MOCK=true`、`K3S_SNAPSHOTTER=fuse-overlayfs`。Helm 创建的面板 Deployment 环境变量中没有 `LOCAL_MOCK`，因此只确认命令环境使用测试模式，不把 Pod 视作 mock 模式。

## 第一轮：发现后续 Gateway API 冲突

第一版镜像 `crd-fix-20261127` 包含面板 CRD 限定导入修复。安装器退出码 0、节点 Ready；升级 Job 越过 CRD 导入后，在 `kodata/shell/upgrade.sh` 对 `experimental-install.yaml` 的 server-side apply 与 `gateway-api-crd` HelmChart 管理字段冲突，Job `Failed 0/1`。该结果不算完整通过。原始日志：`/tmp/w7panel-install-20261127-round1.log`、`/tmp/w7panel-upgrade-20261127-round1.log`。

据此在 `../w7panel-server/kodata/shell/upgrade.sh` 增加检查：存在 `kube-system/gateway-api-crd` HelmChart 时，跳过重复导入实验版 Gateway API 清单。脚本 `sh -n` 与 `git diff --check` 均通过。第一轮结束后，仅在专用容器内运行官方 `k3s-uninstall.sh`，确认 K3s 二进制和数据目录已移除。

## 第二轮：核心安装通过

使用第二版镜像，从新目录 `/tmp/install-round-fix2-20261127` 完整重装；显式设置测试容器内网地址 `172.20.0.3`，避免卸载残留的 Cilium 网卡影响地址选择。安装器退出码 **0**，节点 `server1` **Ready**。Cilium、Gateway API CRD、Higress、cert-manager 和 `w7panel-offline` 的 Helm 安装 Job 均 **Complete 1/1**。官方面板升级 Job `w7panel-offline-1-upgrade-lsezm` **Complete 1/1**，日志显示跳过 HelmChart 管理的 Gateway API CRD 并执行到脚本末尾。面板 Deployment 和 Agent Pod 均 `Running 1/1`，`http://172.20.0.3:9090` 返回 **HTTP 200**。原始日志：`/tmp/w7panel-install-20261127-round2.log`、`/tmp/w7panel-upgrade-20261127-round2.log`。

验收时 cert-manager CSI Driver 达到 `Running 3/3`；`cert-manager-lego-webhook` 仍为 `ContainerCreating 0/1`，事件持续显示拉取 `ghcr.io/yxwuxuanl/cert-manager-lego-webhook:v1.6.2`。宿主 Docker 对原 GHCR 和项目代理地址拉取同一镜像也卡在相同层，未完成导入。因此核心面板安装和升级 Job 已通过，**全部长期 Pod Ready 尚未通过**。该容器随后在第三轮为释放磁盘空间而删除，报告和容器外日志保留；未发布镜像或修改正式安装器的镜像引用。

## 第三轮：全新 Docker 容器中的干净 K3s 集群

用户要求从未安装过 K3s 的集群验收。新建专用容器 `w7panel-k3s137-clean-20261127`（`172.20.0.4`），启动前确认 `/usr/local/bin/k3s` 与 `/var/lib/rancher/k3s` 均不存在，且与宿主的 mount namespace 不同。仅在该容器执行 `mount --make-rshared /`。HTTPS 安装器下载通过本轮 CA 校验，下载和发布副本 SHA256 同为 `7c15a7df260a3a1ecaf2cf3a778aa843d07add03313c55b8851b9180c3bfa466`。从 `/tmp/install-clean-20261127` 执行安装器，环境为 `LOCAL_MOCK=true`、`K3S_SNAPSHOTTER=fuse-overlayfs`、`INTERNAL_IP=PUBLIC_IP=172.20.0.4`；同一个本地镜像 tar 预置于该工作目录的 `w7panel/images/`，由安装器在提交 HelmChart 前自动导入。面板 Pod 本身没有 `LOCAL_MOCK` 环境变量。

**原样 installer 首次执行未通过：退出码 1。** K3s `v1.37.2+k3s1` 节点曾达到 Ready，但宿主同时运行上一轮测试容器，磁盘使用率升至约 93%，两套 K3s 的 CPU 和磁盘 I/O 负载导致 API 多次超时。installer 的 10 分钟面板就绪等待超时。Kubelet 后续触发 `DiskPressure`、驱逐 Pod，并按安装器设置的 `image-gc-high-threshold=70` / `low-threshold=60` 删除了刚导入但尚未被 Pod 使用的本地镜像；面板和升级 Job 随后尝试从 Docker Hub 镜像代理拉取该测试标签，返回 403。原始安装日志及退出码保存在容器 `/tmp/w7panel-install-clean-20261127.log` 和 `.exit`。

为继续验证 Chart，停止并删除了上一轮专用测试容器（其报告和容器外日志保留），释放约 9.8 GB，宿主剩余空间回升至约 16 GB。在**新测试容器内**将 K3s kubelet 镜像 GC 阈值临时调为 95/90，重启 K3s，重新导入本地镜像；这属于测试环境修正，不是仓库 installer 的原样行为。节点 `DiskPressure=False`，镜像标签留存。`helm-install-w7panel-offline` 最终 `Complete 1/1`，release revision 1 为 `deployed`。首次面板升级 Job `w7panel-offline-1-upgrade-zydzr` 仍因当时 cert-manager webhook 没有端点而 `Failed 0/1`；其日志显示 CRD `--force-conflicts` 已执行，失败点是创建 `webhook-ca.yaml` 时 `no endpoints available for service cert-manager-webhook`。

待 cert-manager webhook Ready 后，用同一 `w7panel-offline-1.1.106.tgz` 执行一次 `helm upgrade --reuse-values`（revision 2）。新官方升级 Job `w7panel-offline-2-upgrade-hdgid` **Complete 1/1，耗时 19 秒**；日志显示面板 CRD 应用成功、跳过 HelmChart 管理的 Gateway API CRD，并执行到脚本末尾。面板及 Agent 均 `Running 1/1`，Cilium Agent 和 Operator 均 `Running 1/1`，节点 Ready，容器内访问 `http://172.20.0.4:9090` 为 **HTTP 200**。cert-manager CSI Driver 最终 `Running 3/3`。`cert-manager-lego-webhook` 仍在拉取 `ghcr.io/yxwuxuanl/cert-manager-lego-webhook:v1.6.2`，未达到 Ready；早期被驱逐的旧 Pod 曾短暂保留 `Error` 状态。

结论：**干净 K3s 起点上的 Chart 安装与修复版升级 Job 最终通过，但不是零干预的 installer 安装通过。** 首次安装器超时、磁盘压力和首次升级 Job 失败都必须保留为本轮结果。当前新测试容器保留供复查；正式 installer 仍引用已发布的 `1.1.106` 面板镜像，未包含本地测试修复。
