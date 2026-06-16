package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/whatsar/whatsar/internal/api/validate"
	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/store"
	"github.com/whatsar/whatsar/internal/wa"
)

type Message struct {
	Manager *wa.Manager
}

type sendMessageReq struct {
	SessionID      string `json:"session_id"`
	To             string `json:"to"`
	Text           string `json:"text"`
	Type           string `json:"type"`
	ImageURL       string `json:"image_url"`
	ImageBase64    string `json:"image_base64"`
	DocumentURL    string `json:"document_url"`
	DocumentBase64 string `json:"document_base64"`
	FileName       string `json:"filename"`
	MimeType       string `json:"mimetype"`
	Caption        string `json:"caption"`
	ReplyTo        string `json:"reply_to"`
	QuotedText     string `json:"quoted_text"`
	Retry          bool   `json:"retry"`
}

func (h *Message) Send(w http.ResponseWriter, r *http.Request) {
	var req sendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Body tidak valid")
		return
	}

	if err := validateSessionSend(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	msgType, _ := validate.MessageType(req.Type)

	result, err := h.Manager.SendOutgoing(r.Context(), wa.OutgoingMessage{
		SessionID:   req.SessionID,
		To:          req.To,
		Type:        msgType,
		Text:        req.Text,
		ImageURL:    req.ImageURL,
		ImageB64:    req.ImageBase64,
		DocumentURL: req.DocumentURL,
		DocumentB64: req.DocumentBase64,
		FileName:    req.FileName,
		MimeType:    req.MimeType,
		Caption:     req.Caption,
		ReplyTo:     req.ReplyTo,
		QuotedText:  req.QuotedText,
		QueueRetry:  req.Retry,
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

func validateSessionSend(req *sendMessageReq) error {
	if err := validate.SessionID(req.SessionID); err != nil {
		return err
	}
	if err := validate.Recipient(req.To); err != nil {
		return err
	}
	msgType, err := validate.MessageType(req.Type)
	if err != nil {
		return err
	}
	if err := validate.Caption(req.Caption); err != nil {
		return err
	}
	if err := validate.Filename(req.FileName); err != nil {
		return err
	}

	switch msgType {
	case "text":
		if err := validate.Text(req.Text, true); err != nil {
			return err
		}
	case "image":
		if req.ImageURL == "" && req.ImageBase64 == "" {
			return errRequired("image_url atau image_base64 wajib untuk tipe image")
		}
		if req.ImageURL != "" {
			if err := validate.HTTPURL(req.ImageURL); err != nil {
				return err
			}
		}
		if err := validate.Base64Payload(req.ImageBase64, "image_base64"); err != nil {
			return err
		}
	case "document":
		if req.DocumentURL == "" && req.DocumentBase64 == "" {
			return errRequired("document_url atau document_base64 wajib untuk tipe document")
		}
		if req.DocumentURL != "" {
			if err := validate.HTTPURL(req.DocumentURL); err != nil {
				return err
			}
		}
		if err := validate.Base64Payload(req.DocumentBase64, "document_base64"); err != nil {
			return err
		}
	}

	if req.Text != "" {
		if err := validate.Text(req.Text, false); err != nil {
			return err
		}
	}
	return nil
}

func errRequired(msg string) error {
	return &validationError{msg}
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

func (h *Message) List(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if err := validate.SessionID(sessionID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset, err := validate.Pagination(limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

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