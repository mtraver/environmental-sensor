package measurement

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
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

func TestStorableMeasurementString(t *testing.T) {
	cases := []struct {
		name string
		m    StorableMeasurement
		want string
	}{
		{"empty", StorableMeasurement{}, " 0.000°C 0001-01-01T00:00:00Z"},
		{"no_upload_timestamp",
			StorableMeasurement{
				DeviceId:  "foo",
				Timestamp: testTimestamp,
				Temp:      18.3748,
			},
			"foo 18.375°C 2018-03-25T00:00:00Z",
		},
		{"upload_timestamp",
			StorableMeasurement{
				DeviceId:        "foo",
				Timestamp:       testTimestamp,
				UploadTimestamp: testTimestamp2,
				Temp:            18.3748,
			},
			"foo 18.375°C 2018-03-25T00:00:00Z (14h40m0s upload delay)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.m.String()
			if got != c.want {
				t.Errorf("Got %q, want %q", got, c.want)
			}
		})
	}
}

func TestNewStorableMeasurement(t *testing.T) {
	deviceID := "foo"
	var temp float32 = 18.5

	cases := []struct {
		name  string
		m     *Measurement
		want  StorableMeasurement
		valid bool
	}{
		{"valid_no_upload_timestamp",
			&Measurement{
				DeviceId:  deviceID,
				Timestamp: pbTimestamp,
				Temp:      temp,
			},
			StorableMeasurement{
				DeviceId:  deviceID,
				Timestamp: testTimestamp,
				Temp:      temp,
			},
			true,
		},
		{"valid_with_upload_timestamp",
			&Measurement{
				DeviceId:        deviceID,
				Timestamp:       pbTimestamp,
				UploadTimestamp: pbTimestamp2,
				Temp:            temp,
			},
			StorableMeasurement{
				DeviceId:        deviceID,
				Timestamp:       testTimestamp,
				UploadTimestamp: testTimestamp2,
				Temp:            temp,
			},
			true,
		},
		{"nil_timestamp",
			&Measurement{
				DeviceId:  deviceID,
				Timestamp: nil,
				Temp:      temp,
			},
			StorableMeasurement{},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NewStorableMeasurement(c.m)
			if err != nil && c.valid {
				t.Errorf("Unexpected error: %v", err)
				return
			} else if err == nil && !c.valid {
				t.Errorf("Expected error, got no error")
				return
			} else if err != nil && !c.valid {
				// For this case the test has passed. We don't enforce any contract on the first
				// return value of NewStorableMeasurement when the error is non-nil.
				return
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Got %v, want %v", got, c.want)
			}
		})
	}
}

func TestDBKey(t *testing.T) {
	m := StorableMeasurement{
		DeviceId:  "foo",
		Timestamp: time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC),
		Temp:      18.5,
	}

	expected := "foo#2018-03-25T00:00:00Z"
	key := m.DBKey()
	if key != expected {
		t.Errorf("Incorrect DB key. Expected %q, got %q", expected, key)
	}
}

func TestCacheKeyLatest(t *testing.T) {
	expected := "foo#latest"
	key := CacheKeyLatest("foo")
	if key != expected {
		t.Errorf("Incorrect key. Expected %q, got %q", expected, key)
	}
}

func TestMeasurementMapToJSON(t *testing.T) {
	cases := []struct {
		name         string
		measurements map[string][]StorableMeasurement
		want         string
	}{
		{"none", map[string][]StorableMeasurement{}, "null"},
		{"empty", map[string][]StorableMeasurement{"foo": {}}, `[{"id":"foo","values":[]}]`},
		{"many", map[string][]StorableMeasurement{
			"foo": {
				{
					DeviceId:  "foo",
					Timestamp: time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceId:  "foo",
					Timestamp: time.Date(2018, time.March, 26, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceId:  "foo",
					Timestamp: time.Date(2018, time.March, 27, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
			},
			"bar": {
				{
					DeviceId:  "bar",
					Timestamp: time.Date(2018, time.March, 25, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceId:  "bar",
					Timestamp: time.Date(2018, time.March, 26, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceId:  "bar",
					Timestamp: time.Date(2018, time.March, 27, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
			},
		}, `[{"id":"bar","values":[{"timestamp":1521997200000,"temp":18.5},{"timestamp":1522083600000,"temp":18.5},{"timestamp":1522170000000,"temp":18.5}]},{"id":"foo","values":[{"timestamp":1521936000000,"temp":18.5},{"timestamp":1522022400000,"temp":18.5},{"timestamp":1522108800000,"temp":18.5}]}]`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			marshalledJSON, err := MeasurementMapToJSON(c.measurements)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if string(marshalledJSON) != c.want {
				t.Errorf("Want %q, got %q", c.want, string(marshalledJSON))
			}
		})
	}
}

func getMeasurement(t *testing.T, deviceID string) Measurement {
	t.Helper()
	return Measurement{
		DeviceId:  deviceID,
		Timestamp: pbTimestamp,
		Temp:      18.5,
	}
}

func TestDeviceID(t *testing.T) {
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
			if valid := m.Validate() == nil; valid != c.valid {
				t.Errorf("Measurement valid is %v, expected %v", valid, c.valid)
			}
		})
	}
}
