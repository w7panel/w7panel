#!/bin/bash
# 使用 BFG Repo-Cleaner 清理大文件（推荐方法）

set -e

echo "=== 使用 BFG Repo-Cleaner 删除大文件 ==="
echo ""

# 检查是否安装了 BFG
if ! command -v bfg &> /dev/null; then
    echo "❌ BFG 未安装，请先安装 BFG Repo-Cleaner："
    echo ""
    echo "方法 1 - 使用 Homebrew (macOS):"
    echo "  brew install bfg"
    echo ""
    echo "方法 2 - 下载 JAR 文件:"
    echo "  wget https://repo1.maven.org/maven2/com/madgag/bfg/1.14.2/bfg-1.14.2.jar"
    echo "  alias bfg='java -jar bfg-1.14.2.jar'"
    echo ""
    echo "方法 3 - 使用 apt (Ubuntu/Debian):"
    echo "  sudo apt install bfg"
    echo ""
    exit 1
fi

echo "开始清理..."
bfg --delete-files w7panel --no-blob-protection .

echo ""
echo "=== 清理引用 ==="
git reflog expire --expire=now --all
git gc --prune=now --aggressive

echo ""
echo "=== 验证结果 ==="
git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | awk '/^blob/ && $3 > 104857600 {print $3/1024/1024" MB", $4}'

echo ""
echo "如果上面没有输出，说明大文件已被成功删除！"
echo ""
echo "=== 下一步操作 ==="
echo "1. 检查本地仓库大小：du -sh .git"
echo "2. 强制推送到远程：git push --force --all"
echo "3. 如果有 tag，也需要推送：git push --force --tags"
echo ""
echo "⚠️  警告：强制推送会改写历史，请确保团队成员知晓！"
