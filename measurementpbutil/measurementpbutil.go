// Package measurementpbutil provides utility functions for working with the generated protobuf type Measurement.
package measurementpbutil

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func String(m mpb.Measurement) string {
	var timestamp time.Time
	if m.GetTimestamp() != nil {
		timestamp = m.GetTimestamp().AsTime()
	}

	delay := ""
	if m.GetUploadTimestamp() != nil {
		uploadts := m.GetUploadTimestamp().AsTime()

		delay = fmt.Sprintf(" (%v upload delay)", uploadts.Sub(timestamp))
	}

	strs := []string{}
	r := m.ProtoReflect()
	r.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		// Get the MeasurementOptions extension and verify that it's not nil
		// and that the metric field is not empty. It's ok for the unit field
		// to be empty.
		options := fd.Options()
		if !proto.HasExtension(options, mpb.E_MeasurementOptions) {
			return true
		}
		opt := proto.GetExtension(options, mpb.E_MeasurementOptions).(*mpb.MeasurementOptions)
		if opt == nil || opt.GetMetric() == "" {
			return true
		}

		// The field must be a Message.
		msg, ok := v.Interface().(protoreflect.Message)
		if !ok {
			return true
		}

		// Furthermore, the Message must be a FloatValue pointer.
		fv, ok := msg.Interface().(*wpb.FloatValue)
		if !ok || fv == nil {
			return true
		}

		strs = append(strs, fmt.Sprintf("%s=%.3f%s", opt.GetMetric(), fv.GetValue(), opt.GetUnit()))
		return true
	})
	sort.Strings(strs)

	if len(strs) == 0 {
		strs = append(strs, "[no measurements]")
	}

	return fmt.Sprintf("%s %s %s%s", m.GetDeviceId(), strings.Join(strs, ", "), timestamp.Format(time.RFC3339), delay)
}

// Validate validates each field of the Measurement against an optional regex provided in the .proto file.
// It returns nil if all fields are valid and no other errors occurred along the way. Example of how to
// provide a regex in a .proto file:
//
// string device_id = 1 [(regex) = "^[a-z][a-z0-9+.%~_-]{2,254}$"];
func Validate(m *mpb.Measurement) error {
	// First validate any required fields because protoreflect.Message.Range only
	// iterates over "populated" fields. From the documentation on Range:
	//
	//   "Range iterates over every populated field in an undefined order,
	//   calling f for each field descriptor and value encountered."
	//
	// But for proto3 what does "populated" even mean? There's no longer any notion
	// of required fields, and the zero value of a field may be a perfectly valid
	// value. Why should the fact that the field contains the zero value mean that
	// the caller doesn't get to operate on it?
	//
	// In this case the regex could fail the empty string if it so chose, but we don't
	// get to do that because a field containing the empty string isn't "populated".
	if m.GetDeviceId() == "" {
		return fmt.Errorf("measurementpbutil: field \"device_id\" is required")
	}

	var retErr error
	r := m.ProtoReflect()
	r.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		options := fd.Options()
		if !proto.HasExtension(options, mpb.E_Regex) {
			// The field doesn't have the regex extension, so nothing to validate.
			return true
		}

		// A field validated with a regex must be a string.
		if fd.Kind() != protoreflect.StringKind {
			panic(fmt.Sprintf("measurementpbutil: field is not a string: %q", fd.Name()))
		}

		// We know the value of this extension to be a string.
		ext := proto.GetExtension(options, mpb.E_Regex)
		re := regexp.MustCompile(ext.(string))

		if !re.MatchString(v.String()) {
			retErr = fmt.Errorf("measurementpbutil: field failed regex validation. Field: %q Value: %q Regex: %q", fd.Name(), v.String(), re)
		}

		return true
	})

	return retErr
}
