package measurement

import (
  "strings"
  "time"

  "github.com/golang/protobuf/ptypes"
)

const keySep = "#"

// IMPORTANT: Keep up to date with the automatically-generated Measurement type
type StorableMeasurement struct {
  DeviceId  string     `json:"device_id,omitempty" datastore:"device_id"`
  Timestamp time.Time  `json:"timestamp,omitempty" datastore:"timestamp"`
  Temp      float32    `json:"temp,omitempty" datastore:"temp"`
}

// ToStorableMeasurement converts the automatically-generated Measurement
// type to a type that contains non protobuf-specific types.
// IMPORTANT: Keep up to date with the automatically-generated Measurement type
func (m *Measurement) ToStorableMeasurement() (StorableMeasurement, error) {
  timestamp, err := ptypes.Timestamp(m.GetTimestamp())
  if err != nil {
    return StorableMeasurement{}, nil
  }

  return StorableMeasurement{
    DeviceId: m.GetDeviceId(),
    Timestamp: timestamp,
    Temp: m.GetTemp(),
  }, nil
}

// DBKey returns a string key suitable for Bigtable and Datastore.
// It promotes Device ID and timestamp into the key, which is especially
// useful for Bigtable. For more info see
// https://cloud.google.com/bigtable/docs/schema-design-time-series.
func (m *StorableMeasurement) DBKey() string {
  return strings.Join([]string{m.DeviceId, m.Timestamp.Format(time.RFC3339)},
                      keySep)
}
