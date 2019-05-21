package measurement

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

// Used for separating substrings in database keys. The octothorpe is fine for this because
// device IDs and timestamps, the two things most likely to be used in keys, can't contain it.
const keySep = "#"

// ErrZeroTimestamp is returned from NewMeasurement if the StorableMeasurement's timestamp is the zero timestamp.
var ErrZeroTimestamp = fmt.Errorf("measurement: timestamp cannot be nil")

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
func NewStorableMeasurement(m *mpb.Measurement) (StorableMeasurement, error) {
	// This will return an error if the timestamp is nil, which is good, because
	// we want to enforce non-nil timestamps.
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

// NewMeasurement converts a StorableMeasurement into the generated Measurement type,
// converting time.Time values into the protobuf-specific timestamp type.
// IMPORTANT: Keep up to date with the generated Measurement type
func NewMeasurement(m *StorableMeasurement) (mpb.Measurement, error) {
	// Enforce a non-zero timestamp.
	if m.Timestamp.IsZero() {
		return mpb.Measurement{}, ErrZeroTimestamp
	}

	timestamp, err := ptypes.TimestampProto(m.Timestamp)
	if err != nil {
		return mpb.Measurement{}, err
	}

	// The upload timestamp may be the zero timestamp. If it is, then the upload timestamp
	// should be nil in the generated Measurement type.
	var uploadTimestamp *tspb.Timestamp
	if !m.UploadTimestamp.IsZero() {
		if uploadTimestamp, err = ptypes.TimestampProto(m.UploadTimestamp); err != nil {
			return mpb.Measurement{}, err
		}
	}

	return mpb.Measurement{
		DeviceId:        m.DeviceID,
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            m.Temp,
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
