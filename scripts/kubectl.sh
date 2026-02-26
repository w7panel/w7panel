export TARGETARCH=amd64
export TARGETOS=linux
wget https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.32/deb/${TARGETARCH}/kubectl_1.32.2-1.1_${TARGETARCH}.deb -O /tmp/kubectl.deb
# RUN ls -al && dpkg -i ./kubectl_1.32.2-1.1_${TARGETARCH}.deb 
dpkg --add-architecture arm64
dpkg -i /tmp/kubectl.deb   