#!/usr/bin/env bash
# Whatsar — Universal Installer (Linux / STB / VPS)
#
# One-liner:
#   curl -fsSL https://raw.githubusercontent.com/arifianilhamnrr/whatsar/main/install.sh | sudo bash
#
# Options:
#   curl -fsSL ... | sudo bash -s -- --port 8080 --with-swap --with-tunnel
#
# Supports: linux amd64, arm64, armv7

set -euo pipefail

VERSION="${WHATSAR_VERSION:-latest}"
PORT="${WHATSAR_PORT:-8080}"
SETUP_TUNNEL=false
SETUP_SWAP=false
SKIP_SYSTEMD=false

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --port)        PORT="$2"; shift 2 ;;
            --version)     VERSION="$2"; shift 2 ;;
            --with-tunnel) SETUP_TUNNEL=true; shift ;;
            --with-swap)   SETUP_SWAP=true; shift ;;
            --no-systemd)  SKIP_SYSTEMD=true; shift ;;
            --help|-h)
                cat <<'HELP'
Usage: install.sh [OPTIONS]

  --port PORT        HTTP port (default: 8080)
  --version VER      GitHub release tag (default: latest)
  --with-tunnel      Install cloudflared + systemd unit (butuh token)
  --with-swap        Setup swap 512MB (auto jika RAM ≤ 2GB)
  --no-systemd       Skip systemd, jalankan manual
HELP
                exit 0
                ;;
            *) echo "[WARN] Unknown option: $1" >&2; shift ;;
        esac
    done
}

load_libs() {
    local raw="https://raw.githubusercontent.com/${REPO}/${BRANCH}"
    local tmp
    tmp="$(mktemp -d)"

    for lib in common detect systemd cloudflared; do
        curl -fsSL "${raw}/scripts/lib/${lib}.sh" -o "${tmp}/${lib}.sh" \
            || { echo "[ERROR] Gagal download scripts/lib/${lib}.sh" >&2; exit 1; }
        # shellcheck source=/dev/null
        source "${tmp}/${lib}.sh"
    done
}

main() {
    [[ "${EUID:-$(id -u)}" -ne 0 ]] && { echo "[ERROR] Jalankan sebagai root: sudo bash" >&2; exit 1; }

    # Defaults before libs load (overridden by common.sh)
    REPO="${WHATSAR_REPO:-arifianilhamnrr/whatsar}"
    BRANCH="${WHATSAR_BRANCH:-main}"

    parse_args "$@"
    load_libs

    detect_all
    install_deps
    setup_user
    setup_dirs
    download_binary "$VERSION"
    generate_env
    setup_swap
    setup_logrotate
    setup_systemd
    setup_cloudflared
    print_summary
}

main "$@"