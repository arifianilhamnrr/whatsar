# Whatsar — Start server
# Usage: .\scripts\start.ps1

$root = Split-Path $PSScriptRoot -Parent
$exe  = Join-Path $root "whatsar.exe"
$lock = Join-Path $root "data\.whatsar.lock"

if (-not (Test-Path $exe)) {
    Write-Host "whatsar.exe tidak ditemukan. Build dulu: go build -o whatsar.exe ./cmd/server" -ForegroundColor Red
    exit 1
}

# Stop instance lama
Get-Process -Name "whatsar" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 500
Remove-Item $lock -Force -ErrorAction SilentlyContinue

# Jalankan di window terpisah (tetap hidup walau terminal ini ditutup)
Start-Process -FilePath $exe -WorkingDirectory $root -WindowStyle Minimized

Start-Sleep -Seconds 2

try {
    $health = Invoke-RestMethod -Uri "http://127.0.0.1:8080/health" -TimeoutSec 5
    Write-Host "Whatsar jalan!" -ForegroundColor Green
    Write-Host "  Admin  : http://127.0.0.1:8080/admin"
    Write-Host "  Health : http://127.0.0.1:8080/health"
    Write-Host "  Status : $($health.data.sessions_connected) session connected"
} catch {
    Write-Host 'Server belum merespons — cek window whatsar atau jalankan: whatsar.exe' -ForegroundColor Yellow
}