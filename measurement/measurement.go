package measurement

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
)

// Used for separating substrings in database keys. The octothorpe is fine for this because
// device IDs and timestamps, the two things most likely to be used in keys, can't contain it.
const keySep = "#"

// StorableMeasurement is equivalent to the generated Measurement type but it contains
// no protobuf-specific types. It therefore can be marshaled to JSON and written to
// Datastore.
// IMPORTANT: Keep up to date with the generated Measurement type
type StorableMeasurement struct {
	DeviceID        string    `json:"device_id,omitempty" datastore:"device_id"`
	Timestamp       time.Time `json:"timestamp,omitempty" datastore:"timestamp"`
	UploadTimestamp time.Time `json:"upload_timestamp,omitempty" datastore:"upload_timestamp,omitempty"`
	Temp            float32   `json:"temp,omitempty" datastore:"temp"`
}

// NewStorableMeasurement converts the generated Measurement type to a StorableMeasurement,
// which contains no protobuf-specific types, and therefore can be marshaled to JSON and
// written to Datastore.
// IMPORTANT: Keep up to date with the generated Measurement type
func NewStorableMeasurement(m *Measurement) (StorableMeasurement, error) {
	timestamp, err := ptypes.Timestamp(m.GetTimestamp())
	if err != nil {
		return StorableMeasurement{}, err
	}

	// The generated protobuf code uses a pointer to ptypes' timestamp.Timestamp, but in
	// StorableMeasurement we use golang's time.Time. If the protobuf field is nil then
	// use the zero value of time.Time. cloud.google.com/go/datastore calls IsZero() on
	// time.Time values so omitempty does work.
	var uploadTimestamp time.Time
	pbUploadTimestamp := m.GetUploadTimestamp()
	if pbUploadTimestamp != nil {
		if uploadTimestamp, err = ptypes.Timestamp(pbUploadTimestamp); err != nil {
			return StorableMeasurement{}, err
		}
	}

	return StorableMeasurement{
		DeviceID:        m.GetDeviceId(),
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            m.GetTemp(),
	}, nil
}

// DBKey returns a string key suitable for Datastore. It promotes Device ID and timestamp into the key.
func (m *StorableMeasurement) DBKey() string {
	return strings.Join([]string{m.DeviceID, m.Timestamp.Format(time.RFC3339)}, keySep)
}

func (m StorableMeasurement) String() string {
	delay := ""
	if !m.UploadTimestamp.IsZero() {
		delay = fmt.Sprintf(" (%v upload delay)", m.UploadTimestamp.Sub(m.Timestamp))
	}

	return fmt.Sprintf("%s %.3f°C %s%s", m.DeviceID, m.Temp, m.Timestamp.Format(time.RFC3339), delay)
}
