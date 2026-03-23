

helm upgrade w7panel-offline ./kodata/charts/k8s-offline \
 --set mock.upgrade=true --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
<<<<<<< HEAD
  --set servicelb.port=9090 --set image.tag=${CNB_BRANCH:-v1.1.20.1} --set image.repository=${DOCKER_BASE_REPO:-docker.cnb.cool/i0358/zpk} --set controller.appWatch=true --set k3k.watch=true  --set mock.upgrade=true \
=======
  --set servicelb.port=9090 --set image.tag=${CNB_BRANCH:-v1.1.19.4} --set image.repository=${DOCKER_BASE_REPO:-docker.cnb.cool/i0358/zpk} --set controller.appWatch=true --set k3k.watch=true  --set mock.upgrade=true \
>>>>>>> 12c5b0d855d4efd01858a8ee88cfa66ffe6c5927
   --set storage.enabled=false --install
