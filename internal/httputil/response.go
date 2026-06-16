package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	RequestID string `json:"request_id"`
}

type envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Meta    Meta       `json:"meta"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{
		Success: true,
		Data:    data,
		Meta:    Meta{RequestID: uuid.New().String()},
	})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
		Meta:    Meta{RequestID: uuid.New().String()},
	})
}

func write(w http.ResponseWriter, status int, v envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}