#!/bin/bash
set -e
# 检查终端是否支持颜色输出
if [ -t 1 ]; then
    INFO_COLOR='\033[0m'
    WARN_COLOR='\033[33m'
    ERROR_COLOR='\033[31m'
    SUCCESS_COLOR='\033[32m'
    NC='\033[0m' # No Color
else
    INFO_COLOR=''
    WARN_COLOR=''
    ERROR_COLOR=''
    SUCCESS_COLOR=''
    NC=''
fi

# 信息日志
info() {
    printf "${INFO_COLOR}[INFO]${NC} %s\n" "$*"
}

# 警告日志
warn() {
    printf "${WARN_COLOR}[WARN] %s${NC}\n" "$*" >&2
}

# 成功日志
success() {
    printf "${SUCCESS_COLOR}[SUCCESS]${NC} %s\n" "$*"
}

# 错误日志
fatal() {
    printf "${ERROR_COLOR}[ERROR] %s${NC}\n" "$*" >&2
    exit 1
}
# 测试变量
# export K3K_NAME=s27
# export STORAGE_CLASS_NAME=disk1                                                      
# export K3K_MODE=shared                                                               
# export K3K_NAMESPACE=k3k-${K3K_NAME}
# export KO_DATA_PATH=/home/workspace/k8s-offline/kodata
# export KUBECONFIG_PATH=/home/workspace/k8s-offline/kodata/shell/k3k-${K3K_NAME}-${K3K_NAME}-kubeconfig.yaml
# export INGRESS_CLASS=higress
# export K3K_POLICY=oqarnvwa
# export LOCAL=1

info "判断LOCAL环境变量是否存在..."
if [[ -n "${LOCAL}" ]]; then
    info "LOCAL环境变量存在，设置测试变量..."
    export K3K_NAME=s99
    export STORAGE_CLASS_NAME=disk1                                                      
    export K3K_MODE=virtual                                                               
    export K3K_NAMESPACE=k3k-${K3K_NAME}
    export KO_DATA_PATH=/home/workspace/k8s-offline/kodata
    export KUBECONFIG_PATH=/home/workspace/k8s-offline/kodata/shell/k3k-${K3K_NAME}-${K3K_NAME}-kubeconfig.yaml
    export INGRESS_CLASS=higress
    export K3K_POLICY=odytwrat
    export K3K_STORAGE_REQUEST_SIZE=5Gi
    export K3K_PVC_STORAGE_REQUEST_SIZE=1Gi
    export DEFAULT_VOLUME_NAME=default-volume
    success "测试变量设置完成"
else
    info "LOCAL环境变量不存在，使用默认变量"
fi

info "创建Addons"
#kubectl create secret generic k3k-virtual --from-file=$KO_DATA_PATH/yaml/k3k/k3k-virtual.yaml -n k3k-system


info "开始创建集群..."
# k3kcli cluster create --storage-class-name=${STORAGE_CLASS_NAME} --mode ${K3K_MODE} --namespace ${K3K_NAMESPACE} --kubeconfig-server=${KUBECONFIG_SERVER} --policy=${K3K_POLICY} --server-args="--system-default-registry=registry.cn-hangzhou.aliyuncs.com" --server-envs="K3S_SYSTEM_DEFAULT_REGISTRY=registry.cn-hangzhou.aliyuncs.com" --storage-request-size=$K3K_STORAGE_REQUEST_SIZE ${K3K_NAME}
k3kcli cluster create --storage-class-name=${STORAGE_CLASS_NAME} --mode ${K3K_MODE} --namespace ${K3K_NAMESPACE} --kubeconfig-server=${KUBECONFIG_SERVER} --policy=${K3K_POLICY}  --storage-request-size=$K3K_STORAGE_REQUEST_SIZE --server-args='--kubelet-arg=$cgroup_root' --server-args="--disable=traefik" --server-args="--embedded-registry" --server-args="--disable-network-policy" ${K3K_NAME} 
# kubectl apply -f - <<EOF
# apiVersion: k3k.io/v1alpha1
# kind: Cluster
# metadata:
#     name: ${K3K_NAME}
#     namespace: k3k-${K3K_NAME}
# spec:
#     agents: 0
#     expose:
#         nodePort: {}
#     mode: ${K3K_MODE}
#     persistence:
#         storageClassName: ${STORAGE_CLASS_NAME}
#         storageRequestSize: ${K3K_STORAGE_REQUEST_SIZE}
#         type: dynamic
#     serverArgs:
#         - '--system-default-registry=registry.cn-hangzhou.aliyuncs.com'
#     serverEnvs:
#         -
#             name: K3S_SYSTEM_DEFAULT_REGISTRY
#             value: registry.cn-hangzhou.aliyuncs.com
#     servers: 1
#     # tlsSANs:
#     #     - k3k-${K3K_NAME}-service.k3k-${K3K_NAME}
#     version: ''
#     serverLimit:
#         cpu: 1
#         memory: 1Gi
# EOF


