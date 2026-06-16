package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/store"
	"github.com/whatsar/whatsar/internal/wa"
)

type Message struct {
	Manager *wa.Manager
}

type sendMessageReq struct {
	SessionID   string `json:"session_id"`
	To          string `json:"to"`
	Text        string `json:"text"`
	Type        string `json:"type"`
	ImageURL    string `json:"image_url"`
	ImageBase64 string `json:"image_base64"`
	Caption     string `json:"caption"`
	ReplyTo     string `json:"reply_to"`
	QuotedText  string `json:"quoted_text"`
	Retry       bool   `json:"retry"`
}

func (h *Message) Send(w http.ResponseWriter, r *http.Request) {
	var req sendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Body tidak valid")
		return
	}
	if req.SessionID == "" || req.To == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "session_id dan to wajib diisi")
		return
	}

	msgType := strings.ToLower(strings.TrimSpace(req.Type))
	if msgType == "" {
		msgType = "text"
	}
	if msgType == "text" && req.Text == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "text wajib untuk tipe text")
		return
	}
	if msgType == "image" && req.ImageURL == "" && req.ImageBase64 == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "image_url atau image_base64 wajib untuk tipe image")
		return
	}

	result, err := h.Manager.SendOutgoing(r.Context(), wa.OutgoingMessage{
		SessionID:  req.SessionID,
		To:         req.To,
		Type:       msgType,
		Text:       req.Text,
		ImageURL:   req.ImageURL,
		ImageB64:   req.ImageBase64,
		Caption:    req.Caption,
		ReplyTo:    req.ReplyTo,
		QuotedText: req.QuotedText,
		QueueRetry: req.Retry,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "SEND_FAILED", err.Error())
		return
	}

	status := http.StatusOK
	if result.Queued {
		status = http.StatusAccepted
	}
	httputil.JSON(w, status, result)
}

func (h *Message) List(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "session_id query param wajib")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	msgs, err := h.Manager.ListMessages(r.Context(), sessionID, limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	if msgs == nil {
		msgs = []*store.MessageRecord{}
	}
	httputil.JSON(w, http.StatusOK, msgs)
}