#!/bin/sh
# ==============================================================================
# Go Module 重命名与项目克隆实例化脚本 (go-mod-rename.sh)
# ==============================================================================
#
# 1. 主要功能 (Core Features):
#    - 【就地重命名 (`rename` 模式)】:
#      - 自动提取当前项目 `go.mod` 的现有 module 命名（作为默认旧模块名）。
#      - 修改 `go.mod` 的模块声明并扫描所有受 Git 追踪的 Go 源码/文本文件。
#      - 优先利用高速的 `sd` 工具（降级支持 `perl`）将代码中所有的旧模块路径引用就地批量替换为新模块路径。
#      - 最终调用 `go mod tidy` 重整依赖关系，确保代码开箱即用。
#    - 【项目实例化克隆 (`copy` 模式)】:
#      - 调用 Go 官方的模块模板实例化工具 `gonew`，从一个已有的 Go 模块/模板（如本项目）
#        克隆并初始化为一个全新的独立项目目录及自定义的 module 路径。
#
# 2. 命令行用法 (Usage):
#    - 重命名当前项目:  ./scripts/go-mod-rename.sh rename <new-module> [old-module]
#    - 模板克隆新项目:  ./scripts/go-mod-rename.sh copy <srcmod> <dstmod> [dir]
#
# 3. 输出与影响 (Outputs & Side Effects):
#    - `rename` 模式下会直接就地修改 `go.mod` 和本地受 Git 追踪的相关源码文件。
#    - `copy` 模式下会在指定路径实例化生成包含新 module 构建链的物理文件夹。
# ==============================================================================
set -eu

usage() {
	cat <<'EOF'
Usage:
  scripts/go-mod-rename.sh rename <new-module> [old-module]
  scripts/go-mod-rename.sh copy <srcmod> <dstmod> [dir]

Modes:
  rename  Rewrite the current repository in place to use a new module path.
  copy    Use gonew to clone a template module into a new directory.

Notes:
  - rename updates go.mod, tracked source files, and text references to the old module path.
  - copy requires gonew: go install golang.org/x/tools/cmd/gonew@latest
EOF
}

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repo_root"

if [ -n "${GOFLAGS:-}" ]; then
	export GOFLAGS="$GOFLAGS -mod=mod"
else
	export GOFLAGS="-mod=mod"
fi

mode="${1:-}"
if [ -z "$mode" ]; then
	usage
	exit 1
fi

case "$mode" in
	rename)
		new_module="${2:-}"
		old_module="${3:-}"
		if [ -z "$new_module" ]; then
			usage
			exit 1
		fi
		if [ ! -f go.mod ]; then
			echo "go.mod not found in $repo_root" >&2
			exit 1
		fi
		if [ -z "$old_module" ]; then
			old_module=$(awk 'NR==1 && $1=="module" {print $2; exit}' go.mod)
		fi
		if [ -z "$old_module" ]; then
			echo "Failed to detect current module path from go.mod" >&2
			exit 1
		fi
		if [ "$old_module" = "$new_module" ]; then
			echo "Module already set to $new_module"
			exit 0
		fi

		go mod edit -module "$new_module"

		# 使用 git grep -lFz 以 null 字节分隔输出，防止特殊文件名解析出错
		files=$(git grep -lFz --full-name -- "$old_module" 2>/dev/null || true)
		if [ -n "$files" ]; then
			# 优先使用高效的 Rust 工具 sd 进行批量文本替换；若未安装，平滑降级回退到 perl
			if command -v sd >/dev/null 2>&1; then
				printf '%s' "$files" | xargs -0 sd -F "$old_module" "$new_module"
			else
				printf '%s' "$files" | tr '\0' '\n' | while IFS= read -r file; do
					[ -n "$file" ] || continue
					OLD_MODULE="$old_module" NEW_MODULE="$new_module" perl -0pi -e 's/\Q$ENV{OLD_MODULE}\E/$ENV{NEW_MODULE}/g' "$file"
				done
			fi
		fi

		go mod tidy
		echo "Module renamed from $old_module to $new_module"
		;;
	copy)
		srcmod="${2:-}"
		dstmod="${3:-}"
		dir="${4:-}"
		if [ -z "$srcmod" ] || [ -z "$dstmod" ]; then
			usage
			exit 1
		fi
		if ! command -v gonew >/dev/null 2>&1; then
			echo "gonew is not installed. Run: go install golang.org/x/tools/cmd/gonew@latest" >&2
			exit 1
		fi
		if [ -n "$dir" ]; then
			gonew "$srcmod" "$dstmod" "$dir"
		else
			gonew "$srcmod" "$dstmod"
		fi
		;;
	-h|--help|help)
		usage
		;;
	*)
		echo "Unknown mode: $mode" >&2
		usage >&2
		exit 1
		;;
esac
