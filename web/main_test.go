package main

import (
	"testing"
	"time"

	"github.com/mtraver/environmental-sensor/measurement"
)

func TestMeasurementMapToJSON(t *testing.T) {
	cases := []struct {
		name         string
		measurements map[string][]measurement.StorableMeasurement
		want         string
	}{
		{"none", map[string][]measurement.StorableMeasurement{}, "null"},
		{"empty", map[string][]measurement.StorableMeasurement{"foo": {}}, `[{"id":"foo","values":[]}]`},
		{"many", map[string][]measurement.StorableMeasurement{
			"foo": {
				{
					DeviceID:  "foo",
					Timestamp: time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceID:  "foo",
					Timestamp: time.Date(2018, time.March, 26, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceID:  "foo",
					Timestamp: time.Date(2018, time.March, 27, 0, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
			},
			"bar": {
				{
					DeviceID:  "bar",
					Timestamp: time.Date(2018, time.March, 25, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceID:  "bar",
					Timestamp: time.Date(2018, time.March, 26, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
				{
					DeviceID:  "bar",
					Timestamp: time.Date(2018, time.March, 27, 17, 0, 0, 0, time.UTC),
					Temp:      18.5,
				},
			},
		}, `[{"id":"bar","values":[{"timestamp":1521997200000,"temp":18.5},{"timestamp":1522083600000,"temp":18.5},{"timestamp":1522170000000,"temp":18.5}]},{"id":"foo","values":[{"timestamp":1521936000000,"temp":18.5},{"timestamp":1522022400000,"temp":18.5},{"timestamp":1522108800000,"temp":18.5}]}]`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			marshalledJSON, err := measurementMapToJSON(c.measurements)
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