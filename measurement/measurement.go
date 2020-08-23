package measurement

import (
	"fmt"
	"strings"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
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
	Temp            *float32  `json:"temp,omitempty" datastore:"temp"`
}

// NewStorableMeasurement converts the generated Measurement type to a StorableMeasurement,
// which contains no protobuf-specific types, and therefore can be marshaled to JSON and
// written to Datastore.
// IMPORTANT: Keep up to date with the generated Measurement type
func NewStorableMeasurement(m *mpb.Measurement) (StorableMeasurement, error) {
	// This will return an error if the timestamp is nil, which is good, because
	// we want to enforce non-nil timestamps.
	if m.GetTimestamp() == nil {
		return StorableMeasurement{}, fmt.Errorf("measurement: nil timestamp")
	}
	if err := m.GetTimestamp().CheckValid(); err != nil {
		return StorableMeasurement{}, err
	}
	timestamp := m.GetTimestamp().AsTime()

	// The generated protobuf code uses a pointer to tspb.Timestamp, but in StorableMeasurement
	// we use golang's time.Time. If the protobuf field is nil then use the zero value of time.Time.
	// cloud.google.com/go/datastore calls IsZero() on time.Time values so omitempty does work.
	var uploadTimestamp time.Time
	pbUploadTimestamp := m.GetUploadTimestamp()
	if pbUploadTimestamp != nil {
		if err := pbUploadTimestamp.CheckValid(); err != nil {
			return StorableMeasurement{}, err
		}
		uploadTimestamp = pbUploadTimestamp.AsTime()
	}

	var temp *float32
	if m.GetTemp() != nil {
		t := m.GetTemp().GetValue()
		temp = &t
	}

	return StorableMeasurement{
		DeviceID:        m.GetDeviceId(),
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            temp,
	}, nil
}

// NewMeasurement converts a StorableMeasurement into the generated Measurement type,
// converting time.Time values into the protobuf-specific timestamp type.
// IMPORTANT: Keep up to date with the generated Measurement type
func NewMeasurement(sm *StorableMeasurement) (mpb.Measurement, error) {
	// Enforce a non-zero timestamp.
	if sm.Timestamp.IsZero() {
		return mpb.Measurement{}, ErrZeroTimestamp
	}

	timestamp := tspb.New(sm.Timestamp)

	// The upload timestamp may be the zero timestamp. If it is, then the upload timestamp
	// should be nil in the generated Measurement type.
	var uploadTimestamp *tspb.Timestamp
	if !sm.UploadTimestamp.IsZero() {
		uploadTimestamp = tspb.New(sm.UploadTimestamp)
	}

	var temp *wpb.FloatValue
	if sm.Temp != nil {
		temp = wpb.Float(*sm.Temp)
	}

	return mpb.Measurement{
		DeviceId:        sm.DeviceID,
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            temp,
	}, nil
}

// DBKey returns a string key suitable for Datastore. It promotes Device ID and timestamp into the key.
func (sm *StorableMeasurement) DBKey() string {
	return strings.Join([]string{sm.DeviceID, sm.Timestamp.Format(time.RFC3339)}, keySep)
}

func (sm StorableMeasurement) String() string {
	delay := ""
	if !sm.UploadTimestamp.IsZero() {
		delay = fmt.Sprintf(" (%v upload delay)", sm.UploadTimestamp.Sub(sm.Timestamp))
	}

	tStr := "[unknown]"
	if sm.Temp != nil {
		tStr = fmt.Sprintf("%.3f°C", *sm.Temp)
	}
	return fmt.Sprintf("%s %s %s%s", sm.DeviceID, tStr, sm.Timestamp.Format(time.RFC3339), delay)
}
