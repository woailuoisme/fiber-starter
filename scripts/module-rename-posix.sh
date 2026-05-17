#!/bin/sh
set -eu

usage() {
	cat <<'EOF'
Usage:
  scripts/module-rename-posix.sh rename <new-module> [old-module]

Description:
  Rewrite the current repository in place to use a new Go module path.

Notes:
  - Updates go.mod first.
  - Rewrites tracked text files that contain the old module path.
  - Uses only POSIX sh plus awk and git.
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

cleanup_files=0
cleanup() {
	if [ "${cleanup_files:-0}" -eq 1 ] && [ -n "${listfile:-}" ] && [ -f "$listfile" ]; then
		rm -f "$listfile"
	fi
}
trap cleanup EXIT HUP INT TERM

rewrite_file() {
	file=$1
	old_module=$2
	new_module=$3
	tmpfile=$(mktemp "${TMPDIR:-/tmp}/module-rename.XXXXXX")

	if awk -v old="$old_module" -v new="$new_module" '
		function replace_all(line,    out, rest, pos) {
			out = ""
			rest = line
			while ((pos = index(rest, old)) > 0) {
				out = out substr(rest, 1, pos - 1) new
				rest = substr(rest, pos + length(old))
			}
			return out rest
		}
		{
			print replace_all($0)
		}
	' "$file" > "$tmpfile"; then
		mv "$tmpfile" "$file"
	else
		rm -f "$tmpfile"
		return 1
	fi
}

mode="${1:-}"
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
			listfile=$(mktemp "${TMPDIR:-/tmp}/module-rename-files.XXXXXX")
			cleanup_files=1
			printf '%s\n' "$files" > "$listfile"
			while IFS= read -r file; do
				[ -n "$file" ] || continue
				rewrite_file "$file" "$old_module" "$new_module"
			done < "$listfile"
		fi

		go mod tidy
		echo "Module renamed from $old_module to $new_module"
		;;
	-h|--help|help|"")
		usage
		;;
	*)
		echo "Unknown mode: $mode" >&2
		usage >&2
		exit 1
		;;
esac
