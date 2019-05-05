package measurement

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	protoc_descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes"
)

// Used for separating substrings in database and cache keys. The octothorpe is
// fine for this because device IDs and timestamps, the two things most likely
// to be used in keys, can't contain it.
const keySep = "#"

// IMPORTANT: Keep up to date with the automatically-generated Measurement type
type StorableMeasurement struct {
	DeviceId        string    `json:"device_id,omitempty" datastore:"device_id"`
	Timestamp       time.Time `json:"timestamp,omitempty" datastore:"timestamp"`
	UploadTimestamp time.Time `json:"upload_timestamp,omitempty" datastore:"upload_timestamp,omitempty"`
	Temp            float32   `json:"temp,omitempty" datastore:"temp"`
}

// IMPORTANT: Keep up to date with the automatically-generated Measurement type.
// This is almost the same as StorableMeasurement except without DeviceId. When
// data is marshaled to JSON for use in the template each record doesn't need to
// include the device ID.
type serializableMeasurement struct {
	// This timestamp is an offset from the epoch in milliseconds
	// (compare to Timestamp in StorableMeasurement).
	Timestamp int64   `json:"timestamp,omitempty" datastore:"timestamp"`
	Temp      float32 `json:"temp,omitempty" datastore:"temp"`
}

// NewStorableMeasurement converts the automatically-generated Measurement type to
// a type that contains no protobuf-specific types. The resulting StorableMeasurement
// can be marshaled to JSON and written to Datastore.
// IMPORTANT: Keep up to date with the automatically-generated Measurement type
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
		DeviceId:        m.GetDeviceId(),
		Timestamp:       timestamp,
		UploadTimestamp: uploadTimestamp,
		Temp:            m.GetTemp(),
	}, nil
}

// getField returns the field of the Measurement corresponding to the given
// FieldDescriptorProto. This is a rather ugly operation as it's not supported
// by the protobuf API; it has to be done via reflection. See these GitHub
// issues, the first of which provided the code on which this method is based:
//   https://github.com/golang/protobuf/issues/457 (What is the idiomatic way
//     to get the corresponding struct field for a FieldDescriptorProto?)
//   https://github.com/golang/protobuf/issues/364 (proto: make the Message
//     interface behaviorally complete)
func (m *Measurement) getField(fd *protoc_descriptor.FieldDescriptorProto) reflect.Value {
	messageVal := reflect.ValueOf(*m)
	props := proto.GetProperties(reflect.TypeOf(m).Elem())

	var field reflect.Value
	for _, p := range props.Prop {
		if int32(p.Tag) == fd.GetNumber() {
			field = messageVal.FieldByName(p.Name)
			break
		}
	}

	if !field.IsValid() {
		// Must be a oneof if not found in the regular fields above
		for _, oneof := range props.OneofTypes {
			if int32(oneof.Prop.Tag) == fd.GetNumber() {
				field = messageVal.Field(oneof.Field).Elem().FieldByName(oneof.Prop.Name)
				break
			}
		}
	}

	if !field.IsValid() {
		panic(fmt.Sprintf("Cannot find struct field for proto field name %q, number/tag %d",
			fd.GetName(), fd.GetNumber()))
	}

	return field
}

func (m *Measurement) Validate() error {
	_, msgDesc := descriptor.ForMessage(m)
	for _, f := range msgDesc.GetField() {
		options := f.GetOptions()
		if options == nil {
			continue
		}

		regexExt, err := proto.GetExtension(options, E_Regex)
		if err == proto.ErrMissingExtension {
			// The field doesn't have the regex extension, so nothing to validate
			continue
		} else if err != nil {
			return err
		}

		// A field validated with a regex must be a string, but reflect.Value's
		// String method doesn't panic if the value isn't a string, so check it.
		field := m.getField(f)
		if field.Kind() != reflect.String {
			panic(fmt.Sprintf("Field is not a string: %q", f.GetName()))
		}

		regex := regexp.MustCompile(*regexExt.(*string))
		if !regex.MatchString(field.String()) {
			return fmt.Errorf("Field failed regex validation. Field: %q Value: %q Regex: %q",
				f.GetName(), field.String(), regex)
		}
	}

	return nil
}

// DBKey returns a string key suitable for Datastore.
// It promotes Device ID and timestamp into the key.
func (m *StorableMeasurement) DBKey() string {
	return strings.Join([]string{m.DeviceId, m.Timestamp.Format(time.RFC3339)}, keySep)
}

// CacheKeyLatest returns the cache key of the latest measurement
// for the given device ID.
func CacheKeyLatest(deviceID string) string {
	return strings.Join([]string{deviceID, "latest"}, keySep)
}

// MeasurementMapToJSON converts a string -> []StorableMeasurement map
// into a marshaled JSON array for use in the template. The JSON is an array
// with one element for each device ID. It's constructed this way, instead of
// as a map where keys are device IDs, because the JavaScript visualization
// package D3 (https://d3js.org/) works better with arrays of data than maps.
func MeasurementMapToJSON(measurements map[string][]StorableMeasurement) ([]byte, error) {
	type dataForTemplate struct {
		ID     string                    `json:"id"`
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
