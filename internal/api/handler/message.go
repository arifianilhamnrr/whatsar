package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/store"
	"github.com/whatsar/whatsar/internal/wa"
)

type Message struct {
	Manager *wa.Manager
}

type sendMessageReq struct {
	SessionID string `json:"session_id"`
	To        string `json:"to"`
	Text      string `json:"text"`
}

func (h *Message) Send(w http.ResponseWriter, r *http.Request) {
	var req sendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Body tidak valid")
		return
	}
	if req.SessionID == "" || req.To == "" || req.Text == "" {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "session_id, to, dan text wajib diisi")
		return
	}

	msgID, err := h.Manager.SendText(r.Context(), req.SessionID, req.To, req.Text)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "SEND_FAILED", err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{
		"message_id": msgID,
		"status":     "sent",
	})
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