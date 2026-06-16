# Whatsar API Reference

Base URL: `http://127.0.0.1:8080` (ganti sesuai deploy kamu)

## Autentikasi

Semua endpoint `/api/v1/*` membutuhkan API key:

```http
X-API-Key: your-api-key
```

Rate limit: **60 request/menit per API key** (konfigurasi: `WHATSAR_RATE_LIMIT`).

`/health` tidak membutuhkan API key.

---

## Format Response

```json
{
  "success": true,
  "data": {},
  "meta": { "request_id": "uuid" }
}
```

```json
{
  "success": false,
  "error": { "code": "VALIDATION_ERROR", "message": "..." },
  "meta": { "request_id": "uuid" }
}
```

### Kode error umum

| Code | HTTP | Arti |
|------|------|------|
| `UNAUTHORIZED` | 401 | API key salah / kosong |
| `RATE_LIMITED` | 429 | Terlalu banyak request |
| `VALIDATION_ERROR` | 400 | Input tidak valid |
| `INVALID_JSON` | 400 | Body JSON rusak |
| `SESSION_NOT_FOUND` | 404 | Session tidak ada |
| `SEND_FAILED` | 400 | Gagal kirim pesan |

---

## Endpoints

### `GET /health`

Health check + statistik.

**Response `data`:**
```json
{
  "status": "ok",
  "uptime_seconds": 3600,
  "sessions_total": 1,
  "sessions_connected": 1,
  "queue_pending": 0
}
```

---

### Session

#### `POST /api/v1/sessions`

Buat session baru untuk pairing QR.

**Body:**
```json
{ "name": "notif-toko" }
```

#### `GET /api/v1/sessions`

List semua session.

#### `GET /api/v1/sessions/{id}`

Detail session.

#### `GET /api/v1/sessions/{id}/status`

Cek status koneksi sebelum kirim pesan.

#### `GET /api/v1/sessions/{id}/qr`

Ambil QR code (field `image_base64` = PNG).

#### `DELETE /api/v1/sessions/{id}`

Hapus session + logout device.

---

### Pesan

#### `POST /api/v1/messages/send`

Kirim pesan WhatsApp.

**Body (teks):**
```json
{
  "session_id": "uuid",
  "to": "6281234567890",
  "text": "Halo dari Whatsar!"
}
```

**Body (gambar):**
```json
{
  "session_id": "uuid",
  "to": "6281234567890",
  "type": "image",
  "image_url": "https://example.com/foto.jpg",
  "caption": "Promo!"
}
```

**Body (dokumen):**
```json
{
  "session_id": "uuid",
  "to": "6281234567890",
  "type": "document",
  "document_url": "https://example.com/file.pdf",
  "filename": "laporan.pdf",
  "mimetype": "application/pdf"
}
```

**Field opsional:**

| Field | Keterangan |
|-------|------------|
| `type` | `text` (default), `image`, `document` |
| `image_base64` | Alternatif `image_url` |
| `document_base64` | Alternatif `document_url` |
| `reply_to` | ID pesan untuk reply |
| `quoted_text` | Teks yang di-quote |
| `retry` | `true` = antrian retry jika gagal (HTTP 202) |

**`to`:** nomor `628...`, JID grup `...@g.us`, atau JID lengkap.

**Response sukses:**
```json
{ "message_id": "3EB0...", "status": "sent" }
```

**Response antrian (retry):**
```json
{ "status": "queued", "queued": true, "queue_id": "uuid" }
```

#### `GET /api/v1/messages?session_id={id}&limit=50&offset=0`

Log pesan masuk/keluar. `limit` maks 200.

---

### Webhook

#### `POST /api/v1/webhooks`

Daftarkan URL untuk terima pesan masuk.

```json
{
  "url": "https://project-kamu.com/api/wa-webhook",
  "session_id": "uuid",
  "events": ["message.in"],
  "secret": "webhook-secret"
}
```

`session_id` kosong = global (semua session).

**Payload ke URL kamu:**
```json
{
  "event": "message.in",
  "session_id": "uuid",
  "timestamp": "2026-06-16T10:00:00Z",
  "data": {
    "from": "6289876543210",
    "chat": "6289876543210@s.whatsapp.net",
    "body": "Oke kak",
    "message_id": "3EB0...",
    "type": "text",
    "is_group": false
  }
}
```

Verifikasi HMAC: header `X-Whatsar-Signature: sha256=<hmac>`.

---

## Validasi Input

| Field | Aturan |
|-------|--------|
| `session_id` | UUID v4 |
| `to` | 8–15 digit atau JID valid |
| `text` | Maks 4096 karakter |
| `caption` | Maks 1024 karakter |
| `filename` | Maks 255 karakter, alphanumeric + `._-` |
| URL | `http`/`https` saja |
| Body request | Maks ~72 MB (untuk dokumen base64) |

---

## Contoh Client

Lihat folder [`examples/`](examples/):

- `examples/curl/` — shell scripts
- `examples/php/` — PHP / Laragon
- `examples/javascript/` — Node.js / fetch

Panduan interaktif juga tersedia di Admin UI → **Dokumentasi** (`/admin/docs`).