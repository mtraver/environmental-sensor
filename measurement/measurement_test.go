package measurement

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
)

var (
	testTimestamp = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
)

func TestToStorableMeasurement(t *testing.T) {
	deviceID := "foo"
	var temp float32 = 18.5

	pbTimestamp, _ := ptypes.TimestampProto(testTimestamp)
	m := Measurement{
		DeviceId:  deviceID,
		Timestamp: pbTimestamp,
		Temp:      temp,
	}

	s, err := m.ToStorableMeasurement()
	if err != nil {
		t.Errorf("Error in ToStorableMeasurement: %v", err)
		return
	}

	if s.DeviceId != deviceID {
		t.Errorf("Incorrect devie ID. Expected %q, got %q", deviceID, s.DeviceId)
	}

	if s.Timestamp != testTimestamp {
		t.Errorf("Incorrect timestamp. Expected %v, got %v", testTimestamp,
			s.Timestamp)
	}

	if s.Temp != temp {
		t.Errorf("Incorrect temp. Expected %v, got %v", temp, s.Temp)
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

func TestMeasurementMapToJSONEmpty(t *testing.T) {
	measurements := map[string][]StorableMeasurement{}
	marshalledJSON, err := MeasurementMapToJSON(measurements)
	if err != nil {
		t.Errorf("Error on valid input: %v", err)
	}

	expected := "null"
	if string(marshalledJSON) != expected {
		t.Errorf("Incorrect JSON. Expected %q, got %q", expected,
			string(marshalledJSON))
	}
}

func TestMeasurementMapToJSONNoMeasurements(t *testing.T) {
	measurements := map[string][]StorableMeasurement{
		"foo": {},
	}

	marshalledJSON, err := MeasurementMapToJSON(measurements)
	if err != nil {
		t.Errorf("Error on valid input: %v", err)
	}

	expected := `[{"id":"foo","values":[]}]`
	if string(marshalledJSON) != expected {
		t.Errorf("Incorrect JSON. Expected %q, got %q", expected,
			string(marshalledJSON))
	}
}

func TestMeasurementMapToJSON(t *testing.T) {
	measurements := map[string][]StorableMeasurement{
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
	}

	marshalledJSON, err := MeasurementMapToJSON(measurements)
	if err != nil {
		t.Errorf("Error on valid input: %v", err)
	}

	expected := `[{"id":"bar","values":[{"timestamp":1521997200000,"temp":18.5},{"timestamp":1522083600000,"temp":18.5},{"timestamp":1522170000000,"temp":18.5}]},{"id":"foo","values":[{"timestamp":1521936000000,"temp":18.5},{"timestamp":1522022400000,"temp":18.5},{"timestamp":1522108800000,"temp":18.5}]}]`
	if string(marshalledJSON) != expected {
		t.Errorf("Incorrect JSON. Expected %q, got %q", expected,
			string(marshalledJSON))
	}
}

func getMeasurement(t *testing.T, deviceID string) Measurement {
	t.Helper()
	pbTimestamp, _ := ptypes.TimestampProto(testTimestamp)
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
