package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"

	"google.golang.org/appengine"
	gaelog "google.golang.org/appengine/log"

	"receiver/aeutil"
	"receiver/db"
	"receiver/device"
	"receiver/measurement"
)

// Data up to this many hours old will be plotted
const defaultDataDisplayAgeHours = 6

var (
	// These environment variables should be defined in app.yaml.
	dbType          = mustGetenv("DB_TYPE")
	iotcoreRegistry = mustGetenv("IOTCORE_REGISTRY")

	// Parse and cache all templates at startup instead of loading on each request
	templates = template.Must(template.New("index.html").Funcs(
		template.FuncMap{
			"millis": func(t time.Time) int64 {
				return t.Unix() * 1000
			},
			"timeAgoString": timeAgoString,
		}).ParseGlob("templates/*"))
)

func mustGetenv(varName string) string {
	val := os.Getenv(varName)
	if val == "" {
		log.Fatalf("Environment variable must be set: %v\n", varName)
	}
	return val
}

// No round function in the std lib before go1.10
func round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

func divmod(a, b int64) (int64, int64) {
	return a / b, a % b
}

// timeAgoString turns a time into a friendly string like "just now" or "10 min ago".
func timeAgoString(t time.Time) string {
	d := time.Now().UTC().Sub(t)

	if d < time.Second*5 {
		return "just now"
	}

	if d < time.Second*60 {
		return fmt.Sprintf("%d s ago", int(round(d.Seconds(), 5)))
	}

	if d < time.Hour {
		return fmt.Sprintf("%d min ago", int(round(d.Minutes(), 1)))
	}

	if d < time.Hour*24 {
		h, m := divmod(int64(d.Minutes()), 60)
		if m == 0 {
			return fmt.Sprintf("%d hr ago", h)
		}

		return fmt.Sprintf("%d hr %d min ago", h, m)
	}

	return "> 24 hr ago"
}

// This is the structure of the JSON payload pushed to the endpoint by
// Cloud Pub/Sub. See https://cloud.google.com/pubsub/docs/push.
type pushRequest struct {
	Message struct {
		Attributes map[string]string
		Data       []byte
		ID         string `json:"message_id"`
	}
	Subscription string
}

func getDatabase(ctx context.Context) (db.Database, error) {
	projectID := aeutil.GetProjectID(ctx)

	var database db.Database = nil
	var err error = nil
	switch dbType {
	case "datastore":
		database = db.NewDatastoreDB(projectID)
	case "bigtable":
		database = db.NewBigtableDB(
			projectID, mustGetenv("BIGTABLE_INSTANCE"),
			mustGetenv("BIGTABLE_TABLE"))
	default:
		err = errors.New(fmt.Sprintf("Unknown database type: %v", dbType))
	}

	return database, err
}

func main() {
	http.HandleFunc("/", rootHandler)
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

	database, err := getDatabase(ctx)
	if err != nil {
		gaelog.Criticalf(ctx, "%v", err)
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

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
			startTime, err = time.Parse(time.RFC3339Nano,
				r.FormValue("startdate-adjusted"))
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad start time: %v", err),
					http.StatusBadRequest)
				return
			}

			endTime, err = time.Parse(time.RFC3339Nano,
				r.FormValue("enddate-adjusted"))
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad end time: %v", err),
					http.StatusBadRequest)
				return
			}

			fillRangeForm = true
			fillHoursAgoForm = false
		case "hoursago":
			hoursAgo, err = strconv.Atoi(r.FormValue("hoursago"))
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
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
		gaelog.Errorf(ctx, "Error fetching data: %v", err)
	} else {
		jsonBytes, err = measurement.MeasurementMapToJSON(measurements)
		if err != nil {
			gaelog.Errorf(ctx, "Error marshaling measurements to JSON: %v", err)
		}
	}

	// Get the latest measurement for each device
	var latest map[string]measurement.StorableMeasurement
	ids, latestErr := device.GetDeviceIDs(ctx, aeutil.GetProjectID(ctx), iotcoreRegistry)
	if latestErr != nil {
		gaelog.Errorf(ctx, "Error getting device IDs: %v", latestErr)
	} else {
		latest, latestErr = database.GetLatestMeasurements(ctx, ids)

		if latestErr != nil {
			gaelog.Errorf(ctx, "Error getting latest measurements: %v", latestErr)
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
		gaelog.Errorf(ctx, "Could not execute template: %v", err)
	}
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	database, err := getDatabase(ctx)
	if err != nil {
		gaelog.Criticalf(ctx, "%v", err)
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	msg := &pushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		gaelog.Criticalf(ctx, "Could not decode body: %v\n", err)
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err),
			http.StatusBadRequest)
		return
	}

	m := &measurement.Measurement{}
	err = proto.Unmarshal(msg.Message.Data, m)
	if err != nil {
		gaelog.Criticalf(ctx, "Failed to unmarshal protobuf: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to unmarshal protobuf: %v", err),
			http.StatusBadRequest)
		return
	}

	if err := m.Validate(); err != nil {
		gaelog.Errorf(ctx, "%v", err)

		// Pub/Sub will only stop re-trying the message if it receives a status 200.
		// The docs say that any of 200, 201, 202, 204, or 102 will have this effect
		// (https://cloud.google.com/pubsub/docs/push), but the local emulator
		// doesn't respect anything other than 200, so return 200 just to be safe.
		// TODO(mtraver) I'd rather return e.g. 202 (http.StatusAccepted) to
		// indicate that it was successfully received but not that all is ok.
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := database.Save(ctx, m); err != nil {
		gaelog.Errorf(ctx, "Failed to save measurement: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
}
