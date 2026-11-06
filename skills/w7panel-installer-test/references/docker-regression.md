# Docker 安装、卸载与重装回归

## 1. 前置检查

按项目 AGENTS.md 同步相关仓库，保留已有改动。记录 Git 提交、安装器 SHA256、Chart 版本、容器镜像和 K3s 版本。检查 Docker、mkcert、Helm、Python 3、磁盘空间及 `/dev/fuse`。

镜像必须包含 systemd `/sbin/init`。本轮尝试的 `build-ubuntu-jammy` 不包含该文件；实际运行使用 `ghcr.registry.cdn.w7.cc/w7panel/sysbox-deploy-k3s:v0.7.1-32`（CentOS Stream 9），补装 `iproute sudo fuse-overlayfs`。不要仅凭镜像名推断能启动 systemd。

以下为环境准备示例，容器名须与现有容器不冲突：

```bash
test_container=w7panel-installer-k3s-test
test_image=ghcr.registry.cdn.w7.cc/w7panel/sysbox-deploy-k3s:v0.7.1-32
docker run --rm --entrypoint /bin/sh "$test_image" -c 'test -x /sbin/init'
docker run -d --name "$test_container" --privileged --cgroupns=private \
  --entrypoint /sbin/init --tmpfs /run --tmpfs /run/lock "$test_image"
docker exec "$test_container" dnf install -y iproute sudo fuse-overlayfs
docker exec "$test_container" systemctl is-system-running
```

`degraded` 时检查失败单元，确认 systemd 和本次必要服务可工作后再继续。确认容器与宿主 mount namespace 不同，再在容器内运行 `mount --make-rshared /`。不要把宿主 `/` 设为 shared 来迁就测试。嵌套 overlayfs 失败优先 `K3S_SNAPSHOTTER=fuse-overlayfs`；`native` 曾启动成功但耗时、占空间，不作为首选。

## 2. mkcert 与本地 HTTPS

使用仓库外的私有临时目录，并设置权限；将 `CAROOT` 指到本轮目录，避免改动已有 CA。运行 `mkcert -cert-file ... -key-file ... w7panel.com` 签发证书，不需要在宿主执行 `mkcert -install`。

在容器内创建 `/srv/installer` 和文档根之外的 `/run/w7panel-test-certs`；安装资源复制到前者，签发的服务证书/私钥复制到后者。证书目录限制访问权限，不通过 HTTPS 发布。仅复制 `rootCA.pem` 到 `/etc/pki/ca-trust/source/anchors/w7panel-test.crt`，执行 `update-ca-trust`；不得复制 CA 私钥。容器 `/etc/hosts` 添加 `127.0.0.1 w7panel.com`。

HTTPS 服务采用本轮实际运行的 Python 结构，在容器内启动并保留 PID/日志：

```python
import http.server
import os
import ssl

os.chdir('/srv')
server = http.server.ThreadingHTTPServer(
    ('0.0.0.0', 443), http.server.SimpleHTTPRequestHandler)
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain('/run/w7panel-test-certs/w7panel.com.pem',
                    '/run/w7panel-test-certs/w7panel.com-key.pem')
server.socket = ctx.wrap_socket(server.socket, server_side=True)
server.serve_forever()
```

只发布专用 `/srv` 内容，所有证书私钥必须位于文档根之外。测试结束按用户意图保留环境供复查或清理，仅清理本轮创建的资源。本轮真实环境已迁移证书到 `/run/w7panel-test-certs`，原私钥 URL 返回 404，安装器 HTTPS 下载仍正常。

用容器内 `curl -fS https://w7panel.com/installer/install.sh -o /tmp/install.sh` 验证信任链，并比对下载内容和仓库源文件 SHA256。四个 Chart 也要从该 URL 下载并核对包内版本。不要以跳过 TLS 验证解决 CA 问题。

## 3. 静态检查与执行

在仓库根目录执行：

```bash
bash -n installer/install.sh
python3 tests/check-installer-assets.py
helm lint charts/w7panel
```

修改 Chart 后重新 `helm package`，同步 manifest 的 URL、版本和包内版本；本轮维持 manifest 目标版本 `1.1.102`，执行 `helm package charts/w7panel --version 1.1.102 --app-version 1.1.102 --destination installer/charts`；后续以发布目标为准，不能直接使用源 Chart 的旧版本。安装器先把 Chart 下载到 K3s 静态目录，再使用 `https://%{KUBERNETES_API}%/static/charts/...`，避免 Helm Pod 依赖测试域名 DNS 和 mkcert CA。容器 `/etc/hosts` 与信任库不会自动传入 Pod。

每次修复源码或重新打包 Chart 后，重新同步 HTTPS 发布目录 `/srv/installer`，再通过 HTTPS 下载并校验安装器与 Chart 的 SHA256、包内版本。仅修改仓库文件不会自动更新已复制的发布资产。