# /home/workspace/k8s-offline/kodata/charts/k3k-k7-k7-kubeconfig.yaml
if [ -z "${LOCAL}" ]; then
    info "修改kubeconfig server端口为443..."
    sed -i 's|:[0-9]\+$|:443|' ${KUBECONFIG_PATH}
else
    info "LOCAL环境变量存在，跳过修改kubeconfig server端口"
fi



info "部署CRDs资源..."
kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f $KO_DATA_PATH/crds --server-side 
#&& kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f $KO_DATA_PATH/crds-kubeblocks --server-side

info "检查并创建默认用户..."
if ! kubectl --kubeconfig=${KUBECONFIG_PATH} get sa $K3K_NAME &> /dev/null; then
    kubectl --kubeconfig=${KUBECONFIG_PATH} create sa $K3K_NAME
    success "默认用户 $K3K_NAME 创建成功"
else
    warn "默认用户 $K3K_NAME 已存在，跳过创建"
fi

info "检查并绑定集群角色..."
if ! kubectl --kubeconfig=${KUBECONFIG_PATH} get clusterrolebinding $K3K_NAME-cluster-admin &> /dev/null; then
    kubectl --kubeconfig=${KUBECONFIG_PATH} create clusterrolebinding $K3K_NAME-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:$K3K_NAME
    success "集群角色绑定 $K3K_NAME-cluster-admin 创建成功"
else
    warn "集群角色绑定 $K3K_NAME-cluster-admin 已存在，跳过创建"
fi


# 仅在shared模式下处理ingressclass
if [ "${K3K_MODE}" == "shared" ]; then
    info "检查并设置默认ingressclass..."
    if ! kubectl --kubeconfig=${KUBECONFIG_PATH} get ingressclass ${INGRESS_CLASS} &> /dev/null; then
        kubectl --kubeconfig=${KUBECONFIG_PATH} create -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: ${INGRESS_CLASS}
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
spec:
  controller: ${INGRESS_CLASS}.io/ingress-controller
EOF
        success "创建${INGRESS_CLASS} ingressclass并设置为默认"
    else
        kubectl --kubeconfig=${KUBECONFIG_PATH} annotate ingressclass ${INGRESS_CLASS} ingressclass.kubernetes.io/is-default-class="true" --overwrite
        success "${INGRESS_CLASS} ingressclass已存在，设置为默认"
    fi
else
    info "当前模式为${K3K_MODE}，跳过ingressclass设置"
fi

# 仅在shared模式下创建和设置默认storageclass
if [ "${K3K_MODE}" == "shared" ]; then
    info "检查并设置默认storageclass..."
    if ! kubectl --kubeconfig=${KUBECONFIG_PATH} get storageclass ${STORAGE_CLASS_NAME} &> /dev/null; then
        info "创建${STORAGE_CLASS_NAME} storageclass..."
        kubectl --kubeconfig=${KUBECONFIG_PATH} create -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ${STORAGE_CLASS_NAME}
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
        success "创建${STORAGE_CLASS_NAME} storageclass成功"
    fi
    # 设置为默认storageclass
    kubectl --kubeconfig=${KUBECONFIG_PATH} patch storageclass ${STORAGE_CLASS_NAME} -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
    success "${STORAGE_CLASS_NAME} storageclass已设置为默认"
