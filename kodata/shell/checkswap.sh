#!/bin/sh

# if [ $(swapon --show | wc -l) -le 1 ]; then
#         echo "未开启"
#         exit 1
# else
#         echo "开启了"
#         exit 0
# fi


if swapon --show | grep -q '^/dev/zram'; then
        echo "ZRAM 内存压缩已开启"
        exit 0
    else
        echo "ZRAM 内存压缩未开启"
        exit 1
fi