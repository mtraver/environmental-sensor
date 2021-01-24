package db

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

func mustTimestampProto(t time.Time) *tspb.Timestamp {
	pbts := tspb.New(t)
	if err := pbts.CheckValid(); err != nil {
		panic(err)
	}

	return pbts
}

var (
	testTimestamp = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
	pbTimestamp   = mustTimestampProto(testTimestamp)

	testMeasurement = mpb.Measurement{
		DeviceId:  "foo",
		Timestamp: pbTimestamp,
		Temp:      wpb.Float(18.5),
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
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("temp", 18.5).SetTime(testTimestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("pm25", 12.0).SetTime(testTimestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("pm10", 20.0).SetTime(testTimestamp),
				influxdb2.NewPointWithMeasurement("stat").AddTag("device", "foo").AddField("rh", 55.0).SetTime(testTimestamp),
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

			if diff := cmp.Diff(got, c.want, cmp.AllowUnexported(write.Point{})); diff != "" {
				t.Errorf("Unexpected result (-got +want):\n%s", diff)
			}
		})
	}

}
