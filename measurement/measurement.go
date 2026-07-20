package measurement

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mtraver/environmental-sensor/aqi"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

// Used for separating substrings in database keys. The octothorpe is fine for this because
// device IDs and timestamps, the two things most likely to be used in keys, can't contain it.
const keySep = "#"

var (
	// ErrZeroTimestamp is returned from NewMeasurement if the StorableMeasurement's timestamp is the zero timestamp.
	ErrZeroTimestamp = fmt.Errorf("measurement: timestamp cannot be nil")

	// jsonKeyToMetric and metricNameToMetric do what they say on the tin in order to allow
	// lookup of metric names and units from the key that's used in serialized measurements
	// (e.g., look it up from the frontend).
	jsonKeyToMetric    = make(map[string]Metric)
	metricNameToMetric = make(map[string]Metric)
)

func init() {
	sm := StorableMeasurement{}

	v := reflect.ValueOf(sm)
	for i := 0; i < v.NumField(); i++ {
		// The metric tag must be present. It marks a field as a measurement.
		metric := getMetric(v, i)
		if metric == "" {
			continue
		}

		// The field must be a float32 pointer.
		if _, ok := getValue(v, i); !ok {
			continue
		}

		jsonKey := strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0]

		// Ensure that no keys are duplicated. Fail in glorious fashion if they are.
		if _, ok := jsonKeyToMetric[jsonKey]; ok {
			panic(fmt.Sprintf("duplicate JSON key %q", jsonKey))
		}
		if _, ok := metricNameToMetric[metric]; ok {
			panic(fmt.Sprintf("duplicate metric name %q", metric))
		}

		m := Metric{
			Name:  metric,
			Abbrv: jsonKey,
			Unit:  getUnit(v, i),
		}

		jsonKeyToMetric[jsonKey] = m
		metricNameToMetric[metric] = m
	}
}

type Metric struct {
	Name  string
	Abbrv string
	Unit  string
}

func GetMetric(nameOrKey string) (Metric, bool) {
	if metric, ok := jsonKeyToMetric[nameOrKey]; ok {
		return metric, ok
	}

	if metric, ok := metricNameToMetric[nameOrKey]; ok {
		return metric, ok
	}

	return Metric{}, false
}

// StorableMeasurement is equivalent to the generated Measurement type but it contains
// no protobuf-specific types. It therefore can be marshaled to JSON and written to
// Datastore. StorableMeasurement is marshaled to JSON in order to pass data to the
// frontend. Timestamp is handled specially in MarshalJSON.
// IMPORTANT: Keep up to date with the generated Measurement type (from measurement.proto).
type StorableMeasurement struct {
	DeviceID        string    `json:"-" datastore:"device_id"`
	Timestamp       time.Time `json:"-" datastore:"timestamp"`
	UploadTimestamp time.Time `json:"-" datastore:"upload_timestamp,omitempty"`

	// These metrics are the raw values reported by sensors. They must match the
	// metrics defined in the generated Measurement type (from measurement.proto).
	Temp     *float32 `json:"temp,omitempty" datastore:"temp,omitempty" metric:"temp" unit:"°C"`
	PM1      *float32 `json:"pm1,omitempty" datastore:"pm1,omitempty" metric:"PM1.0" unit:"μg/m³"`
	PM25     *float32 `json:"pm25,omitempty" datastore:"pm25,omitempty" metric:"PM2.5" unit:"μg/m³"`
	PM4      *float32 `json:"pm4,omitempty" datastore:"pm4,omitempty" metric:"PM4" unit:"μg/m³"`
	PM10     *float32 `json:"pm10,omitempty" datastore:"pm10,omitempty" metric:"PM10" unit:"μg/m³"`
	RH       *float32 `json:"rh,omitempty" datastore:"rh,omitempty" metric:"RH" unit:"%"`
	VOCIndex *float32 `json:"voc_index,omitempty" datastore:"voc_index,omitempty" metric:"VOCIndex" unit:""`
	NOxIndex *float32 `json:"nox_index,omitempty" datastore:"nox_index,omitempty" metric:"NOₓIndex" unit:""`
	HCHO     *float32 `json:"hcho,omitempty" datastore:"hcho,omitempty" metric:"HCHO" unit:"ppb"`
	CO2      *float32 `json:"co2,omitempty" datastore:"co2,omitempty" metric:"CO₂" unit:"ppm"`

	// These metrics are derived from the raw values. They're not stored in the database
	// (the `datastore` tag is set to "-") but they are passed to the frontend in JSON form.
	// These values are populated by the FillDerivedMetrics method.
	AQI *float32 `json:"aqi,omitempty" datastore:"-" metric:"AQI" unit:""`
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
func NewMeasurement(sm *StorableMeasurement) (mpb.Measurement, error) {
	// Enforce a non-zero timestamp.
	if sm.Timestamp.IsZero() {
		return mpb.Measurement{}, ErrZeroTimestamp
	}

	m := mpb.Measurement{
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

// ValueMap returns a map from metric name (as defined in struct tags) to values.
// Nil fields are not included.
func (sm StorableMeasurement) ValueMap() map[string]float32 {
	m := make(map[string]float32)

	v := reflect.ValueOf(sm)
	for i := 0; i < v.NumField(); i++ {
		// The metric tag must be present. It marks a field as a measurement.
		metric := getMetric(v, i)
		if metric == "" {
			continue
		}

		// The field must be a float32 pointer.
		f, ok := getValue(v, i)
		if !ok || f == nil {
			continue
		}

		m[metric] = *f
	}

	return m
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
	v := reflect.ValueOf(sm)
	for i := 0; i < v.NumField(); i++ {
		// The metric tag must be present. It marks a field as a measurement.
		metric := getMetric(v, i)
		if metric == "" {
			continue
		}

		// The field must be a float32 pointer.
		f, ok := getValue(v, i)
		if !ok || f == nil {
			continue
		}

		// There may be no unit tag, which is fine.
		unit := getUnit(v, i)
		strs = append(strs, fmt.Sprintf("%s=%.3f%s", metric, *f, unit))
	}
	sort.Strings(strs)

	if len(strs) == 0 {
		strs = append(strs, "[no measurements]")
	}

	return fmt.Sprintf("%s %s %s%s", sm.DeviceID, strings.Join(strs, ", "), sm.Timestamp.Format(time.RFC3339), delay)
}

func getMetric(v reflect.Value, i int) string {
	return v.Type().Field(i).Tag.Get("metric")
}

func getUnit(v reflect.Value, i int) string {
	return v.Type().Field(i).Tag.Get("unit")
}

func getValue(v reflect.Value, i int) (*float32, bool) {
	f, ok := v.Field(i).Interface().(*float32)
	return f, ok
}
