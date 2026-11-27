#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
state_dir=/tmp/w7panel-k3s-ci
image=w7panel-ci:install-test
kubectl() { sudo k3s kubectl "$@"; }

prepare() {
  test ! -e /usr/local/bin/k3s
  test ! -e /var/lib/rancher/k3s
  mkdir -p "$state_dir/work/w7panel/images" "$state_dir/evidence"
  # Hosted runners expose many CPUs to Vite, but its Less worker sync ports
  # can fail during this build. Compile preprocessors in the main process.
  python3 - "$repo_dir/w7panel-ui/config/vite.config.prod.ts" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
source = path.read_text()
needle = "    mode: 'production',\n"
assert source.count(needle) == 1
path.write_text(source.replace(needle, needle + "    css: { preprocessorMaxWorkers: 0 },\n", 1))
PY
  (cd "$repo_dir/w7panel-ui" && HUSKY=0 npm install --legacy-peer-deps && npm run build)
  cp -R "$repo_dir/w7panel-ui/dist/." "$repo_dir/w7panel-server/kodata/"
  mkdir -p "$repo_dir/w7panel-server/kodata/assets"
  cp "$repo_dir/w7panel-server/kodata/logo.png" "$repo_dir/w7panel-server/kodata/assets/logo.png"
  cp "$repo_dir/w7panel-server/kodata/index.html" "$repo_dir/w7panel-server/kodata/panel.html"
  make -C "$repo_dir/w7panel-server" ko-build \
    LOCAL_IMAGE=w7panel-ci IMAGE_TAG=install-test \
    GO_TOOLCHAIN_ROOT=/nonexistent
  docker image inspect "$image" >/dev/null
  docker save "$image" -o "$state_dir/work/w7panel/images/panel.tar"
  # The checkout is disposable. Keep the installer and Chart unchanged; only
  # point this test's HelmChart at the image built from the checked-out source.
  python3 - "$repo_dir/installer/manifests/w7panel-offline.yaml" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
source = path.read_text()
needle = '  set:\n'
assert source.count(needle) == 1
path.write_text(source.replace(needle, '''  set:
    image.repository: "w7panel-ci"
    image.tag: "install-test"
    image.pullPolicy: "IfNotPresent"
''', 1))
PY
  sudo python3 "$repo_dir/skills/w7panel-installer-test/scripts/local-https.py" start
  curl --noproxy w7panel.com -fsS https://w7panel.com/installer/install.sh \
    -o "$state_dir/work/install.sh"
  cmp "$state_dir/work/install.sh" "$repo_dir/installer/install.sh"
}

wait_job() {
  local namespace=$1 name=$2 deadline=$((SECONDS + 900))
  until kubectl -n "$namespace" get job "$name" >/dev/null 2>&1; do
    (( SECONDS < deadline )) || { echo "Job 未创建: $namespace/$name" >&2; return 1; }
    sleep 5
  done
  kubectl -n "$namespace" wait --for=condition=complete "job/$name" --timeout=900s
}

run_install() {
  cd "$state_dir/work"
  # Use the installer itself to create K3s and submit all HelmCharts.
  export PUBLIC_IP=127.0.0.1
  export INSTALL_K3S_VERSION=v1.37.2+k3s1
  export INTERNAL_IP
  INTERNAL_IP=$(ip -4 route get 1.1.1.1 | awk '{for (i=1;i<=NF;i++) if ($i=="src") {print $(i+1); exit}}')
  test -n "$INTERNAL_IP"
  set +e
  bash ./install.sh >"$state_dir/evidence/install.log" 2>&1
  install_status=$?
  set -e
  printf '%s\n' "$install_status" >"$state_dir/evidence/install.exit"
  test "$install_status" -eq 0 || { cat "$state_dir/evidence/install.log"; return 1; }
  kubectl wait --for=condition=Ready node/server1 --timeout=180s
  kubectl -n kube-system rollout status daemonset/cilium --timeout=180s
  kubectl -n kube-system rollout status deployment/cilium-operator --timeout=180s
  wait_job kube-system helm-install-w7panel-offline
  local upgrade_job='' deadline=$((SECONDS + 180))
  until [ -n "$upgrade_job" ]; do
    upgrade_job=$(kubectl -n default get jobs -o name | grep -E '^job.batch/w7panel-offline-[0-9]+-upgrade-' | tail -1 || true)
    [ -n "$upgrade_job" ] && break
    (( SECONDS < deadline )) || { echo '面板升级 Job 未创建' >&2; return 1; }
    sleep 5
  done
  test -n "$upgrade_job"
  kubectl -n default wait --for=condition=complete "$upgrade_job" --timeout=900s
  kubectl -n default logs "$upgrade_job" >"$state_dir/evidence/upgrade.log"
  test "$(curl --noproxy '*' -sS -o /dev/null -w '%{http_code}' "http://${INTERNAL_IP}:9090")" = 200
  sudo k3s --version | tee "$state_dir/evidence/k3s-version.txt"
  grep -q 'v1.37.2+k3s1' "$state_dir/evidence/k3s-version.txt"
}

evidence() {
  mkdir -p "$state_dir/evidence"
  if command -v k3s >/dev/null 2>&1; then
    kubectl get nodes -o wide >"$state_dir/evidence/nodes.txt" 2>&1 || true
    kubectl get pods,jobs,helmcharts -A -o wide >"$state_dir/evidence/resources.txt" 2>&1 || true
    kubectl get events -A --sort-by=.metadata.creationTimestamp >"$state_dir/evidence/events.txt" 2>&1 || true
    sudo journalctl -u k3s --no-pager -n 400 >"$state_dir/evidence/k3s-journal.txt" 2>&1 || true
  fi
  sudo python3 "$repo_dir/skills/w7panel-installer-test/scripts/local-https.py" stop || true
}

case "${1:-}" in
  prepare) prepare ;;
  run) run_install ;;
  evidence) evidence ;;
  *) echo 'usage: ci-k3s-installer.sh {prepare|run|evidence}' >&2; exit 2 ;;
esac
