package main

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	Jobs []JobSpec `json:"jobs"`
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func (c *Config) sensors() []string {
	seen := make(map[string]struct{})
	for _, job := range c.Jobs {
		for _, sensor := range job.Sensors {
			seen[sensor] = struct{}{}
		}
	}

	sensors := make([]string, 0, len(seen))
	for sensor := range seen {
		sensors = append(sensors, sensor)
	}
	return sensors
}

func (c *Config) validate() error {
	if c == nil {
		return nil
	}

	for i, jobSpec := range c.Jobs {
		if jobSpec.Cronspec == "" {
			return fmt.Errorf("job %d has no cronspec", i)
		}

		if _, ok := allJobTypes[jobSpec.Operation]; !ok {
			return fmt.Errorf("job %d has invalid operation: %q", i, jobSpec.Operation)
		}

		if len(jobSpec.Sensors) == 0 {
			return fmt.Errorf("job %d has no sensors", i)
		}

		if hasDuplicates(jobSpec.Sensors) {
			return fmt.Errorf("job %d has duplicate sensors: %v", i, jobSpec.Sensors)
		}
	}

	return nil
}
