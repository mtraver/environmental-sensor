package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"

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
	ctx := appengine.NewContext(r)

	lg, err := gaelog.New(r)
	if err != nil {
		lg.Errorf("%v", err)
	}
	defer lg.Close()

	endTime := time.Now().UTC()
	startTime := endTime.Add(-h.DelayedUploadsDur)

	measurements, err := h.Database.DelayedMeasurementsSince(ctx, startTime)
	if err != nil {
		lg.Errorf("Error fetching data: %v", err)
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
		lg.Errorf("Could not execute template: %v", err)
	}
}
