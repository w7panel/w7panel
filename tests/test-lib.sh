#!/bin/bash
# W7Panel 测试库
# 提供通用的测试工具方法

# ⚠️ 测试前必须先读取菜单地图
#    了解页面结构、菜单层级、路由规则
#    cat docs/testing/ui/ui-menu-map.md
#    禁止直接开始测试，必须先规划测试路径！

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-123456}"
TOKEN=""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志方法
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

#========================================
# 检查服务状态
#========================================
check_service() {
    local code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" 2>/dev/null)
    if [ "$code" = "200" ]; then
        return 0
    else
        log_fail "服务未运行 (HTTP $code)"
        return 1
    fi
}

#========================================
# 检查验证码配置
# 返回: 0=已禁用, 1=未禁用
#========================================
check_captcha_disabled() {
    local resp=$(curl -s "$BASE_URL/k8s/init-user" 2>/dev/null)
    if echo "$resp" | grep -q '"captchaEnabled":"false"'; then
        return 0
    else
        return 1
    fi
}

#========================================
# API登录 - 获取Token
# 返回: Token字符串
#========================================
api_login() {
    local resp=$(curl -s -X POST "$BASE_URL/k8s/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" 2>/dev/null)
    
    if echo "$resp" | grep -q '"token"'; then
        TOKEN=$(echo "$resp" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "$TOKEN"
        return 0
    else
        echo ""
        return 1
    fi
}

#========================================
# UI登录 - 通过浏览器登录后台
# 参数:
#   $1 - 是否截图 (可选, 默认不截图)
# 返回: 0=成功, 1=失败
#========================================
ui_login() {
    local screenshot="${1:-false}"
    
    log_step "UI登录后台 ($USERNAME)"
    
    # 关闭已有浏览器
    agent-browser close 2>/dev/null || true
    sleep 1
    
    # 打开首页
    agent-browser open "$BASE_URL/" 2>&1 | grep -v "^$" || true
    sleep 5
    
    # 检查登录表单
    local snapshot=$(agent-browser snapshot -i 2>&1)
    if ! echo "$snapshot" | grep -q "用户名"; then
        log_fail "登录表单未显示"
        return 1
    fi
    
    # 输入用户名密码并登录
    agent-browser type e1 "$USERNAME" 2>&1 | grep -v "^$" || true
    agent-browser type e2 "$PASSWORD" 2>&1 | grep -v "^$" || true
    agent-browser click e4 2>&1 | grep -v "^$" || true
    sleep 5
    
    # 截图
    if [ "$screenshot" = "true" ]; then
        agent-browser screenshot /tmp/login-result.png 2>&1 | grep -v "^$" || true
    fi
    
    # 验证登录成功
    snapshot=$(agent-browser snapshot 2>&1)
    if echo "$snapshot" | grep -qE "系统管理|云主机"; then
        log_pass "UI登录成功"
        return 0
    else
        log_fail "UI登录失败"
        agent-browser screenshot /tmp/login-fail.png 2>&1 | grep -v "^$" || true
        return 1
    fi
}

#========================================
# 进入文件管理页面
# 参数:
#   $1 - 应用名称 (可选, 默认选择第一个)
# 返回: 0=成功, 1=失败
#========================================
enter_file_manager() {
    local app_name="${1:-}"
    
    log_step "进入文件管理页面"
    
    # 导航到应用列表
    agent-browser open "$BASE_URL/app/apps" 2>&1 | grep -v "^$" || true
    sleep 5
    
    # 点击应用列表菜单
    agent-browser eval "
    (function() {
        const items = document.querySelectorAll('[class*=menu]');
        for (const item of items) {
            if (item.textContent.trim() === '应用列表') {
                item.click();
                return true;
            }
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 3
    
    # 点击文件管理
    agent-browser eval "
    (function() {
        const cells = document.querySelectorAll('td');
        for (const cell of cells) {
            const spans = cell.querySelectorAll('span.operation, span.c-blue');
            for (const span of spans) {
                if (span.textContent.trim() === '文件管理') {
                    span.click();
                    return true;
                }
            }
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 5
    
    # 验证是否进入文件管理
    local snapshot=$(agent-browser snapshot 2>&1)
    if echo "$snapshot" | grep -q "文件管理"; then
        log_pass "进入文件管理页面成功"
        return 0
    else
        log_fail "进入文件管理页面失败"
        return 1
    fi
}

#========================================
# 选中文件/文件夹
# 参数:
#   $1 - 文件/文件夹名称
# 返回: 0=成功, 1=失败
#========================================
select_file() {
    local name="$1"
    
    agent-browser eval "
    (function() {
        const rows = document.querySelectorAll('tr');
        for (const row of rows) {
            if (row.textContent.includes('$name')) {
                const checkbox = row.querySelector('input[type=checkbox]');
                if (checkbox) {
                    checkbox.click();
                    return true;
                }
            }
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 1
}

#========================================
# 点击按钮
# 参数:
#   $1 - 按钮文本
# 返回: 0=成功, 1=失败
#========================================
click_button() {
    local text="$1"
    
    agent-browser eval "
    (function() {
        const buttons = document.querySelectorAll('button');
        for (const btn of buttons) {
            if (btn.textContent.trim() === '$text') {
                btn.click();
                return true;
            }
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 1
}

#========================================
# 获取输入框值
# 参数:
#   $1 - 包含的文本 (用于定位)
#========================================
get_input_value() {
    local contains="$1"
    
    agent-browser eval "
    (function() {
        const inputs = document.querySelectorAll('input');
        for (const input of inputs) {
            if (input.value && input.value.includes('$contains')) {
                return input.value;
            }
        }
        return '';
    })()
    " 2>&1
}

#========================================
# 选择下拉选项
# 参数:
#   $1 - 选项文本 (部分匹配)
#========================================
select_option() {
    local text="$1"
    
    # 打开下拉框
    agent-browser eval "
    (function() {
        const selects = document.querySelectorAll('.arco-select-view');
        for (const sel of selects) {
            sel.click();
            return true;
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 1
    
    # 选择选项
    agent-browser eval "
    (function() {
        const options = document.querySelectorAll('li');
        for (const opt of options) {
            if (opt.textContent.includes('$text')) {
                opt.click();
                return true;
            }
        }
        return false;
    })()
    " 2>&1 | grep -v "^$" || true
    sleep 1
}

#========================================
# 关闭浏览器
#========================================
close_browser() {
    agent-browser close 2>/dev/null || true
}

#========================================
# 截图
# 参数:
#   $1 - 文件路径
#========================================
take_screenshot() {
    local path="$1"
    agent-browser screenshot "$path" 2>&1 | grep -v "^$" || true
}

 # 如果直接执行此脚本，显示帮助
 if [ "${BASH_SOURCE[0]}" = "$0" ]; then
     echo "W7Panel 测试库"
     echo ""
     echo "可用方法:"
     echo "  check_service          - 检查服务状态"
     echo "  check_captcha_disabled - 检查验证码是否禁用"
     echo "  api_login              - API登录获取Token"
     echo "  ui_login [screenshot]  - UI登录后台"
     echo "  enter_file_manager     - 进入文件管理"
     echo "  select_file <name>     - 选中文件"
     echo "  click_button <text>    - 点击按钮"
     echo "  select_option <text>   - 选择下拉选项"
     echo "  get_input_value <text> - 获取输入框值"
     echo "  close_browser          - 关闭浏览器"
     echo "  take_screenshot <path> - 截图"
     echo ""
     echo "测试规则:"
     echo "  1. 验证文件/目录是否真实存在"
     echo "  2. 注意目录下可能有同名文件（如 /etc/etc）"
     echo "  3. 特殊目录（/proc, /sys）可能不支持WebDAV浏览"
     echo ""
     echo "用法: source test-lib.sh"
 fi
 
 #========================================
 # 验证文件/目录是否真实存在（通过WebDAV）
 # 参数:
 #   $1 - 文件/目录名称
 #   $2 - 父目录路径（可选，默认为当前目录）
 # 返回: 0=存在, 1=不存在
 #========================================
 verify_file_exists() {
     local name="$1"
     local parent_path="${2:-}"
     
     if [ -z "$TOKEN" ]; then
         api_login > /dev/null
     fi
     
     local url="${BASE_URL}/k8s/webdav-agent/1/agent${parent_path}/${name}"
     local code=$(curl -s -o /dev/null -w "%{http_code}" \
         -X PROPFIND "$url" \
         -H "Authorization: Bearer $TOKEN" \
         -H "Depth: 0" 2>/dev/null)
     
     if [ "$code" = "207" ] || [ "$code" = "200" ]; then
         return 0
     else
         return 1
     fi
 }
 
 #========================================
 # 检查文件列表是否包含目录本身
 # 参数:
 #   $1 - 目录名称
 #   $2 - 父目录路径
 # 返回: 0=包含(错误), 1=不包含(正确)
 #========================================
 check_dir_not_self_listed() {
     local dirname="$1"
     local parent_path="$2"
     
     if [ -z "$TOKEN" ]; then
         api_login > /dev/null
     fi
     
     # 获取目录列表
     local response=$(curl -s -X PROPFIND \
         "${BASE_URL}/k8s/webdav-agent/1/agent${parent_path}/${dirname}/" \
         -H "Authorization: Bearer $TOKEN" \
         -H "Depth: 1" 2>/dev/null)
     
     # 检查是否只包含目录本身
     local count=$(echo "$response" | grep -o "<D:displayname>" | wc -l)
     local self_count=$(echo "$response" | grep -o "<D:displayname>${dirname}</D:displayname>" | wc -l)
     
     # 如果只有1个displayname且是目录本身，则返回错误
     if [ "$count" -eq 1 ] && [ "$self_count" -eq 1 ]; then
         return 0  # 包含自身，错误
     else
         return 1  # 不包含自身，正确
     fi
 }
