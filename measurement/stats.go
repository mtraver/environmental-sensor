package measurement

import (
	"math"
)

func Mean(measurements []StorableMeasurement) float32 {
	var sum float64
	var count int
	for _, m := range measurements {
		if m.Temp == nil {
			continue
		}

		sum += float64(*m.Temp)
		count++
	}

	return float32(sum / float64(count))
}

func StdDev(measurements []StorableMeasurement) float32 {
	avg := Mean(measurements)

	var sum float64
	var count int
	for _, m := range measurements {
		if m.Temp == nil {
			continue
		}

		sum += math.Pow(float64(*m.Temp-avg), 2)
		count++
	}

	return float32(math.Sqrt(sum / float64(count)))
}

func Min(measurements []StorableMeasurement) float32 {
	var x float32 = math.MaxFloat32
	for _, m := range measurements {
		if m.Temp == nil {
			continue
		}

		if *m.Temp < x {
			x = *m.Temp
		}
	}

	return x
}

func Max(measurements []StorableMeasurement) float32 {
	var x float32 = -math.MaxFloat32
	for _, m := range measurements {
		if m.Temp == nil {
			continue
		}

		if *m.Temp > x {
			x = *m.Temp
		}
	}

	return x
}
