#!/bin/bash


sudo apt-get update && apt-get install --no-install-recommends -y \
       apt-transport-https \
       ca-certificates \
       curl \
       gnupg-agent \
       software-properties-common

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo sudo apt-key add -

sudo apt-key fingerprint 0EBFCD88

sudo add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"
sudo apt-get update && sudo apt-get install --no-install-recommends -y docker-ce docker-ce-cli containerd.io       