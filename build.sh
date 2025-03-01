#!/bin/bash

# 设置脚本在遇到错误时退出
set -e

OUTPUT_DIR="output"

# 创建输出目录（如果不存在）
mkdir -p "${OUTPUT_DIR}"

# 编译项目
echo "Building ProxyServer..."
go build -C ./ProxyServer -o "../${OUTPUT_DIR}" ./...

# 检查编译是否成功
if [ -f "output/proxy-server" ]; then
    echo "Build successful! Output file: ${OUTPUT_DIR}/proxy-server"
else
    echo "Build failed!"
    exit 1
fi

# 编译项目
echo "Building ProxyAdmin..."
go build -C ./ProxyAdmin -o "../${OUTPUT_DIR}" ./...

# 检查编译是否成功
if [ -f "output/proxy-admin" ]; then
    echo "Build successful! Output file: ${OUTPUT_DIR}/proxy-admin"
else
    echo "Build failed!"
    exit 1
fi
