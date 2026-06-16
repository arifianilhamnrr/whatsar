package middleware

import (
	"net/http"

	"github.com/whatsar/whatsar/internal/admin"
	"github.com/whatsar/whatsar/internal/httputil"
)

func APIKeyAuth(keys *admin.APIKeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if key == "" || !keys.Validate(r.Context(), key) {
				httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "API key tidak valid")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithAPIKey(r.Context(), key)))
		})
	}
}