package sensor

import (
	"iter"
	"maps"
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	mu      sync.RWMutex
	sensors = make(map[string]Sensor)
)

type Sensor interface {
	// Init performs any sensor-specific initialization.
	Init() error
	// Sense queries the sensor for measurements and sets the appropriate
	// field(s) in the given Measurement. This is so that the same Measurement
	// may be passed to a series of sensors that each measure different things.
	Sense(m *mpb.Measurement) error
	// Shutdown performs an sensor-specific shutdown or cleanup operations.
	Shutdown() error
}

// Register adds a sensor to the registry.
func Register(name string, s Sensor) {
	mu.Lock()
	defer mu.Unlock()

	sensors[name] = s
}

// Remove removes a sensor from the registry.
func Remove(name string) {
	mu.Lock()
	defer mu.Unlock()

	delete(sensors, name)
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
func Names() iter.Seq[string] {
	mu.RLock()
	defer mu.RUnlock()

	return maps.Keys(sensors)
}

func UsesI2C(name string) bool {
	switch name {
	case "mcp9808":
		return true
	default:
		return false
	}
}
