package admin

import (
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/whatsar/whatsar/web"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	tplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"statusBadge": statusBadge,
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "…"
		},
		"fmtTime": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Local().Format("02/01 15:04")
		},
		"fmtTimeNow": func() string {
			return time.Now().Local().Format("15:04:05")
		},
	}

	tpl, err := template.New("").Funcs(funcs).ParseFS(tplFS, "*.html", "partials/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{templates: tpl}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func statusBadge(status string) string {
	switch status {
	case "connected":
		return `<span class="badge ok">connected</span>`
	case "qr_ready", "connecting":
		return `<span class="badge warn">` + status + `</span>`
	case "failed", "stopped":
		return `<span class="badge err">` + status + `</span>`
	default:
		return `<span class="badge">` + status + `</span>`
	}
}