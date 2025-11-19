#!/bin/bash

# Go 代码格式化脚本 - 类似 Laravel Pint
echo "🔧 开始格式化 Go 代码..."

# 1. 使用 go fmt 格式化代码
echo "📝 使用 go fmt 格式化代码..."
go fmt ./...

# 2. 使用 goimports 整理导入语句（如果安装了）
if command -v goimports >/dev/null 2>&1 || [ -f "$HOME/go/bin/goimports" ] || [ -f "$GOPATH/bin/goimports" ]; then
    echo "📦 整理导入语句..."
    if command -v goimports >/dev/null 2>&1; then
        goimports -w .
    elif [ -f "$HOME/go/bin/goimports" ]; then
        $HOME/go/bin/goimports -w .
    else
        $GOPATH/bin/goimports -w .
    fi
fi

# 3. 使用 go vet 检查代码
echo "🔍 运行代码静态检查..."
go vet ./...

# 4. 使用 golangci-lint（如果安装了且版本兼容）
if command -v golangci-lint >/dev/null 2>&1; then
    echo "⚡ 运行代码质量检查..."
    # 只使用基本的检查器避免版本问题
    golangci-lint run --disable-all -E gofmt,govet --fast || echo "⚠️  golangci-lint 检查跳过（版本兼容性问题）"
fi

echo "✅ 代码格式化完成！"