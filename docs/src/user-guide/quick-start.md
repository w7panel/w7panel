# 快速开始

第一次安装微擎面板时，你只需要按下面 3 步完成：

1. 安装前确认环境
2. 执行安装命令
3. 首次进入后台

## 第一步：安装前确认环境

在执行安装命令前，请先确认服务器满足下面这些条件：

- 节点服务器配置 **不低于 2 核 4G**
- 支持主流 Linux 发行版
- 推荐使用 **CentOS Stream 9 及以上** 或 **Ubuntu Server 22 及以上**
- 服务器外网端口 **6443、80、443、9090** 可访问
- 建议使用**全新的服务器环境**安装
- 不建议与其他面板系统混用

如果环境过旧、端口受限或存在其他面板冲突，安装后可能会出现组件启动异常、页面无法访问或功能不稳定等问题。

## 第二步：执行安装命令

在目标服务器中执行下面的安装命令：

```bash
curl -sfL https://cdn.w7.cc/w7panel/install.sh | sh -
```

安装过程中建议保持终端连接稳定，并等待脚本完整执行结束。

## 第三步：首次进入后台

安装完成后，可以通过下面的地址首次进入后台：

```text
http://{ip}:9090
```

其中 `{ip}` 替换为你的服务器公网 IP 或可访问地址。

尚未设置管理员时，首次打开后台会进入初始化步骤，你可以：

- 设置管理员账号
- 设置管理员密码

已经初始化的面板使用已有账号登录，不需要重复创建管理员。登录页面显示的其他登录方式由管理员配置决定。多人使用时，请由管理员分配账号和权限。

建议在首次配置时就使用容易记住、但足够安全的管理员账号密码，并妥善保管。

## 安装完成后建议先做什么

首次进入后台后，建议先：

1. 确认页面可以正常打开
2. 确认面板基础状态是否正常
3. 进入应用管理开始创建或接入应用

## 无法访问后台时先检查什么

如果安装完成后仍然无法正常访问后台，建议优先检查：

1. 服务器公网 IP 是否正确
2. `9090`、`80`、`443`、`6443` 端口是否已经放行
3. 当前服务器是否存在和其他面板或服务的冲突


## 本地 HTTPS 安装脚本验证

`installer/` 需要作为网站 `/installer/` 路径发布。使用 mkcert 为
`w7panel.com` 签发证书，在独立测试机中将该域名解析到本地 HTTPS 服务，
并将 mkcert 的根证书导入测试机信任库。不要把证书私钥提交到仓库。

```bash
mkcert -cert-file w7panel.com.pem -key-file w7panel.com-key.pem w7panel.com
curl -fsS https://w7panel.com/installer/install.sh -o /tmp/w7panel-install.sh
LOCAL_MOCK=true bash /tmp/w7panel-install.sh
```

安装环境需要 systemd、curl、wget、sudo 和 iproute2。嵌套容器不支持 overlayfs
时，安装 `fuse-overlayfs` 并为测试命令增加 `K3S_SNAPSHOTTER=fuse-overlayfs`；
也可使用 `K3S_SNAPSHOTTER=native`，但复制文件更慢且占用更多磁盘（脚本默认 `overlayfs`）。
Cilium 还要求测试容器内 `/sys`、`/run/cilium` 等挂载具备 shared 挂载传播；这属于容器环境准备，
应在独立特权测试容器内配置，不能把普通应用容器当成完整安装环境。`LOCAL_MOCK` 是面板进程配置，
仅在安装命令中设置并不会自动改变 Helm 部署中的面板运行模式。

安装脚本使用在线镜像拉取，不再强制下载缺失的镜像 tar 包；已有的
`w7panel/images/*.tar` 仍可导入。Chart 先下载到 K3s 的 `server/static/charts`，
再通过 Kubernetes API 提供给 Helm 安装任务。发布时必须包含 Chart 包：

```bash
helm package charts/w7panel --version 1.1.102 --app-version 1.1.102 --destination installer/charts
```

面板 manifest 的 Chart 版本、文件名必须与打包结果一致。等待面板 HTTP 就绪
最多 600 秒，超时后检查 `k3s kubectl get pods -A` 和 `journalctl -u k3s`。
仅在专用测试机执行 `/usr/local/bin/k3s-uninstall.sh` 卸载，然后重新下载脚本复测。


重装时安装器的服务访问地址按默认路由的源 IP 检测，避免选中残留的 `cilium_host` 虚拟网卡；
多网卡环境可以显式设置 `INTERNAL_IP`（控制安装器访问地址和证书 SAN，不替代 K3s 的 `--node-ip`）。面板初始化 Job 会保留其他 Helm release
管理的 CRD（含没有 release 注解、但由 Helm 创建的 CRD），其版本由对应组件
Chart 维护，防止面板镜像捆绑的旧 Higress 定义造成 server-side apply 冲突。
