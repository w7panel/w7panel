#!/usr/bin/env python3
"""隔离验证安装器等待函数在 Bash/sh 下的成功和超时，不执行安装。"""
from pathlib import Path
import subprocess

source = (Path(__file__).resolve().parents[1] / 'installer/install.sh').read_text()
function = source[source.index('checkW7panelInstalled() {'):source.index('# 导入镜像')]
for shell in ('bash', 'sh'):
    for mode in ('ready', 'timeout'):
        setup = '''set -e
SECONDS=
INFO_COLOR= NC=
attempt=0
internalIP() { echo 127.0.0.1; }
fatal() { echo "$*" >&2; exit 42; }
sleep() { :; }
'''
        if mode == 'ready':
            setup += 'curl() { attempt=$((attempt+1)); [ "$attempt" -ge 2 ]; }\n'
        else:
            setup += '''curl() { attempt=$((attempt+1)); return 1; }
date() { if [ "$attempt" -eq 0 ]; then echo 1000; else echo 1601; fi; }
'''
        result = subprocess.run([shell, '-c', setup + function + '\ncheckW7panelInstalled\n'],
                                capture_output=True, text=True, timeout=5)
        assert result.returncode == (42 if mode == 'timeout' else 0), result.stderr
        assert 'number' not in result.stderr and 'integer' not in result.stderr, result.stderr
        print(f'OK {shell}: {mode}')
