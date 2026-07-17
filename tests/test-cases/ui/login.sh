#!/bin/bash
#========================================
# 测试用例: UI登录测试
#========================================
#
# ## 测试信息
# | 项目 | 内容 |
# |------|------|
# | 类型 | UI |
# | 优先级 | P0 |
# | 用途 | 验证登录功能 |
#
# ## 环境变量
# BASE_URL - 服务地址
# USERNAME - 用户名
# PASSWORD - 密码
#
#========================================

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-123456}"
WAIT_TIME=5

# 检查agent-browser
if ! command -v agent-browser &> /dev/null; then
    echo "错误: agent-browser 未安装"
    exit 1
fi

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

echo "=== 测试: UI登录 ==="

# 打开登录页
log_info "打开登录页: $BASE_URL"
agent-browser open "$BASE_URL"
sleep $WAIT_TIME

# 强制检查
log_info "检查控制台..."
agent-browser console | grep -qi "error" && { log_fail "控制台有错误"; agent-browser close; exit 1; }
log_pass "控制台无错误"

log_info "检查JS错误..."
agent-browser errors && log_pass "无JS错误" || log_fail "有JS错误"

# 获取交互元素
log_info "获取表单元素..."
agent-browser snapshot -i > /dev/null 2>&1 || true

# 填写登录信息
log_info "填写用户名: $USERNAME"
agent-browser fill @e1 "$USERNAME" > /dev/null 2>&1 || { log_fail "填写用户名失败"; agent-browser close; exit 1; }

log_info "填写密码: ********"
agent-browser fill @e2 "$PASSWORD" > /dev/null 2>&1 || { log_fail "填写密码失败"; agent-browser close; exit 1; }

log_info "点击登录按钮..."
agent-browser click @e3 > /dev/null 2>&1 || { log_fail "点击登录失败"; agent-browser close; exit 1; }

sleep 3

# 验证登录结果
log_info "验证登录结果..."
URL=$(agent-browser get url 2>/dev/null || echo "")

if echo "$URL" | grep -q "login"; then
    log_fail "登录失败，停留在登录页"
    agent-browser close
    exit 1
else
    log_pass "登录成功，跳转到: $URL"
fi

agent-browser close
log_pass "所有测试通过"
