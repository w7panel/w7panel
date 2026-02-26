#!/bin/sh
set -ex

echo "导入crd"
kubectl apply -f $KO_DATA_PATH/crds --server-side
# kubeblocks 使用新版配置
# kubectl apply -f $KO_DATA_PATH/crds-kubeblocks --server-side

echo "导入yaml"
kubectl apply -f $KO_DATA_PATH/yaml/nvidia.yaml && kubectl apply -f $KO_DATA_PATH/yaml/higress-compressor.yaml --server-side

# echo "卸载默认的vm-operator"
# helm list -n w7-system --filter 'vm-operator' | grep -q 'vm-operator' && helm uninstall vm-operator -n w7-system


echo "卸载旧版metrics pod " # 之前helm cleanup.enabled=false 导致无法删除，手动删除
kubectl delete deployment.apps/vmsingle-vm-operator-k8s-offline-metrics-single -n w7-system --ignore-not-found
kubectl delete deployment.apps/vmagent-vm-operator-k8s-offline-metrics-agent -n w7-system --ignore-not-found

echo "安装k3k"
helm upgrade --namespace k3k-system --create-namespace k3k $KO_DATA_PATH/charts/k3k-0.3.5.tgz --install --timeout 600s

kubectl create secret generic k3k-virtual --from-file=$KO_DATA_PATH/yaml/k3k/k3k-virtual.yaml -n k3k-system | echo "已存在k3k-virtual"

echo "导入k3k 0.3.5 crd"
kubectl apply -f $KO_DATA_PATH/crds-k3k 

echo "apply longhorn-volumes configmap"
kubectl create -f $KO_DATA_PATH/yaml/longhorn-volumes-config.yaml || echo "已存在longhorn-volumes-config"

echo "创建默认pvc"
# kubectl get pvc default-volume  >/dev/null 2>&1 || kubectl apply -f $KO_DATA_PATH/yaml/default-volume.yaml && kubectl apply -f $KO_DATA_PATH/yaml/default-sc.yaml
kubectl create -f $KO_DATA_PATH/yaml/default-volume.yaml || echo "已存在default-volume"
kubectl create -f $KO_DATA_PATH/yaml/default-sc.yaml || echo "已存在default-sc"
# helm upgrade --namespace k3k-system --create-namespace k3k $KO_DATA_PATH/charts/k3k-0.3.5.tgz --install --timeout 600s
echo "域名白名单插件"
kubectl create -f $KO_DATA_PATH/yaml/w7-white-domain.yaml || echo "已存在wasmplugin w7-white-domain"

kubectl patch wasmplugin w7-white-domain -n higress-system --type=merge -p '{"spec":{"url":"http://w7panel-offline.default.svc:8000/ui/wasm/plugin-domain-1.0.2.wasm"}}'



echo "API示例代码"
kubectl apply -f $KO_DATA_PATH/yaml/code

echo "创建创始人权限，新增某些菜单权限" 
# kubectl get configmap k3k.permission.founder >/dev/null 2>&1 || kubectl apply -f $KO_DATA_PATH/yaml/k3k.permission.founder.yaml --server-side
kubectl create -f $KO_DATA_PATH/yaml/permission || echo "已存在"

echo "卸载异常面板"
k8s-offline uninstall-store-panel

echo "新版metrics  "
k8s-offline metrics:upgrade

echo "升级权限菜单"
k8s-offline qx-upgrade

echo "域名解析配置"
k8s-offline domain-config
