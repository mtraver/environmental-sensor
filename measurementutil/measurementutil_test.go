package measurementutil

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/mtraver/environmental-sensor/measurement"
)

var (
	testTimestamp = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
	pbTimestamp   = mustTimestampProto(testTimestamp)
)

func mustTimestampProto(t time.Time) *timestamp.Timestamp {
	pbts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}

	return pbts
}

func getMeasurement(t *testing.T, deviceID string) *measurement.Measurement {
	t.Helper()
	return &measurement.Measurement{
		DeviceId:  deviceID,
		Timestamp: pbTimestamp,
		Temp:      18.5,
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid", "foo+.%~_-0123", true},
		{"empty", "", false},
		{"short", "a", false},
		{"non_alpha_short", "7abcd", false},
		{"illegal_chars", "foo`!@#$^&*()={}[]<>,?/|\\':;", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := getMeasurement(t, c.id)
			if valid := Validate(m) == nil; valid != c.valid {
				t.Errorf("Measurement valid is %v, expected %v", valid, c.valid)
			}
		})
	}
}
