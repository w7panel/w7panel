#!/bin/bash
#========================================
# 测试用例: 压缩功能API
#========================================
#
# ## 测试信息
# | 项目 | 内容 |
# |------|------|
# | 类型 | API |
# | 优先级 | P1 |
# | 用途 | 验证压缩/解压API |
#
# ## 环境变量
# BASE_URL - 服务地址
# TOKEN    - 认证Token
#
#========================================

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
TOKEN="${TOKEN:-$(cat /var/run/secrets/kubernetes.io/serviceaccount/token 2>/dev/null || echo "")}"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

# 检查服务
echo "=== 测试: 压缩功能API ==="
echo "[1] 检查服务..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ "$HTTP_CODE" = "200" ]; then
    log_pass "服务运行正常"
else
    log_fail "服务异常: $HTTP_CODE"
    exit 1
fi

# 获取compressUrl
echo ""
echo "[2] 获取compressUrl..."
PID_RESP=$(curl -s -G "$BASE_URL/k8s/pid" \
    -d "namespace=default" \
    -d "containerName=w7-python" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo "")

COMPRESS_URL=$(echo "$PID_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('compressUrl',''))" 2>/dev/null || echo "")

if [ -n "$COMPRESS_URL" ]; then
    log_pass "获取compressUrl: $COMPRESS_URL"
else
    log_fail "无法获取compressUrl"
    exit 1
fi

# 测试压缩
echo ""
echo "[3] 压缩测试..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL$COMPRESS_URL/compress" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"sources":["/etc/passwd"], "output":"/tmp/test_w7panel.tar.gz"}')

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    log_pass "压缩成功"
else
    log_fail "压缩失败: $HTTP_CODE"
    exit 1
fi

log_pass "所有测试通过"
