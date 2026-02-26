#!/bin/bash
# W7Panel 服务启动脚本
# 用于启动、停止、重启 w7panel 服务

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
DIST_DIR="$BASE_DIR/dist"
PID_FILE="$DIST_DIR/w7panel.pid"
LOG_FILE="/tmp/w7panel.log"

# 默认配置
EXECUTABLE="$DIST_DIR/w7panel"
DEFAULT_KUBECONFIG="$BASE_DIR/kubeconfig.yaml"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务是否运行
is_running() {
    if [ -f "$PID_FILE" ]; then
        local PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            return 0
        fi
    fi
    # 也检查是否有孤儿进程
    if pgrep -f "w7panel server:start" > /dev/null 2>&1; then
        return 0
    fi
    # 检查 Python wrapper
    if pgrep -f "wrapper.py.*w7panel" > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

# 获取运行中的 PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        local PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            echo "$PID"
            return
        fi
    fi
    pgrep -f "w7panel server:start" | head -1
}

# 清理僵尸进程
clean_zombies() {
    # 检查是否有僵尸进程
    local zombies=$(ps aux | awk '$8 ~ /Z/ && $11 ~ /w7panel/ {print $2}')
    if [ -n "$zombies" ]; then
        log_warn "发现僵尸进程，尝试清理..."
        # 僵尸进程无法直接杀死，只能等待父进程收割
        # 如果父进程已退出，僵尸会被 PID 1 收养
        for pid in $zombies; do
            # 尝试收割（通常无效，但尝试一下）
            wait $pid 2>/dev/null || true
        done
    fi
}

# 启动服务
start() {
    if is_running; then
        local PID=$(get_pid)
        log_warn "服务已在运行 (PID: $PID)"
        return 0
    fi

    log_info "启动 w7panel..."

    # 检查可执行文件
    if [ ! -x "$EXECUTABLE" ]; then
        log_error "可执行文件不存在: $EXECUTABLE"
        return 1
    fi

    # 检查 kubeconfig
    if [ ! -f "$DEFAULT_KUBECONFIG" ]; then
        log_warn "kubeconfig 文件不存在: $DEFAULT_KUBECONFIG"
    fi

    # 切换到 dist 目录
    cd "$DIST_DIR"

    # 设置环境变量
    export CAPTCHA_ENABLED="${CAPTCHA_ENABLED:-false}"
    export LOCAL_MOCK="${LOCAL_MOCK:-true}"
    export KO_DATA_PATH="${KO_DATA_PATH:-$DIST_DIR/kodata}"
    export KUBECONFIG="${KUBECONFIG:-$DEFAULT_KUBECONFIG}"

    # 直接启动 w7panel
    nohup "$EXECUTABLE" server:start >> "$LOG_FILE" 2>&1 &
    local PID=$!
    
    echo $PID > "$PID_FILE"
    
    # 等待服务启动
    sleep 2

    # 验证服务是否启动成功
    if kill -0 "$PID" 2>/dev/null; then
        if curl -s http://localhost:8080/ > /dev/null 2>&1; then
            log_info "服务启动成功 (PID: $PID)"
            log_info "日志文件: $LOG_FILE"
            return 0
        else
            log_warn "服务已启动但端口未就绪，请检查日志: $LOG_FILE"
            return 0
        fi
    else
        log_error "服务启动失败，请检查日志: $LOG_FILE"
        rm -f "$PID_FILE"
        return 1
    fi
}

# 停止服务
stop() {
    if ! is_running; then
        log_info "服务未运行"
        rm -f "$PID_FILE"
        return 0
    fi

    local PID=$(get_pid)
    log_info "停止 w7panel (PID: $PID)..."

    # 找到所有子进程
    local CHILDREN=$(ps -eo pid,ppid 2>/dev/null | awk -v ppid="$PID" '$2 == ppid {print $1}')
    
    # 先发送 SIGTERM 给子进程
    for child in $CHILDREN; do
        kill -TERM "$child" 2>/dev/null || true
    done
    
    # 等待子进程退出
    sleep 1
    
    # 强制杀死子进程
    for child in $CHILDREN; do
        if kill -0 "$child" 2>/dev/null; then
            kill -KILL "$child" 2>/dev/null || true
        fi
    done

    # 终止主进程
    kill -TERM "$PID" 2>/dev/null || true
    
    # 等待主进程退出
    local count=0
    while kill -0 "$PID" 2>/dev/null && [ $count -lt 5 ]; do
        sleep 1
        count=$((count + 1))
    done

    # 如果还在运行，使用 SIGKILL
    if kill -0 "$PID" 2>/dev/null; then
        log_warn "进程未响应 SIGTERM，使用 SIGKILL"
        kill -KILL "$PID" 2>/dev/null || true
        sleep 1
    fi

    # 清理
    rm -f "$PID_FILE"
    
    # 清理可能的孤儿进程
    local orphans=$(pgrep -f "w7panel server:start")
    if [ -n "$orphans" ]; then
        log_warn "发现孤儿进程，清理中..."
        # 优雅停止：先 SIGTERM 等待退出，再 SIGKILL
        for orphan in $orphans; do
            kill -TERM "$orphan" 2>/dev/null || true
        done
        sleep 1
        kill -9 $orphans 2>/dev/null || true
    fi

    log_info "服务已停止"
}

# 重启服务
restart() {
    log_info "重启 w7panel..."
    stop
    sleep 1
    start
}

# 查看状态
status() {
    if is_running; then
        local PID=$(get_pid)
        log_info "服务运行中 (PID: $PID)"
        
        # 显示进程信息
        ps -p "$PID" -o pid,ppid,stat,%cpu,%mem,cmd 2>/dev/null || true
        
        # 检查端口
        if curl -s http://localhost:8080/ > /dev/null 2>&1; then
            log_info "HTTP 服务正常 (http://localhost:8080)"
        else
            log_warn "HTTP 服务异常"
        fi
        
        # 检查僵尸进程
        local zombies=$(ps aux | awk '$8 ~ /Z/ && $11 ~ /w7panel/' | wc -l)
        if [ "$zombies" -gt 0 ]; then
            log_warn "发现 $zombies 个僵尸进程"
        fi
    else
        log_info "服务未运行"
        # 检查是否有残留进程
        local orphans=$(pgrep -f "w7panel server:start")
        if [ -n "$orphans" ]; then
            log_warn "发现孤儿进程: $orphans"
        fi
    fi
}

# 查看日志
logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        log_error "日志文件不存在: $LOG_FILE"
        return 1
    fi
}

# 显示帮助
usage() {
    echo "用法: $0 {start|stop|restart|status|logs}"
    echo ""
    echo "命令:"
    echo "  start   - 启动服务"
    echo "  stop    - 停止服务"
    echo "  restart - 重启服务"
    echo "  status  - 查看状态"
    echo "  logs    - 查看日志"
    echo ""
    echo "环境变量:"
    echo "  CAPTCHA_ENABLED - 验证码开关 (默认: false)"
    echo "  LOCAL_MOCK      - 本地模拟模式 (默认: true)"
    echo "  KO_DATA_PATH    - 数据目录 (默认: \$BASE_DIR/dist/kodata)"
    echo "  KUBECONFIG      - K8S 配置文件 (默认: \$BASE_DIR/kubeconfig.yaml)"
}

# 主入口
case "${1:-}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    *)
        usage
        exit 1
        ;;
esac
