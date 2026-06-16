# Whatsar — Rencana Implementasi

> API WhatsApp ringan berbasis Go + whatsmeow, dirancang untuk device kentang (STB Armbian HG680P 2GB) dan reusable di banyak project.

---

## Tujuan

| Aspek | Target |
|-------|--------|
| Platform | STB Armbian (ARM, 2GB RAM) — production ringan |
| Akses | Dari mana saja via HTTPS (Cloudflare Tunnel) |
| Konsumsi RAM | < 800 MB untuk 1 session aktif |
| Reusability | Satu API dipakai banyak project (Laragon, mobile, dll.) |
| Maintenance | Satu binary Go, tanpa Node/Chromium di server |

---

## Stack Final

| Layer | Teknologi |
|-------|-----------|
| WhatsApp engine | [whatsmeow](https://github.com/tulir/whatsmeow) |
| API & UI server | Go 1.22+ |
| HTTP router | [chi](https://github.com/go-chi/chi) |
| Database | SQLite ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)) |
| UI admin | HTMX + Pico.css (embed di binary) |
| Tunnel | Cloudflare Tunnel (`cloudflared`) |
| Auth API | API Key per client/project |

---

## Fase Implementasi

### Fase 0 — Persiapan (Hari 1)

- [ ] Inisialisasi repo Go (`go mod init github.com/whatsar/whatsar`)
- [ ] Setup struktur folder (lihat `ARSITEKTUR.md`)
- [ ] Buat `Makefile` / script build untuk ARM (`GOOS=linux GOARCH=arm64`)
- [ ] Setup `.env` / config (port, DB path, API keys)
- [ ] Dokumentasi deploy ke STB Armbian

**Deliverable:** Project skeleton bisa `go build` dan jalan di lokal.

---

### Fase 1 — Core WhatsApp Engine (Hari 2–4)

- [x] Wrapper `internal/wa/` di atas whatsmeow
- [x] Session store ke SQLite (credential, device identity)
- [x] Flow pairing QR code
- [x] Auto-reconnect saat koneksi putus
- [x] Event handler: pesan masuk, status delivery, disconnect
- [x] Kirim pesan teks

**Deliverable:** Bisa scan QR, terima & kirim pesan teks via CLI/internal test.

**Endpoint internal (belum REST publik):**
```
SessionManager.Create()   → generate QR
SessionManager.Connect()  → maintain WebSocket
SessionManager.SendText() → kirim pesan
```

---

### Fase 2 — REST API (Hari 5–7)

- [x] Middleware: API key auth, request logging, rate limit
- [x] Response format standar (JSON envelope)
- [x] CRUD session
- [x] Endpoint kirim pesan (teks dulu)
- [x] Webhook outbound — forward pesan masuk ke URL client
- [x] Health check (`GET /health`)

**Endpoint v1:**

| Method | Path | Deskripsi |
|--------|------|-----------|
| `POST` | `/api/v1/sessions` | Buat session baru |
| `GET` | `/api/v1/sessions` | List semua session |
| `GET` | `/api/v1/sessions/{id}` | Detail session |
| `DELETE` | `/api/v1/sessions/{id}` | Hapus session |
| `GET` | `/api/v1/sessions/{id}/qr` | Ambil QR (base64 / PNG) |
| `GET` | `/api/v1/sessions/{id}/status` | Status koneksi |
| `POST` | `/api/v1/messages/send` | Kirim pesan |
| `GET` | `/api/v1/messages` | Riwayat pesan (paginated) |
| `POST` | `/api/v1/webhooks` | Daftarkan webhook URL |
| `GET` | `/health` | Health check |

**Deliverable:** Project eksternal bisa kirim pesan cuma dengan `curl` + API key.

---

### Fase 3 — Admin UI (Hari 8–9)

- [x] Template HTMX di `web/templates/`
- [x] Pico.css + HTMX embed via `go:embed`
- [x] Halaman dashboard: list session & status
- [x] Halaman QR: tampil & auto-refresh sampai connected
- [x] Halaman kirim pesan manual (testing)
- [x] Halaman log pesan masuk/keluar

**Route UI:**

| Path | Halaman |
|------|---------|
| `/admin` | Dashboard |
| `/admin/sessions/new` | Buat session + QR |
| `/admin/sessions/{id}` | Detail session |
| `/admin/messages` | Log pesan |

**Deliverable:** Kelola WhatsApp full dari browser, tanpa CLI.

---

### Fase 4 — Media & Fitur Lanjut (Hari 10–12)

- [x] Kirim gambar (URL / base64) + caption
- [x] Reply & quote message (`reply_to`, `quoted_text`)
- [x] Grup: kirim ke JID `...@g.us`
- [x] Retry queue untuk pesan gagal (SQLite-backed)
- [x] Webhook retry dengan exponential backoff
- [ ] Kirim dokumen/PDF (backlog)

**Prioritas rendah (backlog):**
- Broadcast / bulk send
- Template pesan
- Multi-API-key dengan quota per client

---

### Fase 5 — Installer Bash + Deploy (Hari 13–14)

