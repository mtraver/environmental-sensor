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
		Pm1:             float32PtrToFloat64Ptr(sm.PM1),
		Pm25:            float32PtrToFloat64Ptr(sm.PM25),
		Pm4:             float32PtrToFloat64Ptr(sm.PM4),
		Pm10:            float32PtrToFloat64Ptr(sm.PM10),
		Aqi:             float32PtrToFloat64Ptr(sm.AQI),
		Rh:              float32PtrToFloat64Ptr(sm.RH),
		VocIndex:        float32PtrToFloat64Ptr(sm.VOCIndex),
		NoxIndex:        float32PtrToFloat64Ptr(sm.NOxIndex),
		Hcho:            float32PtrToFloat64Ptr(sm.HCHO),
		Co2:             float32PtrToFloat64Ptr(sm.CO2),
	}
}

func float32PtrToFloat64Ptr(f *float32) *float64 {
	if f == nil {
		return nil
	}

	f64 := float64(*f)
	return &f64
}
