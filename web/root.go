package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/mtraver/environmental-sensor/measurement"
	"github.com/mtraver/environmental-sensor/web/device"
	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that we only serve the root.
	// From https://golang.org/pkg/net/http/#ServeMux:
	//   Note that since a pattern ending in a slash names a rooted subtree, the
	//   pattern "/" matches all paths not matched by other registered patterns,
	//   not just the URL with Path == "/".
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := appengine.NewContext(r)

	lg, err := gaelog.New(r)
	if err != nil {
		lg.Errorf("%v", err)
	}
	defer lg.Close()

	// By default display data up to defaultDataDisplayAgeHours hours old
	hoursAgo := defaultDataDisplayAgeHours
	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hoursAgo) * time.Hour)

	// These control which HTML forms are auto-filled when the page loads, to
	// reflect the data that is being displayed
	fillRangeForm := false
	fillHoursAgoForm := true

	if r.Method == "POST" {
		switch formName := r.FormValue("form-name"); formName {
		case "range":
			var rangeErr error
			startTime, rangeErr = time.Parse(time.RFC3339Nano, r.FormValue("startdate-adjusted"))
			if rangeErr != nil {
				http.Error(w, fmt.Sprintf("Bad start time: %v", rangeErr), http.StatusBadRequest)
				return
			}

			endTime, rangeErr = time.Parse(time.RFC3339Nano, r.FormValue("enddate-adjusted"))
			if rangeErr != nil {
				http.Error(w, fmt.Sprintf("Bad end time: %v", rangeErr), http.StatusBadRequest)
				return
			}

			fillRangeForm = true
			fillHoursAgoForm = false
		case "hoursago":
			var hoursAgoErr error
			hoursAgo, hoursAgoErr = strconv.Atoi(r.FormValue("hoursago"))
			if hoursAgoErr != nil {
				http.Error(w, fmt.Sprintf("%v", hoursAgoErr), http.StatusBadRequest)
				return
			}

			if hoursAgo < 1 {
				http.Error(w, fmt.Sprintf("Hours ago must be >= 1"), http.StatusBadRequest)
				return
			}

			endTime = time.Now().UTC()
			startTime = endTime.Add(-time.Duration(hoursAgo) * time.Hour)

			fillRangeForm = false
			fillHoursAgoForm = true
		default:
			http.Error(w, fmt.Sprintf("Unknown form name"), http.StatusBadRequest)
			return
		}
	}

	// Get measurements and marshal to JSON for use in the template
	measurements, err := database.GetMeasurementsBetween(ctx, startTime, endTime)
	jsonBytes := []byte{}
	if err != nil {
		lg.Errorf("Error fetching data: %v", err)
	} else {
		jsonBytes, err = measurementMapToJSON(measurements)
		if err != nil {
			lg.Errorf("Error marshaling measurements to JSON: %v", err)
		}
	}

	// Get the latest measurement for each device
	var latest map[string]measurement.StorableMeasurement
	ids, latestErr := device.GetDeviceIDs(ctx, projectID, iotcoreRegistry)
	if latestErr != nil {
		lg.Errorf("Error getting device IDs: %v", latestErr)
	} else {
		latest, latestErr = database.GetLatestMeasurements(ctx, ids)

		if latestErr != nil {
			lg.Errorf("Error getting latest measurements: %v", latestErr)
		}
	}

	data := struct {
		Measurements     template.JS
		Error            error
		StartTime        time.Time
		EndTime          time.Time
		HoursAgo         int
		FillRangeForm    bool
		FillHoursAgoForm bool
		Latest           map[string]measurement.StorableMeasurement
		LatestError      error
	}{
		Measurements:     template.JS(jsonBytes),
		Error:            err,
		StartTime:        startTime,
		EndTime:          endTime,
		HoursAgo:         hoursAgo,
		FillRangeForm:    fillRangeForm,
		FillHoursAgoForm: fillHoursAgoForm,
		Latest:           latest,
		LatestError:      latestErr,
	}

	if err := templates.ExecuteTemplate(w, "index", data); err != nil {
		lg.Errorf("Could not execute template: %v", err)
	}
}

// serializableMeasurement is the same as StorableMeasurement except without fields that
// the frontend doesn't need for plotting data.
// IMPORTANT: Keep up to date with the generated Measurement type, at least to the extent that is required.
type serializableMeasurement struct {
	// This timestamp is an offset from the epoch in milliseconds
	// (compare to Timestamp in StorableMeasurement).
	Timestamp int64   `json:"timestamp,omitempty" datastore:"timestamp"`
	Temp      float32 `json:"temp,omitempty" datastore:"temp"`
}

// measurementMapToJSON converts a string -> []StorableMeasurement map into a marshaled
// JSON array for use in the template. The JSON is an array with one element for each
// device ID. It's constructed this way, instead of as a map where keys are device IDs,
// because the JavaScript visualization package D3 (https://d3js.org/) works better with
// arrays of data than maps.
func measurementMapToJSON(measurements map[string][]measurement.StorableMeasurement) ([]byte, error) {
	type dataForTemplate struct {
		ID     string                    `json:"id"`
		Values []serializableMeasurement `json:"values"`
	}

	// Sort the map's keys so that the resulting JSON always has them in the same
	// order. This ensures that e.g. the color assigned to each line on a plot is
	// the same for every page load.
	keys := make([]string, len(measurements))
	i := 0
	for k := range measurements {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var data []dataForTemplate
	for _, k := range keys {
		vals := make([]serializableMeasurement, len(measurements[k]))
		for i, m := range measurements[k] {
			vals[i] = serializableMeasurement{m.Timestamp.Unix() * 1000, m.Temp}
		}
		data = append(data, dataForTemplate{k, vals})
	}
	return json.Marshal(data)
}
