#!/bin/bash
set -e
set -o pipefail

# 检查终端是否支持颜色输出
if [ -t 1 ]; then
    INFO_COLOR='\033[0m'
    WARN_COLOR='\033[33m'
    ERROR_COLOR='\033[31m'
    SUCCESS_COLOR='\033[32m'
    NC='\033[0m' # No Color
else
    INFO_COLOR=''
    WARN_COLOR=''
    ERROR_COLOR=''
    SUCCESS_COLOR=''
    NC=''
fi

# 信息日志
info() {
    printf "${INFO_COLOR}[INFO]${NC} %s\n" "$*"
}

# 警告日志
warn() {
    printf "${WARN_COLOR}[WARN] %s${NC}\n" "$*" >&2
}

# 成功日志
success() {
    printf "${SUCCESS_COLOR}[SUCCESS]${NC} %s\n" "$*"
}

# 错误日志
fatal() {
    printf "${ERROR_COLOR}[ERROR] %s${NC}\n" "$*" >&2
    exit 1
}

# 提示日志
tips() {
    printf "${SUCCESS_COLOR}%s${NC}\n" "$*"
}

# --- add quotes to command arguments ---
# 给命令参数添加引号
quote() {
    for arg in "$@"; do
        printf '%s\n' "$arg" | sed "s/'/'\\\\''/g;1s/^/'/;\$s/\$/'/"
    done
}

# --- add indentation and trailing slash to quoted args ---
# 给引用的参数添加缩进和反斜杠
quote_indent() {
    printf ' \\\n'
    for arg in "$@"; do
        printf '\t%s \\\n' "$(quote "$arg")"
    done
}

# --- escape most punctuation characters, except quotes, forward slash, and space ---
# 转义大多数标点符号，除了引号、斜杠和空格
escape() {
    printf '%s' "$@" | sed -e 's/\([][!\#$%&()*;<=>?\_`{|}]\)/\\\1/g;'
}

# --- escape double quotes ---
# 转义双引号
escape_dq() {
    printf '%s' "$@" | sed -e 's/"/\\"/g'
}

# 处理命令行参数
process_args() {
    eval set -- $(escape "$@") $(quote "$@")
}

