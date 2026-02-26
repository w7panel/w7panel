
export K3K_CLI_VERSION=v0.3.5
export TARGETARCH=amd64
wget https://ghproxy.net/https://github.com/rancher/k3k/releases/download/${K3K_CLI_VERSION}/k3kcli-linux-${TARGETARCH} -O /tmp/k3kcli 
chmod +x /tmp/k3kcli && sudo mv /tmp/k3kcli /usr/local/bin/