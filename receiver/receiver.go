package receiver

import (
  "encoding/json"
  "fmt"
  "html/template"
  "log"
  "net/http"
  "os"
  "sync"

  "github.com/golang/protobuf/proto"

  "google.golang.org/appengine"
  gaelog "google.golang.org/appengine/log"

  "receiver/db"
  "receiver/measurement"
)

// Used to display the latest measurements
var (
  measurementsMu sync.Mutex
  measurements []*measurement.Measurement
)

const maxMeasurements = 20

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

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

func init() {
  projectID := mustGetenv("GOOGLE_CLOUD_PROJECT")

  var database db.Database = nil
  switch dbType := mustGetenv("DB_TYPE"); dbType {
    case "datastore":
      database = db.NewDatastoreDB(projectID)
    case "bigtable":
      database = db.NewBigtableDB(
          projectID, mustGetenv("BIGTABLE_INSTANCE"),
          mustGetenv("BIGTABLE_TABLE"))
    default:
      log.Fatalf("Unknown database type: %v\n", dbType)
  }

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/_ah/push-handlers/telemetry", pushHandler(database))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)

  measurementsMu.Lock()
  defer measurementsMu.Unlock()
  if err := indexTemplate.Execute(w, measurements); err != nil {
    gaelog.Errorf(ctx, "Could not execute template: %v", err)
  }
}

func pushHandler(database db.Database) func(w http.ResponseWriter,
                                            r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    ctx := appengine.NewContext(r)

    msg := &pushRequest{}
    if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
      gaelog.Criticalf(ctx, "Could not decode body: %v\n", err)
      http.Error(w, fmt.Sprintf("Could not decode body: %v", err),
                 http.StatusBadRequest)
      return
    }

    m := &measurement.Measurement{}
    err := proto.Unmarshal(msg.Message.Data, m)
    if err != nil {
      gaelog.Criticalf(ctx, "Failed to unmarshal protobuf: %v\n", err)
      http.Error(w, fmt.Sprintf("Failed to unmarshal protobuf: %v", err),
                 http.StatusBadRequest)
      return
    }

    if err := database.Save(r, m); err != nil {
      gaelog.Errorf(ctx, "Failed to save measurement: %v\n", err)
    }

    measurementsMu.Lock()
    defer measurementsMu.Unlock()

    measurements = append(measurements, m)
    if len(measurements) > maxMeasurements {
      measurements = measurements[len(measurements)-maxMeasurements:]
    }

    w.WriteHeader(http.StatusOK)
  }
}
