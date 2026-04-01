
IV ?= v1.1.28.7

build-charts:
	helm package ./kodata/charts/k8s-offline

build-metrics:
	helm package ./kodata/charts/k8s-offline-metrics
# dev:
#     docker run --name vscode-go -p 9006:3000 -p 9007:8000 -v /home/afan/workspace:/home/workspace docker.cnb.cool/i0358/mydocker/vscode:go-1.25.0
# devc:
#     docker run --name vscode-go -p 9006:3000 -p 9007:8000 --privileged -v /home/afan/workspace:/home/workspace \
# -v /:/host:ro \
#   -v /var/lib/containerd:/var/lib/containerd:rw \
#   -v /run/containerd/containerd.sock:/run/containerd/containerd.sock \
#   -v /run/containerd:/run/containerd:rw \
# docker.cnb.cool/i0358/mydocker/vscode:go-1.25.0
	
publish:
	helm upgrade w7panel-offline ./kodata/charts/k8s-offline \
	 --set mock.upgrade=true --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
	  --set servicelb.port=9090 --set image.tag=$(IV) --set image.repository=docker.cnb.cool/i0358/zpk --set controller.appWatch=true --set k3k.watch=true  --set mock.upgrade=true \
	   --set storage.enabled=false --install
