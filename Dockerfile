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

FROM golang:1.26-alpine AS builder

WORKDIR /home/w7panel-server

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories \
    && apk add --no-cache build-base

COPY w7panel-server/go.mod w7panel-server/go.sum ./
RUN go mod download

COPY w7panel-server/ ./
RUN CGO_CFLAGS="-Wno-return-local-address" go build -trimpath -ldflags="-s -w" -o /out/w7panel .

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
