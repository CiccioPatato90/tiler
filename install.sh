#!/usr/bin/env bash

set -euo pipefail

for command in go sway swaymsg; do
	if ! command -v "$command" >/dev/null 2>&1; then
		printf 'error: required command not found: %s\n' "$command" >&2
		exit 1
	fi
done

repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
install_dir="${HOME}/.local/bin"
binary="${install_dir}/tiler"
config_dir="${XDG_CONFIG_HOME:-${HOME}/.config}/sway"
config="${config_dir}/config"

if [[ ! -f "$config" ]]; then
	printf 'error: Sway config not found: %s\n' "$config" >&2
	exit 1
fi

mkdir -p "$install_dir"
(
	cd "$repo_dir"
	go build -buildvcs=false -o "$binary" .
)
printf 'installed %s\n' "$binary"

normal_binding='bindsym $mod+g exec ~/.local/bin/tiler'
equal_binding='bindsym $mod+Shift+g exec ~/.local/bin/tiler --equal'

check_binding() {
	local key_pattern=$1
	local binding=$2

	if grep -Fqx "$binding" "$config"; then
		return
	fi

	if grep -Eq "^[[:space:]]*bindsym[[:space:]]+${key_pattern}[[:space:]]+" "$config"; then
		printf 'error: key is already bound; refusing to overwrite it:\n' >&2
		grep -E "^[[:space:]]*bindsym[[:space:]]+${key_pattern}[[:space:]]+" "$config" >&2
		exit 1
	fi
}

add_binding() {
	local binding=$1

	if grep -Fqx "$binding" "$config"; then
		printf 'binding already configured: %s\n' "$binding"
		return
	fi
	printf '\n%s\n' "$binding" >>"$config"
	printf 'added binding: %s\n' "$binding"
}

check_binding '\$mod\+g' "$normal_binding"
check_binding '\$mod\+Shift\+g' "$equal_binding"
add_binding "$normal_binding"
add_binding "$equal_binding"

if [[ -n "${SWAYSOCK:-}" ]]; then
	swaymsg reload >/dev/null
	printf 'reloaded Sway configuration\n'
else
	printf 'Sway is not running; bindings will be available next session\n'
fi
