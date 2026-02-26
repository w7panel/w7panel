#!/bin/bash
echo "不再执行升级"
if [ "${K3K_MODE}" == "virtual" ]; then
kubectl delete ing/ing-k3k-agent -n default --ignore-not-found
fi 
# if [ "${K3K_MODE}" == "virtual" ]; then
# kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
# kind: Ingress
# apiVersion: networking.k8s.io/v1
# metadata:
#     name: ing-k3k-agent
#     namespace: default
#     labels:
#         app: w7panel-k3k-agent-${K3K_NAME}
#         group: w7panel-k3k-agent-${K3K_NAME}
#         higress.io/resource-definer: higress
#     annotations:
#         higress.io/resource-definer: higress
#         kubernetes.io/ingress.class: higress
# spec:
#     rules:
#         -
#             host: w7panel-k3k-agent-${K3K_NAME}.w7panel.xyz
#             http:
#                 paths:
#                     -
#                         path: /
#                         pathType: Prefix
#                         backend:
#                             service:
#                                 name: w7panel-k3k-agent-${K3K_NAME}
#                                 port:
#                                     number: 8000
# EOF
# fi

# 低版本agent 没有cert-manager
if [ "${K3K_MODE}" == "virtual" ]; then
helm get notes cert-manager -n cert-manager || helm upgrade cert-manager $KO_DATA_PATH/charts/cert-manager-v1.19.2.tgz \
     --kubeconfig=${KUBECONFIG_PATH} \
     --namespace cert-manager \
     --create-namespace \
     --version v1.19.2 \
     --set crds.enabled=true \
     --set prometheus.enabled=false \
     --set webhook.timeoutSeconds=4 \
     --install
fi

if [ "${K3K_MODE}" == "virtual" ]; then
kubectl get ClusterIssuer/w7-letsencrypt-prod || kubectl --kubeconfig=${KUBECONFIG_PATH} apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: w7-letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: 446897682@qq.com
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: w7-letsencrypt-prod
    # Enable the HTTP-01 challenge provider
    solvers:
      - http01:
          ingress:
            class: higress
EOF
           
fi