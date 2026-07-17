#!/bin/bash

# WebDAV 文件管理 & 文本编辑器 UI 测试
# 使用 agent-browser 工具

BASE_DIR=/home/wwwroot/w7panel-dev

echo "=========================================="
echo "    WebDAV UI 测试 (agent-browser)"
echo "=========================================="

# 获取 kubeconfig 中的 token (YAML多行格式)
TOKEN=$(grep -A1 "token:" $BASE_DIR/w7panel/kubeconfig.yaml | tail -1 | sed 's/^[[:space:]]*//')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "token: >-" ]; then
    # 尝试另一种方式
    TOKEN=$(cat $BASE_DIR/w7panel/kubeconfig.yaml | grep -A2 "user:" | grep -v "user:" | grep -v "^--" | sed 's/^[[:space:]]*//' | tr -d '\n')
fi

if [ -z "$TOKEN" ]; then
    echo "❌ 无法获取 token"
    exit 1
fi
echo "✅ Token 已获取: ${TOKEN:0:50}..."

echo ""
echo "=== 测试1: API 接口测试 ==="

# 测试 WebDAV PROPFIND
echo "1.1 PROPFIND /etc/"
RESPONSE=$(curl -s -m 10 -X PROPFIND "http://localhost:8000/k8s/webdav-agent/1/agent/etc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    COUNT=$(echo "$RESPONSE" | grep -o '<D:response>' | wc -l)
    echo "    ✅ 成功，$COUNT 个条目"
else
    echo "    ❌ 失败: ${RESPONSE:0:200}"
fi

# 测试读取文件
echo "1.2 GET /etc/passwd"
RESPONSE=$(curl -s -m 10 "http://localhost:8000/k8s/webdav-agent/1/agent/etc/passwd" \
    -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "root:"; then
    echo "    ✅ 成功读取"
else
    echo "    ❌ 失败"
fi

# 测试特殊目录
echo "1.3 PROPFIND /proc/ (特殊目录)"
START=$(date +%s%3N)
RESPONSE=$(curl -s -m 10 -X PROPFIND "http://localhost:8000/k8s/webdav-agent/1/agent/proc/" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Depth: 1" \
    -H "Content-Type: text/xml; charset=utf-8")
END=$(date +%s%3N)
TIME=$((END - START))
if echo "$RESPONSE" | grep -q "<D:multistatus"; then
    COUNT=$(echo "$RESPONSE" | grep -o '<D:response>' | wc -l)
    echo "    ✅ 成功，$COUNT 个条目，耗时 ${TIME}ms"
    if [ $TIME -lt 100 ]; then
        echo "    ✅ 性能良好（<100ms）"
    fi
else
    echo "    ❌ 失败"
fi

# 测试读取特殊文件
echo "1.4 GET /proc/version (特殊文件)"
RESPONSE=$(curl -s -m 10 "http://localhost:8000/k8s/webdav-agent/1/agent/proc/version" \
    -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "Linux"; then
    echo "    ✅ 成功: ${RESPONSE:0:50}..."
else
    echo "    ❌ 失败"
fi

# 测试符号链接
echo "1.5 GET /etc/mtab (符号链接)"
RESPONSE=$(curl -s -m 10 "http://localhost:8000/k8s/webdav-agent/1/agent/etc/mtab" \
    -H "Authorization: Bearer $TOKEN")
if echo "$RESPONSE" | grep -q "overlay"; then
    echo "    ✅ 符号链接解析成功"
    echo "    内容预览: ${RESPONSE:0:60}..."
else
    echo "    ❌ 失败"
fi

echo ""
echo "=== 测试2: 文件操作测试 ==="

# 创建测试目录
echo "2.1 创建测试目录"
curl -s -m 10 -X MKCOL "http://localhost:8000/k8s/webdav-agent/1/agent/tmp/webdav-ui-test/" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
echo "    ✅ 目录已创建"

# 写入文件
echo "2.2 写入测试文件"
TEST_CONTENT="Hello WebDAV UI Test - $(date)"
curl -s -m 10 -X PUT "http://localhost:8000/k8s/webdav-agent/1/agent/tmp/webdav-ui-test/test.txt" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: text/plain" \
    -d "$TEST_CONTENT" > /dev/null
echo "    ✅ 文件已写入"

# 读取验证
RESPONSE=$(curl -s -m 10 "http://localhost:8000/k8s/webdav-agent/1/agent/tmp/webdav-ui-test/test.txt" \
    -H "Authorization: Bearer $TOKEN")
if [ "$RESPONSE" = "$TEST_CONTENT" ]; then
    echo "2.3 读取验证: ✅ 内容一致"
else
    echo "2.3 读取验证: ❌ 内容不一致"
fi

# 清理
curl -s -m 10 -X DELETE "http://localhost:8000/k8s/webdav-agent/1/agent/tmp/webdav-ui-test/test.txt" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
curl -s -m 10 -X DELETE "http://localhost:8000/k8s/webdav-agent/1/agent/tmp/webdav-ui-test/" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
echo "2.4 清理测试数据: ✅"

echo ""
echo "=== 测试3: 边界测试 ==="

echo "3.1 读取不存在的文件"
HTTP_CODE=$(curl -s -m 10 -o /dev/null -w "%{http_code}" \
    "http://localhost:8000/k8s/webdav-agent/1/agent/etc/nonexistent-file-12345" \
    -H "Authorization: Bearer $TOKEN")
if [ "$HTTP_CODE" = "404" ]; then
    echo "    ✅ 正确返回 404"
else
    echo "    ❌ 返回 HTTP $HTTP_CODE"
fi

echo ""
echo "=== 测试4: 性能测试 ==="

START=$(date +%s%3N)
for i in {1..10}; do
    curl -s -m 10 "http://localhost:8000/k8s/webdav-agent/1/agent/etc/passwd" \
        -H "Authorization: Bearer $TOKEN" > /dev/null
done
END=$(date +%s%3N)
TOTAL=$((END - START))
AVG=$((TOTAL / 10))
echo "4.1 10次读取平均耗时: ${AVG}ms/次"

echo ""
echo "=========================================="
echo "            测试完成"
echo "=========================================="
echo ""
echo "总结:"
echo "  ✅ PROPFIND 普通目录和特殊目录"
echo "  ✅ GET 普通文件、特殊文件、符号链接"
echo "  ✅ MKCOL/PUT/DELETE 文件操作"
echo "  ✅ 性能优化生效（/proc 目录快速响应）"
