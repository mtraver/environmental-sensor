package db

import (
  "time"

  "golang.org/x/net/context"

  "receiver/measurement"
)

// Interface that must be implemented for each database backend
type Database interface {
  // Save saves the given Measurement to the database. If the Measurement
  // already exists in the database it should make no change to the database
  // and return nil as the error.
  Save(ctx context.Context, m *measurement.Measurement) error

  // GetMeasurementsSince gets all measurements with a timestamp greater than
  // or equal to the given time. It returns a map of device ID (a string) to
  // a StorableMeasurement slice, and an error.
  GetMeasurementsSince(
      ctx context.Context,
      startTime time.Time) (map[string][]measurement.StorableMeasurement, error)
}
