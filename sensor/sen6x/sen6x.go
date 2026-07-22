package sen6x

import (
	"sync"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/sen6x"
)

type SEN6x struct {
	dev      *sen6x.Dev
	i2cBusMu *sync.Mutex
}

func New(model sen6x.Model, bus i2c.BusCloser, i2cBusMu *sync.Mutex) (*SEN6x, error) {
	i2cBusMu.Lock()
	defer i2cBusMu.Unlock()

	return &SEN6x{
		dev:      sen6x.New(bus, model),
		i2cBusMu: i2cBusMu,
	}, nil
}

func (s *SEN6x) OnRegister() error {
	s.i2cBusMu.Lock()
	defer s.i2cBusMu.Unlock()

	return s.dev.StartContinuousMeasurement()
}

func (s *SEN6x) OnRemove() error {
	s.i2cBusMu.Lock()
	defer s.i2cBusMu.Unlock()

	return s.dev.StopMeasurement()
}

func (s *SEN6x) RunSetupJob() error {
	return nil
}

func (s *SEN6x) RunSenseJob(m *mpb.Measurement) error {
	s.i2cBusMu.Lock()
	defer s.i2cBusMu.Unlock()

	vals, err := s.dev.ReadMeasuredValues()
	if err != nil {
		return err
	}

	if vals.PM1 != nil {
		m.Pm1 = wpb.Float(*vals.PM1)
	}
	if vals.PM25 != nil {
		m.Pm25 = wpb.Float(*vals.PM25)
	}
	if vals.PM4 != nil {
		m.Pm4 = wpb.Float(*vals.PM4)
	}
	if vals.PM10 != nil {
		m.Pm10 = wpb.Float(*vals.PM10)
	}
	if vals.RH != nil {
		m.Rh = wpb.Float(*vals.RH)
	}
	if vals.Temp != nil {
		m.Temp = wpb.Float(*vals.Temp)
	}
	if vals.VOC != nil {
		m.VocIndex = wpb.Float(*vals.VOC)
	}
	if vals.NOx != nil {
		m.NoxIndex = wpb.Float(*vals.NOx)
	}
	if vals.CO2 != nil {
		m.Co2 = wpb.Float(float32(*vals.CO2))
	}
	if vals.HCHO != nil {
		m.Hcho = wpb.Float(*vals.HCHO)
	}

	return nil
}

func (s *SEN6x) RunShutdownJob() error {
	return nil
}
