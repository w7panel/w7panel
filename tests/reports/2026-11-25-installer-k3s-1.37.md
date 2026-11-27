# K3s 1.37 安装器与 w7panel-offline 升级 Job 回归

日期：2026-11-25。测试环境：独立 CentOS Stream 9 systemd Docker 容器，镜像
`ghcr.registry.cdn.w7.cc/w7panel/sysbox-deploy-k3s:v0.7.1-32`；未在宿主安装 K3s。
根仓库 `9605bf8280b995c49fec9007570b217fa72989d7`，后端 `dev-v1`
`03e0c397cd2c8fb14e6f9408e2390cccf9a82a53`。本地未跟踪文件
`w7panel-offline-1.1.106.tgz` 为测试前已有文件，未修改。

## 安装资产与环境

- 安装器从当前仓库 `installer/` 经专用容器的 mkcert HTTPS 服务发布，`https://w7panel.com/installer/install.sh` 下载内容与仓库 SHA256 均为 `7c15a7df260a3a1ecaf2cf3a778aa843d07add03313c55b8851b9180c3bfa466`。根 CA 只导入测试容器，未关闭 TLS 验证，证书私钥在发布目录外，私钥 URL 返回 404。
- 安装器的 K3s stable 通道解析为 `v1.37.2-k3s1`，节点报告 `v1.37.2+k3s1`。命令带 `LOCAL_MOCK=true`，未验证 Helm 面板 Pod 继承此变量。
- 测试机初始根分区仅余 1.4 GB。清理未使用 Docker 构建缓存 2.023 GB、超过 7 天且无容器引用的旧镜像 5.564 GB 后重测。专用容器中把 kubelet 磁盘驱逐阈值调为 1%；这是测试环境修正，非安装器默认配置。
- 第一轮把 K3s 数据置于 2 GB、后扩到 3.5 GB 的 tmpfs，面板镜像解包报 `no space left on device`，安装器 10 分钟超时，退出码 1。第一轮日志保存在 `/tmp/w7panel-installer-137-round1.log`；随后停止并移除第一轮专用容器。

## 第二轮完整安装结果

第二轮容器名：`w7panel-k3s-137-disk-20261125`，K3s 数据使用容器磁盘。
执行 `LOCAL_MOCK=true K3S_SNAPSHOTTER=fuse-overlayfs bash /tmp/install.sh`，
工作目录 `/tmp/install-round-137-disk`。脚本退出码 **1**，最后错误为
`等待面板就绪超时`；日志保存在 `/tmp/w7panel-installer-137-round2.log`。

- `server1` 节点 Ready；Cilium Agent/Operator Running/Ready。
- Cilium、Gateway API CRD、Higress、cert-manager 和 `w7panel-offline` 的 Helm 安装 Job 均 `Complete 1/1`。Helm 初次启动时 cert-manager webhook 尚未就绪，面板 Helm 曾自动重试，最终 Helm Job 完成。
- 官方面板升级 Job `w7panel-offline-1-upgrade-holwr` 为 **Failed 0/1**，两次 Pod 均 Error。面板镜像为 `ccr.ccs.tencentyun.com/afan-public/w7panel1:1.1.106`，尚未包含本次脚本修复。
- 官方升级日志在 `kubectl apply -f /var/run/ko/crds --server-side` 后报两次 `conflict with "helm": .spec.versions`，并明确建议 `--force-conflicts`。日志保存在 `/tmp/w7panel-installer-137-conflict.log`。对仓库旧 `higress2.0.4.yaml` 执行 server dry-run 可重现两次冲突；该文件含 Higress Chart 管理的 CRD。
- 面板 Pod 与 Agent 均达到 `Running 1/1`，容器内 `http://172.20.0.3:9090` 返回 HTTP 200。升级 Job 失败且安装脚本超时，HTTP 200 不能改变整轮安装 **未通过** 的结论。

## 修复验证与限制

后端 `kodata/shell/upgrade.sh` 与子集群 `k3k-agent-upgrade.sh` 现只对 `w7panel.w7.com_*.yaml` 和
`microapp.w7.cc_*.yaml` 使用 `--server-side --force-conflicts`，不再由面板升级脚本覆盖旧 Higress CRD。K3s 1.37 API 上对这 **21 个面板 CRD** 的 server dry-run 和实际 apply 均成功。

另用相同 `1.1.106` 镜像、相同 ServiceAccount 和环境变量，仅用 ConfigMap 覆盖脚本 CRD 导入段，运行独立 Job `w7panel-upgrade-137-patched`。它成功越过 CRD 导入和迁移，完成 cert-manager CSI Driver 的 Helm 安装；随后等待 `ghcr.io/yxwuxuanl/cert-manager-lego-webhook:v1.6.2` 镜像拉取，首次 `helm --wait --timeout 5m` 报 `context deadline exceeded`，Job 进入第二次尝试。此独立 Job 的运行不等于官方安装器通过。

第二次尝试同样在该 webhook Helm 等待阶段超时。验证 Job 最终为 **Failed 0/1**，条件 `BackoffLimitExceeded`；这表明 CRD 冲突已经越过，但整条升级流程仍受该镜像拉取阻断。该验证没有证明修订版升级 Job Complete。
两次日志保存在宿主 `/tmp/w7panel-installer-137-patched-first.log` 和
`/tmp/w7panel-installer-137-patched-second.log`，最终 Job YAML 保存在
`/tmp/w7panel-installer-137-patched-job.yaml`。第二轮专用容器保留运行，供复查；
它使用约 5 GB 宿主磁盘，清理时只操作该专用容器。

要使官方安装器采用修复，还需构建并发布包含新 `upgrade.sh` 的面板镜像，更新 `installer/manifests/w7panel-offline.yaml` 引用及 Chart 包，再从全新工作目录做完整安装、卸载重装回归。当前未发布镜像或 Chart。
