// Package measurementpbutil provides utility functions for working with the generated protobuf type Measurement.
package measurementpbutil

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	protoc_descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

func String(m mpb.Measurement) string {
	var timestamp time.Time
	if m.GetTimestamp() != nil {
		var err error
		timestamp, err = ptypes.Timestamp(m.GetTimestamp())
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}
	}

	delay := ""
	if m.GetUploadTimestamp() != nil {
		uploadts, err := ptypes.Timestamp(m.GetUploadTimestamp())
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		}

		delay = fmt.Sprintf(" (%v upload delay)", uploadts.Sub(timestamp))
	}

	return fmt.Sprintf("%s %.3f°C %s%s", m.GetDeviceId(), m.Temp, timestamp.Format(time.RFC3339), delay)
}

// Validate validates each field of the Measurement against an optional regex provided in the .proto file.
// It returns nil if all fields are valid and no other errors occurred along the way. Example of how to
// provide a regex in a .proto file:
//   string device_id = 1 [(regex) = "^[a-z][a-z0-9+.%~_-]{2,254}$"];
func Validate(m *mpb.Measurement) error {
	_, msgDesc := descriptor.ForMessage(m)
	for _, f := range msgDesc.GetField() {
		options := f.GetOptions()
		if options == nil {
			continue
		}

		regexExt, err := proto.GetExtension(options, mpb.E_Regex)
		if err == proto.ErrMissingExtension {
			// The field doesn't have the regex extension, so nothing to validate
			continue
		} else if err != nil {
			return err
		}

		// A field validated with a regex must be a string, but reflect.Value's
		// String method doesn't panic if the value isn't a string, so check it.
		field := getField(m, f)
		if field.Kind() != reflect.String {
			panic(fmt.Sprintf("Field is not a string: %q", f.GetName()))
		}

		regex := regexp.MustCompile(*regexExt.(*string))
		if !regex.MatchString(field.String()) {
			return fmt.Errorf("Field failed regex validation. Field: %q Value: %q Regex: %q", f.GetName(), field.String(), regex)
		}
	}

	return nil
}

// getField returns the field of the Measurement corresponding to the given FieldDescriptorProto.
// This is a rather ugly operation since it's not supported by the protobuf API; it has to be done
// via reflection. See the following GitHub issues, the first of which provided the code on which
// this method is based:
//
//   "What is the idiomatic way to get the corresponding struct field for a FieldDescriptorProto?"
//     https://github.com/golang/protobuf/issues/457
//
//   "proto: make the Message interface behaviorally complete"
//     https://github.com/golang/protobuf/issues/364
func getField(m *mpb.Measurement, fd *protoc_descriptor.FieldDescriptorProto) reflect.Value {
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
		panic(fmt.Sprintf("Cannot find struct field for proto field name %q, number/tag %d", fd.GetName(), fd.GetNumber()))
	}

	return field
}
