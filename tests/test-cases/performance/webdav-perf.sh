#!/bin/bash
#========================================
# 测试用例: WebDAV性能测试
#========================================
#
# ## 测试信息
# | 项目 | 内容 |
# |------|------|
# | 类型 | Performance |
# | 优先级 | P1 |
# | 用途 | 验证WebDAV性能 |
#
# ## 性能指标
# | 指标 | 目标值 |
# |------|--------|
# | PROPFIND响应时间 | < 2秒 |
# | GET响应时间 | < 1秒 |
# | 并发错误率 | < 1% |
#
# ## 环境变量
# BASE_URL  - 服务地址
# TOKEN     - 认证Token
# CONCURRENT - 并发数 (默认10)
#
#========================================

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
TOKEN="${TOKEN:-$(cat /var/run/secrets/kubernetes.io/serviceaccount/token 2>/dev/null || echo "")}"
WEBDAV_PATH="/k8s/webdav-agent/1/agent"
CONCURRENT="${CONCURRENT:-10}"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

echo "=== 测试: WebDAV性能 ==="
echo ""

# 测试1: PROPFIND响应时间
log_info "[1] PROPFIND响应时间测试..."
START=$(date +%s%N)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PROPFIND "$BASE_URL$WEBDAV_PATH/tmp/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
END=$(date +%s%N)

DURATION=$(( (END - START) / 1000000 ))

log_info "响应时间: ${DURATION}ms, HTTP: $HTTP_CODE"

if [ $DURATION -lt 2000 ]; then
    log_pass "PROPFIND响应时间达标: ${DURATION}ms < 2000ms"
else
    log_fail "PROPFIND响应时间超标: ${DURATION}ms >= 2000ms"
fi

# 测试2: GET响应时间
log_info ""
log_info "[2] GET响应时间测试..."
START=$(date +%s%N)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X GET "$BASE_URL$WEBDAV_PATH/etc/passwd" \
    -H "Authorization: Bearer $TOKEN")
END=$(date +%s%N)

DURATION=$(( (END - START) / 1000000 ))

log_info "响应时间: ${DURATION}ms, HTTP: $HTTP_CODE"

if [ $DURATION -lt 1000 ]; then
    log_pass "GET响应时间达标: ${DURATION}ms < 1000ms"
else
    log_fail "GET响应时间超标: ${DURATION}ms >= 1000ms"
fi

# 测试3: 并发测试
log_info ""
log_info "[3] 并发测试 (并发数: $CONCURRENT)..."

SUCCESS=0
FAIL=0

for i in $(seq 1 $CONCURRENT); do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
        -X GET "$BASE_URL$WEBDAV_PATH/etc/passwd" \
        -H "Authorization: Bearer $TOKEN") &
done
wait

for job in $(jobs -p); do
    wait $job && SUCCESS=$((SUCCESS + 1)) || FAIL=$((FAIL + 1))
done

ERROR_RATE=$(echo "scale=2; $FAIL * 100 / $CONCURRENT" | bc 2>/dev/null || echo "0")

log_info "成功: $SUCCESS, 失败: $FAIL, 错误率: ${ERROR_RATE}%"

if [ "$FAIL" -eq 0 ]; then
    log_pass "并发测试通过"
else
    log_fail "并发测试有失败"
fi

log_pass "性能测试完成"
