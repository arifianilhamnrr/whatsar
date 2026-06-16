package httputil

import (
	"net/http"
	"strings"
)

// BaseURL returns the public base URL for API examples and docs.
// Priority: publicOverride (WHATSAR_PUBLIC_URL) → proxy headers → request host.
func BaseURL(r *http.Request, publicOverride string) string {
	if u := strings.TrimSpace(publicOverride); u != "" {
		return strings.TrimRight(u, "/")
	}

	scheme := requestScheme(r)
	host := requestHost(r)
	if host == "" {
		return scheme + "://localhost"
	}
	return scheme + "://" + host
}

func requestScheme(r *http.Request) string {
	if proto := firstHeaderValue(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	if scheme := firstHeaderValue(r.Header.Get("X-Forwarded-Scheme")); scheme != "" {
		return scheme
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(r *http.Request) string {
	if host := firstHeaderValue(r.Header.Get("X-Forwarded-Host")); host != "" {
		return host
	}
	return r.Host
}

func firstHeaderValue(v string) string {
	if v == "" {
		return ""
	}
	if idx := strings.Index(v, ","); idx >= 0 {
		v = v[:idx]
	}
	return strings.TrimSpace(v)
}