// Package aqi computes US EPA Air Quality Index values for PM2.5 and PM10.
package aqi

import "math"

type bucket struct {
	lowerLimit float32
	upperLimit float32
	lowerIndex float32
	upperIndex float32
}

var (
	pm25Buckets = []bucket{
		{0, 12.0, 0, 50},
		{12.1, 35.4, 51, 100},
		{35.5, 55.4, 101, 150},
		{55.5, 150.4, 151, 200},
		{150.5, 250.4, 201, 300},
		{250.5, 350.4, 301, 400},
		{350.5, 500.4, 401, 500},
	}

	pm10Buckets = []bucket{
		{0, 54, 0, 50},
		{55, 154, 51, 100},
		{155, 254, 101, 150},
		{255, 354, 151, 200},
		{355, 424, 201, 300},
		{425, 504, 301, 400},
		{505, 604, 401, 500},
	}
)

func scale(pm float32, b bucket) float32 {
	return ((b.upperIndex-b.lowerIndex)/(b.upperLimit-b.lowerLimit))*(pm-b.lowerLimit) + b.lowerIndex
}

func aqi(pm float32, buckets []bucket) int {
	if pm < buckets[0].lowerLimit {
		return 0
	}

	for _, b := range buckets {
		if pm <= b.upperLimit {
			return int(math.Round(float64(scale(pm, b))))
		}
	}

	return 500
}

func PM25(pm float32) int {
	return aqi(pm, pm25Buckets)
}

func PM10(pm float32) int {
	return aqi(pm, pm10Buckets)
}

func String(aqi int) string {
	if aqi <= 50 {
		return "Good"
	} else if aqi <= 100 {
		return "Moderate"
	} else if aqi <= 150 {
		return "Unhealthy for Sensitive Groups"
	} else if aqi <= 200 {
		return "Unhealthy"
	} else if aqi <= 300 {
		return "Very Unhealthy"
	} else {
		return "Hazardous"
	}
}
