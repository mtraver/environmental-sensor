package measurementpbutil

import (
	"testing"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/testutil"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func getMeasurement(t *testing.T, deviceID string) *mpb.Measurement {
	t.Helper()
	return &mpb.Measurement{
		DeviceId:  deviceID,
		Timestamp: testutil.TimestampProto,
		Temp:      wpb.Float(18.3748),
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid", "foo+.%~_-0123", false},
		{"empty", "", true},
		{"short", "a", true},
		{"non_alpha_short", "7abcd", true},
		{"illegal_chars", "foo`!@#$^&*()={}[]<>,?/|\\':;", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := getMeasurement(t, c.id)
			err := Validate(m)

			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		name string
		m    *mpb.Measurement
		want string
	}{
		{
			"empty",
			&mpb.Measurement{},
			" [no measurements] 0001-01-01T00:00:00Z",
		},
		{
			"no upload timestamp",
			&mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: testutil.TimestampProto,
				Temp:      wpb.Float(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z",
		},
		{
			"with upload timestamp",
			&mpb.Measurement{
				DeviceId:        "foo",
				Timestamp:       testutil.TimestampProto,
				UploadTimestamp: testutil.TimestampProto2,
				Temp:            wpb.Float(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z (14h40m0s upload delay)",
		},
		{
			"multiple measurements set",
			&mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: testutil.TimestampProto,
				Temp:      wpb.Float(18.3748),
				Pm25:      wpb.Float(12.0),
				Pm10:      wpb.Float(20.0),
				Rh:        wpb.Float(57.0),
			},
			"foo PM10=20.000μg/m³, PM2.5=12.000μg/m³, RH=57.000%, temp=18.375°C 2018-03-25T00:00:00Z",
		},
		{
			"all measurements set",
			testutil.FullyPopulatedMeasurementProto(),
			"foo CO₂=425.000ppm, HCHO=2.000ppb, NOₓIndex=75.000, PM1.0=1.000μg/m³, PM10=20.000μg/m³, PM2.5=12.000μg/m³, PM4=15.000μg/m³, RH=57.000%, VOCIndex=80.000, temp=18.375°C 2018-03-25T00:00:00Z",
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
