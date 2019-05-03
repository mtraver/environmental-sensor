package main

import (
	"time"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
)

func readTempMulti(sensor *mcp9808.Dev, samples int, interval time.Duration) ([]physic.Temperature, error) {
	temps := make([]physic.Temperature, samples)
	for i := 0; i < samples; i++ {
		temp, err := sensor.SenseTemp()
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

func readTemp(sensor *mcp9808.Dev) (physic.Temperature, error) {
	temps, err := readTempMulti(sensor, 1, 0)
	if err != nil {
		return 0, err
	}

	return temps[0], nil
}
