package admin

import (
	"context"
	"net/http"
)

const sessionCookie = "whatsar_admin"

func AuthMiddleware(auth *PasswordStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth.ValidateSession(r.Context(), r) {
				next.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		})
	}
}

func SetSession(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 7,
	})
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// Legacy helper for login page redirect check.
func isAuthed(ctx context.Context, auth *PasswordStore, r *http.Request) bool {
	return auth.ValidateSession(ctx, r)
}