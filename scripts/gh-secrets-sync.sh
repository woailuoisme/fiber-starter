#!/bin/bash

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: gh CLI tool not found. Please install it first: https://cli.github.com/"
    exit 1
fi

# Check if secrets.ini exists
if [ ! -f "secrets.ini" ]; then
    echo "Error: secrets.ini file not found."
    exit 1
fi

echo "Syncing GitHub Secrets from secrets.ini..."

# Read secrets.ini and set each secret
while IFS='=' read -r key value || [ -n "$key" ]; do
    # Skip comments and empty lines
    [[ "$key" =~ ^[[:space:]]*# ]] || [[ -z "$key" ]] && continue

    # 使用纯 Shell 参数展开去除首尾空格，避免在循环中重复生成 xargs 子进程
    key="${key#"${key%%[![:space:]]*}"}"
    key="${key%"${key##*[![:space:]]}"}"
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"

    # 如果清除空格后 key 为空则跳过
    [[ -z "$key" ]] && continue

    echo "Setting secret: $key"
    gh secret set "$key" --body "$value"
done < secrets.ini

echo "Sync completed successfully!"
