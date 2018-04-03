package db

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/bigtable"

	"receiver/measurement"
)

const bigtableFamily = "measurement"

func floatToBytes(f float32) []byte {
	buf := new(bytes.Buffer)

	// TODO(mtraver) handle error
	binary.Write(buf, binary.LittleEndian, f)

	return buf.Bytes()
}

func bytesToFloat(b []byte) float32 {
	var ret float32
	buf := bytes.NewReader(b)

	// TODO(mtraver) handle error
	binary.Read(buf, binary.LittleEndian, &ret)

	return ret
}

type bigtableDB struct {
	projectID string
	instanceName string
	tableName string
}

// Ensure bigtableDB implements Database
var _ Database = &bigtableDB{}

func NewBigtableDB(projectID string, instanceName string,
									 tableName string) Database {
	return &bigtableDB{
		projectID: projectID,
		instanceName: instanceName,
		tableName: tableName,
	}
}

// Save saves a Measurement to Bigtable.
// The measurement's timestamp will be formatted as RFC 3339
// and promoted into the row key along with the device ID.
// It returns an error, nil if nothing went wrong.
func (db *bigtableDB) Save(ctx context.Context,
													 m *measurement.Measurement) error {
	client, err := bigtable.NewClient(ctx, db.projectID, db.instanceName)
	if err != nil {
		return err
	}

	table := client.Open(db.tableName)

	storableMeasurement, err := m.ToStorableMeasurement()
	if err != nil {
		return err
	}

	// Device ID and timestamp are promoted into the row key
	// https://cloud.google.com/bigtable/docs/schema-design-time-series
	rowKey := storableMeasurement.DBKey()

	// Check if the row exists and return if it does
	row, err := table.ReadRow(ctx, rowKey)
	if err != nil {
		return err
	}
	if len(row) != 0 {
		return nil
	}

	mut := bigtable.NewMutation()
	mut.Set(bigtableFamily, "temp", bigtable.Now(),
					floatToBytes(storableMeasurement.Temp))

	return table.Apply(ctx, rowKey, mut)
}

// TODO(mtraver) implement
func (db *bigtableDB) GetMeasurementsSince(
		ctx context.Context,
		startTime time.Time) (map[string][]measurement.StorableMeasurement, error) {
	return make(map[string][]measurement.StorableMeasurement),
			errors.New("Not implemented")
}

// TODO(mtraver) implement
func (db *bigtableDB) GetMeasurementsBetween(
		ctx context.Context, startTime time.Time,
		endTime time.Time) (map[string][]measurement.StorableMeasurement, error) {
	return make(map[string][]measurement.StorableMeasurement),
			errors.New("Not implemented")
}