每轮创建全新工作目录，例如 `/tmp/install-round-1`、`/tmp/install-round-2`，从该目录执行下载好的脚本并保存日志。安装器 `download_files` 会跳过非空文件；官方卸载不会删除工作目录中的 `w7panel/` 缓存，同目录重装可能误用旧 manifest 或 Chart。保留旧工作目录作为证据，不默认删除。

在测试容器的新工作目录执行：

```bash
mkdir /tmp/install-round-1
cd /tmp/install-round-1
LOCAL_MOCK=true K3S_SNAPSHOTTER=fuse-overlayfs bash /tmp/install.sh
```

运行期间检查 `journalctl -u k3s`、`k3s kubectl get nodes -o wide`、`k3s kubectl get pods,jobs -A`、`k3s kubectl get helmcharts -A`。安装器非 TTY 等待循环应每两秒检查一次，十分钟超时；不要用 spinner 当作实际等待。

验收必须保存：脚本退出码；节点 Ready；Cilium Agent/Operator 状态；Helm 安装结果；面板 HTTP 状态；初始化/升级 Job 的 Complete 条件与日志。Job 有短 TTL，及时取证。单独手动创建的修复验证 Job 只证明修复路径，需要最终完整重装确认安装器集成结果。

## 4. 官方卸载、重新安装

先确认安装或镜像缓存导出完全结束，再在测试容器运行 `/usr/local/bin/k3s-uninstall.sh`。等待其退出，检查 `/usr/local/bin/k3s` 和 `/var/lib/rancher/k3s` 已不存在。保存卸载证据后，确认最新源码与 Chart 已同步到 `/srv/installer`，重新下载本地 HTTPS 脚本、比对资产 SHA256，并切换到新的工作目录再执行安装与全部验收。原工作目录保留取证。本轮实际使用过 `/tmp/install-test`、`/tmp/install-second`、`/tmp/install-final`、`/tmp/install-verified`。

官方卸载可能保留 Cilium 虚拟网卡，第二次安装不能用 `hostname -I` 第一项作为访问地址。使用 `ip -4 route get 1.1.1.1` 的 `src`；必要时显式指定 `INTERNAL_IP`。此变量用于安装器 HTTP 访问地址和证书 SAN，不替代 K3s 的 `--node-ip`。保留网卡的同容器重装有助于验证这类回归；若需要完全干净环境另建容器，并说明两者区别。

## 5. 已遇问题与处理

| 现象 | 原因与处理 |
|---|---|
| `/sbin/init` 不存在 | 构建工具镜像不是 systemd 运行环境；启动前检查，换用已验证镜像。 |
| K3s overlayfs 启动失败 | 嵌套文件系统限制；安装 fuse-overlayfs，传入 `K3S_SNAPSHOTTER=fuse-overlayfs`，确认安装器转为 `--snapshotter` 参数。 |
| Cilium 挂载失败 | 挂载传播不满足要求；仅在独立容器 mount namespace 内执行 `mount --make-rshared /`，观察 `/sys`、`/run/cilium`。 |
| 修复后重装仍使用旧内容 | 仓库修复未同步 `/srv/installer`，或当前工作目录的 `w7panel/` 非空缓存被复用；重新同步并验证 HTTPS 资产，每轮使用新工作目录，保留旧目录取证。 |
| 资源 404、Chart 不存在或版本不一致 | 修正实际资产路径，补打包 Chart，运行 `check-installer-assets.py` 并从 HTTPS 逐个下载验证。 |
| Helm Pod 无法下载本地 Chart | Pod 未继承容器域名解析和 CA；用 K3s 静态目录，保留 TLS 验证。 |
| 空下载被缓存、管道失败未传播 | 缓存要求非空；下载明确失败并保留退出状态；限制重试，避免无限等候。 |
| 非交互日志密集轮询、短进程暴增 | spinner 无实际延时；使用 `sleep 2` 和 600 秒超时。 |
| registry 不可解析或镜像拉取卡住 | 检查旧 registry 域名与实际 endpoint；普通阿里云仓库不能充当通用 Docker Hub mirror。记录 ImagePull 事件，不把 API Ready 当作业务 Ready。 |
| 卸载重装选择 Cilium 地址作为访问地址 | 官方卸载残留虚拟网卡；从默认路由 `src` 获取 IP，支持 `INTERNAL_IP` 覆盖。 |
| 新版初始化脚本找不到 jq | 最终运行镜像必须安装 jq；修改构建依赖并验证运行环境，不能仅在开发容器确认命令存在。 |
| 面板 HTTP 200，但初始化 Job 失败 | 捆绑旧 CRD 覆盖独立 Helm 组件的 CRD；保护其他 release 管理的 CRD。除 release 注解外检查 `metadata.managedFields[].manager == "helm"`，读取时带 `--show-managed-fields=true`。修复后检查 Job Complete。 |
| 镜像导出 EOF | 导出尚未结束就启动卸载，破坏 containerd 状态；等待进程退出和归档有效性再卸载。本轮最终重装未使用该缓存。 |
| 声称 mock 部署但 Pod 无变量 | 安装命令环境与 Helm Pod 环境不同；检查 Deployment/Pod 实际配置并准确记录。 |

