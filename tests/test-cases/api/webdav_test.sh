#!/bin/bash
# WebDAV 随机测试脚本

BASE_DIR=${BASE_DIR:-/home/wwwroot/w7panel-dev}
KUBECONFIG=${KUBECONFIG:-$BASE_DIR/w7panel/kubeconfig.yaml}
TOKEN=$(grep -A2 'token:' "$KUBECONFIG" | grep -v 'token:' | sed 's/^[[:space:]]*//' | head -1)

PASS=0
FAIL=0
RESULTS=()

test_move() {
    local pid=$1
    local path=$2
    local new_name=$(basename "$path")_renamed_$$
    local dir=$(dirname "$path")
    
    # 检查源是否存在
    if [ ! -e "/host/proc/$pid/root$path" ]; then
        echo "SKIP: $path (not exist)"
        return
    fi
    
    # 测试 MOVE
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X MOVE \
        "http://localhost:8000/k8s/webdav-agent/$pid/agent$path" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Destination: /k8s/webdav-agent/$pid/agent$dir/$new_name" \
        -H "Overwrite: T")
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "204" ]; then
        echo "✓ MOVE $path -> $new_name (HTTP $HTTP_CODE)"
        # 恢复
        mv "/host/proc/$pid/root$dir/$new_name" "/host/proc/$pid/root$path" 2>/dev/null
        ((PASS++))
        RESULTS+=("PASS: MOVE $path")
    else
        echo "✗ MOVE $path (HTTP $HTTP_CODE)"
        ((FAIL++))
        RESULTS+=("FAIL: MOVE $path (HTTP $HTTP_CODE)")
    fi
}

test_copy() {
    local pid=$1
    local path=$2
    local new_name=$(basename "$path")_copy_$$
    local dir=$(dirname "$path")
    
    # 检查源是否存在
    if [ ! -e "/host/proc/$pid/root$path" ]; then
        echo "SKIP: $path (not exist)"
        return
    fi
    
    # 测试 COPY
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X COPY \
        "http://localhost:8000/k8s/webdav-agent/$pid/agent$path" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Destination: /k8s/webdav-agent/$pid/agent$dir/$new_name" \
        -H "Overwrite: T")
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "204" ]; then
        echo "✓ COPY $path -> $new_name (HTTP $HTTP_CODE)"
        # 清理
        rm -rf "/host/proc/$pid/root$dir/$new_name" 2>/dev/null
        ((PASS++))
        RESULTS+=("PASS: COPY $path")
    else
        echo "✗ COPY $path (HTTP $HTTP_CODE)"
        # 清理可能创建的部分文件
        rm -rf "/host/proc/$pid/root$dir/$new_name" 2>/dev/null
        ((FAIL++))
        RESULTS+=("FAIL: COPY $path (HTTP $HTTP_CODE)")
    fi
}

echo "=========================================="
echo "WebDAV 随机测试"
echo "=========================================="

# 获取可用的 pid 列表
PIDS=$(ls /host/proc | grep -E '^[0-9]+$' | head -5)
echo "可用 PID: $PIDS"
echo ""

# 测试各种路径
for pid in $PIDS; do
    echo "--- 测试 pid=$pid ---"
    
    # 测试目录
    for dir in /etc /var /tmp /var/www /var/log; do
        if [ -d "/host/proc/$pid/root$dir" ]; then
            # 找一个子目录或文件测试
            first_item=$(ls "/host/proc/$pid/root$dir" 2>/dev/null | head -1)
            if [ -n "$first_item" ]; then
                test_move $pid "$dir/$first_item"
                test_copy $pid "$dir/$first_item"
            fi
        fi
    done
done

echo ""
echo "=========================================="
echo "结果: 通过 $PASS / $((PASS+FAIL))"
echo "=========================================="

if [ $FAIL -gt 0 ]; then
    echo ""
    echo "失败详情:"
    for r in "${RESULTS[@]}"; do
        if [[ "$r" == FAIL* ]]; then
            echo "  $r"
        fi
    done
fi
