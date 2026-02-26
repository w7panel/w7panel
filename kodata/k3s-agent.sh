#!/bin/bash
set -e

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
    local path=".$(echo "$url" | sed -E 's|^https?://[^/]+||')"
    local save_path=$(dirname "$path")
    local filename=$(basename "$path")
    sudo mkdir -p "$save_path" || fatal "创建目录失败: $save_path"

    if [ -f "$path" ]; then
        info "文件已存在: ${path}"
        return 0
    fi

    # 重试配置参数
    local max_retries=10
    local retry_delay=5
    local retry_count=0

    while [ $retry_count -le $max_retries ]; do
        # 最后一次重试不显示特殊提示
        if [ $retry_count -gt 0 ]; then
            info "正在尝试第 ${retry_count}/${max_retries} 次重试..."
        fi

        if sudo wget -q --show-progress --progress=bar:force:noscroll --no-check-certificate \
            -c --timeout=10 --tries=10 --retry-connrefused \
            -O "$path" "$url"; then
            return 0
        fi

        # 重试前清理失败的文件
        [ -f "$path" ] && sudo rm -f "$path"
        
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
        https://cdn.w7.cc/w7panel/images/cilium.cilium-v1.16.4.tar
        https://cdn.w7.cc/w7panel/images/cilium.operator-generic-v1.16.4.tar
        https://cdn.w7.cc/w7panel/images/jetstack.cert-manager-cainjector-v1.16.2.tar
        https://cdn.w7.cc/w7panel/images/jetstack.cert-manager-controller-v1.16.2.tar
        https://cdn.w7.cc/w7panel/images/jetstack.cert-manager-webhook-v1.16.2.tar
        https://cdn.w7.cc/w7panel/images/jetstack.cert-manager-startupapicheck-v1.16.2.tar
        https://cdn.w7.cc/w7panel/manifests/cert-manager.yaml
        https://cdn.w7.cc/w7panel/manifests/cilium.yaml
        https://cdn.w7.cc/w7panel/manifests/higress.yaml
        https://cdn.w7.cc/w7panel/manifests/longhorn.yaml
        https://cdn.w7.cc/w7panel/manifests/w7panel-offline.yaml
    "

    for resource in $resources; do
        download_files "$resource"
    done
}

# 获取公网IP
publicNetworkIp() {
    # publicIp 为空，则重新获取publicIp
    if [ -z "$PUBLIC_IP" ]; then
        PUBLIC_IP=$(curl -s ifconfig.me)
        echo "$PUBLIC_IP"
    else
        echo "$PUBLIC_IP"
    fi
}

# 获取内网IP（兼容多系统）
internalIP() {
    if [ -z "$INTERNAL_IP" ]; then
        INTERNAL_IP=$(
            ip -o -4 addr show | \
            awk '{print $4}' | \
            grep -v '127.0.0.1' | \
            cut -d/ -f1 | \
            head -1
        )
        [ -z "$INTERNAL_IP" ] && INTERNAL_IP=$(hostname -I | awk '{print $1}')
        echo "$INTERNAL_IP"
    else
        echo "$INTERNAL_IP"
    fi
}

# 处理sysctl配置
etcSysctl() {
    # 下载资源文件
    download_files "https://cdn.w7.cc/w7panel/etc/sysctl.d/k3s.conf"
    
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
    download_files "https://cdn.w7.cc/w7panel/etc/registries.yaml"
    
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
    # 下载资源文件
    download_files "https://cdn.w7.cc/w7panel/etc/systemd/k3s.service.env"
    
    if systemctl is-active --quiet k3s; then
        local service_name="k3s"
        local env_file="k3s.service.env"
    else
        local service_name="k3s-agent"
        local env_file="k3s-agent.service.env"
    fi

    local ETC_PATH="/etc/systemd/system/"
    local source_env="./w7panel/etc/systemd/k3s.service.env"

    # 处理环境变量文件
    if [ -f "$source_env" ]; then
        cat "$source_env" | sudo tee -a "${ETC_PATH}/${env_file}" > /dev/null || {
            fatal "Failed to append content to ${ETC_PATH}/${env_file}"
        }
    fi

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
        if systemctl is-active --quiet k3s; then
            local type="Server"
            local uninstall="/usr/local/bin/k3s-uninstall.sh"
        else
            local type="Agent"
            local uninstall="/usr/local/bin/k3s-agent-uninstall.sh"
        fi
        
        warn "K3s $type has been installed , Please execute $uninstall to uninstall k3s "
        warn "K3s $type 已安装 , 请先执行 $uninstall 命令卸载 "
        exit
    fi
}

# 检查微擎面板是否安装成功
checkW7panelInstalled() {
    printf "${INFO_COLOR}[INFO]${NC} %s" "微擎面板正在初始化，预计需要3-5分钟，请耐心等待..."
    local spinpid
    while true; do
        curl -s --max-time 2 -I "http://$(internalIP):9090" | grep -q "HTTP/" && break
        sleep 1 &
        show_progress $! &
        spinpid=$!
        wait $spinpid
    done
    echo
}