else
    info "当前模式为${K3K_MODE}，跳过storageclass创建和设置"
fi



# 根据K3K_MODE创建不同的PVC
if [ "${K3K_MODE}" == "shared" ] || [ "${K3K_MODE}" == "virtual" ]; then
    info "检查默认PVC ${DEFAULT_VOLUME_NAME}..."
    if ! kubectl --kubeconfig=${KUBECONFIG_PATH} get pvc ${DEFAULT_VOLUME_NAME} &> /dev/null; then
        info "创建默认PVC ${DEFAULT_VOLUME_NAME}..."
        # 根据模式设置storageClassName
        if [ "${K3K_MODE}" == "shared" ]; then
            STORAGE_CLASS_TO_USE="${STORAGE_CLASS_NAME}"
        else
            STORAGE_CLASS_TO_USE="local-path"
        fi
        
        kubectl --kubeconfig=${KUBECONFIG_PATH} create -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ${DEFAULT_VOLUME_NAME}
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ${STORAGE_CLASS_TO_USE}
  resources:
    requests:
      storage: ${K3K_PVC_STORAGE_REQUEST_SIZE}
EOF
        success "默认PVC ${DEFAULT_VOLUME_NAME} 创建成功"
    else
        info "默认PVC ${DEFAULT_VOLUME_NAME} 已存在，跳过创建"
    fi
fi

info "配置volumes-config"
kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
kind: ConfigMap
apiVersion: v1
metadata:
    name: longhorn-volumes-config
data:
    customs: ${DEFAULT_VOLUME_NAME}
    default: ${DEFAULT_VOLUME_NAME}
EOF

info "配置k3s.config configmap..."
kubectl --kubeconfig=${KUBECONFIG_PATH} -n kube-system create configmap k3s.config --from-literal=k3s.mode=${K3S_MODE} --dry-run=client -o yaml | kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f -
success "k3s.config configmap配置完成"


# 检查并删除主集群networkpolicy k3k-s24（如果存在）
info "检查并清理 networkpolicy k3k-${K3K_NAME}..."
if kubectl get networkpolicies.networking.k8s.io/k3k-${K3K_NAME} &> /dev/null; then
    kubectl delete networkpolicies.networking.k8s.io/k3k-${K3K_NAME} &> /dev/null || true
    success "已删除networkpolicy k3k-${K3K_NAME}"
else
    info "networkpolicy k3k-${K3K_NAME}不存在，跳过删除"
fi

# kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
# kind: ConfigMap
# apiVersion: v1
# metadata:
#     name: registries
#     namespace: default
#     annotations:
#         title: 镜像仓库
# data:
#     default.cnf: |

#         mirrors:
#           docker.io:
#             endpoint:
#               - "https://mirror.ccs.tencentyun.com"
#               - "https://registry.cn-hangzhou.aliyuncs.com"
#               - "https://docker.m.daocloud.io"
#               - "https://docker.1panel.live"
#               - "https://docker.1ms.run"
#           quay.io:
#             endpoint:
#               - "https://quay.m.daocloud.io"
#               - "https://quay.dockerproxy.com"
#           gcr.io:
#             endpoint:
#               - "https://gcr.m.daocloud.io"
#               - "https://gcr.dockerproxy.com"
#           ghcr.io:
#             endpoint:
#               - "https://ghcr.m.daocloud.io"
#               - "https://ghcr.dockerproxy.com"
#           k8s.gcr.io:
#             endpoint:
#               - "https://k8s-gcr.m.daocloud.io"
#               - "https://k8s.dockerproxy.com"
#           registry.k8s.io:
#             endpoint:
#               - "https://k8s.m.daocloud.io"
#               - "https://k8s.dockerproxy.com"
#           mcr.microsoft.com:
#             endpoint:
#               - "https://mcr.m.daocloud.io"
#               - "https://mcr.dockerproxy.com"
#           nvcr.io:
#             endpoint:
#               - "https://nvcr.m.daocloud.io"
#           registry.local.w7.cc:
#           "*": 
# EOF

