# 宿主安装回归（2026-11-06）

用户明确要求在宿主安装一次。宿主为 `debian-test-1`（Debian 13），systemd PID 1，
安装前 K3s/k3s-agent 均 inactive，80/443/6443/9090 未被占用，根分区可用约 23 GiB。
此前 Docker 测试容器处于停止状态，本轮没有启动它。

## 测试方式

- 新建 `/tmp/w7panel-host-test-20261106`，独立 mkcert CA 与叶子证书；证书位于文档根之外。
- 根 CA 安装到 `/usr/local/share/ca-certificates/w7panel-host-test-20261106.crt`。
- `/etc/hosts` 添加带标记的 `127.0.0.1 w7panel.com`，原文件备份为 `hosts.before`。
- `w7panel-host-test-https.service` 仅在 `127.0.0.1:443` 发布本地 installer。
- `w7panel-host-install-test.service` 在全新 `round-1` 目录执行下载脚本，记录 `install.log` 和 `install.exit`。
- 使用默认 `overlayfs`，未修改宿主根挂载传播。安装命令携带 `LOCAL_MOCK=true`，不会自动注入 Helm 面板 Pod。
- `PUBLIC_IP=172.16.1.123` 用于跳过公网 IP 查询，本轮只验证内网地址，不代表公网可达。
- 安装脚本与发布资源 SHA256：`8b57f4710ce80134d05cfe2dd57acbd3fa2b4ccf19baf6d90db15a8545eb2ace`。
- Shell 语法与四个 Chart 路径/版本检查通过。

## 结果

- 安装脚本退出码 `0`，输出安装成功；K3s `v1.36.5+k3s1` 节点 `server1` Ready。
- 默认 snapshotter 为 `overlayfs`，无须 Docker 测试中的 fuse-overlayfs 或挂载传播调整。
- 四个 Helm 安装 Job Complete；初始化 Job `w7panel-offline-1-upgrade-pwzfx` Complete `1/1`，用时 `3m51s`。
- 面板及 Agent Running/Ready；`http://172.16.1.123:9090` 返回 HTTP 200。
- 最终 Cilium/Operator、CoreDNS、metrics、local-path、cert-manager/CSI、Higress、面板及 Agent 全部长期 Pod Running/Ready；CSI 为 3/3，启动期有一次重启。宿主安装验收通过。

## 启动期问题与处理

1. 宿主残留 `/etc/cni/net.d/05-cilium.conflist`（时间为 2026-03-10），因此本轮不是完全空白系统。Cilium 镜像下载期间普通 Pod 报 `cilium.sock` 不存在；新 Agent 启动后自动恢复，没有删除旧配置。
2. cert-manager webhook 镜像下载约 3 分钟，初始化任务前四次尝试因 webhook 无 endpoints 失败，第五次自动成功。未手动重建 Job，未再次修改代码。
3. Higress Gateway 曾等待 `higress-ca-root-cert`，Controller 就绪后自动创建，Gateway 随后 Ready。
4. 安装脚本在面板 HTTP 可用时已返回成功，早于完整初始化及其余组件就绪；测试验收额外等待 Job Complete 和全部长期 Pod Ready。

## 保留状态与范围

宿主 K3s 和面板保留运行，本轮未卸载。安装器测试 HTTPS unit、带标记 hosts 行及测试 CA 保留供复查。
原始日志及 Job JSON 位于 `/tmp/w7panel-host-test-20261106/`，包含 `install.log`、`install.exit`、`upgrade-job.json`、`upgrade-job.log`。
面板 Deployment 未设置 `LOCAL_MOCK`，不能把安装命令携带该变量等同于面板 mock 部署。
本轮验证内网安装及启动，不覆盖公网访问、登录后的全部业务或新版面板镜像完整构建。

## 提交时资源差异

提交前工作区 manifest 已指向 `installer/charts/w7panel-offline/w7panel-offline-1.1.106.tgz`。
本报告的宿主实测使用 `1.1.102`，不能视为 `1.1.106` 已完成安装回归。
`1.1.106` 资源路径与包版本检查通过，但其内置 upgrade-job 未包含当前
`charts/w7panel/templates/upgrade-job.yaml` 的 CRD 保护逻辑；提交保留该包原内容。
