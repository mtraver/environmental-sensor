package measurementpbutil

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	testTimestamp = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
	pbTimestamp   = mustTimestampProto(testTimestamp)

	testTimestamp2 = time.Date(2018, time.March, 25, 14, 40, 0, 0, time.UTC)
	pbTimestamp2   = mustTimestampProto(testTimestamp2)
)

func mustTimestampProto(t time.Time) *timestamp.Timestamp {
	pbts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}

	return pbts
}

func getMeasurement(t *testing.T, deviceID string) *mpb.Measurement {
	t.Helper()
	return &mpb.Measurement{
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

func TestString(t *testing.T) {
	cases := []struct {
		name string
		m    mpb.Measurement
		want string
	}{
		{"empty", mpb.Measurement{}, " 0.000°C 0001-01-01T00:00:00Z"},
		{"no_upload_timestamp",
			mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: pbTimestamp,
				Temp:      18.3748,
			},
			"foo 18.375°C 2018-03-25T00:00:00Z",
		},
		{"upload_timestamp",
			mpb.Measurement{
				DeviceId:        "foo",
				Timestamp:       pbTimestamp,
				UploadTimestamp: pbTimestamp2,
				Temp:            18.3748,
			},
			"foo 18.375°C 2018-03-25T00:00:00Z (14h40m0s upload delay)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := String(c.m)
			if got != c.want {
				t.Errorf("Got %q, want %q", got, c.want)
			}
		})
	}
}
