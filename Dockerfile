# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# Scratch can be used as the base image because the backend is compiled to include all
# its dependencies.
# FROM scratch as final
# FROM ccr.ccs.tencentyun.com/afan-public/alpine:3.20 as final
FROM ccr.ccs.tencentyun.com/afan-public/alpine:3.20 as final
#FROM alpine:3.20 as final
ARG TARGETARCH
ARG TARGETOS
# 使用阿里云 Ubuntu 镜像作为基础镜像


# # 添加阿里云的 APT 源并更新包列表
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
#     && apt-get update

# RUN sed -i 's|http://archive.ubuntu.com|https://mirrors.aliyun.com|g' /etc/apt/sources.list \
#     && apt-get update
    # && apt-get install -y --no-install-recommends ca-certificates \
    # && rm -rf /var/lib/apt/lists/*

# 安装curl（或者其他你需要的软件包）
RUN apk update && apk add curl kubectl helm k9s zip && rm -rf /var/cache/apk/*
ENV TZ Asia/Shanghai
RUN apk add alpine-conf && \
    /sbin/setup-timezone -z Asia/Shanghai && \
    apk del alpine-conf
# RUN apt-get update && apt-get install -y curl
# Add all files from current working directory to the root of the image, i.e., copy dist directory
# layout to the root directory.
# Copy nonroot user
RUN echo "source <(kubectl completion bash)" >> ~/.bashrc
# Declare the port on which the webserver will be exposed.
EXPOSE 8000

# Run the compiled binary.
# ko 构建的基础镜像
#docker buildx build -t ccr.ccs.tencentyun.com/afan-public/ko:base . --push
