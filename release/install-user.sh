#!/bin/sh
set -eu

REPO="duakc/lightddns"
VERSION="${VERSION:-latest}"
RAW_REF="${RAW_REF:-main}" # branch/tag the unit + schema are fetched from

BIN_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/lightddns"
UNIT_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"

# release asset matching this architecture
asset() {
    case "$(uname -m)" in
        x86_64 | amd64) echo "lightddns-linux-amd64" ;;
        aarch64 | arm64) echo "lightddns-linux-arm64" ;;
        armv7l) echo "lightddns-linux-arm-v7" ;;
        armv6l) echo "lightddns-linux-arm-v6" ;;
        i386 | i686) echo "lightddns-linux-386" ;;
        riscv64) echo "lightddns-linux-riscv64" ;;
        loongarch64) echo "lightddns-linux-loong64" ;;
        *) echo "error: unsupported arch $(uname -m)" >&2; exit 1 ;;
    esac
}

fetch() { # URL OUTFILE
    if command -v curl >/dev/null 2>&1; then
        curl -fSL "$1" -o "$2"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        echo "error: need curl or wget" >&2
        exit 1
    fi
}

write_config() {
    [ -e "$CONF_DIR/config.yaml" ] || {
        echo ">> config -> $CONF_DIR/config.yaml"
        cat > "$CONF_DIR/config.yaml" <<EOF
# yaml-language-server: \$schema=https://raw.githubusercontent.com/$REPO/$RAW_REF/release/schema.json
#
# Minimal config. Add providers/datasources/domains per the docs:
#   https://lightddns.duaky.com
# Secrets go in $CONF_DIR/env and are referenced here as {{ .Env.KEY }}.

log:
  level: info
EOF
    }
    [ -e "$CONF_DIR/env" ] || {
        cat > "$CONF_DIR/env" <<'EOF'
# Sourced by the user unit (EnvironmentFile=).
# ARG appends flags to "run", e.g. ARG="--once". Secrets referenced as
# {{ .Env.KEY }} in config.yaml go here too, e.g. <PROVIDER>_TOKEN=...
#ARG=
EOF
        chmod 0600 "$CONF_DIR/env"
    }
}

# Enable through the user manager (no root). Linger keeps it running when you're
# logged out - essential on a VPS; failures are non-fatal.
enable_service() {
    command -v systemctl >/dev/null 2>&1 || return 0
    systemctl --user daemon-reload || true
    systemctl --user enable lightddns.service || true
    loginctl enable-linger "$(id -un)" 2>/dev/null \
        || echo "   note: linger off; run 'loginctl enable-linger $(id -un)' so it survives logout"
}

main() {
    [ "$(uname -s)" = Linux ] || { echo "error: this installer is for Linux (systemd) only" >&2; exit 1; }

    [ "$VERSION" = latest ] \
        && bin_url="https://github.com/$REPO/releases/latest/download/$(asset)" \
        || bin_url="https://github.com/$REPO/releases/download/$VERSION/$(asset)"
    unit_url="https://raw.githubusercontent.com/$REPO/$RAW_REF/release/systemd/user/lightddns.service"

    echo ">> binary -> $BIN_DIR/lightddns"
    mkdir -p "$BIN_DIR"
    fetch "$bin_url" "$BIN_DIR/lightddns"
    chmod 0755 "$BIN_DIR/lightddns"

    echo ">> unit   -> $UNIT_DIR/lightddns.service"
    mkdir -p "$UNIT_DIR"
    fetch "$unit_url" "$UNIT_DIR/lightddns.service"

    mkdir -p "$CONF_DIR" && chmod 0700 "$CONF_DIR"
    write_config
    enable_service

    cat <<EOF

Done. Next:
  1. edit  $CONF_DIR/config.yaml   (provider/domain)
  2. edit  $CONF_DIR/env           (tokens / ARG)
  3. start: systemctl --user start lightddns
     logs:  journalctl --user -u lightddns -f

Ensure $BIN_DIR is on your PATH (the service uses the absolute path regardless).
EOF
}

main "$@"