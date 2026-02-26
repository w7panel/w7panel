
export KO_DOCKER_REPO=ccr.ccs.tencentyun.com/afan-public/w7panel
# export KO_DEFAULTBASEIMAGE=ccr.ccs.tencentyun.com/afan-public/alpine:3.20-k8s #ccr.ccs.tencentyun.com/afan-public/ko:base-tty2
export KO_DEFAULTBASEIMAGE=ccr.ccs.tencentyun.com/afan-public/ubuntu:24.04-offlineui
export KUBECONFIG=/home/workspace/.kube/configonline
# export KO_DATA_PATH=./asset
# export KUBECONFIG=/home/workspace/.kube/config
# ko build gitee.com/we7coreteam/k8s-offline --tags 0.1 --sbom=none
# ko apply -f dev/deployment.yaml --tags=0.1.54 --tag-only --base-import-paths=true --sbom=none 
# ko build --tags=1.0.60.22 --tag-only --base-import-paths=true --sbom=none --platform=all 
ko build --tags=1.0.109.4 --bare --tag-only  --sbom=none --platform=linux/amd64 

# kubectl create configmap registries --from-file=default.cnf=./kodata/registries.yaml
# kubectl --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt --server=https://172.16.1.13:6443 \
# --token=eyJhbGciOiJSUzI1NiIsImtpZCI6Iml3WktfcGw1VkMzdHdMei1SMjNpMzhWelQ2V0Iyb3FLSUd4RlFaWlMxbncifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzI2MTA5MDkwLCJpYXQiOjE3MjYxMDU0OTAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiNzkyNTQ3NTAtNjYyZS00MWI4LWJiMzgtMjM4YzNlYTI3NTE3Iiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiNWJlYmIzYjUtOGVjNS00NWJhLTgwYmMtOTg4OWRjZmU0ZGYwIn19LCJuYmYiOjE3MjYxMDU0OTAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.ASkl4qn18SM972Z9aQ-FRkjWOJkiBVi2O12GWYPIYxEMGiXNHEDzxm8FaPerdYrxX0d-A2amvDwXbiVdIANyIL74xbi7VjdHfPMrR2h5Mw3xJLMYBfgwcHtTY6Q335VQ1ZdEnOBXK8pQ7tqv8N0anWVzpoUcMsvsG0b-jGReQNu6_TTOauUleLANnOweV2mM9ayw1BADbclrJXKage3VwYkGB7ygiAXM0SbEosb9U82YSXXiGtzH-QYjbzI5JefsiXoRFNTzDCApoXmWM5SKqbgnhVyq1l7bMdTdZyrVlH6JglOO2UaTn6Ux7tMHazKYPFQ2qHsoP7D9mYbIruURUg apply -f ./test.txt

# helm list --all-namespaces --short | awk '{print $1}' | xargs -n1 helm uninstall

#helm list  --short | awk '{print $1}' | xargs -n1 helm uninstall

#helm --kubeconfig=/home/workspace/118.config list  --short | awk '{print $1}' | xargs -n1 helm --kubeconfig=/home/workspace/118.config uninstall
# 0979769872
# kubectl create configmap --namespace default alloy-config "--from-file=config.alloy=./config.alloy"

# https://jianghushinian.cn/2025/05/14/mcp-gateway/\