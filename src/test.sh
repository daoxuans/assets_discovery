#!/bin/bash

# 资产发现系统快速测试脚本

set -e

echo "=== 被动式网络资产识别与分析系统 - 快速测试 ==="
echo

# 进入项目目录
cd "$(dirname "$0")"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未安装Go环境，请先安装Go 1.21+"
    exit 1
fi

echo "1. 编译演示程序..."
cd demo
go build -o demo main.go
echo "✓ 演示程序编译完成"

echo
echo "2. 生成示例资产数据..."
./demo generate
echo "✓ 示例数据生成完成"

echo
echo "3. 分析本地网络接口..."
./demo analyze
echo "✓ 本地网络分析完成"

echo
echo "4. 显示统计信息..."
./demo stats

echo
echo "5. 查看生成的文件..."
ls -la *.json

echo
echo "=== 测试完成 ==="
echo
echo "下一步:"
echo "1. 查看生成的JSON文件了解资产数据格式"
echo "2. 运行完整的安装脚本: ../install.sh"
echo "3. 编译完整版本: cd .. && make build"
echo "4. 阅读README.md了解详细用法"
