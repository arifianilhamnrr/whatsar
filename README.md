# Whatsar

Self-hosted **WhatsApp API gateway** — ringan, tanpa Chromium, satu binary Go. Cocok untuk notifikasi order, OTP, reminder, dan integrasi backend (Laravel, Node, Python, dll.).

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](#license)

**Repo:** https://github.com/arifianilhamnrr/whatsar

---

## Kenapa Whatsar?

| | Whatsar | WA Web + Puppeteer |
|---|---|---|
| RAM | ~300–800 MB (1 session) | 1–2 GB+ |
| Runtime | Satu binary Go | Node + Chromium |
| Deploy | VPS, STB ARM, Windows | Butuh GUI/browser |
| Integrasi | REST API + webhook | Scraping / fragile |

Dibangun di atas [whatsmeow](https://github.com/tulir/whatsmeow) — protokol WhatsApp multi-device, tanpa browser headless.

---

## Fitur

- **REST API** — kirim pesan, kelola session, webhook, health check
- **Admin UI** — dashboard, pairing QR, log pesan, pengaturan API key (HTMX + Pico.css)
- **Pesan** — teks, gambar (URL/base64), dokumen/PDF, reply/quote, grup (`@g.us`)
- **Retry queue** — pesan gagal masuk antrian SQLite dengan backoff eksponensial
- **Webhook** — forward pesan masuk ke URL project kamu (+ HMAC signature)
- **Auto-reconnect** — session restore dari SQLite setelah restart

---

## Quick Start

### Windows (Laragon / lokal)

```powershell
git clone https://github.com/arifianilhamnrr/whatsar.git
cd whatsar
copy .env.example .env
# Edit .env — ganti WHATSAR_API_KEY dan WHATSAR_ADMIN_PASSWORD

go build -o whatsar.exe ./cmd/server
.\scripts\start.ps1
```

| URL | Fungsi |
|-----|--------|
| http://127.0.0.1:8080/admin | Admin UI (login pakai password admin) |
| http://127.0.0.1:8080/admin/docs | Dokumentasi API lengkap |
| http://127.0.0.1:8080/health | Health check (tanpa API key) |

**Pairing:** Admin → buat session → scan QR WhatsApp → ambil `session_id` dari dashboard.

### Linux / VPS

```bash
git clone https://github.com/arifianilhamnrr/whatsar.git
cd whatsar
cp .env.example .env
go build -o whatsar ./cmd/server
./whatsar
```

Installer one-liner (Fase 5 — work in progress): `sudo bash install.sh`

---

## Konfigurasi (`.env`)

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `WHATSAR_HOST` | `127.0.0.1` | Bind address |
| `WHATSAR_PORT` | `8080` | Port HTTP |
| `WHATSAR_DB_PATH` | `./data/whatsar.db` | Path SQLite |
| `WHATSAR_API_KEY` | — | API key awal (bisa diganti dari Admin → Pengaturan) |
| `WHATSAR_ADMIN_PASSWORD` | — | Password login admin UI |
| `WHATSAR_MAX_SESSIONS` | `5` | Batas session (`0` = tanpa batas) |
| `WHATSAR_PUBLIC_URL` | — | URL publik untuk contoh di docs (opsional) |
| `WHATSAR_LOG_LEVEL` | `info` | Level log |

---

## API Singkat

Semua endpoint `/api/v1/*` butuh header:

```
X-API-Key: your-api-key
```

### Kirim teks

```bash
curl -X POST http://127.0.0.1:8080/api/v1/messages/send \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "UUID-DARI-DASHBOARD",
    "to": "6281234567890",
    "text": "Halo dari Whatsar!"
  }'
```

### Kirim gambar

```json
{
  "session_id": "UUID",
  "to": "6281234567890",
  "type": "image",
  "image_url": "https://example.com/foto.jpg",
  "caption": "Promo hari ini"
}
```

### Kirim dokumen / PDF

```json
{
  "session_id": "UUID",
  "to": "6281234567890",
  "type": "document",
  "document_url": "https://example.com/laporan.pdf",
  "filename": "laporan.pdf",
  "mimetype": "application/pdf"
}
```

**Windows — file lokal:**

```powershell
.\scripts\send-document.ps1 -To 6281234567890 -File "C:\path\file.pdf"
```

### Endpoint lainnya

| Method | Path | Deskripsi |
|--------|------|-----------|
| `GET` | `/health` | Status server + `queue_pending` |
| `POST` | `/api/v1/sessions` | Buat session baru |
| `GET` | `/api/v1/sessions` | List session |
| `GET` | `/api/v1/sessions/{id}/status` | Cek koneksi sebelum kirim |
| `GET` | `/api/v1/sessions/{id}/qr` | Ambil QR (base64) |
| `POST` | `/api/v1/webhooks` | Daftarkan webhook pesan masuk |
| `GET` | `/api/v1/messages` | Log pesan per session |

Detail lengkap: **Admin → Dokumentasi** atau [web/templates/docs.html](web/templates/docs.html).

---

## Scripts

| Script | Fungsi |
|--------|--------|
| `scripts/start.ps1` | Start server (Windows) |
| `scripts/stop.ps1` | Stop server (Windows) |
| `scripts/send-document.ps1` | Kirim file lokal via API |
| `scripts/setup-go-path.ps1` | Tambah Go Laragon ke PATH |

---

## Integrasi Project

**PHP / Laravel** — simpan di `.env` project:

```env
WHATSAR_URL=http://127.0.0.1:8080
WHATSAR_KEY=your-api-key
WHATSAR_SESSION_ID=uuid-dari-dashboard
```

Contoh kode (PHP, Node, Python) ada di `/admin/docs`.

**Webhook** — daftarkan URL project kamu; Whatsar POST event `message.in` saat ada pesan masuk. Verifikasi via header `X-Whatsar-Signature`.

---

## Development

```bash
# Build server
go build -o whatsar ./cmd/server

# Build CLI (pairing / test lokal)
go build -o whatsar-cli ./cmd/cli

# Cross-compile ARM (STB Armbian)
make build-arm64
```

Struktur project: lihat [ARSITEKTUR.md](ARSITEKTUR.md)  
Roadmap: lihat [PLAN.md](PLAN.md)

---

## Hardware Target

Dirancang untuk jalan di **STB Armbian HG680P (ARM, 2GB RAM)**:

| Skenario | Status |
|----------|--------|
| 1 session + API + UI | Nyaman |
| 2–3 session | Muat, set swap |
| 5+ session | Pindah ke VPS |

---

## Roadmap

- [x] Fase 0–3 — Engine, REST API, Admin UI
- [x] Fase 4 — Media, dokumen, retry queue, webhook backoff
- [ ] Fase 5 — Installer, systemd, GitHub Releases multi-arch
- [ ] Fase 6 — Hardening, integration test, `API.md`

---

## License

MIT