#!/bin/bash

set -e

# Go formatting script (similar to Laravel Pint)
echo "Starting Go formatting..."

# 1. Format with gofumpt
if command -v gofumpt >/dev/null 2>&1; then
    echo "Running gofumpt..."
    gofumpt -w $(command -v rg >/dev/null 2>&1 && rg --files -g '*.go' || find . -name '*.go' -not -path './vendor/*' -not -path './.git/*')
else
    echo "gofumpt is not installed"
    exit 1
fi

# 2. Format with golangci-lint
if command -v golangci-lint >/dev/null 2>&1; then
    echo "Running golangci-lint fmt..."
    golangci-lint fmt
else
    echo "golangci-lint is not installed"
    exit 1
fi

# 3. Run golangci-lint lint checks
echo "Running golangci-lint run..."
golangci-lint run

echo "Formatting completed."
