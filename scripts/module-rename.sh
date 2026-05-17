#!/bin/sh
set -eu

usage() {
	cat <<'EOF'
Usage:
  scripts/module-rename.sh rename <new-module> [old-module]
  scripts/module-rename.sh copy <srcmod> <dstmod> [dir]

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

		files=$(git grep -lF --full-name -- "$old_module" 2>/dev/null || true)
		if [ -n "$files" ]; then
			printf '%s\n' "$files" | while IFS= read -r file; do
				[ -n "$file" ] || continue
				OLD_MODULE="$old_module" NEW_MODULE="$new_module" perl -0pi -e 's/\Q$ENV{OLD_MODULE}\E/$ENV{NEW_MODULE}/g' "$file"
			done
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
