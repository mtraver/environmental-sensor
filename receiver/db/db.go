package db

import (
  "golang.org/x/net/context"

  "receiver/measurement"
)

// Interface that must be implemented for each database backend
type Database interface {
  Save(ctx context.Context, m *measurement.Measurement) error
}
