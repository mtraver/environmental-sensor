package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mtraver/gaelog"

	"github.com/mtraver/environmental-sensor/measurement"
)

// uploadzHandler renders a page displaying data about delayed data uploads from devices.
type uploadzHandler struct {
	// Display delayed uploads up to this duration old.
	DelayedUploadsDur time.Duration
	Database          Database
	Template          *template.Template
}

func (h uploadzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r)

	endTime := time.Now().UTC()
	startTime := endTime.Add(-h.DelayedUploadsDur)

	measurements, err := h.Database.DelayedSince(ctx, startTime)
	if err != nil {
		gaelog.Errorf(ctx, "Error fetching data: %v", err)
	}

	total := 0
	for _, m := range measurements {
		total += len(m)
	}

	data := struct {
		DelayedUploadsDur   time.Duration
		DelayedMeasurements map[string][]measurement.StorableMeasurement
		DelayedTotal        int
		Error               error
	}{
		DelayedUploadsDur:   h.DelayedUploadsDur,
		DelayedMeasurements: measurements,
		DelayedTotal:        total,
		Error:               err,
	}

	if err := h.Template.ExecuteTemplate(w, "uploadz", data); err != nil {
		gaelog.Errorf(ctx, "Could not execute template: %v", err)
	}
}
