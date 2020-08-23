package measurement

import (
	"math"
)

func Mean(measurements []StorableMeasurement) float32 {
	var sum float64
	var count int
	for _, sm := range measurements {
		if sm.Temp == nil {
			continue
		}

		sum += float64(*sm.Temp)
		count++
	}

	return float32(sum / float64(count))
}

func StdDev(measurements []StorableMeasurement) float32 {
	avg := Mean(measurements)

	var sum float64
	var count int
	for _, sm := range measurements {
		if sm.Temp == nil {
			continue
		}

		sum += math.Pow(float64(*sm.Temp-avg), 2)
		count++
	}

	return float32(math.Sqrt(sum / float64(count)))
}

func Min(measurements []StorableMeasurement) float32 {
	var x float32 = math.MaxFloat32
	for _, sm := range measurements {
		if sm.Temp == nil {
			continue
		}

		if *sm.Temp < x {
			x = *sm.Temp
		}
	}

	return x
}

func Max(measurements []StorableMeasurement) float32 {
	var x float32 = -math.MaxFloat32
	for _, sm := range measurements {
		if sm.Temp == nil {
			continue
		}

		if *sm.Temp > x {
			x = *sm.Temp
		}
	}

	return x
}
