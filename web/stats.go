package main

import (
	"github.com/mtraver/environmental-sensor/measurement"
)

type Stats struct {
	Min    float32
	Max    float32
	Mean   float32
	StdDev float32
}

func summaryStats(measurements map[string][]measurement.StorableMeasurement) map[string]Stats {
	stats := make(map[string]Stats)
	for deviceID, m := range measurements {
		// TODO(mtraver) Make this generic.
		stats[deviceID] = Stats{
			Min:    measurement.Min(m)["temp"],
			Max:    measurement.Max(m)["temp"],
			Mean:   measurement.Mean(m)["temp"],
			StdDev: measurement.StdDev(m)["temp"],
		}
	}

	return stats
}
