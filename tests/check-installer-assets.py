#!/usr/bin/env python3
"""检查 installer 发布资源是否完整，以及 HelmChart 与归档版本是否一致。"""
from pathlib import Path
import re
import tarfile

ROOT = Path(__file__).resolve().parents[1]
PREFIX = "https://w7panel.com/"
script = (ROOT / "installer/install.sh").read_text()
for url in re.findall(r'https://w7panel\.com/installer/[^\s"\']+', script):
    if "$" in url or "}" in url:
        continue
    assert (ROOT / url.removeprefix(PREFIX)).is_file(), f"安装资源缺失: {url}"

for manifest in sorted((ROOT / "installer/manifests").glob("*.yaml")):
    source = manifest.read_text()
    url = re.search(r"^  chart: (\S+)", source, re.M).group(1)
    version = re.search(r"^  version: (\S+)", source, re.M).group(1).lstrip("v")
    assert url.startswith(PREFIX), f"非本站 Chart: {url}"
    chart = ROOT / url.removeprefix(PREFIX)
    with tarfile.open(chart) as archive:
        meta = next(m for m in archive.getmembers()
                    if m.name.count("/") == 1 and m.name.endswith("/Chart.yaml"))
        metadata = archive.extractfile(meta).read().decode()
        actual = re.search(r'^version: ["\']?([^\s"\']+)', metadata, re.M).group(1).lstrip("v")
        assert actual == version, f"{manifest}: manifest={version}, chart={actual}"
    print(f"OK {manifest.name}: {chart.relative_to(ROOT)} ({version})")
