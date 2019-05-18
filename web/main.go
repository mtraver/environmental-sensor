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

func main() {
	projectID := mustGetenv("GOOGLE_CLOUD_PROJECT")

	database, err := db.NewDatastoreDB(projectID)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	http.Handle("/", RootHandler{
		ProjectID: projectID,
		// This environment variable should be defined in app.yaml.
		IoTCoreRegistry: mustGetenv("IOTCORE_REGISTRY"),
		Database:        database,
	})

	http.Handle("/uploadz", UploadzHandler{
		DelayedUploadsDur: 48 * time.Hour,
		Database:          database,
	})

	http.Handle("/_ah/push-handlers/telemetry", PushHandler{
		Database: database,
	})

	appengine.Main()
}