# 导入镜像
importImages() {
    local IMAGES_DIR="./w7panel/images"
    [ ! -d "$IMAGES_DIR" ] && return 0

    local total=$(ls $IMAGES_DIR/*.tar 2>/dev/null | wc -l)
    local count=0
    for IMAGE_FILE in $IMAGES_DIR/*.tar; do
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
    sudo cp -r "./w7panel/manifests/." "$M_PATH" || {
        fatal "Failed to copy manifests to $M_PATH"
        return 1
    }
}

# 安装K3S
k3sInstall() {
    info "current server's public network ip: $(publicNetworkIp)"
    curl -sfL https://rancher-mirror.cdn.w7.cc/k3s/k3s-install.sh | \
    K3S_NODE_NAME=server1 K3S_KUBECONFIG_MODE='644' INSTALL_K3S_SKIP_SELINUX_RPM=true INSTALL_K3S_SELINUX_WARN=true INSTALL_K3S_MIRROR=cn INSTALL_K3S_MIRROR_URL=rancher-mirror.cdn.w7.cc \
    sh -s - --write-kubeconfig-mode 644 \
        --tls-san "$(internalIP)" \
        --kubelet-arg="image-gc-high-threshold=70" \
        --kubelet-arg="image-gc-low-threshold=60" \
        --node-label "w7.public-ip=$(publicNetworkIp)" \
        --embedded-registry \
        --flannel-backend "none" \
        --disable-network-policy \
        --disable-kube-proxy \
        --disable "local-storage,traefik"
}

k3sInstallAgent() {
   info "current server's public network ip: $(publicNetworkIp)"
   curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | \
   K3S_URL=${K3S_URL} K3S_TOKEN=${K3S_TOKEN} INSTALL_K3S_SKIP_SELINUX_RPM=true INSTALL_K3S_SELINUX_WARN=true INSTALL_K3S_MIRROR=cn INSTALL_K3S_MIRROR_URL=rancher-mirror.cdn.w7.cc \
   sh -s - --node-label "w7.public-ip=$(publicNetworkIp)"
}

# 启动服务管理
manage_systemd_service() {
    local mode="$1"
    local service_name=""
    local description=""
    local exec_start_pre=""
    local exec_start=""
    local exec_start_post=""
    local exec_reload=""
    local exec_stop=""
    local exec_stop_pre=""
    local exec_stop_post=""
    local environment_files=""
    local service_type="simple"

    shift
    while [ $# -gt 0 ]; do
        case "$1" in
            --service-name)
                service_name="$2"
                shift 2
                ;;
            --description)
                description="$2"
                shift 2
                ;;
            --exec-start-pre)
                if [ -n "$exec_start_pre" ]; then
                    exec_start_pre="${exec_start_pre}\nExecStartPre=$2"
                else
                    exec_start_pre="ExecStartPre=$2"
                fi
                shift 2
                ;;
            --exec-start)
                if [ -n "$exec_start" ]; then
                    exec_start="${exec_start}\nExecStart=$2"
                else
                    exec_start="ExecStart=$2"
                fi
                shift 2
                ;;
            --exec-start-post)
                if [ -n "$exec_start_post" ]; then
                    exec_start_post="${exec_start_post}\nExecStartPost=$2"
                else
                    exec_start_post="ExecStartPost=$2"
                fi
                shift 2
                ;;
            --exec-reload)
                if [ -n "$exec_reload" ]; then
                    exec_reload="${exec_reload}\nExecReload=$2"
                else
                    exec_reload="ExecReload=$2"
                fi
                shift 2
                ;;
            --exec-stop)
                if [ -n "$exec_stop" ]; then
                    exec_stop="${exec_stop}\nExecStop=$2"
                else
                    exec_stop="ExecStop=$2"
                fi
                shift 2
                ;;
            --exec-stop-pre)
                if [ -n "$exec_stop_pre" ]; then
                    exec_stop_pre="${exec_stop_pre}\nExecStopPre=$2"
                else
                    exec_stop_pre="ExecStopPre=$2"
                fi
                shift 2
                ;;
            --exec-stop-post)
                if [ -n "$exec_stop_post" ]; then
                    exec_stop_post="${exec_stop_post}\nExecStopPost=$2"
                else
                    exec_stop_post="ExecStopPost=$2"
                fi
                shift 2
                ;;
            --environment-file)
                if [ -n "$environment_files" ]; then
                    environment_files="${environment_files}\nEnvironmentFile=-$2"
                else
                    environment_files="EnvironmentFile=-$2"
                fi
                shift 2
                ;;
            --type)
                service_type="$2"
                shift 2
                ;;
            *)
                fatal "未知参数: $1"
                ;;
        esac
    done

    case "$mode" in
        "create")
            if [ -z "$service_name" ] || [ -z "$description" ] || [ -z "$exec_start" ]; then
                fatal "缺少必要参数: --service-name, --description, --exec-start"
            fi

            service_content="[Unit]
Description=$description
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=$service_type
"
            if [ "$service_type" = "oneshot" ]; then
                service_content="${service_content}RemainAfterExit=true
"
            else
                service_content="${service_content}KillMode=process
Delegate=yes
TimeoutStartSec=0
Restart=always
RestartSec=5s
"
            fi

            if [ -n "$environment_files" ]; then
                service_content="${service_content}${environment_files}
"
            fi
            if [ -n "$exec_start_pre" ]; then
                service_content="${service_content}${exec_start_pre}
"
            fi
            if [ -n "$exec_start" ]; then
                service_content="${service_content}${exec_start}
"
            fi
            if [ -n "$exec_start_post" ]; then
                service_content="${service_content}${exec_start_post}
"
            fi
            if [ -n "$exec_reload" ]; then
                service_content="${service_content}${exec_reload}
"
            fi
            if [ -n "$exec_stop_pre" ]; then
                service_content="${service_content}${exec_stop_pre}
"
            fi
            if [ -n "$exec_stop" ]; then
                service_content="${service_content}${exec_stop}
"
            fi
            if [ -n "$exec_stop_post" ]; then
                service_content="${service_content}${exec_stop_post}
"
            fi

            service_file_path="/etc/systemd/system/${service_name}.service"
            printf "%b" "$service_content" | sudo tee "$service_file_path" > /dev/null
            if [ $? -eq 0 ]; then
                success "成功创建 ${service_name}.service 文件"
                sudo systemctl daemon-reload
                sudo systemctl enable "${service_name}.service"
                if [ $? -eq 0 ]; then
                    success "成功设置 ${service_name} 开机启动"
                else
                    warn "设置 ${service_name} 开机启动时出错"
                fi
            else
                fatal "创建服务文件时出错"
            fi
            ;;
        "destroy")
            if [ -z "$service_name" ]; then
                fatal "缺少必要参数: --service-name"
            fi
            service_file_path="/etc/systemd/system/${service_name}.service"
            if [ -f "$service_file_path" ]; then
                sudo systemctl stop "${service_name}.service"
                sudo systemctl disable "${service_name}.service"
                sudo rm "$service_file_path"
                sudo systemctl daemon-reload
                success "成功销毁 ${service_name}.service 文件"
            else
                warn "${service_name}.service 文件不存在"
            fi
            ;;
        *)
            fatal "未知模式: $mode。可用模式: create, destroy"
            ;;
    esac
}

# 检测并安装 zram 模块
check_and_install_zram() {
    # 先尝试加载模块
    sudo modprobe zram num_devices=1 > /dev/null 2>&1 || true
    
    if ! lsmod | grep -q zram; then
        info "未检测到 zram 内核模块，尝试安装..."
        # 检测发行版
        if [ -f /etc/redhat-release ]; then
            # Red Hat 系（如 CentOS、Fedora）
            sudo yum update -y
            sudo yum install -y kernel-modules-extra
        elif [ -f /etc/debian_version ]; then
            # Debian 系（如 Ubuntu、Debian）
            sudo apt-get update -y
            sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -y -o Dpkg::Options::="--force-confold"
            sudo DEBIAN_FRONTEND=noninteractive apt-get install -y linux-modules-extra-$(uname -r)
        elif [ -f /etc/arch-release ]; then
            # Arch Linux
            sudo pacman -Syu --noconfirm
            sudo pacman -S --noconfirm linux-headers
        else
            fatal "无法识别的发行版，请手动安装 zram 模块"
        fi
        # 再次尝试加载模块
        sudo modprobe zram num_devices=1
        if [ $? -ne 0 ]; then
            fatal "安装后仍无法加载 zram 模块，请手动检查"
        fi
        success "zram 模块已成功加载"
    else
        info "zram 内核模块已存在"
    fi
}

# 创建内存压缩
setupZram() {
    # 检测并安装 zram 模块
    check_and_install_zram || return 1

    if [ ! -f /etc/systemd/system/zram.service ]; then
        # 创建 ZRAM Swap Service
        manage_systemd_service create \
        --service-name "zram" \
        --description "ZRAM Swap Service" \
        --exec-start-pre "/sbin/modprobe -r zram" \
        --exec-start-pre "/sbin/modprobe zram num_devices=1" \
        --exec-start-pre "-/bin/sh -c 'echo lz4hc > /sys/block/zram0/comp_algorithm'" \
        --exec-start-pre "/bin/sh -c 'echo 4G > /sys/block/zram0/disksize'" \
        --exec-start-pre "/sbin/mkswap /dev/zram0" \
        --exec-start "/sbin/swapon -p 100 /dev/zram0" \
        --exec-stop "/sbin/swapoff /dev/zram0" \
        --exec-stop-post "/sbin/wipefs -a /dev/zram0" \
        --type oneshot

        # 启动服务
        sudo systemctl start zram.service
    else
        # 重启服务
        sudo systemctl restart zram.service  
    fi
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
}

# 主执行函数
main() {
    process_args "$@"
    
    checkDependencies
    handleSELinux
    handleFirewall
    checkK3SInstalled
    
    etcSysctl
    etcPrivaterRegistry
    setupZram
    checkMultipathAndBlacklist

    if [ -n "$K3S_URL" ]; then
        k3sInstallAgent
    else
        k3sInstall
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
