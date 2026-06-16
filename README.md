# Whatsar

Self-hosted WhatsApp API gateway — ringan, tanpa Chromium, cocok untuk notifikasi & integrasi project.

**Stack:** Go · whatsmeow · SQLite · HTMX admin UI

## Quick Start (Windows)

```powershell
git clone https://github.com/arifianilhamnrr/whatsar.git
cd whatsar
copy .env.example .env
# edit .env — ganti API key & password admin

go build -o whatsar.exe ./cmd/server
.\scripts\start.ps1
```

- Admin UI: http://127.0.0.1:8080/admin
- API docs: http://127.0.0.1:8080/admin/docs
- Health: http://127.0.0.1:8080/health

## Quick Start (Linux)

```bash
git clone https://github.com/arifianilhamnrr/whatsar.git
cd whatsar
cp .env.example .env
go build -o whatsar ./cmd/server
./whatsar
```

Atau pakai installer: `sudo bash install.sh`

## Kirim Notif (contoh)

```bash
curl -X POST http://127.0.0.1:8080/api/v1/messages/send \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"UUID","to":"6281234567890","text":"Halo dari whatsar!"}'
```

## Dokumentasi

- [PLAN.md](PLAN.md) — roadmap implementasi
- [ARSITEKTUR.md](ARSITEKTUR.md) — arsitektur teknis
- Admin → **Dokumentasi** — panduan integrasi lengkap

## License

MIT