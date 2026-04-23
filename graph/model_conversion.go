package graph

import (
	"github.com/mtraver/environmental-sensor/graph/model"
	"github.com/mtraver/environmental-sensor/measurement"
)

func storableMeasurementToGQLMeasurement(sm measurement.StorableMeasurement) *model.Measurement {
	return &model.Measurement{
		DeviceID:        sm.DeviceID,
		Timestamp:       timeToGQLTimestamp(sm.Timestamp),
		UploadTimestamp: timeToGQLTimestamp(sm.UploadTimestamp),
		Temp:            float32PtrToFloat64Ptr(sm.Temp),
		Pm25:            float32PtrToFloat64Ptr(sm.PM25),
		Pm10:            float32PtrToFloat64Ptr(sm.PM10),
		Rh:              float32PtrToFloat64Ptr(sm.RH),
		Aqi:             float32PtrToFloat64Ptr(sm.AQI),
	}
}

func float32PtrToFloat64Ptr(f *float32) *float64 {
	if f == nil {
		return nil
	}

	f64 := float64(*f)
	return &f64
}
