package measurement

import (
	"fmt"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func floatPtr(f float32) *float32 {
	return &f
}

var (
	testTimestamp = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
	pbTimestamp   = mustTimestampProto(testTimestamp)

	testTimestamp2 = time.Date(2018, time.March, 25, 14, 40, 0, 0, time.UTC)
	pbTimestamp2   = mustTimestampProto(testTimestamp2)

	// These cases are used to test conversion in both directions between the generated
	// Measurement type and StorableMeasurement.
	conversionCases = []struct {
		name  string
		m     mpb.Measurement
		sm    StorableMeasurement
		valid bool
	}{
		{"valid_no_upload_timestamp",
			mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: pbTimestamp,
				Temp:      wpb.Float(18.5),
				Pm25:      wpb.Float(12.0),
				Pm10:      wpb.Float(20.0),
				Rh:        wpb.Float(55.0),
			},
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testTimestamp,
				Temp:      floatPtr(18.5),
				PM25:      floatPtr(12.0),
				PM10:      floatPtr(20.0),
				RH:        floatPtr(55.0),
			},
			true,
		},
		{"valid_with_upload_timestamp",
			mpb.Measurement{
				DeviceId:        "foo",
				Timestamp:       pbTimestamp,
				UploadTimestamp: pbTimestamp2,
				Temp:            wpb.Float(18.5),
			},
			StorableMeasurement{
				DeviceID:        "foo",
				Timestamp:       testTimestamp,
				UploadTimestamp: testTimestamp2,
				Temp:            floatPtr(18.5),
			},
			true,
		},
		{"nil_timestamp",
			mpb.Measurement{
				DeviceId:  "foo",
				Timestamp: nil,
				Temp:      wpb.Float(18.5),
			},
			StorableMeasurement{},
			false,
		},
	}
)

func mustTimestampProto(t time.Time) *tspb.Timestamp {
	pbts := tspb.New(t)
	if err := pbts.CheckValid(); err != nil {
		panic(err)
	}

	return pbts
}

func TestStorableMeasurementString(t *testing.T) {
	cases := []struct {
		name string
		sm   StorableMeasurement
		want string
	}{
		{"empty", StorableMeasurement{}, " [no measurements] 0001-01-01T00:00:00Z"},
		{"no_upload_timestamp",
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testTimestamp,
				Temp:      floatPtr(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z",
		},
		{"upload_timestamp",
			StorableMeasurement{
				DeviceID:        "foo",
				Timestamp:       testTimestamp,
				UploadTimestamp: testTimestamp2,
				Temp:            floatPtr(18.3748),
			},
			"foo temp=18.375°C 2018-03-25T00:00:00Z (14h40m0s upload delay)",
		},
		{"multiple_measurements_set",
			StorableMeasurement{
				DeviceID:  "foo",
				Timestamp: testTimestamp,
				Temp:      floatPtr(18.3748),
				PM25:      floatPtr(12.0),
				PM10:      floatPtr(20.0),
				RH:        floatPtr(57.0),
			},
			"foo PM10=20.000μg/m^3, PM2.5=12.000μg/m^3, RH=57.000%, temp=18.375°C 2018-03-25T00:00:00Z",
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

func TestNewStorableMeasurement(t *testing.T) {
	for _, c := range conversionCases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NewStorableMeasurement(&c.m)
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

			if diff := pretty.Compare(got, c.sm); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}

func TestNewMeasurement(t *testing.T) {
	for _, c := range conversionCases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NewMeasurement(&c.sm)
			if err != nil && c.valid {
				t.Errorf("Unexpected error: %v", err)
				return
			} else if err == nil && !c.valid {
				t.Errorf("Expected error, got no error")
				return
			} else if err != nil && !c.valid {
				// For this case the test has passed. We don't enforce any contract on the first
				// return value of NewMeasurement when the error is non-nil.
				return
			}

			if diff := pretty.Compare(got, c.m); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}
}

func TestDBKey(t *testing.T) {
	sm := StorableMeasurement{
		DeviceID:  "foo",
		Timestamp: time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC),
		Temp:      floatPtr(18.5),
	}

	want := "foo#2018-03-25T00:00:00Z"
	if got := sm.DBKey(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
