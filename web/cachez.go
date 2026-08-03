package main

import (
	"html/template"
	"net/http"

	"github.com/mtraver/environmental-sensor/database"
	"github.com/mtraver/gaelog"
)

type cachezHandler struct {
	Database database.Database
	Template *template.Template
}

func (h cachezHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.Template.ExecuteTemplate(w, "cachez", h.Database.CacheStats()); err != nil {
		gaelog.Errorf(ctx, "Could not execute template: %v", err)
	}
}
