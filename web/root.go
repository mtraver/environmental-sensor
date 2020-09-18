package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mtraver/environmental-sensor/aqi"
	"github.com/mtraver/environmental-sensor/measurement"
	"github.com/mtraver/environmental-sensor/web/device"
	"github.com/mtraver/gaelog"
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

	ctx := newContext(r)

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
	metrics := []measurement.Metric{}
	elapsed := time.Since(start)
	gaelog.Infof(ctx, "Done getting measurements; took %v", elapsed)

	jsonBytes := []byte{}
	var stats map[string]Stats
	if err != nil {
		gaelog.Errorf(ctx, "Error fetching data: %v", err)
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			jsonBytes, err = measurementMapToJSON(measurements)
			if err != nil {
				gaelog.Errorf(ctx, "Error marshaling measurements to JSON: %v", err)
			}

			elapsed := time.Since(start)
			gaelog.Infof(ctx, "Done marshaling measurements to JSON; took %v", elapsed)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			// Get the full set of metrics included in the data. This is used to
			// set up plots for each metric.
			mm := make(map[string]bool)
			for _, v := range measurements {
				for _, sm := range v {
					for metricName, _ := range sm.ValueMap() {
						mm[metricName] = true
					}
				}
			}
			for metricName, _ := range mm {
				metric, ok := measurement.GetMetric(metricName)
				if !ok {
					gaelog.Errorf(ctx, "Error getting metrics: unknown metric %q", metricName)
					continue
				}
				metrics = append(metrics, metric)
			}
			sort.Slice(metrics, func(i, j int) bool {
				return metrics[i].Name < metrics[j].Name
			})

			elapsed := time.Since(start)
			gaelog.Infof(ctx, "Done getting metric set; took %v", elapsed)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			stats = summaryStats(measurements)

			elapsed := time.Since(start)
			gaelog.Infof(ctx, "Done computing stats; took %v", elapsed)
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
			gaelog.Errorf(ctx, "Error getting device IDs: %v", err)
			return
		}

		latest, latestErr = h.Database.Latest(ctx, ids)
		if latestErr != nil {
			gaelog.Errorf(ctx, "Error getting latest measurements: %v", latestErr)
		}

		elapsed := time.Since(start)
		gaelog.Infof(ctx, "Done getting latest measurements; took %v", elapsed)
	}()

	wg.Wait()

	// Get the full set of metrics included in the latest measurements so that the
	// latest measurements table can be constructed with only the necessary columns.
	latestMetrics := []string{}
	if latestErr == nil {
		mm := make(map[string]bool)
		for _, sm := range latest {
			for metric, _ := range sm.ValueMap() {
				mm[metric] = true
			}
		}

		for metric, _ := range mm {
			latestMetrics = append(latestMetrics, metric)
		}
		sort.Strings(latestMetrics)
	}

	// Compute PM2.5 AQIs.
	aqis := make(map[string]int)
	for _, sm := range latest {
		if sm.PM25 != nil {
			aqis[sm.DeviceID] = aqi.PM25(*sm.PM25)
		}
	}

	data := struct {
		Measurements     template.JS
		Metrics          []measurement.Metric
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
		LatestMetrics    []string
		LatestError      error
		AQI              map[string]int
	}{
		Measurements:     template.JS(jsonBytes),
		Metrics:          metrics,
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
		LatestMetrics:    latestMetrics,
		LatestError:      latestErr,
		AQI:              aqis,
	}

	if err := h.Template.ExecuteTemplate(w, "index", data); err != nil {
		gaelog.Errorf(ctx, "Could not execute template: %v", err)
	}
}

// serializableMeasurement is the same as StorableMeasurement except without fields that
// the frontend doesn't need for plotting data.
// IMPORTANT: Keep up to date with the generated Measurement type, at least to the
// extent that is required. Ensure that the JSON keys are the same as StorableMeasurement's.
type serializableMeasurement struct {
	// This timestamp is an offset from the epoch in milliseconds (compare to Timestamp in StorableMeasurement).
	Timestamp int64    `json:"ts,omitempty"`
	Temp      *float32 `json:"temp,omitempty"`
	PM25      *float32 `json:"pm25,omitempty"`
	PM10      *float32 `json:"pm10,omitempty"`
	RH        *float32 `json:"rh,omitempty"`
}

// metricsSet returns a slice containing the JSON keys of metric fields that are non-nil.
func (srm serializableMeasurement) metricsSet() []string {
	metrics := []string{}

	v := reflect.ValueOf(srm)
	for i := 0; i < v.NumField(); i++ {
		jsonKey := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]
		if jsonKey == "" {
			continue
		}

		// The field must be a float32 pointer.
		f, ok := v.Field(i).Interface().(*float32)
		if !ok || f == nil {
			continue
		}

		metrics = append(metrics, jsonKey)
	}

	return metrics
}

// measurementMapToJSON converts a string -> []StorableMeasurement map into a marshaled
// JSON array for use in the template. The JSON is an array with one element for each
// device ID. It's constructed this way, instead of as a map where keys are device IDs,
// because the JavaScript visualization package D3 (https://d3js.org/) works better with
// arrays of data than maps.
func measurementMapToJSON(measurements map[string][]measurement.StorableMeasurement) ([]byte, error) {
	type dataForTemplate struct {
		ID      string                    `json:"id"`
		Metrics []string                  `json:"metrics"`
		Values  []serializableMeasurement `json:"values"`
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
		// metricsSet will contain the set of the JSON keys of the metrics that are
		// present in the given measurements.
		metricsSet := make(map[string]bool)

		vals := make([]serializableMeasurement, len(measurements[k]))
		for i, sm := range measurements[k] {
			// IMPORTANT: Keep up to date with serializableMeasurement.
			vals[i] = serializableMeasurement{
				Timestamp: sm.Timestamp.Unix() * 1000,
				Temp:      sm.Temp,
				PM25:      sm.PM25,
				PM10:      sm.PM10,
				RH:        sm.RH,
			}

			for _, metric := range vals[i].metricsSet() {
				metricsSet[metric] = true
			}
		}

		// Turn the set into a slice of keys.
		metrics := make([]string, len(metricsSet))
		i := 0
		for k, _ := range metricsSet {
			metrics[i] = k
			i++
		}

		data = append(data, dataForTemplate{
			ID:      k,
			Metrics: metrics,
			Values:  vals,
		})
	}
	return json.Marshal(data)
}
