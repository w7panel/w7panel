#!/bin/bash

echo "=== 前端性能优化测试 ==="
echo ""

# 检查修改的文件是否存在
echo "1. 检查修改的文件..."
FILES=(
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/hooks/request.ts"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/hooks/timer.ts"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/store/modules/namespace.ts"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/store/modules/user/index.ts"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/api/interceptor.ts"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/views/app/apps/index.vue"
  "/home/wwwroot/w7panel-dev/w7panel-ui/src/components/yaml-input.vue"
)

for file in "${FILES[@]}"; do
  if [ -f "$file" ]; then
    echo "  ✓ $file"
  else
    echo "  ✗ $file (不存在)"
  fi
done

echo ""
echo "2. 检查构建产物..."
if [ -f "/home/wwwroot/w7panel-dev/dist/kodata/index.html" ]; then
  echo "  ✓ 前端构建成功"
else
  echo "  ✗ 前端构建失败"
fi

if [ -f "/home/wwwroot/w7panel-dev/dist/w7panel" ]; then
  echo "  ✓ 后端构建成功"
else
  echo "  ✗ 后端构建失败"
fi

echo ""
echo "3. 测试服务..."
if curl -s http://localhost:8000/ | grep -q "html"; then
  echo "  ✓ 服务正常运行"
else
  echo "  ✗ 服务未运行"
fi

echo ""
echo "4. 测试 API..."
TOKEN=$(cat /home/wwwroot/w7panel-dev/kubeconfig.yaml | grep -A1 "token:" | tail -1 | awk '{print $2}')

# 测试 namespace API
RESULT=$(curl -s "http://localhost:8000/api/v1/namespaces" -H "Authorization: Bearer $TOKEN")
if echo "$RESULT" | grep -q "items"; then
  echo "  ✓ Namespace API 正常"
else
  echo "  ✗ Namespace API 失败"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "优化内容总结："
echo "1. useRequest hook - 添加了缓存、取消、重试、超时机制"
echo "2. timer.ts - 新的定时器管理 composable"
echo "3. namespace store - 添加了缓存机制"
echo "4. user store - 优化了登录时的并行请求"
echo "5. interceptor.ts - 添加了全局默认超时"
echo "6. apps/index.vue - 批量请求替代串行请求"
echo "7. yaml-input.vue - 修复了定时器泄漏"
