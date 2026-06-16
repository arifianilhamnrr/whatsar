#!/usr/bin/env bash
# Whatsar — Universal Installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/<user>/whatsar/main/install.sh | bash
#   curl -fsSL ... | bash -s -- --port 8080 --no-tunnel
#
# Supports: Linux arm64, armv7, x86_64 (amd64)
# Tested on: Armbian STB, Debian, Ubuntu, Alpine (limited)

set -euo pipefail

# ─── Config ───────────────────────────────────────────────────────────────────

REPO="whatsar/whatsar"          # TODO: ganti ke GitHub user/repo kamu
INSTALL_DIR="/opt/whatsar"
SERVICE_USER="whatsar"
BINARY_NAME="whatsar"
VERSION="${WHATSAR_VERSION:-latest}"
PORT="${WHATSAR_PORT:-8080}"
SETUP_TUNNEL=false
SETUP_SWAP=false
SKIP_SYSTEMD=false

# ─── Colors ───────────────────────────────────────────────────────────────────

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC}  $*"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ─── Parse args ───────────────────────────────────────────────────────────────

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --port)       PORT="$2"; shift 2 ;;
            --version)    VERSION="$2"; shift 2 ;;
            --with-tunnel) SETUP_TUNNEL=true; shift ;;
            --with-swap)   SETUP_SWAP=true; shift ;;
            --no-systemd)  SKIP_SYSTEMD=true; shift ;;
            --help|-h)
                echo "Usage: install.sh [OPTIONS]"
                echo "  --port PORT        HTTP port (default: 8080)"
                echo "  --version VER      Release version (default: latest)"
                echo "  --with-tunnel      Setup Cloudflare Tunnel"
                echo "  --with-swap        Setup swap (recommended for ≤2GB RAM)"
                echo "  --no-systemd       Skip systemd, manual run only"
                exit 0
                ;;
            *) warn "Unknown option: $1" ; shift ;;
        esac
    done
}

# ─── Detect system ────────────────────────────────────────────────────────────

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
        warn "RAM ≤ 2GB — disarankan pakai --with-swap dan maks 1-2 session WA"
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

# ─── Dependencies ─────────────────────────────────────────────────────────────

install_deps() {
    info "Install dependensi sistem..."

    case "$OS_FAMILY" in
        debian)
            apt-get update -qq
            apt-get install -y -qq curl ca-certificates unzip
            ;;
        alpine)
            apk add --no-cache curl ca-certificates unzip
            ;;
        *)
            warn "OS family ${OS_FAMILY} — skip auto-install deps, pastikan curl tersedia"
            ;;
    esac

    command -v curl >/dev/null || error "curl tidak ditemukan"
    ok "Dependensi OK"
}

# ─── Download binary ──────────────────────────────────────────────────────────

get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | cut -d'"' -f4
}

download_binary() {
    local ver url tmp
    ver="$VERSION"

    if [[ "$ver" == "latest" ]]; then
        ver="$(get_latest_version)" || error "Gagal ambil versi terbaru dari GitHub Releases"
    fi

    url="https://github.com/${REPO}/releases/download/${ver}/${BINARY_NAME}_${ver#v}_linux_${ARCH}.tar.gz"

    info "Download ${BINARY_NAME} ${ver} (${ARCH})..."
    info "URL: ${url}"

    tmp="$(mktemp -d)"
    curl -fsSL "$url" -o "${tmp}/whatsar.tar.gz" \
        || error "Download gagal. Pastikan release ${ver} untuk ${ARCH} sudah ada di GitHub"

    tar -xzf "${tmp}/whatsar.tar.gz" -C "${tmp}"
    install -m 755 "${tmp}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    rm -rf "${tmp}"

    ok "Binary terinstall → ${INSTALL_DIR}/${BINARY_NAME}"
}

# ─── Setup directories & user ─────────────────────────────────────────────────

setup_user() {
    if ! id "$SERVICE_USER" &>/dev/null; then
        info "Buat user sistem: ${SERVICE_USER}"
        useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
    fi
    ok "User ${SERVICE_USER} OK"
}

setup_dirs() {
    info "Setup direktori ${INSTALL_DIR}..."

    mkdir -p "${INSTALL_DIR}"/{data,logs,config}
    chown -R "${SERVICE_USER}:${SERVICE_USER}" "${INSTALL_DIR}"
    chmod 750 "${INSTALL_DIR}"
    chmod 700 "${INSTALL_DIR}/data"

    ok "Direktori OK"
}

