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
  "strings"
  "time"

  "github.com/golang/protobuf/proto"

  "google.golang.org/appengine"
  gaelog "google.golang.org/appengine/log"

  "receiver/db"
  "receiver/measurement"
)

// Data up to this many hours old will be plotted
const defaultDataDisplayAgeHours = 6

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

func getDatabase(ctx context.Context) (db.Database, error) {
  // From the documentation of appengine.AppID:
  //
  //   AppID returns the application ID for the current application. The string
  //   will be a plain application ID (e.g. "appid"), with a domain prefix for
  //   custom domain deployments (e.g. "example.com:appid").
  //
  // Here we just want the app ID (don't care if it's deployed to a custom
  // domain) so split at the first colon. This is fine because an app ID can
  // only have lowercase letters, digits, and hyphens.
  appIDParts := strings.Split(appengine.AppID(ctx), ":")
  projectID := appIDParts[len(appIDParts)-1]

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

      if (hoursAgo < 1) {
        http.Error(w, fmt.Sprintf("Hours ago must be >= 1"),
                   http.StatusBadRequest)
        return
      }

      endTime = time.Now().UTC()
      startTime = endTime.Add(-time.Duration(hoursAgo) * time.Hour)

      fillRangeForm = false
      fillHoursAgoForm = true
    default:
      http.Error(
          w, fmt.Sprintf("Unknown form name"), http.StatusBadRequest)
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

  data := struct {
    Measurements template.JS
    Error error
    StartTime time.Time
    EndTime time.Time
    HoursAgo int
    FillRangeForm bool
    FillHoursAgoForm bool
  }{
    Measurements: template.JS(jsonBytes),
    Error: err,
    StartTime: startTime,
    EndTime: endTime,
    HoursAgo: hoursAgo,
    FillRangeForm: fillRangeForm,
    FillHoursAgoForm: fillHoursAgoForm,
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

  if err := database.Save(ctx, m); err != nil {
    gaelog.Errorf(ctx, "Failed to save measurement: %v\n", err)
  }

  w.WriteHeader(http.StatusOK)
}
