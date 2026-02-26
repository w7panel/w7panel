#!/bin/sh

stopSwap() {
    # 检测是否存在交换空间
    if [ $(swapon --show | wc -l) -le 1 ]; then
        echo "未检测到任何交换空间"
        return
    fi

    # 停用所有已挂载的交换空间
    swapoff -a
    echo "所有交换空间已停用"
}

stopSwap