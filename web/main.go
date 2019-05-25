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
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/web/db"
)

const (
	datastoreKind = "measurement"
)

type Database interface {
	Save(ctx context.Context, m *mpb.Measurement) error
	Since(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	DelayedSince(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	Between(ctx context.Context, startTime time.Time, endTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	Latest(ctx context.Context, deviceIDs []string) (map[string]measurement.StorableMeasurement, error)
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

	// The path to the templates is relative to go.mod, as that's how the path should
	// be specified when deployed to App Engine.
	templates := template.Must(template.New("index.html").Funcs(
		template.FuncMap{
			"millis": func(t time.Time) int64 {
				return t.Unix() * 1000
			},
			"RFC3339": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
		}).ParseGlob("web/templates/*"))

	database, err := db.NewDatastoreDB(projectID, datastoreKind, true)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	http.Handle("/", rootHandler{
		ProjectID: projectID,
		// This environment variable should be defined in app.yaml.
		IoTCoreRegistry:   mustGetenv("IOTCORE_REGISTRY"),
		DefaultDisplayAge: 12 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	http.Handle("/uploadz", uploadzHandler{
		DelayedUploadsDur: 48 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	http.Handle("/_ah/push-handlers/telemetry", pushHandler{
		Database: database,
	})

	appengine.Main()
}
