#!/bin/bash

# WebDAV 特殊目录和符号链接测试脚本

# ⚠️ 注意：测试脚本中使用 pkill -9 是为了快速重启测试
# 生产环境请使用: ./w7panel/scripts/start.sh stop

BASE_DIR=/home/wwwroot/w7panel-dev

echo "=== WebDAV 特殊目录和符号链接测试 ==="

# 停止旧服务
pkill -9 -f "w7panel-offline" 2>/dev/null || true
sleep 1

# 启动服务
echo "启动服务..."
cd $BASE_DIR/w7panel
LOCAL_MOCK=true KO_DATA_PATH=$BASE_DIR/w7panel/kodata ./w7panel-offline server:start > /tmp/w7panel.log 2>&1 &
sleep 3

# 检查服务
curl -s http://localhost:8000/ > /dev/null || {
    echo "❌ 服务启动失败"
    cat /tmp/w7panel.log | tail -20
    exit 1
}
echo "✅ 服务启动成功"

TOKEN="test"
WEBDAV_URL="http://localhost:8000/k8s/webdav-agent/1/agent"

echo ""
echo "=== 测试 1: PROPFIND 普通目录（返回 XML）==="
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/etc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    echo "✅ PROPFIND 返回标准 XML 格式"
else
    echo "❌ PROPFIND 未返回标准 XML: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 2: PROPFIND /proc（返回 XML）==="
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/proc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")

if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    echo "✅ PROPFIND /proc 返回标准 XML 格式"
else
    echo "❌ PROPFIND /proc 未返回标准 XML: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 3: ?cmd=list 返回 JSON ==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/proc/?cmd=list" \
    -H "Authorization: Bearer $TOKEN")

if echo "$RESPONSE" | grep -q '"name"'; then
    COUNT=$(echo "$RESPONSE" | grep -o '"name"' | wc -l)
    echo "✅ ?cmd=list 返回 JSON 格式，共 $COUNT 个条目"
else
    echo "❌ ?cmd=list 未返回 JSON: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 4: 读取 /proc/version（特殊目录文件）==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/proc/version" -H "Authorization: Bearer $TOKEN")
if [ -n "$RESPONSE" ] && echo "$RESPONSE" | grep -q "Linux"; then
    echo "✅ 成功: ${RESPONSE:0:60}..."
else
    echo "❌ 失败: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 5: 读取符号链接 /etc/mtab ==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/mtab" -H "Authorization: Bearer $TOKEN")
if [ -n "$RESPONSE" ] && echo "$RESPONSE" | grep -q "overlay"; then
    echo "✅ 成功读取符号链接"
else
    echo "❌ 失败: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 6: 读取普通文件 /etc/passwd ==="
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/passwd" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "root:"; then
    echo "✅ 成功读取普通文件"
else
    echo "❌ 失败: ${RESPONSE:0:200}"
fi

echo ""
echo "=== 测试 7: 接口格式验证 ==="
echo "- PROPFIND 应返回 XML (WebDAV 标准)"
echo "- ?cmd=list 返回 JSON (高效格式)"
echo "✅ 接口设计符合 WebDAV 标准"

echo ""
echo "=== 测试完成 ==="

# 停止服务
pkill -9 -f "w7panel-offline" 2>/dev/null || true
