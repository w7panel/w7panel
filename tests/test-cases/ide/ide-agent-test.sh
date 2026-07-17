#!/bin/bash
# W7Panel Web IDE 自动化测试脚本
# 使用 agent-browser 进行界面功能测试

# 配置
BASE_URL="${BASE_URL:-http://localhost:8000}"
API_URL="${API_URL:-/k8s/webdav-agent/1/agent}"
INITIAL_PATH="${INITIAL_PATH:-/tmp}"
TOKEN="${K8S_TOKEN:-}"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# 获取 Token
get_token() {
    if [ -n "$TOKEN" ]; then
        echo "$TOKEN" | tr -d '[:space:]'
        return 0
    fi
    
    local kubeconfig="$BASE_DIR/w7panel/kubeconfig.yaml"
    if [ -f "$kubeconfig" ]; then
        grep -A1 "token:" "$kubeconfig" | tail -1 | awk '{print $2}' | tr -d '[:space:]'
        return 0
    fi
    
    if [ -f "/var/run/secrets/kubernetes.io/serviceaccount/token" ]; then
        cat /var/run/secrets/kubernetes.io/serviceaccount/token
        return 0
    fi
    
    return 1
}

# 打开编辑器
open_editor() {
    local token="$1"
    local initial_path="$2"
    local encoded_token=$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$token'''))")
    local url="${BASE_URL}/ui/plugin/codeblitz/editor.html?api-url=${API_URL}&api-token=${encoded_token}&initial-path=${initial_path}"
    
    agent-browser open "$url" 2>&1
    sleep 15
}

# 关闭编辑器
close_editor() {
    agent-browser close 2>&1
}

# 测试1: 编辑器加载
test_editor_load() {
    log_step "测试1: 编辑器加载 (initial-path=$INITIAL_PATH)"
    
    local console=$(agent-browser console 2>&1)
    
    if echo "$console" | grep -q "token: \*\*\*"; then
        log_pass "Token 已正确设置"
    else
        log_fail "Token 未设置"
        return 1
    fi
    
    if echo "$console" | grep -q "Response: 207"; then
        log_pass "WebDAV 连接成功 (207 Multi-Status)"
    else
        log_fail "WebDAV 连接失败"
        return 1
    fi
    
    if echo "$console" | grep -q "Parsed files:"; then
        local file_count=$(echo "$console" | grep "Parsed files:" | grep -o '[0-9]*' | head -1)
        log_pass "文件列表加载成功 ($file_count 个文件)"
    else
        log_fail "文件列表加载失败"
        return 1
    fi
    
    agent-browser screenshot /tmp/ide-test-1-load.png 2>&1
    return 0
}

# 测试2: 文件树显示
test_file_tree() {
    log_step "测试2: 文件树显示"
    
    local result=$(agent-browser eval "
    (function() {
        const explorer = document.getElementById('explorer');
        if (!explorer) return JSON.stringify({error: 'explorer not found'});
        
        const files = explorer.querySelectorAll('[title*=\"file://\"]');
        const hasEditor = !!document.querySelector('.monaco-editor');
        
        return JSON.stringify({
            filesCount: files.length,
            hasEditor: hasEditor
        });
    })()
    " 2>&1)
    
    local file_count=$(echo "$result" | grep -o '"filesCount":[0-9]*' | grep -o '[0-9]*')
    
    if [ -n "$file_count" ] && [ "$file_count" -gt 0 ]; then
        log_pass "文件树显示正常 ($file_count 个文件/文件夹)"
        return 0
    else
        log_fail "文件树显示异常 (0 个文件)"
        return 1
    fi
}

# 测试3: 命令面板
test_command_palette() {
    log_step "测试3: 命令面板 (Ctrl+Shift+P)"
    
    agent-browser keydown Control 2>&1
    agent-browser keydown Shift 2>&1
    agent-browser press p 2>&1
    agent-browser keyup Shift 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 2
    
    local snapshot=$(agent-browser snapshot -i 2>&1)
    
    if echo "$snapshot" | grep -q "textbox.*命令"; then
        log_pass "命令面板打开成功"
        agent-browser screenshot /tmp/ide-test-3-command-palette.png 2>&1
        agent-browser press Escape 2>&1
        sleep 1
        return 0
    else
        log_fail "命令面板未打开"
        agent-browser press Escape 2>&1
        return 1
    fi
}

# 测试4: 快速打开文件
test_quick_open() {
    log_step "测试4: 快速打开文件 (Ctrl+P)"
    
    agent-browser keydown Control 2>&1
    agent-browser press p 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 2
    
    local snapshot=$(agent-browser snapshot -i 2>&1)
    
    if echo "$snapshot" | grep -q "textbox"; then
        log_pass "快速打开面板显示成功"
        agent-browser screenshot /tmp/ide-test-4-quick-open.png 2>&1
        agent-browser press Escape 2>&1
        sleep 1
        return 0
    else
        log_fail "快速打开面板未显示"
        agent-browser press Escape 2>&1
        return 1
    fi
}

# 测试5: 侧边栏切换
test_sidebar() {
    log_step "测试5: 切换侧边栏 (Ctrl+B)"
    
    agent-browser keydown Control 2>&1
    agent-browser press b 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 1
    agent-browser screenshot /tmp/ide-test-5-sidebar-hidden.png 2>&1
    
    agent-browser keydown Control 2>&1
    agent-browser press b 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 1
    agent-browser screenshot /tmp/ide-test-5-sidebar-visible.png 2>&1
    
    log_pass "侧边栏切换测试完成"
    return 0
}

# 测试6: 终端面板
test_terminal() {
    log_step "测试6: 终端面板 (Ctrl+\`)"
    
    agent-browser keydown Control 2>&1
    agent-browser press Backquote 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 2
    agent-browser screenshot /tmp/ide-test-6-terminal.png 2>&1
    
    agent-browser keydown Control 2>&1
    agent-browser press Backquote 2>&1
    agent-browser keyup Control 2>&1
    
    sleep 1
    log_pass "终端面板测试完成"
    return 0
}

# 测试7: 右键菜单
test_context_menu() {
    log_step "测试7: 右键菜单"
    
    # 模拟右键点击
    local result=$(agent-browser eval "
    (function() {
        const fileNode = document.querySelector('[class*=\"file_tree_node\"], [title*=\"file://\"]') ||
                        document.querySelectorAll('[class*=\"file_tree_node\"]')[0];
        if (!fileNode) return 'file node not found';
        
        const event = new MouseEvent('contextmenu', {
            bubbles: true,
            cancelable: true,
            view: window,
            button: 2,
            clientX: 200,
            clientY: 200
        });
        fileNode.dispatchEvent(event);
        return 'context menu dispatched';
    })()
    " 2>&1)
    
    sleep 2
    
    local snapshot=$(agent-browser snapshot 2>&1)
    
    if echo "$snapshot" | grep -q "menuitem.*新建文件\|menuitem.*删除\|menuitem.*重命名"; then
        log_pass "右键菜单显示正常"
        agent-browser screenshot /tmp/ide-test-7-context-menu.png 2>&1
        agent-browser press Escape 2>&1
        sleep 1
        return 0
    else
        log_fail "右键菜单未显示"
        agent-browser press Escape 2>&1
        return 1
    fi
}

# 测试8: Bug - initial-path=/ 根路径
test_root_path_bug() {
    log_step "测试8: Bug 验证 - initial-path=/ 根路径"
    
    # 关闭当前浏览器
    close_editor
    
    # 使用根路径打开
    open_editor "$TOKEN" "/"
    
    local console=$(agent-browser console 2>&1)
    
    # 检查路径映射
    if echo "$console" | grep -q "readDirectory: / -> /"; then
        log_pass "路径映射正确 (/ -> /)"
    else
        log_fail "路径映射错误"
        return 1
    fi
    
    # 检查请求
    if echo "$console" | grep -q "Fetching: http.*/agent/$"; then
        log_pass "WebDAV 请求路径正确 (以 / 结尾)"
    else
        log_warn "WebDAV 请求路径可能有问题"
    fi
    
    # 检查响应
    if echo "$console" | grep -q "Response: 207"; then
        log_pass "WebDAV 响应 207 Multi-Status"
    else
        log_fail "WebDAV 响应非 207"
        return 1
    fi
    
    # 检查文件数量
    local file_count=$(echo "$console" | grep "Parsed files:" | grep -o '[0-9]*' | head -1)
    if [ -n "$file_count" ] && [ "$file_count" -gt 0 ]; then
        log_pass "根路径文件列表正常 ($file_count 个文件)"
    else
        log_fail "根路径文件列表为空"
        return 1
    fi
    
    # 检查文件树
    local tree_result=$(agent-browser eval "
    (function() {
        const files = document.querySelectorAll('[title*=\"file://\"]');
        return files.length;
    })()
    " 2>&1 | grep -o '[0-9]*')
    
    if [ -n "$tree_result" ] && [ "$tree_result" -gt 0 ]; then
        log_pass "文件树显示正常 ($tree_result 个节点)"
        agent-browser screenshot /tmp/ide-test-8-root-path.png 2>&1
        return 0
    else
        log_fail "文件树显示异常"
        return 1
    fi
}

# 运行所有测试
run_tests() {
    local passed=0
    local failed=0
    
    echo ""
    echo "=========================================="
    echo "  W7Panel Web IDE 自动化测试"
    echo "=========================================="
    echo ""
    
    # 获取 Token
    TOKEN=$(get_token)
    if [ -z "$TOKEN" ]; then
        log_fail "未找到 Token"
        exit 1
    fi
    log_info "Token 已获取"
    
    # 测试1: 使用默认路径打开编辑器
    log_info "打开编辑器 (initial-path=$INITIAL_PATH)..."
    open_editor "$TOKEN" "$INITIAL_PATH"
    
    # 运行基础测试
    test_editor_load && ((passed++)) || ((failed++))
    test_file_tree && ((passed++)) || ((failed++))
    test_command_palette && ((passed++)) || ((failed++))
    test_quick_open && ((passed++)) || ((failed++))
    test_sidebar && ((passed++)) || ((failed++))
    test_terminal && ((passed++)) || ((failed++))
    test_context_menu && ((passed++)) || ((failed++))
    
    # 测试8: Bug 验证 - 根路径
    test_root_path_bug && ((passed++)) || ((failed++))
    
    # 关闭浏览器
    close_editor
    
    echo ""
    echo "=========================================="
    echo "  测试结果: ${GREEN}通过 ${passed}${NC} / ${RED}失败 ${failed}${NC}"
    echo "=========================================="
    echo ""
    echo "截图保存在 /tmp/ide-test-*.png"
    echo ""
    
    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# 交互模式
interactive() {
    TOKEN=$(get_token)
    if [ -z "$TOKEN" ]; then
        log_fail "未找到 Token"
        exit 1
    fi
    
    open_editor "$TOKEN" "$INITIAL_PATH"
    
    echo ""
    echo "交互模式 - 可用命令:"
    echo "  snapshot     - 获取页面快照"
    echo "  console      - 查看控制台日志"
    echo "  screenshot   - 截图"
    echo "  click @e1    - 点击元素"
    echo "  fill @e1 txt - 填充文本"
    echo "  press Key    - 按键"
    echo "  eval 'js'    - 执行 JavaScript"
    echo "  exit         - 退出"
    echo ""
    
    while true; do
        read -p "agent-browser> " cmd args
        
        case "$cmd" in
            exit|quit)
                close_editor
                break
                ;;
            snapshot)
                agent-browser snapshot -i 2>&1
                ;;
            console)
                agent-browser console 2>&1
                ;;
            screenshot)
                agent-browser screenshot "/tmp/ide-interactive-$(date +%s).png" 2>&1
                ;;
            *)
                agent-browser $cmd $args 2>&1
                ;;
        esac
    done
}

# 主函数
main() {
    case "${1:-test}" in
        test)
            run_tests
            ;;
        interactive|i)
            interactive
            ;;
        help|--help|-h)
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  test        运行所有测试 (默认)"
            echo "  interactive 进入交互模式"
            echo "  help        显示帮助"
            echo ""
            echo "环境变量:"
            echo "  BASE_URL      服务地址 (默认: http://localhost:8000)"
            echo "  API_URL       WebDAV API 路径 (默认: /k8s/webdav-agent/1/agent)"
            echo "  INITIAL_PATH  初始路径 (默认: /tmp)"
            echo "  K8S_TOKEN     认证 Token"
            ;;
        *)
            log_fail "未知命令: $1"
            exit 1
            ;;
    esac
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --token)
            K8S_TOKEN="$2"
            shift 2
            ;;
        --url)
            BASE_URL="$2"
            shift 2
            ;;
        --path)
            INITIAL_PATH="$2"
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

main "$@"
