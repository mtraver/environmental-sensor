package sds011

import (
	"time"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/sds011"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	numSamples     = 3
	sampleInterval = 1
)

type SDS011 struct {
	dev sds011.Dev
}

func New(name string) (*SDS011, error) {
	d, err := sds011.New(name)
	if err != nil {
		return nil, err
	}

	return &SDS011{
		dev: d,
	}, nil
}

func (s *SDS011) Init() error {
	if err := s.dev.Wake(); err != nil {
		return nil
	}
	if err := s.dev.SetMode(sds011.ModeQuery); err != nil {
		return nil
	}
	return nil
}

func (s *SDS011) Sense(m *mpb.Measurement) error {
	values := make([]sds011.Measurement, numSamples)
	for i := 0; i < numSamples; i++ {
		v, err := s.dev.Sense()
		if err != nil {
			return err
		}

		values[i] = v
		if i < numSamples-1 {
			time.Sleep(sampleInterval * time.Second)
		}
	}

	avg := mean(values)
	m.Pm25 = wpb.Float(avg.PM25)
	m.Pm10 = wpb.Float(avg.PM10)
	return nil
}

func (s *SDS011) Shutdown() error {
	return s.dev.Sleep()
}

func mean(m []sds011.Measurement) sds011.Measurement {
	res := sds011.Measurement{}
	if len(m) == 0 {
		return res
	}

	for _, v := range m {
		res.PM25 += v.PM25
		res.PM10 += v.PM10
	}

	res.PM25 = res.PM25 / float32(len(m))
	res.PM10 = res.PM10 / float32(len(m))
	return res
}
