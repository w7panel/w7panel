# helm install --kubeconfig=/home/workspace/.kube/118.config --set servicelb.loadBalancerClass=io.cilium/node --set servicelb.port=9091 offline-old ./k8s-offline --set image.tag=1.0.37
# helm --kubeconfig=/home/workspace/.kube/218.config install w7panel-offlineold ./k8s-offline --set servicelb.loadBalancerClass=io.cilium/node --set servicelb.port=9091 --set image.tag=1.0.37


# helm --kubeconfig=/home/workspace/.kube/218.config upgrade w7panel-offline ./k8s-offline --set servicelb.loadBalancerClass=io.cilium/node --set servicelb.port=9091 --set controller.appWatch=false --set image.tag=1.0.38  --install

# helm --kubeconfig=/home/workspace/.kube/config upgrade w7panel-offline ./k8s-offline --set vmk.enabled=true --set servicelb.loadBalancerClass=io.cilium/node --set servicelb.port=9091 --set image.tag=1.0.39  --install
#--set register.needInitUser=true

helm upgrade w7panel-offline ./k8s-offline \
 --set mock.upgrade=false --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
  --set servicelb.port=9090 --set image.tag=${CNB_COMMIT_SHORT:-1.0.112.1} --set controller.appWatch=true --set mock.upgrade=false \
   --set storage.enabled=false --install

# helm --kubeconfig=/home/workspace/.kube/config upgrade w7panel-offline ./k8s-offline --set servicelb.port=9091 --set image.tag=1.0.37.40 --install


# helm upgrade w7panel-offline https://cdn.w7.cc/w7panel/charts/w7panel-offline-1.0.83.tgz \
#  --set mock.upgrade=false --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
#   --set servicelb.port=9090 --set image.tag=${CNB_COMMIT_SHORT:-1.0.107.50} --install

# kubectl get storageproviders -o json | jq '.items[] | select(.metadata.finalizers != null) | .metadata.finalizers |= map(select(. != "storage.kubeblocks.io/finalizer"))' | kubectl apply -f -
