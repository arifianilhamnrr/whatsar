package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/whatsar/whatsar/internal/admin"
	"github.com/whatsar/whatsar/internal/api"
	"github.com/whatsar/whatsar/internal/config"
	"github.com/whatsar/whatsar/internal/wa"
	"github.com/whatsar/whatsar/internal/webhook"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dataDir := filepath.Dir(cfg.DBPath)

	// Placeholder manager untuk dispatcher wiring
	var dispatcher *webhook.Dispatcher

	mgr, err := wa.NewManager(wa.Options{
		DataDir:     dataDir,
		AppDBPath:   cfg.DBPath,
		MaxSessions: cfg.MaxSessions,
		LogLevel:    cfg.LogLevel,
		OnMessage: func(msg wa.IncomingMessage) {
			log.Printf("[wa] %s ← %s: %s", msg.SessionID[:8], msg.From, msg.Body)
			if dispatcher != nil {
				dispatcher.Handle(msg)
			}
		},
	})
	if err != nil {
		log.Fatalf("wa manager: %v", err)
	}
	defer mgr.Close()

	dispatcher = webhook.NewDispatcher(mgr)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	mgr.StartQueueWorker(rootCtx)

	// Reconnect paired sessions
	for _, sess := range mgr.List() {
		if sess.Phone() != "" {
			go func(s *wa.Session) {
				if err := mgr.Connect(context.Background(), s.ID); err != nil {
					log.Printf("reconnect %s: %v", s.ID, err)
				}
			}(sess)
		}
	}

	renderer, err := admin.NewRenderer()
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}

	adminAuth := admin.NewPasswordStore(mgr.AppDB(), cfg.AdminPassword)
	adminKeys := admin.NewAPIKeyStore(mgr.AppDB(), cfg.APIKey)
	adminH := &admin.Handler{Cfg: cfg, Mgr: mgr, Render: renderer, Auth: adminAuth, Keys: adminKeys}
	router := api.NewRouter(cfg, mgr, adminH)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("whatsar API → http://%s", addr)
		log.Printf("health check → http://%s/health", addr)
		log.Printf("admin UI    → http://%s/admin", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}