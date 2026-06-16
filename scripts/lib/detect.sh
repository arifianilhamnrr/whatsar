#!/usr/bin/env bash
# OS / arch / RAM detection

detect_os() {
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        . /etc/os-release
        OS_ID="${ID:-unknown}"
        OS_VERSION="${VERSION_ID:-}"
    else
        error "Tidak bisa deteksi OS. Butuh Linux dengan /etc/os-release"
    fi

    case "$OS_ID" in
        debian|ubuntu|armbian|raspbian|linuxmint) OS_FAMILY="debian" ;;
        alpine)                                    OS_FAMILY="alpine" ;;
        centos|rhel|fedora|rocky|alma)            OS_FAMILY="rhel"   ;;
        *)                                         OS_FAMILY="unknown" ;;
    esac

    info "OS: ${OS_ID} ${OS_VERSION} (${OS_FAMILY})"
}

detect_arch() {
    local raw
    raw="$(uname -m)"

    case "$raw" in
        x86_64|amd64)   ARCH="amd64"   ;;
        aarch64|arm64)  ARCH="arm64"   ;;
        armv7l|armv6l)  ARCH="armv7"   ;;
        *) error "Arsitektur tidak didukung: ${raw}" ;;
    esac

    info "Arch: ${raw} → ${ARCH}"
}

detect_ram() {
    local total_kb total_mb
    total_kb="$(grep MemTotal /proc/meminfo | awk '{print $2}')"
    total_mb=$(( total_kb / 1024 ))

    info "RAM: ${total_mb} MB"

    if [[ "$total_mb" -le 2048 ]]; then
        warn "RAM ≤ 2GB — disarankan --with-swap dan maks 1-2 session WA"
        SETUP_SWAP=true
    fi
}

detect_all() {
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║       Whatsar Installer v0.1         ║"
    echo "╚══════════════════════════════════════╝"
    echo ""

    detect_os
    detect_arch
    detect_ram
}

install_deps() {
    info "Install dependensi sistem..."

    case "${OS_FAMILY:-unknown}" in
        debian)
            apt-get update -qq
            apt-get install -y -qq curl ca-certificates tar openssl
            ;;
        alpine)
            apk add --no-cache curl ca-certificates tar openssl
            ;;
        *)
            warn "OS family ${OS_FAMILY} — pastikan curl & tar tersedia"
            ;;
    esac

    command -v curl >/dev/null || error "curl tidak ditemukan"
    command -v tar   >/dev/null || error "tar tidak ditemukan"
    ok "Dependensi OK"
}