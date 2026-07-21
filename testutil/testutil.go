package testutil

import (
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	Timestamp      = time.Date(2018, time.March, 25, 0, 0, 0, 0, time.UTC)
	TimestampProto = mustTimestampProto(Timestamp)

	Timestamp2      = time.Date(2018, time.March, 25, 14, 40, 0, 0, time.UTC)
	TimestampProto2 = mustTimestampProto(Timestamp2)
)

func FullyPopulatedMeasurementProto() *mpb.Measurement {
	return &mpb.Measurement{
		DeviceId:  "foo",
		Timestamp: TimestampProto,
		Temp:      wpb.Float(18.3748),
		Pm1:       wpb.Float(1.0),
		Pm25:      wpb.Float(12.0),
		Pm4:       wpb.Float(15.0),
		Pm10:      wpb.Float(20.0),
		Rh:        wpb.Float(57.0),
		VocIndex:  wpb.Float(80),
		NoxIndex:  wpb.Float(75),
		Hcho:      wpb.Float(2),
		Co2:       wpb.Float(425),
	}
}

func mustTimestampProto(t time.Time) *tspb.Timestamp {
	pbts := tspb.New(t)
	if err := pbts.CheckValid(); err != nil {
		panic(err)
	}

	return pbts
}

func floatPtr(f float32) *float32 {
	return &f
}
