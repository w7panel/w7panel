#!/bin/bash
# W7Panel 进程包装器 - 持久父进程模式
# 父进程永远不退出，wait 子进程，确保子进程不会变成僵尸

PROGRAM="$1"
shift

LOG_FILE="/tmp/w7panel-wrapper.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

log "Starting: $PROGRAM $*"

# 启动子进程
"$PROGRAM" "$@" >> "$LOG_FILE" 2>&1 &
CHILD_PID=$!
log "Child PID: $CHILD_PID"

# 永远等待子进程
while true; do
    if ! kill -0 "$CHILD_PID" 2>/dev/null; then
        # 子进程已退出
        wait "$CHILD_PID" 2>/dev/null
        EXIT_CODE=$?
        log "Child exited with code: $EXIT_CODE"
        break
    fi
    sleep 1
done

log "Wrapper exiting"
