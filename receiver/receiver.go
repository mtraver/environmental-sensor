package receiver

import (
  "encoding/json"
  "errors"
  "fmt"
  "html/template"
  "log"
  "net/http"
  "os"
  "time"

  "github.com/golang/protobuf/proto"

  "google.golang.org/appengine"
  gaelog "google.golang.org/appengine/log"

  "receiver/db"
  "receiver/measurement"
)

// Data up to this many hours old will be plotted
const dataDisplayAgeHours = 3

// Parse and cache all templates at startup instead of loading on each request
var templates = template.Must(template.New("index.html").Funcs(
    template.FuncMap{
      "millis": func(t time.Time) int64 {
        return t.Unix() * 1000
    },
}).ParseGlob("templates/*"))

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

func mustGetenv(varName string) string {
  val := os.Getenv(varName)
  if val == "" {
    log.Fatalf("Environment variable must be set: %v\n", varName)
  }
  return val
}

func getDatabase() (db.Database, error) {
  projectID := mustGetenv("GOOGLE_CLOUD_PROJECT")

  var database db.Database = nil
  var err error = nil
  switch dbType := mustGetenv("DB_TYPE"); dbType {
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

func init() {
  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/_ah/push-handlers/telemetry", pushHandler)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)

  database, err := getDatabase()
  if err != nil {
    gaelog.Criticalf(ctx, "%v", err)
    http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
    return
  }

  now := time.Now().UTC()
  startTime := now.Add(-time.Duration(dataDisplayAgeHours) * time.Hour)

  measurements, err := database.GetMeasurementsSince(ctx, startTime)
  if err != nil {
    gaelog.Errorf(ctx, "Error fetching data: %v", err)
  }

  data := struct {
    Measurements map[string][]measurement.StorableMeasurement
    Error error
    StartTime time.Time
    EndTime time.Time
  }{
    measurements,
    err,
    startTime,
    now,
  }

  if err := templates.ExecuteTemplate(w, "index", data); err != nil {
    gaelog.Errorf(ctx, "Could not execute template: %v", err)
  }
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)

  database, err := getDatabase()
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

  if err := database.Save(ctx, m); err != nil {
    gaelog.Errorf(ctx, "Failed to save measurement: %v\n", err)
  }

  w.WriteHeader(http.StatusOK)
}
