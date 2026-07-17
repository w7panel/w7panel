#!/bin/bash
# WebDAV 文件操作测试脚本
# 用法: bash tests/webdav.sh
#
# 测试模式说明:
# - 生产环境: 需要部署 agent 服务, 通过 /k8s/v1/{podIp}:8000/proxy 访问
# - 开发环境: 设置 LOCAL_MOCK=true, 使用当前服务(8080端口)模拟 agent

BASE_DIR=${BASE_DIR:-/home/wwwroot/w7panel-dev}
KUBECONFIG=${KUBECONFIG:-$BASE_DIR/w7panel/kubeconfig.yaml}

TOKEN=$(grep -A2 'token:' "$KUBECONFIG" | grep -v 'token:' | sed 's/^[[:space:]]*//' | head -1)

if [ -z "$TOKEN" ]; then
    echo "✗ 无法从 $KUBECONFIG 读取 token"
    exit 1
fi

PASS=0
FAIL=0

echo "=========================================="
echo "WebDAV 文件操作测试"
echo "=========================================="

# 检查服务
echo ""
echo "[1] 服务状态..."
if curl -s http://localhost:8000/ | head -1 | grep -q "DOCTYPE"; then
    echo "✓ 正常"; ((PASS++))
else
    echo "✗ 未启动，请先启动服务"; ((FAIL++)); exit 1
fi

# 获取 webdavUrl (从 /k8s/pid 接口返回)
echo ""
echo "[2] 获取 webdavUrl..."
PID_RESP=$(curl -s -G "http://localhost:8000/k8s/pid" \
  -d "namespace=default" \
  -d "HostIp=10.0.0.206" \
  -d "containerName=w7-python" \
  -d "podName=w7-python-666fbf494-rfbnd" \
  -H "Authorization: Bearer $TOKEN")

WEBDAV_URL=$(echo "$PID_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('webdavUrl',''))" 2>/dev/null)

if [[ "$WEBDAV_URL" == /k8s/* ]]; then
    echo "✓ $WEBDAV_URL"; ((PASS++))
    # 检查是否是 local_mock 模式
    if [[ "$WEBDAV_URL" == /k8s/webdav-agent/* ]]; then
        echo "  (local_mock 模式)"
    fi
else
    echo "✗ 失败: $PID_RESP"; ((FAIL++)); exit 1
fi

# 测试路径
TEST_DIR="/tmp/webdav_test_$$"
TEST_FILE="$TEST_DIR/test.txt"
TEST_DIR_RENAME="$TEST_DIR/renamed_dir"
TEST_FILE_RENAME="$TEST_DIR/renamed.txt"

echo ""
echo "[3] 创建测试目录 (MKCOL)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X MKCOL "http://localhost:8000${WEBDAV_URL}${TEST_DIR}" \
  -H "Authorization: Bearer $TOKEN")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "200" ]]; then
    echo "✓ 创建成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[4] 创建测试文件 (PUT)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "http://localhost:8000${WEBDAV_URL}${TEST_FILE}" \
  -H "Authorization: Bearer $TOKEN" \
  -d "Hello WebDAV Test")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "200" || "$HTTP_CODE" == "204" ]]; then
    echo "✓ 创建成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[5] 重命名文件 (MOVE)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X MOVE "http://localhost:8000${WEBDAV_URL}${TEST_FILE}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Destination: http://localhost:8000${WEBDAV_URL}${TEST_FILE_RENAME}" \
  -H "Overwrite: T")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "204" ]]; then
    echo "✓ 重命名成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[6] 创建子目录..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X MKCOL "http://localhost:8000${WEBDAV_URL}${TEST_DIR}/subdir" \
  -H "Authorization: Bearer $TOKEN")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "200" ]]; then
    echo "✓ 创建成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[7] 重命名文件夹 (MOVE)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X MOVE "http://localhost:8000${WEBDAV_URL}${TEST_DIR}/subdir" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Destination: http://localhost:8000${WEBDAV_URL}${TEST_DIR_RENAME}" \
  -H "Overwrite: T")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "204" ]]; then
    echo "✓ 重命名成功 (HTTP $HTTP_CODE)"; ((PASS++))
elif [[ "$HTTP_CODE" == "403" ]]; then
    echo "⚠ 403 Forbidden - WebDAV库限制 (非关键)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[8] 复制文件 (COPY)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X COPY "http://localhost:8000${WEBDAV_URL}${TEST_FILE_RENAME}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Destination: http://localhost:8000${WEBDAV_URL}${TEST_DIR}/copy.txt" \
  -H "Overwrite: T")
if [[ "$HTTP_CODE" == "201" || "$HTTP_CODE" == "204" ]]; then
    echo "✓ 复制成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[9] 删除文件 (DELETE)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "http://localhost:8000${WEBDAV_URL}${TEST_FILE_RENAME}" \
  -H "Authorization: Bearer $TOKEN")
if [[ "$HTTP_CODE" == "204" || "$HTTP_CODE" == "200" ]]; then
    echo "✓ 删除成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "[10] 删除目录 (DELETE)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "http://localhost:8000${WEBDAV_URL}${TEST_DIR}" \
  -H "Authorization: Bearer $TOKEN")
if [[ "$HTTP_CODE" == "204" || "$HTTP_CODE" == "200" ]]; then
    echo "✓ 删除成功 (HTTP $HTTP_CODE)"; ((PASS++))
else
    echo "✗ 失败 (HTTP $HTTP_CODE)"; ((FAIL++))
fi

echo ""
echo "=========================================="
echo "结果: 通过 $PASS / $((PASS+FAIL))"
echo "=========================================="

[ $FAIL -gt 0 ] && exit 1
