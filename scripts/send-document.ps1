# Kirim dokumen lokal via API Whatsar
# Usage: .\scripts\send-document.ps1 -To 6281234567890 -File "C:\path\file.pdf" [-SessionId UUID]

param(
    [Parameter(Mandatory)][string]$To,
    [Parameter(Mandatory)][string]$File,
    [string]$SessionId = "",
    [string]$ApiKey = $env:WHATSAR_API_KEY,
    [string]$BaseUrl = "http://127.0.0.1:8080"
)

$root = Split-Path $PSScriptRoot -Parent
if (-not $ApiKey) {
    $envFile = Join-Path $root ".env"
    if (Test-Path $envFile) {
        Get-Content $envFile | ForEach-Object {
            if ($_ -match '^WHATSAR_API_KEY=(.+)$') { $ApiKey = $Matches[1].Trim() }
        }
    }
}
if (-not $ApiKey) { throw "API key tidak ditemukan - set WHATSAR_API_KEY atau isi .env" }
if (-not (Test-Path $File)) { throw "File tidak ditemukan: $File" }

if (-not $SessionId) {
    $headers = @{ "X-API-Key" = $ApiKey }
    $sessions = (Invoke-RestMethod -Uri "$BaseUrl/api/v1/sessions" -Headers $headers).data
    $connected = $sessions | Where-Object { $_.status -eq "connected" } | Select-Object -First 1
    if (-not $connected) { $connected = $sessions | Select-Object -First 1 }
    if (-not $connected) { throw "Tidak ada session - buat/pair dulu di /admin" }
    $SessionId = $connected.id
}

$bytes = [IO.File]::ReadAllBytes($File)
$name = [IO.Path]::GetFileName($File)
$mime = switch ([IO.Path]::GetExtension($File).ToLower()) {
    ".pdf"  { "application/pdf" }
    ".doc"  { "application/msword" }
    ".docx" { "application/vnd.openxmlformats-officedocument.wordprocessingml.document" }
    ".xls"  { "application/vnd.ms-excel" }
    ".xlsx" { "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }
    default { "application/octet-stream" }
}

$payload = [ordered]@{
    session_id      = $SessionId
    to              = $To
    type            = "document"
    document_base64 = [Convert]::ToBase64String($bytes)
    filename        = $name
    mimetype        = $mime
}
$jsonPath = Join-Path $root "data\send-payload.json"
$json = $payload | ConvertTo-Json -Compress
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText($jsonPath, $json, $utf8NoBom)

Write-Host "Mengirim $name ($($bytes.Length) bytes) ke $To via session $SessionId ..."
$resp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/messages/send" -Method POST `
    -Headers @{ "X-API-Key" = $ApiKey } -InFile $jsonPath -ContentType "application/json" -TimeoutSec 300
$resp | ConvertTo-Json -Depth 5