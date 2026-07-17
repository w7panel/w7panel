#!/bin/bash

BASE_DIR=/home/wwwroot/w7panel-dev
TOKEN="test"
WEBDAV_URL="http://localhost:8000/k8s/webdav-agent/1/agent"

echo "=========================================="
echo "    WebDAV 文件管理 & 文本编辑器 UI 测试"
echo "=========================================="

echo ""
echo "=== 1. 文件管理测试 ==="

echo ""
echo "1.1 PROPFIND 普通目录 /etc/"
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/etc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    COUNT=$(echo "$RESPONSE" | grep -o '<D:response>' | wc -l)
    echo "    ✅ 列出 /etc/ 成功，共 $COUNT 个条目"
else
    echo "    ❌ 失败"
fi

echo ""
echo "1.2 PROPFIND 特殊目录 /proc/"
START=$(date +%s%3N)
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/proc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
END=$(date +%s%3N)
TIME=$((END - START))
if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    COUNT=$(echo "$RESPONSE" | grep -o '<D:response>' | wc -l)
    echo "    ✅ 列出 /proc/ 成功，共 $COUNT 个条目，耗时 ${TIME}ms"
    if [ $TIME -lt 100 ]; then
        echo "    ✅ 性能良好（<100ms）"
    else
        echo "    ⚠️ 性能较慢"
    fi
else
    echo "    ❌ 失败"
fi

echo ""
echo "1.3 读取普通文件 /etc/passwd"
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/passwd" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "root:"; then
    LINES=$(echo "$RESPONSE" | wc -l)
    echo "    ✅ 读取成功，共 $LINES 行"
else
    echo "    ❌ 失败"
fi

echo ""
echo "1.4 读取特殊文件 /proc/version"
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/proc/version" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "Linux"; then
    echo "    ✅ 读取成功: ${RESPONSE:0:60}..."
else
    echo "    ❌ 失败"
fi

echo ""
echo "1.5 读取符号链接 /etc/mtab"
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/etc/mtab" -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "overlay"; then
    echo "    ✅ 符号链接解析成功"
    echo "    内容预览: ${RESPONSE:0:80}..."
else
    echo "    ❌ 失败: ${RESPONSE:0:100}"
fi

echo ""
echo "=== 2. 文件操作测试 ==="

echo ""
echo "2.1 创建测试目录"
curl -s -m 10 -X MKCOL "$WEBDAV_URL/tmp/webdav-test/" -H "Authorization: Bearer $TOKEN" > /dev/null
RESPONSE=$(curl -s -m 10 -X PROPFIND "$WEBDAV_URL/tmp/webdav-test/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    echo "    ✅ 创建目录成功"
else
    echo "    ❌ 创建目录失败"
fi

echo ""
echo "2.2 写入测试文件"
TEST_CONTENT="Hello WebDAV Test - $(date)"
curl -s -m 10 -X PUT "$WEBDAV_URL/tmp/webdav-test/test.txt" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: text/plain" \
    -d "$TEST_CONTENT" > /dev/null

echo ""
echo "2.3 读取测试文件"
RESPONSE=$(curl -s -m 10 "$WEBDAV_URL/tmp/webdav-test/test.txt" -H "Authorization: Bearer $TOKEN")
if [ "$RESPONSE" = "$TEST_CONTENT" ]; then
    echo "    ✅ 文件读写成功，内容一致"
else
    echo "    ❌ 内容不一致"
    echo "    期望: $TEST_CONTENT"
    echo "    实际: $RESPONSE"
fi

echo ""
echo "2.4 删除测试文件"
curl -s -m 10 -X DELETE "$WEBDAV_URL/tmp/webdav-test/test.txt" -H "Authorization: Bearer $TOKEN" > /dev/null
HTTP_CODE=$(curl -s -m 10 -o /dev/null -w "%{http_code}" "$WEBDAV_URL/tmp/webdav-test/test.txt" -H "Authorization: Bearer $TOKEN")
if [ "$HTTP_CODE" = "404" ]; then
    echo "    ✅ 删除文件成功"
else
    echo "    ❌ 删除失败，HTTP $HTTP_CODE"
fi

echo ""
echo "2.5 删除测试目录"
curl -s -m 10 -X DELETE "$WEBDAV_URL/tmp/webdav-test/" -H "Authorization: Bearer $TOKEN" > /dev/null
HTTP_CODE=$(curl -s -m 10 -o /dev/null -w "%{http_code}" -X PROPFIND "$WEBDAV_URL/tmp/webdav-test/" -H "Authorization: Bearer $TOKEN")
if [ "$HTTP_CODE" = "404" ]; then
    echo "    ✅ 删除目录成功"
else
    echo "    ❌ 删除失败"
fi

echo ""
echo "=== 3. 边界测试 ==="

echo ""
echo "3.1 读取不存在的文件"
HTTP_CODE=$(curl -s -m 10 -o /dev/null -w "%{http_code}" "$WEBDAV_URL/etc/nonexistent-file-12345" -H "Authorization: Bearer $TOKEN")
if [ "$HTTP_CODE" = "404" ]; then
    echo "    ✅ 正确返回 404"
else
    echo "    ❌ 返回 HTTP $HTTP_CODE"
fi

echo ""
echo "3.2 访问不允许的目录"
HTTP_CODE=$(curl -s -m 10 -o /dev/null -w "%{http_code}" "$WEBDAV_URL/root/" -H "Authorization: Bearer $TOKEN")
echo "    访问 /root/ 返回 HTTP $HTTP_CODE"

echo ""
echo "=== 4. 性能测试 ==="

echo ""
echo "4.1 连续 10 次读取同一文件"
START=$(date +%s%3N)
for i in {1..10}; do
    curl -s -m 10 "$WEBDAV_URL/etc/passwd" -H "Authorization: Bearer $TOKEN" > /dev/null
done
END=$(date +%s%3N)
TOTAL=$((END - START))
AVG=$((TOTAL / 10))
echo "    10 次读取总耗时: ${TOTAL}ms，平均: ${AVG}ms/次"

echo ""
echo "=========================================="
echo "               测试完成"
echo "=========================================="
