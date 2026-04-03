

sudo ctr -n k8s.io images prune --all

sudo ctr -n k8s.io images pull 127.0.0.1:9007/afan-public/nginx:latest

docker save -o test.tar ccr.ccs.tencentyun.com/afan-public/nginx:latest

sudo ctr -n k8s.io images import ./test.tar



sudo chmod 666 /run/containerd/containerd.sock

ctr run --rm ccr.ccs.tencentyun.com/afan-public/nginx:test nginx-test1

ctr run --rm registry.local.w7.cc/afan-public/n:v1 nginx-test123

sudo ctr task exec --exec-id shell -t nginx-test1 /bin/bash