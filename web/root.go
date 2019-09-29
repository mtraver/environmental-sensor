package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/mtraver/environmental-sensor/measurement"
	"github.com/mtraver/environmental-sensor/web/device"
	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"
)

// rootHandler renders the page for the root URL, which includes a plot and latest measurements.
type rootHandler struct {
	ProjectID         string
	IoTCoreRegistry   string
	DefaultDisplayAge time.Duration
	Database          Database
	Template          *template.Template
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var daysAgo float64
	hoursAgo := int(h.DefaultDisplayAge.Round(time.Hour).Hours())
	endTime := time.Now().UTC()
	startTime := endTime.Add(-time.Duration(hoursAgo) * time.Hour)

	// These control which HTML forms are auto-filled when the page loads, to
	// reflect the data that is being displayed
	fillRangeForm := false
	fillDaysAgoForm := false
	fillHoursAgoForm := true

	if r.Method == "POST" {
		switch formName := r.FormValue("form-name"); formName {
		case "range":
			var err error
			startTime, err = time.Parse(time.RFC3339Nano, r.FormValue("startdate-adjusted"))
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad start time: %v", err), http.StatusBadRequest)
				return
			}

			endTime, err = time.Parse(time.RFC3339Nano, r.FormValue("enddate-adjusted"))
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad end time: %v", err), http.StatusBadRequest)
				return
			}

			fillRangeForm = true
			fillDaysAgoForm = false
			fillHoursAgoForm = false
		case "daysago":
			var err error
			daysAgo, err = strconv.ParseFloat(r.FormValue("daysago"), 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
				return
			}

			if daysAgo <= 0 {
				http.Error(w, fmt.Sprintf("Days ago must be > 0"), http.StatusBadRequest)
				return
			}

			endTime = time.Now().UTC()
			startTime = endTime.Add(-time.Duration(math.Round(daysAgo*24)) * time.Hour)

			fillRangeForm = false
			fillDaysAgoForm = true
			fillHoursAgoForm = false
		case "hoursago":
			var err error
			hoursAgo, err = strconv.Atoi(r.FormValue("hoursago"))
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
				return
			}

			if hoursAgo < 1 {
				http.Error(w, fmt.Sprintf("Hours ago must be >= 1"), http.StatusBadRequest)
				return
			}

			endTime = time.Now().UTC()
			startTime = endTime.Add(-time.Duration(hoursAgo) * time.Hour)

			fillRangeForm = false
			fillDaysAgoForm = false
			fillHoursAgoForm = true
		default:
			http.Error(w, fmt.Sprintf("Unknown form name"), http.StatusBadRequest)
			return
		}
	}

	var wg sync.WaitGroup

	// Get measurements and marshal to JSON for use in the template
	start := time.Now()
	measurements, err := h.Database.Between(ctx, startTime, endTime)
	elapsed := time.Since(start)
	lg.Infof("Done getting measurements; took %v", elapsed)

	jsonBytes := []byte{}
	var stats map[string]Stats
	if err != nil {
		lg.Errorf("Error fetching data: %v", err)
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			jsonBytes, err = measurementMapToJSON(measurements)
			if err != nil {
				lg.Errorf("Error marshaling measurements to JSON: %v", err)
			}

			elapsed := time.Since(start)
			lg.Infof("Done marshaling measurements to JSON; took %v", elapsed)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			stats = summaryStats(measurements)

			elapsed := time.Since(start)
			lg.Infof("Done computing stats; took %v", elapsed)
		}()
	}

	// Get the latest measurement for each device
	var latest map[string]measurement.StorableMeasurement
	var latestErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()

		ids, err := device.GetDeviceIDs(ctx, h.ProjectID, h.IoTCoreRegistry)
		if err != nil {
			latestErr = err
			lg.Errorf("Error getting device IDs: %v", err)
			return
		}

		latest, latestErr = h.Database.Latest(ctx, ids)
		if latestErr != nil {
			lg.Errorf("Error getting latest measurements: %v", latestErr)
		}

		elapsed := time.Since(start)
		lg.Infof("Done getting latest measurements; took %v", elapsed)
	}()

	wg.Wait()

	data := struct {
		Measurements     template.JS
		Stats            map[string]Stats
		Error            error
		StartTime        time.Time
		EndTime          time.Time
		DaysAgo          float64
		HoursAgo         int
		FillRangeForm    bool
		FillDaysAgoForm  bool
		FillHoursAgoForm bool
		Latest           map[string]measurement.StorableMeasurement
		LatestError      error
	}{
		Measurements:     template.JS(jsonBytes),
		Stats:            stats,
		Error:            err,
		StartTime:        startTime,
		EndTime:          endTime,
		DaysAgo:          daysAgo,
		HoursAgo:         hoursAgo,
		FillRangeForm:    fillRangeForm,
		FillDaysAgoForm:  fillDaysAgoForm,
		FillHoursAgoForm: fillHoursAgoForm,
		Latest:           latest,
		LatestError:      latestErr,
	}

	if err := h.Template.ExecuteTemplate(w, "index", data); err != nil {
		lg.Errorf("Could not execute template: %v", err)
	}
}

// serializableMeasurement is the same as StorableMeasurement except without fields that
// the frontend doesn't need for plotting data. The JSON keys are also as short as possible,
// as that reduces the size of the JSON object served to the user.
// IMPORTANT: Keep up to date with the generated Measurement type, at least to the extent that is required.
type serializableMeasurement struct {
	// This timestamp is an offset from the epoch in milliseconds (compare to Timestamp in StorableMeasurement).
	Timestamp int64   `json:"ts,omitempty" datastore:"timestamp"`
	Temp      float32 `json:"t,omitempty" datastore:"temp"`
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
