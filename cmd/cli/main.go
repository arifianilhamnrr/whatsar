package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/whatsar/whatsar/internal/config"
	"github.com/whatsar/whatsar/internal/wa"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dataDir := filepath.Dir(cfg.DBPath)

	mgr, err := wa.NewManager(wa.Options{
		DataDir:     dataDir,
		AppDBPath:   cfg.DBPath,
		MaxSessions: cfg.MaxSessions,
		LogLevel:    cfg.LogLevel,
		OnMessage: func(msg wa.IncomingMessage) {
			fmt.Printf("\n📩 [%s] %s: %s\n> ", msg.SessionID[:8], msg.From, msg.Body)
		},
	})
	if err != nil {
		log.Fatalf("wa manager: %v", err)
	}
	defer mgr.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "pair":
		runPair(ctx, mgr, dataDir, os.Args[2:])
	case "send":
		runSend(ctx, mgr, os.Args[2:])
	case "status":
		runStatus(mgr)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Whatsar CLI — Fase 1

Usage:
  whatsar-cli pair [nama]          Buat session & tampilkan QR
  whatsar-cli send <session> <to> <pesan>
  whatsar-cli status               Lihat session aktif`)
}

func runPair(ctx context.Context, mgr *wa.Manager, dataDir string, args []string) {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	sess, err := mgr.Create(ctx, name)
	if err != nil {
		log.Fatalf("create session: %v", err)
	}

	fmt.Printf("Session dibuat: %s (%s)\n", sess.ID, name)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastQR := ""
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qr := sess.GetQR()
			st := sess.GetStatus()

			if qr != "" && qr != lastQR {
				lastQR = qr
				displayQR(qr, sess.ID, dataDir)
			}

			if st == wa.StatusConnected {
				fmt.Printf("\n✅ Terhubung! Nomor: %s\n", sess.Phone())
				fmt.Println("Listening pesan masuk... (Ctrl+C untuk keluar)")
				interactiveSend(ctx, mgr, sess.ID)
				return
			}

			if st == wa.StatusFailed || st == wa.StatusStopped {
				log.Fatalf("pairing gagal (status: %s)", st)
			}
		}
	}
}

func interactiveSend(ctx context.Context, mgr *wa.Manager, sessionID string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Print("> ")
			continue
		}
		if line == "/quit" || line == "/exit" {
			return
		}

		// format: nomor pesan  →  628123456789 halo cuy
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			fmt.Println("format: <nomor> <pesan>")
			fmt.Print("> ")
			continue
		}

		msgID, err := mgr.SendText(ctx, sessionID, parts[0], parts[1])
		if err != nil {
			fmt.Printf("gagal kirim: %v\n", err)
		} else {
			fmt.Printf("terkirim ✓ (id: %s)\n", msgID)
		}
		fmt.Print("> ")
	}
}

func runSend(ctx context.Context, mgr *wa.Manager, args []string) {
	if len(args) < 3 {
		log.Fatal("usage: send <session_id> <nomor> <pesan>")
	}

	sessionID := args[0]
	to := args[1]
	text := strings.Join(args[2:], " ")

	sess, err := mgr.Get(sessionID)
	if err != nil {
		log.Fatalf("session: %v", err)
	}

	if !sess.IsConnected() {
		if err := mgr.Connect(ctx, sessionID); err != nil {
			log.Fatalf("connect: %v", err)
		}
		time.Sleep(3 * time.Second)
	}

	msgID, err := mgr.SendText(ctx, sessionID, to, text)
	if err != nil {
		log.Fatalf("send: %v", err)
	}

	fmt.Printf("terkirim ✓ (id: %s)\n", msgID)
}

func runStatus(mgr *wa.Manager) {
	sessions := mgr.List()
	if len(sessions) == 0 {
		fmt.Println("belum ada session")
		return
	}

	for _, s := range sessions {
		fmt.Printf("- %s | %s | status: %s | phone: %s | connected: %v\n",
			s.ID, s.Name, s.GetStatus(), s.Phone(), s.IsConnected())
	}
}