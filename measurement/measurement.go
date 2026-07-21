package measurement

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mtraver/environmental-sensor/aqi"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/metric"
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
// Datastore. Timestamp is handled specially in MarshalJSON.
// IMPORTANT: Keep up to date with the generated Measurement type (from measurement.proto).
type StorableMeasurement struct {
	DeviceID        string    `json:"-" datastore:"device_id"`
	Timestamp       time.Time `json:"-" datastore:"timestamp"`
	UploadTimestamp time.Time `json:"-" datastore:"upload_timestamp,omitempty"`

	// These metrics are the raw values reported by sensors. They must match the
	// metrics defined in the generated Measurement type (from measurement.proto).
	Temp     *float32 `json:"temp,omitempty" datastore:"temp,omitempty"`
	PM1      *float32 `json:"pm1,omitempty" datastore:"pm1,omitempty"`
	PM25     *float32 `json:"pm25,omitempty" datastore:"pm25,omitempty"`
	PM4      *float32 `json:"pm4,omitempty" datastore:"pm4,omitempty"`
	PM10     *float32 `json:"pm10,omitempty" datastore:"pm10,omitempty"`
	RH       *float32 `json:"rh,omitempty" datastore:"rh,omitempty"`
	VOCIndex *float32 `json:"voc_index,omitempty" datastore:"voc_index,omitempty"`
	NOxIndex *float32 `json:"nox_index,omitempty" datastore:"nox_index,omitempty"`
	HCHO     *float32 `json:"hcho,omitempty" datastore:"hcho,omitempty"`
	CO2      *float32 `json:"co2,omitempty" datastore:"co2,omitempty"`

	// These metrics are derived from the raw values. They're not stored in the database
	// (the `datastore` tag is set to "-") but they are passed to the frontend.
	// These values are populated by the FillDerivedMetrics method.
	AQI *float32 `json:"aqi,omitempty" datastore:"-"`
}

func (sm StorableMeasurement) MarshalJSON() ([]byte, error) {
	// Alias the type so that we don't infinitely recurse.
	type alias StorableMeasurement

	return json.Marshal(&struct {
		alias
		Ts int64 `json:"ts,omitempty"`
	}{
		alias: (alias)(sm),
		// Convert the original timestamp to an offset from the epoch in milliseconds.
		Ts: sm.Timestamp.Unix() * 1000,
	})
}

// NewStorableMeasurement converts the generated Measurement type to a StorableMeasurement,
// which contains no protobuf-specific types, and therefore can be marshaled to JSON and
// written to Datastore.
// IMPORTANT: Keep up to date with the generated Measurement type (from measurement.proto).
func NewStorableMeasurement(m *mpb.Measurement) (StorableMeasurement, error) {
	// This will return an error if the timestamp is nil, which is good, because
	// we want to enforce non-nil timestamps.
	if m.GetTimestamp() == nil {
		return StorableMeasurement{}, fmt.Errorf("measurement: nil timestamp")
	}
	if err := m.GetTimestamp().CheckValid(); err != nil {
		return StorableMeasurement{}, err
	}

	sm := StorableMeasurement{
		DeviceID:  m.GetDeviceId(),
		Timestamp: m.GetTimestamp().AsTime(),
	}

	// The generated protobuf code uses a pointer to tspb.Timestamp, but in StorableMeasurement
	// we use golang's time.Time. If the protobuf field is nil then use the zero value of time.Time.
	// cloud.google.com/go/datastore calls IsZero() on time.Time values so omitempty does work.
	if pbUploadTimestamp := m.GetUploadTimestamp(); pbUploadTimestamp != nil {
		if err := pbUploadTimestamp.CheckValid(); err != nil {
			return StorableMeasurement{}, err
		}

		sm.UploadTimestamp = pbUploadTimestamp.AsTime()
	}

	if m.GetTemp() != nil {
		v := m.GetTemp().GetValue()
		sm.Temp = &v
	}

	if m.GetPm1() != nil {
		v := m.GetPm1().GetValue()
		sm.PM1 = &v
	}

	if m.GetPm25() != nil {
		v := m.GetPm25().GetValue()
		sm.PM25 = &v
	}

	if m.GetPm4() != nil {
		v := m.GetPm4().GetValue()
		sm.PM4 = &v
	}

	if m.GetPm10() != nil {
		v := m.GetPm10().GetValue()
		sm.PM10 = &v
	}

	if m.GetRh() != nil {
		v := m.GetRh().GetValue()
		sm.RH = &v
	}

	if m.GetVocIndex() != nil {
		v := m.GetVocIndex().GetValue()
		sm.VOCIndex = &v
	}

	if m.GetNoxIndex() != nil {
		v := m.GetNoxIndex().GetValue()
		sm.NOxIndex = &v
	}

	if m.GetHcho() != nil {
		v := m.GetHcho().GetValue()
		sm.HCHO = &v
	}

	if m.GetCo2() != nil {
		v := m.GetCo2().GetValue()
		sm.CO2 = &v
	}

	return sm, nil
}

