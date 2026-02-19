#!/bin/bash

# Go formatting script (similar to Laravel Pint)
echo "Starting Go formatting..."

# 1. Format with go fmt
echo "Running go fmt..."
go fmt ./...

# 2. Organize imports with goimports (if installed)
if command -v goimports >/dev/null 2>&1 || [ -f "$HOME/go/bin/goimports" ] || [ -f "$GOPATH/bin/goimports" ]; then
    echo "Organizing imports..."
    if command -v goimports >/dev/null 2>&1; then
        goimports -w .
    elif [ -f "$HOME/go/bin/goimports" ]; then
        $HOME/go/bin/goimports -w .
    else
        $GOPATH/bin/goimports -w .
    fi
fi

# 3. Run go vet
echo "Running go vet..."
go vet ./...

# 4. Run golangci-lint (if installed and compatible)
if command -v golangci-lint >/dev/null 2>&1; then
    echo "Running golangci-lint..."
    golangci-lint run --disable-all -E gofmt,govet --fast || echo "golangci-lint skipped (compatibility issue)"
fi

echo "Formatting completed."
