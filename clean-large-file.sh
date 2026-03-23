#!/bin/bash
# 从 git 历史中彻底删除大文件 w7panel

set -e

echo "=== 从 git 历史中删除 w7panel 大文件 ==="

# 方法 1: 使用 git filter-branch
echo "方法 1: 使用 git filter-branch..."
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch w7panel' \
  --prune-empty --tag-name-filter cat -- --all

echo ""
echo "=== 清理引用 ==="
# 清理备份引用
rm -rf .git/refs/original/
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
