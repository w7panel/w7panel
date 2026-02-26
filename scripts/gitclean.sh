#!/bin/bash

# 下载https://gitee.com/oschina/git-repo-clean/releases/download/v1.4.2/git-repo-clean-1.4.2-Linux-64.tar 解压保存的/usr/local/bin/git-repo-clean

wget https://gitee.com/oschina/git-repo-clean/releases/download/v1.4.2/git-repo-clean-1.4.2-Linux-64.tar -O /tmp/git-repo-clean.tar
tar -xf /tmp/git-repo-clean.tar -C /tmp
sudo mv /tmp/releases/1.4.2/Linux-64/git-repo-clean /usr/local/bin/