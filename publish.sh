

helm upgrade w7panel-offline ./kodata/charts/k8s-offline \
 --set mock.upgrade=true --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
  --set servicelb.port=9090 --set image.tag=${CNB_BRANCH:-1.0.123} --set image.repository=docker.cnb.cool/i0358/zpk --set controller.appWatch=true --set k3k.watch=true  --set mock.upgrade=true \
   --set storage.enabled=false --install
