#!/bin/sh
set -eu

PURGE="${PURGE:-}"

BIN_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/lightddns"
UNIT_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/lightddns"

stop_service() {
    command -v systemctl >/dev/null 2>&1 || return 0
    systemctl --user disable --now lightddns.service 2>/dev/null || true
}

main() {
    stop_service
    rm -f "$UNIT_DIR/lightddns.service"
    command -v systemctl >/dev/null 2>&1 && systemctl --user daemon-reload 2>/dev/null || true
    rm -f "$BIN_DIR/lightddns"

    [ -n "$PURGE" ] \
        && { rm -rf "$CONF_DIR" "$STATE_DIR"; echo ">> purged config + state"; } \
        || echo ">> kept config ($CONF_DIR) and state; set PURGE=1 to remove them"

    echo ">> uninstalled lightddns user service"
}

main "$@"