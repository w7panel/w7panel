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

# 卸载内存压缩
cleanZram() {
    manage_systemd_service destroy --service-name "zram"
}

# 检测内存压缩是否开启
checkZram() {
    if swapon --show | grep -q '^/dev/zram'; then
        info "ZRAM 内存压缩已开启"
    else
        info "ZRAM 内存压缩未开启"
        exit 1
    fi
}

# 主函数，根据参数执行相应操作
main() {
    if [ $# -ne 1 ]; then
        fatal "用法: $0 [setup|clean|check]"
    fi

    case "$1" in
        "setup")
            setupZram
            ;;
        "clean")
            cleanZram
            ;;
        "check")
            checkZram
            ;;
        *)
            fatal "未知参数: $1, 可用参数: install, uninstall, check"
            ;;
    esac
}

# 调用主函数
main "$@" 