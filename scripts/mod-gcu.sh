#!/bin/sh
set -eu

usage() {
	cat <<'EOF'
Usage:
  scripts/mod-gcu.sh list
  scripts/mod-gcu.sh up
  scripts/mod-gcu.sh patch

Modes:
  list   Show direct dependencies with newer versions available
  up     Upgrade direct dependencies to @latest
  patch  Upgrade direct dependencies only when a newer patch version exists
EOF
}

mode="${1:-}"
if [ -z "$mode" ]; then
	usage
	exit 1
fi

case "$mode" in
	list)
		GOFLAGS=-mod=mod go list -m -u -f '{{if and (not .Indirect) .Update}}{{printf "%-40s %s -> %s\n" .Path .Version .Update.Version}}{{end}}' all
		;;
	up)
		mods="$(GOFLAGS=-mod=mod go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}{{end}}' all | sed '/^$/d')"
		if [ -z "$mods" ]; then
			echo "No direct dependencies found"
			exit 0
		fi
		for mod in $mods; do
			echo "Updating $mod..."
			GOFLAGS=-mod=mod go get "$mod@latest"
		done
		go mod tidy
		go mod vendor
		;;
	patch)
		mods="$(GOFLAGS=-mod=mod go list -m -u -f '{{if and (not .Main) (not .Indirect) .Update}}{{.Path}} {{.Version}} {{.Update.Version}}{{end}}' all | sed '/^$/d')"
		if [ -z "$mods" ]; then
			echo "No direct dependencies found"
			exit 0
		fi
		printf '%s\n' "$mods" | while read -r mod cur latest; do
			cur_base="${cur#v}"
			latest_base="${latest#v}"
			cur_major="$(printf '%s' "$cur_base" | cut -d. -f1)"
			cur_minor="$(printf '%s' "$cur_base" | cut -d. -f2)"
			cur_patch="$(printf '%s' "$cur_base" | cut -d. -f3)"
			latest_major="$(printf '%s' "$latest_base" | cut -d. -f1)"
			latest_minor="$(printf '%s' "$latest_base" | cut -d. -f2)"
			latest_patch="$(printf '%s' "$latest_base" | cut -d. -f3)"

			if [ -z "$cur_major" ] || [ -z "$cur_minor" ] || [ -z "$cur_patch" ] || [ -z "$latest_major" ] || [ -z "$latest_minor" ] || [ -z "$latest_patch" ]; then
				echo "Skipping $mod $cur -> $latest (unsupported version format)"
				continue
			fi

			if [ "$cur_major" = "$latest_major" ] && [ "$cur_minor" = "$latest_minor" ] && [ "$latest_patch" -gt "$cur_patch" ]; then
				echo "Updating $mod $cur -> $latest"
				GOFLAGS=-mod=mod go get "$mod@$latest"
			else
				echo "Skipping $mod $cur -> $latest (not a patch update)"
			fi
		done
		go mod tidy
		go mod vendor
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
