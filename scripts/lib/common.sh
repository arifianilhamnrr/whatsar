#!/usr/bin/env bash
# Shared helpers for install.sh

REPO="${WHATSAR_REPO:-arifianilhamnrr/whatsar}"
BRANCH="${WHATSAR_BRANCH:-main}"
INSTALL_DIR="${WHATSAR_INSTALL_DIR:-/opt/whatsar}"
SERVICE_USER="${WHATSAR_USER:-whatsar}"
BINARY_NAME="whatsar"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC}  $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

rand_hex() {
    local n="${1:-32}"
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex "$n"
    else
        head -c "$n" /dev/urandom | xxd -p | tr -d '\n'
    fi
}

rand_pass() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 18
    else
        head -c 18 /dev/urandom | base64 | tr -d '\n'
    fi
}

release_version() {
    local ver="${1:-latest}"
    if [[ "$ver" != "latest" ]]; then
        echo "$ver"
        return
    fi
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | cut -d'"' -f4
}

release_asset_name() {
    local ver="$1" os="$2" arch="$3"
    local ver_num="${ver#v}"
    echo "${BINARY_NAME}_${ver_num}_${os}_${arch}"
}