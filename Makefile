hello:
	echo "xxx"

build:
	CGO_CFLAGS="-Wno-return-local-address" go build -o $(or $(OUTPUT),./w7panel) .

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_CFLAGS="-Wno-return-local-address" go build -o $(or $(OUTPUT),./w7panel-offline) .

dev:
	CGO_CFLAGS="-Wno-return-local-address" go run main.go server:start

test:
	go test ./...

publish:
	helm upgrade w7panel-offline ./kodata/charts/k8s-offline \
	 --set mock.upgrade=true --set servicelb.loadBalancerClass=io.cilium/node --set webhook.enabled=true \
	  --set servicelb.port=9090 --set image.tag=${CNB_BRANCH:-v1.1.26.3} --set image.repository=${DOCKER_BASE_REPO:-docker.cnb.cool/i0358/zpk} --set controller.appWatch=true --set k3k.watch=true  --set mock.upgrade=true \
	   --set storage.enabled=false --install
