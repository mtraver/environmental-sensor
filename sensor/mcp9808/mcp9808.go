package mcp9808

import (
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/mcp9808"
)

const (
	numSamples     = 3
	sampleInterval = 1
)

type MCP9808 struct {
	dev *mcp9808.Dev
}

func New(bus i2c.BusCloser) (*MCP9808, error) {
	d, err := mcp9808.New(bus, &mcp9808.DefaultOpts)
	if err != nil {
		return nil, err
	}

	return &MCP9808{
		dev: d,
	}, nil
}

func (s *MCP9808) Init() error {
	return nil
}

func (s *MCP9808) Sense(m *mpb.Measurement) error {
	temps, err := s.readTempMulti(numSamples, time.Duration(sampleInterval)*time.Second)
	if err != nil {
		return err
	}

	m.Temp = wpb.Float(mean(temps))
	return nil
}

func (s *MCP9808) Shutdown() error {
	return nil
}

func (s *MCP9808) readTempMulti(samples int, interval time.Duration) ([]physic.Temperature, error) {
	temps := make([]physic.Temperature, samples)
	for i := 0; i < samples; i++ {
		temp, err := s.dev.SenseTemp()
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

func mean(s []physic.Temperature) float32 {
	var sum float64
	for _, t := range s {
		sum += t.Celsius()
	}

	return float32(sum / float64(len(s)))
}
