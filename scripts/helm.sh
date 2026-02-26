export HELM_VERSION=3.16.3
export TARGETARCH=amd64
curl -L "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" | tar xz \
            && mv linux-${TARGETARCH}/helm /usr/local/bin/helm \
            && rm -rf linux-${TARGETARCH} && chmod +x /usr/local/bin/helm

            