// NewMeasurement converts a StorableMeasurement into the generated Measurement type,
// converting time.Time values into the protobuf-specific timestamp type.
// IMPORTANT: Keep up to date with the generated Measurement type (from measurement.proto).
func NewMeasurement(sm *StorableMeasurement) (*mpb.Measurement, error) {
	// Enforce a non-zero timestamp.
	if sm.Timestamp.IsZero() {
		return nil, ErrZeroTimestamp
	}

	m := &mpb.Measurement{
		DeviceId:  sm.DeviceID,
		Timestamp: tspb.New(sm.Timestamp),
	}

	// The upload timestamp may be the zero timestamp. If it is, then the upload timestamp
	// should be nil in the generated Measurement type.
	if !sm.UploadTimestamp.IsZero() {
		m.UploadTimestamp = tspb.New(sm.UploadTimestamp)
	}

	if sm.Temp != nil {
		m.Temp = wpb.Float(*sm.Temp)
	}

	if sm.PM1 != nil {
		m.Pm1 = wpb.Float(*sm.PM1)
	}

	if sm.PM25 != nil {
		m.Pm25 = wpb.Float(*sm.PM25)
	}

	if sm.PM4 != nil {
		m.Pm4 = wpb.Float(*sm.PM4)
	}

	if sm.PM10 != nil {
		m.Pm10 = wpb.Float(*sm.PM10)
	}

	if sm.RH != nil {
		m.Rh = wpb.Float(*sm.RH)
	}

	if sm.VOCIndex != nil {
		m.VocIndex = wpb.Float(*sm.VOCIndex)
	}

	if sm.NOxIndex != nil {
		m.NoxIndex = wpb.Float(*sm.NOxIndex)
	}

	if sm.HCHO != nil {
		m.Hcho = wpb.Float(*sm.HCHO)
	}

	if sm.CO2 != nil {
		m.Co2 = wpb.Float(*sm.CO2)
	}

	return m, nil
}

// DBKey returns a string key suitable for Datastore. It promotes Device ID and timestamp into the key.
func (sm *StorableMeasurement) DBKey() string {
	return strings.Join([]string{sm.DeviceID, sm.Timestamp.Format(time.RFC3339)}, keySep)
}

// ValueMap returns a map from metric key to value, which may be nil.
func (sm StorableMeasurement) ValueMap() map[metric.Key]*float32 {
	return map[metric.Key]*float32{
		metric.Temp:     sm.Temp,
		metric.PM1:      sm.PM1,
		metric.PM25:     sm.PM25,
		metric.PM4:      sm.PM4,
		metric.PM10:     sm.PM10,
		metric.RH:       sm.RH,
		metric.VOCIndex: sm.VOCIndex,
		metric.NOxIndex: sm.NOxIndex,
		metric.HCHO:     sm.HCHO,
		metric.CO2:      sm.CO2,
	}
}

func (sm *StorableMeasurement) FillDerivedMetrics() {
	if sm.PM25 != nil {
		v := float32(aqi.PM25(*sm.PM25))
		sm.AQI = &v
	}
}

func (sm StorableMeasurement) String() string {
	delay := ""
	if !sm.UploadTimestamp.IsZero() {
		delay = fmt.Sprintf(" (%v upload delay)", sm.UploadTimestamp.Sub(sm.Timestamp))
	}

	strs := []string{}
	for key, v := range sm.ValueMap() {
		if v == nil {
			continue
		}

		info := metric.All[key]
		strs = append(strs, fmt.Sprintf("%s=%.3f%s", info.Name, *v, info.Unit))
	}
	sort.Strings(strs)

	if len(strs) == 0 {
		strs = append(strs, "[no measurements]")
	}

	return fmt.Sprintf("%s %s %s%s", sm.DeviceID, strings.Join(strs, ", "), sm.Timestamp.Format(time.RFC3339), delay)
}
