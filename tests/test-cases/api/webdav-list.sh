#!/bin/bash
#========================================
# 测试用例: WebDAV目录列表
#========================================
#
# ## 测试信息
# | 项目 | 内容 |
# |------|------|
# | 类型 | API |
# | 优先级 | P1 |
# | 用途 | 验证WebDAV PROPFIND方法 |
#
# ## 前置条件
# - 服务已启动
# - Token已配置
#
# ## 环境变量
# BASE_URL - 服务地址 (默认: http://localhost:8000)
# TOKEN    - 认证Token
#
#========================================

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
TOKEN="${TOKEN:-$(cat /var/run/secrets/kubernetes.io/serviceaccount/token 2>/dev/null || echo "")}"
WEBDAV_PATH="/k8s/webdav-agent/1/agent"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

# 检查Token
if [ -z "$TOKEN" ]; then
    log_fail "TOKEN未配置"
    exit 1
fi

echo "=== 测试: WebDAV目录列表 ==="
echo "URL: $BASE_URL$WEBDAV_PATH/"

# 测试1: 列出根目录
echo ""
echo "[1] 列出根目录..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PROPFIND "$BASE_URL$WEBDAV_PATH/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if [ "$HTTP_CODE" = "207" ]; then
    log_pass "PROPFIND返回207 Multi-Status"
else
    log_fail "预期207，实际: $HTTP_CODE"
    exit 1
fi

# 测试2: 列出/tmp目录
echo ""
echo "[2] 列出/tmp目录..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PROPFIND "$BASE_URL$WEBDAV_PATH/tmp/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if [ "$HTTP_CODE" = "207" ]; then
    log_pass "列出/tmp目录成功"
else
    log_fail "预期207，实际: $HTTP_CODE"
    exit 1
fi

# 测试3: 无认证访问
echo ""
echo "[3] 无认证访问..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PROPFIND "$BASE_URL$WEBDAV_PATH/" \
    -H "Depth: 1")

if [ "$HTTP_CODE" = "401" ]; then
    log_pass "无认证正确返回401"
else
    log_fail "预期401，实际: $HTTP_CODE"
    exit 1
fi

echo ""
log_pass "所有测试通过"
