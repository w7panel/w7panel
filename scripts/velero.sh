#!/bin/sh

export TARGETARCH=amd64
RUN wget https://ghproxy.net/https://github.com/vmware-tanzu/velero/releases/download/v1.16.2/velero-v1.16.2-linux-${TARGETARCH}.tar.gz -O /tmp/velero && \
    chmod +x /tmp/velero && mv /tmp/velero /usr/local/bin/

#https://velero.io/docs/v1.16/contributions/tencent-config/


velero install  --provider aws --plugins velero/velero-plugin-for-aws:v1.1.0 --bucket  <BucketName> \
--secret-file ./credentials-velero \
--use-node-agent \
--default-volumes-to-fs-backup \
--backup-location-config \
region=ap-guangzhou,s3ForcePathStyle="true",s3Url=https://cos.ap-guangzhou.myqcloud.com


velero install  --provider aws --plugins velero/velero-plugin-for-aws:v1.5.0 --bucket w7panel-offline-1251743857 \
--secret-file /root/.config/cos.txt \
--use-volume-snapshots=false \
--backup-location-config \
region=ap-guangzhou,s3ForcePathStyle="false",s3Url=https://cos.ap-guangzhou.myqcloud.com


velero backup create backup-20260126 --include-namespaces default

velero backup create nginx-backup --include-namespaces k3k-s82
velero backup describe backup-default
velero restore create --from-backup nginx-backup
