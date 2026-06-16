#!/usr/bin/env bash
# Kirim pesan teks via Whatsar API
# Usage: WHATSAR_KEY=xxx WHATSAR_SESSION=uuid ./send-text.sh 6281234567890 "Halo!"

set -euo pipefail

BASE_URL="${WHATSAR_URL:-http://127.0.0.1:8080}"
API_KEY="${WHATSAR_KEY:?set WHATSAR_KEY}"
SESSION="${WHATSAR_SESSION:?set WHATSAR_SESSION}"
TO="${1:?nomor tujuan}"
TEXT="${2:?pesan}"

curl -sS -X POST "${BASE_URL}/api/v1/messages/send" \
  -H "X-API-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n \
    --arg sid "$SESSION" \
    --arg to "$TO" \
    --arg text "$TEXT" \
    '{session_id:$sid,to:$to,text:$text}')"
echo