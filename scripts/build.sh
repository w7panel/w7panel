#!/bin/bash
# W7Panel 完整构建脚本

set -e

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="$BASE_DIR/dist"

echo "=== 开始构建 ==="

# 1. 清理构建产物
echo "[1/7] 清理构建产物..."
rm -rf "$DIST_DIR/kodata/assets"
rm -rf "$DIST_DIR/kodata/plugin/codeblitz"
rm -f "$DIST_DIR/w7panel"
mkdir -p "$DIST_DIR/kodata/plugin/codeblitz"

# 2. 构建后端
echo "[2/7] 构建后端..."
cd "$BASE_DIR/w7panel"
export PATH=/opt/tools/go/bin:$PATH
export GOPROXY=https://goproxy.cn,direct
CGO_CFLAGS="-Wno-return-local-address" go build -o "$DIST_DIR/w7panel" .

# 3. 构建前端
echo "[3/7] 构建前端..."
cd "$BASE_DIR/w7panel-ui"
npm run build

# 4. 先复制后端静态资源（logo.png, k3s-*.sh, ip2region 等）
echo "[4/7] 复制后端静态资源..."
cp -r "$BASE_DIR/w7panel/kodata/"* "$DIST_DIR/kodata/"

# 5. 再复制前端资源（确保前端资源覆盖后端同名文件）
echo "[5/7] 复制前端资源..."
cp -r "$BASE_DIR/w7panel-ui/dist/"* "$DIST_DIR/kodata/"

# 6. 构建编辑器
echo "[6/7] 构建编辑器..."
cd "$BASE_DIR/codeblitz"
npm run build

# 7. 复制编辑器资源
echo "[7/7] 复制编辑器资源..."
cp -r "$BASE_DIR/codeblitz/dist/"* "$DIST_DIR/kodata/plugin/codeblitz/"
cp "$BASE_DIR/codeblitz/node_modules/@codeblitzjs/ide-core/bundle/"*.wasm "$DIST_DIR/kodata/plugin/codeblitz/"

# 8. 复制启动脚本
echo "[8/8] 复制启动脚本..."
cp "$BASE_DIR/w7panel/scripts/w7panel-ctl.sh" "$DIST_DIR/w7panel-ctl.sh"
chmod +x "$DIST_DIR/w7panel-ctl.sh"

echo "=== 构建完成 ==="
echo "启动命令: cd $DIST_DIR && ./w7panel-ctl.sh start"
