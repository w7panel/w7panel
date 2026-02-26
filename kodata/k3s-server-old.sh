#!/bin/bash
set -e
set -o noglob

# --- helper functions for logs ---
info()
{
    echo '[INFO] ' "$@"
}

warn()
{
    RED='\033[0;31m'
    NC='\033[0m' # No Color
    echo -e "${RED}[WARN] ${NC}" "${RED}$@${NC}" >&2
}
fatal()
{
    echo '[ERROR] ' "$@" >&2
    exit 1
}

# --- add quotes to command arguments ---
quote() {
    for arg in "$@"; do
        printf '%s\n' "$arg" | sed "s/'/'\\\\''/g;1s/^/'/;\$s/\$/'/"
    done
}

# --- add indentation and trailing slash to quoted args ---
quote_indent() {
    printf ' \\\n'
    for arg in "$@"; do
        printf '\t%s \\\n' "$(quote "$arg")"
    done
}

# --- escape most punctuation characters, except quotes, forward slash, and space ---
escape() {
    printf '%s' "$@" | sed -e 's/\([][!#$%&()*;<=>?\_`{|}]\)/\\\1/g;'
}

# --- escape double quotes ---
escape_dq() {
    printf '%s' "$@" | sed -e 's/"/\\"/g'
}

eval set -- $(escape "$@") $(quote "$@")
# K3S_DATASTORE_ENDPOINT=${DATASTORE_ENDPOINT:-"SQLite"}

publicNetworkIp(){
    curl -s ifconfig.me;
    return $?
}

checkK3SInstalled()
{
    info 'start check k3s is installed 检测k3s '
    if  [ -x /usr/local/bin/k3s ]; then
            warn "K3s has been installed , Please execute /usr/local/bin/k3s-agent-uninstall.sh to uninstall k3s "
            warn "K3s 已安装 , 请先执行　/usr/local/bin/k3s-agent-uninstall.sh 命令卸载 "
            exit
        fi
}

# Install k3s #   --prefer-bundled-bin \
k3sInstall(){
   info "start join node as server 开始加入节点为server"
   curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh  | K3S_URL=${K3S_URL} K3S_KUBECONFIG_MODE='644' K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_MIRROR=cn INSTALL_K3S_SKIP_SELINUX_RPM=true \
   INSTALL_K3S_SELINUX_WARN=true INSTALL_K3S_MIRROR_URL=rancher-mirror.rancher.cn  \
   sh -s - server --embedded-registry \
   --flannel-backend "none" \
   --disable-network-policy \
   --disable-kube-proxy \
   --disable "local-storage,traefik"
}

privateDockerRegistry(){
FILE="/etc/rancher/k3s/registries.yaml"
FILE_TMP="./registries.yaml"

cat > ${FILE_TMP} <<- EOF
mirrors:
  docker.io:
    endpoint:
      - "https://mirror.ccs.tencentyun.com"
      - "https://registry.cn-hangzhou.aliyuncs.com"
      - "https://docker.m.daocloud.io"
      - "https://docker.1panel.live"
  quay.io:
    endpoint:
      - "https://quay.m.daocloud.io"
      - "https://quay.dockerproxy.com"
  gcr.io:
    endpoint:
      - "https://gcr.m.daocloud.io"
      - "https://gcr.dockerproxy.com"
  ghcr.io:
    endpoint:
      - "https://ghcr.m.daocloud.io"
      - "https://ghcr.dockerproxy.com"
  k8s.gcr.io:
    endpoint:
      - "https://k8s-gcr.m.daocloud.io"
      - "https://k8s.dockerproxy.com"
  registry.k8s.io:
    endpoint:
      - "https://k8s.m.daocloud.io"
      - "https://k8s.dockerproxy.com"
  mcr.microsoft.com:
    endpoint:
      - "https://mcr.m.daocloud.io"
      - "https://mcr.dockerproxy.com"
  nvcr.io:
    endpoint:
      - "https://nvcr.m.daocloud.io"
  registry.local.w7.cc:
  "*":
EOF

sudo mkdir -p /etc/rancher/k3s
sudo mv $FILE_TMP $FILE
rm -rf $FILE_TMP
}

tip() {
    info '注册节点成功'
}

{
    echo 'start install'
    privateDockerRegistry
    k3sInstall
    tip
}

#curl -sfL '.$baseUrl.'/k3s-agent.sh | K3S_URL=%s K3S_TOKEN=%s UUID=%s sh -

#　网关是否安装完成
#kubectl wait --for=condition=complete --timeout=10m job.batch/helm-install-traefik
# 网关是否启动成功
#kubectl wait --for=condition=available --timeout=5m deployment/traefik


