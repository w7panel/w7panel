#!/usr/bin/env bash
set -euo pipefail

# 通过 /k8s/pid 接口验证 users 的获取与缓存行为

# 配置项：可以通过环境变量覆盖默认值
API_URL="${K8S_PID_API_URL:-http://testcode.fan.b2.sz.w7.com/k8s/pid}"
NAMESPACE="${NAMESPACE:-default}"
HOST_IP="${HOST_IP:-10.0.0.206}"
CONTAINER_ID="${CONTAINER_ID:-containerd://717ae6ce8aed05ffbbdb3fc2a60219a4a1c18be6075fa49c3af60c147ff287d4}"
CONTAINER_NAME="${CONTAINER_NAME:-ai-opencode-gizqppwi}"
POD_NAME="${POD_NAME:-ai-opencode-gizqppwi-65cc89b7b7-zjcfz}"
TOKEN="${TOKEN:-}"

if [[ -z "${TOKEN}" ]]; then
  echo "TOKEN 未设置，跳过需要认证的测试。请设置 TOKEN 环境变量后重跑测试，例如: export TOKEN='your-jwt'"
  exit 0
fi

HEADERS=("-H" "Accept: application/json, text/plain, */*")
if [[ -n "$TOKEN" ]]; then
  HEADERS+=("-H" "Authorization: Bearer $TOKEN")
fi
HEADERS+=("--insecure")

make_request() {
  if command -v jq >/dev/null; then
    RESP=$(curl -sS "${HEADERS[@]}" -G "$API_URL" \
      --data-urlencode "namespace=$NAMESPACE" \
      --data-urlencode "HostIp=$HOST_IP" \
      --data-urlencode "containerId=$CONTAINER_ID" \
      --data-urlencode "containerName=$CONTAINER_NAME" \
      --data-urlencode "podName=$POD_NAME" 2>/dev/null)
  else
    RESP=$(curl -sS "${HEADERS[@]}" -G "$API_URL" \
      --data-urlencode "namespace=$NAMESPACE" \
      --data-urlencode "HostIp=$HOST_IP" \
      --data-urlencode "containerId=$CONTAINER_ID" \
      --data-urlencode "containerName=$CONTAINER_NAME" \
      --data-urlencode "podName=$POD_NAME" 2>/dev/null)
  fi
  echo "$RESP"
}

echo "=== First call: /k8s/pid ---" 
RESP=$(make_request)
echo "$RESP" | head -n 20

if command -v jq >/dev/null; then
  COUNT=$(echo "$RESP" | jq -r '.users | length // 0')
  echo "Users count: $COUNT"
  echo "fromCache: $(echo "$RESP" | jq -r '.fromCache')"
else
  echo "Users field presence (fallback):"
  echo "$RESP" | grep -o '"users"' || true
fi

echo
echo "=== Second call: verify caching ==="
RESP2=$(make_request)
echo "$RESP2" | head -n 20
if command -v jq >/dev/null; then
  echo "fromCache (second): $(echo "$RESP2" | jq -r '.fromCache')"
else
  echo "fromCache (second):"
  echo "$RESP2" | grep -o '"fromCache": *[^,}]*' || true
fi

exit 0
