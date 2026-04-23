package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/mtraver/environmental-sensor/graph"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/envtools"
	"github.com/mtraver/gaelog"
)

const (
	datastoreKind = "measurement"

	// A comma-separated string of device IDs to ignore. Devices pushing data will be checked
	// for whether their IDs contain any of the strings specified in this env var.
	ignoredDevicesEnvVar = "IGNORED_DEVICES"

	// A comma-separated string of Pub/Sub "source" attributes to ignore. Incoming Pub/Sub
	// messages will be checked for a "source" attribute matching any of the strings specified
	// in this env var.
	ignoredSourcesEnvVar = "IGNORED_SOURCES"

	// awsRoleARNEnvVar is the name of the env var that should contain the ARN of the
	// AWS role that we'll assume and use to authenticate with AWS IoT to fetch the
	// list of devices.
	awsRoleARNEnvVar = "AWS_ROLE_ARN"

	awsRegionEnvVar = "AWS_REGION"

	// debugServeClientEnvVar controls whether the client is served from the Go web server
	// along with the backend. This is used for local development.
	debugServeClientEnvVar = "DEBUG_SERVE_CLIENT"

	// debugGraphQLPlaygroundEnvVar controls whether we serve the GraphQL playground.
	debugGraphQLPlaygroundEnvVar = "DEBUG_GQL_PLAYGROUND"
	graphQLPlaygroundURL         = "/debug/graphql"
)

func filter[T any](s []T, test func(T) bool) []T {
	ret := []T{}
	for _, e := range s {
		if test(e) {
			ret = append(ret, e)
		}
	}
	return ret
}

func main() {
	// Get the project ID from the metadata service if possible, and fall back to
	// the env var otherwise. The first check is called "OnGCE" but it will return
	// true when running on Cloud Run as well.
	onGCE := metadata.OnGCE()
	var projectID string
	if onGCE {
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

	roleARN := os.Getenv(awsRoleARNEnvVar)
	if roleARN == "" && onGCE {
		log.Printf("On GCE and $%s is not set. Fetching devices will probably fail.", awsRoleARNEnvVar)
	}

	// The path to the templates is relative to go.mod, as that's how they are placed in the Docker image.
	templates := template.Must(template.New("index.html").Option("missingkey=error").ParseGlob("web/templates/*"))

	cache := newCache()
	database, err := db.NewDatastoreDB(projectID, datastoreKind, cache)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	influxDB := db.NewInfluxDB(envtools.MustGetenv("INFLUXDB_SERVER"), envtools.MustGetenv("INFLUXDB_TOKEN"), envtools.MustGetenv("INFLUXDB_ORG"), envtools.MustGetenv("INFLUXDB_BUCKET"))

	mux := http.NewServeMux()

	if envtools.IsTruthy(debugServeClientEnvVar) {
		log.Printf("Serving client because %s is set", debugServeClientEnvVar)

		// Disable caching so that clients don't try to use an old version
		// which will request old resources that may or may not still exist.
		mux.Handle("/", noCache(http.FileServer(http.Dir("client/build/client/"))))
	} else {
		mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("content-type", "text/plain")
			fmt.Fprintln(w, "OK")
		})
	}

	gqlHandler := graphQLHandler(&graph.Resolver{
		Database:   database,
		AWSRegion:  envtools.MustGetenv(awsRegionEnvVar),
		AWSRoleARN: roleARN,
	})
	mux.Handle("/query", gqlHandler)
	if envtools.IsTruthy(debugGraphQLPlaygroundEnvVar) {
		log.Printf("Serving GraphQL playground at %s because %s is set", graphQLPlaygroundURL, debugGraphQLPlaygroundEnvVar)
		mux.Handle(graphQLPlaygroundURL, playground.Handler("GraphQL playground", "/query"))
	}

	mux.Handle("/debug/uploadz", uploadzHandler{
		DelayedUploadsDur: 48 * time.Hour,
		Database:          database,
		Template:          templates,
	})

	mux.Handle("/debug/cachez", cachezHandler{
		Cache:    cache,
		Template: templates,
	})

	ignoredDevices := strings.Split(os.Getenv(ignoredDevicesEnvVar), ",")
	ignoredDevices = filter(ignoredDevices, func(s string) bool { return s != "" })
	if len(ignoredDevices) > 0 {
		log.Printf("%s is set to %q. Will ignore devices with IDs containing these strings: %v",
			ignoredDevicesEnvVar, os.Getenv(ignoredDevicesEnvVar), ignoredDevices)
	}

	ignoredSources := strings.Split(os.Getenv(ignoredSourcesEnvVar), ",")
	ignoredSources = filter(ignoredSources, func(s string) bool { return s != "" })
	if len(ignoredSources) > 0 {
		log.Printf("%s is set to %q. Will ignore messages with \"source\" attribute equal to any of these strings: %v",
			ignoredSourcesEnvVar, os.Getenv(ignoredSourcesEnvVar), ignoredSources)
	}

	mux.Handle("/push-handlers/telemetry", pushHandler{
		PubSubToken:    envtools.MustGetenv("PUBSUB_VERIFICATION_TOKEN"),
		PubSubAudience: envtools.MustGetenv("PUBSUB_AUDIENCE"),
		Database:       database,
		InfluxDB:       influxDB,
		IgnoredDevices: ignoredDevices,
		IgnoredSources: ignoredSources,
	})

	serve(gaelog.Wrap(mux))
}

func noCache(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	}
}
