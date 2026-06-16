package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/store"
	"github.com/whatsar/whatsar/internal/wa"
)

type Webhook struct {
	Manager *wa.Manager
}

type registerWebhookReq struct {
	URL       string   `json:"url"`
	SessionID string   `json:"session_id"`
	Events    []string `json:"events"`
	Secret    string   `json:"secret"`
}

func (h *Webhook) Register(w http.ResponseWriter, r *http.Request) {
	var req registerWebhookReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Body tidak valid")
		return
	}
	if req.URL == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "url wajib diisi")
		return
	}
	if len(req.Events) == 0 {
		req.Events = []string{"message.in"}
	}

	wh := &store.WebhookRecord{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		URL:       req.URL,
		Events:    req.Events,
		Secret:    req.Secret,
		Active:    true,
	}

	if err := h.Manager.RegisterWebhook(r.Context(), wh); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "REGISTER_FAILED", err.Error())
		return
	}

	httputil.JSON(w, http.StatusCreated, wh)
}