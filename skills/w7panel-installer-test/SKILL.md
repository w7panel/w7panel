---
name: w7panel-installer-test
description: 使用 mkcert 为 w7panel.com 启动本地 HTTPS，在独立 systemd Docker 中回归 W7Panel installer 的 K3s 安装、故障修复、官方卸载和重新安装，并更新测试证据。支持用户明确要求的宿主安装验收，适用于本仓库安装器联调。
---

# W7Panel 安装器回归

按 [回归操作与故障处理](references/docker-regression.md) 准备环境、执行和取证；本轮历史证据在仓库 `tests/reports/2026-11-06-installer.md`。历史通过不能替代当前版本回归。

- 优先独立、可丢弃的 privileged systemd Docker，不挂载宿主根目录，不使用宿主网络。镜像先确认 `/sbin/init` 存在。
- 默认优先 Docker；用户明确要求宿主测试，或 Docker 确实受环境限制且用户已授权时，可转宿主。先确认宿主无业务集群，已有授权不重复询问；无法确认现有集群用途时，不覆盖安装或卸载。宿主准备见参考文档的“宿主测试”一节。
- 安装器会修改防火墙、网络、systemd 与 K3s 状态；卸载仅作用于本轮专用环境。`mount --make-rshared /` 仅在测试容器的独立挂载命名空间执行。
- HTTPS 叶子证书私钥放在文档根之外；mkcert 私钥和根 CA 私钥不入库；只向测试容器导入根证书，保留 HTTPS 证书验证，不用 `curl -k`。
- 默认测试命令携带 `LOCAL_MOCK=true`，但此变量不会自动进入 Helm 创建的面板 Pod。分别记录命令环境和 Pod 实际环境，不将两者等同；需要验证 mock 面板时明确配置并核查 Pod 环境。
- 等待安装、导出等进程真正完成后再执行依赖操作。工具返回 `session_id` 表示进程尚未结束，不能立即卸载。
- 验收同时覆盖脚本退出码、K3s Ready、网络组件、Helm 安装、面板响应与初始化 Job Complete。HTTP 200 单独不能判定通过。
- 修复后更新本地 Chart 包及版本/路径引用，并重新同步 HTTPS 发布资产、核对哈希。执行官方卸载后，从全新工作目录再次获取安装器重装，避免复用旧 `w7panel/` 缓存；旧目录保留取证，区分在线安装与镜像缓存测试。
- 测试结束更新报告与本 Skill 的已验证事实、故障处理和限制；未完成、超时、手动补救必须明确记录。
