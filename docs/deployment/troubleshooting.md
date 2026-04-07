# w7panel 部署与运维排障

本文档用于补充安装、初始化和服务器运维阶段的常见问题。它面向部署者、运维人员和需要自行排查环境问题的用户，不替代面板内功能使用层面的 [FAQ](../user-guide/faq.md)。

## 1. NAT 网关导致公网 IP 识别错误

### 现象

服务器本身通过 NAT 出网，安装过程中自动识别到的公网 IP 不正确，导致后续网络访问异常。

### 解决方法

安装时显式传入 `PUBLIC_IP`：

```bash
PUBLIC_IP=123.123.123.123 sh install.sh
```

---

## 2. IPv6 导致网络组件安装异常

### 现象

公网 IP 或内网 IP 使用 IPv6 时，可能导致网络组件安装错误，出现访问异常或服务未正常启动。

### 解决方法

优先关闭 IPv6，使用 IPv4 安装；如果无法自动识别，可显式指定公网 IP 与内网 IP：

```bash
PUBLIC_IP=123.123.123.123 INTERNAL_IP=123.123.123.123 sh install.sh
```

---

## 3. 多节点场景下 Server 节点 IP 变更，Agent 无法通信

### 现象

Server 节点 IP 发生变化后，Agent 节点仍然指向旧地址，导致部分服务启动异常或节点状态异常。

### 排查步骤

1. 执行以下命令，确认节点 IP 和状态是否正确：

   ```bash
   kubectl get node -o wide
   ```

2. 如果节点状态为 `NotReady`，进入 Agent 节点服务器，编辑：

   ```bash
   /etc/systemd/system/k3s-agent.service.env
   ```

3. 将其中错误的 `K3S_URL` 改为新的 Server 节点地址。

4. 保存后重启 Agent：

   ```bash
   systemctl restart k3s-agent
   ```

---

## 4. SELinux 阻止 k3s 启动

### 现象

`k3s` 或 `k3s-agent` 启动失败，并出现类似以下报错：

```bash
Job for k3s-agent.service failed because the control process exited with error code.
See "systemctl status k3s-agent.service" and "journalctl -xeu k3s-agent.service" for details.
```

### 进一步确认

查看系统日志：

```bash
cat /var/log/messages | grep k3s
```

如果看到类似以下信息，通常可以确认受 SELinux 策略影响：

```bash
systemd[35864]: k3s-agent.service: Failed to locate executable /usr/local/bin/k3s: Permission denied
setroubleshoot[10723]: SELinux is preventing /usr/lib/systemd/systemd from execute access on the file k3s
```

### 解决方法

1. 临时关闭 SELinux：

   ```bash
   setenforce 0
   ```

2. 重启对应服务：

   ```bash
   # Server 节点
   systemctl restart k3s.service

   # Agent 节点
   systemctl restart k3s-agent.service
   ```

3. 如果确认问题由 SELinux 引起，可修改 `/etc/selinux/config`：

   ```bash
   SELINUX=disabled
   ```

4. 保存后重启系统，使配置永久生效。

---

## 5. 阿里云安全组件导致面板启动失败

### 现象

在阿里云服务器上安装后，面板或相关组件启动异常，常见原因是系统中存在阿里云安全加固或阿里云盾客户端。

### 解决方法

1. 普通云服务器：重装系统时可勾选关闭安全加固，或手动卸载阿里云盾。
2. 轻量应用服务器：进入阿里云控制台 → 云安全中心 → 功能设置 → 客户端，找到服务器后执行卸载。

更多信息可参考阿里云官方文档：

https://help.aliyun.com/zh/security-center/user-guide/uninstall-the-security-center-agent

---

## 6. 腾讯云云镜占用资源过高

### 现象

小配置服务器上安装腾讯云主机安全防护（云镜）后，可能额外占用较多内存，影响面板和集群组件运行。

### 解决方法

可尝试执行以下卸载命令：

```bash
/usr/local/qcloud/YunJing/uninst.sh
/usr/local/qcloud/stargate/admin/uninstall.sh
/usr/local/qcloud/monitor/barad/admin/uninstall.sh
```

---

## 7. 服务器负载飙升，面板无法访问

### 现象

部署了较多应用后，服务器 CPU / 内存资源被大量占用，导致面板访问缓慢甚至无法打开。

### 应急处理方法

1. 查看高占用进程：

   ```bash
   top
   ```

   在 `top` 中可使用 `Shift + M` 按内存排序，必要时结合 `e` 调整显示单位。

2. 找到异常高占用进程（通常排除 `k3s-server`），临时终止高占用进程以降低负载。

3. 如果需要先恢复面板访问，可将其他应用实例临时缩容为 0：

   ```bash
   kubectl get deploy,statefulset,daemonset -n default -o name | grep -v 'w7panel-offline' | xargs -I {} sh -c '
       kubectl scale --replicas=0 {} -n default;
       kubectl get pods -n default -o json | jq -r ".items[] | select(.metadata.ownerReferences[]?.name == \"$(echo {} | cut -d'/' -f 2)\") | .metadata.name" | xargs -r kubectl delete pods -n default
   '
   ```

4. 等服务器负载恢复后，再进入面板逐步排查资源使用问题。

---

## 8. 误删微擎面板应用后无法访问

### 现象

误删 `w7panel-offline` 应用后，`http://{ip}:9090` 无法访问。

### 解决方法

1. 先查看最新版 chart 包地址：

   https://cdn.w7.cc/w7panel/manifests/w7panel-offline.yaml

2. 登录服务器后使用 Helm 重新安装，以下示例以 `1.0.55` 为例：

   ```bash
   helm upgrade --install \
   w7panel-offline \
   https://cdn.w7.cc/w7panel/charts/w7panel-offline-1.0.55.tgz \
   --version 1.0.55 \
   --namespace default \
   --atomic \
   --set servicelb.loadBalancerClass=io.cilium/node \
   --set servicelb.port=9090
   ```

---

## 9. 非大陆地区服务器镜像拉取缓慢或失败

### 现象

在香港、新加坡、美国等非大陆地区服务器上安装时，可能长时间卡在等待步骤，本质上通常是镜像拉取失败或过慢。

### 处理方法

1. 可以先使用 `Ctrl + C` 结束当前等待步骤。一般不会影响已经开始执行的安装流程。

2. 编辑 `/etc/rancher/k3s/registries.yaml`，替换为更简化的配置：

   ```yaml
   mirrors:
     registry.local.w7.cc:
     "*":
   ```

3. 重启 k3s：

   ```bash
   systemctl restart k3s.service
   ```

4. 再使用以下命令确认 Pod 是否已全部正常启动：

   ```bash
   kubectl get pod -A
   ```

   如果大部分 Pod 为 `Running` 或 `Completed`，说明安装过程已经恢复正常。

---

## 10. 忘记管理员密码

### 解决方法

在 master 服务器执行：

```bash
kubectl exec -it $(kubectl get pods -n default -l app=w7panel-offline | awk 'NR>1{print $1}') -- ko-app/k8s-offline auth:register --username=admin --password=123456
```

可以通过 `--username` 和 `--password` 传入新的管理员用户名与密码。
