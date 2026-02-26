FROM ccr.ccs.tencentyun.com/afan-public/ubuntu:24.04-offlineui-k3k

ARG TARGETARCH
RUN wget -O /tmp/go.linux.tar.gz https://mirrors.aliyun.com/golang/go1.23.4.linux-${TARGETARCH}.tar.gz
# Extracts the Go binary archive (go.linux-amd64.tar.gz) into the /home/ directory
RUN tar -xf /tmp/go.linux.tar.gz -C /tmp/
RUN rm -rf /tmp/go.linux.tar.gz
RUN ln -s /tmp/go/bin/go /usr/local/bin/go
RUN go env -w  GOPROXY=https://goproxy.cn,direct
RUN export PATH=$PATH:/home/workspace/go/bin
RUN go install github.com/google/ko@latest 
RUN export PATH=$PATH:/root/go/bin

# docker buildx build --platform linux/amd64,linux/arm64 -t ccr.ccs.tencentyun.com/afan-public/ubuntu:go-build . -f ./build.Dockerfile --push