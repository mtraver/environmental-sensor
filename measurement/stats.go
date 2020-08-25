package measurement

import (
	"math"
)

func Mean(measurements []StorableMeasurement) map[string]float32 {
	sums := make(map[string]float64)
	counts := make(map[string]int)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if _, ok := sums[k]; !ok {
				sums[k] = 0.0
			}
			if _, ok := counts[k]; !ok {
				counts[k] = 0
			}

			sums[k] += float64(v)
			counts[k]++
		}
	}

	means := make(map[string]float32)
	for k, v := range sums {
		means[k] = float32(v / float64(counts[k]))
	}

	return means
}

func StdDev(measurements []StorableMeasurement) map[string]float32 {
	avg := Mean(measurements)

	sums := make(map[string]float64)
	counts := make(map[string]int)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if _, ok := sums[k]; !ok {
				sums[k] = 0.0
			}
			if _, ok := counts[k]; !ok {
				counts[k] = 0
			}

			sums[k] += math.Pow(float64(v-avg[k]), 2)
			counts[k]++
		}
	}

	devs := make(map[string]float32)
	for k, v := range sums {
		devs[k] = float32(math.Sqrt(v / float64(counts[k])))
	}

	return devs
}

func Min(measurements []StorableMeasurement) map[string]float32 {
	x := make(map[string]float32)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if _, ok := x[k]; !ok {
				x[k] = math.MaxFloat32
			}

			if v < x[k] {
				x[k] = v
			}
		}
	}

	return x
}

func Max(measurements []StorableMeasurement) map[string]float32 {
	x := make(map[string]float32)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if _, ok := x[k]; !ok {
				x[k] = -math.MaxFloat32
			}

			if v > x[k] {
				x[k] = v
			}
		}
	}

	return x
}
