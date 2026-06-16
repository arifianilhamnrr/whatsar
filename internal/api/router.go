package api

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/whatsar/whatsar/internal/admin"
	"github.com/whatsar/whatsar/internal/api/handler"
	"github.com/whatsar/whatsar/internal/api/middleware"
	"github.com/whatsar/whatsar/internal/config"
	"github.com/whatsar/whatsar/internal/wa"
	"github.com/whatsar/whatsar/web"
)

func NewRouter(cfg *config.Config, mgr *wa.Manager, adminH *admin.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Logger)

	health := &handler.Health{Manager: mgr}
	r.Get("/health", health.ServeHTTP)

	sessH := &handler.Session{Manager: mgr}
	msgH := &handler.Message{Manager: mgr}
	whH := &handler.Webhook{Manager: mgr}

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(adminH.Keys))
		r.Use(middleware.RateLimit(60))

		r.Post("/sessions", sessH.Create)
		r.Get("/sessions", sessH.List)
		r.Get("/sessions/{id}", sessH.Get)
		r.Delete("/sessions/{id}", sessH.Delete)
		r.Get("/sessions/{id}/qr", sessH.QR)
		r.Get("/sessions/{id}/status", sessH.Status)

		r.Post("/messages/send", msgH.Send)
		r.Get("/messages", msgH.List)

		r.Post("/webhooks", whH.Register)
	})

	// Static assets (Pico + HTMX)
	staticFS, _ := fs.Sub(web.Static, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Admin UI
	r.Get("/admin/login", adminH.LoginPage)
	r.Post("/admin/login", adminH.LoginPost)

	r.Route("/admin", func(r chi.Router) {
		r.Use(admin.AuthMiddleware(adminH.Auth))
		r.Get("/logout", adminH.Logout)
		r.Get("/", adminH.Dashboard)
		r.Get("/partials/dashboard", adminH.DashboardPartial)
		r.Get("/partials/sessions", adminH.SessionsPartial)
		r.Get("/sessions/new", adminH.NewSessionQR)
		r.Post("/sessions", adminH.CreateSession)
		r.Get("/sessions/{id}", adminH.SessionDetail)
		r.Get("/sessions/{id}/qr-partial", adminH.QRPartial)
		r.Get("/sessions/{id}/status-partial", adminH.StatusPartial)
		r.Delete("/sessions/{id}", adminH.DeleteSession)
		r.Post("/sessions/{id}/send", adminH.SendMessage)
		r.Get("/messages", adminH.MessagesPage)
		r.Get("/partials/messages", adminH.MessagesPartial)
		r.Get("/docs", adminH.DocsPage)
		r.Get("/settings", adminH.SettingsPage)
		r.Post("/settings/password", adminH.ChangePasswordPost)
		r.Post("/settings/api-key/regenerate", adminH.RegenerateAPIKeyPost)
		r.Post("/settings/api-key", adminH.SetAPIKeyPost)
	})

	return r
}