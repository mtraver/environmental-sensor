package sds011

import (
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/sds011"
	wpb "google.golang.org/protobuf/types/known/wrapperspb"
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
	v, err := s.dev.Sense()
	if err != nil {
		return err
	}

	m.Pm25 = wpb.Float(v.PM25)
	m.Pm10 = wpb.Float(v.PM10)
	return nil
}

func (s *SDS011) Shutdown() error {
	return s.dev.Sleep()
}
