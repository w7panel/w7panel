export K9S_VERSION=0.50.6
export TARGETARCH=amd64
curl -L "https://ghproxy.net/https://github.com/derailed/k9s/releases/download/v${K9S_VERSION}/k9s_Linux_${TARGETARCH}.tar.gz" | tar xz \
    && chmod +x k9s \
    && mv k9s /usr/local/bin/ && chmod +x /usr/local/bin/k9s