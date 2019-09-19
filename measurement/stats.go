package measurement

import "math"

func Mean(measurements []StorableMeasurement) float32 {
	var sum float64
	for _, m := range measurements {
		sum += float64(m.Temp)
	}

	return float32(sum / float64(len(measurements)))
}

func StdDev(measurements []StorableMeasurement) float32 {
	avg := Mean(measurements)

	var sum float64
	for _, m := range measurements {
		sum += math.Pow(float64(m.Temp-avg), 2)
	}

	return float32(math.Sqrt(sum / float64(len(measurements))))
}

func Min(measurements []StorableMeasurement) float32 {
	var x float32 = math.MaxFloat32
	for _, m := range measurements {
		if m.Temp < x {
			x = m.Temp
		}
	}

	return x
}

func Max(measurements []StorableMeasurement) float32 {
	var x float32 = -math.MaxFloat32
	for _, m := range measurements {
		if m.Temp > x {
			x = m.Temp
		}
	}

	return x
}
