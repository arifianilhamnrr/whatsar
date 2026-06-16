package admin

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/whatsar/whatsar/internal/config"
	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/wa"
)

type Handler struct {
	Cfg    *config.Config
	Mgr    *wa.Manager
	Render *Renderer
	Auth   *PasswordStore
	Keys   *APIKeyStore
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if isAuthed(r.Context(), h.Auth, r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	h.Render.Render(w, "login.html", map[string]any{})
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token, err := h.Auth.Login(r.Context(), r.FormValue("password"))
	if err != nil {
		h.Render.Render(w, "login.html", map[string]any{"Error": "Password salah"})
		return
	}
	SetSession(w, token)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, "dashboard", map[string]any{"Title": "Dashboard"})
}

func (h *Handler) DashboardPartial(w http.ResponseWriter, r *http.Request) {
	data, err := h.Mgr.GetDashboard(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "partials/dashboard_content.html", data)
}

func (h *Handler) SessionsPartial(w http.ResponseWriter, r *http.Request) {
	data, _ := h.Mgr.GetDashboard(r.Context())
	h.Render.Render(w, "partials/sessions_table.html", data)
}

func (h *Handler) NewSessionPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, "session_new", map[string]any{
		"Title":     "Session Baru",
		"SessionID": "",
	})
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := r.FormValue("name")
	sess, err := h.Mgr.Create(r.Context(), name)
	if err != nil {
		h.Render.Render(w, "session_new", map[string]any{
			"Title":     "Session Baru",
			"SessionID": "",
			"Error":     err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/admin/sessions/new?id="+sess.ID, http.StatusSeeOther)
}

func (h *Handler) NewSessionQR(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.NewSessionPage(w, r)
		return
	}
	h.Render.Render(w, "session_new", map[string]any{
		"Title":     "Scan QR",
		"SessionID": id,
	})
}

func (h *Handler) SessionDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var info wa.SessionInfo
	if full, err := h.Mgr.GetDashboard(r.Context()); err == nil {
		for _, s := range full.Sessions {
			if s.ID == id {
				info = s
				break
			}
		}
	}
	if info.ID == "" {
		basic, err := h.Mgr.GetSessionInfo(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		info = *basic
	}

	h.Render.Render(w, "session_detail", map[string]any{
		"Title": "Detail Session",
		"Info":  info,
	})
}

func (h *Handler) QRPartial(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	qr, err := h.Mgr.GetQRInfo(r.Context(), id)
	if err != nil {
		h.Render.Render(w, "partials/qr_box.html", map[string]any{
			"SessionID": id,
			"Error":     err.Error(),
		})
		return
	}
	h.Render.Render(w, "partials/qr_box.html", map[string]any{
		"SessionID":   id,
		"Status":      qr.Status,
		"Code":        qr.Code,
		"ImageBase64": qr.ImageBase64,
	})
}

func (h *Handler) StatusPartial(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	info, err := h.Mgr.GetSessionInfo(r.Context(), id)
	if err != nil {
		h.Render.Render(w, "partials/status_box.html", map[string]any{
			"Error": err.Error(),
		})
		return
	}
	h.Render.Render(w, "partials/status_box.html", map[string]any{
		"Status":    info.Status,
		"Connected": info.Connected,
		"Phone":     info.Phone,
	})
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = h.Mgr.Delete(r.Context(), id)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	to := r.FormValue("to")
	text := r.FormValue("text")

	msgID, err := h.Mgr.SendText(r.Context(), id, to, text)
	if err != nil {
		h.Render.Render(w, "partials/send_result.html", map[string]any{"Error": err.Error()})
		return
	}
	h.Render.Render(w, "partials/send_result.html", map[string]any{"MessageID": msgID})
}

func (h *Handler) MessagesPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, "messages", map[string]any{"Title": "Log Pesan"})
}

func (h *Handler) DocsPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, "docs", map[string]any{
		"Title":            "Dokumentasi API",
		"BaseURL":          httputil.BaseURL(r, h.Cfg.PublicURL),
		"PublicURLOverride": h.Cfg.PublicURL != "",
	})
}

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, "settings", h.settingsData(r.Context(), nil))
}

func (h *Handler) settingsData(ctx context.Context, extra map[string]any) map[string]any {
	data := map[string]any{"Title": "Pengaturan"}
	if key, err := h.Keys.Get(ctx); err == nil {
		data["APIKeyMasked"] = h.Keys.Mask(key)
		data["HasAPIKey"] = key != ""
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func (h *Handler) ChangePasswordPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	current := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirm := r.FormValue("confirm_password")

	token, err := h.Auth.ChangePassword(r.Context(), current, newPassword, confirm)
	if err != nil {
		h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{"Error": err.Error()}))
		return
	}
	SetSession(w, token)
	h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{"Success": "Password berhasil diubah."}))
}

func (h *Handler) RegenerateAPIKeyPost(w http.ResponseWriter, r *http.Request) {
	key, err := h.Keys.Regenerate(r.Context())
	if err != nil {
		h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{"APIKeyError": err.Error()}))
		return
	}
	h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{
		"APIKeySuccess": "API key baru dibuat. Salin sekarang — tidak ditampilkan lagi setelah refresh.",
		"NewAPIKey":     key,
	}))
}

func (h *Handler) SetAPIKeyPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	key := r.FormValue("api_key")
	if err := h.Keys.Set(r.Context(), key); err != nil {
		h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{"APIKeyError": err.Error()}))
		return
	}
	h.Render.Render(w, "settings", h.settingsData(r.Context(), map[string]any{
		"APIKeySuccess": "API key berhasil disimpan.",
	}))
}

func (h *Handler) MessagesPartial(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	var msgs any
	var err error
	if sessionID != "" {
		msgs, err = h.Mgr.ListMessages(r.Context(), sessionID, 50, 0)
	} else {
		msgs, err = h.Mgr.ListAllMessages(r.Context(), 50, 0)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "partials/messages_table.html", map[string]any{"Messages": msgs})
}