# 进度条函数
show_progress() {
    local pid=$1
    local delay=0.75
    local spinstr="|/-\\"
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf "${WARN_COLOR} [%c]  ${NC}" "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# 下载文件函数（带重试机制）
download_files() {
    local url=$1
    local path=".$(echo "$url" | sed -E 's|^https?://[^/]+/installer/|/w7panel/|')"
    local save_path=$(dirname "$path")
    local filename=$(basename "$path")
    sudo mkdir -p "$save_path" || fatal "创建目录失败: $save_path"

    if [ -s "$path" ]; then
        info "文件已存在: ${path}"
        return 0
    fi

    # 重试配置参数
    local max_retries=3
    local retry_delay=5
    local retry_count=0

    while [ $retry_count -le $max_retries ]; do
        # 最后一次重试不显示特殊提示
        if [ $retry_count -gt 0 ]; then
            info "正在尝试第 ${retry_count}/${max_retries} 次重试..."
        fi

        if sudo wget -q --show-progress --progress=bar:force:noscroll \
            -c --timeout=10 --tries=1 --retry-connrefused \
            -O "$path.part" "$url"; then
            sudo mv "$path.part" "$path"
            return 0
        fi

        # 重试前清理失败的文件
        [ ! -f "$path.part" ] || sudo rm -f "$path.part"
        
        # 未超过最大重试次数时等待
        if [ $retry_count -lt $max_retries ]; then
            sleep $retry_delay
        fi
        
        retry_count=$((retry_count + 1))
    done

    fatal "下载失败: $url（已重试 ${max_retries} 次），请检查网络连接或稍后再试！"
}

# 载资源函数
downloadResource() {
    info "开始下载微擎面板资源..."
    local resources="
        https://w7panel.com/installer/manifests/cert-manager.yaml
        https://w7panel.com/installer/manifests/cilium.yaml
        https://w7panel.com/installer/manifests/higress.yaml
        https://w7panel.com/installer/manifests/w7panel-offline.yaml
    "

    for resource in $resources; do
        download_files "$resource"
    done
}

# 获取公网IP
publicNetworkIp() {
    # publicIp 为空，则重新获取publicIp
    if [ -z "$PUBLIC_IP" ]; then
        PUBLIC_IP=$(curl -fsS --max-time 10 https://ifconfig.me)
        echo "$PUBLIC_IP"
    else
        echo "$PUBLIC_IP"
    fi
}

# 获取内网IP（兼容多系统）
internalIP() {
    if [ -z "$INTERNAL_IP" ]; then
        # 使用默认路由的源地址，避免重装后选中残留的 cilium_host 等虚拟网卡。
        INTERNAL_IP=$(ip -4 route get 1.1.1.1 | awk '{for (i=1;i<=NF;i++) if ($i=="src") {print $(i+1); exit}}')
        [ -n "$INTERNAL_IP" ] || fatal "无法检测节点 IP，请显式设置 INTERNAL_IP"
        echo "$INTERNAL_IP"
    else
        echo "$INTERNAL_IP"
    fi
}

# 处理sysctl配置
etcSysctl() {
    # 下载资源文件
    download_files "https://w7panel.com/installer/etc/sysctl.d/k3s.conf"
    
    if command -v sysctl &> /dev/null; then
        local ETC_PATH="/etc/sysctl.d"
        sudo mkdir -p "$ETC_PATH" || {
            fatal "Failed to create directory: $ETC_PATH"
            return 1
        }
        sudo chmod -R 755 "$ETC_PATH"
        sudo cp "./w7panel/etc/sysctl.d/k3s.conf" "$ETC_PATH" || {
            fatal "Failed to copy k3s.conf to $ETC_PATH"
            return 1
        }
        sudo sysctl --system >/dev/null 2>&1 || {
            warn "Failed to reload sysctl settings"
        }
    fi
}

# 处理私有仓库配置
etcPrivaterRegistry() {
    # 下载资源文件
    download_files "https://w7panel.com/installer/etc/registries.yaml"
    
    local ETC_PATH="/etc/rancher/k3s/"
    sudo mkdir -p "$ETC_PATH" || {
        fatal "Failed to create directory: $ETC_PATH"
        return 1
    }
    sudo cp "./w7panel/etc/registries.yaml" "$ETC_PATH" || {
        fatal "Failed to copy registries.yaml to $ETC_PATH"
        return 1
    }
}

# 处理systemd配置
etcSystemd() {
    download_files "https://w7panel.com/installer/etc/k3s.service.d/override.conf"
    local service_name="k3s"
    if [ -n "$K3S_URL" ] && [[ "$K3S_NODE_NAME" == agent* ]]; then
        service_name="k3s-agent"
    fi
    local target="/etc/systemd/system/${service_name}.service.d"
    sudo mkdir -p "$target"
    sudo cp ./w7panel/etc/k3s.service.d/override.conf "$target/override.conf"

    # 重新加载 systemd 管理器配置
    sudo systemctl daemon-reload || {
        fatal "Failed to reload systemd manager configuration"
    }

    # 重启对应类型的服务
    sudo systemctl restart "${service_name}.service" || {
        fatal "Failed to restart ${service_name}.service"
    }
    info "${service_name}.service has been restarted successfully."
}

# 检查K3S是否已安装
checkK3SInstalled() {
    info 'start check server is installed 检测k3s是否已安装'
    if [ -x "/usr/local/bin/k3s" ]; then
        if [ -f /usr/local/bin/k3s-uninstall.sh ]; then
            local type="Server"
            local uninstall="/usr/local/bin/k3s-uninstall.sh"
        else
            local type="Agent"
            local uninstall="/usr/local/bin/k3s-agent-uninstall.sh"
        fi
        
        warn "K3s $type has been installed , Please execute $uninstall to uninstall k3s "
        warn "K3s $type 已安装 , 请先执行 $uninstall 命令卸载 "
        exit 1
    fi
}

# 检查微擎面板是否安装成功
checkW7panelInstalled() {
    printf "${INFO_COLOR}[INFO]${NC} %s" "微擎面板正在初始化，预计需要3-5分钟，请耐心等待..."
    local deadline=$((SECONDS + 600))
    while true; do
        curl -fsS --max-time 2 -o /dev/null "http://$(internalIP):9090" 2>/dev/null && break
        [ "$SECONDS" -lt "$deadline" ] || fatal "等待面板就绪超时，请检查 k3s kubectl get pods -A 和 journalctl -u k3s"
        sleep 2
    done
    echo
}

# 导入镜像
importImages() {
    local IMAGES_DIR="./w7panel/images"
    [ ! -d "$IMAGES_DIR" ] && return 0

    local total
    total=$(find "$IMAGES_DIR" -maxdepth 1 -type f -name "*.tar" | wc -l)
    local count=0
    for IMAGE_FILE in "$IMAGES_DIR"/*.tar; do
        [ -f "$IMAGE_FILE" ] || continue
        count=$((count+1))
        info "导入镜像中 [$count/$total] $(basename $IMAGE_FILE)"
        sudo /usr/local/bin/k3s ctr -n=k8s.io images import "$IMAGE_FILE" >/dev/null 2>&1 || {
            warn "镜像导入失败: $(basename $IMAGE_FILE)"
        }
    done
}

# 安装Helm Charts
installHelmCharts() {
    info 'start install helm charts'
    local M_PATH="/var/lib/rancher/k3s/server/manifests/"
    # 先下载 Chart 到 K3s 静态目录，避免启动期 Pod DNS 和私有 CA 依赖。
    local static_path="/var/lib/rancher/k3s/server/static/charts"
    sudo mkdir -p "$M_PATH" "$static_path"
    local manifest chart_url chart_name
    for manifest in ./w7panel/manifests/*.yaml; do
        chart_url=$(awk '$1 == "chart:" {print $2}' "$manifest")
        [ -n "$chart_url" ] || continue
        download_files "$chart_url"
        chart_name=$(basename "$chart_url")
        sudo cp "./w7panel/${chart_url#https://w7panel.com/installer/}" "$static_path/$chart_name"
        sed "s|$chart_url|https://%{KUBERNETES_API}%/static/charts/$chart_name|" "$manifest" | \
            sudo tee "$M_PATH/$(basename "$manifest")" >/dev/null
    done
}

# 安装K3S
_setup_k3s_env() {
    export INSTALL_K3S_SKIP_SELINUX_RPM=true
    export INSTALL_K3S_SELINUX_WARN=true
    export INSTALL_K3S_MIRROR=cn
    export INSTALL_K3S_MIRROR_URL=https://rancher-mirror.cdn.w7.cc
}

k3sInstallServer() {
    info "current server's public network ip: $(publicNetworkIp)"

    # 初始化公共变量
    _setup_k3s_env

    # Server 特有变量
    export K3S_NODE_NAME="${K3S_NODE_NAME:-server1}"
    export K3S_KUBECONFIG_MODE="644"

    # HA模式追加变量
    if [ -n "$K3S_URL" ] && [ -n "$K3S_TOKEN" ]; then
        export K3S_URL="$K3S_URL"
        export K3S_TOKEN="$K3S_TOKEN"
    fi

    curl -sfL https://rancher-mirror.cdn.w7.cc/k3s/k3s-install.sh | \
    sh -s - server \
        --snapshotter "${K3S_SNAPSHOTTER:-overlayfs}" \
        --tls-san "$(internalIP)" \
        --kubelet-arg="image-gc-high-threshold=70" \
        --kubelet-arg="image-gc-low-threshold=60" \
        --embedded-registry \
        --flannel-backend "none" \
        --disable-network-policy \
        --disable-kube-proxy \
        --disable "traefik"
}

k3sInstallAgent() {
    info "current server's public network ip: $(publicNetworkIp)"

    # 初始化公共变量
    _setup_k3s_env

    # HA模式追加变量
    if [ -n "$K3S_URL" ] && [ -n "$K3S_TOKEN" ]; then
        export K3S_URL="$K3S_URL"
        export K3S_TOKEN="$K3S_TOKEN"
    fi

    curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | \
    sh -s - agent --snapshotter "${K3S_SNAPSHOTTER:-overlayfs}"
}

# 检测multipath并配置blacklist
checkMultipathAndBlacklist() {
    # 检查 multipath 是否安装
    if ! command -v multipath >/dev/null 2>&1; then
        info "未安装 multipath，跳过配置"
        return 0
    fi

    info "检测到 multipath 已安装，开始配置黑名单"
    
    # 确保配置目录存在
    sudo mkdir -p /etc/multipath/conf.d
    
    # 黑名单配置内容（根据 longhorn 官方建议）
    local blacklist_conf="/etc/multipath/conf.d/99-longhorn.conf"
    local config_content=$(cat << 'EOF'
blacklist {
    devnode "^sd[a-z0-9]+"
    devnode "^vd[a-z0-9]+"
}
EOF
    )

    # 写入配置文件
    echo "$config_content" | sudo tee "$blacklist_conf" > /dev/null || {
        fatal "无法写入 multipath 黑名单配置文件: $blacklist_conf"
    }

    # 重新加载配置
    if sudo systemctl is-active --quiet multipathd; then
        sudo systemctl reload multipathd || {
            warn "重新加载 multipathd 服务失败，尝试重启服务"
            sudo systemctl restart multipathd || fatal "重启 multipathd 服务失败"
        }
    else
        # 如果服务未运行，尝试启动
        sudo systemctl start multipathd || {
            warn "启动 multipathd 服务失败，将在系统重启后生效配置"
        }
    fi

    success "multipath 黑名单配置已完成"
}

# 检测并处理SELinux状态
handleSELinux() {
    if command -v getenforce >/dev/null 2>&1; then
        local selinux_status=$(getenforce)
        if [ "$selinux_status" = "Enforcing" ] || [ "$selinux_status" = "Permissive" ]; then
            info "当前SELinux状态: $selinux_status，需要关闭以避免影响k3s运行"
            
            # 临时关闭SELinux
            if sudo setenforce 0 >/dev/null 2>&1; then
                success "已临时关闭SELinux"
            else
                warn "临时关闭SELinux失败，将尝试永久关闭"
            fi
            
            # 永久关闭SELinux
            if [ -f /etc/selinux/config ]; then
                sudo sed -i 's/^SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
                sudo sed -i 's/^SELINUX=permissive/SELINUX=disabled/' /etc/selinux/config
                success "已配置SELinux永久关闭，系统重启后生效"
            else
                warn "/etc/selinux/config文件不存在，无法配置永久关闭SELinux"
            fi
        else
            info "SELinux已关闭，无需处理"
        fi
    else
        info "未检测到SELinux管理工具，跳过处理"
    fi
}

# 检测并处理防火墙状态
handleFirewall() {
    # 定义常见的防火墙服务
    local firewalld_services="firewalld ufw iptables firewalld.service ufw.service"
    local firewall_active=false
    local service_name=""
    
    # 检测是否有活跃的防火墙服务
    for service in $firewalld_services; do
        if command -v systemctl >/dev/null 2>&1; then
            if systemctl is-active --quiet "$service"; then
                firewall_active=true
                service_name="$service"
                break
            fi
        elif command -v service >/dev/null 2>&1; then
            if service "$service" status >/dev/null 2>&1; then
                firewall_active=true
                service_name="$service"
                break
            fi
        fi
    done
    
    # 如果防火墙活跃，则进行处理
    if [ "$firewall_active" = true ]; then
        info "检测到活跃的防火墙服务: $service_name，需要关闭以避免影响服务运行"
        
        # 临时关闭防火墙
        if command -v systemctl >/dev/null 2>&1; then
            if sudo systemctl stop "$service_name" >/dev/null 2>&1; then
                success "已临时关闭防火墙服务: $service_name"
            else
                warn "临时关闭防火墙服务 $service_name 失败，将尝试永久关闭"
            fi
        elif command -v service >/dev/null 2>&1; then
            if sudo service "$service_name" stop >/dev/null 2>&1; then
                success "已临时关闭防火墙服务: $service_name"
            else
                warn "临时关闭防火墙服务 $service_name 失败，将尝试永久关闭"
            fi
        fi
        
        # 永久关闭防火墙（禁止开机启动）
        if command -v systemctl >/dev/null 2>&1; then
            if sudo systemctl disable "$service_name" >/dev/null 2>&1; then
                success "已配置防火墙服务 $service_name 永久关闭，系统重启后生效"
            else
                warn "无法配置防火墙服务 $service_name 永久关闭"
            fi
        elif command -v chkconfig >/dev/null 2>&1; then
            if sudo chkconfig "$service_name" off >/dev/null 2>&1; then
                success "已配置防火墙服务 $service_name 永久关闭，系统重启后生效"
            else
                warn "无法配置防火墙服务 $service_name 永久关闭"
            fi
        fi
    else
        info "未检测到活跃的防火墙服务，无需处理"
    fi
}

# 系统检查
checkDependencies() {
    command -v curl >/dev/null || fatal "请先安装 curl"
    command -v wget >/dev/null || fatal "请先安装 wget"
    command -v ip >/dev/null || fatal "需要 iproute2 工具包"
    command -v sudo >/dev/null || fatal "请先安装 sudo"
    [ -d /run/systemd/system ] || fatal "安装需要使用 systemd 的 Linux 系统"
}

# 主执行函数
main() {
    
    checkDependencies
    INTERNAL_IP=$(internalIP)
    export INTERNAL_IP
    checkK3SInstalled
    handleSELinux
    handleFirewall
    
    etcSysctl
    etcPrivaterRegistry
    checkMultipathAndBlacklist

    if [ -n "$K3S_URL" ]; then
        if [ -z "$K3S_NODE_NAME" ]; then
            fatal "添加子节点时必须传入K3S_NODE_NAME参数"
        fi

        case "$K3S_NODE_NAME" in
            server*)
                k3sInstallServer
                ;;
            agent*)
                k3sInstallAgent
                ;;
            *)
                fatal "K3S_NODE_NAME必须包含server或agent关键字，当前值: $K3S_NODE_NAME"
                ;;
        esac

        tips "=================================================================="
        tips "公网IP: $(publicNetworkIp)"
        tips "内网IP: $(internalIP)"
        tips "微擎面板添加节点成功！"
        tips "=================================================================="
    else
        k3sInstallServer
        downloadResource
        importImages
        installHelmCharts
        etcSystemd
        checkW7panelInstalled
    
        tips "=================================================================="
        tips "公网地址: http://$(publicNetworkIp):9090"
        tips "内网地址: http://$(internalIP):9090"
        tips "微擎面板安装成功，请访问后台设置登录密码！"
        tips ""
        warn "如果您的面板无访问："
        warn "请确认服务器安全组是否放通 (80|443|6443|9090) 端口"
        tips "=================================================================="
    fi
}

main "$@"
