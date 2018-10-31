package db

import (
	"time"

	"golang.org/x/net/context"

	"github.com/mtraver/environmental-sensor/receiver/measurement"
)

// Interface that must be implemented for each database backend
type Database interface {
	// Save saves the given Measurement to the database. If the Measurement
	// already exists in the database it should make no change to the database
	// and return nil as the error.
	Save(ctx context.Context, m *measurement.Measurement) error

	// GetMeasurementsSince gets all measurements with a timestamp greater than
	// or equal to startTime. It returns a map of device ID (a string) to a
	// StorableMeasurement slice, and an error.
	GetMeasurementsSince(
		ctx context.Context,
		startTime time.Time) (map[string][]measurement.StorableMeasurement, error)

	// GetMeasurementsBetween gets all measurements with a timestamp greater than
	// or equal to startTime and less than or equal to endTime. It returns a map
	// of device ID (a string) to a StorableMeasurement slice, and an error.
	GetMeasurementsBetween(
		ctx context.Context, startTime time.Time,
		endTime time.Time) (map[string][]measurement.StorableMeasurement, error)

	// GetLatestMeasurements gets the most recent measurement for each of the given
	// device IDs. It returns a map of device ID to StorableMeasurement, and an
	// error. If no measurement is found for a device ID then the returned map will
	// not contain that device ID.
	GetLatestMeasurements(ctx context.Context, deviceIDs []string) (
		map[string]measurement.StorableMeasurement, error)
}
