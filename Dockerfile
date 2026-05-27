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

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /home/w7panel-server

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories \
    && apk add --no-cache build-base zig

COPY w7panel-server/go.mod w7panel-server/go.sum ./
RUN go mod download

COPY w7panel-server/ ./
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) zig_target="x86_64-linux-musl" ;; \
        arm64) zig_target="aarch64-linux-musl" ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    CGO_ENABLED=1 \
    GOOS="${TARGETOS}" \
    GOARCH="${TARGETARCH}" \
    CC="zig cc -target ${zig_target}" \
    CXX="zig c++ -target ${zig_target}" \
    CGO_CFLAGS="-Wno-return-local-addr -D_LARGEFILE64_SOURCE" \
    go build -trimpath -ldflags="-s -w" -o /out/w7panel .

FROM alpine:3.20 AS final

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates curl kubectl helm k9s zip tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

ENV TZ=Asia/Shanghai
ENV KO_DATA_PATH=/var/run/ko

WORKDIR /app
COPY --from=builder /out/w7panel /app/w7panel
COPY w7panel-server/kodata/ /var/run/ko/

EXPOSE 8000

ENTRYPOINT ["/app/w7panel"]
