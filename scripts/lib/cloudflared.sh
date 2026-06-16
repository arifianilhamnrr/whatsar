#!/usr/bin/env bash
# Cloudflare Tunnel (opsional)

setup_cloudflared() {
    if [[ "${SETUP_TUNNEL:-false}" != "true" ]]; then return; fi

    info "Setup Cloudflare Tunnel..."

    if ! command -v cloudflared &>/dev/null; then
        local cf_url cf_arch
        case "${ARCH}" in
            amd64) cf_arch="amd64" ;;
            arm64) cf_arch="arm64" ;;
            armv7) cf_arch="arm"   ;;
        esac
        cf_url="https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${cf_arch}"
        curl -fsSL "$cf_url" -o /usr/local/bin/cloudflared
        chmod +x /usr/local/bin/cloudflared
        ok "cloudflared terinstall"
    fi

    cat > /etc/systemd/system/cloudflared.service <<'EOF'
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
    warn "cloudflared.service dibuat — ganti YOUR_TUNNEL_TOKEN lalu: systemctl enable --now cloudflared"
    warn "Panduan: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/"
}