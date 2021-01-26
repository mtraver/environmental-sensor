package main

import (
	"html/template"
	"net/http"

	"github.com/mtraver/environmental-sensor/web/cache"
	"github.com/mtraver/gaelog"
)

type cachezHandler struct {
	Cache    cache.Cache
	Template *template.Template
}

func (h cachezHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r)

	if err := h.Template.ExecuteTemplate(w, "cachez", h.Cache.Stats()); err != nil {
		gaelog.Errorf(ctx, "Could not execute template: %v", err)
	}
}
