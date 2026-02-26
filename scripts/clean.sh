#!/bin/bash
# 清理构建产物脚本

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="$BASE_DIR/dist"

echo "=== 清理构建产物 ==="

# 清理前端资源
rm -rf "$DIST_DIR/kodata/assets"
rm -rf "$DIST_DIR/kodata/plugin/codeblitz"

# 清理后端二进制
rm -f "$DIST_DIR/w7panel"

echo "清理完成"
