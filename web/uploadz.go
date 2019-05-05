package main

import (
	"net/http"
	"time"

	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"

	"github.com/mtraver/environmental-sensor/measurement"
)

// uploadz will display delayed uploads up to this many hours old.
const delayedUploadsHours = 48

func uploadzHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	lg, err := gaelog.New(r)
	if err != nil {
		lg.Errorf("%v", err)
	}
	defer lg.Close()

	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(delayedUploadsHours) * time.Hour)

	measurements, err := database.GetDelayedMeasurementsSince(ctx, startTime)
	if err != nil {
		lg.Errorf("Error fetching data: %v", err)
	}

	total := 0
	for _, m := range measurements {
		total += len(m)
	}

	data := struct {
		DelayedUploadsHours int
		DelayedMeasurements map[string][]measurement.StorableMeasurement
		DelayedTotal        int
		Error               error
	}{
		DelayedUploadsHours: delayedUploadsHours,
		DelayedMeasurements: measurements,
		DelayedTotal:        total,
		Error:               err,
	}

	if err := templates.ExecuteTemplate(w, "uploadz", data); err != nil {
		lg.Errorf("Could not execute template: %v", err)
	}
}
