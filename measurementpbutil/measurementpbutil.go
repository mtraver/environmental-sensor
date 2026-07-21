// Package measurementpbutil provides utility functions for working with the generated protobuf type Measurement.
package measurementpbutil

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/metric"
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

	values := map[metric.Key]*wpb.FloatValue{
		metric.Temp:     m.GetTemp(),
		metric.PM1:      m.GetPm1(),
		metric.PM25:     m.GetPm25(),
		metric.PM4:      m.GetPm4(),
		metric.PM10:     m.GetPm10(),
		metric.RH:       m.GetRh(),
		metric.VOCIndex: m.GetVocIndex(),
		metric.NOxIndex: m.GetNoxIndex(),
		metric.HCHO:     m.GetHcho(),
		metric.CO2:      m.GetCo2(),
	}

	var strs []string
	for key, v := range values {
		if v == nil {
			continue
		}

		info := metric.All[key]
		strs = append(strs, fmt.Sprintf("%s=%.3f%s", info.Name, v.GetValue(), info.Unit))
	}
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
