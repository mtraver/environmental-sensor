package sensor

import (
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	sensorsMu sync.Mutex
	sensors   map[string]Sensor
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

// Register adds a Sensor to the set of available sensors.
func Register(name string, s Sensor) {
	sensorsMu.Lock()
	defer sensorsMu.Unlock()

	if sensors == nil {
		sensors = make(map[string]Sensor)
	}
	sensors[name] = s
}

// Get gets a sensor by name. It returns false if no sensor with the given name is found.
func Get(name string) (Sensor, bool) {
	sensorsMu.Lock()
	defer sensorsMu.Unlock()

	if sensors == nil {
		sensors = make(map[string]Sensor)
	}

	s, ok := sensors[name]
	return s, ok
}
