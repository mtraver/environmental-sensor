package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mtraver/environmental-sensor/aqi"
	"github.com/mtraver/environmental-sensor/measurement"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/gaelog"
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
			"PrintfPtr": func(format string, f *float32) string {
				if f == nil {
					return "null"
				}
				return fmt.Sprintf(format, *f)
			},
			"AQIStr": func(v float32) string {
				return aqi.String(int(v))
			},
			"AQIAbbrv": func(v float32) string {
				return aqi.Abbrv(int(v))
			},
		}).ParseGlob("web/templates/*"))

	database, err := db.NewDatastoreDB(projectID, datastoreKind, newCache())
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", rootHandler{
		ProjectID: projectID,
		// This environment variable should be defined in app.yaml.
		IoTCoreRegistry:   mustGetenv("IOTCORE_REGISTRY"),
		DefaultDisplayAge: 12 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	mux.Handle("/uploadz", uploadzHandler{
		DelayedUploadsDur: 48 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	mux.Handle("/_ah/push-handlers/telemetry", pushHandler{
		Database: database,
	})

	serve(gaelog.Wrap(mux))
}
