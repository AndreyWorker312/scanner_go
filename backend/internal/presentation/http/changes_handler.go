package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/domain/models"
)

// ChangesRepository is the minimal interface the handler needs.
type ChangesRepository interface {
	GetChangeEvents(limit int, severity string) ([]models.ChangeEvent, error)
	GetChangeEventsSince(since time.Time) ([]models.ChangeEvent, error)
	DeleteChangeEvents() error
}

// ChangesHandler serves change-detection endpoints.
type ChangesHandler struct {
	repo ChangesRepository
}

func NewChangesHandler(repo ChangesRepository) *ChangesHandler {
	return &ChangesHandler{repo: repo}
}

// GET /api/changes?limit=100&severity=CRITICAL
func (h *ChangesHandler) GetChanges(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	severity := r.URL.Query().Get("severity")

	events, err := h.repo.GetChangeEvents(limit, severity)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if events == nil {
		events = []models.ChangeEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    events,
		"count":   len(events),
	})
}

// DELETE /api/changes/delete
func (h *ChangesHandler) DeleteChanges(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteChangeEvents(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GET /api/changes/stream  — Server-Sent Events
//
// The client receives:
//   - event: ping     on connect and every 30 s (keepalive)
//   - event: change   each time a new ChangeEvent is detected
//
// Nginx must have `proxy_buffering off` and a long read-timeout for this
// endpoint (already set in /api/ block).
func (h *ChangesHandler) StreamChanges(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type",                "text/event-stream")
	w.Header().Set("Cache-Control",               "no-cache")
	w.Header().Set("Connection",                  "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering",           "no") // nginx: disable response buffer

	// ── initial burst: last 24 hours ────────────────────────────────────────
	since := time.Now().Add(-24 * time.Hour)
	initial, err := h.repo.GetChangeEvents(500, "")
	if err != nil {
		log.Printf("[SSE] initial load error: %v", err)
	} else {
		for i := len(initial) - 1; i >= 0; i-- { // oldest first
			evt := initial[i]
			if evt.CreatedAt.After(since) {
				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "event: change\ndata: %s\n\n", data)
			}
		}
	}
	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	// ── poll for new events every 5 s ───────────────────────────────────────
	lastCheck := time.Now()

	poll    := time.NewTicker(5 * time.Second)
	keepalive := time.NewTicker(30 * time.Second)
	defer poll.Stop()
	defer keepalive.Stop()

	for {
		select {
		case <-poll.C:
			news, err := h.repo.GetChangeEventsSince(lastCheck)
			if err != nil {
				continue
			}
			for _, evt := range news { // ascending — oldest first
				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "event: change\ndata: %s\n\n", data)
				if evt.CreatedAt.After(lastCheck) {
					lastCheck = evt.CreatedAt
				}
			}
			if len(news) > 0 {
				flusher.Flush()
			}

		case <-keepalive.C:
			fmt.Fprintf(w, "event: ping\ndata: keepalive\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

