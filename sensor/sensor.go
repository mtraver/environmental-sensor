package sensor

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	mu      sync.RWMutex
	sensors = make(map[string]Sensor)
)

type Sensor interface {
	// OnRegister performs one-time sensor initialization tasks. It is called once
	// when the sensor is registered.
	OnRegister() error

	// OnRemove performs one-time sensor shutdown/cleanup tasks. It is called once
	// when the sensor is removed from the registry.
	OnRemove() error

	// RunSetupJob performs recurring, cron-scheduled sensor setup tasks.
	RunSetupJob() error

	// RunSenseJob performs recurring, cron-scheduled sensing. It should query the
	// sensor for measurements and set the appropriate field(s) on the given
	// [*mpb.Measurement]. Note that the same [*mpb.Measurement] is passed to each
	// sensor specified in a job so that each sensor may fill in those measurements
	// that it supports.
	RunSenseJob(m *mpb.Measurement) error

	// RunShutdownJob performs recurring, cron-scheduled sensor shutdown/cleanup tasks.
	RunShutdownJob() error
}

// Register adds a sensor to the registry.
func Register(name string, s Sensor) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := sensors[name]; ok {
		return fmt.Errorf("sensor with name already registered: %q", name)
	}

	if err := s.OnRegister(); err != nil {
		return fmt.Errorf("failed to register sensor with name %q: %w", name, err)
	}

	sensors[name] = s

	return nil
}

// Remove removes a sensor from the registry.
func Remove(name string) error {
	mu.Lock()
	defer mu.Unlock()

	s, ok := sensors[name]
	if !ok {
		return nil
	}

	if err := s.OnRemove(); err != nil {
		return fmt.Errorf("failed to remove sensor with name %q: %w", name, err)
	}

	delete(sensors, name)

	return nil
}

// RemoveAll removes all sensors from the registry.
func RemoveAll() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []error
	for _, name := range slices.Collect(maps.Keys(sensors)) {
		s := sensors[name]

		if err := s.OnRemove(); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove sensor with name %q: %w", name, err))
		} else {
			delete(sensors, name)
		}
	}

	return errors.Join(errs...)
}

// Get gets a sensor by name. It returns nil if no sensor with the given name is found.
func Get(name string) Sensor {
	mu.RLock()
	defer mu.RUnlock()

	s, ok := sensors[name]
	if !ok {
		return nil
	}

	return s
}

// Names returns the names of all sensors in the registry.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()

	return slices.Collect(maps.Keys(sensors))
}

func UsesI2C(name string) bool {
	switch name {
	case "mcp9808", "sen6x":
		return true
	default:
		return false
	}
}
