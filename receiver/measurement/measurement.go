package measurement

import (
  "encoding/json"
  "sort"
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

// IMPORTANT: Keep up to date with the automatically-generated Measurement type.
// This is almost the same as StorableMeasurement except without DeviceId. When
// data is marshaled to JSON for use in the template each record doesn't need to
// include the device ID.
type serializableMeasurement struct {
  // This timestamp is an offset from the epoch in millieconds
  // (compare to Timestamp in StorableMeasurement).
  Timestamp int64    `json:"timestamp,omitempty" datastore:"timestamp"`
  Temp      float32  `json:"temp,omitempty" datastore:"temp"`
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

// MeasurementMapToJSON converts a string -> []StorableMeasurement map
// into a marshaled JSON array for use in the template. The JSON is an array
// with one element for each device ID. It's constructed this way, instead of
// as a map where keys are device IDs, because the JavaScript visualization
// package D3 (https://d3js.org/) works better with arrays of data than maps.
func MeasurementMapToJSON(measurements map[string][]StorableMeasurement) (
    []byte, error) {
  type dataForTemplate struct {
    ID string `json:"id"`
    Values []serializableMeasurement `json:"values"`
  }

  // Sort the map's keys so that the resulting JSON always has them in the same
  // order. This ensures that e.g. the color assigned to each line on a plot is
  // the same for every page load.
  keys := make([]string, len(measurements))
  i := 0
  for k := range measurements {
    keys[i] = k
    i++
  }
  sort.Strings(keys)

  var data []dataForTemplate
  for _, k := range keys {
    vals := make([]serializableMeasurement, len(measurements[k]))
    for i, m := range measurements[k] {
      vals[i] = serializableMeasurement{m.Timestamp.Unix() * 1000, m.Temp}
    }
    data = append(data, dataForTemplate{k, vals})
  }
  return json.Marshal(data)
}