generate_config() {
    local api_key admin_pass config_file
    api_key="$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p)"
    admin_pass="$(openssl rand -base64 16 2>/dev/null || head -c 16 /dev/urandom | base64)"
    config_file="${INSTALL_DIR}/config/config.yaml"

    if [[ -f "$config_file" ]]; then
        warn "Config sudah ada, skip generate"
        return
    fi

    cat > "$config_file" <<EOF
# Whatsar config — generated by install.sh
server:
  port: ${PORT}
  host: "127.0.0.1"

database:
  path: "${INSTALL_DIR}/data/whatsar.db"

auth:
  api_key: "${api_key}"
  admin_password: "${admin_pass}"

sessions:
  max_active: 2

log:
  level: "info"
  file: "${INSTALL_DIR}/logs/whatsar.log"
EOF

    chown "${SERVICE_USER}:${SERVICE_USER}" "$config_file"
    chmod 600 "$config_file"

    # Simpan credential untuk ditampilkan sekali
    CREDENTIAL_FILE="${INSTALL_DIR}/config/.credentials"
    cat > "$CREDENTIAL_FILE" <<EOF
API_KEY=${api_key}
ADMIN_PASSWORD=${admin_pass}
EOF
    chmod 600 "$CREDENTIAL_FILE"
    chown "${SERVICE_USER}:${SERVICE_USER}" "$CREDENTIAL_FILE"

    ok "Config generated → ${config_file}"
}

# ─── Swap (device kentang) ────────────────────────────────────────────────────

setup_swap() {
    if [[ "$SETUP_SWAP" != "true" ]]; then return; fi
    if swapon --show | grep -q '/swapfile'; then
        ok "Swap sudah aktif"
        return
    fi

    info "Setup swap 512MB (device kentang)..."

    fallocate -l 512M /swapfile 2>/dev/null || dd if=/dev/zero of=/swapfile bs=1M count=512
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile

    if ! grep -q '/swapfile' /etc/fstab; then
        echo '/swapfile none swap sw 0 0' >> /etc/fstab
    fi

    ok "Swap 512MB aktif"
}

# ─── Systemd ──────────────────────────────────────────────────────────────────

setup_systemd() {
    if [[ "$SKIP_SYSTEMD" == "true" ]]; then
        warn "Skip systemd — jalankan manual: ${INSTALL_DIR}/${BINARY_NAME}"
        return
    fi

    info "Setup systemd service..."

    cat > /etc/systemd/system/whatsar.service <<EOF
[Unit]
Description=Whatsar WhatsApp API
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} --config ${INSTALL_DIR}/config/config.yaml
Restart=on-failure
RestartSec=10
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${INSTALL_DIR}/data ${INSTALL_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable whatsar
    systemctl restart whatsar

    ok "Service whatsar aktif"
}

# ─── Cloudflare Tunnel (opsional) ─────────────────────────────────────────────

setup_cloudflared() {
    if [[ "$SETUP_TUNNEL" != "true" ]]; then return; fi

    info "Setup Cloudflare Tunnel..."

    if ! command -v cloudflared &>/dev/null; then
        local cf_url cf_arch
        case "$ARCH" in
            amd64)  cf_arch="amd64" ;;
            arm64)  cf_arch="arm64" ;;
            armv7)  cf_arch="arm"   ;;
        esac
        cf_url="https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${cf_arch}"
        curl -fsSL "$cf_url" -o /usr/local/bin/cloudflared
        chmod +x /usr/local/bin/cloudflared
    fi

    warn "Cloudflare Tunnel butuh token/credential manual."
    warn "Jalankan: cloudflared tunnel login"
    warn "Lalu edit /etc/systemd/system/cloudflared.service dengan tunnel ID kamu"
    warn "Lihat: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/"

    cat > /etc/systemd/system/cloudflared.service <<EOF
[Unit]
Description=Cloudflare Tunnel for Whatsar
After=network-online.target whatsar.service
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudflared tunnel --no-autoupdate run --token YOUR_TUNNEL_TOKEN
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    warn "cloudflared.service dibuat tapi BELUM di-enable (butuh token dulu)"
}

# ─── Print summary ────────────────────────────────────────────────────────────

print_summary() {
    local api_key admin_pass
    # shellcheck source=/dev/null
    source "${INSTALL_DIR}/config/.credentials" 2>/dev/null || true

    echo ""
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║              Whatsar berhasil terinstall!                ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo ""
    echo "  URL lokal  : http://127.0.0.1:${PORT}"
    echo "  Admin UI   : http://127.0.0.1:${PORT}/admin"
    echo "  API base   : http://127.0.0.1:${PORT}/api/v1"
    echo ""
    echo "  API Key    : ${API_KEY:-<lihat ${INSTALL_DIR}/config/.credentials>}"
    echo "  Admin pass : ${ADMIN_PASSWORD:-<lihat ${INSTALL_DIR}/config/.credentials>}"
    echo ""
    echo "  Status     : systemctl status whatsar"
    echo "  Log        : journalctl -u whatsar -f"
    echo ""
    warn "Simpan API Key & admin password di atas — tidak ditampilkan lagi!"
    echo ""
}

# ─── Main ─────────────────────────────────────────────────────────────────────

main() {
    [[ "$EUID" -ne 0 ]] && error "Jalankan sebagai root: sudo bash install.sh"

    parse_args "$@"
    detect_all
    install_deps
    setup_user
    setup_dirs
    download_binary
    generate_config
    setup_swap
    setup_systemd
    setup_cloudflared
    print_summary
}

main "$@"