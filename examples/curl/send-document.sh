#!/usr/bin/env bash
# Kirim dokumen lokal (base64) via Whatsar API
# Usage: WHATSAR_KEY=xxx WHATSAR_SESSION=uuid ./send-document.sh 6281234567890 /path/file.pdf

set -euo pipefail

BASE_URL="${WHATSAR_URL:-http://127.0.0.1:8080}"
API_KEY="${WHATSAR_KEY:?set WHATSAR_KEY}"
SESSION="${WHATSAR_SESSION:?set WHATSAR_SESSION}"
TO="${1:?nomor tujuan}"
FILE="${2:?path file}"

B64="$(base64 -w0 "$FILE" 2>/dev/null || base64 < "$FILE" | tr -d '\n')"
NAME="$(basename "$FILE")"

curl -sS -X POST "${BASE_URL}/api/v1/messages/send" \
  -H "X-API-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  --max-time 300 \
  -d "$(jq -n \
    --arg sid "$SESSION" \
    --arg to "$TO" \
    --arg b64 "$B64" \
    --arg name "$NAME" \
    '{session_id:$sid,to:$to,type:"document",document_base64:$b64,filename:$name,mimetype:"application/pdf"}')"
echo