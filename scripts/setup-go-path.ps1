# Tambahkan Go ke PATH session Laragon
# Usage: . .\scripts\setup-go-path.ps1

$goBin = "C:\laragon\bin\go\go\bin"
if (Test-Path $goBin) {
    if ($env:Path -notlike "*$goBin*") {
        $env:Path = "$goBin;$env:Path"
    }
    Write-Host "Go ready: $(go version)"
} else {
    Write-Warning "Go tidak ditemukan di $goBin — jalankan install Go dulu"
}