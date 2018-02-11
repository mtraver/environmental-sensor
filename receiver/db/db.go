package db

import (
  "net/http"

  "receiver/measurement"
)

// Interface that must be implemented for each database backend
type Database interface {
  Save(req *http.Request, m *measurement.Measurement) error
}
