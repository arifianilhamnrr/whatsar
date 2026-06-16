package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/whatsar/whatsar/internal/wa"
)

func TestHealthEndpoint(t *testing.T) {
	tmp := t.TempDir()
	mgr, err := wa.NewManager(wa.Options{
		DataDir:     tmp,
		AppDBPath:   filepath.Join(tmp, "whatsar.db"),
		MaxSessions: 2,
		LogLevel:    "error",
	})
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	defer mgr.Close()

	h := &Health{Manager: mgr}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Data.Status != "ok" {
		t.Fatalf("unexpected body: %+v", body)
	}
}