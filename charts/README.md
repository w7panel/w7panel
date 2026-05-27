# W7Panel Helm Charts

W7Panel 离线版 Helm Charts 包，用于 Kubernetes 集群部署。

## 目录结构

```
w7panel/install/charts/w7panel/
├── Chart.yaml          # Chart 元数据
├── values.yaml         # 默认配置值
└── templates/          # K8s 资源模板
    ├── deployment.yaml     # 面板部署
    ├── daemonset.yaml      # Agent DaemonSet
    ├── service.yaml        # 服务暴露
    ├── ingress.yaml        # 入口配置
    ├── serviceaccount.yaml # 服务账户
    └── ...
```

## 快速部署

```bash
# 部署到默认命名空间
helm upgrade --install w7panel ./w7panel/install/charts/w7panel -n default

# 自定义配置部署
helm upgrade --install w7panel ./w7panel/install/charts/w7panel -n default \
  --set image.tag=1.0.20 \
  --set service.type=LoadBalancer
```

## 配置说明

### 镜像配置

```yaml
image:
  repository: ccr.ccs.tencentyun.com/afan/w7panel
  pullPolicy: IfNotPresent
  tag: "1.0.19"
```

### 副本数

```yaml
replicaCount: 1
```

### 服务配置

```yaml
service:
  type: ClusterIP
  port: 80
```

### Agent DaemonSet

```yaml
daemonset:
  create: true
  name: "w7"
```

## 公测部署流程

### 1. 构建镜像

```bash
# 在项目根目录构建
cd $BASE_DIR
docker build -t ccr.ccs.tencentyun.com/afan/w7panel:1.0.20 .
```

### 2. 推送镜像

```bash
docker push ccr.ccs.tencentyun.com/afan/w7panel:1.0.20
```

### 3. 部署（通过 helm --set 指定镜像版本）

```bash
helm upgrade --install w7panel ./w7panel/install/charts/w7panel -n default \
  --set image.tag=1.0.20
```

## 版本管理

- Chart 版本定义在 `Chart.yaml` 的 `version` 字段
- 镜像版本定义在 `values.yaml` 的 `image.tag` 字段
- 两者应保持一致

## 常见问题

### Q: 如何查看部署状态？

```bash
kubectl get pods -n default -l app.kubernetes.io/name=w7panel
```

### Q: 如何查看日志？

```bash
kubectl logs -n default -l app.kubernetes.io/name=w7panel
```

### Q: 如何回滚？

```bash
helm rollback w7panel -n default
```
