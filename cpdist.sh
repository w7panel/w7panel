#!/bin/bash

rm -rf ./asset/assets
rm -rf ./asset/index.html

cp -r ../w7panel-ui/dist/assets ./asset/

cp -r ../w7panel-ui/dist/index.html ./asset/micro.html
cp -r ../w7panel-ui/dist/index.html ./asset/index.html
# 替换index.html中的assets路径 ./ 改为 /
sed -i 's|./assets|/assets|g' asset/index.html
