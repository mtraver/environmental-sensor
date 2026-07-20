package db

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/testutil"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	testMeasurement = mpb.Measurement{
		DeviceId:  "foo",
		Timestamp: testutil.TimestampProto,
		Temp:      wpb.Float(18.3748),
		Pm25:      wpb.Float(12.0),
		Pm10:      wpb.Float(20.0),
		Rh:        wpb.Float(55.0),
	}
)

func TestNewInfluxDBPoints(t *testing.T) {
	cases := []struct {
		name string
		m    mpb.Measurement
		want []*write.Point
	}{
		{
			name: "many",
			m:    testMeasurement,
			want: []*write.Point{
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("temp", 18.3748).SetTime(testutil.Timestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("pm25", 12.0).SetTime(testutil.Timestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("pm10", 20.0).SetTime(testutil.Timestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("rh", 55.0).SetTime(testutil.Timestamp),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := newInfluxDBPoints(&testMeasurement)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Sort slices before comparing them. Sort by the key of the first field, which is
			// brittle, but since in this case we only have one field per Point it works.
			sort.Slice(got, func(i, j int) bool {
				return got[i].FieldList()[0].Key < got[j].FieldList()[0].Key
			})
			sort.Slice(c.want, func(i, j int) bool {
				return c.want[i].FieldList()[0].Key < c.want[j].FieldList()[0].Key
			})

			if diff := cmp.Diff(got, c.want, cmp.AllowUnexported(write.Point{}), cmpopts.EquateApprox(0, 0.0001)); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}

}