# 安装higress
if [ "${K3K_MODE}" == "virtual" ]; then
   helm upgrade higress https://cdn.w7.cc/w7panel/charts/higress-2.1.6.tgz \
     --kubeconfig=${KUBECONFIG_PATH} \
     --namespace higress-system \
     --create-namespace \
     --version v2.1.6 \
     --set global.ingressClass=higress \
     --set higress-core.gateway.replicas=1 \
     --set higress-core.gateway.resources.limits.cpu=0 \
     --set higress-core.gateway.resources.limits.memory=0 \
     --set higress-core.gateway.resources.requests.cpu=0 \
     --set higress-core.gateway.resources.requests.memory=0 \
     --set higress-core.controller.replicas=1 \
     --set higress-core.controller.resources.requests.cpu=0 \
     --set higress-core.controller.resources.requests.memory=0 \
     --set higress-core.controller.resources.limits.cpu=0 \
     --set higress-core.controller.resources.limits.memory=0 \
     --set higress-core.pilot.replicaCount=1 \
     --set higress-core.pilot.resources.requests.cpu=0 \
     --set higress-core.pilot.resources.requests.memory=0 \
     --set higress-console.replicaCount=0 \
     --set higress-console.resources.requests.cpu=0 \
     --set higress-console.resources.requests.memory=0 --install
fi

if [ "${K3K_MODE}" == "virtual" ]; then
kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
    name: higress-gateway-global-route-config
    namespace: higress-system
spec:
    configPatches:
        -
            applyTo: NETWORK_FILTER
            match:
                context: GATEWAY
                listener:
                    filterChain:
                        filter:
                            name: envoy.filters.network.http_connection_manager
            patch:
                operation: MERGE
                value:
                    typed_config:
                        '@type': >-
                            type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                        skip_xff_append: true
                        xff_num_trusted_hops: 2
        -
            applyTo: ROUTE_CONFIGURATION
            match:
                context: GATEWAY
            patch:
                operation: MERGE
                value:
                    request_headers_to_add:
                        -
                            append: false
                            header:
                                key: x-real-ip
                                value: '%REQ(X-Forwarded-For)%'
                        -
                            append: false
                            header:
                                key: X-Forwarded-Proto
                                value: '%REQ(X-Forwarded-Proto)%'
EOF
           
fi

# 必须用helm 安装否则 clusterissuer 无法创建(因为crds未创建出来) 放到agent initcontainer里了
# if [ "${K3K_MODE}" == "virtual" ]; then
# helm --kubeconfig=${KUBECONFIG_PATH} get notes cert-manager -n cert-manager || helm upgrade cert-manager https://cdn.w7.cc/w7panel/charts/cert-manager-v1.16.2.tgz \
#      --kubeconfig=${KUBECONFIG_PATH} \
#      --namespace cert-manager \
#      --create-namespace \
#      --version v1.16.2 \
#      --set crds.enabled=true \
#      --set prometheus.enabled=false \
#      --set webhook.timeoutSeconds=4 \
#      --install
# fi

# if [ "${K3K_MODE}" == "virtual" ]; then
# kubectl --kubeconfig=${KUBECONFIG_PATH} get ClusterIssuer/w7-letsencrypt-prod || kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
# apiVersion: cert-manager.io/v1
# kind: ClusterIssuer
# metadata:
#   name: w7-letsencrypt-prod
# spec:
#   acme:
#     # The ACME server URL
#     server: https://acme-v02.api.letsencrypt.org/directory
#     # Email address used for ACME registration
#     email: 446897682@qq.com
#     # Name of a secret used to store the ACME account private key
#     privateKeySecretRef:
#       name: w7-letsencrypt-prod
#     # Enable the HTTP-01 challenge provider
#     solvers:
#       - http01:
#           ingress:
#             class: higress
# EOF
           
# fi
success "集群创建完成"