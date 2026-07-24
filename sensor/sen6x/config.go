package sen6x

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"periph.io/x/devices/v3/sen6x"
)

type Config struct {
	AltitudeMeters     *uint16                                  `json:"altitudeM,omitempty"`
	AmbientPressureHPa *uint16                                  `json:"pressureHPa,omitempty"`
	CO2AutoCalibration *bool                                    `json:"co2AutoCalibration,omitempty"`
	VOCTuning          *sen6x.VOCNOxAlgorithmTuningParameters   `json:"vocTuning,omitempty"`
	NOxTuning          *sen6x.VOCNOxAlgorithmTuningParameters   `json:"noxTuning,omitempty"`
	TempAcceleration   *sen6x.TemperatureAccelerationParameters `json:"tempAcceleration,omitempty"`
	TempOffset         *sen6x.TemperatureOffsetParameters       `json:"tempOffset,omitempty"`
}

func (c Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func (s *SEN6x) Configure(raw json.RawMessage) (err error) {
	if raw == nil {
		return nil
	}

	var cfg Config
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("invalid sen6x config: %w", err)
	}

	log.Printf("%s: configure:\n%s", Name, cfg)

	s.i2cBusMu.Lock()
	defer s.i2cBusMu.Unlock()

	// If we do change any config values (we may not if current config matches desired
	// config) we'll need to put the sensor in idle mode because writing config is only
	// available in that mode. If we did put the sensor in idle mode then put it back in
	// measurement mode when we're done. The error returned by the deferred call to
	// startMeasurement is captured via the named return value.
	shouldRestartMeasurement := s.mode == modeMeasurement
	defer func() {
		if shouldRestartMeasurement && s.mode != modeMeasurement {
			err = errors.Join(err, s.startMeasurement())
		}
	}()

	if s.applied.AltitudeMeters, err = applyIfChanged(s.applied.AltitudeMeters, cfg.AltitudeMeters, func(alt uint16) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetSensorAltitude(alt)
	}); err != nil {
		return err
	}

	if s.applied.AmbientPressureHPa, err = applyIfChanged(s.applied.AmbientPressureHPa, cfg.AmbientPressureHPa, func(p uint16) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetAmbientPressure(p)
	}); err != nil {
		return err
	}

	if s.applied.CO2AutoCalibration, err = applyIfChanged(s.applied.CO2AutoCalibration, cfg.CO2AutoCalibration, func(autoCal bool) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetCO2SensorAutomaticSelfCalibration(autoCal)
	}); err != nil {
		return err
	}

	if s.applied.VOCTuning, err = applyIfChanged(s.applied.VOCTuning, cfg.VOCTuning, func(params sen6x.VOCNOxAlgorithmTuningParameters) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetVOCAlgorithmTuningParameters(params)
	}); err != nil {
		return err
	}

	if s.applied.NOxTuning, err = applyIfChanged(s.applied.NOxTuning, cfg.NOxTuning, func(params sen6x.VOCNOxAlgorithmTuningParameters) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetNOxAlgorithmTuningParameters(params)
	}); err != nil {
		return err
	}

	if s.applied.TempAcceleration, err = applyIfChanged(s.applied.TempAcceleration, cfg.TempAcceleration, func(params sen6x.TemperatureAccelerationParameters) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetTemperatureAccelerationParameters(params)
	}); err != nil {
		return err
	}

	if s.applied.TempOffset, err = applyIfChanged(s.applied.TempOffset, cfg.TempOffset, func(params sen6x.TemperatureOffsetParameters) error {
		if err := s.stopMeasurement(); err != nil {
			return err
		}
		return s.dev.SetTemperatureOffsetParameters(params)
	}); err != nil {
		return err
	}

	return nil
}

func (s *SEN6x) CurrentConfig() Config {
	return s.applied
}

// readConfig reads the current config from the device. It puts the sensor in idle mode
// (calls StopMeasurement) because reading these config values is only available in
// idle mode. The caller is responsible for putting the sensor back in measurement
// mode if required.
func (s *SEN6x) readConfig() (Config, error) {
	var cfg Config

	if err := s.stopMeasurement(); err != nil {
		return cfg, err
	}

	if alt, err := s.dev.GetSensorAltitude(); err != nil {
		return cfg, err
	} else {
		cfg.AltitudeMeters = &alt
	}

	if p, err := s.dev.GetAmbientPressure(); err != nil {
		return cfg, err
	} else {
		cfg.AmbientPressureHPa = &p
	}

	if co2AutoCal, err := s.dev.GetCO2SensorAutomaticSelfCalibration(); err != nil {
		return cfg, err
	} else {
		cfg.CO2AutoCalibration = &co2AutoCal
	}

	if vocTuning, err := s.dev.GetVOCAlgorithmTuningParameters(); err != nil {
		return cfg, err
	} else {
		cfg.VOCTuning = vocTuning
	}

	if noxTuning, err := s.dev.GetNOxAlgorithmTuningParameters(); err != nil {
		return cfg, err
	} else {
		cfg.NOxTuning = noxTuning
	}

	// There are no commands to get either temperature acceleration parameters or
	// temperature offset parameters.

	return cfg, nil
}

// applyIfChanged calls set(*newVal) if newVal is non-nil and differs from current.
// It returns the value that should be recorded as applied afterward.
func applyIfChanged[T comparable](current, newVal *T, set func(T) error) (*T, error) {
	if newVal == nil {
		return current, nil
	}

	if current != nil && *current == *newVal {
		return current, nil
	}

	if err := set(*newVal); err != nil {
		return current, err
	}

	return newVal, nil
}
