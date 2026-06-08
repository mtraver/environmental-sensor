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

func (h *rootHandler) formatTimestamp(t *time.Time, now time.Time, nilVal string) string {
	if t == nil {
		return nilVal
	}

	dur := now.Sub(*t).Round(time.Millisecond)
	return fmt.Sprintf("%s (%v ago)", t.Format(timeFormat), dur)
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

	h.mon.connectionMetricsMu.RLock()
	defer h.mon.connectionMetricsMu.RUnlock()

	h.mon.publishMetricsMu.RLock()
	defer h.mon.publishMetricsMu.RUnlock()

	data := struct {
		DeviceID            string
		ConfigVersion       int
		Config              string
		FirstConnected      string
		LastConnected       string
		ReconnectCount      int
		LastPublish         string
		PublishCount        int
		PublishFailureCount int
		GitRevision         string
		BuildTime           string
	}{
		DeviceID:            h.mon.device.ID(),
		ConfigVersion:       h.mon.configVersion,
		Config:              h.mon.config.String(),
		FirstConnected:      h.formatTimestamp(h.mon.firstConnectTime, now, "disconnected"),
		LastConnected:       h.formatTimestamp(h.mon.lastConnectTime, now, "disconnected"),
		ReconnectCount:      int(math.Max(0, float64(h.mon.connectionCount-1))),
		LastPublish:         h.formatTimestamp(h.mon.lastPublishTime, now, "never"),
		PublishCount:        h.mon.successfulPublishCount,
		PublishFailureCount: h.mon.publishFailureCount,
		GitRevision:         gitRevision,
		BuildTime:           buildTime,
	}

	if err := h.templates.ExecuteTemplate(w, "index", data); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
