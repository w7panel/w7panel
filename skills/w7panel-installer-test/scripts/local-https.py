#!/usr/bin/env python3
"""Debian/systemd 本机 installer HTTPS 映射；不执行安装或卸载。"""
import argparse
import hashlib
import http.server
import os
from pathlib import Path
import shutil
import socket
import ssl
import subprocess
import sys
from urllib.parse import unquote, urlsplit

STATE = Path('/var/lib/w7panel-local-https')
CA = Path('/usr/local/share/ca-certificates/w7panel-local-https.crt')
UNIT = 'w7panel-local-https.service'
MARK = '# w7panel-local-https'
SCRIPT = Path(__file__).resolve()
REPO = SCRIPT.parents[3]


def run(*args, **kwargs):
    return subprocess.run(args, check=True, **kwargs)


def hosts_remove():
    path = Path('/etc/hosts')
    path.write_text(''.join(line for line in path.read_text().splitlines(True)
                            if not line.rstrip().endswith(MARK)))


def active():
    return subprocess.run(['systemctl', 'is-active', '--quiet', UNIT]).returncode == 0


def check():
    data = subprocess.check_output(['curl', '--noproxy', 'w7panel.com', '-fsS',
                                    '--max-time', '10',
                                    'https://w7panel.com/installer/install.sh'])
    assert data == (REPO / 'installer/install.sh').read_bytes(), 'HTTPS 内容与当前仓库不一致'
    print('HTTPS 证书校验通过，安装脚本与当前仓库一致：' + hashlib.sha256(data).hexdigest())


def serve():
    root = (REPO / 'installer').resolve()

    class Handler(http.server.SimpleHTTPRequestHandler):
        def translate_path(self, path):
            decoded = unquote(urlsplit(path).path)
            target = (root / decoded.removeprefix('/installer/')).resolve()
            if not decoded.startswith('/installer/') or not target.is_relative_to(root):
                return '/dev/null/w7panel-denied'
            return str(target)

        def list_directory(self, path):
            self.send_error(403)
            return None

    server = http.server.ThreadingHTTPServer(('127.0.0.1', 443), Handler)
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(STATE / 'cert.pem', STATE / 'key.pem')
    server.socket = context.wrap_socket(server.socket, server_side=True)
    server.serve_forever()


def start():
    if active():
        check()
        return
    for cmd in ('mkcert', 'curl', 'systemctl', 'systemd-run', 'update-ca-certificates'):
        if not shutil.which(cmd):
            raise RuntimeError('缺少依赖：' + cmd)
    if not (REPO / 'installer/install.sh').is_file():
        raise RuntimeError('请在当前项目的 skills 目录中运行此脚本')
    with socket.socket() as sock:
        sock.bind(('127.0.0.1', 443))
    for line in Path('/etc/hosts').read_text().splitlines():
        fields = line.split('#', 1)[0].split()
        if 'w7panel.com' in fields[1:] and not line.rstrip().endswith(MARK):
            raise RuntimeError('/etc/hosts 已有 w7panel.com 配置，请先移除旧测试映射：' + line)
    STATE.mkdir(mode=0o700, parents=True, exist_ok=True)
    STATE.chmod(0o700)
    ca_root = STATE / 'ca'
    ca_root.mkdir(mode=0o700, exist_ok=True)
    env = dict(os.environ, CAROOT=str(ca_root))
    run('mkcert', '-cert-file', str(STATE / 'cert.pem'), '-key-file', str(STATE / 'key.pem'),
        'w7panel.com', env=env)
    shutil.copyfile(ca_root / 'rootCA.pem', CA)
    CA.chmod(0o644)
    try:
        run('update-ca-certificates')
        hosts_remove()
        with Path('/etc/hosts').open('a') as output:
            output.write('\n127.0.0.1 w7panel.com ' + MARK + '\n')
        run('systemd-run', '--unit=' + UNIT, '--collect', '--property=Restart=on-failure',
            sys.executable, str(SCRIPT), '_serve')
        # systemd-run 返回时监听可能尚未建立。
        import time
        for attempt in range(20):
            try:
                check()
                break
            except subprocess.CalledProcessError:
                if attempt == 19:
                    raise
                time.sleep(0.25)
    except BaseException:
        stop()
        raise
    print('仓库：' + str(REPO))
    print('仅本机生效；保持终端 NO_PROXY/no_proxy 包含 w7panel.com（见 Skill）。')


def stop():
    subprocess.run(['systemctl', 'stop', UNIT], check=False)
    hosts_remove()
    CA.unlink(missing_ok=True)
    run('update-ca-certificates', '--fresh')
    print('已停止映射并移除本脚本添加的 hosts 和 CA；K3s 不受影响。')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=['start', 'status', 'stop', '_serve'])
    action = parser.parse_args().action
    if os.geteuid() != 0:
        parser.error('请用 sudo 或 root 执行')
    if action == '_serve':
        serve()
    elif action == 'start':
        start()
    elif action == 'stop':
        stop()
    else:
        if not active():
            raise RuntimeError('本地 HTTPS 服务未运行')
        check()


if __name__ == '__main__':
    main()
