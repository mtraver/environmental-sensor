package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"

	"github.com/mtraver/environmental-sensor/measurement"
	"github.com/mtraver/environmental-sensor/web/db"
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
