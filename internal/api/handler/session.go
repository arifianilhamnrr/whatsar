package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/wa"
)

type Session struct {
	Manager *wa.Manager
}

type createSessionReq struct {
	Name string `json:"name"`
}

func (h *Session) Create(w http.ResponseWriter, r *http.Request) {
	var req createSessionReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	sess, err := h.Manager.Create(r.Context(), req.Name)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	info, _ := h.Manager.GetSessionInfo(r.Context(), sess.ID)
	httputil.JSON(w, http.StatusCreated, info)
}

func (h *Session) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.Manager.ListSessionInfos(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	if list == nil {
		list = []wa.SessionInfo{}
	}
	httputil.JSON(w, http.StatusOK, list)
}

func (h *Session) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	info, err := h.Manager.GetSessionInfo(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, info)
}

func (h *Session) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Manager.Delete(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (h *Session) QR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	qr, err := h.Manager.GetQRInfo(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, qr)
}

func (h *Session) Status(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	info, err := h.Manager.GetSessionInfo(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{
		"id":        info.ID,
		"status":    info.Status,
		"connected": info.Connected,
		"phone":     info.Phone,
	})
}