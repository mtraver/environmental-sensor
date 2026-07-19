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

var (
	deviceIDRegex = regexp.MustCompile(`^[a-z][a-z0-9+.%~_-]{2,254}$`)
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

// Validate validates each field of the Measurement.
func Validate(m *mpb.Measurement) error {
	if m.GetDeviceId() == "" {
		return fmt.Errorf("measurementpbutil: device_id is required")
	}

	if !deviceIDRegex.MatchString(m.GetDeviceId()) {
		return fmt.Errorf("measurementpbutil: device_id failed validation: %q", m.GetDeviceId())
	}

	return nil
}
