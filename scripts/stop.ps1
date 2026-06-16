# Whatsar — Stop server
# Usage: .\scripts\stop.ps1

$root = Split-Path $PSScriptRoot -Parent
$lock = Join-Path $root "data\.whatsar.lock"

Get-Process -Name "whatsar" -ErrorAction SilentlyContinue | Stop-Process -Force
Remove-Item $lock -Force -ErrorAction SilentlyContinue
Write-Host "Whatsar dihentikan." -ForegroundColor Green