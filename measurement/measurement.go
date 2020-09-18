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
	Temp *float32 `json:"temp,omitempty" datastore:"temp,omitempty" metric:"temp" unit:"°C"`
	PM25 *float32 `json:"pm25,omitempty" datastore:"pm25,omitempty" metric:"PM2.5" unit:"μg/m³"`
	PM10 *float32 `json:"pm10,omitempty" datastore:"pm10,omitempty" metric:"PM10" unit:"μg/m³"`
	RH   *float32 `json:"rh,omitempty" datastore:"rh,omitempty" metric:"RH" unit:"%"`

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
		v := m.GetTemp().GetValue()
		temp = &v
	}

	var pm25 *float32
	if m.GetPm25() != nil {
		v := m.GetPm25().GetValue()
		pm25 = &v
	}

	var pm10 *float32
	if m.GetPm10() != nil {
		v := m.GetPm10().GetValue()
		pm10 = &v
	}

	var rh *float32
	if m.GetRh() != nil {
		v := m.GetRh().GetValue()
		rh = &v
	}

	return StorableMeasurement{
		DeviceID:        m.GetDeviceId(),
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            temp,
		PM25:            pm25,
		PM10:            pm10,
		RH:              rh,
	}, nil
}

// NewMeasurement converts a StorableMeasurement into the generated Measurement type,
// converting time.Time values into the protobuf-specific timestamp type.
// IMPORTANT: Keep up to date with the generated Measurement type (from measurement.proto).
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

	var pm25 *wpb.FloatValue
	if sm.PM25 != nil {
		pm25 = wpb.Float(*sm.PM25)
	}

	var pm10 *wpb.FloatValue
	if sm.PM10 != nil {
		pm10 = wpb.Float(*sm.PM10)
	}

	var rh *wpb.FloatValue
	if sm.RH != nil {
		rh = wpb.Float(*sm.RH)
	}

	return mpb.Measurement{
		DeviceId:        sm.DeviceID,
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            temp,
		Pm25:            pm25,
		Pm10:            pm10,
		Rh:              rh,
	}, nil
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

// StringValueMap returns a map from metric name (as defined in struct tags) to
// string-formatted values including the unit. Nil fields are not included.
func (sm StorableMeasurement) StringValueMap() map[string]string {
	m := make(map[string]string)

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
		m[metric] = fmt.Sprintf("%.3f%s", *f, unit)
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
