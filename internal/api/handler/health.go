package handler

import (
	"net/http"
	"time"

	"github.com/whatsar/whatsar/internal/httputil"
	"github.com/whatsar/whatsar/internal/wa"
)

var startTime = time.Now()

type Health struct {
	Manager *wa.Manager
}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessions := h.Manager.List()
	connected := 0
	for _, s := range sessions {
		if s.IsConnected() {
			connected++
		}
	}

	pendingQueue, _ := h.Manager.AppDB().CountPendingQueue(r.Context())

	httputil.JSON(w, http.StatusOK, map[string]any{
		"status":             "ok",
		"uptime_seconds":     int(time.Since(startTime).Seconds()),
		"sessions_total":     len(sessions),
		"sessions_connected": connected,
		"queue_pending":      pendingQueue,
	})
}