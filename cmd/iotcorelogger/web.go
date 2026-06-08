package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05 MST (-0700)"
)

type rootHandler struct {
	templates *template.Template
	mon       *Monitor
}

func (h *rootHandler) formatConnectionTimestamp(t *time.Time, now time.Time) string {
	s := "disconnected"
	if t != nil {
		dur := now.Sub(*t).Round(time.Millisecond)
		s = fmt.Sprintf("%s (%v ago)", t.Format(timeFormat), dur)
	}

	return s
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

	h.mon.connectionMetricsMu.RLock()
	defer h.mon.connectionMetricsMu.RUnlock()

	data := struct {
		DeviceID       string
		ConfigVersion  int
		Config         string
		FirstConnected string
		LastConnected  string
		ReconnectCount int
		GitRevision    string
		BuildTime      string
	}{
		DeviceID:       h.mon.device.ID(),
		ConfigVersion:  h.mon.configVersion,
		Config:         h.mon.config.String(),
		FirstConnected: h.formatConnectionTimestamp(h.mon.firstConnectTime, now),
		LastConnected:  h.formatConnectionTimestamp(h.mon.lastConnectTime, now),
		ReconnectCount: int(math.Max(0, float64(h.mon.connectionCount-1))),
		GitRevision:    gitRevision,
		BuildTime:      buildTime,
	}

	if err := h.templates.ExecuteTemplate(w, "index", data); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
