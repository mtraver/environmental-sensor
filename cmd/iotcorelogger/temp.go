package main

import (
	"time"

	"github.com/mtraver/mcp9808"
)

func readTempMulti(samples int, interval time.Duration) ([]float32, error) {
	sensor, err := mcp9808.New()
	if err != nil {
		return []float32{}, err
	}
	defer sensor.Close()

	if err = sensor.Check(); err != nil {
		return []float32{}, err
	}

	temps := make([]float32, samples)
	for i := 0; i < samples; i++ {
		temp, err := sensor.ReadTemp()
		if err != nil {
			return temps, err
		}

		temps[i] = temp
		if i < samples-1 {
			time.Sleep(interval)
		}
	}

	return temps, nil
}

func readTemp() (float32, error) {
	temps, err := readTempMulti(1, 0)
	if err != nil {
		return 0, err
	}

	return temps[0], nil
}
