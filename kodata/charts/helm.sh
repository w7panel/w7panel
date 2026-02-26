
helm package ./k8s-offline 


# helm install test ./helm --set register.apiServerUrl=https://172.16.1.117:6443/ \
# --set register.username=admin --set register.password=123456 --set register.thirdPartyCDToken="123456" \
# --set register.offlineUrl=http://172.16.1.117:9090 --set register.consoleBaseUrl=http://172.16.1.13:8086

# helm install k8s-offline ./helm --set register.apiServerUrl=https://106.54.230.41:6443/ \
# --set register.username=admin --set register.password=123456 --set register.thirdPartyCDToken="iSDqJkSDa7KgGhfK" \
# --set register.offlineUrl=http://106.54.230.41:9090 --set register.create=true


# helm install k8s-offline ./k8s-offline-0.1.205-arm.tgz \
# --set register.username=admin --set register.password=123456 --set register.create=true

# # helm template ./helm --set register.apiServerUrl=https://172.16.1.117:6443/ --set register.username=admin --set register.password=123456 --set register.thirdPartyCDToken=123456


# curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | INSTALL_K3S_MIRROR=cn sh -s - server \
#  --token=mysecret \
#    --datastore-endpoint="postgres://postgres:mdhnr2j9@172.16.1.117:31583/k3s"

# helm install w7panel ./k8s-offline --set servicelb.loadBalancerClass=io.cilium/node --set servicelb.port=9090 --set image.tag=1.0.27-5

helm upgrade cert-manager ./cert-manager-v1.19.2.tgz \
     --namespace cert-manager \
     --create-namespace \
     --version v1.19.2 \
     --set crds.enabled=true \
     --set prometheus.enabled=false \
     --set webhook.timeoutSeconds=4 \
     --install