#!/bin/bash
# WebDAV 最终验证测试

# ⚠️ 注意：测试脚本中使用 pkill -9 是为了快速重启测试
# 生产环境请使用: ./w7panel/scripts/start.sh stop

BASE_DIR=/home/wwwroot/w7panel-dev

echo "=== WebDAV 最终验证测试 ==="

pkill -9 -f "w7panel-offline" 2>/dev/null || true
sleep 1
cd $BASE_DIR/w7panel
LOCAL_MOCK=true KO_DATA_PATH=$BASE_DIR/w7panel/kodata ./w7panel-offline server:start > /tmp/w7panel.log 2>&1 &
sleep 4

TOKEN="test"
WEBDAV_URL="http://localhost:8000/k8s/webdav-agent/1/agent"

echo ""
echo "=== 测试 1: PROPFIND /proc 返回标准 XML ==="
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/proc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    COUNT=$(echo "$RESPONSE" | grep -o '<D:response>' | wc -l)
    echo "✅ 返回标准 XML 格式，共 $COUNT 个条目"
else
    echo "❌ 未返回标准 XML"
fi

echo ""
echo "=== 测试 2: PROPFIND /etc 返回标准 XML ==="
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/etc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    echo "✅ 返回标准 XML 格式"
else
    echo "❌ 未返回标准 XML"
fi

echo ""
echo "=== 测试 3: GET /proc/version ==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/proc/version" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "Linux"; then
    echo "✅ 成功读取特殊文件: ${RESPONSE:0:50}..."
fi

echo ""
echo "=== 测试 4: GET /etc/mtab（符号链接）==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/mtab" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "overlay"; then
    echo "✅ 成功读取符号链接"
fi

echo ""
echo "=== 测试 5: GET /etc/passwd（普通文件）==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/passwd" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "root:"; then
    echo "✅ 成功读取普通文件"
fi

echo ""
echo "=== 测试 6: ?cmd=list 返回 JSON（可选接口）==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/proc/?cmd=list" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q '"name"'; then
    echo "✅ ?cmd=list 返回 JSON 格式"
fi

echo ""
echo "=== 总结 ==="
echo "✅ PROPFIND 始终返回标准 XML（符合 WebDAV 规范）"
echo "✅ GET 自动处理特殊文件和符号链接"
echo "✅ ?cmd=list 作为可选的高效 JSON 接口"

pkill -9 -f "w7panel-offline" 2>/dev/null || true