## 6. 结果维护

更新 `tests/reports/<日期>-installer.md`，区分每轮问题、环境修正、代码修复、卸载证据和最终结果。同步项目版本日志、相关文档与本 Skill；只纳入真实执行过的事实。

2026-11-06 最终第四轮完整回归通过：

- 安装脚本退出码 `0`，`/tmp/install-verified.log` 打印安装成功，正确选择访问地址 `172.20.0.3`。
- K3s `v1.36.5+k3s1`，节点 `server1` Ready；Cilium、Operator、CoreDNS、metrics、local-path、cert-manager、CSI、Higress、面板与 Agent 的长期 Pod 均 Running/Ready。
- 初始化 Job `w7panel-offline-1-upgrade-pzgme` 达到 Complete `1/1`，用时 `2m41s`。启动初期 cert-manager webhook 未就绪触发自动重试，最终成功；不能将这种短暂重试直接判为 CRD 修复失败。
- `http://172.20.0.3:9090` 返回 HTTP 200。本轮经过官方卸载后重新在线下载，未使用早期失败的镜像导出缓存。
- 新增运行依赖 `jq` 已通过 `alpine:3.20` 实际 `apk add jq` 验证；这不等于已完成整个新版镜像构建。
- 保留专用测试容器 `w7panel-installer-k3s-test` 供复查；该 Docker 阶段宿主仅只读检查，后续宿主安装另见第 7 节。安装命令携带 `LOCAL_MOCK=true`，不代表 Helm 面板 Pod 已启用该模式。

原始结果与各轮差异见 `tests/reports/2026-11-06-installer.md`。后续测试替换本节事实，不把本轮成功当作新版本通过。


## 7. 宿主测试

用户明确要求宿主安装时，先检查 hostname、PID 1、K3s/kubelet 服务、K3s 二进制和
80/443/6443/9090 端口，确认没有需要保留的业务集群；不能套用 Docker 中的卸载命令
清除不明宿主数据。已有用户授权无需再次请求业务确认，执行权限限制仍需按环境处理。

宿主使用默认 overlayfs；不要照搬容器的 `mount --make-rshared /` 或强制 fuse-overlayfs。
采用独立临时目录，HTTPS 资源服务仅监听 `127.0.0.1:443`，证书位于文档根外。
Debian 的测试根 CA 放入 `/usr/local/share/ca-certificates/<本轮标记>.crt` 后运行
`update-ca-certificates`。修改 hosts 前备份，新增行带本轮标记，以便只清理本轮变更。

可通过独立的 systemd transient unit 运行 HTTPS 服务及安装脚本，指定全新 WorkingDirectory，
将安装日志和最终退出码写入本轮目录；不能把 `systemd-run` 成功或 unit active 当作安装成功。
使用本轮 unit 名查询状态，并读取实际安装退出码。其余验收与 Docker 一致。
PUBLIC_IP 若设为内网地址以跳过外部查询，报告须说明未验证公网可达性。

宿主测试结束按用户意图保留或卸载服务。需要清理时，仅移除本轮创建的 HTTPS unit、
hosts 标记行、CA 文件和临时资源，再刷新 CA 信任库；不能覆盖用户测试期间的新配置。
本次执行证据记录在 `tests/reports/2026-11-06-installer-host.md`，以报告最终状态为准。

2026-11-06 宿主回归：Debian 13、systemd PID 1，默认 overlayfs 安装成功（退出码 0），
节点 Ready，初始化 Job 自动重试后 Complete 1/1（3m51s），内网面板返回 HTTP 200。
宿主虽无运行中的旧 K3s，仍残留旧 Cilium CNI 配置；镜像下载期间缺少 cilium.sock，
Agent 启动后自动恢复。cert-manager webhook 未就绪导致前四次初始化尝试失败，
第五次成功。检查实际 Job 状态，不能把脚本提前输出成功或短暂重试当作最终验收。
最终全部长期 Pod Running/Ready（包括 CSI 3/3、Higress Gateway 1/1）；宿主服务保留运行，未执行宿主卸载。

手动使用 `curl ... | sh -` 时，第 263 行曾报数字错误：`SECONDS` 是 Bash 变量，在 sh 中为空。现改用 `date +%s` 计算截止时间，并将 `[[ ... ]]` 改为 case。`tests/check-installer-wait.py` 隔离验证 Bash/sh 的等待成功与超时路径；旧版 sh 的 pipefail 支持仍有差异，推荐 Bash。此修复未重新执行完整安装。
