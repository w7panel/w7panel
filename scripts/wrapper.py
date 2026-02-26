#!/usr/bin/env python3
import os
import sys
import signal
import time
import subprocess

log_file = open("/tmp/w7panel-wrapper.log", "a")

def log(msg):
    ts = time.strftime("%Y-%m-%d %H:%M:%S")
    log_file.write(f"[{ts}] {msg}\n")
    log_file.flush()
    print(f"[{ts}] {msg}")

if len(sys.argv) < 2:
    log("Usage: wrapper.py <program> [args...]")
    sys.exit(1)

program = sys.argv[1]
args = sys.argv[1:]

log(f"Starting: {program} {' '.join(args[1:])}")

# 启动子进程
child = subprocess.Popen(args, stdout=log_file, stderr=subprocess.STDOUT)
log(f"Child PID: {child.pid}")

def signal_handler(signum, frame):
    sig = signal.Signals(signum).name
    log(f"Received {sig}, shutting down...")
    
    # 发送 SIGTERM 给子进程
    if child.poll() is None:
        log(f"Sending SIGTERM to child {child.pid}")
        child.terminate()
        
        # 等待子进程退出
        try:
            child.wait(timeout=5)
            log("Child exited gracefully")
        except subprocess.TimeoutExpired:
            log("Child didn't exit, sending SIGKILL")
            child.kill()
            child.wait()
    
    log("Exiting")
    log_file.close()
    sys.exit(0)

# 设置信号处理
signal.signal(signal.SIGTERM, signal_handler)
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGQUIT, signal_handler)

# 等待子进程
while True:
    ret = child.poll()
    if ret is not None:
        log(f"Child exited with code: {ret}")
        break
    time.sleep(0.5)

log_file.close()
sys.exit(ret if ret is not None else 0)
