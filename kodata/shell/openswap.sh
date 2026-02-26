#!/bin/sh

setupZram() {
    # 检测 Swap 交换空间并删除
    non_zram_swap=$(grep -E '^[^#].*\sswap\s' /etc/fstab | awk '{print $1}')
    if [ -n "$non_zram_swap" ]; then
        echo "检测到 Swap 交换空间，开始删除..."
        for swap in $non_zram_swap; do
            # 检查交换空间是否已挂载
            if swapon --show | grep -q "^$swap"; then
                swapoff "$swap"
            fi
            if [ -f "$swap" ]; then
                rm "$swap"
            fi
            # 从 /etc/fstab 中删除对应的挂载信息
            temp_file=$(mktemp)
            grep -v "^$swap " /etc/fstab > "$temp_file"
            mv "$temp_file" /etc/fstab
        done
        echo "Swap 交换空间已删除"
    fi

    # 检查是否已经存在 zram 设备作为交换空间
    if ! swapon --show | grep -q '^/dev/zram'; then
        echo "未检测到 ZRAM Swap 空间，开始创建并设置 4GB 的 ZRAM Swap 空间..."
        # 加载 zram 模块
        modprobe zram num_devices=1

        # 设置 zram 设备的压缩算法为 lz4hc
        echo lz4hc > /sys/block/zram0/comp_algorithm 2>/dev/null || true

        # 设置 zram 设备的大小为 4GB
        echo 4G > /sys/block/zram0/disksize 2>/dev/null || true

        # 格式化 zram 设备为交换空间
        mkswap /dev/zram0 2>/dev/null || true

        # 启用 zram 设备作为交换空间
        swapon /dev/zram0

        echo "ZRAM Swap 空间已成功创建"
    else
        echo "已检测到 ZRAM Swap 空间，跳过创建步骤"
    fi
}
setupZram