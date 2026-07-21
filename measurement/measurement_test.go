package measurement

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/metric"
	"github.com/mtraver/environmental-sensor/testutil"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func floatPtr(f float32) *float32 {
	return &f
}

var (
	fullyPopulatedStorableMeasurement = StorableMeasurement{
		DeviceID:  "foo",
		Timestamp: testutil.Timestamp,
		Temp:      floatPtr(18.3748),
		PM1:       floatPtr(1.0),
		PM25:      floatPtr(12.0),
		PM4:       floatPtr(15.0),
		PM10:      floatPtr(20.0),
		RH:        floatPtr(57.0),
		VOCIndex:  floatPtr(80),
		NOxIndex:  floatPtr(75),
		HCHO:      floatPtr(2),
		CO2:       floatPtr(425),
	}

	// These cases are used to test conversion in both directions between the generated
	// Measurement type and StorableMeasurement.
	conversionCases = []struct {
		name  string
		m     mpb.Measurement
		sm    StorableMeasurement
		valid bool
	}{
		{
			"valid no upload timestamp",
			mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: testutil.TimestampProto,
				Temp:      wpb.Float(18.3748),
				Pm25:      wpb.Float(12.0),
				Pm10:      wpb.Float(20.0),
				Rh:        wpb.Float(55.0),
			},
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testutil.Timestamp,
				Temp:      floatPtr(18.3748),
				PM25:      floatPtr(12.0),
				PM10:      floatPtr(20.0),
				RH:        floatPtr(55.0),
			},
			true,
		},
		{
			"valid with upload timestamp",
			mpb.Measurement{
				DeviceId:        "foo",
				Timestamp:       testutil.TimestampProto,
				UploadTimestamp: testutil.TimestampProto2,
				Temp:            wpb.Float(18.3748),
			},
			StorableMeasurement{
				DeviceID:        "foo",
				Timestamp:       testutil.Timestamp,
				UploadTimestamp: testutil.Timestamp2,
				Temp:            floatPtr(18.3748),
			},
			true,
		},
		{
			"all measurements set",
			testutil.FullyPopulatedMeasurementProto(),
			fullyPopulatedStorableMeasurement,
			true,
		},
		{
			"nil timestamp",
			mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: nil,
				Temp:      wpb.Float(18.3748),
			},
			StorableMeasurement{},
			false,
		},
	}
)

func TestStorableMeasurementString(t *testing.T) {
	cases := []struct {
		name string
		sm   StorableMeasurement
		want string
	}{
		{
			"empty",
			StorableMeasurement{},
			" [no measurements] 0001-01-01T00:00:00Z",
		},
		{
			"no upload timestamp",
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testutil.Timestamp,
				Temp:      floatPtr(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z",
		},
		{
			"with upload timestamp",
			StorableMeasurement{
				DeviceID:        "foo",
				Timestamp:       testutil.Timestamp,
				UploadTimestamp: testutil.Timestamp2,
				Temp:            floatPtr(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z (14h40m0s upload delay)",
		},
		{
			"two measurements set",
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testutil.Timestamp,
				Temp:      floatPtr(18.3748),
				RH:        floatPtr(57.0),
			},
			"foo RH=57.000%, temp=18.375°C 2018-03-25T00:00:00Z",
		},
		{
			"all measurements set",
			fullyPopulatedStorableMeasurement,
			"foo CO₂=425.000ppm, HCHO=2.000ppb, NOₓIndex=75.000, PM1.0=1.000μg/m³, PM10=20.000μg/m³, PM2.5=12.000μg/m³, PM4=15.000μg/m³, RH=57.000%, VOCIndex=80.000, temp=18.375°C 2018-03-25T00:00:00Z",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fmt.Sprintf("%v", c.sm)
			if got != c.want {
				t.Errorf("Got %q, want %q", got, c.want)
			}
		})
	}
}

func TestStorableMeasurementDBKey(t *testing.T) {
	sm := StorableMeasurement{
		DeviceID:  "foo",
		Timestamp: time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC),
		Temp:      floatPtr(18.3748),
	}

	want := "foo#2018-03-25T00:00:00Z"
	if got := sm.DBKey(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStorableMeasurementValueMap(t *testing.T) {
	cases := []struct {
		name string
		sm   StorableMeasurement
		want map[metric.Key]*float32
	}{
		{
			"empty",
			StorableMeasurement{},
			map[metric.Key]*float32{
				metric.Temp:     nil,
				metric.PM1:      nil,
				metric.PM25:     nil,
				metric.PM4:      nil,
				metric.PM10:     nil,
				metric.RH:       nil,
				metric.VOCIndex: nil,
				metric.NOxIndex: nil,
				metric.HCHO:     nil,
				metric.CO2:      nil,
			},
		},
		{
			"all measurements set",
			fullyPopulatedStorableMeasurement,
			map[metric.Key]*float32{
				metric.Temp:     floatPtr(18.3748),
				metric.PM1:      floatPtr(1.0),
				metric.PM25:     floatPtr(12.0),
				metric.PM4:      floatPtr(15.0),
				metric.PM10:     floatPtr(20.0),
				metric.RH:       floatPtr(57.0),
				metric.VOCIndex: floatPtr(80.0),
				metric.NOxIndex: floatPtr(75.0),
				metric.HCHO:     floatPtr(2.0),
				metric.CO2:      floatPtr(425.0),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.sm.ValueMap()
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestNewStorableMeasurement(t *testing.T) {
	for _, tc := range conversionCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewStorableMeasurement(&tc.m)
			if err != nil && tc.valid {
				t.Errorf("unexpected error: %v", err)
				return
			} else if err == nil && !tc.valid {
				t.Errorf("Expected error, got no error")
				return
			} else if err != nil && !tc.valid {
				// For this case the test has passed. We don't enforce any contract on the first
				// return value of NewStorableMeasurement when the error is non-nil.
				return
			}

			if diff := cmp.Diff(got, tc.sm); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestNewMeasurement(t *testing.T) {
	for _, tc := range conversionCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewMeasurement(&tc.sm)
			if err != nil && tc.valid {
				t.Errorf("unexpected error: %v", err)
				return
			} else if err == nil && !tc.valid {
				t.Errorf("Expected error, got no error")
				return
			} else if err != nil && !tc.valid {
				// For this case the test has passed. We don't enforce any contract on the first
				// return value of NewMeasurement when the error is non-nil.
				return
			}

			if diff := cmp.Diff(got, tc.m, cmpopts.IgnoreUnexported(mpb.Measurement{}, tspb.Timestamp{}, wpb.FloatValue{})); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
