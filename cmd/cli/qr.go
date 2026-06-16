package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mdp/qrterminal/v3"
	qrcode "github.com/skip2/go-qrcode"
)

func displayQR(code, sessionID, dataDir string) {
	fmt.Println()
	fmt.Println("Scan QR di bawah via WhatsApp → Perangkat Tertaut → Tautkan perangkat")
	fmt.Println("(QR refresh otomatis setiap ~20 detik, tunggu QR baru jika expired)")
	fmt.Println()

	// Render ASCII QR di terminal
	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)

	// Simpan PNG — lebih gampang discan di Windows
	if err := os.MkdirAll(dataDir, 0o750); err == nil {
		pngPath := filepath.Join(dataDir, fmt.Sprintf("qr-%s.png", sessionID[:8]))
		if err := qrcode.WriteFile(code, qrcode.Medium, 256, pngPath); err == nil {
			fmt.Printf("\nQR disimpan: %s\n", pngPath)
			openFile(pngPath)
		}
	}

	fmt.Println()
	fmt.Println("─────────────────────────────────────────")
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}