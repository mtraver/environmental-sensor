package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/mtraver/environmental-sensor/aqi"
	"github.com/mtraver/environmental-sensor/measurement"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/envtools"
	"github.com/mtraver/gaelog"
)

const (
	datastoreKind = "measurement"

	// If this env var is set the Go server will serve static files. If not then
	// static file serving must be achieved another way.
	serveStaticEnvVar = "SERVE_STATIC"
)

type Database interface {
	Save(ctx context.Context, m *mpb.Measurement) error
	Since(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	DelayedSince(ctx context.Context, startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	Between(ctx context.Context, startTime time.Time, endTime time.Time) (map[string][]measurement.StorableMeasurement, error)
	Latest(ctx context.Context, deviceIDs []string) (map[string]measurement.StorableMeasurement, error)
}

func main() {
	// Get the project ID from the metadata service if possible, and fall back to
	// the env var otherwise. The first check is called "OnGCE" but it will return
	// true when running on Cloud Run as well.
	var projectID string
	if metadata.OnGCE() {
		var err error
		projectID, err = metadata.ProjectID()
		if err != nil {
			log.Fatalf("Failed to get project ID from metadata service: %v", err)
		}
		log.Printf("Got project ID from metadata service: %q", projectID)
	} else {
		log.Printf("Not on Google Cloud infra, falling back to env var for project ID")
		projectID = envtools.MustGetenv("GOOGLE_CLOUD_PROJECT")
	}

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

	cache := newCache()
	database, err := db.NewDatastoreDB(projectID, datastoreKind, cache)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	influxDB := db.NewInfluxDB(envtools.MustGetenv("INFLUXDB_SERVER"), envtools.MustGetenv("INFLUXDB_TOKEN"), envtools.MustGetenv("INFLUXDB_ORG"), envtools.MustGetenv("INFLUXDB_BUCKET"))

	mux := http.NewServeMux()

	mux.Handle("/", rootHandler{
		ProjectID: projectID,
		// This environment variable should be defined in app.yaml.
		IoTCoreRegistry:   envtools.MustGetenv("IOTCORE_REGISTRY"),
		DefaultDisplayAge: 12 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	mux.Handle("/uploadz", uploadzHandler{
		DelayedUploadsDur: 48 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	mux.Handle("/cachez", cachezHandler{
		Cache:    cache,
		Template: templates,
	})

	mux.Handle("/push-handlers/telemetry", pushHandler{
		PubSubToken:    envtools.MustGetenv("PUBSUB_VERIFICATION_TOKEN"),
		PubSubAudience: envtools.MustGetenv("PUBSUB_AUDIENCE"),
		Database:       database,
		InfluxDB:       influxDB,
	})

	if envtools.IsTruthy(serveStaticEnvVar) {
		log.Printf("Serving static files because $%s is set", serveStaticEnvVar)
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	}

	serve(gaelog.Wrap(mux))
}