- [ ] `install.sh` di root repo — one-liner install dari GitHub
- [ ] Deteksi otomatis: OS, arch (arm64/armv7/amd64), RAM
- [ ] Download binary dari GitHub Releases (pre-built per arch)
- [ ] Setup user, direktori, permission (`/opt/whatsar`)
- [ ] Generate config & API key otomatis
- [ ] Systemd service `whatsar.service`
- [ ] Systemd service `cloudflared.service` (opsional, interaktif)
- [ ] Swap 512MB–1GB jika RAM ≤ 2GB
- [ ] Log rotation (`logrotate`)
- [ ] GitHub Actions: build & publish release multi-arch

**Cara install (target):**

Linux / STB / VPS:
```bash
curl -fsSL https://raw.githubusercontent.com/<user>/whatsar/main/install.sh | sudo bash
```

Windows (PowerShell as Admin):
```powershell
irm https://raw.githubusercontent.com/<user>/whatsar/main/install.ps1 | iex
```

**Deliverable:** API live di `https://whatsar.<domain>.com`, install satu command di STB/VPS/Windows.

---

### Fase 6 — Hardening & Dokumentasi (Hari 15+)

- [ ] Rate limiting per API key
- [ ] Validasi input ketat
- [ ] Graceful shutdown (simpan state WA)
- [ ] Integration test dasar
- [ ] `API.md` — dokumentasi endpoint untuk developer
- [ ] Contoh client: PHP (Laragon), curl, JavaScript fetch

---

## Struktur Folder

```
whatsar/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # Auth, logging, rate limit
│   │   └── router.go            # Route registration
│   ├── wa/
│   │   ├── client.go            # whatsmeow wrapper
│   │   ├── events.go            # Event dispatcher
│   │   └── session.go           # Session lifecycle
│   ├── store/
│   │   ├── sqlite.go            # DB connection
│   │   ├── session.go           # Session repo
│   │   ├── message.go           # Message repo
│   │   └── webhook.go           # Webhook repo
│   ├── webhook/
│   │   └── dispatcher.go        # Outbound webhook sender
│   └── config/
│       └── config.go            # Env & config loader
├── web/
│   ├── templates/               # HTMX HTML templates
│   └── static/
│       ├── pico.min.css
│       └── htmx.min.js
├── data/                        # SQLite DB (gitignored)
├── scripts/
│   ├── lib/                     # Helper functions untuk install.sh
│   │   ├── detect.sh            # Deteksi OS, arch, RAM
│   │   ├── systemd.sh           # Setup systemd service
│   │   └── cloudflared.sh       # Setup CF Tunnel (opsional)
│   └── build-release.sh         # Build multi-arch untuk GitHub Release
├── install.sh                   # Installer Linux (curl | bash)
├── install.ps1                  # Installer Windows (PowerShell)
├── go.mod
├── go.sum
├── Makefile
├── PLAN.md
└── ARSITEKTUR.md
```

---

## Batasan & Ekspektasi Hardware

| Skenario | HG680P 2GB | Rekomendasi |
|----------|------------|-------------|
| 1 session + API + UI | ✅ Nyaman | Target utama |
| 2–3 session | ⚠️ Muat, monitor RAM | Set swap wajib |
| 5+ session | ❌ | Pindah ke VPS |
| Kirim media besar | ⚠️ Lambat (CPU lemah) | Batasi ukuran file |
| 24/7 uptime | ⚠️ | STB bukan server DC — backup plan siap |

---

## Konvensi API Response

**Sukses:**
```json
{
  "success": true,
  "data": { ... },
  "meta": { "request_id": "uuid" }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session id tidak ditemukan"
  }
}
```

---

## Contoh Integrasi Project Lain

**PHP (Laragon):**
```php
$response = Http::withHeaders([
    'X-API-Key' => env('WHATSAR_API_KEY'),
])->post('https://whatsar.example.com/api/v1/messages/send', [
    'session_id' => 'default',
    'to'         => '628123456789',
    'text'       => 'Halo dari Laragon!',
]);
```

**curl:**
```bash
curl -X POST https://whatsar.example.com/api/v1/messages/send \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"default","to":"628123456789","text":"Test"}'
```

---

## Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|--------|--------|----------|
| WhatsApp update protokol | Engine rusak sementara | Pin versi whatsmeow, pantau upstream |
| STB mati/listrik padam | API down | UPS kecil; notif health check |
| Akun WA banned | Kehilangan nomor | Jangan spam; hormati rate limit |
| RAM penuh | OOM kill | Limit max session; swap; monitor |
| eMMC lambat | DB lock | WAL mode SQLite; batch write |

---

## Definisi "Selesai" (MVP)

MVP dianggap selesai jika:

1. ✅ Scan QR & maintain 1 session 24 jam di STB
2. ✅ REST API kirim & terima pesan teks
3. ✅ Webhook forward pesan masuk ke URL eksternal
4. ✅ Admin UI: dashboard + QR + log pesan
5. ✅ Akses HTTPS via Cloudflare Tunnel
6. ✅ Project Laragon bisa kirim pesan lewat API key

---

## Timeline Ringkas

```
Minggu 1   → Fase 0–2  (skeleton + engine + REST API)
Minggu 2   → Fase 3–5  (UI admin + deploy STB)
Minggu 3+  → Fase 4,6  (media + hardening + docs)
```

---

## Langkah Berikutnya

1. Baca `ARSITEKTUR.md` untuk detail komponen & alur data
2. Mulai **Fase 0** — scaffold project Go
3. Implement **Fase 1** — whatsmeow wrapper + QR pairing