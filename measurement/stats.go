package measurement

import (
	"math"

	"github.com/mtraver/environmental-sensor/metric"
)

func Mean(measurements []StorableMeasurement) map[metric.Key]float32 {
	sums := make(map[metric.Key]float64)
	counts := make(map[metric.Key]int)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if v == nil {
				continue
			}

			if _, ok := sums[k]; !ok {
				sums[k] = 0.0
			}
			if _, ok := counts[k]; !ok {
				counts[k] = 0
			}

			sums[k] += float64(*v)
			counts[k]++
		}
	}

	means := make(map[metric.Key]float32)
	for k, v := range sums {
		means[k] = float32(v / float64(counts[k]))
	}

	return means
}

func StdDev(measurements []StorableMeasurement) map[metric.Key]float32 {
	avg := Mean(measurements)

	sums := make(map[metric.Key]float64)
	counts := make(map[metric.Key]int)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if v == nil {
				continue
			}

			if _, ok := sums[k]; !ok {
				sums[k] = 0.0
			}
			if _, ok := counts[k]; !ok {
				counts[k] = 0
			}

			sums[k] += math.Pow(float64(*v-avg[k]), 2)
			counts[k]++
		}
	}

	devs := make(map[metric.Key]float32)
	for k, v := range sums {
		devs[k] = float32(math.Sqrt(v / float64(counts[k])))
	}

	return devs
}

func Min(measurements []StorableMeasurement) map[metric.Key]float32 {
	x := make(map[metric.Key]float32)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if v == nil {
				continue
			}

			if _, ok := x[k]; !ok {
				x[k] = math.MaxFloat32
			}

			if *v < x[k] {
				x[k] = *v
			}
		}
	}

	return x
}

func Max(measurements []StorableMeasurement) map[metric.Key]float32 {
	x := make(map[metric.Key]float32)
	for _, sm := range measurements {
		vals := sm.ValueMap()
		if len(vals) == 0 {
			continue
		}

		for k, v := range vals {
			if v == nil {
				continue
			}

			if _, ok := x[k]; !ok {
				x[k] = -math.MaxFloat32
			}

			if *v > x[k] {
				x[k] = *v
			}
		}
	}

	return x
}
