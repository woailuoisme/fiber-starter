#!/bin/sh
# mod-gcu.sh — Go Check/Upgrade direct dependencies (inspired by npm-check-updates)
# Usage: scripts/mod-gcu.sh {list|up|patch}
set -eu

GOFLAGS="-mod=mod"
export GOFLAGS

# ── colors (only when stdout is a TTY) ───────────────────────────────────────

if [ -t 1 ]; then
	RESET=$(printf '\033[0m')
	BOLD=$(printf '\033[1m')
	CYAN=$(printf '\033[0;36m')
	GREEN=$(printf '\033[0;32m')
	YELLOW=$(printf '\033[0;33m')
	RED=$(printf '\033[0;31m')
	DIM=$(printf '\033[2m')
else
	RESET='' BOLD='' CYAN='' GREEN='' YELLOW='' RED='' DIM=''
fi

info()  { printf "${CYAN}  info${RESET}  %s\n" "$*"; }
ok()    { printf "${GREEN}    ok${RESET}  %s\n" "$*"; }
warn()  { printf "${YELLOW}  skip${RESET}  %s\n" "$*"; }
bump()  { printf "${GREEN}  bump${RESET}  %s\n" "$*"; }
die()   { printf "${RED} error${RESET}  %s\n" "$*" >&2; exit 1; }

# ── helpers ───────────────────────────────────────────────────────────────────

usage() {
	cat <<EOF
${BOLD}Usage:${RESET}
  scripts/mod-gcu.sh list   Show direct dependencies with newer versions
  scripts/mod-gcu.sh up     Upgrade direct dependencies to @latest
  scripts/mod-gcu.sh patch  Upgrade direct dependencies (patch-only)

  -h, --help    Show this help
EOF
}

# Strip "v" prefix and pre-release/build-metadata: v1.2.3+incompatible → 1.2.3
strip_semver() { printf '%s' "$1" | sed 's/^v//; s/[-+].*//'; }
semver_field() { printf '%s' "$1" | cut -d. -f"$2"; }

list_direct() {
	go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}{{end}}' all \
		| sed '/^$/d'
}

list_direct_updates() {
	go list -m -u \
		-f '{{if and (not .Main) (not .Indirect) .Update}}{{.Path}} {{.Version}} {{.Update.Version}}{{end}}' \
		all | sed '/^$/d'
}

# ── list ──────────────────────────────────────────────────────────────────────

cmd_list() {
	info "Checking for updates..."
	updates="$(list_direct_updates)"
	if [ -z "$updates" ]; then
		ok "All direct dependencies are up to date."
		return
	fi
	printf "\n${BOLD}%-44s  %-22s  %s${RESET}\n" "MODULE" "CURRENT" "LATEST"
	printf "${DIM}%-44s  %-22s  %s${RESET}\n"    "------" "-------" "------"
	printf '%s\n' "$updates" | while read -r mod cur latest; do
		printf "${CYAN}%-44s${RESET}  ${DIM}%-22s${RESET}  ${GREEN}→ %s${RESET}\n" \
			"$mod" "$cur" "$latest"
	done
	printf "\n"
}

# ── up ────────────────────────────────────────────────────────────────────────

cmd_up() {
	mods="$(list_direct)"
	if [ -z "$mods" ]; then
		warn "No direct dependencies found."
		return
	fi
	n="$(printf '%s\n' "$mods" | wc -l | tr -d ' ')"
	info "Upgrading ${BOLD}${n}${RESET} direct dependencies to @latest..."
	# shellcheck disable=SC2046
	go get $(printf '%s@latest ' $mods)
	go mod tidy
	go mod verify
	ok "All done."
}

# ── patch ─────────────────────────────────────────────────────────────────────

cmd_patch() {
	info "Checking for patch updates..."
	updates="$(list_direct_updates)"
	if [ -z "$updates" ]; then
		ok "No updates available."
		return
	fi
	printf '\n'
	printf '%s\n' "$updates" | while read -r mod cur latest; do
		cur_clean="$(strip_semver "$cur")"
		lat_clean="$(strip_semver "$latest")"

		cur_major="$(semver_field "$cur_clean" 1)"
		cur_minor="$(semver_field "$cur_clean" 2)"
		cur_patch="$(semver_field "$cur_clean" 3)"
		lat_major="$(semver_field "$lat_clean" 1)"
		lat_minor="$(semver_field "$lat_clean" 2)"
		lat_patch="$(semver_field "$lat_clean" 3)"

		for v in "$cur_major" "$cur_minor" "$cur_patch" "$lat_major" "$lat_minor" "$lat_patch"; do
			case "$v" in
			'' | *[!0-9]*)
				warn "$(printf '%-44s  %s → %s  (non-standard version)' "$mod" "$cur" "$latest")"
				continue 2
				;;
			esac
		done

		if [ "$cur_major" = "$lat_major" ] \
			&& [ "$cur_minor" = "$lat_minor" ] \
			&& [ "$lat_patch" -gt "$cur_patch" ]; then
			bump "$(printf '%-44s  %s → %s' "$mod" "$cur" "$latest")"
			go get "$mod@$latest"
		else
			warn "$(printf '%-44s  %s → %s  (major/minor change)' "$mod" "$cur" "$latest")"
		fi
	done
	printf '\n'
	go mod tidy
	go mod verify
	ok "Patch update complete."
}

# ── entry ─────────────────────────────────────────────────────────────────────

mode="${1:-}"
case "$mode" in
list)  cmd_list  ;;
up)    cmd_up    ;;
patch) cmd_patch ;;
-h | --help | help) usage ;;
'') usage; exit 1 ;;
*) die "Unknown mode: $mode" ;;
esac
