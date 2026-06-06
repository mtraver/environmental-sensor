package main

import (
	"html/template"
	"log"
	"net/http"
)

type rootHandler struct {
	templates *template.Template
	mon       *Monitor
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := struct {
		DeviceID      string
		ConfigVersion int
		Config        string
	}{
		DeviceID:      h.mon.device.ID(),
		ConfigVersion: h.mon.configVersion,
		Config:        h.mon.config.String(),
	}

	if err := h.templates.ExecuteTemplate(w, "index", data); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
