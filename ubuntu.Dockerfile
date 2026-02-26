# 使用中科大Debian镜像作为基础
# FROM --platform=$BUILDPLATFORM docker.m.daocloud.io/library/debian:bullseye-slim
FROM --platform=$BUILDPLATFORM docker.m.daocloud.io/library/ubuntu:24.04

# 设置目标架构参数
ARG TARGETARCH

# 替换APT为清华源
RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
    && sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list




# 安装基础工具
RUN apt-get update && apt-get install -y \
    bash \
    bash-completion \
    ca-certificates \
    curl \
    git \
    unzip \
    zip \
    wget \
    vim \
    jq
# && rm -rf /var/lib/apt/lists/*

RUN wget https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.32/deb/${TARGETARCH}/kubectl_1.32.2-1.1_${TARGETARCH}.deb -O /tmp/kubectl.deb
# RUN ls -al && dpkg -i ./kubectl_1.32.2-1.1_${TARGETARCH}.deb 
RUN dpkg --add-architecture arm64
RUN dpkg -i /tmp/kubectl.deb   
# 设置默认shell为bash
SHELL ["/bin/bash", "-c"]
# https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.32/deb/amd64/kubectl_1.32.2-1.1_amd64.deb
# 安装helm（使用华为云镜像）
# ARG HELM_VERSION=3.14.0
# RUN arch=${TARGETARCH/amd64/amd64} \
#     && curl -L "https://mirrors.huaweicloud.com/helm/v${HELM_VERSION}/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" | tar xz \
#     && mv linux-${TARGETARCH}/helm /usr/local/bin/helm \
#     && rm -rf linux-${TARGETARCH}


# 安装k9s（使用GitHub代理）
ARG K9S_VERSION=0.50.6
RUN curl -L "https://ghproxy.net/https://github.com/derailed/k9s/releases/download/v${K9S_VERSION}/k9s_Linux_${TARGETARCH}.tar.gz" | tar xz \
    && chmod +x k9s \
    && mv k9s /usr/local/bin/ && chmod +x /usr/local/bin/k9s

ARG HELM_VERSION=3.16.3
RUN curl -L "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" | tar xz \
    && mv linux-${TARGETARCH}/helm /usr/local/bin/helm \
    && rm -rf linux-${TARGETARCH} && chmod +x /usr/local/bin/helm

# RUN mkdir -p /etc/bash_completion.d \
#             && kubectl completion bash > /etc/bash_completion.d/kubectl \
#             && helm completion bash > /etc/bash_completion.d/helm \
#             && k9s completion bash > /etc/bash_completion.d/k9s \
#             && echo "source /usr/share/bash-completion/bash_completion" >> /etc/bash.bashrc            
#             # 安装kubectl（使用阿里云镜像）
# ARG KUBECTL_VERSION=1.32.1
# RUN arch=${TARGETARCH/amd64/x86_64} \
#         && curl -L "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" \
#         -o /usr/local/bin/kubectl \
#         && chmod +x /usr/local/bin/kubectl

# # 配置自动补全
RUN apt-get install -y bash-completion
RUN mkdir -p /etc/bash_completion.d 
RUN echo "source <(kubectl completion bash)" >> ~/.bashrc
RUN echo "source <(helm completion bash)" >> ~/.bashrc
RUN echo "source /usr/share/bash-completion/bash_completion" >> /etc/bash.bashrc

# //https://github.com/rancher/k3k/releases/download/v0.3.3-rc1/k3kcli-linux-amd64
ARG K3K_CLI_VERSION=v0.3.3

RUN wget https://ghproxy.net/https://github.com/rancher/k3k/releases/download/${K3K_CLI_VERSION}/k3kcli-linux-${TARGETARCH} -O /tmp/k3kcli && \
    chmod +x /tmp/k3kcli && mv /tmp/k3kcli /usr/local/bin/
# RUN apt-get install -y vi
# 设置工作目录
# WORKDIR /root
# wget https://ghproxy.net/https://github.com/rancher/k3k/releases/download/v0.3.3-rc1/k3kcli-linux-amd64 -O /tmp/k3kcli && chmod +x /tmp/k3kcli && mv /tmp/k3kcli /usr/local/bin/
# 验证安装
# RUN kubectl version --client --short \
#     && helm version --short \
#     && k9s version --short

RUN wget https://ghproxy.net/https://github.com/vmware-tanzu/velero/releases/download/v1.16.2/velero-v1.16.2-linux-${TARGETARCH}.tar.gz -O /tmp/velero.tar.gz && \
    cd /tmp && tar -xvf /tmp/velero.tar.gz && \
    # RUN ls -al /tmp
    chmod +x /tmp/velero-v1.16.2-linux-${TARGETARCH}/velero && mv /tmp/velero-v1.16.2-linux-${TARGETARCH}/velero /usr/local/bin/

RUN curl -fsSL https://kubeblocks.io/installer/install_cli.sh | bash -s v1.0.1

CMD ["bash"]

#docker buildx build --platform linux/amd64,linux/arm64 -t ccr.ccs.tencentyun.com/afan-public/ubuntu:24.04-offlineui . -f ./ubuntu.Dockerfile --push



# wget https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.32/deb/amd64/kubectl_1.32.2-1.1_amd64.deb -O /tmp/kubectl.deb