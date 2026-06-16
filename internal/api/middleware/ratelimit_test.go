package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitPerKey(t *testing.T) {
	hits := 0
	handler := RateLimit(3)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))

	keyA := "key-a"
	keyB := "key-b"

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(WithAPIKey(req.Context(), keyA))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: %d", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithAPIKey(req.Context(), keyA))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithAPIKey(req.Context(), keyB))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("other key should pass: %d", rec.Code)
	}
}