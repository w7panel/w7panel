#!/bin/bash
# W7Panel 服务启动脚本
# 用于启动、停止、重启 w7panel 服务
# 
# 使用方式：
#   构建镜像时将脚本放到与 w7panel 同目录
#   容器内：./w7panel-ctl.sh start
#
# 环境变量（通过 K8s deployment 传入）：
#   CAPTCHA_ENABLED - 验证码开关 (默认: false)
#   LOCAL_MOCK     - 本地模拟模式 (默认: 自动检测)
#   KO_DATA_PATH   - 数据目录 (默认: ./kodata)
#   KUBECONFIG     - K8S配置文件 (仅开发模式需要)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXECUTABLE="$SCRIPT_DIR/w7panel"
PID_FILE="$SCRIPT_DIR/w7panel.pid"
LOG_FILE="/tmp/w7panel.log"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

is_running() {
    if [ -f "$PID_FILE" ]; then
        local PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            return 0
        fi
    fi
    if pgrep -f "w7panel server:start" > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

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

start() {
    if is_running; then
        local PID=$(get_pid)
        log_warn "服务已在运行 (PID: $PID)"
        return 0
    fi

    log_info "启动 w7panel..."

    if [ ! -x "$EXECUTABLE" ]; then
        log_error "可执行文件不存在: $EXECUTABLE"
        return 1
    fi

    # 设置默认环境变量
    export CAPTCHA_ENABLED="${CAPTCHA_ENABLED:-false}"
    export KO_DATA_PATH="${KO_DATA_PATH:-$SCRIPT_DIR/kodata}"
    
    # 自动检测模式：开发模式(LOCAL_MOCK=true) 或 生产模式(LOCAL_MOCK=false)
    # 如果设置了 KUBECONFIG，则为开发模式
    if [ -n "$KUBECONFIG" ]; then
        export LOCAL_MOCK="${LOCAL_MOCK:-true}"
        log_info "开发模式: 使用本地 kubeconfig ($KUBECONFIG)"
    else
        export LOCAL_MOCK="${LOCAL_MOCK:-false}"
        log_info "生产模式: 使用 ServiceAccount"
    fi

    cd "$SCRIPT_DIR"

    nohup "$EXECUTABLE" server:start >> "$LOG_FILE" 2>&1 &
    local PID=$!
    
    echo $PID > "$PID_FILE"
    
    sleep 2

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

stop() {
    if ! is_running; then
        log_info "服务未运行"
        rm -f "$PID_FILE"
        return 0
    fi

    local PID=$(get_pid)
    log_info "停止 w7panel (PID: $PID)..."

    local CHILDREN=$(ps -eo pid,ppid 2>/dev/null | awk -v ppid="$PID" '$2 == ppid {print $1}')
    
    for child in $CHILDREN; do
        kill -TERM "$child" 2>/dev/null || true
    done
    
    sleep 1
    
    for child in $CHILDREN; do
        if kill -0 "$child" 2>/dev/null; then
            kill -KILL "$child" 2>/dev/null || true
        fi
    done

    kill -TERM "$PID" 2>/dev/null || true
    
    local count=0
    while kill -0 "$PID" 2>/dev/null && [ $count -lt 5 ]; do
        sleep 1
        count=$((count + 1))
    done

    if kill -0 "$PID" 2>/dev/null; then
        log_warn "进程未响应 SIGTERM，使用 SIGKILL"
        kill -KILL "$PID" 2>/dev/null || true
        sleep 1
    fi

    rm -f "$PID_FILE"
    
    local orphans=$(pgrep -f "w7panel server:start")
    if [ -n "$orphans" ]; then
        log_warn "发现孤儿进程，清理中..."
        for orphan in $orphans; do
            kill -TERM "$orphan" 2>/dev/null || true
        done
        sleep 1
        kill -9 $orphans 2>/dev/null || true
    fi

    log_info "服务已停止"
}

restart() {
    log_info "重启 w7panel..."
    stop
    sleep 1
    start
}

status() {
    if is_running; then
        local PID=$(get_pid)
        log_info "服务运行中 (PID: $PID)"
        ps -p "$PID" -o pid,ppid,stat,%cpu,%mem,cmd 2>/dev/null || true
        
        if curl -s http://localhost:8080/ > /dev/null 2>&1; then
            log_info "HTTP 服务正常 (http://localhost:8080)"
        else
            log_warn "HTTP 服务异常"
        fi
        
        local zombies=$(ps aux | awk '$8 ~ /Z/ && $11 ~ /w7panel/' | wc -l)
        if [ "$zombies" -gt 0 ]; then
            log_warn "发现 $zombies 个僵尸进程"
        fi
    else
        log_info "服务未运行"
        local orphans=$(pgrep -f "w7panel server:start")
        if [ -n "$orphans" ]; then
            log_warn "发现孤儿进程: $orphans"
        fi
    fi
}

logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        log_error "日志文件不存在: $LOG_FILE"
        return 1
    fi
}

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
    echo "环境变量（通过 K8s deployment 传入）:"
    echo "  CAPTCHA_ENABLED - 验证码开关 (默认: false)"
    echo "  LOCAL_MOCK     - 模式: true=开发, false=生产 (默认: 自动检测)"
    echo "  KO_DATA_PATH   - 数据目录 (默认: ./kodata)"
    echo "  KUBECONFIG     - K8S配置文件 (仅开发模式需要)"
    echo ""
    echo "模式说明:"
    echo "  开发模式 (LOCAL_MOCK=true):  需要设置 KUBECONFIG"
    echo "  生产模式 (LOCAL_MOCK=false): 不需要 KUBECONFIG，使用 ServiceAccount"
    echo ""
    echo "示例:"
    echo "  # 开发模式（需要 kubeconfig.yaml）:"
    echo "  export KUBECONFIG=/path/to/kubeconfig.yaml"
    echo "  ./w7panel-ctl.sh start"
    echo ""
    echo "  # 生产模式（使用 ServiceAccount）:"
    echo "  export LOCAL_MOCK=false"
    echo "  ./w7panel-ctl.sh start"
    echo ""
    echo "  # K8s deployment 配置（生产模式）:"
    echo "  env:"
    echo "    - name: LOCAL_MOCK"
    echo "      value: \"false\""
    echo "    - name: CAPTCHA_ENABLED"  
    echo "      value: \"false\""
}

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
