package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mtraver/gaelog"
	"google.golang.org/appengine"

	"github.com/mtraver/environmental-sensor/measurement"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/environmental-sensor/web/device"
)

// Data up to this many hours old will be plotted
const defaultDataDisplayAgeHours = 12

var (
	projectID = mustGetenv("GOOGLE_CLOUD_PROJECT")

	// This environment variable should be defined in app.yaml.
	iotcoreRegistry = mustGetenv("IOTCORE_REGISTRY")

	// Parse and cache all templates at startup instead of loading on each request.
	// The path to the templates is relative to go.mod, as that's how the path should
	// be specified when deployed to App Engine.
	templates = template.Must(template.New("index.html").Funcs(
		template.FuncMap{
			"millis": func(t time.Time) int64 {
				return t.Unix() * 1000
			},
			"RFC3339": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
		}).ParseGlob("web/templates/*"))

	database Database
)

type Database interface {
	Save(ctx context.Context, m *measurement.Measurement) error
	GetMeasurementsSince(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	GetDelayedMeasurementsSince(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	GetMeasurementsBetween(ctx context.Context, startTime time.Time, endTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	GetLatestMeasurements(ctx context.Context, deviceIDs []string) (map[string]measurement.StorableMeasurement, error)
}

func mustGetenv(varName string) string {
	val := os.Getenv(varName)
	if val == "" {
		log.Fatalf("Environment variable must be set: %v\n", varName)
	}
	return val
}

func init() {
	var err error
	database, err = db.NewDatastoreDB(projectID)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/uploadz", uploadzHandler)
	http.HandleFunc("/_ah/push-handlers/telemetry", pushHandler)

	appengine.Main()
}

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
			startTime, rangeErr = time.Parse(time.RFC3339Nano,
				r.FormValue("startdate-adjusted"))
			if rangeErr != nil {
				http.Error(w, fmt.Sprintf("Bad start time: %v", rangeErr),
					http.StatusBadRequest)
				return
			}

			endTime, rangeErr = time.Parse(time.RFC3339Nano,
				r.FormValue("enddate-adjusted"))
			if rangeErr != nil {
				http.Error(w, fmt.Sprintf("Bad end time: %v", rangeErr),
					http.StatusBadRequest)
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
				http.Error(w, fmt.Sprintf("Hours ago must be >= 1"),
					http.StatusBadRequest)
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
		jsonBytes, err = measurement.MeasurementMapToJSON(measurements)
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
